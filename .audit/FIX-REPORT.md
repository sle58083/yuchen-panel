# FIX-REPORT — Batch 1 安全修复 + C10 状态索引

2026-10-13 后续状态：C10 已完成代码与配置检查，实际 1.12.0/1.14.2 × Go/TS × mixed/TUN 共 8 次 check 退出 0；完整最新证据见 [CLIENT-FIX-REPORT.md](CLIENT-FIX-REPORT.md)。下方 Batch 1 原始证据保留，coverage 最新 backend 19.7%、Agent 12.6%。T1–T6 仍 MANUAL_REQUIRED，C5–C8 尚待实测后推进。

## 环境自检（真实输出）

Ubuntu 26.04.1 LTS / WSL2，原生 Linux compiler 与 Docker daemon 已验证。没有运行面板安装/卸载或 T1–T6 实机命令。

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

## 本阶段修复与独立提交

| 提交 | 审计 ID | 改动与兼容性 |
| --- | --- | --- |
| C1 | P0-001 | router/store/Compose 同批 fail-closed；凭据缺失或弱 JWT 会拒绝启动，空历史密码哈希禁用；根安装器在停服务前刷新弱 JWT。手动部署需先改 .env；登录 token 会在密钥轮换后失效。 |
| C2 | P1-008 | 所有 runner 字段通过返回 error 的单引号编码；拒绝 CR/LF/NUL；manifest 限制同 host HTTPS 重定向，ZIP 仅 HTTPS。 |
| C3 | P1-001、P2-005 | jq 真正解析 JSON，白名单校验 version/package/root/URL；mktemp 解压不删除旧目录；两套依赖清单加入 jq。 |
| C4 | P1-010、P1-011，P1-016 部分 | 普通节点仅显式授权；删除最后授权禁用受影响客户；正式移除 SS；远程 Agent HTTPS，loopback HTTP 可用；管理入口打印 SSH 隧道建议。 |

四个提交分别使用 Conventional Commits，并在 body 中给出兼容性与升级步骤。C1 不混入独立 P2 修复；账户检查是 P0 闭环的必要组成部分。

升级前必须：备份数据与 .env；手动/Docker 用户配置强随机 YC_JWT_SECRET（至少 32 字符）、显式管理员名与初始密码（至少 12 字符）；将历史已知弱密码主动改掉。旧客户必须明确重新绑定允许的节点；旧 SS 节点改为 VLESS/Reality。远程 Agent 在重启前配置 HTTPS，否则会明确拒绝启动。本阶段不合 main、不打 tag、不触发 release。

## 验证结果与实际覆盖率

- backend：go build / go vet / go test -race -coverprofile 均退出 0；整体覆盖率 1.1% → 14.6%。
- Agent：同一组检查均退出 0；整体覆盖率 0% → 12.6%。
- 前端：npm ci、2 项 Node 分享测试、npm run build、vue-tsc 均退出 0。
- 6 个 Shell 入口：bash -n 均退出 0；ShellCheck error 级检查退出 0。
- 默认 ShellCheck 退出 1，唯一诊断为 deploy/install.sh 的 SC2034 warning（循环变量未使用），并未伪装成默认检查全绿。
- Compose config：缺凭据拒绝，提供显式 fixture 凭据通过；没有启动面板容器。
- 恶意 runner 输入的完整脚本通过真实 bash -n；只执行赋值部分验证字符原样保留，命令替换没有运行。按用户确认，不再要求安全字面量文本中不能出现 $(id)。
- 恶意 manifest fixtures 不启动任何安装流程；版本/路径/HTTP 下载均被拒绝，原有 release sentinel 保留。

原始输出完整保存在 raw/，没有仅用“全部通过”替代日志。

**依赖告警未修复：** npm audit 报 7 项 high、0 critical。涉及 Vue、Vite 及 server-renderer、brace-expansion、nanoid、postcss、source-map-js；详见 raw/npm-audit.json。此次保持 lockfile，依赖升级单列工程批次，不自动使用 npm audit fix --force。

race 通过只证明本阶段测试覆盖的运行路径；已知 currentXrayVersionText 并发遍历缺锁（P1-007）尚未修复，不宣称全仓库竞态已消除。

## 未完成的 P1/P2

| ID | 问题 | 原因/后续批次 |
| --- | --- | --- |
| P1-002 | 支持 Debian/Ubuntu 却不检测架构，fast 发布与 Agent 强制 amd64 | 不属于 C1–C4，保留待后续专项修复 |
| P1-003 | Fresh 安装备份失败仍清空数据 | C7：fresh 数据备份门禁待数据安全批次 |
| P1-004 | 停旧服务和删除 Nginx 配置发生在完整预检之前，失败无运行时回滚 | C8：安装预检与回滚事务待稳定性批次 |
| P1-005 | Xray 下载失败可能被空 bash 命令掩盖，捆绑安装函数中的失败也可被条件上下文掩盖 | C7：Xray 固定版本下载与显式错误检查待数据安全批次 |
| P1-006 | CLI upgrade 命令省略 SHA256 并把远端清单字段直接拼入 Shell | C7：CLI 下载命令校验与编码待数据安全批次 |
| P2-001 | deploy/uninstall.sh 使用旧服务名，fast 面板与 Agent/Nginx 留在运行 | 不属于 C1–C4，保留待后续专项修复 |
| P2-002 | rsync --delete 会先删除 .env，write_env 的自定义变量保留逻辑失效 | 不属于 C1–C4，保留待后续专项修复 |
| P2-003 | 健康检查和重启失败被吞，仍打印成功 | 不属于 C1–C4，保留待后续专项修复 |
| P2-004 | systemd/env/Nginx 文本变量不转义，路径空格与配置特殊字符破坏生成文件 | 不属于 C1–C4，保留待后续专项修复 |
| P1-007 | 升级接口无锁遍历 Servers，与心跳并发写 map 可造成进程级 fatal | C5：并发压力与锁覆盖修复待实测后继续 |
| P1-009 | Agent 重启失败后磁盘哈希已更新，下一轮同步认为已完成，不再尝试恢复服务 | C8：Agent 回滚与重试待稳定性批次 |
| P1-012 | CLI reset-password/restore 与内存 Store 并发，修改可能被旧快照立即覆盖 | C8：CLI 离线操作待稳定性批次 |
| P2-006 | JWT 不核对账户状态/角色，改密码或禁用不能立即撤销旧 token | 部分：账户状态/角色校验已覆盖；改密撤销旧 JWT 的版本机制尚未实现 |
| P2-007 | CRUD 大量忽略 SaveLocked 错误，返回成功但重启丢配置 | 不属于 C1–C4，保留待后续专项修复 |
| P2-008 | 客户流量配额仅存储展示，未采集按客户用量或限额执行 | C10 已标注 UI 仅展示；限额执行按范围暂不接入 |
| P2-009 | 异步出口测试写回整个旧对象，会覆盖修改或复活已删除记录 | 不属于 C1–C4，保留待后续专项修复 |
| P2-010 | 升级任务没有真正互斥锁，状态文件非原子写，可能重复启动与误清理 | 不属于 C1–C4，保留待后续专项修复 |
| P2-011 | 节点端口冲突与引用检查不对称 | 不属于 C1–C4，保留待后续专项修复 |
| P2-012 | TLS 入站缺证书模型，Trojan none 与客户端强制 TLS 不一致；gRPC 默认已在 C10 对齐 | 不属于 C1–C4，保留待后续专项修复 |
| P2-013 | 固定目标中转却开启覆盖嗅探 | 不属于 C1–C4，保留待后续专项修复 |
| P2-014 | 自实现 HTTP body 解析不解码 chunked | 不属于 C1–C4，保留待后续专项修复 |
| P2-015 | 请求无body限额/服务器超时/登录速率限制，慢请求与密码计算可耗资源 | 不属于 C1–C4，保留待后续专项修复 |
| P2-016 | 低熵或无校验凭据被接受，编码损坏的密码哈希可接受空derived key | 不属于 C1–C4，保留待后续专项修复 |
| P2-017 | 缓存型状态与监测值误导 | 不属于 C1–C4，保留待后续专项修复 |
| P1-013 | 服务器页面生成的 Agent 安装命令仍要求旧版包，当前发布包不能直接使用 | C6：统一 Agent 安装包版本待稳定性批次 |
| P2-018 | SOCKS5密码永久保存在全账户共享的localStorage，退出登录不清除 | 不属于 C1–C4，保留待后续专项修复 |
| P2-019 | Reality watcher 强制推荐模板，会把编辑中的 Trojan 改为 VLESS | 不属于 C1–C4，保留待后续专项修复 |
| P2-020 | 切换服务器时 Host 不清空，节点/中转被保存到错误入口地址 | 不属于 C1–C4，保留待后续专项修复 |
| P2-024 | 部分列表加载/删除没有捕获错误，加载中与空数据混淆 | 不属于 C1–C4，保留待后续专项修复 |
| P2-025 | 已存书签只JSON.parse不校验数组，合法损坏值可使页面方法报错 | 不属于 C1–C4，保留待后续专项修复 |
| P1-014 | Health仍返回0.7.7.1，托管升级按目标1.x校验必然不匹配 | C6：统一 health/Agent/前端版本待稳定性批次 |
| P1-015 | Release先发布checksum到main再上传资产，且不统一更新版本/包URL | C9：发布流程待工程批次，当前禁止触发 release |
| P2-026 | 自定义安装目录与CLI/诊断/升级固定路径分叉 | 不属于 C1–C4，保留待后续专项修复 |
| P2-027 | 自定义API端口能安装，但doctor/体检/升级恢复固定检查8088 | C2 留有 TODO；自定义端口检查待稳定性批次 |
| P2-028 | 示例变量名/文件路径不匹配真实运行配置，照文档配置会被忽略 | 不属于 C1–C4，保留待后续专项修复 |
| P2-029 | Makefile本地前端与代理目标不一致，普通开发访问API会失败 | 不属于 C1–C4，保留待后续专项修复 |
| P1-016 | 默认公网管理页面为HTTP，远程Agent示例同样明文下发凭据与配置 | 部分：远程 Agent 已强制 HTTPS；无域名，公网管理 TLS 按要求暂缓 |
| P1-017 | 本地发布打包不排除.env/私钥/备份，Git忽略规则不保护rsync归档 | C7：发布白名单打包待数据安全批次 |
| P2-030 | Agent配置降为0644、reset替换DB按umask建文件，升级日志可能含明文凭据 | 不属于 C1–C4，保留待后续专项修复 |
| P2-031 | Docker前端忽略lockfile安装依赖，镜像/Action均使用可移动标签 | 不属于 C1–C4，保留待后续专项修复 |
| P2-032 | 客户更新不是文档宣称的merge，固定客户非法到期格式被静默当作长期 | 不属于 C1–C4，保留待后续专项修复 |
| P2-033 | 操作日志无保留上限，所有Agent同步都重写完整含日志DB并长时间持全局锁 | 不属于 C1–C4，保留待后续专项修复 |

## SKIPPED / 待续

- C5–C8：按任务时间线，等待用户实测后再决定稳定性批次。
- C9：CI/release/Docker/依赖工程门禁待续；本阶段真实本地检查输出已附。
- C10：已单独完成代码/配置检查；设备实测待用户，详见 CLIENT-FIX-REPORT.md。
- sing-box check / schema：C10 已通过自动配置验证；设备导入/实际连通未执行，不声称通过。
- VERSION/health/Agent 的旧版本漂移仍存在，留待 C6；本阶段不声称版本一致验收已通过。

## MANUAL_REQUIRED

尚未收到任何用户实机日志，各项均待用户执行：

- T1：测试机 sing-box 1.13.0 / 1.14.2 和 legacy 对照。
- T2：v2rayNG / Shadowrocket，特殊字符密码与真实 IPv6；2001:db8::1 仅可测试解析。
- T3：Clash Verge / Stash 等客户端的 10+ 节点订阅。
- T4：VLESS TCP/TLS/Reality/gRPC、VMess、Trojan 导入与连通。
- T5：iOS/Android 后台恢复。
- T6：SS 创建选项、旧节点迁移提示、订阅不含 SS。

## 原始检查输出

### baseline-backend.txt

```text
COMMAND: go build ./...
COMMAND: go vet ./...
COMMAND: go test -race -coverprofile=<output> ./...
	yuchen-panel/backend/cmd/server		coverage: 0.0% of statements
ok  	yuchen-panel/backend/internal/api	1.048s	coverage: 1.3% of statements
?   	yuchen-panel/backend/internal/model	[no test files]
	yuchen-panel/backend/internal/security		coverage: 0.0% of statements
	yuchen-panel/backend/internal/store		coverage: 0.0% of statements
	yuchen-panel/backend/internal/xray		coverage: 0.0% of statements
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
yuchen-panel/backend/internal/api/nodes.go:15:			nodes					0.0%
yuchen-panel/backend/internal/api/nodes.go:72:			nodeByID				0.0%
yuchen-panel/backend/internal/api/nodes.go:167:			defaultServerIDLocked			0.0%
yuchen-panel/backend/internal/api/nodes.go:190:			normalizeNodeLocked			0.0%
yuchen-panel/backend/internal/api/relays.go:20:			relays					0.0%
yuchen-panel/backend/internal/api/relays.go:56:			relayByID				0.0%
yuchen-panel/backend/internal/api/relays.go:109:		normalizeRelayLocked			0.0%
yuchen-panel/backend/internal/api/response.go:11:		writeJSON				0.0%
yuchen-panel/backend/internal/api/response.go:16:		readJSON				0.0%
yuchen-panel/backend/internal/api/response.go:20:		methodNotAllowed			0.0%
yuchen-panel/backend/internal/api/response.go:23:		clientIP				0.0%
yuchen-panel/backend/internal/api/router.go:24:			NewRouter				0.0%
yuchen-panel/backend/internal/api/router.go:66:			withAuth				0.0%
yuchen-panel/backend/internal/api/router.go:83:			currentClaims				0.0%
yuchen-panel/backend/internal/api/router.go:88:			cors					0.0%
yuchen-panel/backend/internal/api/router.go:100:		getenv					0.0%
yuchen-panel/backend/internal/api/servers.go:13:		servers					0.0%
yuchen-panel/backend/internal/api/servers.go:51:		serverByID				0.0%
yuchen-panel/backend/internal/api/subscription.go:19:		isShareableNodeProtocol			0.0%
yuchen-panel/backend/internal/api/subscription.go:27:		subscription				0.0%
yuchen-panel/backend/internal/api/subscription.go:114:		shortShare				0.0%
yuchen-panel/backend/internal/api/subscription.go:197:		yamlQuote				0.0%
yuchen-panel/backend/internal/api/subscription.go:205:		nodeHost				0.0%
yuchen-panel/backend/internal/api/subscription.go:209:		nodeShareName				0.0%
yuchen-panel/backend/internal/api/subscription.go:216:		relayShareName				0.0%
yuchen-panel/backend/internal/api/subscription.go:223:		buildClashProxyLines			0.0%
yuchen-panel/backend/internal/api/subscription.go:262:		buildClashMetaSubscription		0.0%
yuchen-panel/backend/internal/api/subscription.go:320:		buildNodeShareLink			0.0%
yuchen-panel/backend/internal/api/subscription.go:333:		buildVlessShareLink			0.0%
yuchen-panel/backend/internal/api/subscription.go:365:		buildRelayShareLink			0.0%
yuchen-panel/backend/internal/api/subscription.go:385:		valueOr					0.0%
yuchen-panel/backend/internal/api/subscription.go:393:		buildVmessShareLink			0.0%
yuchen-panel/backend/internal/api/subscription.go:431:		buildTrojanShareLink			0.0%
yuchen-panel/backend/internal/api/subscription.go:463:		buildSSShareLink			0.0%
yuchen-panel/backend/internal/api/subscription.go:471:		buildClashProxyLinesForNode		0.0%
yuchen-panel/backend/internal/api/subscription.go:525:		appendClashTransportAndTLS		0.0%
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
yuchen-panel/backend/internal/api/updates.go:30:		updateManifestURL			0.0%
yuchen-panel/backend/internal/api/updates.go:34:		updateStatus				0.0%
yuchen-panel/backend/internal/api/updates.go:57:		updateCheck				0.0%
yuchen-panel/backend/internal/api/updates.go:98:		updatePanelCommand			0.0%
yuchen-panel/backend/internal/api/updates.go:143:		updateXrayStatus			0.0%
yuchen-panel/backend/internal/api/updates.go:155:		fetchUpdateManifest			0.0%
yuchen-panel/backend/internal/api/updates.go:182:		currentXrayVersionText			0.0%
yuchen-panel/backend/internal/api/updates.go:191:		currentXrayVersionText			0.0%
yuchen-panel/backend/internal/api/updates.go:208:		updateMessage				0.0%
yuchen-panel/backend/internal/api/updates.go:218:		isRemoteVersionNewer			100.0%
yuchen-panel/backend/internal/api/updates.go:222:		comparePanelVersions			100.0%
yuchen-panel/backend/internal/api/updates.go:247:		numericVersionParts			84.0%
yuchen-panel/backend/internal/api/updates.go:283:		shellEscape				0.0%
yuchen-panel/backend/internal/api/updates.go:284:		shellQuoteSafe				0.0%
yuchen-panel/backend/internal/api/updates.go:285:		firstNonEmpty				0.0%
yuchen-panel/backend/internal/api/updates.go:294:		fileSHA256				0.0%
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
yuchen-panel/backend/internal/api/upgrade_tasks.go:459:		upgradePrechecks			0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:477:		commandExists				0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:479:		diskFreeCheck				0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:497:		precheckMessage				0.0%
yuchen-panel/backend/internal/api/upgrade_tasks.go:504:		tailFile				0.0%
yuchen-panel/backend/internal/security/jwt.go:21:		SignJWT					0.0%
yuchen-panel/backend/internal/security/jwt.go:30:		VerifyJWT				0.0%
yuchen-panel/backend/internal/security/jwt.go:55:		b64url					0.0%
yuchen-panel/backend/internal/security/jwt.go:56:		sign					0.0%
yuchen-panel/backend/internal/security/password.go:16:		HashPassword				0.0%
yuchen-panel/backend/internal/security/password.go:25:		VerifyPassword				0.0%
yuchen-panel/backend/internal/security/password.go:46:		b64					0.0%
yuchen-panel/backend/internal/security/password.go:48:		pbkdf2SHA256				0.0%
yuchen-panel/backend/internal/store/store.go:25:		Open					0.0%
yuchen-panel/backend/internal/store/store.go:36:		loadOrInit				0.0%
yuchen-panel/backend/internal/store/store.go:78:		newData					0.0%
yuchen-panel/backend/internal/store/store.go:91:		normalize				0.0%
yuchen-panel/backend/internal/store/store.go:130:		defaultNetworkPolicy			0.0%
yuchen-panel/backend/internal/store/store.go:147:		normalizeNetworkPolicy			0.0%
yuchen-panel/backend/internal/store/store.go:175:		SaveLocked				0.0%
yuchen-panel/backend/internal/store/store.go:188:		seedDefaultAdmin			0.0%
yuchen-panel/backend/internal/store/store.go:203:		seedLocalServer				0.0%
yuchen-panel/backend/internal/store/store.go:218:		ensureSingleModeLocalServer		0.0%
yuchen-panel/backend/internal/store/store.go:290:		sameServerEndpoint			0.0%
yuchen-panel/backend/internal/store/store.go:295:		isLoopbackEndpoint			0.0%
yuchen-panel/backend/internal/store/store.go:304:		shouldReplaceLocalEndpoint		0.0%
yuchen-panel/backend/internal/store/store.go:308:		pickLocalServer				0.0%
yuchen-panel/backend/internal/store/store.go:332:		EnsureSingleModeLocalServerLocked	0.0%
yuchen-panel/backend/internal/store/store.go:343:		getenv					0.0%
yuchen-panel/backend/internal/store/store.go:350:		FindAdminByUsername			0.0%
yuchen-panel/backend/internal/store/store.go:361:		UpdateAdminLogin			0.0%
yuchen-panel/backend/internal/store/store.go:374:		AddLog					0.0%
yuchen-panel/backend/internal/store/store.go:379:		NewID					0.0%
yuchen-panel/backend/internal/store/store.go:382:		NewToken				0.0%
yuchen-panel/backend/internal/store/store.go:384:		NewPassword				0.0%
yuchen-panel/backend/internal/store/store.go:385:		NewUUID					0.0%
yuchen-panel/backend/internal/store/store.go:392:		randHex					0.0%
yuchen-panel/backend/internal/store/store.go:400:		ChangeAdminPassword			0.0%
yuchen-panel/backend/internal/xray/config.go:16:		GenerateInboundConfig			0.0%
yuchen-panel/backend/internal/xray/config.go:22:		GenerateServerConfig			0.0%
yuchen-panel/backend/internal/xray/config.go:250:		makeRelayVlessClients			0.0%
yuchen-panel/backend/internal/xray/config.go:270:		makeInboundClients			0.0%
yuchen-panel/backend/internal/xray/config.go:303:		clientSortKey				0.0%
yuchen-panel/backend/internal/xray/config.go:313:		applyRealitySettings			0.0%
yuchen-panel/backend/internal/xray/config.go:334:		ConfigHash				0.0%
yuchen-panel/backend/internal/xray/config.go:340:		contains				0.0%
yuchen-panel/backend/internal/xray/config.go:349:		trimSlash				0.0%
yuchen-panel/backend/internal/xray/config.go:362:		dnsConfigFromPolicy			0.0%
yuchen-panel/backend/internal/xray/config.go:390:		applyNetworkPolicy			0.0%
yuchen-panel/backend/internal/xray/config.go:420:		normalizeNetworkPolicy			0.0%
yuchen-panel/backend/internal/xray/credentials.go:15:		TrojanClientPassword			0.0%
yuchen-panel/backend/internal/xray/credentials.go:26:		SSCredentials				0.0%
yuchen-panel/backend/internal/xray/reality.go:13:		NewRealityKeys				0.0%
yuchen-panel/backend/internal/xray/reality.go:21:		PublicKeyFromPrivate			0.0%
yuchen-panel/backend/internal/xray/reality.go:33:		NewRealityShortID			0.0%
yuchen-panel/backend/internal/xray/reality.go:41:		encodeRealityKey			0.0%
total:								(statements)				1.1%

```

### baseline-agent.txt

```text
COMMAND: go build ./...
COMMAND: go vet ./...
COMMAND: go test -race -coverprofile=<output> ./...
	yuchen-panel/agent/cmd/agent		coverage: 0.0% of statements
COMMAND: go tool cover -func=<profile>
yuchen-panel/agent/cmd/agent/main.go:65:	main			0.0%
yuchen-panel/agent/cmd/agent/main.go:117:	syncConfig		0.0%
yuchen-panel/agent/cmd/agent/main.go:126:	heartbeat		0.0%
yuchen-panel/agent/cmd/agent/main.go:131:	postJSON		0.0%
yuchen-panel/agent/cmd/agent/main.go:156:	applyXrayConfig		0.0%
yuchen-panel/agent/cmd/agent/main.go:191:	collectHeartbeat	0.0%
yuchen-panel/agent/cmd/agent/main.go:199:	runShell		0.0%
yuchen-panel/agent/cmd/agent/main.go:205:	xrayVersion		0.0%
yuchen-panel/agent/cmd/agent/main.go:213:	fileHash		0.0%
yuchen-panel/agent/cmd/agent/main.go:222:	loadAverage		0.0%
yuchen-panel/agent/cmd/agent/main.go:235:	memoryUsage		0.0%
yuchen-panel/agent/cmd/agent/main.go:256:	diskUsage		0.0%
yuchen-panel/agent/cmd/agent/main.go:269:	networkTotals		0.0%
yuchen-panel/agent/cmd/agent/main.go:296:	round2			0.0%
yuchen-panel/agent/cmd/agent/main.go:297:	getenv			0.0%
yuchen-panel/agent/cmd/agent/main.go:303:	getenvInt		0.0%
yuchen-panel/agent/cmd/agent/main.go:311:	getenvBool		0.0%
total:						(statements)		0.0%

```

### after-backend.txt

```text
COMMAND: go build ./...
COMMAND: go vet ./...
COMMAND: go test -race -coverprofile=<output> ./...
	yuchen-panel/backend/cmd/server		coverage: 0.0% of statements
ok  	yuchen-panel/backend/internal/api	2.415s	coverage: 11.9% of statements
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
yuchen-panel/backend/internal/api/subscription.go:19:		isShareableNodeProtocol			100.0%
yuchen-panel/backend/internal/api/subscription.go:27:		subscription				69.5%
yuchen-panel/backend/internal/api/subscription.go:107:		shortShare				62.7%
yuchen-panel/backend/internal/api/subscription.go:179:		yamlQuote				0.0%
yuchen-panel/backend/internal/api/subscription.go:187:		nodeHost				100.0%
yuchen-panel/backend/internal/api/subscription.go:191:		nodeShareName				0.0%
yuchen-panel/backend/internal/api/subscription.go:198:		relayShareName				0.0%
yuchen-panel/backend/internal/api/subscription.go:205:		buildClashProxyLines			0.0%
yuchen-panel/backend/internal/api/subscription.go:244:		buildClashMetaSubscription		0.0%
yuchen-panel/backend/internal/api/subscription.go:302:		buildNodeShareLink			50.0%
yuchen-panel/backend/internal/api/subscription.go:316:		buildVlessShareLink			55.0%
yuchen-panel/backend/internal/api/subscription.go:348:		buildRelayShareLink			84.6%
yuchen-panel/backend/internal/api/subscription.go:368:		valueOr					100.0%
yuchen-panel/backend/internal/api/subscription.go:376:		buildVmessShareLink			0.0%
yuchen-panel/backend/internal/api/subscription.go:414:		buildTrojanShareLink			0.0%
yuchen-panel/backend/internal/api/subscription.go:446:		buildClashProxyLinesForNode		0.0%
yuchen-panel/backend/internal/api/subscription.go:492:		appendClashTransportAndTLS		0.0%
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
total:								(statements)				14.6%

```

### after-agent.txt

```text
COMMAND: go build ./...
COMMAND: go vet ./...
COMMAND: go test -race -coverprofile=<output> ./...
ok  	yuchen-panel/agent/cmd/agent	1.035s	coverage: 12.6% of statements
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

### npm-ci.txt

```text

added 53 packages, and audited 54 packages in 42s

8 packages are looking for funding
  run `npm fund` for details

7 high severity vulnerabilities

To address issues that do not require attention, run:
  npm audit fix

To address all issues, run:
  npm audit fix --force

Run `npm audit` for details.
npm warn install-scripts 1 package has install scripts not yet covered by allowScripts:
npm warn install-scripts   esbuild@0.25.12 (postinstall: node install.js)
npm warn install-scripts
npm warn install-scripts Run `npm install-scripts ls` to review, or `npm install-scripts approve <pkg>` to allow.
npm notice
npm notice New major version of npm available! 11.19.0 -> 12.2.0
npm notice Changelog: https://github.com/npm/cli/releases/tag/v12.2.0
npm notice To update run: npm install -g npm@12.2.0
npm notice

```

### frontend-tests.txt

```text
✔ empty, disabled, expired and fixed-exit clients have no ordinary node grant (3.290401ms)
✔ retired protocols cannot be shared in any format (1.312301ms)
ℹ tests 2
ℹ suites 0
ℹ pass 2
ℹ fail 0
ℹ cancelled 0
ℹ skipped 0
ℹ todo 0
ℹ duration_ms 540.334651

```

### frontend-build.txt

```text

> yuchen-panel-frontend@0.7.7.1 build
> vue-tsc -b && vite build

vite v6.3.5 building for production...
transforming...
✓ 59 modules transformed.
rendering chunks...
computing gzip size...
dist/index.html                   0.66 kB │ gzip:  0.38 kB
dist/assets/index-D9XQUgiS.css   32.80 kB │ gzip:  7.25 kB
dist/assets/index-qoRWDX86.js   268.91 kB │ gzip: 87.77 kB
✓ built in 4.23s

```

### frontend-typecheck.txt

```text

```

### shell-syntax.txt

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

```

### shellcheck-errors.txt

```text
COMMAND: shellcheck --severity=error <six scripts>
EXIT_CODE: 0

```

### shellcheck-all.txt

```text
COMMAND: shellcheck <six scripts>

In deploy/install.sh line 549:
  for i in $(seq 1 90); do
  ^-^ SC2034 (warning): i appears unused. Verify use (or export if used externally).

For more information:
  https://www.shellcheck.net/wiki/SC2034 -- i appears unused. Verify use (or ...
EXIT_CODE: 1

```

### compose-validation.txt

```text
COMMAND: compose config without credentials (expected rejection)
error while interpolating services.yuchen-panel-api.environment.[]: required variable YC_JWT_SECRET is missing a value: Set YC_JWT_SECRET to a random secret of at least 32 characters
COMMAND: compose config with explicit fixture credentials
EXIT_CODE: 0

```
