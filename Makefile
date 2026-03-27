.PHONY: all local windows pi wasm clean bump release

# --- Variables ---
BINARY_NAME=playstation41
BUILD_DIR=build

# 1. Versioning: Get latest tag (v0.1.2) or default to v0.0.0
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0)
# 2. Timestamp for local dev builds
TIMESTAMP := $(shell date +"%-m-%-d-%y_%-I:%M%p")

# Check for Go 1.24+ wasm_exec location
WASM_EXEC_PATH=$(shell go env GOROOT)/$(shell if [ -d "$(shell go env GOROOT)/lib/wasm" ]; then echo "lib/wasm"; else echo "misc/wasm"; fi)/wasm_exec.js

all: local windows pi wasm

# --- Build Targets ---

local:
	@echo "Building $(VERSION) for macOS..."
	@mkdir -p $(BUILD_DIR)/macos
	# Build with timestamp for your personal use
	go build -o $(BUILD_DIR)/macos/$(BINARY_NAME)_$(TIMESTAMP) .
	# Build a clean name for the GitHub distribution
	go build -o $(BUILD_DIR)/macos/$(BINARY_NAME)_macos .
	@echo "Done: $(BUILD_DIR)/macos/$(BINARY_NAME)_macos"

windows:
	@echo "Building $(VERSION) for Windows..."
	@mkdir -p $(BUILD_DIR)/windows
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/windows/$(BINARY_NAME).exe .

pi:
	@echo "Building $(VERSION) for Raspberry Pi (64-bit ARM) via Docker..."
	@mkdir -p $(BUILD_DIR)/pi
	docker build --platform linux/arm64 -t $(BINARY_NAME)-pi-builder .
	@docker create --name temp-builder $(BINARY_NAME)-pi-builder
	@docker cp temp-builder:/app/$(BINARY_NAME)_pi $(BUILD_DIR)/pi/$(BINARY_NAME)_pi
	@docker rm temp-builder

wasm:
	@echo "Building $(VERSION) for Web (WASM)..."
	@mkdir -p $(BUILD_DIR)/wasm
	GOOS=js GOARCH=wasm go build -o $(BUILD_DIR)/wasm/$(BINARY_NAME).wasm .
	cp "$(WASM_EXEC_PATH)" $(BUILD_DIR)/wasm/
	@if [ -f index.html ]; then cp index.html $(BUILD_DIR)/wasm/; fi

# --- Utility Targets ---

# Increments the PATCH version (e.g., v0.1.1 -> v0.1.2)
bump:
	@echo "Current version: $(VERSION)"
	@NEXT_VERSION=$$(echo $(VERSION) | awk -F. '{print $$1"."$$2"."$$3+1}'); \
	echo "Bumping to $$NEXT_VERSION..."; \
	git tag $$NEXT_VERSION
	@echo "New tag $$NEXT_VERSION created locally."
	@echo "Run 'git push --tags' to sync with GitHub."

# Creates a GitHub Release and uploads all 3 binaries
release: all
	@echo "Creating GitHub Release for $(VERSION)..."
	gh release create $(VERSION) \
		$(BUILD_DIR)/macos/$(BINARY_NAME)_macos \
		$(BUILD_DIR)/pi/$(BINARY_NAME)_pi \
		$(BUILD_DIR)/windows/$(BINARY_NAME).exe \
		--title "Release $(VERSION)" \
		--notes "Automated multi-platform build for $(VERSION)"

clean:
	@echo "Cleaning build directory..."
	rm -rf $(BUILD_DIR)