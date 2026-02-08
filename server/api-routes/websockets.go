package api

import (
	"crypto/md5"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"nadhi.dev/sarvar/fun/auth"
	"nadhi.dev/sarvar/fun/pipeline"
	"nadhi.dev/sarvar/fun/server"

	sheet "nadhi.dev/sarvar/fun/sheets"
	ws "nadhi.dev/sarvar/fun/websocket"
)

func RegisterWebsocketRoutes() {
	// Middleware to check if connection is websocket
	server.Route.Use("/api/v1/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// Websocket connection endpoint
	server.Route.Get("/api/v1/ws/notifications", websocket.New(ws.GetManager().ConnectHandler))

	// Websocket status endpoint (HTTP)
	server.Route.Get("/api/v1/ws/status", func(c *fiber.Ctx) error {
		return c.JSON(ws.GetManager().Status())
	})

	server.Route.Get("/api/v1/ws/job/:jobid", websocket.New(func(c *websocket.Conn) {
		jobID := c.Params("jobid")
		sessionID := c.Query("session")

		// Validate session
		isValid, err := auth.IsSessionValid(sessionID)
		if err != nil || !isValid {
			_ = c.WriteJSON(ws.Error(
				"Invalid session",
				"Authentication failed",
				map[string]interface{}{},
			))
			c.Close()
			return
		}

		// Prefer pipeline websocket if available
		if sheet.GlobalPipelineQueue != nil && sheet.GlobalPipelineStore != nil {
			if jobUUID, parseErr := uuid.Parse(jobID); parseErr == nil {
				registerPipelineJobListener(c, jobUUID)
				return
			}
		}

		// Fallback to legacy queue
		if sheet.GlobalSheetGenerator == nil || sheet.GlobalSheetGenerator.Queue == nil {
			_ = c.WriteJSON(ws.Error(
				"Server error",
				"Sheet generator not initialized",
				map[string]interface{}{},
			))
			c.Close()
			return
		}

		lastSent := make(map[string]string)
		sheet.GlobalSheetGenerator.Queue.RegisterJobListener(jobID, func(update sheet.StatusUpdate) {
			hashInput := fmt.Sprintf("%s|%v|%v", update.Status, update.Result, update.Data)
			hash := fmt.Sprintf("%x", md5.Sum([]byte(hashInput)))
			if lastSent[jobID] == hash {
				return
			}
			lastSent[jobID] = hash

			msg := ws.Push(update.Status, map[string]interface{}{
				"message": fmt.Sprintf("Job %s status: %s", update.ID, update.Status),
			})
			msg["jobId"] = update.ID
			if update.Result != nil {
				msg["result"] = update.Result
			}
			if update.Data != nil {
				msg["data"] = update.Data
			}
			_ = c.WriteJSON(msg)
		})

		if job, exists := sheet.GlobalSheetGenerator.Queue.GetJobStatus(jobID); exists {
			message := fmt.Sprintf("Initial status for job %s: %s", jobID, job.Status)
			msg := ws.Start(message, map[string]interface{}{})
			msg["jobId"] = jobID
			_ = c.WriteJSON(msg)
		}

		for {
			if _, _, err := c.ReadMessage(); err != nil {
				break
			}
		}
	}))
}

func registerPipelineJobListener(c *websocket.Conn, jobID uuid.UUID) {
	lastSent := make(map[string]string)

	sheet.GlobalPipelineQueue.RegisterJobListener(jobID, func(update pipeline.StatusUpdate) {
		hashInput := fmt.Sprintf("%s|%s|%v", update.Status, update.Message, update.Data)
		hash := fmt.Sprintf("%x", md5.Sum([]byte(hashInput)))
		if lastSent[jobID.String()] == hash {
			return
		}
		lastSent[jobID.String()] = hash

		payload := map[string]interface{}{}
		if update.Data != nil {
			payload = update.Data
		}
		if _, ok := payload["type"]; !ok {
			payload["type"] = "processing"
			payload["message"] = update.Message
			payload["step"] = string(update.Step)
		}

		msg := map[string]interface{}{
			"jobId": jobID.String(),
			"data":  payload,
		}
		_ = c.WriteJSON(msg)
	})

	if job, err := sheet.GlobalPipelineStore.GetJob(jobID); err == nil {
		payload := map[string]interface{}{
			"type":    "stage",
			"stage":   "Pipeline",
			"step":    fmt.Sprintf("Status: %s", job.Status),
			"message": fmt.Sprintf("Job %s is %s", job.ID.String(), job.Status),
		}
		if job.Status == pipeline.StatusCompleted {
			metadata := map[string]interface{}{}
			if job.Metadata != nil {
				if md, ok := job.Metadata["metadata"].(map[string]interface{}); ok {
					metadata = md
				}
			}
			payload = ws.Completed("Sheet generation completed", map[string]interface{}{
				"pdf_url":  job.PDFURL,
				"metadata": metadata,
			}, map[string]interface{}{})["data"].(map[string]interface{})
		}

		_ = c.WriteJSON(map[string]interface{}{
			"jobId": job.ID.String(),
			"data":  payload,
		})
	}

	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
}
