package screens

import (
	"fmt"
	"officina/database"
	"officina/utils"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CommessaMode rappresenta la modalità corrente
type CommessaMode int

const (
	CModeList CommessaMode = iota
	CModeAdd
	CModeEdit
)

// StandardWork rappresenta un lavoro standard
type StandardWork struct {
	Code string
	Desc string
}

// StandardWorks contiene i lavori predefiniti categorizzati
var StandardWorks = []StandardWork{
	// Tagliandi e manutenzione ordinaria
	{Code: "MEC-TAG-001", Desc: "Tagliando completo"},
	{Code: "MEC-TAG-002", Desc: "Tagliando intermedio"},
	{Code: "MEC-TAG-003", Desc: "Tagliando pre-revisione"},
	{Code: "MEC-TAG-004", Desc: "Reset spia service"},
	{Code: "MEC-TAG-005", Desc: "Diagnosi elettronica"},
	// Motore e distribuzione
	{Code: "MEC-MOT-001", Desc: "Sostituzione cinghia distribuzione"},
	{Code: "MEC-MOT-002", Desc: "Sostituzione catena distribuzione"},
	{Code: "MEC-MOT-003", Desc: "Sostituzione pompa acqua"},
	{Code: "MEC-MOT-004", Desc: "Revisione turbocompressore"},
	{Code: "MEC-MOT-005", Desc: "Sostituzione iniettori"},
	{Code: "MEC-MOT-006", Desc: "Pulizia valvola EGR"},
	{Code: "MEC-MOT-007", Desc: "Rigenerazione DPF"},
	{Code: "MEC-MOT-008", Desc: "Rifacimento guarnizione testata"},
	// Freni
	{Code: "MEC-FRE-001", Desc: "Sostituzione pastiglie anteriori"},
	{Code: "MEC-FRE-002", Desc: "Sostituzione pastiglie posteriori"},
	{Code: "MEC-FRE-003", Desc: "Sostituzione dischi anteriori"},
	{Code: "MEC-FRE-004", Desc: "Sostituzione dischi posteriori"},
	{Code: "MEC-FRE-005", Desc: "Spurgo impianto frenante"},
	// Trasmissione
	{Code: "MEC-TRA-001", Desc: "Sostituzione frizione"},
	{Code: "MEC-TRA-002", Desc: "Sostituzione volano"},
	{Code: "MEC-TRA-003", Desc: "Revisione cambio manuale"},
	{Code: "MEC-TRA-004", Desc: "Manutenzione cambio automatico"},
	// Carrozzeria
	{Code: "CAR-STR-001", Desc: "Riparazione post-incidente"},
	{Code: "CAR-STR-002", Desc: "Raddrizzatura telaio"},
	// Verniciatura
	{Code: "CAR-VER-001", Desc: "Verniciatura completa"},
	{Code: "CAR-VER-002", Desc: "Verniciatura paraurti"},
	{Code: "CAR-VER-003", Desc: "Verniciatura cofano"},
	{Code: "CAR-VER-004", Desc: "Verniciatura portiera"},
	// Cristalli
	{Code: "CAR-VET-001", Desc: "Sostituzione parabrezza"},
	{Code: "CAR-VET-002", Desc: "Riparazione parabrezza"},
	// Detailing
	{Code: "CAR-DET-001", Desc: "Lucidatura completa"},
	{Code: "CAR-DET-002", Desc: "Trattamento protettivo nano"},
}

// CommesseModel gestisce la schermata commesse
type CommesseModel struct {
	db                     *database.DB
	table                  table.Model
	inputs                 []textinput.Model
	mode                   CommessaMode
	focusIndex             int
	selectedID             int
	err                    error
	msg                    string
	width                  int
	height                 int
	selectionMode          bool
	veicoloTable           table.Model
	veicoloFilter          textinput.Model
	selectionModeLavori    bool
	lavoriTable            table.Model
	lavoriFilter           textinput.Model
	showLavoriPopup        bool
	popupLavoriText        string
	popupTitle             string
	showConfirm            bool
	deletingID             int
	veicoloID              int
	veicoloInfo            string
	deleteWarningMovimenti int
	deleteWarningTotale    float64
}

// NewCommesseModel crea una nuova istanza del model commesse
func NewCommesseModel(db *database.DB) CommesseModel {
	// Tabella principale commesse
	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Numero", Width: 10},
		{Title: "Veicolo", Width: 12},
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

	// Configurazione inputs
	inputs := make([]textinput.Model, 6)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "[ INVIO PER SCEGLIERE VEICOLO ]"
	inputs[0].Width = 50

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "[ INVIO AGGIUNGI STANDARD | SCRIVI LIBERO ]"
	inputs[1].Width = 60

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Costo manodopera (es. 150.00)"
	inputs[2].Width = 40

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Costo ricambi (es. 250.00)"
	inputs[3].Width = 40

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "Note interne..."
	inputs[4].Width = 60

	inputs[5] = textinput.New()
	inputs[5].Placeholder = "Stato: A (Aperta) / C (Chiusa)"
	inputs[5].CharLimit = 1
	inputs[5].Width = 10

	// Tabella selezione veicoli
	vCols := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Targa", Width: 10},
		{Title: "Modello", Width: 22},
		{Title: "Proprietario", Width: 22},
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

	// Tabella selezione lavori
	lCols := []table.Column{
		{Title: "Codice", Width: 15},
		{Title: "Descrizione", Width: 45},
	}
	lt := table.New(
		table.WithColumns(lCols),
		table.WithFocused(true),
		table.WithHeight(12),
	)
	lt.SetStyles(GetTableStyles())

	lf := textinput.New()
	lf.Placeholder = "🔍 Cerca lavorazione..."
	lf.Width = 50

	m := CommesseModel{
		db:            db,
		table:         t,
		inputs:        inputs,
		mode:          CModeList,
		veicoloTable:  vt,
		veicoloFilter: vf,
		lavoriTable:   lt,
		lavoriFilter:  lf,
	}

	m.Refresh()
	m.updateLavoriTable()
	return m
}

// Refresh aggiorna la lista delle commesse
func (m *CommesseModel) Refresh() {
	list, _ := m.db.ListCommesse()
	movimenti, _ := m.db.ListMovimenti()

	// Calcola acconti per commessa
	accontiMap := make(map[int]float64)
	for _, mov := range movimenti {
		if mov.CommessaID > 0 && mov.Tipo == "Entrata" {
			accontiMap[mov.CommessaID] += mov.Importo
		}
	}

	rows := []table.Row{}
	for _, c := range list {
		v, _ := m.db.GetVeicolo(c.VeicoloID)
		vLabel := fmt.Sprintf("ID:%d", c.VeicoloID)
		if v != nil {
			vLabel = v.Targa
		}

		versato := accontiMap[c.ID]
		residuo := c.Totale - versato
		if residuo < 0 {
			residuo = 0
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("%d", c.ID),
			c.Numero,
			vLabel,
			c.Stato,
			utils.FormatEuro(c.Totale),
			utils.FormatEuro(versato),
			utils.FormatEuro(residuo),
		})
	}

	m.table.SetRows(rows)
}

// countMovimentiForCommessa conta i movimenti associati a una commessa
func (m *CommesseModel) countMovimentiForCommessa(commessaID int) (int, float64) {
	movimenti, _ := m.db.ListMovimenti()
	count := 0
	totale := 0.0

	for _, mov := range movimenti {
		if mov.CommessaID == commessaID {
			count++
			if mov.Tipo == "Entrata" {
				totale += mov.Importo
			}
		}
	}

	return count, totale
}

// updateVeicoloTable aggiorna la tabella veicoli con filtro
func (m *CommesseModel) updateVeicoloTable() {
	veicoli, _ := m.db.ListVeicoli()
	filter := strings.ToUpper(strings.TrimSpace(m.veicoloFilter.Value()))
	rows := []table.Row{}

	for _, v := range veicoli {
		propName := "N/D"
		if v.ClienteID > 0 {
			cli, _ := m.db.GetCliente(v.ClienteID)
			if cli != nil {
				propName = fmt.Sprintf("%s %s", cli.Cognome, cli.Nome)
			}
		}

		targa := strings.ToUpper(v.Targa)
		modello := strings.ToUpper(v.Marca + " " + v.Modello)
		prop := strings.ToUpper(propName)

		if filter == "" ||
			strings.Contains(targa, filter) ||
			strings.Contains(modello, filter) ||
			strings.Contains(prop, filter) {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", v.ID),
				v.Targa,
				utils.Truncate(v.Marca+" "+v.Modello, 22),
				utils.Truncate(propName, 22),
			})
		}
	}

	m.veicoloTable.SetRows(rows)
}

// updateLavoriTable aggiorna la tabella lavori con filtro
func (m *CommesseModel) updateLavoriTable() {
	filter := strings.ToUpper(strings.TrimSpace(m.lavoriFilter.Value()))
	rows := []table.Row{}

	for _, w := range StandardWorks {
		code := strings.ToUpper(w.Code)
		desc := strings.ToUpper(w.Desc)

		if filter == "" ||
			strings.Contains(code, filter) ||
			strings.Contains(desc, filter) {
			rows = append(rows, table.Row{
				w.Code,
				w.Desc,
			})
		}
	}

	// Ordina per codice
	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	m.lavoriTable.SetRows(rows)
}

// resetForm resetta il form
func (m *CommesseModel) resetForm() {
	for i := range m.inputs {
		m.inputs[i].SetValue("")
	}
	m.inputs[5].SetValue("A") // Default: Aperta
	m.veicoloID = 0
	m.veicoloInfo = ""
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

	v, _ := m.db.GetVeicolo(c.VeicoloID)
	if v != nil {
		m.inputs[0].SetValue(v.Targa)
		propName := ""
		if v.ClienteID > 0 {
			cli, _ := m.db.GetCliente(v.ClienteID)
			if cli != nil {
				propName = cli.Cognome
			}
		}
		m.veicoloInfo = fmt.Sprintf("%s %s (%s)", v.Marca, v.Modello, propName)
	} else {
		m.inputs[0].SetValue("???")
		m.veicoloInfo = "Veicolo eliminato"
	}

	m.inputs[1].SetValue(c.LavoriEseguiti)
	m.inputs[2].SetValue(fmt.Sprintf("%.2f", c.CostoManodopera))
	m.inputs[3].SetValue(fmt.Sprintf("%.2f", c.CostoRicambi))
	m.inputs[4].SetValue(c.Note)

	state := "A"
	if c.Stato == "Chiusa" {
		state = "C"
	}
	m.inputs[5].SetValue(state)

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

	val := strings.ToUpper(strings.TrimSpace(m.inputs[5].Value()))
	if val != "A" && val != "C" {
		return fmt.Errorf("stato non valido: usa 'A' (Aperta) o 'C' (Chiusa)")
	}

	return nil
}

// save salva la commessa corrente
func (m *CommesseModel) save() error {
	if err := m.validate(); err != nil {
		return err
	}

	val := strings.ToUpper(strings.TrimSpace(m.inputs[5].Value()))
	stato := "Aperta"
	if val == "C" {
		stato = "Chiusa"
	}

	cm, _ := utils.ParseFloat(m.inputs[2].Value())
	cr, _ := utils.ParseFloat(m.inputs[3].Value())

	c := &database.Commessa{
		VeicoloID:       m.veicoloID,
		LavoriEseguiti:  m.inputs[1].Value(),
		CostoManodopera: cm,
		CostoRicambi:    cr,
		Note:            m.inputs[4].Value(),
		Stato:           stato,
	}

	if m.mode == CModeAdd {
		if err := m.db.CreateCommessa(c); err != nil {
			return fmt.Errorf("errore creazione: %w", err)
		}
		m.msg = "✓ Commessa creata con successo"
	} else {
		c.ID = m.selectedID
		old, _ := m.db.GetCommessa(m.selectedID)
		if old != nil {
			c.DataApertura = old.DataApertura
			c.Numero = old.Numero
		}
		if err := m.db.UpdateCommessa(c); err != nil {
			return fmt.Errorf("errore aggiornamento: %w", err)
		}
		m.msg = "✓ Commessa aggiornata con successo"
	}

	m.mode = CModeList
	m.Refresh()
	return nil
}

// Init implementa tea.Model
func (m CommesseModel) Init() tea.Cmd {
	return nil
}

// Update implementa tea.Model
func (m CommesseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Gestione resize
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	// Popup visualizzazione lavori (read-only)
	if m.showLavoriPopup {
		if k, ok := msg.(tea.KeyMsg); ok {
			if k.String() == "esc" || k.String() == "enter" || k.String() == "q" {
				m.showLavoriPopup = false
			}
		}
		return m, nil
	}

	// Popup selezione veicolo
	if m.selectionMode {
		return m.handleVehicleSelection(msg)
	}

	// Popup selezione lavori
	if m.selectionModeLavori {
		return m.handleWorkSelection(msg)
	}

	// Conferma eliminazione
	if m.showConfirm {
		return m.handleDeleteConfirmation(msg)
	}

	// Gestione ESC
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		if m.mode != CModeList {
			m.mode = CModeList
			m.err = nil
			m.msg = ""
			return m, nil
		}
		return m, func() tea.Msg { return ChangeScreenMsg(StateMenu) }
	}

	// Modalità Lista
	if m.mode == CModeList {
		return m.handleListMode(msg)
	}

	// Modalità Form (Add/Edit)
	return m.handleFormMode(msg)
}

// handleVehicleSelection gestisce la selezione del veicolo
func (m *CommesseModel) handleVehicleSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				m.veicoloInfo = fmt.Sprintf("%s (%s)", row[2], row[3])
				m.inputs[0].SetValue(row[1])
				m.selectionMode = false
				m.focusIndex = 1
				m.updateFocus()
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

// handleWorkSelection gestisce la selezione dei lavori
func (m *CommesseModel) handleWorkSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.selectionModeLavori = false
			m.updateFocus()
			return m, nil

		case "enter":
			if row := m.lavoriTable.SelectedRow(); len(row) > 0 {
				selectedDesc := row[1]
				currentVal := m.inputs[1].Value()
				if strings.TrimSpace(currentVal) == "" {
					m.inputs[1].SetValue(selectedDesc)
				} else {
					m.inputs[1].SetValue(currentVal + ", " + selectedDesc)
				}

				// Scrolla alla fine del campo
				m.inputs[1].CursorEnd()

				m.selectionModeLavori = false
				m.inputs[1].Focus()
			}
			return m, nil
		}

		var cmdF, cmdT tea.Cmd
		m.lavoriFilter, cmdF = m.lavoriFilter.Update(msg)
		m.updateLavoriTable()
		m.lavoriTable, cmdT = m.lavoriTable.Update(msg)
		return m, tea.Batch(cmdF, cmdT)
	}

	return m, nil
}

// handleDeleteConfirmation gestisce la conferma eliminazione
func (m *CommesseModel) handleDeleteConfirmation(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "y", "Y":
			if err := m.db.DeleteCommessa(m.deletingID); err != nil {
				m.err = fmt.Errorf("errore eliminazione: %w", err)
			} else {
				m.msg = "✓ Commessa eliminata con successo"
			}
			m.Refresh()
			m.showConfirm = false
			m.deleteWarningMovimenti = 0
			m.deleteWarningTotale = 0

		case "n", "N", "esc":
			m.showConfirm = false
			m.deleteWarningMovimenti = 0
			m.deleteWarningTotale = 0
		}
	}
	return m, nil
}

// handleListMode gestisce gli eventi in modalità lista
func (m *CommesseModel) handleListMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			m.mode = CModeAdd
			m.resetForm()
			return m, nil

		case "e", "enter":
			if row := m.table.SelectedRow(); len(row) > 0 {
				id, _ := strconv.Atoi(row[0])
				m.loadIntoForm(id)
				m.mode = CModeEdit
			}
			return m, nil

		case "x", "d":
			if row := m.table.SelectedRow(); len(row) > 0 {
				id, _ := strconv.Atoi(row[0])
				m.deletingID = id

				// Conta movimenti associati
				numMovimenti, totaleMovimenti := m.countMovimentiForCommessa(id)
				m.deleteWarningMovimenti = numMovimenti
				m.deleteWarningTotale = totaleMovimenti

				m.showConfirm = true
			}
			return m, nil

		case "l", "L":
			// Visualizza lavori della commessa selezionata
			if row := m.table.SelectedRow(); len(row) > 0 {
				id, _ := strconv.Atoi(row[0])
				commessa, _ := m.db.GetCommessa(id)
				if commessa != nil {
					lavoriList := strings.Split(commessa.LavoriEseguiti, ",")
					var formatted strings.Builder
					for _, l := range lavoriList {
						formatted.WriteString("• " + strings.TrimSpace(l) + "\n")
					}
					m.popupTitle = fmt.Sprintf("LAVORI COMMESSA #%s", commessa.Numero)
					m.popupLavoriText = formatted.String()
					m.showLavoriPopup = true
				}
			}
			return m, nil
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// handleFormMode gestisce gli eventi in modalità form
func (m *CommesseModel) handleFormMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		// Campo veicolo: apre selezione
		if m.focusIndex == 0 && (k.String() == "enter" || k.String() == " ") {
			m.selectionMode = true
			m.veicoloFilter.SetValue("")
			m.updateVeicoloTable()
			m.veicoloFilter.Focus()
			return m, nil
		}

		// Campo lavori: apre selezione standard
		if m.focusIndex == 1 && k.String() == "enter" {
			m.selectionModeLavori = true
			m.lavoriFilter.SetValue("")
			m.updateLavoriTable()
			m.lavoriFilter.Focus()
			return m, nil
		}

		// Salvataggio su campo stato (ultimo)
		if m.focusIndex == 5 && k.String() == "enter" {
			if err := m.save(); err != nil {
				m.err = err
				return m, nil
			}
			return m, nil
		}

		switch k.String() {
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

		// Auto-uppercase per stato
		if m.focusIndex == 5 {
			m.inputs[5].SetValue(strings.ToUpper(m.inputs[5].Value()))
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
func (m CommesseModel) View() string {
	width := 100
	if m.width > 0 {
		width = min(m.width, 120)
	}

	// Popup visualizzazione lavori
	if m.showLavoriPopup {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			TitleStyle.Render(m.popupTitle),
			"",
			m.popupLavoriText,
			"",
			HelpStyle.Render("[Esc/↵/Q] Chiudi"),
		)
		return RenderPopup("DETTAGLIO LAVORI", content, m.width, m.height)
	}

	// Popup selezione veicolo
	if m.selectionMode {
		return m.renderVehicleSelection(width)
	}

	// Popup selezione lavori
	if m.selectionModeLavori {
		return m.renderWorkSelection(width)
	}

	// Dialog conferma eliminazione
	if m.showConfirm {
		return m.renderDeleteConfirmation(width)
	}

	// Vista principale
	if m.mode == CModeList {
		return m.renderListView(width)
	}

	return m.renderFormView(width)
}

// renderDeleteConfirmation renderizza il dialog di conferma eliminazione con avviso
func (m CommesseModel) renderDeleteConfirmation(width int) string {
	var message strings.Builder

	message.WriteString(fmt.Sprintf("⚠️  ELIMINAZIONE COMMESSA #%d\n\n", m.deletingID))

	if m.deleteWarningMovimenti > 0 {
		message.WriteString(ErrorStyle.Render(fmt.Sprintf(
			"ATTENZIONE: Questa commessa ha %d movimenti di Prima Nota associati\n"+
				"per un totale di %s!\n\n"+
				"Eliminando la commessa verranno eliminati ANCHE TUTTI i movimenti\n"+
				"di cassa collegati.\n\n",
			m.deleteWarningMovimenti,
			utils.FormatEuro(m.deleteWarningTotale),
		)))
	} else {
		message.WriteString("Questa commessa non ha movimenti di Prima Nota associati.\n\n")
	}

	message.WriteString(WarningStyle.Render("Sei sicuro di voler procedere?\n"))
	message.WriteString(HelpStyle.Render("\n[Y] Sì, elimina • [N/Esc] Annulla"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorError).
		Padding(1, 2).
		Width(70).
		Render(message.String())

	if m.width > 0 && m.height > 0 {
		return CenterContent(m.width, m.height, box)
	}

	return box
}

// renderListView renderizza la vista lista
func (m CommesseModel) renderListView(width int) string {
	title := "GESTIONE COMMESSE"
	header := RenderHeader(title, width)

	helpText := lipgloss.NewStyle().
		MarginBottom(1).
		Foreground(ColorSubText).
		Render("[N] Nuova • [E/↵] Modifica • [L] Lavori • [X/D] Elimina • [ESC] Menu")

	body := lipgloss.JoinVertical(
		lipgloss.Left,
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
func (m CommesseModel) renderFormView(width int) string {
	title := "NUOVA COMMESSA"
	if m.mode == CModeEdit {
		title = fmt.Sprintf("MODIFICA COMMESSA #%d", m.selectedID)
	}

	header := RenderHeader(title, width)

	// Form
	var form strings.Builder
	labels := []string{
		"🚗 Veicolo",
		"🔧 Lavori",
		"💰 Manodopera €",
		"🔩 Ricambi €",
		"📝 Note",
		"📊 Stato (A/C)",
	}

	for i := 0; i < len(m.inputs); i++ {
		labelStyle := LabelStyle
		if i == m.focusIndex {
			labelStyle = LabelFocusedStyle
		}

		view := m.inputs[i].View()

		// Campo veicolo (readonly)
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

		// Campo lavori (con hint)
		if i == 1 && m.focusIndex == 1 && m.inputs[1].Value() == "" {
			view = lipgloss.NewStyle().
				Foreground(ColorHighlight).
				Render("[ INVIO = Lavori Standard | Scrivi = Testo Libero ]")
		}

		form.WriteString(fmt.Sprintf("%s %s\n",
			labelStyle.Render(labels[i]+":"),
			view))
		if i == 0 || i == 1 || i == 3 {
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

// renderVehicleSelection renderizza il popup selezione veicolo
func (m CommesseModel) renderVehicleSelection(width int) string {
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

// renderWorkSelection renderizza il popup selezione lavori
func (m CommesseModel) renderWorkSelection(width int) string {
	title := TitleStyle.Render("🔧 LAVORI STANDARD")

	filter := fmt.Sprintf("Cerca: %s", m.lavoriFilter.View())

	inner := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		filter,
		"",
		m.lavoriTable.View(),
		"",
		HelpStyle.Render("[↑↓] Naviga • [↵] Aggiungi • [Esc] Annulla"),
	)

	box := MainBoxStyle.Copy().
		Width(min(width-10, 75)).
		Render(inner)

	if m.width > 0 && m.height > 0 {
		return CenterContent(m.width, m.height, box)
	}

	return box
}
