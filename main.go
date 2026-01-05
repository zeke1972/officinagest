package main

import (
    "fmt"
    "log"
    "os"

    "officina/config"
    "officina/database"
    "officina/logger"
    "officina/ui"

    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    cfg, err := config.LoadOrDefault()
    if err != nil {
        log.Fatalf("Errore caricamento configurazione: %v", err)
    }

    if err := logger.Init(cfg.App.LogFile, cfg.App.DebugMode); err != nil {
        log.Fatalf("Errore inizializzazione logger: %v", err)
    }
    defer logger.Close()

    logger.Info("Avvio %s v%s", cfg.App.Name, cfg.App.Version)

    db, err := database.InitDB(cfg.Database.Path)
    if err != nil {
        logger.Error("Errore apertura database: %v", err)
        log.Fatalf("Errore apertura DB: %v", err)
    }
    defer db.Close()

    logger.Info("Database aperto: %s", cfg.Database.Path)

    if cfg.Backup.Enabled {
        backupMgr := database.NewBackupManager(db, cfg.App.BackupPath, cfg.Backup.MaxFiles)
        if backupFile, err := backupMgr.CreateBackup(); err != nil {
            logger.Warn("Impossibile creare backup iniziale: %v", err)
        } else {
            logger.Info("Backup creato: %s", backupFile)
        }
    }

    if f, err := tea.LogToFile(cfg.App.LogFile, "debug"); err != nil {
        fmt.Println("Errore creazione log file bubbletea:", err)
        os.Exit(1)
    } else {
        defer f.Close()
    }

    logger.Info("Avvio interfaccia utente")
    p := tea.NewProgram(ui.NewApp(db), tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        logger.Error("Errore esecuzione: %v", err)
        fmt.Printf("Errore esecuzione: %v\n", err)
        os.Exit(1)
    }

    logger.Info("Applicazione terminata correttamente")
}
