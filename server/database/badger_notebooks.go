package store

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// CreateNotebookBadger creates a new notebook in BadgerDB
func CreateNotebookBadger(bdb *BadgerDB, username, name, description string, optional Optional) (*Notebook, error) {
	// Generate unique ID
	id := 1 + rand.Intn(99999)

	// Check if ID exists and regenerate if needed
	for {
		key := fmt.Sprintf("notebooks:%s:%d", username, id)
		exists, err := bdb.Exists(key)
		if err != nil {
			return nil, err
		}
		if !exists {
			break
		}
		id = 1 + rand.Intn(99999)
	}

	now := time.Now()
	notebook := Notebook{
		ID:          id,
		Name:        name,
		Username:    username,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
		Optional:    optional,
		Items:       make(map[string]string),
	}

	key := fmt.Sprintf("notebooks:%s:%d", username, id)
	if err := bdb.Set(key, notebook); err != nil {
		return nil, err
	}

	return &notebook, nil
}

// GetNotebookBadger retrieves a notebook from BadgerDB
func GetNotebookBadger(bdb *BadgerDB, username string, id int) (*Notebook, error) {
	key := fmt.Sprintf("notebooks:%s:%d", username, id)
	var notebook Notebook
	err := bdb.Get(key, &notebook)
	if err == badger.ErrKeyNotFound {
		return nil, fmt.Errorf("notebook %d not found", id)
	}
	if err != nil {
		return nil, err
	}
	return &notebook, nil
}

// GetAllNotebooksBadger retrieves all notebooks for a user from BadgerDB
func GetAllNotebooksBadger(bdb *BadgerDB, username string) ([]Notebook, error) {
	var notebooks []Notebook
	prefix := fmt.Sprintf("notebooks:%s:", username)

	err := bdb.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(prefix)
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			err := item.Value(func(val []byte) error {
				var notebook Notebook
				if err := jsonUnmarshal(val, &notebook); err != nil {
					return err
				}
				notebooks = append(notebooks, notebook)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return notebooks, nil
}

// AddItemToNotebookBadger adds a sheet to a notebook in BadgerDB
func AddItemToNotebookBadger(bdb *BadgerDB, username string, id int, sheetName, url string) error {
	notebook, err := GetNotebookBadger(bdb, username, id)
	if err != nil {
		return err
	}

	notebook.Items[sheetName] = url
	notebook.UpdatedAt = time.Now()

	key := fmt.Sprintf("notebooks:%s:%d", username, id)
	return bdb.Set(key, notebook)
}

// DeleteNotebookBadger removes a notebook from BadgerDB
func DeleteNotebookBadger(bdb *BadgerDB, username string, id int) error {
	key := fmt.Sprintf("notebooks:%s:%d", username, id)
	return bdb.Delete(key)
}

// DeleteItemFromNotebookBadger removes a sheet from a notebook in BadgerDB
func DeleteItemFromNotebookBadger(bdb *BadgerDB, username string, id int, itemName string) error {
	notebook, err := GetNotebookBadger(bdb, username, id)
	if err != nil {
		return err
	}

	if _, exists := notebook.Items[itemName]; !exists {
		return fmt.Errorf("item %s not found in notebook", itemName)
	}

	delete(notebook.Items, itemName)
	notebook.UpdatedAt = time.Now()

	key := fmt.Sprintf("notebooks:%s:%d", username, id)
	return bdb.Set(key, notebook)
}

// GetItemsInNotebookBadger gets all sheets in a notebook from BadgerDB
func GetItemsInNotebookBadger(bdb *BadgerDB, username string, id int) (map[string]string, error) {
	notebook, err := GetNotebookBadger(bdb, username, id)
	if err != nil {
		return nil, err
	}
	return notebook.Items, nil
}

// UpdateNotebookBadger updates a notebook in BadgerDB
func UpdateNotebookBadger(bdb *BadgerDB, username string, notebook Notebook) error {
	// Get original to preserve CreatedAt
	original, err := GetNotebookBadger(bdb, username, notebook.ID)
	if err != nil {
		return err
	}

	notebook.CreatedAt = original.CreatedAt
	notebook.UpdatedAt = time.Now()

	key := fmt.Sprintf("notebooks:%s:%d", username, notebook.ID)
	return bdb.Set(key, notebook)
}
