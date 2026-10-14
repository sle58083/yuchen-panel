# Task 8：Docker 安装路径与端到端测试

## BUG-K

`SRC_DIR` 是源码/构建目录；`APP_DIR` 是安装运行目录，持有 `.env` 和数据。预检现在显式使用 `SRC_DIR/docker-compose.yml`：已有 `.env` 时通过 `--env-file APP_DIR/.env` 提供真实值；全新安装时在独立子 shell 注入仅供校验的占位值。父 shell 不导出这些值。实际启动显式使用真实 `.env`，并在子 shell 中 unset 外部 JWT/管理员/Agent 密钥变量，避免环境覆盖。

Docker/Compose/配置预检仍在备份与 cleanup 之前；失败保持 `existing runtime was left untouched`，不撤回 C22 的 v1 拒绝逻辑。

## BUG-L

apt 按顺序检查并尝试 `docker-compose-v2`、`docker-compose-plugin`、`docker-compose`；不在 apt 源中的包先跳过，避免缺包报错。yum 尝试 plugin/compose。只有 `docker compose version` 成功才接受安装结果，v1 包安装成功不算通过。全部失败时显示 Ubuntu/Debian/CentOS 官方仓库文档 URL、手工安装命令和验证命令。

## 测试链与隔离

新增 `scripts/e2e-docker-install-test.sh`，实际执行源码 `deploy/install.sh`，包含要求的八个断言。临时源码无 `.env`，临时 APP_DIR/CONFIG_DIR 均在 `/tmp/yc-e2e-*`；随机可用端口、唯一 Compose project 和容器名防止碰撞。退出只删除该项目容器和 mktemp 目录。

`YC_DOCKER_ISOLATED=1` 仅接受 disposable 路径和 docker 模式，跳过宿主 systemd/nginx/CLI/firewall/Agent/doctor 集成，绝不触碰真实 `/opt/yuchen-panel`、`/etc/yuchen-panel` 或旧服务。这条测试验证源码容器安装全流程，不声称验证宿主 Nginx/systemd。正式安装默认仍走原完整路径。

CI 新增 `ubuntu-22.04` e2e job；前置条件缺失打印 SKIPPED，CI 必须看到 E2E_PASS，否则 job 失败，避免跳过冒充通过。完整安装日志上传 artifact（仅将 Password/API Token 行脱敏）。包名自动安装路径另外由模拟命令的回归测试覆盖；已预装 v2 的 CI runner 不会自动覆盖缺插件场景，Ubuntu 官方源的实际安装能力仍需该环境复验。

新增 `scripts/test-docker-install-preflight.sh`，检查真实 Compose 只读 config（若可用）、占位值子 shell 隔离、启动 env-file、包名候选回退、v1 拒绝和手工指引。mock 检查只证明分支逻辑，不等于容器运行。

集成测试还发现 Task 7 的 VERSION 文件解析测试需要 Docker builder 包含仓库 VERSION，已补 `COPY VERSION ./VERSION`；否则镜像构建会因缺文件失败。

## 本地执行记录

- Bash syntax / ShellCheck（新脚本和安装器，全部级别）：PASS。
- preflight 回归脚本：PASS；命令为模拟 apt/docker 的逻辑验证与真实 Compose 只读校验。
- backend `go test ./...`：PASS。测试代码来自当前源码，未使用 Release 二进制。
- 完整 e2e 安装：**SKIPPED（本地未执行）**。Task 8 限定完整安装由 CI runner 或显式授权的指定机器执行；本轮交由新增 CI job，不伪造成功输出。完整日志在 CI 实际运行后由 artifact 获取。
- C19/C21/C23/Agent reload 新代码真机验收：本轮未重新执行，不能用旧 fast Release 产物代替。

## 发布前确认

1. 安装器每次生成 64 字符 JWT_SECRET，写 `.env` 时排除旧 YC_JWT_SECRET；旧短密钥因此被替换，旧登录令牌失效。
2. VERSION 文件仍为发行标识 `1.0.2-stable`；Go/API/比较值为 `1.0.2`，panel.info 分别写 VERSION=1.0.2 和 CHANNEL=stable，这是明确拆分，不是要求两个字符串逐字一致。

未合并 main、未打 tag、未触发 Release。

## CI 实际执行结果（完成后补充）

[CI run 38518483171](https://github.com/sle58083/yuchen-panel/actions/runs/38518483171) 的六个 job 全部 success。源码 head 为 `7e0e3c9`，PR 合并检查产物为 `adf5a56`；runner 为 Ubuntu 22.04.5。完整 docker 安装从无 `.env` 的临时源码目录执行，八个断言全部 PASS，耗时约 54 秒。服务端 health 报告 1.0.2；JWT 容器环境与真实 `.env` 逐字节一致。

源码构建 API image：`sha256:ecdf88709f9b8ca2154a066e80ea98fa37e846a7a22a35a37bc0827e43dd9876`，没有使用 fast Release 二进制。八个检查点的实际输出提取见 `task8-e2e-result.txt`。

[下载完整安装日志 artifact](https://github.com/sle58083/yuchen-panel/actions/runs/38518483171/artifacts/11872802386)：含 `.audit/task8-e2e/docker-install.log`（全量、仅凭据行脱敏）及原始 `task8-e2e-result.txt`。Artifact ZIP SHA256：`e78999ec05b7c2693981074901b4421a75b155f8a8d2c9c47e67c237f14c1113`。本地 SKIPPED 状态只适用于本机完整安装；CI 此项现已实际通过。

限制仍然保留：CI runner 预装 Compose v2，因此包名安装循环已做分支逻辑回归，但不把此 e2e 结果宣称为“无插件官方源自动安装真机通过”。宿主 Nginx/systemd/Agent 集成也不属于隔离 e2e 覆盖范围。
