#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
IMAGES_DIR="${IMAGES_DIR:-../images}"
RESULTS_DIR="results"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

cd "$SCRIPT_DIR"
mkdir -p "$RESULTS_DIR"

echo "═══════════════════════════════════════════════════════════"
echo "  LOAD TEST SUITE — Go API POST /api/search"
echo "═══════════════════════════════════════════════════════════"
echo ""

echo "→ Checking Go API at ${BASE_URL}..."
if ! curl -sf "${BASE_URL}/health" > /dev/null 2>&1; then
  echo "  ✗ Go API not reachable at ${BASE_URL}"
  echo "  Start services first: see RUN_COMMANDS.md"
  exit 1
fi
echo "  ✓ Go API is healthy"
echo ""

IMG_COUNT=$(ls "$IMAGES_DIR"/*.jpg 2>/dev/null | wc -l | tr -d ' ')
echo "→ Found ${IMG_COUNT} images in ${IMAGES_DIR}"
if [ "$IMG_COUNT" -lt 10 ]; then
  echo "  ✗ Need at least 10 images. Run the downloader first."
  exit 1
fi
echo ""

run_scenario() {
  local name="$1"
  echo "───────────────────────────────────────────────────────────"
  echo "  Running: ${name}"
  echo "───────────────────────────────────────────────────────────"
  echo ""
  k6 run \
    --env BASE_URL="$BASE_URL" \
    --env IMAGES_DIR="$IMAGES_DIR" \
    --env SCENARIO="$name" \
    k6-search.js
  echo ""
}

SCENARIO="${1:-all}"

case "$SCENARIO" in
  baseline|concurrency|spike|soak)
    run_scenario "$SCENARIO"
    ;;
  all)
    echo "Running all scenarios sequentially..."
    echo ""
    run_scenario "baseline"
    sleep 5
    run_scenario "concurrency"
    sleep 5
    run_scenario "spike"
    sleep 5
    run_scenario "soak"
    echo "═══════════════════════════════════════════════════════════"
    echo "  ALL SCENARIOS COMPLETE"
    echo "═══════════════════════════════════════════════════════════"
    echo ""
    echo "  Results in: ${RESULTS_DIR}/"
    echo ""
    ;;
  *)
    echo "Usage: $0 [baseline|concurrency|spike|soak|all]"
    exit 1
    ;;
esac
