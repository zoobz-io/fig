.PHONY: test lint lint-fix coverage clean check ci help test-unit test-integration test-bench install-hooks install-tools tidy sync test-provider

## Testing
test:              ## Run all tests (workspace-aware)
	go test -race -tags testing ./... ./awssm/... ./gcpsm/... ./vault/...

test-unit:         ## Run unit tests only (workspace-aware)
	go test -race -tags testing -short ./... ./awssm/... ./gcpsm/... ./vault/...

test-integration:  ## Run integration tests
	go test -race -tags testing ./testing/integration/...

test-bench:        ## Run benchmarks
	go test -tags testing -bench=. -benchmem ./testing/benchmarks/...

## Linting
lint:              ## Run linter
	golangci-lint run

lint-fix:          ## Run linter with auto-fix
	golangci-lint run --fix

## Coverage
coverage:          ## Generate coverage report (unit + integration, workspace-aware)
	@go test -tags testing -coverprofile=coverage-unit.out -covermode=atomic ./... ./awssm/... ./gcpsm/... ./vault/...
	@go test -tags testing -coverprofile=coverage-integration.out -covermode=atomic ./testing/integration/... 2>/dev/null || true
	@echo "mode: atomic" > coverage.out
	@tail -n +2 coverage-unit.out >> coverage.out
	@tail -n +2 coverage-integration.out >> coverage.out 2>/dev/null || true
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | tail -1
	@echo "Coverage report: coverage.html"

## Tooling
install-hooks:     ## Install git hooks
	@echo "#!/bin/sh" > .git/hooks/pre-commit
	@echo "make check" >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Pre-commit hook installed"

install-tools:     ## Install development tools
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.7.2

## Workspace
tidy:              ## Tidy all modules
	@go mod tidy
	@for dir in awssm gcpsm vault testing/integration; do (cd $$dir && go mod tidy); done

sync:              ## Sync workspace
	@go work sync

test-provider:     ## Run tests for a specific provider (PROVIDER=awssm|gcpsm|vault)
	@go test -race -v ./$(PROVIDER)/...

## Maintenance
clean:             ## Remove generated files
	rm -f coverage.out coverage.html coverage-unit.out coverage-integration.out

## Workflow
check:             ## Quick validation (test + lint)
	$(MAKE) test
	$(MAKE) lint

ci:                ## Full CI simulation
	$(MAKE) clean
	$(MAKE) lint
	$(MAKE) test
	$(MAKE) coverage

## Help
help:              ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
