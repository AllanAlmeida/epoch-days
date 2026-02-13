# Epoch Days API

API REST em Go para calcular epoch atual e epoch futuro com base em uma quantidade de dias.

## Estrutura do projeto

```text
.
├── api
│   ├── integration_test.go
│   ├── openapi.yaml
│   └── swagger.yaml
├── cmd
│   └── app
│       └── main.go
├── internal
│   └── handlers
│       ├── handlers.go
│       └── handlers_test.go
├── .gitignore
├── Dockerfile
├── go.mod
└── README.md
```

## Endpoint

### `GET /epoch/{days}`

- `days`: inteiro na URL (pode ser negativo)
- Retorno `200`:

```json
{
  "now_epoch": 1700000000,
  "future_epoch": 1700172800,
  "days_added": 2
}
```

### Erro de validação

Se `days` não for inteiro, retorna `400`:

```json
{
  "error": "invalid days: must be an integer"
}
```

## Swagger (OpenAPI)

- Especificação principal: `api/swagger.yaml`
- Versão JSON: `api/swagger.json`
- Especificação inicial mantida: `api/openapi.yaml`
- Endpoint da especificação em runtime: `GET /epoch/swagger`

## Como rodar localmente (Go)

### Pré-requisitos

- Go `1.25+`

### Comandos

```bash
go run ./cmd/app
```

Servidor sobe por padrão em `http://localhost:8080`.

Você pode definir porta customizada:

```bash
PORT=9090 go run ./cmd/app
```

### Teste rápido de chamada

```bash
curl http://localhost:8080/epoch/5
curl http://localhost:8080/epoch/abc
```

## Como testar

```bash
go test ./...
```

## Docker

### Gerar imagem

```bash
docker build -t epoch-days:latest .
```

### Executar container

```bash
docker run --rm -p 8080:8080 epoch-days:latest
```

### Testar endpoint no container

```bash
curl http://localhost:8080/epoch/10
```

## Deploy no Render

1. Suba o projeto para um repositório Git (GitHub/GitLab).
2. No Render, clique em **New +** > **Web Service**.
3. Conecte o repositório.
4. Configure:
- Runtime: `Go`
- Build Command: `go build -o app ./cmd/app`
- Start Command: `./app`
5. O Render injeta `PORT` automaticamente; a aplicação já usa essa variável.
6. Finalize o deploy e teste:

```bash
curl https://SEU-SERVICO.onrender.com/epoch/3
```

## Deploy no Railway

### Opção 1: Deploy com Dockerfile (recomendado)

1. Faça push do projeto para GitHub.
2. No Railway, clique em **New Project** > **Deploy from GitHub repo**.
3. Selecione o repositório.
4. Railway detecta o `Dockerfile` e faz build/deploy automaticamente.
5. Após deploy, teste:

```bash
curl https://SEU-PROJETO.up.railway.app/epoch/3
```

### Opção 2: Railway CLI

Pré-requisitos: Node.js + Railway CLI.

```bash
npm i -g @railway/cli
railway login
railway init
railway up
```

## Boas práticas aplicadas

- `net/http` com roteamento por padrão Go (`GET /epoch/{days}`).
- Uso de `context.Context` no handler para respeitar cancelamento/timeout.
- Tratamento explícito de erros com respostas JSON consistentes.
- Testes unitários e teste de integração com `httptest`.
