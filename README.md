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

## 🎓 Exemplo Completo - Passo a Passo

### Opção 1: Teste Rápido (Recomendado)

Execute o exemplo quickstart automatizado:

```bash
# Build do GEEE
make build-local

# Execute o exemplo completo
cd examples/quickstart
./run.sh
```

O script irá:
- ✅ Verificar instalação
- ✅ Executar pipeline de exemplo
- ✅ Mostrar resultados formatados

### Opção 2: Execução Manual (CLI)

```bash
# 1. Crie um pipeline (pipeline.yaml)
cat > my-pipeline.yaml << 'EOF'
name: exemplo-simples
description: Extrai email e gera relatório
plugins:
  - id: regex-extractor
    config:
      patterns:
        - name: email
          pattern: '\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b'

  - id: template-renderer
    config:
      template: "Email encontrado: {{.email}}"
    depends_on:
      - regex-extractor
EOF

# 2. Crie um input (input.json)
cat > input.json << 'EOF'
{
  "text": "Contato: usuario@exemplo.com"
}
EOF

# 3. Execute
./bin/geee run --config my-pipeline.yaml --input input.json

# 4. Ver resultado formatado
./bin/geee run --config my-pipeline.yaml --input input.json --output result.json
cat result.json | jq -r '.data.result'
# Output: "Email encontrado: usuario@exemplo.com"
```

### Opção 3: HTTP Server

```bash
# Terminal 1: Inicie o servidor
make server

# Terminal 2: Envie requisição
curl -X POST http://localhost:8080/run \
  -H "Content-Type: application/json" \
  -d '{
    "config": "name: test\nplugins:\n  - id: json-transformer\n    config:\n      mappings:\n        - from: input\n          to: output",
    "input": {"input": "hello"}
  }' | jq '.'

# Verifique health
curl http://localhost:8080/health | jq '.'

# Liste plugins disponíveis
curl http://localhost:8080/plugins | jq '.'
```

### Exemplo Real Completo

Veja um exemplo completo funcionando em: **`examples/quickstart/`**

Este exemplo demonstra:
- 📧 Extração de email e telefone com regex
- 🔄 Transformação de campos
- 🌐 Busca de dados em API externa
- 📝 Geração de relatório formatado

```bash
cd examples/quickstart
cat README.md  # Documentação completa
./run.sh       # Execução automatizada
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
