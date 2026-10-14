# Task 5 交付报告：Docker 真机部署阻塞修复

本轮基于 PR #5 的 `fix/audit-batch4-test-harness`，修复协调者在 Ubuntu 22.04.5 / x86_64 真机 Docker 部署中确认的三个问题。没有修改 `main`，没有创建 tag 或 Release。

## BUG-A：构建上下文缺少根目录安装脚本

原配置把 API 构建上下文设为 `./backend`，`backend/Dockerfile` 只执行 `COPY . .`。安装器回归测试从 `internal/api` 读取 `../../../install.sh`，容器中实际解析到 `/install.sh`，导致三个 installer 测试失败。

修复内容：

- `docker-compose.yml` 的 API context 改为仓库根，Dockerfile 显式指定 `backend/Dockerfile`。
- `backend/Dockerfile` 显式复制 `backend/` 和根 `install.sh`，测试工作目录保持 `/src/backend`，因此 `../../../install.sh` 在镜像中真实存在。
- 新增根 `.dockerignore`，排除 Git、依赖目录、数据、备份、构建产物、日志和工作文件。
- 构建上下文实测为约 3.25 kB，远低于 50 MB 要求。

Docker build 日志中真实执行并通过了：

```text
=== RUN   TestInstallerRefreshesWeakSecretWithoutEvaluatingEnv
--- PASS: TestInstallerRefreshesWeakSecretWithoutEvaluatingEnv
=== RUN   TestInstallerRejectsMaliciousManifestBeforeExtraction
--- PASS: TestInstallerRejectsMaliciousManifestBeforeExtraction
=== RUN   TestInstallerExtractsIntoFreshDirectory
--- PASS: TestInstallerExtractsIntoFreshDirectory
...
 yuchen-panel-api  Built
```

完整输出见 [docker-build.log](docker-build.log)。

## BUG-B：格式测试依赖容器中不存在的 Python/PyYAML

`TestSubscriptionFormatRouting` 原先调用 `python3 -c 'import yaml'` 解析 Clash YAML；Go builder 镜像既没有 PyYAML，也不应把系统 Python 作为 Go 测试前置条件。

测试现在使用 Go 标准库对本测试所需的稳定结构进行检查：确认 `proxies:` 段存在，并统计该段的代理项。测试不再导入 Python、PyYAML 或启动外部解析进程。Docker builder 仍安装 `bash curl jq python3 unzip`，因为 installer 安全测试实际执行根 `install.sh` 的脚本路径；这属于被测安装器的运行依赖，和订阅格式测试无关。安装器测试没有 skip。

Dockerfile 还先以 `PATH=/usr/local/go/bin:/bin`（不含 Python）单独执行 `TestSubscriptionFormatRouting`，再用完整工具集执行全部 installer 回归测试，证明该格式测试本身不依赖 Python。

## BUG-C：开发机与 Docker/PR 门禁不一致

新增 `.github/workflows/ci.yml`，在 pull request、`main` push 和 `fix/**` push 触发：

- backend/agent 分别执行 `go build ./...`、`go vet ./...`、`go test -race ./...`；
- `docker compose build yuchen-panel-api` 和 `docker compose build yuchen-panel-frontend`；
- 全部 ShellCheck 与 Bash 语法检查；
- 前端 `npm ci`、`npm run build`（包含 `vue-tsc`）。

## 验证结果

| 检查 | 结果 |
| --- | --- |
| Docker API build，根上下文，installer 测试真实执行 | PASS，`yuchen-panel-api Built` |
| backend build/vet/race | PASS |
| agent build/vet/race | PASS |
| frontend 9 项测试与 `npm run build` | PASS |
| Bash syntax + ShellCheck | PASS |
| installer 测试 | PASS；日志显示真实 RUN/PASS，未因缺少脚本而 skip |
| T2–T6 真机设备连通 | `MANUAL_REQUIRED` |

前端 Docker 镜像的本地构建曾因 `registry.npmmirror.com` 网络请求无输出而中止；仓库现已加入 CI 构建门禁，API 镜像的阻塞根因和完整 Docker 验证已经通过。前端源码构建在 WSL2 本机通过，CI 会在独立 runner 上继续验证前端镜像。
