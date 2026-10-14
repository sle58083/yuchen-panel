#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
# Loaded dynamically from the repository root.
# shellcheck disable=SC1091
source "$root/deploy/install.sh"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
SRC_DIR="$tmp/source"; APP_DIR="$tmp/runtime"
mkdir -p "$SRC_DIR" "$APP_DIR"
# Actual Compose interpolation test is read-only; no images/containers/services are started.
if command docker compose version >/dev/null 2>&1; then
  SRC_DIR="$root"; COMPOSE="docker compose"
  unset YC_JWT_SECRET YC_ADMIN_USERNAME YC_ADMIN_PASSWORD
  validate_compose_source
  [[ ! ${YC_JWT_SECRET+x} && ! ${YC_ADMIN_USERNAME+x} && ! ${YC_ADMIN_PASSWORD+x} ]]
  echo 'PASS: actual Compose v2 source validation without runtime .env (read-only)'
fi
SRC_DIR="$tmp/source"
export COMPOSE=compose_stub
compose_stub() {
  case "$*" in
    *config*)
      [[ -n ${YC_JWT_SECRET:-} || "$*" == *'--env-file'* ]]
      ;;
    *'up -d'*)
      [[ ! ${YC_JWT_SECRET+x} && ! ${YC_ADMIN_USERNAME+x} && ! ${YC_ADMIN_PASSWORD+x} ]]
      [[ "$*" == *"--env-file $APP_DIR/.env"* ]]
      ;;
    *) return 1 ;;
  esac
}
unset YC_JWT_SECRET YC_ADMIN_USERNAME YC_ADMIN_PASSWORD
validate_compose_source
[[ ! ${YC_JWT_SECRET+x} && ! ${YC_ADMIN_USERNAME+x} && ! ${YC_ADMIN_PASSWORD+x} ]]
echo 'PASS: fresh source without .env validated; placeholders confined to child shell'
printf 'YC_JWT_SECRET=real-generated-secret-with-at-least-32-characters\n' > "$APP_DIR/.env"
validate_compose_source
export YC_JWT_SECRET=external-placeholder YC_ADMIN_USERNAME=external-placeholder YC_ADMIN_PASSWORD=external-placeholder
start_docker_runtime
echo 'PASS: real env-file explicitly selected at validation/start; exported values excluded from startup'
# Simulate apt distributions and verify success requires the v2 command, not package exit 0.
apt_update_once() { :; }
systemctl() { :; }
apt-cache() { [[ $1 == show ]]; }
installed=''
# Mock only installer preflight calls; cleanup/xargs is not invoked.
# shellcheck disable=SC2032
docker() {
  if [[ $1 == --version ]]; then echo 'mock Docker'; return 0; fi
  [[ "$*" == 'compose version' && "$installed" == "$accepted" ]]
}
apt-get() { installed=${!#}; echo "ATTEMPT:$installed"; return 0; }
accepted=docker-compose-v2
install_docker_if_missing > "$tmp/apt.log" 2>&1
grep -Fq 'ATTEMPT:docker-compose-v2' "$tmp/apt.log"
accepted=docker-compose-plugin; installed=''
install_docker_if_missing > "$tmp/apt.log" 2>&1
grep -Fq 'ATTEMPT:docker-compose-plugin' "$tmp/apt.log"
accepted=never; installed=''
if install_docker_if_missing > "$tmp/apt.log" 2>&1; then echo 'FAIL: package v1 was accepted'; exit 1; fi
grep -Fq 'https://docs.docker.com/engine/install/ubuntu/' "$tmp/apt.log"
grep -Fq 'sudo apt-get install -y docker-compose-plugin' "$tmp/apt.log"
grep -Fq 'existing runtime was left untouched' "$tmp/apt.log"
echo 'PASS: Ubuntu package, official plugin fallback, v1 rejection and manual recovery guidance (mock commands)'
