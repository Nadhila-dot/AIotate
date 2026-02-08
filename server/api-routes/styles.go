package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	store "nadhi.dev/sarvar/fun/database"
	"nadhi.dev/sarvar/fun/db"
	"nadhi.dev/sarvar/fun/server"
)

func StylesIndex() error {
	server.Route.Get("/api/v1/styles", func(c *fiber.Ctx) error {
		username, err := getUsernameFromAuth(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
		}
		styles, err := store.GetAllStyles(db.StylesDB, username)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to get styles"})
		}
		return c.JSON(styles)
	})

	server.Route.Get("/api/v1/styles/default", func(c *fiber.Ctx) error {
		username, err := getUsernameFromAuth(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
		}
		style, err := store.GetDefaultStyle(db.StylesDB, username)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "default style not set"})
		}
		return c.JSON(style)
	})

	server.Route.Get("/api/v1/styles/:name", func(c *fiber.Ctx) error {
		username, err := getUsernameFromAuth(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
		}
		name := strings.TrimSpace(c.Params("name"))
		if name == "" {
			return c.Status(400).JSON(fiber.Map{"error": "invalid style name"})
		}
		style, err := store.GetStyle(db.StylesDB, username, name)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "style not found"})
		}
		return c.JSON(style)
	})

	server.Route.Post("/api/v1/styles", func(c *fiber.Ctx) error {
		username, err := getUsernameFromAuth(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
		}
		var body struct {
			Name        string `json:"name"`
			Prompt      string `json:"prompt"`
			Description string `json:"description"`
			IsDefault   bool   `json:"isDefault"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
		}
		body.Name = strings.TrimSpace(body.Name)
		body.Prompt = strings.TrimSpace(body.Prompt)

		if body.Name == "" || body.Prompt == "" {
			return c.Status(400).JSON(fiber.Map{"error": "name and prompt are required"})
		}

		style, err := store.CreateStyle(db.StylesDB, username, body.Name, body.Prompt, body.Description, body.IsDefault)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(style)
	})

	server.Route.Put("/api/v1/styles/:name", func(c *fiber.Ctx) error {
		username, err := getUsernameFromAuth(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
		}
		name := strings.TrimSpace(c.Params("name"))
		if name == "" {
			return c.Status(400).JSON(fiber.Map{"error": "invalid style name"})
		}
		var body struct {
			Prompt      string `json:"prompt"`
			Description string `json:"description"`
			IsDefault   bool   `json:"isDefault"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
		}
		body.Prompt = strings.TrimSpace(body.Prompt)

		if body.Prompt == "" {
			return c.Status(400).JSON(fiber.Map{"error": "prompt is required"})
		}

		style, err := store.UpdateStyle(db.StylesDB, username, name, body.Prompt, body.Description)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		if body.IsDefault {
			if _, err := store.SetDefaultStyle(db.StylesDB, username, name); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
		}

		return c.JSON(style)
	})

	server.Route.Delete("/api/v1/styles/:name", func(c *fiber.Ctx) error {
		username, err := getUsernameFromAuth(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
		}
		name := strings.TrimSpace(c.Params("name"))
		if name == "" {
			return c.Status(400).JSON(fiber.Map{"error": "invalid style name"})
		}
		if err := store.DeleteStyle(db.StylesDB, username, name); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to delete style"})
		}
		return c.JSON(fiber.Map{"status": "deleted"})
	})

	server.Route.Post("/api/v1/styles/:name/default", func(c *fiber.Ctx) error {
		username, err := getUsernameFromAuth(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
		}
		name := strings.TrimSpace(c.Params("name"))
		if name == "" {
			return c.Status(400).JSON(fiber.Map{"error": "invalid style name"})
		}
		style, err := store.SetDefaultStyle(db.StylesDB, username, name)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(style)
	})

	return nil
}
