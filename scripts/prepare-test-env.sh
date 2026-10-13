#!/usr/bin/env bash
set -Eeuo pipefail
umask 077
[[ ${YC_ENV:-test} != production ]] || { echo 'Refusing to run with YC_ENV=production' >&2; exit 1; }
base_url=${1:-${SUBSCRIPTION_URL:-http://127.0.0.1:18088/sub/audit-demo-local}}
out=${2:-./client-test-samples}
mkdir -p -- "$out"
script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
command -v bash >/dev/null; command -v curl >/dev/null; command -v jq >/dev/null; command -v python3 >/dev/null
echo "Test-only preparation; production guard passed."
echo "Use branch fix/audit-client-compat and deploy its API/web build on a test machine before using a non-local URL."
bash "$script_dir/verify-clients.sh" "$base_url" "$out"
echo "Samples ready in: $(cd -- "$out" && pwd)"
