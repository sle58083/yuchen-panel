# v1.0.3-rc.1 — 审计修复 RC

- 安全：弱 JWT 密钥拒绝启动、禁用默认凭据、管理员必须启用。
- 安全：升级脚本字面量转义、清单限制同源重定向、安装包路径穿越防护。
- 修复：统一版本来源，升级版本使用语义比较。
- 客户端：sing-box 现代 DNS 与 v2rayN/Clash/Shadowrocket 格式兼容。
- 节点：极速、均衡、抗封锁三档，默认抗封锁档（按用户墙内客户策略）。
- 安装：提前拒绝 Compose v1、Docker Nginx 路由修复、密钥轮换提示。
- 测试：隔离式 Docker 全新安装八项断言及 Nginx 反代复验。

升级会轮换 JWT 签名密钥，所有已登录用户需要重新登录。
CPU/内存档位建议依赖新版 Agent，旧 Agent 未上报时提示不出现。

安装/升级：
```bash
bash <(curl -Ls https://raw.githubusercontent.com/sle58083/yuchen-panel/main/install.sh)
```

这是 RC 预发布版，正式版需用户实测后确认。
