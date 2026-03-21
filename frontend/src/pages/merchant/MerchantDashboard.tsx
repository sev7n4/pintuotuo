import { useEffect } from 'react'
import { Card, Row, Col, Statistic, Table, Tag } from 'antd'
import {
  ShoppingCartOutlined,
  DollarOutlined,
  AppstoreOutlined,
  RiseOutlined,
} from '@ant-design/icons'
import { useMerchantStore } from '@/stores/merchantStore'
import { useAuthStore } from '@/stores/authStore'
import styles from './MerchantDashboard.module.css'

const MerchantDashboard = () => {
  const { stats, orders, fetchStats, fetchOrders, isLoading } = useMerchantStore()
  const { user } = useAuthStore()

  useEffect(() => {
    if (user && user.role === 'merchant') {
      fetchStats()
      fetchOrders(1, 5)
    }
  }, [fetchStats, fetchOrders, user])

  const orderColumns = [
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
      title: '数量',
      dataIndex: 'quantity',
      key: 'quantity',
      width: 80,
    },
    {
      title: '金额',
      dataIndex: 'total_price',
      key: 'total_price',
      width: 100,
      render: (price: number) => `¥${price.toFixed(2)}`,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => {
        const statusMap: Record<string, { color: string; text: string }> = {
          pending: { color: 'default', text: '待支付' },
          paid: { color: 'processing', text: '已支付' },
          completed: { color: 'success', text: '已完成' },
          failed: { color: 'error', text: '失败' },
        }
        const { color, text } = statusMap[status] || { color: 'default', text: status }
        return <Tag color={color}>{text}</Tag>
      },
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 120,
      render: (date: string) => new Date(date).toLocaleDateString('zh-CN'),
    },
  ]

  return (
    <div className={styles.dashboard}>
      <h2 className={styles.pageTitle}>数据概览</h2>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="商品总数"
              value={stats?.total_products || 0}
              prefix={<AppstoreOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="在售商品"
              value={stats?.active_products || 0}
              prefix={<AppstoreOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="本月销售额"
              value={stats?.month_sales || 0}
              precision={2}
              prefix={<DollarOutlined />}
              valueStyle={{ color: '#faad14' }}
              suffix="元"
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="本月订单"
              value={stats?.month_orders || 0}
              prefix={<ShoppingCartOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      <Card
        title="销售趋势"
        className={styles.chartCard}
        extra={<RiseOutlined />}
      >
        <div className={styles.chartPlaceholder}>
          <p>销售趋势图表区域</p>
          <p className={styles.hint}>可集成 ECharts 或其他图表库展示详细数据</p>
        </div>
      </Card>

      <Card title="最近订单" className={styles.orderCard}>
        <Table
          columns={orderColumns}
          dataSource={orders}
          rowKey="id"
          loading={isLoading}
          pagination={false}
          size="small"
        />
      </Card>
    </div>
  )
}

export default MerchantDashboard
