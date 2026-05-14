.PHONY: all local windows pi pi32 wasm clean bump-major bump-minor bump-patch release

# --- Variables ---
BINARY_NAME=playstation41
BUILD_DIR=build
# Pulls the most recent tag or defaults to v0.0.0 if none exist
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0)
TIMESTAMP := $(shell date +"%-m-%-d-%y_%-I:%M%p")
WASM_EXEC_PATH=$(shell go env GOROOT)/$(shell if [ -d "$(shell go env GOROOT)/lib/wasm" ]; then echo "lib/wasm"; else echo "misc/wasm"; fi)/wasm_exec.js

# Removed pi32 from the 'all' target
all: local windows pi wasm

# --- Build Targets ---

local:
	@echo "Building $(VERSION) for macOS..."
	@mkdir -p $(BUILD_DIR)/macos
	go build -o $(BUILD_DIR)/macos/$(BINARY_NAME)_macos .

windows:
	@echo "Building $(VERSION) for Windows..."
	@mkdir -p $(BUILD_DIR)/windows
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/windows/$(BINARY_NAME).exe .

pi:
	@echo "Building $(VERSION) for Raspberry Pi (64-bit ARM)..."
	@mkdir -p $(BUILD_DIR)/pi
	docker build --platform linux/arm64 \
		--build-arg BUILD_PLATFORM=linux/arm64 \
		--build-arg TARGET_ARCH=arm64 \
		--build-arg BINARY_OUT=$(BINARY_NAME)_pi \
		-t $(BINARY_NAME)-pi64-builder .
	@docker create --name temp-pi64 $(BINARY_NAME)-pi64-builder
	@docker cp temp-pi64:/app/$(BINARY_NAME)_pi $(BUILD_DIR)/pi/$(BINARY_NAME)_pi
	@docker rm temp-pi64

pi32:
	@echo "⚠️ Legacy 32-bit build. Not included in release."
	@echo "Building $(VERSION) for Raspberry Pi (32-bit ARM)..."
	@mkdir -p $(BUILD_DIR)/pi
	docker build --platform linux/arm/v7 \
		--build-arg BUILD_PLATFORM=linux/arm/v7 \
		--build-arg TARGET_ARCH=arm \
		--build-arg BINARY_OUT=$(BINARY_NAME)_pi32 \
		-t $(BINARY_NAME)-pi32-builder .
	@docker create --name temp-pi32 $(BINARY_NAME)-pi32-builder
	@docker cp temp-pi32:/app/$(BINARY_NAME)_pi32 $(BUILD_DIR)/pi/$(BINARY_NAME)_pi32bit
	@docker rm temp-pi32

wasm:
	@echo "Building $(VERSION) for Web (WASM)..."
	@mkdir -p $(BUILD_DIR)/wasm
	GOOS=js GOARCH=wasm go build -o $(BUILD_DIR)/wasm/$(BINARY_NAME).wasm .
	cp "$(WASM_EXEC_PATH)" $(BUILD_DIR)/wasm/
	@if [ -f index.html ]; then cp index.html $(BUILD_DIR)/wasm/; fi

# --- Versioning Targets ---

bump-major:
	@CURRENT=$(VERSION); \
	MAJOR=$$(echo $$CURRENT | sed 's/v\([0-9]*\)\..*/\1/'); \
	NEW_MAJOR=$$((MAJOR + 1)); \
	NEW_TAG="v$$NEW_MAJOR.0.0"; \
	echo "Bumping MAJOR: $$CURRENT → $$NEW_TAG"; \
	git tag -a $$NEW_TAG -m "Release $$NEW_TAG"; \
	echo "Created tag $$NEW_TAG. Run 'git push origin $$NEW_TAG' to push it."

bump-minor:
	@CURRENT=$(VERSION); \
	MAJOR=$$(echo $$CURRENT | sed 's/v\([0-9]*\)\..*/\1/'); \
	MINOR=$$(echo $$CURRENT | sed 's/v[0-9]*\.\([0-9]*\)\..*/\1/'); \
	NEW_MINOR=$$((MINOR + 1)); \
	NEW_TAG="v$$MAJOR.$$NEW_MINOR.0"; \
	echo "Bumping MINOR: $$CURRENT → $$NEW_TAG"; \
	git tag -a $$NEW_TAG -m "Release $$NEW_TAG"; \
	echo "Created tag $$NEW_TAG. Run 'git push origin $$NEW_TAG' to push it."

bump-patch:
	@CURRENT=$(VERSION); \
	MAJOR=$$(echo $$CURRENT | sed 's/v\([0-9]*\)\..*/\1/'); \
	MINOR=$$(echo $$CURRENT | sed 's/v[0-9]*\.\([0-9]*\)\..*/\1/'); \
	PATCH=$$(echo $$CURRENT | sed 's/v[0-9]*\.[0-9]*\.\([0-9]*\).*/\1/'); \
	NEW_PATCH=$$((PATCH + 1)); \
	NEW_TAG="v$$MAJOR.$$MINOR.$$NEW_PATCH"; \
	echo "Bumping PATCH: $$CURRENT → $$NEW_TAG"; \
	git tag -a $$NEW_TAG -m "Release $$NEW_TAG"; \
	echo "Created tag $$NEW_TAG. Run 'git push origin $$NEW_TAG' to push it."

# --- Release & Cleanup ---

release: all
	@echo "Creating GitHub Release for $(VERSION)..."
	gh release create $(VERSION) \
		$(BUILD_DIR)/macos/$(BINARY_NAME)_macos \
		$(BUILD_DIR)/pi/$(BINARY_NAME)_pi \
		$(BUILD_DIR)/windows/$(BINARY_NAME).exe \
		$(BUILD_DIR)/wasm/$(BINARY_NAME).wasm \
		--title "Release $(VERSION)" \
		--notes "Automated multi-platform build for $(VERSION)"

clean:
	@echo "Cleaning build directory..."
	rm -rf $(BUILD_DIR)