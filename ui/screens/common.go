package screens

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// AppState rappresenta lo stato corrente dell'applicazione
type AppState int

const (
	StateMenu AppState = iota
	StateClienti
	StateFornitori
	StateVeicoli
	StateCommesse
	StateAgenda
	StatePrimaNota
	StateOperatori
	StatePreventivi
	StateFatture
)

// ChangeScreenMsg è il messaggio per cambiare schermata
type ChangeScreenMsg AppState

// Colori e stili globali
var (
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
	ColorBgLight   = lipgloss.Color("#2D2D2D")
)

// Stili comuni
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Background(ColorBgDark).
			Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Italic(true)

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
			Padding(1, 2)

	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorSubText).
			Width(25).
			Align(lipgloss.Right)

	LabelFocusedStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				Width(25).
				Align(lipgloss.Right)
)

// GetTableStyles restituisce gli stili per la tabella
func GetTableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(ColorBorder).
		BorderBottom(true).
		Bold(true).
		Foreground(ColorPrimary)

	s.Selected = s.Selected.
		Foreground(ColorText).
		Background(ColorPrimary).
		Bold(false)

	return s
}

// RenderHeader renderizza l'header comune
func RenderHeader(title string, width int) string {
	titleBar := lipgloss.NewStyle().
		Background(ColorPrimary).
		Foreground(lipgloss.Color("#000000")).
		Bold(true).
		Padding(0, 2).
		Width(width).
		Align(lipgloss.Center).
		Render(title)

	return titleBar
}

// RenderFooter renderizza il footer comune
func RenderFooter(width int) string {
	footer := lipgloss.NewStyle().
		Foreground(ColorSubText).
		Width(width).
		Align(lipgloss.Center).
		Render("Officina Management System v1.0 • [Q] Esci • [ESC] Indietro")

	return footer
}

// CenterContent centra il contenuto sullo schermo
func CenterContent(termWidth, termHeight int, content string) string {
	return lipgloss.Place(
		termWidth,
		termHeight,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// InfoBadge crea un badge informativo
func InfoBadge(text string) string {
	return lipgloss.NewStyle().
		Background(ColorHighlight).
		Foreground(lipgloss.Color("#000000")).
		Bold(true).
		Padding(0, 1).
		Render(text)
}

// WarningBadge crea un badge di warning
func WarningBadge(text string) string {
	return lipgloss.NewStyle().
		Background(ColorWarning).
		Foreground(lipgloss.Color("#000000")).
		Bold(true).
		Padding(0, 1).
		Render(text)
}

// SuccessBadge crea un badge di successo
func SuccessBadge(text string) string {
	return lipgloss.NewStyle().
		Background(ColorSuccess).
		Foreground(lipgloss.Color("#000000")).
		Bold(true).
		Padding(0, 1).
		Render(text)
}

// ErrorBadge crea un badge di errore
func ErrorBadge(text string) string {
	return lipgloss.NewStyle().
		Background(ColorError).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Padding(0, 1).
		Render(text)
}

// min restituisce il minimo tra due interi
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max restituisce il massimo tra due interi
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
