package process

import (
	"context"
	"encoding/csv"
	"os/exec"
	"strconv"
	"strings"
)

type windowsScanner struct{}

// NewScanner returns a Scanner that uses tasklist on Windows.
func NewScanner() Scanner {
	return &windowsScanner{}
}

func (s *windowsScanner) ScanRunning(ctx context.Context) ([]RunningProcess, error) {
	cmd := exec.CommandContext(ctx, "tasklist", "/V", "/FO", "CSV")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParseTasklistCSV(string(out))
}

// ParseTasklistCSV parses tasklist /V /FO CSV output and returns processes
// that are WindowsTerminal.exe or pwsh.exe with non-empty window titles.
func ParseTasklistCSV(data string) ([]RunningProcess, error) {
	reader := csv.NewReader(strings.NewReader(data))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, nil
	}

	// Find column indices from header.
	header := records[0]
	nameCol := -1
	pidCol := -1
	titleCol := -1
	for i, h := range header {
		lower := strings.ToLower(strings.TrimSpace(h))
		switch {
		case lower == "image name" || lower == "映像名稱":
			nameCol = i
		case lower == "pid" || lower == "pid ":
			pidCol = i
		case lower == "window title" || lower == "視窗標題":
			titleCol = i
		}
	}

	if nameCol < 0 || pidCol < 0 || titleCol < 0 {
		return nil, nil
	}

	var procs []RunningProcess
	for _, row := range records[1:] {
		if len(row) <= titleCol || len(row) <= nameCol {
			continue
		}

		imgName := strings.ToLower(row[nameCol])
		if imgName != "windowsterminal.exe" && imgName != "pwsh.exe" {
			continue
		}

		title := strings.TrimSpace(row[titleCol])
		if title == "" || title == "N/A" || title == "不適用" {
			continue
		}

		pid, _ := strconv.Atoi(strings.TrimSpace(row[pidCol]))
		procs = append(procs, RunningProcess{
			PID:         pid,
			WindowTitle: title,
		})
	}

	return procs, nil
}
