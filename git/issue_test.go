package git

import "testing"

func TestExtractIssueID(t *testing.T) {
	tests := []struct {
		branch string
		want   string
	}{
		{"feature/GST-250-dashboard", "GST-250"},
		{"GST-250-desc", "GST-250"},
		{"fix/JOP-170", "JOP-170"},
		{"hubert/gst-238-v040-health-check", ""},
		{"main", ""},
		{"develop", ""},
		{"release/v1.0", ""},
		{"LEY-42-inspection", "LEY-42"},
		{"feature/ABC-1", "ABC-1"},
		{"HEAD", ""},
	}

	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			got := ExtractIssueID(tt.branch)
			if got != tt.want {
				t.Errorf("ExtractIssueID(%q) = %q, want %q", tt.branch, got, tt.want)
			}
		})
	}
}
