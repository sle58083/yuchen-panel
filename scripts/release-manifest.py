#!/usr/bin/env python3
"""Create complete release metadata from CI-produced, verified archive digest."""
import argparse
import datetime
import json
import re
from pathlib import Path

def release_manifest(version, sha, repository):
    if not re.fullmatch(r"\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?", version):
        raise ValueError("invalid semantic release version")
    if not re.fullmatch(r"[0-9a-f]{64}", sha):
        raise ValueError("digest must come from sha256sum")
    if not re.fullmatch(r"[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+", repository):
        raise ValueError("invalid repository")
    package = f"yuchen-panel-v{version}.zip"
    return {
        "latest": version,
        "version": version,
        "codename": "rc" if "-rc." in version else "stable",
        "package": package,
        "download_url": f"https://github.com/{repository}/releases/download/v{version}/{package}",
        "sha256": sha,
        "min_supported_version": "0.7.5",
        "released_at": datetime.datetime.now(datetime.timezone.utc).isoformat().replace("+00:00", "Z"),
        "changelog": [
            "包月限流量：按客户 ID 汇总 Xray 上传与下载，Stats API 仅监听本机",
            "epoch 与服务器游标持久化，处理重启与重复上报",
            "到期或超限停用；延长已有到期日才续费，清零并恢复启用",
            "长期客户首次设置到期日保留用量及启用状态",
            "区分负数与超范围流量上限提示，移除版本跟踪的过程报告",
            "客户列表按字节显示本周期用量，30 秒刷新且编辑时暂停",
            "正常在线限额执行通常有约 30–60 秒延迟，异常时可能更久",
            "升级会轮换 JWT 密钥，旧会话失效，需要重新登录",
            "需要匹配的新版 API 与 Agent；用户已确认 rc.3 超限停用，rc.4 完整真机验收仍待执行",
        ],
    }

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--version", required=True)
    parser.add_argument("--sha256", required=True)
    parser.add_argument("--repository", required=True)
    parser.add_argument("--output", required=True)
    args = parser.parse_args()
    data = release_manifest(args.version, args.sha256, args.repository)
    Path(args.output).write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
