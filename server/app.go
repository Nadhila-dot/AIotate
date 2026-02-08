package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"nadhi.dev/sarvar/fun/bootstrap"
	_ "nadhi.dev/sarvar/fun/config"
	embeddedAssets "nadhi.dev/sarvar/fun/embed"
)

// ExtractedAssetsPath stores where assets were extracted
var ExtractedAssetsPath string

func init() {
	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal("Failed to get executable path:", err)
	}
	exeDir := filepath.Dir(exePath)

	// Create web directory next to executable
	webDir := filepath.Join(exeDir, "web")
	ExtractedAssetsPath = webDir

	log.Printf("[ASSETS] Extracting embedded assets to: %s", webDir)

	// Extract embedded dist to web/dist
	distDir := filepath.Join(webDir, "dist")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		log.Fatal("Failed to create web/dist directory:", err)
	}

	// Walk through embedded files and extract
	err = fs.WalkDir(embeddedAssets.DistFS, "dist", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root "dist" directory itself
		if path == "dist" {
			return nil
		}

		// Get relative path (remove "dist/" prefix)
		relPath := path[5:] // Remove "dist/"
		targetPath := filepath.Join(distDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// Read embedded file
		data, err := fs.ReadFile(embeddedAssets.DistFS, path)
		if err != nil {
			return err
		}

		// Write to disk
		return os.WriteFile(targetPath, data, 0644)
	})

	if err != nil {
		log.Fatal("Failed to extract embedded assets:", err)
	}

	log.Println("[ASSETS] ✓ Assets extracted successfully")

	// Call bootstrap to initialize application
	bootstrap.Initialize()
}
