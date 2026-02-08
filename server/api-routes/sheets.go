package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"nadhi.dev/sarvar/fun/ai"
	"nadhi.dev/sarvar/fun/auth"
	vela "nadhi.dev/sarvar/fun/bucket"
	"nadhi.dev/sarvar/fun/pipeline"
	"nadhi.dev/sarvar/fun/server"
	sheet "nadhi.dev/sarvar/fun/sheets"
)

// Sheet represents the sheet data structure
type Sheet struct {
	Subject     string   `json:"subject"`
	Course      string   `json:"course"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Visibility  string   `json:"visibility"`
}

const maxUploadBytes = 20 * 1024 * 1024

func parseCreateSheetMultipart(c *fiber.Ctx, req *struct {
	Subject             string          `json:"subject"`
	Course              string          `json:"course"`
	Description         string          `json:"description"`
	Tags                string          `json:"tags"`
	Curriculum          string          `json:"curriculum"`
	SpecialInstructions string          `json:"specialInstructions"`
	Visibility          string          `json:"visibility"`
	StyleName           string          `json:"styleName"`
	Mode                string          `json:"mode"`
	WebSearchQuery      string          `json:"webSearchQuery"`
	WebSearchEnabled    bool            `json:"webSearchEnabled"`
	Attachments         []ai.Attachment `json:"attachments"`
}) error {
	form, err := c.MultipartForm()
	if err != nil {
		return fmt.Errorf("invalid multipart form")
	}

	getValue := func(key string) string {
		if vals, ok := form.Value[key]; ok && len(vals) > 0 {
			return vals[0]
		}
		return ""
	}

	req.Subject = getValue("subject")
	req.Course = getValue("course")
	req.Description = getValue("description")
	req.Tags = getValue("tags")
	req.Curriculum = getValue("curriculum")
	req.SpecialInstructions = getValue("specialInstructions")
	req.Visibility = getValue("visibility")
	req.StyleName = getValue("styleName")
	req.Mode = getValue("mode")
	req.WebSearchQuery = getValue("webSearchQuery")
	req.WebSearchEnabled = strings.ToLower(getValue("webSearchEnabled")) == "true"

	files := []*multipart.FileHeader{}
	if fileList, ok := form.File["files"]; ok {
		files = append(files, fileList...)
	}
	if fileList, ok := form.File["attachments"]; ok {
		files = append(files, fileList...)
	}

	if len(files) == 0 {
		return nil
	}

	var total int64
	for _, fh := range files {
		total += fh.Size
	}
	if total > maxUploadBytes {
		return fmt.Errorf("upload exceeds 20MB limit")
	}

	attachments, err := parseAttachments(files)
	if err != nil {
		return err
	}
	req.Attachments = attachments

	return nil
}

func parseAttachments(files []*multipart.FileHeader) ([]ai.Attachment, error) {
	attachments := make([]ai.Attachment, 0, len(files))
	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %s", fh.Filename)
		}
		data, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %s", fh.Filename)
		}

		mimeType := fh.Header.Get("Content-Type")
		if mimeType == "" {
			mimeType = http.DetectContentType(data)
		}
		if ext := strings.TrimSpace(fh.Filename); ext != "" {
			if m := mime.TypeByExtension("." + strings.Split(ext, ".")[len(strings.Split(ext, "."))-1]); m != "" {
				mimeType = m
			}
		}

		content := ""
		encoding := "base64"
		if utf8.Valid(data) {
			content = string(data)
			encoding = "utf-8"
		} else {
			content = base64.StdEncoding.EncodeToString(data)
		}

		attachments = append(attachments, ai.Attachment{
			Name:     fh.Filename,
			MimeType: mimeType,
			Size:     fh.Size,
			Content:  content,
			Encoding: encoding,
		})
	}

	return attachments, nil
}

// Last request timestamps for cooldown
var lastRequestTimes = make(map[string]time.Time)

// SheetsIndex registers all sheet related routes
func SheetsIndex() error {
	if sheet.GlobalPipelineQueue == nil || sheet.GlobalPipelineStore == nil {
		log.Printf("Warning: Pipeline not initialized; falling back to legacy queue")
		if sheet.GlobalSheetGenerator == nil {
			var err error
			sheet.GlobalSheetGenerator, err = sheet.NewSheetGenerator(nil, "./queue_data", 2)
			if err != nil {
				log.Printf("Failed to initialize GlobalSheetGenerator: %v", err)
				return err
			}
			log.Printf("GlobalSheetGenerator initialized successfully")
		}
	}

	server.Route.Post("/api/v1/sheets/generate-tags", generateTags)
	server.Route.Post("/api/v1/sheets/generate-subject", generateSubject)
	server.Route.Post("/api/v1/sheets/generate-course", generateCourse)
	server.Route.Post("/api/v1/sheets/generate-description", generateDescription)
	server.Route.Post("/api/v1/sheets/queue/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			return c.Status(400).JSON(fiber.Map{"error": "missing job id"})
		}

		if sheet.GlobalPipelineStore != nil {
			if jobID, err := parsePipelineJobID(id); err == nil {
				if err := sheet.GlobalPipelineStore.DeleteJob(jobID); err == nil {
					return c.JSON(fiber.Map{"status": "deleted"})
				}
			}
		}

		if sheet.GlobalSheetGenerator == nil || sheet.GlobalSheetGenerator.Queue == nil {
			return c.Status(500).JSON(fiber.Map{"error": "Sheet queue not initialized"})
		}
		err := sheet.GlobalSheetGenerator.Queue.DeleteJob(id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "deleted"})
	})

	server.Route.Get("/api/v1/sheets/get", func(c *fiber.Ctx) error {
		// Query params
		search := c.Query("search", "")
		latest := c.Query("latest", "true") == "true"
		objNumStr := c.Query("obj_num", "10")
		objNum, err := strconv.Atoi(objNumStr)
		if err != nil || objNum <= 0 {
			objNum = 10
		}

		if sheet.GlobalPipelineStore != nil {
			items, err := getPipelineQueueItems(search, latest, objNum)
			if err == nil {
				return c.JSON(items)
			}
			return c.Status(500).JSON(fiber.Map{"error": "Failed to read pipeline jobs"})
		}

		queuePath := "./queue_data/queue.json"
		items, err := vela.GetQueueItems(queuePath, latest, objNum, search)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get queue items"})
		}
		return c.JSON(items)
	})

	server.Route.Post("/api/v1/sheets/create", func(c *fiber.Ctx) error {
		var req struct {
			Subject             string          `json:"subject"`
			Course              string          `json:"course"`
			Description         string          `json:"description"`
			Tags                string          `json:"tags"`
			Curriculum          string          `json:"curriculum"`
			SpecialInstructions string          `json:"specialInstructions"`
			Visibility          string          `json:"visibility"`
			StyleName           string          `json:"styleName"`
			Mode                string          `json:"mode"`
			WebSearchQuery      string          `json:"webSearchQuery"`
			WebSearchEnabled    bool            `json:"webSearchEnabled"`
			Attachments         []ai.Attachment `json:"attachments"`
		}
		contentType := c.Get("Content-Type")
		if strings.HasPrefix(contentType, "multipart/form-data") {
			if err := parseCreateSheetMultipart(c, &req); err != nil {
				return c.Status(400).JSON(fiber.Map{"error": err.Error()})
			}
		} else {
			if err := c.BodyParser(&req); err != nil {
				return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
			}
		}

		// Validate required fields
		if req.Subject == "" || req.Course == "" || req.Description == "" || req.Tags == "" || req.Curriculum == "" || req.SpecialInstructions == "" || req.Visibility == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request: missing required fields"})
		}

		// Extract and validate session
		authHeader := c.Get("Authorization")
		if len(authHeader) < 8 || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(401).JSON(fiber.Map{"error": "missing or invalid authorization header"})
		}
		sessionID := authHeader[7:]
		valid, err := auth.IsSessionValid(sessionID)
		if err != nil || !valid {
			return c.Status(401).JSON(fiber.Map{"error": "invalid session"})
		}

		// Get user from session
		user, err := auth.GetUserBySession(sessionID)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "user not found or session invalid"})
		}
		userID := user.Username

		// Create a proper GenerationRequest
		genRequest := &ai.GenerationRequest{
			Subject:             req.Subject,
			Course:              req.Course,
			Description:         req.Description,
			Tags:                strings.Split(req.Tags, ","), // Convert comma-separated string to slice
			Curriculum:          req.Curriculum,
			SpecialInstructions: req.SpecialInstructions,
			StyleName:           req.StyleName,
			Username:            userID,
			Mode:                req.Mode,
			WebSearchQuery:      req.WebSearchQuery,
			WebSearchEnabled:    req.WebSearchEnabled,
			Attachments:         req.Attachments,
		}

		requestJSON, err := json.Marshal(genRequest)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to build request"})
		}

		if sheet.GlobalPipelineStore != nil && sheet.GlobalPipelineQueue != nil {
			job := pipeline.NewJob(userID, string(requestJSON), 3)
			job.Metadata["request"] = genRequest
			if err := sheet.GlobalPipelineStore.SaveJob(job); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to save job"})
			}
			conv := pipeline.NewConversation(job.ID)
			_ = sheet.GlobalPipelineStore.SaveConversation(conv)
			if err := sheet.GlobalPipelineQueue.Enqueue(job.ID); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to enqueue sheet"})
			}
			return c.JSON(fiber.Map{"jobId": job.ID.String(), "status": "queued"})
		}

		// Fallback to legacy queue
		if sheet.GlobalSheetGenerator == nil {
			return c.Status(500).JSON(fiber.Map{"error": "Sheet generator not initialized"})
		}

		jobID, err := sheet.GlobalSheetGenerator.CreateSheet(userID, genRequest)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to enqueue sheet"})
		}

		return c.JSON(fiber.Map{"jobId": jobID, "status": "queued"})
	})

	server.Route.Get("/api/v1/sheets/queue", func(c *fiber.Ctx) error {
		userID := c.Locals("username")
		if userID == nil {
			userID = "anonymous"
		}

		if sheet.GlobalPipelineStore != nil {
			jobs, err := sheet.GlobalPipelineStore.GetJobsByUser(userID.(string))
			if err == nil {
				return c.JSON(jobs)
			}
		}

		if sheet.GlobalSheetGenerator == nil {
			return c.Status(500).JSON(fiber.Map{"error": "Sheet generator not initialized"})
		}
		jobs, err := sheet.GlobalSheetGenerator.GetUserJobs(userID.(string))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get queue"})
		}
		return c.JSON(jobs)
	})

	return nil
}

func parsePipelineJobID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

func mapPipelineStatus(status pipeline.JobStatus) string {
	switch status {
	case pipeline.StatusCompleted:
		return "completed"
	case pipeline.StatusError, pipeline.StatusAborted:
		return "error"
	case pipeline.StatusRunning, pipeline.StatusPending, pipeline.StatusWaitingManual, pipeline.StatusWaitingAIFix:
		return "processing"
	default:
		return "processing"
	}
}

func getPipelineQueueItems(search string, latest bool, limit int) ([]map[string]interface{}, error) {
	jobs, err := sheet.GlobalPipelineStore.GetAllJobs()
	if err != nil {
		return nil, err
	}

	searchLower := strings.ToLower(strings.TrimSpace(search))
	items := make([]map[string]interface{}, 0, len(jobs))

	for _, job := range jobs {
		if searchLower != "" && !strings.Contains(strings.ToLower(job.Prompt), searchLower) {
			continue
		}

		result := interface{}(nil)
		if job.Status == pipeline.StatusCompleted {
			metadata := map[string]interface{}{}
			if job.Metadata != nil {
				if md, ok := job.Metadata["metadata"].(map[string]interface{}); ok {
					metadata = md
				}
			}
			result = map[string]interface{}{
				"pdf_url":  job.PDFURL,
				"metadata": metadata,
			}
		}

		items = append(items, map[string]interface{}{
			"id":         job.ID.String(),
			"status":     mapPipelineStatus(job.Status),
			"prompt":     job.Prompt,
			"created_at": job.CreatedAt,
			"updated_at": job.UpdatedAt,
			"result":     result,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		ti, okI := items[i]["updated_at"].(time.Time)
		tj, okJ := items[j]["updated_at"].(time.Time)
		if !okI || !okJ {
			return false
		}
		if latest {
			return ti.After(tj)
		}
		return ti.Before(tj)
	})

	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	return items, nil
}

// getCooldown returns the cooldown time in seconds
func getCooldown() int {
	return 2
}

// checkCooldown checks if the cooldown period has passed for a given endpoint
func checkCooldown(endpoint string) bool {
	cooldown := getCooldown()
	lastTime, exists := lastRequestTimes[endpoint]
	if !exists {
		lastRequestTimes[endpoint] = time.Now()
		return true
	}

	if time.Since(lastTime).Seconds() < float64(cooldown) {
		return false
	}

	lastRequestTimes[endpoint] = time.Now()
	return true
}

// extractTags extracts tags from a response
func extractTags(response string) ([]string, error) {
	var tags []string

	err := json.Unmarshal([]byte(response), &tags)
	if err != nil {
		// Try to extract JSON array from text
		startIdx := strings.Index(response, "[")
		endIdx := strings.LastIndex(response, "]")
		if startIdx >= 0 && endIdx > startIdx {
			jsonStr := response[startIdx : endIdx+1]
			err = json.Unmarshal([]byte(jsonStr), &tags)
			if err != nil {
				// As a fallback, split by commas and clean up
				cleanResponse := strings.Trim(response, "[]\" \n")
				tags = strings.Split(cleanResponse, ",")
				for i, tag := range tags {
					tags[i] = strings.Trim(tag, "\" ")
				}
			}
		}
	}

	return tags, nil
}

// generateTags handles requests to generate tags using AI
func generateTags(c *fiber.Ctx) error {
	// Check cooldown
	if !checkCooldown("tags") {
		return c.Status(429).JSON(fiber.Map{"error": "Too many requests, please wait"})
	}

	var sheet Sheet
	if err := c.BodyParser(&sheet); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request data"})
	}

	if sheet.Subject == "" && sheet.Course == "" && sheet.Description == "" {
		return c.Status(400).JSON(fiber.Map{"error": "At least one of subject, course, or description is required"})
	}

	systemPrompt := `You are a tag generator for educational content. 
Your task is to generate 3-7 relevant tags based on the subject, course title, and description provided.
Return ONLY a JSON array of strings with the tags, nothing else.
Example response: ["mathematics", "algebra", "equations", "polynomials"]`

	userPrompt := fmt.Sprintf(`Generate tags for the following educational content:
Subject: %s
Course: %s
Description: %s`,
		sheet.Subject,
		sheet.Course,
		sheet.Description)

	// Use the new unified AI system
	response, err := ai.GenerateSimple(ai.TaskUtility, systemPrompt, userPrompt)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to generate tags: %v", err)})
	}

	tags, _ := extractTags(response)
	return c.Status(200).JSON(fiber.Map{"tags": tags})
}

// generateSubject generates a subject based on course and/or description
func generateSubject(c *fiber.Ctx) error {
	// Check cooldown
	if !checkCooldown("subject") {
		return c.Status(429).JSON(fiber.Map{"error": "Too many requests, please wait"})
	}

	var request struct {
		Course       string `json:"course"`
		Description  string `json:"description"`
		GenerateTags bool   `json:"generateTags"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request data"})
	}

	if request.Course == "" && request.Description == "" {
		return c.Status(400).JSON(fiber.Map{"error": "At least course or description is required"})
	}

	systemPrompt := `You are an educational content creator. 
Based on the course title and description provided, generate an appropriate subject field.
Return ONLY the subject name, nothing else. Keep it concise (1-3 words).`

	userPrompt := fmt.Sprintf(`Generate a subject name for the following course:
Course: %s
Description: %s`,
		request.Course,
		request.Description)

	// Use the new unified AI system
	response, err := ai.GenerateSimple(ai.TaskUtility, systemPrompt, userPrompt)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to generate subject: %v", err)})
	}

	// Clean the response
	subject := strings.Trim(response, " \n\"")

	result := fiber.Map{"subject": subject}

	// Generate tags only if requested AND the tags query param is set to true
	if request.GenerateTags && c.Query("tags") == "true" {
		if checkCooldown("subject_tags") {
			tagSystemPrompt := `Generate 3-5 tags for this academic subject. Return only a JSON array of strings.
Example: ["physics", "mechanics", "motion"]`

			tagUserPrompt := fmt.Sprintf("Subject: %s\nCourse: %s\nDescription: %s",
				subject, request.Course, request.Description)

			tagResponse, err := ai.GenerateSimple(ai.TaskUtility, tagSystemPrompt, tagUserPrompt)
			if err == nil {
				tags, _ := extractTags(tagResponse)
				result["tags"] = tags
			}
		}
	}

	return c.Status(200).JSON(result)
}

// generateCourse generates a course title based on subject and/or description
func generateCourse(c *fiber.Ctx) error {
	// Check cooldown
	if !checkCooldown("course") {
		return c.Status(429).JSON(fiber.Map{"error": "Too many requests, please wait"})
	}

	var request struct {
		Subject      string `json:"subject"`
		Description  string `json:"description"`
		GenerateTags bool   `json:"generateTags"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request data"})
	}

	if request.Subject == "" && request.Description == "" {
		return c.Status(400).JSON(fiber.Map{"error": "At least subject or description is required"})
	}

	systemPrompt := `You are an educational content creator. 
Based on the subject and description provided, generate an appropriate course title.
Return ONLY the course title, nothing else. Make it sound like an actual academic course.`

	userPrompt := fmt.Sprintf(`Generate a course title for the following:
Subject: %s
Description: %s`,
		request.Subject,
		request.Description)

	// Use the new unified AI system
	response, err := ai.GenerateSimple(ai.TaskUtility, systemPrompt, userPrompt)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to generate course: %v", err)})
	}

	// Clean the response
	course := strings.Trim(response, " \n\"")

	result := fiber.Map{"course": course}

	// Generate tags only if requested AND the tags query param is set to true
	if request.GenerateTags && c.Query("tags") == "true" {
		if checkCooldown("course_tags") {
			tagSystemPrompt := `Generate 3-5 tags for this academic course. Return only a JSON array of strings.
Example: ["calculus", "mathematics", "derivatives"]`

			tagUserPrompt := fmt.Sprintf("Subject: %s\nCourse: %s\nDescription: %s",
				request.Subject, course, request.Description)

			tagResponse, err := ai.GenerateSimple(ai.TaskUtility, tagSystemPrompt, tagUserPrompt)
			if err == nil {
				tags, _ := extractTags(tagResponse)
				result["tags"] = tags
			}
		}
	}

	return c.Status(200).JSON(result)
}

// generateDescription generates a description based on subject and/or course
func generateDescription(c *fiber.Ctx) error {
	// Check cooldown
	if !checkCooldown("description") {
		return c.Status(429).JSON(fiber.Map{"error": "Too many requests, please wait"})
	}

	var request struct {
		Subject      string `json:"subject"`
		Course       string `json:"course"`
		GenerateTags bool   `json:"generateTags"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request data"})
	}

	if request.Subject == "" && request.Course == "" {
		return c.Status(400).JSON(fiber.Map{"error": "At least subject or course is required"})
	}

	systemPrompt := `You are an educational content creator. 
Based on the subject and course title provided, generate an appropriate description.
The description should be 2-3 sentences that explain what the course covers.`

	userPrompt := fmt.Sprintf(`Generate a description for the following course:
Subject: %s
Course: %s
Make an apporiate description with the instructions on how to prepare for the course and create an exam course.`,
		request.Subject,
		request.Course)

	// Use the new unified AI system
	response, err := ai.GenerateSimple(ai.TaskUtility, systemPrompt, userPrompt)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to generate description: %v", err)})
	}

	// Clean the response
	description := strings.Trim(response, " \n\"")

	result := fiber.Map{"description": description}

	// Generate tags only if requested AND the tags query param is set to true
	if request.GenerateTags && c.Query("tags") == "true" {
		if checkCooldown("description_tags") {
			tagSystemPrompt := `Generate 3-5 tags for this course description. Return only a JSON array of strings.
Example: ["chemistry", "organic", "synthesis"]`

			tagUserPrompt := fmt.Sprintf("Subject: %s\nCourse: %s\nDescription: %s",
				request.Subject, request.Course, description)

			tagResponse, err := ai.GenerateSimple(ai.TaskUtility, tagSystemPrompt, tagUserPrompt)
			if err == nil {
				tags, _ := extractTags(tagResponse)
				result["tags"] = tags
			}
		}
	}

	return c.Status(200).JSON(result)
}
