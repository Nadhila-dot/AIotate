package routes

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"nadhi.dev/sarvar/fun/api-routes"
	"nadhi.dev/sarvar/fun/auth"
	"nadhi.dev/sarvar/fun/server"
)

var assetsPath string

func SetAssetsPath(path string) {
	// Path is already the web directory, just append dist
	assetsPath = filepath.Join(path, "dist")
	log.Printf("[ROUTER] Serving assets from: %s", assetsPath)
}

func Register() {
	// Register all routes
	index()
	health()

	server.Route.Use("/api/v1", auth.CheckAuth)
	api.Index()
	api.KasmIndex()
	api.KasmCreateSession()
	api.AuthIndex()
	api.VelaIndex()
	api.SheetsIndex()
	api.StylesIndex()
	api.PipelineIndex()
	api.ToolsIndex()
	api.LatexIndex()
	api.RegisterWebsocketRoutes()
	api.Notebooks()
}

func index() {
	// Serve static files and SPA
	server.Route.Use(func(c *fiber.Ctx) error {
		path := c.Path()
		method := c.Method()

		// Only handle GET requests
		if method != fiber.MethodGet {
			return c.Next()
		}

		// Skip API routes
		if strings.HasPrefix(path, "/api/") {
			return c.Next()
		}

		// Skip special routes
		if strings.HasPrefix(path, "/vela/bucket") || path == "/health" {
			return c.Next()
		}

		// Root redirect
		if path == "/" || path == "" {
			return c.Redirect("/home", fiber.StatusFound)
		}

		// Try to serve static file
		relPath := strings.TrimPrefix(path, "/")
		filePath := filepath.Join(assetsPath, relPath)

		// Check if file exists
		if _, err := os.Stat(filePath); err == nil {
			return c.SendFile(filePath)
		}

		// Fallback to index.html for SPA routing (React Router)
		return c.SendFile(filepath.Join(assetsPath, "index.html"))
	})
}

func health() {
	server.Route.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "200",
			"message": "service is active",
		})
	})
}
