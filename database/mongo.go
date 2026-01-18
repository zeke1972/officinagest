package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB represents a MongoDB database connection
type MongoDB struct {
	client     *mongo.Client
	database   *mongo.Database
	collection map[string]*mongo.Collection
}

// Collection names
const (
	CollClienti      = "clienti"
	CollFornitori    = "fornitori"
	CollVeicoli      = "veicoli"
	CollCommesse     = "commesse"
	CollAppuntamenti = "appuntamenti"
	CollOperatori    = "operatori"
	CollPreventivi   = "preventivi"
	CollFatture      = "fatture"
	CollPrimaNota    = "primanota"
)

// InitMongoDB initializes a new MongoDB connection
func InitMongoDB(ctx context.Context, uri string, dbName string) (*MongoDB, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("errore connessione MongoDB: %w", err)
	}

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("errore ping MongoDB: %w", err)
	}

	db := client.Database(dbName)

	// Create collections map
	collections := map[string]*mongo.Collection{
		CollClienti:      db.Collection(CollClienti),
		CollFornitori:    db.Collection(CollFornitori),
		CollVeicoli:      db.Collection(CollVeicoli),
		CollCommesse:     db.Collection(CollCommesse),
		CollAppuntamenti: db.Collection(CollAppuntamenti),
		CollOperatori:    db.Collection(CollOperatori),
		CollPreventivi:   db.Collection(CollPreventivi),
		CollFatture:      db.Collection(CollFatture),
		CollPrimaNota:    db.Collection(CollPrimaNota),
	}

	// Create unique indexes
	if err := createIndexes(ctx, collections); err != nil {
		return nil, fmt.Errorf("errore creazione indici: %w", err)
	}

	return &MongoDB{
		client:     client,
		database:   db,
		collection: collections,
	}, nil
}

func createIndexes(ctx context.Context, collections map[string]*mongo.Collection) error {
	// Unique indexes for natural keys
	indexes := []struct {
		collection string
		keys       bson.D
		unique     bool
	}{
		{CollVeicoli, bson.D{{Key: "targa", Value: 1}}, true},
		{CollOperatori, bson.D{{Key: "matricola", Value: 1}}, true},
	}

	for _, idx := range indexes {
		model := mongo.IndexModel{
			Keys:    idx.keys,
			Options: options.Index().SetUnique(idx.unique),
		}
		if _, err := collections[idx.collection].Indexes().CreateOne(ctx, model); err != nil {
			return err
		}
	}

	return nil
}

// Close disconnects from MongoDB
func (db *MongoDB) Close(ctx context.Context) error {
	return db.client.Disconnect(ctx)
}

// Helper to convert int ID to ObjectID
func intToObjectID(id int) primitive.ObjectID {
	// If ID is 0, generate new ObjectID
	if id == 0 {
		return primitive.NewObjectID()
	}
	// For migration compatibility, create deterministic ObjectID from int
	// This is a simplified approach - in production, consider storing original ID as separate field
	bytes := make([]byte, 12)
	for i := 0; i < 4 && i < len(bytes); i++ {
		bytes[i] = byte(id >> (24 - i*8))
	}
	return primitive.ObjectIDFromHex(fmt.Sprintf("%024x", id))
}

// ==================== CLIENTI ====================

func (db *MongoDB) CreateCliente(ctx context.Context, c *Cliente) error {
	if err := c.Validate(); err != nil {
		return fmt.Errorf("validazione fallita: %w", err)
	}

	c.ID = 0 // Reset ID for MongoDB
	result, err := db.collection[CollClienti].InsertOne(ctx, c)
	if err != nil {
		return fmt.Errorf("errore creazione cliente: %w", err)
	}

	// Update ID with inserted ID
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		c.ID = int(oid.Timestamp().Unix())
	}
	return nil
}

func (db *MongoDB) GetCliente(ctx context.Context, id int) (*Cliente, error) {
	var cliente Cliente
	oid := intToObjectID(id)
	
	err := db.collection[CollClienti].FindOne(ctx, bson.M{"_id": oid}).Decode(&cliente)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("cliente non trovato")
		}
		return nil, fmt.Errorf("errore ricerca cliente: %w", err)
	}
	
	cliente.ID = id
	return &cliente, nil
}

func (db *MongoDB) UpdateCliente(ctx context.Context, c *Cliente) error {
	if err := c.Validate(); err != nil {
		return fmt.Errorf("validazione fallita: %w", err)
	}

	oid := intToObjectID(c.ID)
	filter := bson.M{"_id": oid}
	update := bson.M{"$set": c}

	result, err := db.collection[CollClienti].UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("errore aggiornamento cliente: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("cliente #%d non trovato", c.ID)
	}

	return nil
}

func (db *MongoDB) DeleteCliente(ctx context.Context, id int) error {
	// Start transaction for cascade delete
	session, err := db.client.StartSession()
	if err != nil {
		return fmt.Errorf("errore avvio sessione: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Get all veicoli for this cliente
		veicoli, err := db.GetVeicoliByCliente(sessCtx, id)
		if err != nil {
			return nil, err
		}

		// Cascade delete: for each veicolo, delete related entities
		for _, veicolo := range veicoli {
			// Delete commesse for this veicolo
			commesse, err := db.GetCommesseByVeicolo(sessCtx, veicolo.ID)
			if err != nil {
				return nil, err
			}

			for _, commessa := range commesse {
				// Delete movimenti for this commessa
				_, err := db.collection[CollPrimaNota].DeleteMany(sessCtx, 
					bson.M{"commessa_id": commessa.ID})
				if err != nil {
					return nil, fmt.Errorf("errore eliminazione movimenti: %w", err)
				}
			}

			// Delete commesse
			_, err = db.collection[CollCommesse].DeleteMany(sessCtx, 
				bson.M{"veicolo_id": veicolo.ID})
			if err != nil {
				return nil, fmt.Errorf("errore eliminazione commesse: %w", err)
			}
		}

		// Delete all veicoli for this cliente
		_, err = db.collection[CollVeicoli].DeleteMany(sessCtx, bson.M{"cliente_id": id})
		if err != nil {
			return nil, fmt.Errorf("errore eliminazione veicoli: %w", err)
		}

		// Delete the cliente
		result, err := db.collection[CollClienti].DeleteOne(sessCtx, bson.M{"_id": intToObjectID(id)})
		if err != nil {
			return nil, fmt.Errorf("errore eliminazione cliente: %w", err)
		}

		if result.DeletedCount == 0 {
			return nil, fmt.Errorf("cliente #%d non trovato", id)
		}

		return nil, nil
	})

	return err
}

func (db *MongoDB) ListClienti(ctx context.Context) ([]Cliente, error) {
	cursor, err := db.collection[CollClienti].Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("errore ricerca clienti: %w", err)
	}
	defer cursor.Close(ctx)

	var clienti []Cliente
	if err := cursor.All(ctx, &clienti); err != nil {
		return nil, fmt.Errorf("errore decodifica clienti: %w", err)
	}

	return clienti, nil
}

// ==================== FORNITORI ====================

func (db *MongoDB) CreateFornitore(ctx context.Context, f *Fornitore) error {
	if err := f.Validate(); err != nil {
		return fmt.Errorf("validazione fallita: %w", err)
	}

	f.ID = 0 // Reset ID for MongoDB
	result, err := db.collection[CollFornitori].InsertOne(ctx, f)
	if err != nil {
		return fmt.Errorf("errore creazione fornitore: %w", err)
	}

	// Update ID with inserted ID
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		f.ID = int(oid.Timestamp().Unix())
	}
	return nil
}

func (db *MongoDB) GetFornitore(ctx context.Context, id int) (*Fornitore, error) {
	var fornitore Fornitore
	oid := intToObjectID(id)
	
	err := db.collection[CollFornitori].FindOne(ctx, bson.M{"_id": oid}).Decode(&fornitore)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("fornitore non trovato")
		}
		return nil, fmt.Errorf("errore ricerca fornitore: %w", err)
	}
	
	fornitore.ID = id
	return &fornitore, nil
}

func (db *MongoDB) UpdateFornitore(ctx context.Context, f *Fornitore) error {
	if err := f.Validate(); err != nil {
		return fmt.Errorf("validazione fallita: %w", err)
	}

	oid := intToObjectID(f.ID)
	filter := bson.M{"_id": oid}
	update := bson.M{"$set": f}

	result, err := db.collection[CollFornitori].UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("errore aggiornamento fornitore: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("fornitore #%d non trovato", f.ID)
	}

	return nil
}

func (db *MongoDB) DeleteFornitore(ctx context.Context, id int) error {
	session, err := db.client.StartSession()
	if err != nil {
		return fmt.Errorf("errore avvio sessione: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Delete related movimenti
		_, err := db.collection[CollPrimaNota].DeleteMany(sessCtx, bson.M{"fornitore_id": id})
		if err != nil {
			return nil, fmt.Errorf("errore eliminazione movimenti: %w", err)
		}

		// Delete fornitore
		result, err := db.collection[CollFornitori].DeleteOne(sessCtx, bson.M{"_id": intToObjectID(id)})
		if err != nil {
			return nil, fmt.Errorf("errore eliminazione fornitore: %w", err)
		}

		if result.DeletedCount == 0 {
			return nil, fmt.Errorf("fornitore #%d non trovato", id)
		}

		return nil, nil
	})

	return err
}

func (db *MongoDB) ListFornitori(ctx context.Context) ([]Fornitore, error) {
	cursor, err := db.collection[CollFornitori].Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("errore ricerca fornitori: %w", err)
	}
	defer cursor.Close(ctx)

	var fornitori []Fornitore
	if err := cursor.All(ctx, &fornitori); err != nil {
		return nil, fmt.Errorf("errore decodifica fornitori: %w", err)
	}

	return fornitori, nil
}

// ==================== VEICOLI ====================

func (db *MongoDB) CreateVeicolo(ctx context.Context, v *Veicolo) error {
	if err := v.Validate(); err != nil {
		return fmt.Errorf("validazione fallita: %w", err)
	}

	v.ID = 0
	result, err := db.collection[CollVeicoli].InsertOne(ctx, v)
	if err != nil {
		return fmt.Errorf("errore creazione veicolo: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		v.ID = int(oid.Timestamp().Unix())
	}
	return nil
}

func (db *MongoDB) GetVeicolo(ctx context.Context, id int) (*Veicolo, error) {
	var veicolo Veicolo
	oid := intToObjectID(id)
	
	err := db.collection[CollVeicoli].FindOne(ctx, bson.M{"_id": oid}).Decode(&veicolo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("veicolo non trovato")
		}
		return nil, fmt.Errorf("errore ricerca veicolo: %w", err)
	}
	
	veicolo.ID = id
	return &veicolo, nil
}

func (db *MongoDB) UpdateVeicolo(ctx context.Context, v *Veicolo) error {
	if err := v.Validate(); err != nil {
		return fmt.Errorf("validazione fallita: %w", err)
	}

	oid := intToObjectID(v.ID)
	filter := bson.M{"_id": oid}
	update := bson.M{"$set": v}

	result, err := db.collection[CollVeicoli].UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("errore aggiornamento veicolo: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("veicolo #%d non trovato", v.ID)
	}

	return nil
}

func (db *MongoDB) DeleteVeicolo(ctx context.Context, id int) error {
	session, err := db.client.StartSession()
	if err != nil {
		return fmt.Errorf("errore avvio sessione: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Get all commesse for this veicolo
		commesse, err := db.GetCommesseByVeicolo(sessCtx, id)
		if err != nil {
			return nil, err
		}

		for _, commessa := range commesse {
			// Delete movimenti for this commessa
			_, err := db.collection[CollPrimaNota].DeleteMany(sessCtx, 
				bson.M{"commessa_id": commessa.ID})
			if err != nil {
				return nil, fmt.Errorf("errore eliminazione movimenti: %w", err)
			}
		}

		// Delete commesse
		_, err = db.collection[CollCommesse].DeleteMany(sessCtx, 
			bson.M{"veicolo_id": id})
		if err != nil {
			return nil, fmt.Errorf("errore eliminazione commesse: %w", err)
		}

		// Delete veicolo
		result, err := db.collection[CollVeicoli].DeleteOne(sessCtx, bson.M{"_id": intToObjectID(id)})
		if err != nil {
			return nil, fmt.Errorf("errore eliminazione veicolo: %w", err)
		}

		if result.DeletedCount == 0 {
			return nil, fmt.Errorf("veicolo #%d non trovato", id)
		}

		return nil, nil
	})

	return err
}

func (db *MongoDB) ListVeicoli(ctx context.Context) ([]Veicolo, error) {
	cursor, err := db.collection[CollVeicoli].Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("errore ricerca veicoli: %w", err)
	}
	defer cursor.Close(ctx)

	var veicoli []Veicolo
	if err := cursor.All(ctx, &veicoli); err != nil {
		return nil, fmt.Errorf("errore decodifica veicoli: %w", err)
	}

	return veicoli, nil
}

// ==================== COMMESSE ====================

func (db *MongoDB) CreateCommessa(ctx context.Context, c *Commessa) error {
	if err := c.Validate(); err != nil {
		return fmt.Errorf("validazione fallita: %w", err)
	}

	c.DataApertura = time.Now()
	c.Totale = c.CostoManodopera + c.CostoRicambi
	c.ID = 0

	result, err := db.collection[CollCommesse].InsertOne(ctx, c)
	if err != nil {
		return fmt.Errorf("errore creazione commessa: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		c.ID = int(oid.Timestamp().Unix())
		c.Numero = fmt.Sprintf("COM-%04d", c.ID)
	}
	return nil
}

func (db *MongoDB) GetCommessa(ctx context.Context, id int) (*Commessa, error) {
	var commessa Commessa
	oid := intToObjectID(id)
	
	err := db.collection[CollCommesse].FindOne(ctx, bson.M{"_id": oid}).Decode(&commessa)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("commessa non trovata")
		}
		return nil, fmt.Errorf("errore ricerca commessa: %w", err)
	}
	
	commessa.ID = id
	return &commessa, nil
}

func (db *MongoDB) UpdateCommessa(ctx context.Context, c *Commessa) error {
	if err := c.Validate(); err != nil {
		return fmt.Errorf("validazione fallita: %w", err)
	}

	c.Totale = c.CostoManodopera + c.CostoRicambi

	if c.Stato == "Chiusa" && c.DataChiusura.IsZero() {
		c.DataChiusura = time.Now()
	}

	oid := intToObjectID(c.ID)
	filter := bson.M{"_id": oid}
	update := bson.M{"$set": c}

	result, err := db.collection[CollCommesse].UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("errore aggiornamento commessa: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("commessa #%d non trovata", c.ID)
	}

	return nil
}

func (db *MongoDB) DeleteCommessa(ctx context.Context, id int) error {
	session, err := db.client.StartSession()
	if err != nil {
		return fmt.Errorf("errore avvio sessione: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Delete related movimenti
		_, err := db.collection[CollPrimaNota].DeleteMany(sessCtx, bson.M{"commessa_id": id})
		if err != nil {
			return nil, fmt.Errorf("errore eliminazione movimenti: %w", err)
		}

		// Delete commessa
		result, err := db.collection[CollCommesse].DeleteOne(sessCtx, bson.M{"_id": intToObjectID(id)})
		if err != nil {
			return nil, fmt.Errorf("errore eliminazione commessa: %w", err)
		}

		if result.DeletedCount == 0 {
			return nil, fmt.Errorf("commessa #%d non trovata", id)
		}

		return nil, nil
	})

	return err
}

func (db *MongoDB) ListCommesse(ctx context.Context) ([]Commessa, error) {
	cursor, err := db.collection[CollCommesse].Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("errore ricerca commesse: %w", err)
	}
	defer cursor.Close(ctx)

	var commesse []Commessa
	if err := cursor.All(ctx, &commesse); err != nil {
		return nil, fmt.Errorf("errore decodifica commesse: %w", err)
	}

	return commesse, nil
}

// ==================== APPUNTAMENTI ====================

func (db *MongoDB) CreateAppuntamento(ctx context.Context, a *Appuntamento) error {
	a.ID = 0
	result, err := db.collection[CollAppuntamenti].InsertOne(ctx, a)
	if err != nil {
		return fmt.Errorf("errore creazione appuntamento: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		a.ID = int(oid.Timestamp().Unix())
	}
	return nil
}

func (db *MongoDB) GetAppuntamento(ctx context.Context, id int) (*Appuntamento, error) {
	var appuntamento Appuntamento
	oid := intToObjectID(id)
	
	err := db.collection[CollAppuntamenti].FindOne(ctx, bson.M{"_id": oid}).Decode(&appuntamento)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("appuntamento non trovato")
		}
		return nil, fmt.Errorf("errore ricerca appuntamento: %w", err)
	}
	
	appuntamento.ID = id
	return &appuntamento, nil
}

func (db *MongoDB) UpdateAppuntamento(ctx context.Context, a *Appuntamento) error {
	oid := intToObjectID(a.ID)
	filter := bson.M{"_id": oid}
	update := bson.M{"$set": a}

	result, err := db.collection[CollAppuntamenti].UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("errore aggiornamento appuntamento: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("appuntamento #%d non trovato", a.ID)
	}

	return nil
}

func (db *MongoDB) DeleteAppuntamento(ctx context.Context, id int) error {
	result, err := db.collection[CollAppuntamenti].DeleteOne(ctx, bson.M{"_id": intToObjectID(id)})
	if err != nil {
		return fmt.Errorf("errore eliminazione appuntamento: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("appuntamento #%d non trovato", id)
	}

	return nil
}

func (db *MongoDB) ListAppuntamenti(ctx context.Context) ([]Appuntamento, error) {
	cursor, err := db.collection[CollAppuntamenti].Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("errore ricerca appuntamenti: %w", err)
	}
	defer cursor.Close(ctx)

	var appuntamenti []Appuntamento
	if err := cursor.All(ctx, &appuntamenti); err != nil {
		return nil, fmt.Errorf("errore decodifica appuntamenti: %w", err)
	}

	return appuntamenti, nil
}

// ==================== OPERATORI ====================

func (db *MongoDB) CreateOperatore(ctx context.Context, o *Operatore) error {
	o.ID = 0
	result, err := db.collection[CollOperatori].InsertOne(ctx, o)
	if err != nil {
		return fmt.Errorf("errore creazione operatore: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		o.ID = int(oid.Timestamp().Unix())
	}
	return nil
}

func (db *MongoDB) GetOperatore(ctx context.Context, id int) (*Operatore, error) {
	var operatore Operatore
	oid := intToObjectID(id)
	
	err := db.collection[CollOperatori].FindOne(ctx, bson.M{"_id": oid}).Decode(&operatore)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("operatore non trovato")
		}
		return nil, fmt.Errorf("errore ricerca operatore: %w", err)
	}
	
	operatore.ID = id
	return &operatore, nil
}

func (db *MongoDB) UpdateOperatore(ctx context.Context, o *Operatore) error {
	oid := intToObjectID(o.ID)
	filter := bson.M{"_id": oid}
	update := bson.M{"$set": o}

	result, err := db.collection[CollOperatori].UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("errore aggiornamento operatore: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("operatore #%d non trovato", o.ID)
	}

	return nil
}

func (db *MongoDB) DeleteOperatore(ctx context.Context, id int) error {
	result, err := db.collection[CollOperatori].DeleteOne(ctx, bson.M{"_id": intToObjectID(id)})
	if err != nil {
		return fmt.Errorf("errore eliminazione operatore: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("operatore #%d non trovato", id)
	}

	return nil
}

func (db *MongoDB) ListOperatori(ctx context.Context) ([]Operatore, error) {
	cursor, err := db.collection[CollOperatori].Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("errore ricerca operatori: %w", err)
	}
	defer cursor.Close(ctx)

	var operatori []Operatore
	if err := cursor.All(ctx, &operatori); err != nil {
		return nil, fmt.Errorf("errore decodifica operatori: %w", err)
	}

	return operatori, nil
}

// ==================== PREVENTIVI ====================

func (db *MongoDB) CreatePreventivo(ctx context.Context, p *Preventivo) error {
	p.Data = time.Now()
	p.Accettato = false
	p.ID = 0

	result, err := db.collection[CollPreventivi].InsertOne(ctx, p)
	if err != nil {
		return fmt.Errorf("errore creazione preventivo: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		p.ID = int(oid.Timestamp().Unix())
		p.Numero = fmt.Sprintf("PREV-%04d", p.ID)
	}
	return nil
}

func (db *MongoDB) GetPreventivo(ctx context.Context, id int) (*Preventivo, error) {
	var preventivo Preventivo
	oid := intToObjectID(id)
	
	err := db.collection[CollPreventivi].FindOne(ctx, bson.M{"_id": oid}).Decode(&preventivo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("preventivo non trovato")
		}
		return nil, fmt.Errorf("errore ricerca preventivo: %w", err)
	}
	
	preventivo.ID = id
	return &preventivo, nil
}

func (db *MongoDB) UpdatePreventivo(ctx context.Context, p *Preventivo) error {
	oid := intToObjectID(p.ID)
	filter := bson.M{"_id": oid}
	update := bson.M{"$set": p}

	result, err := db.collection[CollPreventivi].UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("errore aggiornamento preventivo: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("preventivo #%d non trovato", p.ID)
	}

	return nil
}

func (db *MongoDB) DeletePreventivo(ctx context.Context, id int) error {
	result, err := db.collection[CollPreventivi].DeleteOne(ctx, bson.M{"_id": intToObjectID(id)})
	if err != nil {
		return fmt.Errorf("errore eliminazione preventivo: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("preventivo #%d non trovato", id)
	}

	return nil
}

func (db *MongoDB) ListPreventivi(ctx context.Context) ([]Preventivo, error) {
	cursor, err := db.collection[CollPreventivi].Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("errore ricerca preventivi: %w", err)
	}
	defer cursor.Close(ctx)

	var preventivi []Preventivo
	if err := cursor.All(ctx, &preventivi); err != nil {
		return nil, fmt.Errorf("errore decodifica preventivi: %w", err)
	}

	return preventivi, nil
}

// ==================== FATTURE ====================

func (db *MongoDB) CreateFattura(ctx context.Context, f *Fattura) error {
	f.ID = 0
	result, err := db.collection[CollFatture].InsertOne(ctx, f)
	if err != nil {
		return fmt.Errorf("errore creazione fattura: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		f.ID = int(oid.Timestamp().Unix())
		f.Numero = fmt.Sprintf("FT-%04d/%d", f.ID, f.Data.Year())
	}
	return nil
}

func (db *MongoDB) GetFattura(ctx context.Context, id int) (*Fattura, error) {
	var fattura Fattura
	oid := intToObjectID(id)
	
	err := db.collection[CollFatture].FindOne(ctx, bson.M{"_id": oid}).Decode(&fattura)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("fattura non trovata")
		}
		return nil, fmt.Errorf("errore ricerca fattura: %w", err)
	}
	
	fattura.ID = id
	return &fattura, nil
}

func (db *MongoDB) UpdateFattura(ctx context.Context, f *Fattura) error {
	oid := intToObjectID(f.ID)
	filter := bson.M{"_id": oid}
	update := bson.M{"$set": f}

	result, err := db.collection[CollFatture].UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("errore aggiornamento fattura: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("fattura #%d non trovata", f.ID)
	}

	return nil
}

func (db *MongoDB) DeleteFattura(ctx context.Context, id int) error {
	result, err := db.collection[CollFatture].DeleteOne(ctx, bson.M{"_id": intToObjectID(id)})
	if err != nil {
		return fmt.Errorf("errore eliminazione fattura: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("fattura #%d non trovata", id)
	}

	return nil
}

func (db *MongoDB) ListFatture(ctx context.Context) ([]Fattura, error) {
	cursor, err := db.collection[CollFatture].Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("errore ricerca fatture: %w", err)
	}
	defer cursor.Close(ctx)

	var fatture []Fattura
	if err := cursor.All(ctx, &fatture); err != nil {
		return nil, fmt.Errorf("errore decodifica fatture: %w", err)
	}

	return fatture, nil
}

// ==================== PRIMA NOTA ====================

func (db *MongoDB) CreateMovimento(ctx context.Context, m *MovimentoPrimaNota) error {
	if err := m.Validate(); err != nil {
		return fmt.Errorf("validazione fallita: %w", err)
	}

	m.ID = 0
	result, err := db.collection[CollPrimaNota].InsertOne(ctx, m)
	if err != nil {
		return fmt.Errorf("errore creazione movimento: %w", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		m.ID = int(oid.Timestamp().Unix())
	}
	return nil
}

func (db *MongoDB) GetMovimento(ctx context.Context, id int) (*MovimentoPrimaNota, error) {
	var movimento MovimentoPrimaNota
	oid := intToObjectID(id)
	
	err := db.collection[CollPrimaNota].FindOne(ctx, bson.M{"_id": oid}).Decode(&movimento)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("movimento non trovato")
		}
		return nil, fmt.Errorf("errore ricerca movimento: %w", err)
	}
	
	movimento.ID = id
	return &movimento, nil
}

func (db *MongoDB) UpdateMovimento(ctx context.Context, m *MovimentoPrimaNota) error {
	if err := m.Validate(); err != nil {
		return fmt.Errorf("validazione fallita: %w", err)
	}

	oid := intToObjectID(m.ID)
	filter := bson.M{"_id": oid}
	update := bson.M{"$set": m}

	result, err := db.collection[CollPrimaNota].UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("errore aggiornamento movimento: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("movimento #%d non trovato", m.ID)
	}

	return nil
}

func (db *MongoDB) DeleteMovimento(ctx context.Context, id int) error {
	result, err := db.collection[CollPrimaNota].DeleteOne(ctx, bson.M{"_id": intToObjectID(id)})
	if err != nil {
		return fmt.Errorf("errore eliminazione movimento: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("movimento #%d non trovato", id)
	}

	return nil
}

func (db *MongoDB) ListMovimenti(ctx context.Context) ([]MovimentoPrimaNota, error) {
	cursor, err := db.collection[CollPrimaNota].Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("errore ricerca movimenti: %w", err)
	}
	defer cursor.Close(ctx)

	var movimenti []MovimentoPrimaNota
	if err := cursor.All(ctx, &movimenti); err != nil {
		return nil, fmt.Errorf("errore decodifica movimenti: %w", err)
	}

	return movimenti, nil
}

// ==================== HELPER FUNCTIONS ====================

func (db *MongoDB) GetVeicoliByCliente(ctx context.Context, clienteID int) ([]Veicolo, error) {
	cursor, err := db.collection[CollVeicoli].Find(ctx, bson.M{"cliente_id": clienteID})
	if err != nil {
		return nil, fmt.Errorf("errore ricerca veicoli: %w", err)
	}
	defer cursor.Close(ctx)

	var veicoli []Veicolo
	if err := cursor.All(ctx, &veicoli); err != nil {
		return nil, fmt.Errorf("errore decodifica veicoli: %w", err)
	}

	return veicoli, nil
}

func (db *MongoDB) GetCommesseByVeicolo(ctx context.Context, veicoloID int) ([]Commessa, error) {
	cursor, err := db.collection[CollCommesse].Find(ctx, bson.M{"veicolo_id": veicoloID})
	if err != nil {
		return nil, fmt.Errorf("errore ricerca commesse: %w", err)
	}
	defer cursor.Close(ctx)

	var commesse []Commessa
	if err := cursor.All(ctx, &commesse); err != nil {
		return nil, fmt.Errorf("errore decodifica commesse: %w", err)
	}

	return commesse, nil
}

func (db *MongoDB) GetMovimentiByCommessa(ctx context.Context, commessaID int) ([]MovimentoPrimaNota, error) {
	cursor, err := db.collection[CollPrimaNota].Find(ctx, bson.M{"commessa_id": commessaID})
	if err != nil {
		return nil, fmt.Errorf("errore ricerca movimenti: %w", err)
	}
	defer cursor.Close(ctx)

	var movimenti []MovimentoPrimaNota
	if err := cursor.All(ctx, &movimenti); err != nil {
		return nil, fmt.Errorf("errore decodifica movimenti: %w", err)
	}

	return movimenti, nil
}

// Count functions
func (db *MongoDB) CountClienti(ctx context.Context) (int64, error) {
	count, err := db.collection[CollClienti].CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("errore conteggio clienti: %w", err)
	}
	return count, nil
}

func (db *MongoDB) CountVeicoli(ctx context.Context) (int64, error) {
	count, err := db.collection[CollVeicoli].CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("errore conteggio veicoli: %w", err)
	}
	return count, nil
}

func (db *MongoDB) CountCommesse(ctx context.Context) (int64, error) {
	count, err := db.collection[CollCommesse].CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("errore conteggio commesse: %w", err)
	}
	return count, nil
}

func (db *MongoDB) CountCommesseAperte(ctx context.Context) (int64, error) {
	count, err := db.collection[CollCommesse].CountDocuments(ctx, bson.M{"stato": "Aperta"})
	if err != nil {
		return 0, fmt.Errorf("errore conteggio commesse aperte: %w", err)
	}
	return count, nil
}
