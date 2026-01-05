package database

import (
    "encoding/json"
    "fmt"
    "strings"
    "time"
)

// Cliente rappresenta un cliente dell'officina
type Cliente struct {
    ID                 int    `json:"id"`
    Nome               string `json:"nome"`
    Cognome            string `json:"cognome"`
    Telefono           string `json:"telefono"`
    Email              string `json:"email"`
    PEC                string `json:"pec"`
    CodiceFiscale      string `json:"codice_fiscale"`
    PartitaIVA         string `json:"partita_iva"`
    CodiceDestinatario string `json:"codice_destinatario"` // Codice SDI per fatturazione elettronica
    Indirizzo          string `json:"indirizzo"`
    CAP                string `json:"cap"`
    Citta              string `json:"citta"`
    Provincia          string `json:"provincia"`
}

func (c *Cliente) Validate() error {
    if strings.TrimSpace(c.Nome) == "" {
        return fmt.Errorf("nome non può essere vuoto")
    }
    if strings.TrimSpace(c.Cognome) == "" {
        return fmt.Errorf("cognome non può essere vuoto")
    }
    return nil
}

func (c *Cliente) FullName() string {
    return fmt.Sprintf("%s %s", c.Nome, c.Cognome)
}

// Veicolo rappresenta un veicolo in officina
type Veicolo struct {
    ID        int       `json:"id"`
    Targa     string    `json:"targa"`
    Marca     string    `json:"marca"`
    Modello   string    `json:"modello"`
    Anno      int       `json:"anno"`
    ClienteID int       `json:"cliente_id"` // FK: riferimento al proprietario
    Km        int       `json:"km"`
    UltimaRev time.Time `json:"ultima_rev"`
}

func (v *Veicolo) Validate() error {
    if strings.TrimSpace(v.Targa) == "" {
        return fmt.Errorf("targa non può essere vuota")
    }
    if strings.TrimSpace(v.Marca) == "" {
        return fmt.Errorf("marca non può essere vuota")
    }
    if v.ClienteID <= 0 {
        return fmt.Errorf("cliente_id non valido")
    }
    if v.Anno < 1900 || v.Anno > time.Now().Year()+1 {
        return fmt.Errorf("anno non valido")
    }
    return nil
}

func (v *Veicolo) Description() string {
    return fmt.Sprintf("%s %s (%s)", v.Marca, v.Modello, v.Targa)
}

// Commessa rappresenta un ordine di lavoro
type Commessa struct {
    ID              int       `json:"id"`
    Numero          string    `json:"numero"`           // Generato automaticamente (es. COM-001)
    VeicoloID       int       `json:"veicolo_id"`       // FK: riferimento al veicolo
    DataApertura    time.Time `json:"data_apertura"`    // Data creazione
    DataChiusura    time.Time `json:"data_chiusura"`    // Data chiusura (se applicabile)
    Stato           string    `json:"stato"`            // "Aperta" o "Chiusa"
    LavoriEseguiti  string    `json:"lavori_eseguiti"`  // Descrizione lavori
    Note            string    `json:"note"`             // Note interne
    CostoManodopera float64   `json:"costo_manodopera"` // Costo manodopera
    CostoRicambi    float64   `json:"costo_ricambi"`    // Costo ricambi
    Totale          float64   `json:"totale"`           // Totale calcolato automaticamente
}

func (c *Commessa) Validate() error {
    if c.VeicoloID <= 0 {
        return fmt.Errorf("veicolo_id non valido")
    }
    if !IsValidStatoCommessa(c.Stato) {
        return fmt.Errorf("stato deve essere '%s' o '%s'", StatoCommessaAperta, StatoCommessaChiusa)
    }
    if c.CostoManodopera < 0 {
        return fmt.Errorf("costo manodopera non può essere negativo")
    }
    if c.CostoRicambi < 0 {
        return fmt.Errorf("costo ricambi non può essere negativo")
    }
    return nil
}

func (c *Commessa) CalculateTotal() {
    c.Totale = c.CostoManodopera + c.CostoRicambi
}

func (c *Commessa) IsOpen() bool {
    return c.Stato == StatoCommessaAperta
}

// Appuntamento rappresenta un appuntamento in agenda
type Appuntamento struct {
    ID        int       `json:"id"`
    DataOra   time.Time `json:"data_ora"`   // Data e ora appuntamento
    VeicoloID int       `json:"veicolo_id"` // FK: riferimento al veicolo
    Nota      string    `json:"nota"`       // Note appuntamento
}

// Operatore rappresenta un operatore dell'officina
type Operatore struct {
    ID        int    `json:"id"`
    Matricola string `json:"matricola"` // Codice identificativo operatore
    Nome      string `json:"nome"`
    Cognome   string `json:"cognome"`
    Ruolo     string `json:"ruolo"` // Es: "Meccanico", "Carrozziere", "Elettrauto"
}

// Preventivo rappresenta un preventivo
type Preventivo struct {
    ID          int       `json:"id"`
    Numero      string    `json:"numero"`      // Generato automaticamente (es. PREV-001)
    Cliente     string    `json:"cliente"`     // Nome cliente
    Data        time.Time `json:"data"`        // Data creazione
    Totale      float64   `json:"totale"`      // Importo totale
    Descrizione string    `json:"descrizione"` // Descrizione lavori preventivati
    Accettato   bool      `json:"accettato"`   // true se accettato dal cliente
}

// Fattura rappresenta una fattura emessa
type Fattura struct {
    ID        int       `json:"id"`
    Numero    string    `json:"numero"`     // Generato automaticamente (es. FT-0001/2026)
    Data      time.Time `json:"data"`       // Data emissione
    ClienteID int       `json:"cliente_id"` // FK: riferimento al cliente
    Importo   float64   `json:"importo"`    // Importo totale fattura
}

// MovimentoPrimaNota rappresenta un movimento di prima nota (entrata/uscita)
type MovimentoPrimaNota struct {
    ID          int       `json:"id"`
    Data        time.Time `json:"data"`        // Data movimento
    Descrizione string    `json:"descrizione"` // Descrizione movimento
    Tipo        string    `json:"tipo"`        // "Entrata" o "Uscita"
    Importo     float64   `json:"importo"`     // Importo
    Metodo      string    `json:"metodo"`      // "CASSA", "BANCA", "POS", "ASSEGNO", "BONIFICO"
    CommessaID  int       `json:"commessa_id"` // FK opzionale: collegamento a commessa
}

func (m *MovimentoPrimaNota) Validate() error {
    if !IsValidTipoMovimento(m.Tipo) {
        return fmt.Errorf("tipo deve essere '%s' o '%s'", TipoMovimentoEntrata, TipoMovimentoUscita)
    }
    if m.Importo <= 0 {
        return fmt.Errorf("importo deve essere maggiore di zero")
    }
    if !IsValidMetodoPagamento(m.Metodo) {
        return fmt.Errorf("metodo pagamento non valido (validi: %v)", ValidMetodiPagamento())
    }
    return nil
}

// --- Serializzatori JSON ---

func (c *Cliente) ToJSON() ([]byte, error) {
    return json.Marshal(c)
}

func FromJSONCliente(data []byte) (*Cliente, error) {
    var c Cliente
    err := json.Unmarshal(data, &c)
    return &c, err
}

func (v *Veicolo) ToJSON() ([]byte, error) {
    return json.Marshal(v)
}

func FromJSONVeicolo(data []byte) (*Veicolo, error) {
    var v Veicolo
    err := json.Unmarshal(data, &v)
    return &v, err
}

func (c *Commessa) ToJSON() ([]byte, error) {
    return json.Marshal(c)
}

func FromJSONCommessa(data []byte) (*Commessa, error) {
    var c Commessa
    err := json.Unmarshal(data, &c)
    return &c, err
}

func (a *Appuntamento) ToJSON() ([]byte, error) {
    return json.Marshal(a)
}

func FromJSONAppuntamento(data []byte) (*Appuntamento, error) {
    var a Appuntamento
    err := json.Unmarshal(data, &a)
    return &a, err
}

func (o *Operatore) ToJSON() ([]byte, error) {
    return json.Marshal(o)
}

func FromJSONOperatore(data []byte) (*Operatore, error) {
    var o Operatore
    err := json.Unmarshal(data, &o)
    return &o, err
}

func (p *Preventivo) ToJSON() ([]byte, error) {
    return json.Marshal(p)
}

func FromJSONPreventivo(data []byte) (*Preventivo, error) {
    var p Preventivo
    err := json.Unmarshal(data, &p)
    return &p, err
}

func (f *Fattura) ToJSON() ([]byte, error) {
    return json.Marshal(f)
}

func FromJSONFattura(data []byte) (*Fattura, error) {
    var f Fattura
    err := json.Unmarshal(data, &f)
    return &f, err
}

func (m *MovimentoPrimaNota) ToJSON() ([]byte, error) {
    return json.Marshal(m)
}

func FromJSONMovimentoPrimaNota(data []byte) (*MovimentoPrimaNota, error) {
    var m MovimentoPrimaNota
    err := json.Unmarshal(data, &m)
    return &m, err
}
