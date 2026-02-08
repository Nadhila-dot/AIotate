package embed

import (
	"embed"
)

// Embed the frontend dist directory
//
//go:embed dist
var DistFS embed.FS
