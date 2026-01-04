package database

import (
	"encoding/binary"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

// DB wrappa il database BoltDB
type DB struct {
	*bolt.DB
}

// Bucket names
var (
	BktClienti      = []byte("clienti")
	BktVeicoli      = []byte("veicoli")
	BktCommesse     = []byte("commesse")
	BktAppuntamenti = []byte("appuntamenti")
	BktOperatori    = []byte("operatori")
	BktPreventivi   = []byte("preventivi")
	BktFatture      = []byte("fatture")
	BktPrimaNota    = []byte("primanota")
)

// itob converte un int in []byte per chiavi BoltDB
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// InitDB inizializza il database e crea i bucket necessari
func InitDB(path string) (*DB, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("errore apertura database: %w", err)
	}

	// Creazione bucket
	err = db.Update(func(tx *bolt.Tx) error {
		buckets := [][]byte{
			BktClienti,
			BktVeicoli,
			BktCommesse,
			BktAppuntamenti,
			BktOperatori,
			BktPreventivi,
			BktFatture,
			BktPrimaNota,
		}

		for _, bucketName := range buckets {
			_, err := tx.CreateBucketIfNotExists(bucketName)
			if err != nil {
				return fmt.Errorf("errore creazione bucket %s: %w", string(bucketName), err)
			}
		}

		return nil
	})

	if err != nil {
		db.Close()
		return nil, err
	}

	return &DB{db}, nil
}

// ============================================
// CLIENTI - CRUD Completo
// ============================================

func (db *DB) CreateCliente(c *Cliente) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktClienti)
		seq, err := b.NextSequence()
		if err != nil {
			return err
		}
		c.ID = int(seq)
		data, err := c.ToJSON()
		if err != nil {
			return err
		}
		return b.Put(itob(c.ID), data)
	})
}

func (db *DB) GetCliente(id int) (*Cliente, error) {
	var c *Cliente
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktClienti)
		data := b.Get(itob(id))
		if data == nil {
			return fmt.Errorf("cliente non trovato")
		}
		var err error
		c, err = FromJSONCliente(data)
		return err
	})
	return c, err
}

func (db *DB) UpdateCliente(c *Cliente) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktClienti)
		if b.Get(itob(c.ID)) == nil {
			return fmt.Errorf("cliente %d non trovato", c.ID)
		}
		data, err := c.ToJSON()
		if err != nil {
			return err
		}
		return b.Put(itob(c.ID), data)
	})
}

func (db *DB) DeleteCliente(id int) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktClienti)
		if b.Get(itob(id)) == nil {
			return fmt.Errorf("cliente %d non trovato", id)
		}
		return b.Delete(itob(id))
	})
}

func (db *DB) ListClienti() ([]Cliente, error) {
	var list []Cliente
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktClienti)
		return b.ForEach(func(k, v []byte) error {
			obj, err := FromJSONCliente(v)
			if err == nil {
				list = append(list, *obj)
			}
			return nil
		})
	})
	return list, err
}

// ============================================
// VEICOLI - CRUD Completo
// ============================================

func (db *DB) CreateVeicolo(v *Veicolo) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktVeicoli)
		seq, err := b.NextSequence()
		if err != nil {
			return err
		}
		v.ID = int(seq)
		data, err := v.ToJSON()
		if err != nil {
			return err
		}
		return b.Put(itob(v.ID), data)
	})
}

func (db *DB) GetVeicolo(id int) (*Veicolo, error) {
	var v *Veicolo
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktVeicoli)
		data := b.Get(itob(id))
		if data == nil {
			return fmt.Errorf("veicolo non trovato")
		}
		var err error
		v, err = FromJSONVeicolo(data)
		return err
	})
	return v, err
}

func (db *DB) UpdateVeicolo(v *Veicolo) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktVeicoli)
		if b.Get(itob(v.ID)) == nil {
			return fmt.Errorf("veicolo %d non trovato", v.ID)
		}
		data, err := v.ToJSON()
		if err != nil {
			return err
		}
		return b.Put(itob(v.ID), data)
	})
}

func (db *DB) DeleteVeicolo(id int) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktVeicoli)
		if b.Get(itob(id)) == nil {
			return fmt.Errorf("veicolo %d non trovato", id)
		}
		return b.Delete(itob(id))
	})
}

func (db *DB) ListVeicoli() ([]Veicolo, error) {
	var list []Veicolo
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktVeicoli)
		return b.ForEach(func(k, v []byte) error {
			obj, err := FromJSONVeicolo(v)
			if err == nil {
				list = append(list, *obj)
			}
			return nil
		})
	})
	return list, err
}

// ============================================
// COMMESSE - CRUD Completo
// ============================================

func (db *DB) CreateCommessa(c *Commessa) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktCommesse)
		seq, err := b.NextSequence()
		if err != nil {
			return err
		}
		c.ID = int(seq)
		c.Numero = fmt.Sprintf("COM-%04d", seq)
		c.DataApertura = time.Now()
		c.Totale = c.CostoManodopera + c.CostoRicambi

		data, err := c.ToJSON()
		if err != nil {
			return err
		}
		return b.Put(itob(c.ID), data)
	})
}

func (db *DB) GetCommessa(id int) (*Commessa, error) {
	var c *Commessa
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktCommesse)
		data := b.Get(itob(id))
		if data == nil {
			return fmt.Errorf("commessa non trovata")
		}
		var err error
		c, err = FromJSONCommessa(data)
		return err
	})
	return c, err
}

func (db *DB) UpdateCommessa(c *Commessa) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktCommesse)
		if b.Get(itob(c.ID)) == nil {
			return fmt.Errorf("commessa %d non trovata", c.ID)
		}

		// Ricalcola totale
		c.Totale = c.CostoManodopera + c.CostoRicambi

		// Imposta data chiusura se stato diventa "Chiusa"
		if c.Stato == "Chiusa" && c.DataChiusura.IsZero() {
			c.DataChiusura = time.Now()
		}

		data, err := c.ToJSON()
		if err != nil {
			return err
		}
		return b.Put(itob(c.ID), data)
	})
}

func (db *DB) DeleteCommessa(id int) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktCommesse)
		if b.Get(itob(id)) == nil {
			return fmt.Errorf("commessa %d non trovata", id)
		}
		return b.Delete(itob(id))
	})
}

func (db *DB) ListCommesse() ([]Commessa, error) {
	var list []Commessa
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktCommesse)
		return b.ForEach(func(k, v []byte) error {
			obj, err := FromJSONCommessa(v)
			if err == nil {
				list = append(list, *obj)
			}
			return nil
		})
	})
	return list, err
}

// ============================================
// APPUNTAMENTI - CRUD Completo
// ============================================

func (db *DB) CreateAppuntamento(a *Appuntamento) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktAppuntamenti)
		seq, err := b.NextSequence()
		if err != nil {
			return err
		}
		a.ID = int(seq)
		data, err := a.ToJSON()
		if err != nil {
			return err
		}
		return b.Put(itob(a.ID), data)
	})
}

func (db *DB) GetAppuntamento(id int) (*Appuntamento, error) {
	var a *Appuntamento
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktAppuntamenti)
		data := b.Get(itob(id))
		if data == nil {
			return fmt.Errorf("appuntamento non trovato")
		}
		var err error
		a, err = FromJSONAppuntamento(data)
		return err
	})
	return a, err
}

func (db *DB) UpdateAppuntamento(a *Appuntamento) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktAppuntamenti)
		if b.Get(itob(a.ID)) == nil {
			return fmt.Errorf("appuntamento %d non trovato", a.ID)
		}
		data, err := a.ToJSON()
		if err != nil {
			return err
		}
		return b.Put(itob(a.ID), data)
	})
}

func (db *DB) DeleteAppuntamento(id int) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktAppuntamenti)
		if b.Get(itob(id)) == nil {
			return fmt.Errorf("appuntamento %d non trovato", id)
		}
		return b.Delete(itob(id))
	})
}

func (db *DB) ListAppuntamenti() ([]Appuntamento, error) {
	var list []Appuntamento
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktAppuntamenti)
		return b.ForEach(func(k, v []byte) error {
			obj, err := FromJSONAppuntamento(v)
			if err == nil {
				list = append(list, *obj)
			}
			return nil
		})
	})
	return list, err
}

// ============================================
// OPERATORI - CRUD Completo
// ============================================

func (db *DB) CreateOperatore(o *Operatore) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktOperatori)
		seq, err := b.NextSequence()
		if err != nil {
			return err
		}
		o.ID = int(seq)
		data, err := o.ToJSON()
		if err != nil {
			return err
		}
		return b.Put(itob(o.ID), data)
	})
}

func (db *DB) GetOperatore(id int) (*Operatore, error) {
	var o *Operatore
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktOperatori)
		data := b.Get(itob(id))
		if data == nil {
			return fmt.Errorf("operatore non trovato")
		}
		var err error
		o, err = FromJSONOperatore(data)
		return err
	})
	return o, err
}

func (db *DB) UpdateOperatore(o *Operatore) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktOperatori)
		if b.Get(itob(o.ID)) == nil {
			return fmt.Errorf("operatore %d non trovato", o.ID)
		}
		data, err := o.ToJSON()
		if err != nil {
			return err
		}
		return b.Put(itob(o.ID), data)
	})
}

func (db *DB) DeleteOperatore(id int) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktOperatori)
		if b.Get(itob(id)) == nil {
			return fmt.Errorf("operatore %d non trovato", id)
		}
		return b.Delete(itob(id))
	})
}

func (db *DB) ListOperatori() ([]Operatore, error) {
	var list []Operatore
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktOperatori)
		return b.ForEach(func(k, v []byte) error {
			obj, err := FromJSONOperatore(v)
			if err == nil {
				list = append(list, *obj)
			}
			return nil
		})
	})
	return list, err
}

// ============================================
// PREVENTIVI - CRUD Completo
// ============================================

func (db *DB) CreatePreventivo(p *Preventivo) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktPreventivi)
		seq, err := b.NextSequence()
		if err != nil {
			return err
		}
		p.ID = int(seq)
		p.Numero = fmt.Sprintf("PREV-%04d", seq)
		p.Data = time.Now()
		p.Accettato = false

		data, err := p.ToJSON()
		if err != nil {
			return err
		}
		return b.Put(itob(p.ID), data)
	})
}

func (db *DB) GetPreventivo(id int) (*Preventivo, error) {
	var p *Preventivo
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktPreventivi)
		data := b.Get(itob(id))
		if data == nil {
			return fmt.Errorf("preventivo non trovato")
		}
		var err error
		p, err = FromJSONPreventivo(data)
		return err
	})
	return p, err
}

func (db *DB) UpdatePreventivo(p *Preventivo) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktPreventivi)
		if b.Get(itob(p.ID)) == nil {
			return fmt.Errorf("preventivo %d non trovato", p.ID)
		}
		data, err := p.ToJSON()
		if err != nil {
			return err
		}
		return b.Put(itob(p.ID), data)
	})
}

func (db *DB) DeletePreventivo(id int) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktPreventivi)
		if b.Get(itob(id)) == nil {
			return fmt.Errorf("preventivo %d non trovato", id)
		}
		return b.Delete(itob(id))
	})
}

func (db *DB) ListPreventivi() ([]Preventivo, error) {
	var list []Preventivo
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktPreventivi)
		return b.ForEach(func(k, v []byte) error {
			obj, err := FromJSONPreventivo(v)
			if err == nil {
				list = append(list, *obj)
			}
			return nil
		})
	})
	return list, err
}

// ============================================
// FATTURE - CRUD Completo
// ============================================

func (db *DB) CreateFattura(f *Fattura) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktFatture)
		seq, err := b.NextSequence()
		if err != nil {
			return err
		}
		f.ID = int(seq)
		f.Numero = fmt.Sprintf("FT-%04d/%d", seq, f.Data.Year())

		data, err := f.ToJSON()
		if err != nil {
			return err
		}
		return b.Put(itob(f.ID), data)
	})
}

func (db *DB) GetFattura(id int) (*Fattura, error) {
	var f *Fattura
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktFatture)
		data := b.Get(itob(id))
		if data == nil {
			return fmt.Errorf("fattura non trovata")
		}
		var err error
		f, err = FromJSONFattura(data)
		return err
	})
	return f, err
}

func (db *DB) UpdateFattura(f *Fattura) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktFatture)
		if b.Get(itob(f.ID)) == nil {
			return fmt.Errorf("fattura %d non trovata", f.ID)
		}
		data, err := f.ToJSON()
		if err != nil {
			return err
		}
		return b.Put(itob(f.ID), data)
	})
}

func (db *DB) DeleteFattura(id int) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktFatture)
		if b.Get(itob(id)) == nil {
			return fmt.Errorf("fattura %d non trovata", id)
		}
		return b.Delete(itob(id))
	})
}

func (db *DB) ListFatture() ([]Fattura, error) {
	var list []Fattura
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktFatture)
		return b.ForEach(func(k, v []byte) error {
			obj, err := FromJSONFattura(v)
			if err == nil {
				list = append(list, *obj)
			}
			return nil
		})
	})
	return list, err
}

// ============================================
// PRIMA NOTA - CRUD Completo
// ============================================

func (db *DB) CreateMovimento(m *MovimentoPrimaNota) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktPrimaNota)
		seq, err := b.NextSequence()
		if err != nil {
			return err
		}
		m.ID = int(seq)
		data, err := m.ToJSON()
		if err != nil {
			return err
		}
		return b.Put(itob(m.ID), data)
	})
}

func (db *DB) GetMovimento(id int) (*MovimentoPrimaNota, error) {
	var m *MovimentoPrimaNota
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktPrimaNota)
		data := b.Get(itob(id))
		if data == nil {
			return fmt.Errorf("movimento non trovato")
		}
		var err error
		m, err = FromJSONMovimentoPrimaNota(data)
		return err
	})
	return m, err
}

func (db *DB) UpdateMovimento(m *MovimentoPrimaNota) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktPrimaNota)
		if b.Get(itob(m.ID)) == nil {
			return fmt.Errorf("movimento %d non trovato", m.ID)
		}
		data, err := m.ToJSON()
		if err != nil {
			return err
		}
		return b.Put(itob(m.ID), data)
	})
}

func (db *DB) DeleteMovimento(id int) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktPrimaNota)
		if b.Get(itob(id)) == nil {
			return fmt.Errorf("movimento %d non trovato", id)
		}
		return b.Delete(itob(id))
	})
}

func (db *DB) ListMovimenti() ([]MovimentoPrimaNota, error) {
	var list []MovimentoPrimaNota
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktPrimaNota)
		return b.ForEach(func(k, v []byte) error {
			obj, err := FromJSONMovimentoPrimaNota(v)
			if err == nil {
				list = append(list, *obj)
			}
			return nil
		})
	})
	return list, err
}
