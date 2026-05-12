# 行业调研摘要：OpenAI 兼容 API 与非 Chat 端点（Phase 0.1）

**版本**：1.0  
**日期**：2026-05-12  
**状态**：草案（随矩阵探测结果可修订）

## 1. 结论摘要（给产品 / 研发的共同语言）

1. **「OpenAI 兼容」在业界通常指**：同一类 URL 与请求习惯（如 `base_url` + `Authorization: Bearer` + JSON body），使客户端可复用 OpenAI SDK；**不等于**「实现与 OpenAI 公有云相同的全部子接口」。
2. **子能力差异集中在**：`chat/completions`（含流式、工具调用、多模态输入）覆盖率最高；**embeddings、images、audio、moderations、responses** 是否在同一 `base_url` 下可用，完全取决于各厂商或中间网关（LiteLLM、云厂商 OpenAI 适配层等）的实现与路由策略。
3. **生产侧最佳实践**：维护 **Capability Matrix（厂商 × 模型 × 端点 × 区域 × 认证）**；仅以 `GET /v1/models` 成功**不能**证明 embeddings 等可用；变更应配合 **最小 POST 探测** 或 **厂商文档证据** 留痕。
4. **聚合网关（LiteLLM、OpenRouter、多云统一网关）**：对外仍可能呈现 OpenAI 形状，但后端真实能力取决于路由到的 provider；矩阵中建议拆成 **「网关入口」+「真实后端」** 两行，避免把网关当成单一厂商。

## 2. 与本项目实现的对应关系

- 平台对 C 端暴露路径：见 `backend/routes/routes.go` 中 `RegisterOpenAICompatRoutes`（含 chat、responses、images、audio、embeddings、moderations）。
- 上游 URL 拼接：见 `backend/services/execution_layer.go` 的 `ResolveEndpointByType`——**默认**在去掉末尾 `/v1` 的根上拼接 `/v1/{embeddings|…}`。若某能力实际在不同 host 或前缀，矩阵应标为 `DifferentBase` 或 `Unsupported`，并在 Phase 4 引入独立配置（见总计划）。

## 3. 参考与延伸阅读（链接）

> 以下链接为公开文档，便于团队复核；不表示商业背书。

| 主题 | 参考 |
|------|------|
| OpenAI 官方 API 能力面 | [OpenAI API reference](https://platform.openai.com/docs/api-reference) |
| Google Gemini OpenAI 适配 | [Gemini OpenAI compatibility](https://ai.google.dev/gemini-api/docs/openai) |
| Azure / Microsoft Foundry OpenAI 形态 | [Azure OpenAI / Foundry 文档索引](https://learn.microsoft.com/en-us/azure/ai-services/openai/) |
| 多模型网关模式讨论 | 行业博客如 [OpenAI-compatible gateway 类文章](https://tokenmix.ai/blog/openai-compatible-api)（概念参考） |

## 4. 对本项目 Phase 0 的落地要求（自洽检查）

- [ ] 矩阵模板已创建并与 `endpointPathSuffixes` 列对齐。
- [ ] Runbook 已评审（无密钥落盘、脱敏字段列表明确）。
- [ ] 至少一条「最小 POST」样例与脚本行为一致。
- [ ] 与 Anthropic 原生路径、百度 `api_format` 等特殊行在矩阵中**单独说明**，避免误用 OpenAI 子路径评估。
