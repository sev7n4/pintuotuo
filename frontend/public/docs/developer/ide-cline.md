# Cline（VS Code 扩展）

1. 扩展设置中查找 **OpenAI-compatible** / **Custom API** / **Advanced**。  
2. **Base URL**：`https://YOUR_ORIGIN/api/v1/openai/v1`  
3. **API Key**：`ptd_...`  
4. **Model**：填写 **`api-usage-guide`** 中的模型 ID（`provider/model`）。

若扩展支持 **Anthropic 自定义 Endpoint**，可改用 **`https://YOUR_ORIGIN/api/v1/anthropic/v1`**，模型仍须与权益一致；请求体会**原样转发**至上游原生 `/v1/messages`（平台改写 `model`；LiteLLM 时注入 `user_config`），支持 **tool_use、多模态 content** 等 Claude 官方字段。
