# TASK11 — 移除节点名短 ID 后缀

分支 fix/task11-clean-node-names；代码提交 20a7a22；PR https://github.com/sle58083/yuchen-panel/pull/6。
基于 main 的已验证 RC 后续文档提交 21bceee。本轮没有合并 main、创建 tag 或触发 Release。

## 改动范围

- nodeShareName 保留 Trim 与空名 host:port 兜底，移除短 ID 拼接及 HasSuffix 防重复逻辑。
- uniqueSubscriptionNames 保留去重与节点副本操作，将数字冲突格式由 X #2 改为 X 2；保留 PROXY/direct/DIRECT 的既有保留名规则。
- 前端 safeName 移除短 ID；删除仅服务该后缀的 shortNodeId。
- 前端批量 URI 分享也复用 uniqueProxyNames；VMess ps 改为 safeName，以携带去重后的数字名。
- 前端空名使用与后端相同的 name→host:实际 port 规则，不以 remark 或短 ID 当显示名。此处只改名称，未改 URI 的端口生成逻辑。
- 未修改 Node ID、端口、协议、凭据、C24、Compose 预检、安装器或 Doctor。
- 原“12 同名节点唯一并可追溯短 ID”断言替换为更强的四路径精确列表+唯一性断言，保留 12 节点测试。另验证存储节点不变，仅导出副本的 Name 变化。

## Before / After

Before 根据修改前源码及旧测试期望列出，是行为对比，不冒充本轮重新执行旧版本的日志。

| 场景 | Before | After |
|---|---|---|
| 美国，ID node-80ac | 美国 #80ac | 美国 |
| 同名三个节点 | 美国 #80ac / 美国 #de00 / 美国 #b7ed | 美国 / 美国 2 / 美国 3 |
| 空名，192.0.2.1:24401 | 192.0.2.1:24401 #de00 | 192.0.2.1:24401 |

用户自行填写的名字里已有 # 字符会按原名保留（如原有 IPv6 测试名“测试 节点 #1”）；本补丁只移除自动追加行为，不修改数据库中的名字。

## 实际导出样例

下面输出来自本轮执行的 TestSubscriptionNamesAcrossFourExports，地址是文档保留地址 192.0.2.1，UUID/密码均为测试 fixture，无生产凭据。
```text
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20000?encryption=none&security=none&type=tcp#%E7%BE%8E%E5%9B%BD
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20001?encryption=none&security=none&type=tcp#%E6%97%A5%E6%9C%AC
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20002?encryption=none&security=none&type=tcp#%E9%9F%A9%E5%9B%BD
    subscription_names_test.go:102: sing-box tag: ["美国" "日本" "韩国"]
    subscription_names_test.go:102: Clash name: ["美国" "日本" "韩国"]
    subscription_names_test.go:102: fragment: ["美国" "日本" "韩国"]
    subscription_names_test.go:102: VMess ps: ["美国" "日本" "韩国"]
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20000?encryption=none&security=none&type=tcp#%E7%BE%8E%E5%9B%BD
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20001?encryption=none&security=none&type=tcp#%E7%BE%8E%E5%9B%BD%202
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20002?encryption=none&security=none&type=tcp#%E7%BE%8E%E5%9B%BD%203
    subscription_names_test.go:102: sing-box tag: ["美国" "美国 2" "美国 3"]
    subscription_names_test.go:102: Clash name: ["美国" "美国 2" "美国 3"]
    subscription_names_test.go:102: fragment: ["美国" "美国 2" "美国 3"]
    subscription_names_test.go:102: VMess ps: ["美国" "美国 2" "美国 3"]
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20000?encryption=none&security=none&type=tcp#%E9%87%8D%E5%A4%8D%E5%90%8D%E7%A7%B0
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20001?encryption=none&security=none&type=tcp#%E9%87%8D%E5%A4%8D%E5%90%8D%E7%A7%B0%202
```

## 后端专项完整原始输出（退出 0）

执行命令：
```text
cd backend; go test -v -run "Name|Share|Subscription" ./...
```

真实输出：
```text
?   	yuchen-panel/backend/cmd/server	[no test files]
=== RUN   TestSubscriptionUsesExplicitGrantsAndRetiresShadowsocks
--- PASS: TestSubscriptionUsesExplicitGrantsAndRetiresShadowsocks (0.00s)
=== RUN   TestSubscriptionFormatRouting
--- PASS: TestSubscriptionFormatRouting (0.00s)
=== RUN   TestNodeShareNameMatchesFrontendFixture
--- PASS: TestNodeShareNameMatchesFrontendFixture (0.00s)
=== RUN   TestDisplayNameKeepsOriginalAcrossFormats
--- PASS: TestDisplayNameKeepsOriginalAcrossFormats (0.00s)
=== RUN   TestSubscriptionNamesAcrossFourExports
=== RUN   TestSubscriptionNamesAcrossFourExports/distinct
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20000?encryption=none&security=none&type=tcp#%E7%BE%8E%E5%9B%BD
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20001?encryption=none&security=none&type=tcp#%E6%97%A5%E6%9C%AC
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20002?encryption=none&security=none&type=tcp#%E9%9F%A9%E5%9B%BD
    subscription_names_test.go:102: sing-box tag: ["美国" "日本" "韩国"]
    subscription_names_test.go:102: Clash name: ["美国" "日本" "韩国"]
    subscription_names_test.go:102: fragment: ["美国" "日本" "韩国"]
    subscription_names_test.go:102: VMess ps: ["美国" "日本" "韩国"]
=== RUN   TestSubscriptionNamesAcrossFourExports/duplicates
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20000?encryption=none&security=none&type=tcp#%E7%BE%8E%E5%9B%BD
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20001?encryption=none&security=none&type=tcp#%E7%BE%8E%E5%9B%BD%202
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20002?encryption=none&security=none&type=tcp#%E7%BE%8E%E5%9B%BD%203
    subscription_names_test.go:102: sing-box tag: ["美国" "美国 2" "美国 3"]
    subscription_names_test.go:102: Clash name: ["美国" "美国 2" "美国 3"]
    subscription_names_test.go:102: fragment: ["美国" "美国 2" "美国 3"]
    subscription_names_test.go:102: VMess ps: ["美国" "美国 2" "美国 3"]
=== RUN   TestSubscriptionNamesAcrossFourExports/twelve
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20000?encryption=none&security=none&type=tcp#%E9%87%8D%E5%A4%8D%E5%90%8D%E7%A7%B0
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20001?encryption=none&security=none&type=tcp#%E9%87%8D%E5%A4%8D%E5%90%8D%E7%A7%B0%202
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20002?encryption=none&security=none&type=tcp#%E9%87%8D%E5%A4%8D%E5%90%8D%E7%A7%B0%203
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20003?encryption=none&security=none&type=tcp#%E9%87%8D%E5%A4%8D%E5%90%8D%E7%A7%B0%204
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20004?encryption=none&security=none&type=tcp#%E9%87%8D%E5%A4%8D%E5%90%8D%E7%A7%B0%205
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20005?encryption=none&security=none&type=tcp#%E9%87%8D%E5%A4%8D%E5%90%8D%E7%A7%B0%206
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20006?encryption=none&security=none&type=tcp#%E9%87%8D%E5%A4%8D%E5%90%8D%E7%A7%B0%207
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20007?encryption=none&security=none&type=tcp#%E9%87%8D%E5%A4%8D%E5%90%8D%E7%A7%B0%208
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20008?encryption=none&security=none&type=tcp#%E9%87%8D%E5%A4%8D%E5%90%8D%E7%A7%B0%209
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20009?encryption=none&security=none&type=tcp#%E9%87%8D%E5%A4%8D%E5%90%8D%E7%A7%B0%2010
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20010?encryption=none&security=none&type=tcp#%E9%87%8D%E5%A4%8D%E5%90%8D%E7%A7%B0%2011
    subscription_names_test.go:74: VLESS: vless://f9b77af1-736a-4f5b-8bc5-273ba9b552ea@192.0.2.1:20011?encryption=none&security=none&type=tcp#%E9%87%8D%E5%A4%8D%E5%90%8D%E7%A7%B0%2012
    subscription_names_test.go:102: VMess ps: ["重复名称" "重复名称 2" "重复名称 3" "重复名称 4" "重复名称 5" "重复名称 6" "重复名称 7" "重复名称 8" "重复名称 9" "重复名称 10" "重复名称 11" "重复名称 12"]
    subscription_names_test.go:102: sing-box tag: ["重复名称" "重复名称 2" "重复名称 3" "重复名称 4" "重复名称 5" "重复名称 6" "重复名称 7" "重复名称 8" "重复名称 9" "重复名称 10" "重复名称 11" "重复名称 12"]
    subscription_names_test.go:102: Clash name: ["重复名称" "重复名称 2" "重复名称 3" "重复名称 4" "重复名称 5" "重复名称 6" "重复名称 7" "重复名称 8" "重复名称 9" "重复名称 10" "重复名称 11" "重复名称 12"]
    subscription_names_test.go:102: fragment: ["重复名称" "重复名称 2" "重复名称 3" "重复名称 4" "重复名称 5" "重复名称 6" "重复名称 7" "重复名称 8" "重复名称 9" "重复名称 10" "重复名称 11" "重复名称 12"]
--- PASS: TestSubscriptionNamesAcrossFourExports (0.00s)
    --- PASS: TestSubscriptionNamesAcrossFourExports/distinct (0.00s)
    --- PASS: TestSubscriptionNamesAcrossFourExports/duplicates (0.00s)
    --- PASS: TestSubscriptionNamesAcrossFourExports/twelve (0.00s)
=== RUN   TestNodeShareNameTrimAndEmptyFallback
--- PASS: TestNodeShareNameTrimAndEmptyFallback (0.00s)
PASS
ok  	yuchen-panel/backend/internal/api	0.016s
?   	yuchen-panel/backend/internal/model	[no test files]
?   	yuchen-panel/backend/internal/security	[no test files]
testing: warning: no tests to run
PASS
ok  	yuchen-panel/backend/internal/store	(cached) [no tests to run]
testing: warning: no tests to run
PASS
ok  	yuchen-panel/backend/internal/version	(cached) [no tests to run]
testing: warning: no tests to run
PASS
ok  	yuchen-panel/backend/internal/xray	(cached) [no tests to run]
```


## 后端完整竞态测试（退出 0）

执行命令：
```text
cd backend; go test -race ./...
```

真实输出：
```text
?   	yuchen-panel/backend/cmd/server	[no test files]
ok  	yuchen-panel/backend/internal/api	4.585s
?   	yuchen-panel/backend/internal/model	[no test files]
?   	yuchen-panel/backend/internal/security	[no test files]
ok  	yuchen-panel/backend/internal/store	(cached)
ok  	yuchen-panel/backend/internal/version	(cached)
ok  	yuchen-panel/backend/internal/xray	(cached)
```


## 前端完整测试输出（13 项，退出 0）

执行命令：
```text
cd frontend; node --test tests/*.test.mjs
```

真实输出：
```text
TAP version 13
# Subtest: C24_6_three_preset_parameter_mappings
ok 1 - C24_6_three_preset_parameter_mappings
  ---
  duration_ms: 7.2548
  type: 'test'
  ...
# Subtest: C24_7_two_or_fewer_cores_suggest_fast_without_overriding_default
ok 2 - C24_7_two_or_fewer_cores_suggest_fast_without_overriding_default
  ---
  duration_ms: 0.96
  type: 'test'
  ...
# TASK11 exported names: {"tags":["美国","日本","韩国"],"names":["美国","日本","韩国"],"fragments":["美国","日本","韩国"],"ps":["美国","日本","韩国"]}
# TASK11 exported names: {"tags":["美国","美国 2","美国 3"],"names":["美国","美国 2","美国 3"],"fragments":["美国","美国 2","美国 3"],"ps":["美国","美国 2","美国 3"]}
# TASK11 exported names: {"tags":["美国","example.com:23456","example.com:23456 2"],"names":["美国","example.com:23456","example.com:23456 2"],"fragments":["美国","example.com:23456","example.com:23456 2"],"ps":["美国","example.com:23456","example.com:23456 2"]}
# Subtest: special password and IPv6 URI round-trip preserves endpoint and name
ok 3 - special password and IPv6 URI round-trip preserves endpoint and name
  ---
  duration_ms: 7.0069
  type: 'test'
  ...
# Subtest: display names use numeric duplicate suffixes without short ids
ok 4 - display names use numeric duplicate suffixes without short ids
  ---
  duration_ms: 7.5222
  type: 'test'
  ...
# Subtest: display names do not vary with client username
ok 5 - display names do not vary with client username
  ---
  duration_ms: 1.0194
  type: 'test'
  ...
# Subtest: clean original names and numeric duplicates agree across all four export paths
ok 6 - clean original names and numeric duplicates agree across all four export paths
  ---
  duration_ms: 9.5897
  type: 'test'
  ...
# Subtest: unnamed node uses host and actual port rather than remark or ID
ok 7 - unnamed node uses host and actual port rather than remark or ID
  ---
  duration_ms: 0.9881
  type: 'test'
  ...
# Subtest: VMess network and default WS/gRPC transport parameters survive export
ok 8 - VMess network and default WS/gRPC transport parameters survive export
  ---
  duration_ms: 2.2037
  type: 'test'
  ...
# Subtest: sing-box DNS is modern, selector choices valid, tags unique even when empty
ok 9 - sing-box DNS is modern, selector choices valid, tags unique even when empty
  ---
  duration_ms: 0.8823
  type: 'test'
  ...
# Subtest: subscription format URLs are explicit and token path is encoded
ok 10 - subscription format URLs are explicit and token path is encoded
  ---
  duration_ms: 0.5277
  type: 'test'
  ...
# Subtest: optional VPN profile includes IPv4/IPv6 TUN, DNS hijack and loop prevention
ok 11 - optional VPN profile includes IPv4/IPv6 TUN, DNS hijack and loop prevention
  ---
  duration_ms: 1.2822
  type: 'test'
  ...
# Subtest: empty, disabled, expired and fixed-exit clients have no ordinary node grant
ok 12 - empty, disabled, expired and fixed-exit clients have no ordinary node grant
  ---
  duration_ms: 1.5711
  type: 'test'
  ...
# Subtest: retired protocols cannot be shared in any format
ok 13 - retired protocols cannot be shared in any format
  ---
  duration_ms: 1.7219
  type: 'test'
  ...
1..13
# tests 13
# suites 0
# pass 13
# fail 0
# cancelled 0
# skipped 0
# todo 0
# duration_ms 666.9836
```


## 前端构建（退出 0）

执行命令：
```text
cd frontend; npm run build
```

真实输出：
```text
> yuchen-panel-frontend@1.0.3-rc.1 build
> vue-tsc -b && vite build

vite v6.3.5 building for production...
transforming...
✓ 60 modules transformed.
rendering chunks...
computing gzip size...
dist/index.html                   0.66 kB │ gzip:  0.37 kB
dist/assets/index-D9XQUgiS.css   32.80 kB │ gzip:  7.27 kB
dist/assets/index-KRLGkQly.js   273.76 kB │ gzip: 89.06 kB
✓ built in 5.32s
```


## C24 回归：后端 1–5（含两个存量断言）

执行命令：
```text
cd backend; go test -v -run TestNodePresetDefaultsAndStoredSecurity ./internal/api
```

真实输出：
```text
=== RUN   TestNodePresetDefaultsAndStoredSecurity
=== RUN   TestNodePresetDefaultsAndStoredSecurity/C24_1_NewVLESS_UserConfirmedRealityDefault
    nodes_test.go:50: protocol=vless security=reality generated=reality
=== RUN   TestNodePresetDefaultsAndStoredSecurity/C24_2_NewTrojan_TLS
    nodes_test.go:50: protocol=trojan security=tls generated=tls
=== RUN   TestNodePresetDefaultsAndStoredSecurity/C24_3_NewVMess_None
    nodes_test.go:50: protocol=vmess security=none generated=none
=== RUN   TestNodePresetDefaultsAndStoredSecurity/C24_4_StoredRealityPreserved
    nodes_test.go:50: protocol=vless security=reality generated=reality
=== RUN   TestNodePresetDefaultsAndStoredSecurity/C24_5_StoredNonePreserved
    nodes_test.go:50: protocol=vless security=none generated=none
=== RUN   TestNodePresetDefaultsAndStoredSecurity/ExplicitNewVLESSNoneRemainsSelectable
    nodes_test.go:50: protocol=vless security=none generated=none
=== RUN   TestNodePresetDefaultsAndStoredSecurity/ExplicitVMessRealityCompatibility
    nodes_test.go:50: protocol=vmess security=none generated=none
--- PASS: TestNodePresetDefaultsAndStoredSecurity (0.30s)
    --- PASS: TestNodePresetDefaultsAndStoredSecurity/C24_1_NewVLESS_UserConfirmedRealityDefault (0.00s)
    --- PASS: TestNodePresetDefaultsAndStoredSecurity/C24_2_NewTrojan_TLS (0.00s)
    --- PASS: TestNodePresetDefaultsAndStoredSecurity/C24_3_NewVMess_None (0.00s)
    --- PASS: TestNodePresetDefaultsAndStoredSecurity/C24_4_StoredRealityPreserved (0.00s)
    --- PASS: TestNodePresetDefaultsAndStoredSecurity/C24_5_StoredNonePreserved (0.00s)
    --- PASS: TestNodePresetDefaultsAndStoredSecurity/ExplicitNewVLESSNoneRemainsSelectable (0.00s)
    --- PASS: TestNodePresetDefaultsAndStoredSecurity/ExplicitVMessRealityCompatibility (0.00s)
PASS
ok  	yuchen-panel/backend/internal/api	0.316s
```

C24 第 6/7 条三档映射、≤2 核建议，见前端输出 C24_6/C24_7；五条后端和两条前端全部通过，默认抗封锁遵循之前用户确认。

## 真实测试修正记录

首次专项验证未通过：新增 Go 测试错误地将 Clash 的 PROXY 组名当 JSON 字符串解析；前端旧短 ID 断言替换遗漏，且前端 VMess ps 尚未消费去重显示名。随后改为只解析代理的 quoted name，替换旧 ID 断言为精确数字列表，并修复 VMess ps 的实际命名路径；重跑通过。
首次失败完整日志保存在 .codex-release-inspect/task11-go-names-first-failure.txt 与 task11-frontend-first-failure.txt，未伪称首次全绿。
## CI 与既有功能回归

代码提交 20a7a22 的 PR CI 38593691697、push CI 38593680243 均 success。
命令：gh run list --repo sle58083/yuchen-panel --branch fix/task11-clean-node-names --limit 3
```text
completed	success	fix: remove short node ID display suffixes, preserve numeric deduplication	CI	fix/task11-clean-node-names	pull_request	38593691697	1m36s	2026-10-14T15:56:03Z
completed	success	fix(share): remove short ID labels and retain numeric name deduplication	CI	fix/task11-clean-node-names	push	38593680243	1m8s	2026-10-14T15:55:58Z
```

gh run view 38593691697 --repo sle58083/yuchen-panel --json status,conclusion,jobs（仅投影 job 结论）：
```json
{"conclusion":"success","jobs":[{"conclusion":"success","name":"Docker Compose builds","status":"completed"},{"conclusion":"success","name":"Go build, vet and race tests (agent)","status":"completed"},{"conclusion":"success","name":"Shell syntax and ShellCheck","status":"completed"},{"conclusion":"success","name":"Fresh Docker installer integration (Ubuntu 22.04)","status":"completed"},{"conclusion":"success","name":"Frontend build and typecheck","status":"completed"},{"conclusion":"success","name":"Go build, vet and race tests (backend)","status":"completed"}],"status":"completed"}
```

命令：gh run view 38593691697 --repo sle58083/yuchen-panel --job 115954825492 --log（筛选八项及真实 Nginx 输出）：
```text
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T15:56:30.9162229Z ^[[36;1mgrep -Fq E2E_PASS task8-e2e-result.txt^[[0m
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T15:57:29.5888392Z PASS 1: installer exit 0
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T15:57:29.5909239Z PASS 2: no ContainerConfig
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T15:57:29.5926588Z PASS 3: no missing-package error
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T15:57:29.5943701Z PASS 4: fresh source config validated
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T15:57:29.5964236Z PASS 5: generated secret at least 32 characters
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T15:57:29.6461479Z PASS 6: source-built container health version=1.0.3-rc.1
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T15:57:29.7589615Z PASS 7: Compose API running
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T15:57:29.7770577Z PASS 8: container secret equals runtime .env byte-for-byte
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T15:57:33.0379895Z NGINX_PASS: Docker host root API HTTP=200 {"service":"yuchen-panel-api","status":"ok","version":"1.0.3-rc.1"}
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T15:57:33.0665692Z NGINX_PASS: Docker host prefixed API HTTP=200 {"service":"yuchen-panel-api","status":"ok","version":"1.0.3-rc.1"}
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T15:57:33.0760039Z NGINX_PASS: Docker frontend page HTTP=200
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T15:57:33.0942055Z NGINX_PASS: sub/share invalid tokens reach API (404, no SPA fallback)
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T15:57:33.5670867Z NGINX_PASS: fast template root API HTTP=200 {"service":"yuchen-panel-api","status":"ok","version":"1.0.3-rc.1"}
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T15:57:33.5947701Z NGINX_PASS: fast template prefixed API HTTP=200 {"service":"yuchen-panel-api","status":"ok","version":"1.0.3-rc.1"}
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T15:57:33.5948811Z NGINX_PASS: isolated Docker and fast routing regression checks complete
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T15:57:33.6849135Z E2E_PASS: eight assertions; source-built Docker image=sha256:bcc45edc6dc1a1c2c0da04c9b659a295ce3c862cd79f3e09bea284dc827be8ae; complete redacted log=/home/runner/work/yuchen-panel/yuchen-panel/.audit/task8-e2e/docker-install.log
```

命令：gh run view 38593691697 --repo sle58083/yuchen-panel --job 115954825321 --log（筛选 Compose/通知/Doctor 输出）：
```text
Shell syntax and ShellCheck	Docker installer preflight regression tests	2026-10-14T15:56:21.6135918Z PASS: actual Compose v2 source validation without runtime .env (read-only)
Shell syntax and ShellCheck	Docker installer preflight regression tests	2026-10-14T15:56:21.6142344Z PASS: fresh source without .env validated; placeholders confined to child shell
Shell syntax and ShellCheck	Docker installer preflight regression tests	2026-10-14T15:56:21.6155888Z PASS: real env-file explicitly selected at validation/start; exported values excluded from startup
Shell syntax and ShellCheck	Docker installer preflight regression tests	2026-10-14T15:56:21.6244339Z PASS: Ubuntu package, official plugin fallback, v1 rejection and manual recovery guidance (mock commands)
Shell syntax and ShellCheck	JWT notice and Doctor mode regression tests	2026-10-14T15:56:21.6551814Z NOTICE_PASS: fresh install silent; upgrade/reinstall displays:
Shell syntax and ShellCheck	JWT notice and Doctor mode regression tests	2026-10-14T15:56:21.6965278Z [SKIP] yuchen-panel-api (systemd) - docker mode uses the API container
Shell syntax and ShellCheck	JWT notice and Doctor mode regression tests	2026-10-14T15:56:21.6966267Z [SKIP] yuchen-agent - docker mode does not include Agent; deploy separately for node connectivity
Shell syntax and ShellCheck	JWT notice and Doctor mode regression tests	2026-10-14T15:56:21.6966897Z [SKIP] xray - docker mode does not include Agent; deploy separately for node connectivity
Shell syntax and ShellCheck	JWT notice and Doctor mode regression tests	2026-10-14T15:56:21.6969095Z [SKIP] Xray config - no Agent deployed in docker mode
Shell syntax and ShellCheck	JWT notice and Doctor mode regression tests	2026-10-14T15:56:21.6971042Z [SKIP] Xray restart watch - no Agent/Xray integration configured
Shell syntax and ShellCheck	JWT notice and Doctor mode regression tests	2026-10-14T15:56:21.6971453Z Doctor result: PASS (0 warning(s), 5 skipped)
Shell syntax and ShellCheck	JWT notice and Doctor mode regression tests	2026-10-14T15:56:21.7849704Z DOCTOR_PASS: Docker skip counts, actual container/proxy failures and fast service checks (mock commands)
```

完整原始 CI 日志另留在 .codex-release-inspect/task11-ci-e2e.txt 和 task11-ci-shell.txt。
八项 e2e 全过，额外 Docker/fast Nginx 路由均为真实 HTTP 200；Compose 占位值隔离、包安装兼容、v1 拒绝、JWT 轮换通知和 Doctor SKIP 回归均通过。
BUG-E/F/G/H 相关代理配置/凭据及验证代码未修改，完整后端测试与竞态测试已通过；这些结果不等于重新完成所有用户设备/生产服务器的 T1–T6 实测。
