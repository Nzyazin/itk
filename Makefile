include config.env
export

DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

MIGRATIONS_DIR := migrations

.PHONY: migrate-up
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

.PHONY: migrate-down
migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

.PHONY: test-repo
test-repo:
	@echo "Cleaning port 5433 if in use..."
	@PID=$$(sudo lsof -ti tcp:5433) && \
		if [ -n "$$PID" ]; then \
			echo "Killing process on port 5433 (PID: $$PID)"; \
			sudo kill -9 $$PID; \
			sleep 1; \
		fi || true
	@docker rm -f test-postgres 2>/dev/null || true
	docker run -d \
	  --name test-postgres \
	  -e POSTGRES_USER=test \
	  -e POSTGRES_PASSWORD=test \
	  -e POSTGRES_DB=test_db \
	  -p 5433:5432 \
	  postgres:13
	@sleep 6
	@echo "Using port: $$(docker port test-postgres 5432 | cut -d: -f2)"
	@PG_PORT=$$(docker port test-postgres 5432 | cut -d: -f2) && \
	  migrate -path $(MIGRATIONS_DIR) -database "postgres://test:test@localhost:$$PG_PORT/test_db?sslmode=disable" up && \
	  DB_HOST=localhost DB_PORT=$$PG_PORT go test -v ./internal/core/repository/postgres -count=1
	docker stop test-postgres
