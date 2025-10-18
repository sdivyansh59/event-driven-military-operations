.PHONY: wire run build clean test help migrate-init migrate-up migrate-down migrate-status migrate-create

# Generate wire dependency injection code
wire:
	@echo "Generating wire dependencies..."
	@wire ./app

# Run the application
run:
	@echo "Running application..."
	@go run main.go

# Build the application
build:
	@echo "Building application..."
	@go build -o bin/api-service main.go

# Clean generated files
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f app/wire_gen.go

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Database Migrations
migrate-init:
	@echo "Initializing migration tables..."
	@cd cmd/migrate && go run main.go init

migrate-up:
	@echo "Running migrations..."
	@cd cmd/migrate && go run main.go migrate

migrate-down:
	@echo "Rolling back last migration..."
	@cd cmd/migrate && go run main.go rollback

migrate-status:
	@echo "Checking migration status..."
	@cd cmd/migrate && go run main.go status

migrate-create:
	@echo "Creating new migration..."
	@read -p "Enter migration name: " name; \
	cd cmd/migrate && go run main.go create $$name

# Help command
help:
	@echo "Available commands:"
	@echo "  make wire           - Generate wire dependency injection code"
	@echo "  make run            - Run the application"
	@echo "  make build          - Build the application binary"
	@echo "  make clean          - Clean generated files and binaries"
	@echo "  make test           - Run tests"
	@echo "  make deps           - Install and tidy dependencies"
	@echo ""
	@echo "Database Migrations:"
	@echo "  make migrate-init   - Initialize migration tables in database"
	@echo "  make migrate-up     - Run all pending migrations"
	@echo "  make migrate-down   - Rollback last migration group"
	@echo "  make migrate-status - Show migration status"
	@echo "  make migrate-create - Create a new migration file"
	@echo ""
	@echo "  make help           - Show this help message"
