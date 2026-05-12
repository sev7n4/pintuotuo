# Runbook：环境快照与脱敏导出（Phase 0.3）

**版本**：1.0  
**日期**：2026-05-12  
**目的**：为能力矩阵与上游探测提供**不含密钥**的配置基线；满足审计与复盘。

## 1. 安全原则

- **禁止**将 `merchant_api_keys.api_key_encrypted` 解密结果、明文 Key、JWT 写入仓库或聊天。
- 导出文件默认命名包含环境名与时间戳，存放于**访问受控**目录（如内网对象存储或加密盘）。
- 参与探测的人员使用**个人有权限**的预发 Key；轮换策略按安全制度执行。

## 2. 建议导出的 SQL（PostgreSQL）

在只读账号或只读事务中执行。

### 2.1 `model_providers`（无密钥）

```sql
SELECT
  id, code, name, api_format, status, sort_order,
  COALESCE(api_base_url, '') AS api_base_url,
  COALESCE(provider_region, '') AS provider_region,
  route_strategy, endpoints,
  compat_prefixes,
  COALESCE(litellm_model_template, '') AS litellm_model_template,
  COALESCE(litellm_gateway_api_base, '') AS litellm_gateway_api_base,
  updated_at
FROM model_providers
ORDER BY sort_order, id;
```

### 2.2 `merchant_api_keys`（仅元数据）

```sql
SELECT
  mak.id,
  mak.merchant_id,
  mak.provider,
  mak.status,
  mak.is_default,
  mak.verified_at,
  mak.verification_result,
  mak.route_mode,
  mak.region,
  mak.health_status,
  mak.last_health_check_at,
  mak.updated_at
FROM merchant_api_keys mak
WHERE mak.status = 'active'
ORDER BY mak.provider, mak.merchant_id, mak.id;
```

> **不要**在公共导出中包含 `api_key_encrypted`、`route_config` 内可能含 URL 带 token 的字段；若必须导出 `route_config` 用于排障，先经安全审批并做字段级脱敏。

### 2.3 `provider_models`（可选，用于模型维度矩阵）

```sql
SELECT provider_code, model_id, is_active, synced_at, updated_at
FROM provider_models
WHERE is_active = true
ORDER BY provider_code, model_id;
```

## 3. 与探测命令的衔接

1. 从 `model_providers` 结果中读取各 `code` 与 `api_base_url`（或 `endpoints.direct` 等），确定 **OpenAI 兼容根 URL**（用于人工复核；自动化探测见 [README.md](./README.md) 中的 `capability-probe`）。
2. **在部署环境**执行 `capability-probe`：从 **`merchant_api_keys` 解密 BYOK**，对每条活跃密钥打 `GET /v1/models` 与（对 `api_format=openai`）`POST /v1/embeddings`，生成 CSV。
3. 将 CSV 归档，并在 [matrix-template.md](./matrix-template.md) 中回填状态。

## 4. 完成检查

- [ ] 导出文件已脱敏并存档路径已登记在 Epic/Wiki。
- [ ] 参与人确认无密钥进入 Git 历史。
