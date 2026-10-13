#!/usr/bin/env bash
set -Eeuo pipefail
umask 077
url=${1:-${SUBSCRIPTION_URL:-}}
out=${2:-./diagnostic-$(date +%Y%m%d-%H%M%S)}
[[ -n $url ]] || { echo 'Usage: diagnose-import.sh <subscription-url> [output-dir]' >&2; exit 2; }
mkdir -p -- "$out"
command -v curl >/dev/null; command -v jq >/dev/null; command -v python3 >/dev/null
curl --fail --silent --show-error --location --proto '=https,http' --proto-redir '=https,http' --get --data-urlencode format=singbox --output "$out/subscription.json" --url "${url%%\?*}"
curl --fail --silent --show-error --location --proto '=https,http' --proto-redir '=https,http' --get --data-urlencode format=v2rayn --output "$out/subscription.b64" --url "${url%%\?*}"
jq '{dns,route,inbounds,outbounds: [.outbounds[] | {type,tag,server,server_port,tls,transport}]}' "$out/subscription.json" > "$out/structure.json"
python3 - "$out" "$url" <<'PY'
import base64, pathlib, sys, urllib.parse, json
d=pathlib.Path(sys.argv[1]); raw=base64.b64decode((d/'subscription.b64').read_bytes(), validate=True).decode()
rows=[]
for line in raw.splitlines():
    p=urllib.parse.urlsplit(line)
    rows.append({'scheme':p.scheme,'host':p.hostname,'port':p.port,'fragment':urllib.parse.unquote(p.fragment),'userinfo_present':bool(p.username),'query_keys':sorted(urllib.parse.parse_qs(p.query))})
(d/'uri-summary.json').write_text(json.dumps(rows,ensure_ascii=False,indent=2))
print('URI_COUNT:',len(rows)); print('ENCODING_CHECK: base64 and URL parsing passed')
host=urllib.parse.urlsplit(sys.argv[2]).hostname
if host:
    import socket
    try: print('DNS_REACHABILITY:', sorted({x[4][0] for x in socket.getaddrinfo(host,443)}))
    except OSError as e: print('DNS_REACHABILITY: failed', type(e).__name__)
PY
jq -r '.outbounds[] | select(.tls.enabled == true and .server) | [.server, (.server_port|tostring), (.tls.server_name // .server)] | @tsv' "$out/subscription.json" | head -n 3 > "$out/tls-targets.tsv"
if command -v openssl >/dev/null; then
  : > "$out/certificate.txt"
  while IFS=$'\t' read -r host port sni; do
    [[ -n $host && -n $port ]] || continue
    if [[ $host == *:* ]]; then endpoint="[$host]:$port"; else endpoint="$host:$port"; fi
    printf 'TARGET: %s SNI: %s\n' "$endpoint" "$sni" >> "$out/certificate.txt"
    timeout 8 openssl s_client -connect "$endpoint" -servername "$sni" -showcerts -brief </dev/null >> "$out/certificate.txt" 2>&1 || true
  done < "$out/tls-targets.tsv"
fi
printf 'DIAGNOSTIC_DIR: %s\n' "$(cd -- "$out" && pwd)"
echo 'Sensitive URLs, UUIDs, passwords and public IPs must be removed before sharing this directory.'
