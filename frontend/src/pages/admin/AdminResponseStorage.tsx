import { useEffect, useState } from 'react';
import { Card, Table, Button, Tag, Space, message, Popconfirm, Row, Col, Select } from 'antd';
import { DeleteOutlined, ClearOutlined } from '@ant-design/icons';
import { responseStorageService, ResponseStorageItem } from '@/services/responseStorage';
import { getApiErrorMessage } from '@/utils/apiError';

const statusColors: Record<string, string> = {
  completed: 'success',
  processing: 'processing',
  failed: 'error',
  expired: 'warning',
};

const statusLabels: Record<string, string> = {
  completed: '已完成',
  processing: '处理中',
  failed: '失败',
  expired: '已过期',
};

const AdminResponseStorage = () => {
  const [items, setItems] = useState<ResponseStorageItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 });
  const [statusFilter, setStatusFilter] = useState<string | undefined>(undefined);

  useEffect(() => {
    fetchItems();
  }, [pagination.current, pagination.pageSize, statusFilter]);

  const fetchItems = async () => {
    setLoading(true);
    try {
      const response = await responseStorageService.getList({
        page: pagination.current,
        per_page: pagination.pageSize,
        status: statusFilter,
      });
      setItems(response.data.data || []);
      setPagination((prev) => ({ ...prev, total: response.data.total }));
    } catch (e) {
      message.error(getApiErrorMessage(e, '获取存储响应失败'));
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await responseStorageService.delete(id);
      message.success('删除成功');
      fetchItems();
    } catch (e) {
      message.error(getApiErrorMessage(e, '删除失败'));
    }
  };

  const handleCleanExpired = async () => {
    try {
      const res = await responseStorageService.cleanExpired();
      message.success(`清理完成，共删除 ${res.data.deleted_count} 条过期响应`);
      fetchItems();
    } catch (e) {
      message.error(getApiErrorMessage(e, '清理失败'));
    }
  };

  const columns = [
    {
      title: '响应ID',
      dataIndex: 'response_id',
      key: 'response_id',
      width: 200,
      ellipsis: true,
    },
    {
      title: '用户ID',
      dataIndex: 'user_id',
      key: 'user_id',
      width: 80,
    },
    {
      title: '模型',
      dataIndex: 'model',
      key: 'model',
      width: 120,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={statusColors[status] || 'default'}>{statusLabels[status] || status}</Tag>
      ),
    },
    {
      title: '错误信息',
      dataIndex: 'error_message',
      key: 'error_message',
      width: 200,
      ellipsis: true,
      render: (msg: string) => msg || '-',
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 170,
      render: (v: string) => (v ? new Date(v).toLocaleString() : '-'),
    },
    {
      title: '过期时间',
      dataIndex: 'expires_at',
      key: 'expires_at',
      width: 170,
      render: (v: string) => (v ? new Date(v).toLocaleString() : '-'),
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_: unknown, record: ResponseStorageItem) => (
        <Popconfirm
          title="确定删除？"
          onConfirm={() => handleDelete(record.id)}
          okText="确定"
          cancelText="取消"
        >
          <Button type="link" size="small" danger icon={<DeleteOutlined />}>
            删除
          </Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <div>
      <Card>
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col span={6}>
            <Select
              placeholder="筛选状态"
              allowClear
              style={{ width: '100%' }}
              value={statusFilter}
              onChange={(v) => {
                setStatusFilter(v);
                setPagination((prev) => ({ ...prev, current: 1 }));
              }}
              options={Object.entries(statusLabels).map(([value, label]) => ({ value, label }))}
            />
          </Col>
          <Col span={18} style={{ textAlign: 'right' }}>
            <Space>
              <Button onClick={fetchItems}>刷新</Button>
              <Popconfirm
                title="确定清理所有过期响应？"
                onConfirm={handleCleanExpired}
                okText="确定"
                cancelText="取消"
              >
                <Button icon={<ClearOutlined />}>清理过期</Button>
              </Popconfirm>
            </Space>
          </Col>
        </Row>

        <Table
          columns={columns}
          dataSource={items}
          rowKey="id"
          loading={loading}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (page, pageSize) => setPagination({ ...pagination, current: page, pageSize }),
          }}
        />
      </Card>
    </div>
  );
};

export default AdminResponseStorage;
