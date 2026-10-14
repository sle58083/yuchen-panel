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
            "审计修复 RC：认证、升级与安装路径安全修复",
            "sing-box/v2rayN/Clash/Shadowrocket 客户端兼容",
            "三档节点预设，默认抗封锁档（用户墙内客户策略）",
            "Docker Nginx 路由及 Doctor 模式检查修复",
            "升级会轮换 JWT 密钥，旧会话失效，需要重新登录",
            "服务器规格提示依赖新版 Agent",
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
