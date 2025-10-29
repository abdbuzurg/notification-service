DB_URL := "postgres://postgres:password@localhost:5432/notifications_db?sslmode=disable"

.PHONY: sqlc
sqlc:
	@echo ">> Generating sqlc code..."
	sqlc generate

.PHONY: proto
proto:
	@echo ">> Generating protobuf code..."
	protoc --go_out=. --go-grpc_out=. protos/notification.proto

.PHONY: migrate-create
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir internal/platform/database/migrations -seq $$name

.PHONY: migrate-up
migrate-up:
	@echo ">> Running migrations up..."
	migrate -database "$(DB_URL)" -path internal/platform/database/migrations up
