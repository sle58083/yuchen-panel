# 客户端订阅使用与验收

本分支是待验收版本，尚未合并 main、打 tag 或发布安装包。最低支持 **sing-box ≥1.12.0**。1.13.0 及以上必须使用面板导出的现代 DNS 格式；旧 DNS 可能 FATAL，1.14 已移除旧格式支持。**如你的 sing-box ≥1.13，请勿手工编辑导出的 DNS 段。** 不提供 `format=singbox-legacy`。

## 选择订阅格式

在“客户 → 分享 → 订阅”选择格式并复制链接。订阅只包含该客户明确获授权的有效节点和固定出口；空授权不会授予全部节点。Shadowsocks 已停用，旧记录须迁移到 VLESS/Reality。

| 格式 / URL 参数 | 返回内容 | 适用客户端（待用户实测） |
| --- | --- | --- |
| 自动，无 format | 根据 User-Agent；不能识别时 base64 | 显式格式比 UA 识别更可靠 |
| `format=clash`，别名 clash-meta/mihomo/yaml/yml | Clash/Mihomo YAML | Clash Verge Rev、Mihomo Party、Stash；使用支持 VLESS/Reality 的内核 |
| `format=singbox`，别名 sing-box/sb/json | 独立 sing-box JSON，现代 DNS | sing-box ≥1.12、SFA/SFI，以及支持完整配置导入的客户端构建 |
| `format=v2rayn`，别名 v2rayng/shadowrocket/base64 | base64 URI 列表 | v2rayN、v2rayNG、Shadowrocket；NekoBox/Matsuri 优先检查其具体版本的 URI 导入能力 |
| `format=raw` | 未编码 URI 列表 | 依赖旧版明文响应的工具 |

自动识别：Clash/Mihomo/Stash → YAML；sing-box/singbox/SFA/SFI → JSON；其余 → base64。显式参数优先，未知 format 回退 base64，日志只记选定格式和来源，不记订阅 token 或原始 UA。

**兼容性变化：** 本仓库旧实现默认返回明文 URI；本分支默认和未知格式改为 base64。依赖明文的工具加 `?format=raw`。订阅链接是 bearer 凭据，勿贴真实完整 URL。

## sing-box 桌面与手机模式

默认 JSON 使用 `127.0.0.1:2080` mixed 入站，桌面须让浏览器/系统代理指向此端口。配置自身不会接管系统流量。

手机 SFA/SFI 或桌面 VPN 选“VPN/TUN”，链接增加 `&profile=tun`：IPv4/IPv6 TUN、MTU 1400、自动路由、系统栈、DNS 劫持、自动识别出口网卡以避免回环。SFA/SFI 的 UA 以 `SFA/` 或 `SFI/` 开头且未指定 profile 时也默认 TUN；明确 `profile=mixed` 可覆盖。运行 TUN 时客户端需 VPN/管理员权限；导出 JSON 不改变服务器或本机路由。1400 是起始值，仍需按实际路径 MTU 验证。

NekoBox/Matsuri 的版本、内核和完整 JSON 导入入口存在差异；不要把 sing-box 配置粘进 URI 列表入口。优先验证当前构建，必要时改用通用 base64。流量额度目前仅展示，未接入强制限额。

## 自动下载脚本

Linux/WSL/Git Bash 需要 Bash、curl、jq、Python3。在本分支源码目录运行：

```bash
bash scripts/verify-clients.sh 'https://你的测试面板/sub/测试客户token'
```

脚本创建新的 `client-samples.XXXXXXXX` 目录，输出 `clash.yaml`、`singbox.json`、`singbox-tun.json`、`v2rayn.txt`、`single-node-links.txt` 和故意使用旧 DNS 的 `legacy-singbox.json`。最后一份仅用于 T1 负对照，请勿导入日常客户端。脚本拒绝覆盖已有文件，不运行代理、不运行 Docker、不修改系统网络。脚本打印实际 URI 数量与 YAML 字节数。

直接下载并校验本分支脚本（不安装面板）：

```bash
(
  set -Eeuo pipefail
  umask 077
  task_dir=$(mktemp -d)
  curl -fLsS --proto '=https' --proto-redir '=https' \
    'https://raw.githubusercontent.com/sle58083/yuchen-panel/fix/audit-client-compat/scripts/verify-clients.sh' \
    -o "$task_dir/verify-clients.sh"
  printf '%s  %s\n' 'bdf077119cdfb3f11db383c365c7486bed820f2eb9a1e5632550d40f610c66cc' "$task_dir/verify-clients.sh" | sha256sum -c -
  bash "$task_dir/verify-clients.sh" 'https://你的测试面板/sub/测试客户token'
)
```

没有域名时，可使用已建立的 SSH 隧道下载 `http://127.0.0.1:端口/sub/测试客户token`。生成的文件包含客户凭据，脚本默认使用 umask 077；提交诊断时只发结构/统计/脱敏日志。

## 本地可导入样例

本次交付的样例包含 12 个合成节点，覆盖 VLESS TCP/TLS/Reality/gRPC/WS、VMess、Trojan、IPv6、特殊字符密码和重复名称。订阅中有 12 个代理；sing-box outbounds 为 14 个（12 个节点 + selector + direct）。样例使用 `192.0.2.1`、`2001:db8::1` 文档地址，**只能验证解析、导入和字段，不能验证连通**。T2/T4 要在测试机配置可达节点和一致凭据。

本机临时 API：`http://127.0.0.1:18088/sub/audit-demo-local`。它只在本机运行，进程停止后链接失效；手机应导入下载文件或使用你部署的测试面板。未安装面板系统服务，未启动 Xray。此地址及样例 token 为公开测试数据。

如需从源码准备本分支测试产物，在 Linux/WSL 上克隆 `fix/audit-client-compat`，分别构建 backend/agent 和 frontend（依赖 Go ≥1.23、Node/npm）：

```bash
git clone --branch fix/audit-client-compat --single-branch https://github.com/sle58083/yuchen-panel.git yuchen-client-test
cd yuchen-client-test
mkdir -p test-artifacts
(cd backend && go build -o ../test-artifacts/yuchen-api ./cmd/server)
(cd agent && go build -o ../test-artifacts/yuchen-agent ./cmd/agent)
(cd frontend && npm ci && npm run build)
```

这生成测试产物，不自动安装或覆盖生产服务。使用现有部署方式在预置测试机运行该构建前，按 `.audit/FIX-REPORT.md` 配置强 JWT、显式管理员凭据并备份数据。当前没有已发布的本分支 ZIP，根安装器仍指向发布清单，不能用它验证未发布分支。

## T1–T6 用户实测

所有设备导入、代理运行、连通和负对照均由用户执行。Codex 经单独授权已对样例做 1.12.0/1.14.2 配置检查；这不能代替以下验收。

### T1 — 版本与 legacy 对照

在样例目录执行，使用绝对 volume 路径。官方镜像标签有 **v** 前缀：

```bash
for task_version in v1.13.0 v1.14.2; do
  docker run --rm --network none --read-only --cap-drop ALL \
    -v "$PWD/singbox.json:/c.json:ro" \
    "ghcr.io/sagernet/sing-box:$task_version" check -c /c.json
  printf 'check %s exit=%s\n' "$task_version" "$?"
done
docker run --rm --network none --read-only --cap-drop ALL \
  -v "$PWD/legacy-singbox.json:/c.json:ro" \
  ghcr.io/sagernet/sing-box:v1.13.0 check -c /c.json
printf 'legacy negative control exit=%s\n' "$?"
```

现代配置合格条件：退出码 0、无 FATAL；`check` 成功可没有 stdout，不要求输出固定字符串。旧 DNS 对照应非零，记录真实 legacy DNS 报错；不要设置 `ENABLE_DEPRECATED_LEGACY_DNS_SERVERS` 绕过验证。另对 `singbox-tun.json` 重复现代配置检查，然后在目标客户端真正运行并验证 VPN。

### T2 — Android/iOS URI

从 `single-node-links.txt` 导入含 `p@ss w#rd&test=1` 密码的 Trojan IPv6 链接。v2rayNG/Shadowrocket 应保留名称、密码和 IPv6 端口。样例仅验证解析；在可达 IPv6 测试节点配置同样密码再测试连接。记录客户端版本、内核版本、字段是否保留和真实连接结果。

### T3 — 规模

导入 12 节点的 `clash.yaml` 和 `v2rayn.txt`；Clash 代理组和 sing-box selector 名称应唯一。记录实际导入数量、文件大小和卡顿/截断现象。大量节点优先使用 URL 或文件导入；URI 客户端可改 base64，Clash 仍需 YAML。

### T4 — 协议覆盖

VLESS TCP/TLS/Reality/gRPC/WS、VMess、Trojan 各选择真实节点：导入 → 核对参数 → 选择代理 → 连接。Reality publicKey/shortId/SNI 必须与服务端一致；当前服务端没有配置 Vision flow，导出不会强行添加。TLS 证书加载缺陷 P2-012 尚未完整修复，导入成功不能证明 TLS 服务端可用。不要为掩盖证书错误打开 insecure。

### T5 — 后台恢复

Android/iOS 导入后切后台再返回，确认节点保留、选择状态和订阅刷新表现；记录系统版本与 VPN 权限。`Cache-Control: no-store` 不等于应用会保留或清空本地配置，须看设备实测。

### T6 — SS 迁移 UI

在测试面板确认新建节点没有 SS 选项、旧 SS 记录显示迁移 banner。导出的节点应没有 SS；避免宽泛 `grep 'ss://'`，它也会匹配 `vless://`。可执行：

```bash
if grep -nE '^ss://' single-node-links.txt; then echo 'FAIL: SS URI found'; else echo 'No SS URI'; fi
jq '[.outbounds[] | select(.type == "shadowsocks")] | length' singbox.json
```

### 回贴模板

```text
测试分支/commit:
平台/OS/客户端版本/内核版本:
T1: 1.13.0 exit= ; 1.14.2 exit= ; legacy exit= ; 原始报错:
T2: v2rayNG字段/连通= ; Shadowrocket字段/连通=
T3: 节点数= ; YAML字节= ; 导入耗时/截断=
T4: VLESS TCP/TLS/Reality/gRPC/WS、VMess、Trojan 各自结果/报错=
T5: 后台恢复/刷新结果=
T6: SS选项/旧节点banner/导出SS数量=
```

只发失败项上下文、退出码、数量和脱敏字段；移除公网 IP、UUID、密码、订阅 token、Authorization。未执行的项写“未执行”，不要写推测结果。

## 自动验证范围与剩余限制

真实的编译、竞态、覆盖率、结构校验和 Docker check 原始输出见 `.audit/CLIENT-FIX-REPORT.md`。T1–T6 均为 **MANUAL_REQUIRED**。NekoBox/Matsuri 导入 UI、iOS 平台差异、路径 MTU、网络 QoS 与跨国吞吐量仍需实测。配置修复不能证明已消除最初的“几兆”性能问题。

二维码继续采用现有单节点 URI。本次 12 节点样例的 URI 长度为 133–260 UTF-8 字节；这不是客户端的长度上限。长度随名称、密码、transport 参数增长，没有统一客户端截断阈值；遇到截断应改 URL/文件导入或已有 `/s/{token}/{nodeID}` 短链接，短链接本身也含凭据且并非所有扫码入口都支持 HTTP。此次仅评估，不新增短链机制。

官方依据：[现代 UDP DNS](https://sing-box.sagernet.org/configuration/dns/server/udp/)、[迁移说明](https://sing-box.sagernet.org/migration/)、[TUN](https://sing-box.sagernet.org/configuration/inbound/tun/)、[官方镜像版本](https://github.com/SagerNet/sing-box/pkgs/container/sing-box)。
