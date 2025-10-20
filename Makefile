ifneq (,$(wildcard .env))
    include .env
    export $(shell sed 's/=.*//' .env)
endif


migrate-chat-db:
	@sleep 1
	@goose postgres "user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${CHAT_POSTGRES_DB} host=127.0.0.1 port=5432 sslmode=disable" up -dir ./chat/migrations

.up-chat-db:
	@docker-compose up chat-db -d

.up-chat-service:
	@docker-compose up chat --build -d

.up-auth-service:
	@docker-compose up auth --build -d

.up-users-service:
	@docker-compose up users --build -d

migrate-social-db:
	@sleep 1
	@goose postgres "user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${SOCIAL_POSTGRES_DB} host=127.0.0.1 port=5433 sslmode=disable" up -dir ./social/migrations

.up-social-db:
	@docker-compose up social-db -d

.up-social-service:
	@docker-compose up social --build -d

.up-gateway-service:
	@docker-compose up gateway --build -d

up-chat: .up-chat-db migrate-chat-db .up-chat-service

up-auth: .up-auth-service

up-users: .up-users-service

up-social:  .up-social-db migrate-social-db .up-social-service

up-gateway: .up-gateway-service

up: up-chat up-auth up-users up-social up-gateway

down:
	@docker-compose down

.PHONY:
	migrate-chat-db \
	.up-chat-db \
	.up-chat-service \
	up-chat \
	up \
	down