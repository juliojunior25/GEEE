.PHONY: help dev build test test-race bench lint fmt clean clean-temp plugin-new plugin-test plugin-bench metrics trace profile docker-build docker-run server server-build server-build-prod server-dev server-run

# Cores para output
GREEN  := \033[0;32m
YELLOW := \033[0;33m
BLUE   := \033[0;34m
RED    := \033[0;31m
NC     := \033[0m # No Color

# Variáveis
BINARY_NAME=geee
SERVER_BINARY_NAME=geee-server
BUILD_DIR=./bin
MAIN_PATH=./cmd/geee
SERVER_MAIN_PATH=./cmd/geee-server
COVERAGE_DIR=./coverage

# Flags de build
LDFLAGS=-ldflags "-s -w -X main.Version=$(shell git describe --tags --always --dirty) -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"
DEV_FLAGS=-race -gcflags="all=-N -l"

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

help: ## Mostra esta mensagem de ajuda
	@echo "$(BLUE)GEEE - Generic Extraction & Enrichment Engine$(NC)"
	@echo ""
	@echo "$(GREEN)Comandos disponíveis:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""

##@ Desenvolvimento

dev: ## Inicia modo desenvolvimento com hot reload
	@echo "$(GREEN)🚀 Starting development mode...$(NC)"
	@if command -v air > /dev/null; then \
		air -c .air.toml; \
	else \
		echo "$(YELLOW)⚠️  'air' não instalado. Rodando sem hot reload...$(NC)"; \
		echo "$(YELLOW)💡 Instale com: go install github.com/air-verse/air@latest$(NC)"; \
		$(GORUN) $(DEV_FLAGS) $(MAIN_PATH); \
	fi

build: ## Build de produção (otimizado)
	@echo "$(GREEN)🔨 Building production binary...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)

build-local: ## Build para sistema operacional local
	@echo "$(GREEN)🔨 Building local binary...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

run: build-local ## Build e executa o binário localmente
	@echo "$(GREEN)▶️  Running $(BINARY_NAME)...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME)

##@ Server HTTP

server: server-build ## Build e inicia o servidor HTTP
	@echo "$(GREEN)🚀 Starting server...$(NC)"
	@$(BUILD_DIR)/$(SERVER_BINARY_NAME)

server-build: ## Build do servidor HTTP
	@echo "$(GREEN)🔨 Building server binary...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(SERVER_BINARY_NAME) $(SERVER_MAIN_PATH)
	@echo "$(GREEN)✅ Server build complete: $(BUILD_DIR)/$(SERVER_BINARY_NAME)$(NC)"

server-build-prod: ## Build de produção do servidor (Linux AMD64)
	@echo "$(GREEN)🔨 Building production server binary...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(SERVER_BINARY_NAME) $(SERVER_MAIN_PATH)
	@echo "$(GREEN)✅ Production server build complete: $(BUILD_DIR)/$(SERVER_BINARY_NAME)$(NC)"
	@ls -lh $(BUILD_DIR)/$(SERVER_BINARY_NAME)

server-dev: ## Inicia servidor em modo desenvolvimento
	@echo "$(GREEN)🚀 Starting server in development mode...$(NC)"
	@$(GORUN) $(SERVER_MAIN_PATH) --verbose

server-run: ## Executa servidor já compilado (use PORT=8080 HOST=0.0.0.0)
	@echo "$(GREEN)▶️  Running server...$(NC)"
	@if [ ! -f $(BUILD_DIR)/$(SERVER_BINARY_NAME) ]; then \
		echo "$(YELLOW)⚠️  Server binary not found. Building first...$(NC)"; \
		$(MAKE) server-build; \
	fi
	@$(BUILD_DIR)/$(SERVER_BINARY_NAME)

##@ Testes

test: ## Roda todos os testes
	@echo "$(GREEN)🧪 Running tests...$(NC)"
	@$(GOTEST) -v -cover ./...

test-race: ## Testes com race detector
	@echo "$(GREEN)🏃 Running tests with race detector...$(NC)"
	@$(GOTEST) -race -v ./...

test-coverage: ## Testes com cobertura detalhada
	@echo "$(GREEN)📊 Running tests with coverage...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	@$(GOTEST) -v -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	@$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "$(GREEN)✅ Coverage report: $(COVERAGE_DIR)/coverage.html$(NC)"

bench: ## Executa benchmarks
	@echo "$(GREEN)⚡ Running benchmarks...$(NC)"
	@$(GOTEST) -bench=. -benchmem ./...

##@ Code Quality

lint: ## Executa linter
	@echo "$(GREEN)🔍 Running linter...$(NC)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "$(RED)❌ golangci-lint não instalado$(NC)"; \
		echo "$(YELLOW)💡 Instale com: brew install golangci-lint (Mac) ou veja https://golangci-lint.run/$(NC)"; \
		exit 1; \
	fi

fmt: ## Formata código
	@echo "$(GREEN)✨ Formatting code...$(NC)"
	@$(GOFMT) ./...
	@echo "$(GREEN)✅ Code formatted$(NC)"

vet: ## Executa go vet
	@echo "$(GREEN)🔎 Running go vet...$(NC)"
	@$(GOCMD) vet ./...

##@ Plugins

plugin-new: ## Cria scaffold de novo plugin (use: make plugin-new NAME=my-plugin)
	@if [ -z "$(NAME)" ]; then \
		echo "$(RED)❌ Erro: NAME é obrigatório$(NC)"; \
		echo "$(YELLOW)Uso: make plugin-new NAME=my-plugin$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)🔌 Creating plugin scaffold: $(NAME)$(NC)"
	@mkdir -p plugins/$(NAME)
	@echo "package $(NAME)\n\n// TODO: Implement plugin" > plugins/$(NAME)/plugin.go
	@echo "$(GREEN)✅ Plugin created at plugins/$(NAME)$(NC)"

plugin-test: ## Testa plugin específico (use: make plugin-test PLUGIN=my-plugin)
	@if [ -z "$(PLUGIN)" ]; then \
		echo "$(RED)❌ Erro: PLUGIN é obrigatório$(NC)"; \
		echo "$(YELLOW)Uso: make plugin-test PLUGIN=my-plugin$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)🧪 Testing plugin: $(PLUGIN)$(NC)"
	@$(GOTEST) -v ./plugins/$(PLUGIN)/...

plugin-bench: ## Benchmark de plugin (use: make plugin-bench PLUGIN=my-plugin)
	@if [ -z "$(PLUGIN)" ]; then \
		echo "$(RED)❌ Erro: PLUGIN é obrigatório$(NC)"; \
		echo "$(YELLOW)Uso: make plugin-bench PLUGIN=my-plugin$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)⚡ Benchmarking plugin: $(PLUGIN)$(NC)"
	@$(GOTEST) -bench=. -benchmem ./plugins/$(PLUGIN)/...

##@ Observabilidade

metrics: ## Exporta métricas locais
	@echo "$(GREEN)📊 Exporting metrics...$(NC)"
	@echo "$(YELLOW)⚠️  Not implemented yet$(NC)"

trace: ## Visualiza traces (Jaeger local)
	@echo "$(GREEN)🔍 Opening Jaeger UI...$(NC)"
	@echo "$(YELLOW)⚠️  Not implemented yet$(NC)"
	@echo "$(BLUE)💡 Run: docker run -d --name jaeger -p 16686:16686 -p 6831:6831/udp jaegertracing/all-in-one:latest$(NC)"

profile: ## Profiling de CPU/memória
	@echo "$(GREEN)📈 Starting profiling...$(NC)"
	@$(GOTEST) -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...
	@echo "$(GREEN)✅ Profiles generated: cpu.prof, mem.prof$(NC)"
	@echo "$(BLUE)💡 Analyze with: go tool pprof cpu.prof$(NC)"

##@ Limpeza

clean: ## Remove builds e caches
	@echo "$(GREEN)🧹 Cleaning...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -f cpu.prof mem.prof
	@$(GOCLEAN) -cache
	@echo "$(GREEN)✅ Clean complete$(NC)"

clean-temp: ## Remove arquivos temporários de execuções
	@echo "$(GREEN)🗑️  Removing temporary files...$(NC)"
	@rm -rf /tmp/geee-*
	@echo "$(GREEN)✅ Temporary files removed$(NC)"

##@ Docker

docker-build: ## Build de imagem Docker
	@echo "$(GREEN)🐳 Building Docker image...$(NC)"
	@docker build -t $(BINARY_NAME):latest .
	@echo "$(GREEN)✅ Docker image built: $(BINARY_NAME):latest$(NC)"

docker-run: ## Roda em container
	@echo "$(GREEN)🐳 Running Docker container...$(NC)"
	@docker run --rm -it $(BINARY_NAME):latest

##@ Dependências

deps: ## Instala dependências
	@echo "$(GREEN)📦 Installing dependencies...$(NC)"
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "$(GREEN)✅ Dependencies installed$(NC)"

deps-update: ## Atualiza dependências
	@echo "$(GREEN)🔄 Updating dependencies...$(NC)"
	@$(GOGET) -u ./...
	@$(GOMOD) tidy
	@echo "$(GREEN)✅ Dependencies updated$(NC)"

deps-verify: ## Verifica dependências
	@echo "$(GREEN)🔐 Verifying dependencies...$(NC)"
	@$(GOMOD) verify
	@echo "$(GREEN)✅ Dependencies verified$(NC)"

##@ CI/CD

ci-test: deps test-race lint ## Pipeline de CI completo
	@echo "$(GREEN)✅ CI tests passed$(NC)"

ci-build: deps build ## Pipeline de CI para build
	@echo "$(GREEN)✅ CI build complete$(NC)"

##@ Utilitários

check-tools: ## Verifica ferramentas necessárias
	@echo "$(GREEN)🔧 Checking required tools...$(NC)"
	@command -v go >/dev/null 2>&1 || { echo "$(RED)❌ Go não instalado$(NC)"; exit 1; }
	@command -v air >/dev/null 2>&1 || echo "$(YELLOW)⚠️  air não instalado (opcional para hot reload)$(NC)"
	@command -v golangci-lint >/dev/null 2>&1 || echo "$(YELLOW)⚠️  golangci-lint não instalado (opcional para linting)$(NC)"
	@command -v docker >/dev/null 2>&1 || echo "$(YELLOW)⚠️  docker não instalado (opcional)$(NC)"
	@echo "$(GREEN)✅ Required tools check complete$(NC)"

version: ## Mostra versão do projeto
	@echo "$(BLUE)GEEE v1.0.0$(NC)"
	@echo "Go version: $(shell go version)"
	@echo "Git commit: $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"

# Default target
.DEFAULT_GOAL := help
