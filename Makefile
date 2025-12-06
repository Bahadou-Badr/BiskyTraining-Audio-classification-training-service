.PHONY: up build run test clean

up:
	docker-compose -f deploy/docker-compose.yml up --build

build:
	docker build -t audioml-api:local .

run:
	go run ./cmd/api

trainer-build:
	docker build -t audioml-trainer:local ./trainer

down:
	docker-compose -f deploy/docker-compose.yml down -v
