#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="/root/MP/DS_MP1"
WORKER_REL="worker/main.go"

VM_HOSTNAME="fa25-cs425-10"    # e.g. fa25-cs425-1001.cs.illinois.edu
DOMAIN="cs.illinois.edu"
COUNT="${COUNT:-10}"
REMOTE_LOG_DIR="/root/logs"
PARALLEL="${PARALLEL:-5}"

deploy_host() {
  local host="$1"
  local n="$2"
  local port=$((6000 + n))         # 6001..6010
  local label="vm${n}"
  local glob="vm${n}.log"
  local out="/root/worker_${n}.out"

  echo "[$host] deploy: label=$label port=$port glob=$glob"

  ssh -o BatchMode=yes -o StrictHostKeyChecking=accept-new "root@${host}" \
    REPO_DIR="$REPO_DIR" WORKER_REL="$WORKER_REL" \
    REMOTE_LOG_DIR="$REMOTE_LOG_DIR" PORT="$port" LABEL="$label" GLOB="$glob" OUT="$out" \
    bash -s <<'REMOTE'
set -e

# repo present? pull; else clone (default branch)
if [[ -d "$REPO_DIR/.git" ]]; then
  git -C "$REPO_DIR" fetch --all --prune
  git -C "$REPO_DIR" pull --ff-only
else
  mkdir -p "$(dirname "$REPO_DIR")"
  git clone https://github.com/sudarshanshinde29/DS_MP1.git "$REPO_DIR"
fi

# ensure logs dir
mkdir -p "$REMOTE_LOG_DIR"

# start worker (no pidfile, no kill/verify)
cd "$REPO_DIR"
nohup bash -lc "go run './$WORKER_REL' \
  -addr ':$PORT' \
  -logdir '$REMOTE_LOG_DIR' \
  -glob '$GLOB' \
  -label '$LABEL' \
  2>&1 | cat >> '$OUT'" >/dev/null 2>&1 &
echo "Started: $LABEL on :$PORT (log -> $OUT)"
REMOTE
}

# fan-out across hosts
i=0
for XX in $(seq -w 01 "$COUNT"); do
  host="${VM_HOSTNAME}${XX}.${DOMAIN}"
  n=$((10#$XX))  # 01->1
  deploy_host "$host" "$n" &
  (( ++i % PARALLEL == 0 )) && wait
done
wait
echo "All workers deployed."
