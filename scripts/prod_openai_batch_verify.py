#!/usr/bin/env python3
"""Batch OpenAI-compat calls against deployed API; compare balance vs usage sum."""
from __future__ import annotations

import json
import random
import string
import sys
import time
import urllib.error
import urllib.request

BASE = "http://119.29.173.89/api/v1"
OPENAI_BASE = f"{BASE}/openai/v1"


def req(
    method: str,
    url: str,
    headers: dict | None = None,
    body: dict | None = None,
    timeout: float = 120.0,
) -> tuple[int, dict | list | str]:
    h = dict(headers or {})
    data = None
    if body is not None:
        data = json.dumps(body).encode("utf-8")
        h.setdefault("Content-Type", "application/json")
    r = urllib.request.Request(url, data=data, headers=h, method=method)
    try:
        with urllib.request.urlopen(r, timeout=timeout) as resp:
            raw = resp.read().decode("utf-8", errors="replace")
            code = resp.getcode()
    except urllib.error.HTTPError as e:
        raw = e.read().decode("utf-8", errors="replace")
        code = e.code
    try:
        return code, json.loads(raw)
    except json.JSONDecodeError:
        return code, raw


def unwrap(d):
    if isinstance(d, dict) and "data" in d and ("code" in d or "message" in d):
        return d["data"]
    return d


def login(email: str, password: str) -> str:
    code, body = req(
        "POST",
        f"{BASE}/users/login",
        body={"email": email, "password": password},
    )
    if code != 200:
        raise SystemExit(f"login failed {code}: {body}")
    data = unwrap(body)
    tok = data.get("token") if isinstance(data, dict) else None
    if not tok:
        raise SystemExit(f"no token in {body}")
    return tok


def create_api_key(jwt_token: str, name: str) -> str:
    code, body = req(
        "POST",
        f"{BASE}/tokens/keys",
        headers={"Authorization": f"Bearer {jwt_token}"},
        body={"name": name},
    )
    if code not in (200, 201):
        raise SystemExit(f"create key failed {code}: {body}")
    if isinstance(body, dict) and "key" in body:
        return body["key"]
    data = unwrap(body)
    if isinstance(data, dict) and "key" in data:
        return data["key"]
    raise SystemExit(f"no key in response: {body}")


def get_balance(jwt_token: str) -> float:
    code, body = req(
        "GET",
        f"{BASE}/tokens/balance",
        headers={"Authorization": f"Bearer {jwt_token}"},
    )
    if code != 200:
        raise SystemExit(f"balance {code}: {body}")
    b = unwrap(body)
    return float(b["balance"])


def chat_completion(api_key: str, model: str, user_chars: int) -> dict:
    content = "".join(random.choices(string.ascii_letters + string.digits + " ", k=user_chars))
    payload = {
        "model": model,
        "messages": [{"role": "user", "content": content}],
        "stream": False,
    }
    code, body = req(
        "POST",
        f"{OPENAI_BASE}/chat/completions",
        headers={"Authorization": f"Bearer {api_key}"},
        body=payload,
        timeout=180.0,
    )
    return {"http": code, "body": body}


def main():
    n = int(sys.argv[1]) if len(sys.argv) > 1 else 100
    user_email = "user002@163.com"
    user_pass = "111111"
    model = "stepfun/step-1-8k"

    jwt_tok = login(user_email, user_pass)
    bal0 = get_balance(jwt_tok)
    key_name = f"batch-verify-{int(time.time())}"
    ptd_key = create_api_key(jwt_tok, key_name)
    print(f"created key {key_name}, balance_before={bal0}", flush=True)

    total_pt = 0
    total_ct = 0
    ok = 0
    fail = 0
    for i in range(n):
        ulen = random.randint(8, 120)
        r = chat_completion(ptd_key, model, ulen)
        if r["http"] == 200 and isinstance(r["body"], dict):
            usage = r["body"].get("usage") or {}
            pt = int(usage.get("prompt_tokens", 0))
            ct = int(usage.get("completion_tokens", 0))
            total_pt += pt
            total_ct += ct
            ok += 1
        else:
            fail += 1
            print(f"  fail #{i+1} http={r['http']} body={str(r['body'])[:500]}", flush=True)
        if (i + 1) % 10 == 0:
            print(f"  progress {i+1}/{n} ok={ok} fail={fail}", flush=True)

    time.sleep(1.5)
    bal1 = get_balance(jwt_tok)
    delta = bal0 - bal1
    expect_tokens = total_pt + total_ct
    print(
        json.dumps(
            {
                "balance_before": bal0,
                "balance_after": bal1,
                "balance_delta": delta,
                "sum_prompt_tokens": total_pt,
                "sum_completion_tokens": total_ct,
                "sum_usage_tokens": expect_tokens,
                "ok": ok,
                "fail": fail,
            },
            indent=2,
        )
    )


if __name__ == "__main__":
    main()
