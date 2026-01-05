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

// ClienteMode rappresenta la modalità corrente della schermata
type ClienteMode int

const (
	ClList ClienteMode = iota
	ClAdd
	ClEdit
)

// ClientiModel gestisce la schermata clienti
type ClientiModel struct {
	db                     *database.DB
	table                  table.Model
	inputs                 []textinput.Model
	mode                   ClienteMode
	focusIndex             int
	selectedID             int
	err                    error
	msg                    string
	width                  int
	height                 int
	showConfirm            bool
	deletingID             int
	deleteWarningVeicoli   int
	deleteWarningCommesse  int
	deleteWarningMovimenti int
	deleteWarningTotale    float64
}

// NewClientiModel crea una nuova istanza del model clienti
func NewClientiModel(db *database.DB) ClientiModel {
	// Configurazione tabella
	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Cognome", Width: 18},
		{Title: "Nome", Width: 15},
		{Title: "Telefono", Width: 14},
		{Title: "Email", Width: 25},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(12),
	)
	t.SetStyles(GetTableStyles())

	// Configurazione inputs form
	inputs := make([]textinput.Model, 12)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Nome"
	inputs[0].Width = 40

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Cognome"
	inputs[1].Width = 40

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Telefono (es. 333 1234567)"
	inputs[2].Width = 40

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Email (opzionale)"
	inputs[3].Width = 40

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "PEC (opzionale)"
	inputs[4].Width = 40

	inputs[5] = textinput.New()
	inputs[5].Placeholder = "Codice Fiscale (opzionale)"
	inputs[5].CharLimit = 16
	inputs[5].Width = 40

	inputs[6] = textinput.New()
	inputs[6].Placeholder = "Partita IVA (opzionale)"
	inputs[6].CharLimit = 11
	inputs[6].Width = 40

	inputs[7] = textinput.New()
	inputs[7].Placeholder = "Codice SDI (opzionale)"
	inputs[7].CharLimit = 7
	inputs[7].Width = 40

	inputs[8] = textinput.New()
	inputs[8].Placeholder = "Indirizzo (via, numero)"
	inputs[8].Width = 50

	inputs[9] = textinput.New()
	inputs[9].Placeholder = "CAP"
	inputs[9].CharLimit = 5
	inputs[9].Width = 20

	inputs[10] = textinput.New()
	inputs[10].Placeholder = "Città"
	inputs[10].Width = 40

	inputs[11] = textinput.New()
	inputs[11].Placeholder = "Provincia (es. RM)"
	inputs[11].CharLimit = 2
	inputs[11].Width = 10

	m := ClientiModel{
		db:     db,
		table:  t,
		inputs: inputs,
		mode:   ClList,
	}

	m.Refresh()
	return m
}

// Refresh aggiorna la lista dei clienti
func (m *ClientiModel) Refresh() {
	list, _ := m.db.ListClienti()
	rows := []table.Row{}

	for _, c := range list {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", c.ID),
			utils.Truncate(c.Cognome, 18),
			utils.Truncate(c.Nome, 15),
			c.Telefono,
			utils.Truncate(c.Email, 25),
		})
	}

	m.table.SetRows(rows)
}

// countDataForCliente conta veicoli, commesse e movimenti associati a un cliente
func (m *ClientiModel) countDataForCliente(clienteID int) (int, int, int, float64) {
	veicoli, _ := m.db.ListVeicoli()
	commesse, _ := m.db.ListCommesse()
	movimenti, _ := m.db.ListMovimenti()

	// Conta veicoli del cliente
	numVeicoli := 0
	veicoliIDs := make(map[int]bool)
	for _, v := range veicoli {
		if v.ClienteID == clienteID {
			numVeicoli++
			veicoliIDs[v.ID] = true
		}
	}

	// Conta commesse dei veicoli del cliente
	numCommesse := 0
	commesseIDs := make(map[int]bool)
	for _, c := range commesse {
		if veicoliIDs[c.VeicoloID] {
			numCommesse++
			commesseIDs[c.ID] = true
		}
	}

	// Conta movimenti delle commesse
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

	return numVeicoli, numCommesse, numMovimenti, totaleMov
}

// resetForm resetta il form ai valori predefiniti
func (m *ClientiModel) resetForm() {
	for i := range m.inputs {
		m.inputs[i].SetValue("")
	}
	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.inputs[0].Focus()
}

// loadIntoForm carica un cliente nel form
func (m *ClientiModel) loadIntoForm(id int) {
	c, err := m.db.GetCliente(id)
	if err != nil {
		m.err = fmt.Errorf("errore caricamento cliente: %w", err)
		return
	}

	m.selectedID = id
	m.inputs[0].SetValue(c.Nome)
	m.inputs[1].SetValue(c.Cognome)
	m.inputs[2].SetValue(c.Telefono)
	m.inputs[3].SetValue(c.Email)
	m.inputs[4].SetValue(c.PEC)
	m.inputs[5].SetValue(c.CodiceFiscale)
	m.inputs[6].SetValue(c.PartitaIVA)
	m.inputs[7].SetValue(c.CodiceDestinatario)
	m.inputs[8].SetValue(c.Indirizzo)
	m.inputs[9].SetValue(c.CAP)
	m.inputs[10].SetValue(c.Citta)
	m.inputs[11].SetValue(c.Provincia)

	m.focusIndex = 0
	m.err = nil
	m.msg = ""
	m.inputs[0].Focus()
}

// updateFocus aggiorna il focus tra i campi
func (m *ClientiModel) updateFocus() {
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

// validate valida i dati del form
func (m *ClientiModel) validate() error {
	if err := utils.ValidateNotEmpty(m.inputs[1].Value(), "Cognome"); err != nil {
		return err
	}

	if err := utils.ValidateEmail(m.inputs[3].Value()); err != nil {
		return err
	}

	if err := utils.ValidateEmail(m.inputs[4].Value()); err != nil {
		return fmt.Errorf("PEC non valida: %w", err)
	}

	if err := utils.ValidateCodiceFiscale(m.inputs[5].Value()); err != nil {
		return err
	}

	if err := utils.ValidatePartitaIVA(m.inputs[6].Value()); err != nil {
		return err
	}

	if err := utils.ValidateCAP(m.inputs[9].Value()); err != nil {
		return err
	}

	return nil
}

// save salva il cliente corrente
func (m *ClientiModel) save() error {
	if err := m.validate(); err != nil {
		return err
	}

	c := &database.Cliente{
		Nome:               m.inputs[0].Value(),
		Cognome:            m.inputs[1].Value(),
		Telefono:           m.inputs[2].Value(),
		Email:              m.inputs[3].Value(),
		PEC:                m.inputs[4].Value(),
		CodiceFiscale:      strings.ToUpper(m.inputs[5].Value()),
		PartitaIVA:         m.inputs[6].Value(),
		CodiceDestinatario: strings.ToUpper(m.inputs[7].Value()),
		Indirizzo:          m.inputs[8].Value(),
		CAP:                m.inputs[9].Value(),
		Citta:              m.inputs[10].Value(),
		Provincia:          strings.ToUpper(m.inputs[11].Value()),
	}

	if m.mode == ClAdd {
		if err := m.db.CreateCliente(c); err != nil {
			return fmt.Errorf("errore creazione: %w", err)
		}
		m.msg = "✓ Cliente creato con successo"
	} else {
		c.ID = m.selectedID
		if err := m.db.UpdateCliente(c); err != nil {
			return fmt.Errorf("errore aggiornamento: %w", err)
		}
		m.msg = "✓ Cliente aggiornato con successo"
	}

	m.mode = ClList
	m.Refresh()
	return nil
}

// Init implementa tea.Model
func (m ClientiModel) Init() tea.Cmd {
	return nil
}

// Update implementa tea.Model
func (m ClientiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				if err := m.db.DeleteCliente(m.deletingID); err != nil {
					m.err = fmt.Errorf("errore eliminazione: %w", err)
				} else {
					m.msg = fmt.Sprintf("✓ Cliente, %d veicoli, %d commesse e %d movimenti eliminati",
						m.deleteWarningVeicoli, m.deleteWarningCommesse, m.deleteWarningMovimenti)
				}
				m.Refresh()
				m.showConfirm = false
				m.deleteWarningVeicoli = 0
				m.deleteWarningCommesse = 0
				m.deleteWarningMovimenti = 0
				m.deleteWarningTotale = 0

			case "n", "N", "esc":
				m.showConfirm = false
				m.deleteWarningVeicoli = 0
				m.deleteWarningCommesse = 0
				m.deleteWarningMovimenti = 0
				m.deleteWarningTotale = 0
			}
		}
		return m, nil
	}

	// Gestione ESC
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
		if m.mode != ClList {
			m.mode = ClList
			m.err = nil
			m.msg = ""
			return m, nil
		}
		return m, func() tea.Msg { return ChangeScreenMsg(StateMenu) }
	}

	// Modalità Lista
	if m.mode == ClList {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "n":
				m.mode = ClAdd
				m.resetForm()
				return m, nil

			case "e", "enter":
				if row := m.table.SelectedRow(); len(row) > 0 {
					id, _ := strconv.Atoi(row[0])
					m.loadIntoForm(id)
					m.mode = ClEdit
				}
				return m, nil

			case "x", "d":
				if row := m.table.SelectedRow(); len(row) > 0 {
					id, _ := strconv.Atoi(row[0])
					m.deletingID = id

					// Conta dati associati
					numVeicoli, numCommesse, numMovimenti, totaleMov := m.countDataForCliente(id)
					m.deleteWarningVeicoli = numVeicoli
					m.deleteWarningCommesse = numCommesse
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

	// Modalità Form (Add/Edit)
	if m.mode == ClAdd || m.mode == ClEdit {
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

			// Auto-uppercase per CF, Provincia, SDI
			if m.focusIndex == 5 || m.focusIndex == 7 || m.focusIndex == 11 {
				m.inputs[m.focusIndex].SetValue(strings.ToUpper(m.inputs[m.focusIndex].Value()))
			}
		}

		// Aggiorna inputs
		cmds := make([]tea.Cmd, len(m.inputs))
		for i := range m.inputs {
			m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
		}

		return m, tea.Batch(cmds...)
	}

	return m, nil
}

// renderDeleteConfirmation renderizza il dialog di conferma eliminazione con avviso
func (m ClientiModel) renderDeleteConfirmation(width int) string {
	var message strings.Builder

	c, _ := m.db.GetCliente(m.deletingID)
	nome := "???"
	if c != nil {
		nome = c.Cognome + " " + c.Nome
	}

	message.WriteString(fmt.Sprintf("⚠️  ELIMINAZIONE CLIENTE #%d\n", m.deletingID))
	message.WriteString(fmt.Sprintf("%s\n\n", nome))

	if m.deleteWarningVeicoli > 0 || m.deleteWarningCommesse > 0 {
		message.WriteString(ErrorStyle.Render(fmt.Sprintf(
			"ATTENZIONE: Questo cliente ha dati associati!\n\n"+
				"Eliminando il cliente verranno eliminati:\n"+
				"  • %d veicoli\n"+
				"  • %d commesse\n"+
				"  • %d movimenti di Prima Nota (totale: %s)\n\n"+
				"TUTTI I DATI VERRANNO PERSI IN MODO PERMANENTE!\n\n",
			m.deleteWarningVeicoli,
			m.deleteWarningCommesse,
			m.deleteWarningMovimenti,
			utils.FormatEuro(m.deleteWarningTotale),
		)))
	} else {
		message.WriteString("Questo cliente non ha veicoli, commesse o movimenti associati.\n\n")
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
func (m ClientiModel) View() string {
	width := 85
	if m.width > 0 {
		width = min(m.width, 100)
	}

	// Dialog conferma eliminazione
	if m.showConfirm {
		return m.renderDeleteConfirmation(width)
	}

	// Titolo dinamico
	title := "GESTIONE CLIENTI"
	if m.mode == ClAdd {
		title = "NUOVO CLIENTE"
	} else if m.mode == ClEdit {
		title = fmt.Sprintf("MODIFICA CLIENTE #%d", m.selectedID)
	}

	header := RenderHeader(title, width)
	var body string

	if m.mode == ClList {
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
		labels := []string{
			"Nome", "Cognome", "Telefono", "Email",
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

			// Raggruppamento logico
			if i == 3 || i == 7 {
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
