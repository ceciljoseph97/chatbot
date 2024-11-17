set -e
cleanup() {
  echo "Cleaning up test container and network..." | tee -a "$LOG_FILE"
  docker stop perichatweb_test || true
  docker rm perichatweb_test || true
  docker network rm perichat_test_net || true
}
trap cleanup EXIT
cd "$(dirname "$0")/.." || exit

LOG_FILE="$(pwd)/tests/webHttp_test.log"
mkdir -p "$(pwd)/tests"
> "$LOG_FILE"

docker rm -f perichatweb_test || true
if ! docker network inspect perichat_test_net >/dev/null 2>&1; then
    docker network create perichat_test_net
    echo "Docker network 'perichat_test_net' created." | tee -a "$LOG_FILE"
else
    echo "Docker network 'perichat_test_net' already exists. Skipping creation." | tee -a "$LOG_FILE"
fi

echo "Building perichatweb Docker image..." | tee -a "$LOG_FILE"
docker-compose build perichatweb

if [ "$(docker images -q perichatweb-image)" = "" ]; then
    echo "Error: perichatweb-image not found!" | tee -a "$LOG_FILE"
    exit 1
fi

echo "Docker image built successfully." | tee -a "$LOG_FILE"

echo "Running perichatweb container for port and Web HTTP Access test..." | tee -a "$LOG_FILE"
docker run -d --name perichatweb_test --network perichat_test_net -p 8080:8080 perichatweb-image

echo "Waiting for perichatweb to start..." | tee -a "$LOG_FILE"
MAX_RETRIES=30
RETRY_COUNT=0
until curl -s http://localhost:8080/ > /dev/null; do
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ "$RETRY_COUNT" -ge "$MAX_RETRIES" ]; then
        echo "perichatweb did not start in time." | tee -a "$LOG_FILE"
        exit 1
    fi
    echo "Waiting for perichatweb to start... ($RETRY_COUNT/$MAX_RETRIES)" | tee -a "$LOG_FILE"
    sleep 1
done

echo "perichatweb is up!" | tee -a "$LOG_FILE"

echo "Starting port and Web HTTP Access tests..." | tee -a "$LOG_FILE"
echo "Testing HTTP port 8080 with curl..." | tee -a "$LOG_FILE"

set +e
HTTP_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/)
CURL_EXIT_CODE=$?
set -e

if [ $CURL_EXIT_CODE -ne 0 ]; then
    echo "curl command failed with exit code $CURL_EXIT_CODE" | tee -a "$LOG_FILE"
    exit $CURL_EXIT_CODE
fi

echo "HTTP Response Code: $HTTP_RESPONSE" | tee -a "$LOG_FILE"

if [ "$HTTP_RESPONSE" -eq 200 ]; then
    echo "Port 8080 is available and responding with 200 OK." | tee -a "$LOG_FILE"
else
    echo "Port 8080 test failed: Received HTTP code $HTTP_RESPONSE" | tee -a "$LOG_FILE"
    exit 1
fi

echo "Web HTTP Access test passed." | tee -a "$LOG_FILE"
echo "Web HTTP Access test completed successfully. Check $LOG_FILE for details." | tee -a "$LOG_FILE"
