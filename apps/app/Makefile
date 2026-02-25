.PHONY: all dev dev-go dev-web build build-web build-go clean seed-reset

BINARY := policyflow
WEB_DIR := web
OUT_DIR := web/out

## all: build everything
all: build

## build: build frontend then Go binary
build: build-web build-go

## build-web: install deps and export Next.js static files
build-web:
	cd $(WEB_DIR) && pnpm install && pnpm build

## build-go: compile the Go binary (embeds web/out)
build-go:
	go build -ldflags="-s -w" -o $(BINARY) .

## dev: run backend with dev proxy to Next.js dev server
## Usage: run `make dev-web` in one terminal, `make dev-go` in another
dev-go:
	WEB_DEV_PROXY=http://localhost:3001 go run .

dev-web:
	cd $(WEB_DIR) && pnpm install && pnpm dev

## run: run the compiled binary
run: build
	./$(BINARY)

## clean: remove compiled artifacts
clean:
	rm -f $(BINARY)
	rm -rf $(OUT_DIR)
	mkdir -p $(OUT_DIR)
	@echo '<!DOCTYPE html><html><head><title>PolicyFlow</title></head><body>Run make build-web</body></html>' > $(OUT_DIR)/index.html

## seed-reset: delete the database to force re-seeding on next run
seed-reset:
	rm -f policyflow.db
	@echo "Database removed — will be re-seeded on next run"

## help: list targets
help:
	@grep -E '^## ' Makefile | sed 's/## //'
