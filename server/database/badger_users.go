package store

import (
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

// AddUserBadger adds a user to BadgerDB
func AddUserBadger(bdb *BadgerDB, user User) error {
	key := fmt.Sprintf("users:%s", user.Username)
	return bdb.Set(key, user)
}

// GetUserBadger retrieves a user from BadgerDB
func GetUserBadger(bdb *BadgerDB, username string) (*User, error) {
	key := fmt.Sprintf("users:%s", username)
	var user User
	err := bdb.Get(key, &user)
	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAllUsersBadger retrieves all users from BadgerDB
func GetAllUsersBadger(bdb *BadgerDB) (map[string]User, error) {
	users := make(map[string]User)

	err := bdb.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("users:")
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			err := item.Value(func(val []byte) error {
				var user User
				if err := jsonUnmarshal(val, &user); err != nil {
					return err
				}
				users[user.Username] = user
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return users, err
}

// RemoveUserBadger removes a user from BadgerDB
func RemoveUserBadger(bdb *BadgerDB, username string) error {
	key := fmt.Sprintf("users:%s", username)
	return bdb.Delete(key)
}
