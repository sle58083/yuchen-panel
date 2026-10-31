#!/usr/bin/env bash
# 版本一致性核对 —— 不依赖 jq / node / python，任选其一即可
# 用法: bash tools/check-version-consistency.sh [期望版本]
# 默认期望版本从 VERSION 文件读取

set -u
cd "$(dirname "$0")/.." || exit 1

V="${1:-$(tr -d ' \t\r\n' < VERSION)}"
echo "期望版本: $V"

# 取 JSON 字段：jq > node > python3 > grep
jget() {  # $1=文件 $2=字段路径(点号写法)
  f="$1"; e="$2"
  if command -v jq >/dev/null 2>&1; then jq -r ".$e" "$f"; return; fi
  if command -v node >/dev/null 2>&1; then
    node -e "const d=require('./$f');console.log(('$e').split('.').filter(Boolean).reduce((o,k)=>o?.[k.replace(/^\"|\"\$/g,'')],d) ?? '')" 2>/dev/null && return
  fi
  if command -v python3 >/dev/null 2>&1; then
    python3 -c "import json,sys;d=json.load(open(sys.argv[1],encoding='utf-8'));print(eval('d'+sys.argv[2]) or '')" "$f" "['$e']" 2>/dev/null && return
  fi
  grep -m1 "\"version\"" "$f" | sed 's/.*: *"//; s/".*//'
}

fail=0
chk() {
  local name="$1" got="$2"
  if [ "$got" = "$V" ]; then
    printf '  ✅  %-42s %s\n' "$name" "$got"
  else
    printf '  ❌  %-42s %s  (期望 %s)\n' "$name" "$got" "$V"; fail=1
  fi
}

echo "--- 版本一致性 ---"
chk "VERSION"                             "$(tr -d ' \t\r\n' < VERSION)"
chk "frontend/package.json"               "$(jget frontend/package.json version)"
chk "frontend/package-lock.json 根"       "$(jget frontend/package-lock.json version)"
chk "frontend/package-lock.json packages['']" "$(jget frontend/package-lock.json 'packages[""].version')"
chk "frontend/src/version.ts"             "$(sed -n "s/.*RELEASE_VERSION = '\([^']*\)'.*/\1/p" frontend/src/version.ts | head -1)"
chk "backend/internal/version/version.go" "$(sed -n 's/^const Version = "\([^"]*\)"/\1/p' backend/internal/version/version.go | head -1)"

CH=$(sed -n 's/^const Channel = "\([^"]*\)"/\1/p' backend/internal/version/version.go | head -1)
if [ "$CH" = "stable" ]; then
  printf '  ✅  %-42s %s\n' "version.go Channel（正式版应为 stable）" "$CH"
else
  printf '  ❌  %-42s %s  (期望 stable)\n' "version.go Channel" "$CH"; fail=1
fi

echo "--- 无残留 RC 字面量 ---"
LEFT=$(grep -rn -- '-rc\.' VERSION backend/internal/version/version.go frontend/package.json frontend/src/version.ts 2>/dev/null | grep -v '^\s*$' || true)
if [ -z "$LEFT" ]; then
  printf '  ✅  %-42s 无\n' "版本源文件中的 -rc. 残留"
else
  printf '  ❌  发现 -rc. 残留:\n%s\n' "$LEFT"; fail=1
fi

echo "--- 后端无硬编码版本字面量（与 version_log_test.go 同规则）---"
VIOL=$(find backend -name '*.go' ! -name '*_test.go' ! -path 'backend/internal/version/version.go' -print0 \
       | xargs -0 grep -nE '1\.0\.[0-9]+' 2>/dev/null || true)
if [ -z "$VIOL" ]; then
  printf '  ✅  %-42s 0 处\n' "backend 硬编码版本"
else
  printf '  ❌  发现硬编码:\n%s\n' "$VIOL"; fail=1
fi

echo
if [ "$fail" -eq 0 ]; then echo "结论: 全部一致 ✅"; else echo "结论: 存在不一致 ❌"; fi
exit $fail
