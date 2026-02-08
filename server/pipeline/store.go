package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

// Store provides thread-safe persistence for jobs and conversations
type Store struct {
	jobsPath          string
	conversationsPath string
	jobsBackupPath    string
	convBackupPath    string
	jobsMu            sync.RWMutex
	convMu            sync.RWMutex
}

// NewStore creates a new store with the given base directory
func NewStore(baseDir string) (*Store, error) {
	jobsPath := filepath.Join(baseDir, "jobs.json")
	conversationsPath := filepath.Join(baseDir, "conversations.json")
	jobsBackupPath := filepath.Join(baseDir, "jobs.json.bak")
	conversationsBackupPath := filepath.Join(baseDir, "conversations.json.bak")

	// Ensure directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}

	// Initialize files if they don't exist
	if err := initFileIfNotExists(jobsPath, "{}"); err != nil {
		return nil, err
	}
	if err := initFileIfNotExists(conversationsPath, "{}"); err != nil {
		return nil, err
	}

	return &Store{
		jobsPath:          jobsPath,
		conversationsPath: conversationsPath,
		jobsBackupPath:    jobsBackupPath,
		convBackupPath:    conversationsBackupPath,
	}, nil
}

func initFileIfNotExists(path, initialContent string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.WriteFile(path, []byte(initialContent), 0644)
	}
	return nil
}

// SaveJob persists a job to disk (with write lock)
func (s *Store) SaveJob(job *Job) error {
	s.jobsMu.Lock()
	defer s.jobsMu.Unlock()

	jobs, err := s.loadJobsUnsafe()
	if err != nil {
		return err
	}

	jobs[job.ID.String()] = job

	return s.saveJobsUnsafe(jobs)
}

// GetJob retrieves a job by ID (with read lock)
func (s *Store) GetJob(id uuid.UUID) (*Job, error) {
	s.jobsMu.RLock()
	defer s.jobsMu.RUnlock()

	jobs, err := s.loadJobsUnsafe()
	if err != nil {
		return nil, err
	}

	job, exists := jobs[id.String()]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", id)
	}

	return job, nil
}

// GetJobForUpdate retrieves a job with exclusive write lock
// This simulates SELECT ... FOR UPDATE in SQL
func (s *Store) GetJobForUpdate(id uuid.UUID) (*Job, func() error, error) {
	s.jobsMu.Lock()
	// Don't unlock yet - caller must call commit/rollback

	jobs, err := s.loadJobsUnsafe()
	if err != nil {
		s.jobsMu.Unlock()
		return nil, nil, err
	}

	job, exists := jobs[id.String()]
	if !exists {
		s.jobsMu.Unlock()
		return nil, nil, fmt.Errorf("job not found: %s", id)
	}

	// Return commit function that saves and unlocks
	commit := func() error {
		jobs[id.String()] = job
		err := s.saveJobsUnsafe(jobs)
		s.jobsMu.Unlock()
		return err
	}

	return job, commit, nil
}

// GetAllJobs returns all jobs (with read lock)
func (s *Store) GetAllJobs() (map[string]*Job, error) {
	s.jobsMu.RLock()
	defer s.jobsMu.RUnlock()

	return s.loadJobsUnsafe()
}

// GetJobsByUser returns all jobs for a specific user
func (s *Store) GetJobsByUser(userID string) ([]*Job, error) {
	s.jobsMu.RLock()
	defer s.jobsMu.RUnlock()

	jobs, err := s.loadJobsUnsafe()
	if err != nil {
		return nil, err
	}

	var userJobs []*Job
	for _, job := range jobs {
		if job.UserID == userID {
			userJobs = append(userJobs, job)
		}
	}

	return userJobs, nil
}

// GetJobsByStatus returns all jobs with a specific status
func (s *Store) GetJobsByStatus(status JobStatus) ([]*Job, error) {
	s.jobsMu.RLock()
	defer s.jobsMu.RUnlock()

	jobs, err := s.loadJobsUnsafe()
	if err != nil {
		return nil, err
	}

	var filteredJobs []*Job
	for _, job := range jobs {
		if job.Status == status {
			filteredJobs = append(filteredJobs, job)
		}
	}

	return filteredJobs, nil
}

// DeleteJob removes a job from storage
func (s *Store) DeleteJob(id uuid.UUID) error {
	s.jobsMu.Lock()
	defer s.jobsMu.Unlock()

	jobs, err := s.loadJobsUnsafe()
	if err != nil {
		return err
	}

	delete(jobs, id.String())

	return s.saveJobsUnsafe(jobs)
}

// SaveConversation persists a conversation to disk
func (s *Store) SaveConversation(conv *Conversation) error {
	s.convMu.Lock()
	defer s.convMu.Unlock()

	convs, err := s.loadConversationsUnsafe()
	if err != nil {
		return err
	}

	convs[conv.ID.String()] = conv

	return s.saveConversationsUnsafe(convs)
}

// GetConversation retrieves a conversation by ID
func (s *Store) GetConversation(id uuid.UUID) (*Conversation, error) {
	s.convMu.RLock()
	defer s.convMu.RUnlock()

	convs, err := s.loadConversationsUnsafe()
	if err != nil {
		return nil, err
	}

	conv, exists := convs[id.String()]
	if !exists {
		return nil, fmt.Errorf("conversation not found: %s", id)
	}

	return conv, nil
}

// GetConversationByJobID retrieves a conversation by job ID
func (s *Store) GetConversationByJobID(jobID uuid.UUID) (*Conversation, error) {
	s.convMu.RLock()
	defer s.convMu.RUnlock()

	convs, err := s.loadConversationsUnsafe()
	if err != nil {
		return nil, err
	}

	for _, conv := range convs {
		if conv.JobID == jobID {
			return conv, nil
		}
	}

	return nil, fmt.Errorf("conversation not found for job: %s", jobID)
}

// Internal unsafe methods (must be called with lock held)

func (s *Store) loadJobsUnsafe() (map[string]*Job, error) {
	data, err := os.ReadFile(s.jobsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read jobs file: %w", err)
	}

	var jobs map[string]*Job
	if len(data) == 0 || json.Unmarshal(data, &jobs) != nil {
		backup, berr := os.ReadFile(s.jobsBackupPath)
		if berr == nil && len(backup) > 0 {
			if json.Unmarshal(backup, &jobs) == nil {
				return jobs, nil
			}
		}
		return nil, fmt.Errorf("failed to unmarshal jobs")
	}

	if jobs == nil {
		jobs = make(map[string]*Job)
	}

	return jobs, nil
}

func (s *Store) saveJobsUnsafe(jobs map[string]*Job) error {
	data, err := json.MarshalIndent(jobs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal jobs: %w", err)
	}

	if err := atomicWriteFile(s.jobsPath, s.jobsBackupPath, data); err != nil {
		return fmt.Errorf("failed to write jobs file: %w", err)
	}

	return nil
}

func (s *Store) loadConversationsUnsafe() (map[string]*Conversation, error) {
	data, err := os.ReadFile(s.conversationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read conversations file: %w", err)
	}

	var convs map[string]*Conversation
	if len(data) == 0 || json.Unmarshal(data, &convs) != nil {
		backup, berr := os.ReadFile(s.convBackupPath)
		if berr == nil && len(backup) > 0 {
			if json.Unmarshal(backup, &convs) == nil {
				return convs, nil
			}
		}
		return nil, fmt.Errorf("failed to unmarshal conversations")
	}

	if convs == nil {
		convs = make(map[string]*Conversation)
	}

	return convs, nil
}

func (s *Store) saveConversationsUnsafe(convs map[string]*Conversation) error {
	data, err := json.MarshalIndent(convs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal conversations: %w", err)
	}

	if err := atomicWriteFile(s.conversationsPath, s.convBackupPath, data); err != nil {
		return fmt.Errorf("failed to write conversations file: %w", err)
	}

	return nil
}

func atomicWriteFile(path, backupPath string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		_ = os.Rename(path, backupPath)
	}

	if err := os.Rename(tmp.Name(), path); err != nil {
		return err
	}

	return nil
}
