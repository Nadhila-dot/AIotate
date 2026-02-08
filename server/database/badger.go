package store

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// BadgerDB wraps the badger database
type BadgerDB struct {
	db *badger.DB
}

// InitBadgerDB initializes a new BadgerDB instance
func InitBadgerDB(path string) (*BadgerDB, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil // Disable badger's verbose logging

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db: %w", err)
	}

	log.Printf("[DB] BadgerDB initialized at %s", path)
	return &BadgerDB{db: db}, nil
}

// Close closes the database connection
func (bdb *BadgerDB) Close() error {
	return bdb.db.Close()
}

// Set stores a key-value pair
func (bdb *BadgerDB) Set(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return bdb.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

// Get retrieves a value by key
func (bdb *BadgerDB) Get(key string, out interface{}) error {
	return bdb.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, out)
		})
	})
}

// Delete removes a key
func (bdb *BadgerDB) Delete(key string) error {
	return bdb.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// GetAll retrieves all keys with a given prefix
func (bdb *BadgerDB) GetAll(prefix string, out interface{}) error {
	results := make(map[string]json.RawMessage)

	err := bdb.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(prefix)
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())

			err := item.Value(func(val []byte) error {
				results[key] = json.RawMessage(val)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Marshal and unmarshal to convert to desired type
	data, err := json.Marshal(results)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, out)
}

// Exists checks if a key exists
func (bdb *BadgerDB) Exists(key string) (bool, error) {
	err := bdb.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		return err
	})

	if err == badger.ErrKeyNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// RunGC runs garbage collection on the database
func (bdb *BadgerDB) RunGC() error {
	return bdb.db.RunValueLogGC(0.5)
}

// Backup creates a backup of the database
func (bdb *BadgerDB) Backup(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = bdb.db.Backup(f, 0)
	return err
}

// ExportToJSON exports all data to JSON files for debugging
func (bdb *BadgerDB) ExportToJSON(outputDir string) error {
	collections := []string{"users", "sessions", "notebooks", "queue", "styles"}

	for _, collection := range collections {
		var data map[string]interface{}
		err := bdb.GetAll(collection+":", &data)
		if err != nil && err != badger.ErrKeyNotFound {
			return fmt.Errorf("failed to export %s: %w", collection, err)
		}

		// Write to file (using the old JSON structure for debugging)
		store := &Store{
			Name: collection,
			Path: fmt.Sprintf("%s/%s.json", outputDir, collection),
		}
		if err := store.SetData(data); err != nil {
			log.Printf("[DB] Warning: failed to export %s to JSON: %v", collection, err)
		}
	}

	log.Printf("[DB] Exported database to JSON at %s", outputDir)
	return nil
}

// StartGCRoutine starts a background goroutine for periodic garbage collection
func (bdb *BadgerDB) StartGCRoutine(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			err := bdb.RunGC()
			if err != nil && err != badger.ErrNoRewrite {
				log.Printf("[DB] GC error: %v", err)
			}
		}
	}()
}
