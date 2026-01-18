package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB è l'interfaccia compatibile verso l'esterno
type DB struct {
	mongo *MongoDB
}

// InitMongoDB inizializza il database MongoDB (usato da main.go)
func InitMongoDB(uri, dbName string) (*DB, error) {
	mongo, err := NewMongoDB(uri, dbName)
	if err != nil {
		return nil, err
	}

	return &DB{mongo: mongo}, nil
}

// Close chiude la connessione MongoDB
func (db *DB) Close() error {
	return db.mongo.Close()
}

// ==================== CLIENTI ====================

func (db *DB) CreateCliente(c *Cliente) error {
	return db.mongo.CreateCliente(c)
}

func (db *DB) GetCliente(id int) (*Cliente, error) {
	return db.mongo.GetCliente(id)
}

func (db *DB) UpdateCliente(c *Cliente) error {
	return db.mongo.UpdateCliente(c)
}

func (db *DB) DeleteCliente(id int) error {
	return db.mongo.DeleteCliente(id)
}

func (db *DB) ListClienti() ([]Cliente, error) {
	return db.mongo.ListClienti()
}

// ==================== FORNITORI ====================

func (db *DB) CreateFornitore(f *Fornitore) error {
	return db.mongo.CreateFornitore(f)
}

func (db *DB) GetFornitore(id int) (*Fornitore, error) {
	return db.mongo.GetFornitore(id)
}

func (db *DB) UpdateFornitore(f *Fornitore) error {
	return db.mongo.UpdateFornitore(f)
}

func (db *DB) DeleteFornitore(id int) error {
	return db.mongo.DeleteFornitore(id)
}

func (db *DB) ListFornitori() ([]Fornitore, error) {
	return db.mongo.ListFornitori()
}

// ==================== VEICOLI ====================

func (db *DB) CreateVeicolo(v *Veicolo) error {
	return db.mongo.CreateVeicolo(v)
}

func (db *DB) GetVeicolo(id int) (*Veicolo, error) {
	return db.mongo.GetVeicolo(id)
}

func (db *DB) UpdateVeicolo(v *Veicolo) error {
	return db.mongo.UpdateVeicolo(v)
}

func (db *DB) DeleteVeicolo(id int) error {
	return db.mongo.DeleteVeicolo(id)
}

func (db *DB) ListVeicoli() ([]Veicolo, error) {
	return db.mongo.ListVeicoli()
}

// ==================== COMMESSE ====================

func (db *DB) CreateCommessa(c *Commessa) error {
	return db.mongo.CreateCommessa(c)
}

func (db *DB) GetCommessa(id int) (*Commessa, error) {
	return db.mongo.GetCommessa(id)
}

func (db *DB) UpdateCommessa(c *Commessa) error {
	return db.mongo.UpdateCommessa(c)
}

func (db *DB) DeleteCommessa(id int) error {
	return db.mongo.DeleteCommessa(id)
}

func (db *DB) ListCommesse() ([]Commessa, error) {
	return db.mongo.ListCommesse(map[string]interface{}{})
}

// ==================== APPUNTAMENTI ====================

func (db *DB) CreateAppuntamento(a *Appuntamento) error {
	return db.mongo.CreateAppuntamento(a)
}

func (db *DB) GetAppuntamento(id int) (*Appuntamento, error) {
	return db.mongo.GetAppuntamento(id)
}

func (db *DB) UpdateAppuntamento(a *Appuntamento) error {
	return db.mongo.UpdateAppuntamento(a)
}

func (db *DB) DeleteAppuntamento(id int) error {
	return db.mongo.DeleteAppuntamento(id)
}

func (db *DB) ListAppuntamenti() ([]Appuntamento, error) {
	return db.mongo.ListAppuntamenti(map[string]interface{}{})
}

func (db *DB) ListAppuntamentiByDate(date time.Time) ([]Appuntamento, error) {
	return db.mongo.ListAppuntamenti(map[string]interface{}{"data": date})
}

// ==================== OPERATORI ====================

func (db *DB) CreateOperatore(o *Operatore) error {
	return db.mongo.CreateOperatore(o)
}

func (db *DB) GetOperatore(id int) (*Operatore, error) {
	return db.mongo.GetOperatore(id)
}

func (db *DB) UpdateOperatore(o *Operatore) error {
	return db.mongo.UpdateOperatore(o)
}

func (db *DB) DeleteOperatore(id int) error {
	return db.mongo.DeleteOperatore(id)
}

func (db *DB) ListOperatori() ([]Operatore, error) {
	return db.mongo.ListOperatori()
}

// ==================== PREVENTIVI ====================

func (db *DB) CreatePreventivo(p *Preventivo) error {
	return db.mongo.CreatePreventivo(p)
}

func (db *DB) GetPreventivo(id int) (*Preventivo, error) {
	return db.mongo.GetPreventivo(id)
}

func (db *DB) UpdatePreventivo(p *Preventivo) error {
	return db.mongo.UpdatePreventivo(p)
}

func (db *DB) DeletePreventivo(id int) error {
	return db.mongo.DeletePreventivo(id)
}

func (db *DB) ListPreventivi() ([]Preventivo, error) {
	return db.mongo.ListPreventivi()
}

// ==================== FATTURE ====================

func (db *DB) CreateFattura(f *Fattura) error {
	return db.mongo.CreateFattura(f)
}

func (db *DB) GetFattura(id int) (*Fattura, error) {
	return db.mongo.GetFattura(id)
}

func (db *DB) UpdateFattura(f *Fattura) error {
	return db.mongo.UpdateFattura(f)
}

func (db *DB) DeleteFattura(id int) error {
	return db.mongo.DeleteFattura(id)
}

func (db *DB) ListFatture() ([]Fattura, error) {
	return db.mongo.ListFatture()
}

// ==================== MOVIMENTI PRIMA NOTA ====================

func (db *DB) CreateMovimentoPrimaNota(mov *MovimentoPrimaNota) error {
	return db.mongo.CreateMovimentoPrimaNota(mov)
}

func (db *DB) GetMovimentoPrimaNota(id int) (*MovimentoPrimaNota, error) {
	return db.mongo.GetMovimentoPrimaNota(id)
}

func (db *DB) UpdateMovimentoPrimaNota(mov *MovimentoPrimaNota) error {
	return db.mongo.UpdateMovimentoPrimaNota(mov)
}

func (db *DB) DeleteMovimentoPrimaNota(id int) error {
	return db.mongo.DeleteMovimentoPrimaNota(id)
}

func (db *DB) ListMovimentiPrimaNota(filters map[string]interface{}) ([]MovimentoPrimaNota, error) {
	return db.mongo.ListMovimentiPrimaNota(filters)
}

// ==================== QUERY AGGREGATE ====================

func (db *DB) GetVeicoliByCliente(clienteID int) ([]Veicolo, error) {
	return db.mongo.GetVeicoliByCliente(clienteID)
}

func (db *DB) GetCommesseStats() (aperte int, chiuse int, err error) {
	return db.mongo.GetCommesseStats()
}

func (db *DB) GetPrimaNotaStats(anno int) (entrata float64, uscita float64, err error) {
	return db.mongo.GetPrimaNotaStats(anno)
}

// ==================== EXPORT ====================

func (db *DB) ExportToJSON(collection string) ([]byte, error) {
	ctx := context.Background()

	var cursor *mongo.Cursor
	var err error

	switch collection {
	case "clienti":
		cursor, err = db.mongo.db.Collection("clienti").Find(ctx, bson.M{})
	case "fornitori":
		cursor, err = db.mongo.db.Collection("fornitori").Find(ctx, bson.M{})
	case "veicoli":
		cursor, err = db.mongo.db.Collection("veicoli").Find(ctx, bson.M{})
	case "commesse":
		cursor, err = db.mongo.db.Collection("commesse").Find(ctx, bson.M{})
	case "appuntamenti":
		cursor, err = db.mongo.db.Collection("appuntamenti").Find(ctx, bson.M{})
	case "operatori":
		cursor, err = db.mongo.db.Collection("operatori").Find(ctx, bson.M{})
	case "preventivi":
		cursor, err = db.mongo.db.Collection("preventivi").Find(ctx, bson.M{})
	case "fatture":
		cursor, err = db.mongo.db.Collection("fatture").Find(ctx, bson.M{})
	case "movimenti_primanota":
		cursor, err = db.mongo.db.Collection("movimenti_primanota").Find(ctx, bson.M{})
	default:
		return nil, fmt.Errorf("collezione sconosciuta: %s", collection)
	}

	if err != nil {
		return nil, fmt.Errorf("errore query export: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("errore decodifica export: %w", err)
	}

	data, err := bson.MarshalExtJSON(results, true, true)
	if err != nil {
		return nil, fmt.Errorf("errore serializzazione JSON: %w", err)
	}

	return data, nil
}
