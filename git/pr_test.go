package git

import (
	"context"
	"testing"
)

func TestCheckOpenPR_SkipsMainBranch(t *testing.T) {
	for _, branch := range []string{"main", "master", "develop", "HEAD", ""} {
		result := CheckOpenPR(context.Background(), ".", branch)
		if result {
			t.Errorf("expected false for branch %q", branch)
		}
	}
}

func TestCheckOpenPR_GracefulOnMissingGh(t *testing.T) {
	// This test verifies CheckOpenPR doesn't panic when gh is not available
	// or the directory is not a valid repo. It should return false gracefully.
	result := CheckOpenPR(context.Background(), t.TempDir(), "feature/test-branch")
	if result {
		t.Error("expected false for non-repo directory")
	}
}
