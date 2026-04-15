import React, { useEffect, useMemo, useState } from 'react';
import { Table, Card, Tag, Space, Button, Modal, Form, Input, message, Select } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';

interface User {
  id: number;
  email: string;
  name: string;
  role: string;
  created_at: string;
}

const AdminUsers: React.FC = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [roleFilter, setRoleFilter] = useState('');
  const [form] = Form.useForm();

  useEffect(() => {
    fetchUsers();
  }, []);

  const fetchUsers = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/v1/admin/users', {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('auth_token')}`,
        },
      });
      const result = await response.json();
      if (result.code === 0) {
        setUsers(result.data || []);
      }
    } catch (error) {
      message.error('获取用户列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateAdmin = async (values: { email: string; name: string; password: string }) => {
    try {
      const response = await fetch('/api/v1/admin/users', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${localStorage.getItem('auth_token')}`,
        },
        body: JSON.stringify({ ...values, role: 'admin' }),
      });
      const result = await response.json();
      if (result.code === 0) {
        message.success('创建管理员成功');
        setModalVisible(false);
        form.resetFields();
        fetchUsers();
      } else {
        message.error(result.message || '创建失败');
      }
    } catch (error) {
      message.error('创建管理员失败');
    }
  };

  const filteredUsers = useMemo(() => {
    const text = keyword.trim().toLowerCase();
    return users.filter((u) => {
      if (roleFilter && u.role !== roleFilter) return false;
      if (!text) return true;
      return [u.email, u.name, u.role, `${u.id}`].join(' ').toLowerCase().includes(text);
    });
  }, [users, roleFilter, keyword]);

  const columns: ColumnsType<User> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
      fixed: 'left',
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: '用户名',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      render: (role: string) => {
        const colorMap: Record<string, string> = {
          admin: 'red',
          merchant: 'blue',
          user: 'green',
        };
        const textMap: Record<string, string> = {
          admin: '管理员',
          merchant: '商户',
          user: '用户',
        };
        return <Tag color={colorMap[role] || 'default'}>{textMap[role] || role}</Tag>;
      },
    },
    {
      title: '注册时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 170,
      responsive: ['md'],
      render: (date: string) => new Date(date).toLocaleString('zh-CN'),
    },
  ];

  return (
    <div>
      <Card
        title="用户管理"
        extra={
          <Button type="primary" onClick={() => setModalVisible(true)}>
            创建管理员
          </Button>
        }
      >
        <Space wrap style={{ marginBottom: 16, width: '100%' }}>
          <Select
            style={{ width: 140 }}
            allowClear
            placeholder="角色筛选"
            value={roleFilter || undefined}
            options={[
              { value: 'admin', label: '管理员' },
              { value: 'merchant', label: '商户' },
              { value: 'user', label: '用户' },
            ]}
            onChange={(v) => setRoleFilter(v ?? '')}
          />
          <Input.Search
            style={{ width: 300, maxWidth: '100%' }}
            allowClear
            placeholder="关键词（邮箱/用户名/ID）"
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
          />
          <Button
            onClick={() => {
              setRoleFilter('');
              setKeyword('');
            }}
          >
            重置
          </Button>
        </Space>
        <div style={{ overflowX: 'auto' }}>
          <Table
            columns={columns}
            dataSource={filteredUsers}
            rowKey="id"
            loading={loading}
            pagination={{ pageSize: 10, showSizeChanger: false, showTotal: (total) => `共 ${total} 条` }}
          />
        </div>
      </Card>

      <Modal
        title="创建管理员"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={handleCreateAdmin}>
          <Form.Item
            name="email"
            label="邮箱"
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '邮箱格式不正确' },
            ]}
          >
            <Input prefix={<UserOutlined />} placeholder="admin@example.com" />
          </Form.Item>
          <Form.Item
            name="name"
            label="用户名"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input placeholder="管理员名称" />
          </Form.Item>
          <Form.Item
            name="password"
            label="密码"
            rules={[
              { required: true, message: '请输入密码' },
              { min: 6, message: '密码至少6位' },
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="设置密码" />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                创建
              </Button>
              <Button onClick={() => setModalVisible(false)}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default AdminUsers;
