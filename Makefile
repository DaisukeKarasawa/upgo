.PHONY: dev dev-backend dev-frontend run build test clean backup bench bench-verbose perf-test test-all

dev:
	@echo "Starting development servers..."
	@echo "Backend: http://localhost:8081"
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
	@curl -X POST http://localhost:8081/api/v1/backup

# Docker Compose targets
# Use 'docker compose' (without hyphen) which is the modern Docker CLI plugin
# Falls back to 'docker-compose' (with hyphen) for older installations
DOCKER_COMPOSE := $(shell command -v docker-compose 2> /dev/null || echo "docker compose")

docker-build:
	@echo "Building Docker image..."
	@$(DOCKER_COMPOSE) build

docker-up:
	@echo "Starting containers..."
	@$(DOCKER_COMPOSE) up -d

docker-down:
	@echo "Stopping containers..."
	@$(DOCKER_COMPOSE) down

docker-logs:
	@$(DOCKER_COMPOSE) logs -f

docker-restart:
	@$(DOCKER_COMPOSE) restart

docker-clean:
	@echo "Stopping and removing containers, volumes..."
	@$(DOCKER_COMPOSE) down -v
