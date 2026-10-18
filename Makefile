SHELL := /bin/bash
.DEFAULT_GOAL := help

BACKEND := backend
IOS_PACKAGE := clients/ios/Packages/NuvroKit
ANDROID := clients/android

.PHONY: help verify loc comments go-deps backend-format backend-format-check \
	backend-test backend-race backend-lint backend-vuln backend-tidy \
	run-api run-worker ios-test android-test local-up local-down

help:
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | sed -E 's/:.*## /\t/' | expand -t 24

verify: ## Run every quality gate available in this environment
	@./scripts/verify.sh

loc: ## Fail on any handwritten source file above 400 lines
	@./scripts/check-loc.sh

comments: ## Fail on explanatory comments in handwritten runtime source
	@./scripts/check-comments.sh

go-deps: ## Fail on forbidden Go dependency directions
	@./scripts/check-go-deps.sh

backend-format: ## Format Go sources in place
	@cd $(BACKEND) && gofmt -w .

backend-format-check: ## Fail when Go sources are not gofmt clean
	@cd $(BACKEND) && test -z "$$(gofmt -l .)" || { echo "gofmt: files need formatting:"; gofmt -l .; exit 1; }

backend-test: ## Run backend unit and handler tests
	@cd $(BACKEND) && go vet ./... && go test -timeout 120s ./...

backend-race: ## Run backend tests under the race detector
	@cd $(BACKEND) && go test -race -timeout 300s ./...

backend-lint: ## Run golangci-lint when installed
	@if command -v golangci-lint >/dev/null 2>&1; then cd $(BACKEND) && golangci-lint run; else echo "SKIPPED backend-lint: golangci-lint is not installed"; fi

backend-vuln: ## Run govulncheck when installed
	@if command -v govulncheck >/dev/null 2>&1; then cd $(BACKEND) && govulncheck ./...; else echo "SKIPPED backend-vuln: govulncheck is not installed"; fi

backend-tidy: ## Tidy and verify the backend module graph
	@cd $(BACKEND) && go mod tidy && go mod verify

run-api: ## Run the API binary locally
	@cd $(BACKEND) && go run ./cmd/api

run-worker: ## Run the worker binary locally
	@cd $(BACKEND) && go run ./cmd/worker

ios-test: ## Run the NuvroKit Swift package tests
	@test -f $(IOS_PACKAGE)/Package.swift \
		&& (cd $(IOS_PACKAGE) && swift test) \
		|| { echo "SKIPPED ios-test: $(IOS_PACKAGE)/Package.swift does not exist yet"; exit 0; }

android-test: ## Run Android unit tests once the Gradle project exists
	@test -f $(ANDROID)/settings.gradle.kts \
		&& (cd $(ANDROID) && ./gradlew testDebugUnitTest) \
		|| { echo "SKIPPED android-test: the Android Gradle project is not initialised yet"; echo "See clients/android/README.md"; exit 0; }

local-up: ## Start local PostgreSQL and Redis
	@docker compose -f infra/local/compose.yaml up -d

local-down: ## Stop local PostgreSQL and Redis
	@docker compose -f infra/local/compose.yaml down

.PHONY: api-generate api-check api-validate sql-generate sql-check db-migrate backend-integration
api-generate: ## Generate the strict Chi server from released OpenAPI
	@./scripts/api-generate.sh
ios-contract: ## Check the bundled OpenAPI document the iOS client is generated from
	./scripts/ios-contract-sync.sh check

ios-contract-sync: ## Refresh the bundled OpenAPI document the iOS client is generated from
	./scripts/ios-contract-sync.sh

api-check: ## Check committed generated API output without changing it
	@./scripts/api-generate.sh check
api-validate: ## Validate OpenAPI 3.1 and resolve all released references
	@cd backend && go run ./cmd/contract
sql-generate: ## Generate PostgreSQL queries using sqlc v1.31.1
	@./scripts/sql-generate.sh
sql-check: ## Check committed sqlc output without changing it
	@./scripts/sql-generate.sh check
db-migrate: ## Apply Goose migrations to the configured database
	@cd backend && go run ./cmd/migrate
backend-integration: ## Exercise real PostgreSQL and Redis with explicit test URLs
	@cd backend && go test -tags=integration -count=1 ./tests/integration/...
