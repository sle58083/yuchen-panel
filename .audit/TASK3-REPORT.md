# TASK3-REPORT — C11–C14

日期：2026-10-13（Asia/Shanghai）
分支：`fix/audit-batch3-client-uri`（C11–C14；未合 main/tag/release）

## C11：URI 编码（BUG-1）

验收复核确认 `encodeURIComponent` 会编码 `&`、`=`，但它保留 RFC3986 userinfo 中危险的 `!'()*`。前端新增 `encodeUserInfo`，只用于 Trojan 密码；订阅 token、format、fragment 仍使用 `encodeURIComponent`，并在调用点注释其组件语义。后端继续在最终 URI 生成时由 `net/url` 编码，model.Client.Password 不做预编码。

端到端测试使用同一密码：`&=+,\'" /中文😀:@`，覆盖 Go Trojan 链接和前端 Trojan 链接，标准库解析后逐字节还原。还验证 IPv6、默认 gRPC 与 VMess 传输。前后端新增测试均实际运行通过。

## C12：sing-box 三版本矩阵（GAP-1）

现代配置三版本退出 0；legacy 1.12 退出 0 并输出 deprecation warning；1.13/1.14 退出 1 并输出 FATAL。所有镜像标签使用 `v` 前缀；失败的无 v 标签拉取仅保留在先前日志，不计入矩阵。

```text
COMMAND: docker run --rm --network none --read-only --cap-drop ALL -v <sample>:/c.json:ro ghcr.io/sagernet/sing-box:v1.12.0 check -c /c.json file=modern version=1.12.0
EXIT_CODE: 0
COMMAND: docker run --rm --network none --read-only --cap-drop ALL -v <sample>:/c.json:ro ghcr.io/sagernet/sing-box:v1.12.0 check -c /c.json file=legacy version=1.12.0
[33mWARN[0m[0000] legacy DNS servers is deprecated in sing-box 1.12.0 and will be removed in sing-box 1.14.0, checkout documentation for migration: https://sing-box.sagernet.org/migration/#migrate-to-new-dns-server-formats
EXIT_CODE: 0
COMMAND: docker run --rm --network none --read-only --cap-drop ALL -v <sample>:/c.json:ro ghcr.io/sagernet/sing-box:v1.13.0 check -c /c.json file=modern version=1.13.0
EXIT_CODE: 0
COMMAND: docker run --rm --network none --read-only --cap-drop ALL -v <sample>:/c.json:ro ghcr.io/sagernet/sing-box:v1.13.0 check -c /c.json file=legacy version=1.13.0
[31mERROR[0m[0000] legacy DNS servers is deprecated in sing-box 1.12.0 and will be removed in sing-box 1.14.0, checkout documentation for migration: https://sing-box.sagernet.org/migration/#migrate-to-new-dns-server-formats
[31mFATAL[0m[0000] to continuing using this feature, set environment variable ENABLE_DEPRECATED_LEGACY_DNS_SERVERS=true
EXIT_CODE: 1
COMMAND: docker run --rm --network none --read-only --cap-drop ALL -v <sample>:/c.json:ro ghcr.io/sagernet/sing-box:v1.14.2 check -c /c.json file=modern version=1.14.2
EXIT_CODE: 0
COMMAND: docker run --rm --network none --read-only --cap-drop ALL -v <sample>:/c.json:ro ghcr.io/sagernet/sing-box:v1.14.2 check -c /c.json file=legacy version=1.14.2
[31mFATAL[0m[0000] decode config at /c.json: dns.servers[0]: legacy DNS server formats are deprecated in sing-box 1.12.0 and removed in sing-box 1.14.0, checkout migration: https://sing-box.sagernet.org/migration/#migrate-to-new-dns-server-formats
EXIT_CODE: 1

```

## C13：节点命名（GAP-2 / P2-021）

前后端都把名称输出为 `显示名 #短ID`。短 ID 从节点 ID 的稳定字母数字后缀生成；Clash proxy、sing-box tag、VMess remark、URI fragment 和 selector/group 引用共享该值。12 个同名节点测试断言 name/tag 唯一，并能通过短 ID 反查来源节点。

## C14：实机支持包

- `scripts/prepare-test-env.sh`：要求 `YC_ENV != production`，调用下载器生成独立样例目录；不会写生产配置。
- `docs/TEST-CHECKLIST.md`：包含具体客户端版本、导入入口、判定标准、失败症状和填空回贴模板。
- `scripts/diagnose-import.sh`：收集 sing-box JSON、base64 URI 摘要、结构字段和脱敏前证书探针；输出目录默认 0700，分享前必须删掉 token/UUID/密码/IP。
- T1–T6 仍由用户执行。Codex 没有代跑设备导入、真实连接或后台恢复。

## 自动验证原始结果

- 前端 Node tests：8/8 pass，0 fail。
- backend `go test -race ./internal/api`：退出 0；backend 总覆盖率 20.1%，agent 12.6%，race 退出 0。
- `bash -n`：全部脚本退出 0。
- `shellcheck --severity=error scripts/*.sh deploy/install.sh`：退出 0。
- 完整 ShellCheck：退出 0，既有 SC2034 已改为 `_`。
- 1.12.0/1.13.0/1.14.2 × modern/legacy：原始六条结果见 `.audit/singbox-matrix.txt`。

## 未完成与诚实边界

T2–T6 必须等用户回贴真实结果后才能改变 `MANUAL_REQUIRED`。文档 IP 不可连接；配置 check 不能证明客户端连接、IPv6 路由、证书链、QoS 或吞吐量。C5–C8、版本单一来源、升级回滚和 C9 CI/release 仍待后续批次。P2-012 TLS 证书模型与 P2-008 配额执行仍未完成。

## 仓库卫生

提交前检查目标：无 `.env`、data、backups、node_modules、构建产物、真实凭据或私钥；仅提交测试 fixture、原始输出和文档。法律/SPDX 文件未修改。
