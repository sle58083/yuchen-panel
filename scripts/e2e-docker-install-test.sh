#!/usr/bin/env bash
set -euo pipefail
umask 077
skip() { echo "SKIPPED: $*"; exit 0; }
command -v docker >/dev/null || skip "docker unavailable"
docker compose version >/dev/null 2>&1 || skip "Compose v2 unavailable (BUG-L prerequisite)"
docker info >/dev/null 2>&1 || skip "Docker daemon unavailable"
[[ $EUID -eq 0 ]] || skip "root required"
for cmd in python3 rsync curl jq; do command -v "$cmd" >/dev/null || skip "$cmd unavailable"; done
repo=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
workspace=$(mktemp -d /tmp/yc-e2e-XXXXXXXX)
export APP_DIR="$workspace/app" CONFIG_DIR="$workspace/config"
project_suffix=${workspace##*-}
export COMPOSE_PROJECT_NAME="yc-e2e-${project_suffix,,}"
export YC_API_CONTAINER_NAME="$COMPOSE_PROJECT_NAME-api" YC_WEB_CONTAINER_NAME="$COMPOSE_PROJECT_NAME-web"
export YC_DOCKER_ISOLATED=1 YC_INSTALL_MODE=docker AUTO_AGENT=false
export YC_LOCAL_SERVER_IP=127.0.0.1 YC_LOCAL_SERVER_HOST=127.0.0.1
# Allocate three unused loopback ports. Compose will fail safely if another process races us.
read -r API_PORT WEB_PORT PANEL_PORT < <(python3 - <<'PY_PORTS'
import socket
sockets=[socket.socket() for _ in range(3)]
for s in sockets: s.bind(('127.0.0.1',0))
print(*(s.getsockname()[1] for s in sockets))
for s in sockets: s.close()
PY_PORTS
)
export API_PORT WEB_PORT PANEL_PORT
log_dir=${YC_E2E_LOG_DIR:-$repo/.audit/task8-e2e}
mkdir -p "$log_dir" "$workspace/source"
chmod 755 "$log_dir"
log="$log_dir/docker-install.log"
cleanup() {
  if [[ -f "$APP_DIR/.env" && -f "$APP_DIR/docker-compose.yml" ]]; then
    (unset YC_JWT_SECRET YC_ADMIN_USERNAME YC_ADMIN_PASSWORD; cd "$APP_DIR"; docker compose --env-file "$APP_DIR/.env" down --volumes) >/dev/null 2>&1 || true
  fi
  # Only the canonical mktemp directory is removed.
  [[ "$workspace" == /tmp/yc-e2e-* ]] && rm -rf -- "$workspace"
}
trap cleanup EXIT
fail() { echo "FAIL: $*"; tail -n 100 "$log"; exit 1; }
rsync -a --exclude .git --exclude .env --exclude node_modules --exclude .audit --exclude dist-release "$repo/" "$workspace/source/"
[[ ! -f "$workspace/source/.env" && ! -f "$APP_DIR/.env" ]] || fail "not a fresh source/runtime directory"
# Run the actual installer and source Dockerfiles; no Release binary is used.
set +e
bash "$workspace/source/deploy/install.sh" >"$workspace/raw.log" 2>&1
rc=$?
set -e
# Preserve the complete log, redacting only credentials before publication.
python3 - "$workspace/raw.log" "$log" <<'PY_LOG'
import re,sys
text=open(sys.argv[1],encoding='utf-8',errors='replace').read()
text=re.sub(r'(?m)^(Password:|API Token:).*$',r'\1 [REDACTED]',text)
open(sys.argv[2],'w',encoding='utf-8').write(text)
PY_LOG
chmod 644 "$log" # Complete redacted log is readable by CI artifact uploader.
[[ $rc -eq 0 ]] || fail "1 installer exit=$rc"
echo "PASS 1: installer exit 0"
! grep -Fq ContainerConfig "$log" || fail "2 Compose v1 regression"
echo "PASS 2: no ContainerConfig"
! grep -Fq 'Unable to locate package' "$log" || fail "3 package regression"
echo "PASS 3: no missing-package error"
! grep -Fq 'configuration validation failed' "$log" || fail "4 source validation regression"
echo "PASS 4: fresh source config validated"
[[ -f "$APP_DIR/.env" ]] || fail "5 runtime .env missing"
secret_in_env=$(sed -n 's/^YC_JWT_SECRET=//p' "$APP_DIR/.env")
[[ ${#secret_in_env} -ge 32 ]] || fail "5 JWT secret too short"
echo "PASS 5: generated secret at least 32 characters"
health=$(curl -fsS "http://127.0.0.1:$API_PORT/api/health")
[[ -n $(jq -er '.version | select(type == "string" and length > 0)' <<<"$health") ]] || fail "6 health version empty"
echo "PASS 6: source-built container health version=$(jq -r .version <<<"$health")"
container=$(cd "$APP_DIR"; docker compose --env-file "$APP_DIR/.env" ps -q yuchen-panel-api)
[[ -n "$container" && $(docker inspect --format '{{.State.Status}}' "$container") == running ]] || fail "7 API container not running"
echo "PASS 7: Compose API running"
secret_in_container=$(docker inspect --format '{{range .Config.Env}}{{println .}}{{end}}' "$container" | sed -n 's/^YC_JWT_SECRET=//p')
[[ "$secret_in_env" == "$secret_in_container" ]] || fail "8 placeholder leaked into container"
echo "PASS 8: container secret equals runtime .env byte-for-byte"
! grep -Fq '本次安装已轮换 JWT 签名密钥' "$log" || fail "fresh install incorrectly reported key rotation"
bash "$repo/scripts/test-nginx-proxy.sh"
echo "E2E_PASS: eight assertions; source-built Docker image=$(docker inspect --format '{{.Image}}' "$container"); complete redacted log=$log"
