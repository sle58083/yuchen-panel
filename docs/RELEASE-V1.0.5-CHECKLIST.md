# v1.0.5 发布核验清单

本清单对应 `release/v1.0.5`。在协调者确认最终 diff 前，不创建 tag；在 CI 完成并通过发布核验前，不将 Release 标记为可交付。

## 打 tag 前

- [ ] 分支从包含 PR #11 merge commit `243d170` 的 `main` 创建，工作树干净。
- [ ] `VERSION`、`frontend/package.json`、`frontend/package-lock.json` 根版本、`frontend/package-lock.json` 的 `packages[""]`、`frontend/src/version.ts`、`backend/internal/version/version.go` 均为 `1.0.4`。
- [ ] `backend/internal/version/version.go` 的 `Channel` 为 `stable`，版本源文件没有 `-rc.` 残留。
- [ ] `bash tools/check-version-consistency.sh 1.0.4` 通过。
- [ ] PR #11 的 SOCKS5 随机用户名修复仍在当前提交中；同时确认新建 gRPC 入站会自动生成随机 12 位 serviceName，编辑已有节点仍保留原值。
- [ ] `rg` 确认 gRPC 默认 `serviceName` 的 `yuchen` 仍保留，且 PR #11 没有改动该协议字段。
- [ ] `go test ./...`、`go test -race ./...` 通过。
- [ ] `cd frontend && npm ci --no-audit --no-fund --progress=false && node --test tests/*.test.mjs && VITE_BASE_PATH=/ npm run build` 通过。
- [ ] 记录最终提交 SHA、改动文件白名单和上述检查输出，等待协调者确认。

## CI 构建与 Release

- [ ] 只创建带注解的 `v1.0.5` tag，并推送到远程；不复用或移动已有 tag。
- [ ] Release workflow 从该 tag 构建 API、Agent 和前端，生成 `yuchen-panel-v1.0.5.zip` 与 `SHA256SUMS`。
- [ ] CI 发布 Release 后下载资产，重新计算 ZIP SHA256，并确认与 `SHA256SUMS` 一致。
- [ ] CI 生成的 `version.json` 将 `sha256` 替换为本次 ZIP 的真实摘要，并 fast-forward 更新 `main`；发布前分支中的旧摘要不能被当作最终摘要。
- [ ] Release 为正式版：`prerelease=false`，标题为 `Yuchen Panel v1.0.5`，资产和 tag 指向同一提交。

## 发布后核验

所有 GitHub API、`version.json` 和 tag/blob URL 都追加 cache-buster，避免缓存导致误判：

```bash
CACHE_BUSTER="$(date +%s%N)"
curl -fsSL "https://api.github.com/repos/sle58083/yuchen-panel/releases?per_page=10&ts=${CACHE_BUSTER}"
curl -fsSL "https://raw.githubusercontent.com/sle58083/yuchen-panel/main/version.json?ts=${CACHE_BUSTER}"
curl -fsSL "https://github.com/sle58083/yuchen-panel/blob/main/VERSION?ts=${CACHE_BUSTER}" >/dev/null
```

- [ ] API 返回 Release `v1.0.5`，状态为正式版，资产名和下载 URL 正确。
- [ ] `version.json` 的 `latest`、`version`、`package`、`download_url`、`sha256` 与 Release 资产一致。
- [ ] 旧版本 `v1.0.3` 的 Release 资产和 SHA256 未改变。
- [ ] tag 为 annotated tag，且注解指向发布提交。
- [ ] 安装包内 grep 证明 PR #11 的随机 SOCKS5 用户名修复存在；同时证明 gRPC `serviceName` 默认 `yuchen` 未被改动。
- [ ] 测试机升级后用 `curl -s http://127.0.0.1:8088/api/health` 验证 API 返回 `version: 1.0.4`；不以 `docker logs` 或 `journalctl` 作为唯一依据。

