package health

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gather-system/gst-agent-launcher/config"
)

func TestCheckAll_PathNotExist(t *testing.T) {
	agents := []config.Agent{
		{Name: "missing", Path: filepath.Join(t.TempDir(), "no-such-dir"), Group: "Test"},
	}
	checker := NewChecker()
	results, _ := checker.CheckAll(context.Background(), agents)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.PathValid {
		t.Error("expected PathValid=false for non-existent path")
	}
	if r.IsGitRepo {
		t.Error("expected IsGitRepo=false for non-existent path")
	}
}

func TestCheckAll_PathExistsNoGit(t *testing.T) {
	dir := t.TempDir()
	agents := []config.Agent{
		{Name: "no-git", Path: dir, Group: "Test"},
	}
	checker := NewChecker()
	results, _ := checker.CheckAll(context.Background(), agents)

	r := results[0]
	if !r.PathValid {
		t.Error("expected PathValid=true")
	}
	if r.IsGitRepo {
		t.Error("expected IsGitRepo=false for directory without .git")
	}
}

func TestCheckAll_PathExistsWithGit(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	agents := []config.Agent{
		{Name: "git-repo", Path: dir, Group: "Test"},
	}
	checker := NewChecker()
	results, _ := checker.CheckAll(context.Background(), agents)

	r := results[0]
	if !r.PathValid {
		t.Error("expected PathValid=true")
	}
	if !r.IsGitRepo {
		t.Error("expected IsGitRepo=true for directory with .git")
	}
	if r.HasConflict {
		t.Error("expected HasConflict=false for temp dir")
	}
}

func TestCheckAll_MultipleAgents(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir2, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	agents := []config.Agent{
		{Name: "no-git", Path: dir1, Group: "Test"},
		{Name: "git-repo", Path: dir2, Group: "Test"},
		{Name: "missing", Path: filepath.Join(t.TempDir(), "nope"), Group: "Test"},
	}
	checker := NewChecker()
	results, _ := checker.CheckAll(context.Background(), agents)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Agent 0: exists, no git.
	if !results[0].PathValid || results[0].IsGitRepo {
		t.Errorf("agent 0: want PathValid=true IsGitRepo=false, got %v %v", results[0].PathValid, results[0].IsGitRepo)
	}
	// Agent 1: exists, git repo.
	if !results[1].PathValid || !results[1].IsGitRepo {
		t.Errorf("agent 1: want PathValid=true IsGitRepo=true, got %v %v", results[1].PathValid, results[1].IsGitRepo)
	}
	// Agent 2: does not exist.
	if results[2].PathValid {
		t.Errorf("agent 2: want PathValid=false, got true")
	}
}

func TestCheckAll_GitAvailable(t *testing.T) {
	agents := []config.Agent{}
	checker := NewChecker()
	_, gitAvailable := checker.CheckAll(context.Background(), agents)

	// On CI/dev machines git should be available; just verify the function runs.
	if !gitAvailable {
		t.Log("git not found on this machine — skipping availability assertion")
	}
}

func TestCheckAll_ContextCancelled(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	agents := []config.Agent{
		{Name: "cancelled", Path: dir, Group: "Test"},
	}
	checker := NewChecker()
	results, _ := checker.CheckAll(ctx, agents)

	r := results[0]
	if !r.PathValid {
		t.Error("PathValid should still be true (no I/O timeout for os.Stat)")
	}
}
