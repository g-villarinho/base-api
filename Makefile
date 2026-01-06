VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

setup: ## Instala bibliotecas necessárias do projeto
	@go install github.com/vektra/mockery/v2@v2.53.4
	@go install github.com/air-verse/air@v1.63.4
	@go install github.com/swaggo/swag/cmd/swag@v1.16.4
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0
	@go install gotest.tools/gotestsum@v1.13.0

run: build ## Roda o servidor com .env padrão
	@./bin/api

swagger: ## Generate Swagger/OpenAPI documentation
	@swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
	@sed -i '' 's/github_com_g-villarinho_voxel-api_internal_api_model\.//g' docs/swagger.json

docs: swagger ## Alias for swagger generation

build:  ## Build includes version injection
	@echo "Building with version: $(VERSION)"
	@go build -ldflags "-X main.Version=$(VERSION)" -o bin/api cmd/api/main.go

test: ## Executa todos os testes
	@gotestsum --format pkgname --format-hide-empty-pkg -- ./...

mocks: ## Gera mock de services, repositories e commons
	@mockery

sqlc: ## Gera código SQLC a partir das queries SQL
	@sqlc generate

migrate: ## Aplica todas as migrations pendentes
	@go run cmd/migrate/main.go --direction=up

migrate-down: ## Reverte a última migration aplicada
	@go run cmd/migrate/main.go --direction=down

migrate-status: ## Mostra status das migrations
	@go run cmd/migrate/main.go --direction=status
