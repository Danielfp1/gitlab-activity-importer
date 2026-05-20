package main

import (
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/furmanp/gitlab-activity-importer/internal"
	"github.com/furmanp/gitlab-activity-importer/internal/services"
)

func main() {
	localRepoFlag := flag.Bool("local-repo", false, "Use a local Git repository backup instead of the GitLab API")
	flag.Parse()

	startNow := time.Now()
	err := internal.SetupEnv()
	if err != nil {
		log.Fatalf("Error during loading environmental variables: %v", err)
	}

	localMode := *localRepoFlag || os.Getenv("LOCAL_REPO_PATH") != ""

	if localMode {
		runLocalMode()
	} else {
		runGitLabMode()
	}

	log.Printf("Operation took: %v in total.", time.Since(startNow))
}

// runLocalMode reads commits from a local Git repository and recreates them in
// the destination repository without making any GitLab API calls.
func runLocalMode() {
	log.Println("Using local repository mode")

	repoPath := os.Getenv("LOCAL_REPO_PATH")

	// SOURCE_AUTHOR_EMAIL is the email used to author commits in the source
	// repository (e.g. company GitLab). When not set, COMMITER_EMAIL is used
	// as fallback so that single-email setups continue to work unchanged.
	authorEmail := os.Getenv("SOURCE_AUTHOR_EMAIL")
	if authorEmail == "" {
		authorEmail = os.Getenv("COMMITER_EMAIL")
	}
	log.Printf("Filtering commits by author email: %q", authorEmail)

	since := parseOptionalDate("SINCE_DATE")
	until := parseOptionalDate("UNTIL_DATE")

	commits, err := services.ReadLocalCommits(repoPath, authorEmail, since, until)
	if err != nil {
		log.Fatalf("Error reading local repository: %v", err)
	}
	if len(commits) == 0 {
		log.Printf("No commits found for %q in local repository. Closing the program.", authorEmail)
		return
	}

	importCommits(commits)
}

// runGitLabMode fetches commits via the GitLab API. If the API is unreachable
// and LOCAL_REPO_PATH is set, it falls back to runLocalMode automatically.
func runGitLabMode() {
	gitlabUser, err := services.GetGitlabUser()
	if err != nil {
		localPath := os.Getenv("LOCAL_REPO_PATH")
		if localPath != "" {
			log.Printf("GitLab API unavailable (%v). LOCAL_REPO_PATH is set — falling back to local repository mode.", err)
			runLocalMode()
			return
		}
		log.Fatalf("Error during reading GitLab User data: %v", err)
	}

	projectIds, err := services.GetUsersProjectsIds(gitlabUser.ID)
	if err != nil {
		localPath := os.Getenv("LOCAL_REPO_PATH")
		if localPath != "" {
			log.Printf("GitLab API unavailable (%v). LOCAL_REPO_PATH is set — falling back to local repository mode.", err)
			runLocalMode()
			return
		}
		log.Fatalf("Error during getting users projects: %v", err)
	}
	if len(projectIds) == 0 {
		log.Print("No contributions found for this user. Closing the program.")
		return
	}

	log.Printf("Found contributions in %d projects", len(projectIds))

	commitChannel := make(chan []internal.Commit, len(projectIds))

	var wg sync.WaitGroup
	wg.Add(1)

	var totalCommitsCreated int
	repo := services.OpenOrInitClone()

	err = services.PullLatestChanges(repo)
	if err != nil {
		log.Fatalf("Error pulling latest changes: %v", err)
	}

	go func() {
		defer wg.Done()
		totalCommits := 0
		for commits := range commitChannel {
			if localCommits, err := services.CreateLocalCommit(repo, commits); err == nil {
				totalCommits += localCommits
			} else {
				log.Printf("Error creating local commit: %v", err)
			}
		}
		totalCommitsCreated = totalCommits
		log.Printf("Imported %v commits.\n", totalCommits)
	}()

	services.FetchAllCommits(projectIds, os.Getenv("GITLAB_USERNAME"), commitChannel)

	wg.Wait()

	if totalCommitsCreated > 0 {
		if err := services.PushLocalCommits(repo); err != nil {
			log.Fatalf("Error pushing local commits: %v", err)
		}
		log.Println("Successfully pushed commits to remote repository.")
	} else {
		log.Println("No new commits were created, skipping push operation.")
	}
}

// importCommits opens (or initialises) the destination repository, pulls the
// latest state, writes the provided commits, and pushes the result.
func importCommits(commits []internal.Commit) {
	repo := services.OpenOrInitClone()

	err := services.PullLatestChanges(repo)
	if err != nil {
		log.Fatalf("Error pulling latest changes: %v", err)
	}

	localCommits, err := services.CreateLocalCommit(repo, commits)
	if err != nil {
		log.Fatalf("Error creating local commits: %v", err)
	}
	log.Printf("Imported %v commits.\n", localCommits)

	if localCommits > 0 {
		if err := services.PushLocalCommits(repo); err != nil {
			log.Fatalf("Error pushing local commits: %v", err)
		}
		log.Println("Successfully pushed commits to remote repository.")
	} else {
		log.Println("No new commits were created, skipping push operation.")
	}
}

// parseOptionalDate reads an environment variable and parses it as YYYY-MM-DD.
// Returns nil if the variable is unset or empty.
func parseOptionalDate(envVar string) *time.Time {
	val := os.Getenv(envVar)
	if val == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", val)
	if err != nil {
		log.Printf("Warning: %s=%q is not a valid date (expected YYYY-MM-DD), ignoring.", envVar, val)
		return nil
	}
	return &t
}
