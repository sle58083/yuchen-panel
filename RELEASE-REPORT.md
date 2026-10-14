# v1.0.3-rc.3 发布报告

2026-10-15 Asia/Shanghai（下列 GitHub 原始时间带 Z，使用 UTC）。仅授权 RC；没有创建 v1.0.3 正式版。

发布已完成：PR #8 非 squash 合并、两轮 main CI 全绿、注解 tag、CI Release、ZIP/SHA256SUMS/main 清单三者一致。VPS 客户实际流量与停用验收仍由用户执行，本轮没有把静态/CI 结果伪装为真机验收。

## 0. 发布前核查与授权例外

远端 main=8ace907f77e22d7a05bbb9e7f85a65439e7bb858；Task13=6f3d0ac41a15ec1fd46d2150e05d59f35ea5c08f。
实际 git log origin/main..origin/fix/task13-monthly-traffic 仅 f96f56a、6f3d0ac；反向范围为空。version.json diff为空。PR #8 状态 OPEN、MERGEABLE，12 项 checks 均 SUCCESS（两个六 job run）。rc.3 tag 不存在。

发现源码仍为 rc.2，release.yml 强制 tag 版本与 VERSION、frontend/package.json 一致，直接打 rc.3 会失败。先提供具体准备清单 .codex-release-inspect/rc3/RC3-PREPARATION.md，并询问范围例外。
用户明确答复：**“允许单独的发布准备提交，继续 rc.3”**。
因此先保存 Task13 两条＋merge 的干净历史证据，再独立准备五个版本文件及 RELEASE-NOTES.md、manifest 生成器仅 changelog 文案；增加一轮 main CI，不改变业务逻辑或手工改 version.json。

## Step 1 · 使用远端恢复本地分支

实际命令（关键 Git 命令均检查退出码）：
```text
git fetch origin refs/heads/main:refs/remotes/origin/main refs/heads/fix/task13-monthly-traffic:refs/remotes/origin/fix/task13-monthly-traffic
git config --add remote.origin.fetch '+refs/heads/fix/task13-monthly-traffic:refs/remotes/origin/fix/task13-monthly-traffic'
git fetch origin
git checkout -B fix/task13-monthly-traffic origin/fix/task13-monthly-traffic
git log --oneline -1
git branch -vv
```
本地同名分支原本就等于远端 6f3d0ac、工作区干净，因此没有陈旧状态需备份；仍按指示由远端重建，未复用未核对的旧状态。已有 backup/stale-audit-7bd7628 保留未删。
实际输出：
```text
Reset branch 'fix/task13-monthly-traffic'
branch 'fix/task13-monthly-traffic' set up to track 'origin/fix/task13-monthly-traffic'.
Your branch is up to date with 'origin/fix/task13-monthly-traffic'.
6f3d0ac fix(billing): refresh usage details and document quota validation evidence
```
Git 的 Reset branch 文字来自 checkout -B；未执行 reset --hard/rebase/force push。

本地分支快照

实际命令：
```text
git branch -vv
```

原始输出：
```text
backup/stale-audit-7bd7628    7bd7628 docs(test): record Task 9 green CI and installer regression results
  fix/audit-batch1-security     b3cb4a1 fix(access)!: enforce explicit grants and retire Shadowsocks
  fix/audit-batch10-nginx       5fc4312 docs(test): record Task 10 real nginx and green CI evidence
  fix/audit-batch3-client-uri   7854bf1 fix(clients): close URI encoding and validation gaps
  fix/audit-batch4-test-harness ceb95f4 [origin/fix/audit-batch4-test-harness] chore(release): prepare unified v1.0.3-rc.2 version
  fix/task11-clean-node-names   47c193e docs(task11): record clean-name exports and actual regression evidence
* fix/task13-monthly-traffic    6f3d0ac [origin/fix/task13-monthly-traffic] fix(billing): refresh usage details and document quota validation evidence
  main                          8ace907 [origin/main] docs(release): record verified rc.2 publication and preserved assets
  release/prepare-rc1           5738677 fix(release): preserve RC version and publish verified complete metadata
```

## Step 2 · 合并 PR #8，保留分支

实际命令：
```text
gh pr merge 8 --repo sle58083/yuchen-panel --merge --delete-branch=false --match-head-commit 6f3d0ac41a15ec1fd46d2150e05d59f35ea5c08f
git fetch origin main
git log --oneline 8ace907..origin/main
```
merge 命令退出 0；无冲突。没有 squash。

合并未夹带的关键范围证据（版本准备之前）

实际命令：
```text
git log --oneline 8ace907..origin/main
```

原始输出：
```text
29a3c6e Merge pull request #8 from sle58083/fix/task13-monthly-traffic
6f3d0ac fix(billing): refresh usage details and document quota validation evidence
f96f56a feat(billing): meter per-client traffic and enforce renewable monthly quotas
```


PR #8 合并状态

实际命令：
```text
gh pr view 8 --repo sle58083/yuchen-panel --json state,mergedAt,mergeCommit
```

原始输出：
```text
{"mergeCommit":{"oid":"29a3c6e5bb3e187f173755a4e062533cbab21a05"},"mergedAt":"2026-10-14T21:17:02Z","state":"MERGED"}
```


合并提交双父证据

实际命令：
```text
git show -s --format='commit %H%nparents %P%n%B' 29a3c6e
```

原始输出：
```text
commit 29a3c6e5bb3e187f173755a4e062533cbab21a05
parents 8ace907f77e22d7a05bbb9e7f85a65439e7bb858 6f3d0ac41a15ec1fd46d2150e05d59f35ea5c08f
Merge pull request #8 from sle58083/fix/task13-monthly-traffic

feat: renewable monthly traffic quotas with Xray per-client stats
```


## Step 3 · 合并后的第一轮 main CI

实际命令：
```text
gh run list --repo sle58083/yuchen-panel --branch main --limit 3
```

原始输出：
```text
completed	success	Merge pull request #8 from sle58083/fix/task13-monthly-traffic	CI	main	push	38639041909	1m12s	2026-10-14T21:17:05Z
completed	success	docs(release): record verified rc.2 publication and preserved assets	CI	main	push	38601161818	1m23s	2026-10-14T16:48:10Z
completed	success	Merge pull request #7 from sle58083/fix/audit-batch4-test-harness	CI	main	push	38600379832	1m8s	2026-10-14T16:42:38Z
```


第一轮六 job 全绿

实际命令：
```text
gh run view 38639041909 --repo sle58083/yuchen-panel --json status,conclusion,jobs --jq '{status,conclusion,jobs:[.jobs[]|{name,status,conclusion}]}'
```

原始输出：
```text
{"conclusion":"success","jobs":[{"conclusion":"success","name":"Shell syntax and ShellCheck","status":"completed"},{"conclusion":"success","name":"Go build, vet and race tests (backend)","status":"completed"},{"conclusion":"success","name":"Docker Compose builds","status":"completed"},{"conclusion":"success","name":"Go build, vet and race tests (agent)","status":"completed"},{"conclusion":"success","name":"Frontend build and typecheck","status":"completed"},{"conclusion":"success","name":"Fresh Docker installer integration (Ubuntu 22.04)","status":"completed"}],"status":"completed"}
```

gh run watch 38639041909 --repo sle58083/yuchen-panel --exit-status 退出 0。

已获准的独立发布准备提交

实际命令：
```text
git show --stat --format=oneline c41bcc2
```

原始输出：
```text
c41bcc2b84fa6a14209baf8f3cf7b2b27de45277 chore(release): prepare rc.3 version and monthly quota release notes
 RELEASE-NOTES.md                    | 39 +++++++++++++++++++++++++------------
 VERSION                             |  2 +-
 backend/internal/version/version.go |  2 +-
 frontend/package-lock.json          |  4 ++--
 frontend/package.json               |  2 +-
 frontend/src/version.ts             |  2 +-
 scripts/release-manifest.py         | 11 ++++++-----
 7 files changed, 39 insertions(+), 23 deletions(-)
```

准备仅七个文件：五个版本文件从 1.0.3-rc.2 → 1.0.3-rc.3，Release 文案及生成器 changelog；Agent/安装器已复用版本注入，无需额外改源码。git diff --stat -- version.json 仍为空。
实际本地 python3 scripts/test-release-manifest.py：
```text
test_rc_fields_are_consistent (__main__.ReleaseMetadataTest.test_rc_fields_are_consistent) ... ok
test_rejects_bad_digest (__main__.ReleaseMetadataTest.test_rejects_bad_digest) ... ok

----------------------------------------------------------------------
Ran 2 tests in 0.001s

OK
```
实际提交／普通推送输出：
```text
[main c41bcc2] chore(release): prepare rc.3 version and monthly quota release notes
 7 files changed, 39 insertions(+), 23 deletions(-)
To https://github.com/sle58083/yuchen-panel.git
   29a3c6e..c41bcc2  main -> main
```

独立准备之后的完整历史（不隐瞒额外获准提交）

实际命令：
```text
git log --oneline 8ace907..c41bcc2
```

原始输出：
```text
c41bcc2 chore(release): prepare rc.3 version and monthly quota release notes
29a3c6e Merge pull request #8 from sle58083/fix/task13-monthly-traffic
6f3d0ac fix(billing): refresh usage details and document quota validation evidence
f96f56a feat(billing): meter per-client traffic and enforce renewable monthly quotas
```


版本准备后的第二轮 main CI

实际命令：
```text
gh run list --repo sle58083/yuchen-panel --branch main --limit 3
```

原始输出：
```text
completed	success	chore(release): prepare rc.3 version and monthly quota release notes	CI	main	push	38639513321	6m44s	2026-10-14T21:20:37Z
completed	success	Merge pull request #8 from sle58083/fix/task13-monthly-traffic	CI	main	push	38639041909	1m12s	2026-10-14T21:17:05Z
completed	success	docs(release): record verified rc.2 publication and preserved assets	CI	main	push	38601161818	1m23s	2026-10-14T16:48:10Z
```


第二轮六 job 全绿

实际命令：
```text
gh run view 38639513321 --repo sle58083/yuchen-panel --json status,conclusion,jobs --jq '{status,conclusion,jobs:[.jobs[]|{name,status,conclusion}]}'
```

原始输出：
```text
{"conclusion":"success","jobs":[{"conclusion":"success","name":"Docker Compose builds","status":"completed"},{"conclusion":"success","name":"Shell syntax and ShellCheck","status":"completed"},{"conclusion":"success","name":"Go build, vet and race tests (agent)","status":"completed"},{"conclusion":"success","name":"Frontend build and typecheck","status":"completed"},{"conclusion":"success","name":"Fresh Docker installer integration (Ubuntu 22.04)","status":"completed"},{"conclusion":"success","name":"Go build, vet and race tests (backend)","status":"completed"}],"status":"completed"}
```

gh run watch 38639513321 --repo sle58083/yuchen-panel --exit-status 退出 0。Docker 阶段耗时较长，最终 run 6m44s success；没有超时跳过或提前打 tag。

## Step 4 · 只创建带注解 rc.3 tag

实际命令：
```text
git tag -a v1.0.3-rc.3 -F .git/rc3-tag-message.txt; git cat-file -t v1.0.3-rc.3; git cat-file -p v1.0.3-rc.3
```

原始输出：
```text
object c41bcc2b84fa6a14209baf8f3cf7b2b27de45277
type commit
tag v1.0.3-rc.3
tagger sle58083 <281434516+sle58083@users.noreply.github.com> 1792013472 +0800

v1.0.3-rc.3

包月限流量售卖功能。

新增：
- 按客户维度流量统计（Xray Stats API，仅本机监听）
- 入站标识改用客户 ID，避免同名客户串账
- 流量超限自动停用（到期同样生效）
- 包月续费自动清零并恢复启用；改备注不触发清零
- 客户列表 30 秒自动刷新（编辑期间暂停）

实现要点：
- epoch 游标法处理 Xray 重启与重复上报
- 溢出保护与跨服务器来源授权

⚠️ 限额执行存在 30-60 秒采样延迟（Agent 心跳间隔 30 秒）。
建议流量上限预留余量。
```

cat-file -t 实际输出 tag；推送实际输出：
```text
To https://github.com/sle58083/yuchen-panel.git
 * [new tag]         v1.0.3-rc.3 -> v1.0.3-rc.3
```

远端注解对象及剥离后的源码提交

实际命令：
```text
git ls-remote origin refs/tags/v1.0.3-rc.3 'refs/tags/v1.0.3-rc.3^{}'
```

原始输出：
```text
f7ecadba99b0b8837fb02b9ca8ec526762b7cda7	refs/tags/v1.0.3-rc.3
c41bcc2b84fa6a14209baf8f3cf7b2b27de45277	refs/tags/v1.0.3-rc.3^{}
```


## Step 5 · Release 由 tag 自动触发并成功

实际命令：
```text
gh run list --repo sle58083/yuchen-panel --workflow release.yml --limit 3
```

原始输出：
```text
completed	success	chore(release): prepare rc.3 version and monthly quota release notes	Build Release	v1.0.3-rc.3	push	38640897788	1m6s	2026-10-14T21:31:16Z
completed	success	Merge pull request #7 from sle58083/fix/audit-batch4-test-harness	Build Release	v1.0.3-rc.2	push	38600612485	1m11s	2026-10-14T16:44:18Z
completed	success	merge: audit round 1 security and client compatibility fixes	Build Release	v1.0.3-rc.1	push	38548305975	48s	2026-10-14T10:15:37Z
```


Release job 结论

实际命令：
```text
gh run view 38640897788 --repo sle58083/yuchen-panel --json status,conclusion,jobs（投影 job 结论）
```

原始输出：
```text
{"conclusion":"success","jobs":[{"conclusion":"success","name":"build","status":"completed"}],"status":"completed"}
```

gh run watch 38640897788 --repo sle58083/yuchen-panel --exit-status 退出 0；完整日志保存在 .codex-release-inspect/rc3/release-full-log.txt。由 CI 编译、打包、创建新资产、下载校验及回填，未手动上传资产。

实际 Release 页面信息和 CI 资产

实际命令：
```text
gh release view v1.0.3-rc.3 --repo sle58083/yuchen-panel --json tagName,isPrerelease,url,publishedAt,assets
```

原始输出：
```text
{"assets":[{"apiUrl":"https://api.github.com/repos/sle58083/yuchen-panel/releases/assets/641732658","contentType":"application/octet-stream","createdAt":"2026-10-14T21:32:13Z","digest":"sha256:df6e41934043a7f62e2efbf95e994cbd44308b7edaf1110cd4739398cd5dae42","downloadCount":0,"id":"RA_kwDOUcPjjc4mQBAy","label":"","name":"SHA256SUMS","size":95,"state":"uploaded","updatedAt":"2026-10-14T21:32:14Z","url":"https://github.com/sle58083/yuchen-panel/releases/download/v1.0.3-rc.3/SHA256SUMS"},{"apiUrl":"https://api.github.com/repos/sle58083/yuchen-panel/releases/assets/641732657","contentType":"application/zip","createdAt":"2026-10-14T21:32:13Z","digest":"sha256:b9c8e2b67a49d465e5024105b15ee89c4fd51b042dac9dc7f2f4631e8d4a1fbe","downloadCount":0,"id":"RA_kwDOUcPjjc4mQBAx","label":"","name":"yuchen-panel-v1.0.3-rc.3.zip","size":5707667,"state":"uploaded","updatedAt":"2026-10-14T21:32:14Z","url":"https://github.com/sle58083/yuchen-panel/releases/download/v1.0.3-rc.3/yuchen-panel-v1.0.3-rc.3.zip"}],"isPrerelease":true,"publishedAt":"2026-10-14T21:32:15Z","tagName":"v1.0.3-rc.3","url":"https://github.com/sle58083/yuchen-panel/releases/tag/v1.0.3-rc.3"}
```


独立下载的 ZIP 实测 SHA256

实际命令：
```text
gh release download v1.0.3-rc.3 --repo sle58083/yuchen-panel --dir .codex-release-inspect/rc3/download --pattern 'yuchen-panel-v1.0.3-rc.3.zip' --pattern SHA256SUMS
wsl.exe -d Ubuntu -- sha256sum '/mnt/e/gpt保存/2026-10-13/linux-xray-sing-box-v2ray-shell/work/audit-fix/.codex-release-inspect/rc3/download/yuchen-panel-v1.0.3-rc.3.zip'
```

原始输出：
```text
b9c8e2b67a49d465e5024105b15ee89c4fd51b042dac9dc7f2f4631e8d4a1fbe  /mnt/e/gpt保存/2026-10-13/linux-xray-sing-box-v2ray-shell/work/audit-fix/.codex-release-inspect/rc3/download/yuchen-panel-v1.0.3-rc.3.zip
```


三者逐字符比对输出（全部相同）

实际命令：
```text
Get-FileHash -Algorithm SHA256；读取下载的 SHA256SUMS 与 main 清单；用 PowerShell -cne 进行大小写敏感比较
```

原始输出：
```text
ZIP=b9c8e2b67a49d465e5024105b15ee89c4fd51b042dac9dc7f2f4631e8d4a1fbe
SHA256SUMS=b9c8e2b67a49d465e5024105b15ee89c4fd51b042dac9dc7f2f4631e8d4a1fbe
manifest=b9c8e2b67a49d465e5024105b15ee89c4fd51b042dac9dc7f2f4631e8d4a1fbe
```


main/version.json 全文

实际命令：
```text
(Invoke-WebRequest -UseBasicParsing 'https://raw.githubusercontent.com/sle58083/yuchen-panel/main/version.json').Content
```

原始输出：
```text
{
  "latest": "1.0.3-rc.3",
  "version": "1.0.3-rc.3",
  "codename": "rc",
  "package": "yuchen-panel-v1.0.3-rc.3.zip",
  "download_url": "https://github.com/sle58083/yuchen-panel/releases/download/v1.0.3-rc.3/yuchen-panel-v1.0.3-rc.3.zip",
  "sha256": "b9c8e2b67a49d465e5024105b15ee89c4fd51b042dac9dc7f2f4631e8d4a1fbe",
  "min_supported_version": "0.7.5",
  "released_at": "2026-10-14T21:32:16.487868Z",
  "changelog": [
    "包月限流量：按客户 ID 汇总 Xray 上传与下载，Stats API 仅监听本机",
    "epoch 与服务器游标持久化，处理重启与重复上报",
    "到期或超限停用；延长到期即续费，清零并恢复启用",
    "客户列表按字节显示本周期用量，30 秒刷新且编辑时暂停",
    "正常在线限额执行通常有约 30–60 秒延迟，异常时可能更久",
    "升级会轮换 JWT 密钥，旧会话失效，需要重新登录",
    "需要匹配的新版 API 与 Agent，VPS 真机流量验收仍待用户执行"
  ]
}
```


CI 回填 manifest 的实际提交

实际命令：
```text
git pull --ff-only origin main; git show -s --format='%H %s' HEAD
```

原始输出：
```text
38a24ea3ad17941c590823f217996ee706e7c3c4 release: publish verified manifest for v1.0.3-rc.3
```

latest/version 都为 1.0.3-rc.3、codename=rc、package/download_url 指向新 ZIP。对所有字段作严格比对，未手工改 version.json 或 SHA。

既有 Release 列表

实际命令：
```text
gh release list --repo sle58083/yuchen-panel --limit 6
```

原始输出：
```text
Yuchen Panel v1.0.3-rc.3	Pre-release	v1.0.3-rc.3	2026-10-14T21:32:15Z
Yuchen Panel v1.0.3-rc.2	Pre-release	v1.0.3-rc.2	2026-10-14T16:45:21Z
Yuchen Panel v1.0.3-rc.1	Pre-release	v1.0.3-rc.1	2026-10-14T10:16:18Z
Yuchen Panel v1.0.2	Latest	v1.0.2	2026-09-16T09:57:04Z
Yuchen Panel v1.0.1		v1.0.1	2026-09-15T17:30:03Z
Yuchen Panel v1.0.0		v1.0.0	2026-09-15T16:43:49Z
```


既有资产发布前后核对

实际命令：
```text
gh api repos/sle58083/yuchen-panel/releases --paginate；比较旧 Release id/updated_at，旧资产 id/name/size/digest/created_at/updated_at，并断言下载计数不下降
```

原始输出：
```text
v1.0.3-rc.2: id=640999541 name=SHA256SUMS size=95 digest=sha256:4521f28d1c671cc322b9ae027eacdc99a5ffcab3d2fb66afb005bff291f53a8c downloads=1->1; metadata unchanged
v1.0.3-rc.2: id=640999538 name=yuchen-panel-v1.0.3-rc.2.zip size=5667107 digest=sha256:91baadeaed1615eb079931c76fe6c90884b2f749c951d752e36868c5ea4b98dd downloads=5->5; metadata unchanged
v1.0.3-rc.1: id=640000957 name=SHA256SUMS size=95 digest=sha256:c6137b9a6a580c31c2c20f95313e565013d1bb009776589cfc5936bdb8c66802 downloads=1->1; metadata unchanged
v1.0.3-rc.1: id=640000960 name=yuchen-panel-v1.0.3-rc.1.zip size=5655759 digest=sha256:f4b97180e4b7102757974fd773a1beb5067c9c896c734a219084394543af8530 downloads=4->4; metadata unchanged
v1.0.2: id=567675591 name=yuchen-panel-v1.0.2-stable.zip size=5533673 digest=sha256:7b64a19b908f7485b26858ad80340f2af6333605e2223fadb85c497d195602ea downloads=13->13; metadata unchanged
v1.0.1: id=566170691 name=yuchen-panel-v1.0.1-stable.zip size=5532482 digest=sha256:230da902bdf030d072c5064f476b9a8b3e5069ecc7b7899a0c69ac45eadcb3f8 downloads=6->6; metadata unchanged
v1.0.0: id=566095185 name=yuchen-panel-v1.0.0-stable.zip size=5515117 digest=sha256:7104ef9fd6e30332d991fa165284102474fe2c77f2d9d378520bc31b7bb3d976 downloads=8->8; metadata unchanged
```


一行安装入口 HTTP 状态及前五行

实际命令：
```text
Invoke-WebRequest -UseBasicParsing 'https://raw.githubusercontent.com/sle58083/yuchen-panel/main/install.sh'
```

原始输出：
```text
HTTP 200
#!/usr/bin/env bash
set -Eeuo pipefail

export DEBIAN_FRONTEND="${DEBIAN_FRONTEND:-noninteractive}"
export NEEDRESTART_MODE="${NEEDRESTART_MODE:-a}"
```

## 当前状态与范围

- PR #8 已合并，分支 fix/task13-monthly-traffic 保留；旧本地备份分支保留。
- rc.3 注解对象 f7ecadba99b0b8837fb02b9ca8ec526762b7cda7，指向准备提交 c41bcc2b84fa6a14209baf8f3cf7b2b27de45277。
- CI 回填提交 38a24ea3ad17941c590823f217996ee706e7c3c4；本报告随后为纯文档提交，不改变 tag 或已发布 ZIP。
- rc.2 ZIP 下载计数 5→5，rc.1 4→4，v1.0.2 13→13；资产 ID、摘要和时间元数据不变。所有旧 Release 未被修改。
- 没有正式版 tag、force push、rebase、squash、clobber、手工修改 manifest/SHA 或手动上传。
- 无本轮失败步骤；第二轮 Docker 耗时较长但最终真实 success。

## 升级到 v1.0.3-rc.3

```bash
bash <(curl -Ls https://raw.githubusercontent.com/sle58083/yuchen-panel/main/install.sh)
```

⚠️ 需要重新登录：安装／升级轮换 JWT 签名密钥，是刻意的安全行为。
要同时更新 API 和 Agent；只刷新前端不能提供按客户计量。Task13 报告里的“当时未发布”是历史阶段说明，此版本已由 CI 正式创建为 RC。

### 包月限流量售卖

计费周期：

- 到期时间与流量上限任一触发即停用。
- 续费（延长到期时间）后流量归零，进入新周期并恢复启用。
- 改备注、提前到期不会清零。
- 默认套餐 30 天；起点为新到期减套餐天数，延长会立即清零。

流量统计：

- 按客户 ID 统计上传＋下载，从启用新 Stats 后的该客户连接开始；过去未启用统计的流量不能按客户追溯。
- Xray 重启保留已结算累计；未采样的异常退出尾部不可恢复，详见 TASK13-REPORT.md。
- 达到限额自动停用；0 为不限；GB 按 2^30 bytes 换算。
- 客户列表 30 秒刷新，编辑窗口打开时暂停。

### 限额执行延迟

Agent 默认每 30 秒上报。正常在线时，超量到配置撤销通常约 30–60 秒，期间仍可使用；离线、网络失败或配置应用失败可能更久，这不是硬时限保证。

任务给出的余量试用值：

| 售卖规格 | 任务建议的内部试用上限 |
|---|---|
| 100 GB/月 | 90 GB |
| 500 GB/月 | 450 GB |

这不是经过本次真机流量测试得到的承诺。较低限额可能提前停用，不能把它当作足量交付保证；余量需依据峰值并发吞吐与真实同步延迟验证。

### 五步自测（纠正整数 GB 限额限制）

1. 新建测试客户，绑定测试节点，设未来到期和 1 GB 上限。
2. 客户端下载约 100 MiB。
3. 等待上报与列表刷新，核对非零 traffic_used_bytes 和界面用量，计入上传＋下载。
4. 继续使本周期总量达到／超过 1 GB，核对 Enabled=false、Agent 新配置应用以及缓存凭据无法连接。
5. 延长到期，核对用量归零、启用恢复，并在配置同步后重连。

现有字段仅接受整数 GB，不能设置任务示例中的 10 MB。本轮不增加 MB 功能、不修改数据库来冒充验收。
VPS 实际流量、重启计量与停用／续费验证仍待用户执行，本轮状态为 **SKIPPED（未执行 VPS 部署或客户真实流量测试）**。不拿 CI、源码配置 check 或本地空 StatsService 查询替代这一结论。

发布页：https://github.com/sle58083/yuchen-panel/releases/tag/v1.0.3-rc.3
合并 CI：https://github.com/sle58083/yuchen-panel/actions/runs/38639041909
准备 CI：https://github.com/sle58083/yuchen-panel/actions/runs/38639513321
Release CI：https://github.com/sle58083/yuchen-panel/actions/runs/38640897788