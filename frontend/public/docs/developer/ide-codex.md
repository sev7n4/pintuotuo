# OpenAI Codex CLI / 官方 Codex 工具链

默认连接 **OpenAI 官方**。若所用版本支持 **自定义 Base URL**（环境变量或配置文件名称以官方为准）：

```bash
export OPENAI_API_KEY="ptd_YOUR_KEY_HERE"
export OPENAI_BASE_URL="https://YOUR_ORIGIN/api/v1/openai/v1"
# 再在命令或配置中指定 model = 权益内的 provider/model
```

**注意**：须确认版本是否允许覆盖 Base URL；**`model` 不能填未购权益的 OpenAI 商品名**（strict 下会 403）。
