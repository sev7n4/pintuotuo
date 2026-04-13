#!/usr/bin/env bash
# 解析并打印应使用的 GOMODCACHE 绝对路径（单行，无换行）。
# 当 IDE/沙箱将缓存指到不完整目录（常见：路径含 cursor-sandbox-cache）时，回退到 $HOME/go/pkg/mod。
# Makefile 与 setup.sh 使用；也可： eval "$(scripts/print-go-modcache-export.sh)"（若需 export）
set -euo pipefail

_eff="${GOMODCACHE:-}"
if [[ -z "${_eff}" ]]; then
  _eff="$(go env GOMODCACHE 2>/dev/null || true)"
fi
if [[ -z "${_eff}" ]]; then
  _eff="${HOME}/go/pkg/mod"
fi

_out="${_eff}"
# Cursor / 部分沙箱注入的 mod 缓存路径不完整，导致 go test 报源文件缺失
if [[ "${_eff}" == *cursor-sandbox* ]] || [[ "${_eff}" == *Cursor/sandbox* ]]; then
  _out="${HOME}/go/pkg/mod"
elif [[ -n "${_eff}" ]] && [[ ! -d "${_eff}" ]]; then
  _out="${HOME}/go/pkg/mod"
fi

mkdir -p "${_out}" 2>/dev/null || true
printf '%s' "${_out}"
