.PHONY: all fmt build start migrate migrate-up migrate-down setup swagger sqlc gen-store test lint vet tidy clean

APP_NAME		= echo-admin
BUILD_ROOT		= build
CONFIG			= ./config/config.yaml

all: start

fmt:
	@find . -name "*.go" -type f -not -path "./vendor/*"|xargs gofmt -s -w

build:
	@go build -ldflags "-w -s" -o $(BUILD_ROOT)/$(APP_NAME)

start:
	@go run ./main.go runserver --config=$(CONFIG) --casbin_model=./config/casbin_model.conf

# migrate applies golang-migrate migrations (sqlc-owned tables) and then GORM
# AutoMigrate for the not-yet-migrated modules.
migrate migrate-up:
	@go run ./main.go migrate --config=$(CONFIG)

# migrate-down rolls back the most recent golang-migrate migration.
migrate-down:
	@go run ./main.go migrate --config=$(CONFIG) --down

setup:
	@go run ./main.go setup --config=$(CONFIG) --menu=./config/menu.yaml

swagger:
	@swag init --parseDependency --parseInternal -g main.go

# sqlc regenerates the per-engine type-safe packages from db/queries + db/migrations.
# Install once: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
# Always follow with `make gen-store` so the neutral store layer stays in sync.
sqlc:
	@sqlc generate

# gen-store regenerates the engine-neutral data layer (db/store interface +
# neutral models) and the per-engine adapters (db/pgstore, db/mysqlstore,
# db/sqlitestore) from the sqlc output. Run after every `make sqlc`.
gen-store:
	@go run ./tools/gen-store

test:
	@go test ./...

vet:
	@go vet ./...

# lint requires golangci-lint (https://golangci-lint.run/usage/install/)
lint:
	@golangci-lint run ./...

tidy:
	@go mod tidy

clean:
	@rm -rf $(BUILD_ROOT)
