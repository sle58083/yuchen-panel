# 在一次性 Ubuntu/Debian 测试机部署 Task 4

只在可丢弃的测试机执行。下列分支包含 [PR #4](https://github.com/sle58083/yuchen-panel/pull/4) 的修复和 Task 4 验证脚本；**不要使用 `main`**。建议 Ubuntu 22.04 或 Debian 12，root/sudo、Git、curl 和公网出站可用。安装会配置系统服务、Nginx、Docker/Xray，并占用测试端口；先确认测试机没有业务数据。

## 0. 确认机器和分支

```bash
cat /etc/os-release
git clone https://github.com/sle58083/yuchen-panel.git ~/yuchen-panel-task4
cd ~/yuchen-panel-task4
git fetch origin fix/audit-batch4-test-harness
git switch --detach FETCH_HEAD
git log -1 --oneline
```

若要单独核查父分支，可用 `git fetch origin pull/4/head:pr4 && git switch pr4`。实际安装和执行 C15/C18 时应返回 `fix/audit-batch4-test-harness` 的提交。全新安装可用；从旧版升级时，先备份并检查 `/opt/yuchen-panel/.env` 中 `YC_JWT_SECRET` 是否至少 32 个字符且不是公开占位值。短密钥会使新版 API 启动时 panic；更换密钥会使已有登录令牌失效。不要把密钥粘贴到报告。

## 1. 从本分支源码构建并安装

```bash
cd ~/yuchen-panel-task4
sudo env YC_INSTALL_MODE=docker bash deploy/install.sh
sudo yuchen-panel status
curl -fsS http://127.0.0.1:8088/api/health
```

`deploy/install.sh` 的 Docker 模式从当前检出的源码构建 API 和前端。不要运行 `bash <(curl -Ls .../install.sh)`：仓库根目录安装器会读取该分支的 `version.json`，其 `download_url` 仍指向旧版 GitHub Release ZIP，因此即使脚本 URL 是 PR 分支，也**不会**安装 PR 代码。也不要把此测试分支当作稳定版发布。

Docker 模式明确跳过宿主机 Agent 安装。Agent 需要单独提供与测试机架构匹配的 `bin/yuchen-agent-linux-amd64` 后执行 `deploy/agent-install.sh`；在 Agent 部署前，T2/T4/T5 的 B 档节点连通测试不可用。安装输出会明确打印这条提示，不能把 API/前端容器启动当作代理节点已就绪。

保存安装输出中的 `ACCESS_URL`、面板管理员账号与密码。`panel.info` 的 `API_TOKEN` 是 **Agent 共享密钥**，不是下文 `YC_TOKEN` 所需的管理员登录 JWT。不要在回贴日志中包含两种令牌或密码。

面板默认 HTTP 明文。关闭安全组里的面板端口公网入口，在操作电脑上建立 SSH 隧道访问 UI：

```bash
ssh -L 2053:127.0.0.1:2053 root@测试机公网IP
```

将两处 `2053` 换成安装输出中的实际面板端口；浏览器打开 `http://127.0.0.1:实际端口/实际WEB_BASE_PATH/`。服务器本地预置脚本走 `http://127.0.0.1:8088`，不需要开放 API 端口。真实节点连接测试才需要按计划开放 TCP 24401–24406，并核对云安全组、本机防火墙、端口占用和公网路由。

## 2. 取得管理员 JWT 并预置测试资源

在测试机进入同一份源码目录。先用面板管理员凭据登录；下列命令不把密码写进命令行参数或 shell 历史：

```bash
cd ~/yuchen-panel-task4
read -r -p '管理员用户名: ' YC_ADMIN_USER
read -r -s -p '管理员密码: ' YC_ADMIN_PASS; echo
YC_TOKEN=$(jq -n --arg username "$YC_ADMIN_USER" --arg password "$YC_ADMIN_PASS" \
  '{username:$username,password:$password}' |
  curl -fsS -H 'Content-Type: application/json' --data-binary @- \
    http://127.0.0.1:8088/api/auth/login | jq -er .token)
unset YC_ADMIN_PASS
export YC_ENV=test YC_TOKEN
```

全新安装也会生成非空 `/opt/yuchen-panel/data/yuchen-panel.json`。确认它**只属于这台一次性测试机**后，在 dry-run 与 apply 都显式传 `--allow-existing`：

```bash
bash scripts/prepare-test-env.sh --dry-run --allow-existing
# 核对输出的路径、面板 URL、IPv6 探测结果和计划端口。
bash scripts/prepare-test-env.sh --apply --allow-existing
```

默认 dry-run 不调用 API，也不创建输出文件。`--apply` 会根据测试机 `/api/servers` 的 host 生成 6 个节点；若 host 不是真实可达的 IPv4，可在运行前设置 `TEST_NODE_HOST=测试机真实IPv4`。多服务器面板需要显式 `TEST_SERVER_ID=目标服务器ID`。不要填 `192.0.2.0/24` 等文档地址。脚本检测到 `::/0`、全局 `2000::/3` 地址且本机 bind 成功时，将 Trojan 节点 host 设为真实 IPv6；否则用 IPv4 并打印 T2 IPv6 连通项跳过原因。

每次 apply 生成新的私有 `client-test-env/samples.*` 目录，`client-test-env/TEST-ENV.json` 记录节点/客户 ID、IP、IPv6 结果及样例 SHA256，不记录密码或订阅 token。重复执行会按固定用户名/节点名复用；若同名节点参数不同，脚本会停止，供人工核查。样例文件本身仍含凭据，勿公开分享。用完后 `unset YC_TOKEN`。

## 3. 实测前的已知服务端限制

该分支的 Xray 配置生成器为 TLS 入站只写 `serverName`，尚未配置证书；同时普通入站监听 `0.0.0.0`，不一定接受 IPv6 连接。因此 C15 的 TLS/IPv6 节点用于**真实字段和部署失败定位**，并不能保证 T2/T4 的 B 档连通通过。先在测试机检查 `yuchen-panel doctor`、`journalctl -u yuchen-agent -n 100 --no-pager`、`systemctl status xray` 与实际监听端口。若服务端 TLS/IPv6 入站未就绪，把相应 B 项记为“阻塞/未通过”并附脱敏日志；不能把 A 档导入通过记为连通通过。

之后按 [TEST-CHECKLIST.md](TEST-CHECKLIST.md) 在用户自己的设备完成 T2–T6。12+ 节点的 T3 A 档仍使用已有的合成样例；C15 的六个真实节点供 T2/T4/T5 使用。
