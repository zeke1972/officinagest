# Officina Manager 🚗⚙️

**Sistema gestionale completo per officine meccaniche** con interfaccia TUI (Terminal User Interface) moderna e intuitiva.

Sviluppato in Go con [Bubble Tea](https://github.com/charmbracelet/bubbletea) e database embedded [BoltDB](https://github.com/etcd-io/bbolt).

## ✨ Funzionalità Principali

### 📋 Gestione Completa
- **Clienti**: Anagrafica completa con dati fiscali italiani (CF, P.IVA, PEC, Codice SDI)
- **Veicoli**: Registrazione targhe, marca/modello, chilometraggio, revisioni
- **Commesse**: Ordini di lavoro con tracking stato, costi manodopera e ricambi
- **Agenda**: Calendario appuntamenti con promemoria
- **Operatori**: Gestione team con ruoli specializzati
- **Preventivi**: Creazione e gestione preventivi con stato accettazione
- **Fatture**: Emissione fatture con numerazione automatica
- **Prima Nota**: Registro entrate/uscite con metodi di pagamento multipli

### 🎨 Interfaccia Moderna
- Design professionale dark theme 2026
- Navigazione intuitiva con tastiera
- Tabelle interattive con ricerca e ordinamento
- Dialoghi di conferma per operazioni critiche
- Statistiche dashboard in tempo reale

### 💾 Database & Backup
- Database embedded BoltDB (zero configurazione)
- Backup automatici programmabili
- Export dati in formato JSON
- Eliminazioni a cascata sicure

### 🛡️ Qualità & Affidabilità
- Validazione dati in input completa
- Gestione errori robusta con logging strutturato
- Transazioni ACID garantite da BoltDB
- Codice modulare e manutenibile

## 📦 Installazione

### Requisiti
- Go 1.20 o superiore
- Sistema operativo: Linux, macOS, Windows

### Quick Start

```bash
# Clona il repository
git clone https://github.com/TUO_USERNAME/officina.git
cd officina

# Scarica dipendenze
go mod download

# Compila
go build -o officina

# Esegui
./officina
```

Il primo avvio creerà automaticamente:
- Directory `~/.officina/` per dati e configurazione
- Database `~/.officina/officina.db`
- Directory backup `~/.officina/backups/`
- File di log `~/.officina/debug.log`

## 🎮 Utilizzo

### Navigazione Base
- **↑↓** o **j/k**: Naviga nelle liste
- **Enter**: Seleziona/Conferma
- **Esc**: Torna indietro/Menu principale
- **q** o **Ctrl+C**: Esci dall'applicazione

### Menu Principale
Dal menu principale puoi accedere a tutte le funzionalità:
1. Gestione Clienti
2. Gestione Veicoli
3. Gestione Commesse
4. Agenda Appuntamenti
5. Gestione Operatori
6. Preventivi
7. Fatture
8. Prima Nota

### Workflow Tipico

#### 1. Registrazione Cliente
```
Menu → Clienti → Nuovo Cliente
```
Inserisci: Nome, Cognome, Telefono, Email, Codice Fiscale, ecc.

#### 2. Registrazione Veicolo
```
Menu → Veicoli → Nuovo Veicolo
```
Collega il veicolo al cliente, inserisci targa, marca, modello.

#### 3. Apertura Commessa
```
Menu → Commesse → Nuova Commessa
```
Seleziona veicolo, inserisci lavori da eseguire, costi.

#### 4. Chiusura Commessa e Fatturazione
```
Menu → Commesse → Seleziona commessa → Modifica stato
Menu → Fatture → Nuova Fattura
```

#### 5. Registrazione Pagamento
```
Menu → Prima Nota → Nuovo Movimento
```
Registra entrata collegandola alla commessa.

## 📁 Struttura Progetto

```
officina/
├── main.go                 # Entry point applicazione
├── config/                 # Gestione configurazione
│   └── config.go
├── logger/                 # Sistema di logging
│   └── logger.go
├── database/               # Layer database
│   ├── db.go              # Operazioni CRUD
│   ├── models.go          # Definizione modelli dati
│   ├── helpers.go         # Utility e query avanzate
│   └── backup.go          # Sistema backup/restore
├── utils/                  # Utility generiche
│   ├── validators.go      # Validatori per dati italiani
│   └── formatters.go      # Formattatori output
└── ui/                     # Interfaccia utente
    ├── app.go             # Router principale
    └── screens/           # Schermate UI
        ├── common.go      # Stili e componenti comuni
        ├── menu.go        # Dashboard principale
        ├── clienti.go     # Gestione clienti
        ├── veicoli.go     # Gestione veicoli
        ├── commesse.go    # Gestione commesse
        ├── agenda.go      # Calendario
        ├── operatori.go   # Gestione operatori
        ├── preventivi.go  # Gestione preventivi
        ├── fatture.go     # Gestione fatture
        └── primanota.go   # Prima nota
```

## 🔧 Configurazione

L'applicazione usa una configurazione di default ottimale per la maggior parte degli utilizzi. 

### Percorsi Default
- **Database**: `~/.officina/officina.db`
- **Backup**: `~/.officina/backups/`
- **Log**: `~/.officina/debug.log`

### Personalizzazione
Puoi modificare la configurazione editando `config/config.go` e ricompilando.

## 💾 Backup e Ripristino

### Backup Automatico
Al primo avvio viene creato automaticamente un backup. L'applicazione mantiene gli ultimi 7 backup.

### Backup Manuale
```go
// Nel codice Go
backupMgr := database.NewBackupManager(db, backupPath, maxFiles)
backupFile, err := backupMgr.CreateBackup()
```

### Export JSON
```go
database.ExportToJSON(db, "export.json")
```

## 🐛 Debug e Logging

I log sono salvati in `~/.officina/debug.log` e includono:
- Eventi dell'applicazione
- Errori database
- Operazioni CRUD
- Timestamp dettagliati

Esempio log:
```
2026-01-08 14:32:15 [INFO]  Avvio Officina Manager v2.0.0
2026-01-08 14:32:15 [INFO]  Database aperto: /home/user/.officina/officina.db
2026-01-08 14:32:15 [INFO]  Backup creato: /home/user/.officina/backups/officina_backup_20260108_143215.db
```

## 🧪 Testing

```bash
# Run tutti i test
go test ./...

# Test con coverage
go test -cover ./...

# Test specifico package
go test ./database
```

## 🏗️ Sviluppo

### Aggiungere una Nuova Entità

1. **Definire il modello** in `database/models.go`:
```go
type NuovaEntita struct {
    ID   int    `json:"id"`
    Nome string `json:"nome"`
}
```

2. **Aggiungere operazioni CRUD** in `database/db.go`

3. **Creare la schermata UI** in `ui/screens/nuovaentita.go`

4. **Registrare nel router** in `ui/app.go`

### Linee Guida
- Seguire le convenzioni di naming Go
- Validare sempre i dati in input
- Gestire errori in modo esplicito
- Documentare funzioni pubbliche con commenti GoDoc
- Usare transazioni BoltDB per operazioni multiple

## 📝 Note Tecniche

### Database BoltDB
- **Tipo**: Key-Value store embedded
- **Transazioni**: ACID compliant
- **Concorrenza**: Single writer, multiple readers
- **File size**: Cresce dinamicamente
- **No server**: Zero setup necessario

### Validatori Italiani
L'applicazione include validatori specifici per:
- Codice Fiscale (16 caratteri, formato standard)
- Partita IVA (11 cifre numeriche)
- CAP (5 cifre)
- Targhe veicoli (formato italiano flessibile)
- PEC e Email

## 🤝 Contribuire

Contributi benvenuti! 

1. Fork il progetto
2. Crea un branch per la feature (`git checkout -b feature/AmazingFeature`)
3. Commit le modifiche (`git commit -m 'Add some AmazingFeature'`)
4. Push al branch (`git push origin feature/AmazingFeature`)
5. Apri una Pull Request

## 📄 Licenza

Distribuito sotto licenza MIT. Vedi `LICENSE` per maggiori informazioni.

## 👨‍💻 Autori

- **emC & Claude** - *Sviluppo iniziale*

## 🙏 Ringraziamenti

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Framework TUI
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [BoltDB](https://github.com/etcd-io/bbolt) - Database embedded
- [Bubbles](https://github.com/charmbracelet/bubbles) - Componenti UI

## 📞 Supporto

Per bug, richieste di feature o domande, apri una issue su GitHub.

---

**Made with ❤️ in Italy** 🇮🇹
