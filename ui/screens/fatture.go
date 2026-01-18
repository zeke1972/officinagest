package screens

import (
	"fmt"
	"officina/database"
	"officina/utils"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FattureMode rappresenta la modalità corrente
type FattureMode int

const (
	FatModeList FattureMode = iota
	FatModeAdd
	FatModeEdit
)

// FattureModel gestisce la schermata fatture
type FattureModel struct {
	db          *database.DB
	table       table.Model
	inputs      []textinput.Model
	mode        FattureMode
	focusIndex  int
	selectedID  int
	err         error
	msg         string
	width       int
	height      int
	showConfirm bool
	deletingID  int
}

// NewFattureModel crea una nuova istanza del model fatture
func NewFattureModel(db *database.DB) FattureModel {
	// Configurazione tabella
	t := table.New(
		table.WithColumns([]table.Column{
			{Title: "ID", Width: 4},
			{Title: "Numero", Width: 15},
			{Title: "Data", Width: 12},
			{Title: "Cliente", Width: 30},
			{Title: "Importo", Width: 12},
		}),
		table.WithHeight(12),
		table.WithFocused(true),
	)

	t.SetStyles(GetTableStyles())

	// Configurazione inputs
	inputs := make([]textinput.Model, 3)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Data (GG/MM/AAAA)"
	inputs[0].CharLimit = 10
	inputs[0].Width = 30

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Cliente o descrizione"
	inputs[1].Width = 50

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Importo (es. 1250.50)"
	inputs[2].Width = 30

	m := FattureModel{
		db:     db,
		table:  t,
		inputs: inputs,
		mode:   FatModeList,
	}

	m.Refresh()
	return m
}

// Refresh aggiorna la lista delle fatture
func (m *FattureModel) Refresh() {
	list, _ := m.db.ListFatture()
	rows := []table.Row{}

	for _, f := range list {
		cliente := "—"
		if f.ClienteID > 0 {
			c, err := m.db.GetCliente(f.ClienteID)
			if err == nil && c != nil {
				cliente = c.RagioneSociale
			}
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("%d", f.ID),
			f.Numero,
			utils.FormatDate(f.Data),
			utils.Truncate(cliente, 30),
			utils.FormatEuro(f.Importo),
		})
	}

	m.table.SetRows(rows)
}

// resetForm resetta il form
func (m *FattureModel) resetForm() {
	for i := range m.inputs {
		m.inputs[i].SetValue("")
	}

	// Imposta data corrente
	m.inputs[0].SetValue(time.Now().Format("02/01/2006"))
	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.inputs[0].Focus()
}

// loadIntoForm carica una fattura nel form
func (m *FattureModel) loadIntoForm(id int) {
	f, err := m.db.GetFattura(id)
	if err != nil {
		m.err = fmt.Errorf("errore caricamento fattura: %w", err)
		return
	}

	m.selectedID = id
	m.inputs[0].SetValue(f.Data.Format("02/01/2006"))

	cliente := ""
	if f.ClienteID > 0 {
		c, _ := m.db.GetCliente(f.ClienteID)
		if c != nil {
			cliente = c.RagioneSociale
		}
	}

	m.inputs[1].SetValue(cliente)
	m.inputs[2].SetValue(fmt.Sprintf("%.2f", f.Importo))

	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.inputs[0].Focus()
}

// updateFocus aggiorna il focus tra i campi
func (m *FattureModel) updateFocus() {
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

// validate valida i dati del form
func (m *FattureModel) validate() error {
	dateStr := strings.TrimSpace(m.inputs[0].Value())
	if dateStr == "" {
		return fmt.Errorf("data obbligatoria")
	}

	_, err := time.Parse("02/01/2006", dateStr)
	if err != nil {
		return fmt.Errorf("formato data non valido (usa GG/MM/AAAA)")
	}

	if err := utils.ValidateNotEmpty(m.inputs[1].Value(), "Cliente"); err != nil {
		return err
	}

	importo, err := utils.ParseFloat(m.inputs[2].Value())
	if err != nil {
		return fmt.Errorf("importo non valido")
	}

	if err := utils.ValidateImportoPositivo(importo); err != nil {
		return err
	}

	return nil
}

// save salva la fattura corrente
func (m *FattureModel) save() error {
	if err := m.validate(); err != nil {
		return err
	}

	data, _ := time.Parse("02/01/2006", strings.TrimSpace(m.inputs[0].Value()))
	importo, _ := utils.ParseFloat(m.inputs[2].Value())

	f := &database.Fattura{
		Data:    data,
		Importo: importo,
	}

	if m.mode == FatModeAdd {
		if err := m.db.CreateFattura(f); err != nil {
			return fmt.Errorf("errore creazione: %w", err)
		}
		m.msg = "✓ Fattura creata con successo"
	} else {
		f.ID = m.selectedID
		// Mantieni numero esistente
		old, _ := m.db.GetFattura(m.selectedID)
		if old != nil {
			f.Numero = old.Numero
			f.ClienteID = old.ClienteID
		}

		if err := m.db.UpdateFattura(f); err != nil {
			return fmt.Errorf("errore aggiornamento: %w", err)
		}
		m.msg = "✓ Fattura aggiornata con successo"
	}

	m.mode = FatModeList
	m.Refresh()
	return nil
}

// Init implementa tea.Model
func (m FattureModel) Init() tea.Cmd {
	return nil
}

// Update implementa tea.Model
func (m FattureModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				if err := m.db.DeleteFattura(m.deletingID); err != nil {
					m.err = fmt.Errorf("errore eliminazione: %w", err)
				} else {
					m.msg = "✓ Fattura eliminata con successo"
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
		if m.mode != FatModeList {
			m.mode = FatModeList
			m.err = nil
			m.msg = ""
			return m, nil
		}
		return m, func() tea.Msg { return ChangeScreenMsg(StateMenu) }
	}

	// Modalità Lista
	if m.mode == FatModeList {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "n":
				m.mode = FatModeAdd
				m.resetForm()
				return m, nil
			case "e", "enter":
				if row := m.table.SelectedRow(); len(row) > 0 {
					id, _ := strconv.Atoi(row[0])
					m.loadIntoForm(id)
					m.mode = FatModeEdit
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

	// Modalità Form (Add/Edit)
	if m.mode == FatModeAdd || m.mode == FatModeEdit {
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
func (m FattureModel) View() string {
	width := 85
	if m.width > 0 {
		width = min(m.width, 95)
	}

	// Titolo dinamico
	title := "GESTIONE FATTURE"
	if m.mode == FatModeAdd {
		title = "NUOVA FATTURA"
	} else if m.mode == FatModeEdit {
		title = fmt.Sprintf("MODIFICA FATTURA #%d", m.selectedID)
	}

	header := RenderHeader(title, width)
	var body string

	// Dialog conferma eliminazione
	if m.showConfirm {
		var message strings.Builder
		message.WriteString(fmt.Sprintf("⚠️  ELIMINAZIONE FATTURA #%d\n\n", m.deletingID))
		message.WriteString(WarningStyle.Render("Sei sicuro di voler procedere?\n"))
		message.WriteString(HelpStyle.Render("\n[Y] Sì, elimina • [N/Esc] Annulla"))

		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Padding(1, 2).
			Width(50).
			Render(message.String())

		return CenterContent(m.width, m.height, box)
	} else if m.mode == FatModeList {
		// Vista lista
		helpText := lipgloss.NewStyle().
			MarginBottom(1).
			Foreground(ColorSubText).
			Render("[N] Nuova • [E/↵] Modifica • [X/D] Elimina • [ESC] Menu")

		body = lipgloss.JoinVertical(
			lipgloss.Left,
			helpText,
			m.table.View(),
		)
	} else {
		// Vista form
		var form strings.Builder
		labels := []string{"Data", "Cliente", "Importo €"}

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
