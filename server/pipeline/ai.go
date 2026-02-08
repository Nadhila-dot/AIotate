package pipeline

import (
	"context"
	"fmt"
	"strings"

	"nadhi.dev/sarvar/fun/ai"
)

// SystemPrompt is the fixed system prompt for deterministic generation
const SystemPrompt = `You are a deterministic document generation engine.

Rules:
- Output ONLY valid LaTeX
- Do not explain
- Do not apologize
- Do not include markdown code blocks
- Use only standard packages
- Never invent data
- Never use placeholders or TODOs
- If uncertain, choose the simplest valid solution`

// GenerateDesign creates a design specification from the prompt
func GenerateDesign(ctx context.Context, conv *Conversation, prompt string, attachments []ai.Attachment) (string, error) {
	// Add user prompt to conversation
	conv.AddMessage("user", prompt)

	// Build conversation history for context
	messages := buildMessages(conv, fmt.Sprintf(`Create a detailed design specification for an educational worksheet based on this request:

%s

Output a structured design that includes:
- Document type and purpose
- Content sections and topics
- Question types and difficulty levels
- Layout and formatting requirements
- Any special requirements

Be specific and detailed. This design will be used to generate LaTeX code.`, prompt))

	// Call AI with utility model (fast)
	var result string
	var err error
	if len(attachments) > 0 {
		result, err = ai.GenerateWithAttachments(ctx, ai.TaskUtility, messages, attachments)
	} else {
		result, err = ai.Generate(ctx, ai.TaskUtility, messages)
	}
	if err != nil {
		return "", fmt.Errorf("design generation failed: %w", err)
	}

	design := fmt.Sprintf("%v", result)

	// Add assistant response to conversation
	conv.AddMessage("assistant", design)

	return design, nil
}

// GenerateLatex creates LaTeX code from the design
func GenerateLatex(ctx context.Context, conv *Conversation, design string, stylePrompt string, attachments []ai.Attachment) (string, error) {
	// Build the structured prompt
	userPrompt := fmt.Sprintf(`Generate LaTeX for the following design.

Design:
%s

Visual Style (apply these definitions in the LaTeX):
%s

Constraints:
- Must compile with pdflatex
- No external assets
- No placeholders
- No TODOs
- Use only standard packages (article, amsmath, geometry, etc.)
- Output ONLY the LaTeX code, no explanations
- Do not wrap in markdown code blocks

If uncertain, choose the simplest valid solution.`, design, stylePrompt)

	conv.AddMessage("user", userPrompt)

	messages := buildMessages(conv, userPrompt)

	// Call AI with main model (high quality)
	var result string
	var err error
	if len(attachments) > 0 {
		result, err = ai.GenerateWithAttachments(ctx, ai.TaskLaTeXGeneration, messages, attachments)
	} else {
		result, err = ai.Generate(ctx, ai.TaskLaTeXGeneration, messages)
	}
	if err != nil {
		return "", fmt.Errorf("latex generation failed: %w", err)
	}

	latex := fmt.Sprintf("%v", result)

	// Clean up any markdown artifacts that might have slipped through
	latex = cleanLatex(latex)

	// Add assistant response to conversation
	conv.AddMessage("assistant", latex)

	return latex, nil
}

// FixLatex attempts to fix LaTeX compilation errors using AI
func FixLatex(ctx context.Context, conv *Conversation, latex string, errorLog string) (string, error) {
	fixPrompt := fmt.Sprintf(`The following LaTeX code failed to compile.

LaTeX Code:
%s

Error Log:
%s

Fix the LaTeX code to resolve the compilation error.

Rules:
- Output ONLY the corrected LaTeX code
- Do not explain what you changed
- Do not include markdown code blocks
- Preserve the original content and structure as much as possible
- Only fix what is necessary to make it compile

Output the complete corrected LaTeX code:`, latex, errorLog)

	conv.AddMessage("user", fixPrompt)

	messages := buildMessages(conv, fixPrompt)

	// Use utility model for fixes (faster)
	result, err := ai.Generate(ctx, ai.TaskUtility, messages)
	if err != nil {
		return "", fmt.Errorf("latex fix failed: %w", err)
	}

	fixedLatex := fmt.Sprintf("%v", result)
	fixedLatex = cleanLatex(fixedLatex)

	conv.AddMessage("assistant", fixedLatex)

	return fixedLatex, nil
}

// RefinePrompt allows iterative refinement of the design
func RefinePrompt(ctx context.Context, conv *Conversation, refinement string) (string, error) {
	conv.AddMessage("user", refinement)

	messages := buildMessages(conv, refinement)

	// Use utility model for refinements
	result, err := ai.Generate(ctx, ai.TaskUtility, messages)
	if err != nil {
		return "", fmt.Errorf("refinement failed: %w", err)
	}

	response := fmt.Sprintf("%v", result)
	conv.AddMessage("assistant", response)

	return response, nil
}

// buildMessages constructs the message array for AI generation
func buildMessages(conv *Conversation, currentPrompt string) []ai.Message {
	messages := []ai.Message{
		{
			Role:    "system",
			Content: SystemPrompt,
		},
	}

	// Add conversation history (excluding system messages)
	for _, msg := range conv.Messages {
		if msg.Role != "system" {
			messages = append(messages, ai.Message{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	return messages
}

// cleanLatex removes markdown artifacts and cleans up the LaTeX code
func cleanLatex(latex string) string {
	// Remove markdown code blocks
	latex = strings.TrimPrefix(latex, "```latex\n")
	latex = strings.TrimPrefix(latex, "```latex")
	latex = strings.TrimPrefix(latex, "```\n")
	latex = strings.TrimPrefix(latex, "```")
	latex = strings.TrimSuffix(latex, "\n```")
	latex = strings.TrimSuffix(latex, "```")

	// Trim whitespace
	latex = strings.TrimSpace(latex)

	return latex
}

// GenerateDescription creates a short description for the job
func GenerateDescription(ctx context.Context, prompt string) (string, error) {
	messages := []ai.Message{
		{
			Role:    "system",
			Content: "You are a concise description generator. Output only a brief 1-2 sentence description.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Create a brief description for this worksheet request: %s", prompt),
		},
	}

	result, err := ai.Generate(ctx, ai.TaskUtility, messages)
	if err != nil {
		return "", fmt.Errorf("description generation failed: %w", err)
	}

	return fmt.Sprintf("%v", result), nil
}

// GenerateTags creates tags for the job
func GenerateTags(ctx context.Context, prompt string) ([]string, error) {
	messages := []ai.Message{
		{
			Role:    "system",
			Content: "You are a tag generator. Output only comma-separated tags, no explanations.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Generate 3-5 relevant tags for this worksheet: %s", prompt),
		},
	}

	result, err := ai.Generate(ctx, ai.TaskUtility, messages)
	if err != nil {
		return nil, fmt.Errorf("tag generation failed: %w", err)
	}

	tagsStr := fmt.Sprintf("%v", result)
	tags := strings.Split(tagsStr, ",")

	// Clean up tags
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}

	return tags, nil
}
