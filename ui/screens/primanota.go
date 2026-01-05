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

// PrimaNotaMode rappresenta la modalità corrente
type PrimaNotaMode int

const (
	PNModeList PrimaNotaMode = iota
	PNModeAdd
	PNModeEdit
	PNModeFilter
)

// MetodiPagamento contiene i metodi di pagamento disponibili
var MetodiPagamento = []string{"CASSA", "BANCA", "POS", "ASSEGNO", "BONIFICO"}

// PrimaNotaModel gestisce la schermata prima nota
type PrimaNotaModel struct {
	db                     *database.DB
	table                  table.Model
	inputs                 []textinput.Model
	mode                   PrimaNotaMode
	focusIndex             int
	err                    error
	msg                    string
	width                  int
	height                 int
	selectedID             int
	totaleEntrate          float64
	totaleUscite           float64
	saldo                  float64
	selectionMode          bool
	commessaTable          table.Model
	commessaFilter         textinput.Model
	selectedCommID         int
	currentMetodoIdx       int
	showConfirm            bool
	deletingID             int
	selectedCommessaTotale float64
	selectedCommessaNumero string
	selectedCommessaTarga  string
	// Filtri
	filterInputs    []textinput.Model
	filterFocusIdx  int
	filteredList    []database.MovimentoPrimaNota
	hasActiveFilter bool
}

// NewPrimaNotaModel crea una nuova istanza del model prima nota
func NewPrimaNotaModel(db *database.DB) PrimaNotaModel {
	// Tabella principale movimenti
	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Data", Width: 12},
		{Title: "Descrizione", Width: 45},
		{Title: "Tipo", Width: 8},
		{Title: "Metodo", Width: 10},
		{Title: "Importo", Width: 12},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(12),
	)
	t.SetStyles(GetTableStyles())

	// Configurazione inputs
	inputs := make([]textinput.Model, 5)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Data (GG/MM/AAAA)"
	inputs[0].CharLimit = 10
	inputs[0].Width = 30

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Tipo: E (Entrata) / U (Uscita)"
	inputs[1].CharLimit = 1
	inputs[1].Width = 10

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Importo (es. 1250.50)"
	inputs[2].Width = 30

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "[ SPAZIO = Cambia Metodo ]"
	inputs[3].Width = 30

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "[ INVIO = Cerca Commessa ] o scrivi descrizione..."
	inputs[4].Width = 60

	// Configurazione filtri
	filterInputs := make([]textinput.Model, 4)

	filterInputs[0] = textinput.New()
	filterInputs[0].Placeholder = "Data DA (GG/MM/AAAA)"
	filterInputs[0].CharLimit = 10
	filterInputs[0].Width = 30

	filterInputs[1] = textinput.New()
	filterInputs[1].Placeholder = "Data A (GG/MM/AAAA)"
	filterInputs[1].CharLimit = 10
	filterInputs[1].Width = 30

	filterInputs[2] = textinput.New()
	filterInputs[2].Placeholder = "Cerca nella descrizione..."
	filterInputs[2].Width = 50

	filterInputs[3] = textinput.New()
	filterInputs[3].Placeholder = "Importo (es. 100 o >500 o <1000)"
	filterInputs[3].Width = 35

	// Tabella selezione commesse
	cCols := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Numero", Width: 10},
		{Title: "Veicolo", Width: 12},
		{Title: "Stato", Width: 8},
		{Title: "Totale", Width: 12},
		{Title: "Versato", Width: 12},
		{Title: "Residuo", Width: 12},
	}
	ct := table.New(
		table.WithColumns(cCols),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	ct.SetStyles(GetTableStyles())

	cf := textinput.New()
	cf.Placeholder = "🔍 Cerca commessa/targa..."
	cf.Width = 50

	m := PrimaNotaModel{
		db:               db,
		table:            t,
		inputs:           inputs,
		mode:             PNModeList,
		commessaTable:    ct,
		commessaFilter:   cf,
		currentMetodoIdx: 0,
		filterInputs:     filterInputs,
		filterFocusIdx:   0,
		hasActiveFilter:  false,
	}

	m.Refresh()
	return m
}

// Refresh aggiorna la lista dei movimenti e calcola totali
func (m *PrimaNotaModel) Refresh() {
	list, _ := m.db.ListMovimenti()

	// Applica filtri se attivi
	if m.hasActiveFilter {
		list = m.applyFilters(list)
	}

	// Ordina per data decrescente
	sort.Slice(list, func(i, j int) bool {
		return list[i].Data.After(list[j].Data)
	})

	m.totaleEntrate = 0
	m.totaleUscite = 0
	rows := []table.Row{}

	for _, mov := range list {
		if mov.Tipo == "Entrata" {
			m.totaleEntrate += mov.Importo
		} else {
			m.totaleUscite += mov.Importo
		}

		tipo := "🔴 OUT"
		if mov.Tipo == "Entrata" {
			tipo = "🟢 IN"
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("%d", mov.ID),
			utils.FormatDate(mov.Data),
			utils.Truncate(mov.Descrizione, 45),
			tipo,
			mov.Metodo,
			utils.FormatEuro(mov.Importo),
		})
	}

	m.saldo = m.totaleEntrate - m.totaleUscite
	m.table.SetRows(rows)
}

// applyFilters applica i filtri alla lista movimenti
func (m *PrimaNotaModel) applyFilters(list []database.MovimentoPrimaNota) []database.MovimentoPrimaNota {
	var filtered []database.MovimentoPrimaNota

	dataDa := strings.TrimSpace(m.filterInputs[0].Value())
	dataA := strings.TrimSpace(m.filterInputs[1].Value())
	descrizione := strings.ToUpper(strings.TrimSpace(m.filterInputs[2].Value()))
	importoStr := strings.TrimSpace(m.filterInputs[3].Value())

	var dataInizio, dataFine time.Time
	var errDa, errA error

	if dataDa != "" {
		dataInizio, errDa = time.Parse("02/01/2006", dataDa)
	}
	if dataA != "" {
		dataFine, errA = time.Parse("02/01/2006", dataA)
	}

	for _, mov := range list {
		// Filtro data DA
		if errDa == nil && dataDa != "" {
			if mov.Data.Before(dataInizio) {
				continue
			}
		}

		// Filtro data A
		if errA == nil && dataA != "" {
			if mov.Data.After(dataFine) {
				continue
			}
		}

		// Filtro descrizione
		if descrizione != "" {
			if !strings.Contains(strings.ToUpper(mov.Descrizione), descrizione) {
				continue
			}
		}

		// Filtro importo
		if importoStr != "" {
			if !m.matchImporto(mov.Importo, importoStr) {
				continue
			}
		}

		filtered = append(filtered, mov)
	}

	return filtered
}

// matchImporto verifica se un importo corrisponde al filtro
func (m *PrimaNotaModel) matchImporto(importo float64, filter string) bool {
	filter = strings.TrimSpace(filter)

	// Operatori: >, <, >=, <=, =
	if strings.HasPrefix(filter, ">=") {
		val, err := utils.ParseFloat(filter[2:])
		return err == nil && importo >= val
	}
	if strings.HasPrefix(filter, "<=") {
		val, err := utils.ParseFloat(filter[2:])
		return err == nil && importo <= val
	}
	if strings.HasPrefix(filter, ">") {
		val, err := utils.ParseFloat(filter[1:])
		return err == nil && importo > val
	}
	if strings.HasPrefix(filter, "<") {
		val, err := utils.ParseFloat(filter[1:])
		return err == nil && importo < val
	}
	if strings.HasPrefix(filter, "=") {
		val, err := utils.ParseFloat(filter[1:])
		return err == nil && importo == val
	}

	// Nessun operatore = uguaglianza
	val, err := utils.ParseFloat(filter)
	return err == nil && importo == val
}

// clearFilters pulisce tutti i filtri
func (m *PrimaNotaModel) clearFilters() {
	for i := range m.filterInputs {
		m.filterInputs[i].SetValue("")
	}
	m.hasActiveFilter = false
	m.Refresh()
}

// updateCommessaTable aggiorna la tabella commesse con filtro
func (m *PrimaNotaModel) updateCommessaTable() {
	commesse, _ := m.db.ListCommesse()
	movimenti, _ := m.db.ListMovimenti()

	filter := strings.ToUpper(strings.TrimSpace(m.commessaFilter.Value()))
	rows := []table.Row{}

	// Calcola versato per commessa
	accontiMap := make(map[int]float64)
	for _, mov := range movimenti {
		if mov.CommessaID > 0 && mov.Tipo == "Entrata" {
			// Escludi movimento in modifica
			if m.mode == PNModeEdit && mov.ID == m.selectedID {
				continue
			}
			accontiMap[mov.CommessaID] += mov.Importo
		}
	}

	for _, c := range commesse {
		v, _ := m.db.GetVeicolo(c.VeicoloID)
		targa := "???"
		if v != nil {
			targa = v.Targa
		}

		versato := accontiMap[c.ID]
		residuo := c.Totale - versato
		if residuo < 0 {
			residuo = 0
		}

		numero := strings.ToUpper(c.Numero)
		targaUpper := strings.ToUpper(targa)

		if filter == "" ||
			strings.Contains(numero, filter) ||
			strings.Contains(targaUpper, filter) {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", c.ID),
				c.Numero,
				targa,
				c.Stato,
				utils.FormatEuro(c.Totale),
				utils.FormatEuro(versato),
				utils.FormatEuro(residuo),
			})
		}
	}

	m.commessaTable.SetRows(rows)
}

// calcolaVersatoCommessa calcola quanto già versato per una commessa
func (m *PrimaNotaModel) calcolaVersatoCommessa(commessaID int) float64 {
	movimenti, _ := m.db.ListMovimenti()
	var versato float64

	for _, mov := range movimenti {
		if mov.CommessaID == commessaID && mov.Tipo == "Entrata" {
			// Escludi movimento in modifica
			if m.mode == PNModeEdit && mov.ID == m.selectedID {
				continue
			}
			versato += mov.Importo
		}
	}

	return versato
}

// resetForm resetta il form
func (m *PrimaNotaModel) resetForm() {
	for i := range m.inputs {
		m.inputs[i].SetValue("")
	}
	m.inputs[0].SetValue(time.Now().Format("02/01/2006"))
	m.inputs[1].SetValue("E")
	// Default: CASSA
	m.currentMetodoIdx = 0
	m.inputs[3].SetValue(MetodiPagamento[0])
	m.selectedCommID = 0
	m.selectedCommessaTotale = 0
	m.selectedCommessaNumero = ""
	m.selectedCommessaTarga = ""
	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.inputs[0].Focus()
}

// loadIntoForm carica un movimento nel form
func (m *PrimaNotaModel) loadIntoForm(id int) {
	list, _ := m.db.ListMovimenti()
	var mov *database.MovimentoPrimaNota

	for i := range list {
		if list[i].ID == id {
			mov = &list[i]
			break
		}
	}

	if mov == nil {
		m.err = fmt.Errorf("movimento non trovato")
		return
	}

	m.selectedID = mov.ID
	m.selectedCommID = mov.CommessaID
	m.inputs[0].SetValue(mov.Data.Format("02/01/2006"))

	tipoChar := "E"
	if mov.Tipo == "Uscita" {
		tipoChar = "U"
	}
	m.inputs[1].SetValue(tipoChar)

	m.inputs[2].SetValue(fmt.Sprintf("%.2f", mov.Importo))

	// Trova indice metodo
	found := false
	for i, v := range MetodiPagamento {
		if v == mov.Metodo {
			m.currentMetodoIdx = i
			found = true
			break
		}
	}
	if !found {
		m.currentMetodoIdx = 0
	}
	m.inputs[3].SetValue(MetodiPagamento[m.currentMetodoIdx])

	m.inputs[4].SetValue(mov.Descrizione)

	// Carica info commessa se presente
	if mov.CommessaID > 0 {
		comm, _ := m.db.GetCommessa(mov.CommessaID)
		if comm != nil {
			m.selectedCommessaTotale = comm.Totale
			m.selectedCommessaNumero = comm.Numero
			if v, _ := m.db.GetVeicolo(comm.VeicoloID); v != nil {
				m.selectedCommessaTarga = v.Targa
			}
		}
	}

	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.inputs[0].Focus()
}

// updateFocus aggiorna il focus tra i campi
func (m *PrimaNotaModel) updateFocus() {
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

// updateFilterFocus aggiorna il focus tra i campi filtro
func (m *PrimaNotaModel) updateFilterFocus() {
	for i := range m.filterInputs {
		if i == m.filterFocusIdx {
			m.filterInputs[i].Focus()
		} else {
			m.filterInputs[i].Blur()
		}
	}
}

// validate valida i dati del form
func (m *PrimaNotaModel) validate() error {
	dateStr := strings.TrimSpace(m.inputs[0].Value())
	if len(dateStr) != 10 {
		return fmt.Errorf("data incompleta (formato: GG/MM/AAAA)")
	}

	_, err := time.Parse("02/01/2006", dateStr)
	if err != nil {
		return fmt.Errorf("data non valida")
	}

	tipoInput := strings.ToUpper(strings.TrimSpace(m.inputs[1].Value()))
	if tipoInput != "E" && tipoInput != "U" {
		return fmt.Errorf("tipo errato: usa 'E' (Entrata) o 'U' (Uscita)")
	}

	importo, err := utils.ParseFloat(m.inputs[2].Value())
	if err != nil {
		return fmt.Errorf("importo non valido")
	}

	if err := utils.ValidateImportoPositivo(importo); err != nil {
		return err
	}

	if err := utils.ValidateNotEmpty(m.inputs[4].Value(), "Descrizione"); err != nil {
		return err
	}

	return nil
}

// save salva il movimento corrente
func (m *PrimaNotaModel) save() error {
	if err := m.validate(); err != nil {
		return err
	}

	dateStr := strings.TrimSpace(m.inputs[0].Value())
	data, _ := time.Parse("02/01/2006", dateStr)

	tipoInput := strings.ToUpper(strings.TrimSpace(m.inputs[1].Value()))
	tipo := "Entrata"
	if tipoInput == "U" {
		tipo = "Uscita"
	}

	importo, _ := utils.ParseFloat(m.inputs[2].Value())
	metodo := m.inputs[3].Value()
	descrizione := m.inputs[4].Value()

	// Se è un'entrata collegata a una commessa, calcola la descrizione corretta
	if tipo == "Entrata" && m.selectedCommID > 0 {
		// Calcola quanto già versato
		versato := m.calcolaVersatoCommessa(m.selectedCommID)
		totaleConQuesto := versato + importo

		// Recupera i dati della commessa
		numero := m.selectedCommessaNumero
		targa := m.selectedCommessaTarga
		totale := m.selectedCommessaTotale

		// Se i dati non sono stati memorizzati, recuperali dal database
		if numero == "" {
			if comm, _ := m.db.GetCommessa(m.selectedCommID); comm != nil {
				totale = comm.Totale
				numero = comm.Numero
				if v, _ := m.db.GetVeicolo(comm.VeicoloID); v != nil {
					targa = v.Targa
				}
			}
		}

		// Determina la descrizione in base all'importo totale versato
		if totaleConQuesto >= totale {
			descrizione = fmt.Sprintf("Saldo Commessa #%s - Targa %s", numero, targa)
		} else {
			descrizione = fmt.Sprintf("Acconto Commessa #%s - Targa %s", numero, targa)
		}
	}

	mov := &database.MovimentoPrimaNota{
		Data:        data,
		Tipo:        tipo,
		Importo:     importo,
		Metodo:      metodo,
		Descrizione: descrizione,
		CommessaID:  m.selectedCommID,
	}

	if m.mode == PNModeAdd {
		if err := m.db.CreateMovimento(mov); err != nil {
			return fmt.Errorf("errore creazione: %w", err)
		}
		m.msg = "✓ Movimento registrato con successo"
	} else {
		mov.ID = m.selectedID
		if err := m.db.UpdateMovimento(mov); err != nil {
			return fmt.Errorf("errore aggiornamento: %w", err)
		}
		m.msg = "✓ Movimento aggiornato con successo"
	}

	m.mode = PNModeList
	m.Refresh()
	return nil
}

// Init implementa tea.Model
func (m PrimaNotaModel) Init() tea.Cmd {
	return nil
}

// Update implementa tea.Model
func (m PrimaNotaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Gestione resize
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	// Popup selezione commessa
	if m.selectionMode {
		return m.handleCommessaSelection(msg)
	}

	// Conferma eliminazione
	if m.showConfirm {
		return m.handleDeleteConfirmation(msg)
	}

	// Gestione ESC
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		if m.mode == PNModeFilter {
			m.mode = PNModeList
			m.err = nil
			return m, nil
		}
		if m.mode != PNModeList {
			m.mode = PNModeList
			m.err = nil
			m.msg = ""
			return m, nil
		}
		return m, func() tea.Msg { return ChangeScreenMsg(StateMenu) }
	}

	// Modalità Filtro
	if m.mode == PNModeFilter {
		return m.handleFilterMode(msg)
	}

	// Modalità Lista
	if m.mode == PNModeList {
		return m.handleListMode(msg)
	}

	// Modalità Form (Add/Edit)
	return m.handleFormMode(msg)
}

// handleFilterMode gestisce la modalità filtro
func (m *PrimaNotaModel) handleFilterMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "enter":
			// Valida le date prima di applicare i filtri
			dataDa := strings.TrimSpace(m.filterInputs[0].Value())
			dataA := strings.TrimSpace(m.filterInputs[1].Value())

			// Valida data DA
			if dataDa != "" {
				if len(dataDa) != 10 {
					m.err = fmt.Errorf("data DA incompleta (formato: GG/MM/AAAA)")
					return m, nil
				}
				if _, err := time.Parse("02/01/2006", dataDa); err != nil {
					m.err = fmt.Errorf("data DA non valida")
					return m, nil
				}
			}

			// Valida data A
			if dataA != "" {
				if len(dataA) != 10 {
					m.err = fmt.Errorf("data A incompleta (formato: GG/MM/AAAA)")
					return m, nil
				}
				if _, err := time.Parse("02/01/2006", dataA); err != nil {
					m.err = fmt.Errorf("data A non valida")
					return m, nil
				}
			}

			// Verifica che data DA sia <= data A
			if dataDa != "" && dataA != "" {
				da, _ := time.Parse("02/01/2006", dataDa)
				a, _ := time.Parse("02/01/2006", dataA)
				if da.After(a) {
					m.err = fmt.Errorf("data DA deve essere precedente o uguale alla data A")
					return m, nil
				}
			}

			// Valida formato importo se presente
			importoStr := strings.TrimSpace(m.filterInputs[3].Value())
			if importoStr != "" {
				if !m.validateImportoFilter(importoStr) {
					m.err = fmt.Errorf("formato importo non valido (usa: 100, >500, <1000, >=200, <=800, =150)")
					return m, nil
				}
			}

			// Applica filtri
			m.hasActiveFilter = true
			m.mode = PNModeList
			m.err = nil
			m.Refresh()
			m.msg = "✓ Filtri applicati"
			return m, nil

		case "ctrl+r":
			// Reset filtri
			m.clearFilters()
			m.mode = PNModeList
			m.err = nil
			m.msg = "✓ Filtri rimossi"
			return m, nil

		case "tab", "down":
			m.filterFocusIdx++
			if m.filterFocusIdx >= len(m.filterInputs) {
				m.filterFocusIdx = 0
			}
			m.err = nil
			m.updateFilterFocus()
			return m, nil

		case "shift+tab", "up":
			m.filterFocusIdx--
			if m.filterFocusIdx < 0 {
				m.filterFocusIdx = len(m.filterInputs) - 1
			}
			m.err = nil
			m.updateFilterFocus()
			return m, nil
		}
	}

	// Update filtro corrente
	cmds := make([]tea.Cmd, len(m.filterInputs))
	for i := range m.filterInputs {
		if i == m.filterFocusIdx {
			m.filterInputs[i], cmds[i] = m.filterInputs[i].Update(msg)
		}
	}

	return m, tea.Batch(cmds...)
}

// validateImportoFilter valida il formato del filtro importo
func (m *PrimaNotaModel) validateImportoFilter(filter string) bool {
	filter = strings.TrimSpace(filter)

	// Controlla operatori
	if strings.HasPrefix(filter, ">=") {
		_, err := utils.ParseFloat(filter[2:])
		return err == nil
	}
	if strings.HasPrefix(filter, "<=") {
		_, err := utils.ParseFloat(filter[2:])
		return err == nil
	}
	if strings.HasPrefix(filter, ">") {
		_, err := utils.ParseFloat(filter[1:])
		return err == nil
	}
	if strings.HasPrefix(filter, "<") {
		_, err := utils.ParseFloat(filter[1:])
		return err == nil
	}
	if strings.HasPrefix(filter, "=") {
		_, err := utils.ParseFloat(filter[1:])
		return err == nil
	}

	// Nessun operatore = deve essere un numero valido
	_, err := utils.ParseFloat(filter)
	return err == nil
}

// handleCommessaSelection gestisce la selezione della commessa
func (m *PrimaNotaModel) handleCommessaSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.selectionMode = false
			m.inputs[4].Focus()
			return m, nil

		case "enter":
			if row := m.commessaTable.SelectedRow(); len(row) > 0 {
				cID, _ := strconv.Atoi(row[0])
				totaleStr := row[4]
				residuoStr := row[6]

				totale, _ := utils.ParseFloat(totaleStr)
				residuo, _ := utils.ParseFloat(residuoStr)

				m.selectedCommID = cID
				m.selectedCommessaTotale = totale
				m.selectedCommessaNumero = row[1]
				m.selectedCommessaTarga = row[2]

				// Controlla se il campo importo è vuoto
				importoVuoto := strings.TrimSpace(m.inputs[2].Value()) == ""

				if importoVuoto {
					// Campo vuoto: riempi con residuo e usa descrizione temporanea
					m.inputs[2].SetValue(fmt.Sprintf("%.2f", residuo))
					desc := fmt.Sprintf("Saldo Commessa #%s - Targa %s", row[1], row[2])
					m.inputs[4].SetValue(desc)
				} else {
					// Campo già compilato: NON modificare l'importo, descrizione temporanea
					desc := fmt.Sprintf("Commessa #%s - Targa %s", row[1], row[2])
					m.inputs[4].SetValue(desc)
				}

				m.inputs[1].SetValue("E")
				m.selectionMode = false
				m.focusIndex = 2
				m.updateFocus()
			}
			return m, nil
		}

		var cmdF, cmdT tea.Cmd
		m.commessaFilter, cmdF = m.commessaFilter.Update(msg)
		m.updateCommessaTable()
		m.commessaTable, cmdT = m.commessaTable.Update(msg)
		return m, tea.Batch(cmdF, cmdT)
	}

	return m, nil
}

// handleDeleteConfirmation gestisce la conferma eliminazione
func (m *PrimaNotaModel) handleDeleteConfirmation(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "y", "Y":
			if err := m.db.DeleteMovimento(m.deletingID); err != nil {
				m.err = fmt.Errorf("errore eliminazione: %w", err)
			} else {
				m.msg = "✓ Movimento eliminato con successo"
			}
			m.Refresh()
			m.showConfirm = false

		case "n", "N", "esc":
			m.showConfirm = false
		}
	}
	return m, nil
}

// handleListMode gestisce gli eventi in modalità lista
func (m *PrimaNotaModel) handleListMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			m.mode = PNModeAdd
			m.resetForm()
			return m, nil

		case "f", "ctrl+f":
			m.mode = PNModeFilter
			m.filterFocusIdx = 0
			m.updateFilterFocus()
			return m, nil

		case "ctrl+r":
			if m.hasActiveFilter {
				m.clearFilters()
				m.msg = "✓ Filtri rimossi"
			}
			return m, nil

		case "e", "enter":
			if row := m.table.SelectedRow(); len(row) > 0 {
				id, _ := strconv.Atoi(row[0])
				m.loadIntoForm(id)
				m.mode = PNModeEdit
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

// handleFormMode gestisce gli eventi in modalità form
func (m *PrimaNotaModel) handleFormMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		// Campo descrizione: apre selezione commessa se vuoto
		if m.focusIndex == 4 && k.String() == "enter" && m.inputs[4].Value() == "" {
			m.selectionMode = true
			m.commessaFilter.SetValue("")
			m.updateCommessaTable()
			m.commessaFilter.Focus()
			return m, nil
		}

		// Campo metodo: SPAZIO per ciclare
		if m.focusIndex == 3 && k.String() == " " {
			m.currentMetodoIdx++
			if m.currentMetodoIdx >= len(MetodiPagamento) {
				m.currentMetodoIdx = 0
			}
			m.inputs[3].SetValue(MetodiPagamento[m.currentMetodoIdx])
			return m, nil
		}

		// Gestione ENTER
		if k.String() == "enter" {
			// Validazione data
			if m.focusIndex == 0 {
				dateStr := strings.TrimSpace(m.inputs[0].Value())
				if len(dateStr) != 10 {
					m.err = fmt.Errorf("data incompleta")
					return m, nil
				}
				if _, err := time.Parse("02/01/2006", dateStr); err != nil {
					m.err = fmt.Errorf("data non valida")
					return m, nil
				}
			}

			// Validazione tipo
			if m.focusIndex == 1 {
				val := strings.ToUpper(strings.TrimSpace(m.inputs[1].Value()))
				if val != "E" && val != "U" {
					m.err = fmt.Errorf("usa 'E' o 'U'")
					return m, nil
				}
				m.inputs[1].SetValue(val)
			}

			// Salvataggio (campo descrizione non vuoto)
			if m.focusIndex == 4 && m.inputs[4].Value() != "" {
				if err := m.save(); err != nil {
					m.err = err
					return m, nil
				}
				return m, nil
			}

			// Passa al campo successivo
			m.focusIndex++
			if m.focusIndex >= len(m.inputs) {
				m.focusIndex = 0
			}
			m.updateFocus()
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

		// Auto-uppercase per tipo
		if m.focusIndex == 1 {
			m.inputs[1].SetValue(strings.ToUpper(m.inputs[1].Value()))
		}
	}

	// Update inputs SOLO per il campo corrente (escluso metodo)
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		if i == m.focusIndex && i != 3 {
			m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
		}
	}

	return m, tea.Batch(cmds...)
}

// View implementa tea.Model
func (m PrimaNotaModel) View() string {
	width := 100
	if m.width > 0 {
		width = min(m.width, 115)
	}

	// Popup selezione commessa
	if m.selectionMode {
		return m.renderCommessaSelection(width)
	}

	// Dialog conferma eliminazione
	if m.showConfirm {
		return RenderConfirmDialog(
			fmt.Sprintf("Eliminare il movimento #%d?", m.deletingID),
			m.width,
			m.height,
		)
	}

	// Vista filtro
	if m.mode == PNModeFilter {
		return m.renderFilterView(width)
	}

	// Vista principale
	if m.mode == PNModeList {
		return m.renderListView(width)
	}

	return m.renderFormView(width)
}

// renderListView renderizza la vista lista
func (m PrimaNotaModel) renderListView(width int) string {
	title := "PRIMA NOTA & CASSA"
	header := RenderHeader(title, width)

	// Dashboard riepilogo
	stats := lipgloss.JoinHorizontal(
		lipgloss.Center,
		SuccessBadge("ENTRATE: "+utils.FormatEuro(m.totaleEntrate)),
		"  ",
		ErrorBadge("USCITE: "+utils.FormatEuro(m.totaleUscite)),
		"  ",
		PrimaryBadge("SALDO: "+utils.FormatEuro(m.saldo)),
	)

	filterBadge := ""
	if m.hasActiveFilter {
		filterBadge = "\n" + InfoBadge("🔍 Filtri attivi - [Ctrl+R] Rimuovi")
	}

	helpText := lipgloss.NewStyle().
		MarginTop(1).
		MarginBottom(1).
		Foreground(ColorSubText).
		Render("[N] Nuovo • [F] Filtri • [E/↵] Modifica • [X/D] Elimina • [ESC] Menu")

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		stats,
		filterBadge,
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

// renderFilterView renderizza la vista filtri
func (m PrimaNotaModel) renderFilterView(width int) string {
	title := "FILTRA MOVIMENTI"
	header := RenderHeader(title, width)

	var form strings.Builder
	labels := []string{
		"📅 Data DA",
		"📅 Data A",
		"📝 Descrizione",
		"💰 Importo",
	}

	for i := 0; i < len(m.filterInputs); i++ {
		labelStyle := LabelStyle
		if i == m.filterFocusIdx {
			labelStyle = LabelFocusedStyle
		}

		form.WriteString(fmt.Sprintf("%s %s\n",
			labelStyle.Render(labels[i]+":"),
			m.filterInputs[i].View()))
	}

	form.WriteString("\n")
	form.WriteString(HelpStyle.Render("💡 Importo: usa >, <, >=, <=, = (es: >500 o <1000)\n"))
	form.WriteString(HelpStyle.Render("[Tab/↑↓] Naviga • [↵] Applica • [Ctrl+R] Reset • [Esc] Annulla"))

	body := form.String()

	// Footer
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

// renderFormView renderizza la vista form
func (m PrimaNotaModel) renderFormView(width int) string {
	title := "NUOVO MOVIMENTO"
	if m.mode == PNModeEdit {
		title = fmt.Sprintf("MODIFICA MOVIMENTO #%d", m.selectedID)
	}

	header := RenderHeader(title, width)

	// Form
	var form strings.Builder
	labels := []string{
		"📅 Data",
		"📊 Tipo (E/U)",
		"💰 Importo €",
		"💳 Metodo",
		"📝 Descrizione",
	}

	for i := 0; i < len(m.inputs); i++ {
		labelStyle := LabelStyle
		if i == m.focusIndex {
			labelStyle = LabelFocusedStyle
		}

		view := m.inputs[i].View()

		// Campo descrizione con hint
		if i == 4 && m.focusIndex == 4 && m.inputs[4].Value() == "" {
			view = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Render("[ INVIO = Cerca Commessa | Scrivi = Libero ]")
		}

		// Campo metodo con hint
		if i == 3 && m.focusIndex == 3 {
			view = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				Render(m.inputs[3].Value()) + " " +
				HelpStyle.Render("[ SPAZIO = Cambia ]")
		}

		form.WriteString(fmt.Sprintf("%s %s\n",
			labelStyle.Render(labels[i]+":"),
			view))
		if i == 1 || i == 3 {
			form.WriteString("\n")
		}
	}

	form.WriteString("\n")
	form.WriteString(HelpStyle.Render("[Tab/↑↓] Naviga • [↵] Conferma/Prossimo • [Esc] Annulla"))

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

// renderCommessaSelection renderizza il popup selezione commessa
func (m PrimaNotaModel) renderCommessaSelection(width int) string {
	title := TitleStyle.Render("💶 SELEZIONA COMMESSA DA INCASSARE")

	filter := fmt.Sprintf("Cerca: %s", m.commessaFilter.View())

	hint := HelpStyle.Render("💡 Se importo vuoto = riempie residuo | Se compilato = mantiene importo")

	inner := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		hint,
		"",
		filter,
		"",
		m.commessaTable.View(),
		"",
		HelpStyle.Render("[↑↓] Naviga • [↵] Seleziona • [Esc] Manuale"),
	)

	box := MainBoxStyle.Copy().
		Width(min(width-10, 90)).
		Render(inner)

	if m.width > 0 && m.height > 0 {
		return CenterContent(m.width, m.height, box)
	}

	return box
}
