# GEEE - Generic Extraction & Enrichment Engine

**Motor genérico de extração e enriquecimento de dados**, orientado a plugins, escrito em Go, com foco em baixo consumo de recursos e alta extensibilidade.

## 🎯 Características Principais

- **Zero Memory Leaks**: Cleanup automático garantido
- **Low Resource**: < 50MB idle, < 10MB por pipeline
- **Plugin-Based**: Extensível sem modificar o core
- **AI Planning**: IA planeja transformações, não executa código
- **Full Observability**: Traces, métricas, logs estruturados
- **Developer Friendly**: Makefile como interface única

## 🚀 Quick Start

### Pré-requisitos

- Go 1.21+
- Make
- (Opcional) Air para hot reload: `go install github.com/air-verse/air@latest`
- (Opcional) golangci-lint: `brew install golangci-lint`

### Instalação

```bash
# Clone o repositório
git clone <repo-url>
cd GEEE

# Instale dependências
make deps

# Verifique ferramentas
make check-tools
```

### Desenvolvimento

```bash
# Modo desenvolvimento com hot reload
make dev

# Build de produção
make build

# Executar build local
make run
```

## 📚 Comandos Disponíveis

### Desenvolvimento
```bash
make dev          # Modo desenvolvimento com hot reload
make build        # Build de produção otimizado
make run          # Build e executa localmente
make fmt          # Formata código
make lint         # Executa linter
```

### Testes
```bash
make test         # Roda todos os testes
make test-race    # Testes com race detector
make test-coverage # Cobertura detalhada
make bench        # Benchmarks
```

### Plugins
```bash
make plugin-new NAME=my-plugin     # Cria scaffold de plugin
make plugin-test PLUGIN=my-plugin  # Testa plugin específico
make plugin-bench PLUGIN=my-plugin # Benchmark de plugin
```

### Observabilidade
```bash
make metrics      # Exporta métricas
make trace        # Visualiza traces
make profile      # Profiling CPU/memória
```

### Limpeza
```bash
make clean        # Remove builds e caches
make clean-temp   # Remove arquivos temporários
```

## 🏗️ Arquitetura

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

## 🔌 Criando um Plugin

```bash
# Cria scaffold
make plugin-new NAME=geo-resolver

# Implemente a interface
# plugins/geo-resolver/plugin.go
```

Exemplo de manifest:
```json
{
  "id": "geo-resolver",
  "version": "1.0.0",
  "type": "enrichment",
  "required": true,
  "parallel_safe": true,
  "max_memory_bytes": 52428800,
  "max_duration_seconds": 30,
  "inputSchema": {
    "required": ["place_reference"],
    "properties": {
      "place_reference": {"type": "string"}
    }
  },
  "outputSchema": {
    "properties": {
      "latitude": {"type": "number"},
      "longitude": {"type": "number"}
    }
  }
}
```

## 📊 Observabilidade

### Métricas Prometheus

```
geee_plugin_duration_seconds
geee_plugin_errors_total
geee_plugin_memory_bytes
geee_active_executions
geee_state_size_bytes
```

### Tracing

Toda execução gera traces OpenTelemetry compatíveis.

### Logs Estruturados

```json
{
  "timestamp": "2025-01-11T10:30:45Z",
  "level": "info",
  "execution_id": "exec-abc123",
  "plugin_id": "speech-to-text",
  "event": "plugin_completed",
  "duration_ms": 1200
}
```

## 🧪 Testes

```bash
# Testes unitários
make test

# Testes com race detector (obrigatório no CI)
make test-race

# Coverage report
make test-coverage
```

## 📖 Documentação

- [PRP.md](PRP.md) - Product Requirements Prompt completo
- [CRITICAL_IMPROVEMENTS.md](CRITICAL_IMPROVEMENTS.md) - Melhorias críticas implementadas

## 🔒 Garantias de Qualidade

- ✅ Zero race conditions
- ✅ Cleanup automático 100% garantido
- ✅ Limites de recursos respeitados
- ✅ Erros estruturados com suggestions
- ✅ Observabilidade completa

## 📝 Critérios de Aceite

### Funcionais
- Core funciona sem plugins de domínio
- Plugins opcionais não abortam pipeline
- IA só usa capacidades registradas
- Output validado por schema

### Não-Funcionais
- Memória idle < 50MB
- Estado < 10MB por execução
- Cleanup 100% garantido
- Health check < 100ms

## 🤝 Contribuindo

```bash
# Fork & clone
git clone <your-fork>

# Crie branch
git checkout -b feature/my-feature

# Desenvolva
make dev

# Teste
make test-race
make lint

# Commit
git commit -m "feat: add my feature"

# Push
git push origin feature/my-feature
```

## 📄 Licença

[MIT License](LICENSE)

## 🔗 Links

- [Go Documentation](https://golang.org/doc/)
- [OpenTelemetry](https://opentelemetry.io/)
- [Prometheus](https://prometheus.io/)

---

**Status**: 🚧 Em desenvolvimento (v1.1)

Desenvolvido com ❤️ usando Go
