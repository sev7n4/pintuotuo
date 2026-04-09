import { useEffect, useState, useMemo } from 'react';
import {
  Card,
  Table,
  Button,
  Tag,
  Space,
  Modal,
  Form,
  Input,
  Select,
  InputNumber,
  message,
  Popconfirm,
  Row,
  Col,
  Switch,
  Divider,
  Grid,
  Segmented,
  Checkbox,
  Alert,
} from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { skuService } from '@/services/sku';
import type { SKUWithSPU, SPU, SKUCreateRequest, SKUUpdateRequest } from '@/types/sku';
import { SKU_TYPE_LABELS, MODEL_TIER_LABELS, SUBSCRIPTION_PERIOD_LABELS } from '@/types/sku';
import { getApiErrorMessage } from '@/utils/apiError';

/** 管理端 InputNumber 可能产生小数，PostgreSQL 整型列会拒绝（生产曾出现 stock=9.9 导致更新失败） */
function intFormValue(v: unknown, fallback: number): number {
  if (v === null || v === undefined) return fallback;
  const n = Number(v);
  if (!Number.isFinite(n)) return fallback;
  return Math.trunc(n);
}

const { useBreakpoint } = Grid;

const AdminSKUs = () => {
  const screens = useBreakpoint();
  const [skus, setSKUs] = useState<SKUWithSPU[]>([]);
  const [spus, setSPUs] = useState<SPU[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingSKU, setEditingSKU] = useState<SKUWithSPU | null>(null);
  const [form] = Form.useForm();
  const [listScope, setListScope] = useState<'sellable' | 'all'>('sellable');
  const [filters, setFilters] = useState({
    spu_id: '',
    type: '',
    sku_status: 'all' as 'all' | 'active' | 'inactive',
    spu_status: '' as '' | 'active' | 'inactive',
    provider: '',
    q: '',
    misaligned: false,
  });
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });
  const [selectedSKUType, setSelectedSKUType] = useState<string>('token_pack');

  const providerOptions = useMemo(
    () => [...new Set(spus.map((s) => s.model_provider))].filter(Boolean).sort(),
    [spus]
  );

  useEffect(() => {
    fetchSKUs();
  }, [listScope, filters, pagination.current, pagination.pageSize]);

  useEffect(() => {
    fetchSPUs();
  }, []);

  const fetchSKUs = async () => {
    setLoading(true);
    try {
      const params: Parameters<typeof skuService.getSKUs>[0] = {
        page: pagination.current,
        per_page: pagination.pageSize,
        scope: listScope,
      };
      if (filters.spu_id) params.spu_id = parseInt(filters.spu_id, 10);
      if (filters.type) params.type = filters.type;
      if (listScope === 'all') {
        params.status = filters.sku_status === 'all' ? 'all' : filters.sku_status;
        if (filters.spu_status) params.spu_status = filters.spu_status;
        if (filters.provider) params.provider = filters.provider;
        const qq = filters.q.trim();
        if (qq) params.q = qq;
        if (filters.misaligned) params.misaligned = '1';
      } else {
        if (filters.provider) params.provider = filters.provider;
        const qq = filters.q.trim();
        if (qq) params.q = qq;
      }
      const response = await skuService.getSKUs(params);
      setSKUs(response.data.data || []);
      setPagination((prev) => ({ ...prev, total: response.data.total }));
    } catch (e) {
      message.error(getApiErrorMessage(e, '获取SKU列表失败'));
    } finally {
      setLoading(false);
    }
  };

  const fetchSPUs = async () => {
    try {
      const response = await skuService.getSPUs({ status: 'all', per_page: 200 });
      setSPUs(response.data.data || []);
    } catch {
      console.error('Failed to fetch SPUs');
    }
  };

  const handleAdd = () => {
    setEditingSKU(null);
    form.resetFields();
    form.setFieldsValue({
      sku_type: 'token_pack',
      valid_days: 365,
      stock: -1,
      status: 'active',
      group_enabled: true,
      min_group_size: 2,
      max_group_size: 10,
      is_trial: false,
      is_unlimited: false,
      is_promoted: false,
      inherit_spu_cost: true,
    });
    setSelectedSKUType('token_pack');
    setModalVisible(true);
  };

  const handleEdit = (record: SKUWithSPU) => {
    setEditingSKU(record);
    form.setFieldsValue(record);
    setSelectedSKUType(record.sku_type);
    setModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await skuService.deleteSKU(id);
      message.success('SKU已删除');
      fetchSKUs();
    } catch (e) {
      message.error(getApiErrorMessage(e, '删除失败'));
    }
  };

  const confirmInactiveSPUWithActiveSKU = (): Promise<boolean> =>
    new Promise((resolve) => {
      Modal.confirm({
        title: 'SPU 已下架',
        content:
          '商户端仍无法勾选该 SKU（需 SPU 在售）。若仍保存为「在售」，仅作后台记录，请确认是否继续。',
        okText: '仍要保存',
        cancelText: '取消',
        onOk: () => resolve(true),
        onCancel: () => resolve(false),
      });
    });

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      const spu = spus.find((s) => s.id === values.spu_id);
      if (spu && spu.status !== 'active' && values.status === 'active') {
        const ok = await confirmInactiveSPUWithActiveSKU();
        if (!ok) return;
      }

      if (editingSKU) {
        const updateData: SKUUpdateRequest = {
          retail_price: values.retail_price,
          wholesale_price: values.wholesale_price,
          original_price: values.original_price,
          stock: intFormValue(values.stock, editingSKU.stock ?? -1),
          daily_limit: intFormValue(values.daily_limit, editingSKU.daily_limit ?? 0),
          group_enabled: values.group_enabled,
          min_group_size: intFormValue(values.min_group_size, editingSKU.min_group_size ?? 2),
          max_group_size: intFormValue(values.max_group_size, editingSKU.max_group_size ?? 10),
          group_discount_rate: values.group_discount_rate,
          status: values.status,
          is_promoted: values.is_promoted,
          cost_input_rate: values.cost_input_rate,
          cost_output_rate: values.cost_output_rate,
          inherit_spu_cost: values.inherit_spu_cost,
        };
        await skuService.updateSKU(editingSKU.id, updateData);
        message.success('SKU已更新');
      } else {
        const { sku_code: codeRaw, ...rest } = values;
        const createData: SKUCreateRequest = {
          ...rest,
          sku_code: codeRaw.toUpperCase(),
          token_amount: rest.token_amount != null ? intFormValue(rest.token_amount, 0) : undefined,
          fair_use_limit:
            rest.fair_use_limit != null ? intFormValue(rest.fair_use_limit, 0) : undefined,
          valid_days: rest.valid_days != null ? intFormValue(rest.valid_days, 365) : undefined,
          concurrent_requests:
            rest.concurrent_requests != null
              ? intFormValue(rest.concurrent_requests, 0)
              : undefined,
          tpm_limit: rest.tpm_limit != null ? intFormValue(rest.tpm_limit, 0) : undefined,
          rpm_limit: rest.rpm_limit != null ? intFormValue(rest.rpm_limit, 0) : undefined,
          stock: rest.stock != null ? intFormValue(rest.stock, -1) : undefined,
          daily_limit: rest.daily_limit != null ? intFormValue(rest.daily_limit, 0) : undefined,
          min_group_size:
            rest.min_group_size != null ? intFormValue(rest.min_group_size, 2) : undefined,
          max_group_size:
            rest.max_group_size != null ? intFormValue(rest.max_group_size, 10) : undefined,
          trial_duration_days:
            rest.trial_duration_days != null
              ? intFormValue(rest.trial_duration_days, 0)
              : undefined,
        };
        await skuService.createSKU(createData);
        message.success('SKU已创建');
      }
      setModalVisible(false);
      fetchSKUs();
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'errorFields' in error) {
        return;
      }
      const fallback = editingSKU ? '更新失败' : '创建失败';
      message.error(getApiErrorMessage(error, fallback));
    }
  };

  const columns = [
    {
      title: 'SKU编码',
      dataIndex: 'sku_code',
      key: 'sku_code',
      width: 150,
    },
    {
      title: '关联SPU',
      dataIndex: 'spu_name',
      key: 'spu_name',
      width: 150,
      render: (name: string, record: SKUWithSPU) => (
        <Space direction="vertical" size={0}>
          <span>{name}</span>
          <Tag>{MODEL_TIER_LABELS[record.model_tier || 'lite']}</Tag>
        </Space>
      ),
    },
    {
      title: '类型',
      dataIndex: 'sku_type',
      key: 'sku_type',
      width: 100,
      render: (type: string) => <Tag color="blue">{SKU_TYPE_LABELS[type]}</Tag>,
    },
    {
      title: 'SPU状态',
      key: 'spu_status',
      width: 96,
      render: (_: unknown, record: SKUWithSPU) => (
        <Tag color={record.spu_status === 'active' ? 'processing' : 'default'}>
          {record.spu_status === 'active' ? 'SPU在售' : 'SPU下架'}
        </Tag>
      ),
    },
    {
      title: '商户可选',
      key: 'sellable',
      width: 92,
      render: (_: unknown, record: SKUWithSPU) =>
        record.sellable === true ? <Tag color="success">是</Tag> : <Tag color="warning">否</Tag>,
    },
    {
      title: '规格',
      key: 'spec',
      width: 150,
      render: (_: unknown, record: SKUWithSPU) => {
        if (record.sku_type === 'token_pack') {
          return (
            <span>
              {record.token_amount?.toLocaleString()} Tokens
              <br />({record.compute_points?.toLocaleString()} 算力点)
            </span>
          );
        }
        if (record.sku_type === 'subscription') {
          return (
            <span>
              {SUBSCRIPTION_PERIOD_LABELS[record.subscription_period || 'monthly']}
              {record.is_unlimited && ' (无限量)'}
            </span>
          );
        }
        if (record.sku_type === 'concurrent') {
          return <span>{record.concurrent_requests} 并发</span>;
        }
        return '-';
      },
    },
    {
      title: '价格',
      dataIndex: 'retail_price',
      key: 'retail_price',
      width: 100,
      render: (price: number, record: SKUWithSPU) => (
        <Space direction="vertical" size={0}>
          <span style={{ color: '#f5222d', fontWeight: 'bold' }}>¥{price.toFixed(2)}</span>
          {record.original_price && record.original_price > price && (
            <span style={{ textDecoration: 'line-through', color: '#999' }}>
              ¥{record.original_price.toFixed(2)}
            </span>
          )}
        </Space>
      ),
    },
    {
      title: '库存',
      dataIndex: 'stock',
      key: 'stock',
      width: 80,
      render: (stock: number) => (stock === -1 ? '无限' : stock),
    },
    {
      title: '拼团',
      key: 'group',
      width: 100,
      render: (_: unknown, record: SKUWithSPU) =>
        record.group_enabled ? (
          <Tag color="green">
            {record.min_group_size}-{record.max_group_size}人
          </Tag>
        ) : (
          <Tag>不支持</Tag>
        ),
    },
    {
      title: '单位成本(元/1K)',
      key: 'unit_cost',
      width: 200,
      render: (_: unknown, record: SKUWithSPU) => (
        <Space direction="vertical" size={0}>
          <Tag color={record.inherit_spu_cost ? 'blue' : 'gold'}>
            {record.inherit_spu_cost ? '继承SPU' : 'SKU覆盖'}
          </Tag>
          <span>
            in {Number(record.cost_input_rate || 0).toFixed(6)} / out{' '}
            {Number(record.cost_output_rate || 0).toFixed(6)}
          </span>
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: string) => (
        <Tag color={status === 'active' ? 'success' : 'warning'}>
          {status === 'active' ? '在售' : '下架'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: unknown, record: SKUWithSPU) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这个SKU吗？"
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
    <div>
      <Card>
        <Alert
          type="info"
          showIcon
          message="说明：商户上架可勾选的 SKU =「SKU 在售」且「所属 SPU 在售」。默认列表与商户端一致；切换到「全部 SKU」可查看下架、配置异常等。"
          style={{ marginBottom: 16 }}
        />
        <Segmented
          value={listScope}
          onChange={(v) => {
            setListScope(v as 'sellable' | 'all');
            setPagination((p) => ({ ...p, current: 1 }));
          }}
          options={[
            { label: '商户可选（默认）', value: 'sellable' },
            { label: '全部 SKU', value: 'all' },
          ]}
          style={{ marginBottom: 16 }}
        />
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col xs={24} sm={12} md={6}>
            <Select
              value={filters.spu_id || undefined}
              onChange={(v) => setFilters({ ...filters, spu_id: v ?? '' })}
              style={{ width: '100%' }}
              placeholder="筛选 SPU"
              allowClear
              showSearch
              optionFilterProp="label"
            >
              {spus.map((s) => (
                <Select.Option key={s.id} value={String(s.id)} label={s.name}>
                  {s.name}
                </Select.Option>
              ))}
            </Select>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Select
              value={filters.type || undefined}
              onChange={(v) => setFilters({ ...filters, type: v ?? '' })}
              style={{ width: '100%' }}
              placeholder="SKU 类型"
              allowClear
            >
              <Select.Option value="token_pack">Token包</Select.Option>
              <Select.Option value="subscription">订阅套餐</Select.Option>
              <Select.Option value="concurrent">并发套餐</Select.Option>
              <Select.Option value="trial">试用套餐</Select.Option>
            </Select>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Select
              value={filters.provider || undefined}
              onChange={(v) => setFilters({ ...filters, provider: v ?? '' })}
              style={{ width: '100%' }}
              placeholder="厂商 model_provider"
              allowClear
              showSearch
            >
              {providerOptions.map((p) => (
                <Select.Option key={p} value={p}>
                  {p}
                </Select.Option>
              ))}
            </Select>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Input
              allowClear
              placeholder="编码 / SPU名 / spu_code"
              value={filters.q}
              onChange={(e) => setFilters({ ...filters, q: e.target.value })}
            />
          </Col>
        </Row>
        {listScope === 'all' && (
          <Row gutter={[16, 16]} style={{ marginBottom: 16 }} align="middle">
            <Col xs={24} sm={8} md={5}>
              <Select
                value={filters.sku_status}
                onChange={(v) => setFilters({ ...filters, sku_status: v })}
                style={{ width: '100%' }}
                placeholder="SKU 状态"
              >
                <Select.Option value="all">全部状态</Select.Option>
                <Select.Option value="active">SKU 在售</Select.Option>
                <Select.Option value="inactive">SKU 下架</Select.Option>
              </Select>
            </Col>
            <Col xs={24} sm={8} md={5}>
              <Select
                value={filters.spu_status || undefined}
                onChange={(v) =>
                  setFilters({ ...filters, spu_status: (v ?? '') as '' | 'active' | 'inactive' })
                }
                style={{ width: '100%' }}
                placeholder="SPU 状态"
                allowClear
              >
                <Select.Option value="active">SPU 在售</Select.Option>
                <Select.Option value="inactive">SPU 下架</Select.Option>
              </Select>
            </Col>
            <Col xs={24} sm={12} md={10}>
              <Checkbox
                checked={filters.misaligned}
                onChange={(e) => setFilters({ ...filters, misaligned: e.target.checked })}
              >
                仅「SKU在售但SPU下架」
              </Checkbox>
            </Col>
          </Row>
        )}
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col xs={24} style={{ textAlign: screens?.md ? 'right' : 'left' }}>
            <Space wrap>
              <Button onClick={() => fetchSKUs()}>刷新</Button>
              <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
                新增SKU
              </Button>
            </Space>
          </Col>
        </Row>

        <Table
          columns={columns}
          dataSource={skus}
          rowKey="id"
          loading={loading}
          scroll={{ x: 1560 }}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (page, pageSize) => setPagination({ ...pagination, current: page, pageSize }),
          }}
        />
      </Card>

      <Modal
        title={editingSKU ? '编辑SKU' : '新增SKU'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        okText="保存"
        cancelText="取消"
        width={800}
      >
        <Form form={form} layout="vertical">
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="spu_id"
                label="关联SPU"
                rules={[{ required: true, message: '请选择SPU' }]}
              >
                <Select
                  placeholder="请选择SPU"
                  showSearch
                  optionFilterProp="label"
                  disabled={!!editingSKU}
                >
                  {spus.map((s) => (
                    <Select.Option key={s.id} value={s.id} label={s.name}>
                      {s.name} ({s.spu_code}){s.status !== 'active' ? ' · SPU下架' : ''}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={24}>
              <Form.Item noStyle shouldUpdate={(prev, cur) => prev.spu_id !== cur.spu_id}>
                {() => {
                  const sid = form.getFieldValue('spu_id') as number | undefined;
                  const spu = spus.find((x) => x.id === sid);
                  if (!spu || spu.status === 'active') return null;
                  return (
                    <Alert
                      type="warning"
                      showIcon
                      style={{ marginBottom: 16 }}
                      message="该 SPU 已下架：商户端无法勾选此 SKU；请先上架 SPU 或仍保存为后台记录。"
                    />
                  );
                }}
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="sku_code"
                label="SKU编码"
                rules={[{ required: true, message: '请输入SKU编码' }]}
                extra="唯一标识，如: DEEPSEEK-V3-100K"
              >
                <Input placeholder="DEEPSEEK-V3-100K" disabled={!!editingSKU} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="sku_type"
                label="SKU类型"
                rules={[{ required: true, message: '请选择SKU类型' }]}
              >
                <Select
                  placeholder="请选择"
                  disabled={!!editingSKU}
                  onChange={(v) => {
                    setSelectedSKUType(v);
                    form.setFieldsValue({
                      is_unlimited: false,
                      group_enabled: v !== 'trial',
                    });
                  }}
                >
                  <Select.Option value="token_pack">Token包</Select.Option>
                  <Select.Option value="subscription">订阅套餐</Select.Option>
                  <Select.Option value="concurrent">并发套餐</Select.Option>
                  <Select.Option value="trial">试用套餐</Select.Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="status" label="状态">
                <Select>
                  <Select.Option value="active">在售</Select.Option>
                  <Select.Option value="inactive">下架</Select.Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          {selectedSKUType === 'token_pack' && (
            <>
              <Divider>Token包配置</Divider>
              <Row gutter={16}>
                <Col span={8}>
                  <Form.Item
                    name="token_amount"
                    label="Token数量"
                    rules={[{ required: true, message: '请输入Token数量' }]}
                  >
                    <InputNumber
                      min={1}
                      precision={0}
                      step={1}
                      style={{ width: '100%' }}
                      placeholder="100000"
                      formatter={(value) => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
                    />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    name="compute_points"
                    label="算力点数量"
                    rules={[{ required: true, message: '请输入算力点数量' }]}
                  >
                    <InputNumber
                      min={1}
                      style={{ width: '100%' }}
                      placeholder="100"
                      formatter={(value) => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
                    />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name="valid_days" label="有效天数">
                    <InputNumber min={1} precision={0} step={1} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>
            </>
          )}

          {selectedSKUType === 'subscription' && (
            <>
              <Divider>订阅配置</Divider>
              <Row gutter={16}>
                <Col span={8}>
                  <Form.Item
                    name="subscription_period"
                    label="订阅周期"
                    rules={[{ required: true, message: '请选择订阅周期' }]}
                  >
                    <Select placeholder="请选择">
                      <Select.Option value="monthly">月度</Select.Option>
                      <Select.Option value="quarterly">季度</Select.Option>
                      <Select.Option value="yearly">年度</Select.Option>
                    </Select>
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name="is_unlimited" label="是否无限量" valuePropName="checked">
                    <Switch />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    name="fair_use_limit"
                    label="公平使用上限"
                    extra="无限量套餐的每日使用上限"
                  >
                    <InputNumber
                      min={1}
                      precision={0}
                      step={1}
                      style={{ width: '100%' }}
                      placeholder="Tokens/天"
                    />
                  </Form.Item>
                </Col>
              </Row>
            </>
          )}

          {selectedSKUType === 'concurrent' && (
            <>
              <Divider>并发配置</Divider>
              <Row gutter={16}>
                <Col span={8}>
                  <Form.Item
                    name="concurrent_requests"
                    label="并发请求数"
                    rules={[{ required: true, message: '请输入并发请求数' }]}
                  >
                    <InputNumber min={1} precision={0} step={1} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name="tpm_limit" label="TPM限制">
                    <InputNumber
                      min={1}
                      precision={0}
                      step={1}
                      style={{ width: '100%' }}
                      placeholder="Tokens/分钟"
                    />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name="rpm_limit" label="RPM限制">
                    <InputNumber
                      min={1}
                      precision={0}
                      step={1}
                      style={{ width: '100%' }}
                      placeholder="请求/分钟"
                    />
                  </Form.Item>
                </Col>
              </Row>
            </>
          )}

          <Divider>定价与库存</Divider>
          <Row gutter={16}>
            <Col span={8}>
              <Form.Item
                name="retail_price"
                label="零售价"
                rules={[{ required: true, message: '请输入零售价' }]}
              >
                <InputNumber
                  min={0}
                  precision={2}
                  style={{ width: '100%' }}
                  placeholder="9.90"
                  prefix="¥"
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="original_price" label="原价">
                <InputNumber
                  min={0}
                  precision={2}
                  style={{ width: '100%' }}
                  placeholder="19.90"
                  prefix="¥"
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="stock" label="库存" extra="-1表示无限库存">
                <InputNumber
                  min={-1}
                  precision={0}
                  step={1}
                  style={{ width: '100%' }}
                  placeholder="-1"
                />
              </Form.Item>
            </Col>
          </Row>

          <Divider orientation="left">高级：单位成本（元/1K tokens）</Divider>
          <Form.Item name="inherit_spu_cost" label="继承 SPU 参考价" valuePropName="checked">
            <Switch checkedChildren="继承" unCheckedChildren="手填" />
          </Form.Item>
          <Form.Item
            noStyle
            shouldUpdate={(prev, cur) => prev.inherit_spu_cost !== cur.inherit_spu_cost}
          >
            {({ getFieldValue }) => {
              const inherit = getFieldValue('inherit_spu_cost') !== false;
              const sid = getFieldValue('spu_id') as number | undefined;
              const spu = spus.find((x) => x.id === sid);
              return (
                <>
                  {inherit ? (
                    <Alert
                      type="info"
                      showIcon
                      style={{ marginBottom: 12 }}
                      message={`当前将继承 SPU 参考价：输入 ${Number(spu?.provider_input_rate || 0).toFixed(6)} / 输出 ${Number(spu?.provider_output_rate || 0).toFixed(6)}（元/1K）`}
                    />
                  ) : (
                    <Alert
                      type="warning"
                      showIcon
                      style={{ marginBottom: 12 }}
                      message="已关闭继承：请填写非空单位成本，系统会用于商户上架默认值与路由成本基线。"
                    />
                  )}
                  {!inherit && (
                    <>
                      <Row gutter={16}>
                        <Col span={12}>
                          <Form.Item
                            name="cost_input_rate"
                            label="输入成本（元/1K）"
                            rules={[{ required: true, message: '请填写输入成本' }]}
                          >
                            <InputNumber
                              min={0.000001}
                              step={0.000001}
                              precision={6}
                              style={{ width: '100%' }}
                            />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item
                            name="cost_output_rate"
                            label="输出成本（元/1K）"
                            rules={[{ required: true, message: '请填写输出成本' }]}
                          >
                            <InputNumber
                              min={0.000001}
                              step={0.000001}
                              precision={6}
                              style={{ width: '100%' }}
                            />
                          </Form.Item>
                        </Col>
                      </Row>
                      <Form.Item noStyle shouldUpdate>
                        {({ getFieldValue }) => {
                          const inCost = Number(getFieldValue('cost_input_rate') || 0);
                          const outCost = Number(getFieldValue('cost_output_rate') || 0);
                          const spuIn = Number(spu?.provider_input_rate || 0);
                          const spuOut = Number(spu?.provider_output_rate || 0);
                          const inDelta = spuIn > 0 ? Math.abs(inCost - spuIn) / spuIn : 0;
                          const outDelta = spuOut > 0 ? Math.abs(outCost - spuOut) / spuOut : 0;
                          if (inDelta <= 0.2 && outDelta <= 0.2) return null;
                          return (
                            <Alert
                              type="warning"
                              showIcon
                              style={{ marginBottom: 12 }}
                              message="手填成本与 SPU 参考价偏差超过 20%"
                              description={`SPU参考：输入 ${spuIn.toFixed(6)} / 输出 ${spuOut.toFixed(6)}；当前：输入 ${inCost.toFixed(6)} / 输出 ${outCost.toFixed(6)}（元/1K）`}
                            />
                          );
                        }}
                      </Form.Item>
                    </>
                  )}
                </>
              );
            }}
          </Form.Item>

          {selectedSKUType !== 'trial' && (
            <>
              <Divider>拼团配置</Divider>
              <Row gutter={16}>
                <Col span={6}>
                  <Form.Item name="group_enabled" label="支持拼团" valuePropName="checked">
                    <Switch />
                  </Form.Item>
                </Col>
                <Col span={6}>
                  <Form.Item name="min_group_size" label="最小人数">
                    <InputNumber min={2} precision={0} step={1} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={6}>
                  <Form.Item name="max_group_size" label="最大人数">
                    <InputNumber min={2} precision={0} step={1} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={6}>
                  <Form.Item
                    name="group_discount_rate"
                    label="拼团折扣率(%)"
                    extra="如: 20 表示8折"
                  >
                    <InputNumber min={0} max={100} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>
            </>
          )}

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="is_promoted" label="推荐商品" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="daily_limit" label="每日购买限制">
                <InputNumber
                  min={0}
                  precision={0}
                  step={1}
                  style={{ width: '100%' }}
                  placeholder="0表示不限制"
                />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </div>
  );
};

export default AdminSKUs;
