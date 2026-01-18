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

// CommessaMode rappresenta la modalità corrente
type CommessaMode int

const (
	CommList CommessaMode = iota
	CommAdd
	CommEdit
	CommDetail
)

// CommesseModel gestisce la schermata commesse
type CommesseModel struct {
	db               *database.DB
	table            table.Model
	inputs           []textinput.Model
	mode             CommessaMode
	focusIndex       int
	selectedID       int
	veicoloID        int
	veicoloInfo      string
	err              error
	msg              string
	width            int
	height           int
	selectionMode    bool
	veicoloTable     table.Model
	veicoloFilter    textinput.Model
	showConfirm      bool
	deletingID       int
	viewport         viewport.Model
	showOverlay      bool
	deleteWarningMov int
	deleteWarningTot float64
}

// CommessaViewItem contiene i dati di visualizzazione di una commessa
type CommessaViewItem struct {
	Commessa database.Commessa
	Versato  float64
	Residuo  float64
}

// NewCommesseModel crea una nuova istanza del model commesse
func NewCommesseModel(db *database.DB) CommesseModel {
	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Numero", Width: 10},
		{Title: "Veicolo", Width: 15},
		{Title: "Data", Width: 12},
		{Title: "Stato", Width: 8},
		{Title: "Totale", Width: 12},
		{Title: "Versato", Width: 12},
		{Title: "Residuo", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(12),
	)

	t.SetStyles(GetTableStyles())

	inputs := make([]textinput.Model, 5)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "[ INVIO PER SCEGLIERE VEICOLO ]"
	inputs[0].Width = 40

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Lavori (separa con virgola)"
	inputs[1].Width = 60

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Costo Manodopera (es. 150.00)"
	inputs[2].Width = 30

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Costo Ricambi (es. 200.00)"
	inputs[3].Width = 30

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "Note (opzionale)"
	inputs[4].Width = 60

	veicoloCols := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Targa", Width: 10},
		{Title: "Modello", Width: 22},
		{Title: "Proprietario", Width: 25},
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

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(1, 2)

	m := CommesseModel{
		db:            db,
		table:         t,
		inputs:        inputs,
		mode:          CommList,
		veicoloTable:  vt,
		veicoloFilter: vf,
		viewport:      vp,
	}

	m.Refresh()
	return m
}

// Refresh aggiorna la lista delle commesse
func (m *CommesseModel) Refresh() {
	commesse, _ := m.db.ListCommesse()
	movimenti, _ := m.db.ListMovimenti()

	accontiMap := make(map[int]float64)
	for _, mov := range movimenti {
		if mov.CommessaID > 0 && mov.Tipo == "Entrata" {
			accontiMap[mov.CommessaID] += mov.Importo
		}
	}

	var viewItems []CommessaViewItem
	for _, c := range commesse {
		versato := accontiMap[c.ID]
		residuo := c.Totale - versato
		if residuo < 0 {
			residuo = 0
		}

		viewItems = append(viewItems, CommessaViewItem{
			Commessa: c,
			Versato:  versato,
			Residuo:  residuo,
		})
	}

	sort.Slice(viewItems, func(i, j int) bool {
		if viewItems[i].Commessa.Stato == "Aperta" && viewItems[j].Commessa.Stato == "Chiusa" {
			return true
		}
		if viewItems[i].Commessa.Stato == "Chiusa" && viewItems[j].Commessa.Stato == "Aperta" {
			return false
		}
		return viewItems[i].Commessa.DataApertura.After(viewItems[j].Commessa.DataApertura)
	})

	rows := []table.Row{}
	for _, item := range viewItems {
		c := item.Commessa
		v, _ := m.db.GetVeicolo(c.VeicoloID)
		veicoloInfo := "N/D"
		if v != nil {
			veicoloInfo = v.Targa
		}

		stato := "🔴 Aperta"
		if c.Stato == "Chiusa" {
			stato = "🟢 Chiusa"
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("%d", c.ID),
			c.Numero,
			veicoloInfo,
			utils.FormatDate(c.DataApertura),
			stato,
			utils.FormatEuro(c.Totale),
			utils.FormatEuro(item.Versato),
			utils.FormatEuro(item.Residuo),
		})
	}

	m.table.SetRows(rows)
}

// updateVeicoloTable aggiorna la tabella veicoli con filtro
func (m *CommesseModel) updateVeicoloTable() {
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
				utils.Truncate(proprietario, 25),
			})
		}
	}

	m.veicoloTable.SetRows(rows)
}

// countMovimentiForCommessa conta i movimenti associati a una commessa
func (m *CommesseModel) countMovimentiForCommessa(commessaID int) (int, float64) {
	movimenti, _ := m.db.ListMovimenti()
	numMovimenti := 0
	totale := 0.0

	for _, mov := range movimenti {
		if mov.CommessaID == commessaID {
			numMovimenti++
			if mov.Tipo == "Entrata" {
				totale += mov.Importo
			}
		}
	}

	return numMovimenti, totale
}

// loadDetail carica i dettagli di una commessa
func (m *CommesseModel) loadDetail(id int) {
	comm, err := m.db.GetCommessa(id)
	if err != nil {
		m.err = fmt.Errorf("errore caricamento commessa: %w", err)
		return
	}

	v, _ := m.db.GetVeicolo(comm.VeicoloID)
	movimenti, _ := m.db.ListMovimenti()

	var sb strings.Builder

	title := fmt.Sprintf("📋 DETTAGLIO COMMESSA #%s", comm.Numero)
	sb.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Render(title) + "\n\n")

	stato := "🔴 APERTA"
	stStyle := ErrorStyle
	if comm.Stato == "Chiusa" {
		stato = "🟢 CHIUSA"
		stStyle = SuccessStyle
	}

	sb.WriteString(stStyle.Render(stato) + "\n\n")

	sb.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorHighlight).
		Render("VEICOLO") + "\n")

	if v != nil {
		sb.WriteString(fmt.Sprintf("🚗 Targa: %s\n", strings.ToUpper(v.Targa)))
		sb.WriteString(fmt.Sprintf("   Marca/Modello: %s %s\n", v.Marca, v.Modello))

		if v.ClienteID > 0 {
			cli, _ := m.db.GetCliente(v.ClienteID)
			if cli != nil {
				sb.WriteString(fmt.Sprintf("   Proprietario: %s\n", cli.RagioneSociale))
			}
		}
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorHighlight).
		Render("DATE") + "\n")
	sb.WriteString(fmt.Sprintf("📅 Apertura: %s\n", utils.FormatDate(comm.DataApertura)))
	if !comm.DataChiusura.IsZero() {
		sb.WriteString(fmt.Sprintf("📅 Chiusura: %s\n", utils.FormatDate(comm.DataChiusura)))
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorHighlight).
		Render("LAVORI ESEGUITI") + "\n")

	lavori := strings.Split(comm.LavoriEseguiti, ",")
	for _, lavoro := range lavori {
		lavoro = strings.TrimSpace(lavoro)
		if lavoro != "" {
			sb.WriteString(fmt.Sprintf("🔧 %s\n", lavoro))
		}
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorHighlight).
		Render("COSTI") + "\n")
	sb.WriteString(fmt.Sprintf("💼 Manodopera: %s\n", utils.FormatEuro(comm.CostoManodopera)))
	sb.WriteString(fmt.Sprintf("🔩 Ricambi: %s\n", utils.FormatEuro(comm.CostoRicambi)))
	sb.WriteString(fmt.Sprintf("💰 Totale: %s\n", utils.FormatEuro(comm.Totale)))

	versato := 0.0
	var pagamenti []database.MovimentoPrimaNota
	for _, mov := range movimenti {
		if mov.CommessaID == comm.ID && mov.Tipo == "Entrata" {
			versato += mov.Importo
			pagamenti = append(pagamenti, mov)
		}
	}

	residuo := comm.Totale - versato
	if residuo < 0 {
		residuo = 0
	}

	sb.WriteString(fmt.Sprintf("💵 Versato: %s\n", utils.FormatEuro(versato)))
	sb.WriteString(fmt.Sprintf("📊 Residuo: %s\n", utils.FormatEuro(residuo)))

	if len(pagamenti) > 0 {
		sb.WriteString("\n")
		sb.WriteString(lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorHighlight).
			Render("PAGAMENTI") + "\n")

		sort.Slice(pagamenti, func(i, j int) bool {
			return pagamenti[i].Data.Before(pagamenti[j].Data)
		})

		for _, pag := range pagamenti {
			sb.WriteString(fmt.Sprintf("💳 %s - %s - %s\n",
				utils.FormatDate(pag.Data),
				utils.FormatEuro(pag.Importo),
				pag.Metodo))
		}
	}

	if comm.Note != "" {
		sb.WriteString("\n")
		sb.WriteString(lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorHighlight).
			Render("NOTE") + "\n")
		sb.WriteString(fmt.Sprintf("📝 %s\n", comm.Note))
	}

	m.viewport.SetContent(sb.String())
	m.viewport.GotoTop()
	m.showOverlay = true
}

// resetForm resetta il form
func (m *CommesseModel) resetForm() {
	for i := range m.inputs {
		m.inputs[i].SetValue("")
	}

	m.inputs[2].SetValue("0.00")
	m.inputs[3].SetValue("0.00")
	m.veicoloID = 0
	m.veicoloInfo = "Nessun veicolo"
	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.inputs[0].Focus()
}

// loadIntoForm carica una commessa nel form
func (m *CommesseModel) loadIntoForm(id int) {
	c, err := m.db.GetCommessa(id)
	if err != nil {
		m.err = fmt.Errorf("errore caricamento commessa: %w", err)
		return
	}

	m.selectedID = id
	m.veicoloID = c.VeicoloID

	if c.VeicoloID > 0 {
		v, err := m.db.GetVeicolo(c.VeicoloID)
		if err == nil && v != nil {
			m.veicoloInfo = fmt.Sprintf("%s %s (%s)", v.Marca, v.Modello, v.Targa)
			m.inputs[0].SetValue(m.veicoloInfo)
		}
	} else {
		m.veicoloInfo = "Nessun veicolo"
		m.inputs[0].SetValue("")
	}

	m.inputs[1].SetValue(c.LavoriEseguiti)
	m.inputs[2].SetValue(fmt.Sprintf("%.2f", c.CostoManodopera))
	m.inputs[3].SetValue(fmt.Sprintf("%.2f", c.CostoRicambi))
	m.inputs[4].SetValue(c.Note)

	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.inputs[0].Focus()
}

// updateFocus aggiorna il focus tra i campi
func (m *CommesseModel) updateFocus() {
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

// validate valida i dati del form
func (m *CommesseModel) validate() error {
	if m.veicoloID == 0 {
		return fmt.Errorf("seleziona un veicolo")
	}

	if err := utils.ValidateNotEmpty(m.inputs[1].Value(), "Lavori eseguiti"); err != nil {
		return err
	}

	manodopera, err := utils.ParseFloat(m.inputs[2].Value())
	if err != nil {
		return fmt.Errorf("costo manodopera non valido")
	}

	if manodopera < 0 {
		return fmt.Errorf("costo manodopera non può essere negativo")
	}

	ricambi, err := utils.ParseFloat(m.inputs[3].Value())
	if err != nil {
		return fmt.Errorf("costo ricambi non valido")
	}

	if ricambi < 0 {
		return fmt.Errorf("costo ricambi non può essere negativo")
	}

	return nil
}

// save salva la commessa corrente
func (m *CommesseModel) save() error {
	if err := m.validate(); err != nil {
		return err
	}

	manodopera, _ := utils.ParseFloat(m.inputs[2].Value())
	ricambi, _ := utils.ParseFloat(m.inputs[3].Value())

	c := &database.Commessa{
		VeicoloID:       m.veicoloID,
		LavoriEseguiti:  m.inputs[1].Value(),
		CostoManodopera: manodopera,
		CostoRicambi:    ricambi,
		Note:            m.inputs[4].Value(),
		Stato:           "Aperta",
	}

	if m.mode == CommAdd {
		if err := m.db.CreateCommessa(c); err != nil {
			return fmt.Errorf("errore creazione: %w", err)
		}
		m.msg = "✓ Commessa creata con successo"
	} else {
		c.ID = m.selectedID
		oldComm, _ := m.db.GetCommessa(m.selectedID)
		if oldComm != nil {
			c.Stato = oldComm.Stato
			c.DataApertura = oldComm.DataApertura
			c.DataChiusura = oldComm.DataChiusura
		}

		if err := m.db.UpdateCommessa(c); err != nil {
			return fmt.Errorf("errore aggiornamento: %w", err)
		}
		m.msg = "✓ Commessa aggiornata con successo"
	}

	m.mode = CommList
	m.Refresh()
	return nil
}

// toggleStato cambia lo stato di una commessa
func (m *CommesseModel) toggleStato(id int) error {
	c, err := m.db.GetCommessa(id)
	if err != nil {
		return err
	}

	if c.Stato == "Aperta" {
		c.Stato = "Chiusa"
		c.DataChiusura = time.Now()
	} else {
		c.Stato = "Aperta"
		c.DataChiusura = time.Time{}
	}

	if err := m.db.UpdateCommessa(c); err != nil {
		return err
	}

	m.msg = fmt.Sprintf("✓ Commessa %s", c.Stato)
	m.Refresh()
	return nil
}

// Init implementa tea.Model
func (m CommesseModel) Init() tea.Cmd {
	return nil
}

// Update implementa tea.Model
func (m CommesseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	if m.showOverlay {
		if k, ok := msg.(tea.KeyMsg); ok {
			if k.String() == "esc" || k.String() == "d" || k.String() == "q" {
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
				if row := m.veicoloTable.SelectedRow(); len(row) > 0 {
					m.veicoloID, _ = strconv.Atoi(row[0])
					m.veicoloInfo = fmt.Sprintf("%s (%s)", row[2], row[1])
					m.inputs[0].SetValue(m.veicoloInfo)
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
				if err := m.db.DeleteCommessa(m.deletingID); err != nil {
					m.err = fmt.Errorf("errore eliminazione: %w", err)
				} else {
					m.msg = fmt.Sprintf("✓ Commessa e %d movimenti eliminati", m.deleteWarningMov)
				}
				m.Refresh()
				m.showConfirm = false
				m.deleteWarningMov = 0
				m.deleteWarningTot = 0
			case "n", "N", "esc":
				m.showConfirm = false
				m.deleteWarningMov = 0
				m.deleteWarningTot = 0
			}
			return m, nil
		}
	}

	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		if m.mode != CommList {
			m.mode = CommList
			m.err = nil
			m.msg = ""
			return m, nil
		}
		return m, func() tea.Msg { return ChangeScreenMsg(StateMenu) }
	}

	if m.mode == CommList {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "n":
				m.mode = CommAdd
				m.resetForm()
				return m, nil
			case "e", "enter":
				if row := m.table.SelectedRow(); len(row) > 0 {
					id, _ := strconv.Atoi(row[0])
					m.loadIntoForm(id)
					m.mode = CommEdit
				}
				return m, nil
			case "d":
				if row := m.table.SelectedRow(); len(row) > 0 {
					id, _ := strconv.Atoi(row[0])
					m.loadDetail(id)
				}
				return m, nil
			case "s":
				if row := m.table.SelectedRow(); len(row) > 0 {
					id, _ := strconv.Atoi(row[0])
					if err := m.toggleStato(id); err != nil {
						m.err = err
					}
				}
				return m, nil
			case "x":
				if row := m.table.SelectedRow(); len(row) > 0 {
					id, _ := strconv.Atoi(row[0])
					m.deletingID = id
					numMov, totMov := m.countMovimentiForCommessa(id)
					m.deleteWarningMov = numMov
					m.deleteWarningTot = totMov
					m.showConfirm = true
				}
				return m, nil
			}
		}
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	if m.mode == CommAdd || m.mode == CommEdit {
		if k, ok := msg.(tea.KeyMsg); ok {
			if m.focusIndex == 0 && (k.String() == "enter" || k.String() == " ") {
				m.selectionMode = true
				m.veicoloFilter.SetValue("")
				m.updateVeicoloTable()
				m.veicoloFilter.Focus()
				return m, nil
			}

			switch k.String() {
			case "enter":
				if m.focusIndex != 0 {
					if err := m.save(); err != nil {
						m.err = err
						return m, nil
					}
				}
				return m, nil
			case "tab", "down":
				m.focusIndex++
				if m.focusIndex > 4 {
					m.focusIndex = 0
				}
				m.updateFocus()
				return m, nil
			case "shift+tab", "up":
				m.focusIndex--
				if m.focusIndex < 0 {
					m.focusIndex = 4
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
func (m CommesseModel) View() string {
	width := 100
	if m.width > 0 {
		width = min(m.width, 110)
	}

	if m.showOverlay {
		return CenterContent(m.width, m.height, m.viewport.View())
	}

	if m.showConfirm {
		var message strings.Builder
		c, _ := m.db.GetCommessa(m.deletingID)
		numero := "???"
		if c != nil {
			numero = c.Numero
		}

		message.WriteString(fmt.Sprintf("⚠️  ELIMINAZIONE COMMESSA #%s\n\n", numero))

		if m.deleteWarningMov > 0 {
			message.WriteString(ErrorStyle.Render(fmt.Sprintf(
				"ATTENZIONE: Questa commessa ha dati associati!\n\n"+
					"Eliminando la commessa verranno eliminati:\n"+
					" • %d movimenti di Prima Nota (totale: %s)\n\n"+
					"TUTTI I DATI VERRANNO PERSI IN MODO PERMANENTE!\n\n",
				m.deleteWarningMov,
				utils.FormatEuro(m.deleteWarningTot),
			)))
		} else {
			message.WriteString("Questa commessa non ha movimenti associati.\n\n")
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

	title := "GESTIONE COMMESSE"
	if m.mode == CommAdd {
		title = "NUOVA COMMESSA"
	} else if m.mode == CommEdit {
		title = fmt.Sprintf("MODIFICA COMMESSA #%d", m.selectedID)
	}

	header := RenderHeader(title, width)
	var body string

	if m.mode == CommList {
		commesse, _ := m.db.ListCommesse()
		openCount := 0
		for _, c := range commesse {
			if c.Stato == "Aperta" {
				openCount++
			}
		}

		openBadge := ""
		if openCount > 0 {
			openBadge = WarningBadge(fmt.Sprintf(" %d aperte ", openCount)) + " "
		}

		helpText := lipgloss.NewStyle().
			MarginBottom(1).
			Foreground(ColorSubText).
			Render(openBadge + "[N] Nuova • [E/↵] Modifica • [D] Dettaglio • [S] Cambia Stato • [X] Elimina • [ESC] Menu")

		body = lipgloss.JoinVertical(
			lipgloss.Left,
			helpText,
			m.table.View(),
		)
	} else {
		var form strings.Builder
		labels := []string{"Veicolo", "Lavori Eseguiti", "Costo Manodopera", "Costo Ricambi", "Note"}

		for i, inp := range m.inputs {
			labelStyle := LabelStyle
			if i == m.focusIndex {
				labelStyle = LabelFocusedStyle
			}

			form.WriteString(fmt.Sprintf("%s %s\n",
				labelStyle.Render(labels[i]+":"),
				inp.View()))

			if i == 1 || i == 3 {
				form.WriteString("\n")
			}
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
