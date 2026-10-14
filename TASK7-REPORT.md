# Task 7 收尾报告

## C19 版本单一来源

Agent 使用 `var version = "dev"`，发布构建通过 `-ldflags "-X main.version=$(cut -d- -f1 VERSION)"` 注入。该注入已同步到 `Makefile`、`.github/workflows/release.yml`、`scripts/build-fast-release.sh` 和 `deploy/agent-install.sh`；CI 输出并断言 `agent/` 与 `backend/internal/version/` 中只有 `backend/internal/version/version.go` 的语义版本定义。未注入时 Agent 明确报告 `dev`。

## C20 Version 与 Channel

`VERSION` 保留发布标识 `1.0.2-stable`；安装器解析出 `VERSION=1.0.2` 和 `CHANNEL=stable` 写入 `panel.info`。API health、存储数据和升级比较使用纯语义版本 `1.0.2`。`stable` 只作为发行频道元数据，升级校验将其视为版本元数据而不是预发布标识。

## C21 升级校验

移除了 `HasPrefix` 版本匹配；更新可用性比较对现代版本也使用同一比较器，历史四/五段版本保留数字兼容路径。Go 侧使用零依赖的最小语义版本解析器：容忍前导 `v`，缺失 patch 按 0，`rc` 低于正式版，`stable` 为频道元数据。runner（独立于会被重启的 API 进程）使用现有 Python 依赖，从 `/api/health` 解析 JSON 的 `version` 字段后再比较；没有另起 API 内的 Go 轮询协程，未使用 `grep -q "$TARGET_VERSION"`。

真实命令 `go test -v -run "Version|Compare" ./...` 已通过，覆盖：`1.0.2 != 1.0.20`、`1.10.0 > 1.9.0`、`1.0 == 1.0.0`、前导 `v`、空值、`1.0.2-rc.1 < 1.0.2`。

## C22 Docker 选择

按用户确认保留选项②：Docker 模式明确跳过宿主 Agent。部署文档和测试清单顶部均说明 API/UI 可用，但没有 Agent 时节点不会写入并重载 Xray，B 档不可行，正式路径是 fast 一行安装。安装器现在在备份和删除旧容器前完成 Docker、Compose v2 版本和 `compose config` 预检（拒绝 v1），失败会保留旧运行时。

## C23 Reality 默认值

新建节点未指定安全方式时，VLESS 默认 `reality`，Trojan 默认 `tls`；显式 `none` 保留。更新请求省略 security 时沿用存量值，因此旧的 `none` 不会被改写。VMess 继续使用 `none`，因为 Xray VMess 入站不支持 Reality。前端把 Reality 置首并标注“推荐”，但仍保留 `none`/`tls` 可选项；Reality 保存前自动调用 `/api/nodes/reality-keys`。

Before（VLESS 未指定 security）：`streamSettings.security = "none"`。
After：`streamSettings.security = "reality"`，并生成 Reality key、SNI、short ID。
Trojan 未指定 security 的生成结果保持 `streamSettings.security = "tls"`。

## Agent reload 与延迟

Agent 只有在写入临时文件、Xray 检查、原子替换成功后才执行 reload；原逻辑每轮重新读取磁盘 hash，reload 失败后可能误判已同步。本轮加入 reloadPending：失败后向同步 API 报告空 hash，直到 reload 成功才清除，下一轮会重试，避免 P1-009 的“失败后错误记完成”。新增重载成功日志；空 reload 命令返回错误；待重载标记跨 Agent 进程重启保留。成功日志记录实际命令。端口在 systemd/Xray 启动后可能晚于配置写入出现，`config applied` 只证明 reload 命令成功，不等同于端口已就绪。

## 验证

- `go test -v -run "Version|Compare" ./...`：通过。
- backend、agent：build/vet/race 全部通过。
- 前端生产构建与 9 项分享测试：通过（Linux Node 环境）。
- `bash -n scripts/*.sh deploy/*.sh install.sh`：通过（WSL）。
- `git diff --check`：通过。
- 升级密钥：安装器已有 `JWT_SECRET=$(random_string 64)`，写 `.env` 时过滤旧 `YC_JWT_SECRET`；因此旧短密钥被替换，已有登录令牌失效。
- 未合并 main、未打 tag、未触发发布。
