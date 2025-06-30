.PHONY: all
all: help tidy lint test test/cover

.PHONY: help
help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: tidy
tidy: tidy/tools ## Tidy
	go mod tidy -v
	go mod verify
	go fmt ./... 
	go vet ./...

.PHONY: tidy/tools
tidy/tools: ## Tidy dev tools
	cd tools && go mod tidy -v
	cd tools && go mod verify

.PHONY: lint
lint: ## Lint
	golangci-lint run -v -c .golangci.yaml

.PHONY: test
test: ## Test
	go test -race -buildvcs ./...

.PHONY: test/cover
test/cover: ## Test and cover
	go test -race -buildvcs -coverprofile=cover.out ./...
	go tool cover -html=cover.out

.PHONY: sqlc/generate
sqlc/generate: ## Generate sqlc
	go tool -modfile tools/go.mod sqlc generate

.PHONY: migrate/up
migrate/up: ## Run UP migrations
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest -path migration -database ${POSTGRESQL_URL} up

.PHONY: migrate/down
migrate/down: ## Run DOWN migrations
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest -path migration -database ${POSTGRESQL_URL} down