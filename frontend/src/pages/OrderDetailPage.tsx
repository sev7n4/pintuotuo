import React, { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  Card,
  Button,
  Descriptions,
  Tag,
  Spin,
  Empty,
  Space,
  Typography,
} from 'antd'
import { ArrowLeftOutlined, PayCircleOutlined, TeamOutlined } from '@ant-design/icons'
import { useOrderStore } from '@/stores/orderStore'
import { useProductStore } from '@/stores/productStore'

const { Title, Text } = Typography

const statusMap: Record<string, { color: string; label: string }> = {
  pending: { color: 'orange', label: '待支付' },
  paid: { color: 'blue', label: '已支付' },
  processing: { color: 'cyan', label: '处理中' },
  completed: { color: 'green', label: '已完成' },
  failed: { color: 'red', label: '失败' },
  cancelled: { color: 'gray', label: '已取消' },
}

const groupStatusMap: Record<string, { color: string; label: string }> = {
  active: { color: 'processing', label: '拼团中' },
  completed: { color: 'success', label: '已成团' },
  failed: { color: 'error', label: '拼团失败' },
}

export const OrderDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { currentOrder, isLoading, error, fetchOrderByID } = useOrderStore()
  const { fetchProductByID } = useProductStore()
  const [product, setProduct] = useState<any>(null)

  useEffect(() => {
    if (id) {
      fetchOrderByID(parseInt(id))
    }
  }, [id, fetchOrderByID])

  useEffect(() => {
    if (currentOrder?.product_id) {
      const loadProduct = async () => {
        const p = await fetchProductByID(currentOrder.product_id)
        setProduct(p)
      }
      loadProduct()
    }
  }, [currentOrder, fetchProductByID])

  if (isLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}>
        <Spin size="large" tip="加载中..." />
      </div>
    )
  }

  if (error) {
    return <Empty description={`错误: ${error}`} />
  }

  if (!currentOrder) {
    return <Empty description="订单不存在" />
  }

  const statusInfo = statusMap[currentOrder.status] || statusMap.pending
  const groupStatusInfo = currentOrder.group_id ? groupStatusMap.active : null

  const handlePay = () => {
    navigate(`/payment/${currentOrder.id}`)
  }

  const handleViewGroup = () => {
    if (currentOrder.group_id) {
      navigate(`/groups/${currentOrder.group_id}`)
    }
  }

  return (
    <div style={{ padding: '20px', maxWidth: 800, margin: '0 auto' }}>
      <Card>
        <div style={{ marginBottom: 20 }}>
          <Space align="center">
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/orders')}>
              返回订单列表
            </Button>
            <Title level={3} style={{ margin: 0 }}>
              订单详情
            </Title>
          </Space>
        </div>

        <Descriptions column={2} bordered>
          <Descriptions.Item label="订单号">
            #{currentOrder.id}
          </Descriptions.Item>
          <Descriptions.Item label="商品名称">
            {product?.name || '加载中...'}
          </Descriptions.Item>
          <Descriptions.Item label="单价">
            ¥{currentOrder.unit_price}
          </Descriptions.Item>
          <Descriptions.Item label="数量">
            {currentOrder.quantity}
          </Descriptions.Item>
          <Descriptions.Item label="总价">
            <Text strong style={{ color: '#f5222d', fontSize: 18 }}>
              ¥{currentOrder.total_price}
            </Text>
          </Descriptions.Item>
          <Descriptions.Item label="订单状态">
            <Tag color={statusInfo.color}>{statusInfo.label}</Tag>
          </Descriptions.Item>
          {currentOrder.group_id && (
            <Descriptions.Item label="拼团状态">
              <Tag color={groupStatusInfo?.color || 'default'}>
                {groupStatusInfo?.label || '未知'}
              </Tag>
            </Descriptions.Item>
          )}
          <Descriptions.Item label="创建时间">
            {new Date(currentOrder.created_at).toLocaleString()}
          </Descriptions.Item>
        </Descriptions>

        <div style={{ marginTop: 24, textAlign: 'center' }}>
          <Space size="large">
            {currentOrder.status === 'pending' && (
              <Button
                type="primary"
                size="large"
                icon={<PayCircleOutlined />}
                onClick={handlePay}
              >
                立即支付
              </Button>
            )}
            {currentOrder.group_id && (
              <Button
                size="large"
                icon={<TeamOutlined />}
                onClick={handleViewGroup}
              >
                查看拼团进度
              </Button>
            )}
          </Space>
        </div>
      </Card>
    </div>
  )
}

export default OrderDetailPage
