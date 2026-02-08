package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"nadhi.dev/sarvar/fun/pipeline"
	"nadhi.dev/sarvar/fun/server"
	sheet "nadhi.dev/sarvar/fun/sheets"
	ws "nadhi.dev/sarvar/fun/websocket"
)

func PipelineIndex() error {
	server.Route.Get("/api/v1/pipeline/jobs/:id", func(c *fiber.Ctx) error {
		username, err := getUsernameFromAuth(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
		}

		jobID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid job id"})
		}

		job, err := sheet.GlobalPipelineStore.GetJob(jobID)
		if err != nil || job.UserID != username {
			return c.Status(404).JSON(fiber.Map{"error": "job not found"})
		}

		return c.JSON(job)
	})

	server.Route.Post("/api/v1/pipeline/jobs/:id/design/approve", func(c *fiber.Ctx) error {
		return handlePipelineDesignApprove(c)
	})

	server.Route.Post("/api/v1/pipeline/jobs/:id/design/refine", func(c *fiber.Ctx) error {
		return handlePipelineDesignRefine(c)
	})

	server.Route.Post("/api/v1/pipeline/jobs/:id/latex/approve", func(c *fiber.Ctx) error {
		return handlePipelineLatexApprove(c)
	})

	server.Route.Post("/api/v1/pipeline/jobs/:id/latex/edit", func(c *fiber.Ctx) error {
		return handlePipelineLatexEdit(c)
	})

	server.Route.Post("/api/v1/pipeline/jobs/:id/latex/fix", func(c *fiber.Ctx) error {
		return handlePipelineLatexFix(c)
	})

	server.Route.Post("/api/v1/pipeline/jobs/:id/abort", func(c *fiber.Ctx) error {
		return handlePipelineAbort(c)
	})

	server.Route.Post("/api/v1/pipeline/jobs/:id/retry", func(c *fiber.Ctx) error {
		return handlePipelineRetry(c)
	})

	return nil
}

func handlePipelineDesignApprove(c *fiber.Ctx) error {
	job, _, err := getPipelineJobForUser(c)
	if err != nil {
		return err
	}

	job.CurrentStep = pipeline.StepLatex
	job.Status = pipeline.StatusPending
	job.UpdatedAt = time.Now()

	if err := sheet.GlobalPipelineStore.SaveJob(job); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to save job"})
	}

	sheet.GlobalPipelineQueue.EmitUpdate(job, "Design approved, generating LaTeX", ws.Stage("Design", "Approved", nil)["data"].(map[string]interface{}))
	_ = sheet.GlobalPipelineQueue.Enqueue(job.ID)

	return c.JSON(fiber.Map{"status": "queued", "jobId": job.ID.String()})
}

func handlePipelineDesignRefine(c *fiber.Ctx) error {
	job, _, err := getPipelineJobForUser(c)
	if err != nil {
		return err
	}

	var body struct {
		Refinement string `json:"refinement"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	refinement := strings.TrimSpace(body.Refinement)
	if refinement == "" {
		return c.Status(400).JSON(fiber.Map{"error": "refinement required"})
	}

	conv, convErr := sheet.GlobalPipelineStore.GetConversationByJobID(job.ID)
	if convErr != nil {
		conv = pipeline.NewConversation(job.ID)
		_ = sheet.GlobalPipelineStore.SaveConversation(conv)
	}

	prompt := fmt.Sprintf("Refine the design based on this feedback: %s\n\nCurrent design:\n%s", refinement, job.Design)
	refined, err := pipeline.RefinePrompt(context.Background(), conv, prompt)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to refine design"})
	}

	job.Design = refined
	job.Status = pipeline.StatusWaitingManual
	job.CurrentStep = pipeline.StepDesign
	job.UpdatedAt = time.Now()

	if err := sheet.GlobalPipelineStore.SaveJob(job); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to save job"})
	}

	_ = sheet.GlobalPipelineStore.SaveConversation(conv)

	reviewData := ws.Review_output(
		"Design Review",
		fmt.Sprintf("```text\n%s\n```", refined),
		false,
		map[string]interface{}{
			"pipeline": map[string]interface{}{
				"jobId":   job.ID.String(),
				"step":    "design",
				"actions": []string{"approve", "refine", "regenerate"},
			},
		},
	)["data"].(map[string]interface{})

	sheet.GlobalPipelineQueue.EmitUpdate(job, "Design refined - review required", reviewData)

	return c.JSON(fiber.Map{"status": "updated"})
}

func handlePipelineLatexApprove(c *fiber.Ctx) error {
	job, _, err := getPipelineJobForUser(c)
	if err != nil {
		return err
	}

	job.CurrentStep = pipeline.StepCompile
	job.Status = pipeline.StatusPending
	job.UpdatedAt = time.Now()

	if err := sheet.GlobalPipelineStore.SaveJob(job); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to save job"})
	}

	sheet.GlobalPipelineQueue.EmitUpdate(job, "LaTeX approved, starting compilation", ws.Stage("LaTeX", "Approved", nil)["data"].(map[string]interface{}))
	_ = sheet.GlobalPipelineQueue.Enqueue(job.ID)

	return c.JSON(fiber.Map{"status": "queued", "jobId": job.ID.String()})
}

func handlePipelineLatexEdit(c *fiber.Ctx) error {
	job, _, err := getPipelineJobForUser(c)
	if err != nil {
		return err
	}

	var body struct {
		Latex string `json:"latex"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	latex := strings.TrimSpace(body.Latex)
	if latex == "" {
		return c.Status(400).JSON(fiber.Map{"error": "latex required"})
	}

	job.Latex = latex
	job.CurrentStep = pipeline.StepCompile
	job.Status = pipeline.StatusPending
	job.UpdatedAt = time.Now()

	if err := sheet.GlobalPipelineStore.SaveJob(job); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to save job"})
	}

	sheet.GlobalPipelineQueue.EmitUpdate(job, "LaTeX updated, starting compilation", ws.Stage("LaTeX", "Edited", nil)["data"].(map[string]interface{}))
	_ = sheet.GlobalPipelineQueue.Enqueue(job.ID)

	return c.JSON(fiber.Map{"status": "queued"})
}

func handlePipelineLatexFix(c *fiber.Ctx) error {
	job, _, err := getPipelineJobForUser(c)
	if err != nil {
		return err
	}

	var body struct {
		ErrorLog string `json:"errorLog"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	errorLog := strings.TrimSpace(body.ErrorLog)
	if errorLog == "" {
		return c.Status(400).JSON(fiber.Map{"error": "errorLog required"})
	}

	conv, convErr := sheet.GlobalPipelineStore.GetConversationByJobID(job.ID)
	if convErr != nil {
		conv = pipeline.NewConversation(job.ID)
		_ = sheet.GlobalPipelineStore.SaveConversation(conv)
	}

	fixed, err := pipeline.FixLatex(context.Background(), conv, job.Latex, errorLog)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fix latex"})
	}

	job.Latex = fixed
	job.Status = pipeline.StatusWaitingManual
	job.CurrentStep = pipeline.StepLatex
	job.UpdatedAt = time.Now()

	if err := sheet.GlobalPipelineStore.SaveJob(job); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to save job"})
	}

	_ = sheet.GlobalPipelineStore.SaveConversation(conv)

	reviewData := ws.Review_output(
		"LaTeX Review",
		fmt.Sprintf("```latex\n%s\n```", fixed),
		false,
		map[string]interface{}{
			"pipeline": map[string]interface{}{
				"jobId":   job.ID.String(),
				"step":    "latex",
				"actions": []string{"approve", "edit", "fix"},
			},
		},
	)["data"].(map[string]interface{})

	sheet.GlobalPipelineQueue.EmitUpdate(job, "LaTeX fixed - review required", reviewData)

	return c.JSON(fiber.Map{"status": "updated"})
}

func handlePipelineAbort(c *fiber.Ctx) error {
	job, _, err := getPipelineJobForUser(c)
	if err != nil {
		return err
	}

	job.Status = pipeline.StatusAborted
	job.UpdatedAt = time.Now()

	if err := sheet.GlobalPipelineStore.SaveJob(job); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to save job"})
	}

	sheet.GlobalPipelineQueue.EmitUpdate(job, "Job aborted", ws.Error("Job aborted", "Aborted by user", map[string]interface{}{})["data"].(map[string]interface{}))

	return c.JSON(fiber.Map{"status": "aborted"})
}

func handlePipelineRetry(c *fiber.Ctx) error {
	job, _, err := getPipelineJobForUser(c)
	if err != nil {
		return err
	}

	// Only allow retrying failed or aborted jobs
	if job.Status != pipeline.StatusError && job.Status != pipeline.StatusAborted {
		return c.Status(400).JSON(fiber.Map{"error": fmt.Sprintf("cannot retry job in state: %s", job.Status)})
	}

	// Reset the job to re-run from the beginning
	job.Status = pipeline.StatusPending
	job.CurrentStep = pipeline.StepPrompt
	job.RetryCount = 0
	job.ErrorMessage = nil
	job.ErrorLog = nil
	job.Design = ""
	job.Latex = ""
	job.PDFURL = ""
	job.CompletedAt = nil
	job.UpdatedAt = time.Now()

	if err := sheet.GlobalPipelineStore.SaveJob(job); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to save job"})
	}

	// Create a fresh conversation
	conv := pipeline.NewConversation(job.ID)
	_ = sheet.GlobalPipelineStore.SaveConversation(conv)

	sheet.GlobalPipelineQueue.EmitUpdate(job, "Job retrying from scratch", ws.Stage("Pipeline", "Retrying", nil)["data"].(map[string]interface{}))

	if err := sheet.GlobalPipelineQueue.Enqueue(job.ID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to enqueue retry"})
	}

	return c.JSON(fiber.Map{"status": "retrying", "jobId": job.ID.String()})
}

func getPipelineJobForUser(c *fiber.Ctx) (*pipeline.Job, string, error) {
	if sheet.GlobalPipelineStore == nil || sheet.GlobalPipelineQueue == nil {
		return nil, "", c.Status(500).JSON(fiber.Map{"error": "pipeline not initialized"})
	}

	username, err := getUsernameFromAuth(c)
	if err != nil {
		return nil, "", c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	jobID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return nil, "", c.Status(400).JSON(fiber.Map{"error": "invalid job id"})
	}

	job, err := sheet.GlobalPipelineStore.GetJob(jobID)
	if err != nil || job.UserID != username {
		return nil, "", c.Status(404).JSON(fiber.Map{"error": "job not found"})
	}

	return job, username, nil
}
