package lowwatermark

import (
	"testing"
)

func TestAddEntry(t *testing.T) {
	logManager := NewLogManager()
	logManager.AddEntry(LogEntry{Index: 1, Data: "Entry 1"})
	logManager.AddEntry(LogEntry{Index: 2, Data: "Entry 2"})

	if len(logManager.entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(logManager.entries))
	}
}

func TestSetLowWaterMark(t *testing.T) {
	logManager := NewLogManager()
	logManager.AddEntry(LogEntry{Index: 1, Data: "Entry 1"})
	logManager.AddEntry(LogEntry{Index: 2, Data: "Entry 2"})
	logManager.AddEntry(LogEntry{Index: 3, Data: "Entry 3"})

	logManager.SetLowWaterMark(2)

	if len(logManager.entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(logManager.entries))
	}

	if logManager.entries[0].Index != 2 || logManager.entries[1].Index != 3 {
		t.Errorf("Entries do not match expected indices")
	}
}

func TestCleanup(t *testing.T) {
	logManager := NewLogManager()
	logManager.AddEntry(LogEntry{Index: 1, Data: "Entry 1"})
	logManager.AddEntry(LogEntry{Index: 2, Data: "Entry 2"})
	logManager.AddEntry(LogEntry{Index: 3, Data: "Entry 3"})
	logManager.AddEntry(LogEntry{Index: 4, Data: "Entry 4"})

	logManager.SetLowWaterMark(3)

	if len(logManager.entries) != 2 {
		t.Errorf("Expected 2 entries after cleanup, got %d", len(logManager.entries))
	}

	if logManager.entries[0].Index != 3 || logManager.entries[1].Index != 4 {
		t.Errorf("Entries do not match expected indices after cleanup")
	}
}
