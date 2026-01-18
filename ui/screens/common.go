package screens

import (
	ui "officina/ui"

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

// GetTableStyles restituisce gli stili per la tabella
func GetTableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(ui.ColorBorder).
		BorderBottom(true).
		Bold(true).
		Foreground(ui.ColorPrimary)

	s.Selected = s.Selected.
		Foreground(ui.ColorText).
		Background(ui.ColorPrimary).
		Bold(false)

	return s
}

// RenderHeader renderizza l'header comune
func RenderHeader(title string, width int) string {
	titleBar := lipgloss.NewStyle().
		Background(ui.ColorPrimary).
		Foreground(lipgloss.Color("#000000")).
		Bold(true).
		Padding(0, 2).
		Width(width).
		Align(lipgloss.Center).
		Render(title)

	return titleBar
}

// RenderFooter renderizza il footer comone
func RenderFooter(width int, version string) string {
	footer := ui.FooterStyle.
		Width(width).
		Align(lipgloss.Center).
		Render("Officina Management System " + version + " • [Q] Esci • [ESC] Indietro")

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
	return ui.InfoBadge.Render(text)
}

// WarningBadge crea un badge di warning
func WarningBadge(text string) string {
	return ui.WarningBadge.Render(text)
}

// SuccessBadge crea un badge di successo
func SuccessBadge(text string) string {
	return ui.SuccessBadge.Render(text)
}

// ErrorBadge crea un badge di errore
func ErrorBadge(text string) string {
	return ui.ErrorBadge.Render(text)
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
