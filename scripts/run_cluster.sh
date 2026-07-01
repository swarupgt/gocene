#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
BINARY="$ROOT_DIR/gocene"

if [[ ! -x "$BINARY" ]]; then
  echo "Binary not found at $BINARY. Run 'go build -o gocene .' first."
  exit 1
fi

for i in 0 1 2; do
  env_file="$ROOT_DIR/.env.node$i"
  if [[ ! -f "$env_file" ]]; then
    echo "Missing env file: $env_file"
    exit 1
  fi
done

PIDS=()

cleanup() {
  echo ""
  echo "Shutting down cluster..."
  for pid in "${PIDS[@]}"; do
    kill "$pid" 2>/dev/null || true
  done
  wait
  echo "All nodes stopped."
}

trap cleanup SIGINT SIGTERM

start_node() {
  local i="$1"
  local env_file="$ROOT_DIR/.env.node$i"
  echo "Starting node$i (env: $env_file)..."
  env $(grep -v '^\s*#' "$env_file" | grep -v '^\s*$' | xargs) "$BINARY" &
  PIDS+=($!)
}

# Polls http_addr/status until it responds with an elected Raft leader, so
# nodes joining via RAFT_JOIN_ADDRESS don't hit a leader that isn't up yet.
wait_for_leader() {
  local http_addr="$1"
  local name="$2"
  local timeout_secs="${3:-15}"
  local elapsed=0

  echo "Waiting for $name ($http_addr) to become ready..."
  while (( elapsed < timeout_secs )); do
    body="$(curl -sf -m 1 "http://$http_addr/status" 2>/dev/null || true)"
    if [[ -n "$body" ]] && echo "$body" | grep -q '"leader":{"node_id":"[^"]\+"'; then
      echo "$name is ready."
      return 0
    fi
    sleep 1
    elapsed=$((elapsed + 1))
  done

  echo "Timed out waiting for $name to become ready after ${timeout_secs}s."
  return 1
}

# node0 bootstraps the cluster and elects itself leader; node1 and node2 join
# it over HTTP via RAFT_JOIN_ADDRESS, so node0 must be ready before they start.
start_node 0

node0_http_addr="$(grep -E '^RAFT_SELF_HTTP_ADDRESS=' "$ROOT_DIR/.env.node0" | cut -d= -f2-)"
if ! wait_for_leader "$node0_http_addr" "node0"; then
  cleanup
  exit 1
fi

start_node 1
start_node 2

echo "Cluster running. PIDs: ${PIDS[*]}"
echo "Press Ctrl+C to stop."

wait
