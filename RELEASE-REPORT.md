# v1.0.3-rc.2 发布报告

2026-10-15 Asia/Shanghai；只授权 RC，无 v1.0.3 正式 tag。
Step 1–7 已完成。资产由 CI 构建上传，ZIP / SHA256SUMS / main manifest 三者一致。尚未在用户 VPS 实机安装新 RC。

## 0. 授权修正与原始失败记录

用户明确授权：PR #5 已合并，因此新建替代 PR；另加仅版本号的 rc.2 准备提交，不手改 version.json。用户也确认：完整保留 5050059 范围日志，以 main 为基线证明 Task11 只有两条提交，并单独记录版本提交。

上次 Step1 pull --ff-only 因本地旧历史 7bd7628 与远端 5050059 分叉退出 128，已停止，无远端变更；完整失败报告仍保存在 .codex-release-inspect/rc2/RELEASE-REPORT.md（此文件为失败现场，不覆盖）。
本轮先创建 backup/stale-audit-7bd7628，再按用户指定 checkout -B 从远端恢复。未使用 reset --hard/rebase/force push。尝试设置 upstream 时因 fetch 仅声明 main 被拒绝（fatal: cannot set up tracking information; starting point 'origin/fix/audit-batch4-test-harness' is not a branch）；后续获用户继续确认后，增加本地 fetch refspec 并设置 upstream 成功。未隐藏此前失败。

## Step 1 · 备份、恢复与合入审计分支

实际命令：
```bash
git branch backup/stale-audit-7bd7628 fix/audit-batch4-test-harness
git checkout -B fix/audit-batch4-test-harness origin/fix/audit-batch4-test-harness
git fetch origin fix/task11-clean-node-names:refs/remotes/origin/fix/task11-clean-node-names
git config --add remote.origin.fetch '+refs/heads/fix/audit-batch4-test-harness:refs/remotes/origin/fix/audit-batch4-test-harness'
git fetch origin
git branch --set-upstream-to=origin/fix/audit-batch4-test-harness
git merge --no-ff origin/fix/task11-clean-node-names -F .git/task11-merge-message.txt
```
真实输出：
```text
Reset branch 'fix/audit-batch4-test-harness'
5050059 fix(release): preserve RC version and publish verified complete metadata
branch 'fix/audit-batch4-test-harness' set up to track 'origin/fix/audit-batch4-test-harness'.
Merge made by the 'ort' strategy.
```
checkout -B 只移动已备份的本地分支，不是 reset --hard；备份保留。

本地分支和备份

实际命令：
```text
git branch -vv（恢复后的快照）
```

原始输出：
```text
backup/stale-audit-7bd7628    7bd7628 docs(test): record Task 9 green CI and installer regression results
  fix/audit-batch1-security     b3cb4a1 fix(access)!: enforce explicit grants and retire Shadowsocks
  fix/audit-batch10-nginx       5fc4312 docs(test): record Task 10 real nginx and green CI evidence
  fix/audit-batch3-client-uri   7854bf1 fix(clients): close URI encoding and validation gaps
* fix/audit-batch4-test-harness 5050059 [origin/fix/audit-batch4-test-harness] fix(release): preserve RC version and publish verified complete metadata
  fix/task11-clean-node-names   47c193e docs(task11): record clean-name exports and actual regression evidence
  main                          21bceee [origin/main] docs(release): record verified v1.0.3-rc.1 publication evidence
  release/prepare-rc1           5738677 fix(release): preserve RC version and publish verified complete metadata
```


完整 5050059 范围（合并后、版本准备前）

实际命令：
```text
git log --oneline 5050059..HEAD
```

原始输出：
```text
7494af0 feat(share): drop short ID labels and keep numeric deduplication
47c193e docs(task11): record clean-name exports and actual regression evidence
20a7a22 fix(share): remove short ID labels and retain numeric name deduplication
21bceee docs(release): record verified v1.0.3-rc.1 publication evidence
8856345 release: publish verified manifest for v1.0.3-rc.1
da927c6 merge: audit round 1 security and client compatibility fixes
```


未夹带功能提交的 main 基线证据

实际命令：
```text
git log --oneline origin/main..HEAD（合并后、版本准备前）
```

原始输出：
```text
7494af0 feat(share): drop short ID labels and keep numeric deduplication
47c193e docs(task11): record clean-name exports and actual regression evidence
20a7a22 fix(share): remove short ID labels and retain numeric name deduplication
```

其中 da927c6 / 8856345 / 21bceee 是 main 已有的 rc.1 发布继承历史，并非本轮新功能；这点由用户确认验收口径。7494af0 是本轮 --no-ff merge；真正 Task11 是 20a7a22 + 47c193e。

单独获授权的版本准备提交 ceb95f4，仅五个文件：VERSION、backend/internal/version/version.go、frontend/package.json、frontend/package-lock.json、frontend/src/version.ts。Agent 沿用从 VERSION 注入；没有手工改 version.json。
实际提交/推送输出：
```text
[fix/audit-batch4-test-harness ceb95f4] chore(release): prepare unified v1.0.3-rc.2 version
 5 files changed, 6 insertions(+), 6 deletions(-)
To https://github.com/sle58083/yuchen-panel.git
   5050059..ceb95f4  fix/audit-batch4-test-harness -> fix/audit-batch4-test-harness
```

完整获准增量（含版本准备）

实际命令：
```text
git log --oneline origin/main..HEAD（版本提交后）
```

原始输出：
```text
ceb95f4 chore(release): prepare unified v1.0.3-rc.2 version
7494af0 feat(share): drop short ID labels and keep numeric deduplication
47c193e docs(task11): record clean-name exports and actual regression evidence
20a7a22 fix(share): remove short ID labels and retain numeric name deduplication
```


## Step 2 · 审计分支 CI 全绿

实际命令：
```text
gh run list --repo sle58083/yuchen-panel --branch fix/audit-batch4-test-harness --limit 3
```

原始输出：
```text
completed	success	release: Task11 clean node names and v1.0.3-rc.2 preparation	CI	fix/audit-batch4-test-harness	pull_request	38600129274	1m13s	2026-10-14T16:40:50Z
completed	success	chore(release): prepare unified v1.0.3-rc.2 version	CI	fix/audit-batch4-test-harness	push	38600119060	1m16s	2026-10-14T16:40:46Z
completed	success	release: audited fixes and v1.0.3-rc.1 preparation	CI	fix/audit-batch4-test-harness	pull_request	38545571083	1m6s	2026-10-14T09:52:46Z
```


分支六个 job 的真实结论

实际命令：
```text
gh run view 38600129274 --repo sle58083/yuchen-panel --json status,conclusion,jobs
```

原始输出：
```text
{"conclusion":"success","jobs":[{"completedAt":"2026-10-14T16:41:14Z","conclusion":"success","databaseId":115976925919,"name":"Frontend build and typecheck","startedAt":"2026-10-14T16:40:53Z","status":"completed","steps":[{"completedAt":"2026-10-14T16:40:55Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-14T16:40:54Z","status":"completed"},{"completedAt":"2026-10-14T16:40:55Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-14T16:40:55Z","status":"completed"},{"completedAt":"2026-10-14T16:41:00Z","conclusion":"success","name":"Run actions/setup-node@v4","number":3,"startedAt":"2026-10-14T16:40:55Z","status":"completed"},{"completedAt":"2026-10-14T16:41:02Z","conclusion":"success","name":"Run npm ci --no-audit --no-fund --progress=false","number":4,"startedAt":"2026-10-14T16:41:00Z","status":"completed"},{"completedAt":"2026-10-14T16:41:02Z","conclusion":"success","name":"Run node --test tests/*.test.mjs","number":5,"startedAt":"2026-10-14T16:41:02Z","status":"completed"},{"completedAt":"2026-10-14T16:41:08Z","conclusion":"success","name":"Run npm run build","number":6,"startedAt":"2026-10-14T16:41:02Z","status":"completed"},{"completedAt":"2026-10-14T16:41:12Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":11,"startedAt":"2026-10-14T16:41:08Z","status":"completed"},{"completedAt":"2026-10-14T16:41:12Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":12,"startedAt":"2026-10-14T16:41:12Z","status":"completed"},{"completedAt":"2026-10-14T16:41:12Z","conclusion":"success","name":"Complete job","number":13,"startedAt":"2026-10-14T16:41:12Z","status":"completed"}],"url":"https://github.com/sle58083/yuchen-panel/actions/runs/38600129274/job/115976925919"},{"completedAt":"2026-10-14T16:41:18Z","conclusion":"success","databaseId":115976926381,"name":"Shell syntax and ShellCheck","startedAt":"2026-10-14T16:40:53Z","status":"completed","steps":[{"completedAt":"2026-10-14T16:40:55Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-14T16:40:54Z","status":"completed"},{"completedAt":"2026-10-14T16:40:56Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-14T16:40:55Z","status":"completed"},{"completedAt":"2026-10-14T16:41:06Z","conclusion":"success","name":"Install ShellCheck","number":3,"startedAt":"2026-10-14T16:40:56Z","status":"completed"},{"completedAt":"2026-10-14T16:41:09Z","conclusion":"success","name":"ShellCheck","number":4,"startedAt":"2026-10-14T16:41:06Z","status":"completed"},{"completedAt":"2026-10-14T16:41:14Z","conclusion":"success","name":"Docker installer preflight regression tests","number":5,"startedAt":"2026-10-14T16:41:09Z","status":"completed"},{"completedAt":"2026-10-14T16:41:14Z","conclusion":"success","name":"JWT notice and Doctor mode regression tests","number":6,"startedAt":"2026-10-14T16:41:14Z","status":"completed"},{"completedAt":"2026-10-14T16:41:14Z","conclusion":"success","name":"Release manifest consistency regression","number":7,"startedAt":"2026-10-14T16:41:14Z","status":"completed"},{"completedAt":"2026-10-14T16:41:15Z","conclusion":"success","name":"Bash syntax","number":8,"startedAt":"2026-10-14T16:41:14Z","status":"completed"},{"completedAt":"2026-10-14T16:41:15Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-14T16:41:15Z","status":"completed"},{"completedAt":"2026-10-14T16:41:15Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-14T16:41:15Z","status":"completed"}],"url":"https://github.com/sle58083/yuchen-panel/actions/runs/38600129274/job/115976926381"},{"completedAt":"2026-10-14T16:41:40Z","conclusion":"success","databaseId":115976926434,"name":"Fresh Docker installer integration (Ubuntu 22.04)","startedAt":"2026-10-14T16:40:53Z","status":"completed","steps":[{"completedAt":"2026-10-14T16:40:54Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-14T16:40:53Z","status":"completed"},{"completedAt":"2026-10-14T16:40:54Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-14T16:40:54Z","status":"completed"},{"completedAt":"2026-10-14T16:41:01Z","conclusion":"success","name":"Integration test dependencies","number":3,"startedAt":"2026-10-14T16:40:54Z","status":"completed"},{"completedAt":"2026-10-14T16:41:37Z","conclusion":"success","name":"Execute isolated source installer and eight assertions","number":4,"startedAt":"2026-10-14T16:41:01Z","status":"completed"},{"completedAt":"2026-10-14T16:41:38Z","conclusion":"success","name":"Run actions/upload-artifact@v4","number":5,"startedAt":"2026-10-14T16:41:37Z","status":"completed"},{"completedAt":"2026-10-14T16:41:38Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":10,"startedAt":"2026-10-14T16:41:38Z","status":"completed"},{"completedAt":"2026-10-14T16:41:38Z","conclusion":"success","name":"Complete job","number":11,"startedAt":"2026-10-14T16:41:38Z","status":"completed"}],"url":"https://github.com/sle58083/yuchen-panel/actions/runs/38600129274/job/115976926434"},{"completedAt":"2026-10-14T16:41:28Z","conclusion":"success","databaseId":115976926454,"name":"Go build, vet and race tests (agent)","startedAt":"2026-10-14T16:40:53Z","status":"completed","steps":[{"completedAt":"2026-10-14T16:40:54Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-14T16:40:53Z","status":"completed"},{"completedAt":"2026-10-14T16:40:55Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-14T16:40:54Z","status":"completed"},{"completedAt":"2026-10-14T16:41:02Z","conclusion":"success","name":"Run actions/setup-go@v5","number":3,"startedAt":"2026-10-14T16:40:55Z","status":"completed"},{"completedAt":"2026-10-14T16:41:11Z","conclusion":"success","name":"Run go build ./...","number":4,"startedAt":"2026-10-14T16:41:02Z","status":"completed"},{"completedAt":"2026-10-14T16:41:11Z","conclusion":"success","name":"Build agent with release version injection","number":5,"startedAt":"2026-10-14T16:41:11Z","status":"completed"},{"completedAt":"2026-10-14T16:41:11Z","conclusion":"success","name":"Check release version single source","number":6,"startedAt":"2026-10-14T16:41:11Z","status":"completed"},{"completedAt":"2026-10-14T16:41:12Z","conclusion":"success","name":"Run go vet ./...","number":7,"startedAt":"2026-10-14T16:41:11Z","status":"completed"},{"completedAt":"2026-10-14T16:41:25Z","conclusion":"success","name":"Run go test -race ./...","number":8,"startedAt":"2026-10-14T16:41:12Z","status":"completed"},{"completedAt":"2026-10-14T16:41:25Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":15,"startedAt":"2026-10-14T16:41:25Z","status":"completed"},{"completedAt":"2026-10-14T16:41:25Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-14T16:41:25Z","status":"completed"},{"completedAt":"2026-10-14T16:41:25Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-14T16:41:25Z","status":"completed"}],"url":"https://github.com/sle58083/yuchen-panel/actions/runs/38600129274/job/115976926454"},{"completedAt":"2026-10-14T16:42:01Z","conclusion":"success","databaseId":115976926459,"name":"Docker Compose builds","startedAt":"2026-10-14T16:40:53Z","status":"completed","steps":[{"completedAt":"2026-10-14T16:40:54Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-14T16:40:53Z","status":"completed"},{"completedAt":"2026-10-14T16:40:54Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-14T16:40:54Z","status":"completed"},{"completedAt":"2026-10-14T16:41:35Z","conclusion":"success","name":"Run docker compose build yuchen-panel-api","number":3,"startedAt":"2026-10-14T16:40:54Z","status":"completed"},{"completedAt":"2026-10-14T16:42:00Z","conclusion":"success","name":"Run docker compose build yuchen-panel-frontend","number":4,"startedAt":"2026-10-14T16:41:35Z","status":"completed"},{"completedAt":"2026-10-14T16:42:00Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":8,"startedAt":"2026-10-14T16:42:00Z","status":"completed"},{"completedAt":"2026-10-14T16:42:00Z","conclusion":"success","name":"Complete job","number":9,"startedAt":"2026-10-14T16:42:00Z","status":"completed"}],"url":"https://github.com/sle58083/yuchen-panel/actions/runs/38600129274/job/115976926459"},{"completedAt":"2026-10-14T16:41:41Z","conclusion":"success","databaseId":115976926702,"name":"Go build, vet and race tests (backend)","startedAt":"2026-10-14T16:40:53Z","status":"completed","steps":[{"completedAt":"2026-10-14T16:40:55Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-14T16:40:54Z","status":"completed"},{"completedAt":"2026-10-14T16:40:56Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-14T16:40:55Z","status":"completed"},{"completedAt":"2026-10-14T16:41:05Z","conclusion":"success","name":"Run actions/setup-go@v5","number":3,"startedAt":"2026-10-14T16:40:56Z","status":"completed"},{"completedAt":"2026-10-14T16:41:15Z","conclusion":"success","name":"Run go build ./...","number":4,"startedAt":"2026-10-14T16:41:05Z","status":"completed"},{"completedAt":"2026-10-14T16:41:15Z","conclusion":"skipped","name":"Build agent with release version injection","number":5,"startedAt":"2026-10-14T16:41:15Z","status":"completed"},{"completedAt":"2026-10-14T16:41:15Z","conclusion":"skipped","name":"Check release version single source","number":6,"startedAt":"2026-10-14T16:41:15Z","status":"completed"},{"completedAt":"2026-10-14T16:41:17Z","conclusion":"success","name":"Run go vet ./...","number":7,"startedAt":"2026-10-14T16:41:15Z","status":"completed"},{"completedAt":"2026-10-14T16:41:38Z","conclusion":"success","name":"Run go test -race ./...","number":8,"startedAt":"2026-10-14T16:41:17Z","status":"completed"},{"completedAt":"2026-10-14T16:41:38Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":15,"startedAt":"2026-10-14T16:41:38Z","status":"completed"},{"completedAt":"2026-10-14T16:41:38Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-14T16:41:38Z","status":"completed"},{"completedAt":"2026-10-14T16:41:38Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-14T16:41:38Z","status":"completed"}],"url":"https://github.com/sle58083/yuchen-panel/actions/runs/38600129274/job/115976926702"}],"status":"completed"}
```

## Step 3 · 替代 PR #7 非 squash 合并

实际命令：
```bash
gh pr create --repo sle58083/yuchen-panel --head fix/audit-batch4-test-harness --base main --title 'release: Task11 clean node names and v1.0.3-rc.2 preparation' --body-file .codex-release-inspect/rc2/pr-body.txt
gh pr merge 7 --repo sle58083/yuchen-panel --merge --delete-branch=false --match-head-commit ceb95f4f4e8c2adab04ae26970175da5172f430a
```
PR create 实际输出 https://github.com/sle58083/yuchen-panel/pull/7。PR #5 已在 rc.1 合并，不能再次合并，因此按用户确认使用替代 PR。

PR 合并状态

实际命令：
```text
gh pr view 7 --repo sle58083/yuchen-panel --json state,mergedAt,mergeCommit
```

原始输出：
```text
{"mergeCommit":{"oid":"ef0e76d2181c7b04e75bfa852be7923023fed340"},"mergedAt":"2026-10-14T16:42:36Z","state":"MERGED"}
```


main 合并双父提交

实际命令：
```text
git show -s --format='%H%n%P%n%B' HEAD（pull 后）
```

原始输出：
```text
ef0e76d2181c7b04e75bfa852be7923023fed340
21bceee80f1f64ff2fa502bd139e81928f99685e ceb95f4f4e8c2adab04ae26970175da5172f430a
Merge pull request #7 from sle58083/fix/audit-batch4-test-harness

release: Task11 clean node names and v1.0.3-rc.2 preparation
```


## Step 4 · main CI 全绿

实际命令：
```text
gh run list --repo sle58083/yuchen-panel --branch main --limit 3
```

原始输出：
```text
completed	success	Merge pull request #7 from sle58083/fix/audit-batch4-test-harness	CI	main	push	38600379832	1m8s	2026-10-14T16:42:38Z
completed	success	docs(release): record verified v1.0.3-rc.1 publication evidence	CI	main	push	38548765906	1m8s	2026-10-14T10:19:31Z
completed	success	merge: audit round 1 security and client compatibility fixes	CI	main	push	38548087432	1m8s	2026-10-14T10:13:47Z
```


main 全部 job 结论

实际命令：
```text
gh run view 38600379832 --repo sle58083/yuchen-panel --json status,conclusion,jobs
```

原始输出：
```text
{"conclusion":"success","jobs":[{"completedAt":"2026-10-14T16:43:24Z","conclusion":"success","databaseId":115977787909,"name":"Go build, vet and race tests (agent)","startedAt":"2026-10-14T16:42:41Z","status":"completed","steps":[{"completedAt":"2026-10-14T16:42:42Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-14T16:42:41Z","status":"completed"},{"completedAt":"2026-10-14T16:42:43Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-14T16:42:42Z","status":"completed"},{"completedAt":"2026-10-14T16:42:57Z","conclusion":"success","name":"Run actions/setup-go@v5","number":3,"startedAt":"2026-10-14T16:42:43Z","status":"completed"},{"completedAt":"2026-10-14T16:43:06Z","conclusion":"success","name":"Run go build ./...","number":4,"startedAt":"2026-10-14T16:42:57Z","status":"completed"},{"completedAt":"2026-10-14T16:43:06Z","conclusion":"success","name":"Build agent with release version injection","number":5,"startedAt":"2026-10-14T16:43:06Z","status":"completed"},{"completedAt":"2026-10-14T16:43:06Z","conclusion":"success","name":"Check release version single source","number":6,"startedAt":"2026-10-14T16:43:06Z","status":"completed"},{"completedAt":"2026-10-14T16:43:08Z","conclusion":"success","name":"Run go vet ./...","number":7,"startedAt":"2026-10-14T16:43:06Z","status":"completed"},{"completedAt":"2026-10-14T16:43:21Z","conclusion":"success","name":"Run go test -race ./...","number":8,"startedAt":"2026-10-14T16:43:08Z","status":"completed"},{"completedAt":"2026-10-14T16:43:21Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":15,"startedAt":"2026-10-14T16:43:21Z","status":"completed"},{"completedAt":"2026-10-14T16:43:21Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-14T16:43:21Z","status":"completed"},{"completedAt":"2026-10-14T16:43:21Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-14T16:43:21Z","status":"completed"}],"url":"https://github.com/sle58083/yuchen-panel/actions/runs/38600379832/job/115977787909"},{"completedAt":"2026-10-14T16:42:57Z","conclusion":"success","databaseId":115977788168,"name":"Frontend build and typecheck","startedAt":"2026-10-14T16:42:41Z","status":"completed","steps":[{"completedAt":"2026-10-14T16:42:42Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-14T16:42:41Z","status":"completed"},{"completedAt":"2026-10-14T16:42:43Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-14T16:42:42Z","status":"completed"},{"completedAt":"2026-10-14T16:42:47Z","conclusion":"success","name":"Run actions/setup-node@v4","number":3,"startedAt":"2026-10-14T16:42:43Z","status":"completed"},{"completedAt":"2026-10-14T16:42:49Z","conclusion":"success","name":"Run npm ci --no-audit --no-fund --progress=false","number":4,"startedAt":"2026-10-14T16:42:47Z","status":"completed"},{"completedAt":"2026-10-14T16:42:49Z","conclusion":"success","name":"Run node --test tests/*.test.mjs","number":5,"startedAt":"2026-10-14T16:42:49Z","status":"completed"},{"completedAt":"2026-10-14T16:42:55Z","conclusion":"success","name":"Run npm run build","number":6,"startedAt":"2026-10-14T16:42:49Z","status":"completed"},{"completedAt":"2026-10-14T16:42:55Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":11,"startedAt":"2026-10-14T16:42:55Z","status":"completed"},{"completedAt":"2026-10-14T16:42:56Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":12,"startedAt":"2026-10-14T16:42:55Z","status":"completed"},{"completedAt":"2026-10-14T16:42:56Z","conclusion":"success","name":"Complete job","number":13,"startedAt":"2026-10-14T16:42:56Z","status":"completed"}],"url":"https://github.com/sle58083/yuchen-panel/actions/runs/38600379832/job/115977788168"},{"completedAt":"2026-10-14T16:43:41Z","conclusion":"success","databaseId":115977788209,"name":"Fresh Docker installer integration (Ubuntu 22.04)","startedAt":"2026-10-14T16:42:41Z","status":"completed","steps":[{"completedAt":"2026-10-14T16:42:42Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-14T16:42:41Z","status":"completed"},{"completedAt":"2026-10-14T16:42:43Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-14T16:42:42Z","status":"completed"},{"completedAt":"2026-10-14T16:42:49Z","conclusion":"success","name":"Integration test dependencies","number":3,"startedAt":"2026-10-14T16:42:43Z","status":"completed"},{"completedAt":"2026-10-14T16:43:38Z","conclusion":"success","name":"Execute isolated source installer and eight assertions","number":4,"startedAt":"2026-10-14T16:42:49Z","status":"completed"},{"completedAt":"2026-10-14T16:43:39Z","conclusion":"success","name":"Run actions/upload-artifact@v4","number":5,"startedAt":"2026-10-14T16:43:38Z","status":"completed"},{"completedAt":"2026-10-14T16:43:39Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":10,"startedAt":"2026-10-14T16:43:39Z","status":"completed"},{"completedAt":"2026-10-14T16:43:39Z","conclusion":"success","name":"Complete job","number":11,"startedAt":"2026-10-14T16:43:39Z","status":"completed"}],"url":"https://github.com/sle58083/yuchen-panel/actions/runs/38600379832/job/115977788209"},{"completedAt":"2026-10-14T16:43:45Z","conclusion":"success","databaseId":115977788312,"name":"Docker Compose builds","startedAt":"2026-10-14T16:42:41Z","status":"completed","steps":[{"completedAt":"2026-10-14T16:42:43Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-14T16:42:42Z","status":"completed"},{"completedAt":"2026-10-14T16:42:44Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-14T16:42:43Z","status":"completed"},{"completedAt":"2026-10-14T16:43:24Z","conclusion":"success","name":"Run docker compose build yuchen-panel-api","number":3,"startedAt":"2026-10-14T16:42:44Z","status":"completed"},{"completedAt":"2026-10-14T16:43:44Z","conclusion":"success","name":"Run docker compose build yuchen-panel-frontend","number":4,"startedAt":"2026-10-14T16:43:24Z","status":"completed"},{"completedAt":"2026-10-14T16:43:44Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":8,"startedAt":"2026-10-14T16:43:44Z","status":"completed"},{"completedAt":"2026-10-14T16:43:44Z","conclusion":"success","name":"Complete job","number":9,"startedAt":"2026-10-14T16:43:44Z","status":"completed"}],"url":"https://github.com/sle58083/yuchen-panel/actions/runs/38600379832/job/115977788312"},{"completedAt":"2026-10-14T16:42:59Z","conclusion":"success","databaseId":115977788379,"name":"Shell syntax and ShellCheck","startedAt":"2026-10-14T16:42:41Z","status":"completed","steps":[{"completedAt":"2026-10-14T16:42:43Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-14T16:42:42Z","status":"completed"},{"completedAt":"2026-10-14T16:42:44Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-14T16:42:43Z","status":"completed"},{"completedAt":"2026-10-14T16:42:51Z","conclusion":"success","name":"Install ShellCheck","number":3,"startedAt":"2026-10-14T16:42:44Z","status":"completed"},{"completedAt":"2026-10-14T16:42:54Z","conclusion":"success","name":"ShellCheck","number":4,"startedAt":"2026-10-14T16:42:51Z","status":"completed"},{"completedAt":"2026-10-14T16:42:55Z","conclusion":"success","name":"Docker installer preflight regression tests","number":5,"startedAt":"2026-10-14T16:42:54Z","status":"completed"},{"completedAt":"2026-10-14T16:42:55Z","conclusion":"success","name":"JWT notice and Doctor mode regression tests","number":6,"startedAt":"2026-10-14T16:42:55Z","status":"completed"},{"completedAt":"2026-10-14T16:42:55Z","conclusion":"success","name":"Release manifest consistency regression","number":7,"startedAt":"2026-10-14T16:42:55Z","status":"completed"},{"completedAt":"2026-10-14T16:42:55Z","conclusion":"success","name":"Bash syntax","number":8,"startedAt":"2026-10-14T16:42:55Z","status":"completed"},{"completedAt":"2026-10-14T16:42:55Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-14T16:42:55Z","status":"completed"},{"completedAt":"2026-10-14T16:42:55Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-14T16:42:55Z","status":"completed"}],"url":"https://github.com/sle58083/yuchen-panel/actions/runs/38600379832/job/115977788379"},{"completedAt":"2026-10-14T16:43:29Z","conclusion":"success","databaseId":115977788490,"name":"Go build, vet and race tests (backend)","startedAt":"2026-10-14T16:42:41Z","status":"completed","steps":[{"completedAt":"2026-10-14T16:42:44Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-14T16:42:42Z","status":"completed"},{"completedAt":"2026-10-14T16:42:45Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-14T16:42:44Z","status":"completed"},{"completedAt":"2026-10-14T16:42:54Z","conclusion":"success","name":"Run actions/setup-go@v5","number":3,"startedAt":"2026-10-14T16:42:45Z","status":"completed"},{"completedAt":"2026-10-14T16:43:04Z","conclusion":"success","name":"Run go build ./...","number":4,"startedAt":"2026-10-14T16:42:54Z","status":"completed"},{"completedAt":"2026-10-14T16:43:04Z","conclusion":"skipped","name":"Build agent with release version injection","number":5,"startedAt":"2026-10-14T16:43:04Z","status":"completed"},{"completedAt":"2026-10-14T16:43:04Z","conclusion":"skipped","name":"Check release version single source","number":6,"startedAt":"2026-10-14T16:43:04Z","status":"completed"},{"completedAt":"2026-10-14T16:43:06Z","conclusion":"success","name":"Run go vet ./...","number":7,"startedAt":"2026-10-14T16:43:04Z","status":"completed"},{"completedAt":"2026-10-14T16:43:25Z","conclusion":"success","name":"Run go test -race ./...","number":8,"startedAt":"2026-10-14T16:43:06Z","status":"completed"},{"completedAt":"2026-10-14T16:43:26Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":15,"startedAt":"2026-10-14T16:43:25Z","status":"completed"},{"completedAt":"2026-10-14T16:43:26Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-14T16:43:26Z","status":"completed"},{"completedAt":"2026-10-14T16:43:26Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-14T16:43:26Z","status":"completed"}],"url":"https://github.com/sle58083/yuchen-panel/actions/runs/38600379832/job/115977788490"}],"status":"completed"}
```

gh run watch 38600379832 --exit-status 退出 0；全部 success 后才执行 tag。

## Step 5 · 带注解 rc.2 tag

实际命令：
```text
git tag -a v1.0.3-rc.2 -F .git/rc2-tag-message.txt; git cat-file -t v1.0.3-rc.2; git cat-file -p v1.0.3-rc.2
```

原始输出：
```text
object ef0e76d2181c7b04e75bfa852be7923023fed340
type commit
tag v1.0.3-rc.2
tagger sle58083 <281434516+sle58083@users.noreply.github.com> 1791996252 +0800

v1.0.3-rc.2

在 v1.0.3-rc.1 基础上：
- 节点名移除 #xxxx 短 ID 后缀，恢复干净名称（美国 / 日本 / 香港）
- 同名节点仍自动去重（美国 / 美国 2 / 美国 3），四条导出路径一致

升级会轮换 JWT 密钥，需要重新登录。
```

cat-file -t 输出 tag。普通推送实际输出：
```text
To https://github.com/sle58083/yuchen-panel.git
 * [new tag]         v1.0.3-rc.2 -> v1.0.3-rc.2
```

## Step 6 · CI Release 构建与回填

实际命令：
```text
gh run list --repo sle58083/yuchen-panel --workflow release.yml --limit 3
```

原始输出：
```text
completed	success	Merge pull request #7 from sle58083/fix/audit-batch4-test-harness	Build Release	v1.0.3-rc.2	push	38600612485	1m11s	2026-10-14T16:44:18Z
completed	success	merge: audit round 1 security and client compatibility fixes	Build Release	v1.0.3-rc.1	push	38548305975	48s	2026-10-14T10:15:37Z
completed	success	fix: 修复 VMess/Shadowsocks 入站测试超时(-1)	Build Release	v1.0.2	push	35082203560	55s	2026-09-16T09:56:14Z
```


Release workflow 结论

实际命令：
```text
gh run view 38600612485 --repo sle58083/yuchen-panel --json status,conclusion,jobs（投影 job 状态）
```

原始输出：
```text
{"conclusion":"success","jobs":[{"conclusion":"success","name":"build","status":"completed"}],"status":"completed"}
```

gh run watch 38600612485 --exit-status 退出 0。完整日志 .codex-release-inspect/rc2/release-full-log.txt。workflow 完成发布后，独立下载 ZIP 与 SHA256SUMS 校验。

实际 Release 页面和 CI 资产

实际命令：
```text
gh release view v1.0.3-rc.2 --repo sle58083/yuchen-panel --json tagName,isPrerelease,url,publishedAt,assets
```

原始输出：
```text
{"assets":[{"apiUrl":"https://api.github.com/repos/sle58083/yuchen-panel/releases/assets/640999541","contentType":"application/octet-stream","createdAt":"2026-10-14T16:45:20Z","digest":"sha256:4521f28d1c671cc322b9ae027eacdc99a5ffcab3d2fb66afb005bff291f53a8c","downloadCount":0,"id":"RA_kwDOUcPjjc4mNOB1","label":"","name":"SHA256SUMS","size":95,"state":"uploaded","updatedAt":"2026-10-14T16:45:20Z","url":"https://github.com/sle58083/yuchen-panel/releases/download/v1.0.3-rc.2/SHA256SUMS"},{"apiUrl":"https://api.github.com/repos/sle58083/yuchen-panel/releases/assets/640999538","contentType":"application/zip","createdAt":"2026-10-14T16:45:20Z","digest":"sha256:91baadeaed1615eb079931c76fe6c90884b2f749c951d752e36868c5ea4b98dd","downloadCount":0,"id":"RA_kwDOUcPjjc4mNOBy","label":"","name":"yuchen-panel-v1.0.3-rc.2.zip","size":5667107,"state":"uploaded","updatedAt":"2026-10-14T16:45:21Z","url":"https://github.com/sle58083/yuchen-panel/releases/download/v1.0.3-rc.2/yuchen-panel-v1.0.3-rc.2.zip"}],"isPrerelease":true,"publishedAt":"2026-10-14T16:45:21Z","tagName":"v1.0.3-rc.2","url":"https://github.com/sle58083/yuchen-panel/releases/tag/v1.0.3-rc.2"}
```


独立 sha256sum 原始结果

实际命令：
```text
gh release download v1.0.3-rc.2 --repo sle58083/yuchen-panel --dir .codex-release-inspect/rc2/download --pattern 'yuchen-panel-v1.0.3-rc.2.zip' --pattern SHA256SUMS
wsl.exe -d Ubuntu -- sha256sum '/mnt/e/gpt保存/2026-10-13/linux-xray-sing-box-v2ray-shell/work/audit-fix/.codex-release-inspect/rc2/download/yuchen-panel-v1.0.3-rc.2.zip'
```

原始输出：
```text
91baadeaed1615eb079931c76fe6c90884b2f749c951d752e36868c5ea4b98dd  /mnt/e/gpt保存/2026-10-13/linux-xray-sing-box-v2ray-shell/work/audit-fix/.codex-release-inspect/rc2/download/yuchen-panel-v1.0.3-rc.2.zip
```


ZIP / SHA256SUMS / manifest 三者逐字符一致

实际命令：
```text
Get-FileHash -Algorithm SHA256，读取下载的 SHA256SUMS；用 -cne 与 main manifest .sha256 比较
```

原始输出：
```text
ZIP=91baadeaed1615eb079931c76fe6c90884b2f749c951d752e36868c5ea4b98dd
SHA256SUMS=91baadeaed1615eb079931c76fe6c90884b2f749c951d752e36868c5ea4b98dd
manifest=91baadeaed1615eb079931c76fe6c90884b2f749c951d752e36868c5ea4b98dd
```


main/version.json 原始内容

实际命令：
```text
(Invoke-WebRequest -UseBasicParsing 'https://raw.githubusercontent.com/sle58083/yuchen-panel/main/version.json').Content
```

原始输出：
```text
{
  "latest": "1.0.3-rc.2",
  "version": "1.0.3-rc.2",
  "codename": "rc",
  "package": "yuchen-panel-v1.0.3-rc.2.zip",
  "download_url": "https://github.com/sle58083/yuchen-panel/releases/download/v1.0.3-rc.2/yuchen-panel-v1.0.3-rc.2.zip",
  "sha256": "91baadeaed1615eb079931c76fe6c90884b2f749c951d752e36868c5ea4b98dd",
  "min_supported_version": "0.7.5",
  "released_at": "2026-10-14T16:45:22.871404Z",
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


CI 清单回填提交

实际命令：
```text
git pull --ff-only origin main; git show -s --format='%H %s' HEAD
```

原始输出：
```text
b123bfb3de7794e447a6642289997092e4b8f2c1 release: publish verified manifest for v1.0.3-rc.2
```

本轮未手改 version.json。新 Release 创建时 workflow 使用沿用的 RELEASE-NOTES.md，随后只更新新 rc.2 Release 的说明为本轮变化：gh release edit v1.0.3-rc.2 --notes-file .codex-release-inspect/rc2/release-notes.txt。未上传/替换资产、未修改旧 Release。

## Step 7 · 旧资产与安装入口核验

实际命令：
```text
gh release list --repo sle58083/yuchen-panel --limit 6
```

原始输出：
```text
Yuchen Panel v1.0.3-rc.2	Pre-release	v1.0.3-rc.2	2026-10-14T16:45:21Z
Yuchen Panel v1.0.3-rc.1	Pre-release	v1.0.3-rc.1	2026-10-14T10:16:18Z
Yuchen Panel v1.0.2	Latest	v1.0.2	2026-09-16T09:57:04Z
Yuchen Panel v1.0.1		v1.0.1	2026-09-15T17:30:03Z
Yuchen Panel v1.0.0		v1.0.0	2026-09-15T16:43:49Z
```


发布前后旧资产核对

实际命令：
```text
gh api repos/sle58083/yuchen-panel/releases --paginate；逐一核对旧 release/asset id、size、digest、updated_at，并断言下载计数不下降
```

原始输出：
```text
v1.0.3-rc.1: id=640000957 name=SHA256SUMS size=95 digest=sha256:c6137b9a6a580c31c2c20f95313e565013d1bb009776589cfc5936bdb8c66802 download_count=1->1 updated_at=10/14/2026 10:16:17 unchanged
v1.0.3-rc.1: id=640000960 name=yuchen-panel-v1.0.3-rc.1.zip size=5655759 digest=sha256:f4b97180e4b7102757974fd773a1beb5067c9c896c734a219084394543af8530 download_count=4->4 updated_at=10/14/2026 10:16:18 unchanged
v1.0.2: id=567675591 name=yuchen-panel-v1.0.2-stable.zip size=5533673 digest=sha256:7b64a19b908f7485b26858ad80340f2af6333605e2223fadb85c497d195602ea download_count=13->13 updated_at=09/16/2026 09:57:03 unchanged
v1.0.1: id=566170691 name=yuchen-panel-v1.0.1-stable.zip size=5532482 digest=sha256:230da902bdf030d072c5064f476b9a8b3e5069ecc7b7899a0c69ac45eadcb3f8 download_count=6->6 updated_at=09/15/2026 17:30:02 unchanged
v1.0.0: id=566095185 name=yuchen-panel-v1.0.0-stable.zip size=5515117 digest=sha256:7104ef9fd6e30332d991fa165284102474fe2c77f2d9d378520bc31b7bb3d976 download_count=8->8 updated_at=09/15/2026 16:43:49 unchanged
```


安装入口状态和前五行

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

rc.1 ZIP 下载计数 4→4，v1.0.2 ZIP 13→13；ID/digest/时间均不变。旧资产完整保留。

## 升级到 v1.0.3-rc.2

```bash
bash <(curl -Ls https://raw.githubusercontent.com/sle58083/yuchen-panel/main/install.sh)
```

⚠️ 需要重新登录：JWT 签名密钥轮换是刻意的安全行为。

本版变化：节点名不再自动带 #xxxx 后缀，之前“美国 #80ac”现在“美国”；同名仍自动区分“美国 / 美国 2 / 美国 3”。
默认抗封锁档不变；CPU/内存提示依赖新版 Agent。

用户 VPS 安装与设备实际验收待用户执行。未创建 v1.0.3 正式 tag，未 squash/force push/rebase，备份与审计分支保留。本报告为发布校验后的文档提交，不改变已发布包或 tag。