LOGGER_BINARY=logger-service-binary

## up: starts all containers in the background without forcing build
up:
	@echo "Starting Docker containers..."
	docker-compose up -d
	@echo "Docker containers started!"

## up_build: stops docker-compose (if running), builds the logger service and starts docker-compose
up_build: build_logger
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

## build_logger: builds the logger binary as a Linux executable
build_logger:
	@echo "Building logger binary..."
	cd ./src/logger-service && env GOOS=linux CGO_ENABLED=0 go build -o ${LOGGER_BINARY} .
	@echo "Logger binary built!"

## run_logger: runs the logger service locally
run_logger:
	@echo "Running logger service locally..."
	cd ./src/logger-service && go run main.go
	@echo "Logger service is running!"

## test_logger: runs tests for the logger service
test_logger:
	@echo "Running tests for logger service..."
	cd ./src/logger-service && go test ./...
	@echo "Logger service tests completed!"
