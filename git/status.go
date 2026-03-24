package git

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gather-system/gst-agent-launcher/config"
)

// RepoStatus holds the git status of a single repository.
type RepoStatus struct {
	AgentIndex int
	Branch     string
	DirtyCount int
	IssueID    string
	HasOpenPR  bool
	Error      error
}

// GetStatus returns the git status for a single repository.
func GetStatus(ctx context.Context, runner Runner, ghRunner GhRunner, index int, path string) RepoStatus {
	result := RepoStatus{AgentIndex: index}

	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Get current branch.
	branch, err := runner.Run(checkCtx, path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		result.Error = err
		return result
	}
	result.Branch = branch
	result.IssueID = ExtractIssueID(branch)

	// Start PR check concurrently (uses its own 5s timeout).
	var prWg sync.WaitGroup
	prWg.Add(1)
	go func() {
		defer prWg.Done()
		result.HasOpenPR = CheckOpenPR(ctx, ghRunner, path, branch)
	}()

	// Get dirty count (concurrent with PR check).
	porcelain, err := runner.Run(checkCtx, path, "status", "--porcelain")
	if err != nil {
		result.Error = err
		prWg.Wait()
		return result
	}
	if porcelain != "" {
		result.DirtyCount = len(strings.Split(porcelain, "\n"))
	}

	prWg.Wait()
	return result
}

// GetAllStatuses returns git status for all agents with valid git repos, in parallel.
func GetAllStatuses(ctx context.Context, runner Runner, ghRunner GhRunner, agents []config.Agent, isGitRepo func(int) bool) []RepoStatus {
	var results []RepoStatus
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i, agent := range agents {
		if !isGitRepo(i) {
			continue
		}
		wg.Add(1)
		go func(idx int, path string) {
			defer wg.Done()
			status := GetStatus(ctx, runner, ghRunner, idx, path)
			mu.Lock()
			results = append(results, status)
			mu.Unlock()
		}(i, agent.Path)
	}

	wg.Wait()
	return results
}
