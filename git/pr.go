package git

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"time"
)

// GhRunner executes gh CLI commands in a given directory.
type GhRunner interface {
	Run(ctx context.Context, dir string, args ...string) (string, error)
}

type defaultGhRunner struct{}

// NewGhRunner returns a GhRunner that executes real gh CLI commands.
func NewGhRunner() GhRunner {
	return &defaultGhRunner{}
}

func (r *defaultGhRunner) Run(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// CheckOpenPR checks if the current branch has an open PR using gh CLI.
// Returns false gracefully if gh is not installed or not authenticated.
func CheckOpenPR(ctx context.Context, ghRunner GhRunner, dir string, branch string) bool {
	if branch == "" || branch == "main" || branch == "master" || branch == "develop" || branch == "HEAD" {
		return false
	}

	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	out, err := ghRunner.Run(checkCtx, dir, "pr", "list", "--head", branch, "--json", "number", "--limit", "1")
	if err != nil {
		return false
	}

	var prs []struct {
		Number int `json:"number"`
	}
	if err := json.Unmarshal([]byte(out), &prs); err != nil {
		return false
	}
	return len(prs) > 0
}
