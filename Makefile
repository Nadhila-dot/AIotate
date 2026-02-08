# AIotate Makefile
# Build system for frontend + backend single binary distribution

.PHONY: all build build-all frontend backend clean dev install help

# Directories
FRONTEND_DIR := ./frontend
SERVER_DIR := ./server
ASSETS_DIR := $(SERVER_DIR)/assets
BUILD_DIR := $(SERVER_DIR)/builds

# Binary name
BINARY_NAME := aiotate
ifeq ($(OS),Windows_NT)
	BINARY_NAME := aiotate.exe
endif

# Build flags
GO_BUILD_FLAGS := -tags webkit2_41 -ldflags="-s -w"
CGO_ENABLED := 1

# Default target
all: build

# ============================================================================
# Help
# ============================================================================

help:
	@echo ""
	@echo "╔════════════════════════════════════════════════════════════════╗"
	@echo "║              AIotate Build System                              ║"
	@echo "╚════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "Available targets:"
	@echo ""
	@echo "  make build       Build for current platform"
	@echo "  make build-all   Build for all platforms (macOS, Windows, Linux)"
	@echo "  make frontend    Build React frontend only"
	@echo "  make backend     Build Go binary only"
	@echo "  make dev         Run development server"
	@echo "  make install     Install dependencies"
	@echo "  make clean       Clean build artifacts"
	@echo "  make test        Run tests"
	@echo "  make help        Show this help"
	@echo ""
	@echo "Cross-compilation requirements:"
	@echo "  • Zig (for Linux builds from macOS)"
	@echo "    Install: brew install zig"
	@echo "  • MinGW-w64 (for Windows builds from macOS)"
	@echo "    Install: brew install mingw-w64"
	@echo ""

# ============================================================================
# Main Build Targets
# ============================================================================

build: banner frontend backend summary

build-all: banner frontend build-macos build-windows build-linux summary-all

# ============================================================================
# Build Steps
# ============================================================================

banner:
	@echo ""
	@echo "╔════════════════════════════════════════════════════════════════╗"
	@echo "║              AIotate Build Script                              ║"
	@echo "╚════════════════════════════════════════════════════════════════╝"
	@echo ""

frontend:
	@echo "📦 [1/3] Building React frontend..."
	@cd $(FRONTEND_DIR) && \
		if [ ! -d "node_modules" ]; then \
			echo "  → Installing dependencies..."; \
			npm install; \
		fi && \
		echo "  → Running Vite build..." && \
		npm run build
	@echo "  → Copying to server embed..."
	@mkdir -p $(SERVER_DIR)/embed/dist
	@rm -rf $(SERVER_DIR)/embed/dist/*
	@cp -r $(FRONTEND_DIR)/dist/* $(SERVER_DIR)/embed/dist/
	@echo "  ✓ Frontend built and copied to embed"
	@echo ""

inject:
	@echo "📋 [2/3] Preparing embed directory..."
	@mkdir -p $(SERVER_DIR)/embed
	@echo "  ✓ Embed directory ready"
	@echo ""

backend: inject
	@echo "🔨 [3/3] Building Go binary..."
	@mkdir -p $(BUILD_DIR)
	@echo "  → Platform: $$(go env GOOS)/$$(go env GOARCH)"
	@echo "  → Embedding assets..."
	@cd $(SERVER_DIR) && \
		CGO_ENABLED=$(CGO_ENABLED) go build $(GO_BUILD_FLAGS) -o builds/$(BINARY_NAME) .
	@if [ $$? -eq 0 ]; then \
		echo "  ✓ Binary built"; \
		if [ -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
			echo "  → Size: $$(du -h $(BUILD_DIR)/$(BINARY_NAME) | cut -f1)"; \
		fi; \
	else \
		echo "  ✗ Build failed"; \
		exit 1; \
	fi
	@echo ""

# ============================================================================
# Cross-Platform Builds
# ============================================================================

build-macos:
	@echo "🍎 Building for macOS..."
	@mkdir -p $(BUILD_DIR)/darwin-amd64 $(BUILD_DIR)/darwin-arm64
	@echo "  → macOS Intel (amd64)..."
	@cd $(SERVER_DIR) && \
		GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 \
		go build $(GO_BUILD_FLAGS) -o builds/darwin-amd64/aiotate .
	@echo "  → macOS Apple Silicon (arm64)..."
	@cd $(SERVER_DIR) && \
		GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 \
		go build $(GO_BUILD_FLAGS) -o builds/darwin-arm64/aiotate .
	@echo "  ✓ macOS builds complete"
	@echo ""

build-windows:
	@echo "🪟 Building for Windows..."
	@mkdir -p $(BUILD_DIR)/windows-amd64
	@echo "  → Windows x64..."
	@cd $(SERVER_DIR) && \
		GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
		CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ \
		go build $(GO_BUILD_FLAGS) -o builds/windows-amd64/aiotate.exe .
	@echo "  ✓ Windows build complete"
	@echo ""

build-linux:
	@echo "🐧 Building for Linux..."
	@mkdir -p $(BUILD_DIR)/linux-amd64
	@echo "  → Linux x64 (using Zig)..."
	@cd $(SERVER_DIR) && \
		GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
		CC="zig cc -target x86_64-linux-gnu" \
		CXX="zig c++ -target x86_64-linux-gnu" \
		go build $(GO_BUILD_FLAGS) -o builds/linux-amd64/aiotate .
	@echo "  ✓ Linux build complete"
	@echo ""

# ============================================================================
# Summary
# ============================================================================

summary:
	@echo "✨ Build complete!"
	@echo ""
	@echo "╔════════════════════════════════════════════════════════════════╗"
	@echo "║                    Build Summary                               ║"
	@echo "╠════════════════════════════════════════════════════════════════╣"
	@echo "║  Binary:     $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "║  Platform:   $$(go env GOOS)/$$(go env GOARCH)"
	@if [ -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		echo "║  Size:       $$(du -h $(BUILD_DIR)/$(BINARY_NAME) | cut -f1)"; \
	fi
	@echo "║  Assets:     Embedded"
	@echo "╚════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "🚀 Ready to run: cd $(BUILD_DIR) && ./$(BINARY_NAME)"
	@echo ""

summary-all:
	@echo "✨ Multi-platform build complete!"
	@echo ""
	@echo "╔════════════════════════════════════════════════════════════════╗"
	@echo "║                Multi-Platform Build Summary                    ║"
	@echo "╠════════════════════════════════════════════════════════════════╣"
	@if [ -f "$(BUILD_DIR)/darwin-amd64/aiotate" ]; then \
		echo "║  🍎 macOS Intel:     $$(du -h $(BUILD_DIR)/darwin-amd64/aiotate | cut -f1)"; \
	fi
	@if [ -f "$(BUILD_DIR)/darwin-arm64/aiotate" ]; then \
		echo "║  🍎 macOS ARM:       $$(du -h $(BUILD_DIR)/darwin-arm64/aiotate | cut -f1)"; \
	fi
	@if [ -f "$(BUILD_DIR)/windows-amd64/aiotate.exe" ]; then \
		echo "║  🪟 Windows x64:     $$(du -h $(BUILD_DIR)/windows-amd64/aiotate.exe | cut -f1)"; \
	fi
	@if [ -f "$(BUILD_DIR)/linux-amd64/aiotate" ]; then \
		echo "║  🐧 Linux x64:       $$(du -h $(BUILD_DIR)/linux-amd64/aiotate | cut -f1)"; \
	fi
	@echo "╚════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "📦 Binaries located in: $(BUILD_DIR)/"
	@echo ""

# ============================================================================
# Development & Utilities
# ============================================================================

install:
	@echo "📥 Installing dependencies..."
	@cd $(FRONTEND_DIR) && npm install
	@echo "✓ Dependencies installed"

dev:
	@echo "🚀 Starting development server..."
	@echo ""
	@echo "  Frontend: http://localhost:5173"
	@echo "  Backend:  http://localhost:317"
	@echo ""
	@cd $(SERVER_DIR) && go run .

clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(SERVER_DIR)/embed/dist
	@rm -rf $(FRONTEND_DIR)/dist
	@echo "✓ Clean complete"

test:
	@echo "🧪 Running tests..."
	@cd $(SERVER_DIR) && go test ./...
	@echo "✓ Tests passed"

deps:
	@echo "📦 Installing Go dependencies..."
	@cd $(SERVER_DIR) && go mod download && go mod tidy
	@echo "✓ Go dependencies installed"
