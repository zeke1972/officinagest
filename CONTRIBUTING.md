# Contributing to Officina Manager

Prima di tutto, grazie per aver considerato di contribuire a Officina Manager! 🎉

Questo documento fornisce linee guida per contribuire al progetto.

## 📋 Indice

- [Codice di Condotta](#codice-di-condotta)
- [Come Posso Contribuire?](#come-posso-contribuire)
- [Sviluppo Locale](#sviluppo-locale)
- [Linee Guida Codice](#linee-guida-codice)
- [Processo Pull Request](#processo-pull-request)
- [Stile Commit](#stile-commit)

## 🤝 Codice di Condotta

Partecipando a questo progetto, accetti di mantenere un ambiente rispettoso e inclusivo per tutti.

## 💡 Come Posso Contribuire?

### Segnalare Bug

Se trovi un bug:

1. **Verifica** che il bug non sia già stato segnalato nelle [Issues](https://github.com/TUO_USERNAME/officina/issues)
2. **Apri una nuova issue** con:
   - Titolo chiaro e descrittivo
   - Descrizione dettagliata del problema
   - Passi per riprodurre il bug
   - Comportamento atteso vs effettivo
   - Screenshot se applicabile
   - Versione Go e OS

### Suggerire Funzionalità

Se hai un'idea per una nuova funzionalità:

1. **Verifica** che non sia già stata proposta
2. **Apri una issue** con tag `enhancement` contenente:
   - Descrizione chiara della funzionalità
   - Motivazione (problema che risolve)
   - Esempio d'uso
   - Alternative considerate

### Contribuire Codice

Contributi al codice sono benvenuti! Vedi [Processo Pull Request](#processo-pull-request).

### Migliorare Documentazione

Documentazione chiara è fondamentale:
- README.md
- Commenti GoDoc
- Guide d'uso
- Esempi

## 🛠️ Sviluppo Locale

### Prerequisiti

- Go 1.20 o superiore
- Git
- Make (opzionale ma raccomandato)

### Setup Ambiente

```bash
# Clone repository
git clone https://github.com/TUO_USERNAME/officina.git
cd officina

# Download dipendenze
go mod download

# Verifica che compili
make build

# Run tests
make test

# Run applicazione
make run
```

### Struttura Progetto

```
officina/
├── config/         # Configurazione
├── database/       # Layer database
├── logger/         # Sistema logging
├── ui/            # Interfaccia TUI
├── utils/         # Utility generiche
└── main.go        # Entry point
```

### Comandi Utili

```bash
make build          # Compila
make run            # Esegui
make test           # Run test
make test-coverage  # Test con coverage
make fmt            # Format codice
make vet            # Run go vet
make lint           # Run linter
make clean          # Pulisci build
make dev            # Sviluppo veloce (fmt + build + run)
make check          # Pre-commit (fmt + vet + test)
```

## 📝 Linee Guida Codice

### Convenzioni Go

Seguiamo le [Effective Go](https://golang.org/doc/effective_go) guidelines:

1. **Naming**
   - Package: lowercase, singolare
   - Variabili: camelCase
   - Costanti: CamelCase o UPPER_CASE
   - Funzioni pubbliche: CamelCase
   - Funzioni private: camelCase

2. **Commenti**
   - Tutte le funzioni pubbliche devono avere commento GoDoc
   - Formato: `// FunctionName fa qualcosa...`
   - Commenti su codice complesso

3. **Error Handling**
   ```go
   // ✅ Buono
   if err != nil {
       return fmt.Errorf("operazione fallita: %w", err)
   }
   
   // ❌ Evitare
   if err != nil {
       panic(err)
   }
   ```

4. **Validazione**
   - Usare metodo `Validate()` sui modelli
   - Validare input prima del salvataggio
   ```go
   if err := model.Validate(); err != nil {
       return err
   }
   ```

5. **Costanti**
   - Usare costanti definite in `database/constants.go`
   ```go
   // ✅ Buono
   if stato == StatoCommessaAperta { }
   
   // ❌ Evitare
   if stato == "Aperta" { }
   ```

### Stile Codice

- **Format**: Usa `gofmt` o `make fmt`
- **Imports**: Raggruppati in standard lib, external, internal
- **Line length**: Max 100 caratteri quando possibile
- **Functions**: Max 50 linee, se più grande refactoring

### Testing

1. **Test Required**
   - Tutti i nuovi moduli devono avere test
   - Funzioni pubbliche devono essere testate
   - Validatori devono essere testati

2. **Naming**
   ```go
   func TestFunctionName(t *testing.T) { }
   ```

3. **Table-Driven Tests**
   ```go
   tests := []struct {
       name    string
       input   Type
       want    Type
       wantErr bool
   }{
       {"caso normale", input1, output1, false},
       {"caso errore", input2, nil, true},
   }
   
   for _, tt := range tests {
       t.Run(tt.name, func(t *testing.T) {
           // test
       })
   }
   ```

4. **Coverage**
   - Target: >70% per nuovo codice
   - Run: `make test-coverage`

### Database

1. **Transazioni**
   - Usa sempre transazioni per operazioni multiple
   - Gestisci rollback su errori

2. **Helper Functions**
   - Aggiungi helper in `database/helpers.go` per query comuni
   - Documenta con GoDoc

3. **Validazione**
   - Valida sempre prima di `Create/Update`
   ```go
   func (db *DB) CreateCliente(c *Cliente) error {
       if err := c.Validate(); err != nil {
           return err
       }
       // ...
   }
   ```

### UI

1. **Stili**
   - Usa stili definiti in `ui/screens/common.go`
   - Mantieni consistenza colori e layout

2. **Navigazione**
   - Sempre fornire modo di tornare indietro (ESC)
   - Supportare shortcuts standard (j/k, arrow keys)

3. **Messaggi**
   - Errori: chiari e actionable
   - Successo: conferme visibili
   - Loading: indicatori per operazioni lunghe

## 🔄 Processo Pull Request

1. **Fork e Branch**
   ```bash
   # Fork repository su GitHub
   git clone https://github.com/TUO_USERNAME/officina.git
   cd officina
   git checkout -b feature/amazing-feature
   ```

2. **Sviluppo**
   ```bash
   # Lavora sulla feature
   # ...
   
   # Format e test
   make check
   
   # Commit
   git add .
   git commit -m "feat: Add amazing feature"
   ```

3. **Pre-Submit Checklist**
   - [ ] Codice compila (`make build`)
   - [ ] Test passano (`make test`)
   - [ ] Codice formattato (`make fmt`)
   - [ ] Go vet passa (`make vet`)
   - [ ] Aggiunti test per nuovo codice
   - [ ] Documentazione aggiornata
   - [ ] CHANGELOG.md aggiornato (per feature)

4. **Push e PR**
   ```bash
   git push origin feature/amazing-feature
   ```
   
   Apri Pull Request su GitHub con:
   - Titolo chiaro
   - Descrizione dettagliata
   - Link a issue correlate
   - Screenshot se UI changes

5. **Review**
   - Attendi review
   - Indirizza feedback
   - Una volta approvato, verrà merged

## 📏 Stile Commit

Usiamo [Conventional Commits](https://www.conventionalcommits.org/):

### Format
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types
- `feat`: Nuova feature
- `fix`: Bug fix
- `docs`: Solo documentazione
- `style`: Formattazione, mancano semicolon, ecc.
- `refactor`: Refactoring codice
- `test`: Aggiunta test
- `chore`: Build, tool config, ecc.

### Esempi
```
feat: Add backup rotation feature

Implements automatic rotation of backup files keeping only
the last N backups as configured.

Closes #42

---

fix(database): Correct cliente deletion cascade

Previously, deleting a cliente would leave orphaned veicoli.
Now properly cascades deletion.

Fixes #38

---

docs: Update README with backup instructions

---

test(validators): Add tests for partita IVA validation
```

### Scope (opzionale)
- `database`: Database layer
- `ui`: User interface
- `config`: Configuration
- `logger`: Logging
- `utils`: Utilities

## 🎯 Priorità Contributi

Contributi particolarmente benvenuti in queste aree:

### Alta Priorità 🔴
- Test integrazione database
- Performance optimization
- Bug fix

### Media Priorità 🟡
- UI improvements (ricerca, filtri)
- Export features (PDF)
- Documentazione esempi

### Bassa Priorità 🟢
- Web interface
- REST API
- Mobile app

## 📚 Risorse

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Bubble Tea Docs](https://github.com/charmbracelet/bubbletea)
- [BoltDB Docs](https://github.com/etcd-io/bbolt)

## ❓ Domande?

- Apri una [Discussion](https://github.com/TUO_USERNAME/officina/discussions)
- Contatta i maintainer

## 🙏 Riconoscimenti

Ogni contributo, grande o piccolo, è apprezzato! I contributor verranno riconosciuti nel CHANGELOG e README.

---

**Grazie per contribuire a Officina Manager!** 🚗⚙️❤️
