.PHONY: all build build-app build-docs dev dev-app dev-docs \
        docker docker-build docker-up docker-down \
        test test-go test-web clean help

# ── Paths ──────────────────────────────────────────────────────────────────
APP_DIR    := apps/app
DOCS_DIR   := apps/docs
BUILD_DIR  := build
BINARY     := $(BUILD_DIR)/policyflow

# ── Default ────────────────────────────────────────────────────────────────
all: build

# ══════════════════════════════════════════════════════════════════════════
## Build
# ══════════════════════════════════════════════════════════════════════════

## build: build frontend + Go binary into ./build/policyflow
build: build-app

## build-app: build Next.js frontend, then compile Go binary with embedded assets
build-app:
	@echo "▶ Building frontend..."
	cd $(APP_DIR) && pnpm --prefix web install && pnpm --prefix web build
	@echo "▶ Compiling Go binary..."
	@mkdir -p $(BUILD_DIR)
	cd $(APP_DIR) && go build -ldflags="-s -w" -o ../../$(BINARY) .
	@echo "✓ Binary: $(BINARY)"

## build-docs: build the Fumadocs documentation site
build-docs:
	@echo "▶ Building docs..."
	cd $(DOCS_DIR) && pnpm install && pnpm build
	@echo "✓ Docs built"

## build-all: build everything (app + docs)
build-all: build-app build-docs

# ══════════════════════════════════════════════════════════════════════════
## Development
# ══════════════════════════════════════════════════════════════════════════

## dev-app: start Go backend (port 8080) proxying to Next.js dev server (port 3001)
##          Run in TWO terminals: `make dev-frontend` and `make dev-backend`
dev-backend:
	cd $(APP_DIR) && WEB_DEV_PROXY=http://localhost:3001 go run .

## dev-frontend: start Next.js dev server on port 3001
dev-frontend:
	cd $(APP_DIR) && pnpm --prefix web install && pnpm --prefix web dev

## dev-docs: start Fumadocs dev server on port 3000
dev-docs:
	cd $(DOCS_DIR) && pnpm install && pnpm dev

# ══════════════════════════════════════════════════════════════════════════
## Run
# ══════════════════════════════════════════════════════════════════════════

## run: build and run the binary (production-like)
run: build-app
	./$(BINARY)

## run-binary: run an already-compiled binary
run-binary:
	./$(BINARY)

# ══════════════════════════════════════════════════════════════════════════
## Tests
# ══════════════════════════════════════════════════════════════════════════

## test: run all tests
test: test-go

## test-go: run Go unit tests
test-go:
	cd $(APP_DIR) && go test ./... -v -timeout 30s

## test-go-short: run fast Go tests only
test-go-short:
	cd $(APP_DIR) && go test ./... -short -timeout 10s

## lint-go: run Go linter
lint-go:
	cd $(APP_DIR) && go vet ./...

## lint-web: run TypeScript type-check
lint-web:
	cd $(APP_DIR)/web && pnpm exec tsc --noEmit

# ══════════════════════════════════════════════════════════════════════════
## Docker
# ══════════════════════════════════════════════════════════════════════════

## docker-build: build the Docker image
docker-build:
	docker build -f docker/Dockerfile.backend -t policyflow:latest .

## docker-up: start the stack with docker-compose
docker-up:
	docker compose -f docker/docker-compose.yml up -d
	@echo "PolicyFlow running at http://localhost:8080"

## docker-down: stop the stack
docker-down:
	docker compose -f docker/docker-compose.yml down

## docker-logs: tail container logs
docker-logs:
	docker compose -f docker/docker-compose.yml logs -f

## docker-shell: open a shell in the running container
docker-shell:
	docker compose -f docker/docker-compose.yml exec policyflow sh

# ══════════════════════════════════════════════════════════════════════════
## Database
# ══════════════════════════════════════════════════════════════════════════

## db-reset: wipe the local database (will re-seed on next run)
db-reset:
	rm -f $(APP_DIR)/policyflow.db $(APP_DIR)/policyflow.db-wal $(APP_DIR)/policyflow.db-shm
	@echo "✓ Database removed — will be re-seeded on next startup"

# ══════════════════════════════════════════════════════════════════════════
## Cleanup
# ══════════════════════════════════════════════════════════════════════════

## clean: remove build artifacts and Next.js output
clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(APP_DIR)/web/.next $(APP_DIR)/web/out
	@mkdir -p $(APP_DIR)/web/out
	@echo '<!DOCTYPE html><html><head><title>PolicyFlow</title></head><body>Run make build-app</body></html>' \
	  > $(APP_DIR)/web/out/index.html
	@echo "✓ Clean"

## clean-all: also remove node_modules and Go cache
clean-all: clean
	rm -rf $(APP_DIR)/web/node_modules $(DOCS_DIR)/node_modules
	cd $(APP_DIR) && go clean -cache

# ══════════════════════════════════════════════════════════════════════════
## Help
# ══════════════════════════════════════════════════════════════════════════

## help: list all targets with descriptions
help:
	@echo ""
	@echo "PolicyFlow — available make targets:"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/^## /  /'
	@echo ""
