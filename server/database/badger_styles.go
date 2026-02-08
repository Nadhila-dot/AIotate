package store

import (
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

// AddStyleBadger adds a style to BadgerDB
func AddStyleBadger(bdb *BadgerDB, style Style) error {
	key := fmt.Sprintf("styles:%s:%s", style.Username, style.Name)
	return bdb.Set(key, style)
}

// GetStyleBadger retrieves a style from BadgerDB
func GetStyleBadger(bdb *BadgerDB, username, name string) (*Style, error) {
	key := fmt.Sprintf("styles:%s:%s", username, name)
	var style Style
	err := bdb.Get(key, &style)
	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &style, nil
}

// GetAllStylesBadger retrieves all styles for a user from BadgerDB
func GetAllStylesBadger(bdb *BadgerDB, username string) ([]Style, error) {
	var styles []Style
	prefix := fmt.Sprintf("styles:%s:", username)

	err := bdb.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(prefix)
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			err := item.Value(func(val []byte) error {
				var style Style
				if err := jsonUnmarshal(val, &style); err != nil {
					return err
				}
				styles = append(styles, style)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return styles, err
}

// DeleteStyleBadger removes a style from BadgerDB
func DeleteStyleBadger(bdb *BadgerDB, username, name string) error {
	key := fmt.Sprintf("styles:%s:%s", username, name)
	return bdb.Delete(key)
}

// UpdateStyleBadger updates a style in BadgerDB
func UpdateStyleBadger(bdb *BadgerDB, style Style) error {
	key := fmt.Sprintf("styles:%s:%s", style.Username, style.Name)
	return bdb.Set(key, style)
}
