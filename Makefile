APP_NAME ?= epoch-days
CMD_PATH ?= ./cmd/app
DOCKER_IMAGE ?= $(APP_NAME):latest
PORT ?= 8080

.PHONY: help run fmt test test-race build tidy docker-build docker-run clean

help: ## Lista os comandos disponíveis
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*##/ {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Executa a API localmente
	PORT=$(PORT) go run $(CMD_PATH)

fmt: ## Formata o código Go
	gofmt -w $$(find . -type f -name '*.go' -not -path './vendor/*')

test: ## Roda todos os testes
	go test ./...

test-race: ## Roda testes com detector de race
	go test -race ./...

build: ## Gera binário local em ./bin
	mkdir -p bin
	go build -o bin/$(APP_NAME) $(CMD_PATH)

tidy: ## Organiza dependências (go.mod/go.sum)
	go mod tidy

docker-build: ## Gera a imagem Docker
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Sobe container Docker em localhost:8080
	docker run --rm -p 8080:8080 $(DOCKER_IMAGE)

clean: ## Remove artefatos locais de build/teste
	rm -rf bin
	rm -f coverage.out coverage.html
