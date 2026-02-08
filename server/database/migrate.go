package store

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// MigrateJSONToBadger migrates data from JSON files to BadgerDB
func MigrateJSONToBadger(jsonDir string, badgerDB *BadgerDB) error {
	log.Println("[MIGRATION] Starting JSON to BadgerDB migration...")

	// Migrate users
	if err := migrateUsers(jsonDir, badgerDB); err != nil {
		return fmt.Errorf("failed to migrate users: %w", err)
	}

	// Migrate sessions
	if err := migrateSessions(jsonDir, badgerDB); err != nil {
		return fmt.Errorf("failed to migrate sessions: %w", err)
	}

	// Migrate notebooks
	if err := migrateNotebooks(jsonDir, badgerDB); err != nil {
		return fmt.Errorf("failed to migrate notebooks: %w", err)
	}

	// Migrate queue
	if err := migrateQueue(jsonDir, badgerDB); err != nil {
		return fmt.Errorf("failed to migrate queue: %w", err)
	}

	// Migrate styles
	if err := migrateStyles(jsonDir, badgerDB); err != nil {
		return fmt.Errorf("failed to migrate styles: %w", err)
	}

	log.Println("[MIGRATION] Migration completed successfully!")
	return nil
}

func migrateUsers(jsonDir string, badgerDB *BadgerDB) error {
	usersFile := filepath.Join(jsonDir, "users", "users.json")
	if _, err := os.Stat(usersFile); os.IsNotExist(err) {
		log.Println("[MIGRATION] No users.json found, skipping users migration")
		return nil
	}

	db, err := InitDB(jsonDir)
	if err != nil {
		return err
	}

	users, err := GetAllUsers(db)
	if err != nil {
		return err
	}

	count := 0
	for _, user := range users {
		if err := AddUserBadger(badgerDB, user); err != nil {
			return err
		}
		count++
	}

	log.Printf("[MIGRATION] Migrated %d users", count)
	return nil
}

func migrateSessions(jsonDir string, badgerDB *BadgerDB) error {
	sessionsFile := filepath.Join(jsonDir, "sessions", "sessions.json")
	if _, err := os.Stat(sessionsFile); os.IsNotExist(err) {
		log.Println("[MIGRATION] No sessions.json found, skipping sessions migration")
		return nil
	}

	db, err := InitDB(jsonDir)
	if err != nil {
		return err
	}

	store, err := db.GetStore("sessions")
	if err != nil {
		return err
	}

	var sessions map[string]Session
	if err := store.GetData(&sessions); err != nil {
		return err
	}

	count := 0
	for _, session := range sessions {
		if err := AddSessionBadger(badgerDB, session); err != nil {
			return err
		}
		count++
	}

	log.Printf("[MIGRATION] Migrated %d sessions", count)
	return nil
}

func migrateNotebooks(jsonDir string, badgerDB *BadgerDB) error {
	notebooksFile := filepath.Join(jsonDir, "notebooks", "notebooks.json")
	if _, err := os.Stat(notebooksFile); os.IsNotExist(err) {
		log.Println("[MIGRATION] No notebooks.json found, skipping notebooks migration")
		return nil
	}

	db, err := InitDB(jsonDir)
	if err != nil {
		return err
	}

	store, err := db.GetStore("notebooks")
	if err != nil {
		return err
	}

	var notebooks map[string]map[string]Notebook
	if err := store.GetData(&notebooks); err != nil {
		return err
	}

	count := 0
	for username, userNotebooks := range notebooks {
		for _, notebook := range userNotebooks {
			key := fmt.Sprintf("notebooks:%s:%d", username, notebook.ID)
			if err := badgerDB.Set(key, notebook); err != nil {
				return err
			}
			count++
		}
	}

	log.Printf("[MIGRATION] Migrated %d notebooks", count)
	return nil
}

func migrateQueue(jsonDir string, badgerDB *BadgerDB) error {
	queueFile := filepath.Join(jsonDir, "queue", "queue.json")
	if _, err := os.Stat(queueFile); os.IsNotExist(err) {
		log.Println("[MIGRATION] No queue.json found, skipping queue migration")
		return nil
	}

	db, err := InitDB(jsonDir)
	if err != nil {
		return err
	}

	jobs, err := GetAllQueuedJobs(db, "")
	if err != nil {
		return err
	}

	count := 0
	for _, job := range jobs {
		if err := AddQueuedJobBadger(badgerDB, job); err != nil {
			return err
		}
		count++
	}

	log.Printf("[MIGRATION] Migrated %d queue jobs", count)
	return nil
}

func migrateStyles(jsonDir string, badgerDB *BadgerDB) error {
	stylesFile := filepath.Join(jsonDir, "styles", "styles.json")
	if _, err := os.Stat(stylesFile); os.IsNotExist(err) {
		log.Println("[MIGRATION] No styles.json found, skipping styles migration")
		return nil
	}

	db, err := InitDB(jsonDir)
	if err != nil {
		return err
	}

	store, err := db.GetStore("styles")
	if err != nil {
		return err
	}

	var styles map[string]map[string]Style
	if err := store.GetData(&styles); err != nil {
		return err
	}

	count := 0
	for username, userStyles := range styles {
		for _, style := range userStyles {
			key := fmt.Sprintf("styles:%s:%s", username, style.Name)
			if err := badgerDB.Set(key, style); err != nil {
				return err
			}
			count++
		}
	}

	log.Printf("[MIGRATION] Migrated %d styles", count)
	return nil
}
