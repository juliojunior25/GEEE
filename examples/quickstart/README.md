# Quickstart Guide - GEEE

Este exemplo demonstra um pipeline completo de extração e enriquecimento de dados de usuários.

## O que este exemplo faz?

1. **Extrai informações** de texto não estruturado (email e telefone)
2. **Transforma campos** para nomenclatura padronizada
3. **Busca dados adicionais** de uma API externa
4. **Gera relatório** formatado com todas as informações

## Arquivos

```
examples/quickstart/
├── pipeline.yaml      # Configuração do pipeline
├── input.json        # Dados de entrada
├── run.sh            # Script para executar
└── README.md         # Esta documentação
```

## Pré-requisitos

```bash
# Certifique-se de que o GEEE está instalado
cd /Users/julio.junior/developer/GEEE
make build-local

# Verifique a instalação
./bin/geee --version
```

## Opção 1: CLI (Linha de Comando)

### Passo 1: Execute o pipeline

```bash
cd examples/quickstart

# Execução básica (output no terminal)
../../bin/geee run \
  --config pipeline.yaml \
  --input input.json

# Execução com output em arquivo
../../bin/geee run \
  --config pipeline.yaml \
  --input input.json \
  --output result.json

# Ver o resultado
cat result.json | jq '.'
```

### Passo 2: Entenda o resultado

O resultado terá a estrutura:

```json
{
  "success": true,
  "execution_id": "uuid-gerado",
  "data": {
    "user_name": "João Silva",
    "user_email": "joao.silva@exemplo.com.br",
    "user_phone": "+55 11 98765-4321",
    "result": "User Profile Report\n==================\n...",
    "body": {
      "id": 1,
      "username": "Bret",
      "company": {...}
    }
  },
  "metadata": {
    "start_time": "...",
    "end_time": "...",
    "duration_ms": 123
  }
}
```

### Passo 3: Ver apenas o relatório gerado

```bash
../../bin/geee run \
  --config pipeline.yaml \
  --input input.json \
  --output result.json

cat result.json | jq -r '.data.result'
```

Output esperado:
```
User Profile Report
===================
Name: João Silva
Email: joao.silva@exemplo.com.br
Phone: +55 11 98765-4321

Additional Info:
- User ID: 1
- Username: Bret
- Company: Romaguera-Crona
```

## Opção 2: HTTP Server

### Passo 1: Inicie o servidor

```bash
# Terminal 1: Inicie o servidor
cd /Users/julio.junior/developer/GEEE
make server

# Aguarde a mensagem: "Server started"
```

### Passo 2: Execute o pipeline via API

```bash
# Terminal 2: Envie requisição
cd examples/quickstart

curl -X POST http://localhost:8080/run \
  -H "Content-Type: application/json" \
  -d "{
    \"config\": \"$(cat pipeline.yaml | sed 's/"/\\"/g' | tr '\n' ' ')\",
    \"input\": $(cat input.json)
  }" | jq '.'
```

### Passo 3: Verifique os endpoints disponíveis

```bash
# Health check
curl http://localhost:8080/health | jq '.'

# Lista de plugins
curl http://localhost:8080/plugins | jq '.'

# Métricas Prometheus
curl http://localhost:8080/metrics
```

## Script Automatizado

Execute tudo automaticamente:

```bash
./run.sh
```

Este script irá:
1. ✅ Verificar se o GEEE está compilado
2. ✅ Executar o pipeline via CLI
3. ✅ Salvar resultado em `result.json`
4. ✅ Exibir o relatório formatado
5. ✅ Mostrar estatísticas de execução

## Personalize o Exemplo

### Modifique o Input

Edite `input.json` com seus próprios dados:

```json
{
  "name": "Seu Nome",
  "text": "Contato: seu.email@dominio.com ou +55 21 99999-8888",
  "source": "teste"
}
```

### Modifique o Pipeline

Edite `pipeline.yaml`:

```yaml
# Adicione mais steps
plugins:
  - id: json-transformer
    config:
      mappings:
        - from: source
          to: data_source
  # ... adicione mais plugins
```

### Plugins Disponíveis

| Plugin | Função | Exemplo |
|--------|--------|---------|
| `json-transformer` | Transforma/renomeia campos | `from: "old" → to: "new"` |
| `regex-extractor` | Extrai padrões com regex | `pattern: "\\d+"` |
| `http-fetcher` | Busca dados de APIs | `url: "https://..."` |
| `template-renderer` | Renderiza templates Go | `template: "{{.field}}"` |

## Troubleshooting

### Erro: "binary not found"

```bash
cd /Users/julio.junior/developer/GEEE
make build-local
```

### Erro: "plugin not found"

Verifique os plugins disponíveis:
```bash
../../bin/geee plugins
```

### Erro de execução no pipeline

Verifique o YAML está correto:
```bash
# Valida sintaxe YAML
cat pipeline.yaml | python3 -c 'import yaml, sys; yaml.safe_load(sys.stdin)'
```

### Ver logs detalhados

```bash
../../bin/geee run \
  --config pipeline.yaml \
  --input input.json \
  --verbose
```

## Próximos Passos

1. **Explore outros exemplos**: `examples/`
2. **Crie seus próprios plugins**: `make plugin-new NAME=my-plugin`
3. **Integre com seu sistema**: Use a API HTTP
4. **Monitore com Prometheus**: `curl http://localhost:8080/metrics`

## Documentação

- [README Principal](../../README.md)
- [Plano de Implementação](../../IMPLEMENTATION_PLAN.md)
- [Criando Plugins](../../docs/creating-plugins.md)
