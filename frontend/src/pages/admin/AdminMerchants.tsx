import { useCallback, useEffect, useState } from 'react';
import {
  Card,
  Table,
  Button,
  Tag,
  Space,
  Modal,
  message,
  Typography,
  Row,
  Col,
  Image,
  Descriptions,
  Input,
  Tooltip,
  Tabs,
  Form,
  Select,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  CheckOutlined,
  CloseOutlined,
  ReloadOutlined,
  EyeOutlined,
  EditOutlined,
} from '@ant-design/icons';
import {
  adminMerchantService,
  PendingMerchant,
  MerchantAuditLog,
  MERCHANT_BUSINESS_CATEGORY_OPTIONS,
  labelForBusinessCategory,
} from '@/services/adminMerchant';

const { Title } = Typography;
const { TextArea } = Input;

const STATUS_OPTIONS = [
  { value: '', label: '全部状态' },
  { value: 'pending', label: '待审核' },
  { value: 'reviewing', label: '审核中' },
  { value: 'active', label: '已认证' },
  { value: 'rejected', label: '已拒绝' },
  { value: 'suspended', label: '已暂停' },
];

const AUDIT_ACTION_LABELS: Record<string, string> = {
  approve: '通过',
  reject: '拒绝',
  meta_update: '信息维护',
};

const statusTag = (status: string) => {
  const colorMap: Record<string, string> = {
    pending: 'orange',
    reviewing: 'blue',
    active: 'green',
    rejected: 'red',
    suspended: 'gray',
  };
  const textMap: Record<string, string> = {
    pending: '待审核',
    reviewing: '审核中',
    active: '已认证',
    rejected: '已拒绝',
    suspended: '已暂停',
  };
  return <Tag color={colorMap[status] || 'default'}>{textMap[status] || status}</Tag>;
};

const AdminMerchants = () => {
  const [activeTab, setActiveTab] = useState('pending');

  const [merchants, setMerchants] = useState<PendingMerchant[]>([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 });

  const [allMerchants, setAllMerchants] = useState<PendingMerchant[]>([]);
  const [allLoading, setAllLoading] = useState(false);
  const [allPagination, setAllPagination] = useState({ current: 1, pageSize: 20, total: 0 });
  const [filterStatus, setFilterStatus] = useState<string>('');
  const [filterCategory, setFilterCategory] = useState<string>('');
  const [filterKeyword, setFilterKeyword] = useState('');

  const [auditLogs, setAuditLogs] = useState<MerchantAuditLog[]>([]);
  const [auditLoading, setAuditLoading] = useState(false);
  const [auditPagination, setAuditPagination] = useState({ current: 1, pageSize: 20, total: 0 });
  const [auditActionFilter, setAuditActionFilter] = useState<string>('');

  const [detailVisible, setDetailVisible] = useState(false);
  const [selectedMerchant, setSelectedMerchant] = useState<PendingMerchant | null>(null);
  const [rejectVisible, setRejectVisible] = useState(false);
  const [rejectReason, setRejectReason] = useState('');

  const [editVisible, setEditVisible] = useState(false);
  const [editSubmitting, setEditSubmitting] = useState(false);
  const [editForm] = Form.useForm();

  const fetchMerchants = async (page = 1) => {
    setLoading(true);
    try {
      const response = await adminMerchantService.getPendingMerchants(page, pagination.pageSize);
      setMerchants(response.data.data);
      setPagination({
        ...pagination,
        current: page,
        total: response.data.total,
      });
    } catch {
      message.error('获取待审核商户失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchAllMerchants = useCallback(
    async (page = 1) => {
      setAllLoading(true);
      try {
        const response = await adminMerchantService.getAdminMerchants({
          page,
          per_page: allPagination.pageSize,
          status: filterStatus || undefined,
          business_category: filterCategory || undefined,
          keyword: filterKeyword.trim() || undefined,
        });
        setAllMerchants(response.data.data);
        setAllPagination((p) => ({
          ...p,
          current: page,
          total: response.data.total,
        }));
      } catch {
        message.error('获取商户列表失败');
      } finally {
        setAllLoading(false);
      }
    },
    [allPagination.pageSize, filterStatus, filterCategory, filterKeyword]
  );

  const fetchAuditLogs = async (page = 1) => {
    setAuditLoading(true);
    try {
      const response = await adminMerchantService.getMerchantAuditLogs({
        page,
        per_page: auditPagination.pageSize,
        action: auditActionFilter || undefined,
      });
      setAuditLogs(response.data.data);
      setAuditPagination((p) => ({
        ...p,
        current: page,
        total: response.data.total,
      }));
    } catch {
      message.error('获取审核记录失败');
    } finally {
      setAuditLoading(false);
    }
  };

  useEffect(() => {
    if (activeTab === 'pending') {
      fetchMerchants(1);
    }
  }, [activeTab]);

  useEffect(() => {
    if (activeTab === 'all') {
      fetchAllMerchants(1);
    }
  }, [activeTab, fetchAllMerchants]);

  useEffect(() => {
    if (activeTab === 'audit') {
      fetchAuditLogs(1);
    }
  }, [activeTab, auditActionFilter]);

  const handleApprove = (merchant: PendingMerchant) => {
    Modal.confirm({
      title: '批准商户',
      content: `确定批准商户 "${merchant.company_name}" 的入驻申请吗？`,
      okText: '确定',
      cancelText: '取消',
      onOk: async () => {
        try {
          await adminMerchantService.approveMerchant(merchant.id);
          message.success('商户已批准');
          fetchMerchants(pagination.current);
          fetchAuditLogs(auditPagination.current);
        } catch {
          message.error('操作失败');
        }
      },
    });
  };

  const handleReject = (merchant: PendingMerchant) => {
    setSelectedMerchant(merchant);
    setRejectReason(merchant.rejection_reason || '');
    setRejectVisible(true);
  };

  const confirmReject = async () => {
    if (!selectedMerchant) return;
    try {
      await adminMerchantService.rejectMerchant(selectedMerchant.id, rejectReason);
      message.success('商户已拒绝');
      setRejectVisible(false);
      fetchMerchants(pagination.current);
      fetchAuditLogs(auditPagination.current);
    } catch {
      message.error('操作失败');
    }
  };

  const showDetail = (merchant: PendingMerchant) => {
    setSelectedMerchant(merchant);
    setDetailVisible(true);
  };

  const openEdit = (merchant: PendingMerchant) => {
    setSelectedMerchant(merchant);
    editForm.setFieldsValue({
      business_category: merchant.business_category || undefined,
      admin_notes: merchant.admin_notes || '',
    });
    setEditVisible(true);
  };

  const submitEdit = async () => {
    if (!selectedMerchant) return;
    try {
      const values = await editForm.validateFields();
      setEditSubmitting(true);
      await adminMerchantService.patchMerchant(selectedMerchant.id, {
        business_category: values.business_category,
        admin_notes: values.admin_notes,
      });
      message.success('已保存');
      setEditVisible(false);
      fetchAllMerchants(allPagination.current);
      fetchAuditLogs(auditPagination.current);
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return;
      message.error('保存失败');
    } finally {
      setEditSubmitting(false);
    }
  };

  const formatDate = (date: string) => {
    return new Date(date).toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const pendingColumns: ColumnsType<PendingMerchant> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 60, fixed: 'left' },
    {
      title: '公司名称',
      dataIndex: 'company_name',
      key: 'company_name',
      ellipsis: true,
    },
    {
      title: '经营类目',
      dataIndex: 'business_category',
      key: 'business_category',
      width: 120,
      render: (v: string) => labelForBusinessCategory(v),
    },
    {
      title: '营业执照号',
      dataIndex: 'business_license',
      key: 'business_license',
      ellipsis: true,
      responsive: ['md'],
    },
    { title: '联系人', dataIndex: 'contact_name', key: 'contact_name', responsive: ['lg'] },
    { title: '联系电话', dataIndex: 'contact_phone', key: 'contact_phone', responsive: ['lg'] },
    {
      title: '联系邮箱',
      dataIndex: 'contact_email',
      key: 'contact_email',
      ellipsis: true,
      responsive: ['xl'],
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => statusTag(status),
    },
    {
      title: '申请时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 140,
      responsive: ['md'],
      render: (date: string) => formatDate(date),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      fixed: 'right',
      render: (_: unknown, record: PendingMerchant) => (
        <Space size="small" wrap>
          <Tooltip title="查看详情">
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => showDetail(record)}
            >
              详情
            </Button>
          </Tooltip>
          {(record.status === 'pending' || record.status === 'reviewing') && (
            <>
              <Button
                type="primary"
                size="small"
                icon={<CheckOutlined />}
                onClick={() => handleApprove(record)}
              >
                批准
              </Button>
              <Button
                danger
                size="small"
                icon={<CloseOutlined />}
                onClick={() => handleReject(record)}
              >
                拒绝
              </Button>
            </>
          )}
        </Space>
      ),
    },
  ];

  const allColumns: ColumnsType<PendingMerchant> = [
    ...pendingColumns.slice(0, -1),
    {
      title: '操作',
      key: 'action',
      width: 120,
      fixed: 'right',
      render: (_: unknown, record: PendingMerchant) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => showDetail(record)}
          >
            详情
          </Button>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => openEdit(record)}>
            维护
          </Button>
        </Space>
      ),
    },
  ];

  const auditColumns: ColumnsType<MerchantAuditLog> = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 70 },
    { title: '商户ID', dataIndex: 'merchant_id', key: 'merchant_id', width: 90 },
    {
      title: '操作',
      dataIndex: 'action',
      key: 'action',
      width: 100,
      render: (a: string) => AUDIT_ACTION_LABELS[a] || a,
    },
    {
      title: '公司(快照)',
      dataIndex: 'company_name_snapshot',
      key: 'company_name_snapshot',
      ellipsis: true,
    },
    {
      title: '说明/原因',
      dataIndex: 'reason',
      key: 'reason',
      ellipsis: true,
      render: (t: string) => t || '-',
    },
    {
      title: '操作人',
      key: 'admin',
      width: 180,
      render: (_: unknown, r: MerchantAuditLog) =>
        r.admin_email || (r.admin_user_id ? `#${r.admin_user_id}` : '-'),
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (d: string) => formatDate(d),
    },
  ];

  return (
    <div>
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={24} sm={12}>
          <Title level={2} style={{ margin: 0 }}>
            商户管理
          </Title>
        </Col>
        <Col xs={24} sm={12} style={{ textAlign: 'right' }}>
          <Button
            icon={<ReloadOutlined />}
            onClick={() => {
              if (activeTab === 'pending') fetchMerchants(pagination.current);
              if (activeTab === 'all') fetchAllMerchants(allPagination.current);
              if (activeTab === 'audit') fetchAuditLogs(auditPagination.current);
            }}
          >
            刷新
          </Button>
        </Col>
      </Row>

      <Card>
        <Tabs
          activeKey={activeTab}
          onChange={(k) => setActiveTab(k)}
          items={[
            {
              key: 'pending',
              label: '待审核',
              children: (
                <Table
                  columns={pendingColumns}
                  dataSource={merchants}
                  rowKey="id"
                  loading={loading}
                  scroll={{ x: 1000 }}
                  pagination={{
                    ...pagination,
                    onChange: (page) => fetchMerchants(page),
                    showSizeChanger: false,
                    showTotal: (total) => `共 ${total} 条`,
                  }}
                />
              ),
            },
            {
              key: 'all',
              label: '全部商户',
              children: (
                <>
                  <Space wrap style={{ marginBottom: 16 }}>
                    <Select
                      style={{ width: 140 }}
                      placeholder="状态"
                      value={filterStatus || undefined}
                      allowClear
                      options={STATUS_OPTIONS.filter((o) => o.value !== '')}
                      onChange={(v) => setFilterStatus(v ?? '')}
                    />
                    <Select
                      style={{ width: 160 }}
                      placeholder="经营类目"
                      value={filterCategory || undefined}
                      allowClear
                      options={MERCHANT_BUSINESS_CATEGORY_OPTIONS}
                      onChange={(v) => setFilterCategory(v ?? '')}
                    />
                    <Input.Search
                      placeholder="公司名 / 邮箱 / 电话"
                      allowClear
                      style={{ width: 260 }}
                      onSearch={() => fetchAllMerchants(1)}
                      value={filterKeyword}
                      onChange={(e) => setFilterKeyword(e.target.value)}
                    />
                    <Button type="primary" onClick={() => fetchAllMerchants(1)}>
                      查询
                    </Button>
                  </Space>
                  <Table
                    columns={allColumns}
                    dataSource={allMerchants}
                    rowKey="id"
                    loading={allLoading}
                    scroll={{ x: 1100 }}
                    pagination={{
                      ...allPagination,
                      onChange: (page) => fetchAllMerchants(page),
                      showSizeChanger: false,
                      showTotal: (total) => `共 ${total} 条`,
                    }}
                  />
                </>
              ),
            },
            {
              key: 'audit',
              label: '审核记录',
              children: (
                <>
                  <Space style={{ marginBottom: 16 }}>
                    <span>操作类型：</span>
                    <Select
                      style={{ width: 140 }}
                      allowClear
                      placeholder="全部"
                      value={auditActionFilter || undefined}
                      options={[
                        { value: 'approve', label: '通过' },
                        { value: 'reject', label: '拒绝' },
                        { value: 'meta_update', label: '信息维护' },
                      ]}
                      onChange={(v) => setAuditActionFilter(v ?? '')}
                    />
                  </Space>
                  <Table
                    columns={auditColumns}
                    dataSource={auditLogs}
                    rowKey="id"
                    loading={auditLoading}
                    scroll={{ x: 960 }}
                    pagination={{
                      ...auditPagination,
                      onChange: (page) => fetchAuditLogs(page),
                      showSizeChanger: false,
                      showTotal: (total) => `共 ${total} 条`,
                    }}
                  />
                </>
              ),
            },
          ]}
        />
      </Card>

      <Modal
        title="商户详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={700}
      >
        {selectedMerchant && (
          <Descriptions column={{ xs: 1, sm: 2 }} bordered size="small">
            <Descriptions.Item label="公司名称" span={2}>
              {selectedMerchant.company_name}
            </Descriptions.Item>
            <Descriptions.Item label="经营类目">
              {labelForBusinessCategory(selectedMerchant.business_category)}
            </Descriptions.Item>
            <Descriptions.Item label="状态">{statusTag(selectedMerchant.status)}</Descriptions.Item>
            <Descriptions.Item label="营业执照号">
              {selectedMerchant.business_license || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="管理员备注" span={2}>
              {selectedMerchant.admin_notes || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="联系人">
              {selectedMerchant.contact_name || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="联系电话">
              {selectedMerchant.contact_phone || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="联系邮箱" span={2}>
              {selectedMerchant.contact_email || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="地址" span={2}>
              {selectedMerchant.address || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="描述" span={2}>
              {selectedMerchant.description || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="营业执照" span={2}>
              {selectedMerchant.business_license_url ? (
                <Image
                  src={selectedMerchant.business_license_url}
                  alt="营业执照"
                  width={200}
                  style={{ cursor: 'pointer' }}
                />
              ) : (
                '未上传'
              )}
            </Descriptions.Item>
            <Descriptions.Item label="身份证正面">
              {selectedMerchant.id_card_front_url ? (
                <Image
                  src={selectedMerchant.id_card_front_url}
                  alt="身份证正面"
                  width={150}
                  style={{ cursor: 'pointer' }}
                />
              ) : (
                '未上传'
              )}
            </Descriptions.Item>
            <Descriptions.Item label="身份证背面">
              {selectedMerchant.id_card_back_url ? (
                <Image
                  src={selectedMerchant.id_card_back_url}
                  alt="身份证背面"
                  width={150}
                  style={{ cursor: 'pointer' }}
                />
              ) : (
                '未上传'
              )}
            </Descriptions.Item>
            <Descriptions.Item label="申请时间">
              {formatDate(selectedMerchant.created_at)}
            </Descriptions.Item>
            <Descriptions.Item label="更新时间">
              {formatDate(selectedMerchant.updated_at)}
            </Descriptions.Item>
            {selectedMerchant.rejection_reason && (
              <Descriptions.Item label="拒绝原因" span={2}>
                <span style={{ color: 'red' }}>{selectedMerchant.rejection_reason}</span>
              </Descriptions.Item>
            )}
          </Descriptions>
        )}
      </Modal>

      <Modal
        title="拒绝商户申请"
        open={rejectVisible}
        onCancel={() => setRejectVisible(false)}
        onOk={confirmReject}
        okText="确认拒绝"
        cancelText="取消"
        okButtonProps={{ danger: true }}
      >
        <p>确定拒绝商户 &quot;{selectedMerchant?.company_name}&quot; 的入驻申请吗？</p>
        <TextArea
          placeholder="请输入拒绝原因（可选）"
          value={rejectReason}
          onChange={(e) => setRejectReason(e.target.value)}
          rows={3}
        />
      </Modal>

      <Modal
        title="维护商户信息"
        open={editVisible}
        onCancel={() => setEditVisible(false)}
        onOk={submitEdit}
        confirmLoading={editSubmitting}
        okText="保存"
        destroyOnClose
      >
        <Form form={editForm} layout="vertical">
          <Form.Item name="business_category" label="经营类目">
            <Select
              allowClear
              placeholder="选择类目"
              options={[...MERCHANT_BUSINESS_CATEGORY_OPTIONS]}
            />
          </Form.Item>
          <Form.Item name="admin_notes" label="管理员内部备注">
            <TextArea rows={4} placeholder="仅后台可见" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default AdminMerchants;
