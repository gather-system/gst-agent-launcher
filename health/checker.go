package health

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gather-system/gst-agent-launcher/config"
)

// Checker runs pre-flight health checks on agent directories.
type Checker interface {
	CheckAll(ctx context.Context, agents []config.Agent) ([]CheckResult, bool)
}

type defaultChecker struct{}

// NewChecker returns a Checker that performs real filesystem and git checks.
func NewChecker() Checker {
	return &defaultChecker{}
}

// CheckAll checks all agents in parallel and returns results plus git availability.
func (c *defaultChecker) CheckAll(ctx context.Context, agents []config.Agent) ([]CheckResult, bool) {
	gitAvailable := isGitAvailable()

	results := make([]CheckResult, len(agents))
	var wg sync.WaitGroup

	for i, agent := range agents {
		wg.Add(1)
		go func(idx int, a config.Agent) {
			defer wg.Done()
			checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			results[idx] = checkAgent(checkCtx, idx, a, gitAvailable)
		}(i, agent)
	}

	wg.Wait()
	return results, gitAvailable
}

func checkAgent(ctx context.Context, index int, agent config.Agent, gitAvailable bool) CheckResult {
	result := CheckResult{AgentIndex: index}

	// Path check.
	_, err := os.Stat(agent.Path)
	result.PathValid = err == nil
	if !result.PathValid {
		return result
	}

	// Git repo check.
	_, err = os.Stat(filepath.Join(agent.Path, ".git"))
	result.IsGitRepo = err == nil
	if !result.IsGitRepo || !gitAvailable {
		return result
	}

	// Conflict check (only if git repo and git available).
	select {
	case <-ctx.Done():
		result.Error = ctx.Err()
		return result
	default:
	}

	cmd := exec.CommandContext(ctx, "git", "diff", "--name-only", "--diff-filter=U")
	cmd.Dir = agent.Path
	out, err := cmd.Output()
	if err == nil && len(strings.TrimSpace(string(out))) > 0 {
		result.HasConflict = true
	}

	return result
}

func isGitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}
