import React, { useEffect, useState } from 'react'
import {
  Table,
  Button,
  Space,
  Tag,
  Empty,
  Spin,
  Modal,
  Descriptions,
  Select,
  Input,
  message,
  Divider,
  Typography,
} from 'antd'
import { FundOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useOrderStore } from '@stores/orderStore'
import type { Order } from '@/types'

const { Option } = Select
const { TextArea } = Input
const { Text } = Typography

const statusMap: Record<string, { color: string; label: string }> = {
  pending: { color: 'orange', label: '待支付' },
  paid: { color: 'blue', label: '已支付' },
  completed: { color: 'green', label: '已完成' },
  failed: { color: 'red', label: '失败' },
  cancelled: { color: 'gray', label: '已取消' },
  refunding: { color: 'purple', label: '退款中' },
  refunded: { color: 'cyan', label: '已退款' },
}

const cancelReasons = [
  { value: 'changed_mind', label: '不想买了' },
  { value: 'found_better_price', label: '找到更便宜的了' },
  { value: 'wrong_item', label: '拍错商品' },
  { value: 'other', label: '其他原因' },
]

export const OrderListPage: React.FC = () => {
  const navigate = useNavigate()
  const { orders, isLoading, error, fetchOrders, cancelOrder, requestRefund } = useOrderStore()
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null)
  const [modalVisible, setModalVisible] = useState(false)
  const [cancelModalVisible, setCancelModalVisible] = useState(false)
  const [refundModalVisible, setRefundModalVisible] = useState(false)
  const [cancelReason, setCancelReason] = useState<string>('')
  const [cancelReasonText, setCancelReasonText] = useState<string>('')
  const [refundReason, setRefundReason] = useState<string>('')

  useEffect(() => {
    fetchOrders()
  }, [])

  const handleCancelOrder = async () => {
    if (!selectedOrder) return
    
    if (!cancelReason) {
      message.warning('请选择取消原因')
      return
    }

    try {
      await cancelOrder(selectedOrder.id, cancelReason === 'other' ? cancelReasonText : cancelReason)
      message.success('订单已取消')
      setCancelModalVisible(false)
      setCancelReason('')
      setCancelReasonText('')
      fetchOrders()
    } catch (error) {
      message.error('取消订单失败')
    }
  }

  const handleRefundRequest = async () => {
    if (!selectedOrder) return
    
    if (!refundReason) {
      message.warning('请输入退款原因')
      return
    }

    try {
      await requestRefund(selectedOrder.id, refundReason)
      message.success('退款申请已提交')
      setRefundModalVisible(false)
      setRefundReason('')
      fetchOrders()
    } catch (error) {
      message.error('退款申请失败')
    }
  }

  const openCancelModal = (order: Order) => {
    setSelectedOrder(order)
    setCancelModalVisible(true)
  }

  const openRefundModal = (order: Order) => {
    setSelectedOrder(order)
    setRefundModalVisible(true)
  }

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
            <>
              <Button type="link" onClick={() => navigate(`/payment/${record.id}`)}>
                支付
              </Button>
              <Button type="link" danger onClick={() => openCancelModal(record)}>
                取消
              </Button>
            </>
          )}
          {record.status === 'paid' && (
            <Button type="link" icon={<FundOutlined />} onClick={() => openRefundModal(record)}>
              退款
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

      <Modal
        title="取消订单"
        open={cancelModalVisible}
        onCancel={() => {
          setCancelModalVisible(false)
          setCancelReason('')
          setCancelReasonText('')
        }}
        onOk={handleCancelOrder}
        okText="确认取消"
        cancelText="返回"
      >
        <div style={{ marginBottom: 16 }}>
          <Text>请选择取消原因：</Text>
        </div>
        <Select
          style={{ width: '100%', marginBottom: 16 }}
          placeholder="选择取消原因"
          value={cancelReason || undefined}
          onChange={(value) => setCancelReason(value)}
        >
          {cancelReasons.map((reason) => (
            <Option key={reason.value} value={reason.value}>
              {reason.label}
            </Option>
          ))}
        </Select>
        {cancelReason === 'other' && (
          <TextArea
            placeholder="请输入具体原因"
            value={cancelReasonText}
            onChange={(e) => setCancelReasonText(e.target.value)}
            rows={3}
          />
        )}
        {selectedOrder && (
          <>
            <Divider />
            <div style={{ padding: '12px', background: '#f5f5f5', borderRadius: 4 }}>
              <Text strong>退款金额：</Text>
              <Text type="danger" style={{ fontSize: 18, marginLeft: 8 }}>
                ¥{selectedOrder.total_price.toFixed(2)}
              </Text>
            </div>
          </>
        )}
      </Modal>

      <Modal
        title="申请退款"
        open={refundModalVisible}
        onCancel={() => {
          setRefundModalVisible(false)
          setRefundReason('')
        }}
        onOk={handleRefundRequest}
        okText="提交申请"
        cancelText="取消"
      >
        <div style={{ marginBottom: 16 }}>
          <Text>请输入退款原因：</Text>
        </div>
        <TextArea
          placeholder="请详细说明退款原因"
          value={refundReason}
          onChange={(e) => setRefundReason(e.target.value)}
          rows={4}
        />
        {selectedOrder && (
          <>
            <Divider />
            <div style={{ padding: '12px', background: '#f5f5f5', borderRadius: 4 }}>
              <Text strong>可退金额：</Text>
              <Text type="danger" style={{ fontSize: 18, marginLeft: 8 }}>
                ¥{selectedOrder.total_price.toFixed(2)}
              </Text>
              <div style={{ marginTop: 8 }}>
                <Text type="secondary" style={{ fontSize: 12 }}>
                  退款将在1-3个工作日内处理完成
                </Text>
              </div>
            </div>
          </>
        )}
      </Modal>
    </div>
  )
}

export default OrderListPage
