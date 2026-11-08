# 诊断提交模板

- 客户端 OS/版本：
- 客户端程序/版本，实际核心/版本：
- 连接角色：CLIENT → RELAY → EXIT → TARGET（只用角色名称，不填 IP）
- RTT 大致范围；服务器内存/vCPU：
- 测速时间（时区）、工具、单连接/多连接、单位 Mbps 或 MB/s：
- 测试目标是否相同；目标站/测试端容量是否充足：
- 系统代理速度；TUN 速度；TUN MTU；TUN stack：
- 客户端 Mux/multiplex/smux：
- 客户端 Vision flow；服务端 runtime Vision flow：
- iperf 发送端角色/接收端角色，P1/P8/反向结果：
- MTU 探测：每个 payload 的成功/失败；是否经过 TUN：
- 本次只改变的参数；改变前后速度：

粘贴 collect-network-diagnostics.sh 的精简输出和 iperf JSON 摘要。
连接号在每个 ss 样本内编号，跨样本不保证对应同一连接。
脚本默认不提交原始日志；若有错误计数，另行提供脱敏后的 1–3 条对应错误。
不要发送原始配置、订阅 URL、UUID、密码、Reality 私钥或 Agent token。
