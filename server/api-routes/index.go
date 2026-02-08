package api

import (
	"github.com/gofiber/fiber/v2"
	"nadhi.dev/sarvar/fun/config"
	"nadhi.dev/sarvar/fun/server"
	ws "nadhi.dev/sarvar/fun/websocket"
)

func Index() error {
	server.Route.Get("/api/v1", func(c *fiber.Ctx) error {
		// Send broadcast notification when API index is hit
		//  ws.GetManager().Broadcast(ws.Info("info", "Someone accessed the API index"))

		return c.JSON(fiber.Map{
			"status":  "200",
			"message": "service is active",
			"endpoints": []string{
				"/api/",
				"/api/kasm",
				"/api/v1/system",
				"/api/set (GET, POST)",
			},
		})
	})

	server.Route.Get("/api/v1/system", func(c *fiber.Ctx) error {
		// Send broadcast notification when system endpoint is hit
		//  ws.GetManager().Broadcast(ws.Info("system", "System information was requested"))

		return c.JSON(fiber.Map{
			"status":  "200",
			"message": "service is active",
			"data": fiber.Map{
				"build":  "vela-beta-v1.4",
				//"date":   "September 2025",
                "date":   "February 2026",
				"author": "Nadhi",
			},
		})
	})

	// GET /api/set
	server.Route.Get("/api/v1/set", func(c *fiber.Ctx) error {
		cfg, err := config.GetConfig()
		if err != nil {
			ws.GetManager().Broadcast(ws.Error("error", "Failed to read configuration", map[string]interface{}{}))
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to read configuration",
			})
		}

		return c.JSON(cfg)
	})

	// POST /api/set
	server.Route.Post("/api/v1/set", func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if len(auth) < 8 || auth[:7] != "Bearer " {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		var newData map[string]interface{}
		if err := c.BodyParser(&newData); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid JSON",
			})
		}

		if err := config.SaveConfig(newData); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"status": 500,
				"error":  "Failed to save configuration",
			})
		}

		return c.JSON(fiber.Map{
			"status":  200,
			"message": "Configuration updated successfully",
		})
	})

	return nil
}
