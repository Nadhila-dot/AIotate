package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"nadhi.dev/sarvar/fun/ai"
	"nadhi.dev/sarvar/fun/latex"
	"nadhi.dev/sarvar/fun/websearch"
	ws "nadhi.dev/sarvar/fun/websocket"
)

// Queue manages job processing with a simple worker pool
type Queue struct {
	jobs      chan uuid.UUID
	store     *Store
	logger    *log.Logger
	wg        sync.WaitGroup
	updates   chan StatusUpdate
	mu        sync.Mutex
	listeners map[uuid.UUID]func(StatusUpdate)
}

// NewQueue creates a new queue with the specified capacity
func NewQueue(size int, store *Store, logger *log.Logger) *Queue {
	if logger == nil {
		logger = log.Default()
	}

	return &Queue{
		jobs:      make(chan uuid.UUID, size),
		store:     store,
		logger:    logger,
		updates:   make(chan StatusUpdate, 100),
		listeners: make(map[uuid.UUID]func(StatusUpdate)),
	}
}

// RegisterJobListener registers a callback for a specific job ID
func (q *Queue) RegisterJobListener(jobID uuid.UUID, cb func(StatusUpdate)) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.listeners[jobID] = cb
}

// Start initializes worker goroutines
func (q *Queue) Start(ctx context.Context, workers int) {
	q.logger.Printf("Starting queue with %d workers", workers)

	// Start status update handler
	q.wg.Add(1)
	go q.statusUpdateHandler(ctx)

	// Start workers
	for i := 0; i < workers; i++ {
		q.wg.Add(1)
		go q.worker(ctx, i)
	}
}

// Stop gracefully shuts down the queue
func (q *Queue) Stop() {
	q.logger.Println("Stopping queue")
	close(q.jobs)
	close(q.updates)
	q.wg.Wait()
}

// Enqueue adds a job to the processing queue
func (q *Queue) Enqueue(jobID uuid.UUID) error {
	// Verify job exists
	_, err := q.store.GetJob(jobID)
	if err != nil {
		return fmt.Errorf("cannot enqueue non-existent job: %w", err)
	}

	select {
	case q.jobs <- jobID:
		q.logger.Printf("Enqueued job %s", jobID)
		return nil
	default:
		return fmt.Errorf("queue is full")
	}
}

// worker processes jobs from the queue
func (q *Queue) worker(ctx context.Context, id int) {
	defer q.wg.Done()
	q.logger.Printf("Worker %d started", id)

	for {
		select {
		case <-ctx.Done():
			q.logger.Printf("Worker %d shutting down", id)
			return

		case jobID, ok := <-q.jobs:
			if !ok {
				q.logger.Printf("Worker %d shutting down (channel closed)", id)
				return
			}

			q.logger.Printf("Worker %d processing job %s", id, jobID)
			if err := q.processJob(ctx, jobID); err != nil {
				q.logger.Printf("Worker %d: job %s failed: %v", id, jobID, err)
			}
		}
	}
}

// processJob executes the full pipeline for a single job in one pass.
// It holds the store write lock for the duration, so individual steps
// must NOT call Enqueue (which would deadlock on the store mutex).
func (q *Queue) processJob(ctx context.Context, jobID uuid.UUID) error {
	// Acquire exclusive lock on job
	job, commit, err := q.store.GetJobForUpdate(jobID)
	if err != nil {
		return fmt.Errorf("failed to lock job: %w", err)
	}
	defer func() {
		if err := commit(); err != nil {
			q.logger.Printf("Failed to commit job %s: %v", jobID, err)
		}
	}()

	// Check if job is in a processable state
	if job.Status != StatusPending && job.Status != StatusRunning {
		q.logger.Printf("Job %s is in state %s, skipping", jobID, job.Status)
		return nil
	}

	// Mark as running
	job.Status = StatusRunning
	q.sendUpdate(job, "Job processing started", q.stageData("Pipeline", "Job processing started", nil))

	// Run all pipeline steps in sequence
	for {
		var stepErr error
		switch job.CurrentStep {
		case StepPrompt:
			stepErr = q.executePromptStep(ctx, job)
		case StepDesign:
			stepErr = q.executeDesignStep(ctx, job)
		case StepLatex:
			stepErr = q.executeLatexStep(ctx, job)
		case StepCompile:
			stepErr = q.executeCompileStep(ctx, job)
		case StepDone:
			return nil
		default:
			return fmt.Errorf("unknown step: %s", job.CurrentStep)
		}

		if stepErr != nil {
			return stepErr
		}

		// If the step didn't advance (e.g. completed/errored), stop
		if job.Status != StatusPending && job.Status != StatusRunning {
			return nil
		}

		// Reset status for next step
		job.Status = StatusRunning
	}
}

// executePromptStep processes the initial prompt
func (q *Queue) executePromptStep(ctx context.Context, job *Job) error {
	q.sendUpdate(job, "Processing prompt", q.stageData("Prompt", "Validating request", nil))

	if job.Prompt == "" {
		job.SetError("Empty prompt", nil)
		q.sendUpdate(job, "Prompt validation failed", q.errorData("Empty prompt"))
		return fmt.Errorf("empty prompt")
	}

	if _, err := q.parseRequest(job); err != nil {
		msg := fmt.Sprintf("Invalid request: %v", err)
		job.SetError(msg, nil)
		q.sendUpdate(job, "Prompt validation failed", q.errorData(msg))
		return err
	}

	job.AdvanceStep()
	job.Status = StatusPending
	q.sendUpdate(job, "Prompt validated, moving to design", q.stageData("Prompt", "Validated", nil))

	return nil
}

// executeDesignStep generates the design from the prompt
func (q *Queue) executeDesignStep(ctx context.Context, job *Job) error {
	q.sendUpdate(job, "Generating design", q.stageData("Design", "Generating design", nil))

	request, err := q.parseRequest(job)
	if err != nil {
		msg := fmt.Sprintf("Invalid request: %v", err)
		job.SetError(msg, nil)
		q.sendUpdate(job, "Design generation failed", q.errorData(msg))
		return err
	}

	conv, convErr := q.store.GetConversationByJobID(job.ID)
	if convErr != nil {
		conv = NewConversation(job.ID)
		_ = q.store.SaveConversation(conv)
	}

	designPrompt := q.formatDesignPrompt(request)

	if request.WebSearchEnabled && strings.TrimSpace(request.WebSearchQuery) != "" {
		webContext, _, err := websearch.SearchAndExtract(request.WebSearchQuery, 3)
		if err != nil {
			q.sendUpdate(job, "Web search failed, continuing without web context", q.stageData("WebSearch", "Failed", map[string]interface{}{"error": err.Error()}))
		} else {
			designPrompt = designPrompt + "\n\n" + webContext
			q.sendUpdate(job, "Web search completed, context added", q.stageData("WebSearch", "Completed", nil))
		}
	}

	design, err := GenerateDesign(ctx, conv, designPrompt, request.Attachments)
	if err != nil {
		if job.CanRetry() {
			job.IncrementRetry()
			job.ResetToStep(StepDesign)
			job.Status = StatusRunning
			q.sendUpdate(job, "Design generation failed, retrying", q.retryData(job, err))
			return q.executeDesignStep(ctx, job)
		}
		msg := fmt.Sprintf("Design generation failed: %v", err)
		job.SetError(msg, nil)
		q.sendUpdate(job, "Design generation failed", q.errorData(msg))
		return err
	}

	job.Design = design
	_ = q.store.SaveConversation(conv)

	q.sendUpdate(job, "Design generated, advancing to LaTeX", q.stageData("Design", "Design generated", nil))

	job.AdvanceStep()
	job.Status = StatusPending
	return nil
}

// executeLatexStep generates LaTeX from the design
func (q *Queue) executeLatexStep(ctx context.Context, job *Job) error {
	q.sendUpdate(job, "Generating LaTeX", q.stageData("LaTeX", "Generating LaTeX", nil))

	request, err := q.parseRequest(job)
	if err != nil {
		msg := fmt.Sprintf("Invalid request: %v", err)
		job.SetError(msg, nil)
		q.sendUpdate(job, "LaTeX generation failed", q.errorData(msg))
		return err
	}

	conv, convErr := q.store.GetConversationByJobID(job.ID)
	if convErr != nil {
		conv = NewConversation(job.ID)
		_ = q.store.SaveConversation(conv)
	}

	stylePrompt := ai.ResolveStylePrompt(request)
	latexOutput, err := GenerateLatex(ctx, conv, job.Design, stylePrompt, request.Attachments)
	if err != nil {
		if job.CanRetry() {
			job.IncrementRetry()
			job.ResetToStep(StepLatex)
			job.Status = StatusRunning
			q.sendUpdate(job, "LaTeX generation failed, retrying", q.retryData(job, err))
			return q.executeLatexStep(ctx, job)
		}
		msg := fmt.Sprintf("LaTeX generation failed: %v", err)
		job.SetError(msg, nil)
		q.sendUpdate(job, "LaTeX generation failed", q.errorData(msg))
		return err
	}

	job.Latex = latexOutput
	_ = q.store.SaveConversation(conv)

	q.sendUpdate(job, "LaTeX generated, compiling PDF", q.stageData("LaTeX", "LaTeX generated", nil))

	job.AdvanceStep()
	job.Status = StatusPending
	return nil
}

// executeCompileStep compiles the LaTeX to PDF
func (q *Queue) executeCompileStep(ctx context.Context, job *Job) error {
	q.sendUpdate(job, "Compiling LaTeX to PDF", q.stageData("Compile", "Compiling LaTeX", nil))

	if strings.TrimSpace(job.Latex) == "" {
		msg := "No LaTeX available for compilation"
		job.SetError(msg, nil)
		q.sendUpdate(job, "Compilation failed", q.errorData(msg))
		return fmt.Errorf("%s", msg)
	}

	outputDir := filepath.Join("./storage", "bucket")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		msg := fmt.Sprintf("Failed to create bucket directory: %v", err)
		job.SetError(msg, nil)
		q.sendUpdate(job, "Compilation failed", q.errorData(msg))
		return err
	}

	// Save LaTeX and metadata for audit/debug
	generatedDir := filepath.Join("./generated", job.ID.String())
	if err := os.MkdirAll(generatedDir, 0755); err == nil {
		texPath := filepath.Join(generatedDir, fmt.Sprintf("%s.tex", job.ID.String()))
		_ = os.WriteFile(texPath, []byte(job.Latex), 0644)
		metadata := map[string]interface{}{
			"generated": time.Now().Format(time.RFC3339),
			"source":    "pipeline",
		}
		metaPath := filepath.Join(generatedDir, fmt.Sprintf("%s.meta.json", job.ID.String()))
		if metaJSON, err := json.MarshalIndent(metadata, "", "  "); err == nil {
			_ = os.WriteFile(metaPath, metaJSON, 0644)
		}
		if job.Metadata == nil {
			job.Metadata = make(map[string]interface{})
		}
		job.Metadata["metadata"] = metadata
	}

	texFilename := fmt.Sprintf("%s.tex", job.ID.String())
	pdfFilename := fmt.Sprintf("%s.pdf", job.ID.String())
	outputPath := filepath.Join(outputDir, pdfFilename)

	_, err := latex.ConvertLatexToPDFWithRetry(job.Latex, texFilename, outputPath)
	if err != nil {
		msg := fmt.Sprintf("LaTeX compilation failed: %v", err)
		job.SetError(msg, nil)
		q.sendUpdate(job, "Compilation failed", q.errorData(msg))
		return err
	}

	if job.Metadata == nil {
		job.Metadata = make(map[string]interface{})
	}
	if _, ok := job.Metadata["metadata"]; !ok {
		job.Metadata["metadata"] = map[string]interface{}{
			"generated": time.Now().Format(time.RFC3339),
			"source":    "pipeline",
		}
	}

	pdfURL := fmt.Sprintf("/vela/bucket/bucket/%s", pdfFilename)
	job.SetCompleted(pdfURL)

	q.sendUpdate(job, "Compilation completed successfully", ws.Completed("Sheet generation completed", map[string]interface{}{
		"pdf_url":  pdfURL,
		"metadata": job.Metadata["metadata"],
	}, map[string]interface{}{})["data"].(map[string]interface{}))

	return nil
}

// sendUpdate sends a status update to the update channel
func (q *Queue) sendUpdate(job *Job, message string, data map[string]interface{}) {
	job.UpdatedAt = time.Now()
	update := StatusUpdate{
		JobID:     job.ID,
		Status:    job.Status,
		Step:      job.CurrentStep,
		Message:   message,
		Timestamp: job.UpdatedAt,
		Data:      data,
	}

	select {
	case q.updates <- update:
	default:
		q.logger.Printf("Warning: status update channel full, dropping update for job %s", job.ID)
	}

	q.mu.Lock()
	listener, exists := q.listeners[job.ID]
	q.mu.Unlock()
	if exists && listener != nil {
		listener(update)
	}
}

// statusUpdateHandler processes status updates
func (q *Queue) statusUpdateHandler(ctx context.Context) {
	defer q.wg.Done()
	q.logger.Println("Status update handler started")

	for {
		select {
		case <-ctx.Done():
			q.logger.Println("Status update handler shutting down")
			return

		case update, ok := <-q.updates:
			if !ok {
				q.logger.Println("Status update handler shutting down (channel closed)")
				return
			}

			// TODO: Send to websocket
			q.logger.Printf("Status update: job=%s status=%s step=%s message=%s",
				update.JobID, update.Status, update.Step, update.Message)
		}
	}
}

// GetUpdates returns the status update channel for external consumers
func (q *Queue) GetUpdates() <-chan StatusUpdate {
	return q.updates
}

// EmitUpdate allows external callers to emit a status update for a job
func (q *Queue) EmitUpdate(job *Job, message string, data map[string]interface{}) {
	if job == nil {
		return
	}
	q.sendUpdate(job, message, data)
}

func (q *Queue) stageData(stage string, step string, extra map[string]interface{}) map[string]interface{} {
	return ws.Stage(stage, step, extra)["data"].(map[string]interface{})
}

func (q *Queue) errorData(message string) map[string]interface{} {
	return ws.Error("Pipeline error", message, map[string]interface{}{})["data"].(map[string]interface{})
}

func (q *Queue) retryData(job *Job, err error) map[string]interface{} {
	return ws.Retry("Retrying", map[string]interface{}{
		"retries":   job.RetryCount,
		"maxRetry":  job.MaxRetries,
		"willRetry": job.CanRetry(),
		"error":     err.Error(),
	})["data"].(map[string]interface{})
}

func (q *Queue) parseRequest(job *Job) (*ai.GenerationRequest, error) {
	var req ai.GenerationRequest
	if err := json.Unmarshal([]byte(job.Prompt), &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func (q *Queue) formatDesignPrompt(req *ai.GenerationRequest) string {
	if req == nil {
		return ""
	}

	tags := strings.Join(req.Tags, ", ")
	mode := req.Mode
	if mode == "" {
		mode = "notes"
	}

	modeInstructions := getModeInstructions(mode)
	attachmentContext := formatAttachmentContext(req.Attachments)

	return fmt.Sprintf(
		"Subject: %s\nCourse: %s\nDescription: %s\nTags: %s\nCurriculum: %s\nSpecial Instructions: %s\n\nGeneration Mode: %s\n%s\n\nAdditional Context:\n%s",
		req.Subject,
		req.Course,
		req.Description,
		tags,
		req.Curriculum,
		req.SpecialInstructions,
		mode,
		modeInstructions,
		attachmentContext,
	)
}

func formatAttachmentContext(attachments []ai.Attachment) string {
	if len(attachments) == 0 {
		return "(none)"
	}

	var b strings.Builder
	for i, att := range attachments {
		b.WriteString(fmt.Sprintf("[%d] %s (%s, %d bytes, %s)\n", i+1, att.Name, att.MimeType, att.Size, att.Encoding))
		content := att.Content
		if len(content) > 20000 {
			content = content[:20000] + "\n[TRUNCATED]"
		}
		b.WriteString(content)
		b.WriteString("\n---\n")
	}

	return b.String()
}

// getModeInstructions returns mode-specific AI instructions
func getModeInstructions(mode string) string {
	switch mode {
	case "prep-test":
		return `MODE: PREP TEST
You are generating a practice test / exam paper.

Requirements:
- Create a complete test paper with clear sections
- Include a mix of question types: multiple choice, short answer, long answer, and problem-solving
- Vary difficulty: easy (30%), medium (50%), hard (20%)
- Include point values for each question
- Add a clear header with subject, course, date, and time limit
- Include instructions section at the top
- Add space for student name and ID
- Provide an answer key section at the end
- Make questions that genuinely test understanding, not just memorization
- Include at least 15-25 questions depending on complexity
- Group questions by topic or section
- Use professional exam formatting`

	case "super-lazy":
		return `MODE: SUPER LAZY
You are generating a study document optimized for maximum retention with minimum effort.

Requirements:
- Use proven memory techniques: spaced repetition cues, mnemonics, chunking, and visual anchors
- Structure content as KEY POINTS with bold highlights for critical terms
- Use the "explain like I'm 5" approach for complex concepts
- Include quick-fire summary boxes at the end of each section
- Add "Remember This" callout boxes with memory tricks and acronyms
- Use comparison tables to contrast similar concepts
- Include a one-page "cheat sheet" summary at the end with EVERYTHING essential
- Create "If you only read ONE thing" highlights per section
- Use bullet points extensively, avoid long paragraphs
- Add visual separators between concepts
- Include practice recall prompts ("Can you explain X without looking?")
- Make at least 4-5 pages of content
- Design it so someone reading it the night before an exam WILL pass with excellence
- Prioritize the 20% of content that covers 80% of what's tested
- Use casual, engaging tone - not dry textbook language`

	default: // "notes" mode
		return `MODE: NOTES
You are generating comprehensive, professional study notes.

Requirements:
- Create at least 3 pages of thorough, well-structured notes
- Use a clean, professional document design with clear hierarchy
- Include numbered sections and subsections
- Add definitions, theorems, and key concepts in highlighted boxes
- Include worked examples where relevant
- Use proper mathematical notation where applicable
- Add summary points at the end of each major section
- Include diagrams descriptions where they would help understanding
- Use professional typography: proper headings, consistent spacing, clear fonts
- Make it comprehensive enough to be a standalone study resource
- Include a table of contents if content is substantial
- Add page numbers and proper headers/footers`
	}
}
