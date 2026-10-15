# 网络热修复与性能诊断

此工具不保证解除线路限速。先比较同一客户端、同一目标的裸 TCP、系统代理与 TUN 单连接速度。

## 已安装服务器：免编译修复

在本仓库源码目录执行（Debian/Ubuntu、systemd、root Xray 服务）：

```bash
sudo apt-get update
sudo apt-get install -y jq util-linux
sudo bash deploy/network-hotfix/install-xray-hotfix.sh --vision-tags '[]' --restart
```

`--restart` 会中断现有连接；省略它可以先安装，稍后手工重启。
该安装器替换 Xray 的 ExecStart，适用于单个 `xray run -config` 服务；有自定义启动参数、沙箱或配置目录时应先适配。

每次启动时，Hook 读取 `/etc/yuchen-panel/xray/config.json`，删除所有 dokodemo-door 入站的 sniffing，经 jq 与 Xray 校验后启动独立的 `/run/yuchen-xray-hotfix/config.json`。
面板源文件保持不变，避免 Agent 根据源文件哈希反复下发和重启。无需 inotifywait。
自定义源路径用 `--config /absolute/path/config.json`。

## 可选 Vision

默认不修改 flow，避免现有客户端失联。仅支持指定标签的 VLESS + Reality + TCP/RAW 入站：

```bash
sudo jq -r '.inbounds[] | select(.protocol=="vless" and .streamSettings.security=="reality") | [.tag,.port,(.streamSettings.network // "tcp")] | @tsv' /etc/yuchen-panel/xray/config.json
```

首次安装可以使用 `--vision-tags '["selected-inbound-tag"]'`。
已安装时，修改 `/etc/yuchen-panel/hotfix/hook.env` 的 `VISION_TAGS`，再运行：

```bash
sudo /etc/yuchen-panel/hotfix/start.sh --check
sudo systemctl restart xray
```

指定入站的所有用户都会设为 `xtls-rprx-vision`；对应客户端必须同步设置相同 flow。
面板订阅仍输出空 flow，刷新订阅可能覆盖客户端修改。该 Hook 不重写订阅。
透传入口只删除 sniffing；Vision 应设置在实际终止 VLESS/Reality 的落地入站。
WS/gRPC、未知标签、重复标签和不兼容的现有 flow 会被拒绝。

安装器打印 `/var/tmp/yuchen-hotfix-backup.*/uninstall.sh` 的实际路径。
以 root 执行该卸载脚本可移除 Hook 并重启原服务；不要使用通配符选择多个备份。
备份含敏感配置，仅 root 可读。

## 网络调优与脱敏采集

```bash
sudo bash scripts/network/tune-network.sh
# 在持续下载期间，将 25642 换成实际节点端口
sudo bash scripts/network/collect-network-diagnostics.sh 25642 > diagnostic-summary.txt
```

调优脚本备份实时参数并提供回滚；不会自动重启 Xray或覆盖当前网卡 qdisc。
64 MiB 是 socket 缓冲按需增长上限，小内存高并发主机应评估内存压力。
默认 qdisc 改动不保证当前接口立刻使用 fq；用 `tc -s qdisc show` 核实。
文件句柄上限需要服务重启后生效，BBR 应检查实际新建连接。

采集器仅输出性能字段、配置结构投影和日志错误类别计数，不输出端点 IP、UUID、密码、私钥或原始日志。
默认三次 socket 样本，每次最多十条，编号在每个样本内有效。
接口计数是累计值；防火墙特征计数仅用于筛查，不能证明限速规则实际生效。
使用 `scripts/network/diagnostic-submission-template.md` 描述版本、角色、测速条件与客户端参数。

受控测试端启动 iperf3 服务后，可提取不含 IP 的摘要：

```bash
iperf3 -c PEER_IP -t 30 -P 1 -J | jq -f scripts/network/iperf-summary.jq
iperf3 -c PEER_IP -t 30 -P 8 -J | jq -f scripts/network/iperf-summary.jq
iperf3 -c PEER_IP -t 30 -P 1 -R -J | jq -f scripts/network/iperf-summary.jq
```

## 发布边界

源码修复关闭固定 TCP 转发的嗅探；后端重新构建后生效。
上述免编译 Hook 可立即用于已有安装，不必等待面板新二进制。
本变更不修改 version.json，不冒用旧发布包校验和，不覆盖已有 Release。
入口 install.sh 仍下载 version.json 指向的发布包；发布新二进制与更新清单是独立步骤。
