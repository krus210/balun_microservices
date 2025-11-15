ifneq (,$(wildcard .env))
    include .env
    export $(shell sed 's/=.*//' .env)
endif

# Vault must come up first so we can seed credentials before dependent services start.
.up-vault:
	@docker-compose up vault -d

.up-kafka:
	@docker-compose up kafka -d --wait
	@docker-compose up kafka-ui -d

.up-jaeger:
	@docker-compose up jaeger -d --wait

.up-graylog-stack:
	@docker-compose up mongodb -d --wait
	@docker-compose up elasticsearch -d --wait
	@docker-compose up graylog -d --wait
	@docker-compose up filebeat -d

migrate-chat-db:
	@goose postgres "user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${CHAT_POSTGRES_DB} host=127.0.0.1 port=5432 sslmode=disable" up -dir ./chat/migrations

.up-chat-db:
	@docker-compose up chat-db -d --wait

.up-chat-service:
	@docker-compose up chat --build -d

.up-auth-service:
	@docker-compose up auth --build -d

.up-users-db:
	@docker-compose up users-db -d --wait

migrate-users-db:
	@goose postgres "user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${USERS_POSTGRES_DB} host=127.0.0.1 port=5434 sslmode=disable" up -dir ./users/migrations

.up-users-service:
	@docker-compose up users --build -d

migrate-social-db:
	@goose postgres "user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${SOCIAL_POSTGRES_DB} host=127.0.0.1 port=5433 sslmode=disable" up -dir ./social/migrations

.up-social-db:
	@docker-compose up social-db -d --wait

.up-social-service:
	@docker-compose up social --build -d

.up-gateway-service:
	@docker-compose up gateway --build -d

migrate-notifications-db:
	@goose postgres "user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${NOTIFICATIONS_POSTGRES_DB} host=127.0.0.1 port=5429 sslmode=disable" up -dir ./notifications/migrations

.up-notifications-db:
	@docker-compose up notifications-db -d --wait

.up-notifications-service:
	@docker-compose up notifications --build -d

up-chat: .up-chat-db migrate-chat-db .up-chat-service

up-auth: .up-auth-service

up-users: .up-users-db migrate-users-db .up-users-service

up-social:  .up-social-db migrate-social-db .up-social-service

up-gateway: .up-gateway-service

up-notifications: .up-notifications-db migrate-notifications-db .up-notifications-service

up: .up-graylog-stack .up-jaeger .up-kafka up-chat up-auth up-users up-social up-notifications up-gateway

down:
	@docker-compose down

.PHONY:
	migrate-chat-db \
	.up-chat-db \
	.up-chat-service \
	up-chat \
	up \
	down
