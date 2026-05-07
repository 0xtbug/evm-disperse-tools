package tui

import "github.com/charmbracelet/lipgloss"

var (
	bgPrimary = lipgloss.Color("#1a1b26") // Tokyo Night Dark
	accent    = lipgloss.Color("#bb9af7") // Soft Violet
	selection = lipgloss.Color("#2f334d") // Dark Gray Blue
	cyan      = lipgloss.Color("#7dcfff") // Sky Cyan
	emerald   = lipgloss.Color("#9ece6a") // Grass Green
	rose      = lipgloss.Color("#f7768e") // Soft Red/Pink
	amber     = lipgloss.Color("#e0af68") // Soft Orange
	gray      = lipgloss.Color("#565f89") // Muted Blue Gray
	white     = lipgloss.Color("#c0caf5") // Text Color
	border    = lipgloss.Color("#414868") // Border Color
	statusBg  = lipgloss.Color("#16161e") // Status bar background
)

// General Typography
var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true).
			MarginLeft(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(gray).
			Italic(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(gray)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(emerald)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(rose)

	WarnStyle = lipgloss.NewStyle().
			Foreground(amber)
)

// Dashboard Pane Styling
var (
	BasePaneStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(border).
			Padding(0, 1)

	ActivePaneStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Padding(0, 1)

	PaneTitleStyle = lipgloss.NewStyle().
			Foreground(white).
			Background(border).
			Padding(0, 1).
			Bold(true)
)

// Menu Styling
var (
	ToolLabelStyle = lipgloss.NewStyle().
			Foreground(white).
			Bold(true)

	ToolDescStyle = lipgloss.NewStyle().
			Foreground(gray)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(accent).
			Background(selection).
			Bold(true)

	NavItemStyle = lipgloss.NewStyle().
			Foreground(white).
			PaddingLeft(2)

	ActiveNavItemStyle = lipgloss.NewStyle().
				Foreground(accent).
				Bold(true).
				Background(selection).
				PaddingLeft(1)

	KeyStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true)
)

// Section and Header
var (
	SectionHeaderStyle = lipgloss.NewStyle().
				Foreground(gray).
				Bold(true).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(lipgloss.Color("#24283b"))

	SidebarTitleStyle = lipgloss.NewStyle().
				Foreground(accent).
				Bold(true)

	SidebarLabelStyle = lipgloss.NewStyle().
				Foreground(gray).
				Width(10)
)

// Status Bar
var (
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(gray).
			Background(statusBg).
			Padding(0, 1)

	StatusKeyStyle = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true)
)

// Metadata
var (
	MetaKeyStyle = lipgloss.NewStyle().Foreground(gray).Faint(true)
	MetaValStyle = lipgloss.NewStyle().Foreground(cyan).Bold(true)
)

// Logging
var (
	LogTimestampStyle = lipgloss.NewStyle().
				Foreground(cyan).
				Faint(true)

	LogMessageStyle = lipgloss.NewStyle().
			Foreground(gray)

	ClickableStyle = lipgloss.NewStyle().
			Foreground(cyan)
)

// Tabs
var (
	ActiveTabStyle = lipgloss.NewStyle().
			Foreground(white).
			Background(accent).
			Padding(0, 2).
			Bold(true)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(gray).
				Padding(0, 2)
)

// App Layout
var (
	AppStyle = lipgloss.NewStyle().
			Background(bgPrimary)

	DocStyle = lipgloss.NewStyle().
			Padding(1, 2)
)

// Highlight creates highlighted text
func Highlight(text string) string {
	return lipgloss.NewStyle().
		Foreground(accent).
		Bold(true).
		Render(text)
}
