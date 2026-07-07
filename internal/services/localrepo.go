package services

import (
	"fmt"
	"log"
	"time"

	"github.com/furmanp/gitlab-activity-importer/internal"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// ReadLocalCommits opens a local Git repository at repoPath and returns all
// commits authored by authorEmail. It works with bare-style folders, repos
// without a remote, and repos whose remote no longer exists, because it only
// reads object data from the .git directory. since and until are optional date
// range filters; pass nil to disable each bound.
//
// authorEmail should be the email used to author commits in the SOURCE
// repository (e.g. the company GitLab email). It may differ from the
// COMMITER_EMAIL used to sign commits in the destination GitHub repository.
func ReadLocalCommits(repoPath, authorEmail string, since, until *time.Time) ([]internal.Commit, error) {
	log.Println("Reading commits from local Git history")

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open local repository at %q: %w", repoPath, err)
	}

	logOpts := &git.LogOptions{
		All: true,
	}
	if since != nil {
		logOpts.Since = since
	}
	if until != nil {
		logOpts.Until = until
	}

	iter, err := repo.Log(logOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to read commit log: %w", err)
	}
	defer iter.Close()

	var commits []internal.Commit
	err = iter.ForEach(func(c *object.Commit) error {
		if c.Author.Email != authorEmail {
			return nil
		}
		commits = append(commits, internal.Commit{
			ID:           c.Hash.String(),
			Message:      c.Message,
			AuthorName:   c.Author.Name,
			AuthorMail:   c.Author.Email,
			AuthoredDate: c.Author.When,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error iterating commits: %w", err)
	}

	log.Printf("Found %d commits by %q in local repository", len(commits), authorEmail)
	return commits, nil
}
