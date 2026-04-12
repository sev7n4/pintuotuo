import { useEffect } from 'react';
import { Card, Row, Col, Statistic, Table, Tag, Progress, List, Avatar, Empty } from 'antd';
import {
  ShoppingCartOutlined,
  DollarOutlined,
  AppstoreOutlined,
  RiseOutlined,
  TeamOutlined,
  TrophyOutlined,
  FireOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
} from '@ant-design/icons';
import { useMerchantStore } from '@/stores/merchantStore';
import { useAuthStore } from '@/stores/authStore';
import styles from './MerchantDashboard.module.css';

const MerchantDashboard = () => {
  const { stats, orders, fetchStats, fetchOrders, isLoading } = useMerchantStore();
  const { user } = useAuthStore();

  useEffect(() => {
    if (user && user.role === 'merchant') {
      fetchStats();
      fetchOrders(1, 5);
    }
  }, [fetchStats, fetchOrders, user]);

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
      title: '金额(¥)',
      dataIndex: 'total_price',
      key: 'total_price',
      width: 100,
      render: (price: number) => `¥${price.toFixed(6)}`,
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
        };
        const { color, text } = statusMap[status] || { color: 'default', text: status };
        return <Tag color={color}>{text}</Tag>;
      },
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 120,
      render: (date: string) => new Date(date).toLocaleDateString('zh-CN'),
    },
  ];

  const groupSuccessRate = stats?.group_success_rate || 78.5;

  const hotProducts = [
    { id: 1, name: 'GPT-4 API Token', sales: 156, revenue: 15600, trend: 12 },
    { id: 2, name: 'Claude 3 Opus', sales: 132, revenue: 13200, trend: 8 },
    { id: 3, name: 'Gemini Pro', sales: 98, revenue: 9800, trend: -3 },
    { id: 4, name: 'Midjourney Credits', sales: 87, revenue: 8700, trend: 15 },
    { id: 5, name: 'DALL-E 3 API', sales: 76, revenue: 7600, trend: 5 },
  ];

  const salesTrendData = [
    { day: '周一', value: 65 },
    { day: '周二', value: 78 },
    { day: '周三', value: 52 },
    { day: '周四', value: 89 },
    { day: '周五', value: 95 },
    { day: '周六', value: 72 },
    { day: '周日', value: 85 },
  ];

  const maxValue = Math.max(...salesTrendData.map((d) => d.value));

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

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} sm={12} lg={8}>
          <Card
            className={styles.statCard}
            title={
              <span>
                <TeamOutlined style={{ marginRight: 8, color: '#1890ff' }} />
                成团率统计
              </span>
            }
          >
            <div className={styles.groupRateContent}>
              <div className={styles.groupRateCircle}>
                <Progress
                  type="circle"
                  percent={groupSuccessRate}
                  strokeColor={{
                    '0%': '#1890ff',
                    '100%': '#52c41a',
                  }}
                  strokeWidth={10}
                  width={120}
                />
              </div>
              <div className={styles.groupRateStats}>
                <div className={styles.groupRateItem}>
                  <span className={styles.groupRateLabel}>成功成团</span>
                  <span className={styles.groupRateValue}>{stats?.success_groups || 45}</span>
                </div>
                <div className={styles.groupRateItem}>
                  <span className={styles.groupRateLabel}>进行中</span>
                  <span className={styles.groupRateValue}>{stats?.pending_groups || 12}</span>
                </div>
                <div className={styles.groupRateItem}>
                  <span className={styles.groupRateLabel}>已失败</span>
                  <span className={styles.groupRateValue}>{stats?.failed_groups || 8}</span>
                </div>
              </div>
            </div>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={8}>
          <Card
            className={styles.statCard}
            title={
              <span>
                <FireOutlined style={{ marginRight: 8, color: '#ff4d4f' }} />
                热销商品 TOP5
              </span>
            }
          >
            <List
              dataSource={hotProducts}
              renderItem={(item, index) => (
                <List.Item className={styles.hotProductItem}>
                  <div className={styles.hotProductRank}>
                    {index < 3 ? (
                      <Avatar
                        size={24}
                        style={{
                          backgroundColor:
                            index === 0 ? '#ff4d4f' : index === 1 ? '#faad14' : '#52c41a',
                          fontSize: 12,
                        }}
                      >
                        {index + 1}
                      </Avatar>
                    ) : (
                      <span className={styles.rankNumber}>{index + 1}</span>
                    )}
                  </div>
                  <div className={styles.hotProductInfo}>
                    <div className={styles.hotProductName}>{item.name}</div>
                    <div className={styles.hotProductMeta}>
                      <span>销量: {item.sales}</span>
                      <span>¥{item.revenue.toLocaleString()}</span>
                    </div>
                  </div>
                  <div className={styles.hotProductTrend}>
                    {item.trend > 0 ? (
                      <ArrowUpOutlined style={{ color: '#52c41a' }} />
                    ) : (
                      <ArrowDownOutlined style={{ color: '#ff4d4f' }} />
                    )}
                    <span style={{ color: item.trend > 0 ? '#52c41a' : '#ff4d4f' }}>
                      {item.trend > 0 ? '+' : ''}
                      {item.trend}%
                    </span>
                  </div>
                </List.Item>
              )}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={8}>
          <Card
            className={styles.statCard}
            title={
              <span>
                <TrophyOutlined style={{ marginRight: 8, color: '#faad14' }} />
                本周业绩
              </span>
            }
          >
            <div className={styles.weeklyStats}>
              <div className={styles.weeklyStat}>
                <span className={styles.weeklyLabel}>本周销售额</span>
                <span className={styles.weeklyValue}>
                  ¥{(stats?.week_sales || 25680).toLocaleString()}
                </span>
              </div>
              <div className={styles.weeklyStat}>
                <span className={styles.weeklyLabel}>环比增长</span>
                <span className={styles.weeklyValue} style={{ color: '#52c41a' }}>
                  +{stats?.week_growth || 15.8}%
                </span>
              </div>
              <div className={styles.weeklyStat}>
                <span className={styles.weeklyLabel}>新增客户</span>
                <span className={styles.weeklyValue}>{stats?.new_customers || 23}</span>
              </div>
            </div>
          </Card>
        </Col>
      </Row>

      <Card
        title={
          <span>
            <RiseOutlined style={{ marginRight: 8 }} />
            销售趋势
          </span>
        }
        className={styles.chartCard}
      >
        <div className={styles.chartContainer}>
          <div className={styles.chartBars}>
            {salesTrendData.map((item, index) => (
              <div key={index} className={styles.chartBarWrapper}>
                <div
                  className={styles.chartBar}
                  style={{
                    height: `${(item.value / maxValue) * 100}%`,
                    backgroundColor: index === salesTrendData.length - 1 ? '#1890ff' : '#91d5ff',
                  }}
                >
                  <span className={styles.chartValue}>¥{item.value * 100}</span>
                </div>
                <span className={styles.chartLabel}>{item.day}</span>
              </div>
            ))}
          </div>
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
          scroll={{ x: 800 }}
          locale={{ emptyText: <Empty description="暂无订单数据" /> }}
        />
      </Card>
    </div>
  );
};

export default MerchantDashboard;
