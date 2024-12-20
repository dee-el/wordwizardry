# Makefile

.PHONY: build
build:
	docker-compose build

.PHONY: up
up:
	docker-compose up -d

.PHONY: down
down:
	docker-compose down

.PHONY: logs
logs:
	docker-compose logs -f

.PHONY: restart
restart:
	docker-compose restart

.PHONY: clean
clean:
	docker-compose down -v

.PHONY: rebuild
rebuild:
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d

# Development commands
.PHONY: dev
dev:
	go run main.go

.PHONY: test
test:
	go test -v ./...

# Docker cleanup commands
.PHONY: docker-clean
docker-clean:
	docker system prune -f

.PHONY: docker-clean-all
docker-clean-all:
	docker system prune -a -f --volumes

.PHONY: docker-clean-unused
docker-clean-unused:
	docker image prune -a -f
	docker container prune -f
	docker volume prune -f
	docker network prune -f

.PHONY: docker-clean-dangling
docker-clean-dangling:
	docker image prune -f


# Redis commands
.PHONY: redis-up
redis-up:
	docker-compose up -d redis

.PHONY: redis-down
redis-down:
	docker-compose stop redis && docker-compose rm -f redis
