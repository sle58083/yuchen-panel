# Task 9 — C24 性能与抗识别档位

## 前提：由用户直接确认，不由协调者或代码推断

开始修改前明确询问了“客户是否主要在墙内、需要抗封锁”。用户回复：**“主要墙内／需要抗封锁：默认 Reality 档”**。因此本轮不把 VLESS 省略 security 的默认值改为 none；这是用户回复对原任务“默认极速”的覆盖，不是静默跳过。第 1 条测试的期望同步明确为 Reality，第 4/5 条存量兼容断言仍严格执行。

第二项询问发现旧接口只有使用率，没有核数与总内存；用户明确授权：**“允许复用 Agent 现有内存读取，并补充规格字段采集”**。据此扩展现有心跳，未另起采集服务或接口。

## 后端默认值与 before/after

```diff
 case "vless":
+    // Task 9: user confirmed mainly mainland clients; anti-blocking default.
     n.Security = "reality"
```

Before：未指定 VLESS security → Reality。
After：用户确认后仍为 Reality；选择极速/均衡时前端显式传 none，后端完整保留该选择。`if n.Security == ""` 守卫未移除，存量 none/Reality 均保留；Trojan 默认 tls；VMess 默认 none，原有 VMess+Reality 兼容处理不变。

## 前端档位选择器（代码说明）

`frontend/src/nodePresets.ts` 是 UI 和测试共用映射：

| 档位 | 协议 | 传输 | security | 本部署默认 |
| --- | --- | --- | --- | --- |
| 极速 | vless | tcp | none | 否 |
| 均衡 | vless | grpc | none | 否 |
| 抗封锁 | vless | tcp | reality | 是 |

`Nodes.vue` 从 `NODE_PRESETS` 渲染三张选择按钮，每张同时展示性能说明、抗识别能力与“未实测”。没有编造百分比，也没有承诺选档位能跑满线路。均衡的“中”仅指相对裸 TCP 的封装，不把明文 gRPC 当作 TLS 或封锁保证。

用户点击档位后执行 `Object.assign(form.value, presetParameters(next))`。gRPC service name 为 `yuchen`；Reality 档自动生成密钥。高级按钮只展开 security/transport，不修改当前字段；WS、Trojan、VMess 和 SOCKS5 的高级能力保留。编辑节点按实际 protocol/transport/security 识别档位，不把 Trojan Reality 误当 VLESS 抗封锁模板。

## 服务器规格与建议

Agent 在原心跳中增加 `cpu_cores`（进程可用逻辑核数）和 `memory_total_bytes`。`/proc/meminfo` 只读取一次，同时计算原 memory_usage 和总量；CPU 使用 `runtime.NumCPU()`。服务端保存这些字段并在原 `/api/system/checks` 的 `server_specs` 中按 server ID 返回，前端只读取当前节点服务器的数据，绝不拿面板容器资源替代远程节点。

- ≤ 2 核：提示吞吐优先时建议极速档；不覆盖用户墙内抗封锁默认。
- ≥ 4 核：提示可用 Reality，不保证单连接吞吐。
- 内存 < 1 GiB：提示 OOM/并发负载风险。
- 缺规格、读取失败、旧 Agent、接口不可用：静默隐藏相应提示。

旧 Agent 不带规格字段时，不擦除曾成功上报的容量；没有任何已知值时不捏造规格。规格需要新版源码构建 Agent 上报后才出现，旧 fast Release 二进制无法验证新增字段。

## 七条验收用例与真实输出

| 编号 | 用例 | 结果 |
| --- | --- | --- |
| 1 | 新 VLESS 不指定 security | Reality（按本轮用户直接确认）；生成配置同步验证 |
| 2 | 新 Trojan | tls |
| 3 | 新 VMess | none；显式 Reality 仍按原规则变 none |
| 4 | 存量 Reality 经 normalize | 保持 Reality |
| 5 | 存量 none 经 normalize | 保持 none |
| 6 | 三档参数映射 | fast=tcp/none，balanced=grpc/none，anti=tcp/Reality |
| 7 | ≤ 2 核建议 | 提示极速，不修改默认；另覆盖 ≥4 核/低内存/缺失规格 |

实际输出：`task9-backend-tests.txt`、`task9-frontend-tests.txt`、`task9-agent-tests.txt`。另验证 Agent 内存解析、CPU 字段序列化、心跳→服务器记录→system/checks 数据链、旧心跳兼容。前端共 11 项测试，包含原 9 项客户端分享回归。

本地源码 backend/agent build、vet、race 通过；前端生产构建与测试通过；ShellCheck/Bash syntax 通过。所有本地 Go/前端产物来自当前源码，未使用正式 Release 二进制。

Task 8 的 deploy/install.sh、Compose 文件、BUG-K/BUG-L/C22/C25 脚本未修改。CI 接入前端测试；端到端八项断言将在本轮 CI 中复验，完成后补实际结果。设备连接、链路测速、本轮默认档位真机 UI 仍需用户执行，不能拿源码测试代替。

## 发布前两项确认

1. 升级路径：安装器生成新的 64 字符 JWT_SECRET，写 .env 时剔除旧 YC_JWT_SECRET；旧短密钥不会被继续使用。旧登录令牌会失效。相关逻辑本轮没有改动。
2. 版本格式：VERSION 文件为发行标识 `1.0.2-stable`；API/Go/比较使用 `1.0.2`；panel.info 分写 VERSION=1.0.2、CHANNEL=stable。已明确统一语义，未强求发行标识与语义版本逐字一致。

未合并 main、未打 tag、未触发发布。

## CI 实际复验结果

[CI run 38534346317](https://github.com/sle58083/yuchen-panel/actions/runs/38534346317) 六个 job 全部 success，包括前端 11 项测试、前后端 Docker 构建、backend/agent build/vet/race、Shell 检查和 Ubuntu 22.04 完整隔离安装。

代码提交 `8bebf1e`；CI 使用该 PR 的源码合并检查产物构建 API，八个安装断言全部 PASS，包括 JWT 容器环境与真实 .env 逐字节一致。实际输出见 `task9-e2e-result.txt`；API image 为 `sha256:45eaebd82707d762ab24b995f4a37ab18068c654d1ba29acf08c5a09af9e0565`。

[完整安装日志 artifact](https://github.com/sle58083/yuchen-panel/actions/runs/38534346317/artifacts/11879841731)，ZIP SHA256：`b550cf8d4aa087cea66dd76f3f0fb3b4c11226e471fab0aa13acce09885a1487`。没有触碰真实 /opt/yuchen-panel、/etc/yuchen-panel，没有用旧 Release 二进制替代源码验证。

CI 安装验证仍不证明跨国线路测速、设备连接或新版 Agent 的真机规格提示；这些由用户真机复验。Task 8 预装 Compose v2 的包安装覆盖限制继续保留。
