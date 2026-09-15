#!/usr/bin/env bash
set -euo pipefail

PANEL_BASE="${PANEL_BASE:-http://127.0.0.1:8088}"
SERVER_ID="${SERVER_ID:-}"
AGENT_TOKEN="${AGENT_TOKEN:-}"
APPLY_CONFIG="${APPLY_CONFIG:-true}"
XRAY_CONFIG="${XRAY_CONFIG:-/etc/yuchen-panel/xray/config.json}"
XRAY_TEST_CMD="${XRAY_TEST_CMD:-xray run -test -config {config}}"
XRAY_RELOAD_CMD="${XRAY_RELOAD_CMD:-systemctl restart xray}"
INSTALL_XRAY="${INSTALL_XRAY:-true}"
SETUP_XRAY_SERVICE="${SETUP_XRAY_SERVICE:-true}"
YC_FORCE_INSTALL_XRAY="${YC_FORCE_INSTALL_XRAY:-0}"
YC_SKIP_XRAY_INSTALL="${YC_SKIP_XRAY_INSTALL:-0}"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DST="/usr/local/bin/yuchen-agent"

if [[ -z "$SERVER_ID" || -z "$AGENT_TOKEN" ]]; then
  echo "ERROR: SERVER_ID and AGENT_TOKEN are required."
  echo "Example: PANEL_BASE='http://1.2.3.4:8088' SERVER_ID='srv_xxx' AGENT_TOKEN='token' ./deploy/agent-install.sh"
  exit 1
fi

install_local_xray_if_available() {
  local xray_src=""
  if [[ -x "$ROOT_DIR/bin/xray-linux-amd64" ]]; then
    xray_src="$ROOT_DIR/bin/xray-linux-amd64"
  elif [[ -x "$ROOT_DIR/bin/xray" ]]; then
    xray_src="$ROOT_DIR/bin/xray"
  fi

  if [[ -z "$xray_src" ]]; then
    return 1
  fi

  echo "Installing bundled Xray-core: $xray_src"
  install -m 0755 "$xray_src" /usr/local/bin/xray
  mkdir -p /usr/local/share/xray /usr/local/etc/xray /var/log/xray /etc/yuchen-panel/xray

  if [[ -f "$ROOT_DIR/bin/geoip.dat" ]]; then
    install -m 0644 "$ROOT_DIR/bin/geoip.dat" /usr/local/share/xray/geoip.dat
  fi
  if [[ -f "$ROOT_DIR/bin/geosite.dat" ]]; then
    install -m 0644 "$ROOT_DIR/bin/geosite.dat" /usr/local/share/xray/geosite.dat
  fi
  if [[ ! -f /usr/local/etc/xray/config.json ]]; then
    cat > /usr/local/etc/xray/config.json <<'JSON'
{
  "log": { "loglevel": "warning" },
  "inbounds": [],
  "outbounds": [ { "protocol": "freedom", "tag": "direct" } ]
}
JSON
  fi
  return 0
}

if [[ "${YC_SKIP_XRAY_INSTALL}" == "1" || "${YC_SKIP_XRAY_INSTALL}" == "true" ]]; then
  echo "YC_SKIP_XRAY_INSTALL is enabled, skip Xray-core installation."
elif [[ "${INSTALL_XRAY}" == "true" ]]; then
  if [[ "${YC_FORCE_INSTALL_XRAY}" != "1" && "${YC_FORCE_INSTALL_XRAY}" != "true" ]] && command -v xray >/dev/null 2>&1; then
    echo "Xray-core already installed, skip download: $(xray version | head -1)"
  elif install_local_xray_if_available; then
    echo "Bundled Xray-core installed: $(xray version | head -1)"
  else
    if [[ "${YC_FORCE_INSTALL_XRAY}" == "1" || "${YC_FORCE_INSTALL_XRAY}" == "true" ]]; then
      echo "YC_FORCE_INSTALL_XRAY is enabled, reinstalling Xray-core using official community install script..."
    else
      echo "Xray-core not found, installing using official community install script..."
    fi
    bash -c "$(curl -L https://github.com/XTLS/Xray-install/raw/main/install-release.sh)" @ install
  fi
else
  echo "INSTALL_XRAY is false, skip Xray-core installation."
fi

BIN_SRC=""
if [[ -x "$ROOT_DIR/bin/yuchen-agent-linux-amd64" ]]; then
  BIN_SRC="$ROOT_DIR/bin/yuchen-agent-linux-amd64"
elif [[ -x "$ROOT_DIR/bin/yuchen-agent" ]]; then
  BIN_SRC="$ROOT_DIR/bin/yuchen-agent"
fi

if [[ -z "$BIN_SRC" ]]; then
  echo "Agent binary not found, trying to build from source..."
  if ! command -v go >/dev/null 2>&1; then
    echo "ERROR: Go is not installed and prebuilt binary is missing."
    exit 1
  fi
  BIN_SRC="$ROOT_DIR/bin/yuchen-agent-linux-amd64"
  mkdir -p "$ROOT_DIR/bin"
  (cd "$ROOT_DIR/agent" && go build -o "$BIN_SRC" ./cmd/agent)
fi

install -m 0755 "$BIN_SRC" "$BIN_DST"
mkdir -p /etc/yuchen-panel/xray
cat > /etc/yuchen-panel/agent.env <<ENV
YC_PANEL_BASE=${PANEL_BASE}
YC_SERVER_ID=${SERVER_ID}
YC_AGENT_TOKEN=${AGENT_TOKEN}
YC_XRAY_CONFIG=${XRAY_CONFIG}
YC_XRAY_TEST_CMD=${XRAY_TEST_CMD}
YC_XRAY_RELOAD_CMD=${XRAY_RELOAD_CMD}
YC_APPLY_CONFIG=${APPLY_CONFIG}
YC_AGENT_INTERVAL_SECONDS=30
ENV
chmod 0600 /etc/yuchen-panel/agent.env

if [[ "${SETUP_XRAY_SERVICE}" == "true" ]]; then
  XRAY_BIN="$(command -v xray || true)"
  if [[ -n "$XRAY_BIN" ]]; then
    mkdir -p /etc/systemd/system/xray.service.d
    cat > /etc/systemd/system/xray.service.d/99-yuchen-panel.conf <<EOF_SERVICE
[Service]
User=root
Group=root
ExecStart=
ExecStart=$XRAY_BIN run -config ${XRAY_CONFIG}
EOF_SERVICE
    systemctl daemon-reload
  fi
fi

cat > /etc/systemd/system/yuchen-agent.service <<'SERVICE'
[Unit]
Description=Yuchen Panel Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
EnvironmentFile=/etc/yuchen-panel/agent.env
ExecStart=/usr/local/bin/yuchen-agent
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
SERVICE

systemctl daemon-reload
systemctl enable yuchen-agent
systemctl restart yuchen-agent

echo "Yuchen Agent installed."
echo "Check status: systemctl status yuchen-agent --no-pager"
echo "Check logs:   journalctl -u yuchen-agent -n 80 --no-pager"
