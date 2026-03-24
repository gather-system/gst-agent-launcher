package git

import (
	"context"
	"errors"
	"testing"

	"github.com/gather-system/gst-agent-launcher/config"
)

type batchMockRunner struct {
	responses map[string]mockResponse
}

func (m *batchMockRunner) Run(_ context.Context, dir string, args ...string) (string, error) {
	key := dir
	resp, ok := m.responses[key]
	if !ok {
		return "", errors.New("unexpected dir: " + key)
	}
	return resp.output, resp.err
}

func TestPullAll_Success(t *testing.T) {
	runner := &batchMockRunner{responses: map[string]mockResponse{
		"/a": {output: "Already up to date."},
		"/b": {output: "Updating abc..def\n1 file changed"},
	}}

	agents := []config.Agent{
		{Name: "A", Path: "/a"},
		{Name: "B", Path: "/b"},
		{Name: "C", Path: "/c"},
	}
	pathValid := map[int]bool{0: true, 1: true, 2: false}

	results := PullAll(context.Background(), runner, agents, pathValid)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for _, r := range results {
		if !r.Success {
			t.Errorf("agent %s: expected success, got error: %v", r.AgentName, r.Error)
		}
	}
}

func TestPullAll_MixedResults(t *testing.T) {
	runner := &batchMockRunner{responses: map[string]mockResponse{
		"/a": {output: "Already up to date."},
		"/b": {err: errors.New("could not resolve host")},
	}}

	agents := []config.Agent{
		{Name: "A", Path: "/a"},
		{Name: "B", Path: "/b"},
	}
	pathValid := map[int]bool{0: true, 1: true}

	results := PullAll(context.Background(), runner, agents, pathValid)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	byName := make(map[string]BatchResult)
	for _, r := range results {
		byName[r.AgentName] = r
	}

	if !byName["A"].Success {
		t.Error("A should succeed")
	}
	if byName["B"].Success {
		t.Error("B should fail")
	}
}

func TestStatusAll(t *testing.T) {
	runner := &batchMockRunner{responses: map[string]mockResponse{
		"/a": {output: " M file.go"},
	}}

	agents := []config.Agent{
		{Name: "A", Path: "/a"},
	}
	pathValid := map[int]bool{0: true}

	results := StatusAll(context.Background(), runner, agents, pathValid)
	if len(results) != 1 || !results[0].Success {
		t.Fatalf("expected 1 success result, got %+v", results)
	}
	if results[0].Output != " M file.go" {
		t.Errorf("unexpected output: %q", results[0].Output)
	}
}
