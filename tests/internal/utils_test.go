package services_test

import (
	"os"
	"strings"
	"testing"

	"github.com/furmanp/gitlab-activity-importer/internal"
)

func clearEnvVars(t *testing.T) {
	vars := []string{
		"ENV",
		"BASE_URL",
		"GITLAB_TOKEN",
		"GITLAB_USERNAME",
		"GH_USERNAME",
		"COMMITER_EMAIL",
		"ORIGIN_REPO_URL",
		"ORIGIN_TOKEN",
		"COMMITS_IMPORTER_PATH",
		"LOCAL_REPO_PATH",
	}

	for _, v := range vars {
		if err := os.Unsetenv(v); err != nil {
			t.Fatalf("failed to unset %s: %v", v, err)
		}
	}
}

func TestCheckEnvVariables(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name: "all required variables set",
			setupEnv: map[string]string{
				"BASE_URL":        "http://test-url.com",
				"GITLAB_TOKEN":    "token123",
				"GITLAB_USERNAME": "gitlab_user",
				"GH_USERNAME":     "github_user",
				"COMMITER_EMAIL":  "test@example.com",
				"ORIGIN_REPO_URL": "http://repo.com",
				"ORIGIN_TOKEN":    "origintoken123",
			},
			expectError: false,
		},
		{
			name: "missing one variable",
			setupEnv: map[string]string{
				"BASE_URL":        "http://test-url.com",
				"GITLAB_TOKEN":    "token123",
				"GITLAB_USERNAME": "gitlab_user",
				"GH_USERNAME":     "github_user",
				"COMMITER_EMAIL":  "test@example.com",
				"ORIGIN_TOKEN":    "origintoken123",
			},
			expectError: true,
			errorMsg:    "ORIGIN_REPO_URL",
		},
		{
			name: "missing multiple variables",
			setupEnv: map[string]string{
				"BASE_URL": "http://test-url.com",
			},
			expectError: true,
			errorMsg:    "GITLAB_TOKEN, GITLAB_USERNAME, GH_USERNAME, COMMITER_EMAIL, ORIGIN_REPO_URL, ORIGIN_TOKEN",
		},
		{
			name:        "no variables set",
			setupEnv:    map[string]string{},
			expectError: true,
			errorMsg:    "BASE_URL, GITLAB_TOKEN, GITLAB_USERNAME, GH_USERNAME, COMMITER_EMAIL, ORIGIN_REPO_URL, ORIGIN_TOKEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars(t)

			for k, v := range tt.setupEnv {
				if err := os.Setenv(k, v); err != nil {
					t.Fatalf("failed to set %s: %v", k, err)
				}
			}

			err := internal.SetupEnv()

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectError && err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
			}
		})
	}
}

func TestGetCommitsImporterPath(t *testing.T) {
	clearEnvVars(t)

	t.Run("uses COMMITS_IMPORTER_PATH when set", func(t *testing.T) {
		custom := `D:\custom\clone`
		if err := os.Setenv("COMMITS_IMPORTER_PATH", custom); err != nil {
			t.Fatalf("failed to set COMMITS_IMPORTER_PATH: %v", err)
		}
		defer os.Unsetenv("COMMITS_IMPORTER_PATH")

		got := internal.GetCommitsImporterPath()
		if got != custom {
			t.Fatalf("expected %q, got %q", custom, got)
		}
	})

	t.Run("default path when env unset", func(t *testing.T) {
		_ = os.Unsetenv("COMMITS_IMPORTER_PATH")
		got := internal.GetCommitsImporterPath()
		if got == "" {
			t.Fatal("expected non-empty default path")
		}
		if !strings.Contains(strings.ToLower(got), "commits-importer") {
			t.Fatalf("expected path to contain commits-importer, got %q", got)
		}
	})
}
