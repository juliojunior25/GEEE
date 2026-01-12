# Plano de Implementação - GEEE

**Generic Extraction & Enrichment Engine**

**Versão**: 1.0.0
**Data**: 2025-01-11
**Status**: Aprovado para desenvolvimento

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

### Fase 3: CLI
**Duração estimada**: 2-3 dias

#### Entregáveis
- [ ] Command runner: `geee run --config <file>`
- [ ] Flags: `--input`, `--output`, `--verbose`, `--config`
- [ ] Input/Output via arquivos JSON/YAML
- [ ] Structured logging (JSON format)
- [ ] Help e autocompletar (bash)
- [ ] Verbose/debug mode
- [ ] Error reporting estruturado

#### Commandos CLI
```bash
geee run --config configs/pipeline.yaml --input data.json --output result.json
geee run --config configs/pipeline.yaml --input data.json --output result.json --verbose
geee --help
geee plugins list
```

#### Criteria de Aceite
- [ ] CLI funciona com configuração YAML
- [ ] Output é válido e bem formatado
- [ ] Erros são úteis e acionáveis
- [ ] Modo verbose mostra detalhes de execução

---

### Fase 4: Plugins de Exemplo
**Duração estimada**: 3-4 dias

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

#### Criteria de Aceite
- [ ] Cada plugin tem manifest válido
- [ ] Cada plugin tem testes unitários
- [ ] Plugins compilam sem warnings
- [ ] Documentação em cada plugin

---

### Fase 5: HTTP Server (Extensível)
**Duração estimada**: 2-3 dias

#### Entregáveis
- [ ] Entry point: `cmd/geee-server/main.go`
- [ ] Chi router configurável
- [ ] Endpoints:
  - `POST /run` - Executa pipeline
  - `GET /health` - Health check
  - `GET /plugins` - Lista plugins
  - `GET /ready` - Readiness probe
- [ ] Middlewares:
  - Logging
  - Recovery
  - CORS (configurável)
- [ ] Request/Response validation
- [ ] Configuração via flags/env vars

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

#### Criteria de Aceite
- [ ] Server inicia sem erros
- [ ] Endpoints respondem corretamente
- [ ] Health check funciona
- [ ] Pronto para integração MCP/API

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
