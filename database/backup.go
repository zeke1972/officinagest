package database

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

// BackupManager gestisce i backup del database
type BackupManager struct {
	db       *DB
	basePath string
	maxFiles int
}

// NewBackupManager crea un nuovo gestore di backup
func NewBackupManager(db *DB, basePath string, maxFiles int) *BackupManager {
	return &BackupManager{
		db:       db,
		basePath: basePath,
		maxFiles: maxFiles,
	}
}

// CreateBackup crea un backup del database
func (bm *BackupManager) CreateBackup() (string, error) {
	// Crea la directory di backup se non esiste
	if err := os.MkdirAll(bm.basePath, 0755); err != nil {
		return "", fmt.Errorf("impossibile creare directory backup: %w", err)
	}

	// Genera nome file backup con timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(bm.basePath, fmt.Sprintf("officina_backup_%s.db", timestamp))

	// Crea il file di backup
	dst, err := os.Create(backupFile)
	if err != nil {
		return "", fmt.Errorf("impossibile creare file backup: %w", err)
	}
	defer dst.Close()

	// Esegui backup usando BoltDB
	err = bm.db.View(func(tx *bolt.Tx) error {
		_, err := tx.WriteTo(dst)
		return err
	})

	if err != nil {
		os.Remove(backupFile)
		return "", fmt.Errorf("errore durante il backup: %w", err)
	}

	// Pulisci vecchi backup
	if err := bm.cleanOldBackups(); err != nil {
		return backupFile, fmt.Errorf("backup creato ma pulizia fallita: %w", err)
	}

	return backupFile, nil
}

// cleanOldBackups rimuove i backup più vecchi se superano il limite
func (bm *BackupManager) cleanOldBackups() error {
	files, err := os.ReadDir(bm.basePath)
	if err != nil {
		return err
	}

	// Filtra solo i file di backup
	var backupFiles []os.DirEntry
	for _, f := range files {
		if !f.IsDir() && strings.HasPrefix(f.Name(), "officina_backup_") && strings.HasSuffix(f.Name(), ".db") {
			backupFiles = append(backupFiles, f)
		}
	}

	// Se non superiamo il limite, non fare nulla
	if len(backupFiles) <= bm.maxFiles {
		return nil
	}

	// Ordina per data (dal più vecchio al più recente)
	sort.Slice(backupFiles, func(i, j int) bool {
		return backupFiles[i].Name() < backupFiles[j].Name()
	})

	// Elimina i file più vecchi
	toDelete := len(backupFiles) - bm.maxFiles
	for i := 0; i < toDelete; i++ {
		filePath := filepath.Join(bm.basePath, backupFiles[i].Name())
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("impossibile eliminare backup vecchio %s: %w", filePath, err)
		}
	}

	return nil
}

// RestoreBackup ripristina il database da un file di backup
func (bm *BackupManager) RestoreBackup(backupFile string) error {
	// Verifica che il file di backup esista
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("file di backup non trovato: %s", backupFile)
	}

	// Chiudi il database corrente
	dbPath := bm.db.Path()
	if err := bm.db.Close(); err != nil {
		return fmt.Errorf("impossibile chiudere database: %w", err)
	}

	// Copia il backup sul database corrente
	src, err := os.Open(backupFile)
	if err != nil {
		return fmt.Errorf("impossibile aprire backup: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(dbPath)
	if err != nil {
		return fmt.Errorf("impossibile creare file database: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("errore durante il ripristino: %w", err)
	}

	// Riapri il database
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return fmt.Errorf("impossibile riaprire database: %w", err)
	}

	bm.db.DB = db
	return nil
}

// ListBackups elenca tutti i backup disponibili
func (bm *BackupManager) ListBackups() ([]string, error) {
	files, err := os.ReadDir(bm.basePath)
	if err != nil {
		return nil, err
	}

	var backups []string
	for _, f := range files {
		if !f.IsDir() && strings.HasPrefix(f.Name(), "officina_backup_") && strings.HasSuffix(f.Name(), ".db") {
			backups = append(backups, filepath.Join(bm.basePath, f.Name()))
		}
	}

	// Ordina dal più recente al più vecchio
	sort.Slice(backups, func(i, j int) bool {
		return backups[i] > backups[j]
	})

	return backups, nil
}
