package screens

import (
	"fmt"
	"officina/database"
	"officina/utils"
	"sort"
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
	AgendaList AgendaMode = iota
	AgendaAdd
	AgendaEdit
)

// AgendaModel gestisce la schermata agenda
type AgendaModel struct {
	db            *database.DB
	table         table.Model
	inputs        []textinput.Model
	mode          AgendaMode
	focusIndex    int
	selectedID    int
	veicoloID     int
	veicoloInfo   string
	err           error
	msg           string
	width         int
	height        int
	selectionMode bool
	veicoloTable  table.Model
	veicoloFilter textinput.Model
	showConfirm   bool
	deletingID    int
}

// NewAgendaModel crea una nuova istanza del model agenda
func NewAgendaModel(db *database.DB) AgendaModel {
	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Data e Ora", Width: 18},
		{Title: "Veicolo", Width: 18},
		{Title: "Proprietario", Width: 20},
		{Title: "Nota", Width: 30},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(12),
	)

	t.SetStyles(GetTableStyles())

	inputs := make([]textinput.Model, 4)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Data (GG/MM/AAAA)"
	inputs[0].CharLimit = 10
	inputs[0].Width = 30

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Ora (HH:MM)"
	inputs[1].CharLimit = 5
	inputs[1].Width = 20

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "[ INVIO PER SCEGLIERE VEICOLO ]"
	inputs[2].Width = 40

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Nota appuntamento (opzionale)"
	inputs[3].Width = 60

	veicoloCols := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Targa", Width: 10},
		{Title: "Modello", Width: 22},
		{Title: "Proprietario", Width: 20},
	}

	vt := table.New(
		table.WithColumns(veicoloCols),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	vt.SetStyles(GetTableStyles())

	vf := textinput.New()
	vf.Placeholder = "🔍 Cerca veicolo/targa/proprietario..."
	vf.Width = 50

	m := AgendaModel{
		db:            db,
		table:         t,
		inputs:        inputs,
		mode:          AgendaList,
		veicoloTable:  vt,
		veicoloFilter: vf,
	}

	m.Refresh()
	return m
}

// Refresh aggiorna la lista degli appuntamenti
func (m *AgendaModel) Refresh() {
	list, _ := m.db.ListAppuntamenti()

	sort.Slice(list, func(i, j int) bool {
		return list[i].DataOra.Before(list[j].DataOra)
	})

	rows := []table.Row{}
	now := time.Now()

	for _, a := range list {
		v, _ := m.db.GetVeicolo(a.VeicoloID)
		veicoloInfo := "N/D"
		proprietario := "N/D"

		if v != nil {
			veicoloInfo = fmt.Sprintf("%s (%s)", v.Marca+" "+v.Modello, v.Targa)

			if v.ClienteID > 0 {
				c, err := m.db.GetCliente(v.ClienteID)
				if err == nil && c != nil {
					proprietario = utils.Truncate(c.RagioneSociale, 20)
				}
			}
		}

		dataOraStr := a.DataOra.Format("02/01/2006 15:04")

		if a.DataOra.Before(now) {
			dataOraStr = "⏰ " + dataOraStr
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("%d", a.ID),
			dataOraStr,
			utils.Truncate(veicoloInfo, 18),
			proprietario,
			utils.Truncate(a.Nota, 30),
		})
	}

	m.table.SetRows(rows)
}

// updateVeicoloTable aggiorna la tabella veicoli con filtro
func (m *AgendaModel) updateVeicoloTable() {
	veicoli, _ := m.db.ListVeicoli()
	filter := strings.ToUpper(strings.TrimSpace(m.veicoloFilter.Value()))
	rows := []table.Row{}

	for _, v := range veicoli {
		targa := strings.ToUpper(v.Targa)
		modello := strings.ToUpper(v.Marca + " " + v.Modello)
		proprietario := "N/D"

		if v.ClienteID > 0 {
			c, err := m.db.GetCliente(v.ClienteID)
			if err == nil && c != nil {
				proprietario = c.RagioneSociale
			}
		}

		proprietarioUpper := strings.ToUpper(proprietario)

		if filter == "" ||
			strings.Contains(targa, filter) ||
			strings.Contains(modello, filter) ||
			strings.Contains(proprietarioUpper, filter) {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", v.ID),
				v.Targa,
				utils.Truncate(v.Marca+" "+v.Modello, 22),
				utils.Truncate(proprietario, 20),
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

	m.inputs[0].SetValue(time.Now().Format("02/01/2006"))
	m.inputs[1].SetValue("09:00")
	m.veicoloID = 0
	m.veicoloInfo = "Nessun veicolo"
	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.inputs[0].Focus()
}

// loadIntoForm carica un appuntamento nel form
func (m *AgendaModel) loadIntoForm(id int) {
	a, err := m.db.GetAppuntamento(id)
	if err != nil {
		m.err = fmt.Errorf("errore caricamento appuntamento: %w", err)
		return
	}

	m.selectedID = id
	m.inputs[0].SetValue(a.DataOra.Format("02/01/2006"))
	m.inputs[1].SetValue(a.DataOra.Format("15:04"))
	m.inputs[3].SetValue(a.Nota)
	m.veicoloID = a.VeicoloID

	if a.VeicoloID > 0 {
		v, err := m.db.GetVeicolo(a.VeicoloID)
		if err == nil && v != nil {
			m.veicoloInfo = fmt.Sprintf("%s %s (%s)", v.Marca, v.Modello, v.Targa)
			m.inputs[2].SetValue(m.veicoloInfo)
		}
	} else {
		m.veicoloInfo = "Nessun veicolo"
		m.inputs[2].SetValue("")
	}

	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.inputs[0].Focus()
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
	dateStr := strings.TrimSpace(m.inputs[0].Value())
	if len(dateStr) != 10 {
		return fmt.Errorf("data incompleta (formato: GG/MM/AAAA)")
	}

	_, err := time.Parse("02/01/2006", dateStr)
	if err != nil {
		return fmt.Errorf("data non valida")
	}

	timeStr := strings.TrimSpace(m.inputs[1].Value())
	if len(timeStr) != 5 {
		return fmt.Errorf("ora incompleta (formato: HH:MM)")
	}

	_, err = time.Parse("15:04", timeStr)
	if err != nil {
		return fmt.Errorf("ora non valida")
	}

	if m.veicoloID == 0 {
		return fmt.Errorf("seleziona un veicolo")
	}

	return nil
}

// save salva l'appuntamento corrente
func (m *AgendaModel) save() error {
	if err := m.validate(); err != nil {
		return err
	}

	dateStr := strings.TrimSpace(m.inputs[0].Value())
	timeStr := strings.TrimSpace(m.inputs[1].Value())

	dateTimeStr := dateStr + " " + timeStr
	dataOra, _ := time.Parse("02/01/2006 15:04", dateTimeStr)

	a := &database.Appuntamento{
		DataOra:   dataOra,
		VeicoloID: m.veicoloID,
		Nota:      m.inputs[3].Value(),
	}

	if m.mode == AgendaAdd {
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

	m.mode = AgendaList
	m.Refresh()
	return nil
}

// Init implementa tea.Model
func (m AgendaModel) Init() tea.Cmd {
	return nil
}

// Update implementa tea.Model
func (m AgendaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	if m.selectionMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.selectionMode = false
				m.updateFocus()
				return m, nil
			case "enter":
				if row := m.veicoloTable.SelectedRow(); len(row) > 0 {
					m.veicoloID, _ = strconv.Atoi(row[0])
					m.veicoloInfo = fmt.Sprintf("%s (%s)", row[2], row[1])
					m.inputs[2].SetValue(m.veicoloInfo)
					m.selectionMode = false
					m.focusIndex = 0
					m.updateFocus()
				}
				return m, nil
			}
		}

		var cmdF, cmdT tea.Cmd
		m.veicoloFilter, cmdF = m.veicoloFilter.Update(msg)
		m.updateVeicoloTable()
		m.veicoloTable, cmdT = m.veicoloTable.Update(msg)
		return m, tea.Batch(cmdF, cmdT)
	}

	if m.showConfirm {
		if k, ok := msg.(tea.KeyMsg); ok {
			switch k.String() {
			case "y", "Y":
				if err := m.db.DeleteAppuntamento(m.deletingID); err != nil {
					m.err = fmt.Errorf("errore eliminazione: %w", err)
				} else {
					m.msg = "✓ Appuntamento eliminato con successo"
				}
				m.Refresh()
				m.showConfirm = false
			case "n", "N", "esc":
				m.showConfirm = false
			}
			return m, nil
		}
	}

	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		if m.mode != AgendaList {
			m.mode = AgendaList
			m.err = nil
			m.msg = ""
			return m, nil
		}
		return m, func() tea.Msg { return ChangeScreenMsg(StateMenu) }
	}

	if m.mode == AgendaList {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "n":
				m.mode = AgendaAdd
				m.resetForm()
				return m, nil
			case "e", "enter":
				if row := m.table.SelectedRow(); len(row) > 0 {
					id, _ := strconv.Atoi(row[0])
					m.loadIntoForm(id)
					m.mode = AgendaEdit
				}
				return m, nil
			case "x", "d":
				if row := m.table.SelectedRow(); len(row) > 0 {
					id, _ := strconv.Atoi(row[0])
					m.deletingID = id
					m.showConfirm = true
				}
				return m, nil
			}
		}
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	if m.mode == AgendaAdd || m.mode == AgendaEdit {
		if k, ok := msg.(tea.KeyMsg); ok {
			if m.focusIndex == 2 && (k.String() == "enter" || k.String() == " ") {
				m.selectionMode = true
				m.veicoloFilter.SetValue("")
				m.updateVeicoloTable()
				m.veicoloFilter.Focus()
				return m, nil
			}

			switch k.String() {
			case "enter":
				if m.focusIndex != 2 {
					if err := m.save(); err != nil {
						m.err = err
						return m, nil
					}
				}
				return m, nil
			case "tab", "down":
				m.focusIndex++
				if m.focusIndex > 3 {
					m.focusIndex = 0
				}
				m.updateFocus()
				return m, nil
			case "shift+tab", "up":
				m.focusIndex--
				if m.focusIndex < 0 {
					m.focusIndex = 3
				}
				m.updateFocus()
				return m, nil
			}
		}

		cmds := make([]tea.Cmd, len(m.inputs))
		for i := range m.inputs {
			m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
		}
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

// View implementa tea.Model
func (m AgendaModel) View() string {
	width := 95
	if m.width > 0 {
		width = min(m.width, 110)
	}

	if m.showConfirm {
		var message strings.Builder
		message.WriteString(fmt.Sprintf("⚠️  ELIMINAZIONE APPUNTAMENTO #%d\n\n", m.deletingID))
		message.WriteString(WarningStyle.Render("Sei sicuro di voler procedere?\n"))
		message.WriteString(HelpStyle.Render("\n[Y] Sì, elimina • [N/Esc] Annulla"))

		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Padding(1, 2).
			Width(50).
			Render(message.String())

		return CenterContent(m.width, m.height, box)
	}

	if m.selectionMode {
		title := RenderHeader("SELEZIONA VEICOLO", width)
		filterView := lipgloss.NewStyle().
			MarginBottom(1).
			Render(m.veicoloFilter.View())

		body := lipgloss.JoinVertical(
			lipgloss.Left,
			filterView,
			m.veicoloTable.View(),
		)

		helpText := HelpStyle.Render("\n[↑↓] Naviga • [↵] Seleziona • [Esc] Annulla")

		content := lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			lipgloss.NewStyle().Padding(0, 2).Render(body),
			helpText,
		)

		box := MainBoxStyle.Copy().Width(width - 4).Render(content)
		return CenterContent(m.width, m.height, box)
	}

	title := "AGENDA & APPUNTAMENTI"
	if m.mode == AgendaAdd {
		title = "NUOVO APPUNTAMENTO"
	} else if m.mode == AgendaEdit {
		title = fmt.Sprintf("MODIFICA APPUNTAMENTO #%d", m.selectedID)
	}

	header := RenderHeader(title, width)
	var body string

	if m.mode == AgendaList {
		todayCount := 0
		today := time.Now().Format("2006-01-02")
		list, _ := m.db.ListAppuntamenti()

		for _, a := range list {
			if a.DataOra.Format("2006-01-02") == today {
				todayCount++
			}
		}

		todayBadge := ""
		if todayCount > 0 {
			todayBadge = WarningBadge(fmt.Sprintf(" %d appuntamenti oggi ", todayCount)) + " "
		}

		helpText := lipgloss.NewStyle().
			MarginBottom(1).
			Foreground(ColorSubText).
			Render(todayBadge + "[N] Nuovo • [E/↵] Modifica • [X/D] Elimina • [ESC] Menu")

		body = lipgloss.JoinVertical(
			lipgloss.Left,
			helpText,
			m.table.View(),
		)
	} else {
		var form strings.Builder
		labels := []string{"Data", "Ora", "Veicolo", "Nota"}

		for i, inp := range m.inputs {
			labelStyle := LabelStyle
			if i == m.focusIndex {
				labelStyle = LabelFocusedStyle
			}

			form.WriteString(fmt.Sprintf("%s %s\n",
				labelStyle.Render(labels[i]+":"),
				inp.View()))
		}

		form.WriteString("\n")
		form.WriteString(HelpStyle.Render("[Tab/↑↓] Naviga • [↵] Conferma • [Esc] Annulla"))
		body = form.String()
	}

	footer := RenderFooter(width)
	if m.err != nil {
		footer = "\n" + ErrorStyle.Render("✗ "+m.err.Error()) + "\n" + footer
	}

	if m.msg != "" {
		footer = "\n" + SuccessStyle.Render(m.msg) + "\n" + footer
	}

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
