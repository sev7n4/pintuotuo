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
  Select,
  InputNumber,
  message,
  Popconfirm,
  Row,
  Col,
  Checkbox,
  Segmented,
  Alert,
} from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, TagsOutlined } from '@ant-design/icons';
import { skuService } from '@/services/sku';
import type { SPU, ModelProvider, SPUCreateRequest } from '@/types/sku';
import { MODEL_TIER_LABELS } from '@/types/sku';
import { getApiErrorMessage } from '@/utils/apiError';

interface SPUScenario {
  id: number;
  code: string;
  name: string;
  description?: string;
  is_linked: boolean;
  is_primary: boolean;
}

const AdminSPUs = () => {
  const [spus, setSPUs] = useState<SPU[]>([]);
  const [providers, setProviders] = useState<ModelProvider[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingSPU, setEditingSPU] = useState<SPU | null>(null);
  const [form] = Form.useForm();
  const [listScope, setListScope] = useState<'active' | 'all'>('active');
  const [filters, setFilters] = useState({
    provider: '',
    tier: '',
    q: '',
  });
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });
  const [scenarioModalVisible, setScenarioModalVisible] = useState(false);
  const [currentSPU, setCurrentSPU] = useState<SPU | null>(null);
  const [scenarios, setScenarios] = useState<SPUScenario[]>([]);
  const [selectedScenarios, setSelectedScenarios] = useState<number[]>([]);
  const [primaryScenario, setPrimaryScenario] = useState<number | undefined>();
  const [scenarioLoading, setScenarioLoading] = useState(false);
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<SPU | null>(null);
  const [deleteConfirmCode, setDeleteConfirmCode] = useState('');
  const [deleteLoading, setDeleteLoading] = useState(false);

  useEffect(() => {
    fetchSPUs();
    fetchProviders();
  }, [listScope, filters, pagination.current, pagination.pageSize]);

  const fetchSPUs = async () => {
    setLoading(true);
    try {
      const qq = filters.q.trim();
      const response = await skuService.getSPUs({
        page: pagination.current,
        per_page: pagination.pageSize,
        provider: filters.provider || undefined,
        tier: filters.tier || undefined,
        status: listScope === 'active' ? 'active' : 'all',
        q: qq || undefined,
      });
      setSPUs(response.data.data || []);
      setPagination((prev) => ({ ...prev, total: response.data.total }));
    } catch (e) {
      message.error(getApiErrorMessage(e, '获取SPU列表失败'));
    } finally {
      setLoading(false);
    }
  };

  const fetchProviders = async () => {
    try {
      const response = await skuService.getAllModelProviders();
      setProviders(response.data.data || []);
    } catch {
      console.error('Failed to fetch providers');
    }
  };

  const confirmDeactivateWithActiveSkus = (activeSkuCount: number): Promise<boolean> =>
    new Promise((resolve) => {
      Modal.confirm({
        title: 'SPU 将下架（禁用）',
        content: `当前仍有 ${activeSkuCount} 个在售 SKU；禁用后商户端将无法勾选这些 SKU。确定继续？`,
        okText: '仍要禁用',
        cancelText: '取消',
        onOk: () => resolve(true),
        onCancel: () => resolve(false),
      });
    });

  const handleAdd = () => {
    setEditingSPU(null);
    form.resetFields();
    const timestamp = Date.now().toString(36).toUpperCase().slice(-6);
    form.setFieldsValue({
      spu_code: `SPU-${timestamp}`,
      base_compute_points: 1.0,
      provider_input_rate: 0,
      provider_output_rate: 0,
      status: 'active',
      sort_order: 0,
    });
    setModalVisible(true);
  };

  const handleEdit = async (record: SPU) => {
    setEditingSPU(record);
    form.setFieldsValue(record);
    setModalVisible(true);
    try {
      const res = await skuService.getSPU(record.id);
      const fresh = res.data.data;
      if (fresh) {
        setEditingSPU(fresh);
        form.setFieldsValue(fresh);
      }
    } catch (e) {
      message.warning(getApiErrorMessage(e, '已用列表数据编辑，SKU 计数可能不是最新'));
    }
  };

  const runDelete = async (id: number) => {
    try {
      await skuService.deleteSPU(id);
      message.success('SPU已删除');
      setDeleteModalOpen(false);
      setDeleteTarget(null);
      setDeleteConfirmCode('');
      fetchSPUs();
    } catch (e) {
      message.error(getApiErrorMessage(e, '删除失败'));
    } finally {
      setDeleteLoading(false);
    }
  };

  const openDeleteFlow = (record: SPU) => {
    const n = record.sku_count ?? 0;
    if (n > 0) {
      setDeleteTarget(record);
      setDeleteConfirmCode('');
      setDeleteModalOpen(true);
    }
  };

  const handleDeleteModalOk = async () => {
    if (!deleteTarget) return;
    const expect = deleteTarget.spu_code.trim().toUpperCase();
    if (deleteConfirmCode.trim().toUpperCase() !== expect) {
      message.error('请输入正确的 SPU 编码以确认删除');
      return;
    }
    setDeleteLoading(true);
    await runDelete(deleteTarget.id);
  };

  const handleManageScenarios = async (record: SPU) => {
    setCurrentSPU(record);
    setScenarioLoading(true);
    setScenarioModalVisible(true);
    try {
      const response = await skuService.getSPUScenarios(record.id);
      const scenarioList = response.data.scenarios || [];
      setScenarios(scenarioList);
      const linkedIds = scenarioList.filter((s) => s.is_linked).map((s) => s.id);
      setSelectedScenarios(linkedIds);
      const primary = scenarioList.find((s) => s.is_primary);
      setPrimaryScenario(primary?.id);
    } catch (e) {
      message.error(getApiErrorMessage(e, '获取场景失败'));
    } finally {
      setScenarioLoading(false);
    }
  };

  const handleSaveScenarios = async () => {
    if (!currentSPU) return;
    setScenarioLoading(true);
    try {
      await skuService.updateSPUScenarios(currentSPU.id, {
        scenario_ids: selectedScenarios,
        primary_id: primaryScenario,
      });
      message.success('场景关联已更新');
      setScenarioModalVisible(false);
      fetchSPUs();
    } catch (e) {
      message.error(getApiErrorMessage(e, '更新失败'));
    } finally {
      setScenarioLoading(false);
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      const data: SPUCreateRequest = {
        ...values,
      };
      if (typeof values.spu_code === 'string' && values.spu_code.trim()) {
        data.spu_code = values.spu_code.toUpperCase();
      }

      if (editingSPU) {
        const wasActive = editingSPU.status === 'active';
        const willInactive = values.status === 'inactive';
        const activeSkus = editingSPU.active_sku_count ?? 0;
        if (wasActive && willInactive && activeSkus > 0) {
          const ok = await confirmDeactivateWithActiveSkus(activeSkus);
          if (!ok) return;
        }
        await skuService.updateSPU(editingSPU.id, data);
        message.success('SPU已更新');
      } else {
        await skuService.createSPU(data);
        message.success('SPU已创建');
      }
      setModalVisible(false);
      fetchSPUs();
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'errorFields' in error) {
        return;
      }
      message.error(getApiErrorMessage(error, editingSPU ? '更新失败' : '创建失败'));
    }
  };

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 60,
    },
    {
      title: 'SPU编码',
      dataIndex: 'spu_code',
      key: 'spu_code',
      width: 150,
    },
    {
      title: '产品名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '模型厂商',
      dataIndex: 'model_provider',
      key: 'model_provider',
      width: 120,
      render: (code: string) => {
        const p = providers.find((x) => x.code === code);
        const label = p ? `${p.name}${p.status !== 'active' ? '（停用）' : ''}` : code;
        return label;
      },
    },
    {
      title: '模型层级',
      dataIndex: 'model_tier',
      key: 'model_tier',
      width: 100,
      render: (tier: string) => (
        <Tag
          color={
            tier === 'pro' ? 'red' : tier === 'lite' ? 'blue' : tier === 'mini' ? 'green' : 'purple'
          }
        >
          {MODEL_TIER_LABELS[tier] || tier}
        </Tag>
      ),
    },
    {
      title: 'SKU数',
      key: 'sku_count',
      width: 72,
      render: (_: unknown, r: SPU) => r.sku_count ?? 0,
    },
    {
      title: '在售SKU',
      key: 'active_sku_count',
      width: 88,
      render: (_: unknown, r: SPU) => {
        const n = r.active_sku_count ?? 0;
        return n > 0 ? <Tag color="processing">{n}</Tag> : <span>0</span>;
      },
    },
    {
      title: '算力点系数',
      dataIndex: 'base_compute_points',
      key: 'base_compute_points',
      width: 100,
      render: (v: number) => v.toFixed(4),
    },
    {
      title: '参考成本(输入/输出)',
      key: 'provider_rates',
      width: 170,
      render: (_: unknown, r: SPU) => `${(r.provider_input_rate ?? 0).toFixed(6)} / ${(r.provider_output_rate ?? 0).toFixed(6)}`,
    },
    {
      title: '销量',
      dataIndex: 'total_sales_count',
      key: 'total_sales_count',
      width: 80,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: string) => (
        <Tag color={status === 'active' ? 'success' : 'default'}>
          {status === 'active' ? '在售' : '下架'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 220,
      render: (_: unknown, record: SPU) => {
        const skuCount = record.sku_count ?? 0;
        return (
          <Space size="small">
            <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
              编辑
            </Button>
            <Button
              type="link"
              size="small"
              icon={<TagsOutlined />}
              onClick={() => handleManageScenarios(record)}
            >
              场景
            </Button>
            {skuCount > 0 ? (
              <Button type="link" size="small" danger icon={<DeleteOutlined />} onClick={() => openDeleteFlow(record)}>
                删除
              </Button>
            ) : (
              <Popconfirm
                title="确定删除该 SPU？（当前无关联 SKU）"
                onConfirm={() => runDelete(record.id)}
                okText="确定"
                cancelText="取消"
              >
                <Button type="link" size="small" danger icon={<DeleteOutlined />}>
                  删除
                </Button>
              </Popconfirm>
            )}
          </Space>
        );
      },
    },
  ];

  return (
    <div>
      <Alert
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
        message="SPU 与商户端可见性"
        description="SPU 下架（禁用）后，其下所有 SKU 在商户端均不可勾选。默认列表仅显示在售 SPU；切换到「全部 SPU」可查看已下架项。删除有关联 SKU 的 SPU 将级联删除 SKU，请谨慎操作。"
      />

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={24} md={12}>
          <h2 style={{ margin: 0 }}>SPU 管理</h2>
        </Col>
        <Col xs={24} md={12}>
          <Segmented
            value={listScope}
            onChange={(v) => {
              setListScope(v as 'active' | 'all');
              setPagination((p) => ({ ...p, current: 1 }));
            }}
            options={[
              { label: '在售 SPU（默认）', value: 'active' },
              { label: '全部 SPU', value: 'all' },
            ]}
          />
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={24} sm={8} md={6}>
          <Select
            value={filters.provider || undefined}
            onChange={(v) => setFilters({ ...filters, provider: v ?? '' })}
            style={{ width: '100%' }}
            placeholder="厂商"
            allowClear
          >
            {providers.map((p) => (
              <Select.Option key={p.code} value={p.code}>
                {p.name}
                {p.status !== 'active' ? '（停用）' : ''}
              </Select.Option>
            ))}
          </Select>
        </Col>
        <Col xs={24} sm={8} md={6}>
          <Select
            value={filters.tier || undefined}
            onChange={(v) => setFilters({ ...filters, tier: v ?? '' })}
            style={{ width: '100%' }}
            placeholder="层级"
            allowClear
          >
            <Select.Option value="pro">旗舰版</Select.Option>
            <Select.Option value="lite">标准版</Select.Option>
            <Select.Option value="mini">轻量版</Select.Option>
            <Select.Option value="vision">多模态版</Select.Option>
          </Select>
        </Col>
        <Col xs={24} sm={8} md={8}>
          <Input.Search
            allowClear
            placeholder="关键词：SPU 编码 / 名称"
            value={filters.q}
            onChange={(e) => setFilters({ ...filters, q: e.target.value })}
            onSearch={() => setPagination((p) => ({ ...p, current: 1 }))}
          />
        </Col>
        <Col xs={24} md={4}>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd} block>
            新增SPU
          </Button>
        </Col>
      </Row>

      <Card>
        <Table
          columns={columns}
          dataSource={spus}
          rowKey="id"
          loading={loading}
          scroll={{ x: 1280 }}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (page, pageSize) => setPagination({ ...pagination, current: page, pageSize }),
          }}
        />
      </Card>

      <Modal
        title={editingSPU ? '编辑SPU' : '新增SPU'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        okText="保存"
        cancelText="取消"
        width={700}
      >
        <Form form={form} layout="vertical">
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="spu_code"
                label="SPU编码"
                rules={[{ required: true, message: '请输入SPU编码' }]}
                extra="唯一标识，如: DEEPSEEK-V3"
              >
                <Input placeholder="DEEPSEEK-V3" disabled={!!editingSPU} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="name"
                label="产品名称"
                rules={[{ required: true, message: '请输入产品名称' }]}
              >
                <Input placeholder="DeepSeek V3 模型服务" />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="model_provider"
                label="模型厂商"
                rules={[{ required: true, message: '请选择模型厂商' }]}
              >
                <Select placeholder="请选择">
                  {providers.map((p) => (
                    <Select.Option key={p.code} value={p.code}>
                      {p.name}
                      {p.status !== 'active' ? '（停用）' : ''}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="model_name"
                label="模型名称"
                rules={[{ required: true, message: '请输入模型名称' }]}
                extra="API调用时使用的模型标识"
              >
                <Input placeholder="deepseek-chat" />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={6}>
              <Form.Item name="model_version" label="模型版本">
                <Input placeholder="v3" />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item
                name="model_tier"
                label="模型层级"
                rules={[{ required: true, message: '请选择模型层级' }]}
              >
                <Select placeholder="请选择">
                  <Select.Option value="pro">旗舰版 (Pro)</Select.Option>
                  <Select.Option value="lite">标准版 (Lite)</Select.Option>
                  <Select.Option value="mini">轻量版 (Mini)</Select.Option>
                  <Select.Option value="vision">多模态版 (Vision)</Select.Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item
                name="base_compute_points"
                label="算力点消耗系数"
                rules={[{ required: true, message: '请输入算力点消耗系数' }]}
                extra="1.0 = 基准系数"
              >
                <InputNumber min={0.0001} step={0.1} precision={4} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item
                name="provider_input_rate"
                label="参考输入成本(元/1K)"
                rules={[{ required: true, message: '请输入参考输入成本' }]}
              >
                <InputNumber min={0} step={0.000001} precision={6} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={6}>
              <Form.Item
                name="provider_output_rate"
                label="参考输出成本(元/1K)"
                rules={[{ required: true, message: '请输入参考输出成本' }]}
              >
                <InputNumber min={0} step={0.000001} precision={6} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="context_window" label="上下文窗口 (K)">
                <InputNumber min={1} style={{ width: '100%' }} placeholder="64" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="max_output_tokens" label="最大输出Token">
                <InputNumber min={1} style={{ width: '100%' }} placeholder="4096" />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="description" label="产品描述">
            <Input.TextArea rows={3} placeholder="请输入产品描述" />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="status" label="状态">
                <Select>
                  <Select.Option value="active">在售</Select.Option>
                  <Select.Option value="inactive">下架</Select.Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="sort_order" label="排序权重">
                <InputNumber min={0} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>

      <Modal
        title="确认删除 SPU"
        open={deleteModalOpen}
        okText="确认删除"
        cancelText="取消"
        confirmLoading={deleteLoading}
        onCancel={() => {
          setDeleteModalOpen(false);
          setDeleteTarget(null);
          setDeleteConfirmCode('');
        }}
        onOk={handleDeleteModalOk}
      >
        {deleteTarget && (
          <>
            <p>
              该 SPU 下仍有 <strong>{deleteTarget.sku_count ?? 0}</strong> 个 SKU，删除将<strong>级联删除</strong>这些 SKU。
            </p>
            <p style={{ marginBottom: 8 }}>请输入 SPU 编码 <strong>{deleteTarget.spu_code}</strong> 以确认：</p>
            <Input
              placeholder={deleteTarget.spu_code}
              value={deleteConfirmCode}
              onChange={(e) => setDeleteConfirmCode(e.target.value)}
            />
          </>
        )}
      </Modal>

      <Modal
        title={`场景管理 - ${currentSPU?.name || ''}`}
        open={scenarioModalVisible}
        onOk={handleSaveScenarios}
        onCancel={() => setScenarioModalVisible(false)}
        okText="保存"
        cancelText="取消"
        confirmLoading={scenarioLoading}
        width={600}
      >
        <div style={{ marginBottom: 16 }}>
          <p style={{ color: '#666', marginBottom: 8 }}>
            选择该SPU适用的使用场景，可多选。设置主要场景用于前端展示优先级。
          </p>
        </div>
        {scenarioLoading ? (
          <div style={{ textAlign: 'center', padding: 20 }}>加载中...</div>
        ) : (
          <div>
            <div style={{ marginBottom: 16 }}>
              <span style={{ fontWeight: 500, marginRight: 8 }}>选择场景：</span>
            </div>
            <Row gutter={[16, 16]}>
              {scenarios.map((scenario) => (
                <Col span={12} key={scenario.id}>
                  <Card
                    size="small"
                    hoverable
                    style={{
                      borderColor: selectedScenarios.includes(scenario.id) ? '#1890ff' : undefined,
                      backgroundColor: selectedScenarios.includes(scenario.id) ? '#e6f7ff' : undefined,
                    }}
                    onClick={() => {
                      if (selectedScenarios.includes(scenario.id)) {
                        setSelectedScenarios(selectedScenarios.filter((id) => id !== scenario.id));
                        if (primaryScenario === scenario.id) {
                          setPrimaryScenario(undefined);
                        }
                      } else {
                        setSelectedScenarios([...selectedScenarios, scenario.id]);
                      }
                    }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <div>
                        <Checkbox
                          checked={selectedScenarios.includes(scenario.id)}
                          style={{ marginRight: 8 }}
                        />
                        <span style={{ fontWeight: 500 }}>{scenario.name}</span>
                        {scenario.description && (
                          <div style={{ fontSize: 12, color: '#999', marginTop: 4 }}>
                            {scenario.description}
                          </div>
                        )}
                      </div>
                      {selectedScenarios.includes(scenario.id) && (
                        <Checkbox
                          checked={primaryScenario === scenario.id}
                          onChange={(e) => {
                            e.stopPropagation();
                            setPrimaryScenario(e.target.checked ? scenario.id : undefined);
                          }}
                          onClick={(e) => e.stopPropagation()}
                        >
                          主要
                        </Checkbox>
                      )}
                    </div>
                  </Card>
                </Col>
              ))}
            </Row>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default AdminSPUs;
