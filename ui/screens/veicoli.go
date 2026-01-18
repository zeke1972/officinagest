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
	db                     *database.DB
	table                  table.Model
	inputs                 []textinput.Model
	mode                   VeicoloMode
	focusIndex             int
	selectedID             int
	clienteID              int
	clienteInfo            string
	err                    error
	msg                    string
	width                  int
	height                 int
	selectionMode          bool
	clientTable            table.Model
	clientFilter           textinput.Model
	showConfirm            bool
	deletingID             int
	viewport               viewport.Model
	showOverlay            bool
	deleteWarningCommesse  int
	deleteWarningMovimenti int
	deleteWarningTotale    float64
}

// NewVeicoliModel crea una nuova istanza del model veicoli
func NewVeicoliModel(db *database.DB) VeicoliModel {
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

	clientCols := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Ragione Sociale/N.C.", Width: 35},
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

	rows := []table.Row{}
	for _, item := range viewItems {
		v := item.Veicolo
		prop := "N/D"
		if v.ClienteID > 0 {
			c, err := m.db.GetCliente(v.ClienteID)
			if err == nil && c != nil {
				prop = utils.Truncate(c.RagioneSociale, 20)
			}
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("%d", v.ID),
			strings.ToUpper(v.Targa),
			utils.Truncate(v.Marca+" "+v.Modello, 22),
			item.StatusString,
			prop,
		})
	}

	m.table.SetRows(rows)
}

// countDataForVeicolo conta commesse e movimenti associati a un veicolo
func (m *VeicoliModel) countDataForVeicolo(veicoloID int) (int, int, float64) {
	commesse, _ := m.db.ListCommesse()
	movimenti, _ := m.db.ListMovimenti()

	numCommesse := 0
	commesseIDs := make(map[int]bool)

	for _, c := range commesse {
		if c.VeicoloID == veicoloID {
			numCommesse++
			commesseIDs[c.ID] = true
		}
	}

	numMovimenti := 0
	totaleMov := 0.0
	for _, mov := range movimenti {
		if commesseIDs[mov.CommessaID] {
			numMovimenti++
			if mov.Tipo == "Entrata" {
				totaleMov += mov.Importo
			}
		}
	}

	return numCommesse, numMovimenti, totaleMov
}

// updateClientTable aggiorna la tabella clienti con filtro
func (m *VeicoliModel) updateClientTable() {
	clients, _ := m.db.ListClienti()
	filter := strings.ToUpper(strings.TrimSpace(m.clientFilter.Value()))
	rows := []table.Row{}

	for _, c := range clients {
		ragioneSociale := strings.ToUpper(c.RagioneSociale)
		telefono := strings.ToUpper(c.Telefono)

		if filter == "" ||
			strings.Contains(ragioneSociale, filter) ||
			strings.Contains(telefono, filter) {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", c.ID),
				utils.Truncate(c.RagioneSociale, 35),
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
	m.inputs[0].SetValue(strings.ToUpper(v.Targa))
	m.inputs[1].SetValue(v.Marca)
	m.inputs[2].SetValue(v.Modello)
	m.clienteID = v.ClienteID

	if v.ClienteID > 0 {
		c, err := m.db.GetCliente(v.ClienteID)
		if err == nil && c != nil {
			m.clienteInfo = c.RagioneSociale
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

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].DataApertura.After(filtered[j].DataApertura)
	})

	v, _ := m.db.GetVeicolo(veicoloID)
	var sb strings.Builder

	title := fmt.Sprintf("📋 STORICO INTERVENTI: %s %s (%s)",
		v.Marca, v.Modello, strings.ToUpper(v.Targa))

	sb.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Render(title) + "\n\n")

	if len(filtered) == 0 {
		sb.WriteString(HelpStyle.Render("Nessuna commessa registrata per questo veicolo."))
	} else {
		for _, c := range filtered {
			dateStr := utils.FormatDate(c.DataApertura)

			status := "🔴 APERTA"
			stStyle := ErrorStyle
			if c.Stato == "Chiusa" {
				status = "🟢 CHIUSA"
				stStyle = SuccessStyle
			}

			versato := 0.0
			for _, mov := range movimenti {
				if mov.CommessaID == c.ID && mov.Tipo == "Entrata" {
					versato += mov.Importo
				}
			}

			residuo := c.Totale - versato

			sb.WriteString(lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorHighlight).
				Render(fmt.Sprintf("Commessa #%s", c.Numero)) + " ")
			sb.WriteString(stStyle.Render(status) + "\n")
			sb.WriteString(fmt.Sprintf("📅 Data: %s\n", dateStr))
			sb.WriteString(fmt.Sprintf("💰 Totale: %s\n", utils.FormatEuro(c.Totale)))
			sb.WriteString(fmt.Sprintf("💵 Versato: %s | Residuo: %s\n",
				utils.FormatEuro(versato), utils.FormatEuro(residuo)))

			sb.WriteString("🔧 Lavori:\n")
			lavori := strings.Split(c.LavoriEseguiti, ",")
			for _, lavoro := range lavori {
				lavoro = strings.TrimSpace(lavoro)
				if lavoro != "" {
					sb.WriteString(fmt.Sprintf("  • %s\n", lavoro))
				}
			}

			if c.Note != "" {
				sb.WriteString(fmt.Sprintf("📝 Note: %s\n", c.Note))
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

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

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
					m.clienteInfo = row[1]
					m.inputs[3].SetValue(m.clienteInfo)
					m.selectionMode = false
					m.focusIndex = 0
					m.updateFocus()
				}
				return m, nil
			}
		}

		var cmdF, cmdT tea.Cmd
		m.clientFilter, cmdF = m.clientFilter.Update(msg)
		m.updateClientTable()
		m.clientTable, cmdT = m.clientTable.Update(msg)
		return m, tea.Batch(cmdF, cmdT)
	}

	if m.showConfirm {
		if k, ok := msg.(tea.KeyMsg); ok {
			switch k.String() {
			case "y", "Y":
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
					m.msg = fmt.Sprintf("✓ Veicolo, %d commesse e relativi movimenti eliminati", count)
				}

				m.Refresh()
				m.showConfirm = false
				m.deleteWarningCommesse = 0
				m.deleteWarningMovimenti = 0
				m.deleteWarningTotale = 0
			case "n", "N", "esc":
				m.showConfirm = false
				m.deleteWarningCommesse = 0
				m.deleteWarningMovimenti = 0
				m.deleteWarningTotale = 0
			}
			return m, nil
		}
	}

	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		if m.mode != ModeList {
			m.mode = ModeList
			m.err = nil
			m.msg = ""
			return m, nil
		}
		return m, func() tea.Msg { return ChangeScreenMsg(StateMenu) }
	}

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
					numCommesse, numMovimenti, totaleMovimenti := m.countDataForVeicolo(id)
					m.deleteWarningCommesse = numCommesse
					m.deleteWarningMovimenti = numMovimenti
					m.deleteWarningTotale = totaleMovimenti
					m.showConfirm = true
				}
				return m, nil
			}
		}
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	if m.mode == ModeAdd || m.mode == ModeEdit {
		if k, ok := msg.(tea.KeyMsg); ok {
			if m.focusIndex == 3 && (k.String() == "enter" || k.String() == " ") {
				m.selectionMode = true
				m.clientFilter.SetValue("")
				m.updateClientTable()
				m.clientFilter.Focus()
				return m, nil
			}

			switch k.String() {
			case "enter":
				if m.focusIndex != 3 {
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

		if m.focusIndex == 0 {
			m.inputs[0].SetValue(strings.ToUpper(m.inputs[0].Value()))
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
func (m VeicoliModel) View() string {
	width := 90
	if m.width > 0 {
		width = min(m.width, 100)
	}

	if m.showOverlay {
		return CenterContent(m.width, m.height, m.viewport.View())
	}

	if m.showConfirm {
		var message strings.Builder
		v, _ := m.db.GetVeicolo(m.deletingID)
		targa := "???"
		if v != nil {
			targa = fmt.Sprintf("%s %s (%s)", v.Marca, v.Modello, v.Targa)
		}

		message.WriteString(fmt.Sprintf("⚠️  ELIMINAZIONE VEICOLO #%d\n", m.deletingID))
		message.WriteString(fmt.Sprintf("%s\n\n", targa))

		if m.deleteWarningCommesse > 0 {
			message.WriteString(ErrorStyle.Render(fmt.Sprintf(
				"ATTENZIONE: Questo veicolo ha dati associati!\n\n"+
					"Eliminando il veicolo verranno eliminati:\n"+
					" • %d commesse\n"+
					" • %d movimenti di Prima Nota (totale: %s)\n\n"+
					"TUTTI I DATI VERRANNO PERSI IN MODO PERMANENTE!\n\n",
				m.deleteWarningCommesse,
				m.deleteWarningMovimenti,
				utils.FormatEuro(m.deleteWarningTotale),
			)))
		} else {
			message.WriteString("Questo veicolo non ha commesse o movimenti associati.\n\n")
		}

		message.WriteString(WarningStyle.Render("Sei sicuro di voler procedere?\n"))
		message.WriteString(HelpStyle.Render("\n[Y] Sì, elimina TUTTO • [N/Esc] Annulla"))

		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Padding(1, 2).
			Width(75).
			Render(message.String())

		return CenterContent(m.width, m.height, box)
	}

	if m.selectionMode {
		title := RenderHeader("SELEZIONA PROPRIETARIO", width)
		filterView := lipgloss.NewStyle().
			MarginBottom(1).
			Render(m.clientFilter.View())

		body := lipgloss.JoinVertical(
			lipgloss.Left,
			filterView,
			m.clientTable.View(),
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

	title := "GESTIONE VEICOLI"
	if m.mode == ModeAdd {
		title = "NUOVO VEICOLO"
	} else if m.mode == ModeEdit {
		title = fmt.Sprintf("MODIFICA VEICOLO #%d", m.selectedID)
	}

	header := RenderHeader(title, width)
	var body string

	if m.mode == ModeList {
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
		var form strings.Builder
		labels := []string{"Targa", "Marca", "Modello", "Proprietario"}

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
