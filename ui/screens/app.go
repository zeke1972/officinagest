package screens

import (
    "officina/database"

    tea "github.com/charmbracelet/bubbletea"
)

// AppModel è il model principale dell'applicazione
type AppModel struct {
    db            *database.DBCompat
    currentScreen AppState
    menu          MenuModel
    clienti       ClientiModel
    fornitori     FornitoriModel
    veicoli       VeicoliModel
    commesse      CommesseModel
    agenda        AgendaModel
    primanota     PrimaNotaModel
    operatori     OperatoriModel
    preventivi    PreventiviModel
    fatture       FattureModel
    width         int
    height        int
}

// NewModel crea una nuova istanza del model principale
func NewModel(db *database.DBCompat) AppModel {
    return AppModel{
        db:            db,
        currentScreen: StateMenu,
        menu:          NewMenuModel(db),
        clienti:       NewClientiModel(db),
        fornitori:     NewFornitoriModel(db),
        veicoli:       NewVeicoliModel(db),
        commesse:      NewCommesseModel(db),
        agenda:        NewAgendaModel(db),
        primanota:     NewPrimaNotaModel(db),
        operatori:     NewOperatoriModel(db),
        preventivi:    NewPreventiviModel(db),
        fatture:       NewFattureModel(db),
    }
}

// Init implementa tea.Model
func (m AppModel) Init() tea.Cmd {
    return nil
}

// Update implementa tea.Model
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height

    case ChangeScreenMsg:
        m.currentScreen = AppState(msg)
        return m, nil

    case tea.KeyMsg:
        if msg.String() == "ctrl+c" {
            return m, tea.Quit
        }
    }

    var cmd tea.Cmd

    switch m.currentScreen {
    case StateMenu:
        var model tea.Model
        model, cmd = m.menu.Update(msg)
        m.menu = model.(MenuModel)
    case StateClienti:
        var model tea.Model
        model, cmd = m.clienti.Update(msg)
        m.clienti = model.(ClientiModel)
    case StateFornitori:
        var model tea.Model
        model, cmd = m.fornitori.Update(msg)
        m.fornitori = model.(FornitoriModel)
    case StateVeicoli:
        var model tea.Model
        model, cmd = m.veicoli.Update(msg)
        m.veicoli = model.(VeicoliModel)
    case StateCommesse:
        var model tea.Model
        model, cmd = m.commesse.Update(msg)
        m.commesse = model.(CommesseModel)
    case StateAgenda:
        var model tea.Model
        model, cmd = m.agenda.Update(msg)
        m.agenda = model.(AgendaModel)
    case StatePrimaNota:
        var model tea.Model
        model, cmd = m.primanota.Update(msg)
        m.primanota = model.(PrimaNotaModel)
    case StateOperatori:
        var model tea.Model
        model, cmd = m.operatori.Update(msg)
        m.operatori = model.(OperatoriModel)
    case StatePreventivi:
        var model tea.Model
        model, cmd = m.preventivi.Update(msg)
        m.preventivi = model.(PreventiviModel)
    case StateFatture:
        var model tea.Model
        model, cmd = m.fatture.Update(msg)
        m.fatture = model.(FattureModel)
    }

    return m, cmd
}

// View implementa tea.Model
func (m AppModel) View() string {
    switch m.currentScreen {
    case StateMenu:
        return m.menu.View()
    case StateClienti:
        return m.clienti.View()
    case StateFornitori:
        return m.fornitori.View()
    case StateVeicoli:
        return m.veicoli.View()
    case StateCommesse:
        return m.commesse.View()
    case StateAgenda:
        return m.agenda.View()
    case StatePrimaNota:
        return m.primanota.View()
    case StateOperatori:
        return m.operatori.View()
    case StatePreventivi:
        return m.preventivi.View()
    case StateFatture:
        return m.fatture.View()
    }

    return "Schermata sconosciuta"
}
