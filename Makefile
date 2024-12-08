DOCKER_COMPOSE_FILE = docker-compose.yml

up:
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d

down:
	docker-compose -f $(DOCKER_COMPOSE_FILE) down

restart: down up

logs:
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

build:
	docker-compose -f $(DOCKER_COMPOSE_FILE) build