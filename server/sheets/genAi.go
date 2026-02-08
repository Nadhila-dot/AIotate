package sheet

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"nadhi.dev/sarvar/fun/ai"
	"nadhi.dev/sarvar/fun/latex"
	logg "nadhi.dev/sarvar/fun/logs"
	websocket "nadhi.dev/sarvar/fun/websocket"
)

func (sq *SheetQueue) generateWithAI(job *QueuedJob) (interface{}, error) {
	//sq.logger.Printf("Starting AI generation for job %s", job.ID)
	logg.Info(fmt.Sprintf("Starting AI generation for job %s", job.ID))

	var request ai.GenerationRequest
	if err := json.Unmarshal([]byte(job.Prompt), &request); err != nil {
		return nil, fmt.Errorf("failed to parse job request: %w", err)
	}

	// 1. AI Generation Started
	sq.statusUpdates <- StatusUpdate{
		ID:     job.ID,
		Status: "processing",
		Data:   websocket.Stage("AI", "AI generation started", nil)["data"].(map[string]interface{}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6000000*time.Second)
	defer cancel()

	// Validate AI configuration
	if err := ai.ValidateAIConfig(); err != nil {
		sq.statusUpdates <- StatusUpdate{
			ID:     job.ID,
			Status: "failed",
			Data:   websocket.Review_output("AI Configuration Error", fmt.Sprintf("# AI Configuration Error\n\n%s\n\nPlease check your set.json configuration.\n\n## Gemini Setup\nSet `AI_PROVIDER` to `gemini` and add your `GEMINI_API_KEY`.\nGet a key from: https://aistudio.google.com/app/apikey\n\n## OpenRouter Setup\nSet `AI_PROVIDER` to `openrouter` and add your `OPENROUTER_API_KEY`.\nGet a key from: https://openrouter.ai/keys", err.Error()), true, map[string]interface{}{}),
		}
		sq.statusUpdates <- StatusUpdate{
			ID:     job.ID,
			Status: "failed",
			Data:   websocket.End("AI not configured", map[string]interface{}{})["data"].(map[string]interface{}),
		}
		return nil, fmt.Errorf("AI configuration error: %w", err)
	}

	// 2. AI Generation
	result, err := ai.ProcessGeneration(ctx, job.ID, &request)
	if err != nil {
		sq.statusUpdates <- StatusUpdate{
			ID:     job.ID,
			Status: "processing",
			Data:   websocket.Retry("AI generation failed..", map[string]interface{}{}),
		}
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}
	sq.statusUpdates <- StatusUpdate{
		ID:     job.ID,
		Status: "processing",
		Data:   websocket.Stage("AI", "AI generation completed, parsing LaTeX", nil)["data"].(map[string]interface{}),
	}

	// 3. Parse LaTeX
	// 3. Parse LaTeX
	rawContent := fmt.Sprintf("%v", result)
	latexContent, metadata, err := latex.SplitContent(rawContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LaTeX: %w", err)
	}

	// Get the raw LaTeX content from the result map if available
	var rawLatex string
	if rawLatexValue, ok := result["latexContent"].(string); ok {
		rawLatex = rawLatexValue
	} else {
		rawLatex = latexContent // Fallback to parsed content
	}

	// Remove any markdown code block markers if they somehow got included
	rawLatex = strings.TrimPrefix(rawLatex, "```latex\n")
	rawLatex = strings.TrimPrefix(rawLatex, "```latex")
	rawLatex = strings.TrimPrefix(rawLatex, "```\n")
	rawLatex = strings.TrimPrefix(rawLatex, "```")
	rawLatex = strings.TrimSuffix(rawLatex, "\n```")
	rawLatex = strings.TrimSuffix(rawLatex, "```")
	rawLatex = strings.TrimSpace(rawLatex)

	// Send LaTeX to client for review (as Markdown in modal)
	sq.statusUpdates <- StatusUpdate{
		ID:     job.ID,
		Status: "processing",
		Data: websocket.Review_output(
			"Review Generated LaTeX - Please check for any errors",
			fmt.Sprintf("```latex\n%s\n```", rawLatex), // Wrap for display only
			false,
			map[string]interface{}{
				"metadata": metadata,
			},
		)["data"].(map[string]interface{}),
	}

	sq.statusUpdates <- StatusUpdate{
		ID:     job.ID,
		Status: "processing",
		Data:   websocket.Stage("LaTeX", "LaTeX parsed, converting to PDF", nil)["data"].(map[string]interface{}),
	}

	texFilename := fmt.Sprintf("%s.tex", job.ID)
	pdfFilename := fmt.Sprintf("%s.pdf", job.ID)
	//generatedTexPath := filepath.Join("./generated", job.ID, texFilename)
	bucketPath := filepath.Join("./storage/bucket", pdfFilename)

	sq.statusUpdates <- StatusUpdate{
		ID:     job.ID,
		Status: "processing",
		Data:   websocket.Stage("LaTeX", "Converting LaTeX to PDF...", nil)["data"].(map[string]interface{}),
	}

	// Show that we're trying to fix LaTeX if needed
	sq.statusUpdates <- StatusUpdate{
		ID:     job.ID,
		Status: "processing",
		Data:   websocket.Stage("LaTeX", "Starting PDF conversion, may attempt AI fixes if needed", nil)["data"].(map[string]interface{}),
	}

	// Now convert with the cleaned content
	if _, err := latex.ConvertLatexToPDFWithRetry(rawLatex,
		texFilename, bucketPath); err != nil {
		errorDetails := fmt.Sprintf("PDF conversion failed: %v", err)

		// Log the detailed error
		logg.Error(errorDetails)

		sq.statusUpdates <- StatusUpdate{
			ID:     job.ID,
			Status: "failed",
			Data: websocket.Retry("PDF conversion failed after multiple attempts. Please try again or simplify your request.", map[string]interface{}{
				"details": errorDetails,
			})["data"].(map[string]interface{}),
		}
	}

	// Final check for PDF conversion error

	sq.statusUpdates <- StatusUpdate{
		ID:     job.ID,
		Status: "processing",
		Data:   websocket.Stage("PDF", "PDF conversion done, saving to bucket", nil)["data"].(map[string]interface{}),
	}

	// 5. Construct URL
	// add an adiditonal /bucket cuz of the way the server serves static files
	url := fmt.Sprintf("/vela/bucket/bucket/%s", pdfFilename)
	sq.statusUpdates <- StatusUpdate{
		ID:     job.ID,
		Status: "completed",
		Result: map[string]interface{}{
			"pdf_url":  url,
			"metadata": metadata,
		},
		Data: websocket.Review_output("Completed Generation", fmt.Sprintf("# Sheet generation completed: %s\n- Generated by Gemini 2.5 Pro", url), true, map[string]interface{}{
			"generatedWith": "Gemini 2.5 Pro",
		})["data"].(map[string]interface{}),
	}
	sq.statusUpdates <- StatusUpdate{
		ID:     job.ID,
		Status: "completed",
		Result: map[string]interface{}{
			"pdf_url":  url,
			"metadata": metadata,
		},
		Data: websocket.Completed("Sheet generation completed", url, map[string]interface{}{
			"generatedWith": "Gemini 2.5 Pro",
		})["data"].(map[string]interface{}),
	}

	return map[string]interface{}{
		"pdf_url":  url,
		"metadata": metadata,
	}, nil
}
