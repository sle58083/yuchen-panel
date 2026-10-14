# Task 10 — 发布前 Nginx 与用户提示收尾

## BUG-M 定位与修复

旧 `write_host_nginx_docker` 模板只有 `location = /` 和 `location /<WEB_BASE_PATH>/`。`/api/health` 不匹配两者，进入宿主 Nginx 默认静态文件处理并产生 404；并非 API 端口不通，也不是 frontend build 目录缺失。测试将用原模板路由结构在独立 Nginx 容器复现 404，并输出 error/access log 中的静态文件路径及 `upstream=-`，不把协调者实机日志伪造成本机证据。

修复宿主模板的根 `/api/`、`/sub/`、`/s/` location，保持 URI 通过 loopback WEB_PORT 转发给 frontend 容器。宿主 Nginx 不在 Docker 网络中，不能直接解析 `yuchen-panel-api`；因此保留正确的两层架构：host Nginx → 127.0.0.1:WEB_PORT → frontend 容器 Nginx → Docker DNS `yuchen-panel-api:8088`。

同时修复 frontend Nginx 根 API/sub/share location 的优先级（加 `^~`），避免被宽泛 Vue history regex 覆盖。随机前缀 API 路由使用 rewrite + 静态 upstream，避免 URI 变量触发没有 resolver 的运行时 DNS 解析。

Docker/fast 模板拆为纯 render 与原写入启动函数，方便隔离测试；fast heredoc 模板内容逐字一致，仅抽出渲染函数，没有修改 fast 路由语义。

`test-nginx-proxy.sh` 挂载实际生成模板，在 host-network 的独立 Nginx 容器测试：旧模板 404；新 Docker 根/前缀 API health 200 JSON、前端页面 200、无效 sub/share token 到达 API 404而非 SPA HTML；fast 根/前缀 health 200。它只在 `/tmp/yc-e2e-*` 中运行，不修改宿主 /etc/nginx，不启动/重启宿主 nginx/systemd。fast 检查验证模板路由，API upstream 由本次源码构建容器提供，不冒充真实 systemd 安装验证。

## BUG-N 密钥轮换提示

安装器在复制/清理运行目录前检测 `APP_DIR/.env` 是否存在，记录本次安装是否轮换旧密钥。成功输出 `print_result` 仅在升级/重装时打印：

```text
⚠️ 安全提示：本次安装已轮换 JWT 签名密钥。
   所有已登录用户需要重新登录，管理后台的旧会话将失效。
```

全新安装不显示此提示；生成 64 字符密钥的既有逻辑不变。CHANGELOG 的 Unreleased 段记录行为。回归测试调用实际 `print_result`，验证新装与重装两种输出；e2e 也断言新装完整日志没有轮换提示。

## Doctor SKIP 与真正失败

Docker 模式 API 的 systemd 检查显示 SKIP，但额外验证 Compose API 容器确实 running。未部署 Agent（无 agent.env）时，Agent/Xray/config/restart-watch 检查显示 SKIP，并注明 Docker 模式不含 Agent、节点连通需单独部署。已经配置 Agent 时仍按真实服务状态检查，不吞掉错误。

API health、面板反代、Nginx 语法、数据文件失败仍显示 FAIL。API health 从运行 .env 读取实际 API_PORT，不再硬编码 8088。末尾单独统计 failure/warning/skipped，SKIP 不计入失败。CLI 增加 APP_DIR/CONFIG_DIR 环境覆盖仅用于独立测试，正式默认路径未变。

mock 回归验证 Docker 无 Agent 时 PASS（5 skipped）、真正停止的容器/失败的反代仍 FAIL、fast 模式仍检查全部 systemd 服务，没有 SKIP。

## 验证与边界

- 原 fast Nginx heredoc body 字节对比：PASS。
- Bash/ShellCheck、JWT notice/Doctor mock 回归、Task 8 preflight 回归：PASS。
- backend 源码 go test：PASS。
- C24 未修改；七条用例和前端测试由 CI 重跑。
- Docker 安装八项断言保留，追加真实 Nginx 路由验证，交由 Ubuntu 22.04 CI 执行。完成后补实际日志；本机没有执行完整安装，没有伪造 curl 200 证据。

Task 8 的 BUG-K/BUG-L/C22 预检、包名候选与提前退出逻辑未改动。正式 fast Release 仍是旧产物，不能用于验收本轮源码修复。

## 告知用户

- 升级/重装后密钥轮换，旧会话失效，需要重新登录。
- serverSpecAdvice 的容量提示依赖新版 Agent；旧 Agent 不上报时提示不出现。

未合并 main、未打 tag、未触发 Release。
