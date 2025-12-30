.PHONY: dev run build test clean backup

dev:
	@echo "Starting development server..."
	@go run cmd/server/main.go

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

clean:
	@rm -rf bin/ dist/ web/dist/

backup:
	@curl -X POST http://localhost:8080/api/v1/backup
