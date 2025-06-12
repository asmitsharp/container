.PHONY: dev build test clean run

# Development commands
dev:
	docker compose -f docker-compose.dev.yml up --build

dev-shell:
	docker compose -f docker-compose.dev.yml run --rm dev /bin/bash

# Build the container runtime
build:
	docker compose -f docker-compose.dev.yml run --rm dev go build -o bin/container-runtime ./cmd/main.go

# Run tests
test:
	docker compose -f docker-compose.dev.yml run --rm dev go test ./...

# Clean up
clean:
	docker compose -f docker-compose.dev.yml down
	docker system prune -f

# Run the container runtime
run:
	docker compose -f docker-compose.dev.yml run --rm --privileged dev ./bin/container-runtime

# Format code
fmt:
	docker compose -f docker-compose.dev.yml run --rm dev go fmt ./...

# Install dependencies
deps:
	docker compose -f docker-compose.dev.yml run --rm dev go mod tidy