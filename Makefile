run-up:
	go run ./cmd/server/migrate/main.go up

rollback:
	go run ./cmd/server/migrate/main.go rollback

reset-db:
	psql -U sabiroadadmin -c "DROP DATABASE IF EXISTS sabiroad;" -d postgres
	psql -U sabiroadadmin -c "CREATE DATABASE sabiroad;" -d postgres

dev:
	air

run:
	go run ./cmd/server/main.go

# Test targets
test-smoke:
	@echo "Running smoke tests..."
	@if [ ! -f tests/.env ]; then \
		echo "❌ tests/.env not found. Copy tests/.env.example and configure it first."; \
		exit 1; \
	fi
	@set -a && source tests/.env && set +a && ./scripts/smoke.sh

test-setup:
	@echo "Setting up test environment..."
	@cp tests/.env.example tests/.env
	@echo "✅ Created tests/.env from template"
	@echo "📝 Please edit tests/.env with your actual credentials"

test-deps:
	@echo "Checking test dependencies..."
	@command -v jq >/dev/null 2>&1 || { echo "❌ jq is required but not installed. Install with: brew install jq"; exit 1; }
	@command -v curl >/dev/null 2>&1 || { echo "❌ curl is required but not installed"; exit 1; }
	@echo "✅ All test dependencies are installed"

test-help:
	@echo "Available test commands:"
	@echo "  make test-setup  - Create tests/.env from template"
	@echo "  make test-deps   - Check if test dependencies are installed"
	@echo "  make test-smoke  - Run automated smoke tests"
	@echo ""
	@echo "Manual testing:"
	@echo "  - Open tests/api.http in VS Code with REST Client extension"
	@echo "  - Or import into Bruno API client"
	@echo ""
	@echo "Setup:"
	@echo "  1. make test-setup"
	@echo "  2. Edit tests/.env with your credentials"
	@echo "  3. make test-deps"
	@echo "  4. make test-smoke"