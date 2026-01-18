package main

import (
    "fmt"
    "log"
    "os"

    "context"
    "officina/config"
    "officina/database"
    "officina/logger"
    "officina/ui/screens"

    tea "github.com/charmbracelet/bubbletea"
    "go.mongodb.org/mongo-driver/mongo"
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

    // Inizializza database utilizzando il wrapper compatibile
    ctx := context.Background()
    
    if cfg.Database.Username != "" && cfg.Database.Password != "" {
        // Connection string with authentication
        cfg.Database.URI = fmt.Sprintf("mongodb://%s:%s@%s", cfg.Database.Username, cfg.Database.Password, cfg.Database.URI)
    }
    
    db, err := database.InitCompatDB(ctx, cfg.Database.URI, cfg.Database.Name)
    if err != nil {
        logger.Error("Errore connessione MongoDB: %v", err)
        log.Fatalf("Errore connessione MongoDB: %v", err)
    }
    
    defer func() {
        if db != nil {
            if err := db.Close(); err != nil {
                logger.Error("Errore chiusura MongoDB: %v", err)
            }
        }
    }()

    logger.Info("MongoDB connesso: %s/%s", cfg.Database.URI, cfg.Database.Name)

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
