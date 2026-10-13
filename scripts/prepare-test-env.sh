#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

usage() {
  cat <<'USAGE'
Usage: YC_ENV=test bash scripts/prepare-test-env.sh [--dry-run|--apply] [--allow-existing]
       [--panel-url=URL] [--output=DIR]

Dry-run is the default. --apply creates or reuses one test client and six test
nodes on the selected disposable test panel. It requires YC_TOKEN (a login JWT).
Set TEST_NODE_HOST to override the server's advertised IPv4 address.
Set TEST_SERVER_ID when the panel has more than one server.
USAGE
}

mode=dry-run
allow_existing=0
panel_url=${PANEL_URL:-http://127.0.0.1:8088}
output_dir=${TEST_ENV_DIR:-./client-test-env}
for arg in "$@"; do
  case "$arg" in
    --dry-run) mode=dry-run ;;
    --apply) mode=apply ;;
    --allow-existing) allow_existing=1 ;;
    --panel-url=*) panel_url=${arg#*=} ;;
    --output=*) output_dir=${arg#*=} ;;
    -h|--help) usage; exit 0 ;;
    *) printf 'Unknown argument: %s\n' "$arg" >&2; usage >&2; exit 2 ;;
  esac
done

[[ ${YC_ENV:-} == test ]] || { echo 'YC_ENV=test is required; refusing to operate on production or an unspecified environment.' >&2; exit 1; }
data_path=/opt/yuchen-panel/data/yuchen-panel.json
for guarded_path in "$data_path" "${YC_DATA_PATH:-$data_path}"; do
  if [[ -s $guarded_path && $allow_existing -ne 1 ]]; then
    printf 'Non-empty panel data at %s; use --allow-existing only on a disposable test panel.\n' "$guarded_path" >&2
    exit 1
  fi
done
[[ $panel_url == http://* || $panel_url == https://* ]] || { echo 'Panel URL must begin with http:// or https://.' >&2; exit 2; }
panel_url=${panel_url%/}
command -v python3 >/dev/null || { echo 'python3 is required.' >&2; exit 2; }

printf 'Environment: YC_ENV=%s\nPanel URL: %s\nGuarded data path: %s\nOutput path: %s\nMode: %s\n' \
  "$YC_ENV" "$panel_url" "$data_path" "$output_dir" "$mode"
printf 'Plan: inspect the test server; create/reuse uri-edge-case and six named nodes (VLESS TCP/Reality/gRPC/WS, VMess TCP, Trojan TCP).\n'
printf 'Node ports: 24401-24406. Verify that these ports are free and reachable before real connection tests.\n'

ipv6_addr=''
ipv6_reason=''
if ! command -v ip >/dev/null; then
  ipv6_reason='ip command unavailable'
elif ! ip -6 route show default | grep -q '^default'; then
  ipv6_reason='no IPv6 default route (::/0)'
else
  ipv6_addr=$(python3 - <<'PY'
import ipaddress
import socket
import subprocess

try:
    rows = subprocess.check_output(
        ['ip', '-6', '-o', 'addr', 'show', 'scope', 'global'],
        text=True, stderr=subprocess.DEVNULL,
    ).splitlines()
except (OSError, subprocess.CalledProcessError):
    rows = []
for row in rows:
    fields = row.split()
    if len(fields) < 4:
        continue
    address = fields[3].split('/')[0]
    try:
        ip = ipaddress.IPv6Address(address)
        if not ip.is_global or ip not in ipaddress.IPv6Network('2000::/3'):
            continue
        with socket.socket(socket.AF_INET6, socket.SOCK_STREAM) as probe:
            probe.bind((address, 0))
        print(address)
        break
    except (OSError, ValueError):
        pass
PY
)
  [[ -n $ipv6_addr ]] || ipv6_reason='no global 2000::/3 address that passes bind/socket probe'
fi
if [[ -n $ipv6_addr ]]; then
  ipv6_status=available
  printf 'IPv6: available, %s (external reachability still needs a device test)\n' "$ipv6_addr"
else
  ipv6_status=unavailable
  printf 'IPv6: unavailable; T2 IPv6 connection is SKIPPED: %s\n' "$ipv6_reason"
fi

if [[ $mode == dry-run ]]; then
  echo 'DRY_RUN: no API calls, no output files, no clients or nodes created.'
  exit 0
fi

[[ -n ${YC_TOKEN:-} ]] || { echo 'Set YC_TOKEN to the JWT returned by /api/auth/login before --apply.' >&2; exit 2; }
for tool in curl jq sha256sum; do command -v "$tool" >/dev/null || { printf '%s is required.\n' "$tool" >&2; exit 2; }; done
script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
mkdir -p -- "$output_dir"
output_dir=$(cd -- "$output_dir" && pwd)
tmp=$(mktemp -d "$output_dir/.prepare.XXXXXXXX")
trap 'rm -rf -- "$tmp"' EXIT
api=$panel_url/api
auth=(-H "Authorization: Bearer $YC_TOKEN")
api_get() { curl --fail --silent --show-error --proto '=http,https' "${auth[@]}" "$1"; }
api_post() { curl --fail --silent --show-error --proto '=http,https' "${auth[@]}" -H 'Content-Type: application/json' -X POST --data-binary @"$2" "$1"; }

servers=$(api_get "$api/servers")
server_count=$(jq 'length' <<<"$servers")
if [[ -n ${TEST_SERVER_ID:-} ]]; then
  server=$(jq -ce --arg id "$TEST_SERVER_ID" '[.[] | select(.id == $id)] | if length == 1 then .[0] else error("TEST_SERVER_ID not found or duplicated") end' <<<"$servers")
elif [[ $server_count -eq 1 ]]; then
  server=$(jq -ce '.[0]' <<<"$servers")
else
  printf 'Expected exactly one test server, got %s; set TEST_SERVER_ID explicitly.\n' "$server_count" >&2
  exit 1
fi
server_id=$(jq -r '.id' <<<"$server")
node_host=${TEST_NODE_HOST:-$(jq -r '.host // .ip // empty' <<<"$server")}
if ! python3 - "$node_host" <<'PY'
import ipaddress
import sys

try:
    ip = ipaddress.IPv4Address(sys.argv[1])
except ValueError:
    raise SystemExit(1)
blocked = [ipaddress.IPv4Network(net) for net in (
    '0.0.0.0/8', '127.0.0.0/8', '169.254.0.0/16',
    '192.0.2.0/24', '198.51.100.0/24', '203.0.113.0/24',
    '224.0.0.0/4', '240.0.0.0/4',
)]
if any(ip in net for net in blocked):
    raise SystemExit(1)
PY
then
  printf 'Server host %s is not a usable test-machine IPv4 address. Set TEST_NODE_HOST to the machine\047s real private/public IPv4.\n' "$node_host" >&2
  exit 1
fi
printf 'Selected server: id=%s, test-machine IPv4=%s. Use TEST_NODE_HOST to override before --apply if this is wrong.\n' "$server_id" "$node_host"
echo 'Warning: this branch does not configure TLS certificates for TLS inbounds; Trojan/WS/gRPC connection checks may fail. Use only an isolated test machine.'

clients=$(api_get "$api/clients")
nodes=$(api_get "$api/nodes")
[[ $(jq '[.[] | select(.username == "uri-edge-case")] | length' <<<"$clients") -le 1 ]] || { echo 'Duplicate uri-edge-case clients found; resolve manually.' >&2; exit 1; }
client_id=$(jq -r '[.[] | select(.username == "uri-edge-case")][0].id // empty' <<<"$clients")
client_token=$(jq -r '[.[] | select(.username == "uri-edge-case")][0].subscribe_token // empty' <<<"$clients")
edge_password="&=+,'\" /中文😀:@"
node_ids=()
resources="$tmp/resources.jsonl"
: > "$resources"

create_node() {
  local name=$1 protocol=$2 transport=$3 security=$4 host=$5 port=$6
  local existing id payload response keys status
  existing=$(jq -c --arg n "$name" '[.[] | select(.name == $n)] | if length <= 1 then .[0] // empty else error("duplicate test node names") end' <<<"$nodes")
  if [[ -n $existing ]]; then
    if ! jq -e --arg sid "$server_id" --arg p "$protocol" --arg t "$transport" --arg s "$security" --arg h "$host" --argjson port "$port" \
      '.server_id == $sid and .protocol == $p and .transport == $t and .security == $s and .host == $h and .port == $port' <<<"$existing" >/dev/null; then
      printf 'Existing node %s has different settings; review it manually before retrying.\n' "$name" >&2
      exit 1
    fi
    id=$(jq -r '.id' <<<"$existing")
    status=reused
  else
    payload="$tmp/node-$port.json"
    if [[ $security == reality ]]; then
      keys=$(curl --fail --silent --show-error --proto '=http,https' "${auth[@]}" -X POST "$api/nodes/reality-keys")
      jq -n --arg sid "$server_id" --arg n "$name" --arg h "$host" --argjson port "$port" --argjson k "$keys" \
        '{server_id:$sid,name:$n,host:$h,port:$port,protocol:"vless",transport:"tcp",security:"reality",reality_private_key:$k.private_key,reality_public_key:$k.public_key,reality_short_id:$k.short_id,reality_dest:$k.dest,sni:$k.sni,reality_spider_x:"/",fingerprint:"chrome"}' > "$payload"
    else
      jq -n --arg sid "$server_id" --arg n "$name" --arg h "$host" --arg p "$protocol" --arg t "$transport" --arg s "$security" --argjson port "$port" \
        '{server_id:$sid,name:$n,host:$h,port:$port,protocol:$p,transport:$t,security:$s}' > "$payload"
    fi
    response=$(api_post "$api/nodes" "$payload")
    id=$(jq -r '.id // empty' <<<"$response")
    [[ -n $id ]] || { printf 'Node creation returned no id: %s\n' "$name" >&2; exit 1; }
    nodes=$(jq -n --argjson old "$nodes" --argjson item "$response" '$old + [$item]')
    status=created
  fi
  node_ids+=("$id")
  jq -nc --arg id "$id" --arg name "$name" --arg state "$status" '{id:$id,name:$name,state:$state}' >> "$resources"
  printf 'NODE_%s: %s (%s)\n' "${status^^}" "$name" "$id"
}

create_node uri-edge-vless-tcp vless tcp none "$node_host" 24401
create_node uri-edge-vless-reality vless tcp reality "$node_host" 24402
create_node uri-edge-vless-grpc vless grpc tls "$node_host" 24403
create_node uri-edge-vless-ws vless ws tls "$node_host" 24404
create_node uri-edge-vmess vmess tcp none "$node_host" 24405
if [[ $ipv6_status == available ]]; then trojan_host=$ipv6_addr; else trojan_host=$node_host; fi
create_node uri-edge-trojan trojan tcp tls "$trojan_host" 24406

node_json=$(printf '%s\n' "${node_ids[@]}" | jq -R . | jq -s .)
client_payload="$tmp/client.json"
jq -n --arg p "$edge_password" --argjson ids "$node_json" \
  '{username:"uri-edge-case",password:$p,node_ids:$ids,traffic_limit_gb:100,enabled:true}' > "$client_payload"
if [[ -z $client_id ]]; then
  response=$(api_post "$api/clients" "$client_payload")
  client_id=$(jq -r '.id // empty' <<<"$response")
  client_token=$(jq -r '.subscribe_token // empty' <<<"$response")
  client_state=created
else
  jq --arg p "$edge_password" --argjson ids "$node_json" '.password=$p | .node_ids=$ids | .enabled=true' \
    <<<"$(jq -c --arg id "$client_id" '.[] | select(.id == $id)' <<<"$clients")" > "$client_payload"
  response=$(curl --fail --silent --show-error --proto '=http,https' "${auth[@]}" -H 'Content-Type: application/json' -X PUT --data-binary @"$client_payload" "$api/clients/$client_id")
  client_token=$(jq -r '.subscribe_token // empty' <<<"$response")
  client_state=reused
fi
[[ -n $client_id && -n $client_token ]] || { echo 'Client creation/update returned no id or subscription token.' >&2; exit 1; }
printf 'CLIENT_%s: uri-edge-case (%s)\n' "${client_state^^}" "$client_id"

sample_dir=$(mktemp -d "$output_dir/samples.XXXXXXXX")
bash "$script_dir/verify-clients.sh" "$panel_url/sub/$client_token" "$sample_dir"

jq -e '.dns.servers | all(.type == "udp" and has("server") and (has("address") | not))' "$sample_dir/singbox.json" >/dev/null
jq -e '[.outbounds[].tag] as $tags | (($tags | length) == ($tags | unique | length)) and ($tags | all(length > 0))' "$sample_dir/singbox.json" >/dev/null
clash_names="$tmp/clash-names"
awk '/^  - name: / {sub(/^  - name: /, ""); print}' "$sample_dir/clash.yaml" > "$clash_names"
[[ -s $clash_names && $(wc -l < "$clash_names") -eq $(sort -u "$clash_names" | wc -l) ]] || { echo 'Clash names are empty or duplicate.' >&2; exit 1; }
export EXPECTED_PASSWORD=$edge_password
python3 - "$sample_dir/single-node-links.txt" <<'PY'
import os
import pathlib
import sys
from urllib.parse import unquote, urlsplit

lines = pathlib.Path(sys.argv[1]).read_text(encoding='utf-8').splitlines()
trojan = [line for line in lines if line.startswith('trojan://')]
if len(trojan) != 1:
    raise SystemExit(f'expected one Trojan URI, found {len(trojan)}')
if unquote(urlsplit(trojan[0]).username or '') != os.environ['EXPECTED_PASSWORD']:
    raise SystemExit('Trojan password round trip failed')
if any(line.startswith('ss://') for line in lines):
    raise SystemExit('SS URI found')
PY
echo 'SELF_CHECK_PASS: modern DNS, unique sing-box tags and Clash names, Trojan password, no SS URI.'

python3 - "$output_dir" "$sample_dir" "$panel_url" "$node_host" "$ipv6_status" "$ipv6_addr" "$ipv6_reason" "$client_id" "$client_state" "$resources" <<'PY'
import hashlib
import json
import pathlib
import sys

out, samples = (pathlib.Path(v) for v in sys.argv[1:3])
files = {}
for path in sorted(samples.iterdir()):
    if path.is_file():
        files[path.name] = {
            'path': str(path.resolve()),
            'sha256': hashlib.sha256(path.read_bytes()).hexdigest(),
            'bytes': path.stat().st_size,
        }
resources = [json.loads(row) for row in pathlib.Path(sys.argv[10]).read_text().splitlines()]
record = {
    'panel_url': sys.argv[3],
    'test_machine_ip': sys.argv[4],
    'ipv6': {'status': sys.argv[5], 'address': sys.argv[6] or None, 'reason': sys.argv[7]},
    'client': {'id': sys.argv[8], 'username': 'uri-edge-case', 'state': sys.argv[9]},
    'nodes': resources,
    'samples': files,
}
with (out / 'TEST-ENV.json').open('w', encoding='utf-8') as stream:
    json.dump(record, stream, ensure_ascii=False, indent=2)
PY
printf 'APPLY_COMPLETE: resources=%s\nTEST_ENV: %s/TEST-ENV.json\n' "${#node_ids[@]} nodes + 1 client" "$output_dir"
