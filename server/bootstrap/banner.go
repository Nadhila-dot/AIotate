package bootstrap

import (
	"fmt"
	"runtime"
)

// ShowBanner displays the AIotate ASCII art banner
func ShowBanner(port int) {
	fmt.Println()
	fmt.Println(`
   /\_/\  
  ( o.o )   AIotate - AI-Powered Educational Worksheet Generator
   > ^ <    Single Distribution Binary
  
  Made by Nadhi.dev
`)
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Printf("║  Server:     http://127.0.0.1:%-5d                            ║\n", port)
	fmt.Printf("║  Platform:   %-50s║\n", runtime.GOOS+"/"+runtime.GOARCH)
	fmt.Printf("║  Go Version: %-50s║\n", runtime.Version())
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// ShowStartupComplete displays completion message
func ShowStartupComplete() {
	fmt.Println()
	fmt.Println("✓ AIotate is ready!")
	fmt.Println("  Press Ctrl+C to stop the server")
	fmt.Println()
}

// ShowShutdown displays shutdown message
func ShowShutdown() {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    Shutting down AIotate...                    ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()
}
