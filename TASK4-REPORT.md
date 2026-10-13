# Task 4 交付报告：T2–T6 实机验证前置

本分支 `fix/audit-batch4-test-harness` 基于 PR #4 的提交 `7854bf1`，补齐 C15–C18。没有执行真实测试设备导入/连接，也没有在用户面板上执行 `--apply`；T2–T6 的 B 档仍为 `MANUAL_REQUIRED`。

## C15：真实测试环境脚本

`scripts/prepare-test-env.sh` 现在默认 dry-run，只有显式 `--apply` 才会调用 API。脚本要求 `YC_ENV=test`，检查 `/opt/yuchen-panel/data/yuchen-panel.json`（以及可选的 `YC_DATA_PATH`）非空时必须传 `--allow-existing`，并在开始处打印面板 URL、数据路径、输出目录、节点端口和操作模式。

apply 流程使用管理员 JWT `YC_TOKEN`，按固定用户名和节点名查重并复用，创建或更新特殊密码客户 `uri-edge-case`，生成 VLESS TCP/Reality/gRPC/WS、VMess TCP、Trojan TLS 六个资源。节点地址来自测试面板服务器或显式 `TEST_NODE_HOST`；文档地址和 loopback 被拒绝。IPv6 只有在存在 `::/0`、全局 `2000::/3` 地址且 bind/socket 探测成功时才使用；否则输出实际原因，节点使用可核对的 IPv4，并将 T2 IPv6 B 项标记为跳过。

apply 后调用 `verify-clients.sh`，并阻断以下任一失败：sing-box modern DNS 字段、sing-box tag 唯一性、Clash name 唯一性、Trojan 特殊密码 URL 往返、输出中出现 `ss://`。`TEST-ENV.json` 只记录测试机地址、IPv6 状态、资源 ID、样例路径和 SHA256，不记录密码或订阅 token；样例目录仍含凭据，应保持本地权限并脱敏后再提交。

隔离临时面板的 apply 冒烟结果：实际创建 6 个节点和 1 个客户，下载 6 条 URI，生成 Clash YAML 2483 字节，四项自检全部打印 `SELF_CHECK_PASS`，最终 `TASK4_APPLY_SMOKE_PASS`。该冒烟使用临时 JSON 数据和本地 API，不能证明公网连接成功。

真实 dry-run 输出（WSL2，未发 API 请求）：

```text
Environment: YC_ENV=test
Panel URL: http://127.0.0.1:8088
Guarded data path: /opt/yuchen-panel/data/yuchen-panel.json
Output path: ./.audit/raw/task4/client-test-env
Mode: dry-run
Plan: inspect the test server; create/reuse uri-edge-case and six named nodes (VLESS TCP/Reality/gRPC/WS, VMess TCP, Trojan TCP).
Node ports: 24401-24406. Verify that these ports are free and reachable before real connection tests.
IPv6: unavailable; T2 IPv6 connection is SKIPPED: no IPv6 default route (::/0)
DRY_RUN: no API calls, no output files, no clients or nodes created.
```

## C16：干净测试机部署路径

新增 [docs/DEPLOY-TO-TEST-MACHINE.md](docs/DEPLOY-TO-TEST-MACHINE.md)，说明 Ubuntu 22.04/Debian 12 检查、检出 PR 分支、源码 Docker 构建、JWT 长度升级风险、HTTP 默认状态和 SSH 隧道。文档明确指出根 `install.sh` 会按 `version.json` 下载旧 Release，不能用它安装未发布 PR；测试部署必须从当前 PR 源码运行 `deploy/install.sh`。

## C17：A/B 判定分档

重写 [docs/TEST-CHECKLIST.md](docs/TEST-CHECKLIST.md)：T2 为 A+B，T3 为 A，T4 为 A+B，T5 为 B，T6 为 A+UI。合成地址只能用于 A 档，IPv6 不可用时必须写“本次跳过 + 实际原因”，不能把导入成功推断为连通成功。回贴模板会过滤 token、JWT、UUID、密码和公网 IP。

## C18：前后端命名一致

前端 `safeName` 去掉客户用户名后缀，并使用与后端 `nodeShareName` 相同的持久化 node ID 短后缀。前端内部去重不再覆盖原始 `name`，避免第二次格式化得到 `节点 #id #id`。新增前端测试和 Go 测试，确认同一节点显示为 `重复名称 #de00`，不随客户用户名变化。

## 验证记录

| 检查 | 结果 |
| --- | --- |
| `bash -n scripts/prepare-test-env.sh` | PASS |
| `shellcheck --severity=error scripts/*.sh deploy/install.sh` | PASS |
| 全部脚本 `shellcheck scripts/*.sh deploy/install.sh` | PASS |
| 前端 `node --test frontend/tests/share.test.mjs` | 9/9 PASS |
| backend `go build`, `go vet`, `go test -race ./...` | PASS，覆盖率 20.1% |
| agent `go build`, `go vet`, `go test -race ./...` | PASS，覆盖率 12.6% |
| C15 dry-run 生产保护和 apply 隔离冒烟 | PASS |
| T2–T6 真实设备导入/连接 | `MANUAL_REQUIRED` |

原始日志位于 `.audit/raw/task4/`，包括 dry-run、backend、agent、frontend-tests 和 shellcheck。未运行真实设备，也没有伪造任何连接日志。
