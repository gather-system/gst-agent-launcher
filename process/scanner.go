package process

import "context"

// RunningProcess represents a detected running process.
type RunningProcess struct {
	PID         int
	WindowTitle string
}

// Scanner detects running processes that may be agent terminals.
type Scanner interface {
	ScanRunning(ctx context.Context) ([]RunningProcess, error)
}

// MatchAgentNames returns a set of agent indices whose names appear in any window title.
func MatchAgentNames(processes []RunningProcess, agentNames []string) map[int]bool {
	running := make(map[int]bool)
	for _, proc := range processes {
		for i, name := range agentNames {
			if containsIgnoreCase(proc.WindowTitle, name) {
				running[i] = true
			}
		}
	}
	return running
}

func containsIgnoreCase(s, substr string) bool {
	sLower := toLower(s)
	subLower := toLower(substr)
	return len(subLower) > 0 && contains(sLower, subLower)
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
