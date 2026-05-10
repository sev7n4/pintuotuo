import { useEffect, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  Empty,
  Popconfirm,
  Space,
  Table,
  Typography,
  message,
  Tag,
} from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';
import { useTokenStore } from '@/stores/tokenStore';
import type { UserAPIKey } from '@/types';
import { PlatformAPIKeySecretCell } from '@/components/user/PlatformAPIKeySecretCell';
import { CreatePlatformAPIKeyModal } from '@/components/developer/CreatePlatformAPIKeyModal';

const { Title, Paragraph, Text } = Typography;

export default function DeveloperKeysPage() {
  const { apiKeys, fetchAPIKeys, deleteAPIKey, isLoading } = useTokenStore();
  const [modalOpen, setModalOpen] = useState(false);

  useEffect(() => {
    fetchAPIKeys();
  }, [fetchAPIKeys]);

  const handleDelete = async (id: number) => {
    const ok = await deleteAPIKey(id);
    if (ok) {
      message.success('已删除');
      fetchAPIKeys();
    }
  };

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <div>
        <Title level={3} style={{ marginTop: 0 }}>
          密钥与安全
        </Title>
        <Paragraph type="secondary">
          此处为<strong>平台调用密钥</strong>（前缀一般为 <Text code>ptd_</Text>），用于{' '}
          <Text code>Authorization: Bearer</Text> 调用拼脱脱 OpenAI 兼容接口；<strong>不是</strong>{' '}
          OpenAI 等厂商的 <Text code>sk-...</Text>。
        </Paragraph>
      </div>

      <Alert
        type="info"
        showIcon
        message="完整 Token 与充值仍在我的 Token"
        description={<Link to="/my-tokens">前往「我的 Token」查看余额、充值与流水</Link>}
      />

      <Card
        title="API 密钥"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
            创建密钥
          </Button>
        }
      >
        <Table<UserAPIKey>
          rowKey="id"
          loading={isLoading}
          pagination={false}
          dataSource={apiKeys}
          locale={{ emptyText: <Empty description="暂无密钥，请先创建" /> }}
          columns={[
            { title: '名称', dataIndex: 'name', key: 'name' },
            {
              title: '密钥',
              key: 'secret',
              render: (_: unknown, r: UserAPIKey) => <PlatformAPIKeySecretCell record={r} />,
            },
            {
              title: '状态',
              dataIndex: 'status',
              key: 'status',
              width: 100,
              render: (s: string) => (
                <Tag color={s === 'active' ? 'success' : 'default'}>
                  {s === 'active' ? '启用' : '禁用'}
                </Tag>
              ),
            },
            {
              title: '操作',
              key: 'act',
              width: 120,
              render: (_: unknown, r: UserAPIKey) => (
                <Popconfirm title="确定删除？" onConfirm={() => handleDelete(r.id)}>
                  <Button type="link" danger size="small" icon={<DeleteOutlined />}>
                    删除
                  </Button>
                </Popconfirm>
              ),
            },
          ]}
        />
      </Card>

      <CreatePlatformAPIKeyModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onSuccess={() => fetchAPIKeys()}
      />
    </Space>
  );
}
