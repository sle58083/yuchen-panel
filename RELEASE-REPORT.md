# v1.0.3-rc.1 发布报告

2026-10-14（Asia/Shanghai）。用户只授权 RC，未创建 v1.0.3 正式版。
Step 1–7 已完成；以下是实际命令与输出。安装脚本/manifest 已验证可获取及字段正确，尚未在用户 VPS 执行新 RC 安装或实机验收。

## 0. 前置、授权与失败证据

用户授权先修正 RC 版本及发布流程；准备提交 5050059d5dd8f45769c23b3adc3bfc34fd257f99 的两次 CI 均 success（38545571083、38545562983）。
客户主要在墙内、需抗封锁，由用户明确答复确定，所以 C24 默认 Reality。
GitHub CLI 登录账户 sle58083，auth setup-git 完成。仓库 Actions 默认权限为 read，release.yml 显式声明 contents:write；后续实际发布与 main 回填成功验证其有效性。

两次失败分别为缺少远端 tracking ref、未跟踪报告阻挡 merge，均在失败时停止，未推送。
本轮先将仓库外完整报告复制至 .codex-release-inspect/RELEASE-REPORT-attempt2.md，并对两个文件执行 Get-FileHash -Algorithm SHA256：两者均为 FDF7C930BF4EF9278E2197C5A084D6A47ED3463784A6F60007FE059F0C4BF531。
没有删除证据；git check-ignore 输出 .codex-release-inspect/RELEASE-REPORT-attempt2.md；git status --short 为空。

## Step 1 · 非 squash 合并

实际命令（每条关键命令失败即停止）：
```bash
git checkout main
git pull --ff-only origin main
git merge --no-ff origin/fix/audit-batch4-test-harness -F .git/rc-merge-message.txt
git show -s --format='commit %H%nparents %P%n%B' HEAD
git push origin main
```
实际输出摘录（省略大量文件统计，保留完整 merge 元数据）：
```text
Already on 'main'
Your branch is up to date with 'origin/main'.
Already up to date.
Merge made by the 'ort' strategy.
To https://github.com/sle58083/yuchen-panel.git
   665df78..da927c6  main -> main
```

合并提交的双父节点证据

实际命令：
```text
git show -s --format='commit %H%nparents %P%n%B' da927c67
```

实际输出：
```text
commit da927c67c43afe9e3fc43e7479fad2c261e9c9d6
parents 665df7866ccbb404f77832c3657498ac164b9bde 5050059d5dd8f45769c23b3adc3bfc34fd257f99
merge: audit round 1 security and client compatibility fixes

P0: reject weak JWT secrets, forbid default credentials, require active admin
P1: escape upgrade runner literals, restrict manifest redirects, isolate extraction
P1: unify version source, replace substring version match with semantic compare
C24: node performance presets (anti-blocking default per user policy)
BUG-M/N: repair docker nginx proxy routes, notify JWT rotation
C22: reject compose v1 before touching existing services
C25: add isolated docker installer end-to-end test

Coordinated by WorkBuddy. Verified on Ubuntu 22.04 (38.34.21.9).
```


## Step 2 · main CI 全绿

实际命令：
```text
gh run list --repo sle58083/yuchen-panel --branch main --limit 3
```

实际输出：
```text
completed	success	merge: audit round 1 security and client compatibility fixes	CI	main	push	38548087432	1m8s	2026-10-14T10:13:47Z
```


六个 job 的结论

实际命令：
```text
gh run view 38548087432 --repo sle58083/yuchen-panel --json status,conclusion,jobs --jq '{status,conclusion,jobs:[.jobs[]|{name,status,conclusion}]}'
```

实际输出：
```text
{"conclusion":"success","jobs":[{"conclusion":"success","name":"Fresh Docker installer integration (Ubuntu 22.04)","status":"completed"},{"conclusion":"success","name":"Frontend build and typecheck","status":"completed"},{"conclusion":"success","name":"Shell syntax and ShellCheck","status":"completed"},{"conclusion":"success","name":"Go build, vet and race tests (agent)","status":"completed"},{"conclusion":"success","name":"Go build, vet and race tests (backend)","status":"completed"},{"conclusion":"success","name":"Docker Compose builds","status":"completed"}],"status":"completed"}
```

实际 gh run watch 38548087432 --repo sle58083/yuchen-panel --exit-status 退出 0；完整 watch 输出留在 .codex-release-inspect/main-ci-watch.txt。

## Step 3 · 仅 RC 注解 tag

实际命令：
```text
git tag -a v1.0.3-rc.1 -F .git/rc-tag-message.txt; git cat-file -t v1.0.3-rc.1; git cat-file -p v1.0.3-rc.1
```

实际输出：
```text
object da927c67c43afe9e3fc43e7479fad2c261e9c9d6
type commit
tag v1.0.3-rc.1
tagger sle58083 <281434516+sle58083@users.noreply.github.com> 1791972926 +0800

v1.0.3-rc.1

审计修复 RC。相对 v1.0.2：安全（弱密钥拒绝、默认凭据禁用、升级转义、
清单重定向限制、解压隔离）、修复（版本单一来源、语义比较替代子串匹配）、
客户端全格式兼容、三档节点预设（默认抗封锁）、安装流程修复
（拒绝 compose v1、nginx 路由、密钥轮换提示）、新增隔离式 e2e 安装测试。

升级会轮换 JWT 密钥，所有用户需重新登录。
```

cat-file -t 实际输出 tag；git push origin v1.0.3-rc.1 实际输出：
```text
To https://github.com/sle58083/yuchen-panel.git
 * [new tag]         v1.0.3-rc.1 -> v1.0.3-rc.1
```
远端 ls-remote：
```text
aa0716f65e71ea505097cd577eb5ba26346927fe refs/tags/v1.0.3-rc.1
da927c67c43afe9e3fc43e7479fad2c261e9c9d6 refs/tags/v1.0.3-rc.1^{}
```

## Step 4 · 自动触发 CI Release

实际命令：
```text
gh run list --repo sle58083/yuchen-panel --workflow release.yml --limit 3
```

实际输出：
```text
completed	success	merge: audit round 1 security and client compatibility fixes	Build Release	v1.0.3-rc.1	push	38548305975	48s	2026-10-14T10:15:37Z
completed	success	fix: 修复 VMess/Shadowsocks 入站测试超时(-1)	Build Release	v1.0.2	push	35082203560	55s	2026-09-16T09:56:14Z
completed	success	feat: 客户管理支持多协议（VMess/Trojan/Shadowsocks）绑定与订阅导出	Build Release	v1.0.1	push	35001552993	1m6s	2026-09-15T17:29:05Z
```


Release job 和每个步骤的真实结论

实际命令：
```text
gh run view 38548305975 --repo sle58083/yuchen-panel --json status,conclusion,jobs（投影步骤状态）
```

实际输出：
```text
{"conclusion":"success","jobs":[{"conclusion":"success","name":"build","status":"completed","steps":[{"conclusion":"success","name":"Set up job","status":"completed"},{"conclusion":"success","name":"Run actions/checkout@v4","status":"completed"},{"conclusion":"success","name":"Run actions/setup-go@v5","status":"completed"},{"conclusion":"success","name":"Run actions/setup-node@v4","status":"completed"},{"conclusion":"success","name":"Resolve and validate annotated release version","status":"completed"},{"conclusion":"success","name":"Build backend API from tag source","status":"completed"},{"conclusion":"success","name":"Build matching Agent from tag source","status":"completed"},{"conclusion":"success","name":"Build frontend","status":"completed"},{"conclusion":"success","name":"Package release zip","status":"completed"},{"conclusion":"success","name":"Create CI-built release assets (no replacement)","status":"completed"},{"conclusion":"success","name":"Verify published archive and generate full manifest","status":"completed"},{"conclusion":"success","name":"Publish verified metadata on current main (fast-forward only)","status":"completed"},{"conclusion":"success","name":"Post Run actions/setup-node@v4","status":"completed"},{"conclusion":"success","name":"Post Run actions/setup-go@v5","status":"completed"},{"conclusion":"success","name":"Post Run actions/checkout@v4","status":"completed"},{"conclusion":"success","name":"Complete job","status":"completed"}]}],"status":"completed"}
```

gh run watch 38548305975 --repo sle58083/yuchen-panel --exit-status 退出 0；完整 workflow 日志保存在 .codex-release-inspect/release-full-log.txt。由 CI 构建和上传 ZIP/SHA256SUMS，没有手动上传。

Release 页面与资产

实际命令：
```text
gh release view v1.0.3-rc.1 --repo sle58083/yuchen-panel --json tagName,isPrerelease,url,publishedAt,assets
```

实际输出：
```text
{"assets":[{"apiUrl":"https://api.github.com/repos/sle58083/yuchen-panel/releases/assets/640000957","contentType":"application/octet-stream","createdAt":"2026-10-14T10:16:17Z","digest":"sha256:c6137b9a6a580c31c2c20f95313e565013d1bb009776589cfc5936bdb8c66802","downloadCount":0,"id":"RA_kwDOUcPjjc4mJaO9","label":"","name":"SHA256SUMS","size":95,"state":"uploaded","updatedAt":"2026-10-14T10:16:17Z","url":"https://github.com/sle58083/yuchen-panel/releases/download/v1.0.3-rc.1/SHA256SUMS"},{"apiUrl":"https://api.github.com/repos/sle58083/yuchen-panel/releases/assets/640000960","contentType":"application/zip","createdAt":"2026-10-14T10:16:17Z","digest":"sha256:f4b97180e4b7102757974fd773a1beb5067c9c896c734a219084394543af8530","downloadCount":0,"id":"RA_kwDOUcPjjc4mJaPA","label":"","name":"yuchen-panel-v1.0.3-rc.1.zip","size":5655759,"state":"uploaded","updatedAt":"2026-10-14T10:16:18Z","url":"https://github.com/sle58083/yuchen-panel/releases/download/v1.0.3-rc.1/yuchen-panel-v1.0.3-rc.1.zip"}],"isPrerelease":true,"publishedAt":"2026-10-14T10:16:18Z","tagName":"v1.0.3-rc.1","url":"https://github.com/sle58083/yuchen-panel/releases/tag/v1.0.3-rc.1"}
```


## Step 5 · 下载包 SHA256 与 manifest 一致

实际命令：
```text
gh release download v1.0.3-rc.1 --repo sle58083/yuchen-panel --dir .codex-release-inspect/rc-verify --pattern 'yuchen-panel-v1.0.3-rc.1.zip'
wsl.exe -d Ubuntu -- sha256sum '/mnt/e/gpt保存/2026-10-13/linux-xray-sing-box-v2ray-shell/work/audit-fix/.codex-release-inspect/rc-verify/yuchen-panel-v1.0.3-rc.1.zip'
```

实际输出：
```text
f4b97180e4b7102757974fd773a1beb5067c9c896c734a219084394543af8530  /mnt/e/gpt保存/2026-10-13/linux-xray-sing-box-v2ray-shell/work/audit-fix/.codex-release-inspect/rc-verify/yuchen-panel-v1.0.3-rc.1.zip
```


main/version.json 原始内容

实际命令：
```text
(Invoke-WebRequest -UseBasicParsing 'https://raw.githubusercontent.com/sle58083/yuchen-panel/main/version.json').Content
```

实际输出：
```text
{
  "latest": "1.0.3-rc.1",
  "version": "1.0.3-rc.1",
  "codename": "rc",
  "package": "yuchen-panel-v1.0.3-rc.1.zip",
  "download_url": "https://github.com/sle58083/yuchen-panel/releases/download/v1.0.3-rc.1/yuchen-panel-v1.0.3-rc.1.zip",
  "sha256": "f4b97180e4b7102757974fd773a1beb5067c9c896c734a219084394543af8530",
  "min_supported_version": "0.7.5",
  "released_at": "2026-10-14T10:16:19.811084Z",
  "changelog": [
    "审计修复 RC：认证、升级与安装路径安全修复",
    "sing-box/v2rayN/Clash/Shadowrocket 客户端兼容",
    "三档节点预设，默认抗封锁档（用户墙内客户策略）",
    "Docker Nginx 路由及 Doctor 模式检查修复",
    "升级会轮换 JWT 密钥，旧会话失效，需要重新登录",
    "服务器规格提示依赖新版 Agent"
  ]
}
```


逐字符及全字段比较结果

实际命令：
```text
Get-FileHash -Algorithm SHA256；PowerShell -cne 比较实测 digest 与 .sha256，以及 latest/version/codename/package/download_url
```

实际输出：
```text
Downloaded ZIP SHA256: f4b97180e4b7102757974fd773a1beb5067c9c896c734a219084394543af8530
SHA256 exact match; all RC manifest fields match.
```

清单由 workflow 校验下载 ZIP 后生成，bot 回填提交为 88563450b0bdf30974b705652375a3c451e4c540；没有手工编辑 version.json 或 SHA。

## Step 6 · 一行安装可解析

实际命令：
```text
Invoke-WebRequest -UseBasicParsing 'https://raw.githubusercontent.com/sle58083/yuchen-panel/main/install.sh'（HTTP 状态、前五行）
```

实际输出：
```text
HTTP 200
#!/usr/bin/env bash
set -Eeuo pipefail

export DEBIAN_FRONTEND="${DEBIAN_FRONTEND:-noninteractive}"
export NEEDRESTART_MODE="${NEEDRESTART_MODE:-a}"
```

version.json 同样读取成功，version=1.0.3-rc.1（见 Step 5 原始 JSON）。这是下载与清单解析验证，不宣称已完成用户 VPS 安装。

## Step 7 · PR #5 已合并

实际命令：
```text
gh pr view 5 --repo sle58083/yuchen-panel --json state,mergedAt,mergeCommit
```

实际输出：
```text
{"mergeCommit":{"oid":"da927c67c43afe9e3fc43e7479fad2c261e9c9d6"},"mergedAt":"2026-10-14T10:13:46Z","state":"MERGED"}
```


## 既有 Release 资产完整性

实际命令：
```text
gh api repos/sle58083/yuchen-panel/releases --paginate；发布前后比较 release id/updated_at，以及 asset id/name/size/digest/created_at/updated_at/browser_download_url
```

实际输出：
```text
v1.0.2 unchanged; release id=389806375; assets={"id":567675591,"name":"yuchen-panel-v1.0.2-stable.zip","size":5533673,"digest":"sha256:7b64a19b908f7485b26858ad80340f2af6333605e2223fadb85c497d195602ea","created_at":"2026-09-16T09:57:03Z","updated_at":"2026-09-16T09:57:03Z","browser_download_url":"https://github.com/sle58083/yuchen-panel/releases/download/v1.0.2/yuchen-panel-v1.0.2-stable.zip"}
v1.0.1 unchanged; release id=389343392; assets={"id":566170691,"name":"yuchen-panel-v1.0.1-stable.zip","size":5532482,"digest":"sha256:230da902bdf030d072c5064f476b9a8b3e5069ecc7b7899a0c69ac45eadcb3f8","created_at":"2026-09-15T17:30:02Z","updated_at":"2026-09-15T17:30:02Z","browser_download_url":"https://github.com/sle58083/yuchen-panel/releases/download/v1.0.1/yuchen-panel-v1.0.1-stable.zip"}
v1.0.0 unchanged; release id=389313795; assets={"id":566095185,"name":"yuchen-panel-v1.0.0-stable.zip","size":5515117,"digest":"sha256:7104ef9fd6e30332d991fa165284102474fe2c77f2d9d378520bc31b7bb3d976","created_at":"2026-09-15T16:43:48Z","updated_at":"2026-09-15T16:43:49Z","browser_download_url":"https://github.com/sle58083/yuchen-panel/releases/download/v1.0.0/yuchen-panel-v1.0.0-stable.zip"}
```

## 升级到 v1.0.3-rc.1

```bash
bash <(curl -Ls https://raw.githubusercontent.com/sle58083/yuchen-panel/main/install.sh)
```

⚠️ 升级需要重新登录：本次安装/升级轮换 JWT 签名密钥，旧登录会话失效。
新建节点默认「抗封锁档」（VLESS + Reality），也可选极速/均衡。
CPU/内存建议依赖新版 Agent 上报；旧 Agent 不上报时提示不出现，需要安装匹配的新版 Agent。

发布页面：https://github.com/sle58083/yuchen-panel/releases/tag/v1.0.3-rc.1
main CI：https://github.com/sle58083/yuchen-panel/actions/runs/38548087432
Release CI：https://github.com/sle58083/yuchen-panel/actions/runs/38548305975

用户完成 RC 实机安装验证后，再另行授权正式发布。本轮无正式版 tag、force push、rebase、squash、clobber 或既有资产修改。
本报告在 RC 资产校验完成后以文档提交保存；RC tag 始终指向 da927c67，报告提交不改变已发布 ZIP。