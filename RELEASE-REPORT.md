# v1.0.3-rc.1 发布报告（执行中）

本轮只授权 RC；没有授权 v1.0.3 正式 tag。以下记录真实执行，尚未执行的步骤不会标成功。

## 0. 发布前检查与获准修复

实际命令：
```powershell
Get-Content .github/workflows/release.yml
Get-Content VERSION
Get-Content version.json
Get-Content backend/internal/version/version.go
Invoke-RestMethod https://api.github.com/repos/sle58083/yuchen-panel/releases?per_page=5
git ls-remote origin refs/heads/main refs/heads/fix/audit-batch4-test-harness refs/tags/v1.0.3-rc.1
```

实际结果：main=`665df7866ccbb404f77832c3657498ac164b9bde`，修复分支=`5ac73c9d906314d99bc2257c673283718d2ee1d3`，RC tag 未存在；无本地 releases 目录。

既有 Release：v1.0.2（2026-09-16T09:57:04Z），资产 ID `567675591`，文件 `yuchen-panel-v1.0.2-stable.zip`，大小 `5533673` bytes；原 version.json 的 download_url 指向该资产，sha256=`7b64a19b908f7485b26858ad80340f2af6333605e2223fadb85c497d195602ea`。另有 v1.0.1 和 v1.0.0 Release。均未修改/删除。

原 workflow 同时支持 push tags v* 和 workflow_dispatch，声明 contents:write；原逻辑从 tag 得 RC 文件名，却仅更新旧 manifest 的 SHA256/时间，旧下载 URL 会与新 digest 不一致，且 API/Agent/安装器仍1.0.2。

按任务要求先询问，用户明确回复：**“先修正 RC 版本与流程，再继续发布”**。因此准备补丁统一运行版本为1.0.3-rc.1，并修复为：CI 编译→创建全新 RC 资产（不允许 clobber）→下载核验→生成全字段 manifest→fast-forward 回填当前 main。源 version.json 此时保留原1.0.2，绝不手写新 sha。

本机预检命令：
```powershell
$env:GIT_TERMINAL_PROMPT='0'
git -c credential.interactive=false push --dry-run origin HEAD:refs/heads/fix/audit-batch10-nginx
```
实际输出：`fatal: unable to get password from user`，没有任何远端变更。用户要求调起登录窗口，已用 winget 安装 GitHub 官方 CLI 2.102.0：`已成功验证安装程序哈希`、`已成功安装`。以正常执行策略打开 gh auth login 可见窗口；登录待用户完成。此前 ExecutionPolicy Bypass 组合启动命令被自动策略拒绝，改用包管理器与直接登录命令，没有绕过执行策略。

版本/元数据修复本地测试：Go race测试、Python RC manifest一致性两项测试、前端11项测试与RC构建、Shell及旧安装前置回归通过。后续补原始日志及真实CI。

## Step 1–7

尚未开始：等待本机 GitHub 写入认证就绪，随后验证发布补丁CI并按序执行非squash合并、main全绿、带注解RC tag、Release构建、资产校验、main manifest读取、PR merged确认。没有合并main或打tag。

## 升级到 v1.0.3-rc.1（仅发布校验成功后执行）

```bash
bash <(curl -Ls https://raw.githubusercontent.com/sle58083/yuchen-panel/main/install.sh)
```

升级会轮换 JWT 签名密钥；之前登录的所有用户需要重新登录。
新建节点默认抗封锁档（VLESS+Reality），也可选择极速/均衡档。
CPU/内存建议依赖新版 Agent 上报容量；旧 Agent 不上报时提示不出现，重新安装匹配新版 Agent 即可。
