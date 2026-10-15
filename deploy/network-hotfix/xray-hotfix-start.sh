#!/usr/bin/env bash
set -Eeuo pipefail
umask 077
source /etc/yuchen-panel/hotfix/hook.env
[[ $EUID -eq 0 ]] || { echo 'Hotfix requires the audited root Xray service' >&2; exit 1; }
install -d -m 0700 /run/yuchen-xray-hotfix
exec 9>/run/yuchen-xray-hotfix/build.lock
flock -x 9
tmp=$(mktemp /run/yuchen-xray-hotfix/config.XXXXXXXX.json)
trap 'rm -f -- "$tmp"' EXIT
# Read an independent snapshot. The Agent-owned source and its hash stay untouched.
jq -e --argjson tags "$VISION_TAGS" -f /etc/yuchen-panel/hotfix/filter.jq \
  "$SOURCE_CONFIG" > "$tmp"
"$XRAY_BIN" run -test -config "$tmp"
if [[ ${1:-} == --check ]]; then
  echo 'Hotfix transform and Xray validation passed'
  exit 0
fi
mv -f -- "$tmp" /run/yuchen-xray-hotfix/config.json
trap - EXIT
flock -u 9
exec 9>&-
echo 'Starting Xray with validated hotfix runtime config; panel source was preserved'
exec "$XRAY_BIN" run -config /run/yuchen-xray-hotfix/config.json
