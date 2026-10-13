# Task 6 交付报告：真机 `--apply` 自检与部署阻塞修复

本轮基于 Task 5 提交，修复真机复核发现的 BUG-E、BUG-F、BUG-G、BUG-H。`model.Client.Password` 的设计语义确认是“管理员可指定的客户 Trojan 密码”：`xray.TrojanClientPassword` 优先读取该字段，普通客户创建和固定出口客户创建现在都保留请求中的密码，空值才生成随机密码。

## BUG-E：密码自检使用了错误的基准

`scripts/prepare-test-env.sh` 现在在调用 API 前把原始密码写入临时的 `expected-password.txt`。导出后同时读取：

1. 独立基准文件中的原始密码；
2. `/api/clients/{id}` 返回的数据库密码；
3. `single-node-links.txt` 中 Trojan URI 经 Python 标准库解析后的密码。

只有三者逐字节相等才输出 `SELF_CHECK_PASS`。脚本不再用“数据库值和由数据库值生成的链接”进行自读自比。

固定出口客户创建请求结构也增加了 `password` 字段，创建逻辑使用非空请求密码，否则才调用 `store.NewPassword()`。

## BUG-F：节点名称重复短 ID

`nodeShareName` 现在检测原始名称是否已经以相同 `#<short-id>` 结尾，避免 `uniqueSubscriptionNames` 预先加工后下游再次追加。新增测试覆盖：

- sing-box tag；
- Clash proxy name；
- v2rayN VMess `ps`；
- 分享链接 fragment。

四条路径都断言短 ID 只出现一次。

## BUG-G：版本号不一致

backend 新增 `internal/version.Version = "1.0.2"`，health 与新建 PanelData 使用该来源；Agent、前端 package/package-lock、前端版本提示和诊断文案统一为 `1.0.2`。`backend/agent/frontend` 中不再出现 `0.7.7.1`。

## BUG-H：Docker 模式 Agent 失败

Docker 模式只构建 API 和前端镜像，宿主机没有 Agent 二进制时不会再静默尝试 Go 编译。`deploy/install.sh` 会明确打印：

```text
Docker mode: skipping Agent installation. Deploy yuchen-agent separately before B-tier node connectivity tests.
```

部署文档同步说明：需要单独提供匹配架构的 `bin/yuchen-agent-linux-amd64` 后运行 `deploy/agent-install.sh`；Agent 就绪前 T2/T4/T5 的 B 档不可用。

## 验证

| 检查 | 结果 |
| --- | --- |
| backend build/vet/race | PASS |
| agent build/vet/race | PASS |
| 前端 9 项测试与生产构建 | PASS |
| Bash syntax 与 ShellCheck | PASS |
| `0.7.7.1` 在 backend/agent/frontend 中搜索 | 0 命中 |
| 隔离临时 API `--apply` | PASS：6 节点、1 客户、独立密码基准、三方密码比较、四格式名称检查 |
| 真实 38.34.21.9 重新 apply | `MANUAL_REQUIRED`，需协调者执行 |

隔离 apply 的完整输出保存在 [apply-verify.txt](apply-verify.txt)。其中 `TASK6_APPLY_SMOKE_PASS` 只证明临时面板路径，不能替代协调者在真实服务器上的独立复核。未执行设备连接，也没有伪造 T2–T6 B 档结果。
