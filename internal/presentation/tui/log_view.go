package tui

import (
	"fmt"
	"strings"
	"time"
)

// LogScreen displays execution logs
type LogScreen struct {
	entries []logEntry
	maxSize int
}

// logEntry represents a single log entry
type logEntry struct {
	message   string
	timestamp time.Time
	level     string // info, warn, error, success
}

// NewLogScreen creates a new log screen
func NewLogScreen() *LogScreen {
	return &LogScreen{
		entries: []logEntry{},
		maxSize: 100,
	}
}

// AddLog adds a log entry
func (ls *LogScreen) AddLog(msg string) {
	ls.AddLogLevel(msg, "info")
}

// AddLogLevel adds a log entry with a specific level
func (ls *LogScreen) AddLogLevel(msg string, level string) {
	entry := logEntry{
		message:   msg,
		timestamp: time.Now(),
		level:     level,
	}

	ls.entries = append(ls.entries, entry)

	// Keep only last maxSize entries
	if len(ls.entries) > ls.maxSize {
		ls.entries = ls.entries[len(ls.entries)-ls.maxSize:]
	}
}

// View renders the log screen with responsive width
func (ls *LogScreen) View(width int) string {
	var sb strings.Builder

	sb.WriteString(TitleStyle.Render("  Execution Log") + "\n\n")

	if len(ls.entries) == 0 {
		sb.WriteString(MutedStyle.Render("  No logs yet"))
		return sb.String()
	}

	// Calculate visible entries based on available width
	maxLineLen := max(20, width-6)

	// Show the last 20 entries
	startIdx := 0
	if len(ls.entries) > 20 {
		startIdx = len(ls.entries) - 20
	}

	for _, entry := range ls.entries[startIdx:] {
		timeStr := entry.timestamp.Format("15:04:05")

		var logLine string
		switch entry.level {
		case "error":
			logLine = ErrorStyle.Render(timeStr + " [" + entry.level + "] " + entry.message)
		case "warn":
			logLine = WarnStyle.Render(timeStr + " [" + entry.level + "] " + entry.message)
		case "success":
			logLine = SuccessStyle.Render(timeStr + " [" + entry.level + "] " + entry.message)
		default:
			logLine = MutedStyle.Render(timeStr + " [" + entry.level + "] " + entry.message)
		}

		// Truncate if too long
		if len(logLine) > maxLineLen {
			logLine = logLine[:maxLineLen-3] + "..."
		}
		sb.WriteString("  " + logLine + "\n")
	}

	sb.WriteString("\n" + SectionHeaderStyle.Width(width).Render(" SUMMARY"))
	sb.WriteString("\n" + MutedStyle.Render(fmt.Sprintf("  Total entries: %d", len(ls.entries))))

	return sb.String()
}

// Clear clears all logs
func (ls *LogScreen) Clear() {
	ls.entries = []logEntry{}
}

// GetEntries returns all log entries
func (ls *LogScreen) GetEntries() []string {
	var result []string
	for _, entry := range ls.entries {
		result = append(result, entry.message)
	}
	return result
}
