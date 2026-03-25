import { useEffect, useState } from 'react'
import { Card, Table, Tag, Select, Space, Button, message } from 'antd'
import { useMerchantStore } from '@/stores/merchantStore'
import styles from './MerchantOrders.module.css'

const MerchantOrders = () => {
  const { orders, fetchOrders, isLoading } = useMerchantStore()
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [page, setPage] = useState(1)

  useEffect(() => {
    fetchOrders(page, 20, statusFilter === 'all' ? undefined : statusFilter)
  }, [fetchOrders, page, statusFilter])

  const handleExport = () => {
    message.info('导出功能开发中')
  }

  const statusMap: Record<string, { color: string; text: string }> = {
    pending: { color: 'default', text: '待支付' },
    paid: { color: 'processing', text: '已支付' },
    completed: { color: 'success', text: '已完成' },
    failed: { color: 'error', text: '失败' },
    cancelled: { color: 'warning', text: '已取消' },
  }

  const columns = [
    {
      title: '订单ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '商品名称',
      dataIndex: 'product_name',
      key: 'product_name',
    },
    {
      title: '用户ID',
      dataIndex: 'user_id',
      key: 'user_id',
      width: 80,
    },
    {
      title: '数量',
      dataIndex: 'quantity',
      key: 'quantity',
      width: 80,
    },
    {
      title: '订单金额',
      dataIndex: 'total_price',
      key: 'total_price',
      width: 120,
      render: (price: number) => `¥${price.toFixed(2)}`,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => {
        const { color, text } = statusMap[status] || { color: 'default', text: status }
        return <Tag color={color}>{text}</Tag>
      },
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (date: string) => new Date(date).toLocaleString('zh-CN'),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 160,
      render: (date: string) => new Date(date).toLocaleString('zh-CN'),
    },
  ]

  return (
    <div className={styles.orders}>
      <div className={styles.header}>
        <h2 className={styles.pageTitle}>订单管理</h2>
        <Space>
          <Select
            value={statusFilter}
            onChange={(value) => {
              setStatusFilter(value)
              setPage(1)
            }}
            style={{ width: 120 }}
            options={[
              { value: 'all', label: '全部状态' },
              { value: 'pending', label: '待支付' },
              { value: 'paid', label: '已支付' },
              { value: 'completed', label: '已完成' },
              { value: 'cancelled', label: '已取消' },
            ]}
          />
          <Button onClick={handleExport}>导出数据</Button>
        </Space>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={orders}
          rowKey="id"
          loading={isLoading}
          scroll={{ x: 900 }}
          pagination={{
            current: page,
            pageSize: 20,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (newPage) => setPage(newPage),
          }}
        />
      </Card>
    </div>
  )
}

export default MerchantOrders
