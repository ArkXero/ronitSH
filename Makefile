BINARY_SERVER  := ronit-sh
BINARY_CTL     := ronit-sh-ctl
BUILD_DIR      := dist
SERVER_PKG     := ./cmd/ronit-sh
CTL_PKG        := ./cmd/ronit-sh-ctl

# Cross-compile target for the Ubuntu server.
GOOS   := linux
GOARCH := amd64

.PHONY: dev build test lint deploy clean

# Run the SSH server locally on 127.0.0.1:23234
dev:
	go run $(SERVER_PKG)

# Cross-compile for linux/amd64
build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BUILD_DIR)/$(BINARY_SERVER) $(SERVER_PKG)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BUILD_DIR)/$(BINARY_CTL)    $(CTL_PKG)
	@echo "Built: $(BUILD_DIR)/$(BINARY_SERVER) and $(BUILD_DIR)/$(BINARY_CTL)"

# Run all tests
test:
	go test ./...

# Static analysis
lint:
	go vet ./...

# Deploy to the Ubuntu server (prompts for confirmation)
deploy: build
	./scripts/deploy.sh

# Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)
