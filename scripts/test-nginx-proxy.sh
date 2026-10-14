#!/usr/bin/env bash
# Called after the isolated source installer; tests real Nginx without changing host /etc.
set -euo pipefail
: "${APP_DIR:?}" "${CONFIG_DIR:?}" "${API_PORT:?}" "${WEB_PORT:?}" "${PANEL_PORT:?}" "${COMPOSE_PROJECT_NAME:?}"
repo=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
# shellcheck source=deploy/install.sh
source "$repo/deploy/install.sh"
WEB_BASE_PATH=$(sed -n 's/^WEB_BASE_PATH=//p' "$APP_DIR/.env")
[[ "$APP_DIR" == /tmp/yc-e2e-*/app ]] || { echo 'FAIL: proxy test requires isolated app'; exit 1; }
tmp=$(mktemp -d /tmp/yc-nginx-XXXXXXXX)
name="${COMPOSE_PROJECT_NAME}-proxy-check"
cleanup() { docker rm -f "$name" >/dev/null 2>&1 || true; rm -rf -- "$tmp"; }
trap cleanup EXIT
fail() { echo "FAIL: $*"; docker logs "$name" 2>&1 || true; exit 1; }
cat > "$tmp/nginx.conf" <<'MAIN'
user root;
error_log /dev/stderr info;
events {}
http {
 log_format route '$status uri=$uri upstream=$upstream_addr';
 access_log /dev/stdout route;
 include /etc/nginx/conf.d/default.conf;
}
MAIN
run_nginx() {
 docker rm -f "$name" >/dev/null 2>&1 || true
 docker run -d --name "$name" --network host \
   -v "$tmp/nginx.conf:/etc/nginx/nginx.conf:ro" \
   -v "$tmp/server.conf:/etc/nginx/conf.d/default.conf:ro" \
   -v "$APP_DIR/frontend/dist:$APP_DIR/frontend/dist:ro" nginx:alpine >/dev/null
 for _ in $(seq 1 40); do
   if curl -sS --max-time 1 -o /dev/null "http://127.0.0.1:$PANEL_PORT/" 2>/dev/null; then return; fi
   sleep 0.25
 done
 fail 'Nginx did not start'
}
assert_health() {
 local route=$1 label=$2 code
 code=$(curl -sS --max-time 5 -o "$tmp/body" -w '%{http_code}' "http://127.0.0.1:$PANEL_PORT$route")
 [[ "$code" == 200 ]] || fail "$label status=$code"
 jq -e '.status == "ok" and (.version | length > 0)' "$tmp/body" >/dev/null || fail "$label returned non-health body"
 echo "NGINX_PASS: $label HTTP=$code $(cat "$tmp/body")"
}
mkdir -p "$APP_DIR/frontend/dist"
printf '<html>isolated fast-mode static fixture</html>' > "$APP_DIR/frontend/dist/index.html"
# Exact routing structure of the previous Docker host template: no root API location.
cat > "$tmp/server.conf" <<OLD
server {
 listen $PANEL_PORT;
 server_name _;
 client_max_body_size 20m;
 location = / { return 302 /$WEB_BASE_PATH/; }
 location /$WEB_BASE_PATH/ {
  proxy_pass http://127.0.0.1:$WEB_PORT/$WEB_BASE_PATH/;
  proxy_http_version 1.1;
  proxy_set_header Host \$host;
  proxy_set_header X-Real-IP \$remote_addr;
  proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
  proxy_set_header X-Forwarded-Proto \$scheme;
 }
}
OLD
run_nginx
old_code=$(curl -sS --max-time 5 -o "$tmp/body" -w '%{http_code}' "http://127.0.0.1:$PANEL_PORT/api/health")
[[ "$old_code" == 404 ]] || fail "old host route reproduction expected 404, got $old_code"
echo 'NGINX_REPRODUCED: old Docker host template /api/health HTTP=404 (default static handler, no upstream)'
docker logs "$name" 2>&1 | tail -n 8
render_host_nginx_docker > "$tmp/server.conf"
run_nginx
assert_health /api/health 'Docker host root API'
assert_health "/$WEB_BASE_PATH/api/health" 'Docker host prefixed API'
[[ $(curl -sS --max-time 5 -o "$tmp/body" -w '%{http_code}' "http://127.0.0.1:$PANEL_PORT/$WEB_BASE_PATH/") == 200 ]] || fail 'Docker frontend page not served'
grep -Fqi '<html' "$tmp/body" || fail 'Docker frontend page not HTML'
echo 'NGINX_PASS: Docker frontend page HTTP=200'
for route in /sub/missing /s/missing/node; do
 code=$(curl -sS --max-time 5 -o "$tmp/body" -w '%{http_code}' "http://127.0.0.1:$PANEL_PORT$route")
 [[ "$code" == 404 ]] || fail "invalid token $route expected API 404, got $code"
 ! grep -Fqi '<html' "$tmp/body" || fail "invalid token $route returned Vue/host fallback"
done
echo 'NGINX_PASS: sub/share invalid tokens reach API (404, no SPA fallback)'
render_host_nginx_fast > "$tmp/server.conf"
run_nginx
assert_health /api/health 'fast template root API'
assert_health "/$WEB_BASE_PATH/api/health" 'fast template prefixed API'
echo 'NGINX_PASS: isolated Docker and fast routing regression checks complete'
