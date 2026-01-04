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
)

// MetodiPagamento contiene i metodi di pagamento disponibili
var MetodiPagamento = []string{"CASSA", "BANCA", "POS", "ASSEGNO", "BONIFICO"}

// PrimaNotaModel gestisce la schermata prima nota
type PrimaNotaModel struct {
	db               *database.DB
	table            table.Model
	inputs           []textinput.Model
	mode             PrimaNotaMode
	focusIndex       int
	err              error
	msg              string
	width            int
	height           int
	selectedID       int
	totaleEntrate    float64
	totaleUscite     float64
	saldo            float64
	selectionMode    bool
	commessaTable    table.Model
	commessaFilter   textinput.Model
	selectedCommID   int
	currentMetodoIdx int
	showConfirm      bool
	deletingID       int
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
	}

	m.Refresh()
	return m
}

// Refresh aggiorna la lista dei movimenti e calcola totali
func (m *PrimaNotaModel) Refresh() {
	list, _ := m.db.ListMovimenti()

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

	mov := &database.MovimentoPrimaNota{
		Data:        data,
		Tipo:        tipo,
		Importo:     importo,
		Metodo:      metodo,
		Descrizione: m.inputs[4].Value(),
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
		if m.mode != PNModeList {
			m.mode = PNModeList
			m.err = nil
			m.msg = ""
			return m, nil
		}
		return m, func() tea.Msg { return ChangeScreenMsg(StateMenu) }
	}

	// Modalità Lista
	if m.mode == PNModeList {
		return m.handleListMode(msg)
	}

	// Modalità Form (Add/Edit)
	return m.handleFormMode(msg)
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
				//totaleStr := row[4]
				versatoStr := row[5]
				residuoStr := row[6]

				//totale, _ := utils.ParseFloat(totaleStr)
				versato, _ := utils.ParseFloat(versatoStr)
				residuo, _ := utils.ParseFloat(residuoStr)

				m.selectedCommID = cID

				desc := ""
				if versato == 0 {
					desc = fmt.Sprintf("Acconto Commessa #%s - Targa %s", row[1], row[2])
				} else {
					desc = fmt.Sprintf("Saldo Commessa #%s - Targa %s", row[1], row[2])
				}

				m.inputs[4].SetValue(desc)
				m.inputs[2].SetValue(fmt.Sprintf("%.2f", residuo))
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
		return m, nil
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

		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	return m, nil
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
		if m.focusIndex == 3 {
			if k.String() == " " {
				m.currentMetodoIdx++
				if m.currentMetodoIdx >= len(MetodiPagamento) {
					m.currentMetodoIdx = 0
				}
				m.inputs[3].SetValue(MetodiPagamento[m.currentMetodoIdx])
				return m, nil
			}

			// Invio per conferma e passare avanti
			if k.String() == "enter" {
				m.focusIndex++
				m.updateFocus()
				return m, nil
			}
		}

		// Validazioni step-by-step
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

			// Salvataggio (campo descrizione)
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

	// Update inputs (escluso metodo che è gestito con logica custom)
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		if i != 3 { // Non aggiornare campo metodo
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

	helpText := lipgloss.NewStyle().
		MarginTop(1).
		MarginBottom(1).
		Foreground(ColorSubText).
		Render("[N] Nuovo • [E/↵] Modifica • [X/D] Elimina • [ESC] Menu")

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		stats,
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

	hint := HelpStyle.Render("💡 Suggerimento: Il sistema calcolerà automaticamente il residuo")

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
		HelpStyle.Render("[↑↓] Naviga • [↵] Seleziona (Calcola Residuo) • [Esc] Manuale"),
	)

	box := MainBoxStyle.Copy().
		Width(min(width-10, 90)).
		Render(inner)

	if m.width > 0 && m.height > 0 {
		return CenterContent(m.width, m.height, box)
	}

	return box
}
