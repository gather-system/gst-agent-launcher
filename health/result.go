package health

// CheckResult holds the health check outcome for a single agent.
type CheckResult struct {
	AgentIndex  int
	PathValid   bool
	IsGitRepo   bool
	HasConflict bool
	Error       error
}
