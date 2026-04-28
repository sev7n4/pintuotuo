import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Card,
  Table,
  Button,
  Tag,
  Space,
  Modal,
  message,
  Typography,
  Form,
  Select,
  Input,
  Tooltip,
  Popconfirm,
  Spin,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  ReloadOutlined,
  SettingOutlined,
  ThunderboltOutlined,
  SafetyCertificateOutlined,
  SearchOutlined,
} from '@ant-design/icons';
import {
  adminByokRoutingService,
  BYOKRoutingItem,
  UpdateRouteConfigRequest,
} from '@/services/adminByokRouting';

const { Title } = Typography;

const byokTypeTag = (byokType: string) => {
  const colorMap: Record<string, string> = {
    official: 'blue',
    reseller: 'orange',
    self_hosted: 'purple',
  };
  const labelMap: Record<string, string> = {
    official: '官方',
    reseller: '代理商',
    self_hosted: '自建商',
  };
  return <Tag color={colorMap[byokType] || 'default'}>{labelMap[byokType] || byokType}</Tag>;
};

const routeModeTag = (routeMode: string) => {
  const colorMap: Record<string, string> = {
    auto: 'cyan',
    direct: 'green',
    litellm: 'blue',
    proxy: 'purple',
  };
  const labelMap: Record<string, string> = {
    auto: '自动',
    direct: '直连',
    litellm: 'LiteLLM',
    proxy: '代理',
  };
  return <Tag color={colorMap[routeMode] || 'default'}>{labelMap[routeMode] || routeMode}</Tag>;
};

const healthStatusTag = (healthStatus: string) => {
  const colorMap: Record<string, string> = {
    healthy: 'success',
    degraded: 'warning',
    unhealthy: 'error',
    unknown: 'default',
  };
  const labelMap: Record<string, string> = {
    healthy: '健康',
    degraded: '降级',
    unhealthy: '不健康',
    unknown: '未知',
  };
  return <Tag color={colorMap[healthStatus] || 'default'}>{labelMap[healthStatus] || healthStatus}</Tag>;
};

const regionTag = (region: string) => {
  const colorMap: Record<string, string> = {
    domestic: 'green',
    overseas: 'blue',
  };
  const labelMap: Record<string, string> = {
    domestic: '国内',
    overseas: '海外',
  };
  return <Tag color={colorMap[region] || 'default'}>{labelMap[region] || region}</Tag>;
};

const AdminByokRouting = () => {
  const [data, setData] = useState<BYOKRoutingItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);

  const [byokTypeFilter, setByokTypeFilter] = useState<string>('');
  const [providerFilter, setProviderFilter] = useState<string>('');
  const [regionFilter, setRegionFilter] = useState<string>('');
  const [routeModeFilter, setRouteModeFilter] = useState<string>('');
  const [healthFilter, setHealthFilter] = useState<string>('');
  const [keywordFilter, setKeywordFilter] = useState<string>('');

  const [configModalVisible, setConfigModalVisible] = useState(false);
  const [configLoading, setConfigLoading] = useState(false);
  const [selectedItem, setSelectedItem] = useState<BYOKRoutingItem | null>(null);
  const [configForm] = Form.useForm();

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const response = await adminByokRoutingService.getByokRoutingList({
        byok_type: byokTypeFilter || undefined,
        provider: providerFilter || undefined,
        region: regionFilter || undefined,
        route_mode: routeModeFilter || undefined,
        health_status: healthFilter || undefined,
      });
      let items = response.data.data || [];
      if (keywordFilter.trim()) {
        const kw = keywordFilter.trim().toLowerCase();
        items = items.filter(
          (item) =>
            item.name.toLowerCase().includes(kw) ||
            item.company_name.toLowerCase().includes(kw) ||
            item.provider.toLowerCase().includes(kw)
        );
      }
      setData(items);
      setTotal(response.data.total);
    } catch {
      message.error('获取BYOK路由列表失败');
    } finally {
      setLoading(false);
    }
  }, [byokTypeFilter, providerFilter, regionFilter, routeModeFilter, healthFilter, keywordFilter]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const providerOptions = useMemo(() => {
    const providers = new Set(data.map((item) => item.provider));
    return Array.from(providers).sort();
  }, [data]);

  const handleOpenConfig = (record: BYOKRoutingItem) => {
    setSelectedItem(record);
    configForm.setFieldsValue({
      route_mode: record.route_mode || 'auto',
      endpoint_url: record.endpoint_url || '',
      fallback_endpoint_url: record.fallback_endpoint_url || '',
    });
    setConfigModalVisible(true);
  };

  const handleSaveConfig = async () => {
    if (!selectedItem) return;
    const values = await configForm.validateFields();
    setConfigLoading(true);
    try {
      const payload: UpdateRouteConfigRequest = {
        route_mode: values.route_mode,
        endpoint_url: values.endpoint_url?.trim() || '',
        fallback_endpoint_url: values.fallback_endpoint_url?.trim() || '',
      };
      await adminByokRoutingService.updateRouteConfig(selectedItem.id, payload);
      message.success('路由配置更新成功');
      setConfigModalVisible(false);
      fetchData();
    } catch {
      message.error('更新路由配置失败');
    } finally {
      setConfigLoading(false);
    }
  };

  const handleTriggerProbe = async (id: number) => {
    try {
      await adminByokRoutingService.triggerProbe(id);
      message.success('探测已触发');
    } catch {
      message.error('触发探测失败');
    }
  };

  const handleLightVerify = async (id: number) => {
    try {
      await adminByokRoutingService.triggerLightVerify(id);
      message.success('轻量验证已触发');
    } catch {
      message.error('触发轻量验证失败');
    }
  };

  const resetFilters = () => {
    setByokTypeFilter('');
    setProviderFilter('');
    setRegionFilter('');
    setRouteModeFilter('');
    setHealthFilter('');
    setKeywordFilter('');
  };

  const columns: ColumnsType<BYOKRoutingItem> = [
    {
      title: '商户',
      dataIndex: 'company_name',
      key: 'company_name',
      width: 150,
      ellipsis: true,
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 120,
      ellipsis: true,
    },
    {
      title: '提供商',
      dataIndex: 'provider',
      key: 'provider',
      width: 80,
      render: (provider: string) => <Tag color="blue">{provider.toUpperCase()}</Tag>,
    },
    {
      title: 'BYOK类型',
      dataIndex: 'byok_type',
      key: 'byok_type',
      width: 90,
      render: (byokType: string) => byokTypeTag(byokType),
    },
    {
      title: '区域',
      dataIndex: 'region',
      key: 'region',
      width: 70,
      render: (region: string) => regionTag(region),
    },
    {
      title: '路由模式',
      dataIndex: 'route_mode',
      key: 'route_mode',
      width: 90,
      render: (routeMode: string) => routeModeTag(routeMode),
    },
    {
      title: '端点URL',
      dataIndex: 'endpoint_url',
      key: 'endpoint_url',
      width: 180,
      ellipsis: true,
      render: (url: string) =>
        url ? (
          <Tooltip title={url}>
            <span style={{ fontSize: 12 }}>{url}</span>
          </Tooltip>
        ) : (
          <span style={{ color: '#999' }}>—</span>
        ),
    },
    {
      title: '健康状态',
      dataIndex: 'health_status',
      key: 'health_status',
      width: 90,
      render: (status: string) => healthStatusTag(status),
    },
    {
      title: '验证状态',
      dataIndex: 'verification_result',
      key: 'verification_result',
      width: 90,
      render: (result: string) => {
        const colorMap: Record<string, string> = {
          success: 'success',
          failed: 'error',
          in_progress: 'processing',
          pending: 'default',
        };
        const labelMap: Record<string, string> = {
          success: '已验证',
          failed: '失败',
          in_progress: '验证中',
          pending: '未验证',
        };
        return <Tag color={colorMap[result] || 'default'}>{labelMap[result] || result}</Tag>;
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 70,
      render: (status: string) => (
        <Tag color={status === 'active' ? 'green' : 'red'}>
          {status === 'active' ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      fixed: 'right',
      render: (_, record) => (
        <Space size="small">
          <Button size="small" icon={<SettingOutlined />} onClick={() => handleOpenConfig(record)}>
            配置
          </Button>
          <Popconfirm title="确定触发探测？" onConfirm={() => handleTriggerProbe(record.id)}>
            <Button size="small" icon={<ThunderboltOutlined />}>
              探测
            </Button>
          </Popconfirm>
          <Popconfirm title="确定触发轻量验证？" onConfirm={() => handleLightVerify(record.id)}>
            <Button size="small" icon={<SafetyCertificateOutlined />}>
              轻验
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Card>
        <Space direction="vertical" style={{ width: '100%' }} size="middle">
          <div style={{ display: 'flex', justifyContent: 'space-between', width: '100%' }}>
            <Title level={4} style={{ margin: 0 }}>
              BYOK 路由管理
            </Title>
            <Button icon={<ReloadOutlined />} onClick={fetchData} loading={loading}>
              刷新
            </Button>
          </div>

          <Space wrap size={12}>
            <Input
              allowClear
              style={{ width: 180 }}
              placeholder="搜索商户/名称/提供商"
              prefix={<SearchOutlined />}
              value={keywordFilter}
              onChange={(e) => setKeywordFilter(e.target.value)}
            />
            <Select
              style={{ width: 130 }}
              value={byokTypeFilter}
              onChange={setByokTypeFilter}
              allowClear
              placeholder="BYOK类型"
              options={[
                { value: 'official', label: '官方' },
                { value: 'reseller', label: '代理商' },
                { value: 'self_hosted', label: '自建商' },
              ]}
            />
            <Select
              style={{ width: 120 }}
              value={providerFilter}
              onChange={setProviderFilter}
              allowClear
              placeholder="提供商"
              options={providerOptions.map((p) => ({ value: p.toLowerCase(), label: p.toUpperCase() }))}
            />
            <Select
              style={{ width: 100 }}
              value={regionFilter}
              onChange={setRegionFilter}
              allowClear
              placeholder="区域"
              options={[
                { value: 'domestic', label: '国内' },
                { value: 'overseas', label: '海外' },
              ]}
            />
            <Select
              style={{ width: 110 }}
              value={routeModeFilter}
              onChange={setRouteModeFilter}
              allowClear
              placeholder="路由模式"
              options={[
                { value: 'auto', label: '自动' },
                { value: 'direct', label: '直连' },
                { value: 'litellm', label: 'LiteLLM' },
                { value: 'proxy', label: '代理' },
              ]}
            />
            <Select
              style={{ width: 110 }}
              value={healthFilter}
              onChange={setHealthFilter}
              allowClear
              placeholder="健康状态"
              options={[
                { value: 'healthy', label: '健康' },
                { value: 'degraded', label: '降级' },
                { value: 'unhealthy', label: '不健康' },
                { value: 'unknown', label: '未知' },
              ]}
            />
            <Button onClick={resetFilters}>重置筛选</Button>
          </Space>

          <Table
            columns={columns}
            dataSource={data}
            rowKey="id"
            loading={loading}
            scroll={{ x: 1400 }}
            pagination={{
              total,
              pageSize: 20,
              showSizeChanger: false,
              showTotal: (t) => `共 ${t} 条`,
            }}
          />
        </Space>
      </Card>

      <Modal
        title="路由配置"
        open={configModalVisible}
        onCancel={() => setConfigModalVisible(false)}
        onOk={handleSaveConfig}
        confirmLoading={configLoading}
        width={500}
      >
        <Spin spinning={configLoading}>
          <Form form={configForm} layout="vertical">
            <Form.Item name="route_mode" label="路由模式">
              <Select placeholder="选择路由模式">
                <Select.Option value="auto">自动（系统决策）</Select.Option>
                <Select.Option value="direct">直连（直接访问上游）</Select.Option>
                <Select.Option value="litellm">LiteLLM（通过LiteLLM网关）</Select.Option>
                <Select.Option value="proxy">代理（通过代理访问）</Select.Option>
              </Select>
            </Form.Item>
            <Form.Item name="endpoint_url" label="端点URL">
              <Input placeholder="自定义端点URL（可选）" />
            </Form.Item>
            <Form.Item name="fallback_endpoint_url" label="备用端点URL">
              <Input placeholder="备用端点URL（可选）" />
            </Form.Item>
          </Form>
        </Spin>
      </Modal>
    </div>
  );
};

export default AdminByokRouting;
