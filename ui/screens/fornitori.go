package screens

import (
	"fmt"
	"officina/database"
	"officina/utils"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FornitoreMode rappresenta la modalità corrente della schermata
type FornitoreMode int

const (
	FornList FornitoreMode = iota
	FornAdd
	FornEdit
)

// FornitoriModel gestisce la schermata fornitori
type FornitoriModel struct {
	db                     *database.DB
	table                  table.Model
	inputs                 []textinput.Model
	mode                   FornitoreMode
	focusIndex             int
	selectedID             int
	err                    error
	msg                    string
	width                  int
	height                 int
	showConfirm            bool
	deletingID             int
	deleteWarningMovimenti int
	deleteWarningTotale    float64
}

// NewFornitoriModel crea una nuova istanza del model fornitori
func NewFornitoriModel(db *database.DB) FornitoriModel {
	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Ragione Sociale", Width: 35},
		{Title: "Telefono", Width: 14},
		{Title: "Email", Width: 25},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(12),
	)

	t.SetStyles(GetTableStyles())

	inputs := make([]textinput.Model, 11)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Ragione Sociale"
	inputs[0].Width = 50

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Telefono (es. 06 12345678)"
	inputs[1].Width = 40

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Email (opzionale)"
	inputs[2].Width = 40

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "PEC (opzionale)"
	inputs[3].Width = 40

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "Codice Fiscale (opzionale)"
	inputs[4].CharLimit = 16
	inputs[4].Width = 40

	inputs[5] = textinput.New()
	inputs[5].Placeholder = "Partita IVA (opzionale)"
	inputs[5].CharLimit = 11
	inputs[5].Width = 40

	inputs[6] = textinput.New()
	inputs[6].Placeholder = "Codice SDI (opzionale)"
	inputs[6].CharLimit = 7
	inputs[6].Width = 40

	inputs[7] = textinput.New()
	inputs[7].Placeholder = "Indirizzo (via, numero)"
	inputs[7].Width = 50

	inputs[8] = textinput.New()
	inputs[8].Placeholder = "CAP"
	inputs[8].CharLimit = 5
	inputs[8].Width = 20

	inputs[9] = textinput.New()
	inputs[9].Placeholder = "Città"
	inputs[9].Width = 40

	inputs[10] = textinput.New()
	inputs[10].Placeholder = "Provincia (es. RM)"
	inputs[10].CharLimit = 2
	inputs[10].Width = 10

	m := FornitoriModel{
		db:     db,
		table:  t,
		inputs: inputs,
		mode:   FornList,
	}

	m.Refresh()
	return m
}

// Refresh aggiorna la lista dei fornitori
func (m *FornitoriModel) Refresh() {
	list, _ := m.db.ListFornitori()
	rows := []table.Row{}

	for _, f := range list {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", f.ID),
			utils.Truncate(f.RagioneSociale, 35),
			f.Telefono,
			utils.Truncate(f.Email, 25),
		})
	}

	m.table.SetRows(rows)
}

// countDataForFornitore conta movimenti associati a un fornitore
func (m *FornitoriModel) countDataForFornitore(fornitoreID int) (int, float64) {
	movimenti, _ := m.db.ListMovimenti()

	numMovimenti := 0
	totaleMov := 0.0
	for _, mov := range movimenti {
		if mov.FornitoreID == fornitoreID {
			numMovimenti++
			if mov.Tipo == "Uscita" {
				totaleMov += mov.Importo
			}
		}
	}

	return numMovimenti, totaleMov
}

// resetForm resetta il form ai valori predefiniti
func (m *FornitoriModel) resetForm() {
	for i := range m.inputs {
		m.inputs[i].SetValue("")
	}

	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.inputs[0].Focus()
}

// loadIntoForm carica un fornitore nel form
func (m *FornitoriModel) loadIntoForm(id int) {
	f, err := m.db.GetFornitore(id)
	if err != nil {
		m.err = fmt.Errorf("errore caricamento fornitore: %w", err)
		return
	}

	m.selectedID = id
	m.inputs[0].SetValue(f.RagioneSociale)
	m.inputs[1].SetValue(f.Telefono)
	m.inputs[2].SetValue(f.Email)
	m.inputs[3].SetValue(f.PEC)
	m.inputs[4].SetValue(f.CodiceFiscale)
	m.inputs[5].SetValue(f.PartitaIVA)
	m.inputs[6].SetValue(f.CodiceDestinatario)
	m.inputs[7].SetValue(f.Indirizzo)
	m.inputs[8].SetValue(f.CAP)
	m.inputs[9].SetValue(f.Citta)
	m.inputs[10].SetValue(f.Provincia)

	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.inputs[0].Focus()
}

// updateFocus aggiorna il focus tra i campi
func (m *FornitoriModel) updateFocus() {
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

// validate valida i dati del form
func (m *FornitoriModel) validate() error {
	if err := utils.ValidateNotEmpty(m.inputs[0].Value(), "Ragione Sociale"); err != nil {
		return err
	}

	if err := utils.ValidateEmail(m.inputs[2].Value()); err != nil {
		return err
	}

	if err := utils.ValidateEmail(m.inputs[3].Value()); err != nil {
		return fmt.Errorf("PEC non valida: %w", err)
	}

	if err := utils.ValidateCodiceFiscale(m.inputs[4].Value()); err != nil {
		return err
	}

	if err := utils.ValidatePartitaIVA(m.inputs[5].Value()); err != nil {
		return err
	}

	if err := utils.ValidateCAP(m.inputs[8].Value()); err != nil {
		return err
	}

	return nil
}

// save salva il fornitore corrente
func (m *FornitoriModel) save() error {
	if err := m.validate(); err != nil {
		return err
	}

	f := &database.Fornitore{
		RagioneSociale:     m.inputs[0].Value(),
		Telefono:           m.inputs[1].Value(),
		Email:              m.inputs[2].Value(),
		PEC:                m.inputs[3].Value(),
		CodiceFiscale:      strings.ToUpper(m.inputs[4].Value()),
		PartitaIVA:         m.inputs[5].Value(),
		CodiceDestinatario: strings.ToUpper(m.inputs[6].Value()),
		Indirizzo:          m.inputs[7].Value(),
		CAP:                m.inputs[8].Value(),
		Citta:              m.inputs[9].Value(),
		Provincia:          strings.ToUpper(m.inputs[10].Value()),
	}

	if m.mode == FornAdd {
		if err := m.db.CreateFornitore(f); err != nil {
			return fmt.Errorf("errore creazione: %w", err)
		}
		m.msg = "✓ Fornitore creato con successo"
	} else {
		f.ID = m.selectedID
		if err := m.db.UpdateFornitore(f); err != nil {
			return fmt.Errorf("errore aggiornamento: %w", err)
		}
		m.msg = "✓ Fornitore aggiornato con successo"
	}

	m.mode = FornList
	m.Refresh()
	return nil
}

// Init implementa tea.Model
func (m FornitoriModel) Init() tea.Cmd {
	return nil
}

// Update implementa tea.Model
func (m FornitoriModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	if m.showConfirm {
		if k, ok := msg.(tea.KeyMsg); ok {
			switch k.String() {
			case "y", "Y":
				if err := m.db.DeleteFornitore(m.deletingID); err != nil {
					m.err = fmt.Errorf("errore eliminazione: %w", err)
				} else {
					m.msg = fmt.Sprintf("✓ Fornitore e %d movimenti eliminati",
						m.deleteWarningMovimenti)
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
			return m, nil
		}
	}

	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		if m.mode != FornList {
			m.mode = FornList
			m.err = nil
			m.msg = ""
			return m, nil
		}
		return m, func() tea.Msg { return ChangeScreenMsg(StateMenu) }
	}

	if m.mode == FornList {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "n":
				m.mode = FornAdd
				m.resetForm()
				return m, nil
			case "e", "enter":
				if row := m.table.SelectedRow(); len(row) > 0 {
					id, _ := strconv.Atoi(row[0])
					m.loadIntoForm(id)
					m.mode = FornEdit
				}
				return m, nil
			case "x", "d":
				if row := m.table.SelectedRow(); len(row) > 0 {
					id, _ := strconv.Atoi(row[0])
					m.deletingID = id
					numMovimenti, totaleMov := m.countDataForFornitore(id)
					m.deleteWarningMovimenti = numMovimenti
					m.deleteWarningTotale = totaleMov
					m.showConfirm = true
				}
				return m, nil
			}
		}
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	if m.mode == FornAdd || m.mode == FornEdit {
		if k, ok := msg.(tea.KeyMsg); ok {
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

		if m.focusIndex == 4 || m.focusIndex == 6 || m.focusIndex == 10 {
			m.inputs[m.focusIndex].SetValue(strings.ToUpper(m.inputs[m.focusIndex].Value()))
		}

		cmds := make([]tea.Cmd, len(m.inputs))
		for i := range m.inputs {
			m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
		}
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

// renderDeleteConfirmation renderizza il dialog di conferma eliminazione con avviso
func (m FornitoriModel) renderDeleteConfirmation(width int) string {
	var message strings.Builder
	f, _ := m.db.GetFornitore(m.deletingID)
	nome := "???"
	if f != nil {
		nome = f.RagioneSociale
	}

	message.WriteString(fmt.Sprintf("⚠️  ELIMINAZIONE FORNITORE #%d\n", m.deletingID))
	message.WriteString(fmt.Sprintf("%s\n\n", nome))

	if m.deleteWarningMovimenti > 0 {
		message.WriteString(ErrorStyle.Render(fmt.Sprintf(
			"ATTENZIONE: Questo fornitore ha dati associati!\n\n"+
				"Eliminando il fornitore verranno eliminati:\n"+
				" • %d movimenti di Prima Nota (totale: %s)\n\n"+
				"TUTTI I DATI VERRANNO PERSI IN MODO PERMANENTE!\n\n",
			m.deleteWarningMovimenti,
			utils.FormatEuro(m.deleteWarningTotale),
		)))
	} else {
		message.WriteString("Questo fornitore non ha movimenti associati.\n\n")
	}

	message.WriteString(WarningStyle.Render("Sei sicuro di voler procedere?\n"))
	message.WriteString(HelpStyle.Render("\n[Y] Sì, elimina TUTTO • [N/Esc] Annulla"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorError).
		Padding(1, 2).
		Width(75).
		Render(message.String())

	if m.width > 0 && m.height > 0 {
		return CenterContent(m.width, m.height, box)
	}

	return box
}

// View implementa tea.Model
func (m FornitoriModel) View() string {
	width := 85
	if m.width > 0 {
		width = min(m.width, 100)
	}

	if m.showConfirm {
		return m.renderDeleteConfirmation(width)
	}

	title := "GESTIONE FORNITORI"
	if m.mode == FornAdd {
		title = "NUOVO FORNITORE"
	} else if m.mode == FornEdit {
		title = fmt.Sprintf("MODIFICA FORNITORE #%d", m.selectedID)
	}

	header := RenderHeader(title, width)
	var body string

	if m.mode == FornList {
		helpText := lipgloss.NewStyle().
			MarginBottom(1).
			Foreground(ColorSubText).
			Render("[N] Nuovo • [E/↵] Modifica • [X/D] Elimina • [ESC] Menu")

		body = lipgloss.JoinVertical(
			lipgloss.Left,
			helpText,
			m.table.View(),
		)
	} else {
		var form strings.Builder
		labels := []string{
			"Ragione Sociale", "Telefono", "Email",
			"PEC", "Cod. Fiscale", "P.IVA", "SDI",
			"Indirizzo", "CAP", "Città", "Provincia",
		}

		for i, inp := range m.inputs {
			labelStyle := LabelStyle
			if i == m.focusIndex {
				labelStyle = LabelFocusedStyle
			}

			form.WriteString(fmt.Sprintf("%s %s\n",
				labelStyle.Render(labels[i]+":"),
				inp.View()))

			if i == 2 || i == 6 {
				form.WriteString("\n")
			}
		}

		form.WriteString("\n")
		form.WriteString(HelpStyle.Render("[Tab/↑↓] Naviga • [↵] Conferma/Prossimo • [Esc] Annulla"))
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
