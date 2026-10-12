#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

subscription_url=${1:-${SUBSCRIPTION_URL:-}}
if [[ -z $subscription_url ]]; then
  printf 'Usage: bash scripts/verify-clients.sh <subscription-url> [new-output-directory]\n' >&2
  exit 2
fi
command -v curl >/dev/null
command -v jq >/dev/null
command -v python3 >/dev/null
base_url=${subscription_url%%\?*}
base_url=${base_url%%\#*}
output_dir=${2:-$(mktemp -d ./client-samples.XXXXXXXX)}
mkdir -p -- "$output_dir"
for file in clash.yaml singbox.json singbox-tun.json v2rayn.txt single-node-links.txt legacy-singbox.json; do
  [[ ! -e $output_dir/$file ]] || { printf 'Refusing to overwrite: %s\n' "$output_dir/$file" >&2; exit 1; }
done
download() {
  local format=$1 file=$2 profile=${3:-mixed}
  curl --fail --silent --show-error --location --retry 2 \
    --proto '=https,http' --proto-redir '=https,http' \
    --get --data-urlencode "format=$format" --data-urlencode "profile=$profile" \
    --output "$output_dir/$file" --url "$base_url"
}
download clash clash.yaml
download singbox singbox.json
download singbox singbox-tun.json tun
download v2rayn v2rayn.txt
jq -e '.dns.servers | all(.type == "udp" and has("server") and (has("address") | not))' "$output_dir/singbox.json" >/dev/null
jq '.dns.servers |= map({tag: .tag, address: .server})' "$output_dir/singbox.json" > "$output_dir/legacy-singbox.json"
python3 - "$output_dir" <<'PY_CLIENTS'
import base64, pathlib, sys
directory = pathlib.Path(sys.argv[1])
raw = base64.b64decode((directory / 'v2rayn.txt').read_bytes(), validate=True).decode('utf-8')
links = [line for line in raw.splitlines() if line]
if any(not line.startswith(('vless://', 'vmess://', 'trojan://')) for line in links):
    raise SystemExit('Unexpected or retired URI protocol in subscription')
(directory / 'single-node-links.txt').write_text('\n'.join(links), encoding='utf-8')
print('Downloaded URI count:', len(links))
print('Clash YAML bytes:', (directory / 'clash.yaml').stat().st_size)
print('Samples:', directory.resolve())
PY_CLIENTS
printf 'Files generated. sing-box check and T1–T6 device tests remain for the operator.\n'
