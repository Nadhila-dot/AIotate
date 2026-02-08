package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"nadhi.dev/sarvar/fun/auth"
	"nadhi.dev/sarvar/fun/server"
	"nadhi.dev/sarvar/fun/websearch"
)

// ToolsIndex registers utility routes like web search
func ToolsIndex() error {
	server.Route.Post("/api/v1/tools/web-search", func(c *fiber.Ctx) error {
		var req struct {
			Query string `json:"query"`
			Limit int    `json:"limit"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
		}

		q := strings.TrimSpace(req.Query)
		if q == "" {
			return c.Status(400).JSON(fiber.Map{"error": "query is required"})
		}

		// Validate session
		authHeader := c.Get("Authorization")
		if len(authHeader) < 8 || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(401).JSON(fiber.Map{"error": "missing or invalid authorization header"})
		}
		sessionID := authHeader[7:]
		valid, err := auth.IsSessionValid(sessionID)
		if err != nil || !valid {
			return c.Status(401).JSON(fiber.Map{"error": "invalid session"})
		}

		context, results, err := websearch.SearchAndExtract(q, req.Limit)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"query":   q,
			"results": results,
			"context": context,
		})
	})

	return nil
}
