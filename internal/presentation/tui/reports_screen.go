package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/0xtbug/evm-disperse-tools/internal/domain/entity"
	"github.com/0xtbug/evm-disperse-tools/internal/domain/port"
)

// ReportsScreen displays persisted disperse execution reports from disk.
type ReportsScreen struct {
	repo         port.ReportRepository
	reports      []*entity.ExecutionReport
	scrollOffset int
	width        int
	loaded       bool
}

// NewReportsScreen creates a new reports screen backed by a report repository.
func NewReportsScreen(repo port.ReportRepository) *ReportsScreen {
	return &ReportsScreen{
		repo: repo,
	}
}

// LoadReports loads all reports from the repository into memory.
func (rs *ReportsScreen) LoadReports() {
	if rs.repo == nil {
		return
	}
	reports, err := rs.repo.ListAll()
	if err != nil {
		rs.reports = nil
		rs.loaded = true
		return
	}
	rs.reports = reports
	rs.loaded = true
	rs.scrollOffset = 0
}

// maxVisibleReports returns the max number of report cards shown at once.
func (rs *ReportsScreen) maxVisibleReports() int {
	return 1
}

// ScrollUp scrolls up to older reports.
func (rs *ReportsScreen) ScrollUp() {
	maxOffset := len(rs.reports) - rs.maxVisibleReports()
	if maxOffset < 0 {
		maxOffset = 0
	}
	if rs.scrollOffset < maxOffset {
		rs.scrollOffset++
	}
}

// ScrollDown scrolls down to newer reports.
func (rs *ReportsScreen) ScrollDown() {
	if rs.scrollOffset > 0 {
		rs.scrollOffset--
	}
}

// View renders the reports screen.
func (rs *ReportsScreen) View(width int) string {
	rs.width = width

	// Lazy-load on first view
	if !rs.loaded {
		rs.LoadReports()
	}

	var sb strings.Builder

	sb.WriteString(TitleStyle.Render("  Execution Reports") + "\n\n")

	if len(rs.reports) == 0 {
		sb.WriteString(MutedStyle.Render("  No reports yet — run a disperse operation to generate reports"))
		return sb.String()
	}

	total := len(rs.reports)
	maxVisible := rs.maxVisibleReports()

	// Calculate visible range
	startIdx := rs.scrollOffset
	endIdx := rs.scrollOffset + maxVisible
	if endIdx > total {
		endIdx = total
	}

	// Render each report as a card
	for i, report := range rs.reports[startIdx:endIdx] {
		cardNum := startIdx + i + 1
		sb.WriteString(rs.renderCard(report, cardNum, width))
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderCard renders a single report as a bordered card with full tx hash.
// Box interior width = bw (between the │ characters).
// All lines have visible width = bw + 4 (2 indent + left │ + interior + right │).
func (rs *ReportsScreen) renderCard(report *entity.ExecutionReport, num int, width int) string {
	bw := width - 6
	if bw < 40 {
		bw = 40
	}

	// Status icon and label
	var statusIcon, statusLabel string
	switch report.Status {
	case "confirmed":
		statusIcon = "✓"
		statusLabel = SuccessStyle.Render("CONFIRMED")
	case "reverted":
		statusIcon = "✗"
		statusLabel = ErrorStyle.Render("REVERTED")
	default:
		statusIcon = "◌"
		statusLabel = WarnStyle.Render("PENDING")
	}

	// Card header line
	header := fmt.Sprintf(" %s %s  │  %s  │  %s",
		statusIcon, statusLabel,
		ToolLabelStyle.Render(report.ChainName),
		MutedStyle.Render(report.Timestamp.Format("2006-01-02 15:04:05")),
	)

	// Details line
	amountStr := report.TotalAmount
	if len(amountStr) > 20 {
		amountStr = amountStr[:20] + "…"
	}
	details := fmt.Sprintf(" Recipients: %s  │  Amount: %s  │  Gas: %s  │  Block: %s",
		ToolLabelStyle.Render(fmt.Sprintf("%d", report.Recipients)),
		ToolLabelStyle.Render(amountStr),
		ToolLabelStyle.Render(fmt.Sprintf("%d", report.GasUsed)),
		ToolLabelStyle.Render(fmt.Sprintf("%d", report.BlockNumber)),
	)

	// Token line
	tokenLine := fmt.Sprintf(" Token: %s", ToolLabelStyle.Render(report.Token))

	// Tx hash — full display, no truncation
	txLine := fmt.Sprintf(" Tx: %s", ToolLabelStyle.Render(report.TxHash))

	// Build card with box drawing — all lines align to bw + 4 visible width
	cardTitle := fmt.Sprintf(" Report #%d ", num)
	topDashes := maxInt(0, bw-len(cardTitle)-1)
	topBorder := "  ┌─" + cardTitle + strings.Repeat("─", topDashes) + "┐"
	separator := "  ├" + strings.Repeat("─", bw) + "┤"
	bottomBorder := "  └" + strings.Repeat("─", bw) + "┘"

	var lines []string
	lines = append(lines, topBorder)
	lines = append(lines, cardLine(header, bw))
	lines = append(lines, separator)
	lines = append(lines, cardLine(details, bw))
	lines = append(lines, cardLine(tokenLine, bw))
	lines = append(lines, cardLine(txLine, bw))
	lines = append(lines, bottomBorder)

	return strings.Join(lines, "\n")
}

// cardLine renders a single content line inside a card box.
// bw is the interior width (between the two │ characters).
// Layout: "  │ " + content + padding + "│"  where padding = bw - 1 - visibleLen(content)
func cardLine(content string, bw int) string {
	visLen := visibleStrLen(content)
	padding := bw - 1 - visLen
	if padding < 0 {
		padding = 0
	}
	return "  │ " + content + strings.Repeat(" ", padding) + "│"
}

// visibleStrLen returns the visible character count (ignoring ANSI escape sequences).
func visibleStrLen(s string) int {
	inEscape := false
	count := 0
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
			}
			continue
		}
		count++
	}
	return count
}

// Summary returns a plain-text summary string of all reports.
func (rs *ReportsScreen) Summary() string {
	if !rs.loaded {
		rs.LoadReports()
	}
	if len(rs.reports) == 0 {
		return "No reports found"
	}

	confirmed, reverted, pending := 0, 0, 0
	for _, r := range rs.reports {
		switch r.Status {
		case "confirmed":
			confirmed++
		case "reverted":
			reverted++
		default:
			pending++
		}
	}

	return fmt.Sprintf("Reports: %d total — %d confirmed, %d reverted, %d pending", len(rs.reports), confirmed, reverted, pending)
}

// RecentReport returns the most recent report, or nil if none.
func (rs *ReportsScreen) RecentReport() *entity.ExecutionReport {
	if !rs.loaded {
		rs.LoadReports()
	}
	if len(rs.reports) == 0 {
		return nil
	}
	return rs.reports[0]
}

// AddReport adds a report to the in-memory list (for real-time updates).
func (rs *ReportsScreen) AddReport(report *entity.ExecutionReport) {
	if !rs.loaded {
		rs.LoadReports()
	}
	rs.reports = append([]*entity.ExecutionReport{report}, rs.reports...)
}

// GetReportsByDate returns reports filtered by date.
func (rs *ReportsScreen) GetReportsByDate(date time.Time) []*entity.ExecutionReport {
	if !rs.loaded {
		rs.LoadReports()
	}
	dateStr := date.Format("2006-01-02")
	var filtered []*entity.ExecutionReport
	for _, r := range rs.reports {
		if r.Timestamp.Format("2006-01-02") == dateStr {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// Count returns the total number of reports.
func (rs *ReportsScreen) Count() int {
	if !rs.loaded {
		rs.LoadReports()
	}
	return len(rs.reports)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
