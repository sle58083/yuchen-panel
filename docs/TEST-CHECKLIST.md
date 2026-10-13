# T1–T6 客户端验收清单

这些步骤验证真实客户端和可达节点。合成样例使用 `192.0.2.1`/`2001:db8::1`，只能验证导入字段，不能连接。部署测试分支前备份测试数据；不要把生产订阅 token 粘进日志。

## 准备

```bash
export YC_ENV=test
bash scripts/prepare-test-env.sh 'https://测试面板.example/sub/测试token' ./client-test-samples
```

没有域名时使用 SSH 隧道后的 `http://127.0.0.1:端口/sub/测试token`。脚本拒绝 `YC_ENV=production`，只写新的样例目录。

## T1 — sing-box 版本矩阵（自动）

**执行者：** Codex 已在 WSL2 执行；用户可复核。**判定：** modern 三个版本退出 0；legacy 的 1.12 退出 0 且有 deprecation warning，1.13/1.14 退出非 0 且 FATAL。原始六条命令见 `singbox-matrix.txt`。所有镜像标签有 `v` 前缀。

## T2 — 特殊密码与 IPv6

**客户端：** Android v2rayNG 1.10.19、iOS Shadowrocket 2.2.72。**位置：** v2rayNG 的订阅设置/从剪贴板导入；Shadowrocket 的添加服务器/从剪贴板。**动作：** 从 `single-node-links.txt` 导入 Trojan IPv6 链接，核对密码、节点名和地址，再用真实可达 IPv6 节点连接。**通过：** 密码含 `& = + , ' " 空格 # @ : / 中文 emoji` 仍逐字一致，名称不乱码，连接成功。导入成功但超时→检查地址/端口、防火墙、IPv6 路由或服务端 TLS/Reality；密码截断→客户端 URI 解析或旧内核；名称乱码→剪贴板/文件编码。

☐ 完成　实际结果：________________

## T3 — 12+ 节点规模

**客户端：** Windows Clash Verge Rev 2.4.5（Mihomo 内核）与 iOS Stash 1.7.5。**位置：** Clash Verge “Profiles → Import URL”；Stash “配置 → 添加订阅”。**动作：** 导入 `clash.yaml`，再导入 `v2rayn.txt` 到支持的 URI 客户端。**通过：** 12 个节点均出现、代理组可选、无体积/截断告警；记录 YAML 字节数与导入耗时。节点少→检查授权和客户端解析限制；代理组空→检查 YAML 格式/名称引用；卡顿→检查文件大小和客户端限制。

☐ 完成　节点数：____　YAML 字节：____　实际结果：________________

## T4 — 协议覆盖

**客户端：** sing-box 1.14.2 桌面版、Clash Verge Rev 2.4.5、v2rayNG 1.10.19。**动作：** 每种真实节点各导入一条：VLESS TCP/TLS/Reality/gRPC/WS、VMess、Trojan；核对 SNI、Reality publicKey/shortId、gRPC service name、WS path 后连接。**通过：** 参数正确且每条实际连接成功。Reality 超时→公钥/shortId/SNI；TLS 失败→证书链或服务端证书模型（P2-012 尚未完全修复）；gRPC/WS 失败→反代 path/service name 或缓冲；VMess 失败→UUID/transport。

☐ 完成　各协议结果：________________________________

## T5 — 后台恢复

**客户端：** Android v2rayNG 1.10.19、iOS Shadowrocket 2.2.72。**动作：** 连接成功后切后台 60 秒，再回前台；检查节点、选择状态和订阅列表，再手动刷新。**通过：** 节点不丢、VPN 可恢复、刷新不清空配置。失败对照：VPN 被系统杀→电池优化/后台权限；订阅空→缓存策略或 token；恢复慢→系统网络重建。

☐ Android：____　☐ iOS：____　实际结果：________________

## T6 — Shadowsocks 下线

Codex 已验证导出过滤和前端代码；用户需在测试面板确认 UI。新建节点查看协议下拉；打开存量 SS 记录；导出所有格式并检查 SS URI。**通过：** 新建无 SS；存量显示迁移 banner；导出没有以 `ss://` 开头的行。

```bash
if grep -nE '^ss://' single-node-links.txt; then echo FAIL; else echo 'No SS URI'; fi
jq '[.outbounds[] | select(.type == "shadowsocks")] | length' singbox.json
```

☐ 完成　SS 选项：____　banner：____　导出数量：____　实际结果：________________

## 回贴模板

```text
分支/commit：
平台、OS、客户端和内核版本：
T1：modern 1.12= / 1.13= / 1.14=；legacy 1.12= / 1.13= / 1.14=；完整报错：
T2：v2rayNG=；Shadowrocket=；真实 IPv6 连接=；
T3：节点数=；YAML 字节=；耗时=；
T4：VLESS TCP/TLS/Reality/gRPC/WS、VMess、Trojan：
T5：Android=；iOS=；
T6：新建 SS=；旧记录 banner=；导出 SS 数量=；
```

只提交脱敏后的退出码、数量和错误上下文；移除订阅 token、UUID、密码、公网 IP 与 Authorization。未执行项必须写“未执行”，不会被推断为通过。