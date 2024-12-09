DOCKER_COMPOSE_FILE=docker-compose.yml

## up: starts all containers in the background without forcing build
up:
	@echo "Starting Docker containers..."
	docker-compose up -d
	@echo "Docker containers started!"

## up_build: stops docker-compose (if running), builds all images and starts docker compose
up_build:
	@echo "Stopping Docker containers (if running)..."
	docker-compose down
	@echo "Building and starting Docker containers..."
	docker-compose up --build -d
	@echo "Docker containers built and started!"

## down: stops all containers
down:
	@echo "Stopping Docker containers..."
	docker-compose down
	@echo "Done!"

## logs: shows logs from all services
logs:
	@echo "Fetching logs from all services..."
	docker-compose logs -f