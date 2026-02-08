# Pipeline System - Deterministic Job Processing

This package implements a robust, deterministic pipeline for sheet generation with proper error handling, state management, and retry logic.

## Architecture

### Core Principles

1. **Deterministic Pipeline**: Each job progresses through explicit stages
2. **Strong Data Model**: Job lifecycle is stored, not inferred
3. **Database Safety**: Row-level locking prevents corruption
4. **Simple Queue**: Worker pool + channel, no clever abstractions
5. **Conversation Trains**: Persistent dialogue threads for AI context
6. **Explicit Error Handling**: No silent failures, full error capture

### Pipeline Stages

```
Prompt → Design → LaTeX → Compile → SUCCESS
           ↓         ↓        ↓
         Error → Manual Edit OR AI Fix → Retry
```

Each step is:
- **Atomic**: Completes fully or fails cleanly
- **Persisted**: State saved to disk
- **Resumable**: Can continue after restart
- **Isolated**: Cannot corrupt other jobs

## Components

### Types (`types.go`)

**JobStatus**:
- `pending`: Ready to process
- `running`: Currently being processed
- `error`: Failed with error
- `waiting_manual`: Waiting for user intervention
- `waiting_ai_fix`: Waiting for AI fix attempt
- `completed`: Successfully finished
- `aborted`: User cancelled

**PipelineStep**:
- `prompt`: Initial prompt validation
- `design`: Design generation
- `latex`: LaTeX code generation
- `compile`: PDF compilation
- `done`: Completed

**Job**: Complete job state including:
- ID, UserID, Status, CurrentStep
- Prompt, Design, Latex, PDFURL
- ErrorMessage, ErrorLog
- ConversationID for dialogue tracking
- RetryCount, MaxRetries
- Timestamps

**Conversation**: Persistent dialogue thread
- Messages with role (system/user/assistant)
- Linked to job via ConversationID
- Maintains full context across retries

### Store (`store.go`)

Thread-safe persistence layer with:
- **Read locks**: Multiple readers can access simultaneously
- **Write locks**: Exclusive access for updates
- **GetJobForUpdate**: Simulates `SELECT ... FOR UPDATE`
- JSON file storage (easily replaceable with SQL)

**Key Methods**:
```go
SaveJob(job *Job) error
GetJob(id uuid.UUID) (*Job, error)
GetJobForUpdate(id uuid.UUID) (*Job, func() error, error)
GetJobsByStatus(status JobStatus) ([]*Job, error)
```

### Queue (`queue.go`)

Simple, predictable worker pool:
```go
type Queue struct {
    jobs    chan uuid.UUID  // Job IDs to process
    store   *Store          // Persistent storage
    updates chan StatusUpdate // Status notifications
}
```

**Worker Flow**:
1. Pull job ID from channel
2. Acquire exclusive lock via `GetJobForUpdate`
3. Execute current pipeline step
4. Save state and release lock
5. Re-enqueue if more steps remain

**No Race Conditions**: Lock prevents concurrent access to same job.

### AI Integration (`ai.go`)

**System Prompt** (fixed):
```
You are a deterministic document generation engine.
Rules:
- Output ONLY valid LaTeX
- Do not explain
- Do not apologize
- Do not include markdown code blocks
- Use only standard packages
- Never invent data
```

**Functions**:
- `GenerateDesign`: Creates design spec from prompt
- `GenerateLatex`: Generates LaTeX from design
- `FixLatex`: Attempts to fix compilation errors
- `RefinePrompt`: Iterative refinement
- `GenerateDescription`: Short description
- `GenerateTags`: Tag generation

All functions use conversation context for continuity.

## Usage

### Initialize

```go
import "nadhi.dev/sarvar/fun/pipeline"

// Create store
store, err := pipeline.NewStore("./storage/pipeline")
if err != nil {
    log.Fatal(err)
}

// Create queue
queue := pipeline.NewQueue(100, store, logger)

// Start workers
ctx := context.Background()
queue.Start(ctx, 2) // 2 workers
defer queue.Stop()
```

### Create and Enqueue Job

```go
// Create job
job := pipeline.NewJob(userID, prompt, 3) // max 3 retries

// Save to store
if err := store.SaveJob(job); err != nil {
    return err
}

// Enqueue for processing
if err := queue.Enqueue(job.ID); err != nil {
    return err
}
```

### Monitor Status

```go
// Get status updates
updates := queue.GetUpdates()

go func() {
    for update := range updates {
        fmt.Printf("Job %s: %s - %s\n", 
            update.JobID, update.Status, update.Message)
        
        // Send to websocket, etc.
    }
}()
```

### Manual Intervention

```go
// Get job with lock
job, commit, err := store.GetJobForUpdate(jobID)
if err != nil {
    return err
}
defer commit()

// User edits LaTeX
job.Latex = userEditedLatex

// Reset to compile step
job.ResetToStep(pipeline.StepCompile)

// Re-enqueue
queue.Enqueue(job.ID)
```

### AI Fix Attempt

```go
// Load conversation
conv, err := store.GetConversationByJobID(jobID)
if err != nil {
    return err
}

// Attempt fix
fixedLatex, err := pipeline.FixLatex(ctx, conv, job.Latex, errorLog)
if err != nil {
    // Fix failed, wait for manual intervention
    job.SetWaitingManual("AI fix failed")
    return
}

// Apply fix
job.Latex = fixedLatex
job.ResetToStep(pipeline.StepCompile)
job.IncrementRetry()

// Save conversation
store.SaveConversation(conv)

// Re-enqueue
queue.Enqueue(job.ID)
```

## Error Handling

### Retry Policy

```go
const MaxRetries = 3

if job.RetryCount >= job.MaxRetries {
    job.SetWaitingManual("Max retries exceeded")
} else {
    job.IncrementRetry()
    job.ResetToStep(failedStep)
    queue.Enqueue(job.ID)
}
```

### Error States

**StatusError**: Temporary failure, will retry
**StatusWaitingManual**: Needs user intervention
**StatusWaitingAIFix**: AI fix in progress

### LaTeX Errors

```go
type LatexError struct {
    Log     string // Full compilation log
    Snippet string // Relevant error section
    Line    int    // Error line number
}
```

Stored in `job.ErrorLog` for user review.

## Conversation Trains

Maintains context across the entire job lifecycle:

```go
conv := pipeline.NewConversation(jobID)

// Initial prompt
conv.AddMessage("user", prompt)
design, _ := pipeline.GenerateDesign(ctx, conv, prompt)

// Refinement
conv.AddMessage("user", "Make it more minimal")
refinedDesign, _ := pipeline.RefinePrompt(ctx, conv, "Make it more minimal")

// LaTeX generation (with full context)
latex, _ := pipeline.GenerateLatex(ctx, conv, refinedDesign)

// Error fix (with full context)
if compilationFailed {
    fixedLatex, _ := pipeline.FixLatex(ctx, conv, latex, errorLog)
}
```

No state loss. No hallucinated resets.

## Database Safety

### File-Based Locking

```go
// Acquire exclusive lock
job, commit, err := store.GetJobForUpdate(jobID)
if err != nil {
    return err
}

// Lock is held until commit() is called
defer commit()

// Modify job safely
job.Status = StatusRunning
job.Latex = newLatex

// commit() saves and releases lock
```

### Migration to SQL

Easy to replace with Postgres:

```go
func (s *Store) GetJobForUpdate(id uuid.UUID) (*Job, func() error, error) {
    tx, err := db.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelSerializable,
    })
    if err != nil {
        return nil, nil, err
    }

    row := tx.QueryRow("SELECT * FROM jobs WHERE id = $1 FOR UPDATE", id)
    // ... scan job

    commit := func() error {
        return tx.Commit()
    }

    return job, commit, nil
}
```

## Testing

### Unit Tests

```go
func TestJobLifecycle(t *testing.T) {
    store, _ := pipeline.NewStore(t.TempDir())
    
    job := pipeline.NewJob("user1", "test prompt", 3)
    store.SaveJob(job)
    
    // Verify initial state
    assert.Equal(t, pipeline.StatusPending, job.Status)
    assert.Equal(t, pipeline.StepPrompt, job.CurrentStep)
    
    // Advance through pipeline
    job.AdvanceStep()
    assert.Equal(t, pipeline.StepDesign, job.CurrentStep)
}
```

### Integration Tests

```go
func TestFullPipeline(t *testing.T) {
    store, _ := pipeline.NewStore(t.TempDir())
    queue := pipeline.NewQueue(10, store, logger)
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    queue.Start(ctx, 1)
    defer queue.Stop()
    
    job := pipeline.NewJob("user1", "Create a math worksheet", 3)
    store.SaveJob(job)
    queue.Enqueue(job.ID)
    
    // Wait for completion
    time.Sleep(10 * time.Second)
    
    finalJob, _ := store.GetJob(job.ID)
    assert.Equal(t, pipeline.StatusCompleted, finalJob.Status)
}
```

## Migration from Old System

### Key Differences

| Old System | New System |
|------------|------------|
| Status inferred from data | Explicit JobStatus enum |
| No locking | Row-level locking |
| Retry logic scattered | Centralized retry policy |
| No conversation context | Persistent conversation trains |
| Silent failures | Full error capture |
| Forced AI fixes | Manual edit option |
| Chaotic queue | Deterministic pipeline |

### Migration Steps

1. **Create pipeline storage**: `./storage/pipeline/`
2. **Initialize new queue**: Replace `SheetQueue` with `pipeline.Queue`
3. **Migrate existing jobs**: Convert to new `Job` format
4. **Update API endpoints**: Use new job IDs (UUID)
5. **Add manual edit endpoint**: `POST /jobs/{id}/latex`
6. **Update websocket**: Send `StatusUpdate` events
7. **Test thoroughly**: Verify no data loss

## Future Enhancements

- [ ] Postgres backend for production scale
- [ ] Job priority queue
- [ ] Scheduled cleanup of old jobs
- [ ] Job cancellation
- [ ] Batch job processing
- [ ] Metrics and monitoring
- [ ] Job dependencies
- [ ] Webhook notifications

## Troubleshooting

### Job Stuck in Running

```go
// Find stuck jobs
jobs, _ := store.GetJobsByStatus(pipeline.StatusRunning)
for _, job := range jobs {
    if time.Since(job.UpdatedAt) > 10*time.Minute {
        job.Status = pipeline.StatusError
        job.SetError("Job timeout", nil)
        store.SaveJob(job)
    }
}
```

### Queue Full

```go
// Increase queue size
queue := pipeline.NewQueue(1000, store, logger) // was 100
```

### High Retry Rate

Check AI prompts and LaTeX validation logic.

## License

Part of AIotate project.
