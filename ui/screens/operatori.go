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

// OperatoriMode rappresenta la modalità corrente
type OperatoriMode int

const (
	OpModeList OperatoriMode = iota
	OpModeAdd
	OpModeEdit
)

// OperatoriModel gestisce la schermata operatori
type OperatoriModel struct {
	db          *database.DB
	table       table.Model
	inputs      []textinput.Model
	mode        OperatoriMode
	focusIndex  int
	selectedID  int
	err         error
	msg         string
	width       int
	height      int
	showConfirm bool
	deletingID  int
}

// NewOperatoriModel crea una nuova istanza del model operatori
func NewOperatoriModel(db *database.DB) OperatoriModel {
	// Configurazione tabella
	cols := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Matricola", Width: 10},
		{Title: "Nome Completo", Width: 30},
		{Title: "Ruolo", Width: 25},
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithHeight(12),
		table.WithFocused(true),
	)
	t.SetStyles(GetTableStyles())

	// Configurazione inputs
	inputs := make([]textinput.Model, 4)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Matricola (es. OPR001)"
	inputs[0].Width = 40

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Nome"
	inputs[1].Width = 40

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Cognome"
	inputs[2].Width = 40

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Ruolo (es. Meccanico, Carrozziere)"
	inputs[3].Width = 40

	m := OperatoriModel{
		db:     db,
		table:  t,
		inputs: inputs,
		mode:   OpModeList,
	}

	m.Refresh()
	return m
}

// Refresh aggiorna la lista degli operatori
func (m *OperatoriModel) Refresh() {
	list, _ := m.db.ListOperatori()
	rows := []table.Row{}

	for _, o := range list {
		nomeCompleto := fmt.Sprintf("%s %s", o.Nome, o.Cognome)
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", o.ID),
			o.Matricola,
			utils.Truncate(nomeCompleto, 30),
			utils.Truncate(o.Ruolo, 25),
		})
	}

	m.table.SetRows(rows)
}

// resetForm resetta il form
func (m *OperatoriModel) resetForm() {
	for i := range m.inputs {
		m.inputs[i].SetValue("")
	}

	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.inputs[0].Focus()
}

// loadIntoForm carica un operatore nel form
func (m *OperatoriModel) loadIntoForm(id int) {
	o, err := m.db.GetOperatore(id)
	if err != nil {
		m.err = fmt.Errorf("errore caricamento operatore: %w", err)
		return
	}

	m.selectedID = id
	m.inputs[0].SetValue(o.Matricola)
	m.inputs[1].SetValue(o.Nome)
	m.inputs[2].SetValue(o.Cognome)
	m.inputs[3].SetValue(o.Ruolo)

	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.inputs[0].Focus()
}

// updateFocus aggiorna il focus tra i campi
func (m *OperatoriModel) updateFocus() {
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

// validate valida i dati del form
func (m *OperatoriModel) validate() error {
	if err := utils.ValidateNotEmpty(m.inputs[0].Value(), "Matricola"); err != nil {
		return err
	}

	if err := utils.ValidateNotEmpty(m.inputs[1].Value(), "Nome"); err != nil {
		return err
	}

	if err := utils.ValidateNotEmpty(m.inputs[2].Value(), "Cognome"); err != nil {
		return err
	}

	if err := utils.ValidateNotEmpty(m.inputs[3].Value(), "Ruolo"); err != nil {
		return err
	}

	return nil
}

// save salva l'operatore corrente
func (m *OperatoriModel) save() error {
	if err := m.validate(); err != nil {
		return err
	}

	o := &database.Operatore{
		Matricola: strings.ToUpper(m.inputs[0].Value()),
		Nome:      m.inputs[1].Value(),
		Cognome:   m.inputs[2].Value(),
		Ruolo:     m.inputs[3].Value(),
	}

	if m.mode == OpModeAdd {
		if err := m.db.CreateOperatore(o); err != nil {
			return fmt.Errorf("errore creazione: %w", err)
		}
		m.msg = "✓ Operatore creato con successo"
	} else {
		o.ID = m.selectedID
		if err := m.db.UpdateOperatore(o); err != nil {
			return fmt.Errorf("errore aggiornamento: %w", err)
		}
		m.msg = "✓ Operatore aggiornato con successo"
	}

	m.mode = OpModeList
	m.Refresh()
	return nil
}

// Init implementa tea.Model
func (m OperatoriModel) Init() tea.Cmd {
	return nil
}

// Update implementa tea.Model
func (m OperatoriModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				if err := m.db.DeleteOperatore(m.deletingID); err != nil {
					m.err = fmt.Errorf("errore eliminazione: %w", err)
				} else {
					m.msg = "✓ Operatore eliminato con successo"
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
		if m.mode != OpModeList {
			m.mode = OpModeList
			m.err = nil
			m.msg = ""
			return m, nil
		}
		return m, func() tea.Msg { return ChangeScreenMsg(StateMenu) }
	}

	// Modalità Lista
	if m.mode == OpModeList {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "n":
				m.mode = OpModeAdd
				m.resetForm()
				return m, nil

			case "e", "enter":
				if row := m.table.SelectedRow(); len(row) > 0 {
					id, _ := strconv.Atoi(row[0])
					m.loadIntoForm(id)
					m.mode = OpModeEdit
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
	if m.mode == OpModeAdd || m.mode == OpModeEdit {
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

			// Auto-uppercase per matricola
			if m.focusIndex == 0 {
				m.inputs[0].SetValue(strings.ToUpper(m.inputs[0].Value()))
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
func (m OperatoriModel) View() string {
	width := 80
	if m.width > 0 {
		width = min(m.width, 90)
	}

	// Titolo dinamico
	title := "GESTIONE OPERATORI"
	if m.mode == OpModeAdd {
		title = "NUOVO OPERATORE"
	} else if m.mode == OpModeEdit {
		title = fmt.Sprintf("MODIFICA OPERATORE #%d", m.selectedID)
	}

	header := RenderHeader(title, width)

	var body string

	// Dialog conferma eliminazione
	if m.showConfirm {
		body = RenderConfirmDialog(
			fmt.Sprintf("Eliminare l'operatore #%d?", m.deletingID),
			width,
			0,
		)
	} else if m.mode == OpModeList {
		// Vista lista
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
		// Vista form
		var form strings.Builder
		labels := []string{"Matricola", "Nome", "Cognome", "Ruolo"}

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
