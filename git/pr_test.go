package git

import (
	"context"
	"errors"
	"testing"
)

type mockGhRunner struct {
	output string
	err    error
}

func (m *mockGhRunner) Run(_ context.Context, _ string, _ ...string) (string, error) {
	return m.output, m.err
}

func TestCheckOpenPR_SkipsMainBranch(t *testing.T) {
	gh := &mockGhRunner{output: `[{"number":1}]`}
	for _, branch := range []string{"main", "master", "develop", "HEAD", ""} {
		result := CheckOpenPR(context.Background(), gh, ".", branch)
		if result {
			t.Errorf("expected false for branch %q", branch)
		}
	}
}

func TestCheckOpenPR_WithMock_HasPR(t *testing.T) {
	gh := &mockGhRunner{output: `[{"number":42}]`}
	result := CheckOpenPR(context.Background(), gh, "/tmp", "feature/test")
	if !result {
		t.Error("expected true for open PR")
	}
}

func TestCheckOpenPR_WithMock_NoPR(t *testing.T) {
	gh := &mockGhRunner{output: `[]`}
	result := CheckOpenPR(context.Background(), gh, "/tmp", "feature/test")
	if result {
		t.Error("expected false for no open PR")
	}
}

func TestCheckOpenPR_WithMock_GhError(t *testing.T) {
	gh := &mockGhRunner{err: errors.New("gh not found")}
	result := CheckOpenPR(context.Background(), gh, "/tmp", "feature/test")
	if result {
		t.Error("expected false when gh fails")
	}
}

func TestCheckOpenPR_WithMock_InvalidJSON(t *testing.T) {
	gh := &mockGhRunner{output: "not json"}
	result := CheckOpenPR(context.Background(), gh, "/tmp", "feature/test")
	if result {
		t.Error("expected false for invalid JSON")
	}
}
