import { useEffect, useState } from 'react'
import { Card, Table, Button, Tag, Space, Modal, Form, Input, InputNumber, Select, message, Popconfirm } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import { useMerchantStore } from '@/stores/merchantStore'
import { productService } from '@/services/product'
import { Product } from '@/types'
import styles from './MerchantProducts.module.css'

const MerchantProducts = () => {
  const { products, fetchProducts, isLoading } = useMerchantStore()
  const [modalVisible, setModalVisible] = useState(false)
  const [editingProduct, setEditingProduct] = useState<Product | null>(null)
  const [form] = Form.useForm()
  const [statusFilter, setStatusFilter] = useState<string>('all')

  useEffect(() => {
    fetchProducts(1, 20, statusFilter === 'all' ? undefined : statusFilter)
  }, [fetchProducts, statusFilter])

  const handleAdd = () => {
    setEditingProduct(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record: Product) => {
    setEditingProduct(record)
    form.setFieldsValue(record)
    setModalVisible(true)
  }

  const handleDelete = async (id: number) => {
    try {
      await productService.deleteProduct(id)
      message.success('商品已删除')
      fetchProducts(1, 20, statusFilter === 'all' ? undefined : statusFilter)
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      if (editingProduct) {
        await productService.updateProduct(editingProduct.id, values)
        message.success('商品已更新')
      } else {
        await productService.createProduct(values)
        message.success('商品已创建')
      }
      setModalVisible(false)
      fetchProducts(1, 20, statusFilter === 'all' ? undefined : statusFilter)
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'errorFields' in error) {
        const validationError = error as { errorFields: { errors: string[] }[] }
        const firstError = validationError.errorFields[0]?.errors[0]
        if (firstError) {
          message.error(firstError)
          return
        }
      }
      const axiosError = error as { response?: { data?: { message?: string } } }
      const errorMessage = axiosError.response?.data?.message || (editingProduct ? '更新失败' : '创建失败')
      message.error(errorMessage)
    }
  }

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 60,
    },
    {
      title: '商品名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '价格',
      dataIndex: 'price',
      key: 'price',
      width: 100,
      render: (price: number) => `¥${price.toFixed(2)}`,
    },
    {
      title: '库存',
      dataIndex: 'stock',
      key: 'stock',
      width: 80,
    },
    {
      title: '已售',
      dataIndex: 'sold_count',
      key: 'sold_count',
      width: 80,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => {
        const statusMap: Record<string, { color: string; text: string }> = {
          active: { color: 'success', text: '在售' },
          inactive: { color: 'warning', text: '下架' },
          archived: { color: 'default', text: '归档' },
        }
        const { color, text } = statusMap[status] || { color: 'default', text: status }
        return <Tag color={color}>{text}</Tag>
      },
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: unknown, record: Product) => (
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
            title="确定要删除这个商品吗？"
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
    <div className={styles.products}>
      <div className={styles.header}>
        <h2 className={styles.pageTitle}>商品管理</h2>
        <Space>
          <Select
            value={statusFilter}
            onChange={setStatusFilter}
            style={{ width: 120 }}
            options={[
              { value: 'all', label: '全部状态' },
              { value: 'active', label: '在售' },
              { value: 'inactive', label: '下架' },
              { value: 'archived', label: '归档' },
            ]}
          />
          <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
            添加商品
          </Button>
        </Space>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={products}
          rowKey="id"
          loading={isLoading}
          pagination={{
            pageSize: 20,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
          }}
        />
      </Card>

      <Modal
        title={editingProduct ? '编辑商品' : '添加商品'}
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
            label="商品名称"
            rules={[{ required: true, message: '请输入商品名称' }]}
          >
            <Input placeholder="请输入商品名称" />
          </Form.Item>
          <Form.Item name="description" label="商品描述">
            <Input.TextArea rows={3} placeholder="请输入商品描述" />
          </Form.Item>
          <Form.Item
            name="price"
            label="价格"
            rules={[
              { required: true, message: '请输入价格' },
              {
                validator: (_, value) => {
                  if (value !== undefined && value <= 0) {
                    return Promise.reject(new Error('价格必须大于0'))
                  }
                  return Promise.resolve()
                },
              },
            ]}
            validateTrigger="onBlur"
          >
            <InputNumber
              precision={2}
              style={{ width: '100%' }}
              placeholder="请输入价格"
            />
          </Form.Item>
          <Form.Item name="original_price" label="原价">
            <InputNumber
              precision={2}
              style={{ width: '100%' }}
              placeholder="请输入原价（可选）"
            />
          </Form.Item>
          <Form.Item
            name="stock"
            label="库存"
            rules={[
              { required: true, message: '请输入库存' },
              {
                validator: (_, value) => {
                  if (value !== undefined && value < 0) {
                    return Promise.reject(new Error('库存必须大于等于0'))
                  }
                  return Promise.resolve()
                },
              },
            ]}
            validateTrigger="onBlur"
          >
            <InputNumber
              style={{ width: '100%' }}
              placeholder="请输入库存"
            />
          </Form.Item>
          <Form.Item name="category" label="分类">
            <Input placeholder="请输入分类" />
          </Form.Item>
          <Form.Item
            name="status"
            label="状态"
            rules={[{ required: true, message: '请选择状态' }]}
          >
            <Select
              placeholder="请选择状态"
              options={[
                { value: 'active', label: '在售' },
                { value: 'inactive', label: '下架' },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default MerchantProducts
