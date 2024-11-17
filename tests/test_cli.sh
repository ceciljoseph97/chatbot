set -e

cd "$(dirname "$0")/.." || exit

if ! command -v docker-compose &> /dev/null
then
    echo "docker-compose could not be found. Please install it."
    exit 1
fi

echo "Building perichat Docker image..."
docker-compose build perichat

if [[ "$(docker images -q perichat-image)" == "" ]]; then
    echo "Error: perichat-image not found!"
    exit 1
fi

echo "Docker image built successfully."

LOG_FILE="./tests/cli_test.log"
CLI_CHAT_VER_FILE="./tests/cli_chatVer.txt"

mkdir -p ./tests
> "$LOG_FILE"

QUESTIONS=$(cat "$CLI_CHAT_VER_FILE")
SESSION_INPUT="$QUESTIONS
/geronimo"

echo "Starting CLI-based test..." | tee -a "$LOG_FILE"
echo "$SESSION_INPUT" | docker-compose run perichat | tee -a "$LOG_FILE"

echo "CLI-based test completed. Performing schema checks..." | tee -a "$LOG_FILE"
grep -E '^Bot:' "$LOG_FILE" | sed 's/^Bot: //'
SCHEMA_CHECK_PASSED=true
while IFS= read -r reply
do
  if [[ -z "$reply" ]]; then
    echo "Schema check failed: Empty reply detected." | tee -a "$LOG_FILE"
    SCHEMA_CHECK_PASSED=false
    break
  fi
done < ./tests/cli_bot_replies.txt

if $SCHEMA_CHECK_PASSED; then
  echo "CLI schema check passed." | tee -a "$LOG_FILE"
else
  echo "CLI schema check failed." | tee -a "$LOG_FILE"
  exit 1
fi

echo "All CLI tests passed successfully. Check $LOG_FILE for details." | tee -a "$LOG_FILE"
