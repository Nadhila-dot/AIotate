package store

import (
	"log"
	"os"
	"time"
)

// GlobalDB is the global database instance
var GlobalDB *UnifiedDB

// UnifiedDB provides a unified interface for both BadgerDB and JSON export
type UnifiedDB struct {
	Badger    *BadgerDB
	JSONDir   string
	DebugMode bool
}

// InitUnifiedDB initializes the unified database system
func InitUnifiedDB(badgerPath, jsonDir string, debugMode bool) (*UnifiedDB, error) {
	// Initialize BadgerDB
	badger, err := InitBadgerDB(badgerPath)
	if err != nil {
		return nil, err
	}

	// Create JSON directory if it doesn't exist
	if err := os.MkdirAll(jsonDir, 0755); err != nil {
		return nil, err
	}

	udb := &UnifiedDB{
		Badger:    badger,
		JSONDir:   jsonDir,
		DebugMode: debugMode,
	}

	// Start garbage collection routine (every 10 minutes)
	badger.StartGCRoutine(10 * time.Minute)

	// Export to JSON if debug mode is enabled
	if debugMode {
		log.Println("[DB] Debug mode enabled - JSON exports will be created")
		go udb.startJSONExportRoutine(5 * time.Minute)
	}

	GlobalDB = udb
	return udb, nil
}

// Close closes the database
func (udb *UnifiedDB) Close() error {
	if udb.DebugMode {
		// Final export before closing
		if err := udb.Badger.ExportToJSON(udb.JSONDir); err != nil {
			log.Printf("[DB] Warning: failed to export to JSON on close: %v", err)
		}
	}
	return udb.Badger.Close()
}

// startJSONExportRoutine periodically exports BadgerDB to JSON for debugging
func (udb *UnifiedDB) startJSONExportRoutine(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		if err := udb.Badger.ExportToJSON(udb.JSONDir); err != nil {
			log.Printf("[DB] Warning: failed to export to JSON: %v", err)
		}
	}
}

// User operations
func (udb *UnifiedDB) AddUser(user User) error {
	return AddUserBadger(udb.Badger, user)
}

func (udb *UnifiedDB) GetUser(username string) (*User, error) {
	return GetUserBadger(udb.Badger, username)
}

func (udb *UnifiedDB) GetAllUsers() (map[string]User, error) {
	return GetAllUsersBadger(udb.Badger)
}

func (udb *UnifiedDB) RemoveUser(username string) error {
	return RemoveUserBadger(udb.Badger, username)
}

// Session operations
func (udb *UnifiedDB) AddSession(session Session) error {
	return AddSessionBadger(udb.Badger, session)
}

func (udb *UnifiedDB) GetSession(id string) (*Session, error) {
	return GetSessionBadger(udb.Badger, id)
}

func (udb *UnifiedDB) RemoveSession(id string) error {
	return RemoveSessionBadger(udb.Badger, id)
}

// Notebook operations
func (udb *UnifiedDB) CreateNotebook(username, name, description string, optional Optional) (*Notebook, error) {
	return CreateNotebookBadger(udb.Badger, username, name, description, optional)
}

func (udb *UnifiedDB) GetNotebook(username string, id int) (*Notebook, error) {
	return GetNotebookBadger(udb.Badger, username, id)
}

func (udb *UnifiedDB) GetAllNotebooks(username string) ([]Notebook, error) {
	return GetAllNotebooksBadger(udb.Badger, username)
}

func (udb *UnifiedDB) AddItemToNotebook(username string, id int, sheetName, url string) error {
	return AddItemToNotebookBadger(udb.Badger, username, id, sheetName, url)
}

func (udb *UnifiedDB) DeleteNotebook(username string, id int) error {
	return DeleteNotebookBadger(udb.Badger, username, id)
}

func (udb *UnifiedDB) DeleteItemFromNotebook(username string, id int, itemName string) error {
	return DeleteItemFromNotebookBadger(udb.Badger, username, id, itemName)
}

func (udb *UnifiedDB) GetItemsInNotebook(username string, id int) (map[string]string, error) {
	return GetItemsInNotebookBadger(udb.Badger, username, id)
}

func (udb *UnifiedDB) UpdateNotebook(username string, notebook Notebook) error {
	return UpdateNotebookBadger(udb.Badger, username, notebook)
}

// Queue operations
func (udb *UnifiedDB) AddQueuedJob(job QueuedJob) error {
	return AddQueuedJobBadger(udb.Badger, job)
}

func (udb *UnifiedDB) GetQueuedJob(id string) (*QueuedJob, error) {
	return GetQueuedJobBadger(udb.Badger, id)
}

func (udb *UnifiedDB) GetAllQueuedJobs(status string) (map[string]QueuedJob, error) {
	return GetAllQueuedJobsBadger(udb.Badger, status)
}

func (udb *UnifiedDB) UpdateQueuedJobStatus(id, status string, result interface{}) error {
	return UpdateQueuedJobStatusBadger(udb.Badger, id, status, result)
}

func (udb *UnifiedDB) GetQueuedJobsByUser(userID string) ([]QueuedJob, error) {
	return GetQueuedJobsByUserBadger(udb.Badger, userID)
}

func (udb *UnifiedDB) RemoveQueuedJob(id string) error {
	return RemoveQueuedJobBadger(udb.Badger, id)
}

func (udb *UnifiedDB) CleanupOldJobs(maxAge time.Duration) error {
	return CleanupOldJobsBadger(udb.Badger, maxAge)
}

// Style operations
func (udb *UnifiedDB) AddStyle(style Style) error {
	return AddStyleBadger(udb.Badger, style)
}

func (udb *UnifiedDB) GetStyle(username, name string) (*Style, error) {
	return GetStyleBadger(udb.Badger, username, name)
}

func (udb *UnifiedDB) GetAllStyles(username string) ([]Style, error) {
	return GetAllStylesBadger(udb.Badger, username)
}

func (udb *UnifiedDB) DeleteStyle(username, name string) error {
	return DeleteStyleBadger(udb.Badger, username, name)
}

func (udb *UnifiedDB) UpdateStyle(style Style) error {
	return UpdateStyleBadger(udb.Badger, style)
}
