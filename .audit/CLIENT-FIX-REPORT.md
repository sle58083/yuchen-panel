# CLIENT-FIX-REPORT — C10（待设备验收）

## 环境自检（实际输出）

Ubuntu 26.04.1 LTS / WSL2；Linux compiler、CGO/race 和 Linux Docker daemon 可用。测试依赖 Python3 + PyYAML 6.0.3、Bash、jq。未安装生产面板或启动任何代理核心。

```text

go version go1.27.2 linux/amd64
linux
amd64
1
11.19.0
11.19.0
GNU bash, version 5.3.9(1)-release (x86_64-pc-linux-gnu)
Copyright (C) 2025 Free Software Foundation, Inc.
License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>

This is free software; you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.
ShellCheck - shell script analysis tool
version: 0.11.0
license: GNU General Public License, version 3
website: https://www.shellcheck.net
GNU Make 4.4.1
Built for x86_64-pc-linux-gnu
Copyright (C) 1988-2023 Free Software Foundation, Inc.
License GPLv3+: GNU GPL version 3 or later <https://gnu.org/licenses/gpl.html>
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.
Docker version 29.1.3, build 29.1.3-0ubuntu4.1
Docker Compose version 2.40.3+ds1-0ubuntu1

=== Docker engine check ===
ServerVersion=29.1.3 OSType=linux

Environment self-check complete; source code and client checks remain pending.
```

## 提交范围与兼容性

C1–C4 已分别发布到 `fix/audit-batch1-security`，草稿 PR #2（未合并）：
https://github.com/sle58083/yuchen-panel/pull/2
C10 在其后单独提交到 `fix/audit-client-compat`，基于 C4 的 GitHub SHA `81ed942f02b87d3bc8a53e818a47d8c9426543c5`，以安全分支为 PR base。推送时校验远端 Git tree 与本地测试 tree 完全相同；不改写本地 C1–C4 历史。

- C10-1：Go 独立生成 sing-box 订阅，显式格式 switch、JSON Content-Type、节点数 + selector + direct；仅客户获授权、有效、未停用协议。
- C10-2 / P2-023：前后端采用 >=1.12 typed UDP DNS、default_domain_resolver；没有 legacy 导出 API。结构校验含拒绝未知字段、DNS/TLS/Reality/transport/selector 引用和旧 DNS 负对照；另做真实核心 check。
- C10-3 / P2-022：Trojan 密码、Unicode 名称、IPv6 URI 编码和解析回归。Trojan URI 的密码在 userinfo username 位置，所以 Go 使用 url.User(password)，与前端 encodeURIComponent 一致；不是给密码后加一个空冒号。
- C10-4：显式格式优先、UA 兜底、未知格式回退 base64；AddLog 只记格式/来源。编码和日志均在输入快照锁释放后进行。
- C10-5：格式选择器、客户端提示、>=1.12 提示、下载脚本和可导入样例。默认 mixed localhost:2080，可选 profile=tun 用于手机 VPN/桌面 TUN。
- P2-021：重复/保留代理名去重；空授权订阅保留合法 direct selector。
- P2-012（部分）：gRPC 默认服务名统一 yuchen；VMess 的 Clash network 恢复。证书模型/服务端 TLS 缺陷仍待修复。
- P2-008（部分）：额度 UI 明写“仅展示，当前不执行限额”，没有实施限额。
- 修正 C4 引入的 UI 到期判断：Go zero time 表示长期客户，前端不再按已过期过滤。未 amend C4，在 C10 中记录并回归测试；C4 单独分支仍有该 UI 问题。

升级：测试部署需更新本分支 API 和重新构建的 web，保留 C1–C4 凭据/授权迁移步骤，不需重新安装。**默认/未知格式从旧版明文 URI 改为 base64**；需要旧响应的工具用 format=raw。最低 sing-box 1.12，>=1.13 请勿手工编辑 DNS。TUN 导出不改变系统网络，真正运行需用户 VPN/管理员权限。没有 main 合并、tag、release 或发布包。

## 自动验证结果

| 检查 | 实际结果 |
| --- | --- |
| backend build / vet / race+cover | 退出 0；总覆盖率 1.1% 基线 → C4 14.6% → C10 19.7% |
| agent build / vet / race+cover | 退出 0；总覆盖率 0% → 12.6% |
| frontend node:test | 7/7，0 fail、0 skip |
| npm build / vue-tsc | 退出 0 |
| 七个脚本 bash -n | 各退出 0 |
| ShellCheck severity=error | 退出 0 |
| 完整 ShellCheck | 退出 1，仅既有 deploy/install.sh:549 SC2034 warning；未宣称零 warning |
| 新 verify-clients.sh 完整 ShellCheck | 退出 0 |
| 1.12.0 / 1.14.2 × Go/TS × mixed/TUN | 8 次 check，全部退出 0、无 FATAL；网络禁用/只读/无 capabilities |
| 本机 API 下载格式 | 12 URI，Clash YAML 3465 bytes，sing-box 14 outbounds；Windows loopback HTTP 200 |

核心镜像 RepoDigest：
- v1.12.0: sha256:3c1ee82d450df9b336b6f63b1d272cf86e3835392d99596ef8c35f206a457999
- v1.14.2: sha256:afbffd294c5eb3519cc7b4587299ef189bb0a2ca2f667cb6301fdb6b9bce9558

初次使用任务中的无 v 标签拉取返回 not found；已查官方包页更正，保留失败日志，最终使用官方 v 标签检查。check 成功没有 stdout 是正常的，只以真实退出码/错误判断，不伪造 configuration is valid 文案。所有 Docker 调用仅 check，未运行代理或改变宿主路由。

最终下载 JSON 与已 check 的 Go JSON 按字节比较相同，样例 hash/大小见 [sample-checksums.json](raw/client-compat/sample-checksums.json)。前端样例来自真实 share.ts 的 esbuild bundle，对同一合成 DB 生成后检查，不是手写 JSON。合成地址不可连通；全部文件只证明配置/解析，不能证明移动端、QoS 或吞吐量。

TUN 的 inbounds.address 是现代 CIDR 字段，与 DNS.servers.address legacy 完全不同；测试负对照和下载的 legacy-singbox.json 故意保留旧字段。不能用全局 grep address 判断 DNS 是否仍旧。

## 仍未修复 / SKIPPED

C5–C8 根据约定等待用户实测后推进。currentXrayVersionText 并发遍历 P1-007 和 Agent 回滚 P1-009 等已知路径未消除；race 通过不代表未覆盖路径不存在竞态。版本漂移 C6 / P1-013、P1-014 尚在，health/Agent/frontend package 的旧版本字符串仍会命中。C7 下载信任、数据备份、白名单发布及 C8 升级互斥/恢复未完成。C9 CI/release/Docker 工程批次未做，没有触发发布。

完整剩余 P1/P2 清单及原因见 [FIX-REPORT.md](FIX-REPORT.md)；C10 已完成 P2-021/022/023 的静态修复，P2-008/012 仍部分未完成。npm audit 的 7 high、0 critical 依赖告警仍未修复，保持原 lockfile，见 [npm-audit.json](raw/npm-audit.json)。公网管理 TLS 因无域名按要求暂缓，远程 Agent 仍要求 HTTPS。法律文件/SPDX 未修改。

## MANUAL_REQUIRED — 用户尚未提供实机日志

| 项 | 待用户执行 |
| --- | --- |
| T1 | 测试机 1.13.0 / 1.14.2 现代配置与 legacy 负对照；客户端实际运行 |
| T2 | v2rayNG / Shadowrocket 的特殊字符、真实 IPv6 导入和连接 |
| T3 | Clash Verge / Stash 等 12+ 节点数量/体积/导入表现 |
| T4 | 各协议真实节点连接，包括 TLS/Reality/gRPC |
| T5 | Android/iOS 切后台、恢复与刷新 |
| T6 | 测试面板 SS 选项删除、旧 SS banner、实际导出不含 SS |

所有项目都是 MANUAL_REQUIRED，未从 config check 推导设备测试结果。使用说明、下载命令和回贴模板见 [CLIENT-COMPAT.md](../docs/CLIENT-COMPAT.md)。本地订阅 `http://127.0.0.1:18088/sub/audit-demo-local` 已真实响应，进程结束后失效，不是公网服务。

## 原始输出（完整文件及以下原样嵌入）

### backend.txt

```text
COMMAND: go build -o <workspace work build dir>/ ./...
COMMAND: go vet ./...
COMMAND: go test -race -coverprofile=<output> ./...
	yuchen-panel/backend/cmd/server		coverage: 0.0% of statements
ok  	yuchen-panel/backend/internal/api	2.951s	coverage: 18.0% of statements
	yuchen-panel/backend/internal/model		coverage: 0.0% of statements
	yuchen-panel/backend/internal/security		coverage: 0.0% of statements
ok  	yuchen-panel/backend/internal/store	(cached)	coverage: 37.6% of statements
ok  	yuchen-panel/backend/internal/xray	(cached)	coverage: 32.1% of statements
COMMAND: go tool cover -func=<profile>
yuchen-panel/backend/cmd/server/main.go:13:			main					0.0%
yuchen-panel/backend/cmd/server/main.go:28:			getenv					0.0%
yuchen-panel/backend/internal/api/agent.go:13:			agentHeartbeat				0.0%
yuchen-panel/backend/internal/api/agent.go:45:			agentSync				0.0%
yuchen-panel/backend/internal/api/agent.go:120:			validateAgentToken			0.0%
yuchen-panel/backend/internal/api/auth.go:16:			login					0.0%
yuchen-panel/backend/internal/api/auth.go:45:			changePassword				0.0%
yuchen-panel/backend/internal/api/clients.go:15:		clients					0.0%
yuchen-panel/backend/internal/api/clients.go:67:		clientByID				0.0%
yuchen-panel/backend/internal/api/clients.go:149:		validateClientLocked			0.0%
yuchen-panel/backend/internal/api/clients.go:218:		createClientWithSocks5Relay		0.0%
yuchen-panel/backend/internal/api/dashboard.go:10:		dashboard				0.0%
yuchen-panel/backend/internal/api/dashboard.go:136:		dashboardTime				0.0%
yuchen-panel/backend/internal/api/health.go:6:			health					0.0%
yuchen-panel/backend/internal/api/landing_exits.go:20:		landingExits				0.0%
yuchen-panel/backend/internal/api/landing_exits.go:56:		landingExitByID				0.0%
yuchen-panel/backend/internal/api/landing_exits.go:118:		normalizeLandingExit			0.0%
yuchen-panel/backend/internal/api/landing_exits.go:144:		bulkLandingExits			0.0%
yuchen-panel/backend/internal/api/landing_exits.go:200:		splitBulkLine				0.0%
yuchen-panel/backend/internal/api/landing_exits.go:207:		cleanParts				0.0%
yuchen-panel/backend/internal/api/landing_exits.go:215:		testLandingExit				0.0%
yuchen-panel/backend/internal/api/logs.go:12:			logs					0.0%
yuchen-panel/backend/internal/api/network_policy.go:18:		networkPolicy				0.0%
yuchen-panel/backend/internal/api/network_policy.go:29:		networkPolicyPreview			0.0%
yuchen-panel/backend/internal/api/network_policy.go:51:		networkPolicyRollback			0.0%
yuchen-panel/backend/internal/api/network_policy.go:76:		networkPolicyStatus			0.0%
yuchen-panel/backend/internal/api/network_policy.go:100:	networkPolicySave			0.0%
yuchen-panel/backend/internal/api/network_policy.go:135:	normalizePolicyForAPI			0.0%
yuchen-panel/backend/internal/api/network_policy.go:223:	cleanDNSServers				0.0%
yuchen-panel/backend/internal/api/network_policy.go:251:	policySummary				0.0%
yuchen-panel/backend/internal/api/network_policy.go:272:	policyWarnings				0.0%
yuchen-panel/backend/internal/api/network_policy.go:289:	xrayDNSPreview				0.0%
yuchen-panel/backend/internal/api/network_policy.go:306:	routingPreview				0.0%
yuchen-panel/backend/internal/api/network_policy.go:323:	yesNo					0.0%
yuchen-panel/backend/internal/api/network_policy.go:330:	ipv6StrategyLabel			0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:54:		nodeDiagnosisRun			0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:63:		nodeDiagnosisReport			0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:74:		buildNodeDiagnosis			0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:310:	renderNodeDiagnosisReport		0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:341:	writeDiagnosisSection			0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:353:	diagnosisStatusLabel			0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:364:	serviceCheck				0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:381:	canTCPDial				0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:393:	checkManagedPort			0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:409:	readKVFile				0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:430:	firstExistingFile			0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:439:	firstNonEmptyDiag			0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:448:	findChinaDNS				0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:459:	ipv6Enabled				0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:467:	xrayDNSPolicyMessage			0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:479:	networkPolicyHasPublicDNSGuard		0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:497:	detectPublicIPv6			0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:521:	detectPublicIPv4			0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:545:	isLocalEndpoint				0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:550:	errString				0.0%
yuchen-panel/backend/internal/api/node_diagnosis.go:557:	uniqueStrings				0.0%
yuchen-panel/backend/internal/api/nodes.go:15:			nodes					29.6%
yuchen-panel/backend/internal/api/nodes.go:52:			nodeByID				33.3%
yuchen-panel/backend/internal/api/nodes.go:138:			defaultServerIDLocked			0.0%
yuchen-panel/backend/internal/api/nodes.go:161:			normalizeNodeLocked			28.6%
yuchen-panel/backend/internal/api/relays.go:20:			relays					0.0%
yuchen-panel/backend/internal/api/relays.go:56:			relayByID				47.5%
yuchen-panel/backend/internal/api/relays.go:112:		normalizeRelayLocked			0.0%
yuchen-panel/backend/internal/api/response.go:11:		writeJSON				100.0%
yuchen-panel/backend/internal/api/response.go:16:		readJSON				100.0%
yuchen-panel/backend/internal/api/response.go:20:		methodNotAllowed			0.0%
yuchen-panel/backend/internal/api/response.go:23:		clientIP				66.7%
yuchen-panel/backend/internal/api/router.go:24:			NewRouter				100.0%
yuchen-panel/backend/internal/api/router.go:70:			withAuth				77.8%
yuchen-panel/backend/internal/api/router.go:95:			currentClaims				100.0%
yuchen-panel/backend/internal/api/router.go:100:		cors					12.5%
yuchen-panel/backend/internal/api/router.go:112:		getenv					66.7%
yuchen-panel/backend/internal/api/servers.go:13:		servers					0.0%
yuchen-panel/backend/internal/api/servers.go:51:		serverByID				0.0%
yuchen-panel/backend/internal/api/singbox.go:9:			grpcService				100.0%
yuchen-panel/backend/internal/api/singbox.go:13:		addSingBoxTunProfile			100.0%
yuchen-panel/backend/internal/api/singbox.go:21:		buildSingBoxSubscription		90.0%
yuchen-panel/backend/internal/api/singbox.go:69:		singBoxNode				100.0%
yuchen-panel/backend/internal/api/subscription.go:21:		isShareableNodeProtocol			100.0%
yuchen-panel/backend/internal/api/subscription.go:29:		subscription				84.6%
yuchen-panel/backend/internal/api/subscription.go:83:		shortShare				62.7%
yuchen-panel/backend/internal/api/subscription.go:155:		yamlQuote				75.0%
yuchen-panel/backend/internal/api/subscription.go:163:		nodeHost				100.0%
yuchen-panel/backend/internal/api/subscription.go:167:		nodeShareName				100.0%
yuchen-panel/backend/internal/api/subscription.go:174:		relayShareName				0.0%
yuchen-panel/backend/internal/api/subscription.go:181:		buildClashProxyLines			55.2%
yuchen-panel/backend/internal/api/subscription.go:220:		buildClashMetaSubscription		84.4%
yuchen-panel/backend/internal/api/subscription.go:278:		buildNodeShareLink			83.3%
yuchen-panel/backend/internal/api/subscription.go:292:		buildVlessShareLink			57.9%
yuchen-panel/backend/internal/api/subscription.go:323:		buildRelayShareLink			83.3%
yuchen-panel/backend/internal/api/subscription.go:342:		valueOr					100.0%
yuchen-panel/backend/internal/api/subscription.go:350:		buildVmessShareLink			76.5%
yuchen-panel/backend/internal/api/subscription.go:388:		buildTrojanShareLink			55.0%
yuchen-panel/backend/internal/api/subscription.go:419:		buildClashProxyLinesForNode		58.8%
yuchen-panel/backend/internal/api/subscription.go:466:		appendClashTransportAndTLS		57.9%
yuchen-panel/backend/internal/api/subscription_formats.go:12:	subscriptionFormat			94.7%
yuchen-panel/backend/internal/api/subscription_formats.go:48:	subscriptionSnapshot			77.8%
yuchen-panel/backend/internal/api/subscription_formats.go:84:	uniqueSubscriptionNames			92.9%
yuchen-panel/backend/internal/api/system.go:27:			systemChecks				0.0%
yuchen-panel/backend/internal/api/system.go:42:			systemReport				0.0%
yuchen-panel/backend/internal/api/system.go:126:		buildSystemChecks			0.0%
yuchen-panel/backend/internal/api/system.go:427:		relayRouteMode				0.0%
yuchen-panel/backend/internal/api/system.go:435:		relayNetworkOrTCP			0.0%
yuchen-panel/backend/internal/api/system.go:443:		isVLESSRealityTCP			0.0%
yuchen-panel/backend/internal/api/system.go:448:		hasDNSProtection			0.0%
yuchen-panel/backend/internal/api/system.go:463:		hasSniffing				0.0%
yuchen-panel/backend/internal/api/system.go:481:		hasFreedomUseIPv4			0.0%
yuchen-panel/backend/internal/api/system.go:497:		socks5RouteBindingProblems		0.0%
yuchen-panel/backend/internal/api/system.go:570:		emptyDash				0.0%
yuchen-panel/backend/internal/api/system.go:577:		timeOrDash				0.0%
yuchen-panel/backend/internal/api/system.go:584:		isLoopbackForReport			0.0%
yuchen-panel/backend/internal/api/tools.go:32:			testSocks5				0.0%
yuchen-panel/backend/internal/api/tools.go:68:			socks5HTTPGet				0.0%
yuchen-panel/backend/internal/api/tools.go:94:			socks5Handshake				0.0%
yuchen-panel/backend/internal/api/tools.go:187:			socks5ReplyMessage			0.0%
yuchen-panel/backend/internal/api/tools.go:210:			parseHTTPBody				0.0%
yuchen-panel/backend/internal/api/tools.go:236:			trimForMessage				0.0%
yuchen-panel/backend/internal/api/updates.go:31:		updateManifestURL			0.0%
yuchen-panel/backend/internal/api/updates.go:35:		updateStatus				0.0%
yuchen-panel/backend/internal/api/updates.go:58:		updateCheck				0.0%
yuchen-panel/backend/internal/api/updates.go:99:		updatePanelCommand			0.0%
yuchen-panel/backend/internal/api/updates.go:149:		updateXrayStatus			0.0%
yuchen-panel/backend/internal/api/updates.go:161:		fetchUpdateManifest			82.6%
yuchen-panel/backend/internal/api/updates.go:193:		currentXrayVersionText			0.0%
yuchen-panel/backend/internal/api/updates.go:202:		currentXrayVersionText			0.0%
yuchen-panel/backend/internal/api/updates.go:219:		updateMessage				0.0%
yuchen-panel/backend/internal/api/updates.go:229:		isRemoteVersionNewer			100.0%
yuchen-panel/backend/internal/api/updates.go:233:		comparePanelVersions			100.0%
yuchen-panel/backend/internal/api/updates.go:258:		numericVersionParts			84.0%
yuchen-panel/backend/internal/api/updates.go:295:		shellEscape				100.0%
yuchen-panel/backend/internal/api/updates.go:302:		shellEscapeValues			100.0%
yuchen-panel/backend/internal/api/updates.go:314:		manifestRedirectPolicy			100.0%
yuchen-panel/backend/internal/api/updates.go:322:		firstNonEmpty				0.0%
yuchen-panel/backend/internal/api/updates.go:331:		fileSHA256				0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:39:		updateTaskLatest			0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:53:		updateTaskLogs				0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:68:		updateTaskPrecheck			0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:83:		updateTaskClearStale			0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:116:		updateTaskCreate			0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:208:		upgradeStatePath			0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:210:		readUpgradeTaskState			0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:222:		writeUpgradeTaskState			0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:235:		runningUpgradeTask			0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:244:		isActiveUpgradeStatus			0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:253:		normalizeUpgradeTaskState		0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:278:		versionMatchesTarget			0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:287:		managedUpgradeSupported			0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:297:		safePackageName				0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:304:		startUpgradeRunner			0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:320:		writeUpgradeRunner			0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:324:		writeUpgradeRunnerAt			78.6%
yuchen-panel/backend/internal/api/upgrade_tasks.go:475:		upgradePrechecks			0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:493:		commandExists				0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:495:		diskFreeCheck				0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:513:		precheckMessage				0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:520:		tailFile				0.0%
yuchen-panel/backend/internal/model/authorization.go:7:		AllowedNodes				0.0%
yuchen-panel/backend/internal/security/jwt.go:21:		SignJWT					0.0%
yuchen-panel/backend/internal/security/jwt.go:30:		VerifyJWT				0.0%
yuchen-panel/backend/internal/security/jwt.go:55:		b64url					0.0%
yuchen-panel/backend/internal/security/jwt.go:56:		sign					0.0%
yuchen-panel/backend/internal/security/password.go:16:		HashPassword				0.0%
yuchen-panel/backend/internal/security/password.go:25:		VerifyPassword				0.0%
yuchen-panel/backend/internal/security/password.go:46:		b64					0.0%
yuchen-panel/backend/internal/security/password.go:48:		pbkdf2SHA256				0.0%
yuchen-panel/backend/internal/store/store.go:26:		Open					83.3%
yuchen-panel/backend/internal/store/store.go:37:		loadOrInit				57.1%
yuchen-panel/backend/internal/store/store.go:79:		newData					100.0%
yuchen-panel/backend/internal/store/store.go:92:		normalize				66.7%
yuchen-panel/backend/internal/store/store.go:127:		defaultNetworkPolicy			100.0%
yuchen-panel/backend/internal/store/store.go:144:		normalizeNetworkPolicy			66.7%
yuchen-panel/backend/internal/store/store.go:172:		SaveLocked				75.0%
yuchen-panel/backend/internal/store/store.go:185:		seedDefaultAdmin			90.9%
yuchen-panel/backend/internal/store/store.go:203:		seedLocalServer				100.0%
yuchen-panel/backend/internal/store/store.go:218:		ensureSingleModeLocalServer		4.2%
yuchen-panel/backend/internal/store/store.go:290:		sameServerEndpoint			0.0%
yuchen-panel/backend/internal/store/store.go:295:		isLoopbackEndpoint			0.0%
yuchen-panel/backend/internal/store/store.go:304:		shouldReplaceLocalEndpoint		0.0%
yuchen-panel/backend/internal/store/store.go:308:		pickLocalServer				0.0%
yuchen-panel/backend/internal/store/store.go:332:		EnsureSingleModeLocalServerLocked	0.0%
yuchen-panel/backend/internal/store/store.go:343:		getenv					66.7%
yuchen-panel/backend/internal/store/store.go:350:		FindAdminByUsername			0.0%
yuchen-panel/backend/internal/store/store.go:361:		UpdateAdminLogin			0.0%
yuchen-panel/backend/internal/store/store.go:374:		AddLog					0.0%
yuchen-panel/backend/internal/store/store.go:379:		NewID					100.0%
yuchen-panel/backend/internal/store/store.go:382:		NewToken				100.0%
yuchen-panel/backend/internal/store/store.go:385:		NewPassword				0.0%
yuchen-panel/backend/internal/store/store.go:386:		NewUUID					0.0%
yuchen-panel/backend/internal/store/store.go:393:		randHex					75.0%
yuchen-panel/backend/internal/store/store.go:401:		ChangeAdminPassword			0.0%
yuchen-panel/backend/internal/xray/config.go:16:		GenerateInboundConfig			0.0%
yuchen-panel/backend/internal/xray/config.go:22:		GenerateServerConfig			29.2%
yuchen-panel/backend/internal/xray/config.go:245:		makeRelayVlessClients			0.0%
yuchen-panel/backend/internal/xray/config.go:265:		makeInboundClients			68.4%
yuchen-panel/backend/internal/xray/config.go:297:		clientSortKey				0.0%
yuchen-panel/backend/internal/xray/config.go:307:		applyRealitySettings			0.0%
yuchen-panel/backend/internal/xray/config.go:328:		ConfigHash				0.0%
yuchen-panel/backend/internal/xray/config.go:334:		contains				0.0%
yuchen-panel/backend/internal/xray/config.go:343:		trimSlash				0.0%
yuchen-panel/backend/internal/xray/config.go:356:		dnsConfigFromPolicy			88.9%
yuchen-panel/backend/internal/xray/config.go:384:		applyNetworkPolicy			57.1%
yuchen-panel/backend/internal/xray/config.go:414:		normalizeNetworkPolicy			36.8%
yuchen-panel/backend/internal/xray/credentials.go:15:		TrojanClientPassword			0.0%
yuchen-panel/backend/internal/xray/reality.go:13:		NewRealityKeys				0.0%
yuchen-panel/backend/internal/xray/reality.go:21:		PublicKeyFromPrivate			0.0%
yuchen-panel/backend/internal/xray/reality.go:33:		NewRealityShortID			0.0%
yuchen-panel/backend/internal/xray/reality.go:41:		encodeRealityKey			0.0%
total:								(statements)				19.7%

```

### agent.txt

```text
COMMAND: go build -o <workspace work build dir>/ ./...
COMMAND: go vet ./...
COMMAND: go test -race -coverprofile=<output> ./...
ok  	yuchen-panel/agent/cmd/agent	(cached)	coverage: 12.6% of statements
COMMAND: go tool cover -func=<profile>
yuchen-panel/agent/cmd/agent/main.go:67:	main			0.0%
yuchen-panel/agent/cmd/agent/main.go:122:	syncConfig		0.0%
yuchen-panel/agent/cmd/agent/main.go:131:	heartbeat		0.0%
yuchen-panel/agent/cmd/agent/main.go:136:	postJSON		52.2%
yuchen-panel/agent/cmd/agent/main.go:171:	validatePanelURL	100.0%
yuchen-panel/agent/cmd/agent/main.go:186:	applyXrayConfig		0.0%
yuchen-panel/agent/cmd/agent/main.go:221:	collectHeartbeat	0.0%
yuchen-panel/agent/cmd/agent/main.go:229:	runShell		0.0%
yuchen-panel/agent/cmd/agent/main.go:235:	xrayVersion		0.0%
yuchen-panel/agent/cmd/agent/main.go:243:	fileHash		0.0%
yuchen-panel/agent/cmd/agent/main.go:252:	loadAverage		0.0%
yuchen-panel/agent/cmd/agent/main.go:265:	memoryUsage		0.0%
yuchen-panel/agent/cmd/agent/main.go:286:	diskUsage		0.0%
yuchen-panel/agent/cmd/agent/main.go:299:	networkTotals		0.0%
yuchen-panel/agent/cmd/agent/main.go:326:	round2			0.0%
yuchen-panel/agent/cmd/agent/main.go:327:	getenv			0.0%
yuchen-panel/agent/cmd/agent/main.go:333:	getenvInt		0.0%
yuchen-panel/agent/cmd/agent/main.go:341:	getenvBool		0.0%
total:						(statements)		12.6%

```

### frontend.txt

```text
COMMAND: node --test tests/*.test.mjs (also export synthetic frontend fixtures)
✔ special password and IPv6 URI round-trip preserves endpoint and name (2.4692ms)
✔ VMess network and default WS/gRPC transport parameters survive export (5.8946ms)
✔ sing-box DNS is modern, selector choices valid, tags unique even when empty (0.7927ms)
✔ subscription format URLs are explicit and token path is encoded (0.482ms)
✔ optional VPN profile includes IPv4/IPv6 TUN, DNS hijack and loop prevention (0.593799ms)
✔ empty, disabled, expired and fixed-exit clients have no ordinary node grant (3.0847ms)
✔ retired protocols cannot be shared in any format (1.1465ms)
ℹ tests 7
ℹ suites 0
ℹ pass 7
ℹ fail 0
ℹ cancelled 0
ℹ skipped 0
ℹ todo 0
ℹ duration_ms 605.117328
COMMAND: npm run build

> yuchen-panel-frontend@0.7.7.1 build
> vue-tsc -b && vite build

vite v6.3.5 building for production...
transforming...
✓ 59 modules transformed.
rendering chunks...
computing gzip size...
dist/index.html                   0.66 kB │ gzip:  0.38 kB
dist/assets/index-D9XQUgiS.css   32.80 kB │ gzip:  7.25 kB
dist/assets/index-BFI-kDd3.js   272.39 kB │ gzip: 89.06 kB
✓ built in 6.00s
COMMAND: npx vue-tsc --noEmit
EXIT_CODE: 0

```

### shell.txt

```text
COMMAND: bash -n install.sh
EXIT_CODE: 0
COMMAND: bash -n deploy/install.sh
EXIT_CODE: 0
COMMAND: bash -n deploy/agent-install.sh
EXIT_CODE: 0
COMMAND: bash -n deploy/uninstall.sh
EXIT_CODE: 0
COMMAND: bash -n scripts/build-fast-release.sh
EXIT_CODE: 0
COMMAND: bash -n scripts/yuchen-panel
EXIT_CODE: 0
COMMAND: bash -n scripts/verify-clients.sh
EXIT_CODE: 0
COMMAND: shellcheck --severity=error <seven scripts>
EXIT_CODE: 0
COMMAND: shellcheck <seven scripts>

In deploy/install.sh line 549:
  for i in $(seq 1 90); do
  ^-^ SC2034 (warning): i appears unused. Verify use (or export if used externally).

For more information:
  https://www.shellcheck.net/wiki/SC2034 -- i appears unused. Verify use (or ...
EXIT_CODE: 1
COMMAND: shellcheck scripts/verify-clients.sh
EXIT_CODE: 0

```

### singbox-v1.12.0.txt

```text
COMMAND: docker pull ghcr.io/sagernet/sing-box:v1.12.0
v1.12.0: Pulling from sagernet/sing-box
Digest: sha256:3c1ee82d450df9b336b6f63b1d272cf86e3835392d99596ef8c35f206a457999
Status: Image is up to date for ghcr.io/sagernet/sing-box:v1.12.0
ghcr.io/sagernet/sing-box:v1.12.0
COMMAND: inspect image digest
ghcr.io/sagernet/sing-box@sha256:3c1ee82d450df9b336b6f63b1d272cf86e3835392d99596ef8c35f206a457999
COMMAND: docker run --rm --network none --read-only --cap-drop ALL -v <sample>:/c.json:ro ghcr.io/sagernet/sing-box:v1.12.0 check -c /c.json (file=singbox.json)
EXIT_CODE: 0
COMMAND: docker run --rm --network none --read-only --cap-drop ALL -v <sample>:/c.json:ro ghcr.io/sagernet/sing-box:v1.12.0 check -c /c.json (file=singbox-tun.json)
EXIT_CODE: 0
COMMAND: frontend-generated sample check; image=ghcr.io/sagernet/sing-box:v1.12.0 file=frontend-singbox.json
EXIT_CODE: 0
COMMAND: frontend-generated sample check; image=ghcr.io/sagernet/sing-box:v1.12.0 file=frontend-singbox-tun.json
EXIT_CODE: 0

```

### singbox-v1.14.2.txt

```text
COMMAND: docker pull ghcr.io/sagernet/sing-box:v1.14.2
v1.14.2: Pulling from sagernet/sing-box
Digest: sha256:afbffd294c5eb3519cc7b4587299ef189bb0a2ca2f667cb6301fdb6b9bce9558
Status: Image is up to date for ghcr.io/sagernet/sing-box:v1.14.2
ghcr.io/sagernet/sing-box:v1.14.2
COMMAND: inspect image digest
ghcr.io/sagernet/sing-box@sha256:afbffd294c5eb3519cc7b4587299ef189bb0a2ca2f667cb6301fdb6b9bce9558
COMMAND: docker run --rm --network none --read-only --cap-drop ALL -v <sample>:/c.json:ro ghcr.io/sagernet/sing-box:v1.14.2 check -c /c.json (file=singbox.json)
EXIT_CODE: 0
COMMAND: docker run --rm --network none --read-only --cap-drop ALL -v <sample>:/c.json:ro ghcr.io/sagernet/sing-box:v1.14.2 check -c /c.json (file=singbox-tun.json)
EXIT_CODE: 0
COMMAND: frontend-generated sample check; image=ghcr.io/sagernet/sing-box:v1.14.2 file=frontend-singbox.json
EXIT_CODE: 0
COMMAND: frontend-generated sample check; image=ghcr.io/sagernet/sing-box:v1.14.2 file=frontend-singbox-tun.json
EXIT_CODE: 0

```

### initial-pull-1.12.0.txt

```text
COMMAND: docker pull ghcr.io/sagernet/sing-box:1.12.0
Error response from daemon: failed to resolve reference "ghcr.io/sagernet/sing-box:1.12.0": ghcr.io/sagernet/sing-box:1.12.0: not found

```

### initial-pull-1.14.2.txt

```text
COMMAND: docker pull ghcr.io/sagernet/sing-box:1.14.2
Error response from daemon: failed to resolve reference "ghcr.io/sagernet/sing-box:1.14.2": ghcr.io/sagernet/sing-box:1.14.2: not found

```

### environment.txt

```text

go version go1.27.2 linux/amd64
linux
amd64
1
11.19.0
11.19.0
GNU bash, version 5.3.9(1)-release (x86_64-pc-linux-gnu)
Copyright (C) 2025 Free Software Foundation, Inc.
License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>

This is free software; you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.
ShellCheck - shell script analysis tool
version: 0.11.0
license: GNU General Public License, version 3
website: https://www.shellcheck.net
GNU Make 4.4.1
Built for x86_64-pc-linux-gnu
Copyright (C) 1988-2023 Free Software Foundation, Inc.
License GPLv3+: GNU GPL version 3 or later <https://gnu.org/licenses/gpl.html>
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.
Docker version 29.1.3, build 29.1.3-0ubuntu4.1
Docker Compose version 2.40.3+ds1-0ubuntu1

=== Docker engine check ===
ServerVersion=29.1.3 OSType=linux

Environment self-check complete; source code and client checks remain pending.

```


### loopback.txt

```text
COMMAND: Windows Invoke-WebRequest <public fixture subscription> ?format=singbox
HTTP_STATUS: 200
OUTBOUNDS: 14
DNS_TYPES: udp,udp
```
