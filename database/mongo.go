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

// MongoDB
type MongoDB struct {
	client *mongo.Client
	db     *mongo.Database
	ctx    context.Context
}

// NewMongoDB crea una nuova connessione MongoDB
func NewMongoDB(uri, dbName string) (*MongoDB, error) {
	ctx := context.Background()

	opts := options.Client().ApplyURI(uri)
	opts.SetConnectTimeout(5 * time.Second)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("errore connessione mongo: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("errore ping mongo: %w", err)
	}

	db := client.Database(dbName)

	// Creazione indici
	if err := setupIndexes(ctx, db); err != nil {
		return nil, fmt.Errorf("errore setup indici: %w", err)
	}

	return &MongoDB{
		client: client,
		db:     db,
		ctx:    ctx,
	}, nil
}

func setupIndexes(ctx context.Context, db *mongo.Database) error {
	// Indici per query comuni
	indexes := map[string][]mongo.IndexModel{
		"clienti": {
			{Keys: bson.D{{Key: "ragione_sociale", Value: 1}}},
			{Keys: bson.D{{Key: "partita_iva", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true)},
		},
		"veicoli": {
			{Keys: bson.D{{Key: "cliente_id", Value: 1}}},
			{Keys: bson.D{{Key: "targa", Value: 1}}, Options: options.Index().SetUnique(true)},
		},
		"commesse": {
			{Keys: bson.D{{Key: "veicolo_id", Value: 1}}},
			{Keys: bson.D{{Key: "stato", Value: 1}}},
			{Keys: bson.D{{Key: "data_apertura", Value: -1}}},
		},
		"movimenti_primanota": {
			{Keys: bson.D{{Key: "commessa_id", Value: 1}}},
			{Keys: bson.D{{Key: "fornitore_id", Value: 1}}},
			{Keys: bson.D{{Key: "data", Value: -1}}},
		},
		"appuntamenti": {
			{Keys: bson.D{{Key: "data_ora", Value: 1}}},
			{Keys: bson.D{{Key: "veicolo_id", Value: 1}}},
		},
		"fornitori": {
			{Keys: bson.D{{Key: "ragione_sociale", Value: 1}}},
			{Keys: bson.D{{Key: "partita_iva", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true)},
		},
	}

	for collection, idxs := range indexes {
		if _, err := db.Collection(collection).Indexes().CreateMany(ctx, idxs); err != nil {
			return fmt.Errorf("errore creazione indici per %s: %w", collection, err)
		}
	}

	return nil
}

// Close chiude la connessione
func (m *MongoDB) Close() error {
	return m.client.Disconnect(m.ctx)
}

// generaID genera un ObjectID compatibile con INT legacy
func generaID() int {
	return int(primitive.NewObjectID().Timestamp().Unix())
}

// ==================== CLIENTI ====================

func (m *MongoDB) CreateCliente(c *Cliente) error {
	c.ID = generaID()
	_, err := m.db.Collection("clienti").InsertOne(m.ctx, c)
	return err
}

func (m *MongoDB) GetCliente(id int) (*Cliente, error) {
	var c Cliente
	err := m.db.Collection("clienti").FindOne(m.ctx, bson.M{"id": id}).Decode(&c)
	if err != nil {
		return nil, fmt.Errorf("cliente non trovato: %w", err)
	}
	return &c, nil
}

func (m *MongoDB) UpdateCliente(c *Cliente) error {
	result := m.db.Collection("clienti").FindOneAndReplace(m.ctx, bson.M{"id": c.ID}, c)
	if result.Err() != nil {
		return fmt.Errorf("cliente #%d non trovato", c.ID)
	}
	return nil
}

func (m *MongoDB) DeleteCliente(id int) error {
	// Cascade deletion
	return m.db.Client().UseSession(m.ctx, func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			return err
		}

		defer sessionContext.EndSession(m.ctx)

		// Trova veicoli
		var veicoliIDs []int
		cursor, err := m.db.Collection("veicoli").Find(sessionContext, bson.M{"cliente_id": id})
		if err != nil {
			sessionContext.AbortTransaction(sessionContext)
			return err
		}

		var veicoli []Veicolo
		if err := cursor.All(sessionContext, &veicoli); err != nil {
			sessionContext.AbortTransaction(sessionContext)
			return err
		}

		for _, v := range veicoli {
			veicoliIDs = append(veicoliIDs, v.ID)
		}

		// Trova commesse
		var commesseIDs []int
		cursor, err = m.db.Collection("commesse").Find(sessionContext, bson.M{"veicolo_id": bson.M{"$in": veicoliIDs}})
		if err != nil {
			sessionContext.AbortTransaction(sessionContext)
			return err
		}

		var commesse []Commessa
		if err := cursor.All(sessionContext, &commesse); err != nil {
			sessionContext.AbortTransaction(sessionContext)
			return err
		}

		for _, c := range commesse {
			commesseIDs = append(commesseIDs, c.ID)
		}

		// Elimina movimenti
		if _, err := m.db.Collection("movimenti_primanota").DeleteMany(sessionContext, bson.M{"commessa_id": bson.M{"$in": commesseIDs}}); err != nil {
			sessionContext.AbortTransaction(sessionContext)
			return err
		}

		// Elimina commesse
		if _, err := m.db.Collection("commesse").DeleteMany(sessionContext, bson.M{"veicolo_id": bson.M{"$in": veicoliIDs}}); err != nil {
			sessionContext.AbortTransaction(sessionContext)
			return err
		}

		// Elimina veicoli
		if _, err := m.db.Collection("veicoli").DeleteMany(sessionContext, bson.M{"cliente_id": id}); err != nil {
			sessionContext.AbortTransaction(sessionContext)
			return err
		}

		// Elimina cliente
		if _, err := m.db.Collection("clienti").DeleteOne(sessionContext, bson.M{"id": id}); err != nil {
			sessionContext.AbortTransaction(sessionContext)
			return err
		}

		return sessionContext.CommitTransaction(sessionContext)
	})
}

func (m *MongoDB) ListClienti() ([]Cliente, error) {
	var list []Cliente
	cursor, err := m.db.Collection("clienti").Find(m.ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "ragione_sociale", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(m.ctx)
	return list, cursor.All(m.ctx, &list)
}

// ==================== FORNITORI ====================

func (m *MongoDB) CreateFornitore(f *Fornitore) error {
	f.ID = generaID()
	_, err := m.db.Collection("fornitori").InsertOne(m.ctx, f)
	return err
}

func (m *MongoDB) GetFornitore(id int) (*Fornitore, error) {
	var f Fornitore
	err := m.db.Collection("fornitori").FindOne(m.ctx, bson.M{"id": id}).Decode(&f)
	if err != nil {
		return nil, fmt.Errorf("fornitore non trovato: %w", err)
	}
	return &f, nil
}

func (m *MongoDB) UpdateFornitore(f *Fornitore) error {
	result := m.db.Collection("fornitori").FindOneAndReplace(m.ctx, bson.M{"id": f.ID}, f)
	if result.Err() != nil {
		return fmt.Errorf("fornitore #%d non trovato", f.ID)
	}
	return nil
}

func (m *MongoDB) DeleteFornitore(id int) error {
	// Cascade movimenti
	return m.db.Client().UseSession(m.ctx, func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			return err
		}

		if _, err := m.db.Collection("movimenti_primanota").DeleteMany(sessionContext, bson.M{"fornitore_id": id}); err != nil {
			sessionContext.AbortTransaction(sessionContext)
			return err
		}

		if _, err := m.db.Collection("fornitori").DeleteOne(sessionContext, bson.M{"id": id}); err != nil {
			sessionContext.AbortTransaction(sessionContext)
			return err
		}

		return sessionContext.CommitTransaction(sessionContext)
	})
}

func (m *MongoDB) ListFornitori() ([]Fornitore, error) {
	var list []Fornitore
	cursor, err := m.db.Collection("fornitori").Find(m.ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "ragione_sociale", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(m.ctx)
	return list, cursor.All(m.ctx, &list)
}

// ==================== VEICOLI ====================

func (m *MongoDB) CreateVeicolo(v *Veicolo) error {
	v.ID = generaID()
	_, err := m.db.Collection("veicoli").InsertOne(m.ctx, v)
	return err
}

func (m *MongoDB) GetVeicolo(id int) (*Veicolo, error) {
	var v Veicolo
	err := m.db.Collection("veicoli").FindOne(m.ctx, bson.M{"id": id}).Decode(&v)
	if err != nil {
		return nil, fmt.Errorf("veicolo non trovato: %w", err)
	}
	return &v, nil
}

func (m *MongoDB) UpdateVeicolo(v *Veicolo) error {
	result := m.db.Collection("veicoli").FindOneAndReplace(m.ctx, bson.M{"id": v.ID}, v)
	if result.Err() != nil {
		return fmt.Errorf("veicolo #%d non trovato", v.ID)
	}
	return nil
}

func (m *MongoDB) DeleteVeicolo(id int) error {
	return m.db.Client().UseSession(m.ctx, func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			return err
		}

		// Trova commesse
		var commesseIDs []int
		cursor, err := m.db.Collection("commesse").Find(sessionContext, bson.M{"veicolo_id": id})
		if err != nil {
			sessionContext.AbortTransaction(sessionContext)
			return err
		}

		var commesse []Commessa
		if err := cursor.All(sessionContext, &commesse); err != nil {
			sessionContext.AbortTransaction(sessionContext)
			return err
		}

		for _, c := range commesse {
			commesseIDs = append(commesseIDs, c.ID)
		}

		// Elimina movimenti
		if _, err := m.db.Collection("movimenti_primanota").DeleteMany(sessionContext, bson.M{"commessa_id": bson.M{"$in": commesseIDs}}); err != nil {
			sessionContext.AbortTransaction(sessionContext)
			return err
		}

		// Elimina commesse
		if _, err := m.db.Collection("commesse").DeleteMany(sessionContext, bson.M{"veicolo_id": id}); err != nil {
			sessionContext.AbortTransaction(sessionContext)
			return err
		}

		// Elimina veicolo
		if _, err := m.db.Collection("veicoli").DeleteOne(sessionContext, bson.M{"id": id}); err != nil {
			sessionContext.AbortTransaction(sessionContext)
			return err
		}

		return sessionContext.CommitTransaction(sessionContext)
	})
}

func (m *MongoDB) ListVeicoli() ([]Veicolo, error) {
	var list []Veicolo
	cursor, err := m.db.Collection("veicoli").Find(m.ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "marca", Value: 1}, {Key: "modello", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(m.ctx)
	return list, cursor.All(m.ctx, &list)
}

// ==================== COMMESSE ====================

func (m *MongoDB) CreateCommessa(c *Commessa) error {
	c.ID = generaID()
	c.Numero = fmt.Sprintf("COM-%04d", c.ID)
	c.DataApertura = time.Now()
	c.Totale = c.CostoManodopera + c.CostoRicambi

	_, err := m.db.Collection("commesse").InsertOne(m.ctx, c)
	return err
}

func (m *MongoDB) GetCommessa(id int) (*Commessa, error) {
	var c Commessa
	err := m.db.Collection("commesse").FindOne(m.ctx, bson.M{"id": id}).Decode(&c)
	if err != nil {
		return nil, fmt.Errorf("commessa non trovata: %w", err)
	}
	return &c, nil
}

func (m *MongoDB) UpdateCommessa(c *Commessa) error {
	c.Totale = c.CostoManodopera + c.CostoRicambi
	if c.Stato == "Chiusa" && c.DataChiusura.IsZero() {
		c.DataChiusura = time.Now()
	}

	result := m.db.Collection("commesse").FindOneAndReplace(m.ctx, bson.M{"id": c.ID}, c)
	if result.Err() != nil {
		return fmt.Errorf("commessa #%d non trovata", c.ID)
	}
	return nil
}

func (m *MongoDB) DeleteCommessa(id int) error {
	return m.db.Client().UseSession(m.ctx, func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			return err
		}

		if _, err := m.db.Collection("movimenti_primanota").DeleteMany(sessionContext, bson.M{"commessa_id": id}); err != nil {
			sessionContext.AbortTransaction(sessionContext)
			return err
		}

		if _, err := m.db.Collection("commesse").DeleteOne(sessionContext, bson.M{"id": id}); err != nil {
			sessionContext.AbortTransaction(sessionContext)
			return err
		}

		return sessionContext.CommitTransaction(sessionContext)
	})
}

func (m *MongoDB) ListCommesse(filters map[string]interface{}) ([]Commessa, error) {
	var list []Commessa
	query := bson.M{}

	if stato, ok := filters["stato"].(string); ok && stato != "" {
		query["stato"] = stato
	}

	opts := options.Find().SetSort(bson.D{{Key: "data_apertura", Value: -1}})
	cursor, err := m.db.Collection("commesse").Find(m.ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(m.ctx)
	return list, cursor.All(m.ctx, &list)
}

func (m *MongoDB) AggregateCommesseStats() (aperte int, chiuse int, err error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.M{"stato": "$stato"}},
			{Key: "count", Value: bson.M{"$sum": 1}},
		}}},
	}

	cursor, err := m.db.Collection("commesse").Aggregate(m.ctx, pipeline)
	if err != nil {
		return 0, 0, err
	}
	defer cursor.Close(m.ctx)

	stats := make(map[string]int)
	for cursor.Next(m.ctx) {
		var result struct {
			ID    map[string]string `bson:"_id"`
			Count int               `bson:"count"`
		}

		if err := cursor.Decode(&result); err != nil {
			return 0, 0, err
		}

		if stato, ok := result.ID["stato"]; ok {
			stats[stato] = result.Count
		}
	}

	return stats[StatoCommessaAperta], stats[StatoCommessaChiusa], nil
}

// ==================== APPUNTAMENTI ====================

func (m *MongoDB) CreateAppuntamento(a *Appuntamento) error {
	a.ID = generaID()
	_, err := m.db.Collection("appuntamenti").InsertOne(m.ctx, a)
	return err
}

func (m *MongoDB) GetAppuntamento(id int) (*Appuntamento, error) {
	var a Appuntamento
	err := m.db.Collection("appuntamenti").FindOne(m.ctx, bson.M{"id": id}).Decode(&a)
	if err != nil {
		return nil, fmt.Errorf("appuntamento non trovato: %w", err)
	}
	return &a, nil
}

func (m *MongoDB) UpdateAppuntamento(a *Appuntamento) error {
	result := m.db.Collection("appuntamenti").FindOneAndReplace(m.ctx, bson.M{"id": a.ID}, a)
	if result.Err() != nil {
		return fmt.Errorf("appuntamento #%d non trovato", a.ID)
	}
	return nil
}

func (m *MongoDB) DeleteAppuntamento(id int) error {
	_, err := m.db.Collection("appuntamenti").DeleteOne(m.ctx, bson.M{"id": id})
	return err
}

func (m *MongoDB) ListAppuntamenti(filters map[string]interface{}) ([]Appuntamento, error) {
	var list []Appuntamento
	query := bson.M{}

	if data, ok := filters["data"].(time.Time); ok {
		start := time.Date(data.Year(), data.Month(), data.Day(), 0, 0, 0, 0, time.Local)
		end := start.AddDate(0, 0, 1)
		query["data_ora"] = bson.M{"$gte": start, "$lt": end}
	}

	if veicoloID, ok := filters["veicolo_id"].(int); ok && veicoloID > 0 {
		query["veicolo_id"] = veicoloID
	}

	opts := options.Find().SetSort(bson.D{{Key: "data_ora", Value: 1}})
	cursor, err := m.db.Collection("appuntamenti").Find(m.ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(m.ctx)
	return list, cursor.All(m.ctx, &list)
}

// ==================== OPERATORI ====================

func (m *MongoDB) CreateOperatore(o *Operatore) error {
	o.ID = generaID()
	_, err := m.db.Collection("operatori").InsertOne(m.ctx, o)
	return err
}

func (m *MongoDB) GetOperatore(id int) (*Operatore, error) {
	var o Operatore
	err := m.db.Collection("operatori").FindOne(m.ctx, bson.M{"id": id}).Decode(&o)
	if err != nil {
		return nil, fmt.Errorf("operatore non trovato: %w", err)
	}
	return &o, nil
}

func (m *MongoDB) UpdateOperatore(o *Operatore) error {
	result := m.db.Collection("operatori").FindOneAndReplace(m.ctx, bson.M{"id": o.ID}, o)
	if result.Err() != nil {
		return fmt.Errorf("operatore #%d non trovato", o.ID)
	}
	return nil
}

func (m *MongoDB) DeleteOperatore(id int) error {
	_, err := m.db.Collection("operatori").DeleteOne(m.ctx, bson.M{"id": id})
	return err
}

func (m *MongoDB) ListOperatori() ([]Operatore, error) {
	var list []Operatore
	cursor, err := m.db.Collection("operatori").Find(m.ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "cognome", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(m.ctx)
	return list, cursor.All(m.ctx, &list)
}

// ==================== PREVENTIVI ====================

func (m *MongoDB) CreatePreventivo(p *Preventivo) error {
	p.ID = generaID()
	_, err := m.db.Collection("preventivi").InsertOne(m.ctx, p)
	return err
}

func (m *MongoDB) GetPreventivo(id int) (*Preventivo, error) {
	var p Preventivo
	err := m.db.Collection("preventivi").FindOne(m.ctx, bson.M{"id": id}).Decode(&p)
	if err != nil {
		return nil, fmt.Errorf("preventivo non trovato: %w", err)
	}
	return &p, nil
}

func (m *MongoDB) UpdatePreventivo(p *Preventivo) error {
	result := m.db.Collection("preventivi").FindOneAndReplace(m.ctx, bson.M{"id": p.ID}, p)
	if result.Err() != nil {
		return fmt.Errorf("preventivo #%d non trovato", p.ID)
	}
	return nil
}

func (m *MongoDB) DeletePreventivo(id int) error {
	_, err := m.db.Collection("preventivi").DeleteOne(m.ctx, bson.M{"id": id})
	return err
}

func (m *MongoDB) ListPreventivi() ([]Preventivo, error) {
	var list []Preventivo
	cursor, err := m.db.Collection("preventivi").Find(m.ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "data", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(m.ctx)
	return list, cursor.All(m.ctx, &list)
}

// ==================== FATTURE ====================

func (m *MongoDB) CreateFattura(f *Fattura) error {
	f.ID = generaID()
	_, err := m.db.Collection("fatture").InsertOne(m.ctx, f)
	return err
}

func (m *MongoDB) GetFattura(id int) (*Fattura, error) {
	var f Fattura
	err := m.db.Collection("fatture").FindOne(m.ctx, bson.M{"id": id}).Decode(&f)
	if err != nil {
		return nil, fmt.Errorf("fattura non trovata: %w", err)
	}
	return &f, nil
}

func (m *MongoDB) UpdateFattura(f *Fattura) error {
	result := m.db.Collection("fatture").FindOneAndReplace(m.ctx, bson.M{"id": f.ID}, f)
	if result.Err() != nil {
		return fmt.Errorf("fattura #%d non trovata", f.ID)
	}
	return nil
}

func (m *MongoDB) DeleteFattura(id int) error {
	_, err := m.db.Collection("fatture").DeleteOne(m.ctx, bson.M{"id": id})
	return err
}

func (m *MongoDB) ListFatture() ([]Fattura, error) {
	var list []Fattura
	cursor, err := m.db.Collection("fatture").Find(m.ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "data", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(m.ctx)
	return list, cursor.All(m.ctx, &list)
}

// ==================== MOVIMENTI PRIMA NOTA ====================

func (m *MongoDB) CreateMovimentoPrimaNota(mov *MovimentoPrimaNota) error {
	mov.ID = generaID()
	_, err := m.db.Collection("movimenti_primanota").InsertOne(m.ctx, mov)
	return err
}

func (m *MongoDB) GetMovimentoPrimaNota(id int) (*MovimentoPrimaNota, error) {
	var mov MovimentoPrimaNota
	err := m.db.Collection("movimenti_primanota").FindOne(m.ctx, bson.M{"id": id}).Decode(&mov)
	if err != nil {
		return nil, fmt.Errorf("movimento non trovato: %w", err)
	}
	return &mov, nil
}

func (m *MongoDB) UpdateMovimentoPrimaNota(mov *MovimentoPrimaNota) error {
	result := m.db.Collection("movimenti_primanota").FindOneAndReplace(m.ctx, bson.M{"id": mov.ID}, mov)
	if result.Err() != nil {
		return fmt.Errorf("movimento #%d non trovato", mov.ID)
	}
	return nil
}

func (m *MongoDB) DeleteMovimentoPrimaNota(id int) error {
	_, err := m.db.Collection("movimenti_primanota").DeleteOne(m.ctx, bson.M{"id": id})
	return err
}

func (m *MongoDB) ListMovimentiPrimaNota(filters map[string]interface{}) ([]MovimentoPrimaNota, error) {
	var list []MovimentoPrimaNota
	query := bson.M{}

	if tipo, ok := filters["tipo"].(string); ok && tipo != "" {
		query["tipo"] = tipo
	}

	if commessaID, ok := filters["commessa_id"].(int); ok && commessaID > 0 {
		query["commessa_id"] = commessaID
	}

	if data, ok := filters["data"].(time.Time); ok && !data.IsZero() {
		query["data"] = bson.M{"$gte": data}
	}

	opts := options.Find().SetSort(bson.D{{Key: "data", Value: -1}})
	cursor, err := m.db.Collection("movimenti_primanota").Find(m.ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(m.ctx)
	return list, cursor.All(m.ctx, &list)
}

// ==================== AGGREGATE QUERIES ====================

func (m *MongoDB) GetVeicoliByCliente(clienteID int) ([]Veicolo, error) {
	var list []Veicolo
	cursor, err := m.db.Collection("veicoli").Find(m.ctx, bson.M{"cliente_id": clienteID}, options.Find().SetSort(bson.D{{Key: "marca", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(m.ctx)
	return list, cursor.All(m.ctx, &list)
}

func (m *MongoDB) GetCommesseStats() (aperte int, chiuse int, err error) {
	return m.AggregateCommesseStats()
}

func (m *MongoDB) GetPrimaNotaStats(anno int) (entrata float64, uscita float64, err error) {
	start := time.Date(anno, 1, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(1, 0, 0)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"data": bson.M{"$gte": start, "$lt": end}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$tipo"},
			{Key: "total", Value: bson.M{"$sum": "$importo"}},
		}}},
	}

	cursor, err := m.db.Collection("movimenti_primanota").Aggregate(m.ctx, pipeline)
	if err != nil {
		return 0, 0, err
	}
	defer cursor.Close(m.ctx)

	stats := make(map[string]float64)
	for cursor.Next(m.ctx) {
		var result struct {
			ID struct {
				Tipo string `bson:"tipo"`
			} `bson:"_id"`
			Total float64 `bson:"total"`
		}

		if err := cursor.Decode(&result); err != nil {
			return 0, 0, err
		}

		stats[result.ID.Tipo] = result.Total
	}

	return stats[TipoMovimentoEntrata], stats[TipoMovimentoUscita], nil
}
