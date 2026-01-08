#!/bin/bash

# ==============================================================================
# Configurações e Variáveis
# ==============================================================================
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

API_URL="http://localhost:8090"
BANK_URL="http://localhost:8080"
PROJECT_ROOT="../../" 

# ==============================================================================
# Verificação de Pré-requisitos
# ==============================================================================
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Erro: 'jq' não está instalado.${NC}"
    exit 1
fi

# ==============================================================================
# Inicialização do Ambiente (Docker Compose)
# ==============================================================================
echo -e "\n[1/5] Iniciando Ambiente (API + Bank Simulator)..."
pushd $PROJECT_ROOT > /dev/null

docker-compose up -d --build

if [ $? -ne 0 ]; then
    echo -e "${RED}Falha ao subir containers via docker-compose.${NC}"
    popd > /dev/null
    exit 1
fi
popd > /dev/null

# ==============================================================================
# Healthcheck (Aguardando Serviços)
# ==============================================================================
echo "Aguardando Bank Simulator ($BANK_URL)..."
for i in {1..30}; do
    if curl -s $BANK_URL > /dev/null; then
        echo -e "${GREEN}Bank Simulator online!${NC}"
        break
    fi
    sleep 1
done

echo "Aguardando API ($API_URL)..."
API_READY=false
for i in {1..30}; do
    if curl -s "$API_URL/ping" > /dev/null; then
        echo -e "${GREEN}API online!${NC}"
        API_READY=true
        break
    fi
    sleep 1
done

if [ "$API_READY" = false ]; then
    echo -e "${RED}API falhou ao iniciar ou não está respondendo.${NC}"
    echo -e "${RED}Logs dos containers:${NC}"
    pushd $PROJECT_ROOT > /dev/null
    docker-compose logs --tail=20
    docker-compose down
    popd > /dev/null
    exit 1
fi

# ==============================================================================
# Função Auxiliar de Teste
# ==============================================================================
run_test() {
    local test_name=$1
    local curl_cmd=$2
    local expected_status=$3
    local check_field=$4
    local check_value=$5

    echo -n "Testando: $test_name ... " >&2

    TMP_BODY=$(mktemp)
    HTTP_CODE=$(eval "$curl_cmd -o $TMP_BODY -w '%{http_code}'")

    if [ "$HTTP_CODE" -ne "$expected_status" ]; then
        echo -e "${RED}FALHOU (Status: $HTTP_CODE, Esperado: $expected_status)${NC}" >&2
        echo "Response Body:" >&2
        cat $TMP_BODY >&2
        echo "" >&2
        rm "$TMP_BODY"
        return 1
    fi

    if [ ! -z "$check_field" ]; then
        actual_value=$(cat $TMP_BODY | jq -r "$check_field")
        if [[ "$actual_value" != "$check_value" ]]; then
            echo -e "${RED}FALHOU (Campo '$check_field' valor: '$actual_value', Esperado: '$check_value')${NC}" >&2
            rm "$TMP_BODY"
            return 1
        fi
    fi

    echo -e "${GREEN}SUCESSO${NC}" >&2
    
    cat $TMP_BODY | jq -r '.id // empty'
    
    rm "$TMP_BODY"
}

# ==============================================================================
# Execução dos Cenários
# ==============================================================================
echo -e "\n[3/5] Executando Cenários..."

# Cenário 1: Sucesso (Cartão final ímpar)
PAYMENT_ID=$(run_test "Pagamento Autorizado" \
    "curl -s -X POST $API_URL/api/payments -H 'Content-Type: application/json' -d '{\"card_number\": \"1234567890123451\", \"expiry_month\": 12, \"expiry_year\": 2030, \"currency\": \"USD\", \"amount\": 1000, \"cvv\": \"123\"}'" \
    200 ".payment_status" "Authorized")

# Cenário 2: GET
if [ ! -z "$PAYMENT_ID" ] && [ "$PAYMENT_ID" != "null" ]; then
    run_test "Recuperar Pagamento (GET)" \
        "curl -s -X GET $API_URL/api/payments/$PAYMENT_ID" \
        200 ".payment_status" "Authorized" > /dev/null
else
    echo -e "${RED}Skipping GET test (No ID captured)${NC}" >&2
fi

# Cenário 3: Recusa (Cartão final par)
run_test "Pagamento Recusado" \
    "curl -s -X POST $API_URL/api/payments -H 'Content-Type: application/json' -d '{\"card_number\": \"1234567890123452\", \"expiry_month\": 12, \"expiry_year\": 2030, \"currency\": \"EUR\", \"amount\": 500, \"cvv\": \"123\"}'" \
    200 ".payment_status" "Declined" > /dev/null

# Cenário 4: Validação (Moeda Inválida)
# ATENÇÃO: Mudado de BRL para JPY
run_test "Erro de Validação" \
    "curl -s -X POST $API_URL/api/payments -H 'Content-Type: application/json' -d '{\"card_number\": \"1234567890123451\", \"expiry_month\": 12, \"expiry_year\": 2030, \"currency\": \"JPY\", \"amount\": 1000, \"cvv\": \"123\"}'" \
    400 ".payment_status" "Rejected" > /dev/null

# Cenário 5: Erro do Banco (Cartão final 0)
run_test "Banco Indisponível" \
    "curl -s -X POST $API_URL/api/payments -H 'Content-Type: application/json' -d '{\"card_number\": \"1234567890123450\", \"expiry_month\": 12, \"expiry_year\": 2030, \"currency\": \"USD\", \"amount\": 1000, \"cvv\": \"123\"}'" \
    502 ".payment_status" "Failed" > /dev/null

# ==============================================================================
# Limpeza e Encerramento
# ==============================================================================
echo -e "\n[4/5] Limpando ambiente..."
pushd $PROJECT_ROOT > /dev/null
docker-compose down
popd > /dev/null

echo -e "\n[5/5] Finalizado."