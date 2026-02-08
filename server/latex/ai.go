package latex

import (
	"fmt"
	"log"

	"nadhi.dev/sarvar/fun/ai"
)

// FixLatexWithAI attempts to fix LaTeX content using the configured AI provider
func FixLatexWithAI(texContent, errorMsg string) (string, error) {
	// Create prompt for AI
	prompt := fmt.Sprintf(`You are an expert LaTeX engineer whose sole job is to fix LaTeX sources so they compile with Tectonic. Using the ERROR MESSAGE and the LATEX DOCUMENT below, produce a corrected LaTeX source that will compile with Tectonic. Follow these rules strictly:
1) Diagnose the error from the provided message and make minimal, targeted fixes (syntax, missing braces, unclosed environments, incorrect environment names, missing math delimiters, mismatched \begin/\end, and missing common packages that are needed by the document).
2) Preserve the original document structure, macros, comments and intent; change only what is necessary to make it compile.
3) If adding packages is required, add only widely-available packages in the preamble (no external files). Prefer safety and compatibility with Tectonic.
4) Do not add explanations, diagnostics, or any text outside the LaTeX source. Do not use markdown or code fences.
5) If a best-effort fix still may have issues, return the best corrected LaTeX source you can produce (still with no explanations).

ERROR MESSAGE FOR THE TECTONIC LATEX ENGINE:
%s

LATEX DOCUMENT:
%s`, errorMsg, texContent)

	// Use utility model for LaTeX fixing
	systemPrompt := "You are an expert LaTeX engineer. Fix LaTeX compilation errors."
	response, err := ai.GenerateSimple(ai.TaskUtility, systemPrompt, prompt)
	if err != nil {
		return "", fmt.Errorf("AI error: %w", err)
	}

	// Clean up the response - remove any markdown code block markers
	fixedLatex := RemoveCodeBlockMarkers(response)

	log.Printf("Successfully received fixed LaTeX from AI")
	return fixedLatex, nil
}
