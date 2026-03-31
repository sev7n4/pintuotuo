import { useEffect, useState } from 'react';
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
} from 'antd';
import { CheckOutlined, CloseOutlined, ReloadOutlined, EyeOutlined } from '@ant-design/icons';
import { adminMerchantService, PendingMerchant } from '@/services/adminMerchant';

const { Title } = Typography;
const { TextArea } = Input;

const AdminMerchants = () => {
  const [merchants, setMerchants] = useState<PendingMerchant[]>([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 });
  const [detailVisible, setDetailVisible] = useState(false);
  const [selectedMerchant, setSelectedMerchant] = useState<PendingMerchant | null>(null);
  const [rejectVisible, setRejectVisible] = useState(false);
  const [rejectReason, setRejectReason] = useState('');

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
    } catch (error) {
      message.error('获取商户列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMerchants();
  }, []);

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
        } catch (error) {
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
      await adminMerchantService.rejectMerchant(selectedMerchant.id);
      message.success('商户已拒绝');
      setRejectVisible(false);
      fetchMerchants(pagination.current);
    } catch (error) {
      message.error('操作失败');
    }
  };

  const showDetail = (merchant: PendingMerchant) => {
    setSelectedMerchant(merchant);
    setDetailVisible(true);
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

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 60,
      fixed: 'left' as const,
    },
    {
      title: '公司名称',
      dataIndex: 'company_name',
      key: 'company_name',
      ellipsis: true,
    },
    {
      title: '营业执照号',
      dataIndex: 'business_license',
      key: 'business_license',
      ellipsis: true,
      responsive: ['md'] as const,
    },
    {
      title: '联系人',
      dataIndex: 'contact_name',
      key: 'contact_name',
      responsive: ['lg'] as const,
    },
    {
      title: '联系电话',
      dataIndex: 'contact_phone',
      key: 'contact_phone',
      responsive: ['lg'] as const,
    },
    {
      title: '联系邮箱',
      dataIndex: 'contact_email',
      key: 'contact_email',
      ellipsis: true,
      responsive: ['xl'] as const,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => {
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
      },
    },
    {
      title: '申请时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 140,
      responsive: ['md'] as const,
      render: (date: string) => formatDate(date),
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
      fixed: 'right' as const,
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

  return (
    <div>
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={24} sm={12}>
          <Title level={2} style={{ margin: 0 }}>
            商户审核
          </Title>
        </Col>
        <Col xs={24} sm={12} style={{ textAlign: 'right' }}>
          <Button icon={<ReloadOutlined />} onClick={() => fetchMerchants(pagination.current)}>
            刷新
          </Button>
        </Col>
      </Row>

      <Card>
        <Table
          columns={columns}
          dataSource={merchants}
          rowKey="id"
          loading={loading}
          scroll={{ x: 900 }}
          pagination={{
            ...pagination,
            onChange: (page) => fetchMerchants(page),
            showSizeChanger: false,
            showTotal: (total) => `共 ${total} 条`,
          }}
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
            <Descriptions.Item label="营业执照号">
              {selectedMerchant.business_license || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag
                color={
                  selectedMerchant.status === 'pending'
                    ? 'orange'
                    : selectedMerchant.status === 'reviewing'
                      ? 'blue'
                      : 'default'
                }
              >
                {selectedMerchant.status === 'pending'
                  ? '待审核'
                  : selectedMerchant.status === 'reviewing'
                    ? '审核中'
                    : selectedMerchant.status}
              </Tag>
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
        <p>确定拒绝商户 "{selectedMerchant?.company_name}" 的入驻申请吗？</p>
        <TextArea
          placeholder="请输入拒绝原因（可选）"
          value={rejectReason}
          onChange={(e) => setRejectReason(e.target.value)}
          rows={3}
        />
      </Modal>
    </div>
  );
};

export default AdminMerchants;
