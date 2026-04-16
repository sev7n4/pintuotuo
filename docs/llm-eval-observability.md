# LLM 评测与可观测性使用说明

## Promptfoo 回归

1. 配置环境变量：
   - `PROMPTFOO_BASE_URL`
   - `PROMPTFOO_API_KEY`
   - `PROMPTFOO_MODEL`
2. 运行：
   - `bash scripts/run_prompt_evals.sh`
3. 查看报告：
   - `evals/promptfoo/reports/promptfoo-report.json`

## Langfuse 灰度追踪

- 默认关闭：`LANGFUSE_ENABLED=false`
- 启用后会输出兼容 Langfuse 的追踪日志字段（request_id、provider、model、latency、status、error_code）。
- 当前为最小侵入式接入，先保证链路可观测，再按需要替换为官方 SDK。
