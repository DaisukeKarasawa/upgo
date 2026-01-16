.PHONY: dev dev-backend dev-frontend run build test clean backup bench bench-verbose perf-test test-all \
        legacy-dev legacy-backend legacy-frontend skillgen

# New CLI tool commands
build:
	@echo "Building skillgen CLI..."
	@go build -o bin/upgo cmd/skillgen/main.go

skillgen:
	@go run cmd/skillgen/main.go

test:
	@go test ./...

# Run benchmark tests
bench:
	@echo "Running benchmark tests..."
	@go test -bench=. -benchmem ./...

# Run benchmark tests with verbose output
bench-verbose:
	@echo "Running benchmark tests with verbose output..."
	@go test -bench=. -benchmem -v ./...

# Run performance tests (includes detailed metrics)
perf-test:
	@echo "Running performance tests..."
	@go test -v -run TestPerformance ./...

# Run all tests including benchmarks and performance tests
test-all:
	@echo "Running all tests..."
	@go test -v -bench=. -benchmem ./...

clean:
	@rm -rf bin/ dist/ legacy/web/dist/

# Legacy Web UI commands
legacy-dev:
	@echo "Starting legacy development servers..."
	@echo "Backend: http://localhost:8081"
	@echo "Frontend: http://localhost:5173"
	@make -j2 legacy-backend legacy-frontend

legacy-backend:
	@go run legacy/cmd/server/main.go

legacy-frontend:
	@cd legacy/web && npm run dev

legacy-run:
	@echo "Building legacy frontend..."
	@cd legacy/web && npm run build
	@echo "Starting legacy server..."
	@go run legacy/cmd/server/main.go

legacy-build:
	@echo "Building legacy backend..."
	@go build -o bin/upgo-legacy legacy/cmd/server/main.go
	@echo "Building legacy frontend..."
	@cd legacy/web && npm run build

backup:
	@curl -X POST http://localhost:8081/api/v1/backup
