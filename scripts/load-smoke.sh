#!/usr/bin/env bash
# Serangan ringan ke /healthz — bukan benchmark sejati. Untuk beban puncak pakai: hey, k6, vegeta.
set -euo pipefail
BASE_URL="${1:-http://127.0.0.1:8080}"
N="${2:-3000}"
echo "GET $BASE_URL/healthz sebanyak $N kali..."
ok=0
for ((i=1; i<=N; i++)); do
  if curl -fsS -o /dev/null --max-time 5 "$BASE_URL/healthz"; then
    ((ok++)) || true
  else
    echo "gagal pada request ke-$i" >&2
    exit 1
  fi
done
echo "selesai: $ok/$N OK"
