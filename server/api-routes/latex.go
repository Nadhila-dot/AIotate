package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"nadhi.dev/sarvar/fun/latex"
	"nadhi.dev/sarvar/fun/server"
)

const maxLatexPreviewBytes = 1_000_000

// LatexIndex registers LaTeX preview routes.
func LatexIndex() error {
	server.Route.Post("/api/v1/latex/preview", func(c *fiber.Ctx) error {
		var body struct {
			Latex string `json:"latex"`
		}

		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid json"})
		}

		latexSource := strings.TrimSpace(body.Latex)
		if latexSource == "" {
			return c.Status(400).JSON(fiber.Map{"error": "latex required"})
		}
		if len(latexSource) > maxLatexPreviewBytes {
			return c.Status(413).JSON(fiber.Map{"error": "latex too large"})
		}

		prepared, err := latex.PreparePreviewLatex(latexSource)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		html, err := latex.ConvertLatexToHTML(prepared, "preview.tex")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	})

	return nil
}
