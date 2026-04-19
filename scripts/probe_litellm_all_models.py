#!/usr/bin/env python3
"""
一键探测：仅读取 litellm_proxy_config.yaml 中的 model_name，直连 LiteLLM 网关
POST /v1/chat/completions（不经业务后端），用于校验 Master Key + deployment + 上游。

依赖：Python 3.10+ 标准库（无 PyYAML / 无 curl 依赖）。

示例：

  export LITELLM_MASTER_KEY=...
  python3 scripts/probe_litellm_all_models.py
  python3 scripts/probe_litellm_all_models.py --url http://127.0.0.1:4000 --env-file .env
  make probe-litellm

可选：--stream 对每条模型额外做一次流式请求（内部用 curl，需系统有 curl）。
"""
from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
import urllib.error
import urllib.request
from pathlib import Path


def load_master_key(env_file: str | None) -> str:
    if env_file and Path(env_file).is_file():
        for line in Path(env_file).read_text(encoding="utf-8", errors="replace").splitlines():
            line = line.strip()
            if line.startswith("LITELLM_MASTER_KEY="):
                v = line.split("=", 1)[1].strip().strip('"').strip("'")
                if v:
                    return v
    mk = os.environ.get("LITELLM_MASTER_KEY", "").strip()
    if mk:
        return mk
    sys.stderr.write("错误: 请设置环境变量 LITELLM_MASTER_KEY 或使用 --env-file 指向含该变量的 .env\n")
    sys.exit(1)


def parse_model_names(yaml_path: str) -> list[str]:
    text = Path(yaml_path).read_text(encoding="utf-8", errors="replace")
    # 与 yaml 中 `  - model_name: xxx` 对齐；忽略行尾注释
    found = re.findall(r"^\s*-\s*model_name:\s*([^\s#]+)", text, re.MULTILINE)
    return list(dict.fromkeys(found))


def short_err(body: str, limit: int = 220) -> str:
    try:
        j = json.loads(body)
    except json.JSONDecodeError:
        return body[:limit].replace("\n", " ")
    err = j.get("error")
    if isinstance(err, dict):
        msg = err.get("message") or err.get("type") or json.dumps(err, ensure_ascii=False)
    else:
        msg = str(err or j)
    return msg.replace("\n", " ")[:limit]


def post_chat_completions(
    base_url: str,
    master_key: str,
    model: str,
    *,
    max_tokens: int,
    timeout: float,
) -> tuple[int, str]:
    """返回 (HTTP 状态码, 响应体文本)。失败连接记为 -1 与原因字符串。"""
    url = base_url.rstrip("/") + "/v1/chat/completions"
    payload = json.dumps(
        {
            "model": model,
            "messages": [{"role": "user", "content": "ping"}],
            "max_tokens": max_tokens,
        },
        ensure_ascii=False,
    ).encode("utf-8")
    req = urllib.request.Request(
        url,
        data=payload,
        method="POST",
        headers={
            "Authorization": f"Bearer {master_key}",
            "Content-Type": "application/json; charset=utf-8",
            "User-Agent": "pintuotuo-probe-litellm/1.0",
        },
    )
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            body = resp.read().decode("utf-8", errors="replace")
            return int(resp.status), body
    except urllib.error.HTTPError as e:
        body = e.read().decode("utf-8", errors="replace")
        return int(e.code), body
    except urllib.error.URLError as e:
        reason = getattr(e, "reason", e)
        return -1, str(reason)


def probe_stream_curl(
    base_url: str,
    master_key: str,
    model: str,
    *,
    max_tokens: int,
    timeout: int,
) -> tuple[str, str]:
    """返回 (状态摘要, 详情)。依赖系统 curl。"""
    body = {
        "model": model,
        "messages": [{"role": "user", "content": "ping"}],
        "max_tokens": max_tokens,
        "stream": True,
    }
    payload = json.dumps(body, ensure_ascii=False)
    cmd = [
        "curl",
        "-sS",
        "--max-time",
        str(timeout),
        "-H",
        f"Authorization: Bearer {master_key}",
        "-H",
        "Content-Type: application/json",
        "-d",
        payload,
        f"{base_url.rstrip('/')}/v1/chat/completions",
    ]
    p = subprocess.run(cmd, capture_output=True, text=True)
    out = (p.stdout or "") + (p.stderr or "")
    if p.returncode != 0:
        return "ERR", f"curl_rc={p.returncode} {out[:200].replace(chr(10), ' ')}"
    if out.lstrip().startswith("data:") or "\ndata:" in out[:8000]:
        return "200", "OK_SSE"
    try:
        j = json.loads(out)
    except json.JSONDecodeError:
        return "?", out[:200].replace("\n", " ")
    if j.get("choices"):
        return "200", "OK_JSON"
    return "ERR", short_err(out, 320)


def main() -> None:
    ap = argparse.ArgumentParser(description="从 YAML 读取 model_name，探测 LiteLLM /v1/chat/completions")
    ap.add_argument(
        "--yaml",
        default="deploy/litellm/litellm_proxy_config.yaml",
        help="litellm_proxy_config.yaml 路径",
    )
    ap.add_argument("--url", default="http://127.0.0.1:4000", help="LiteLLM 根 URL（无尾斜杠）")
    ap.add_argument("--env-file", default="", help="含 LITELLM_MASTER_KEY 的 .env 路径")
    ap.add_argument("--max-tokens", type=int, default=4, help="chat/completions 的 max_tokens")
    ap.add_argument("--timeout", type=float, default=45.0, help="单次非流式请求超时（秒）")
    ap.add_argument(
        "--stream",
        action="store_true",
        help="每条模型在非流式之外再测一次流式（需系统安装 curl）",
    )
    ap.add_argument("--timeout-stream", type=int, default=30, help="流式请求超时（秒），仅 --stream 时有效")
    ap.add_argument("--tsv", default="", help="可选：将结果写入 TSV 文件")
    ap.add_argument(
        "--strict",
        action="store_true",
        help="若任一模型非流式 HTTP 非 200 则 exit 1",
    )
    args = ap.parse_args()

    mk = load_master_key(args.env_file or None)
    ypath = Path(args.yaml)
    if not ypath.is_file():
        sys.stderr.write(f"错误: 找不到 YAML: {ypath}\n")
        sys.exit(1)

    models = parse_model_names(str(ypath))
    print(f"YAML: {ypath}")
    print(f"LiteLLM: {args.url}")
    print(f"model_name 数: {len(models)} | max_tokens={args.max_tokens} | timeout={args.timeout}s")
    print("=" * 100)

    rows: list[tuple[str, ...]] = []
    any_non_200 = False
    for i, m in enumerate(models, 1):
        code, body = post_chat_completions(
            args.url,
            mk,
            m,
            max_tokens=args.max_tokens,
            timeout=args.timeout,
        )
        ok = code == 200
        if not ok:
            any_non_200 = True
        detail = "OK" if ok else short_err(body)
        if code < 0:
            detail = body[:300]
        row = (m, str(code), detail)
        line = f"[{i}/{len(models)}] {m}\tHTTP {code}\t{detail[:300]}"
        print(line)
        if args.stream:
            st, std = probe_stream_curl(
                args.url,
                mk,
                m,
                max_tokens=args.max_tokens,
                timeout=args.timeout_stream,
            )
            print(f"         stream\t{st}\t{std[:300]}")
            row = row + (st, std)
        rows.append(row)
        sys.stdout.flush()

    ok_ct = sum(1 for r in rows if r[1] == "200")
    print("=" * 100)
    print(f"非流式 HTTP 200: {ok_ct}/{len(models)}")
    if args.tsv:
        outp = Path(args.tsv)
        outp.parent.mkdir(parents=True, exist_ok=True)
        with outp.open("w", encoding="utf-8") as f:
            if args.stream:
                f.write("model_name\thttp_status\tnon_stream_detail\tstream_status\tstream_detail\n")
                for r in rows:
                    f.write("\t".join(str(x).replace("\t", " ") for x in r) + "\n")
            else:
                f.write("model_name\thttp_status\tdetail\n")
                for r in rows:
                    f.write("\t".join(str(x).replace("\t", " ") for x in r) + "\n")
        print(f"TSV: {outp.resolve()}")

    if args.strict and any_non_200:
        sys.exit(1)


if __name__ == "__main__":
    main()
