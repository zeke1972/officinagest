package ui

import (
	"officina/database"
	"officina/ui/screens"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MainModel è il model principale che gestisce il routing tra le schermate
type MainModel struct {
	state        screens.AppState
	db           *database.DB
	currentModel tea.Model
	width        int
	height       int
}

// NewApp crea una nuova istanza dell'applicazione
func NewApp(db *database.DB) MainModel {
	return MainModel{
		state:        screens.StateMenu,
		db:           db,
		currentModel: screens.NewMenuModel(db),
	}
}

// Init implementa tea.Model
func (m MainModel) Init() tea.Cmd {
	return nil
}

// Update implementa tea.Model
func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Propaga il resize al model corrente
		m.currentModel, cmd = m.currentModel.Update(msg)
		return m, cmd

	case screens.ChangeScreenMsg:
		// Cambio di schermata
		m.state = screens.AppState(msg)

		// Crea il nuovo model in base allo stato
		switch m.state {
		case screens.StateMenu:
			m.currentModel = screens.NewMenuModel(m.db)
		case screens.StateClienti:
			m.currentModel = screens.NewClientiModel(m.db)
		case screens.StateVeicoli:
			m.currentModel = screens.NewVeicoliModel(m.db)
		case screens.StateCommesse:
			m.currentModel = screens.NewCommesseModel(m.db)
		case screens.StateAgenda:
			m.currentModel = screens.NewAgendaModel(m.db)
		case screens.StateOperatori:
			m.currentModel = screens.NewOperatoriModel(m.db)
		case screens.StatePreventivi:
			m.currentModel = screens.NewPreventiviModel(m.db)
		case screens.StateFatture:
			m.currentModel = screens.NewFattureModel(m.db)
		case screens.StatePrimaNota:
			m.currentModel = screens.NewPrimaNotaModel(m.db)
		}

		// Propaga le dimensioni correnti al nuovo model
		if m.width > 0 && m.height > 0 {
			m.currentModel, _ = m.currentModel.Update(tea.WindowSizeMsg{
				Width:  m.width,
				Height: m.height,
			})
		}

		return m, m.currentModel.Init()
	}

	// Propaga l'update al model corrente
	m.currentModel, cmd = m.currentModel.Update(msg)
	return m, cmd
}

// View implementa tea.Model
func (m MainModel) View() string {
	// Wrapper con margini esterni
	return lipgloss.NewStyle().
		Margin(1, 2).
		Render(m.currentModel.View())
}
