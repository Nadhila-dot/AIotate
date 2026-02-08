package bootstrap

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// SystemChecks performs all prerequisite checks
func SystemChecks() error {
	log.Println("[CHECKS] Running system prerequisite checks...")

	// Get executable directory and change to it
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(exePath)

	log.Printf("[CHECKS] Executable directory: %s", exeDir)

	// Change working directory to executable location
	if err := os.Chdir(exeDir); err != nil {
		return fmt.Errorf("failed to change to executable directory: %w", err)
	}

	log.Printf("[CHECKS] Working directory set to: %s", exeDir)

	// Check for tectonic
	if err := checkTectonic(); err != nil {
		return err
	}

	// Create required directories
	if err := createDirectories(); err != nil {
		return err
	}

	// Check/create config file
	if err := checkConfigFile(); err != nil {
		return err
	}

	log.Println("[CHECKS] All prerequisite checks passed ✓")
	return nil
}

// checkTectonic verifies tectonic is installed and accessible
func checkTectonic() error {
	log.Println("[CHECKS] Checking for tectonic LaTeX compiler...")

	cmd := exec.Command("tectonic", "--version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf(`
╔════════════════════════════════════════════════════════════════╗
║                    TECTONIC NOT FOUND                          ║
╚════════════════════════════════════════════════════════════════╝

AIotate requires Tectonic LaTeX compiler to generate PDFs.

Installation instructions:
%s

After installation, restart AIotate.
`, getInstallInstructions())
	}

	log.Printf("[CHECKS] Tectonic found: %s", string(output))
	return nil
}

// getInstallInstructions returns platform-specific installation instructions
func getInstallInstructions() string {
	switch runtime.GOOS {
	case "darwin":
		return `
  macOS (Homebrew):
    brew install tectonic

  macOS (Manual):
    Download from: https://github.com/tectonic-typesetting/tectonic/releases
`
	case "linux":
		return `
  Linux (Cargo):
    cargo install tectonic

  Linux (Package Manager):
    # Debian/Ubuntu
    sudo apt install tectonic
    
    # Arch Linux
    sudo pacman -S tectonic

  Linux (Manual):
    Download from: https://github.com/tectonic-typesetting/tectonic/releases
`
	case "windows":
		return `
  Windows (Scoop):
    scoop install tectonic

  Windows (Cargo):
    cargo install tectonic

  Windows (Manual):
    Download from: https://github.com/tectonic-typesetting/tectonic/releases
`
	default:
		return `
  Visit: https://tectonic-typesetting.github.io/install.html
`
	}
}

// createDirectories creates all required directories
func createDirectories() error {
	log.Println("[CHECKS] Creating required directories...")

	dirs := []string{
		"./badger-db",
		"./zp-database/users",
		"./zp-database/sessions",
		"./zp-database/notebooks",
		"./zp-database/queue",
		"./zp-database/styles",
		"./storage/bucket",
		"./storage/queue_data",
		"./generated",
		"./generated/gemini_fixes",
		"./logs",
	}

	// Get current working directory for logging
	cwd, _ := os.Getwd()

	for _, dir := range dirs {
		fullPath := filepath.Join(cwd, dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		log.Printf("[CHECKS]   ✓ %s", fullPath)
	}

	log.Printf("[CHECKS] Created %d directories ✓", len(dirs))
	return nil
}

// checkConfigFile ensures set.json exists with defaults
func checkConfigFile() error {
	log.Println("[CHECKS] Checking configuration file...")

	configPath := "./set.json"

	// Get absolute path for logging
	absPath, _ := filepath.Abs(configPath)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Println("[CHECKS] set.json not found, creating default configuration...")

		defaultConfig := `{
  "AI_PROVIDER": "gemini",
  "GEMINI_API_KEY": "",
  "OPENROUTER_API_KEY": "",
  "AI_MAIN_MODEL": "",
  "AI_UTILITY_MODEL": "",
  "MAX_SESSIONS": 2,
  "SHEET_QUEUE_DIR": "./storage/queue_data"
}
`
		if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}

		log.Printf("[CHECKS] ✓ Created: %s", absPath)
		log.Println("[CHECKS] ⚠ Please configure your API keys in Settings")
	} else {
		log.Printf("[CHECKS] ✓ Found: %s", absPath)
	}

	return nil
}
