package db

import (
	store "nadhi.dev/sarvar/fun/database"
)

var SessionsDB *store.DB
var UsersDB *store.DB
var QueueDB *store.DB
var NotebooksDB *store.DB
var StylesDB *store.DB

func InitSessionsDB() error {
	var err error
	SessionsDB, err = store.InitDB("sessions")
	return err
}

func InitUsersDB() error {
	var err error
	UsersDB, err = store.InitDB("users")
	return err
}

func InitQueueDB() error {
	var err error
	QueueDB, err = store.InitDB("queue")
	return err
}

func InitNotebooksDB() error {
	var err error
	NotebooksDB, err = store.InitDB("notebooks")
	return err
}

func InitStylesDB() error {
	var err error
	StylesDB, err = store.InitDB("styles")
	return err
}
