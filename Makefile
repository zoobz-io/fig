.PHONY: test test-unit test-integration test-bench lint lint-fix security coverage clean help check ci install-tools install-hooks tidy sync test-provider

.DEFAULT_GOAL := help

help: ## Display available commands
	@echo "fig Development Commands"
	@echo "=============================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

test: ## Run all tests with race detector
	@go test -v -race -tags testing ./... ./awssm/... ./gcpsm/... ./vault/...

test-unit: ## Run unit tests only (short mode)
	@go test -v -race -tags testing -short ./... ./awssm/... ./gcpsm/... ./vault/...

test-integration: ## Run integration tests
	@go test -v -race -tags testing ./testing/integration/...

test-bench: ## Run benchmarks
	@go test -tags testing -bench=. -benchmem -benchtime=1s ./testing/benchmarks/...

test-provider: ## Run tests for a specific provider (PROVIDER=awssm|gcpsm|vault)
	@go test -race -tags testing -v ./$(PROVIDER)/...

lint: ## Run linters
	@golangci-lint run --config=.golangci.yml --timeout=5m

lint-fix: ## Run linters with auto-fix
	@golangci-lint run --config=.golangci.yml --fix

security: ## Run security scanner
	@gosec -quiet ./...

coverage: ## Generate coverage report (HTML)
	@go test -tags testing -coverprofile=coverage-unit.out -covermode=atomic ./... ./awssm/... ./gcpsm/... ./vault/...
	@go test -tags testing -coverprofile=coverage-integration.out -covermode=atomic ./testing/integration/... 2>/dev/null || true
	@echo "mode: atomic" > coverage.out
	@tail -n +2 coverage-unit.out >> coverage.out
	@tail -n +2 coverage-integration.out >> coverage.out 2>/dev/null || true
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | tail -1
	@echo "Coverage report: coverage.html"

clean: ## Remove generated files
	@rm -f coverage.out coverage.html coverage-unit.out coverage-integration.out coverage.txt
	@find . -name "*.test" -delete
	@find . -name "*.prof" -delete
	@find . -name "*.out" -delete

tidy: ## Tidy all modules
	@go mod tidy
	@for dir in awssm gcpsm vault testing/integration; do (cd $$dir && go mod tidy); done

sync: ## Sync workspace
	@go work sync

install-tools: ## Install development tools
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.7.2
	@go install github.com/securego/gosec/v2/cmd/gosec@latest

install-hooks: ## Install git pre-commit hook
	@mkdir -p .git/hooks
	@echo '#!/bin/sh' > .git/hooks/pre-commit
	@echo 'make check' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Pre-commit hook installed"

check: lint test security ## Run lint, tests, and security scan
	@echo "All checks passed!"

ci: clean check coverage test-bench ## Full CI simulation
	@echo "CI simulation complete!"
