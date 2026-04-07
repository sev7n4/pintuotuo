import { useEffect, useState } from 'react';
import {
  Card,
  Table,
  Button,
  Tag,
  Space,
  Modal,
  Form,
  Input,
  InputNumber,
  Select,
  message,
  Popconfirm,
  Progress,
  Tooltip,
  Badge,
  Descriptions,
  Divider,
  Spin,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  SyncOutlined,
  ApiOutlined,
} from '@ant-design/icons';
import { useMerchantStore } from '@/stores/merchantStore';
import { MerchantAPIKey, VerificationResult } from '@/types';
import type { ModelProvider } from '@/types/sku';
import api from '@/services/api';
import { merchantService } from '@/services/merchant';
import styles from './MerchantAPIKeys.module.css';

const MerchantAPIKeys = () => {
  const {
    apiKeys,
    apiKeyUsage,
    fetchAPIKeys,
    fetchAPIKeyUsage,
    createAPIKey,
    updateAPIKey,
    deleteAPIKey,
    isLoading,
  } = useMerchantStore();
  const [modalVisible, setModalVisible] = useState(false);
  const [editingKey, setEditingKey] = useState<MerchantAPIKey | null>(null);
  const [form] = Form.useForm();
  const [verificationModalVisible, setVerificationModalVisible] = useState(false);
  const [verificationResult, setVerificationResult] = useState<VerificationResult | null>(null);
  const [verificationLoading, setVerificationLoading] = useState(false);
  const [modelProviders, setModelProviders] = useState<ModelProvider[]>([]);
  const [providersLoading, setProvidersLoading] = useState(false);

  useEffect(() => {
    fetchAPIKeys();
    fetchAPIKeyUsage();
  }, [fetchAPIKeys, fetchAPIKeyUsage]);

  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      setProvidersLoading(true);
      try {
        const res = await merchantService.getMerchantModelProviders();
        if (!cancelled && res.data?.data) {
          setModelProviders(res.data.data);
        }
      } catch {
        if (!cancelled) {
          message.error('加载提供商列表失败');
          setModelProviders([]);
        }
      } finally {
        if (!cancelled) setProvidersLoading(false);
      }
    };
    load();
    return () => {
      cancelled = true;
    };
  }, []);

  const handleAdd = () => {
    setEditingKey(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (record: MerchantAPIKey) => {
    setEditingKey(record);
    form.setFieldsValue({
      name: record.name,
      quota_limit: record.quota_limit,
      status: record.status,
      endpoint_url: record.endpoint_url,
      health_check_level: record.health_check_level,
      cost_input_rate: record.cost_input_rate,
      cost_output_rate: record.cost_output_rate,
      profit_margin: record.profit_margin,
    });
    setModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    const success = await deleteAPIKey(id);
    if (success) {
      message.success('API密钥已删除');
      fetchAPIKeys();
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (editingKey) {
        const success = await updateAPIKey(editingKey.id, values);
        if (success) {
          message.success('API密钥已更新');
          setModalVisible(false);
          fetchAPIKeys();
        }
      } else {
        const success = await createAPIKey(values);
        if (success) {
          message.success('API密钥已创建');
          setModalVisible(false);
          fetchAPIKeys();
        }
      }
    } catch (error) {
      message.error('操作失败');
    }
  };

  const handleVerify = async (id: number) => {
    setVerificationModalVisible(true);
    setVerificationLoading(true);
    setVerificationResult(null);

    try {
      await api.post(`/merchants/api-keys/${id}/verify`);
      message.success('验证已启动，正在后台执行...');

      await pollVerificationResult(id);
    } catch (error) {
      message.error('启动验证失败');
      setVerificationLoading(false);
    }
  };

  const pollVerificationResult = async (id: number) => {
    const maxAttempts = 30;
    const interval = 2000;
    let attempts = 0;

    const poll = async (): Promise<void> => {
      try {
        const response = await api.get<VerificationResult>(
          `/merchants/api-keys/${id}/verification`
        );
        const result = response.data;

        setVerificationResult(result);

        if (result.status === 'pending' || result.status === 'in_progress') {
          attempts++;
          if (attempts < maxAttempts) {
            await new Promise((resolve) => setTimeout(resolve, interval));
            await poll();
          } else {
            message.warning('验证超时，请稍后查看结果');
            setVerificationLoading(false);
          }
        } else {
          setVerificationLoading(false);
          fetchAPIKeys();
        }
      } catch (error) {
        attempts++;
        if (attempts < maxAttempts) {
          await new Promise((resolve) => setTimeout(resolve, interval));
          await poll();
        } else {
          message.error('获取验证结果失败');
          setVerificationLoading(false);
        }
      }
    };

    await poll();
  };

  const getHealthStatusTag = (status?: string) => {
    if (!status) return <Tag>未知</Tag>;

    const statusConfig: Record<string, { color: string; icon: React.ReactNode; text: string }> = {
      healthy: { color: 'success', icon: <CheckCircleOutlined />, text: '健康' },
      degraded: { color: 'warning', icon: <ExclamationCircleOutlined />, text: '降级' },
      unhealthy: { color: 'error', icon: <ExclamationCircleOutlined />, text: '不健康' },
      unknown: { color: 'default', icon: null, text: '未知' },
    };

    const config = statusConfig[status] || statusConfig.unknown;
    return (
      <Tag color={config.color} icon={config.icon}>
        {config.text}
      </Tag>
    );
  };

  const getVerificationStatusBadge = (result?: string) => {
    if (!result) return <Badge status="default" text="未验证" />;

    const statusMap: Record<string, 'success' | 'error' | 'processing' | 'default'> = {
      success: 'success',
      failed: 'error',
      pending: 'processing',
    };

    return (
      <Badge
        status={statusMap[result] || 'default'}
        text={result === 'success' ? '已验证' : result === 'failed' ? '验证失败' : '待验证'}
      />
    );
  };

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '提供商',
      dataIndex: 'provider',
      key: 'provider',
      render: (provider: string) => <Tag color="blue">{provider.toUpperCase()}</Tag>,
    },
    {
      title: '端点URL',
      dataIndex: 'endpoint_url',
      key: 'endpoint_url',
      render: (url?: string) => (
        <Tooltip title={url || '使用默认端点'}>
          <span>{url ? url.substring(0, 30) + '...' : '默认'}</span>
        </Tooltip>
      ),
    },
    {
      title: '配额',
      dataIndex: 'quota_limit',
      key: 'quota_limit',
      render: (_: unknown, record: MerchantAPIKey) => {
        const usage = apiKeyUsage.find((u) => u.id === record.id);
        if (!usage || usage.quota_limit === 0) {
          return '无限制';
        }
        const percent = Math.min(usage.usage_percentage, 100);
        return (
          <div className={styles.quotaCell}>
            <Progress percent={percent} size="small" />
            <span className={styles.quotaText}>
              ${usage.quota_used.toFixed(2)} / ${usage.quota_limit.toFixed(2)}
            </span>
          </div>
        );
      },
    },
    {
      title: '健康状态',
      dataIndex: 'health_status',
      key: 'health_status',
      render: (status: string, record: MerchantAPIKey) => (
        <Space direction="vertical" size="small">
          {getHealthStatusTag(status)}
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
      render: (result: string, record: MerchantAPIKey) => (
        <Space direction="vertical" size="small">
          {getVerificationStatusBadge(result)}
          {record.verified_at && (
            <span style={{ fontSize: '12px', color: '#999' }}>
              {new Date(record.verified_at).toLocaleString('zh-CN')}
            </span>
          )}
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'active' ? 'success' : 'default'}>
          {status === 'active' ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 220,
      render: (_: unknown, record: MerchantAPIKey) => (
        <Space size="small">
          <Tooltip title="验证API Key">
            <Button
              type="link"
              size="small"
              icon={<ApiOutlined />}
              onClick={() => handleVerify(record.id)}
            >
              验证
            </Button>
          </Tooltip>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这个API密钥吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div className={styles.apiKeys}>
      <div className={styles.header}>
        <h2 className={styles.pageTitle}>API密钥管理</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
          添加密钥
        </Button>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={apiKeys}
          rowKey="id"
          loading={isLoading}
          pagination={false}
          scroll={{ x: 'max-content' }}
        />
      </Card>

      <Modal
        title={editingKey ? '编辑API密钥' : '添加API密钥'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        okText="保存"
        cancelText="取消"
        width={700}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="name"
            label="密钥名称"
            rules={[{ required: true, message: '请输入密钥名称' }]}
          >
            <Input placeholder="例如：生产环境密钥" disabled={!!editingKey} />
          </Form.Item>

          {!editingKey && (
            <>
              <Form.Item
                name="provider"
                label="提供商"
                rules={[{ required: true, message: '请选择提供商' }]}
              >
                <Spin spinning={providersLoading}>
                  <Select
                    placeholder="请选择提供商"
                    allowClear
                    showSearch
                    optionFilterProp="children"
                    notFoundContent={providersLoading ? '加载中…' : '暂无可用提供商'}
                  >
                    {modelProviders.map((p) => (
                      <Select.Option key={p.id} value={p.code}>
                        {p.name}
                      </Select.Option>
                    ))}
                  </Select>
                </Spin>
              </Form.Item>
              <Form.Item
                name="api_key"
                label="API Key"
                rules={[{ required: true, message: '请输入API Key' }]}
              >
                <Input.Password placeholder="请输入API Key" />
              </Form.Item>
              <Form.Item name="api_secret" label="API Secret">
                <Input.Password placeholder="请输入API Secret（可选）" />
              </Form.Item>
            </>
          )}

          <Form.Item name="endpoint_url" label="端点URL">
            <Input placeholder="自定义端点URL（可选，留空使用默认）" />
          </Form.Item>

          <Form.Item name="quota_limit" label="配额限制（美元）">
            <InputNumber
              min={0}
              precision={2}
              style={{ width: '100%' }}
              placeholder="0表示无限制"
            />
          </Form.Item>

          <Form.Item name="health_check_level" label="健康检查级别">
            <Select placeholder="选择健康检查频率">
              <Select.Option value="high">高频（每5分钟）</Select.Option>
              <Select.Option value="medium">中频（每15分钟）</Select.Option>
              <Select.Option value="low">低频（每30分钟）</Select.Option>
              <Select.Option value="daily">每日一次</Select.Option>
            </Select>
          </Form.Item>

          <Divider>成本定价配置</Divider>

          <Form.Item name="cost_input_rate" label="输入成本率（$/1K tokens）">
            <InputNumber
              min={0}
              precision={6}
              style={{ width: '100%' }}
              placeholder="输入token成本"
            />
          </Form.Item>

          <Form.Item name="cost_output_rate" label="输出成本率（$/1K tokens）">
            <InputNumber
              min={0}
              precision={6}
              style={{ width: '100%' }}
              placeholder="输出token成本"
            />
          </Form.Item>

          <Form.Item name="profit_margin" label="利润率（%）">
            <InputNumber
              min={0}
              max={100}
              precision={2}
              style={{ width: '100%' }}
              placeholder="利润率百分比"
            />
          </Form.Item>

          {editingKey && (
            <Form.Item name="status" label="状态">
              <Select>
                <Select.Option value="active">启用</Select.Option>
                <Select.Option value="inactive">禁用</Select.Option>
              </Select>
            </Form.Item>
          )}
        </Form>
      </Modal>

      <Modal
        title="API Key 验证"
        open={verificationModalVisible}
        onCancel={() => {
          setVerificationModalVisible(false);
          setVerificationResult(null);
        }}
        footer={null}
        width={600}
      >
        {verificationLoading && (
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <Spin size="large" />
            <p style={{ marginTop: 16 }}>正在验证 API Key...</p>
          </div>
        )}

        {verificationResult && (
          <Descriptions bordered column={1}>
            <Descriptions.Item label="验证状态">
              {verificationResult.status === 'success' ? (
                <Tag color="success" icon={<CheckCircleOutlined />}>
                  验证成功
                </Tag>
              ) : verificationResult.status === 'failed' ? (
                <Tag color="error" icon={<ExclamationCircleOutlined />}>
                  验证失败
                </Tag>
              ) : (
                <Tag color="processing" icon={<SyncOutlined spin />}>
                  验证中
                </Tag>
              )}
            </Descriptions.Item>

            <Descriptions.Item label="连接测试">
              {verificationResult.connection_test ? (
                <Tag color="success">成功 ({verificationResult.connection_latency_ms}ms)</Tag>
              ) : (
                <Tag color="error">失败</Tag>
              )}
            </Descriptions.Item>

            {verificationResult.models_found && verificationResult.models_found.length > 0 && (
              <Descriptions.Item label="支持的模型">
                <Space wrap>
                  {verificationResult.models_found.map((model) => (
                    <Tag key={model}>{model}</Tag>
                  ))}
                </Space>
              </Descriptions.Item>
            )}

            <Descriptions.Item label="定价验证">
              {verificationResult.pricing_verified ? (
                <Tag color="success">已验证</Tag>
              ) : (
                <Tag color="warning">未验证</Tag>
              )}
            </Descriptions.Item>

            {verificationResult.error_message && (
              <Descriptions.Item label="错误信息">
                <span style={{ color: '#ff4d4f' }}>{verificationResult.error_message}</span>
              </Descriptions.Item>
            )}

            <Descriptions.Item label="重试次数">{verificationResult.retry_count}</Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  );
};

export default MerchantAPIKeys;
