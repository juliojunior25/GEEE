# Product Requirement Prompt (PRP)

## Nome do Projeto

**Generic Extraction & Enrichment Engine (GEEE)**

## Linguagem Base

**Go (Golang)**

## Objetivo do PRP

Este PRP define, de forma **precisa, executável e não ambígua**, os requisitos para que uma IA ou equipe de engenharia implemente um **motor genérico de extração e enriquecimento de dados**, orientado a **plugins**, escrito em **Go**, com foco em:

* Baixo consumo de recursos
* Execução local (serverless / low-cost)
* Extensibilidade extrema via plugins
* IA atuando como **planner**, não como executor
* Zero acoplamento entre domínio, plugins e core

Este documento **não descreve um app específico**, mas sim um **framework/motor reutilizável**.

---

## 1. Visão Geral do Produto

O sistema é um **motor genérico de transformação de dados não estruturados em dados estruturados**, por meio da composição dinâmica de plugins.

O motor recebe:

* Um **estado inicial de dados** (ex: URL, arquivo, texto)
* Um **estado alvo (schema desejado)**
* Um **catálogo de plugins disponíveis**

E produz:

* Um **plano de execução**
* A **execução validada desse plano**
* Um **output final estruturado**, com métricas de confiança

A IA **não executa código** e **não conhece implementações** — ela apenas **planeja transformações possíveis**, com base nos contratos declarados pelos plugins.

---

## 2. Princípios Arquiteturais (Obrigatórios)

1. **Core Imutável**
   O core do sistema não deve mudar ao adicionar novos plugins ou domínios.

2. **Plugins Declarativos**
   Toda capacidade deve ser descrita por manifesto legível por máquina e IA.

3. **Schema-Driven**
   Todas as entradas e saídas são definidas por JSON Schema.

4. **Fail-Fast e Validado**
   Nenhum plugin pode executar se os requisitos de entrada não forem satisfeitos.

5. **Low Resource First**
   O sistema deve priorizar baixo uso de CPU e memória.

---

## 3. Escopo da Versão 1.0

### INCLUSO

* Core Engine em Go
* Sistema de Plugin Registry
* Plugin Loader (estático)
* Planejamento via IA (LLM externo ou mock)
* Execução sequencial de plano
* Validação de schemas
* Geração de output final em JSON

### FORA DO ESCOPO

* Interface gráfica
* Execução distribuída
* Plugins carregados remotamente
* Persistência em banco de dados

---

## 4. Arquitetura de Alto Nível

```
┌──────────────┐
│   Input      │
└──────┬───────┘
       ↓
┌──────────────┐
│ Plugin       │◄───┐
│ Registry     │    │
└──────┬───────┘    │
       ↓            │
┌──────────────┐    │
│ IA Planner   │────┘
└──────┬───────┘
       ↓
┌──────────────┐
│ Plan Executor│
└──────┬───────┘
       ↓
┌──────────────┐
│ Output JSON  │
└──────────────┘
```

---

## 5. Conceitos Fundamentais

### 5.1 Estado de Dados

O estado é um objeto JSON incremental que **NUNCA** deve carregar dados binários grandes diretamente em memória.

#### Estratégia de Referências (Obrigatório)

Dados binários ou grandes devem ser armazenados como referências a arquivos temporários:

```json
{
  "video_url": "string",
  "audio_ref": "file://tmp/exec-abc123/audio.wav",
  "transcript": "string",
  "_metadata": {
    "temp_files": ["file://tmp/exec-abc123/audio.wav"],
    "cleanup_on_finish": true,
    "max_state_size_bytes": 10485760
  }
}
```

#### Limites de Tamanho (Obrigatório)

* Estado em memória: **máximo 10MB por execução**
* Strings individuais: **máximo 1MB**
* Arrays: **máximo 1000 elementos**
* Dados binários: **sempre como referência de arquivo**

#### Garbage Collection de Estado

Plugins podem declarar campos descartáveis após uso:

```json
{
  "id": "video-to-audio",
  "disposable_outputs": ["raw_frames", "temp_buffer"]
}
```

---

### 5.2 Plugin

Um plugin representa **uma transformação de estado**.

#### Interface Base (Go)

```go
type Plugin interface {
    // Identificação e metadados
    ID() string
    Version() string // Semantic versioning obrigatório
    Type() PluginType
    Description() string

    // Schemas de dados
    InputSchema() JSONSchema
    OutputSchema() JSONSchema

    // Execução
    Execute(ctx context.Context, input map[string]any) (PluginResult, error)

    // Resource management (OBRIGATÓRIO para prevenir leaks)
    Cleanup(ctx context.Context) error
    MaxMemoryBytes() int64
    MaxDurationSeconds() int
    IsParallelSafe() bool

    // Observability
    HealthCheck(ctx context.Context) error
}

type PluginResult struct {
    Data       map[string]any
    Confidence float64 // 0.0 a 1.0
    Metadata   ResultMetadata
}

type ResultMetadata struct {
    ExecutionTimeMs int64
    TempFilesCreated []string
    WarningsCollected []string
}
```

---

### 5.3 Plugin Manifest (Obrigatório)

Cada plugin deve expor um manifesto declarativo.

```json
{
  "id": "geo-resolver",
  "version": "1.2.0",
  "type": "enrichment",
  "description": "Resolve referência textual em coordenadas geográficas",
  "required": true,
  "parallel_safe": true,
  "max_memory_bytes": 52428800,
  "max_duration_seconds": 30,
  "disposable_outputs": [],
  "inputSchema": {
    "required": ["place_reference"],
    "properties": {
      "place_reference": {"type": "string"}
    }
  },
  "outputSchema": {
    "properties": {
      "latitude": {"type": "number"},
      "longitude": {"type": "number"},
      "confidence": {"type": "number", "minimum": 0, "maximum": 1}
    }
  },
  "error_schema": {
    "properties": {
      "error_code": {"type": "string"},
      "error_message": {"type": "string"},
      "field_path": {"type": "string"},
      "suggestion": {"type": "string"}
    }
  }
}
```

#### Campos Obrigatórios do Manifest

* **version**: Semantic versioning (MAJOR.MINOR.PATCH)
* **required**: Se false, falha não aborta pipeline
* **parallel_safe**: Pode executar concorrentemente
* **max_memory_bytes**: Limite de memória (hard limit)
* **max_duration_seconds**: Timeout de execução
* **disposable_outputs**: Campos que podem ser GC após uso
* **error_schema**: Schema estruturado de erros

---

## 6. Plugin Registry

### Responsabilidades

* Registrar plugins
* Expor catálogo de capacidades
* Validar conflitos de IDs

```go
type PluginRegistry struct {
    plugins map[string]Plugin
}
```

### API Esperada

```go
func (r *PluginRegistry) Register(p Plugin) error
func (r *PluginRegistry) List() []PluginDescriptor
func (r *PluginRegistry) Get(id string) (Plugin, bool)
```

---

## 7. IA Planner

### Papel da IA

* Receber estado atual
* Receber schema alvo
* Receber catálogo de plugins
* Produzir **plano de execução**

### Output Esperado

```json
{
  "plan": [
    {
      "plugin_id": "video-to-audio",
      "produces": ["audio"]
    },
    {
      "plugin_id": "speech-to-text",
      "produces": ["transcript"]
    }
  ]
}
```

### Regras

* A IA **não pode inventar plugins**
* A IA **não pode pular requisitos de schema**
* Campos não resolvíveis devem ser explicitados

---

## 8. Plan Executor

### Responsabilidades

* Validar inputs de cada passo
* Executar plugins com DAG de dependências (paralelo quando possível)
* Atualizar estado global
* Cleanup de recursos temporários
* Handling de falhas parciais (plugins opcionais)

```go
type PlanExecutor struct {
    registry       *PluginRegistry
    maxWorkers     int
    tempDir        string
    cleanupTimeout time.Duration
}

type ExecutionContext struct {
    ExecutionID   string
    StartTime     time.Time
    State         map[string]any
    TempFiles     []string
    Metadata      ExecutionMetadata
    cancelFunc    context.CancelFunc
}

type ExecutionMetadata struct {
    PluginsExecuted   []string
    ConfidenceScores  map[string]float64
    Warnings          []string
    ResourceUsage     ResourceMetrics
}

type ResourceMetrics struct {
    PeakMemoryBytes int64
    TotalDurationMs int64
    PluginDurations map[string]int64
}
```

### Execução com Backpressure

```go
type PipelineStage struct {
    Plugin     Plugin
    InputChan  chan map[string]any
    OutputChan chan PluginResult
    ErrChan    chan error
}

// Bounded channels previnem memory bloat
const maxChannelBuffer = 10
```

### Cleanup Automático

```go
func (e *ExecutionContext) Cleanup(ctx context.Context) error {
    // 1. Cancelar plugins em execução
    if e.cancelFunc != nil {
        e.cancelFunc()
    }

    // 2. Cleanup de cada plugin
    for _, pluginID := range e.Metadata.PluginsExecuted {
        plugin.Cleanup(ctx)
    }

    // 3. Remover arquivos temporários
    for _, file := range e.TempFiles {
        os.Remove(file)
    }

    // 4. Cleanup do diretório temporário
    return os.RemoveAll(e.tempDir)
}
```

---

## 9. Output Builder

### Responsabilidades

* Validar estado final contra schema alvo
* Agregar confidence scores de plugins
* Adicionar metadados e métricas
* Gerar JSON final estruturado

```json
{
  "data": {
    "transcript": "...",
    "latitude": 40.7128,
    "longitude": -74.0060
  },
  "confidence": {
    "overall": 0.87,
    "by_field": {
      "transcript": 0.95,
      "latitude": 0.82,
      "longitude": 0.82
    },
    "aggregation_method": "weighted_average"
  },
  "execution_metadata": {
    "execution_id": "exec-abc123",
    "duration_ms": 3450,
    "plugins_executed": ["video-to-audio", "speech-to-text", "geo-resolver"],
    "plugins_skipped": [],
    "plugins_failed": [],
    "warnings": []
  },
  "resource_usage": {
    "peak_memory_bytes": 8388608,
    "total_duration_ms": 3450,
    "plugin_durations": {
      "video-to-audio": 2000,
      "speech-to-text": 1200,
      "geo-resolver": 250
    }
  }
}
```

### Agregação de Confidence

```go
func AggregateConfidence(scores map[string]float64, method string) float64 {
    switch method {
    case "minimum":
        return min(scores)
    case "weighted_average":
        return weightedAvg(scores)
    default:
        return average(scores)
    }
}
```

---

## 10. Requisitos Não-Funcionais

### 10.1 Concorrência e Execução Paralela (Obrigatório)

O sistema **DEVE** suportar a execução de múltiplas instâncias do motor em paralelo, seja:

* Em múltiplas goroutines dentro do mesmo processo
* Em múltiplos processos do binário
* Como workers independentes atrás de uma API

#### Requisitos Específicos

* O Core Engine **DEVE ser stateless** entre execuções
* Nenhum estado global mutável pode ser compartilhado entre pipelines
* Cada execução deve operar sobre seu próprio **Execution Context** isolado

```go
type ExecutionContext struct {
    ExecutionID string
    StartTime   time.Time
    State       map[string]any
}
```

* O Plugin Registry deve ser **read-only** após inicialização
* Plugins **NÃO PODEM** manter estado interno mutável compartilhado

#### Paralelismo Interno

* O motor **PODE** executar etapas independentes em paralelo quando:

  * Não houver dependência de schema
  * Plugins forem declarados como `parallel_safe = true`

Exemplo de manifesto:

```json
{
  "id": "speech-to-text",
  "parallel_safe": true
}
```

#### Controle de Recursos

* Limite máximo de execuções simultâneas configurável
* Pool de workers configurável
* Cada plugin deve respeitar `context.Context` para cancelamento

---

* Memória idle < 50MB
* Estado em execução: máximo 10MB por pipeline
* Execuções concorrentes sem race conditions
* Compatível com uso como backend de API
* Escalável horizontalmente via múltiplos processos
* Plugins com timeout configurável
* Context cancellation obrigatório
* Logs estruturados (JSON format)
* Cleanup automático de recursos em todas as condições (sucesso, erro, cancelamento)

---

### 10.2 Observabilidade (Obrigatório)

O sistema **DEVE** implementar observabilidade completa para debugging e otimização.

#### Tracing

* Toda execução gera trace distribuído (OpenTelemetry compatible)
* Spans por plugin com atributos:
  * `plugin.id`
  * `plugin.version`
  * `plugin.input_size_bytes`
  * `plugin.output_size_bytes`
  * `plugin.confidence`

```go
import "go.opentelemetry.io/otel"

func (p *PlanExecutor) ExecutePlugin(ctx context.Context, plugin Plugin) {
    ctx, span := otel.Tracer("geee").Start(ctx, plugin.ID())
    defer span.End()

    span.SetAttributes(
        attribute.String("plugin.id", plugin.ID()),
        attribute.String("plugin.version", plugin.Version()),
    )

    result, err := plugin.Execute(ctx, input)
    // ...
}
```

#### Métricas (Prometheus format)

```
# Latência por plugin
geee_plugin_duration_seconds{plugin_id="geo-resolver",status="success"} histogram

# Taxa de erro
geee_plugin_errors_total{plugin_id="geo-resolver",error_type="timeout"} counter

# Uso de memória
geee_plugin_memory_bytes{plugin_id="video-to-audio"} gauge

# Execuções ativas
geee_active_executions gauge

# Tamanho de estado
geee_state_size_bytes{execution_id="..."} gauge
```

#### Logs Estruturados

```json
{
  "timestamp": "2025-01-11T10:30:45Z",
  "level": "info",
  "execution_id": "exec-abc123",
  "plugin_id": "speech-to-text",
  "event": "plugin_completed",
  "duration_ms": 1200,
  "confidence": 0.95,
  "memory_bytes": 4194304
}
```

#### Health Checks

```go
type HealthStatus struct {
    Status      string            // "healthy", "degraded", "unhealthy"
    Plugins     map[string]string // plugin_id -> status
    LastChecked time.Time
}

func (r *PluginRegistry) HealthCheck(ctx context.Context) HealthStatus {
    status := HealthStatus{
        Status:  "healthy",
        Plugins: make(map[string]string),
    }

    for id, plugin := range r.plugins {
        if err := plugin.HealthCheck(ctx); err != nil {
            status.Plugins[id] = "unhealthy"
            status.Status = "degraded"
        } else {
            status.Plugins[id] = "healthy"
        }
    }

    return status
}
```

---

### 10.3 Tratamento de Erros Estruturado

Todos os erros devem seguir formato estruturado para facilitar debugging:

```go
type StructuredError struct {
    Code       string         `json:"code"`
    Message    string         `json:"message"`
    PluginID   string         `json:"plugin_id,omitempty"`
    FieldPath  string         `json:"field_path,omitempty"`
    Suggestion string         `json:"suggestion,omitempty"`
    Details    map[string]any `json:"details,omitempty"`
}

// Exemplo de erro
{
  "code": "SCHEMA_VALIDATION_FAILED",
  "message": "Required field missing",
  "plugin_id": "geo-resolver",
  "field_path": "$.place_reference",
  "suggestion": "Add 'place_reference' field to input or use optional plugin",
  "details": {
    "expected_type": "string",
    "received": null
  }
}
```

#### Códigos de Erro Padrão

* `SCHEMA_VALIDATION_FAILED`: Input/output não atende schema
* `PLUGIN_TIMEOUT`: Execução excedeu max_duration_seconds
* `PLUGIN_OOM`: Plugin excedeu max_memory_bytes
* `PLUGIN_UNAVAILABLE`: Plugin não está healthy
* `STATE_SIZE_EXCEEDED`: Estado excedeu 10MB
* `CLEANUP_FAILED`: Recursos não foram liberados corretamente

---

## 11. Workflow de Desenvolvimento (Obrigatório)

### 11.1 Makefile como Interface Única

**TODOS os comandos de desenvolvimento DEVEM ser executados via Makefile.**

Nenhum comando Go direto (`go build`, `go run`, `go test`) deve ser necessário. O Makefile abstrai complexidades e garante consistência.

#### Comandos Obrigatórios

```makefile
# Desenvolvimento
make dev          # Roda em modo desenvolvimento com hot reload
make build        # Build de produção (otimizado, sem debug symbols)
make test         # Roda todos os testes
make test-race    # Testes com race detector
make bench        # Benchmarks
make lint         # Linting (golangci-lint)
make fmt          # Formatação de código

# Plugins
make plugin-new NAME=<plugin_name>   # Scaffold de novo plugin
make plugin-test PLUGIN=<plugin_id>  # Testa plugin específico
make plugin-bench PLUGIN=<plugin_id> # Benchmark de plugin

# Observabilidade
make metrics      # Exporta métricas locais
make trace        # Visualiza traces (Jaeger local)
make profile      # Profiling de CPU/memória

# Limpeza
make clean        # Remove builds e caches
make clean-temp   # Remove arquivos temporários de execuções

# Docker
make docker-build # Build de imagem Docker
make docker-run   # Roda em container
```

#### Exemplo de Makefile Base

```makefile
.PHONY: dev build test clean

# Variáveis
BINARY_NAME=geee
BUILD_DIR=./bin
MAIN_PATH=./cmd/geee

# Flags
LDFLAGS=-ldflags "-s -w"
DEV_FLAGS=-race -gcflags="all=-N -l"

dev:
	@echo "🚀 Starting development mode..."
	@air -c .air.toml || go run $(DEV_FLAGS) $(MAIN_PATH)

build:
	@echo "🔨 Building production binary..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

test:
	@echo "🧪 Running tests..."
	@go test -v -cover ./...

test-race:
	@echo "🏃 Running tests with race detector..."
	@go test -race -v ./...

clean:
	@echo "🧹 Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf /tmp/geee-*
	@go clean -cache
```

### 11.2 Hot Reload em Desenvolvimento

Para facilitar iteração rápida, o comando `make dev` deve suportar hot reload:

```toml
# .air.toml
[build]
  cmd = "go build -o ./tmp/main ./cmd/geee"
  bin = "./tmp/main"
  include_ext = ["go", "json"]
  exclude_dir = ["tmp", "vendor"]
```

### 11.3 CI/CD Integration

O Makefile deve ser a interface para CI/CD:

```yaml
# .github/workflows/test.yml
- name: Run tests
  run: make test-race

- name: Build
  run: make build

- name: Lint
  run: make lint
```

---

## 12. Critérios de Aceite

### Funcionais

* Core funciona sem plugins de domínio
* Adição de plugin não exige mudança no core
* IA só utiliza capacidades registradas
* Output validado por schema
* Plugins opcionais não abortam pipeline em caso de falha

### Não-Funcionais

* Memória idle < 50MB
* Estado por execução < 10MB
* Cleanup completo em 100% das execuções (medido por testes)
* Zero race conditions (validado por `make test-race`)
* Todos os comandos via Makefile
* 100% de cobertura de logs estruturados
* Métricas Prometheus exportadas
* Health check responde em < 100ms

### Observabilidade

* Todo plugin gera trace com spans
* Erros estruturados com suggestions
* Métricas de latência p50, p95, p99 por plugin
* Logs com execution_id para correlação

---

## 13. Resultado Esperado

Um **motor genérico**, extensível, orientado a contratos, capaz de compor pipelines complexos de extração e enriquecimento com:

* **Baixo custo operacional**: < 50MB idle, < 10MB por pipeline
* **Alta previsibilidade**: Erros estruturados, timeouts, limites de recursos
* **Zero leaks**: Cleanup automático garantido
* **Observabilidade completa**: Traces, métricas, logs, health checks
* **Developer Experience**: Makefile como única interface

---

**Fim do PRP v1.1 (Critical Improvements)**
