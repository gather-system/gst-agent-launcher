package git

import (
	"context"
	"errors"
	"testing"

	"github.com/gather-system/gst-agent-launcher/config"
)

// mockRunner is a test double for Runner.
type mockRunner struct {
	responses map[string]mockResponse
}

type mockResponse struct {
	output string
	err    error
}

func (m *mockRunner) Run(_ context.Context, dir string, args ...string) (string, error) {
	key := dir + ":" + args[0]
	if len(args) > 1 {
		key += ":" + args[1]
	}
	resp, ok := m.responses[key]
	if !ok {
		return "", errors.New("unexpected command: " + key)
	}
	return resp.output, resp.err
}

func TestGetStatus_CleanRepo(t *testing.T) {
	runner := &mockRunner{responses: map[string]mockResponse{
		"/repo:rev-parse:--abbrev-ref": {output: "main"},
		"/repo:status:--porcelain":     {output: ""},
	}}

	s := GetStatus(context.Background(), runner, &mockGhRunner{}, 0, "/repo")
	if s.Error != nil {
		t.Fatalf("unexpected error: %v", s.Error)
	}
	if s.Branch != "main" {
		t.Errorf("expected branch main, got %s", s.Branch)
	}
	if s.DirtyCount != 0 {
		t.Errorf("expected dirty 0, got %d", s.DirtyCount)
	}
	if s.IssueID != "" {
		t.Errorf("expected empty issue ID, got %s", s.IssueID)
	}
}

func TestGetStatus_DirtyRepo(t *testing.T) {
	runner := &mockRunner{responses: map[string]mockResponse{
		"/repo:rev-parse:--abbrev-ref": {output: "feature/GST-250-dashboard"},
		"/repo:status:--porcelain":     {output: " M file1.go\n M file2.go\n?? file3.go"},
	}}

	s := GetStatus(context.Background(), runner, &mockGhRunner{}, 1, "/repo")
	if s.Error != nil {
		t.Fatalf("unexpected error: %v", s.Error)
	}
	if s.Branch != "feature/GST-250-dashboard" {
		t.Errorf("expected branch feature/GST-250-dashboard, got %s", s.Branch)
	}
	if s.DirtyCount != 3 {
		t.Errorf("expected dirty 3, got %d", s.DirtyCount)
	}
	if s.IssueID != "GST-250" {
		t.Errorf("expected issue ID GST-250, got %s", s.IssueID)
	}
}

func TestGetStatus_DetachedHead(t *testing.T) {
	runner := &mockRunner{responses: map[string]mockResponse{
		"/repo:rev-parse:--abbrev-ref": {output: "HEAD"},
		"/repo:status:--porcelain":     {output: ""},
	}}

	s := GetStatus(context.Background(), runner, &mockGhRunner{}, 0, "/repo")
	if s.Error != nil {
		t.Fatalf("unexpected error: %v", s.Error)
	}
	if s.Branch != "HEAD" {
		t.Errorf("expected branch HEAD, got %s", s.Branch)
	}
}

func TestGetStatus_BranchError(t *testing.T) {
	runner := &mockRunner{responses: map[string]mockResponse{
		"/repo:rev-parse:--abbrev-ref": {err: errors.New("not a git repo")},
	}}

	s := GetStatus(context.Background(), runner, &mockGhRunner{}, 0, "/repo")
	if s.Error == nil {
		t.Fatal("expected error")
	}
	if s.Branch != "" {
		t.Errorf("expected empty branch, got %s", s.Branch)
	}
}

func TestGetAllStatuses(t *testing.T) {
	runner := &mockRunner{responses: map[string]mockResponse{
		"/a:rev-parse:--abbrev-ref": {output: "main"},
		"/a:status:--porcelain":     {output: ""},
		"/c:rev-parse:--abbrev-ref": {output: "develop"},
		"/c:status:--porcelain":     {output: " M x.go"},
	}}

	agents := []config.Agent{
		{Name: "A", Path: "/a"},
		{Name: "B", Path: "/b"},
		{Name: "C", Path: "/c"},
	}

	isGitRepo := func(i int) bool { return i != 1 } // skip B

	results := GetAllStatuses(context.Background(), runner, &mockGhRunner{}, agents, isGitRepo)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	byIndex := make(map[int]RepoStatus)
	for _, r := range results {
		byIndex[r.AgentIndex] = r
	}

	if byIndex[0].Branch != "main" {
		t.Errorf("agent 0: expected main, got %s", byIndex[0].Branch)
	}
	if byIndex[2].Branch != "develop" {
		t.Errorf("agent 2: expected develop, got %s", byIndex[2].Branch)
	}
	if byIndex[2].DirtyCount != 1 {
		t.Errorf("agent 2: expected dirty 1, got %d", byIndex[2].DirtyCount)
	}
}
