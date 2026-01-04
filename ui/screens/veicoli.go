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
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// VeicoloMode rappresenta la modalità corrente
type VeicoloMode int

const (
	ModeList VeicoloMode = iota
	ModeAdd
	ModeEdit
)

// VeicoloViewItem contiene i dati di visualizzazione di un veicolo
type VeicoloViewItem struct {
	Veicolo      database.Veicolo
	LastOpenDate time.Time
	HasOpen      bool
	StatusString string
}

// VeicoliModel gestisce la schermata veicoli
type VeicoliModel struct {
	db            *database.DB
	table         table.Model
	inputs        []textinput.Model
	mode          VeicoloMode
	focusIndex    int
	selectedID    int
	clienteID     int
	clienteInfo   string
	err           error
	msg           string
	width         int
	height        int
	selectionMode bool
	clientTable   table.Model
	clientFilter  textinput.Model
	showConfirm   bool
	deletingID    int
	viewport      viewport.Model
	showOverlay   bool
}

// NewVeicoliModel crea una nuova istanza del model veicoli
func NewVeicoliModel(db *database.DB) VeicoliModel {
	// Tabella principale veicoli
	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Targa", Width: 10},
		{Title: "Modello", Width: 22},
		{Title: "Stato Commessa", Width: 28},
		{Title: "Proprietario", Width: 20},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(12),
	)
	t.SetStyles(GetTableStyles())

	// Input form
	inputs := make([]textinput.Model, 4)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Targa (es. AB123CD)"
	inputs[0].CharLimit = 10
	inputs[0].Width = 40

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Marca (es. Fiat)"
	inputs[1].Width = 40

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Modello (es. Panda 1.2)"
	inputs[2].Width = 40

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "[ INVIO PER SCEGLIERE PROPRIETARIO ]"
	inputs[3].Width = 40

	// Tabella selezione cliente
	clientCols := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Cognome", Width: 20},
		{Title: "Nome", Width: 20},
		{Title: "Telefono", Width: 14},
	}

	ct := table.New(
		table.WithColumns(clientCols),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	ct.SetStyles(GetTableStyles())

	cf := textinput.New()
	cf.Placeholder = "🔍 Cerca proprietario..."
	cf.Width = 50

	// Viewport per storico
	vp := viewport.New(70, 20)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(1, 2)

	m := VeicoliModel{
		db:           db,
		table:        t,
		inputs:       inputs,
		mode:         ModeList,
		viewport:     vp,
		clientTable:  ct,
		clientFilter: cf,
	}

	m.Refresh()
	return m
}

// Refresh aggiorna la lista dei veicoli con priorità commesse aperte
func (m *VeicoliModel) Refresh() {
	list, _ := m.db.ListVeicoli()
	commesse, _ := m.db.ListCommesse()

	var viewItems []VeicoloViewItem

	for _, v := range list {
		var lastOpen time.Time
		hasOpen := false
		statusStr := "—"

		// Cerca commesse aperte per questo veicolo
		for _, c := range commesse {
			if c.VeicoloID == v.ID && c.Stato == "Aperta" {
				if c.DataApertura.After(lastOpen) {
					lastOpen = c.DataApertura
					hasOpen = true
				}
			}
		}

		if hasOpen {
			statusStr = "🔴 APERTA: " + utils.FormatDate(lastOpen)
		}

		viewItems = append(viewItems, VeicoloViewItem{
			Veicolo:      v,
			LastOpenDate: lastOpen,
			HasOpen:      hasOpen,
			StatusString: statusStr,
		})
	}

	// Ordina: prima veicoli con commesse aperte (per data recente), poi altri
	sort.Slice(viewItems, func(i, j int) bool {
		if viewItems[i].HasOpen && viewItems[j].HasOpen {
			return viewItems[i].LastOpenDate.After(viewItems[j].LastOpenDate)
		}
		if viewItems[i].HasOpen && !viewItems[j].HasOpen {
			return true
		}
		if !viewItems[i].HasOpen && viewItems[j].HasOpen {
			return false
		}
		return viewItems[i].Veicolo.ID > viewItems[j].Veicolo.ID
	})

	// Genera righe tabella
	rows := []table.Row{}
	for _, item := range viewItems {
		v := item.Veicolo
		prop := "N/D"

		if v.ClienteID > 0 {
			c, err := m.db.GetCliente(v.ClienteID)
			if err == nil && c != nil {
				prop = utils.Truncate(fmt.Sprintf("%s %s", c.Cognome, c.Nome), 20)
			}
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("%d", v.ID),
			v.Targa,
			utils.Truncate(v.Marca+" "+v.Modello, 22),
			item.StatusString,
			prop,
		})
	}

	m.table.SetRows(rows)
}

// updateClientTable aggiorna la tabella clienti con filtro
func (m *VeicoliModel) updateClientTable() {
	clients, _ := m.db.ListClienti()
	filter := strings.ToUpper(strings.TrimSpace(m.clientFilter.Value()))
	rows := []table.Row{}

	for _, c := range clients {
		cognome := strings.ToUpper(c.Cognome)
		nome := strings.ToUpper(c.Nome)
		telefono := strings.ToUpper(c.Telefono)

		if filter == "" ||
			strings.Contains(cognome, filter) ||
			strings.Contains(nome, filter) ||
			strings.Contains(telefono, filter) {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", c.ID),
				utils.Truncate(c.Cognome, 20),
				utils.Truncate(c.Nome, 20),
				c.Telefono,
			})
		}
	}

	m.clientTable.SetRows(rows)
}

// resetForm resetta il form
func (m *VeicoliModel) resetForm() {
	for i := range m.inputs {
		m.inputs[i].SetValue("")
	}

	m.clienteID = 0
	m.clienteInfo = "Nessun proprietario"
	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.inputs[0].Focus()
}

// loadIntoForm carica un veicolo nel form
func (m *VeicoliModel) loadIntoForm(id int) {
	v, err := m.db.GetVeicolo(id)
	if err != nil {
		m.err = fmt.Errorf("errore caricamento veicolo: %w", err)
		return
	}

	m.selectedID = id
	m.inputs[0].SetValue(v.Targa)
	m.inputs[1].SetValue(v.Marca)
	m.inputs[2].SetValue(v.Modello)
	m.clienteID = v.ClienteID

	if v.ClienteID > 0 {
		c, err := m.db.GetCliente(v.ClienteID)
		if err == nil && c != nil {
			m.clienteInfo = fmt.Sprintf("%s %s", c.Cognome, c.Nome)
			m.inputs[3].SetValue(m.clienteInfo)
		}
	} else {
		m.clienteInfo = "Nessun proprietario"
		m.inputs[3].SetValue("")
	}

	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.inputs[0].Focus()
}

// loadHistory carica lo storico commesse di un veicolo
func (m *VeicoliModel) loadHistory(veicoloID int) {
	commesse, _ := m.db.ListCommesse()
	movimenti, _ := m.db.ListMovimenti()

	var filtered []database.Commessa
	for _, c := range commesse {
		if c.VeicoloID == veicoloID {
			filtered = append(filtered, c)
		}
	}

	// Ordina per data decrescente
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].DataApertura.After(filtered[j].DataApertura)
	})

	v, _ := m.db.GetVeicolo(veicoloID)

	var sb strings.Builder

	// Titolo
	title := fmt.Sprintf("📋 STORICO INTERVENTI: %s %s (%s)",
		v.Marca, v.Modello, v.Targa)
	sb.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Render(title) + "\n\n")

	if len(filtered) == 0 {
		sb.WriteString(HelpStyle.Render("Nessuna commessa registrata per questo veicolo."))
	} else {
		for _, c := range filtered {
			dateStr := utils.FormatDate(c.DataApertura)

			// Stato
			status := "🔴 APERTA"
			stStyle := ErrorStyle
			if c.Stato == "Chiusa" {
				status = "🟢 CHIUSA"
				stStyle = SuccessStyle
			}

			// Calcola versato
			versato := 0.0
			for _, mov := range movimenti {
				if mov.CommessaID == c.ID && mov.Tipo == "Entrata" {
					versato += mov.Importo
				}
			}
			residuo := c.Totale - versato

			// Rendering commessa
			sb.WriteString(lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorHighlight).
				Render(fmt.Sprintf("Commessa #%s", c.Numero)) + " ")
			sb.WriteString(stStyle.Render(status) + "\n")
			sb.WriteString(fmt.Sprintf("📅 Data: %s\n", dateStr))
			sb.WriteString(fmt.Sprintf("💰 Totale: %s\n", utils.FormatEuro(c.Totale)))
			sb.WriteString(fmt.Sprintf("💵 Versato: %s | Residuo: %s\n",
				utils.FormatEuro(versato), utils.FormatEuro(residuo)))
			sb.WriteString(fmt.Sprintf("🔧 Lavori: %s\n", utils.Truncate(c.LavoriEseguiti, 60)))

			if c.Note != "" {
				sb.WriteString(fmt.Sprintf("📝 Note: %s\n", utils.Truncate(c.Note, 60)))
			}

			sb.WriteString(lipgloss.NewStyle().
				Foreground(ColorBorder).
				Render(strings.Repeat("─", 70)) + "\n\n")
		}
	}

	m.viewport.SetContent(sb.String())
	m.viewport.GotoTop()
	m.showOverlay = true
}

// updateFocus aggiorna il focus tra i campi
func (m *VeicoliModel) updateFocus() {
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

// validate valida i dati del form
func (m *VeicoliModel) validate() error {
	if err := utils.ValidateTarga(m.inputs[0].Value()); err != nil {
		return err
	}

	if err := utils.ValidateNotEmpty(m.inputs[1].Value(), "Marca"); err != nil {
		return err
	}

	if err := utils.ValidateNotEmpty(m.inputs[2].Value(), "Modello"); err != nil {
		return err
	}

	if m.clienteID == 0 {
		return fmt.Errorf("seleziona un proprietario")
	}

	return nil
}

// save salva il veicolo corrente
func (m *VeicoliModel) save() error {
	if err := m.validate(); err != nil {
		return err
	}

	v := &database.Veicolo{
		Targa:     strings.ToUpper(m.inputs[0].Value()),
		Marca:     m.inputs[1].Value(),
		Modello:   m.inputs[2].Value(),
		ClienteID: m.clienteID,
	}

	if m.mode == ModeAdd {
		if err := m.db.CreateVeicolo(v); err != nil {
			return fmt.Errorf("errore creazione: %w", err)
		}
		m.msg = "✓ Veicolo creato con successo"
	} else {
		v.ID = m.selectedID
		if err := m.db.UpdateVeicolo(v); err != nil {
			return fmt.Errorf("errore aggiornamento: %w", err)
		}
		m.msg = "✓ Veicolo aggiornato con successo"
	}

	m.mode = ModeList
	m.Refresh()
	return nil
}

// Init implementa tea.Model
func (m VeicoliModel) Init() tea.Cmd {
	return nil
}

// Update implementa tea.Model
func (m VeicoliModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Gestione resize
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	// Overlay storico
	if m.showOverlay {
		if k, ok := msg.(tea.KeyMsg); ok {
			if k.String() == "esc" || k.String() == "h" || k.String() == "q" {
				m.showOverlay = false
				return m, nil
			}
		}
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	// Popup selezione cliente
	if m.selectionMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.selectionMode = false
				m.updateFocus()
				return m, nil

			case "enter":
				if row := m.clientTable.SelectedRow(); len(row) > 0 {
					m.clienteID, _ = strconv.Atoi(row[0])
					m.clienteInfo = fmt.Sprintf("%s %s", row[1], row[2])
					m.inputs[3].SetValue(m.clienteInfo)
					m.selectionMode = false
					m.focusIndex = 0
					m.updateFocus()
				}
				return m, nil
			}

			var cmdF, cmdT tea.Cmd
			m.clientFilter, cmdF = m.clientFilter.Update(msg)
			m.updateClientTable()
			m.clientTable, cmdT = m.clientTable.Update(msg)
			return m, tea.Batch(cmdF, cmdT)
		}
	}

	// Conferma eliminazione
	if m.showConfirm {
		if k, ok := msg.(tea.KeyMsg); ok {
			switch k.String() {
			case "y", "Y":
				// Elimina anche le commesse associate
				commesse, _ := m.db.ListCommesse()
				count := 0
				for _, c := range commesse {
					if c.VeicoloID == m.deletingID {
						m.db.DeleteCommessa(c.ID)
						count++
					}
				}

				if err := m.db.DeleteVeicolo(m.deletingID); err != nil {
					m.err = fmt.Errorf("errore eliminazione: %w", err)
				} else {
					m.msg = fmt.Sprintf("✓ Veicolo e %d commesse eliminati", count)
				}
				m.Refresh()
				m.showConfirm = false

			case "n", "N", "esc":
				m.showConfirm = false
			}
			return m, nil
		}
	}

	// Gestione ESC
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		if m.mode != ModeList {
			m.mode = ModeList
			m.err = nil
			m.msg = ""
			return m, nil
		}
		return m, func() tea.Msg { return ChangeScreenMsg(StateMenu) }
	}

	// Modalità Lista
	if m.mode == ModeList {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "n":
				m.mode = ModeAdd
				m.resetForm()
				return m, nil

			case "e", "enter":
				if row := m.table.SelectedRow(); len(row) > 0 {
					id, _ := strconv.Atoi(row[0])
					m.loadIntoForm(id)
					m.mode = ModeEdit
				}
				return m, nil

			case "h":
				if row := m.table.SelectedRow(); len(row) > 0 {
					id, _ := strconv.Atoi(row[0])
					m.loadHistory(id)
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

			m.table, cmd = m.table.Update(msg)
			return m, cmd
		}
	}

	// Modalità Form (Add/Edit)
	if m.mode == ModeAdd || m.mode == ModeEdit {
		if k, ok := msg.(tea.KeyMsg); ok {
			// Campo cliente: apre selezione
			if m.focusIndex == 3 && (k.String() == "enter" || k.String() == " ") {
				m.selectionMode = true
				m.clientFilter.SetValue("")
				m.updateClientTable()
				m.clientFilter.Focus()
				return m, nil
			}

			switch k.String() {
			case "enter":
				// Se non siamo sul campo cliente, salva
				if m.focusIndex != 3 {
					if err := m.save(); err != nil {
						m.err = err
						return m, nil
					}
					return m, nil
				}

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

			// Auto-uppercase per targa
			if m.focusIndex == 0 {
				m.inputs[0].SetValue(strings.ToUpper(m.inputs[0].Value()))
			}
		}

		// Update inputs (escluso campo cliente readonly)
		cmds := make([]tea.Cmd, 4)
		for i := 0; i < 3; i++ {
			m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
		}
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

// View implementa tea.Model
func (m VeicoliModel) View() string {
	width := 90
	if m.width > 0 {
		width = min(m.width, 110)
	}

	// Overlay storico
	if m.showOverlay {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			m.viewport.View(),
			"",
			HelpStyle.Render("[Esc/H/Q] Chiudi • [↑↓] Scorri"),
		)

		box := MainBoxStyle.Copy().Width(min(width-4, 80)).Render(content)

		if m.width > 0 && m.height > 0 {
			return CenterContent(m.width, m.height, box)
		}
		return box
	}

	// Popup selezione cliente
	if m.selectionMode {
		title := TitleStyle.Render("🔍 SELEZIONA PROPRIETARIO")
		filter := fmt.Sprintf("Cerca: %s", m.clientFilter.View())

		inner := lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			filter,
			"",
			m.clientTable.View(),
			"",
			HelpStyle.Render("[↵] Seleziona • [Esc] Annulla"),
		)

		box := MainBoxStyle.Copy().
			Width(min(width-10, 70)).
			Render(inner)

		if m.width > 0 && m.height > 0 {
			return CenterContent(m.width, m.height, box)
		}
		return box
	}

	// Titolo dinamico
	title := "GESTIONE VEICOLI"
	if m.mode == ModeAdd {
		title = "NUOVO VEICOLO"
	} else if m.mode == ModeEdit {
		title = fmt.Sprintf("MODIFICA VEICOLO #%d", m.selectedID)
	}

	header := RenderHeader(title, width)

	var body string

	// Dialog conferma eliminazione
	if m.showConfirm {
		body = RenderConfirmDialog(
			fmt.Sprintf("Eliminare il veicolo #%d e tutte le sue commesse?", m.deletingID),
			width,
			0,
		)
	} else if m.mode == ModeList {
		// Vista lista
		helpText := lipgloss.NewStyle().
			MarginBottom(1).
			Foreground(ColorSubText).
			Render("[N] Nuovo • [E/↵] Modifica • [H] Storico • [X/D] Elimina • [ESC] Menu")

		body = lipgloss.JoinVertical(
			lipgloss.Left,
			helpText,
			m.table.View(),
		)
	} else {
		// Vista form
		var form strings.Builder
		labels := []string{"Targa", "Marca", "Modello", "Proprietario"}

		for i := 0; i < 4; i++ {
			labelStyle := LabelStyle
			if i == m.focusIndex {
				labelStyle = LabelFocusedStyle
			}

			view := m.inputs[i].View()

			// Campo proprietario (readonly con highlight)
			if i == 3 {
				if m.focusIndex == 3 {
					if m.clienteInfo == "" || m.clienteInfo == "Nessun proprietario" {
						view = lipgloss.NewStyle().
							Foreground(ColorPrimary).
							Render("[ INVIO PER SCEGLIERE ]")
					} else {
						view = lipgloss.NewStyle().
							Foreground(ColorPrimary).
							Bold(true).
							Render(m.clienteInfo)
					}
				} else {
					if m.clienteInfo == "" || m.clienteInfo == "Nessun proprietario" {
						view = HelpStyle.Render("[ Nessuno ]")
					} else {
						view = m.clienteInfo
					}
				}
			}

			form.WriteString(fmt.Sprintf("%s %s\n",
				labelStyle.Render(labels[i]+":"),
				view))
		}

		form.WriteString("\n")
		form.WriteString(HelpStyle.Render("[Tab/↑↓] Naviga • [↵] Salva • [Esc] Annulla"))

		body = form.String()
	}

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
