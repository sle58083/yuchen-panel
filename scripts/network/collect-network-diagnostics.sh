#!/usr/bin/env bash
# Usage: sudo bash collect-network-diagnostics.sh NODE_PORT [SOURCE_CONFIG]
# Run while a single sustained client download is active.
# No credentials, raw JSON, raw firewall rules or raw journal messages are printed.
set -Eeuo pipefail
port=${1:?Usage: collect-network-diagnostics.sh NODE_PORT [SOURCE_CONFIG]}
config=${2:-/etc/yuchen-panel/xray/config.json}
[[ $port =~ ^[0-9]+$ ]] && ((port >= 1 && port <= 65535))
for cmd in ss sysctl jq python3; do command -v "$cmd" >/dev/null; done
export LC_ALL=C
echo '[kernel]'
uname -r
sysctl net.ipv4.tcp_congestion_control net.ipv4.tcp_available_congestion_control \
  net.ipv4.tcp_rmem net.ipv4.tcp_wmem net.core.rmem_max net.core.wmem_max \
  net.ipv4.tcp_moderate_rcvbuf net.ipv4.tcp_window_scaling \
  net.ipv4.tcp_mtu_probing net.core.default_qdisc net.core.somaxconn
echo '[socket samples: only node-port sockets; connection numbers reset each sample]'
for sample in 1 2 3; do
  echo "sample=$sample"
  ss -tminH state established "( sport = :$port or dport = :$port )" | awk '
  /^[^[:space:]]/ { n++; if(n<=10) print "conn="n,"recv_q=" $1,"send_q=" $2; next }
  n<=10 {
    out="";
    for(i=1;i<=NF;i++) {
      if($i ~ /^(bbr|bbr2|cubic|reno)$/ ||
         $i ~ /^(rtt|rto|mss|pmtu|advmss|cwnd|ssthresh|wscale|retrans|bytes_retrans|bytes_acked|bytes_received|segs_out|segs_in|data_segs_out|data_segs_in|lost|rcv_space|rcv_ssthresh|snd_wnd|rwnd_limited|sndbuf_limited|busy|rcv_rtt):/)
         out=out " " $i;
      if($i ~ /^(send|pacing_rate|delivery_rate)$/ && i<NF) out=out " " $i "=" $(i+1);
      if($i ~ /^skmem:/) out=out " " $i;
    }
    print out;
  }
  END {print "matched_connections="n+0,"displayed_max=10"}'
  if ((sample < 3)); then sleep 5; fi
done
echo '[interface counters]'
for path in /sys/class/net/*; do
  dev=${path##*/}
  [[ $dev != lo ]] || continue
  printf 'interface=%s mtu=%s' "$dev" "$(cat "$path/mtu")"
  for metric in rx_bytes tx_bytes rx_dropped tx_dropped rx_errors tx_errors; do
    printf ' %s=%s' "$metric" "$(cat "$path/statistics/$metric")"
  done
  printf '\n'
done
echo '[traffic-control numeric summaries]'
if command -v tc >/dev/null; then
  for path in /sys/class/net/*; do
    dev=${path##*/}
    echo "interface=$dev"
    { tc -s qdisc show dev "$dev"; tc -s class show dev "$dev";
      tc -s filter show dev "$dev"; tc -s filter show dev "$dev" ingress; } |
      python3 -c 'import sys,re
lines=sys.stdin.read().splitlines()
for line in lines[:400]:
 kinds=re.findall(r"\b(?:qdisc|class)\s+(\w+)",line)
 metrics=re.findall(r"\b(?:rate|ceil|burst|limit|backlog|dropped|overlimits|requeues|Sent|police|conform|exceed)\s+[0-9.]+(?:[A-Za-z/]+)?",line)
 if kinds or metrics: print(" ".join(kinds+metrics))
if len(lines)>400: print("truncated: >400 tc lines")'
  done
fi
echo '[firewall feature counts; rules withheld]'
for tool in nft iptables-save; do
  command -v "$tool" >/dev/null || continue
  if [[ $tool == nft ]]; then args=(list ruleset); else args=(-c); fi
  "$tool" "${args[@]}" | python3 -c 'import sys,re
s=sys.stdin.read()
for feature in ["hashlimit","limit","meter","quota","drop","reject"]:
 print(feature+"_occurrences="+str(len(re.findall(r"\b"+feature+r"\b",s,re.I))))' || echo "$tool read_failed"
done
echo '[Xray version and service limits]'
xray version | awk 'NR==1 {print $1,$2}'
systemctl show xray -p LimitNOFILE -p MainPID -p ActiveState
pid=$(systemctl show xray -p MainPID --value)
if [[ $pid =~ ^[1-9][0-9]*$ && -r /proc/$pid/limits ]]; then
  awk '/Max open files/' "/proc/$pid/limits"
  ps -p "$pid" -o %cpu=,rss=,nlwp= | awk '{print "cpu_pct=" $1,"rss_kib=" $2,"threads=" $3}'
fi
echo '[safe config projection: source and optional hotfix runtime]'
for file in "$config" /run/yuchen-xray-hotfix/config.json; do
  [[ -r $file ]] || continue
  [[ $file == "$config" ]] && echo 'config=panel_source' || echo 'config=hook_runtime'
  jq '{inbounds:[(.inbounds // []) | to_entries[] |
    {index:.key,protocol:.value.protocol,port:.value.port,
     transport:.value.streamSettings.network,security:.value.streamSettings.security,
     sniffing_enabled:(.value.sniffing.enabled // false),
     routeOnly:.value.sniffing.routeOnly,
     flows:([.value.settings.clients[]?.flow // ""]|unique)}],
    outbounds:[(.outbounds // [])[]|{protocol,mux_enabled:(.mux.enabled // false)}],
    routing_strategy:.routing.domainStrategy,
    rule_count:((.routing.rules // [])|length)}' "$file"
done
echo '[journal event counts: current boot, last 10 min, max 200 entries, no raw message text]'
journalctl -b --since '-10 min' -u xray -u yuchen-agent -n 200 --no-pager -o cat |
python3 -c 'import sys,re
s=sys.stdin.read().lower()
for event in ["config applied","config already up to date","apply failed","sync failed","xray test failed","xray reload failed","timeout","failed","error","too many open files","connection reset","permission denied"]:
 print(event.replace(" ","_")+"="+str(s.count(event)))'
