package latex

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ConvertLatexToPDFWithRetry tries to convert LaTeX to PDF with AI-powered fixes
func ConvertLatexToPDFWithRetry(latexContent, texFilename, outputPath string) (string, error) {
	const maxAttempts = 3
	var conversionErr error

	// Check if content is empty
	if len(latexContent) == 0 {
		log.Printf("[ERROR] Empty LaTeX content provided to conversion function")
		return "", fmt.Errorf("empty LaTeX content, cannot proceed with conversion")
	}

	log.Printf("[DEBUG] Starting LaTeX conversion for file: %s", texFilename)
	log.Printf("[DEBUG] LaTeX content first 100 chars: %s...", truncateString(latexContent, 100))
	log.Printf("[DEBUG] Output path set to: %s", outputPath)

	// Create temp directory for conversion attempts
	fixesDir := filepath.Join("./generated/gemini_fixes")
	if err := os.MkdirAll(fixesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create fixes directory: %w", err)
	}

	// Initial attempt with original content
	pdfPath, conversionErr := convertToPDF(latexContent, texFilename, outputPath)
	if conversionErr == nil {
		return pdfPath, nil
	}

	log.Printf("[ERROR] Initial conversion failed: %v", conversionErr)

	// Save the original content for debugging
	originalFile := filepath.Join(fixesDir, strings.TrimSuffix(texFilename, ".tex")+".original.tex")
	if err := ioutil.WriteFile(originalFile, []byte(latexContent), 0644); err != nil {
		log.Printf("[WARNING] Could not save original LaTeX: %v", err)
	} else {
		log.Printf("[DEBUG] Original LaTeX saved for debugging at: %s", originalFile)
	}

	// Try with AI fixes
	currentContent := latexContent
	var errorMsg string

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		log.Printf("Gemini fix attempt %d/%d", attempt, maxAttempts)

		// Get error message from last attempt
		errorMsg = extractErrorMessage(conversionErr)

		// Request fix from AI
		fixedContent, err := FixLatexWithAI(currentContent, errorMsg)
		if err != nil {
			log.Printf("Failed to get AI fix: %v", err)
			continue
		}

		// Save this attempt for debugging
		attemptFile := filepath.Join(fixesDir, strings.TrimSuffix(texFilename, ".tex")+fmt.Sprintf(".attempt%d.tex", attempt))
		if err := ioutil.WriteFile(attemptFile, []byte(fixedContent), 0644); err != nil {
			log.Printf("Warning: Could not save attempt %d: %v", attempt, err)
		}

		// Try conversion with fixed content
		pdfPath, conversionErr = convertToPDF(fixedContent, texFilename, outputPath)
		if conversionErr == nil {
			log.Printf("Successfully fixed and converted LaTeX on attempt %d", attempt)
			return pdfPath, nil
		}

		log.Printf("Conversion still failed after fix attempt %d: %v", attempt, conversionErr)

		// Update for next attempt
		currentContent = fixedContent
	}

	// If we get here, all attempts failed
	return "", fmt.Errorf("failed to convert LaTeX to PDF after %d Gemini fix attempts: %v", maxAttempts, conversionErr)
}

// extractErrorMessage gets a clean error message from the conversion error
func extractErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	// Extract relevant part from error message
	errorMsg := err.Error()

	// Look for Tectonic output in the error
	const outputMarker = "Tectonic output:"
	if idx := strings.Index(errorMsg, outputMarker); idx >= 0 {
		errorMsg = errorMsg[idx+len(outputMarker):]
	}

	// Limit error message length for API calls
	if len(errorMsg) > 2000 {
		errorMsg = errorMsg[:2000] + "..."
	}

	return errorMsg
}

// Helper function to truncate string for logging
func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func convertToPDF(latexContent, texFilename, outputPath string) (string, error) {
	// Check if LaTeX content is empty before proceeding
	latexContent = strings.TrimSpace(latexContent)
	if latexContent == "" {
		return "", fmt.Errorf("empty LaTeX content provided, cannot proceed with conversion")
	}

	// Check if content is too small to be valid LaTeX
	if len(latexContent) < 50 {
		return "", fmt.Errorf("LaTeX content too small (%d bytes), likely invalid", len(latexContent))
	}

	// Add basic LaTeX validation
	if !strings.Contains(latexContent, "\\documentclass") ||
		!strings.Contains(latexContent, "\\begin{document}") {
		return "", fmt.Errorf("invalid LaTeX content: missing required elements")
	}

	// Log the first and last 100 characters of the content for debugging
	contentPreview := latexContent
	if len(contentPreview) > 200 {
		contentPreview = contentPreview[:100] + "..." + contentPreview[len(contentPreview)-100:]
	}
	log.Printf("[DEBUG] LaTeX content preview: %s", contentPreview)

	// Create temporary directory for conversion
	tempDir, err := ioutil.TempDir("", "latex-conversion")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Log directory and file details
	log.Printf("[DEBUG] LaTeX conversion temp directory: %s", tempDir)
	log.Printf("[DEBUG] LaTeX content length: %d bytes", len(latexContent))

	// Write LaTeX content to temporary file
	tempTexPath := filepath.Join(tempDir, texFilename)
	if err := ioutil.WriteFile(tempTexPath, []byte(latexContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write LaTeX content: %w", err)
	}

	// Verify the file was written correctly and has content
	if fileInfo, err := os.Stat(tempTexPath); err != nil {
		return "", fmt.Errorf("failed to verify LaTeX file was written: %w", err)
	} else if fileInfo.Size() == 0 {
		return "", fmt.Errorf("LaTeX file was created but is empty")
	} else {
		log.Printf("[DEBUG] LaTeX file written successfully: %s (size: %d bytes)", tempTexPath, fileInfo.Size())
	}

	// Get the filename without extension for PDF output
	fileBase := strings.TrimSuffix(texFilename, filepath.Ext(texFilename))
	tempPDFPath := filepath.Join(tempDir, fileBase+".pdf")

	log.Printf("[DEBUG] Expected PDF output path: %s", tempPDFPath)

	// Run Tectonic command with detailed output capture
	cmd := exec.Command("tectonic", "--outfmt=pdf", "--keep-logs", "-o", tempDir, tempTexPath)

	// Set working directory to temp directory to ensure relative paths work
	cmd.Dir = tempDir
	log.Printf("[DEBUG] Running Tectonic in directory: %s", cmd.Dir)
	log.Printf("[DEBUG] Tectonic command: %v", cmd.Args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Save the error output for debugging
		errorLogPath := filepath.Join("./generated/error_logs", fileBase+".log")
		os.MkdirAll(filepath.Dir(errorLogPath), 0755)
		ioutil.WriteFile(errorLogPath, output, 0644)

		return "", fmt.Errorf("%w\nTectonic output:\n%s", err, string(output))
	}

	// Check if PDF was created
	if _, err := os.Stat(tempPDFPath); os.IsNotExist(err) {
		return "", fmt.Errorf("tectonic completed but PDF file not found\nOutput: %s", string(output))
	}

	log.Printf("[DEBUG] PDF file created successfully at: %s", tempPDFPath)

	// Make sure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Copy the generated PDF to the final output location
	outputData, err := ioutil.ReadFile(tempPDFPath)
	if err != nil {
		return "", fmt.Errorf("failed to read generated PDF: %w", err)
	}

	if err := ioutil.WriteFile(outputPath, outputData, 0644); err != nil {
		return "", fmt.Errorf("failed to copy PDF to final location: %w", err)
	}

	log.Printf("[DEBUG] PDF file copied to final location: %s", outputPath)

	return outputPath, nil
}
