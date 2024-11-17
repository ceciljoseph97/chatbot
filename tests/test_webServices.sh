set -e

cd "$(dirname "$0")/.." || exit

docker rm -f perichatweb_test || true

docker network create perichat_test_net || true

echo "Building perichatweb Docker image..."
docker-compose build perichatweb

if [[ "$(docker images -q perichatweb-image)" == "" ]]; then
    echo "Error: perichatweb-image not found!"
    exit 1
fi

echo "Docker image built successfully."

echo "Running perichatweb container for web-based test..."
docker run -d --name perichatweb_test --network perichat_test_net perichatweb-image

sleep 5

LOG_FILE="./tests/webService_test.log"
WEB_CHAT_VER_FILE="./tests/web_chatVer.json"

mkdir -p ./tests
> "$LOG_FILE"

docker exec perichatweb_test apk add --no-cache curl jq

echo "Performing web-Service based test..." | tee -a "$LOG_FILE"

while IFS= read -r line
do
  MESSAGE_JSON="$line"
  RESPONSE=$(docker exec perichatweb_test curl -s -X POST -H "Content-Type: application/json" -d "$MESSAGE_JSON" http://localhost:8080/chat)
  
  echo "Sent: $MESSAGE_JSON" | tee -a "$LOG_FILE"
  echo "Received: $RESPONSE" | tee -a "$LOG_FILE"

  if echo "$RESPONSE" | jq -e '.reply' >/dev/null 2>&1; then
    echo "Response schema check passed." | tee -a "$LOG_FILE"
  else
    echo "Response schema check failed." | tee -a "$LOG_FILE"
    exit 1
  fi

done < "$WEB_CHAT_VER_FILE"

echo "Cleaning up test container and network..."
docker stop perichatweb_test
docker rm perichatweb_test
docker network rm perichat_test_net || true

echo "Web Service test completed successfully. Check $LOG_FILE for details."
