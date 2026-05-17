import { useEffect, useState } from 'react';
import {
  Card,
  Table,
  Button,
  Tag,
  Alert,
  Modal,
  Form,
  Input,
  Select,
  InputNumber,
  message,
  Space,
  Tabs,
  Divider,
  Tooltip,
  Spin,
} from 'antd';
import {
  EditOutlined,
  PlusOutlined,
  SettingOutlined,
  InfoCircleOutlined,
  QuestionCircleOutlined,
  GlobalOutlined,
  SyncOutlined,
  ApiOutlined,
  DiffOutlined,
} from '@ant-design/icons';
import { skuService } from '@/services/sku';
import { routeConfigService } from '@/services/routeConfig';
import { adminByokRoutingService, BYOKRoutingItem } from '@/services/adminByokRouting';
import type { ModelProvider } from '@/types/sku';
import RouteStrategyConfig from '@/components/admin/RouteStrategyConfig';
import EndpointsConfig from '@/components/admin/EndpointsConfig';

type ModalMode = 'create' | 'edit';

const AdminModelProviders = () => {
  const fallbackCode = '__default__';
  const [rows, setRows] = useState<ModelProvider[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [modalMode, setModalMode] = useState<ModalMode>('edit');
  const [editing, setEditing] = useState<ModelProvider | null>(null);
  const [form] = Form.useForm();
  const [activeTab, setActiveTab] = useState('basic');
  const isFallbackEditing = modalMode === 'edit' && editing?.code === fallbackCode;

  const [syncModalVisible, setSyncModalVisible] = useState(false);
  const [syncingProvider, setSyncingProvider] = useState<ModelProvider | null>(null);
  const [apiKeyList, setApiKeyList] = useState<BYOKRoutingItem[]>([]);
  const [apiKeyListLoading, setApiKeyListLoading] = useState(false);
  const [syncing, setSyncing] = useState(false);
  const [selectedAPIKeyID, setSelectedAPIKeyID] = useState<number | undefined>(undefined);

  const [gapModalVisible, setGapModalVisible] = useState(false);
  const [gapProvider, setGapProvider] = useState<ModelProvider | null>(null);
  const [gapLoading, setGapLoading] = useState(false);
  const [gapData, setGapData] = useState<
    Awaited<ReturnType<typeof skuService.getProviderCatalogGaps>>['data'] | null
  >(null);
  const [selectedPendingModels, setSelectedPendingModels] = useState<string[]>([]);
  const [creatingDrafts, setCreatingDrafts] = useState(false);
  const [syncingAll, setSyncingAll] = useState(false);

  const fetchList = async () => {
    setLoading(true);
    try {
      const res = await skuService.getAllModelProviders();
      setRows(res.data.data || []);
    } catch {
      message.error('获取模型厂商列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchList();
  }, []);

  const openCreate = () => {
    setModalMode('create');
    setEditing(null);
    setActiveTab('basic');
    form.resetFields();
    form.setFieldsValue({
      api_format: 'openai',
      billing_type: 'flat',
      status: 'active',
      sort_order: 0,
      compat_prefixes: [],
      provider_region: 'domestic',
      route_strategy: {},
      endpoints: {},
    });
    setModalVisible(true);
  };

  const handleEdit = async (record: ModelProvider) => {
    setModalMode('edit');
    setEditing(record);
    setActiveTab('basic');
    form.setFieldsValue({
      name: record.name,
      api_base_url: record.api_base_url ?? '',
      api_format: record.api_format,
      billing_type: record.billing_type ?? '',
      status: record.status,
      sort_order: record.sort_order,
      compat_prefixes: record.compat_prefixes?.length ? record.compat_prefixes : [],
      provider_region: record.provider_region || 'domestic',
      route_strategy: record.route_strategy || {},
      endpoints: record.endpoints || {},
      litellm_prefix: record.litellm_model_template
        ? record.litellm_model_template.replace('/{model_id}', '')
        : undefined,
      litellm_model_template: record.litellm_model_template || undefined,
      litellm_gateway_api_key_env: record.litellm_gateway_api_key_env || undefined,
      litellm_gateway_api_base: record.litellm_gateway_api_base || undefined,
    });
    setModalVisible(true);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (modalMode === 'create') {
        await skuService.createModelProvider({
          code: (values.code as string).trim().toLowerCase(),
          name: (values.name as string).trim(),
          api_base_url: values.api_base_url ? String(values.api_base_url).trim() : undefined,
          api_format: values.api_format,
          billing_type: values.billing_type ? String(values.billing_type).trim() : undefined,
          status: values.status,
          sort_order: values.sort_order as number,
          compat_prefixes: Array.isArray(values.compat_prefixes)
            ? (values.compat_prefixes as string[]).map((s) => String(s).trim()).filter(Boolean)
            : undefined,
          litellm_model_template: values.litellm_model_template?.trim() || undefined,
          litellm_gateway_api_key_env: values.litellm_gateway_api_key_env?.trim() || undefined,
          litellm_gateway_api_base: values.litellm_gateway_api_base?.trim() || undefined,
        });
        if (values.provider_region || values.route_strategy || values.endpoints) {
          try {
            await routeConfigService.updateProviderRouteConfig(
              (values.code as string).trim().toLowerCase(),
              {
                provider_region: values.provider_region,
                route_strategy: values.route_strategy,
                endpoints: values.endpoints,
              }
            );
          } catch {
            message.warning('厂商创建成功，但路由配置保存失败');
          }
        }
        message.success('已创建');
      } else {
        if (!editing) return;
        await skuService.patchModelProvider(editing.id, {
          name: values.name,
          api_base_url: values.api_base_url || undefined,
          api_format: values.api_format,
          billing_type: values.billing_type || undefined,
          status: values.status,
          sort_order: values.sort_order,
          compat_prefixes: Array.isArray(values.compat_prefixes)
            ? (values.compat_prefixes as string[]).map((s) => String(s).trim()).filter(Boolean)
            : [],
          litellm_model_template: values.litellm_model_template?.trim() || '',
          litellm_gateway_api_key_env: values.litellm_gateway_api_key_env?.trim() || '',
          litellm_gateway_api_base: values.litellm_gateway_api_base?.trim() || '',
        });
        if (!isFallbackEditing) {
          try {
            await routeConfigService.updateProviderRouteConfig(editing.code, {
              provider_region: values.provider_region,
              route_strategy: values.route_strategy,
              endpoints: values.endpoints,
            });
          } catch {
            message.warning('基础信息已保存，但路由配置保存失败');
          }
        }
        message.success('已保存');
      }
      setModalVisible(false);
      setEditing(null);
      fetchList();
    } catch (e: unknown) {
      if (e && typeof e === 'object' && 'errorFields' in e) return;
      message.error(modalMode === 'create' ? '创建失败' : '保存失败');
    }
  };

  const handleOpenSyncModal = async (record: ModelProvider) => {
    setSyncingProvider(record);
    setSelectedAPIKeyID(undefined);
    setSyncModalVisible(true);
    setApiKeyListLoading(true);

    try {
      const response = await adminByokRoutingService.getByokRoutingList({
        provider: record.code,
      });
      const keys = response.data.data || [];
      setApiKeyList(keys);
      if (keys.length > 0) {
        setSelectedAPIKeyID(keys[0].id);
      }
    } catch {
      message.error('获取API Key列表失败');
      setApiKeyList([]);
    } finally {
      setApiKeyListLoading(false);
    }
  };

  const handleSyncModels = async () => {
    if (!syncingProvider) return;

    setSyncing(true);
    try {
      const response = await skuService.syncProviderModels(syncingProvider.code, selectedAPIKeyID);
      message.success(`同步成功，共同步 ${response.data.synced_count} 个模型`);
      setSyncModalVisible(false);
      const provider = syncingProvider;
      setSyncingProvider(null);
      if (provider) {
        await handleOpenCatalogGapModal(provider);
      }
    } catch (error) {
      message.error(error instanceof Error ? error.message : '同步失败');
    } finally {
      setSyncing(false);
    }
  };

  const handleOpenCatalogGapModal = async (record: ModelProvider) => {
    setGapProvider(record);
    setGapModalVisible(true);
    setGapLoading(true);
    setSelectedPendingModels([]);
    try {
      const response = await skuService.getProviderCatalogGaps(record.code);
      setGapData(response.data);
      setSelectedPendingModels(
        (response.data.pending_onboard || []).map((m) => m.model_id).slice(0, 20)
      );
    } catch {
      message.error('获取目录差异失败');
      setGapData(null);
    } finally {
      setGapLoading(false);
    }
  };

  const handleCreateSPUDrafts = async () => {
    if (!gapProvider || selectedPendingModels.length === 0) {
      message.warning('请选择待上架模型');
      return;
    }
    setCreatingDrafts(true);
    try {
      const res = await skuService.createProviderSPUDrafts(gapProvider.code, selectedPendingModels);
      const created = (res.data.results || []).filter((r) => r.spu_id).length;
      message.success(`已创建 ${created} 个 SPU 草稿（inactive）`);
      await handleOpenCatalogGapModal(gapProvider);
    } catch {
      message.error('创建 SPU 草稿失败');
    } finally {
      setCreatingDrafts(false);
    }
  };

  const handleSyncAllProviders = async () => {
    setSyncingAll(true);
    try {
      const res = await skuService.syncAllProviderModels();
      const results = res.data.results || {};
      const ok = Object.values(results).filter((n) => n >= 0).length;
      message.success(`全量同步完成：${ok} 个厂商成功`);
      await fetchList();
    } catch {
      message.error('全量同步失败');
    } finally {
      setSyncingAll(false);
    }
  };

  const columns = [
    {
      title: '代码',
      dataIndex: 'code',
      key: 'code',
      width: 120,
      fixed: 'left' as const,
      render: (code: string) =>
        code === fallbackCode ? (
          <Space size={6}>
            <span>{code}</span>
            <Tag color="purple">兜底</Tag>
          </Space>
        ) : (
          code
        ),
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 120,
    },
    {
      title: '区域',
      dataIndex: 'provider_region',
      key: 'provider_region',
      width: 80,
      render: (v: string) => (
        <Tag color={v === 'overseas' ? 'blue' : 'green'} icon={<GlobalOutlined />}>
          {v === 'overseas' ? '海外' : '国内'}
        </Tag>
      ),
    },
    {
      title: 'API Base URL',
      dataIndex: 'api_base_url',
      key: 'api_base_url',
      ellipsis: true,
      render: (v: string) => v || '—',
    },
    {
      title: 'API 格式',
      dataIndex: 'api_format',
      key: 'api_format',
      width: 120,
    },
    {
      title: '路由配置',
      key: 'route_config',
      width: 100,
      render: (_: unknown, record: ModelProvider) => {
        const hasRouteConfig =
          record.route_strategy && Object.keys(record.route_strategy || {}).length > 0;
        return hasRouteConfig ? <Tag color="blue">已配置</Tag> : <Tag color="default">未配置</Tag>;
      },
    },
    {
      title: 'LiteLLM',
      key: 'litellm_config',
      width: 100,
      render: (_: unknown, record: ModelProvider) => {
        if (!record.litellm_model_template) return <Tag color="default">未配置</Tag>;
        const prefix = record.litellm_model_template.replace('/{model_id}', '');
        return <Tag color="green">{prefix}</Tag>;
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (s: string) => (s === 'active' ? <Tag color="green">启用</Tag> : <Tag>停用</Tag>),
    },
    {
      title: '排序',
      dataIndex: 'sort_order',
      key: 'sort_order',
      width: 60,
    },
    {
      title: '操作',
      key: 'action',
      width: 220,
      fixed: 'right' as const,
      render: (_: unknown, record: ModelProvider) => (
        <Space size="small" wrap>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          {record.code !== fallbackCode && (
            <>
              <Button
                type="link"
                size="small"
                icon={<SyncOutlined />}
                onClick={() => handleOpenSyncModal(record)}
              >
                同步模型
              </Button>
              <Button
                type="link"
                size="small"
                icon={<DiffOutlined />}
                onClick={() => handleOpenCatalogGapModal(record)}
              >
                目录差异
              </Button>
            </>
          )}
        </Space>
      ),
    },
  ];

  const apiFormatOptions = [
    { value: 'openai', label: 'openai（OpenAI 兼容 /chat/completions）' },
    { value: 'anthropic', label: 'anthropic（/messages + x-api-key）' },
    { value: 'baidu', label: 'baidu（千帆等，需与代理实现一致）' },
  ];

  const commonLitellmPrefixes = [
    { value: 'openai', label: 'openai（OpenAI）' },
    { value: 'anthropic', label: 'anthropic（Anthropic）' },
    { value: 'deepseek', label: 'deepseek（DeepSeek）' },
    { value: 'zai', label: 'zai（智谱 Z.AI 海外）' },
    { value: 'dashscope', label: 'dashscope（阿里通义）' },
    { value: 'gemini', label: 'gemini（Google）' },
    { value: 'moonshot', label: 'moonshot（Moonshot）' },
    { value: 'minimax', label: 'minimax（MiniMax）' },
  ];

  const deriveLitellmConfig = (prefix: string) => {
    const p = (prefix || '').trim().toLowerCase();
    if (!p) return { template: '', envVar: '', apiBase: '' };
    return {
      template: `${p}/{model_id}`,
      envVar: `${p.toUpperCase().replace(/-/g, '_')}_API_KEY`,
      apiBase: '',
    };
  };

  const BasicInfoTab = () => (
    <>
      {modalMode === 'create' ? (
        <Form.Item
          name="code"
          label="代码（唯一，小写字母/数字/下划线）"
          rules={[
            { required: true, message: '请输入代码' },
            {
              pattern: /^[a-z][a-z0-9_]{0,48}$/,
              message: '须以小写字母开头，仅 a-z、0-9、下划线，最长 50 字符',
            },
            {
              validator: async (_rule, value) => {
                if (
                  String(value || '')
                    .trim()
                    .toLowerCase() === fallbackCode
                ) {
                  throw new Error('__default__ 为系统保留兜底代码，请勿新建');
                }
              },
            },
          ]}
        >
          <Input placeholder="例如 minimax" autoComplete="off" />
        </Form.Item>
      ) : (
        editing && (
          <Form.Item label="代码（只读）">
            <Input value={editing.code} disabled />
          </Form.Item>
        )
      )}
      <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
        <Input disabled={isFallbackEditing} />
      </Form.Item>
      <Form.Item name="api_base_url" label="API Base URL">
        <Input placeholder="例如 https://api.openai.com/v1" disabled={isFallbackEditing} />
      </Form.Item>
      <Form.Item name="api_format" label="API 格式" rules={[{ required: true }]}>
        <Select options={apiFormatOptions} placeholder="选择格式" disabled={isFallbackEditing} />
      </Form.Item>
      <Alert
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
        message="商户 SKU 与 Anthropic（Claude）出站"
        description={
          <div>
            <p>
              若同一订阅既要 OpenAI 兼容又要走 Anthropic Messages 出站，除目录
              <strong>主厂商</strong>（如 <code>alibaba</code>）外，请再建一条厂商代码为「主厂商小写
              + <code>_anthropic</code>」的记录（例如 <code>alibaba_anthropic</code>），且 API
              格式选 <strong>anthropic</strong>。
            </p>
            <p style={{ marginTop: 8 }}>
              商户上架 SKU 时可在「Anthropic 出站 Key」绑定 <code>provider</code> 为该代码的 API
              Key；strict 权益白名单会合并该槽位。
            </p>
            <p style={{ marginTop: 8, marginBottom: 0 }}>
              完整检查清单与配置步骤：{' '}
              <a
                href="https://github.com/sev7n4/pintuotuo/blob/main/documentation/capability/anthropic-sibling-provider-setup.md"
                target="_blank"
                rel="noopener noreferrer"
              >
                anthropic-sibling-provider-setup.md
              </a>
            </p>
          </div>
        }
      />
      <Form.Item
        name="compat_prefixes"
        label="OpenAI 兼容前缀（无前缀 model 名匹配）"
        tooltip="小写字母/数字/点/下划线/连字符；用于 /openai/v1 下仅写模型名时路由到本厂商。可多个，最长匹配优先。"
      >
        <Select
          mode="tags"
          placeholder="输入后回车添加，如 deepseek、glm-"
          tokenSeparators={[',']}
          disabled={isFallbackEditing}
        />
      </Form.Item>
      <Form.Item name="billing_type" label="计费类型">
        <Input placeholder="可选，默认 flat" disabled={isFallbackEditing} />
      </Form.Item>
      <Form.Item name="status" label="状态" rules={[{ required: true }]}>
        <Select
          options={[
            { value: 'active', label: '启用' },
            { value: 'inactive', label: '停用' },
          ]}
        />
      </Form.Item>
      <Form.Item name="sort_order" label="排序" rules={[{ required: true }]}>
        <InputNumber min={0} style={{ width: '100%' }} disabled={isFallbackEditing} />
      </Form.Item>
    </>
  );

  const RouteStrategyTab = () => (
    <>
      <Alert
        message="路由策略配置"
        description="配置不同用户类型的路由策略，系统会根据商户类型和区域自动选择最优路由模式。"
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />
      <Form.Item
        name="provider_region"
        label={
          <Space>
            <span>厂商区域</span>
            <Tooltip title="标识厂商服务器所在区域，影响路由决策">
              <QuestionCircleOutlined style={{ color: '#999' }} />
            </Tooltip>
          </Space>
        }
      >
        <Select
          options={[
            { value: 'domestic', label: '国内' },
            { value: 'overseas', label: '海外' },
          ]}
        />
      </Form.Item>
      <Divider orientation="left">
        <Space>
          <span>路由策略</span>
          <Tooltip title="为不同用户类型配置路由策略">
            <InfoCircleOutlined style={{ color: '#1890ff' }} />
          </Tooltip>
        </Space>
      </Divider>
      <Form.Item name="route_strategy" noStyle>
        <RouteStrategyConfig providerRegion={form.getFieldValue('provider_region')} />
      </Form.Item>
    </>
  );

  const EndpointsTab = () => (
    <>
      <Alert
        message="端点配置（默认模板）"
        description={
          <div>
            <p>
              此配置为<strong>厂商级默认模板</strong>，用于生成 LiteLLM YAML 配置和作为商户 API Key
              的回退端点。
            </p>
            <p style={{ marginTop: 8 }}>
              <InfoCircleOutlined style={{ marginRight: 4 }} />
              商户上传的 API Key 可通过 BYOK 路由管理页面配置独立端点，优先级高于此默认模板。
            </p>
          </div>
        }
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />
      <Form.Item name="endpoints" noStyle>
        <EndpointsConfig providerCode={editing?.code} />
      </Form.Item>
    </>
  );

  const LitellmConfigTab = () => {
    const litellmPrefix = Form.useWatch('litellm_prefix', form);
    const derived = deriveLitellmConfig(litellmPrefix);

    const handlePrefixChange = (prefix: string) => {
      const d = deriveLitellmConfig(prefix);
      form.setFieldsValue({
        litellm_prefix: prefix,
        litellm_model_template: d.template || undefined,
        litellm_gateway_api_key_env: d.envVar || undefined,
      });
    };

    return (
      <>
        <Alert
          message="LiteLLM 网关配置"
          description={
            <div>
              <p>
                配置厂商在 LiteLLM 网关中的模型名称映射。只需填写 LiteLLM Provider
                前缀，系统会自动生成模板和环境变量名。
              </p>
              <p style={{ marginTop: 8 }}>
                <InfoCircleOutlined style={{ marginRight: 4 }} />
                此配置为 BYOK 深度验证和 LiteLLM 路由模式的核心数据源（SSOT）。
              </p>
            </div>
          }
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />

        <Form.Item
          name="litellm_prefix"
          label={
            <Space>
              <span>LiteLLM Provider 前缀</span>
              <Tooltip title="输入 LiteLLM 官方前缀，系统会自动推导模型模板和环境变量名。如不确定，请查阅 litellm.ai/docs/providers">
                <QuestionCircleOutlined style={{ color: '#999' }} />
              </Tooltip>
            </Space>
          }
        >
          <Select
            showSearch
            allowClear
            placeholder="选择或输入前缀"
            options={commonLitellmPrefixes}
            onChange={handlePrefixChange}
            disabled={isFallbackEditing}
            filterOption={(input, option) =>
              (option?.label as string)?.toLowerCase().includes(input.toLowerCase()) ||
              (option?.value as string)?.toLowerCase().includes(input.toLowerCase())
            }
          />
        </Form.Item>

        {derived.template && (
          <Alert
            style={{ marginBottom: 16 }}
            type="success"
            showIcon
            message="自动推导预览"
            description={
              <div style={{ fontSize: 13 }}>
                <div>
                  模型模板：
                  <code style={{ background: '#f5f5f5', padding: '2px 6px', borderRadius: 3 }}>
                    {derived.template}
                  </code>
                </div>
                <div style={{ marginTop: 4 }}>
                  环境变量：
                  <code style={{ background: '#f5f5f5', padding: '2px 6px', borderRadius: 3 }}>
                    {derived.envVar}
                  </code>
                </div>
                <div style={{ marginTop: 4 }}>
                  示例：
                  <code style={{ background: '#f5f5f5', padding: '2px 6px', borderRadius: 3 }}>
                    glm-4.5 → {derived.template.replace('{model_id}', 'glm-4.5')}
                  </code>
                </div>
              </div>
            }
          />
        )}

        <Form.Item
          name="litellm_model_template"
          label={
            <Space>
              <span>模型模板</span>
              <Tooltip title="自动生成，通常无需修改。格式：prefix/{model_id}">
                <QuestionCircleOutlined style={{ color: '#999' }} />
              </Tooltip>
            </Space>
          }
        >
          <Input placeholder="例如 zai/{model_id}" disabled={isFallbackEditing} />
        </Form.Item>

        <Form.Item
          name="litellm_gateway_api_key_env"
          label={
            <Space>
              <span>API Key 环境变量名</span>
              <Tooltip title="LiteLLM 容器内读取 API Key 的环境变量名，自动生成，通常无需修改">
                <QuestionCircleOutlined style={{ color: '#999' }} />
              </Tooltip>
            </Space>
          }
        >
          <Input placeholder="例如 ZAI_API_KEY" disabled={isFallbackEditing} />
        </Form.Item>

        <Form.Item
          name="litellm_gateway_api_base"
          label={
            <Space>
              <span>自定义 API Base</span>
              <Tooltip title="仅 OpenAI 兼容厂商需要填写（如阶跃星辰），大多数厂商留空即可">
                <QuestionCircleOutlined style={{ color: '#999' }} />
              </Tooltip>
            </Space>
          }
        >
          <Input
            placeholder="仅 OpenAI 兼容厂商需要，如 https://api.stepfun.com/v1"
            disabled={isFallbackEditing}
          />
        </Form.Item>
      </>
    );
  };

  const tabItems = [
    { key: 'basic', label: '基础信息', children: <BasicInfoTab /> },
    {
      key: 'route',
      label: '路由策略',
      children: <RouteStrategyTab />,
      disabled: isFallbackEditing,
    },
    {
      key: 'endpoints',
      label: '端点配置',
      children: <EndpointsTab />,
      disabled: isFallbackEditing,
    },
    {
      key: 'litellm',
      label: (
        <Space size={4}>
          <ApiOutlined />
          LiteLLM 配置
        </Space>
      ),
      children: <LitellmConfigTab />,
      disabled: isFallbackEditing,
    },
  ];

  return (
    <div>
      <Card
        title="厂商配置"
        extra={
          <Space>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openCreate()}>
              新增厂商
            </Button>
            <Button loading={syncingAll} icon={<SyncOutlined />} onClick={handleSyncAllProviders}>
              全量同步模型
            </Button>
            <Button onClick={() => fetchList()}>刷新</Button>
          </Space>
        }
      >
        <Alert
          type="info"
          showIcon
          style={{ marginBottom: 12 }}
          message="统一厂商配置"
          description="在此页面可配置厂商基础信息、路由策略和端点地址。同步模型后可用「目录差异」对比厂商列表与 SPU，一键生成 inactive 草稿 SPU。定时任务需设置环境变量 PROVIDER_MODEL_SYNC_ENABLED=true。"
        />
        <Table
          rowKey="id"
          loading={loading}
          columns={columns}
          dataSource={rows}
          scroll={{ x: 1200 }}
          pagination={{ pageSize: 50, showSizeChanger: true }}
        />
      </Card>

      <Modal
        title={
          <Space>
            {modalMode === 'create' ? '新增厂商' : editing ? `编辑：${editing.code}` : '编辑'}
            {editing && editing.code !== fallbackCode && (
              <Tag color="blue">
                <SettingOutlined /> 统一配置
              </Tag>
            )}
          </Space>
        }
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => {
          setModalVisible(false);
          setEditing(null);
        }}
        destroyOnClose
        width={900}
        style={{ top: 20 }}
        bodyStyle={{ maxHeight: '70vh', overflowY: 'auto' }}
      >
        <Form form={form} layout="vertical">
          <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} />
        </Form>
      </Modal>

      <Modal
        title={
          <Space>
            <SyncOutlined />
            同步模型 - {syncingProvider?.name || syncingProvider?.code}
          </Space>
        }
        open={syncModalVisible}
        onOk={handleSyncModels}
        onCancel={() => {
          setSyncModalVisible(false);
          setSyncingProvider(null);
        }}
        confirmLoading={syncing}
        okText="开始同步"
        cancelText="取消"
        width={600}
      >
        <Alert
          message="同步厂商模型列表"
          description={
            <div>
              <p>从厂商的 /v1/models 接口获取最新的模型列表，并更新到系统数据库中。</p>
              <p style={{ marginTop: 8 }}>
                <InfoCircleOutlined style={{ marginRight: 4 }} />
                需要选择一个有效的 API Key 用于认证，如果该厂商没有可用的 API Key，请先在 BYOK
                路由管理中添加。
              </p>
            </div>
          }
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />

        {apiKeyListLoading ? (
          <div style={{ textAlign: 'center', padding: '20px 0' }}>
            <Spin />
            <div style={{ marginTop: 8, color: '#999' }}>加载 API Key 列表...</div>
          </div>
        ) : apiKeyList.length === 0 ? (
          <Alert
            message="无可用的 API Key"
            description={
              <div>
                <p>该厂商没有可用的 BYOK API Key，无法同步模型列表。</p>
                <p style={{ marginTop: 8 }}>
                  请先在 <a href="/admin/byok-routing">BYOK 路由管理</a> 中为该厂商添加 API Key。
                </p>
              </div>
            }
            type="warning"
            showIcon
          />
        ) : (
          <Form layout="vertical">
            <Form.Item
              label="选择 API Key"
              required
              tooltip="选择用于认证的 API Key，将使用该 Key 访问厂商的 /v1/models 接口"
            >
              <Select
                value={selectedAPIKeyID}
                onChange={setSelectedAPIKeyID}
                placeholder="请选择 API Key"
                options={apiKeyList.map((key) => ({
                  value: key.id,
                  label: (
                    <Space>
                      <span>{key.name}</span>
                      <Tag color={key.health_status === 'healthy' ? 'green' : 'default'}>
                        {key.health_status === 'healthy' ? '健康' : key.health_status}
                      </Tag>
                      <Tag color="blue">{key.company_name}</Tag>
                    </Space>
                  ),
                }))}
              />
            </Form.Item>
          </Form>
        )}
      </Modal>

      <Modal
        title={gapProvider ? `目录差异：${gapProvider.name}` : '目录差异'}
        open={gapModalVisible}
        onCancel={() => {
          setGapModalVisible(false);
          setGapProvider(null);
          setGapData(null);
        }}
        width={720}
        footer={[
          <Button key="close" onClick={() => setGapModalVisible(false)}>
            关闭
          </Button>,
          <Button
            key="draft"
            type="primary"
            loading={creatingDrafts}
            disabled={selectedPendingModels.length === 0}
            onClick={handleCreateSPUDrafts}
          >
            生成 SPU 草稿
          </Button>,
        ]}
      >
        {gapLoading ? (
          <div style={{ textAlign: 'center', padding: 24 }}>
            <Spin />
          </div>
        ) : gapData ? (
          <>
            <Alert
              type="info"
              showIcon
              style={{ marginBottom: 12 }}
              message={`厂商快照 ${gapData.provider_model_count} 个 · 目录 SPU ${gapData.spu_model_count} 个`}
              description={
                gapData.last_synced_at
                  ? `最近同步：${new Date(gapData.last_synced_at).toLocaleString()}`
                  : '尚未同步厂商模型，请先「同步模型」'
              }
            />
            {(gapData.stale_spus?.length ?? 0) > 0 && (
              <Alert
                type="warning"
                showIcon
                style={{ marginBottom: 12 }}
                message={`${gapData.stale_spus.length} 个在售 SPU 已不在厂商最新列表中`}
                description={
                  <ul style={{ margin: 0, paddingLeft: 18 }}>
                    {gapData.stale_spus.slice(0, 8).map((s) => (
                      <li key={s.spu_id}>
                        {s.spu_code} · {s.model_id}
                        {s.active_sku_count > 0 ? `（${s.active_sku_count} 个在售 SKU）` : ''}
                      </li>
                    ))}
                  </ul>
                }
              />
            )}
            <div style={{ marginBottom: 8, fontWeight: 500 }}>待上架（厂商有、SPU 无）</div>
            {(gapData.pending_onboard?.length ?? 0) === 0 ? (
              <Alert type="success" message="无待上架模型" showIcon />
            ) : (
              <Select
                mode="multiple"
                style={{ width: '100%' }}
                placeholder="选择要生成草稿的模型"
                value={selectedPendingModels}
                onChange={setSelectedPendingModels}
                options={(gapData.pending_onboard || []).map((m) => ({
                  value: m.model_id,
                  label: m.display_name ? `${m.model_id} (${m.display_name})` : m.model_id,
                }))}
              />
            )}
          </>
        ) : null}
      </Modal>
    </div>
  );
};

export default AdminModelProviders;
