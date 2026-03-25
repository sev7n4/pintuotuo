import { useEffect, useState } from 'react'
import { Card, Table, Button, Tag, Space, Modal, Form, Input, InputNumber, Select, message, Popconfirm, Progress } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import { useMerchantStore } from '@/stores/merchantStore'
import { MerchantAPIKey } from '@/types'
import styles from './MerchantAPIKeys.module.css'

const MerchantAPIKeys = () => {
  const { apiKeys, apiKeyUsage, fetchAPIKeys, fetchAPIKeyUsage, createAPIKey, updateAPIKey, deleteAPIKey, isLoading } = useMerchantStore()
  const [modalVisible, setModalVisible] = useState(false)
  const [editingKey, setEditingKey] = useState<MerchantAPIKey | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchAPIKeys()
    fetchAPIKeyUsage()
  }, [fetchAPIKeys, fetchAPIKeyUsage])

  const handleAdd = () => {
    setEditingKey(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record: MerchantAPIKey) => {
    setEditingKey(record)
    form.setFieldsValue({
      name: record.name,
      quota_limit: record.quota_limit,
      status: record.status,
    })
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    const success = await deleteAPIKey(id)
    if (success) {
      message.success('API密钥已删除')
      fetchAPIKeys()
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      if (editingKey) {
        const success = await updateAPIKey(editingKey.id, values)
        if (success) {
          message.success('API密钥已更新')
          setModalVisible(false)
          fetchAPIKeys()
        }
      } else {
        const success = await createAPIKey(values)
        if (success) {
          message.success('API密钥已创建')
          setModalVisible(false)
          fetchAPIKeys()
        }
      }
    } catch (error) {
      message.error('操作失败')
    }
  }

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
      render: (provider: string) => (
        <Tag color="blue">{provider.toUpperCase()}</Tag>
      ),
    },
    {
      title: '配额',
      dataIndex: 'quota_limit',
      key: 'quota_limit',
      render: (_: unknown, record: MerchantAPIKey) => {
        const usage = apiKeyUsage.find(u => u.id === record.id)
        if (!usage || usage.quota_limit === 0) {
          return '无限制'
        }
        const percent = Math.min(usage.usage_percentage, 100)
        return (
          <div className={styles.quotaCell}>
            <Progress percent={percent} size="small" />
            <span className={styles.quotaText}>
              ${usage.quota_used.toFixed(2)} / ${usage.quota_limit.toFixed(2)}
            </span>
          </div>
        )
      },
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
      title: '最后使用',
      dataIndex: 'last_used_at',
      key: 'last_used_at',
      render: (date: string) => date ? new Date(date).toLocaleString('zh-CN') : '从未使用',
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleDateString('zh-CN'),
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: unknown, record: MerchantAPIKey) => (
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
  ]

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
        width={600}
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
                <Select placeholder="请选择提供商">
                  <Select.Option value="openai">OpenAI</Select.Option>
                  <Select.Option value="anthropic">Anthropic</Select.Option>
                  <Select.Option value="google">Google AI</Select.Option>
                  <Select.Option value="azure">Azure OpenAI</Select.Option>
                </Select>
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
          <Form.Item name="quota_limit" label="配额限制（美元）">
            <InputNumber
              min={0}
              precision={2}
              style={{ width: '100%' }}
              placeholder="0表示无限制"
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
    </div>
  )
}

export default MerchantAPIKeys
