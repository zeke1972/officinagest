package database

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	bolt "go.etcd.io/bbolt"
)

type BackupManager struct {
	db         *DB
	backupPath string
	maxBackups int
}

func NewBackupManager(db *DB, backupPath string, maxBackups int) *BackupManager {
	return &BackupManager{
		db:         db,
		backupPath: backupPath,
		maxBackups: maxBackups,
	}
}

func (bm *BackupManager) CreateBackup() (string, error) {
	if err := os.MkdirAll(bm.backupPath, 0755); err != nil {
		return "", fmt.Errorf("impossibile creare directory backup: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(bm.backupPath, fmt.Sprintf("officina_backup_%s.db", timestamp))

	err := bm.db.View(func(tx *bolt.Tx) error {
		dst, err := os.Create(backupFile)
		if err != nil {
			return err
		}
		defer dst.Close()

		_, err = tx.WriteTo(dst)
		return err
	})

	if err != nil {
		return "", fmt.Errorf("errore durante il backup: %w", err)
	}

	if err := bm.cleanOldBackups(); err != nil {
		return backupFile, fmt.Errorf("backup creato ma errore durante pulizia: %w", err)
	}

	return backupFile, nil
}

func (bm *BackupManager) cleanOldBackups() error {
	files, err := filepath.Glob(filepath.Join(bm.backupPath, "officina_backup_*.db"))
	if err != nil {
		return err
	}

	if len(files) <= bm.maxBackups {
		return nil
	}

	type fileInfo struct {
		path    string
		modTime time.Time
	}

	var fileInfos []fileInfo
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, fileInfo{
			path:    file,
			modTime: info.ModTime(),
		})
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].modTime.Before(fileInfos[j].modTime)
	})

	toDelete := len(fileInfos) - bm.maxBackups
	for i := 0; i < toDelete; i++ {
		if err := os.Remove(fileInfos[i].path); err != nil {
			return fmt.Errorf("errore eliminazione backup vecchio: %w", err)
		}
	}

	return nil
}

func (bm *BackupManager) RestoreBackup(backupFile string) error {
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("file di backup non trovato: %s", backupFile)
	}

	src, err := os.Open(backupFile)
	if err != nil {
		return fmt.Errorf("impossibile aprire file di backup: %w", err)
	}
	defer src.Close()

	dbPath := bm.db.Path()

	if err := bm.db.Close(); err != nil {
		return fmt.Errorf("impossibile chiudere database: %w", err)
	}

	dst, err := os.Create(dbPath)
	if err != nil {
		return fmt.Errorf("impossibile creare file database: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("errore durante il ripristino: %w", err)
	}

	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return fmt.Errorf("impossibile riaprire database: %w", err)
	}

	bm.db.DB = db

	return nil
}

func (bm *BackupManager) ListBackups() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(bm.backupPath, "officina_backup_*.db"))
	if err != nil {
		return nil, err
	}

	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	return files, nil
}

func ExportToJSON(db *DB, outputPath string) error {
	return db.View(func(tx *bolt.Tx) error {
		f, err := os.Create(outputPath)
		if err != nil {
			return err
		}
		defer f.Close()

		f.WriteString("{\n")

		buckets := []string{
			string(BktClienti),
			string(BktVeicoli),
			string(BktCommesse),
			string(BktAppuntamenti),
			string(BktOperatori),
			string(BktPreventivi),
			string(BktFatture),
			string(BktPrimaNota),
		}

		for i, bucketName := range buckets {
			f.WriteString(fmt.Sprintf("  \"%s\": [\n", bucketName))

			b := tx.Bucket([]byte(bucketName))
			if b != nil {
				first := true
				b.ForEach(func(k, v []byte) error {
					if !first {
						f.WriteString(",\n")
					}
					f.WriteString("    ")
					f.Write(v)
					first = false
					return nil
				})
			}

			f.WriteString("\n  ]")
			if i < len(buckets)-1 {
				f.WriteString(",")
			}
			f.WriteString("\n")
		}

		f.WriteString("}\n")
		return nil
	})
}
