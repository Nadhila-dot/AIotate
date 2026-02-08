package ai

import (
	_ "bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	_ "strconv"
	"strings"
	"time"

	store "nadhi.dev/sarvar/fun/database"
	"nadhi.dev/sarvar/fun/db"
)

const (
	// Default model to use for sheet generation
	DefaultModel = "gemini-2.5-pro"

	// Gemini API endpoint
	GeminiEndpoint = "https://generativelanguage.googleapis.com/v1/models/%s:generateContent?key=%s"

	// Maximum retries for API requests
	MaxRetries = 3

	// Delay between retries
	RetryDelay = 2 * time.Second
)

// GenerationRequest represents the request structure for sheet generation
type GenerationRequest struct {
	Subject             string       `json:"subject"`
	Course              string       `json:"course"`
	Description         string       `json:"description"`
	Tags                []string     `json:"tags"`
	Curriculum          string       `json:"curriculum"`
	SpecialInstructions string       `json:"specialInstructions"`
	StyleName           string       `json:"styleName"`
	Username            string       `json:"username"`
	Mode                string       `json:"mode"`
	WebSearchQuery      string       `json:"webSearchQuery"`
	WebSearchEnabled    bool         `json:"webSearchEnabled"`
	Attachments         []Attachment `json:"attachments"`
}

// GenerationResult contains the generated content and metadata
type GenerationResult struct {
	LaTeX    string            `json:"latex"`
	Metadata map[string]string `json:"metadata"`
}

// GenerateSheet takes user input and produces LaTeX content using Gemini 2.5 Pro
// GenerateSheet takes user input and produces LaTeX content using the configured AI provider
func GenerateSheet(request *GenerationRequest) (*GenerationResult, error) {
	if request == nil {
		return nil, fmt.Errorf("generation request cannot be nil")
	}

	systemPrompt := buildSystemPrompt(request)
	userPrompt := buildUserPrompt(request)

	log.Printf("Generating sheet with configured AI provider")

	// Use the unified API with LaTeX generation task type
	response, err := GenerateWithRetry(TaskLaTeXGeneration, systemPrompt, userPrompt, MaxRetries)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Parse the response to extract LaTeX content and metadata
	latex, metadata, err := extractContent(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse generated content: %w", err)
	}

	return &GenerationResult{
		LaTeX:    latex,
		Metadata: metadata,
	}, nil
}

// defaultStylePrompt contains the visual styling section of the system prompt.
const defaultStylePrompt = `% Document styling for visual appeal
\\definecolor{primary}{RGB}{25,103,210}
\\definecolor{secondary}{RGB}{234,67,53}
\\definecolor{accent}{RGB}{251,188,4}
\\definecolor{light}{RGB}{242,242,242}

\\hypersetup{colorlinks=true,linkcolor=primary}
\\setlength{\\parindent}{0pt}
\\setlength{\\parskip}{6pt}

\\title{\\textcolor{primary}{\\Large TITLE_OF_WORKSHEET}}
\\author{\\textcolor{secondary}{Course: COURSE_NAME}}
\\date{\\today}`

// getStylePromptForRequest resolves the style prompt from the NoSQL store.
// Priority: explicit style name > user's default style > defaultStylePrompt
func getStylePromptForRequest(request *GenerationRequest) string {
	if request == nil {
		return defaultStylePrompt
	}

	username := strings.TrimSpace(request.Username)
	if username == "" {
		return defaultStylePrompt
	}

	styleName := strings.TrimSpace(request.StyleName)
	if styleName != "" {
		if style, err := store.GetStyle(db.StylesDB, username, styleName); err == nil {
			if style != nil && strings.TrimSpace(style.Prompt) != "" {
				return style.Prompt
			}
		}
	}

	if style, err := store.GetDefaultStyle(db.StylesDB, username); err == nil {
		if style != nil && strings.TrimSpace(style.Prompt) != "" {
			return style.Prompt
		}
	}

	return defaultStylePrompt
}

// ResolveStylePrompt exposes the style prompt lookup for other packages.
func ResolveStylePrompt(request *GenerationRequest) string {
	return getStylePromptForRequest(request)
}

// buildSystemPrompt creates the system prompt for the Gemini model
func buildSystemPrompt(request *GenerationRequest) string {
	stylePrompt := getStylePromptForRequest(request)
	return `You are a professional educator and LaTeX expert. Your task is to generate 
a comprehensive educational worksheet that will maximize student learning outcomes.

IMPORTANT: Your response MUST follow this exact format with NO DEVIATIONS:

<Output>
\\documentclass[12pt]{article}

% Essential packages for reliable compilation
\\usepackage[margin=1in]{geometry}
\\usepackage{amsmath,amssymb,amsthm}
\\usepackage{graphicx}
\\usepackage{enumitem}
\\usepackage{xcolor}
\\usepackage{tcolorbox}
\\usepackage{hyperref}

` + stylePrompt + `

\\begin{document}

\\maketitle

\\begin{tcolorbox}[colback=light,colframe=primary]
\\textbf{Instructions:} Clear instructions here...
\\end{tcolorbox}

% ... complete worksheet content organized in sections ...

\\end{document}
</Output>

<meta-data>
Subject: The subject of the worksheet
Level: Educational level
EstimatedTime: Estimated completion time
Keywords: comma,separated,keywords
Notes: Teacher notes
</meta-data>

STRICT REQUIREMENTS:
1. Focus EXCLUSIVELY on content quality and LaTeX correctness
2. Always use the provided LaTeX packages - DO NOT add custom package imports
3. Create syntactically perfect, compilable LaTeX with no undefined commands
4. Replace placeholder text (TITLE_OF_WORKSHEET, COURSE_NAME) appropriately
5. Organize content with clear section headings (\\section{}, \\subsection{})
6. Use \\begin{enumerate}[label=\\arabic*.] or \\begin{itemize} for lists
7. Include white space appropriately for readability

EDUCATIONAL BEST PRACTICES:
1. Start with easier questions and gradually increase difficulty
2. Include worked examples before challenging problems
3. Cover all curriculum topics specified by the user
4. Design content that targets common misconceptions
5. Use color strategically to highlight important concepts
6. Create visually distinct sections for different types of activities
7. Include "Knowledge Check" questions throughout the document`
}

// buildUserPrompt creates the user prompt based on the generation request
func buildUserPrompt(request *GenerationRequest) string {
	tagsStr := strings.Join(request.Tags, ", ")

	return fmt.Sprintf(`Please create an educational worksheet with the following specifications, make sure you do your research on this material:

Subject: %s
Course: %s
Description: %s
Tags/Keywords: %s

Curriculum Topics to Cover:
%s

Special Instructions:
%s

Remember to provide the content in the required format with both the LaTeX code and metadata.`,
		request.Subject,
		request.Course,
		request.Description,
		tagsStr,
		request.Curriculum,
		request.SpecialInstructions)
}

// geminiRequest represents the request structure for the Gemini API
type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Role  string `json:"role"`
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

// geminiResponse represents the response structure from the Gemini API
type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// generateGeminiResponse calls the Gemini API to generate content
func generateGeminiResponse(apiKey, model, systemPrompt, userPrompt string, maxRetries int) (string, error) {
	url := fmt.Sprintf(GeminiEndpoint, model, apiKey)

	// Combine system prompt and user prompt since Gemini doesn't support system role
	combinedPrompt := systemPrompt + "\n\n" + userPrompt

	// Create request with valid roles only
	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Role: "user", // Use "user" role instead of "system"
				Parts: []part{
					{Text: combinedPrompt},
				},
			},
		},
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Debug request being sent
	log.Printf("[DEBUG] Sending request to Gemini API with model: %s", model)
	log.Printf("[DEBUG] Request length: %d bytes", len(reqJSON))

	var response string
	var lastErr error

	// Implement retry logic
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retrying API request (attempt %d/%d)", attempt+1, maxRetries)
			time.Sleep(RetryDelay * time.Duration(attempt)) // Increase delay with each retry
		}

		req, err := http.NewRequest("POST", url, strings.NewReader(string(reqJSON)))
		if err != nil {
			lastErr = err
			log.Printf("[ERROR] Failed to create request: %v", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		// Set a longer timeout for the HTTP client
		// Longer timeouts aka timeout
		client := &http.Client{Timeout: 6000 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			log.Printf("[ERROR] Failed to execute request: %v", err)
			continue
		}

		// Read the response body for debugging if there's an error
		var respBody []byte
		if resp.StatusCode != http.StatusOK {
			respBody, _ = io.ReadAll(resp.Body)
			resp.Body.Close()

			log.Printf("[ERROR] API returned status %d: %s", resp.StatusCode, string(respBody))
			lastErr = fmt.Errorf("API returned non-200 status: %d - %s", resp.StatusCode, string(respBody))

			// For rate limiting errors, increase the delay significantly
			if resp.StatusCode == 429 {
				log.Printf("[WARN] Rate limited by API, increasing delay before retry")
				time.Sleep(5 * time.Second) // Add extra delay for rate limiting
			}

			continue
		}

		var geminiResp geminiResponse
		if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
			lastErr = err
			log.Printf("[ERROR] Failed to decode API response: %v", err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
			lastErr = fmt.Errorf("empty response from API")
			log.Printf("[ERROR] Empty response from API")
			continue
		}

		response = geminiResp.Candidates[0].Content.Parts[0].Text
		log.Printf("[DEBUG] Successfully received response from Gemini API (length: %d bytes)", len(response))
		return response, nil
	}

	return "", fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// extractContent parses the generated response to extract LaTeX and metadata
func extractContent(response string) (string, map[string]string, error) {
	// Extract LaTeX content between <Output> tags
	latexPattern := regexp.MustCompile(`(?s)<Output>(.*?)</Output>`)
	latexMatches := latexPattern.FindStringSubmatch(response)

	// Extract metadata between <meta-data> tags
	metadataPattern := regexp.MustCompile(`(?s)<meta-data>(.*?)</meta-data>`)
	metadataMatches := metadataPattern.FindStringSubmatch(response)

	if len(latexMatches) < 2 {
		return "", nil, fmt.Errorf("could not extract LaTeX content")
	}

	latexContent := strings.TrimSpace(latexMatches[1])

	metadata := make(map[string]string)

	// Parse metadata if available
	if len(metadataMatches) >= 2 {
		metadataText := metadataMatches[1]
		metadataLines := strings.Split(metadataText, "\n")

		for _, line := range metadataLines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				metadata[key] = value
			}
		}
	}

	return latexContent, metadata, nil
}

// GenerateSheetForQueue is used by the sheet queue system to process generation requests
// GenerateSheetForQueue is used by the sheet queue system to process generation requests
func GenerateSheetForQueue(ctx context.Context, request *GenerationRequest) (*GenerationResult, error) {
	log.Printf("[DEBUG] Starting GenerateSheetForQueue")

	if request == nil {
		log.Printf("[ERROR] Request is nil in GenerateSheetForQueue")
		return nil, fmt.Errorf("generation request cannot be nil")
	}

	// Debug request contents
	log.Printf("[DEBUG] Request - Subject: %s, Course: %s", request.Subject, request.Course)
	log.Printf("[DEBUG] Request - Description length: %d", len(request.Description))
	log.Printf("[DEBUG] Request - Tags count: %d", len(request.Tags))

	// Create a channel to receive results
	resultChan := make(chan *GenerationResult, 1)
	errChan := make(chan error, 1)

	// Run generation in a goroutine to respect context cancellation
	go func() {
		log.Printf("[DEBUG] Starting generation goroutine")
		result, err := GenerateSheet(request)
		if err != nil {
			log.Printf("[ERROR] GenerateSheet failed: %v", err)
			errChan <- err
			return
		}
		log.Printf("[DEBUG] GenerateSheet succeeded")
		resultChan <- result
	}()

	// Wait for generation or context cancellation
	log.Printf("[DEBUG] Waiting for generation result or context cancellation")
	select {
	case result := <-resultChan:
		log.Printf("[DEBUG] Received result from channel")
		return result, nil
	case err := <-errChan:
		log.Printf("[ERROR] Received error from channel: %v", err)
		return nil, err
	case <-ctx.Done():
		log.Printf("[ERROR] Context cancelled: %v", ctx.Err())
		return nil, ctx.Err()
	}
}

// SaveGeneratedContent saves the LaTeX content and metadata to separate files
func SaveGeneratedContent(outputDir, jobID string, result *GenerationResult) error {
	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Save LaTeX content to .tex file
	latexPath := fmt.Sprintf("%s/%s.tex", outputDir, jobID)
	if err := os.WriteFile(latexPath, []byte(result.LaTeX), 0644); err != nil {
		return fmt.Errorf("failed to save LaTeX content: %w", err)
	}
	log.Printf("Successfully saved .tex file to: %s", latexPath)

	// Save metadata as JSON
	metadataJSON, err := json.MarshalIndent(result.Metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	metadataPath := fmt.Sprintf("%s/%s.meta.json", outputDir, jobID)
	if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}
	log.Printf("Successfully saved metadata to: %s", metadataPath)

	return nil
}

func ProcessGeneration(ctx context.Context, jobID string, req *GenerationRequest) (map[string]interface{}, error) {
	log.Printf("[DEBUG] Starting AI generation for job ID: %s", jobID)

	// Debug the request content
	if req == nil {
		log.Printf("[ERROR] Request is nil for job ID: %s", jobID)
		return nil, fmt.Errorf("generation request cannot be nil")
	}

	// Log the request details
	reqJSON, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		log.Printf("[ERROR] Failed to marshal request for debugging: %v", err)
	} else {
		log.Printf("[DEBUG] Generation request for job %s: %s", jobID, string(reqJSON))
	}

	// Generate sheet content using the unified API
	log.Printf("[DEBUG] Calling GenerateSheetForQueue for job ID: %s", jobID)
	result, err := GenerateSheetForQueue(ctx, req)
	if err != nil {
		log.Printf("[ERROR] Generation failed for job ID %s: %v", jobID, err)
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	log.Printf("[DEBUG] Successfully generated content for job ID: %s", jobID)

	// Save generated content to files
	outputDir := fmt.Sprintf("./generated/%s", jobID)
	log.Printf("[DEBUG] Saving content to directory: %s", outputDir)

	if err := SaveGeneratedContent(outputDir, jobID, result); err != nil {
		log.Printf("[ERROR] Failed to save content for job ID %s: %v", jobID, err)
		return nil, fmt.Errorf("failed to save content: %w", err)
	}

	log.Printf("[DEBUG] Successfully processed job ID: %s", jobID)

	// Return results for the queue system, including the raw LaTeX content
	return map[string]interface{}{
		"jobID":        jobID,
		"latexPath":    fmt.Sprintf("%s/%s.tex", outputDir, jobID),
		"metaPath":     fmt.Sprintf("%s/%s.meta.json", outputDir, jobID),
		"metadata":     result.Metadata,
		"successful":   true,
		"latexContent": result.LaTeX, // Include the raw LaTeX content
		"rawResponse":  result.LaTeX, // Add the raw response for review
	}, nil
}
