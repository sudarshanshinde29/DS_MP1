#!/usr/bin/env bash
USER="root"
BASE="fa25-cs425-10"

CMD="cd /root/MP/DS_MP1 && export GOTOOLCHAIN=auto && git pull && go mod tidy"  #
#CMD="go mod tidy"

for i in $(seq -w 01 10); do
  HOST="${BASE}${i}.cs.illinois.edu"
  echo "=== $HOST ==="
  ssh -o BatchMode=yes "$USER@$HOST" "$CMD"
done
