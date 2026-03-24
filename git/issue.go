package git

import "regexp"

var issueIDPattern = regexp.MustCompile(`([A-Z]+-\d+)`)

// ExtractIssueID extracts the first issue ID (e.g. GST-250, JOP-170) from a branch name.
// Returns empty string if no match is found.
func ExtractIssueID(branch string) string {
	match := issueIDPattern.FindString(branch)
	return match
}
