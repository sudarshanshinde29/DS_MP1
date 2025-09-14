#!/usr/bin/env bash
USER="root"
BASE="fa25-cs425-10"

CMD="export GOTOOLCHAIN=auto \
      go version"  # command to execute passed as arg



for i in $(seq -w 01 10); do
  HOST="${BASE}${i}.cs.illinois.edu"
  echo "=== $HOST ==="
  ssh -o BatchMode=yes "$USER@$HOST" "$CMD"
done
