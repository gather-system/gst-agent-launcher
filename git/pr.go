package git

import (
	"context"
	"encoding/json"
	"os/exec"
	"time"
)

// CheckOpenPR checks if the current branch has an open PR using gh CLI.
// Returns false gracefully if gh is not installed or not authenticated.
func CheckOpenPR(ctx context.Context, dir string, branch string) bool {
	if branch == "" || branch == "main" || branch == "master" || branch == "develop" || branch == "HEAD" {
		return false
	}

	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(checkCtx, "gh", "pr", "list", "--head", branch, "--json", "number", "--limit", "1")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return false // gh not available or auth issue — graceful
	}

	var prs []struct {
		Number int `json:"number"`
	}
	if err := json.Unmarshal(out, &prs); err != nil {
		return false
	}
	return len(prs) > 0
}
