package screens

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// AppState rappresenta lo stato corrente dell'applicazione
type AppState int

const (
	StateMenu AppState = iota
	StateVeicoli
	StateCommesse
	StateAgenda
	StateOperatori
	StatePreventivi
	StateFatture
	StateClienti
	StatePrimaNota
)

// ChangeScreenMsg è il messaggio per cambiare schermata
type ChangeScreenMsg AppState

// === PALETTE COLORI 2026 - Professional Dark Theme ===
var (
	ColorPrimary   = lipgloss.Color("#00D4FF") // Cyan brillante
	ColorSecondary = lipgloss.Color("#1E3A8A") // Navy scuro
	ColorText      = lipgloss.Color("#E2E8F0") // Slate chiaro
	ColorSubText   = lipgloss.Color("#94A3B8") // Slate medio
	ColorSuccess   = lipgloss.Color("#10B981") // Verde smeraldo
	ColorError     = lipgloss.Color("#EF4444") // Rosso
	ColorWarning   = lipgloss.Color("#F59E0B") // Arancione
	ColorHighlight = lipgloss.Color("#0EA5E9") // Sky
	ColorBgDark    = lipgloss.Color("#0F172A") // Slate 900
	ColorBgLight   = lipgloss.Color("#1E293B") // Slate 800
	ColorBorder    = lipgloss.Color("#334155") // Slate 700
)

// === STILI GLOBALI ===
var (
	// Base
	BaseStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	// Contenitore principale
	MainBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2).
			Foreground(ColorText)

	// Titoli
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Padding(0, 1)

	// Header applicazione
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorPrimary).
			Padding(0, 2)

	// Sottotitoli
	SubTitleStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Bold(true).
			Padding(0, 0, 1, 0)

	// Messaggi di errore
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	// Messaggi di successo
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	// Messaggi di avviso
	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	// Testo di aiuto
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorSubText).
			Italic(true)

	// Item selezionato
	SelectedItemStyle = lipgloss.NewStyle().
				Background(ColorPrimary).
				Foreground(ColorBgDark).
				Bold(true).
				Padding(0, 1)

	// Item normale
	NormalItemStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Padding(0, 1)

	// Label form
	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorSubText).
			Width(15)

	// Label form con focus
	LabelFocusedStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				Width(15)
)

// RenderHeader renderizza l'header dell'applicazione con titolo e timestamp
func RenderHeader(title string, width int) string {
	appTitle := HeaderStyle.Render("⚙️  OFFICINA MANAGER v2.0 by emC & claude")
	dateStr := lipgloss.NewStyle().
		Foreground(ColorSubText).
		Render(time.Now().Format("02/01/2006 15:04"))

	// Calcola spacing
	totalWidth := lipgloss.Width(appTitle) + lipgloss.Width(dateStr)
	spacing := width - totalWidth - 8

	if spacing < 0 {
		spacing = 2
	}

	spacer := strings.Repeat(" ", spacing)

	headerLine := lipgloss.JoinHorizontal(
		lipgloss.Top,
		appTitle,
		spacer,
		dateStr,
	)

	// Titolo sezione
	sectionTitle := SubTitleStyle.Render(title)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerLine,
		"",
		sectionTitle,
	)
}

// RenderFooter renderizza il footer con help text
func RenderFooter(width int) string {
	helpText := HelpStyle.Render("↑↓/jk Naviga • ↵ Seleziona • esc Menu/Indietro • q/ctrl+c Esci")
	separator := lipgloss.NewStyle().
		Foreground(ColorBorder).
		Render(strings.Repeat("─", min(width-4, 100)))

	return separator + "\n" + helpText
}

// RenderSeparator renderizza una linea di separazione
func RenderSeparator(width int) string {
	return lipgloss.NewStyle().
		Foreground(ColorBorder).
		Render(strings.Repeat("─", min(width, 100)))
}

// === BADGE HELPERS ===

// SuccessBadge crea un badge verde per messaggi di successo
func SuccessBadge(text string) string {
	return lipgloss.NewStyle().
		Background(ColorSuccess).
		Foreground(ColorBgDark).
		Bold(true).
		Padding(0, 1).
		Render(text)
}

// ErrorBadge crea un badge rosso per messaggi di errore
func ErrorBadge(text string) string {
	return lipgloss.NewStyle().
		Background(ColorError).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Padding(0, 1).
		Render(text)
}

// WarningBadge crea un badge arancione per avvisi
func WarningBadge(text string) string {
	return lipgloss.NewStyle().
		Background(ColorWarning).
		Foreground(ColorBgDark).
		Bold(true).
		Padding(0, 1).
		Render(text)
}

// InfoBadge crea un badge blu per informazioni
func InfoBadge(text string) string {
	return lipgloss.NewStyle().
		Background(ColorSecondary).
		Foreground(ColorText).
		Bold(true).
		Padding(0, 1).
		Render(text)
}

// PrimaryBadge crea un badge con colore primario
func PrimaryBadge(text string) string {
	return lipgloss.NewStyle().
		Background(ColorPrimary).
		Foreground(ColorBgDark).
		Bold(true).
		Padding(0, 1).
		Render(text)
}

// GetTableStyles restituisce gli stili per le tabelle
func GetTableStyles() table.Styles {
	s := table.DefaultStyles()

	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(ColorBorder).
		BorderBottom(true).
		Bold(true).
		Foreground(ColorPrimary).
		Padding(0, 1)

	s.Selected = s.Selected.
		Foreground(ColorBgDark).
		Background(ColorHighlight).
		Bold(true)

	s.Cell = s.Cell.
		Foreground(ColorText)

	return s
}

// CenterContent centra il contenuto nella finestra
func CenterContent(width, height int, content string) string {
	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// RenderPopup renderizza un popup centrato
func RenderPopup(title, content string, width, height int) string {
	popupStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 2).
		Width(min(width-10, 80))

	titleRendered := TitleStyle.Render(title)

	popup := lipgloss.JoinVertical(
		lipgloss.Left,
		titleRendered,
		"",
		content,
	)

	boxed := popupStyle.Render(popup)

	if width > 0 && height > 0 {
		return CenterContent(width, height, boxed)
	}

	return boxed
}

// RenderConfirmDialog renderizza un dialog di conferma
func RenderConfirmDialog(message string, width, height int) string {
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		ErrorStyle.Render("⚠️  "+message),
		"",
		HelpStyle.Render("[Y] Conferma  [N] Annulla"),
	)

	return RenderPopup("CONFERMA OPERAZIONE", content, width, height)
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
