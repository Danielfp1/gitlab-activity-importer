[![Go Report Card](https://goreportcard.com/badge/github.com/furmanp/gitlab-activity-importer)](https://goreportcard.com/report/github.com/furmanp/gitlab-activity-importer)
![Latest Release](https://img.shields.io/github/v/release/furmanp/gitlab-activity-importer)

# Git activity Importer (Gitlab -> Github)
A tool to transfer your GitLab commit history to GitHub, reflecting your GitLab activity on GitHub’s contribution graph.
# Table of Contents
- [Git activity Importer (Gitlab -\> Github)](#git-activity-importer-gitlab---github)
- [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Features](#features)
  - [Setup](#setup)
    - [1. Environmental Variables](#1-environmental-variables)
    - [2. Automatic Imports (Recommended)](#2-automatic-imports-recommended)
    - [3. Manual Imports using repository](#3-manual-imports-using-repository)
    - [4. Manual Imports using binary](#4-manual-imports-using-binary)
  - [Local Import (Offline / Repository Backup)](#local-import-offline--repository-backup)
    - [When to use local mode](#when-to-use-local-mode)
    - [Required variables for local mode](#required-variables-for-local-mode)
    - [Optional date range variables](#optional-date-range-variables)
    - [Example .env for local mode](#example-env-for-local-mode)
    - [Running local mode on Linux / macOS](#running-local-mode-on-linux--macos)
    - [Running local mode on Windows (PowerShell)](#running-local-mode-on-windows-powershell)
    - [Using the --local-repo flag](#using-the---local-repo-flag)
    - [Fallback behaviour](#fallback-behaviour)
  - [Configuration](#configuration)
    - [Important Notes:](#important-notes)
  - [License](#license)


## Overview
This tool fetches your commit history from private GitLab repositories and imports it into a specified GitHub repository, creating a visual representation of your activity on GitHub’s contribution graph. It can be configured for automated daily imports or manual runs.

## Features
-	Automated Daily Imports: Syncs your GitLab activity with GitHub automatically each day.
-	Manual Imports: Allows on-demand updates.
-	Secure Data Handling: Requires minimal permissions and uses GitHub repository secrets for configuration.

## Setup
### 1. Environmental Variables

        | Secret Name       | Description                                                            |
        | ----------------- | ---------------------------------------------------------------------- |
        | `BASE_URL`        | URL of your GitLab instance (e.g., `https://gitlab.com`)               |
        | `GITLAB_USERNAME` | Your GitLab username                                                   |
        | `GH_USERNAME`     | Your GitHub username                                                   | 
        | `COMMITER_EMAIL`  | Email associated with your GitHub profile                              |
        | `GITLAB_TOKEN`    | GitLab personal access token (read permissions only)                   |
        | `ORIGIN_TOKEN`    | GitHub personal access token (with write permissions for auto-push)    |
        | `ORIGIN_REPO_URL` | HTTPS URL of your GitHub repository (ensure it has a `.git` extension) |

### 2. Automatic Imports (Recommended)
This approach will automatically keep your activity up to date. The program is being run daily at midnight UTC.
It imports your latest commits and automatically pushes them to specified GitHub repository.

To do that follow these steps:
1. **Fork this repository** to your GitHub account.
2. **Create an empty repository** in your GitHub profile where the commits will be pushed.
3. **Configure repository secrets** in your forked repository:
   - Go to your forked repository settings.
   - Under **Security**, navigate to **Secrets and variables > Actions**.
     ![Repository Secrets Configuration](assets/image.png)
   - Add the secrets from section [1](#1-environmental-variables):



Once these variables are saved in your Repository secrets, your commits will be automatically updated every day.

### 3. Manual Imports using repository
>You need to have GO installed on your computer

If you prefer to run the importer manually:
1. Clone the repository
2. Create an `.env` file in the root of your project and provide necessary variables
3. Run the tool locally whenever you want to sync your activity using `go run ./cmd/go/main.go

### 4. Manual Imports using binary
1. **Download the latest release** of the tool.
2. Set up the same environment variables on your local machine:
```
export BASE_URL=https://gitlab.com
export GITLAB_USERNAME=your_gitlab_username
export GH_USERNAME=your_github_username
export COMMITER_EMAIL=your_email@example.com
...
```
3. Run the tool binary whenever you want to sync your activity.


## Local Import (Offline / Repository Backup)

The tool supports a **local mode** that reads commit history directly from a `.git` directory on disk, without contacting the GitLab API or requiring any network access to the original remote repository.

### When to use local mode

- The original GitLab instance is no longer available or has been migrated.
- You have a local backup of a repository (the `.git` folder is enough — no working tree required).
- You want to import history from any Git repository regardless of its remote origin.
- The remote is misconfigured, invalid, or simply absent.

### Required variables for local mode

| Variable          | Description                                                            |
| ----------------- | ---------------------------------------------------------------------- |
| `LOCAL_REPO_PATH` | Absolute path to the folder that contains the `.git` directory         |
| `COMMITER_EMAIL`  | Email linked to your **GitHub** profile (used to sign destination commits) |
| `GH_USERNAME`     | Your GitHub username (used when writing commits to the destination)    |
| `ORIGIN_REPO_URL` | HTTPS URL of the destination GitHub repository                         |
| `ORIGIN_TOKEN`    | GitHub personal access token with write access to the destination repo |

> The GitLab variables (`BASE_URL`, `GITLAB_TOKEN`, `GITLAB_USERNAME`) are **not required** in local mode.

> **Using two different emails?** If the email you used to commit in the source repository (e.g. your company GitLab) is different from the email on your GitHub profile, set `SOURCE_AUTHOR_EMAIL` to the source email. The tool will filter source commits by `SOURCE_AUTHOR_EMAIL` and sign destination commits with `COMMITER_EMAIL`. When `SOURCE_AUTHOR_EMAIL` is not set, `COMMITER_EMAIL` is used for both.

### Optional variables for local mode

| Variable              | Format       | Description                                                                              |
| --------------------- | ------------ | ---------------------------------------------------------------------------------------- |
| `SOURCE_AUTHOR_EMAIL` | email        | Email used in the **source** repo. Required when it differs from `COMMITER_EMAIL`       |
| `SINCE_DATE`          | `YYYY-MM-DD` | Only import commits on or after this date                                                |
| `UNTIL_DATE`          | `YYYY-MM-DD` | Only import commits on or before this date                                               |

### Example .env for local mode

```env
LOCAL_REPO_PATH=/path/to/your/repo-backup

# Email used to author commits in the SOURCE repository (e.g. company GitLab):
SOURCE_AUTHOR_EMAIL=you@company.com

# Email linked to your GitHub profile (used to sign destination commits):
COMMITER_EMAIL=you@personal.com

GH_USERNAME=your_github_username
ORIGIN_REPO_URL=https://github.com/your_github_username/activity-mirror.git
ORIGIN_TOKEN=ghp_yourPersonalAccessToken

# Optional: restrict the date range
SINCE_DATE=2023-01-01
UNTIL_DATE=2024-12-31
```

### Running local mode on Linux / macOS

```bash
# Using the .env file (place it in the project root):
go run ./cmd/main.go

# Or pass the variable inline without a .env file:
LOCAL_REPO_PATH=/home/user/backups/my-project \
  COMMITER_EMAIL=you@example.com \
  GH_USERNAME=your_github_username \
  ORIGIN_REPO_URL=https://github.com/your_github_username/activity-mirror.git \
  ORIGIN_TOKEN=ghp_yourToken \
  go run ./cmd/main.go

# Using a pre-built binary:
LOCAL_REPO_PATH=/home/user/backups/my-project \
  COMMITER_EMAIL=you@example.com \
  GH_USERNAME=your_github_username \
  ORIGIN_REPO_URL=https://github.com/your_github_username/activity-mirror.git \
  ORIGIN_TOKEN=ghp_yourToken \
  ./GitLab-Importer-linux-amd64
```

### Running local mode on Windows (PowerShell)

```powershell
# Set variables in the current session, then run:
$env:LOCAL_REPO_PATH = "C:\Backups\my-project"
$env:COMMITER_EMAIL  = "you@example.com"
$env:GH_USERNAME     = "your_github_username"
$env:ORIGIN_REPO_URL = "https://github.com/your_github_username/activity-mirror.git"
$env:ORIGIN_TOKEN    = "ghp_yourPersonalAccessToken"

go run .\cmd\main.go

# Using a pre-built binary:
.\GitLab-Importer-windows-amd64.exe
```

### Using the --local-repo flag

You can also activate local mode with the `--local-repo` flag. When this flag is present the tool will read `LOCAL_REPO_PATH` from the environment (or the `.env` file):

```bash
# Linux / macOS
go run ./cmd/main.go --local-repo

# Windows PowerShell
go run .\cmd\main.go --local-repo
```

### Fallback behaviour

If the GitLab API is unreachable (network error, instance offline, invalid token) **and** `LOCAL_REPO_PATH` is set, the tool automatically falls back to local mode instead of exiting with an error. A warning is printed to the log:

```
GitLab API unavailable (<error>). LOCAL_REPO_PATH is set — falling back to local repository mode.
```

## Configuration
This project uses GitHub Actions to automate builds and daily synchronization:

- GitHub Actions Workflow: The .github/workflows/schedule.yml defines the automation steps for building and running the tool.
- Secrets Configuration: The secrets allow secure storage and retrieval of required tokens and URLs during automation.

### Important Notes:
- **GitLab permissions:** The tool requires read-only access to your GitLab user and Gitlab repositories (`read_user` and `read_repository`)
- **GitHub permissions:** Your GitHub token must have write access to the destination repository for automatic pushes.

## License
This project is licensed under the MIT License, which allows for free, unrestricted use, copying, modification, and distribution with attribution.
