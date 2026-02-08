package store

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// Helper function for JSON unmarshaling
func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// AddQueuedJobBadger adds a job to the queue in BadgerDB
func AddQueuedJobBadger(bdb *BadgerDB, job QueuedJob) error {
	key := fmt.Sprintf("queue:%s", job.ID)
	return bdb.Set(key, job)
}

// GetQueuedJobBadger gets a job by ID from BadgerDB
func GetQueuedJobBadger(bdb *BadgerDB, id string) (*QueuedJob, error) {
	key := fmt.Sprintf("queue:%s", id)
	var job QueuedJob
	err := bdb.Get(key, &job)
	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// GetAllQueuedJobsBadger gets all jobs from BadgerDB, optionally filtered by status
func GetAllQueuedJobsBadger(bdb *BadgerDB, status string) (map[string]QueuedJob, error) {
	jobs := make(map[string]QueuedJob)

	err := bdb.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("queue:")
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())

			err := item.Value(func(val []byte) error {
				var job QueuedJob
				if err := jsonUnmarshal(val, &job); err != nil {
					return err
				}

				// Filter by status if provided
				if status == "" || job.Status == status {
					jobs[key] = job
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return jobs, err
}

// UpdateQueuedJobStatusBadger updates a job's status and result in BadgerDB
func UpdateQueuedJobStatusBadger(bdb *BadgerDB, id, status string, result interface{}) error {
	job, err := GetQueuedJobBadger(bdb, id)
	if err != nil {
		return err
	}
	if job == nil {
		return nil // Job not found, silently ignore
	}

	job.Status = status
	job.Result = result
	job.UpdatedAt = time.Now()

	key := fmt.Sprintf("queue:%s", id)
	return bdb.Set(key, job)
}

// GetQueuedJobsByUserBadger gets all jobs for a specific user from BadgerDB
func GetQueuedJobsByUserBadger(bdb *BadgerDB, userID string) ([]QueuedJob, error) {
	var userJobs []QueuedJob

	err := bdb.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("queue:")
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			err := item.Value(func(val []byte) error {
				var job QueuedJob
				if err := jsonUnmarshal(val, &job); err != nil {
					return err
				}

				if job.UserID == userID {
					userJobs = append(userJobs, job)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return userJobs, err
}

// RemoveQueuedJobBadger removes a job from the queue in BadgerDB
func RemoveQueuedJobBadger(bdb *BadgerDB, id string) error {
	key := fmt.Sprintf("queue:%s", id)
	return bdb.Delete(key)
}

// CleanupOldJobsBadger removes jobs older than the specified duration from BadgerDB
func CleanupOldJobsBadger(bdb *BadgerDB, maxAge time.Duration) error {
	now := time.Now()
	var keysToDelete []string

	err := bdb.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("queue:")
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())

			err := item.Value(func(val []byte) error {
				var job QueuedJob
				if err := jsonUnmarshal(val, &job); err != nil {
					return err
				}

				if now.Sub(job.UpdatedAt) > maxAge {
					keysToDelete = append(keysToDelete, key)
				}
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

	// Delete old jobs
	for _, key := range keysToDelete {
		if err := bdb.Delete(key); err != nil {
			return err
		}
	}

	return nil
}
