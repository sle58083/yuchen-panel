# TASK13 — 包月限流量售卖

需求：时间与流量任一触发停用；延长到期即续费，归零并恢复。
分支 fix/task13-monthly-traffic；代码提交 f96f56a；PR https://github.com/sle58083/yuchen-panel/pull/8。
基线 main 8ace907（rc.2 发布后报告）。本轮不合并 main、不打 tag、不发布；当前一行安装仍取旧 rc.2 包，不含本 PR 功能。验收必须部署本分支源码构建的 API 与 Agent，不能只升级前端或重跑旧 fast 包。

## 四层实现

1. Xray：stats={}，level 0 启用 statsUserUplink/Downlink，api=StatsService；API 入站只监听 127.0.0.1，并把 api→api 路由置于策略规则前。默认端口 10000；若该服务器节点/中转占用它则顺序选空闲配置端口。Agent 从实际配置读取该 loopback 端口，不硬编码。
   普通 VLESS/VMess/Trojan 与固定出口 relay 的 email 都使用 client_<客户ID>，同名不共享统计身份。
2. Agent：调用已安装 Xray 的 `xray api statsquery --server=127.0.0.1:<port> -pattern user>>>client_ -reset=false`，CLI 自带 gRPC 客户端；无 grpcurl 安装和新 Go module。JSON int64 字符串/数字/省略的零值均支持，拒绝负数、重复、越界及不完整方向快照。只解析用户 ID 计数，整机 networkTotals 保留为机器展示，绝不用于客户计费。
   heartbeat 带 traffic_by_client、traffic_epoch、traffic_sample_at。epoch 取 boot ID + Xray PID + /proc 进程启动 ticks；Agent 重启不改变它，Xray/主机重启会改变。配置 apply 前先采样并成功上报；上报失败会推迟本次 apply。Stats 查询失败不伪造零计数，Agent 日志输出 Stats unavailable。
3. 落库：权威 TrafficUsedBytes=int64；保留旧 TrafficUsedGB 整数作兼容，旧非零 GB 可迁移至字节。UI 以字节换算小数显示，因此几十 MB 不再显示成整数 0 GB。
   客户持久化每服务器累计游标。相同 epoch 只计新累计值减去上次值；新 epoch 仅加入新进程首次计数。低于前值的单方向重置以新值计一次并更新游标，重复同值仍为零增量。拒绝陈旧时间戳、未知客户、非绑定服务器和负值；加法饱和防止溢出。保存失败回滚计数游标与客户状态，重试不会漏计/双计。
   任务中“total < TrafficUsedBytes 就加 total”的示例比较的是周期历史总量，重启后重复 30 会变成 330→360→390；本实现用源游标，测试实际证明 300→330→330→350，避免这种错误。
4. 执行限额：heartbeat 与 sync 在存储锁内检查到期和字节限额；limit=0 不限。达限额/到期置 Enabled=false；同一 Agent 紧接着 sync，其他 Agent 下次 sync 时重新生成配置、hash 改变、RestartRequired=true，移除客户凭据并重启 Xray。现有 ExpireAt 订阅与生成配置守卫未改（实际订阅拒绝码是 403，不把任务文字的 401 当实际输出）。

## 周期与 UI

- 新客户周期起点为创建时间；套餐天数默认 30，可在高级表单设置 1–366。
- PUT 只有 ExpireAt 延长时清零、BillingCycleStart=新到期-套餐天数、自动启用。保持旧累计游标，不把旧周期的绝对计数重新收费；拒绝续费前采样的迟到包。
- 备注、缩短到期、普通编辑不清零；客户端提交的用量/游标/周期起点不可信，更新时保留服务器已有计费数据。
- 客户列表每 30 秒仅刷新客户记录；编辑/创建窗口打开期间暂停；请求发出后若进入编辑也不覆盖；不重叠请求，卸载清理 timer。
- 显示「本周期用量 X / Y GB（周期起止）」。0 显示不限；前端原 `0 || 100` 已修复，后台固定出口原零值默认 100 也移除。无历史周期起点的旧客户显示“未记录”，不编造日期；下一次续费记录起点。

## 十五项断言对应

| # | 用例 | 真实测试 |
|---|---|---|
|1|stats/policy/api|TestStatsConfigAndUniqueClientEmail|
|2|email=client_ID，同名隔离|TestStatsConfigAndUniqueClientEmail / TestStatsRelayEmailUsesClientID|
|3|gRPC loopback|TestStatsConfigAndUniqueClientEmail / TestStatsAPIAddressLoopbackOnly|
|4|上传+下载字节与换算|TestTrafficNormalBytesAndConversion|
|5|Xray 重启补算|TestTrafficRestartAndRepeatedValueAreIdempotent|
|6|同值幂等、面板重启仍幂等|TestTrafficRestartAndRepeatedValueAreIdempotent / TestTrafficDuplicateHeartbeatPersistsAcrossPanelRestart|
|7|超限停用及下发|TestTrafficLimitStopsAndSyncRemovesClient|
|8|未超限不误停|TestTrafficLimitUnderAndUnlimited|
|9|0 不限|TestTrafficLimitUnderAndUnlimited / 前端 Traffic form retains explicit zero|
|10|到期停用/订阅守卫|TestExpireBillingStopsAndSubscriptionGuardRemains|
|11|延长到期归零/起点|TestBillingRenewalBoundaryCases/Billing11_extend_resets|
|12|停用客户续费恢复|TestBillingRenewalBoundaryCases/Billing12_disabled_renew_restores|
|13|改备注保留|TestBillingRenewalBoundaryCases/Billing13_remark_preserves|
|14|缩短到期保留|TestBillingRenewalBoundaryCases/Billing14_shorter_preserves|
|15|已过期续费恢复|TestBillingRenewalBoundaryCases/Billing15_expired_renew_restores|

另有多服务器/迟到样本、整机总量绝不计费、持久化失败回滚、续费后同绝对值不重计、端口碰撞、Agent epoch、partial/zero JSON 等断言。

## 后端专项完整输出（退出 0）

实际命令：
```text
cd backend; go test -v -run "Traffic|Stats|Limit|Expire|Billing" ./...（额外 YC_STATS_FIXTURE=绝对临时文件路径用于导出生成配置）
```

原始输出：
```text
?   	yuchen-panel/backend/cmd/server	[no test files]
=== RUN   TestTrafficPersistenceFailureRollsBackBeforeRetry
--- PASS: TestTrafficPersistenceFailureRollsBackBeforeRetry (0.26s)
=== RUN   TestTrafficNormalBytesAndConversion
    traffic_test.go:74: traffic_used_bytes=95226307 used_GB=0.088686
--- PASS: TestTrafficNormalBytesAndConversion (0.27s)
=== RUN   TestTrafficRestartAndRepeatedValueAreIdempotent
    traffic_test.go:93: restart: 300+30=330; repeat=330; subsequent increment=350
--- PASS: TestTrafficRestartAndRepeatedValueAreIdempotent (0.24s)
=== RUN   TestTrafficMachineTotalsNeverBillClients
--- PASS: TestTrafficMachineTotalsNeverBillClients (0.27s)
=== RUN   TestTrafficDuplicateHeartbeatPersistsAcrossPanelRestart
--- PASS: TestTrafficDuplicateHeartbeatPersistsAcrossPanelRestart (0.24s)
=== RUN   TestTrafficLimitStopsAndSyncRemovesClient
    traffic_test.go:152: Enabled=false; sync RestartRequired=true; client credential removed
--- PASS: TestTrafficLimitStopsAndSyncRemovesClient (0.27s)
=== RUN   TestTrafficLimitUnderAndUnlimited
--- PASS: TestTrafficLimitUnderAndUnlimited (0.53s)
=== RUN   TestExpireBillingStopsAndSubscriptionGuardRemains
--- PASS: TestExpireBillingStopsAndSubscriptionGuardRemains (0.31s)
=== RUN   TestBillingRenewalBoundaryCases
=== RUN   TestBillingRenewalBoundaryCases/Billing11_extend_resets
    traffic_test.go:237: used_bytes=0 enabled=true cycle_start=2026-10-15T20:26:28Z
=== RUN   TestBillingRenewalBoundaryCases/Billing12_disabled_renew_restores
    traffic_test.go:237: used_bytes=0 enabled=true cycle_start=2026-10-15T20:26:29Z
=== RUN   TestBillingRenewalBoundaryCases/Billing13_remark_preserves
    traffic_test.go:237: used_bytes=123456 enabled=true cycle_start=2026-09-14T20:26:29Z
=== RUN   TestBillingRenewalBoundaryCases/Billing14_shorter_preserves
    traffic_test.go:237: used_bytes=123456 enabled=true cycle_start=2026-09-14T20:26:29Z
=== RUN   TestBillingRenewalBoundaryCases/Billing15_expired_renew_restores
    traffic_test.go:237: used_bytes=0 enabled=true cycle_start=2026-10-15T20:26:30Z
--- PASS: TestBillingRenewalBoundaryCases (1.47s)
    --- PASS: TestBillingRenewalBoundaryCases/Billing11_extend_resets (0.31s)
    --- PASS: TestBillingRenewalBoundaryCases/Billing12_disabled_renew_restores (0.29s)
    --- PASS: TestBillingRenewalBoundaryCases/Billing13_remark_preserves (0.29s)
    --- PASS: TestBillingRenewalBoundaryCases/Billing14_shorter_preserves (0.30s)
    --- PASS: TestBillingRenewalBoundaryCases/Billing15_expired_renew_restores (0.29s)
=== RUN   TestTrafficMultiServerAndStaleSourceIsolation
--- PASS: TestTrafficMultiServerAndStaleSourceIsolation (0.32s)
=== RUN   TestBillingRenewalSameCumulativeValueDoesNotRebill
--- PASS: TestBillingRenewalSameCumulativeValueDoesNotRebill (0.27s)
PASS
ok  	yuchen-panel/backend/internal/api	4.471s
?   	yuchen-panel/backend/internal/model	[no test files]
?   	yuchen-panel/backend/internal/security	[no test files]
testing: warning: no tests to run
PASS
ok  	yuchen-panel/backend/internal/store	(cached) [no tests to run]
testing: warning: no tests to run
PASS
ok  	yuchen-panel/backend/internal/version	(cached) [no tests to run]
=== RUN   TestStatsConfigAndUniqueClientEmail
    stats_test.go:46: generated Stats config:
        {
          "api": {
            "services": [
              "StatsService"
            ],
            "tag": "api"
          },
          "dns": {
            "queryStrategy": "UseIPv4",
            "servers": [
              "1.1.1.1",
              "8.8.8.8",
              "9.9.9.9"
            ]
          },
          "inbounds": [
            {
              "listen": "0.0.0.0",
              "port": 24401,
              "protocol": "vless",
              "settings": {
                "clients": [
                  {
                    "email": "client_first",
                    "flow": "",
                    "id": "test-uuid-1"
                  },
                  {
                    "email": "client_second",
                    "flow": "",
                    "id": "test-uuid-2"
                  }
                ],
                "decryption": "none"
              },
              "sniffing": {
                "destOverride": [
                  "http",
                  "tls",
                  "quic"
                ],
                "enabled": true,
                "routeOnly": false
              },
              "streamSettings": {
                "network": "tcp",
                "security": "none"
              },
              "tag": "n"
            },
            {
              "listen": "127.0.0.1",
              "port": 10000,
              "protocol": "dokodemo-door",
              "settings": {
                "address": "127.0.0.1"
              },
              "tag": "api"
            }
          ],
          "log": {
            "loglevel": "warning"
          },
          "outbounds": [
            {
              "protocol": "freedom",
              "settings": {
                "domainStrategy": "UseIPv4"
              },
              "tag": "direct"
            },
            {
              "protocol": "blackhole",
              "tag": "blocked"
            }
          ],
          "policy": {
            "levels": {
              "0": {
                "statsUserDownlink": true,
                "statsUserUplink": true
              }
            }
          },
          "routing": {
            "domainStrategy": "IPIfNonMatch",
            "rules": [
              {
                "inboundTag": [
                  "api"
                ],
                "outboundTag": "api",
                "type": "field"
              }
            ]
          },
          "stats": {}
        }
--- PASS: TestStatsConfigAndUniqueClientEmail (0.00s)
=== RUN   TestStatsPortAvoidsNodeCollision
--- PASS: TestStatsPortAvoidsNodeCollision (0.00s)
=== RUN   TestStatsRelayEmailUsesClientID
--- PASS: TestStatsRelayEmailUsesClientID (0.00s)
PASS
ok  	yuchen-panel/backend/internal/xray	(cached)
```


## Agent 专项与本地真实 gRPC 输出（退出 0）

实际命令：
```text
cd agent; YC_XRAY_TEST_BINARY=<已核验官方 Xray 绝对路径> go test -v -run "Traffic|Stats|Limit|Expire|Billing" ./...
```

原始输出：
```text
=== RUN   TestStatsTrafficParserAcceptsProtoStringsAndNumbers
    traffic_test.go:21: traffic_by_client=map[id1:{Uplink:4030040 Downlink:91196267}] (machine/username counters ignored)
--- PASS: TestStatsTrafficParserAcceptsProtoStringsAndNumbers (0.00s)
=== RUN   TestStatsDefaultZeroValueInProtoJSON
--- PASS: TestStatsDefaultZeroValueInProtoJSON (0.00s)
=== RUN   TestStatsCollectorWithRealCore
    traffic_test.go:68: real local StatsService queried on 127.0.0.1:46051; process epoch identified; no proxy traffic generated
--- PASS: TestStatsCollectorWithRealCore (1.39s)
=== RUN   TestStatsEpochSurvivesAgentRestartButDetectsXrayRestart
--- PASS: TestStatsEpochSurvivesAgentRestartButDetectsXrayRestart (0.00s)
=== RUN   TestStatsTrafficParserRejectsInvalidCounters
--- PASS: TestStatsTrafficParserRejectsInvalidCounters (0.00s)
=== RUN   TestStatsAPIAddressLoopbackOnly
--- PASS: TestStatsAPIAddressLoopbackOnly (0.00s)
=== RUN   TestStatsPartialDirectionDoesNotInventCounterReset
--- PASS: TestStatsPartialDirectionDoesNotInventCounterReset (0.00s)
PASS
ok  	yuchen-panel/agent/cmd/agent	1.400s
```


## 后端完整 race（退出 0）

实际命令：
```text
cd backend; go test -race ./...
```

原始输出：
```text
?   	yuchen-panel/backend/cmd/server	[no test files]
ok  	yuchen-panel/backend/internal/api	26.128s
?   	yuchen-panel/backend/internal/model	[no test files]
?   	yuchen-panel/backend/internal/security	[no test files]
ok  	yuchen-panel/backend/internal/store	(cached)
ok  	yuchen-panel/backend/internal/version	(cached)
ok  	yuchen-panel/backend/internal/xray	(cached)
```


## Agent 完整 race（退出 0）

实际命令：
```text
cd agent; go test -race ./...
```

原始输出：
```text
ok  	yuchen-panel/agent/cmd/agent	1.085s
```

常规 race 与 CI 未设置 YC_XRAY_TEST_BINARY，所以可选 TestStatsCollectorWithRealCore 在这些运行中 SKIP；上述本地专项明确设置已核验的二进制并 PASS。不会把整体 job success 等同于该可选测试在 CI 执行。

## 前端完整输出（16 项，退出 0）

实际命令：
```text
cd frontend; node --test tests/*.test.mjs
```

原始输出：
```text
TAP version 13
# Subtest: Traffic polling uses 30 seconds, pauses during editing, rejects late responses, and cleans up
ok 1 - Traffic polling uses 30 seconds, pauses during editing, rejects late responses, and cleans up
  ---
  duration_ms: 10.4143
  type: 'test'
  ...
# Subtest: Traffic display uses byte precision and marks zero limit unlimited
ok 2 - Traffic display uses byte precision and marks zero limit unlimited
  ---
  duration_ms: 34.2623
  type: 'test'
  ...
# Subtest: Traffic form retains explicit zero and rejects invalid limits
ok 3 - Traffic form retains explicit zero and rejects invalid limits
  ---
  duration_ms: 2.1482
  type: 'test'
  ...
# Subtest: C24_6_three_preset_parameter_mappings
ok 4 - C24_6_three_preset_parameter_mappings
  ---
  duration_ms: 7.2799
  type: 'test'
  ...
# Subtest: C24_7_two_or_fewer_cores_suggest_fast_without_overriding_default
ok 5 - C24_7_two_or_fewer_cores_suggest_fast_without_overriding_default
  ---
  duration_ms: 0.9218
  type: 'test'
  ...
# TASK11 exported names: {"tags":["美国","日本","韩国"],"names":["美国","日本","韩国"],"fragments":["美国","日本","韩国"],"ps":["美国","日本","韩国"]}
# TASK11 exported names: {"tags":["美国","美国 2","美国 3"],"names":["美国","美国 2","美国 3"],"fragments":["美国","美国 2","美国 3"],"ps":["美国","美国 2","美国 3"]}
# TASK11 exported names: {"tags":["美国","example.com:23456","example.com:23456 2"],"names":["美国","example.com:23456","example.com:23456 2"],"fragments":["美国","example.com:23456","example.com:23456 2"],"ps":["美国","example.com:23456","example.com:23456 2"]}
# Subtest: special password and IPv6 URI round-trip preserves endpoint and name
ok 6 - special password and IPv6 URI round-trip preserves endpoint and name
  ---
  duration_ms: 5.772
  type: 'test'
  ...
# Subtest: display names use numeric duplicate suffixes without short ids
ok 7 - display names use numeric duplicate suffixes without short ids
  ---
  duration_ms: 5.5939
  type: 'test'
  ...
# Subtest: display names do not vary with client username
ok 8 - display names do not vary with client username
  ---
  duration_ms: 0.6903
  type: 'test'
  ...
# Subtest: clean original names and numeric duplicates agree across all four export paths
ok 9 - clean original names and numeric duplicates agree across all four export paths
  ---
  duration_ms: 10.4478
  type: 'test'
  ...
# Subtest: unnamed node uses host and actual port rather than remark or ID
ok 10 - unnamed node uses host and actual port rather than remark or ID
  ---
  duration_ms: 0.8
  type: 'test'
  ...
# Subtest: VMess network and default WS/gRPC transport parameters survive export
ok 11 - VMess network and default WS/gRPC transport parameters survive export
  ---
  duration_ms: 2.2413
  type: 'test'
  ...
# Subtest: sing-box DNS is modern, selector choices valid, tags unique even when empty
ok 12 - sing-box DNS is modern, selector choices valid, tags unique even when empty
  ---
  duration_ms: 0.9954
  type: 'test'
  ...
# Subtest: subscription format URLs are explicit and token path is encoded
ok 13 - subscription format URLs are explicit and token path is encoded
  ---
  duration_ms: 0.5178
  type: 'test'
  ...
# Subtest: optional VPN profile includes IPv4/IPv6 TUN, DNS hijack and loop prevention
ok 14 - optional VPN profile includes IPv4/IPv6 TUN, DNS hijack and loop prevention
  ---
  duration_ms: 1.1406
  type: 'test'
  ...
# Subtest: empty, disabled, expired and fixed-exit clients have no ordinary node grant
ok 15 - empty, disabled, expired and fixed-exit clients have no ordinary node grant
  ---
  duration_ms: 1.593
  type: 'test'
  ...
# Subtest: retired protocols cannot be shared in any format
ok 16 - retired protocols cannot be shared in any format
  ---
  duration_ms: 1.7239
  type: 'test'
  ...
1..16
# tests 16
# suites 0
# pass 16
# fail 0
# cancelled 0
# skipped 0
# todo 0
# duration_ms 742.7071
```


## 前端构建（退出 0）

实际命令：
```text
cd frontend; npm run build
```

原始输出：
```text
> yuchen-panel-frontend@1.0.3-rc.2 build
> vue-tsc -b && vite build

vite v6.3.5 building for production...
transforming...
✓ 60 modules transformed.
rendering chunks...
computing gzip size...
dist/index.html                   0.66 kB │ gzip:  0.37 kB
dist/assets/index-D9XQUgiS.css   32.80 kB │ gzip:  7.27 kB
dist/assets/index-G06w9aTC.js   275.19 kB │ gzip: 89.61 kB
✓ built in 4.91s
```


## 官方核心来源与校验

实际命令：
```text
gh api repos/XTLS/Xray-core/releases/latest；下载 v26.3.27 的 Xray-linux-64.zip，用 Get-FileHash SHA256 与官方资产 digest 核对
```

原始输出：
```text
Official XTLS/Xray-core v26.3.27 verified SHA256=23cd9af937744d97776ee35ecad4972cf4b2109d1e0fe6be9930467608f7c8ae
```


官方真实配置校验（退出 0，未启动代理）

实际命令：
```text
.codex-release-inspect/task13-core/extracted/xray run -test -config .codex-release-inspect/task13-xray-config.json
```

原始输出：
```text
Xray 26.3.27 (Xray, Penetrates Everything.) d2758a0 (go1.26.1 linux/amd64)
A unified platform for anti-censorship.
2026/10/15 04:12:07.774146 [Info] infra/conf/serial: Reading config: &{Name:.codex-release-inspect/task13-xray-config.json Format:json}
Configuration OK.
```


## 实际生成的完整 Xray 配置（仅合成凭据）

实际命令：
```text
TestStatsConfigAndUniqueClientEmail 生成，YC_STATS_FIXTURE 写入样例；此文件已由官方核心 check
```

原始输出：
```text
{
  "api": {
    "services": [
      "StatsService"
    ],
    "tag": "api"
  },
  "dns": {
    "queryStrategy": "UseIPv4",
    "servers": [
      "1.1.1.1",
      "8.8.8.8",
      "9.9.9.9"
    ]
  },
  "inbounds": [
    {
      "listen": "0.0.0.0",
      "port": 24401,
      "protocol": "vless",
      "settings": {
        "clients": [
          {
            "email": "client_first",
            "flow": "",
            "id": "test-uuid-1"
          },
          {
            "email": "client_second",
            "flow": "",
            "id": "test-uuid-2"
          }
        ],
        "decryption": "none"
      },
      "sniffing": {
        "destOverride": [
          "http",
          "tls",
          "quic"
        ],
        "enabled": true,
        "routeOnly": false
      },
      "streamSettings": {
        "network": "tcp",
        "security": "none"
      },
      "tag": "n"
    },
    {
      "listen": "127.0.0.1",
      "port": 10000,
      "protocol": "dokodemo-door",
      "settings": {
        "address": "127.0.0.1"
      },
      "tag": "api"
    }
  ],
  "log": {
    "loglevel": "warning"
  },
  "outbounds": [
    {
      "protocol": "freedom",
      "settings": {
        "domainStrategy": "UseIPv4"
      },
      "tag": "direct"
    },
    {
      "protocol": "blackhole",
      "tag": "blocked"
    }
  ],
  "policy": {
    "levels": {
      "0": {
        "statsUserDownlink": true,
        "statsUserUplink": true
      }
    }
  },
  "routing": {
    "domainStrategy": "IPIfNonMatch",
    "rules": [
      {
        "inboundTag": [
          "api"
        ],
        "outboundTag": "api",
        "type": "field"
      }
    ]
  },
  "stats": {}
}
```

## 真机验证记录

**SKIPPED**：本轮未在目标 ECS 38.246.250.89 部署新 API/Agent，未具备该 VPS 的可用执行环境/登录凭据，没有生成真实客户流量，也没有取得面板非零用量截图/接口记录。
用户/协调者提供的 Ubuntu 22.04.5、rc.2、整机上传 4030040 / 下载 91196267 / traffic_by_client=null 是任务基线，本轮未独立重测。
测试中 95226307 bytes / 0.088686 GB 是合成 heartbeat 的验证值，虽然采用与基线相同的数字，但不是把整机流量归给真实客户、不是部署证据。
本地真实 Xray 校验与 StatsService gRPC 查询只证明配置和采集接口可用，没有代理流量，不代替真机收费与断连验收。

部署此 PR 源码构建的新版 API/Agent 后，服务器可读取这些不含凭据的摘要：
```bash
sudo jq '{stats,policy,api,stats_listener:[.inbounds[]|select(.tag=="api")|{listen,port,protocol,tag}]}' /etc/yuchen-panel/xray/config.json
STATS_PORT=$(sudo jq -r '.inbounds[]|select(.tag=="api" and .listen=="127.0.0.1")|.port' /etc/yuchen-panel/xray/config.json)
xray api statsquery --server="127.0.0.1:$STATS_PORT" -pattern 'user>>>client_' -reset=false
journalctl -u yuchen-agent --since '5 minutes ago' --no-pager | rg 'Stats unavailable|heartbeat failed|config applied'
```
不要用 -reset=true。管理员用正常客户端产生流量后，在列表等待一次 Agent 上报和刷新，记录该客户 traffic_used_bytes、billing_cycle_start、expire_at、enabled；提交时仅保留客户匿名编号和这些字段，不发送密码/UUID/token。
验收重启、重复上报、续费和超限需由协调者在部署后实测；本报告未伪造这些输出。

## 计量精度与执行边界

- 采用成功采样的累计水位，不是逐字节硬配额；默认 30 秒，限额发现与其他服务器撤销有同步延迟，会有一个采样窗口的超额。
- 保存成功后已结算用量与水位持久化，重启不会清零历史或把同值翻倍。异常 kill/断电时，自最后成功采样到退出之间的未观察字节无法恢复；正常配置 apply 前会先采样上报，但最后采样到进程停止的短尾部仍可能漏计。没有声称全程无损计费。
- 续费基于最后成功的源水位；跨续费边界的一个采样窗口不能按包精确切分。此前没有任何水位的源，续费后第一份样本仅建立基线，避免对未知旧累计重复收费；这一情况会少计首次样本中无法归期的部分。Agent/API 时钟应同步。
- Stats 查询不可用会在 Agent 日志报错，不把整机指标充当计费；旧 Agent 不上报按客户统计，不能据此宣称计费验收通过。
- 旧核心在 Stats 未启用时期的流量无法按客户追溯，不将整机计数回填给客户。所有操作仍需要可用 Agent 成功同步；离线时不能保证远程即时撤销已经缓存的配置。

## 包月限流量售卖

### 计费周期
- 到期时间与流量上限任一触发即停用。
- 续费（延长到期时间）后流量自动归零，进入新周期。
- 改备注、提前到期不会清零流量。
- 此前因超限停用的客户续费后自动恢复启用。
- 默认套餐 30 天，周期起点记录为新到期减套餐天数；延长到期会立即清零，不等待旧到期时间。

### 流量统计
- 按客户 ID 统计，依赖新版 Agent 与 Xray Stats API。
- 已结算的累计用量不会因 Xray 重启清零；完整精度边界见上节。
- 达上限自动停用；不限流量填写 0。
- 计费口径为上传+下载，界面 GB 按 2^30 bytes 换算，与后台限额位移规则一致。

### 到期处理
- 原有到期订阅拒绝逻辑保持；Agent heartbeat/sync 自动置停用并撤销 Xray 客户。
- 停用依赖后续配置同步成功。紧急撤销应核实 Agent 配置已应用，不能只看到订阅拒绝就声称缓存连接已断开。

## 诚实性和范围

首次专项测试的计费断言已通过，但样例导出使用相对路径造成文件写入失败；原始日志保存在 .codex-release-inspect/task13-first-export-failure.txt。改绝对路径后重跑通过，未删计费断言。
Go mod/sum、VERSION、version.json 无修改；新增命令仅用既有 Xray CLI 和 Go 标准库。没有合并 main 或触发 Release。
## CI 真实结果

代码 f96f56a 的 PR CI 38633302425 与 push CI 38633292190 均 success。

```text
completed	success	feat: renewable monthly traffic quotas with Xray per-client stats	CI	fix/task13-monthly-traffic	pull_request	38633302425	2m0s	2026-10-14T20:35:05Z
completed	success	feat(billing): meter per-client traffic and enforce renewable monthly…	CI	fix/task13-monthly-traffic	push	38633292190	1m15s	2026-10-14T20:35:01Z
```

```json
{"conclusion":"success","jobs":[{"conclusion":"success","id":116090080597,"name":"Fresh Docker installer integration (Ubuntu 22.04)","status":"completed"},{"conclusion":"success","id":116090080958,"name":"Shell syntax and ShellCheck","status":"completed"},{"conclusion":"success","id":116090081023,"name":"Go build, vet and race tests (backend)","status":"completed"},{"conclusion":"success","id":116090081129,"name":"Docker Compose builds","status":"completed"},{"conclusion":"success","id":116090081220,"name":"Frontend build and typecheck","status":"completed"},{"conclusion":"success","id":116090081317,"name":"Go build, vet and race tests (agent)","status":"completed"}],"status":"completed"}
```

命令 gh run view 38633302425 --repo sle58083/yuchen-panel --job 116090080597 --log，筛选真实 e2e 断言输出：
```text
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T20:35:16.7305429Z ^[[36;1mgrep -Fq E2E_PASS task8-e2e-result.txt^[[0m
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T20:36:03.4552190Z PASS 1: installer exit 0
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T20:36:03.4586827Z PASS 2: no ContainerConfig
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T20:36:03.4601782Z PASS 3: no missing-package error
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T20:36:03.4612343Z PASS 4: fresh source config validated
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T20:36:03.4627502Z PASS 5: generated secret at least 32 characters
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T20:36:03.5097530Z PASS 6: source-built container health version=1.0.3-rc.2
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T20:36:03.6062491Z PASS 7: Compose API running
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T20:36:03.6220495Z PASS 8: container secret equals runtime .env byte-for-byte
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T20:36:06.8166153Z NGINX_PASS: Docker host root API HTTP=200 {"service":"yuchen-panel-api","status":"ok","version":"1.0.3-rc.2"}
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T20:36:06.8448543Z NGINX_PASS: Docker host prefixed API HTTP=200 {"service":"yuchen-panel-api","status":"ok","version":"1.0.3-rc.2"}
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T20:36:06.8528905Z NGINX_PASS: Docker frontend page HTTP=200
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T20:36:06.8684278Z NGINX_PASS: sub/share invalid tokens reach API (404, no SPA fallback)
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T20:36:07.3378473Z NGINX_PASS: fast template root API HTTP=200 {"service":"yuchen-panel-api","status":"ok","version":"1.0.3-rc.2"}
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T20:36:07.3654215Z NGINX_PASS: fast template prefixed API HTTP=200 {"service":"yuchen-panel-api","status":"ok","version":"1.0.3-rc.2"}
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T20:36:07.3655498Z NGINX_PASS: isolated Docker and fast routing regression checks complete
Fresh Docker installer integration (Ubuntu 22.04)	Execute isolated source installer and eight assertions	2026-10-14T20:36:07.4597998Z E2E_PASS: eight assertions; source-built Docker image=sha256:0724acf992b2332294be73c7e264ce1745c5921794ae9b536326ce3445bfde16; complete redacted log=/home/runner/work/yuchen-panel/yuchen-panel/.audit/task8-e2e/docker-install.log
```

后续仅补充轮询时同步详情／分享窗口引用及本报告；不修改计费核心。最新提交 CI 将在 PR 中记录。

官方资料：[Stats API](https://xtls.github.io/config/api.html)、[统计格式](https://xtls.github.io/en/config/stats.html)、[本次固定核心的 CLI 源码](https://github.com/XTLS/Xray-core/blob/v26.3.27/main/commands/all/api/stats_query.go)。
