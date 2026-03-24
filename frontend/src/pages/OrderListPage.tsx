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
  Tabs,
  Card,
} from 'antd'
import { FundOutlined, ReloadOutlined, TeamOutlined, ShoppingOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useOrderStore } from '@stores/orderStore'
import { useCartStore } from '@stores/cartStore'
import { useProductStore } from '@stores/productStore'
import type { Order } from '@/types'

const { Option } = Select
const { TextArea } = Input
const { Text } = Typography

const statusMap: Record<string, { color: string; label: string }> = {
  pending: { color: 'orange', label: '待支付' },
  paid: { color: 'blue', label: '已支付' },
  processing: { color: 'cyan', label: '处理中' },
  completed: { color: 'green', label: '已完成' },
  failed: { color: 'red', label: '失败' },
  cancelled: { color: 'gray', label: '已取消' },
  refunding: { color: 'purple', label: '退款中' },
  refunded: { color: 'cyan', label: '已退款' },
}

const groupStatusMap: Record<string, { color: string; label: string }> = {
  active: { color: 'processing', label: '拼团中' },
  completed: { color: 'success', label: '已成团' },
  failed: { color: 'error', label: '拼团失败' },
}

const cancelReasons = [
  { value: 'changed_mind', label: '不想买了' },
  { value: 'found_better_price', label: '找到更便宜的了' },
  { value: 'wrong_item', label: '拍错商品' },
  { value: 'other', label: '其他原因' },
]

const statusTabs = [
  { key: 'all', label: '全部订单' },
  { key: 'pending', label: '待支付' },
  { key: 'processing', label: '进行中' },
  { key: 'completed', label: '已完成' },
  { key: 'cancelled', label: '已取消' },
]

export const OrderListPage: React.FC = () => {
  const navigate = useNavigate()
  const { orders, isLoading, error, fetchOrders, cancelOrder, requestRefund } = useOrderStore()
  const { addItem } = useCartStore()
  const { fetchProductByID } = useProductStore()
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null)
  const [modalVisible, setModalVisible] = useState(false)
  const [cancelModalVisible, setCancelModalVisible] = useState(false)
  const [refundModalVisible, setRefundModalVisible] = useState(false)
  const [cancelReason, setCancelReason] = useState<string>('')
  const [cancelReasonText, setCancelReasonText] = useState<string>('')
  const [refundReason, setRefundReason] = useState<string>('')
  const [activeTab, setActiveTab] = useState<string>('all')

  useEffect(() => {
    fetchOrders()
  }, [fetchOrders])

  const filteredOrders = orders.filter((order) => {
    if (activeTab === 'all') return true
    if (activeTab === 'processing') {
      return ['pending', 'paid', 'processing'].includes(order.status)
    }
    return order.status === activeTab
  })

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
    } catch {
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
    } catch {
      message.error('退款申请失败')
    }
  }

  const handleBuyAgain = async (order: Order) => {
    try {
      const product = await fetchProductByID(order.product_id)
      if (product) {
        addItem(product, order.quantity)
        message.success('已添加到购物车')
        navigate('/cart')
      }
    } catch {
      message.error('商品不存在或已下架')
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
      render: (id: number) => <Text strong>#{id}</Text>,
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
      render: (price: number) => <Text type="danger">¥{price.toFixed(2)}</Text>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string, record: Order) => {
        const s = statusMap[status] || { color: 'default', label: status }
        const groupStatus = record.group_id ? groupStatusMap[record.group_status || 'active'] : null
        return (
          <Space direction="vertical" size="small">
            <Tag color={s.color}>{s.label}</Tag>
            {groupStatus && (
              <Tag color={groupStatus.color} icon={<TeamOutlined />}>
                {groupStatus.label}
              </Tag>
            )}
          </Space>
        )
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
      render: (_: unknown, record: Order) => (
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
          {record.status === 'completed' && (
            <Button 
              type="link" 
              icon={<ReloadOutlined />} 
              onClick={() => handleBuyAgain(record)}
            >
              再次购买
            </Button>
          )}
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
          {record.group_id && record.group_status === 'active' && (
            <Button 
              type="link" 
              icon={<TeamOutlined />} 
              onClick={() => navigate(`/groups/${record.group_id}`)}
            >
              拼团进度
            </Button>
          )}
        </Space>
      ),
    },
  ]

  if (error) {
    return <Empty description={`错误: ${error}`} />
  }

  const tabItems = statusTabs.map((tab) => ({
    key: tab.key,
    label: (
      <span>
        {tab.label}
        {tab.key === 'all' && <Tag style={{ marginLeft: 8 }}>{orders.length}</Tag>}
        {tab.key !== 'all' && (
          <Tag style={{ marginLeft: 8 }}>
            {orders.filter((o) => 
              tab.key === 'processing' 
                ? ['pending', 'paid', 'processing'].includes(o.status)
                : o.status === tab.key
            ).length}
          </Tag>
        )}
      </span>
    ),
  }))

  return (
    <div style={{ padding: '20px', maxWidth: 1200, margin: '0 auto' }}>
      <Card>
        <Space style={{ marginBottom: 16, justifyContent: 'space-between', width: '100%' }}>
          <Text strong style={{ fontSize: 18 }}>
            <ShoppingOutlined style={{ marginRight: 8 }} />
            我的订单
          </Text>
        </Space>

        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={tabItems}
        />

        <Spin spinning={isLoading}>
          <Table
            columns={columns}
            dataSource={filteredOrders}
            rowKey="id"
            pagination={{ pageSize: 10 }}
            locale={{ emptyText: <Empty description="暂无订单" /> }}
          />
        </Spin>
      </Card>

      <Modal
        title={`订单详情 #${selectedOrder?.id}`}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
      >
        {selectedOrder && (
          <Descriptions column={2}>
            <Descriptions.Item label="订单号">{selectedOrder.id}</Descriptions.Item>
            <Descriptions.Item label="产品ID">{selectedOrder.product_id}</Descriptions.Item>
            <Descriptions.Item label="数量">{selectedOrder.quantity}</Descriptions.Item>
            <Descriptions.Item label="单价">
              ¥{(selectedOrder.total_price / selectedOrder.quantity).toFixed(2)}
            </Descriptions.Item>
            <Descriptions.Item label="总价">
              <Text type="danger">¥{selectedOrder.total_price.toFixed(2)}</Text>
            </Descriptions.Item>
            <Descriptions.Item label="状态">
              {statusMap[selectedOrder.status]?.label || selectedOrder.status}
            </Descriptions.Item>
            <Descriptions.Item label="创建时间">
              {new Date(selectedOrder.created_at).toLocaleString('zh-CN')}
            </Descriptions.Item>
            {selectedOrder.group_id && (
              <>
                <Descriptions.Item label="拼团ID">
                  <Button 
                    type="link" 
                    onClick={() => {
                      setModalVisible(false)
                      navigate(`/groups/${selectedOrder.group_id}`)
                    }}
                  >
                    #{selectedOrder.group_id}
                  </Button>
                </Descriptions.Item>
                <Descriptions.Item label="拼团状态">
                  {groupStatusMap[selectedOrder.group_status || 'active']?.label || '进行中'}
                </Descriptions.Item>
              </>
            )}
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
