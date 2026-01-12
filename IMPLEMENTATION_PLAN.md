# Plano de Implementação - GEEE

**Generic Extraction & Enrichment Engine**

**Versão**: 1.0.0
**Data**: 2025-01-11
**Status**: Aprovado para desenvolvimento

---

## Status Geral do Projeto

| Fase | Status | Data Conclusão | Duração Estimada | Duração Real |
|------|--------|----------------|------------------|--------------|
| Fase 1: Foundation | ✅ Concluída | 2026-01-12 | 2-3 dias | 1 dia |
| Fase 2: Plugin System | ✅ Concluída | 2026-01-12 | 3-4 dias | 1 dia |
| Fase 3: CLI | ✅ Concluída | 2026-01-12 | 2-3 dias | 1 dia |
| Fase 4: Plugins de Exemplo | ✅ Concluída | 2026-01-12 | 3-4 dias | 1 dia |
| Fase 5: HTTP Server | ✅ Concluída | 2026-01-12 | 2-3 dias | 1 dia |
| Fase 6: Observabilidade | 🔜 Pendente | - | 2-3 dias | - |
| Fase 7: Quality Assurance | 🔜 Pendente | - | 2-3 dias | - |

**Progresso Geral**: 5/7 fases concluídas (71.43%)

---

## Decisões de Arquitetura

| Aspecto | Escolha | Justificativa |
|---------|---------|---------------|
| **Config** | YAML | Legível, hierárquico, padrão em projetos Go |
| **Plugin Loader** | Compile-time (imports) | Mais simples, zero runtime complexity |
| **HTTP Server** | chi | Minimalista, middleware-based, fácil de estender |
| **Formato Plugins** | Go packages | Integração nativa com Go |

---

## Estrutura de Diretórios

```
geee/
├── cmd/
│   ├── geee/main.go           # CLI entrypoint
│   └── geee-server/main.go    # HTTP server entrypoint
├── internal/
│   ├── config/                # Config loader
│   ├── core/                  # Core interfaces
│   ├── executor/              # Plan executor
│   ├── registry/              # Plugin registry
│   └── observability/         # Logs, metrics, tracing
├── plugins/
│   ├── base/                  # Base plugin struct
│   ├── json-transformer/
│   ├── regex-extractor/
│   ├── http-fetcher/
│   └── template-renderer/
├── pkg/
│   ├── errors/                # Structured errors
│   └── types/                 # Shared types
├── configs/
│   └── default.yaml           # Configuração padrão
├── test/
│   ├── fixtures/              # Test data
│   └── integration/           # Integration tests
├── bin/                       # Compiled binaries
├── Makefile
├── go.mod
└── README.md
```

---

## Fases de Desenvolvimento

### Fase 1: Foundation ✅ CONCLUÍDA
**Duração estimada**: 2-3 dias
**Duração real**: 1 dia
**Data de conclusão**: 2026-01-12

#### Entregáveis
- [x] go.mod inicializado com Go 1.21+
- [x] Estrutura de diretórios criada
- [x] Interfaces core definidas:
  - `Plugin` interface
  - `PluginRegistry` interface
  - `Executor` interface
- [x] Types básicos implementados:
  - `PluginResult`
  - `StructuredError`
  - `ExecutionContext`
  - `PluginManifest`
- [x] Makefile básico funcional
- [x] Config loader (YAML)
- [x] Tests unitários core (100% coverage)

#### Criteria de Aceite
- [x] `make build` compila sem erros
- [x] `make test` passa com coverage 100%
- [x] Interfaces documentadas com Go docs

#### Arquivos Criados
- `pkg/types/types.go` - Types básicos (PluginResult, ExecutionContext, PluginManifest, StructuredError)
- `pkg/errors/errors.go` - Funções auxiliares para criação de erros estruturados
- `pkg/errors/errors_test.go` - Testes unitários (100% coverage)
- `internal/core/interfaces.go` - Interfaces core (Plugin, PluginRegistry, Executor, ExecutionPlan)
- `internal/core/interfaces_test.go` - Testes unitários (100% coverage)
- `internal/config/loader.go` - Config loader YAML
- `internal/config/loader_test.go` - Testes unitários (100% coverage)
- `configs/default.yaml` - Configuração de exemplo
- `cmd/geee/main.go` - Entry point CLI (stub para Fase 3)

---

### Fase 2: Plugin System ✅ CONCLUÍDA
**Duração estimada**: 3-4 dias
**Duração real**: 1 dia
**Data de conclusão**: 2026-01-12

#### Entregáveis
- [x] PluginRegistry implementado
- [x] Base plugin struct (todos os plugins herdão)
- [x] Plugin registration system
- [x] Manifest parser (JSON Schema validation)
- [x] Execution DAG básico
- [x] Resource limits (memória, tempo)
- [x] Cleanup automático
- [x] Error handling estruturado

#### Criteria de Aceite
- [x] Plugins podem ser registrados via código
- [x] Manifest valida inputs/outputs
- [x] Cleanup é chamado em sucesso/erro/cancelamento
- [x] `make test-race` passa sem race conditions

#### Arquivos Criados
- `internal/registry/registry.go` - DefaultRegistry implementation com thread-safe operations
- `internal/registry/registry_test.go` - Testes unitários (100% coverage)
- `plugins/base/base.go` - BasePlugin struct com cleanup automático e validação
- `plugins/base/base_test.go` - Testes unitários para BasePlugin
- `plugins/base/resource_tracker.go` - ResourceTracker para monitorar memória e tempo
- `plugins/base/validator.go` - Validação de schemas JSON simplificada
- `internal/executor/executor.go` - DefaultExecutor com execução paralela e DAG
- `internal/executor/executor_test.go` - Testes unitários para Executor (88.5% coverage)
- `internal/executor/dag.go` - DAG para gerenciar dependências entre plugins
- `internal/executor/dag_test.go` - Testes unitários para DAG (100% coverage)
- `pkg/errors/errors.go` - Funções adicionais: NewPluginError, NewResourceLimitError

#### Funcionalidades Implementadas
- **PluginRegistry**: Thread-safe, suporta registro/remoção/listagem de plugins
- **BasePlugin**: Estrutura base com cleanup automático, validação de schemas e resource tracking
- **ResourceTracker**: Monitora uso de memória e tempo de execução
- **Validator**: Validação simplificada de JSON schemas
- **Executor**: Execução de pipelines com suporte a DAG e paralelismo
- **DAG**: Gerenciamento de dependências e execução paralela de plugins independentes
- **Error Handling**: Erros estruturados com sugestões e contexto

---

### Fase 3: CLI ✅ CONCLUÍDA
**Duração estimada**: 2-3 dias
**Duração real**: 1 dia
**Data de conclusão**: 2026-01-12

#### Entregáveis
- [x] Command runner: `geee run --config <file>`
- [x] Flags: `--input`, `--output`, `--verbose`, `--config`
- [x] Input/Output via arquivos JSON/YAML
- [x] Structured logging (JSON format)
- [x] Help e autocompletar (bash)
- [x] Verbose/debug mode
- [x] Error reporting estruturado

#### Commandos CLI
```bash
geee run --config configs/pipeline.yaml --input data.json --output result.json
geee run --config configs/pipeline.yaml --input data.json --output result.json --verbose
geee --help
geee plugins list
```

#### Criteria de Aceite
- [x] CLI funciona com configuração YAML
- [x] Output é válido e bem formatado
- [x] Erros são úteis e acionáveis
- [x] Modo verbose mostra detalhes de execução

#### Arquivos Criados
- `internal/observability/logger.go` - JSONLogger com suporte a níveis de log (debug, info, warn, error)
- `internal/observability/logger_test.go` - Testes unitários para logger (85% coverage)
- `internal/observability/metrics.go` - NoOpMetricsCollector (implementação stub para Fase 6)
- `internal/cli/cli.go` - Implementação completa da CLI com comandos run e plugins list
- `internal/cli/cli_test.go` - Testes unitários para CLI (57.1% coverage)
- `cmd/geee/main.go` - Entry point CLI completo com flags e subcomandos
- `test/fixtures/input.json` - Arquivo de exemplo para testes
- `test/integration/cli_test.go` - Testes de integração (1 skipped, aguardando Fase 4)

#### Funcionalidades Implementadas
- **CLI Commands**: `run` e `plugins list` totalmente funcionais
- **Flags**: `--config`, `--input`, `--output`, `--verbose` implementados
- **Input/Output**: Suporte para JSON e YAML em input e output
- **Structured Logging**: Logs em formato JSON com timestamp, level e fields
- **Verbose Mode**: Debug logs habilitados quando --verbose é usado
- **Error Handling**: Mensagens de erro claras e acionáveis
- **Help System**: Help detalhado com exemplos de uso
- **Validation**: Validação de flags obrigatórios e formatos de arquivo

---

### Fase 4: Plugins de Exemplo ✅ CONCLUÍDA
**Duração estimada**: 3-4 dias
**Duração real**: 1 dia
**Data de conclusão**: 2026-01-12

#### Entregáveis

##### Plugin 1: json-transformer
- **Função**: Manipula e renomeia campos no estado
- **Input**: `{ "a": 1, "b": "text" }`
- **Output**: `{ "x": 1, "y": "text" }`
- **Configuração**:
  ```yaml
  - id: json-transformer
    mappings:
      - from: "a"
        to: "x"
      - from: "b"
        to: "y"
  ```

##### Plugin 2: regex-extractor
- **Função**: Extrai padrões via expressões regulares
- **Input**: `{ "text": "email: foo@bar.com, phone: 123" }`
- **Output**: `{ "email": "foo@bar.com", "phone": "123" }`
- **Configuração**:
  ```yaml
  - id: regex-extractor
    patterns:
      - name: "email"
        pattern: "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}"
      - name: "phone"
        pattern: "\\d{3,}"
  ```

##### Plugin 3: http-fetcher
- **Função**: Busca conteúdo de URLs
- **Input**: `{ "url": "https://api.example.com/data" }`
- **Output**: `{ "response": { "status": 200, "body": "..." } }`
- **Configuração**:
  ```yaml
  - id: http-fetcher
    url_field: "url"
    response_field: "response"
    timeout: 30
  ```

##### Plugin 4: template-renderer
- **Função**: Renderiza templates Go
- **Input**: `{ "template": "Hello {{.Name}}!", "data": {"Name": "World"} }`
- **Output**: `{ "result": "Hello World!" }`
- **Configuração**:
  ```yaml
  - id: template-renderer
    template_field: "template"
    data_field: "data"
    output_field: "result"
  ```

#### Critérios de Aceite
- [x] Cada plugin tem manifest válido
- [x] Cada plugin tem testes unitários
- [x] Plugins compilam sem warnings
- [x] Documentação em cada plugin

#### Arquivos Criados
- `plugins/json-transformer/json_transformer.go` - Plugin de transformação de campos JSON
- `plugins/json-transformer/json_transformer_test.go` - Testes unitários (81% coverage)
- `plugins/json-transformer/README.md` - Documentação completa do plugin
- `plugins/regex-extractor/regex_extractor.go` - Plugin de extração via regex
- `plugins/regex-extractor/regex_extractor_test.go` - Testes unitários (74.2% coverage)
- `plugins/regex-extractor/README.md` - Documentação completa do plugin
- `plugins/http-fetcher/http_fetcher.go` - Plugin de requisições HTTP
- `plugins/http-fetcher/http_fetcher_test.go` - Testes unitários (74.5% coverage)
- `plugins/http-fetcher/README.md` - Documentação completa do plugin
- `plugins/template-renderer/template_renderer.go` - Plugin de renderização de templates
- `plugins/template-renderer/template_renderer_test.go` - Testes unitários (77.8% coverage)
- `plugins/template-renderer/README.md` - Documentação completa do plugin
- `plugins/registry.go` - Registro centralizado de todos os plugins
- `configs/example-pipeline.yaml` - Pipeline de exemplo usando múltiplos plugins
- `configs/simple-test-pipeline.yaml` - Pipeline simples para testes
- `test/fixtures/example-input.json` - Dados de exemplo para testes

#### Funcionalidades Implementadas

**Plugin 1: json-transformer** ✅
- Mapeamento de campos flexível (from/to)
- Preservação de campos não mapeados
- Suporte a múltiplos mapeamentos simultâneos
- Validação de configuração via JSON Schema
- Coverage: 81.0%

**Plugin 2: regex-extractor** ✅
- Múltiplos padrões regex em uma única execução
- Campo de entrada configurável
- Compilação e validação de padrões
- Extração do primeiro match de cada padrão
- Coverage: 74.2%

**Plugin 3: http-fetcher** ✅
- Suporte a todos os métodos HTTP padrão
- Timeout configurável
- Captura completa de resposta (status, headers, body)
- Cleanup automático de conexões
- Context-aware para cancelamento
- Coverage: 74.5%

**Plugin 4: template-renderer** ✅
- Suporte completo à sintaxe de templates Go
- Campos configuráveis (template, data, output)
- Suporte a estruturas aninhadas
- Condicionais e loops
- Validação de sintaxe de template
- Coverage: 77.8%

#### Integração com CLI
- Todos os plugins registrados automaticamente via `plugins.RegisterAll()`
- CLI lista todos os 4 plugins com `geee plugins list`
- Plugins prontos para uso em pipelines YAML

---

### Fase 5: HTTP Server ✅ CONCLUÍDA
**Duração estimada**: 2-3 dias
**Duração real**: 1 dia
**Data de conclusão**: 2026-01-12

#### Entregáveis
- [x] Entry point: `cmd/geee-server/main.go`
- [x] Chi router configurável
- [x] Endpoints:
  - `POST /run` - Executa pipeline
  - `GET /health` - Health check
  - `GET /plugins` - Lista plugins
  - `GET /ready` - Readiness probe
- [x] Middlewares:
  - Logging
  - Recovery
  - CORS (configurável)
- [x] Request/Response validation
- [x] Configuração via flags/env vars

#### Endpoints

```
POST /run
{
  "config": "pipeline.yaml",
  "input": { "data": "..." }
}

Response:
{
  "success": true,
  "data": { "result": "..." },
  "execution_id": "abc123",
  "duration_ms": 150
}
```

```
GET /health
Response:
{
  "status": "healthy",
  "plugins": {
    "json-transformer": "healthy",
    "regex-extractor": "healthy"
  }
}
```

#### Critérios de Aceite
- [x] Server inicia sem erros
- [x] Endpoints respondem corretamente
- [x] Health check funciona
- [x] Pronto para integração MCP/API

#### Arquivos Criados
- `internal/server/types.go` - Tipos de request/response para os endpoints
- `internal/server/middleware.go` - Middlewares (logging, recovery, CORS)
- `internal/server/handlers.go` - Handlers dos endpoints HTTP
- `internal/server/server.go` - Estrutura principal do servidor HTTP
- `internal/server/handlers_test.go` - Testes unitários dos handlers (13 testes, 100% pass)
- `internal/server/middleware_test.go` - Testes unitários dos middlewares (4 testes, 100% pass)
- `cmd/geee-server/main.go` - Entry point do servidor HTTP
- `pkg/types/types.go` - Adicionado método Error() ao StructuredError
- `Makefile` - Adicionados comandos: server, server-build, server-build-prod, server-dev, server-run

#### Funcionalidades Implementadas

**Servidor HTTP** ✅
- Chi router v5 configurável
- Graceful shutdown com timeout configurável
- Suporte a variáveis de ambiente e flags
- Health check com status de plugins
- Readiness probe
- Configuração via flags: --port, --host, --cors, --allowed-origins, --verbose
- Configuração via env vars: GEEE_PORT, GEEE_HOST, GEEE_ENABLE_CORS, GEEE_ALLOWED_ORIGINS, GEEE_VERBOSE

**Middlewares** ✅
- Logging estruturado de todas as requisições (método, path, status, duração)
- Recovery automático de panics com stack trace
- CORS configurável com origens permitidas
- Content-Type JSON automático

**Endpoints** ✅
1. **POST /run** - Executa pipelines
   - Aceita configuração inline YAML ou path para arquivo
   - Validação de request/response
   - Timeout configurável por pipeline
   - Retorna execution_id único e duração em ms
   - Tratamento de erros estruturados

2. **GET /health** - Health check
   - Status geral do servidor
   - Status individual de cada plugin
   - Útil para load balancers

3. **GET /plugins** - Lista plugins
   - Informações detalhadas de cada plugin (ID, nome, descrição, versão)
   - Extração automática de inputs/outputs dos schemas
   - Count total de plugins

4. **GET /ready** - Readiness probe
   - Indica se servidor está pronto para receber tráfego
   - Timestamp de quando ficou ready
   - Útil para Kubernetes

**Testes** ✅
- 17 testes unitários (100% pass)
- Coverage dos handlers e middlewares
- Testes de sucesso e erro
- Testes de validação de request
- Testes de panic recovery
- Testes de logging

**Comandos Makefile** ✅
```bash
make server            # Build e inicia o servidor
make server-build      # Build do servidor local
make server-build-prod # Build de produção (Linux AMD64)
make server-dev        # Inicia servidor em modo verbose
make server-run        # Executa servidor já compilado
```

#### Testes Manuais Realizados
- ✅ Servidor inicia sem erros na porta 8080
- ✅ Endpoint /health retorna status healthy com 4 plugins
- ✅ Endpoint /ready retorna ready=true
- ✅ Endpoint /plugins lista todos os 4 plugins
- ✅ Endpoint POST /run executa pipeline com sucesso
- ✅ Endpoint POST /run retorna erro estruturado quando config está vazio
- ✅ Logs estruturados em formato JSON
- ✅ Graceful shutdown funcionando

---

### Fase 6: Observabilidade
**Duração estimada**: 2-3 dias

#### Entregáveis
- [ ] Structured errors com suggestions
- [ ] Structured logging (JSON format)
- [ ] Métricas Prometheus:
  - `geee_execution_duration_seconds`
  - `geee_plugin_execution_duration_seconds`
  - `geee_plugin_errors_total`
  - `geee_active_executions`
- [ ] OpenTelemetry tracing (spans por plugin)
- [ ] Health check endpoint

#### Structured Error Format
```json
{
  "code": "SCHEMA_VALIDATION_FAILED",
  "message": "Required field 'email' is missing",
  "field_path": "$.input.email",
  "suggestion": "Add 'email' field to input or use optional plugin",
  "plugin_id": "regex-extractor"
}
```

#### Structured Log Format
```json
{
  "timestamp": "2025-01-11T10:30:45Z",
  "level": "info",
  "execution_id": "exec-abc123",
  "plugin_id": "http-fetcher",
  "event": "plugin_completed",
  "duration_ms": 1200
}
```

#### Criteria de Aceite
- [ ] Erros são estruturados e úteis
- [ ] Logs são JSON formatados
- [ ] Métricas exportadas em /metrics
- [ ] Traces gerados por execução

---

### Fase 7: Quality Assurance
**Duração estimada**: 2-3 dias

#### Entregáveis
- [ ] Testes de integração CLI
- [ ] Testes de integração HTTP
- [ ] Race detector tests (`make test-race`)
- [ ] Coverage report (>80%)
- [ ] Benchmarking baseline
- [ ] Linting (golangci-lint)
- [ ] Makefile commands validados

#### Testes de Integração CLI
```bash
# Testa execução completa
geee run --config configs/test-pipeline.yaml --input test/fixtures/input.json --output /tmp/result.json

# Valida output
jq -e '.success == true' /tmp/result.json
```

#### Testes de Integração HTTP
```bash
# Testa endpoint /run
curl -X POST http://localhost:8080/run \
  -H "Content-Type: application/json" \
  -d '{"config": "pipeline.yaml", "input": {"data": "test"}}'
```

#### Criteria de Aceite
- [ ] Todos os testes passam
- [ ] `make test-race` passa sem race conditions
- [ ] Coverage > 80%
- [ ] Linting sem errors
- [ ] Benchmarks documentados

---

## Plugins de Exemplo (Deliverables)

| Plugin | Input | Output | Uso |
|--------|-------|--------|-----|
| `json-transformer` | `{ "a": 1 }` | `{ "b": 1 }` | Renomeia/transforma campos |
| `regex-extractor` | `"email: foo@bar.com"` | `{ "email": "foo@bar.com" }` | Extrai padrões |
| `http-fetcher` | `{ "url": "..." }` | `{ "response": "..." }` | Busca URLs |
| `template-renderer` | `{ "template": "...", "data": {...} }` | `{ "result": "..." }` | Renderiza templates Go |

---

## Comandos Makefile

```bash
# Desenvolvimento
make dev          # Modo desenvolvimento com hot reload
make build        # Build de produção
make run          # Build e executa CLI

# Testes
make test         # Testes unitários
make test-race    # Testes com race detector
make test-coverage # Testes com cobertura

# Qualidade
make lint         # Executa linter
make fmt          # Formata código

# Plugins
make plugin-new NAME=my-plugin  # Cria scaffold de plugin

# Server
make server       # Inicia HTTP server
make server-build # Build do server

# Limpeza
make clean        # Remove builds e caches
```

---

## Timeline

| Fase | Duração | Entrega |
|------|---------|---------|
| Fase 1: Foundation | 2-3 dias | Core interfaces, types, config loader |
| Fase 2: Plugin System | 3-4 dias | Registry, base plugin, execution DAG |
| Fase 3: CLI | 2-3 dias | CLI funcional com flags |
| Fase 4: Plugins de Exemplo | 3-4 dias | 4 plugins implementados |
| Fase 5: HTTP Server | 2-3 dias | Server extensível |
| Fase 6: Observabilidade | 2-3 dias | Logs, métricas, tracing |
| Fase 7: Quality Assurance | 2-3 dias | Testes, coverage, linting |

**Tempo total estimado**: 14-20 dias

---

## Critérios de Aceite Gerais

### Funcionais
- [ ] Core funciona sem plugins de domínio
- [ ] Plugins opcionais não abortam pipeline
- [ ] CLI funciona conforme especificado
- [ ] Server extensível para futuras integrações

### Não-Funcionais
- [ ] Memória idle < 50MB
- [ ] Estado por execução < 10MB
- [ ] Cleanup 100% garantido
- [ ] Zero race conditions
- [ ] Coverage > 80%
- [ ] Health check < 100ms

### Observabilidade
- [ ] Logs estruturados (JSON)
- [ ] Erros estruturados com suggestions
- [ ] Métricas Prometheus exportadas
- [ ] Traces por execução

---

## Próximos Passos

1. Revisar e aprovar este plano
2. Iniciar Fase 1: Foundation
3. Criar primeira versão do go.mod
4. Definir interfaces core

---

**Documento criado em**: 2025-01-11
**Versão**: 1.0.0
