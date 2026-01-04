package main

import (
	"fmt"
	"log"
	"os"

	"officina/database"
	"officina/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	dbPath := "officina.db"
	db, err := database.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Errore apertura DB: %v", err)
	}
	defer db.Close()

	if f, err := tea.LogToFile("debug.log", "debug"); err != nil {
		fmt.Println("Errore creazione log file:", err)
		os.Exit(1)
	} else {
		defer f.Close()
	}

	p := tea.NewProgram(ui.NewApp(db), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Errore esecuzione: %v\n", err)
		os.Exit(1)
	}
}
