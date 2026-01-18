package database

import (
	"context"
	"fmt"
	"time"
)

// DBCompat è un wrapper compatibile per mantenere l'interfaccia BoltDB originale
type DBCompat struct {
	mongoDB *MongoDB
	ctx     context.Context
}

// NewDBCompat crea un wrapper compatibile
type NewDBCompat func(mongoDB *MongoDB, ctx context.Context) *DBCompat

// InitCompatDB inizializza il database compatibile (wrapper per MongoDB)
func InitCompatDB(ctx context.Context, uri string, dbName string) (*DBCompat, error) {
	mongoDB, err := InitMongoDB(ctx, uri, dbName)
	if err != nil {
		return nil, err
	}
	
	return &DBCompat{
		mongoDB: mongoDB,
		ctx:     ctx,
	}, nil
}

// Close chiude la connessione
func (db *DBCompat) Close() error {
	return db.mongoDB.Close(db.ctx)
}

// ==================== CLIENTI ====================

func (db *DBCompat) CreateCliente(c *Cliente) error {
	return db.mongoDB.CreateCliente(db.ctx, c)
}

func (db *DBCompat) GetCliente(id int) (*Cliente, error) {
	return db.mongoDB.GetCliente(db.ctx, id)
}

func (db *DBCompat) UpdateCliente(c *Cliente) error {
	return db.mongoDB.UpdateCliente(db.ctx, c)
}

func (db *DBCompat) DeleteCliente(id int) error {
	return db.mongoDB.DeleteCliente(db.ctx, id)
}

func (db *DBCompat) ListClienti() ([]Cliente, error) {
	return db.mongoDB.ListClienti(db.ctx)
}

// ==================== FORNITORI ====================

func (db *DBCompat) CreateFornitore(f *Fornitore) error {
	return db.mongoDB.CreateFornitore(db.ctx, f)
}

func (db *DBCompat) GetFornitore(id int) (*Fornitore, error) {
	return db.mongoDB.GetFornitore(db.ctx, id)
}

func (db *DBCompat) UpdateFornitore(f *Fornitore) error {
	return db.mongoDB.UpdateFornitore(db.ctx, f)
}

func (db *DBCompat) DeleteFornitore(id int) error {
	return db.mongoDB.DeleteFornitore(db.ctx, id)
}

func (db *DBCompat) ListFornitori() ([]Fornitore, error) {
	return db.mongoDB.ListFornitori(db.ctx)
}

// ==================== VEICOLI ====================

func (db *DBCompat) CreateVeicolo(v *Veicolo) error {
	return db.mongoDB.CreateVeicolo(db.ctx, v)
}

func (db *DBCompat) GetVeicolo(id int) (*Veicolo, error) {
	return db.mongoDB.GetVeicolo(db.ctx, id)
}

func (db *DBCompat) UpdateVeicolo(v *Veicolo) error {
	return db.mongoDB.UpdateVeicolo(db.ctx, v)
}

func (db *DBCompat) DeleteVeicolo(id int) error {
	return db.mongoDB.DeleteVeicolo(db.ctx, id)
}

func (db *DBCompat) ListVeicoli() ([]Veicolo, error) {
	return db.mongoDB.ListVeicoli(db.ctx)
}

// ==================== COMMESSE ====================

func (db *DBCompat) CreateCommessa(c *Commessa) error {
	return db.mongoDB.CreateCommessa(db.ctx, c)
}

func (db *DBCompat) GetCommessa(id int) (*Commessa, error) {
	return db.mongoDB.GetCommessa(db.ctx, id)
}

func (db *DBCompat) UpdateCommessa(c *Commessa) error {
	return db.mongoDB.UpdateCommessa(db.ctx, c)
}

func (db *DBCompat) DeleteCommessa(id int) error {
	return db.mongoDB.DeleteCommessa(db.ctx, id)
}

func (db *DBCompat) ListCommesse() ([]Commessa, error) {
	return db.mongoDB.ListCommesse(db.ctx)
}

// ==================== APPUNTAMENTI ====================

func (db *DBCompat) CreateAppuntamento(a *Appuntamento) error {
	return db.mongoDB.CreateAppuntamento(db.ctx, a)
}

func (db *DBCompat) GetAppuntamento(id int) (*Appuntamento, error) {
	return db.mongoDB.GetAppuntamento(db.ctx, id)
}

func (db *DBCompat) UpdateAppuntamento(a *Appuntamento) error {
	return db.mongoDB.UpdateAppuntamento(db.ctx, a)
}

func (db *DBCompat) DeleteAppuntamento(id int) error {
	return db.mongoDB.DeleteAppuntamento(db.ctx, id)
}

func (db *DBCompat) ListAppuntamenti() ([]Appuntamento, error) {
	return db.mongoDB.ListAppuntamenti(db.ctx)
}

// ==================== OPERATORI ====================

func (db *DBCompat) CreateOperatore(o *Operatore) error {
	return db.mongoDB.CreateOperatore(db.ctx, o)
}

func (db *DBCompat) GetOperatore(id int) (*Operatore, error) {
	return db.mongoDB.GetOperatore(db.ctx, id)
}

func (db *DBCompat) UpdateOperatore(o *Operatore) error {
	return db.mongoDB.UpdateOperatore(db.ctx, o)
}

func (db *DBCompat) DeleteOperatore(id int) error {
	return db.mongoDB.DeleteOperatore(db.ctx, id)
}

func (db *DBCompat) ListOperatori() ([]Operatore, error) {
	return db.mongoDB.ListOperatori(db.ctx)
}

// ==================== PREVENTIVI ====================

func (db *DBCompat) CreatePreventivo(p *Preventivo) error {
	return db.mongoDB.CreatePreventivo(db.ctx, p)
}

func (db *DBCompat) GetPreventivo(id int) (*Preventivo, error) {
	return db.mongoDB.GetPreventivo(db.ctx, id)
}

func (db *DBCompat) UpdatePreventivo(p *Preventivo) error {
	return db.mongoDB.UpdatePreventivo(db.ctx, p)
}

func (db *DBCompat) DeletePreventivo(id int) error {
	return db.mongoDB.DeletePreventivo(db.ctx, id)
}

func (db *DBCompat) ListPreventivi() ([]Preventivo, error) {
	return db.mongoDB.ListPreventivi(db.ctx)
}

// ==================== FATTURE ====================

func (db *DBCompat) CreateFattura(f *Fattura) error {
	return db.mongoDB.CreateFattura(db.ctx, f)
}

func (db *DBCompat) GetFattura(id int) (*Fattura, error) {
	return db.mongoDB.GetFattura(db.ctx, id)
}

func (db *DBCompat) UpdateFattura(f *Fattura) error {
	return db.mongoDB.UpdateFattura(db.ctx, f)
}

func (db *DBCompat) DeleteFattura(id int) error {
	return db.mongoDB.DeleteFattura(db.ctx, id)
}

func (db *DBCompat) ListFatture() ([]Fattura, error) {
	return db.mongoDB.ListFatture(db.ctx)
}

// ==================== PRIMA NOTA ====================

func (db *DBCompat) CreateMovimento(m *MovimentoPrimaNota) error {
	return db.mongoDB.CreateMovimento(db.ctx, m)
}

func (db *DBCompat) GetMovimento(id int) (*MovimentoPrimaNota, error) {
	return db.mongoDB.GetMovimento(db.ctx, id)
}

func (db *DBCompat) UpdateMovimento(m *MovimentoPrimaNota) error {
	return db.mongoDB.UpdateMovimento(db.ctx, m)
}

func (db *DBCompat) DeleteMovimento(id int) error {
	return db.mongoDB.DeleteMovimento(db.ctx, id)
}

func (db *DBCompat) ListMovimenti() ([]MovimentoPrimaNota, error) {
	return db.mongoDB.ListMovimenti(db.ctx)
}

// ==================== HELPER FUNCTIONS ====================

func (db *DBCompat) GetVeicoliByCliente(clienteID int) ([]Veicolo, error) {
	return db.mongoDB.GetVeicoliByCliente(db.ctx, clienteID)
}

func (db *DBCompat) GetCommesseByVeicolo(veicoloID int) ([]Commessa, error) {
	return db.mongoDB.GetCommesseByVeicolo(db.ctx, veicoloID)
}

func (db *DBCompat) GetMovimentiByCommessa(commessaID int) ([]MovimentoPrimaNota, error) {
	return db.mongoDB.GetMovimentiByCommessa(db.ctx, commessaID)
}

// Count functions with compatibility layer
func (db *DBCompat) CountClienti() (int, error) {
	count, err := db.mongoDB.CountClienti(db.ctx)
	return int(count), err
}

func (db *DBCompat) CountVeicoli() (int, error) {
	count, err := db.mongoDB.CountVeicoli(db.ctx)
	return int(count), err
}

func (db *DBCompat) CountCommesse() (int, error) {
	count, err := db.mongoDB.CountCommesse(db.ctx)
	return int(count), err
}

func (db *DBCompat) CountCommesseAperte() (int, error) {
	count, err := db.mongoDB.CountCommesseAperte(db.ctx)
	return int(count), err
}

// Remove BackupManager for MongoDB (MongoDB has its own backup)
type BackupManager struct{}

func NewBackupManager(whatever interface{}, path string, maxFiles int) *BackupManager {
	logger.Info("Backup MongoDB non implementato - usare mongodump")
	return &BackupManager{}
}

func (bm *BackupManager) CreateBackup() (string, error) {
	return "", fmt.Errorf("backup non supportato - usare mongodump")
}
