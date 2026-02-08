package pipeline

import (
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the current state of a job
type JobStatus string

const (
	StatusPending       JobStatus = "pending"
	StatusRunning       JobStatus = "running"
	StatusError         JobStatus = "error"
	StatusWaitingManual JobStatus = "waiting_manual"
	StatusWaitingAIFix  JobStatus = "waiting_ai_fix"
	StatusCompleted     JobStatus = "completed"
	StatusAborted       JobStatus = "aborted"
)

// PipelineStep represents a stage in the generation pipeline
type PipelineStep string

const (
	StepPrompt  PipelineStep = "prompt"
	StepDesign  PipelineStep = "design"
	StepLatex   PipelineStep = "latex"
	StepCompile PipelineStep = "compile"
	StepDone    PipelineStep = "done"
)

// Job represents a sheet generation job with full state tracking
type Job struct {
	ID             uuid.UUID              `json:"id"`
	UserID         string                 `json:"userId"`
	Status         JobStatus              `json:"status"`
	CurrentStep    PipelineStep           `json:"currentStep"`
	Prompt         string                 `json:"prompt"`
	Design         string                 `json:"design"`
	Latex          string                 `json:"latex"`
	PDFURL         string                 `json:"pdfUrl,omitempty"`
	ErrorMessage   *string                `json:"errorMessage,omitempty"`
	ErrorLog       *string                `json:"errorLog,omitempty"`
	ConversationID uuid.UUID              `json:"conversationId"`
	RetryCount     int                    `json:"retryCount"`
	MaxRetries     int                    `json:"maxRetries"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
	CompletedAt    *time.Time             `json:"completedAt,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// Conversation represents a persistent dialogue thread for a job
type Conversation struct {
	ID        uuid.UUID `json:"id"`
	JobID     uuid.UUID `json:"jobId"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Message represents a single message in a conversation
type Message struct {
	Role      string    `json:"role"` // system, user, assistant
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// LatexError contains detailed information about LaTeX compilation failures
type LatexError struct {
	Log     string `json:"log"`
	Snippet string `json:"snippet"`
	Line    int    `json:"line,omitempty"`
}

// StatusUpdate represents a job status change event
type StatusUpdate struct {
	JobID     uuid.UUID              `json:"jobId"`
	Status    JobStatus              `json:"status"`
	Step      PipelineStep           `json:"step"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewJob creates a new job with initial state
func NewJob(userID, prompt string, maxRetries int) *Job {
	now := time.Now()
	return &Job{
		ID:             uuid.New(),
		UserID:         userID,
		Status:         StatusPending,
		CurrentStep:    StepPrompt,
		Prompt:         prompt,
		ConversationID: uuid.New(),
		RetryCount:     0,
		MaxRetries:     maxRetries,
		CreatedAt:      now,
		UpdatedAt:      now,
		Metadata:       make(map[string]interface{}),
	}
}

// NewConversation creates a new conversation for a job
func NewConversation(jobID uuid.UUID) *Conversation {
	now := time.Now()
	return &Conversation{
		ID:        uuid.New(),
		JobID:     jobID,
		Messages:  []Message{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddMessage adds a message to the conversation
func (c *Conversation) AddMessage(role, content string) {
	c.Messages = append(c.Messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	c.UpdatedAt = time.Now()
}

// CanRetry checks if the job can be retried
func (j *Job) CanRetry() bool {
	return j.RetryCount < j.MaxRetries
}

// IncrementRetry increments the retry counter
func (j *Job) IncrementRetry() {
	j.RetryCount++
	j.UpdatedAt = time.Now()
}

// SetError sets the job to error state with a message
func (j *Job) SetError(message string, log *string) {
	j.Status = StatusError
	j.ErrorMessage = &message
	j.ErrorLog = log
	j.UpdatedAt = time.Now()
}

// SetWaitingManual sets the job to waiting for manual intervention
func (j *Job) SetWaitingManual(message string) {
	j.Status = StatusWaitingManual
	j.ErrorMessage = &message
	j.UpdatedAt = time.Now()
}

// SetCompleted marks the job as completed
func (j *Job) SetCompleted(pdfURL string) {
	j.Status = StatusCompleted
	j.CurrentStep = StepDone
	j.PDFURL = pdfURL
	now := time.Now()
	j.CompletedAt = &now
	j.UpdatedAt = now
}

// AdvanceStep moves the job to the next pipeline step
func (j *Job) AdvanceStep() {
	switch j.CurrentStep {
	case StepPrompt:
		j.CurrentStep = StepDesign
	case StepDesign:
		j.CurrentStep = StepLatex
	case StepLatex:
		j.CurrentStep = StepCompile
	case StepCompile:
		j.CurrentStep = StepDone
	}
	j.UpdatedAt = time.Now()
}

// ResetToStep resets the job to a specific step (for retries)
func (j *Job) ResetToStep(step PipelineStep) {
	j.CurrentStep = step
	j.Status = StatusPending
	j.ErrorMessage = nil
	j.ErrorLog = nil
	j.UpdatedAt = time.Now()
}
