package screens

import (
    "fmt"
    "officina/database"
    "strings"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// MenuItem rappresenta una voce del menu
type MenuItem struct {
    Label string
    Icon  string
    State AppState
}

// MenuModel gestisce il menu principale
type MenuModel struct {
    db                *database.DBCompat
    cursor            int
    items             []MenuItem
    width             int
    height            int
    todayAppointments int
    openCommesse      int
}

// NewMenuModel crea una nuova istanza del menu
func NewMenuModel(db *database.DBCompat) MenuModel {
    m := MenuModel{
        db:     db,
        cursor: 0,
        items: []MenuItem{
            {Label: "Gestione Clienti", Icon: "👥", State: StateClienti},
            {Label: "Gestione Veicoli", Icon: "🚗", State: StateVeicoli},
            {Label: "Gestione Fornitori", Icon: "🏢", State: StateFornitori},
            {Label: "Gestione Commesse", Icon: "🔧", State: StateCommesse},
            {Label: "Agenda & Appuntamenti", Icon: "📅", State: StateAgenda},
            {Label: "Prima Nota & Cassa", Icon: "💶", State: StatePrimaNota},
            {Label: "Operatori", Icon: "👨‍🔧", State: StateOperatori},
            {Label: "Preventivi", Icon: "💰", State: StatePreventivi},
            {Label: "Fatture & Ricevute", Icon: "📄", State: StateFatture},
        },
    }

    m.RefreshStats()
    return m
}

// RefreshStats aggiorna le statistiche del menu
func (m *MenuModel) RefreshStats() {
    if m.db == nil {
        return
    }

    // Update open commesse count
    var count int
    count, _ = m.db.CountCommesseAperte()
    m.openCommesse = count

    // Count today's appointments (simplified - just get all and filter)
    // For efficiency in production, consider adding a dedicated count method
    list, _ := m.db.ListAppuntamenti()
    today := time.Now().Format("2006-01-02")
    appointmentCount := 0
    for _, a := range list {
        if a.DataOra.Format("2006-01-02") == today {
            appointmentCount++
        }
    }
    m.todayAppointments = appointmentCount
}

// Init implementa tea.Model
func (m MenuModel) Init() tea.Cmd {
    return nil
}

// Update implementa tea.Model
func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.RefreshStats()

    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        return m, nil

    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit

        case "up", "k":
            if m.cursor > 0 {
                m.cursor--
            } else {
                m.cursor = len(m.items) - 1
            }

        case "down", "j":
            if m.cursor < len(m.items)-1 {
                m.cursor++
            } else {
                m.cursor = 0
            }

        case "enter", " ":
            target := m.items[m.cursor].State
            return m, func() tea.Msg { return ChangeScreenMsg(target) }

        case "1", "2", "3", "4", "5", "6", "7", "8", "9":
            num := int(msg.String()[0]-'0') - 1
            if num >= 0 && num < len(m.items) {
                target := m.items[num].State
                return m, func() tea.Msg { return ChangeScreenMsg(target) }
            }
        }
    }

    return m, nil
}

// View implementa tea.Model
func (m MenuModel) View() string {
    width := 70
    if m.width > 0 {
        width = min(m.width, 80)
    }

    header := RenderHeader("MENU PRINCIPALE", width)

    var statsBuilder strings.Builder
    if m.todayAppointments > 0 || m.openCommesse > 0 {
        statsBuilder.WriteString(lipgloss.NewStyle().
            Foreground(ColorSubText).
            Render("📊 Dashboard rapida") + "\n\n")

        if m.todayAppointments > 0 {
            badge := WarningBadge(fmt.Sprintf(" %d Appuntamenti oggi ", m.todayAppointments))
            statsBuilder.WriteString("  " + badge + "\n")
        }

        if m.openCommesse > 0 {
            badge := InfoBadge(fmt.Sprintf(" %d Commesse aperte ", m.openCommesse))
            statsBuilder.WriteString("  " + badge + "\n")
        }

        statsBuilder.WriteString("\n")
    }

    var menuBuilder strings.Builder
    for i, item := range m.items {
        badge := ""
        if item.State == StateAgenda && m.todayAppointments > 0 {
            badge = lipgloss.NewStyle().
                Foreground(ColorWarning).
                Bold(true).
                Render(fmt.Sprintf(" [%d]", m.todayAppointments))
        }

        if item.State == StateCommesse && m.openCommesse > 0 {
            badge = lipgloss.NewStyle().
                Foreground(ColorHighlight).
                Bold(true).
                Render(fmt.Sprintf(" [%d]", m.openCommesse))
        }

        var numLabel string
        if i == m.cursor {
            numLabel = lipgloss.NewStyle().
                Foreground(ColorPrimary).
                Background(ColorBgLight).
                Bold(true).
                Render(fmt.Sprintf("[%d]", i+1))
        } else {
            numLabel = lipgloss.NewStyle().
                Foreground(ColorSubText).
                Bold(true).
                Render(fmt.Sprintf("[%d]", i+1))
        }

        cursor := "  "
        if i == m.cursor {
            cursor = "▶ "
        }

        label := fmt.Sprintf("%s %s %s%s", numLabel, item.Icon, item.Label, badge)
        lineStyle := lipgloss.NewStyle().
            Foreground(ColorText).
            Padding(0, 1)

        menuBuilder.WriteString(lineStyle.Render(cursor+label) + "\n")

        if i < len(m.items)-1 {
            menuBuilder.WriteString("\n")
        }
    }

    footer := RenderFooter(width)

    content := lipgloss.JoinVertical(
        lipgloss.Left,
        header,
        "",
        statsBuilder.String(),
        lipgloss.NewStyle().Padding(0, 4).Render(menuBuilder.String()),
        "",
        footer,
    )

    box := MainBoxStyle.Copy().Width(width - 4).Render(content)

    if m.width > 0 && m.height > 0 {
        return CenterContent(m.width, m.height, box)
    }

    return "\n" + box
}
