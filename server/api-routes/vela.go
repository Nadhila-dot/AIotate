package api

import (
	_ "encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"nadhi.dev/sarvar/fun/server"
)

// List all files in ./storage and send as JSON array (Fiber version)
func ListStorageFilesFiber(c *fiber.Ctx) error {
	files := []string{}
	root := "./storage"

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath := strings.TrimPrefix(path, root+"/")
			files = append(files, relPath)
		}
		return nil
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to list files"})
	}

	return c.JSON(files)
}

func ServeStorageFileFiber(c *fiber.Ctx) error {
	relPath := c.Params("*")
	filePath := filepath.Join("./storage", relPath)

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "File not found"})
	}

	raw := c.Query("raw") == "true"

	switch ext := strings.ToLower(filepath.Ext(filePath)); ext {
	case ".png":
		c.Type("png")
	case ".jpg", ".jpeg":
		c.Type("jpeg")
	case ".pdf":
		c.Type("pdf")
	case ".txt", ".md", ".json", ".csv":
		// Force text display in browser like GitHub Raw
		c.Set("Content-Type", "text/plain; charset=utf-8")
		c.Set("Content-Disposition", "inline")
		c.Set("X-Content-Type-Options", "nosniff")

		if raw {
			return c.Send(data)
		}
		return c.SendString(string(data))
	default:
		c.Set("Content-Type", "text/plain; charset=utf-8")
		c.Set("Content-Disposition", "inline")
		c.Set("X-Content-Type-Options", "nosniff")

	}

	return c.Send(data)
}

// Check if our lovely tex engine is working
// so add a web helper for ts.
func CheckTectonic(c *fiber.Ctx) error {
	// Run tectonic --version to check if the LaTeX engine is working
	cmd := exec.Command("tectonic", "--version")
	output, err := cmd.Output()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status": "not working",
			"error":  err.Error(),
		})
	}

	// If successful, return the version info
	version := strings.TrimSpace(string(output))
	return c.JSON(fiber.Map{
		"status":  "working",
		"version": version,
	})
}

// Register Vela storage routes
func VelaIndex() error {
	// List all files in ./storage as JSON array
	server.Route.Get("/vela/list", ListStorageFilesFiber)

	// Serve a file from ./storage/bucket/whatever
	server.Route.Get("/vela/bucket/*", ServeStorageFileFiber)

	// Check if tectonic LaTeX engine is working
	server.Route.Get("/vela/check-tectonic", CheckTectonic)

	return nil
}
