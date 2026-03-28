import { useEffect, useState } from 'react';
import { Card, Table, Button, Tag, Space, Modal, message, Typography } from 'antd';
import { CheckOutlined, CloseOutlined, ReloadOutlined } from '@ant-design/icons';
import { adminMerchantService, PendingMerchant } from '@/services/adminMerchant';

const { Title } = Typography;

const AdminMerchants = () => {
  const [merchants, setMerchants] = useState<PendingMerchant[]>([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 });

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
    Modal.confirm({
      title: '拒绝商户',
      content: `确定拒绝商户 "${merchant.company_name}" 的入驻申请吗？`,
      okText: '确定',
      cancelText: '取消',
      onOk: async () => {
        try {
          await adminMerchantService.rejectMerchant(merchant.id);
          message.success('商户已拒绝');
          fetchMerchants(pagination.current);
        } catch (error) {
          message.error('操作失败');
        }
      },
    });
  };

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 60,
    },
    {
      title: '公司名称',
      dataIndex: 'company_name',
      key: 'company_name',
    },
    {
      title: '联系人',
      dataIndex: 'contact_name',
      key: 'contact_name',
    },
    {
      title: '联系电话',
      dataIndex: 'contact_phone',
      key: 'contact_phone',
    },
    {
      title: '联系邮箱',
      dataIndex: 'contact_email',
      key: 'contact_email',
    },
    {
      title: '地址',
      dataIndex: 'address',
      key: 'address',
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const colorMap: Record<string, string> = {
          pending: 'orange',
          active: 'green',
          rejected: 'red',
          suspended: 'gray',
        };
        const textMap: Record<string, string> = {
          pending: '待审核',
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
      render: (date: string) => new Date(date).toLocaleString('zh-CN'),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: PendingMerchant) => (
        <Space>
          {record.status === 'pending' && (
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
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 16,
        }}
      >
        <Title level={2}>商户管理</Title>
        <Button icon={<ReloadOutlined />} onClick={() => fetchMerchants(pagination.current)}>
          刷新
        </Button>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={merchants}
          rowKey="id"
          loading={loading}
          pagination={{
            ...pagination,
            onChange: (page) => fetchMerchants(page),
            showSizeChanger: false,
          }}
        />
      </Card>
    </div>
  );
};

export default AdminMerchants;
