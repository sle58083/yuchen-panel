def scaled($n; $d): if $n == null then null else $n / $d end;
{sender_mbps:scaled(.end.sum_sent.bits_per_second;1000000),
 receiver_mbps:scaled(.end.sum_received.bits_per_second;1000000),
 retransmits:.end.sum_sent.retransmits,
 sender_cpu_pct:.end.cpu_utilization_percent.host_total,
 receiver_cpu_pct:.end.cpu_utilization_percent.remote_total,
 failed:has("error"),
 streams:[.end.streams[]? | .sender |
   {mean_rtt_ms:scaled(.mean_rtt;1000),
    max_snd_cwnd_bytes:.max_snd_cwnd,
    retransmits,
    sender_mbps:scaled(.bits_per_second;1000000)}]}
