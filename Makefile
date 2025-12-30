.PHONY: dev dev-backend dev-frontend run build test clean backup bench bench-verbose perf-test test-all

dev:
	@echo "Starting development servers..."
	@echo "Backend: http://localhost:8080"
	@echo "Frontend: http://localhost:5173"
	@make -j2 dev-backend dev-frontend

dev-backend:
	@go run cmd/server/main.go

dev-frontend:
	@cd web && npm run dev

run:
	@echo "Building frontend..."
	@cd web && npm run build
	@echo "Starting server..."
	@go run cmd/server/main.go

build:
	@echo "Building backend..."
	@go build -o bin/upgo cmd/server/main.go
	@echo "Building frontend..."
	@cd web && npm run build

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
	@rm -rf bin/ dist/ web/dist/

backup:
	@curl -X POST http://localhost:8080/api/v1/backup
