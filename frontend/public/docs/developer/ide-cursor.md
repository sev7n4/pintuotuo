# Cursor

1. 打开 **Settings → Models**（或 **Cursor Settings → OpenAI API**；版本不同菜单位置可能不同）。  
2. **Override OpenAI Base URL** / **Custom OpenAI API**：  
   `https://YOUR_ORIGIN/api/v1/openai/v1`  
3. **API Key**：`ptd_...`  
4. 对话或模型列表中选择 **`api-usage-guide`** 里的 **`provider/model`**。  

若只使用 Cursor 内置 **Anthropic** 通道而未改 Base URL，**不会**走拼脱脱计费与权益；需改为 **OpenAI 兼容 + 自定义 Base URL**，或若产品支持 **Custom Anthropic Base URL**，则填 **`https://YOUR_ORIGIN/api/v1/anthropic/v1`** 与 `ptd_`，模型仍须为权益内 ID。
