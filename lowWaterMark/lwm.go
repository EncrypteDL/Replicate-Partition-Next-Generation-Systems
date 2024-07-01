package lowwatermark

type LogEntry struct {
	Index uint64
	Data  string
}

type LogManager struct {
	entries      []LogEntry
	lowWaterMark uint64
}

func NewLogManager() *LogManager {
	return &LogManager{
		entries:      []LogEntry{},
		lowWaterMark: 0,
	}
}

func (lm *LogManager) AddEntry(entry LogEntry) {
	lm.entries = append(lm.entries, entry)
}

func (lm *LogManager) SetLowWaterMark(index uint64) {
	lm.lowWaterMark = index
	lm.cleanup()
}

func (lm *LogManager) cleanup() {
	var newEntries []LogEntry
	for _, entry := range lm.entries {
		if entry.Index >= lm.lowWaterMark {
			newEntries = append(newEntries, entry)
		}
	}
	lm.entries = newEntries
}
