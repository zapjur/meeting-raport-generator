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

up_build_clean: clean up_build

up_gpu:
	@echo "Stopping Docker containers (if running)..."
	docker-compose down
	@echo "Building and starting Docker containers..."
	docker-compose -f docker-compose.yml up --build -d --gpus all
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

init-mongo:
	@echo "Initializing MongoDB..."
	docker exec -it mongodb mongoimport --db database --collection transcriptions --file /docker-entrypoint-initdb.d/transcriptions.json --jsonArray -u admin -p password --authenticationDatabase admin
	@echo "MongoDB initialized!"

## clean: cleans embedding, transcription, and summary collections
## clean: cleans embedding, transcription, and summary collections
clean:
	@echo "Cleaning MongoDB collections: embeddings, transcriptions, and summaries..."
	docker exec -it mongodb mongosh database -u admin -p password --authenticationDatabase admin --eval "db.embeddings.deleteMany({})"
	docker exec -it mongodb mongosh database -u admin -p password --authenticationDatabase admin --eval "db.transcriptions.deleteMany({})"
	docker exec -it mongodb mongosh database -u admin -p password --authenticationDatabase admin --eval "db.summaries.deleteMany({})"
	@echo "MongoDB collections cleaned!"



