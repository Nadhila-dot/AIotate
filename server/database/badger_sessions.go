package store

import (
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

// AddSessionBadger adds a session to BadgerDB
func AddSessionBadger(bdb *BadgerDB, session Session) error {
	key := fmt.Sprintf("sessions:%s", session.ID)
	return bdb.Set(key, session)
}

// GetSessionBadger retrieves a session from BadgerDB
func GetSessionBadger(bdb *BadgerDB, id string) (*Session, error) {
	key := fmt.Sprintf("sessions:%s", id)
	var session Session
	err := bdb.Get(key, &session)
	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// RemoveSessionBadger removes a session from BadgerDB
func RemoveSessionBadger(bdb *BadgerDB, id string) error {
	key := fmt.Sprintf("sessions:%s", id)
	return bdb.Delete(key)
}
