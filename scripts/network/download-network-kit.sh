#!/usr/bin/env bash
# Downloads a reviewed immutable commit; does not install or restart services.
set -Eeuo pipefail
revision=bff728061acca57666e5f77cf5e9676b66585602
url="https://raw.githubusercontent.com/sle58083/yuchen-panel/$revision"
destination=${1:-./yuchen-network-kit}
for cmd in curl sha256sum mktemp install; do
  command -v "$cmd" >/dev/null || { echo "Missing dependency: $cmd" >&2; exit 1; }
done
mkdir -p -- "$destination"
destination=$(cd -- "$destination" && pwd)
stage=$(mktemp -d "$destination/.download.XXXXXXXX")
trap 'rm -rf -- "$stage"' EXIT
while read -r digest path; do
  name=${path##*/}
  echo "Downloading $name"
  curl --proto '=https' --proto-redir '=https' -fLSs \
    --connect-timeout 15 --max-time 120 --retry 3 \
    "$url/$path" -o "$stage/$name"
  printf '%s  %s\n' "$digest" "$name" >> "$stage/SHA256SUMS"
done <<'MANIFEST'
e5b11b1b378393cdc40439ca7cf5ab1576cf94157848f83510b4704844d257bf deploy/network-hotfix/install-xray-hotfix.sh
5bdff27dfa1c0f6bd74a0319bd24c2d05fd4d548dbe264f76fe53619861856b8 deploy/network-hotfix/xray-hotfix-start.sh
3558618121ac31e383aea6f1f818a6d1d4eb8e37365a8e7321bf809bd4a24918 deploy/network-hotfix/xray-hotfix.jq
2ae84f80cc37f41e1eb4f04061271448e847df02637f105f13609bfc110850a5 scripts/network/tune-network.sh
756d4711d4880aaee71ebc42a7667e003a122878fbb8d9c106f32d4a59e3ca68 scripts/network/collect-network-diagnostics.sh
955a46efb388119549e283638a15fad539855c27752be2dcdeb1e63e62c884fe scripts/network/iperf-summary.jq
1495b076bd63d0377976fe99267c4f68ec132cf5825d32d90b2b52bc20cbd22f scripts/network/diagnostic-submission-template.md
MANIFEST
(cd -- "$stage" && sha256sum -c SHA256SUMS)
for file in "$stage"/*; do
  install -m 0644 -- "$file" "$destination/${file##*/}"
done
echo "Verified kit downloaded to: $destination"
echo 'Nothing was installed. Next: sudo bash install-xray-hotfix.sh --vision-tags '\''[]'\'' --restart'

