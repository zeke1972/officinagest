package main

import (
	"fmt"
	"log"
	"os"

	"officina/config"
	"officina/database"
	"officina/logger"
	"officina/ui/screens"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Carica configurazione
	cfg, err := config.LoadOrDefault()
	if err != nil {
		log.Fatalf("Errore caricamento configurazione: %v", err)
	}

	// Inizializza logger
	if err := logger.Init(cfg.App.LogFile, cfg.App.DebugMode); err != nil {
		log.Fatalf("Errore inizializzazione logger: %v", err)
	}
	defer logger.Close()

	logger.Info("Avvio %s v%s", cfg.App.Name, cfg.App.Version)

	// Inizializza database MongoDB
	db, err := database.InitMongoDB(cfg.Database.URI, cfg.Database.Name)
	if err != nil {
		logger.Error("Errore connessione database MongoDB: %v", err)
		log.Fatalf("Errore connessione MongoDB: %v", err)
	}
	defer db.Close()

	logger.Info("Database MongoDB connesso: %s/%s", cfg.Database.URI, cfg.Database.Name)

	// Backup automatico
	if cfg.Backup.Enabled {
		backupMgr := database.NewBackupManager(db, cfg.App.BackupPath, cfg.Backup.MaxFiles)
		if backupFile, err := backupMgr.CreateBackup(); err != nil {
			logger.Warn("Impossibile creare backup iniziale: %v", err)
		} else {
			logger.Info("Backup creato: %s", backupFile)
		}
	}

	// Log Bubbletea (opzionale per debug)
	if cfg.App.DebugMode {
		if f, err := tea.LogToFile(cfg.App.LogFile, "debug"); err != nil {
			fmt.Println("Errore creazione log file bubbletea:", err)
		} else {
			defer f.Close()
		}
	}

	// Avvia interfaccia utente
	logger.Info("Avvio interfaccia utente")

	p := tea.NewProgram(
		screens.NewModel(db),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		logger.Error("Errore esecuzione: %v", err)
		fmt.Printf("Errore esecuzione: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Applicazione terminata correttamente")
}
