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
	filterInputs           []textinput.Model
	filterFocusIdx         int
	filteredList           []database.MovimentoPrimaNota
	hasActiveFilter        bool
	fornitoreSelectionMode bool
	fornitoreTable         table.Model
	fornitoreFilter        textinput.Model
	selectedFornitoreID    int
	selectedFornitoreNome  string
}

// NewPrimaNotaModel crea una nuova istanza del model prima nota
func NewPrimaNotaModel(db *database.DB) PrimaNotaModel {
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

	inputs := make([]textinput.Model, 7)
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
	inputs[4].Placeholder = "[ INVIO = Cerca Commessa/Fornitore ] o descrizione..."
	inputs[4].Width = 60

	inputs[5] = textinput.New()
	inputs[5].Placeholder = "Numero Fattura/Ricevuta (opzionale)"
	inputs[5].Width = 40

	inputs[6] = textinput.New()
	inputs[6].Placeholder = "Data Fattura (GG/MM/AAAA) (opzionale)"
	inputs[6].CharLimit = 10
	inputs[6].Width = 30

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

	fCols := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Ragione Sociale", Width: 35},
		{Title: "Telefono", Width: 14},
	}

	ft := table.New(
		table.WithColumns(fCols),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	ft.SetStyles(GetTableStyles())

	ff := textinput.New()
	ff.Placeholder = "🔍 Cerca fornitore..."
	ff.Width = 50

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
		fornitoreTable:   ft,
		fornitoreFilter:  ff,
	}

	m.Refresh()
	return m
}

// Refresh aggiorna la lista dei movimenti e calcola totali
func (m *PrimaNotaModel) Refresh() {
	list, _ := m.db.ListMovimenti()

	if m.hasActiveFilter {
		list = m.applyFilters(list)
	}

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
		if errDa == nil && dataDa != "" {
			if mov.Data.Before(dataInizio) {
				continue
			}
		}

		if errA == nil && dataA != "" {
			if mov.Data.After(dataFine) {
				continue
			}
		}

		if descrizione != "" {
			if !strings.Contains(strings.ToUpper(mov.Descrizione), descrizione) {
				continue
			}
		}

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

	accontiMap := make(map[int]float64)
	for _, mov := range movimenti {
		if mov.CommessaID > 0 && mov.Tipo == "Entrata" {
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

// updateFornitoreTable aggiorna la tabella fornitori con filtro
func (m *PrimaNotaModel) updateFornitoreTable() {
	fornitori, _ := m.db.ListFornitori()
	filter := strings.ToUpper(strings.TrimSpace(m.fornitoreFilter.Value()))
	rows := []table.Row{}

	for _, f := range fornitori {
		ragioneSociale := strings.ToUpper(f.RagioneSociale)
		telefono := strings.ToUpper(f.Telefono)

		if filter == "" ||
			strings.Contains(ragioneSociale, filter) ||
			strings.Contains(telefono, filter) {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", f.ID),
				utils.Truncate(f.RagioneSociale, 35),
				f.Telefono,
			})
		}
	}

	m.fornitoreTable.SetRows(rows)
}

// calcolaVersatoCommessa calcola quanto già versato per una commessa
func (m *PrimaNotaModel) calcolaVersatoCommessa(commessaID int) float64 {
	movimenti, _ := m.db.ListMovimenti()
	var versato float64

	for _, mov := range movimenti {
		if mov.CommessaID == commessaID && mov.Tipo == "Entrata" {
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
	m.currentMetodoIdx = 0
	m.inputs[3].SetValue(MetodiPagamento[0])
	m.selectedCommID = 0
	m.selectedCommessaTotale = 0
	m.selectedCommessaNumero = ""
	m.selectedCommessaTarga = ""
	m.selectedFornitoreID = 0
	m.selectedFornitoreNome = ""
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
	m.selectedFornitoreID = mov.FornitoreID

	m.inputs[0].SetValue(mov.Data.Format("02/01/2006"))

	tipoChar := "E"
	if mov.Tipo == "Uscita" {
		tipoChar = "U"
	}
	m.inputs[1].SetValue(tipoChar)
	m.inputs[2].SetValue(fmt.Sprintf("%.2f", mov.Importo))

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
	m.inputs[5].SetValue(mov.NumeroFattura)

	if !mov.DataFattura.IsZero() {
		m.inputs[6].SetValue(mov.DataFattura.Format("02/01/2006"))
	}

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

	if mov.FornitoreID > 0 {
		forn, _ := m.db.GetFornitore(mov.FornitoreID)
		if forn != nil {
			m.selectedFornitoreNome = forn.RagioneSociale
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

	dataFatturaStr := strings.TrimSpace(m.inputs[6].Value())
	if dataFatturaStr != "" && len(dataFatturaStr) == 10 {
		_, err := time.Parse("02/01/2006", dataFatturaStr)
		if err != nil {
			return fmt.Errorf("data fattura non valida")
		}
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
	numeroFattura := strings.TrimSpace(m.inputs[5].Value())

	var dataFattura time.Time
	dataFatturaStr := strings.TrimSpace(m.inputs[6].Value())
	if dataFatturaStr != "" && len(dataFatturaStr) == 10 {
		dataFattura, _ = time.Parse("02/01/2006", dataFatturaStr)
	}

	if tipo == "Entrata" && m.selectedCommID > 0 {
		versato := m.calcolaVersatoCommessa(m.selectedCommID)
		totaleConQuesto := versato + importo

		numero := m.selectedCommessaNumero
		targa := m.selectedCommessaTarga
		totale := m.selectedCommessaTotale

		if numero == "" {
			if comm, _ := m.db.GetCommessa(m.selectedCommID); comm != nil {
				totale = comm.Totale
				numero = comm.Numero
				if v, _ := m.db.GetVeicolo(comm.VeicoloID); v != nil {
					targa = v.Targa
				}
			}
		}

		if totaleConQuesto >= totale {
			descrizione = fmt.Sprintf("Saldo Commessa #%s - Targa %s", numero, targa)
		} else {
			descrizione = fmt.Sprintf("Acconto Commessa #%s - Targa %s", numero, targa)
		}
	} else if tipo == "Uscita" && m.selectedFornitoreID > 0 {
		fornitoreNome := m.selectedFornitoreNome
		if fornitoreNome == "" {
			if forn, _ := m.db.GetFornitore(m.selectedFornitoreID); forn != nil {
				fornitoreNome = forn.RagioneSociale
			}
		}

		if numeroFattura != "" {
			descrizione = fmt.Sprintf("Fornitura da %s - Fatt. %s", fornitoreNome, numeroFattura)
		} else {
			descrizione = fmt.Sprintf("Fornitura da %s", fornitoreNome)
		}
	}

	mov := &database.MovimentoPrimaNota{
		Data:          data,
		Tipo:          tipo,
		Importo:       importo,
		Metodo:        metodo,
		Descrizione:   descrizione,
		CommessaID:    m.selectedCommID,
		FornitoreID:   m.selectedFornitoreID,
		NumeroFattura: numeroFattura,
		DataFattura:   dataFattura,
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
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	if m.selectionMode {
		return m.handleCommessaSelection(msg)
	}

	if m.fornitoreSelectionMode {
		return m.handleFornitoreSelection(msg)
	}

	if m.showConfirm {
		return m.handleDeleteConfirmation(msg)
	}

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

	if m.mode == PNModeFilter {
		return m.handleFilterMode(msg)
	}

	if m.mode == PNModeList {
		return m.handleListMode(msg)
	}

	return m.handleFormMode(msg)
}

// handleCommessaSelection gestisce la selezione commessa
func (m PrimaNotaModel) handleCommessaSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.selectionMode = false
			m.updateFocus()
			return m, nil
		case "enter":
			if row := m.commessaTable.SelectedRow(); len(row) > 0 {
				m.selectedCommID, _ = strconv.Atoi(row[0])
				m.selectedCommessaNumero = row[1]
				m.selectedCommessaTarga = row[2]
				totaleStr := strings.ReplaceAll(row[4], "€", "")
				totaleStr = strings.ReplaceAll(totaleStr, ".", "")
				totaleStr = strings.ReplaceAll(totaleStr, ",", ".")
				totaleStr = strings.TrimSpace(totaleStr)
				m.selectedCommessaTotale, _ = utils.ParseFloat(totaleStr)

				m.inputs[4].SetValue(fmt.Sprintf("Commessa #%s - %s", m.selectedCommessaNumero, m.selectedCommessaTarga))
				m.selectionMode = false
				m.focusIndex = 0
				m.updateFocus()
			}
			return m, nil
		}
	}

	var cmdF, cmdT tea.Cmd
	m.commessaFilter, cmdF = m.commessaFilter.Update(msg)
	m.updateCommessaTable()
	m.commessaTable, cmdT = m.commessaTable.Update(msg)
	return m, tea.Batch(cmdF, cmdT)
}

// handleFornitoreSelection gestisce la selezione fornitore
func (m PrimaNotaModel) handleFornitoreSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.fornitoreSelectionMode = false
			m.updateFocus()
			return m, nil
		case "enter":
			if row := m.fornitoreTable.SelectedRow(); len(row) > 0 {
				m.selectedFornitoreID, _ = strconv.Atoi(row[0])
				m.selectedFornitoreNome = row[1]

				m.inputs[4].SetValue(fmt.Sprintf("Fornitore: %s", m.selectedFornitoreNome))
				m.fornitoreSelectionMode = false
				m.focusIndex = 0
				m.updateFocus()
			}
			return m, nil
		}
	}

	var cmdF, cmdT tea.Cmd
	m.fornitoreFilter, cmdF = m.fornitoreFilter.Update(msg)
	m.updateFornitoreTable()
	m.fornitoreTable, cmdT = m.fornitoreTable.Update(msg)
	return m, tea.Batch(cmdF, cmdT)
}

// handleDeleteConfirmation gestisce la conferma eliminazione
func (m PrimaNotaModel) handleDeleteConfirmation(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "y", "Y":
			if err := m.db.DeleteMovimento(m.deletingID); err != nil {
				m.err = fmt.Errorf("errore eliminazione: %w", err)
			} else {
				m.msg = "✓ Movimento eliminato"
			}
			m.Refresh()
			m.showConfirm = false
		case "n", "N", "esc":
			m.showConfirm = false
		}
	}
	return m, nil
}

// handleFilterMode gestisce la modalità filtro
func (m PrimaNotaModel) handleFilterMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "enter":
			m.hasActiveFilter = true
			m.mode = PNModeList
			m.Refresh()
			return m, nil
		case "tab", "down":
			m.filterFocusIdx++
			if m.filterFocusIdx >= len(m.filterInputs) {
				m.filterFocusIdx = 0
			}
			m.updateFilterFocus()
			return m, nil
		case "shift+tab", "up":
			m.filterFocusIdx--
			if m.filterFocusIdx < 0 {
				m.filterFocusIdx = len(m.filterInputs) - 1
			}
			m.updateFilterFocus()
			return m, nil
		case "ctrl+r":
			m.clearFilters()
			m.mode = PNModeList
			return m, nil
		}
	}

	cmds := make([]tea.Cmd, len(m.filterInputs))
	for i := range m.filterInputs {
		m.filterInputs[i], cmds[i] = m.filterInputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

// handleListMode gestisce la modalità lista
func (m PrimaNotaModel) handleListMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			m.mode = PNModeAdd
			m.resetForm()
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
		case "f":
			m.mode = PNModeFilter
			m.filterFocusIdx = 0
			m.updateFilterFocus()
			return m, nil
		case "ctrl+r":
			m.clearFilters()
			return m, nil
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// handleFormMode gestisce la modalità form
func (m PrimaNotaModel) handleFormMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		tipoInput := strings.ToUpper(strings.TrimSpace(m.inputs[1].Value()))

		if m.focusIndex == 4 && k.String() == "enter" {
			if tipoInput == "E" {
				m.selectionMode = true
				m.commessaFilter.SetValue("")
				m.updateCommessaTable()
				m.commessaFilter.Focus()
				return m, nil
			} else if tipoInput == "U" {
				m.fornitoreSelectionMode = true
				m.fornitoreFilter.SetValue("")
				m.updateFornitoreTable()
				m.fornitoreFilter.Focus()
				return m, nil
			}
		}

		if m.focusIndex == 3 && k.String() == " " {
			m.currentMetodoIdx++
			if m.currentMetodoIdx >= len(MetodiPagamento) {
				m.currentMetodoIdx = 0
			}
			m.inputs[3].SetValue(MetodiPagamento[m.currentMetodoIdx])
			return m, nil
		}

		switch k.String() {
		case "enter":
			if m.focusIndex == len(m.inputs)-1 {
				if err := m.save(); err != nil {
					m.err = err
					return m, nil
				}
				return m, nil
			}
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

	if m.focusIndex == 1 {
		m.inputs[1].SetValue(strings.ToUpper(m.inputs[1].Value()))
	}

	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

// View implementa tea.Model
func (m PrimaNotaModel) View() string {
	width := 100
	if m.width > 0 {
		width = min(m.width, 110)
	}

	if m.showConfirm {
		var message strings.Builder
		message.WriteString(fmt.Sprintf("⚠️  ELIMINAZIONE MOVIMENTO #%d\n\n", m.deletingID))
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
		title := RenderHeader("SELEZIONA COMMESSA", width)
		filterView := lipgloss.NewStyle().
			MarginBottom(1).
			Render(m.commessaFilter.View())

		body := lipgloss.JoinVertical(
			lipgloss.Left,
			filterView,
			m.commessaTable.View(),
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

	if m.fornitoreSelectionMode {
		title := RenderHeader("SELEZIONA FORNITORE", width)
		filterView := lipgloss.NewStyle().
			MarginBottom(1).
			Render(m.fornitoreFilter.View())

		body := lipgloss.JoinVertical(
			lipgloss.Left,
			filterView,
			m.fornitoreTable.View(),
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

	title := "PRIMA NOTA & CASSA"
	if m.mode == PNModeAdd {
		title = "NUOVO MOVIMENTO"
	} else if m.mode == PNModeEdit {
		title = fmt.Sprintf("MODIFICA MOVIMENTO #%d", m.selectedID)
	} else if m.mode == PNModeFilter {
		title = "FILTRI PRIMA NOTA"
	}

	header := RenderHeader(title, width)
	var body string

	if m.mode == PNModeList {
		filterStatus := ""
		if m.hasActiveFilter {
			filterStatus = WarningBadge(" FILTRI ATTIVI ") + " "
		}

		helpText := lipgloss.NewStyle().
			MarginBottom(1).
			Foreground(ColorSubText).
			Render(filterStatus + "[N] Nuovo • [E/↵] Modifica • [X/D] Elimina • [F] Filtri • [Ctrl+R] Reset Filtri • [ESC] Menu")

		statsLine := fmt.Sprintf("💰 Totale Entrate: %s | Totale Uscite: %s | Saldo: %s",
			utils.FormatEuro(m.totaleEntrate),
			utils.FormatEuro(m.totaleUscite),
			utils.FormatEuro(m.saldo))

		saldoStyle := SuccessStyle
		if m.saldo < 0 {
			saldoStyle = ErrorStyle
		}

		body = lipgloss.JoinVertical(
			lipgloss.Left,
			helpText,
			m.table.View(),
			"",
			saldoStyle.Render(statsLine),
		)
	} else if m.mode == PNModeFilter {
		var form strings.Builder
		labels := []string{"Data DA", "Data A", "Descrizione", "Importo"}

		for i, inp := range m.filterInputs {
			labelStyle := LabelStyle
			if i == m.filterFocusIdx {
				labelStyle = LabelFocusedStyle
			}

			form.WriteString(fmt.Sprintf("%s %s\n",
				labelStyle.Render(labels[i]+":"),
				inp.View()))
		}

		form.WriteString("\n")
		form.WriteString(HelpStyle.Render("[Tab/↑↓] Naviga • [↵] Applica Filtri • [Ctrl+R] Reset • [Esc] Annulla"))
		body = form.String()
	} else {
		var form strings.Builder
		labels := []string{"Data", "Tipo", "Importo", "Metodo", "Descrizione/Commessa/Fornitore", "N. Fattura/Ricevuta", "Data Fattura"}

		tipoInput := strings.ToUpper(strings.TrimSpace(m.inputs[1].Value()))

		for i, inp := range m.inputs {
			if i == 5 || i == 6 {
				if tipoInput != "U" {
					continue
				}
			}

			labelStyle := LabelStyle
			if i == m.focusIndex {
				labelStyle = LabelFocusedStyle
			}

			form.WriteString(fmt.Sprintf("%s %s\n",
				labelStyle.Render(labels[i]+":"),
				inp.View()))

			if i == 2 {
				form.WriteString("\n")
			}
		}

		form.WriteString("\n")

		if tipoInput == "E" {
			form.WriteString(HelpStyle.Render("[Tab/↑↓] Naviga • Campo Descrizione: [↵] Seleziona Commessa • [Esc] Annulla"))
		} else if tipoInput == "U" {
			form.WriteString(HelpStyle.Render("[Tab/↑↓] Naviga • Campo Descrizione: [↵] Seleziona Fornitore • [Esc] Annulla"))
		} else {
			form.WriteString(HelpStyle.Render("[Tab/↑↓] Naviga • [↵] Conferma/Prossimo • [Esc] Annulla"))
		}

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
