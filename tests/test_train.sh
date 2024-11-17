set -e
cleanup() {
  echo "Cleaning up test container and network..." | tee -a "$TRAIN_LOG"
  docker stop perichatweb_test || true
  docker rm perichatweb_test || true
  docker network rm perichat_test_net || true
}
trap cleanup EXIT
cd "$(dirname "$0")/.." || exit
TRAIN_LOG="$(pwd)/tests/train_test.log"
mkdir -p "$(pwd)/tests"
> "$TRAIN_LOG"
cd "cli/train" || exit
{
  echo "Starting training process..."
  go run train.go -d Corpus/en -m -o ../chat/PMFuncOverview.gob -config ../config_local.yaml
  echo "Training test completed. Check the tests directory for logs."
} | tee -a "$TRAIN_LOG"
