#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# ==============================================================================
# Dirs and var configs
# ==============================================================================
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
K6_DIR="$SCRIPT_DIR/k6"

NETWORK_NAME="payment-gateway-challenge-go_default"
API_CONTAINER_NAME="api_load_test"
BANK_CONTAINER_NAME="bank_simulator"
IMAGE_NAME="payment-api:loadtest"

if [ ! -f "$K6_DIR/script.js" ]; then
    echo -e "${RED}ERROR: script.js not found in $K6_DIR${NC}"
    exit 1
fi

if [ ! -f "$PROJECT_ROOT/Dockerfile" ]; then
    echo -e "${RED}ERROR: Dockerfile not found at root: $PROJECT_ROOT/Dockerfile${NC}"
    exit 1
fi

# ==============================================================================
# 1. Env prep
# ==============================================================================
echo -e "\n[1/5] Cleaning up previous environment..."
docker rm -f $API_CONTAINER_NAME 2>/dev/null || true
docker-compose -f "$PROJECT_ROOT/docker-compose.yml" down 2>/dev/null || true

echo -e "\n[2/5] Building API image..."

docker build --network=host -t $IMAGE_NAME -f "$PROJECT_ROOT/Dockerfile" "$PROJECT_ROOT"

if [ $? -ne 0 ]; then
    echo -e "${RED}Docker image build failed!${NC}"
    exit 1
fi

# ==============================================================================
# 2. Start infra
# ==============================================================================
echo -e "\n[3/5] Starting dependencies..."
docker-compose -f "$PROJECT_ROOT/docker-compose.yml" up -d

echo "Waiting for Bank Simulator..."
for i in {1..30}; do
    if docker run --rm --network $NETWORK_NAME curlimages/curl -s "http://$BANK_CONTAINER_NAME:8080" > /dev/null; then
        echo -e "${GREEN}Bank Simulator online!${NC}"
        break
    fi
    sleep 1
done

# ==============================================================================
# 3. Start API
# ==============================================================================
echo -e "\n[4/5] Starting API in Container..."

docker run -d \
  --name $API_CONTAINER_NAME \
  --network $NETWORK_NAME \
  -e BANK_URL="http://$BANK_CONTAINER_NAME:8080" \
  $IMAGE_NAME

echo "Waiting for API..."
API_READY=false
for i in {1..30}; do
    if ! docker ps | grep -q $API_CONTAINER_NAME; then
        echo -e "${RED}API Container died! Logs:${NC}"
        docker logs $API_CONTAINER_NAME
        break
    fi

    if docker run --rm --network $NETWORK_NAME curlimages/curl -s "http://$API_CONTAINER_NAME:8090/ping" > /dev/null; then
        echo -e "${GREEN}API online!${NC}"
        API_READY=true
        break
    fi
    sleep 1
done

if [ "$API_READY" = false ]; then
    echo -e "${RED}Failed to start API.${NC}"
    docker-compose -f "$PROJECT_ROOT/docker-compose.yml" down
    exit 1
fi

# ==============================================================================
# 4. Execute K6 Load Test as sidecar
# ==============================================================================
echo -e "\n[5/5] Running K6 Load Test..."

docker run --rm -i \
  --network="container:$API_CONTAINER_NAME" \
  -v "$K6_DIR:/scripts" \
  grafana/k6 run /scripts/script.js

EXIT_CODE=$?

# ==============================================================================
# 5. Cleanup
# ==============================================================================
echo -e "\nCleaning up..."
if [ $EXIT_CODE -ne 0 ]; then
    echo -e "${RED}API Logs:${NC}"
    docker logs $API_CONTAINER_NAME | tail -n 20
fi

docker rm -f $API_CONTAINER_NAME > /dev/null
docker-compose -f "$PROJECT_ROOT/docker-compose.yml" down > /dev/null
docker rmi $IMAGE_NAME > /dev/null 2>/dev/null || true

if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}Load Test Finished SUCCESSFULLY!${NC}"
else
    echo -e "${RED}Load Test FAILED!${NC}"
fi

exit $EXIT_CODE