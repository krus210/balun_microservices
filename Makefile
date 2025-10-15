ifneq (,$(wildcard .env))
    include .env
    export $(shell sed 's/=.*//' .env)
endif


migrate-chat-db:
	@sleep 1
	@goose postgres "user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${CHAT_POSTGRES_DB} host=127.0.0.1 port=5432 sslmode=disable" up -dir ./chat/internal/app/migrations

.up-chat-db:
	@docker-compose up chat-db -d

.up-chat-service:
	@docker-compose up chat --build -d

up-chat: .up-chat-db migrate-chat-db .up-chat-service

down:
	@docker-compose down

.PHONY:
	migrate-chat-db \
	.up-chat-db \
	.up-chat-service \
	up-chat \
	down