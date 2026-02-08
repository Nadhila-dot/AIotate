package latex

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const previewTemplate = `\documentclass[11pt]{article}
\usepackage{xcolor}
\usepackage{hyperref}
\usepackage{geometry}
\usepackage{graphicx}
\geometry{margin=1in}

% Style prompt
%s

\begin{document}
\section*{Style Preview}
This preview uses your current style prompt to render a sample layout.\\
	extcolor{primary}{Primary Accent}\\
	extcolor{secondary}{Secondary Accent}\\
	extcolor{accent}{Accent}\\
	extcolor{light}{Light Accent}

\vspace{12pt}
\fcolorbox{primary}{light}{\parbox{0.88\linewidth}{\centering
	extbf{Sample callout}\\
Use this block to verify your primary/secondary palette, spacing, and typography.
}}

\end{document}`

// PreparePreviewLatex normalizes user input into a compilable LaTeX document.
func PreparePreviewLatex(input string) (string, error) {
	src := strings.TrimSpace(input)
	if src == "" {
		return "", fmt.Errorf("latex content is empty")
	}

	if extracted, err := ExtractOutput(src); err == nil && strings.TrimSpace(extracted) != "" {
		src = extracted
	}

	src = RemoveCodeBlockMarkers(src)
	src = strings.TrimSpace(src)
	if src == "" {
		return "", fmt.Errorf("latex content is empty")
	}

	if !strings.Contains(src, "\\documentclass") {
		return fmt.Sprintf(previewTemplate, src), nil
	}

	if !strings.Contains(src, "\\begin{document}") {
		return src + "\n\\begin{document}\n\\section*{Preview}\nPreview content.\n\\end{document}\n", nil
	}
	if !strings.Contains(src, "\\end{document}") {
		return src + "\n\\end{document}\n", nil
	}

	return src, nil
}

// ConvertLatexToHTML renders LaTeX to HTML using Tectonic.
func ConvertLatexToHTML(latexContent, texFilename string) (string, error) {
	if strings.TrimSpace(latexContent) == "" {
		return "", fmt.Errorf("latex content is empty")
	}

	if texFilename == "" {
		texFilename = "preview.tex"
	}

	tempDir, err := ioutil.TempDir("", "latex-preview")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	texPath := filepath.Join(tempDir, texFilename)
	if err := ioutil.WriteFile(texPath, []byte(latexContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write latex file: %w", err)
	}

	fileBase := strings.TrimSuffix(texFilename, filepath.Ext(texFilename))
	htmlPath := filepath.Join(tempDir, fileBase+".html")

	cmd := exec.Command("tectonic", "--outfmt=html", "--keep-logs", "-o", tempDir, texPath)
	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("tectonic html failed: %w\nTectonic output:\n%s", err, string(output))
	}

	data, err := ioutil.ReadFile(htmlPath)
	if err != nil {
		return "", fmt.Errorf("failed to read html: %w", err)
	}

	return string(data), nil
}
