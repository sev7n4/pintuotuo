# LiteLLM 聚合网关配置

- 主配置：[litellm_proxy_config.yaml](./litellm_proxy_config.yaml)（`model_list` + `router_settings` + `litellm_settings`）。
- **目录单一事实来源**：商品以库表 `spus` + `model_providers` 为准；与网关路由对齐请使用：

```bash
# 需 DATABASE_URL 指向含迁移后的库
cd backend
go run ./cmd/litellm-catalog-sync -verify \
  -config ../deploy/litellm/litellm_proxy_config.yaml \
  -map ../deploy/litellm/catalog_provider_map.json
```

- 种子库与「P0 全量模型」yaml 可能暂时不一致时，可使用 `-soft`（仅打印缺失，退出码 0）。
- 从目录**生成**可合并的 `model_list` 片段：

```bash
go run ./cmd/litellm-catalog-sync -generate -map ../deploy/litellm/catalog_provider_map.json -out /tmp/catalog_models.yaml
```

- 厂商 code → LiteLLM 前缀与 API Key 环境变量见 [catalog_provider_map.json](./catalog_provider_map.json)；新增厂商时请同步更新。
