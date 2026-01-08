#!/bin/bash

# ==============================================================================
# Configurações
# ==============================================================================
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

API_URL="http://localhost:8090"
BANK_URL="http://localhost:8080"
PID_FILE="api_load.pid"
API_LOG="api_output.log" # Arquivo de log para debug
PROJECT_ROOT="../../" 

# ==============================================================================
# Inicialização
# ==============================================================================
echo -e "\n[1/4] Iniciando Bank Simulator..."
pushd $PROJECT_ROOT > /dev/null
docker-compose up -d
popd > /dev/null

echo "Aguardando Bank Simulator..."
for i in {1..30}; do
    if curl -s $BANK_URL > /dev/null; then
        echo -e "${GREEN}Bank Simulator online!${NC}"
        break
    fi
    sleep 1
done

echo -e "\n[2/4] Iniciando API (com logs em $API_LOG)..."
pushd $PROJECT_ROOT > /dev/null
# Agora salvamos o log para entender o erro
go run main.go > "test/load/$API_LOG" 2>&1 &
SERVER_PID=$!
popd > /dev/null

echo $SERVER_PID > $PID_FILE

echo "Aguardando API..."
API_READY=false
for i in {1..30}; do
    # Verifica se o processo ainda existe
    if ! kill -0 $SERVER_PID 2>/dev/null; then
        echo -e "${RED}API morreu antes de iniciar!${NC}"
        break
    fi

    if curl -s "$API_URL/ping" > /dev/null; then
        echo -e "${GREEN}API online!${NC}"
        API_READY=true
        break
    fi
    sleep 1
done

if [ "$API_READY" = false ]; then
    echo -e "${RED}API falhou. Mostrando últimos logs:${NC}"
    cat "$API_LOG"
    
    if [ -f $PID_FILE ]; then rm $PID_FILE; fi
    pushd $PROJECT_ROOT > /dev/null
    docker-compose down
    popd > /dev/null
    exit 1
fi

# ==============================================================================
# Execução do K6 (Via Docker)
# ==============================================================================
echo -e "\n[3/4] Executando K6 Load Test (Docker)..."

K6_DIR="$(pwd)/k6"

# --network="host" só funciona bem no Linux nativo.
# Se estiver no Mac/Windows, precisaria usar "host.docker.internal" no script do k6.
# Assumindo Linux baseado no seu path (/home/luzucchi).
docker run --rm -i \
  --network="host" \
  -v "$K6_DIR:/scripts" \
  grafana/k6 run /scripts/script.js

EXIT_CODE=$?

# ==============================================================================
# Limpeza
# ==============================================================================
echo -e "\n[4/4] Limpando ambiente..."

# Verifica se a API ainda está rodando antes de matar
if kill -0 $SERVER_PID 2>/dev/null; then
    kill $SERVER_PID
else
    echo -e "${RED}AVISO: A API caiu durante o teste!${NC}"
    echo -e "${RED}Logs da API (últimas 20 linhas):${NC}"
    tail -n 20 "$API_LOG"
    EXIT_CODE=1 # Força erro se a API caiu
fi

if [ -f $PID_FILE ]; then rm $PID_FILE; fi

pushd $PROJECT_ROOT > /dev/null
docker-compose down
popd > /dev/null

if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}Teste de Carga Finalizado com Sucesso!${NC}"
else
    echo -e "${RED}Teste de Carga Falhou!${NC}"
fi

exit $EXIT_CODE