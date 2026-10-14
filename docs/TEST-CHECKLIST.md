# T1–T6 客户端验收清单

> ⚠️ **Docker 模式不包含 yuchen-agent。** 面板 API 与 Web UI 可用，但没有 Agent 写入并重载 Xray，创建的节点不会生效，B 档真实连通测试不可行。需另行部署匹配架构的 `yuchen-agent`。正式用户请使用 fast 模式一行安装；Release 包含预编译 Agent。


先按 [DEPLOY-TO-TEST-MACHINE.md](DEPLOY-TO-TEST-MACHINE.md) 在一次性 Ubuntu/Debian 机器部署 Task 4 分支，执行 `prepare-test-env.sh --dry-run --allow-existing`，确认计划后由操作员执行 `--apply --allow-existing`。脚本输出的 `TEST-ENV.json` 是本次验收的资源及样例 SHA256 对照。所有设备导入、代理启动和连接测试均由用户亲自执行；在用户提交结果前，T2–T6 状态为 **MANUAL_REQUIRED**。

判定分两档，回贴时每个用例必须写实际执行了哪一档：

| 档位 | 核对内容 | 前置条件 |
| --- | --- | --- |
| A — 导入字段 | 节点数量、名称、协议与传输参数、密码逐字一致，且没有乱码、重复或体积告警 | 合成样例也可用；`192.0.2.0/24` 和 `2001:db8::/32` 只能用于这一档 |
| B — 实际连通 | 客户端发起连接，服务端握手成功，指定节点可访问测试 URL；记录握手/路由错误 | 必须是实际可达的测试机节点、Xray 入站、证书、开放端口和路由 |

A 通过不代表 B 通过。IPv6 探测只证明测试机有默认路由、全局地址和本机 bind 能力，不能证明外部设备可达；若 C15 显示 IPv6 不可用，T2 的 IPv6 B 项写“**本次跳过，原因：<脚本原文>**”，不能写“通过”。TLS 与 IPv6 入站仍有服务端限制，详见部署文档第 3 节。

## T1 — sing-box 配置矩阵（自动配置检查）

Codex 在 WSL2 对 modern 和 legacy 配置运行过 1.12.0、1.13.0、1.14.2 的 `check`。modern 三版退出 0；legacy 在 1.12 只有 WARN、1.13/1.14 为 FATAL。原始输出见 `.audit/singbox-matrix.txt`。这属于静态配置检查，不能替代设备连接。

## T2 — 特殊密码与 IPv6（A + B）

**客户端：** Android v2rayNG、iOS Shadowrocket，记录实际版本。**A 档：** 从 `single-node-links.txt` 导入 `uri-edge-trojan`，核对密码 `&=+,'" /中文😀:@`、节点名、端口、SNI、IPv4/IPv6 地址；密码应逐字一致，没有乱码。现有合成样例也能做 A 档。**B 档：** 只用 C15 产出的真实测试机节点，实际连接并访问测试 URL，记录成功或客户端/服务端错误；若无 IPv6，IPv6 B 项必须跳过并抄录脚本原因。TLS 入站可能因缺证书而不能启动，需单独记录服务端证据。

```text
T2 A：未执行 / 通过 / 未通过；客户端与版本：________；密码/名称/地址结果：________
T2 B IPv4：未执行 / 通过 / 未通过 / 阻塞；端口/握手/访问结果：________
T2 B IPv6：未执行 / 通过 / 未通过 / 本次跳过；原因或证据：________
```

## T3 — 12+ 节点规模（仅 A）

**客户端：** Windows Clash Verge Rev、iOS Stash，记录实际版本。C15 只建六个真实节点；此项继续使用交付的 **12 节点合成** `clash.yaml` 与 `v2rayn.txt`，不要尝试连接文档地址。导入后核对 12 个节点、唯一名称、代理组可选、无截断或体积告警，并记录 YAML 字节数和导入耗时。节点缺失时检查订阅权限及客户端解析；代理组空时检查 YAML 名称引用。

```text
T3 A：未执行 / 通过 / 未通过；客户端与版本：________；节点数：____；YAML 字节：____；耗时：____
```

## T4 — 协议覆盖（A + B）

**客户端：** sing-box 桌面版、Clash Verge Rev、v2rayNG，记录实际版本。**A 档：** 导入 C15 的六个真实节点，逐个核对 VLESS TCP、VLESS Reality 的 SNI/publicKey/shortId、VLESS gRPC 的 service name、VLESS WS 的 path、VMess TCP 的加密 `auto`、Trojan TLS 的密码。**B 档：** 逐个实际连接、记录握手与访问结果，不把 Xray 未监听、TLS 证书缺失或 IPv6 路由失败归咎于客户端。gRPC/WS/Trojan TLS 在当前服务端配置模型下可能被阻塞；任何未连通节点不能算 T4 整体通过。

```text
T4 A：未执行 / 通过 / 未通过；六种节点参数逐项结果：________
T4 B：未执行 / 通过 / 未通过 / 阻塞；VLESS TCP：____ Reality：____ gRPC：____ WS：____ VMess：____ Trojan：____
```

## T5 — 后台恢复（仅 B）

**客户端：** Android v2rayNG、iOS Shadowrocket。先在真实节点上确认 B 档可连接，再切后台 60 秒、返回前台、访问测试 URL 并手动刷新订阅。记录 VPN 状态、节点是否保留、刷新是否清空。若所有可用节点因服务端问题不能建立初始连接，写“阻塞”，不能用合成样例冒充通过。

```text
T5 B：未执行 / 通过 / 未通过 / 阻塞；Android：____；iOS：____；60 秒后访问/刷新结果：________
```

## T6 — Shadowsocks 下线（A + UI 人工确认）

**A 档：** 在 C15 导出目录中确认 `single-node-links.txt` 无 `ss://` 行、`singbox.json` 无 Shadowsocks outbound。**UI：** 测试面板新建节点的协议下拉无 SS，打开存量 SS 记录时看到迁移 banner。UI 必须由用户在测试机确认；没有存量记录则该子项写“未执行”。

```bash
if grep -nE '^ss://' single-node-links.txt; then echo FAIL; else echo 'No SS URI'; fi
jq '[.outbounds[] | select(.type == "shadowsocks")] | length' singbox.json
```

```text
T6 A：未执行 / 通过 / 未通过；导出 SS 数量：____
T6 UI：未执行 / 通过 / 未通过；新建 SS 选项：____；存量迁移 banner：____
```

## 脱敏回贴模板

```text
分支/commit：
TEST-ENV.json 的样例 SHA256（不要贴订阅文件原文）：
测试机 IPv6 探测状态与原因（IP 脱敏）：
客户端 / OS / 内核版本：
T2 A：      T2 B IPv4：      T2 B IPv6：
T3 A：
T4 A：      T4 B（六节点逐项）：
T5 B：
T6 A：      T6 UI：
服务端 Xray/Agent 监听及错误摘要（脱敏）：
```

删除订阅 token、JWT、UUID、密码、公网 IP 和 Authorization 后再提交日志。未执行项写“未执行”，被服务端缺陷阻塞则写“阻塞”及证据；不会由 A 档结果推断 B 档成功。
