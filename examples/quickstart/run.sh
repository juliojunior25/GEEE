#!/bin/bash

# Script de teste automatizado do GEEE Quickstart
# Usage: ./run.sh

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Banner
echo -e "${BLUE}"
cat << "EOF"
   _____ ______ ______ ______
  / ____|  ____|  ____|  ____|
 | |  __| |__  | |__  | |__
 | | |_ |  __| |  __| |  __|
 | |__| | |____| |____| |____
  \_____|______|______|______|

  Generic Extraction & Enrichment Engine
  Quickstart Example
EOF
echo -e "${NC}"

# Step 1: Check if binary exists
echo -e "${YELLOW}[1/5]${NC} Verificando binário GEEE..."
GEEE_BIN="../../bin/geee"

if [ ! -f "$GEEE_BIN" ]; then
    echo -e "${RED}✗${NC} Binário não encontrado em $GEEE_BIN"
    echo -e "${YELLOW}→${NC} Compilando GEEE..."
    cd ../..
    make build-local
    cd examples/quickstart
    echo -e "${GREEN}✓${NC} Binário compilado com sucesso"
else
    echo -e "${GREEN}✓${NC} Binário encontrado"
fi

# Step 2: Verify files
echo -e "\n${YELLOW}[2/5]${NC} Verificando arquivos de exemplo..."
if [ ! -f "pipeline.yaml" ]; then
    echo -e "${RED}✗${NC} pipeline.yaml não encontrado"
    exit 1
fi

if [ ! -f "input.json" ]; then
    echo -e "${RED}✗${NC} input.json não encontrado"
    exit 1
fi
echo -e "${GREEN}✓${NC} Arquivos encontrados"

# Step 3: Show input
echo -e "\n${YELLOW}[3/5]${NC} Input de exemplo:"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
cat input.json | jq '.' 2>/dev/null || cat input.json
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Step 4: Execute pipeline
echo -e "\n${YELLOW}[4/5]${NC} Executando pipeline..."
echo -e "${BLUE}→${NC} Comando: $GEEE_BIN run --config pipeline.yaml --input input.json --output result.json"

if $GEEE_BIN run --config pipeline.yaml --input input.json --output result.json; then
    echo -e "${GREEN}✓${NC} Pipeline executado com sucesso!"
else
    echo -e "${RED}✗${NC} Falha na execução do pipeline"
    exit 1
fi

# Step 5: Show results
echo -e "\n${YELLOW}[5/5]${NC} Resultados:"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

if [ -f "result.json" ]; then
    # Check if jq is available
    if command -v jq &> /dev/null; then
        SUCCESS=$(cat result.json | jq -r '.success')
        DURATION=$(cat result.json | jq -r '.metadata.duration_ms')

        echo -e "${GREEN}✓ Success:${NC} $SUCCESS"
        echo -e "${GREEN}✓ Duration:${NC} ${DURATION}ms"
        echo ""
        echo -e "${YELLOW}Relatório Gerado:${NC}"
        echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        cat result.json | jq -r '.data.result'
        echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

        echo -e "\n${YELLOW}Dados Extraídos:${NC}"
        echo -e "${GREEN}✓ Email:${NC} $(cat result.json | jq -r '.data.user_email')"
        echo -e "${GREEN}✓ Phone:${NC} $(cat result.json | jq -r '.data.user_phone')"
        echo -e "${GREEN}✓ Name:${NC} $(cat result.json | jq -r '.data.user_name')"

        echo -e "\n${YELLOW}Resultado completo salvo em:${NC} result.json"
        echo -e "${BLUE}→${NC} Ver JSON completo: ${YELLOW}cat result.json | jq '.'${NC}"
    else
        echo -e "${YELLOW}⚠${NC} jq não instalado, mostrando resultado bruto:"
        cat result.json
    fi
else
    echo -e "${RED}✗${NC} Arquivo result.json não foi gerado"
    exit 1
fi

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Summary
echo -e "\n${GREEN}🎉 Teste concluído com sucesso!${NC}"
echo -e "\n${YELLOW}Próximos passos:${NC}"
echo "  1. Edite input.json com seus dados"
echo "  2. Modifique pipeline.yaml para customizar"
echo "  3. Execute: ./run.sh"
echo -e "\n${YELLOW}Ou teste o HTTP Server:${NC}"
echo "  1. Terminal 1: cd ../.. && make server"
echo "  2. Terminal 2: curl -X POST http://localhost:8080/run ..."
echo -e "\n${BLUE}Documentação completa: README.md${NC}"
