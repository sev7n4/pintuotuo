# 架构改进方案：路由配置 SSOT 下沉到 merchant_api_keys

**文档版本**: v1.7  
**创建日期**: 2026-04-28  
**最后更新**: 2026-04-28

---

## 版本历史

| 版本 | 日期 | 变更内容 |
|---|---|---|
| v1.7 | 2026-04-28 | 商户端增加BYOK类型选择、列表展示和筛选功能 |
| v1.6 | 2026-04-28 | 将 merchant_type 重命名为 byok_type，明确是 API Key 的属性而非商户属性 |
| v1.5 | 2026-04-28 | 增加商户类别（官方/代理商/自建商）字段，支持BYOK分类筛选 |
| v1.4 | 2026-04-28 | 删除重复的5.3端点探测API；修正文档序号 |
| v1.3 | 2026-04-28 | 重命名管理端为BYOK路由管理；商户端后续优化；厂商端简化路由配置 |
| v1.2 | 2026-04-28 | 管理端API Key运维页面增强：增加筛选、红绿灯、验证能力；厂商侧UI暂不修改 |
| v1.1 | 2026-04-28 | 商户端表单保留成本信息录入字段 |
| v1.0 | 2026-04-28 | 初始版本，包含职责边界设计和 LiteLLM 影响评估 |

---

## 1. 背景与问题

### 1.1 当前架构

| 表 | 职责 | 数据层级 |
|---|---|---|
| `model_providers` | 厂商全局配置模板 | 全局（管理员维护） |
| `merchant_api_keys` | 商户API Key + 健康状态 | 商户级（BYOK） |

### 1.2 存在的问题

1. **路由配置不够灵活**：所有使用同一厂商的商户共享相同的路由策略（`model_providers.route_strategy`）
2. **端点配置分离**：
   - `model_providers.endpoints` 是全局模板
   - `merchant_api_keys.endpoint_url` 是商户自定义
   - 两者关系不清晰，路由决策时优先级不明确
3. **SSOT 不明确**：路由决策时应该参考哪个表？
4. **管理界面分离**：
   - 管理端在 `model_providers` 配置端点
   - 商户端在 `merchant_api_keys` 配置端点
   - 两边配置可能冲突

### 1.3 目标

将路由配置的 SSOT（Single Source of Truth）下沉到 `merchant_api_keys` 表，实现：
- 每个商户的每个 API Key 都有独立的路由配置
- 路由决策优先使用 `merchant_api_keys` 的配置
- `model_providers` 作为默认模板和 fallback

### 1.4 职责边界设计（托管模式）

采用**商户托管模式**设计，降低商户使用门槛（0解释、0成本参与）：

| 角色 | 职责 | 功能 |
|---|---|---|
| **商户端** | 自助上传 + 成本录入 + 轻量验证 | 上传 API Key、录入成本信息、轻量验证、深度验证、立即探测 |
| **管理端** | 运维托管 | API Key 验证管理、路由健康探测、路由配置维护 |
| **数据 SSOT** | `merchant_api_keys` | 唯一数据真相来源 |

**设计原则**：
- 商户只需上传 API Key 并录入成本信息，无需理解路由配置
- 路由运维工作由管理运营端代为维护
- 数据存储在 `merchant_api_keys`，保证 SSOT

---

## 2. 数据模型变更

### 2.1 merchant_api_keys 表增强

新增字段：

```sql
-- BYOK类别（API Key来源分类）
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS byok_type VARCHAR(20) DEFAULT 'official';
-- 可选值: official(官方), reseller(代理商), self_hosted(自建商)

-- 路由出站模式
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS route_mode VARCHAR(20) DEFAULT 'auto';
-- 可选值: auto, direct, litellm, proxy

-- 主端点URL（已有，明确职责）
-- endpoint_url VARCHAR(500) -- 商户自定义主端点

-- 备用端点URL
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS fallback_endpoint_url VARCHAR(500);

-- 完整路由配置（JSONB）
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS route_config JSONB DEFAULT '{}'::jsonb;
-- 示例内容:
-- {
--   "gateway_mode": "litellm",
--   "litellm_endpoint": "http://119.29.173.89:4000/v1",
--   "proxy_endpoint": "",
--   "fallback_mode": "direct",
--   "timeout_ms": 30000,
--   "retry_count": 3
-- }
```

### 2.2 BYOK 类别定义

| BYOK类别 | 英文标识 | 说明 | 示例 |
|---|---|---|---|
| **官方** | `official` | 大模型厂商官方API Key | 智普、OpenAI、Anthropic |
| **代理商** | `reseller` | 代理商提供的API Key | OpenRouter、硅基流动、DeepInfra |
| **自建商** | `self_hosted` | 自建数据中心部署的API Key | 私有化部署的智普、千问、DeepSeek |

> **说明**: `byok_type` 是 API Key 的属性，而非商户的属性。一个商户可能拥有多个不同类型的 API Key。

**区域字段**（已有 `region`）：
- `domestic` - 国内端点
- `overseas` - 海外端点

### 2.3 model_providers 表职责调整

保留字段但调整职责：

| 字段 | 新职责 |
|---|---|
| `code`, `name` | 厂商标识（不变） |
| `api_format` | API 格式模板（不变） |
| `billing_type`, `segment_config` | 计费配置（不变） |
| `provider_region` | 厂商默认区域（作为 fallback） |
| `route_strategy` | 默认路由策略模板（作为 fallback） |
| `endpoints` | 默认端点模板（作为 fallback） |

### 2.4 职责边界明确

| 表 | 新职责 |
|---|---|
| `model_providers` | 厂商元数据 + 计费配置 + 默认模板（fallback） |
| `merchant_api_keys` | **SSOT**：API Key + 路由配置 + 端点 + 健康状态 |

---

## 3. 实施步骤

### Phase 1: 数据库迁移

**文件**: `backend/migrations/069_route_config_ssot_to_api_keys.sql`

```sql
-- 1. 新增字段
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS byok_type VARCHAR(20) DEFAULT 'official';
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS route_mode VARCHAR(20) DEFAULT 'auto';
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS fallback_endpoint_url VARCHAR(500);
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS route_config JSONB DEFAULT '{}'::jsonb;

-- 2. 添加注释
COMMENT ON COLUMN merchant_api_keys.byok_type IS 'BYOK类别: official(官方), reseller(代理商), self_hosted(自建商)';
COMMENT ON COLUMN merchant_api_keys.route_mode IS '路由出站模式: auto(自动), direct(直连), litellm, proxy(代理)';
COMMENT ON COLUMN merchant_api_keys.fallback_endpoint_url IS '备用端点URL，主端点不可用时使用';
COMMENT ON COLUMN merchant_api_keys.route_config IS '完整路由配置（JSONB）: gateway_mode, endpoints, timeout, retry等';

-- 3. 创建索引
CREATE INDEX IF NOT EXISTS idx_merchant_api_keys_byok_type ON merchant_api_keys(byok_type);
CREATE INDEX IF NOT EXISTS idx_merchant_api_keys_route_mode ON merchant_api_keys(route_mode);
CREATE INDEX IF NOT EXISTS idx_merchant_api_keys_route_config ON merchant_api_keys USING GIN(route_config);

-- 4. 数据迁移：从 model_providers 复制默认配置到 merchant_api_keys
UPDATE merchant_api_keys mak
SET route_config = jsonb_build_object(
    'gateway_mode', COALESCE(
        mp.route_strategy->>'default_mode',
        'auto'
    ),
    'endpoints', COALESCE(mp.endpoints, '{}'::jsonb)
)
FROM model_providers mp
WHERE mak.provider = mp.code
  AND (mak.route_config IS NULL OR mak.route_config = '{}'::jsonb);
```

### Phase 2: 后端模型更新

**文件**: `backend/models/models.go`

```go
type MerchantAPIKey struct {
    // ... 现有字段 ...
    
    // BYOK类别（新增）
    BYOKType            string                 `json:"byok_type,omitempty"`            // official, reseller, self_hosted
    
    // 路由配置（新增）
    RouteMode           string                 `json:"route_mode,omitempty"`           // auto, direct, litellm, proxy
    FallbackEndpointURL string                 `json:"fallback_endpoint_url,omitempty"`
    RouteConfig         map[string]interface{} `json:"route_config,omitempty"`         // JSONB
}
```

### Phase 3: 路由决策逻辑修改

**文件**: `backend/services/execution_layer.go`

修改 `resolveEndpoint` 方法，优先使用 `merchant_api_keys` 的配置：

```go
func (l *ExecutionLayer) resolveEndpoint(cfg *ExecutionProviderConfig, apiKey *MerchantAPIKey) string {
    // 1. 优先使用 API Key 的自定义端点
    if apiKey != nil && apiKey.EndpointURL != "" {
        return apiKey.EndpointURL
    }
    
    // 2. 使用 API Key 的路由配置
    if apiKey != nil && apiKey.RouteConfig != nil {
        if endpoint := extractEndpointFromConfig(apiKey.RouteConfig, cfg.GatewayMode); endpoint != "" {
            return endpoint
        }
    }
    
    // 3. Fallback 到 model_providers 的端点模板
    return resolveFromProviderEndpoints(cfg.Endpoints, cfg.GatewayMode)
}
```

### Phase 4: 前端管理界面更新

#### 4.1 商户端 API Key 管理（修改）

**文件**: `frontend/src/pages/merchant/MerchantAPIKeys.tsx`

商户端需要增加 BYOK 类型相关功能：

**变更内容**：

1. **上传/创建 API Key 表单**：增加 BYOK 类型选择

```tsx
<Form.Item name="byok_type" label="API Key 类型" rules={[{ required: true }]}>
  <Select placeholder="选择 API Key 来源类型">
    <Select.Option value="official">官方（大模型厂商官方API）</Select.Option>
    <Select.Option value="reseller">代理商（如 OpenRouter、硅基流动）</Select.Option>
    <Select.Option value="self_hosted">自建商（私有化部署）</Select.Option>
  </Select>
</Form.Item>

<Form.Item name="region" label="端点区域">
  <Select placeholder="选择端点区域">
    <Select.Option value="domestic">国内</Select.Option>
    <Select.Option value="overseas">海外</Select.Option>
  </Select>
</Form.Item>
```

2. **API Key 列表**：增加 BYOK 类型列展示

```tsx
const columns = [
  // ... 现有列 ...
  {
    title: 'BYOK类型',
    dataIndex: 'byok_type',
    key: 'byok_type',
    render: (type) => {
      const typeMap = {
        'official': { color: 'blue', text: '官方' },
        'reseller': { color: 'orange', text: '代理商' },
        'self_hosted': { color: 'purple', text: '自建商' },
      };
      const config = typeMap[type] || typeMap['official'];
      return <Tag color={config.color}>{config.text}</Tag>;
    }
  },
  {
    title: '区域',
    dataIndex: 'region',
    key: 'region',
    render: (region) => region === 'domestic' ? '国内' : region === 'overseas' ? '海外' : region,
  },
  // ... 其他列 ...
];
```

3. **列表筛选**：增加 BYOK 类型筛选

```tsx
<Form.Item label="API Key 类型">
  <Select 
    placeholder="选择类型" 
    allowClear
    style={{ width: 120 }}
    onChange={handleByokTypeFilter}
  >
    <Select.Option value="official">官方</Select.Option>
    <Select.Option value="reseller">代理商</Select.Option>
    <Select.Option value="self_hosted">自建商</Select.Option>
  </Select>
</Form.Item>
```

4. **编辑 API Key**：支持修改 BYOK 类型和区域

#### 4.2 管理端 BYOK 路由管理（新增）

**新增文件**: `frontend/src/pages/admin/AdminBYOKRouting.tsx`

管理端新增 BYOK 路由管理页面，功能包括：
- 查看所有商户的 API Key 列表（含商户类别、区域、商户ID、商户名称等字段）
- 支持按商户类别筛选：官方、代理商、自建商
- 支持筛选查询后编辑路由配置和端点配置
- 提供不同出站模式下的验证能力：轻量验证、深度验证、立即探测
- 参考商户端红绿灯实现（所见即所得）

**列表字段设计**：

| 字段 | 说明 |
|---|---|
| 商户ID | merchant_id |
| 商户名称 | merchant_name |
| **BYOK类别** | byok_type（官方/代理商/自建商） |
| 厂商 | provider |
| API Key名称 | name |
| 区域 | region（国内/海外） |
| 路由模式 | route_mode（auto/direct/litellm/proxy） |
| 端点URL | endpoint_url |
| 健康状态 | health_status（红绿灯：绿=健康，黄=降级，红=不健康，灰=未知） |
| 验证状态 | verification_result |
| 状态 | status |

**筛选条件**：
- 商户ID/名称
- **BYOK类别**（官方/代理商/自建商）
- 厂商
- 区域
- 路由模式
- 健康状态
- 状态

```tsx
import { useEffect, useState } from 'react';
import { Table, Card, Button, Modal, Form, Select, Input, Divider, Space, Tag, Tooltip, message, InputNumber, Badge } from 'antd';
import { InfoCircleOutlined } from '@ant-design/icons';

const AdminBYOKRouting = () => {
  const [keys, setKeys] = useState([]);
  const [editingKey, setEditingKey] = useState(null);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();
  const [filters, setFilters] = useState({
    merchant_id: '',
    byok_type: '',
    provider: '',
    region: '',
    route_mode: '',
    health_status: '',
  });

  // BYOK类别映射
  const byokTypeLabel = (type: string) => {
    const map = {
      'official': { color: 'blue', text: '官方' },
      'reseller': { color: 'orange', text: '代理商' },
      'self_hosted': { color: 'purple', text: '自建商' },
    };
    return map[type] || { color: 'default', text: type };
  };

  // 健康状态红绿灯样式（参考商户端实现）
  const healthDotClass = (status: string) => {
    const map = {
      'healthy': styles.dotGreen,
      'degraded': styles.dotYellow,
      'unhealthy': styles.dotRed,
      'unknown': styles.dotGray,
    };
    return map[status?.toLowerCase()] || styles.dotGray;
  };

  const healthLabel = (status: string) => {
    const map = {
      'healthy': '健康',
      'degraded': '降级',
      'unhealthy': '不健康',
      'unknown': '未知',
    };
    return map[status?.toLowerCase()] || '未知';
  };

  const columns = [
    { title: '商户ID', dataIndex: 'merchant_id', key: 'merchant_id', width: 100 },
    { title: '商户名称', dataIndex: 'merchant_name', key: 'merchant_name', width: 120 },
    { 
      title: 'BYOK类别', 
      dataIndex: 'byok_type', 
      key: 'byok_type',
      width: 100,
      render: (type) => {
        const config = byokTypeLabel(type);
        return <Tag color={config.color}>{config.text}</Tag>;
      }
    },
    { title: '厂商', dataIndex: 'provider', key: 'provider', width: 100 },
    { title: 'API Key名称', dataIndex: 'name', key: 'name', width: 150 },
    { 
      title: '区域', 
      dataIndex: 'region', 
      key: 'region',
      width: 80,
      render: (region) => region === 'domestic' ? '国内' : region === 'overseas' ? '海外' : region,
    },
    { 
      title: '路由模式', 
      dataIndex: 'route_mode', 
      key: 'route_mode',
      width: 100,
      render: (mode) => {
        const modeMap = {
          'auto': { color: 'blue', text: '自动' },
          'direct': { color: 'green', text: '直连' },
          'litellm': { color: 'orange', text: 'LiteLLM' },
          'proxy': { color: 'purple', text: '代理' },
        };
        const config = modeMap[mode] || modeMap['auto'];
        return <Tag color={config.color}>{config.text}</Tag>;
      }
    },
    { title: '端点URL', dataIndex: 'endpoint_url', key: 'endpoint_url', ellipsis: true, width: 200 },
    { 
      title: (
        <span>
          健康状态{' '}
          <Tooltip title="绿灯=健康，黄灯=降级，红灯=不健康，灰灯=未知">
            <InfoCircleOutlined style={{ color: '#8c8c8c' }} />
          </Tooltip>
        </span>
      ),
      dataIndex: 'health_status', 
      key: 'health_status',
      width: 120,
      render: (status, record) => (
        <Space direction="vertical" size={4}>
          <Tooltip title={`当前健康状态：${healthLabel(status)}`}>
            <span className={styles.statusLightRow}>
              <span className={`${styles.statusDot} ${healthDotClass(status)}`} />
              <span className={styles.statusLightLabel}>{healthLabel(status)}</span>
            </span>
          </Tooltip>
          {record.last_health_check_at && (
            <span style={{ fontSize: '12px', color: '#999' }}>
              {new Date(record.last_health_check_at).toLocaleString('zh-CN')}
            </span>
          )}
        </Space>
      ),
    },
    { 
      title: '验证状态', 
      dataIndex: 'verification_result', 
      key: 'verification_result',
      width: 100,
      render: (result) => {
        const resultMap = {
          'verified': { color: 'success', text: '已验证' },
          'failed': { color: 'error', text: '失败' },
          'pending': { color: 'processing', text: '进行中' },
          'unverified': { color: 'default', text: '未验证' },
        };
        const config = resultMap[result] || resultMap['unverified'];
        return <Tag color={config.color}>{config.text}</Tag>;
      }
    },
    { 
      title: '状态', 
      dataIndex: 'status', 
      key: 'status',
      width: 80,
      render: (status) => (
        <Badge status={status === 'active' ? 'success' : 'default'} text={status === 'active' ? '启用' : '禁用'} />
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" onClick={() => handleEdit(record)}>配置</Button>
          <Button type="link" size="small" onClick={() => handleLightVerify(record)}>轻量验证</Button>
          <Button type="link" size="small" onClick={() => handleDeepVerify(record)}>深度验证</Button>
          <Button type="link" size="small" onClick={() => handleProbe(record)}>立即探测</Button>
        </Space>
      ),
    },
  ];

  // 筛选区域
  const FilterSection = () => (
    <Card size="small" style={{ marginBottom: 16 }}>
      <Form layout="inline">
        <Form.Item label="商户">
          <Input 
            placeholder="商户ID/名称" 
            value={filters.merchant_id}
            onChange={(e) => setFilters({ ...filters, merchant_id: e.target.value })}
            style={{ width: 150 }}
          />
        </Form.Item>
        <Form.Item label="BYOK类别">
          <Select 
            placeholder="选择类别" 
            allowClear
            value={filters.byok_type || undefined}
            onChange={(v) => setFilters({ ...filters, byok_type: v || '' })}
            style={{ width: 120 }}
          >
            <Select.Option value="official">官方</Select.Option>
            <Select.Option value="reseller">代理商</Select.Option>
            <Select.Option value="self_hosted">自建商</Select.Option>
          </Select>
        </Form.Item>
        <Form.Item label="厂商">
          <Select 
            placeholder="选择厂商" 
            allowClear
            value={filters.provider || undefined}
            onChange={(v) => setFilters({ ...filters, provider: v || '' })}
            style={{ width: 120 }}
          >
            {/* 厂商选项 */}
          </Select>
        </Form.Item>
        <Form.Item label="区域">
          <Select 
            placeholder="选择区域" 
            allowClear
            value={filters.region || undefined}
            onChange={(v) => setFilters({ ...filters, region: v || '' })}
            style={{ width: 100 }}
          >
            <Select.Option value="domestic">国内</Select.Option>
            <Select.Option value="overseas">海外</Select.Option>
          </Select>
        </Form.Item>
        <Form.Item label="路由模式">
          <Select 
            placeholder="选择模式" 
            allowClear
            value={filters.route_mode || undefined}
            onChange={(v) => setFilters({ ...filters, route_mode: v || '' })}
            style={{ width: 120 }}
          >
            <Select.Option value="auto">自动</Select.Option>
            <Select.Option value="direct">直连</Select.Option>
            <Select.Option value="litellm">LiteLLM</Select.Option>
            <Select.Option value="proxy">代理</Select.Option>
          </Select>
        </Form.Item>
        <Form.Item label="健康状态">
          <Select 
            placeholder="选择状态" 
            allowClear
            value={filters.health_status || undefined}
            onChange={(v) => setFilters({ ...filters, health_status: v || '' })}
            style={{ width: 100 }}
          >
            <Select.Option value="healthy">健康</Select.Option>
            <Select.Option value="degraded">降级</Select.Option>
            <Select.Option value="unhealthy">不健康</Select.Option>
            <Select.Option value="unknown">未知</Select.Option>
          </Select>
        </Form.Item>
        <Form.Item>
          <Button type="primary" onClick={handleSearch}>查询</Button>
        </Form.Item>
      </Form>
    </Card>
  );

  const handleEdit = (record) => {
    setEditingKey(record);
    form.setFieldsValue({
      route_mode: record.route_mode || 'auto',
      endpoint_url: record.endpoint_url,
      fallback_endpoint_url: record.fallback_endpoint_url,
      region: record.region,
      health_check_level: record.health_check_level || 'medium',
    });
    setModalVisible(true);
  };

  // 不同出站模式下的验证能力
  const handleLightVerify = async (record) => {
    // 轻量验证 - 快速检查 API Key 格式和基本有效性
    message.loading({ content: '正在执行轻量验证...', key: 'verify' });
    // 调用验证API
  };

  const handleDeepVerify = async (record) => {
    // 深度验证 - 完整验证 API Key 权限和配额
    message.loading({ content: '正在执行深度验证...', key: 'verify' });
    // 调用验证API
  };

  const handleProbe = async (record) => {
    // 立即探测 - 触发端点健康探测
    message.loading({ content: '正在执行健康探测...', key: 'probe' });
    // 调用探测API
  };

  return (
    <Card title="商户 API Key 运维管理">
      <FilterSection />
      <Table 
        columns={columns} 
        dataSource={keys} 
        rowKey="id" 
        scroll={{ x: 1500 }}
        pagination={{ showSizeChanger: true, showQuickJumper: true }}
      />
      
      <Modal
        title={`配置路由 - ${editingKey?.merchant_name} / ${editingKey?.provider}`}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={handleSave}
        width={600}
      >
        <Form form={form} layout="vertical">
          <Divider>路由配置（运维托管）</Divider>
          
          <Form.Item name="route_mode" label="路由模式">
            <Select>
              <Select.Option value="auto">自动（系统选择最优）</Select.Option>
              <Select.Option value="direct">直连</Select.Option>
              <Select.Option value="litellm">LiteLLM 网关</Select.Option>
              <Select.Option value="proxy">代理</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item name="region" label="区域">
            <Select>
              <Select.Option value="domestic">国内</Select.Option>
              <Select.Option value="overseas">海外</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item name="endpoint_url" label="主端点URL">
            <Input placeholder="自定义端点URL" />
          </Form.Item>

          <Form.Item name="fallback_endpoint_url" label="备用端点URL">
            <Input placeholder="备用端点URL（可选）" />
          </Form.Item>

          <Divider>健康探测配置</Divider>

          <Form.Item name="health_check_level" label="探测频率">
            <Select>
              <Select.Option value="high">高频（约每1分钟）</Select.Option>
              <Select.Option value="medium">中频（约每5分钟）</Select.Option>
              <Select.Option value="low">低频（约每30分钟）</Select.Option>
              <Select.Option value="daily">每日一次</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

// 样式文件（参考商户端实现）
const styles = {
  statusLightRow: 'status-light-row',
  statusDot: 'status-dot',
  statusLightLabel: 'status-light-label',
  dotGreen: 'dot-green',
  dotYellow: 'dot-yellow',
  dotRed: 'dot-red',
  dotGray: 'dot-gray',
};
```

**配套样式文件**: `frontend/src/pages/admin/AdminBYOKRouting.module.less`

```less
.status-light-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}

.dot-green {
  background-color: #52c41a;
  box-shadow: 0 0 4px #52c41a;
}

.dot-yellow {
  background-color: #faad14;
  box-shadow: 0 0 4px #faad14;
}

.dot-red {
  background-color: #ff4d4f;
  box-shadow: 0 0 4px #ff4d4f;
}

.dot-gray {
  background-color: #d9d9d9;
}

.status-light-label {
  font-size: 13px;
}
```

#### 4.3 BYOK 路由管理 API 接口（新增）

**文件**: `backend/handlers/admin_byok_routing.go`

新增 BYOK 路由管理接口：

```go
// GetBYOKRoutingList 获取所有商户的 API Key 路由列表（管理端）
func GetBYOKRoutingList(c *gin.Context) {
    // 支持按商户、厂商、状态筛选
    // 返回路由配置和健康状态
}

// UpdateBYOKRouteConfig 更新商户 API Key 路由配置（管理端）
func UpdateBYOKRouteConfig(c *gin.Context) {
    // 更新 route_mode, endpoint_url, fallback_endpoint_url, route_config
    // 更新 health_check_level
}

// TriggerBYOKProbe 触发 API Key 端点探测（管理端）
func TriggerBYOKProbe(c *gin.Context) {
    // 立即触发健康探测
}

// LightVerifyBYOK 轻量验证 API Key（管理端）
func LightVerifyBYOK(c *gin.Context) {
    // 快速检查 API Key 格式和基本有效性
}

// DeepVerifyBYOK 深度验证 API Key（管理端）
func DeepVerifyBYOK(c *gin.Context) {
    // 完整验证 API Key 权限和配额
}
```

**路由注册**: `backend/routes/routes.go`

```go
admin.GET("/byok-routing", handlers.GetBYOKRoutingList)
admin.PUT("/byok-routing/:id/route-config", handlers.UpdateBYOKRouteConfig)
admin.POST("/byok-routing/:id/probe", handlers.TriggerBYOKProbe)
admin.POST("/byok-routing/:id/light-verify", handlers.LightVerifyBYOK)
admin.POST("/byok-routing/:id/deep-verify", handlers.DeepVerifyBYOK)
```

#### 4.4 管理端厂商配置（简化路由配置）

**文件**: `frontend/src/pages/admin/AdminModelProviders.tsx`

简化厂商端的路由配置，明确其作为默认模板的职责：

**变更内容**：
1. 端点配置区域添加说明提示
2. 路由策略配置简化为只保留默认模板配置
3. 移除与商户级别配置冲突的功能

```tsx
<Alert
  message="端点配置说明"
  description="此配置为默认模板，用于新商户 API Key 的初始配置。商户的具体路由配置在「BYOK 路由管理」页面维护。"
  type="info"
  style={{ marginBottom: 16 }}
/>
```

> **注意**: 厂商端的路由配置仅作为默认模板，实际路由决策以 `merchant_api_keys` 表中的配置为准。

### Phase 5: 后端 API 更新

#### 5.1 商户 API Key 创建/更新 API（商户端）

**文件**: `backend/handlers/merchant_apikey.go`

商户端 API 需要支持 BYOK 类型字段：

```go
type CreateMerchantAPIKeyRequest struct {
    Provider        string   `json:"provider" binding:"required"`
    Name            string   `json:"name" binding:"required"`
    APIKey          string   `json:"api_key" binding:"required"`
    Description     string   `json:"description"`
    // BYOK 类型（新增）
    BYOKType        string   `json:"byok_type" binding:"required"`  // official, reseller, self_hosted
    Region          string   `json:"region"`                        // domestic, overseas
}

type UpdateMerchantAPIKeyRequest struct {
    Name            string   `json:"name"`
    Description     string   `json:"description"`
    // BYOK 类型（新增）
    BYOKType        string   `json:"byok_type"`
    Region          string   `json:"region"`
}

// 列表查询支持 BYOK 类型筛选
func GetMerchantAPIKeys(c *gin.Context) {
    // 支持筛选: byok_type, provider, status
    // 返回: 包含 byok_type, region 字段
}
```

#### 5.2 BYOK 路由管理 API（管理端）

**文件**: `backend/handlers/admin_byok_routing.go`

管理端 BYOK 路由管理接口，提供完整的路由配置能力：

```go
type UpdateBYOKRouteConfigRequest struct {
    RouteMode           string                 `json:"route_mode"`           // auto, direct, litellm, proxy
    EndpointURL         string                 `json:"endpoint_url"`         // 主端点
    FallbackEndpointURL string                 `json:"fallback_endpoint_url"` // 备用端点
    RouteConfig         map[string]interface{} `json:"route_config"`         // 完整配置
    HealthCheckLevel    string                 `json:"health_check_level"`   // high, medium, low
}

// GetBYOKRoutingList 获取所有商户的 API Key 路由列表
func GetBYOKRoutingList(c *gin.Context) {
    // 支持筛选: merchant_id, provider, status, health_status
    // 返回: 路由配置、健康状态、最后探测时间
}

// UpdateBYOKRouteConfig 更新路由配置
func UpdateBYOKRouteConfig(c *gin.Context) {
    // 验证配置有效性
    // 更新 merchant_api_keys 表
    // 记录操作日志
}

// TriggerBYOKProbe 触发端点探测
func TriggerBYOKProbe(c *gin.Context) {
    // 立即执行健康探测
    // 返回探测结果
}

// LightVerifyBYOK 轻量验证 API Key
func LightVerifyBYOK(c *gin.Context) {
    // 快速检查 API Key 格式和基本有效性
}

// DeepVerifyBYOK 深度验证 API Key
func DeepVerifyBYOK(c *gin.Context) {
    // 完整验证 API Key 权限和配额
}
```

> **说明**: 原有的 `route_config.go` 中的 `ProbeEndpoint` 函数用于厂商级别的端点探测，BYOK 路由管理的探测功能基于此实现，但增加了商户 API Key 上下文。

---

## 4. LiteLLM YAML 生成逻辑影响评估

### 4.1 当前 LiteLLM 配置生成逻辑

**文件**: `backend/cmd/litellm-catalog-sync/main.go`

数据来源：
- `model_providers` 表：`litellm_model_template`, `litellm_gateway_api_key_env`, `litellm_gateway_api_base`
- `spus` 表：活跃模型列表

生成逻辑：
```go
// 从 model_providers 获取厂商级别的网关配置
SELECT code, litellm_model_template, litellm_gateway_api_key_env, litellm_gateway_api_base
FROM model_providers WHERE status = 'active'

// 从 spus 获取模型列表
SELECT mp.code, sp.provider_model_id
FROM spus sp
INNER JOIN model_providers mp ON mp.code = sp.model_provider
WHERE sp.status = 'active'
```

### 4.2 影响分析结论

**结论：LiteLLM YAML 生成逻辑不需要修改**

原因：
1. **职责分离清晰**：
   - `model_providers.litellm_*` 字段：用于生成 LiteLLM YAML 配置（厂商级别）
   - `merchant_api_keys.route_*` 字段：用于路由决策（商户级别）

2. **数据层级不同**：
   - LiteLLM 是网关层面的配置，是厂商级别的
   - 商户的 API Key 是运行时使用的，不影响 LiteLLM 配置生成

3. **运行时 vs 配置时**：
   - LiteLLM YAML 是静态配置文件
   - `merchant_api_keys` 的路由配置用于运行时路由决策

### 4.3 两个表的职责分工

| 表 | LiteLLM 相关字段 | 路由决策相关字段 |
|---|---|---|
| `model_providers` | `litellm_model_template`, `litellm_gateway_api_key_env`, `litellm_gateway_api_base` | `route_strategy`, `endpoints`（作为 fallback） |
| `merchant_api_keys` | 无 | `route_mode`, `endpoint_url`, `fallback_endpoint_url`, `route_config`（SSOT） |

---

## 5. 兼容性处理

### 5.1 向后兼容

- 现有 `merchant_api_keys` 记录的 `route_mode` 默认为 `auto`
- 路由决策时，如果 API Key 没有配置，fallback 到 `model_providers` 的配置
- 现有 API 接口保持兼容

### 5.2 数据迁移

- 迁移脚本自动将 `model_providers` 的默认配置复制到 `merchant_api_keys.route_config`
- 不影响现有功能

---

## 6. 测试计划

### 6.1 单元测试

- `TestResolveEndpoint_WithAPIKeyConfig`: 测试 API Key 配置优先
- `TestResolveEndpoint_FallbackToProvider`: 测试 fallback 到厂商配置
- `TestRouteModeSelection`: 测试路由模式选择逻辑

### 6.2 集成测试

- 端到端路由决策测试
- API Key 创建/更新测试
- 端点探测测试

### 6.3 E2E 测试

- 商户配置自定义端点后请求成功
- 管理员修改厂商默认端点不影响已配置商户

---

## 7. 风险评估

| 风险 | 影响 | 缓解措施 |
|---|---|---|
| 数据迁移失败 | 中 | 在测试环境充分验证，提供回滚脚本 |
| 路由决策逻辑变更 | 高 | 保持 fallback 机制，灰度发布 |
| 前端界面变更 | 低 | 增量添加字段，不删除现有功能 |

---

## 8. 文件变更清单

### 后端

| 文件 | 变更类型 |
|---|---|
| `backend/migrations/069_route_config_ssot_to_api_keys.sql` | 新增 |
| `backend/models/models.go` | 修改 |
| `backend/handlers/merchant_apikey.go` | 修改（增加BYOK类型字段） |
| `backend/handlers/admin_byok_routing.go` | **新增** |
| `backend/routes/routes.go` | 修改 |
| `backend/services/execution_layer.go` | 修改 |
| `backend/handlers/route_config.go` | 已修复 |

### 前端

| 文件 | 变更类型 |
|---|---|
| `frontend/src/pages/merchant/MerchantAPIKeys.tsx` | 修改（增加BYOK类型） |
| `frontend/src/pages/admin/AdminBYOKRouting.tsx` | **新增** |
| `frontend/src/pages/admin/AdminBYOKRouting.module.less` | **新增** |
| `frontend/src/pages/admin/AdminModelProviders.tsx` | 修改（简化路由配置） |
| `frontend/src/types/merchant.ts` | 修改 |
| `frontend/src/services/admin.ts` | 修改 |

---

## 9. 预期成果

1. **SSOT 明确**：路由配置的唯一真相来源是 `merchant_api_keys`
2. **托管模式**：商户只需上传 API Key，运维工作由管理端托管
3. **灵活性提升**：每个商户可以独立配置路由模式和端点
4. **向后兼容**：现有功能不受影响，平滑迁移
5. **LiteLLM 无影响**：LiteLLM YAML 生成逻辑保持不变
