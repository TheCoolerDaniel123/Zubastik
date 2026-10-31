.DEFAULT_GOAL := help

# Go packages of every module in the go.work workspace, except e2e
# (e2e needs a running docker compose stack, see `make e2e`).
UNIT_PKGS := ./shared/... ./auth-service/... ./user-service/... ./apigateway-service/...

.PHONY: help up down clean logs ps build vet fmt test e2e

help: ## Show this help
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "  \033[36m%-8s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## ── Docker ──────────────────────────────────────────────────────────────────

up: ## Build and start the whole stack in the background
	docker compose up --build -d

down: ## Stop the stack (data is kept)
	docker compose down

clean: ## Stop the stack and delete all volumes (databases, Redis)
	docker compose down -v

logs: ## Follow logs of all containers
	docker compose logs -f

ps: ## Show container status
	docker compose ps

## ── Go ──────────────────────────────────────────────────────────────────────

build: ## Compile all modules
	go build $(UNIT_PKGS)

vet: ## Run go vet on all modules
	go vet $(UNIT_PKGS)

fmt: ## Format the code with gofmt
	gofmt -w .

test: ## Run unit tests of all modules
	go test -count=1 $(UNIT_PKGS)

e2e: ## Run end-to-end tests (the stack must be running: make up)
	go test -count=1 -v ./e2e/...
