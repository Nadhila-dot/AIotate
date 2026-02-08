package server

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"nadhi.dev/sarvar/fun/db"
	logg "nadhi.dev/sarvar/fun/logs"
	websocket "nadhi.dev/sarvar/fun/websocket"
)

var Route *fiber.App

func init() {
	Route = fiber.New(fiber.Config{
		DisableStartupMessage: true,             // Disable Fiber's default banner
		BodyLimit:             30 * 1024 * 1024, // 30MB to allow multipart overhead for 20MB uploads
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if err == fiber.ErrRequestEntityTooLarge {
				return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
					"error": "Upload too large. Max total upload size is 20MB.",
				})
			}
			return fiber.DefaultErrorHandler(c, err)
		},
	})
	Route.Use(logger.New())
	logg.Info("Fiber instance created successfully")
	websocket.Init(log.Default())
	logg.Info("WebSocket manager initialized successfully")

	// Initialize databases
	// Sessions, Users and etc
	if err := db.InitSessionsDB(); err != nil {
		logg.Error("Failed to initialize sessions DB: ")
	}
	if err := db.InitUsersDB(); err != nil {
		logg.Error("Failed to initialize users DB: ")
	}
	if err := db.InitQueueDB(); err != nil {
		logg.Error("Failed to initialize queue DB: ")
	}
	if err := db.InitNotebooksDB(); err != nil {
		logg.Error("Failed to initialize notebooks DB: ")
	}
	if err := db.InitStylesDB(); err != nil {
		logg.Error("Failed to initialize styles DB: ")
	}
}
