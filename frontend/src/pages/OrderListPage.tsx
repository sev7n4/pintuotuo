import React, { useEffect } from 'react'
import {
  Table,
  Button,
  Space,
  Tag,
  Empty,
  Spin,
  Modal,
  Descriptions,
} from 'antd'
import { useNavigate } from 'react-router-dom'
import { useOrderStore } from '@stores/orderStore'
import type { Order } from '@types/index'

const statusMap: Record<string, { color: string; label: string }> = {
  pending: { color: 'orange', label: '待支付' },
  paid: { color: 'blue', label: '已支付' },
  completed: { color: 'green', label: '已完成' },
  failed: { color: 'red', label: '失败' },
  cancelled: { color: 'gray', label: '已取消' },
}

export const OrderListPage: React.FC = () => {
  const navigate = useNavigate()
  const { orders, isLoading, error, fetchOrders } = useOrderStore()
  const [selectedOrder, setSelectedOrder] = React.useState<Order | null>(null)
  const [modalVisible, setModalVisible] = React.useState(false)

  useEffect(() => {
    fetchOrders()
  }, [])

  const columns = [
    {
      title: '订单号',
      dataIndex: 'id',
      key: 'id',
      render: (id: number) => `#${id}`,
    },
    {
      title: '产品ID',
      dataIndex: 'product_id',
      key: 'product_id',
    },
    {
      title: '数量',
      dataIndex: 'quantity',
      key: 'quantity',
    },
    {
      title: '总价',
      dataIndex: 'total_price',
      key: 'total_price',
      render: (price: number) => `¥${price.toFixed(2)}`,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const s = statusMap[status] || { color: 'default', label: status }
        return <Tag color={s.color}>{s.label}</Tag>
      },
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
      render: (_: any, record: Order) => (
        <Space>
          <Button
            type="link"
            onClick={() => {
              setSelectedOrder(record)
              setModalVisible(true)
            }}
          >
            详情
          </Button>
          {record.status === 'pending' && (
            <Button type="link" onClick={() => navigate(`/payment/${record.id}`)}>
              支付
            </Button>
          )}
        </Space>
      ),
    },
  ]

  if (error) {
    return <Empty description={`错误: ${error}`} />
  }

  return (
    <div style={{ padding: '20px' }}>
      <h1>订单列表</h1>

      <Spin spinning={isLoading}>
        <Table
          columns={columns}
          dataSource={orders}
          rowKey="id"
          pagination={false}
          locale={{ emptyText: '暂无订单' }}
        />
      </Spin>

      <Modal
        title={`订单详情 #${selectedOrder?.id}`}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
      >
        {selectedOrder && (
          <Descriptions>
            <Descriptions.Item label="订单号">{selectedOrder.id}</Descriptions.Item>
            <Descriptions.Item label="产品ID">{selectedOrder.product_id}</Descriptions.Item>
            <Descriptions.Item label="数量">{selectedOrder.quantity}</Descriptions.Item>
            <Descriptions.Item label="单价">
              ¥{(selectedOrder.total_price / selectedOrder.quantity).toFixed(2)}
            </Descriptions.Item>
            <Descriptions.Item label="总价">
              ¥{selectedOrder.total_price.toFixed(2)}
            </Descriptions.Item>
            <Descriptions.Item label="状态">
              {statusMap[selectedOrder.status]?.label || selectedOrder.status}
            </Descriptions.Item>
            <Descriptions.Item label="创建时间">
              {new Date(selectedOrder.created_at).toLocaleString('zh-CN')}
            </Descriptions.Item>
            <Descriptions.Item label="分组ID">
              {selectedOrder.group_id || '-'}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  )
}

export default OrderListPage
