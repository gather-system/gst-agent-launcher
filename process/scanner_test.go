package process

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestParseTasklistCSV(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "tasklist_sample.csv"))
	if err != nil {
		t.Fatal(err)
	}

	procs, err := ParseTasklistCSV(string(data))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Should have 3 results: WT PM, pwsh GST-develop, WT ProcessVision-App
	// (pwsh N/A and notepad filtered out)
	if len(procs) != 3 {
		t.Fatalf("expected 3 processes, got %d: %+v", len(procs), procs)
	}

	expected := []struct {
		pid   int
		title string
	}{
		{12345, "PM"},
		{23456, "GST-develop"},
		{45678, "ProcessVision-App"},
	}

	for i, e := range expected {
		if procs[i].PID != e.pid {
			t.Errorf("proc[%d] PID: want %d, got %d", i, e.pid, procs[i].PID)
		}
		if procs[i].WindowTitle != e.title {
			t.Errorf("proc[%d] title: want %q, got %q", i, e.title, procs[i].WindowTitle)
		}
	}
}

func TestParseTasklistCSV_Empty(t *testing.T) {
	procs, err := ParseTasklistCSV("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(procs) != 0 {
		t.Errorf("expected 0 processes, got %d", len(procs))
	}
}

func TestParseTasklistCSV_HeaderOnly(t *testing.T) {
	data := `"Image Name","PID","Session Name","Session#","Mem Usage","Status","User Name","CPU Time","Window Title"` + "\n"
	procs, err := ParseTasklistCSV(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(procs) != 0 {
		t.Errorf("expected 0 processes, got %d", len(procs))
	}
}

func TestMatchAgentNames(t *testing.T) {
	procs := []RunningProcess{
		{PID: 100, WindowTitle: "PM"},
		{PID: 200, WindowTitle: "GST-develop"},
		{PID: 300, WindowTitle: "ProcessVision-App"},
	}

	agents := []string{"PM", "Reviewer", "GST-develop", "ProcessVision-App", "SMB-App"}

	running := MatchAgentNames(procs, agents)

	if !running[0] {
		t.Error("expected PM (index 0) to be running")
	}
	if running[1] {
		t.Error("expected Reviewer (index 1) to NOT be running")
	}
	if !running[2] {
		t.Error("expected GST-develop (index 2) to be running")
	}
	if !running[3] {
		t.Error("expected ProcessVision-App (index 3) to be running")
	}
	if running[4] {
		t.Error("expected SMB-App (index 4) to NOT be running")
	}
}

func TestNewScanner(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}
	s := NewScanner()
	if s == nil {
		t.Fatal("NewScanner returned nil")
	}
}
