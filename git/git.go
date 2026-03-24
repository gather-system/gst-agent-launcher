package git

import (
	"context"
	"os/exec"
	"strings"
)

// Runner executes git commands in a given directory.
type Runner interface {
	Run(ctx context.Context, dir string, args ...string) (string, error)
}

type defaultRunner struct{}

// NewRunner returns a Runner that executes real git commands.
func NewRunner() Runner {
	return &defaultRunner{}
}

func (r *defaultRunner) Run(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
