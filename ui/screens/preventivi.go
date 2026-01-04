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

// PreventiviMode rappresenta la modalità corrente
type PreventiviMode int

const (
	PrevModeList PreventiviMode = iota
	PrevModeAdd
	PrevModeEdit
)

// PreventiviModel gestisce la schermata preventivi
type PreventiviModel struct {
	db          *database.DB
	table       table.Model
	inputs      []textinput.Model
	mode        PreventiviMode
	focusIndex  int
	selectedID  int
	err         error
	msg         string
	width       int
	height      int
	showConfirm bool
	deletingID  int
}

// NewPreventiviModel crea una nuova istanza del model preventivi
func NewPreventiviModel(db *database.DB) PreventiviModel {
	// Configurazione tabella
	t := table.New(
		table.WithColumns([]table.Column{
			{Title: "ID", Width: 4},
			{Title: "Numero", Width: 12},
			{Title: "Cliente", Width: 35},
			{Title: "Totale", Width: 12},
			{Title: "Stato", Width: 10},
		}),
		table.WithHeight(12),
		table.WithFocused(true),
	)
	t.SetStyles(GetTableStyles())

	// Configurazione inputs
	inputs := make([]textinput.Model, 3)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Cliente (nome o ragione sociale)"
	inputs[0].Width = 50

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Totale (es. 1250.50)"
	inputs[1].Width = 30

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Descrizione lavori"
	inputs[2].Width = 60

	m := PreventiviModel{
		db:     db,
		table:  t,
		inputs: inputs,
		mode:   PrevModeList,
	}

	m.Refresh()
	return m
}

// Refresh aggiorna la lista dei preventivi
func (m *PreventiviModel) Refresh() {
	list, _ := m.db.ListPreventivi()
	rows := []table.Row{}

	for _, p := range list {
		stato := "Aperto"
		if p.Accettato {
			stato = "Accettato"
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("%d", p.ID),
			p.Numero,
			utils.Truncate(p.Cliente, 35),
			utils.FormatEuro(p.Totale),
			stato,
		})
	}

	m.table.SetRows(rows)
}

// resetForm resetta il form
func (m *PreventiviModel) resetForm() {
	for i := range m.inputs {
		m.inputs[i].SetValue("")
	}

	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.inputs[0].Focus()
}

// loadIntoForm carica un preventivo nel form
func (m *PreventiviModel) loadIntoForm(id int) {
	p, err := m.db.GetPreventivo(id)
	if err != nil {
		m.err = fmt.Errorf("errore caricamento preventivo: %w", err)
		return
	}

	m.selectedID = id
	m.inputs[0].SetValue(p.Cliente)
	m.inputs[1].SetValue(fmt.Sprintf("%.2f", p.Totale))
	m.inputs[2].SetValue(p.Descrizione)

	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.inputs[0].Focus()
}

// updateFocus aggiorna il focus tra i campi
func (m *PreventiviModel) updateFocus() {
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

// validate valida i dati del form
func (m *PreventiviModel) validate() error {
	if err := utils.ValidateNotEmpty(m.inputs[0].Value(), "Cliente"); err != nil {
		return err
	}

	totale, err := utils.ParseFloat(m.inputs[1].Value())
	if err != nil {
		return fmt.Errorf("importo non valido")
	}

	if err := utils.ValidateImportoPositivo(totale); err != nil {
		return err
	}

	return nil
}

// save salva il preventivo corrente
func (m *PreventiviModel) save() error {
	if err := m.validate(); err != nil {
		return err
	}

	totale, _ := utils.ParseFloat(m.inputs[1].Value())

	p := &database.Preventivo{
		Cliente:     m.inputs[0].Value(),
		Totale:      totale,
		Descrizione: m.inputs[2].Value(),
		Accettato:   false,
	}

	if m.mode == PrevModeAdd {
		if err := m.db.CreatePreventivo(p); err != nil {
			return fmt.Errorf("errore creazione: %w", err)
		}
		m.msg = "✓ Preventivo creato con successo"
	} else {
		p.ID = m.selectedID

		// Mantieni lo stato di accettazione esistente
		old, _ := m.db.GetPreventivo(m.selectedID)
		if old != nil {
			p.Numero = old.Numero
			p.Accettato = old.Accettato
			p.Data = old.Data
		}

		if err := m.db.UpdatePreventivo(p); err != nil {
			return fmt.Errorf("errore aggiornamento: %w", err)
		}
		m.msg = "✓ Preventivo aggiornato con successo"
	}

	m.mode = PrevModeList
	m.Refresh()
	return nil
}

// Init implementa tea.Model
func (m PreventiviModel) Init() tea.Cmd {
	return nil
}

// Update implementa tea.Model
func (m PreventiviModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Gestione resize
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	// Conferma eliminazione
	if m.showConfirm {
		if k, ok := msg.(tea.KeyMsg); ok {
			switch k.String() {
			case "y", "Y":
				if err := m.db.DeletePreventivo(m.deletingID); err != nil {
					m.err = fmt.Errorf("errore eliminazione: %w", err)
				} else {
					m.msg = "✓ Preventivo eliminato con successo"
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
		if m.mode != PrevModeList {
			m.mode = PrevModeList
			m.err = nil
			m.msg = ""
			return m, nil
		}
		return m, func() tea.Msg { return ChangeScreenMsg(StateMenu) }
	}

	// Modalità Lista
	if m.mode == PrevModeList {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "n":
				m.mode = PrevModeAdd
				m.resetForm()
				return m, nil

			case "e", "enter":
				if row := m.table.SelectedRow(); len(row) > 0 {
					id, _ := strconv.Atoi(row[0])
					m.loadIntoForm(id)
					m.mode = PrevModeEdit
				}
				return m, nil

			case "x", "d":
				if row := m.table.SelectedRow(); len(row) > 0 {
					id, _ := strconv.Atoi(row[0])
					m.deletingID = id
					m.showConfirm = true
				}
				return m, nil

			case "a":
				// Accetta preventivo
				if row := m.table.SelectedRow(); len(row) > 0 {
					id, _ := strconv.Atoi(row[0])
					p, _ := m.db.GetPreventivo(id)
					if p != nil {
						p.Accettato = !p.Accettato
						m.db.UpdatePreventivo(p)
						m.Refresh()
						if p.Accettato {
							m.msg = "✓ Preventivo accettato"
						} else {
							m.msg = "✓ Preventivo riaperto"
						}
					}
				}
				return m, nil
			}

			m.table, cmd = m.table.Update(msg)
			return m, cmd
		}
	}

	// Modalità Form (Add/Edit)
	if m.mode == PrevModeAdd || m.mode == PrevModeEdit {
		if k, ok := msg.(tea.KeyMsg); ok {
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

		// Update inputs
		cmds := make([]tea.Cmd, len(m.inputs))
		for i := range m.inputs {
			m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
		}
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

// View implementa tea.Model
func (m PreventiviModel) View() string {
	width := 85
	if m.width > 0 {
		width = min(m.width, 95)
	}

	// Titolo dinamico
	title := "GESTIONE PREVENTIVI"
	if m.mode == PrevModeAdd {
		title = "NUOVO PREVENTIVO"
	} else if m.mode == PrevModeEdit {
		title = fmt.Sprintf("MODIFICA PREVENTIVO #%d", m.selectedID)
	}

	header := RenderHeader(title, width)

	var body string

	// Dialog conferma eliminazione
	if m.showConfirm {
		body = RenderConfirmDialog(
			fmt.Sprintf("Eliminare il preventivo #%d?", m.deletingID),
			width,
			0,
		)
	} else if m.mode == PrevModeList {
		// Vista lista
		helpText := lipgloss.NewStyle().
			MarginBottom(1).
			Foreground(ColorSubText).
			Render("[N] Nuovo • [E/↵] Modifica • [A] Accetta/Riapri • [X/D] Elimina • [ESC] Menu")

		body = lipgloss.JoinVertical(
			lipgloss.Left,
			helpText,
			m.table.View(),
		)
	} else {
		// Vista form
		var form strings.Builder
		labels := []string{"Cliente", "Totale €", "Descrizione"}

		for i, inp := range m.inputs {
			labelStyle := LabelStyle
			if i == m.focusIndex {
				labelStyle = LabelFocusedStyle
			}

			form.WriteString(fmt.Sprintf("%s %s\n",
				labelStyle.Render(labels[i]+":"),
				inp.View()))

			if i == 1 {
				form.WriteString("\n")
			}
		}

		form.WriteString("\n")
		form.WriteString(HelpStyle.Render("[Tab/↑↓] Naviga • [↵] Conferma/Prossimo • [Esc] Annulla"))

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
