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
	RagioneSociale     string `json:"ragione_sociale"`
	Telefono           string `json:"telefono"`
	Email              string `json:"email"`
	PEC                string `json:"pec"`
	CodiceFiscale      string `json:"codice_fiscale"`
	PartitaIVA         string `json:"partita_iva"`
	CodiceDestinatario string `json:"codice_destinatario"`
	Indirizzo          string `json:"indirizzo"`
	CAP                string `json:"cap"`
	Citta              string `json:"citta"`
	Provincia          string `json:"provincia"`
}

func (c *Cliente) Validate() error {
	if strings.TrimSpace(c.RagioneSociale) == "" {
		return fmt.Errorf("ragione sociale non può essere vuota")
	}
	return nil
}

func (c *Cliente) FullName() string {
	return c.RagioneSociale
}

// Fornitore rappresenta un fornitore dell'officina
type Fornitore struct {
	ID                 int    `json:"id"`
	RagioneSociale     string `json:"ragione_sociale"`
	Telefono           string `json:"telefono"`
	Email              string `json:"email"`
	PEC                string `json:"pec"`
	CodiceFiscale      string `json:"codice_fiscale"`
	PartitaIVA         string `json:"partita_iva"`
	CodiceDestinatario string `json:"codice_destinatario"`
	Indirizzo          string `json:"indirizzo"`
	CAP                string `json:"cap"`
	Citta              string `json:"citta"`
	Provincia          string `json:"provincia"`
}

func (f *Fornitore) Validate() error {
	if strings.TrimSpace(f.RagioneSociale) == "" {
		return fmt.Errorf("ragione sociale non può essere vuota")
	}
	return nil
}

func (f *Fornitore) FullName() string {
	return f.RagioneSociale
}

// Veicolo rappresenta un veicolo in officina
type Veicolo struct {
	ID        int       `json:"id"`
	Targa     string    `json:"targa"`
	Marca     string    `json:"marca"`
	Modello   string    `json:"modello"`
	Anno      int       `json:"anno"`
	ClienteID int       `json:"cliente_id"`
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
	Numero          string    `json:"numero"`
	VeicoloID       int       `json:"veicolo_id"`
	DataApertura    time.Time `json:"data_apertura"`
	DataChiusura    time.Time `json:"data_chiusura"`
	Stato           string    `json:"stato"`
	LavoriEseguiti  string    `json:"lavori_eseguiti"`
	Note            string    `json:"note"`
	CostoManodopera float64   `json:"costo_manodopera"`
	CostoRicambi    float64   `json:"costo_ricambi"`
	Totale          float64   `json:"totale"`
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
	DataOra   time.Time `json:"data_ora"`
	VeicoloID int       `json:"veicolo_id"`
	Nota      string    `json:"nota"`
}

// Operatore rappresenta un operatore dell'officina
type Operatore struct {
	ID        int    `json:"id"`
	Matricola string `json:"matricola"`
	Nome      string `json:"nome"`
	Cognome   string `json:"cognome"`
	Ruolo     string `json:"ruolo"`
}

// Preventivo rappresenta un preventivo
type Preventivo struct {
	ID          int       `json:"id"`
	Numero      string    `json:"numero"`
	Cliente     string    `json:"cliente"`
	Data        time.Time `json:"data"`
	Totale      float64   `json:"totale"`
	Descrizione string    `json:"descrizione"`
	Accettato   bool      `json:"accettato"`
}

// Fattura rappresenta una fattura emessa
type Fattura struct {
	ID        int       `json:"id"`
	Numero    string    `json:"numero"`
	Data      time.Time `json:"data"`
	ClienteID int       `json:"cliente_id"`
	Importo   float64   `json:"importo"`
}

// MovimentoPrimaNota rappresenta un movimento di prima nota (entrata/uscita)
type MovimentoPrimaNota struct {
	ID            int       `json:"id"`
	Data          time.Time `json:"data"`
	Descrizione   string    `json:"descrizione"`
	Tipo          string    `json:"tipo"`
	Importo       float64   `json:"importo"`
	Metodo        string    `json:"metodo"`
	CommessaID    int       `json:"commessa_id"`
	FornitoreID   int       `json:"fornitore_id"`
	NumeroFattura string    `json:"numero_fattura"`
	DataFattura   time.Time `json:"data_fattura"`
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

func (f *Fornitore) ToJSON() ([]byte, error) {
	return json.Marshal(f)
}

func FromJSONFornitore(data []byte) (*Fornitore, error) {
	var f Fornitore
	err := json.Unmarshal(data, &f)
	return &f, err
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
