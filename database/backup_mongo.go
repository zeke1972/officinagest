package database

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// BackupManagerMongo gestisce i backup del database MongoDB tramite JSON export
type BackupManagerMongo struct {
	db       *DB
	basePath string
	maxFiles int
}

// NewBackupManagerMongo crea un nuovo gestore di backup per MongoDB
func NewBackupManagerMongo(db *DB, basePath string, maxFiles int) *BackupManagerMongo {
	return &BackupManagerMongo{
		db:       db,
		basePath: basePath,
		maxFiles: maxFiles,
	}
}

// CreateBackup crea un backup di tutte le collezioni MongoDB in formato JSON
func (bm *BackupManagerMongo) CreateBackup() (string, error) {
	// Crea la directory di backup se non esiste
	if err := os.MkdirAll(bm.basePath, 0755); err != nil {
		return "", fmt.Errorf("impossibile creare directory backup: %w", err)
	}

	// Genera nome directory backup con timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupDir := filepath.Join(bm.basePath, fmt.Sprintf("officina_backup_%s", timestamp))

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("impossibile creare directory backup: %w", err)
	}

	// Lista delle collezioni da esportare
	collections := []string{
		"clienti",
		"fornitori",
		"veicoli",
		"commesse",
		"appuntamenti",
		"operatori",
		"preventivi",
		"fatture",
		"movimenti_primanota",
	}

	// Esporta ogni collezione in un file JSON separato
	for _, collection := range collections {
		data, err := bm.db.ExportToJSON(collection)
		if err != nil {
			// Continua anche se una collezione fallisce
			fmt.Printf("Warning: errore export collection %s: %v\n", collection, err)
			continue
		}

		backupFile := filepath.Join(backupDir, fmt.Sprintf("%s.json", collection))
		if err := os.WriteFile(backupFile, data, 0644); err != nil {
			return backupDir, fmt.Errorf("errore scrittura backup %s: %w", collection, err)
		}
	}

	// Crea file di metadati
	metadata := map[string]interface{}{
		"timestamp":   timestamp,
		"collections": collections,
		"version":     "2.0.0",
	}

	metadataBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return backupDir, fmt.Errorf("errore creazione metadati: %w", err)
	}

	metadataFile := filepath.Join(backupDir, "metadata.json")
	if err := os.WriteFile(metadataFile, metadataBytes, 0644); err != nil {
		return backupDir, fmt.Errorf("errore scrittura metadati: %w", err)
	}

	// Pulisci vecchi backup
	if err := bm.cleanOldBackups(); err != nil {
		return backupDir, fmt.Errorf("backup creato ma pulizia fallita: %w", err)
	}

	return backupDir, nil
}

// cleanOldBackups rimuove i backup più vecchi se superano il limite
func (bm *BackupManagerMongo) cleanOldBackups() error {
	entries, err := os.ReadDir(bm.basePath)
	if err != nil {
		return err
	}

	// Filtra solo le directory di backup
	var backupDirs []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "officina_backup_") {
			backupDirs = append(backupDirs, entry)
		}
	}

	// Se non superiamo il limite, non fare nulla
	if len(backupDirs) <= bm.maxFiles {
		return nil
	}

	// Ordina per nome (dal più vecchio al più recente - il timestamp nel nome lo garantisce)
	sort.Slice(backupDirs, func(i, j int) bool {
		return backupDirs[i].Name() < backupDirs[j].Name()
	})

	// Elimina le directory più vecchie
	toDelete := len(backupDirs) - bm.maxFiles
	for i := 0; i < toDelete; i++ {
		dirPath := filepath.Join(bm.basePath, backupDirs[i].Name())
		if err := os.RemoveAll(dirPath); err != nil {
			return fmt.Errorf("impossibile eliminare backup vecchio %s: %w", dirPath, err)
		}
	}

	return nil
}

// ListBackups elenca tutti i backup disponibili
func (bm *BackupManagerMongo) ListBackups() ([]string, error) {
	entries, err := os.ReadDir(bm.basePath)
	if err != nil {
		return nil, err
	}

	var backups []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "officina_backup_") {
			backups = append(backups, filepath.Join(bm.basePath, entry.Name()))
		}
	}

	// Ordina dal più recente al più vecchio
	sort.Slice(backups, func(i, j int) bool {
		return backups[i] > backups[j]
	})

	return backups, nil
}

// RestoreBackup ripristina il database da una directory di backup JSON
func (bm *BackupManagerMongo) RestoreBackup(backupDir string) error {
	// Verifica che la directory di backup esista
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return fmt.Errorf("directory di backup non trovata: %s", backupDir)
	}

	// Leggi metadati
	metadataFile := filepath.Join(backupDir, "metadata.json")
	metadataBytes, err := os.ReadFile(metadataFile)
	if err != nil {
		return fmt.Errorf("impossibile leggere metadati: %w", err)
	}

	var metadata struct {
		Collections []string `json:"collections"`
	}
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return fmt.Errorf("errore parsing metadati: %w", err)
	}

	ctx := context.Background()

	// Per ogni collezione, importa i dati
	for _, collection := range metadata.Collections {
		backupFile := filepath.Join(backupDir, fmt.Sprintf("%s.json", collection))

		// Verifica che il file esista
		if _, err := os.Stat(backupFile); os.IsNotExist(err) {
			fmt.Printf("Warning: file backup mancante per collection %s\n", collection)
			continue
		}

		// Leggi file JSON
		data, err := os.ReadFile(backupFile)
		if err != nil {
			return fmt.Errorf("errore lettura backup %s: %w", collection, err)
		}

		// Parse JSON array
		var docs []bson.M
		if err := json.Unmarshal(data, &docs); err != nil {
			return fmt.Errorf("errore parsing JSON %s: %w", collection, err)
		}

		// Cancella collezione esistente
		if err := bm.db.mongo.db.Collection(collection).Drop(ctx); err != nil {
			fmt.Printf("Warning: impossibile droppare collection %s: %v\n", collection, err)
		}

		// Importa documenti
		if len(docs) > 0 {
			var docsInterface []interface{}
			for _, doc := range docs {
				docsInterface = append(docsInterface, doc)
			}

			_, err := bm.db.mongo.db.Collection(collection).InsertMany(ctx, docsInterface)
			if err != nil {
				return fmt.Errorf("errore import collection %s: %w", collection, err)
			}
		}
	}

	return nil
}
