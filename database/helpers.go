package database

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type Serializable interface {
	ToJSON() ([]byte, error)
}

type Validator interface {
	Validate() error
}

func (db *DB) create(bucketName []byte, obj Serializable, getID func() *int, setID func(int)) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		seq, err := b.NextSequence()
		if err != nil {
			return fmt.Errorf("errore generazione sequence: %w", err)
		}

		id := int(seq)
		setID(id)

		if validator, ok := obj.(Validator); ok {
			if err := validator.Validate(); err != nil {
				return fmt.Errorf("validazione fallita: %w", err)
			}
		}

		data, err := obj.ToJSON()
		if err != nil {
			return fmt.Errorf("errore serializzazione: %w", err)
		}

		return b.Put(itob(id), data)
	})
}

func (db *DB) get(bucketName []byte, id int, deserialize func([]byte) (interface{}, error)) (interface{}, error) {
	var result interface{}
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		data := b.Get(itob(id))
		if data == nil {
			return fmt.Errorf("record non trovato")
		}

		var err error
		result, err = deserialize(data)
		return err
	})
	return result, err
}

func (db *DB) update(bucketName []byte, id int, obj Serializable) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b.Get(itob(id)) == nil {
			return fmt.Errorf("record %d non trovato", id)
		}

		if validator, ok := obj.(Validator); ok {
			if err := validator.Validate(); err != nil {
				return fmt.Errorf("validazione fallita: %w", err)
			}
		}

		data, err := obj.ToJSON()
		if err != nil {
			return fmt.Errorf("errore serializzazione: %w", err)
		}

		return b.Put(itob(id), data)
	})
}

func (db *DB) delete(bucketName []byte, id int) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b.Get(itob(id)) == nil {
			return fmt.Errorf("record %d non trovato", id)
		}
		return b.Delete(itob(id))
	})
}

func (db *DB) exists(bucketName []byte, id int) (bool, error) {
	var exists bool
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		exists = b.Get(itob(id)) != nil
		return nil
	})
	return exists, err
}

func (db *DB) count(bucketName []byte) (int, error) {
	var count int
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		count = b.Stats().KeyN
		return nil
	})
	return count, err
}

type ListFilter func([]byte) (interface{}, bool, error)

func (db *DB) listWithFilter(bucketName []byte, filter ListFilter) ([]interface{}, error) {
	var list []interface{}
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.ForEach(func(k, v []byte) error {
			obj, include, err := filter(v)
			if err != nil {
				return err
			}
			if include {
				list = append(list, obj)
			}
			return nil
		})
	})
	return list, err
}

func (db *DB) GetVeicoliByCliente(clienteID int) ([]Veicolo, error) {
	var list []Veicolo
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktVeicoli)
		return b.ForEach(func(k, v []byte) error {
			veic, err := FromJSONVeicolo(v)
			if err == nil && veic.ClienteID == clienteID {
				list = append(list, *veic)
			}
			return nil
		})
	})
	return list, err
}

func (db *DB) GetCommesseByVeicolo(veicoloID int) ([]Commessa, error) {
	var list []Commessa
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktCommesse)
		return b.ForEach(func(k, v []byte) error {
			comm, err := FromJSONCommessa(v)
			if err == nil && comm.VeicoloID == veicoloID {
				list = append(list, *comm)
			}
			return nil
		})
	})
	return list, err
}

func (db *DB) GetMovimentiByCommessa(commessaID int) ([]MovimentoPrimaNota, error) {
	var list []MovimentoPrimaNota
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktPrimaNota)
		return b.ForEach(func(k, v []byte) error {
			mov, err := FromJSONMovimentoPrimaNota(v)
			if err == nil && mov.CommessaID == commessaID {
				list = append(list, *mov)
			}
			return nil
		})
	})
	return list, err
}

func (db *DB) CountClienti() (int, error) {
	return db.count(BktClienti)
}

func (db *DB) CountVeicoli() (int, error) {
	return db.count(BktVeicoli)
}

func (db *DB) CountCommesse() (int, error) {
	return db.count(BktCommesse)
}

func (db *DB) CountCommesseAperte() (int, error) {
	var count int
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BktCommesse)
		return b.ForEach(func(k, v []byte) error {
			comm, err := FromJSONCommessa(v)
			if err == nil && comm.IsOpen() {
				count++
			}
			return nil
		})
	})
	return count, err
}
