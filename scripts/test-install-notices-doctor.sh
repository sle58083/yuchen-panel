#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
# shellcheck source=deploy/install.sh
source "$root/deploy/install.sh"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
export APP_DIR="$tmp/app" CONFIG_DIR="$tmp/config"
mkdir -p "$APP_DIR/data" "$APP_DIR/backups" "$CONFIG_DIR" "$tmp/bin"
record_jwt_rotation
fresh_output=$(print_result)
! grep -Fq '本次安装已轮换 JWT 签名密钥' <<<"$fresh_output" || { echo 'FAIL: fresh install warning'; exit 1; }
printf 'API_PORT=37888\nYC_UPDATE_MANIFEST_URL=https://example.com/version.json\nYC_JWT_SECRET=old-weak-secret\n' > "$APP_DIR/.env"
record_jwt_rotation
notice=$(print_result | sed -n '/⚠️ 安全提示/,+1p')
grep -Fq '本次安装已轮换 JWT 签名密钥' <<<"$notice"
grep -Fq '重新登录' <<<"$notice"
echo 'NOTICE_PASS: fresh install silent; upgrade/reinstall displays:'
echo "$notice"
printf '{}\n' > "$APP_DIR/data/yuchen-panel.json"
touch "$APP_DIR/docker-compose.yml"
printf 'fixture\n' > "$APP_DIR/backups/yuchen-panel-backup-fixture.tar.gz"
cat > "$CONFIG_DIR/panel.info" <<INFO
INSTALL_MODE=docker
VERSION=1.0.2
PORT=39071
WEB_BASE_PATH=fixture
INFO
cat > "$tmp/bin/systemctl" <<'MOCK'
#!/bin/bash
exit 0
MOCK
cat > "$tmp/bin/docker" <<'MOCK'
#!/bin/bash
case "$*" in
 'compose version') echo v2 ;;
 'compose ps -q yuchen-panel-api') echo fixture-container ;;
 'inspect --format {{.State.Status}} fixture-container') if [[ ${FAIL_CONTAINER:-0} == 1 ]]; then echo exited; else echo running; fi ;;
 *) exit 1 ;;
esac
MOCK
cat > "$tmp/bin/curl" <<'MOCK'
#!/bin/bash
if [[ ${FAIL_PROXY:-0} == 1 && "$*" == *':39071/'* ]]; then exit 22; fi
[[ "$*" == *':37888/api/health'* || "$*" == *':39071/api/health'* ]]
MOCK
cat > "$tmp/bin/nginx" <<'MOCK'
#!/bin/bash
exit 0
MOCK
cat > "$tmp/bin/journalctl" <<'MOCK'
#!/bin/bash
exit 0
MOCK
chmod +x "$tmp/bin/"*
PATH="$tmp/bin:$PATH" bash "$root/scripts/yuchen-panel" doctor > "$tmp/doctor.txt"
grep -Fq '[SKIP] yuchen-panel-api (systemd)' "$tmp/doctor.txt"
grep -Fq '[SKIP] yuchen-agent' "$tmp/doctor.txt"
grep -Fq '[OK]   API container' "$tmp/doctor.txt"
grep -Fq 'Doctor result: PASS' "$tmp/doctor.txt"
! grep -Fq '[FAIL]' "$tmp/doctor.txt" || { echo 'FAIL: Docker doctor reported failure'; exit 1; }
cat "$tmp/doctor.txt"
if FAIL_CONTAINER=1 PATH="$tmp/bin:$PATH" bash "$root/scripts/yuchen-panel" doctor > "$tmp/fail.txt"; then echo 'FAIL: stopped container accepted'; exit 1; fi
grep -Fq '[FAIL] API container' "$tmp/fail.txt"
if FAIL_PROXY=1 PATH="$tmp/bin:$PATH" bash "$root/scripts/yuchen-panel" doctor > "$tmp/fail.txt"; then echo 'FAIL: proxy failure accepted'; exit 1; fi
grep -Fq '[FAIL] Nginx API proxy' "$tmp/fail.txt"
sed -i 's/INSTALL_MODE=docker/INSTALL_MODE=fast/' "$CONFIG_DIR/panel.info"
PATH="$tmp/bin:$PATH" bash "$root/scripts/yuchen-panel" doctor > "$tmp/fast.txt"
! grep -Fq '[SKIP]' "$tmp/fast.txt" || { echo 'FAIL: fast service check was skipped'; exit 1; }
grep -Fq '[OK]   yuchen-panel-api' "$tmp/fast.txt"
echo 'DOCTOR_PASS: Docker skip counts, actual container/proxy failures and fast service checks (mock commands)'
