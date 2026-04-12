import { useEffect, useState, useCallback } from 'react';
import {
  Card,
  Table,
  Button,
  Tag,
  Space,
  Modal,
  Select,
  message,
  Popconfirm,
  Empty,
  Typography,
  Tooltip,
  Row,
  Col,
  Grid,
  InputNumber,
  Segmented,
  Alert,
} from 'antd';
import { PlusOutlined, DeleteOutlined, ApiOutlined, ShopOutlined } from '@ant-design/icons';
import { merchantSkuService } from '@/services/merchantSku';
import type {
  MerchantSKUDetail,
  AvailableSKU,
  MerchantSKUCreateRequest,
} from '@/types/merchantSku';
import type { MerchantAPIKey } from '@/types';
import { merchantService } from '@/services/merchant';
import MerchantGuard from '@/components/MerchantGuard';
import styles from './MerchantProducts.module.css';

const { Text } = Typography;
const { useBreakpoint } = Grid;

const MerchantSKUs = () => {
  const screens = useBreakpoint();
  const [merchantSKUs, setMerchantSKUs] = useState<MerchantSKUDetail[]>([]);
  const [availableSKUs, setAvailableSKUs] = useState<AvailableSKU[]>([]);
  const [apiKeys, setApiKeys] = useState<MerchantAPIKey[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectModalVisible, setSelectModalVisible] = useState(false);
  const [statusFilter, setStatusFilter] = useState<string>('active');
  const [providerFilter, setProviderFilter] = useState<string>('');
  const [typeFilter, setTypeFilter] = useState<string>('');
  const [selectedSKUs, setSelectedSKUs] = useState<number[]>([]);
  const [selectedAPIKey, setSelectedAPIKey] = useState<number | undefined>();
  const [costGuideMode, setCostGuideMode] = useState<'official' | 'self_hosted'>('official');
  const [customPricingEnabled, setCustomPricingEnabled] = useState(false);
  const [customInputRate, setCustomInputRate] = useState<number | undefined>();
  const [customOutputRate, setCustomOutputRate] = useState<number | undefined>();
  const [customProfitMargin, setCustomProfitMargin] = useState<number>(20);
  const [submitting, setSubmitting] = useState(false);
  const [providerOptions, setProviderOptions] = useState<string[]>([]);
  const [typeOptions, setTypeOptions] = useState<string[]>([]);

  const selectedOfficialRef = selectedSKUs.reduce(
    (acc, id) => {
      const hit = availableSKUs.find((s) => s.id === id);
      if (!hit) return acc;
      return {
        input: acc.input + Number(hit.spu_input_rate || 0),
        output: acc.output + Number(hit.spu_output_rate || 0),
        count: acc.count + 1,
      };
    },
    { input: 0, output: 0, count: 0 }
  );
  const avgOfficialInput =
    selectedOfficialRef.count > 0 ? selectedOfficialRef.input / selectedOfficialRef.count : 0;
  const avgOfficialOutput =
    selectedOfficialRef.count > 0 ? selectedOfficialRef.output / selectedOfficialRef.count : 0;

  const fetchMerchantSKUs = useCallback(async () => {
    setLoading(true);
    try {
      const data = await merchantSkuService.getMerchantSKUs(statusFilter);
      setMerchantSKUs(data);
    } catch {
      message.error('获取SKU列表失败');
    } finally {
      setLoading(false);
    }
  }, [statusFilter]);

  const fetchAvailableSKUs = useCallback(async () => {
    try {
      const data = await merchantSkuService.getAvailableSKUs(providerFilter, typeFilter);
      setAvailableSKUs(data);
    } catch {
      message.error('获取可用SKU列表失败');
    }
  }, [providerFilter, typeFilter]);

  const fetchAPIKeys = useCallback(async () => {
    try {
      const response = await merchantService.getAPIKeys();
      setApiKeys(response.data.data || []);
    } catch {
      console.error('获取API Key列表失败');
    }
  }, []);

  useEffect(() => {
    fetchAPIKeys();
  }, [fetchAPIKeys]);

  useEffect(() => {
    fetchMerchantSKUs();
  }, [fetchMerchantSKUs]);

  useEffect(() => {
    if (selectModalVisible) {
      fetchAvailableSKUs();
      // 使用“无筛选”数据构造筛选项，避免筛选项被当前过滤结果反向限制
      merchantSkuService
        .getAvailableSKUs('', '')
        .then((all) => {
          setProviderOptions([...new Set(all.map((sku) => sku.model_provider))]);
          setTypeOptions([...new Set(all.map((sku) => sku.sku_type))]);
        })
        .catch(() => {
          setProviderOptions([]);
          setTypeOptions([]);
        });
    }
  }, [selectModalVisible, providerFilter, typeFilter, fetchAvailableSKUs]);

  const handleSelectSKU = () => {
    setSelectedSKUs([]);
    setSelectedAPIKey(undefined);
    setCostGuideMode('official');
    setCustomPricingEnabled(false);
    setCustomInputRate(undefined);
    setCustomOutputRate(undefined);
    setCustomProfitMargin(20);
    setSelectModalVisible(true);
  };

  const handleBatchSelect = async () => {
    if (selectedSKUs.length === 0) {
      message.warning('请选择要上架的SKU');
      return;
    }

    setSubmitting(true);
    try {
      if (customPricingEnabled && (customInputRate == null || customOutputRate == null)) {
        message.warning('自定义成本模式下，请填写输入/输出成本');
        setSubmitting(false);
        return;
      }
      for (const skuId of selectedSKUs) {
        const data: MerchantSKUCreateRequest = {
          sku_id: skuId,
          api_key_id: selectedAPIKey,
          custom_pricing_enabled: customPricingEnabled,
          cost_input_rate: customPricingEnabled ? customInputRate : undefined,
          cost_output_rate: customPricingEnabled ? customOutputRate : undefined,
          profit_margin: customPricingEnabled ? customProfitMargin : undefined,
        };
        await merchantSkuService.createMerchantSKU(data);
      }
      message.success(`已成功上架 ${selectedSKUs.length} 个SKU`);
      setSelectModalVisible(false);
      fetchMerchantSKUs();
    } catch (error: unknown) {
      const axiosError = error as { response?: { data?: { message?: string; code?: string } } };
      if (axiosError.response?.data?.code === 'MERCHANT_NOT_FOUND') {
        message.warning('您的商户申请正在审核中');
        Modal.confirm({
          title: '提交商户资料',
          content: '您需要提交商户资料才能上架商品，是否现在提交？',
          okText: '去提交',
          cancelText: '取消',
          onOk: () => {
            window.location.href = '/merchant/settings';
          },
        });
      } else if (axiosError.response?.data?.code === 'MERCHANT_SKU_KEY_CONFLICT') {
        message.error(
          axiosError.response?.data?.message ||
            '该 API Key 已绑定其他在售商品，请更换 Key 或先下架占用该 Key 的商品'
        );
      } else {
        message.error(axiosError.response?.data?.message || '上架失败');
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleToggleStatus = async (id: number, currentStatus: string) => {
    try {
      const newStatus = currentStatus === 'active' ? 'inactive' : 'active';
      await merchantSkuService.updateMerchantSKU(id, { status: newStatus });
      message.success(newStatus === 'active' ? 'SKU已上架' : 'SKU已下架');
      fetchMerchantSKUs();
    } catch (error: unknown) {
      const axiosError = error as { response?: { data?: { message?: string; code?: string } } };
      if (axiosError.response?.data?.code === 'MERCHANT_SKU_KEY_CONFLICT') {
        message.error(
          axiosError.response?.data?.message ||
            '该 API Key 已绑定其他在售商品，请更换 Key 或先下架占用该 Key 的商品'
        );
      } else {
        message.error('操作失败');
      }
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await merchantSkuService.deleteMerchantSKU(id);
      message.success('SKU已下架');
      fetchMerchantSKUs();
    } catch {
      message.error('下架失败');
    }
  };

  const handleUpdateAPIKey = async (id: number, apiKeyId: number | undefined) => {
    try {
      await merchantSkuService.updateMerchantSKU(id, { api_key_id: apiKeyId });
      message.success('API Key已更新');
      fetchMerchantSKUs();
    } catch (error: unknown) {
      const axiosError = error as { response?: { data?: { message?: string; code?: string } } };
      if (axiosError.response?.data?.code === 'MERCHANT_SKU_KEY_CONFLICT') {
        message.error(
          axiosError.response?.data?.message ||
            '该 API Key 已绑定其他在售商品，请更换 Key 或先下架占用该 Key 的商品'
        );
      } else {
        message.error('更新失败');
      }
    }
  };

  const formatPrice = (price: number) => `¥${price.toFixed(2)}`;

  const formatTokenAmount = (amount?: number) => {
    if (!amount) return '-';
    if (amount >= 1000000) return `${(amount / 1000000).toFixed(0)}M`;
    if (amount >= 1000) return `${(amount / 1000).toFixed(0)}K`;
    return amount.toString();
  };

  const getSKUTypeLabel = (type: string) => {
    const typeMap: Record<string, string> = {
      token_pack: 'Token包',
      subscription: '订阅',
      concurrent: '并发',
    };
    return typeMap[type] || type;
  };

  const getProviderLabel = (provider: string) => {
    const providerMap: Record<string, string> = {
      deepseek: 'DeepSeek',
      openai: 'OpenAI',
      anthropic: 'Anthropic',
      baidu: '百度',
      alibaba: '阿里',
      zhipu: '智谱',
    };
    return providerMap[provider] || provider;
  };

  const columns = [
    {
      title: 'SKU编码',
      dataIndex: 'sku_code',
      key: 'sku_code',
      width: screens.xs ? 100 : 150,
    },
    {
      title: 'SPU名称',
      dataIndex: 'spu_name',
      key: 'spu_name',
      width: screens.xs ? 100 : 150,
      render: (name: string, record: MerchantSKUDetail) => (
        <Tooltip title={`${record.model_name} (${record.model_tier})`}>
          <span>{name}</span>
        </Tooltip>
      ),
    },
    {
      title: '厂商',
      dataIndex: 'model_provider',
      key: 'model_provider',
      width: screens.xs ? 80 : 100,
      render: (provider: string) => getProviderLabel(provider),
    },
    {
      title: '类型',
      dataIndex: 'sku_type',
      key: 'sku_type',
      width: 80,
      render: (type: string) => <Tag>{getSKUTypeLabel(type)}</Tag>,
    },
    {
      title: '规格',
      key: 'spec',
      width: 80,
      render: (_: unknown, record: MerchantSKUDetail) => {
        if (record.sku_type === 'token_pack') {
          return formatTokenAmount(record.token_amount);
        }
        return `${record.valid_days}天`;
      },
    },
    {
      title: '价格',
      dataIndex: 'retail_price',
      key: 'retail_price',
      width: 80,
      render: (price: number) => <Text strong>{formatPrice(price)}</Text>,
    },
    {
      title: '官方参考成本(元/1K)',
      key: 'spu_ref_cost',
      width: 180,
      render: (_: unknown, record: MerchantSKUDetail) => (
        <span>
          in {Number(record.spu_input_rate || 0).toFixed(6)} / out{' '}
          {Number(record.spu_output_rate || 0).toFixed(6)}
        </span>
      ),
    },
    {
      title: '关联API Key',
      dataIndex: 'api_key_name',
      key: 'api_key_name',
      width: screens.xs ? 100 : 150,
      render: (_: unknown, record: MerchantSKUDetail) => (
        <Select
          style={{ width: '100%' }}
          placeholder="选择API Key"
          value={record.api_key_id}
          onChange={(value) => handleUpdateAPIKey(record.id, value)}
          allowClear
          size="small"
        >
          {apiKeys
            .filter((key) => key.provider === record.model_provider)
            .map((key) => (
              <Select.Option key={key.id} value={key.id}>
                {key.name}
              </Select.Option>
            ))}
        </Select>
      ),
    },
    {
      title: '销量',
      dataIndex: 'sales_count',
      key: 'sales_count',
      width: 60,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 70,
      render: (status: string) => (
        <Tag color={status === 'active' ? 'success' : 'warning'}>
          {status === 'active' ? '在售' : '下架'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: screens.xs ? 100 : 150,
      render: (_: unknown, record: MerchantSKUDetail) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            onClick={() => handleToggleStatus(record.id, record.status)}
          >
            {record.status === 'active' ? '下架' : '上架'}
          </Button>
          <Popconfirm
            title="确定要下架这个SKU吗？"
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

  const availableColumns = [
    {
      title: 'SKU编码',
      dataIndex: 'sku_code',
      key: 'sku_code',
      width: 120,
    },
    {
      title: 'SPU名称',
      dataIndex: 'spu_name',
      key: 'spu_name',
      width: 120,
    },
    {
      title: '厂商',
      dataIndex: 'model_provider',
      key: 'model_provider',
      width: 100,
      render: (provider: string) => getProviderLabel(provider),
    },
    {
      title: '类型',
      dataIndex: 'sku_type',
      key: 'sku_type',
      width: 80,
      render: (type: string) => <Tag>{getSKUTypeLabel(type)}</Tag>,
    },
    {
      title: '规格',
      key: 'spec',
      width: 80,
      render: (_: unknown, record: AvailableSKU) => {
        if (record.sku_type === 'token_pack') {
          return formatTokenAmount(record.token_amount);
        }
        return `${record.valid_days}天`;
      },
    },
    {
      title: '价格',
      dataIndex: 'retail_price',
      key: 'retail_price',
      width: 80,
      render: (price: number) => <Text strong>{formatPrice(price)}</Text>,
    },
    {
      title: '状态',
      dataIndex: 'is_selected',
      key: 'is_selected',
      width: 80,
      render: (isSelected: boolean) =>
        isSelected ? <Tag color="success">已选择</Tag> : <Tag>未选择</Tag>,
    },
  ];

  const rowSelection = {
    selectedRowKeys: selectedSKUs,
    onChange: (keys: React.Key[]) => {
      setSelectedSKUs(keys as number[]);
    },
    getCheckboxProps: (record: AvailableSKU) => ({
      disabled: record.is_selected,
    }),
  };

  const providers = providerOptions;
  const types = typeOptions;

  return (
    <MerchantGuard>
      <div className={styles.container}>
        <Card
          title={
            <Space>
              <ShopOutlined />
              <span>选品与SKU管理</span>
            </Space>
          }
          extra={
            merchantSKUs.length > 0 ? (
              <Space wrap>
                <Select
                  style={{ width: 100 }}
                  value={statusFilter}
                  onChange={setStatusFilter}
                  options={[
                    { value: 'all', label: '全部' },
                    { value: 'active', label: '在售' },
                    { value: 'inactive', label: '已下架' },
                  ]}
                />
                <Button type="primary" icon={<PlusOutlined />} onClick={handleSelectSKU}>
                  选择商品上架
                </Button>
              </Space>
            ) : null
          }
        >
          <Table
            columns={columns}
            dataSource={merchantSKUs}
            rowKey="id"
            loading={loading}
            scroll={{ x: 900 }}
            pagination={{
              pageSize: 10,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 条`,
            }}
            size={screens.xs ? 'small' : 'middle'}
            locale={{
              emptyText: (
                <Empty description="暂无SKU" image={Empty.PRESENTED_IMAGE_SIMPLE}>
                  <Button type="primary" icon={<PlusOutlined />} onClick={handleSelectSKU}>
                    选择商品上架
                  </Button>
                </Empty>
              ),
            }}
          />
        </Card>

        <Modal
          title="选择要上架的商品"
          open={selectModalVisible}
          onCancel={() => setSelectModalVisible(false)}
          width={screens.xs ? '95%' : 900}
          footer={[
            <Button key="cancel" onClick={() => setSelectModalVisible(false)}>
              取消
            </Button>,
            <Button
              key="submit"
              type="primary"
              loading={submitting}
              onClick={handleBatchSelect}
              disabled={selectedSKUs.length === 0}
            >
              确认上架 ({selectedSKUs.length})
            </Button>,
          ]}
        >
          <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
            <Col xs={24} sm={12}>
              <Select
                style={{ width: '100%' }}
                placeholder="筛选厂商"
                value={providerFilter || undefined}
                onChange={(value) => setProviderFilter(value || '')}
                allowClear
                options={providers.map((p) => ({ value: p, label: getProviderLabel(p) }))}
              />
            </Col>
            <Col xs={24} sm={12}>
              <Select
                style={{ width: '100%' }}
                placeholder="筛选类型"
                value={typeFilter || undefined}
                onChange={(value) => setTypeFilter(value || '')}
                allowClear
                options={types.map((t) => ({ value: t, label: getSKUTypeLabel(t) }))}
              />
            </Col>
          </Row>

          <Table
            columns={availableColumns}
            dataSource={availableSKUs}
            rowKey="id"
            rowSelection={rowSelection}
            pagination={{ pageSize: 5 }}
            size="small"
            scroll={{ x: 700 }}
          />

          <div style={{ marginTop: 16 }}>
            <Alert
              type={costGuideMode === 'official' ? 'info' : 'warning'}
              showIcon
              style={{ marginBottom: 12 }}
              message={costGuideMode === 'official' ? '官方目录默认成本向导' : '自部署自定义成本向导'}
              description={
                costGuideMode === 'official'
                  ? '默认继承平台目录（SPU）参考成本并同步到上架 SKU；适用于官方代理/托管场景。'
                  : `若为自部署同名模型，建议填写真实推理成本。当前所选SKU官方参考均值：输入 ${avgOfficialInput.toFixed(6)} / 输出 ${avgOfficialOutput.toFixed(6)} 元/1K（仅供对照）。`
              }
            />
            <Segmented
              value={costGuideMode}
              onChange={(v) => {
                const mode = v as 'official' | 'self_hosted';
                setCostGuideMode(mode);
                setCustomPricingEnabled(mode === 'self_hosted');
              }}
              options={[
                { label: '官方目录默认', value: 'official' },
                { label: '自部署自定义', value: 'self_hosted' },
              ]}
              style={{ marginBottom: 12 }}
            />

            <label style={{ display: 'block', marginBottom: 8 }}>选择API Key（可选）</label>
            <Select
              style={{ width: '100%' }}
              placeholder="选择要关联的API Key"
              value={selectedAPIKey}
              onChange={setSelectedAPIKey}
              allowClear
            >
              {apiKeys.map((key) => (
                <Select.Option key={key.id} value={key.id}>
                  <Space>
                    <ApiOutlined />
                    {key.name} ({getProviderLabel(key.provider)})
                  </Space>
                </Select.Option>
              ))}
            </Select>

            {customPricingEnabled && (
              <Row gutter={12} style={{ marginTop: 12 }}>
                <Col xs={24} md={8}>
                  <label style={{ display: 'block', marginBottom: 6 }}>输入成本(元/1K)</label>
                  <InputNumber
                    min={0}
                    step={0.000001}
                    precision={6}
                    value={customInputRate}
                    onChange={(v) => setCustomInputRate(v ?? undefined)}
                    style={{ width: '100%' }}
                  />
                </Col>
                <Col xs={24} md={8}>
                  <label style={{ display: 'block', marginBottom: 6 }}>输出成本(元/1K)</label>
                  <InputNumber
                    min={0}
                    step={0.000001}
                    precision={6}
                    value={customOutputRate}
                    onChange={(v) => setCustomOutputRate(v ?? undefined)}
                    style={{ width: '100%' }}
                  />
                </Col>
                <Col xs={24} md={8}>
                  <label style={{ display: 'block', marginBottom: 6 }}>利润率(%)</label>
                  <InputNumber
                    min={0}
                    max={200}
                    step={0.1}
                    precision={2}
                    value={customProfitMargin}
                    onChange={(v) => setCustomProfitMargin(v ?? 20)}
                    style={{ width: '100%' }}
                  />
                </Col>
              </Row>
            )}
          </div>
        </Modal>
      </div>
    </MerchantGuard>
  );
};

export default MerchantSKUs;
