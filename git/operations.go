package git

import (
	"context"
	"sync"
	"time"

	"github.com/gather-system/gst-agent-launcher/config"
)

// BatchResult holds the result of a batch git operation on a single repo.
type BatchResult struct {
	AgentName  string
	AgentIndex int
	Success    bool
	Output     string
	Error      error
}

// PullAll runs git pull on multiple repos in parallel (max 5 concurrent).
func PullAll(ctx context.Context, runner Runner, agents []config.Agent, pathValid map[int]bool) []BatchResult {
	return batchRun(ctx, runner, agents, pathValid, "pull")
}

// StatusAll runs git status --short on multiple repos in parallel (max 5 concurrent).
func StatusAll(ctx context.Context, runner Runner, agents []config.Agent, pathValid map[int]bool) []BatchResult {
	return batchRun(ctx, runner, agents, pathValid, "status", "--short")
}

func batchRun(ctx context.Context, runner Runner, agents []config.Agent, pathValid map[int]bool, args ...string) []BatchResult {
	type job struct {
		index int
		agent config.Agent
	}

	var jobs []job
	for i, a := range agents {
		if pathValid[i] {
			jobs = append(jobs, job{index: i, agent: a})
		}
	}

	results := make([]BatchResult, len(jobs))
	sem := make(chan struct{}, 5) // concurrency limit
	var wg sync.WaitGroup

	for i, j := range jobs {
		wg.Add(1)
		go func(ri int, jb job) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			out, err := runner.Run(opCtx, jb.agent.Path, args...)
			results[ri] = BatchResult{
				AgentName:  jb.agent.Name,
				AgentIndex: jb.index,
				Success:    err == nil,
				Output:     out,
				Error:      err,
			}
		}(i, j)
	}

	wg.Wait()
	return results
}
