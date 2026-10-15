#!/usr/bin/env bash
# Debian 12 / Ubuntu 22.04+, host network namespace, root.
# No service restart and no replacement of existing interface qdiscs.
set -Eeuo pipefail
[[ $EUID -eq 0 ]] || { echo 'Run as root' >&2; exit 1; }
for cmd in sysctl systemctl install mktemp awk; do
  command -v "$cmd" >/dev/null || { echo "Missing: $cmd" >&2; exit 1; }
done
[[ -d /run/systemd/system ]] || { echo 'systemd host required' >&2; exit 1; }
backup=$(mktemp -d /var/tmp/yuchen-net-backup.XXXXXXXX)
chmod 700 "$backup"
target=/etc/sysctl.d/99-yuchen-throughput.conf
dropin=/etc/systemd/system/xray.service.d/98-yuchen-throughput.conf
keys=(net.core.default_qdisc net.ipv4.tcp_congestion_control
  net.ipv4.tcp_rmem net.ipv4.tcp_wmem net.core.rmem_max net.core.wmem_max
  net.ipv4.tcp_moderate_rcvbuf net.ipv4.tcp_window_scaling
  net.ipv4.tcp_mtu_probing net.core.somaxconn)
for key in "${keys[@]}"; do
  printf '%s = %s\n' "$key" "$(sysctl -n "$key")" >> "$backup/live-before.conf"
done
[[ ! -e $target ]] || cp -a "$target" "$backup/sysctl-before.conf"
[[ ! -e $dropin ]] || cp -a "$dropin" "$backup/dropin-before.conf"
cat > "$backup/rollback.sh" <<'ROLLBACK'
#!/usr/bin/env bash
set -Eeuo pipefail
[[ $EUID -eq 0 ]]
base=$(cd -- "$(dirname -- "$0")" && pwd)
target=/etc/sysctl.d/99-yuchen-throughput.conf
dropin=/etc/systemd/system/xray.service.d/98-yuchen-throughput.conf
if [[ -f $base/sysctl-before.conf ]]; then cp -a "$base/sysctl-before.conf" "$target"; else rm -f -- "$target"; fi
if [[ -f $base/dropin-before.conf ]]; then cp -a "$base/dropin-before.conf" "$dropin"; else rm -f -- "$dropin"; fi
sysctl -p "$base/live-before.conf"
systemctl daemon-reload
echo 'Rolled back. Restart Xray during a maintenance window to restore process limits.'
ROLLBACK
chmod 700 "$backup/rollback.sh"
trap 'echo "Failed; rollback with: bash $backup/rollback.sh" >&2' ERR
if command -v modprobe >/dev/null; then
  if ! modprobe tcp_bbr; then echo 'tcp_bbr module not loaded; checking built-in support' >&2; fi
fi
if ! sysctl -n net.ipv4.tcp_available_congestion_control | tr ' ' '\n' | grep -qx bbr; then
  echo 'BBR unavailable; no persistent tuning applied. Use a supported distro kernel.' >&2
  exit 1
fi
# 64 MiB maximum per socket; allocated on demand, not preallocated.
# Review memory capacity before deploying to high-concurrency small VPS hosts.
cat > "$backup/new.conf" <<'SYSCTL'
net.core.default_qdisc = fq
net.ipv4.tcp_congestion_control = bbr
net.ipv4.tcp_rmem = 4096 131072 67108864
net.ipv4.tcp_wmem = 4096 131072 67108864
net.core.rmem_max = 67108864
net.core.wmem_max = 67108864
net.ipv4.tcp_moderate_rcvbuf = 1
net.ipv4.tcp_window_scaling = 1
net.ipv4.tcp_mtu_probing = 1
net.core.somaxconn = 4096
SYSCTL
install -m 0644 "$backup/new.conf" "$target"
sysctl -p "$target"
install -d -m 0755 /etc/systemd/system/xray.service.d
printf '[Service]\nLimitNOFILE=65536\n' > "$dropin"
systemctl daemon-reload
echo "Backup and rollback: $backup/rollback.sh"
echo 'New TCP sockets use BBR unless overridden by the application.'
echo 'Existing NIC qdiscs were preserved; inspect with: tc -s qdisc show'
echo 'For Xray process limits, restart during a maintenance window: systemctl restart xray'
echo 'Verify live sockets during a download: ss -tinp'
