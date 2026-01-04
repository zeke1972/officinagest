package screens

import (
	"fmt"
	"officina/database"
	"officina/utils"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AgendaMode rappresenta la modalità corrente
type AgendaMode int

const (
	AgModeList AgendaMode = iota
	AgModeAdd
	AgModeEdit
)

// AgendaModel gestisce la schermata agenda
type AgendaModel struct {
	db            *database.DB
	table         table.Model
	inputs        []textinput.Model
	mode          AgendaMode
	focusIndex    int
	selectedID    int
	err           error
	msg           string
	width         int
	height        int
	selectionMode bool
	veicoloTable  table.Model
	veicoloFilter textinput.Model
	veicoloID     int
	veicoloInfo   string
	showConfirm   bool
	deletingID    int
	todayCount    int
}

// NewAgendaModel crea una nuova istanza del model agenda
func NewAgendaModel(db *database.DB) *AgendaModel {
	// Tabella principale appuntamenti
	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "📅 Data/Ora", Width: 16},
		{Title: "🚗 Veicolo", Width: 12},
		{Title: "👤 Cliente", Width: 20},
		{Title: "📝 Note", Width: 35},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(12),
	)
	t.SetStyles(GetTableStyles())

	// Configurazione inputs
	inputs := make([]textinput.Model, 4)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "[ INVIO PER SCEGLIERE VEICOLO ]"
	inputs[0].Width = 50

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Data (GG/MM/AAAA)"
	inputs[1].CharLimit = 10
	inputs[1].Width = 30

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Ora (HH:MM)"
	inputs[2].CharLimit = 5
	inputs[2].Width = 20

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Note appuntamento..."
	inputs[3].Width = 60

	// Tabella selezione veicoli
	vCols := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Targa", Width: 10},
		{Title: "Modello", Width: 25},
		{Title: "Proprietario", Width: 25},
	}

	vt := table.New(
		table.WithColumns(vCols),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	vt.SetStyles(GetTableStyles())

	vf := textinput.New()
	vf.Placeholder = "🔍 Cerca veicolo..."
	vf.Width = 50

	m := &AgendaModel{
		db:            db,
		table:         t,
		inputs:        inputs,
		mode:          AgModeList,
		veicoloTable:  vt,
		veicoloFilter: vf,
	}

	m.Refresh()
	return m
}

// Refresh aggiorna la lista degli appuntamenti
func (m *AgendaModel) Refresh() {
	list, err := m.db.ListAppuntamenti()
	if err != nil {
		m.err = fmt.Errorf("errore caricamento appuntamenti: %w", err)
		return
	}

	rows := []table.Row{}
	m.todayCount = 0
	today := time.Now().Format("2006-01-02")

	for _, a := range list {
		veicolo := fmt.Sprintf("ID:%d", a.VeicoloID)
		cliente := "—"

		v, err := m.db.GetVeicolo(a.VeicoloID)
		if err == nil && v != nil {
			veicolo = v.Targa

			if v.ClienteID > 0 {
				c, err := m.db.GetCliente(v.ClienteID)
				if err == nil && c != nil {
					cliente = utils.Truncate(fmt.Sprintf("%s %s", c.Cognome, c.Nome), 20)
				}
			}
		}

		if a.DataOra.Format("2006-01-02") == today {
			m.todayCount++
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("%d", a.ID),
			utils.FormatDateTime(a.DataOra),
			veicolo,
			cliente,
			utils.Truncate(a.Nota, 35),
		})
	}

	m.table.SetRows(rows)
	m.err = nil
}

// updateVeicoloTable aggiorna la tabella veicoli con filtro
func (m *AgendaModel) updateVeicoloTable() {
	veicoli, err := m.db.ListVeicoli()
	if err != nil {
		m.err = fmt.Errorf("errore caricamento veicoli: %w", err)
		return
	}

	filter := strings.ToUpper(strings.TrimSpace(m.veicoloFilter.Value()))
	rows := []table.Row{}

	for _, v := range veicoli {
		prop := "N/D"
		if v.ClienteID > 0 {
			c, err := m.db.GetCliente(v.ClienteID)
			if err == nil && c != nil {
				prop = fmt.Sprintf("%s %s", c.Cognome, c.Nome)
			}
		}

		targa := strings.ToUpper(v.Targa)
		modello := strings.ToUpper(v.Marca + " " + v.Modello)
		proprietario := strings.ToUpper(prop)

		if filter == "" ||
			strings.Contains(targa, filter) ||
			strings.Contains(modello, filter) ||
			strings.Contains(proprietario, filter) {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", v.ID),
				v.Targa,
				utils.Truncate(v.Marca+" "+v.Modello, 25),
				utils.Truncate(prop, 25),
			})
		}
	}

	m.veicoloTable.SetRows(rows)
}

// resetForm resetta il form
func (m *AgendaModel) resetForm() {
	for i := range m.inputs {
		m.inputs[i].SetValue("")
	}

	// Imposta domani alle 9:00 come default
	domani := time.Now().AddDate(0, 0, 1)
	m.inputs[1].SetValue(domani.Format("02/01/2006"))
	m.inputs[2].SetValue("09:00")

	m.veicoloID = 0
	m.veicoloInfo = ""
	m.selectedID = 0
	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.updateFocus()
}

// loadIntoForm carica un appuntamento nel form
func (m *AgendaModel) loadIntoForm(id int) {
	a, err := m.db.GetAppuntamento(id)
	if err != nil {
		m.err = fmt.Errorf("errore caricamento appuntamento: %w", err)
		return
	}

	m.selectedID = id
	m.veicoloID = a.VeicoloID

	// Carica info veicolo
	v, err := m.db.GetVeicolo(a.VeicoloID)
	if err == nil && v != nil {
		m.inputs[0].SetValue(v.Targa)
		m.veicoloInfo = fmt.Sprintf("%s %s", v.Marca, v.Modello)
	} else {
		m.inputs[0].SetValue(fmt.Sprintf("ID:%d", a.VeicoloID))
		m.veicoloInfo = "Veicolo non trovato"
	}

	m.inputs[1].SetValue(a.DataOra.Format("02/01/2006"))
	m.inputs[2].SetValue(a.DataOra.Format("15:04"))
	m.inputs[3].SetValue(a.Nota)

	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.updateFocus()
}

// updateFocus aggiorna il focus tra i campi
func (m *AgendaModel) updateFocus() {
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

// validate valida i dati del form
func (m *AgendaModel) validate() error {
	if m.veicoloID == 0 {
		return fmt.Errorf("seleziona un veicolo")
	}

	dateStr := strings.TrimSpace(m.inputs[1].Value())
	timeStr := strings.TrimSpace(m.inputs[2].Value())

	if dateStr == "" {
		return fmt.Errorf("data obbligatoria")
	}

	if timeStr == "" {
		return fmt.Errorf("ora obbligatoria")
	}

	_, err := time.Parse("02/01/2006 15:04", dateStr+" "+timeStr)
	if err != nil {
		return fmt.Errorf("data/ora non valida (formato: GG/MM/AAAA HH:MM)")
	}

	return nil
}

// save salva l'appuntamento corrente
func (m *AgendaModel) save() error {
	if err := m.validate(); err != nil {
		return err
	}

	dateStr := strings.TrimSpace(m.inputs[1].Value())
	timeStr := strings.TrimSpace(m.inputs[2].Value())
	dt, _ := time.Parse("02/01/2006 15:04", dateStr+" "+timeStr)

	a := &database.Appuntamento{
		VeicoloID: m.veicoloID,
		DataOra:   dt,
		Nota:      strings.TrimSpace(m.inputs[3].Value()),
	}

	if m.mode == AgModeAdd {
		if err := m.db.CreateAppuntamento(a); err != nil {
			return fmt.Errorf("errore creazione: %w", err)
		}
		m.msg = "✓ Appuntamento creato con successo"
	} else {
		a.ID = m.selectedID
		if err := m.db.UpdateAppuntamento(a); err != nil {
			return fmt.Errorf("errore aggiornamento: %w", err)
		}
		m.msg = "✓ Appuntamento aggiornato con successo"
	}

	m.mode = AgModeList
	m.selectedID = 0
	m.Refresh()
	return nil
}

// Init implementa tea.Model
func (m *AgendaModel) Init() tea.Cmd {
	return nil
}

// Update implementa tea.Model
func (m *AgendaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Gestione resize
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	// Popup selezione veicolo
	if m.selectionMode {
		return m.handleVehicleSelection(msg)
	}

	// Conferma eliminazione
	if m.showConfirm {
		return m.handleDeleteConfirmation(msg)
	}

	// Gestione ESC
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		if m.mode != AgModeList {
			m.mode = AgModeList
			m.err = nil
			m.msg = ""
			return m, nil
		}
		return m, func() tea.Msg { return ChangeScreenMsg(StateMenu) }
	}

	// Modalità Lista
	if m.mode == AgModeList {
		return m.handleListMode(msg)
	}

	// Modalità Form (Add/Edit)
	return m.handleFormMode(msg)
}

// handleVehicleSelection gestisce la selezione del veicolo
func (m *AgendaModel) handleVehicleSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.selectionMode = false
			m.updateFocus()
			return m, nil

		case "enter":
			if row := m.veicoloTable.SelectedRow(); len(row) > 0 {
				id, err := strconv.Atoi(row[0])
				if err == nil {
					m.veicoloID = id
					m.inputs[0].SetValue(row[1])
					m.veicoloInfo = row[2] + " - " + row[3]
					m.selectionMode = false
					m.focusIndex = 1
					m.updateFocus()
				}
			}
			return m, nil
		}

		var cmdF, cmdT tea.Cmd
		m.veicoloFilter, cmdF = m.veicoloFilter.Update(msg)
		m.updateVeicoloTable()
		m.veicoloTable, cmdT = m.veicoloTable.Update(msg)
		return m, tea.Batch(cmdF, cmdT)
	}

	return m, nil
}

// handleDeleteConfirmation gestisce la conferma eliminazione
func (m *AgendaModel) handleDeleteConfirmation(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "y", "Y":
			if err := m.db.DeleteAppuntamento(m.deletingID); err != nil {
				m.err = fmt.Errorf("errore eliminazione: %w", err)
			} else {
				m.msg = "✓ Appuntamento eliminato con successo"
			}
			m.showConfirm = false
			m.Refresh()

		case "n", "N", "esc":
			m.showConfirm = false
		}
		return m, nil
	}

	return m, nil
}

// handleListMode gestisce gli eventi in modalità lista
func (m *AgendaModel) handleListMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			m.mode = AgModeAdd
			m.resetForm()
			return m, nil

		case "e", "enter":
			if row := m.table.SelectedRow(); len(row) > 0 {
				id, err := strconv.Atoi(row[0])
				if err == nil {
					m.loadIntoForm(id)
					m.mode = AgModeEdit
				}
			}
			return m, nil

		case "x", "d":
			if row := m.table.SelectedRow(); len(row) > 0 {
				id, err := strconv.Atoi(row[0])
				if err == nil {
					m.deletingID = id
					m.showConfirm = true
				}
			}
			return m, nil
		}

		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleFormMode gestisce gli eventi in modalità form
func (m *AgendaModel) handleFormMode(msg tea.Msg) (tea.Model, tea.Cmd) {

	if k, ok := msg.(tea.KeyMsg); ok {
		// Campo veicolo: apre popup selezione
		if m.focusIndex == 0 && (k.String() == "enter" || k.String() == " ") {
			m.selectionMode = true
			m.veicoloFilter.SetValue("")
			m.updateVeicoloTable()
			m.veicoloFilter.Focus()
			return m, nil
		}

		switch k.String() {
		case "enter":
			// Se siamo sull'ultimo campo, salva
			if m.focusIndex == len(m.inputs)-1 {
				if err := m.save(); err != nil {
					m.err = err
					return m, nil
				}
				return m, nil
			}

			// Altrimenti passa al prossimo campo
			m.focusIndex++
			if m.focusIndex >= len(m.inputs) {
				m.focusIndex = 0
			}
			m.updateFocus()
			return m, nil

		case "tab", "down":
			m.focusIndex++
			if m.focusIndex >= len(m.inputs) {
				m.focusIndex = 0
			}
			m.updateFocus()
			return m, nil

		case "shift+tab", "up":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}
			m.updateFocus()
			return m, nil
		}
	}

	// Update inputs (escluso campo veicolo readonly)
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := 1; i < len(m.inputs); i++ {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

// View implementa tea.Model
func (m *AgendaModel) View() string {
	if m.width == 0 {
		return "Caricamento..."
	}

	width := 95
	if m.width > 0 {
		width = min(m.width, 110)
	}

	// Popup selezione veicolo
	if m.selectionMode {
		return m.renderVehicleSelection(width)
	}

	// Dialog conferma eliminazione
	if m.showConfirm {
		return RenderConfirmDialog(
			fmt.Sprintf("Eliminare l'appuntamento #%d?", m.deletingID),
			m.width,
			m.height,
		)
	}

	// Vista principale
	if m.mode == AgModeList {
		return m.renderListView(width)
	}

	return m.renderFormView(width)
}

// renderListView renderizza la vista lista
func (m *AgendaModel) renderListView(width int) string {
	title := "AGENDA APPUNTAMENTI"
	header := RenderHeader(title, width)

	// Info appuntamenti oggi
	var infoBox string
	if m.todayCount > 0 {
		badge := WarningBadge(fmt.Sprintf(" 🔔 %d appuntamenti oggi ", m.todayCount))
		infoBox = lipgloss.NewStyle().
			MarginBottom(1).
			Render(badge)
	}

	// Help text
	helpText := lipgloss.NewStyle().
		MarginBottom(1).
		Foreground(ColorSubText).
		Render("[N] Nuovo • [E/↵] Modifica • [X/D] Elimina • [ESC] Menu")

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		infoBox,
		helpText,
		m.table.View(),
	)

	// Footer con messaggi
	footer := RenderFooter(width)

	if m.err != nil {
		footer = "\n" + ErrorStyle.Render("✗ "+m.err.Error()) + "\n" + footer
	}

	if m.msg != "" {
		footer = "\n" + SuccessStyle.Render(m.msg) + "\n" + footer
	}

	// Composizione finale
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		lipgloss.NewStyle().Padding(0, 2).Render(body),
		"",
		footer,
	)

	box := MainBoxStyle.Copy().Width(width - 4).Render(content)

	if m.width > 0 && m.height > 0 {
		return CenterContent(m.width, m.height, box)
	}

	return "\n" + box
}

// renderFormView renderizza la vista form
func (m *AgendaModel) renderFormView(width int) string {
	title := "NUOVO APPUNTAMENTO"
	if m.mode == AgModeEdit {
		title = fmt.Sprintf("MODIFICA APPUNTAMENTO #%d", m.selectedID)
	}

	header := RenderHeader(title, width)

	// Form
	var form strings.Builder
	labels := []string{"🚗 Veicolo", "📅 Data", "🕐 Ora", "📝 Note"}

	for i := 0; i < len(m.inputs); i++ {
		labelStyle := LabelStyle
		if i == m.focusIndex {
			labelStyle = LabelFocusedStyle
		}

		view := m.inputs[i].View()

		// Campo veicolo (readonly con highlight)
		if i == 0 {
			if m.focusIndex == 0 {
				if m.veicoloInfo == "" {
					view = lipgloss.NewStyle().
						Foreground(ColorPrimary).
						Render("[ INVIO PER SCEGLIERE ]")
				} else {
					view = lipgloss.NewStyle().
						Foreground(ColorPrimary).
						Bold(true).
						Render(m.inputs[0].Value() + " - " + m.veicoloInfo)
				}
			} else {
				if m.veicoloInfo == "" {
					view = HelpStyle.Render("[ Nessuno ]")
				} else {
					view = m.inputs[0].Value() + " " + HelpStyle.Render("("+m.veicoloInfo+")")
				}
			}
		}

		form.WriteString(fmt.Sprintf("%s %s\n",
			labelStyle.Render(labels[i]+":"),
			view))

		if i == 0 {
			form.WriteString("\n")
		}
	}

	form.WriteString("\n")
	form.WriteString(HelpStyle.Render("[Tab/↑↓] Naviga • [↵] Salva • [Esc] Annulla"))

	body := form.String()

	// Footer con messaggi
	footer := RenderFooter(width)

	if m.err != nil {
		footer = "\n" + ErrorStyle.Render("✗ "+m.err.Error()) + "\n" + footer
	}

	// Composizione finale
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		lipgloss.NewStyle().Padding(0, 2).Render(body),
		"",
		footer,
	)

	box := MainBoxStyle.Copy().Width(width - 4).Render(content)

	if m.width > 0 && m.height > 0 {
		return CenterContent(m.width, m.height, box)
	}

	return "\n" + box
}

// renderVehicleSelection renderizza il popup di selezione veicolo
func (m *AgendaModel) renderVehicleSelection(width int) string {
	title := TitleStyle.Render("🔍 SELEZIONA VEICOLO")
	filter := fmt.Sprintf("Cerca: %s", m.veicoloFilter.View())

	inner := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		filter,
		"",
		m.veicoloTable.View(),
		"",
		HelpStyle.Render("[↑↓] Naviga • [↵] Seleziona • [Esc] Annulla"),
	)

	box := MainBoxStyle.Copy().
		Width(min(width-10, 75)).
		Render(inner)

	if m.width > 0 && m.height > 0 {
		return CenterContent(m.width, m.height, box)
	}

	return box
}
