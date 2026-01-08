#!/bin/bash

# ==============================================================================
# Configuration and Variables
# ==============================================================================
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

API_URL="http://localhost:8090"
BANK_URL="http://localhost:8080"
PROJECT_ROOT="../../" 

# ==============================================================================
# Prerequisites Check
# ==============================================================================
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: 'jq' is not installed.${NC}"
    exit 1
fi

# ==============================================================================
# Environment Initialization (Docker Compose)
# ==============================================================================
echo -e "\n[1/5] Starting Environment (API + Bank Simulator)..."
pushd $PROJECT_ROOT > /dev/null

docker-compose up -d --build

if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to start containers via docker-compose.${NC}"
    popd > /dev/null
    exit 1
fi
popd > /dev/null

# ==============================================================================
# Healthcheck (Waiting for Services)
# ==============================================================================
echo "Waiting for Bank Simulator ($BANK_URL)..."
for i in {1..30}; do
    if curl -s $BANK_URL > /dev/null; then
        echo -e "${GREEN}Bank Simulator online!${NC}"
        break
    fi
    sleep 1
done

echo "Waiting for API ($API_URL)..."
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
    echo -e "${RED}API failed to start or is not responding.${NC}"
    echo -e "${RED}Container logs:${NC}"
    pushd $PROJECT_ROOT > /dev/null
    docker-compose logs --tail=20
    docker-compose down
    popd > /dev/null
    exit 1
fi

# ==============================================================================
# Helper Test Function
# ==============================================================================
run_test() {
    local test_name=$1
    local curl_cmd=$2
    local expected_status=$3
    local check_field=$4
    local check_value=$5

    echo -n "Testing: $test_name ... " >&2

    TMP_BODY=$(mktemp)
    HTTP_CODE=$(eval "$curl_cmd -o $TMP_BODY -w '%{http_code}'")

    if [ "$HTTP_CODE" -ne "$expected_status" ]; then
        echo -e "${RED}FAILED (Status: $HTTP_CODE, Expected: $expected_status)${NC}" >&2
        echo "Response Body:" >&2
        cat $TMP_BODY >&2
        echo "" >&2
        rm "$TMP_BODY"
        return 1
    fi

    if [ ! -z "$check_field" ]; then
        actual_value=$(cat $TMP_BODY | jq -r "$check_field")
        if [[ "$actual_value" != "$check_value" ]]; then
            echo -e "${RED}FAILED (Field '$check_field' value: '$actual_value', Expected: '$check_value')${NC}" >&2
            rm "$TMP_BODY"
            return 1
        fi
    fi

    echo -e "${GREEN}SUCCESS${NC}" >&2
    
    cat $TMP_BODY | jq -r '.id // empty'
    
    rm "$TMP_BODY"
}

# ==============================================================================
# Scenario Execution
# ==============================================================================
echo -e "\n[3/5] Running Scenarios..."

# Scenario 1: Success (Card ending in odd number)
PAYMENT_ID=$(run_test "Authorized Payment" \
    "curl -s -X POST $API_URL/api/payments -H 'Content-Type: application/json' -d '{\"card_number\": \"1234567890123451\", \"expiry_month\": 12, \"expiry_year\": 2030, \"currency\": \"USD\", \"amount\": 1000, \"cvv\": \"123\"}'" \
    200 ".payment_status" "Authorized")

# Scenario 2: GET
if [ ! -z "$PAYMENT_ID" ] && [ "$PAYMENT_ID" != "null" ]; then
    run_test "Retrieve Payment (GET)" \
        "curl -s -X GET $API_URL/api/payments/$PAYMENT_ID" \
        200 ".payment_status" "Authorized" > /dev/null
else
    echo -e "${RED}Skipping GET test (No ID captured)${NC}" >&2
fi

# Scenario 3: Decline (Card ending in even number)
run_test "Declined Payment" \
    "curl -s -X POST $API_URL/api/payments -H 'Content-Type: application/json' -d '{\"card_number\": \"1234567890123452\", \"expiry_month\": 12, \"expiry_year\": 2030, \"currency\": \"EUR\", \"amount\": 500, \"cvv\": \"123\"}'" \
    200 ".payment_status" "Declined" > /dev/null

# Scenario 4: Validation (Invalid Currency)
# NOTE: Changed from BRL to JPY
run_test "Validation Error" \
    "curl -s -X POST $API_URL/api/payments -H 'Content-Type: application/json' -d '{\"card_number\": \"1234567890123451\", \"expiry_month\": 12, \"expiry_year\": 2030, \"currency\": \"JPY\", \"amount\": 1000, \"cvv\": \"123\"}'" \
    400 ".payment_status" "Rejected" > /dev/null

# Scenario 5: Bank Error (Card ending in 0)
run_test "Bank Unavailable" \
    "curl -s -X POST $API_URL/api/payments -H 'Content-Type: application/json' -d '{\"card_number\": \"1234567890123450\", \"expiry_month\": 12, \"expiry_year\": 2030, \"currency\": \"USD\", \"amount\": 1000, \"cvv\": \"123\"}'" \
    502 ".payment_status" "Failed" > /dev/null

# ==============================================================================
# Cleanup and Shutdown
# ==============================================================================
echo -e "\n[4/5] Cleaning up environment..."
pushd $PROJECT_ROOT > /dev/null
docker-compose down
popd > /dev/null

echo -e "\n[5/5] Finished."