package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
const (
	ColorPrimary   = lipgloss.Color("#00D9FF")
	ColorSecondary = lipgloss.Color("#7D56F4")
	ColorSuccess   = lipgloss.Color("#04B575")
	ColorError     = lipgloss.Color("#FF0040")
	ColorWarning   = lipgloss.Color("#FFB86C")
	ColorHighlight = lipgloss.Color("#BD93F9")
	ColorText      = lipgloss.Color("#FAFAFA")
	ColorSubText   = lipgloss.Color("#626262")
	ColorBorder    = lipgloss.Color("#383838")
	ColorBgDark    = lipgloss.Color("#1A1A1A")
	ColorBgMedium  = lipgloss.Color("#2D2D2D")
	ColorBgLight   = lipgloss.Color("#3D3D3D")
)

// Base styling
var (
	BaseStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorBgDark)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Italic(true)

	HeaderStyle = lipgloss.NewStyle().
			Background(ColorPrimary).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 2)

	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorSubText).
			Padding(1, 2)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorSubText)

	MainBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Background(ColorBgDark).
			Padding(1, 2)
)

// Input field styling
var (
	InputFieldStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	InputFieldFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 1)

	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorSubText).
			Width(20).
			Align(lipgloss.Right)

	LabelFocusedStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				Width(20).
				Align(lipgloss.Right)

	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Bold(true)
)

// List/table styling
var (
	ListItemStyle = lipgloss.NewStyle().
			Padding(0, 1)

	ListItemSelectedStyle = lipgloss.NewStyle().
				Background(ColorPrimary).
				Foreground(lipgloss.Color("#000000"))

	ListItemCursorStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)
)

// Badge styling
var (
	InfoBadge = lipgloss.NewStyle().
			Background(ColorHighlight).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1)

	SuccessBadge = lipgloss.NewStyle().
			Background(ColorSuccess).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1)

	WarningBadge = lipgloss.NewStyle().
			Background(ColorWarning).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1)

	ErrorBadge = lipgloss.NewStyle().
			Background(ColorError).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1)
)

// Button styling
var (
	ButtonStyle = lipgloss.NewStyle().
			Background(ColorBgMedium).
			Foreground(ColorText).
			Padding(0, 2).
			MarginRight(1)

	ButtonPrimaryStyle = lipgloss.NewStyle().
				Background(ColorPrimary).
				Foreground(lipgloss.Color("#000000")).
				Bold(true).
				Padding(0, 2).
				MarginRight(1)

	ButtonSuccessStyle = lipgloss.NewStyle().
				Background(ColorSuccess).
				Foreground(lipgloss.Color("#000000")).
				Bold(true).
				Padding(0, 2).
				MarginRight(1)

	ButtonDangerStyle = lipgloss.NewStyle().
				Background(ColorError).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).
				Padding(0, 2).
				MarginRight(1)
)
