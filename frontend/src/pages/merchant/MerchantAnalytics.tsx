import React, { useEffect, useState } from 'react'
import {
  Card,
  Row,
  Col,
  Statistic,
  Typography,
  Progress,
  Space,
  Spin,
  Empty,
  List,
  Avatar,
  Divider,
  Tag,
  Tabs,
} from 'antd'
import {
  UserOutlined,
  RiseOutlined,
  TeamOutlined,
  EnvironmentOutlined,
  ShoppingCartOutlined,
  TrophyOutlined,
  TagOutlined,
  FireOutlined,
} from '@ant-design/icons'
import { useMerchantStore } from '@/stores/merchantStore'
import styles from './Merchant.module.css'

const { Title, Text } = Typography

interface UserStats {
  new_customers: number
  returning_customers: number
  new_customer_rate: number
  repeat_rate: number
  avg_order_value: number
}

interface RegionData {
  region: string
  count: number
  percentage: number
}

interface ModelPreference {
  model: string
  count: number
  percentage: number
}

interface TopUser {
  id: number
  name: string
  avatar?: string
  total_spent: number
  order_count: number
}

const mockUserStats: UserStats = {
  new_customers: 350,
  returning_customers: 650,
  new_customer_rate: 35,
  repeat_rate: 65,
  avg_order_value: 100.38,
}

const mockRegionData: RegionData[] = [
  { region: '北京', count: 1250, percentage: 25 },
  { region: '上海', count: 1000, percentage: 20 },
  { region: '杭州', count: 750, percentage: 15 },
  { region: '深圳', count: 500, percentage: 10 },
  { region: '南京', count: 400, percentage: 8 },
]

const mockModelPreferences: ModelPreference[] = [
  { model: '编码类', count: 4500, percentage: 45 },
  { model: '文本处理', count: 3000, percentage: 30 },
  { model: '多模态', count: 2500, percentage: 25 },
]

const mockTopUsers: TopUser[] = [
  { id: 1, name: '张三', total_spent: 5680, order_count: 23 },
  { id: 2, name: '李四', total_spent: 4320, order_count: 18 },
  { id: 3, name: '王五', total_spent: 3150, order_count: 15 },
  { id: 4, name: '赵六', total_spent: 2890, order_count: 12 },
  { id: 5, name: '钱七', total_spent: 2340, order_count: 10 },
]

interface UserTag {
  tag: string
  count: number
  percentage: number
  color: string
}

interface UserSegment {
  segment: string
  description: string
  count: number
  percentage: number
  tags: string[]
}

const mockUserTags: UserTag[] = [
  { tag: '高频购买', count: 320, percentage: 32, color: 'red' },
  { tag: '价格敏感', count: 280, percentage: 28, color: 'orange' },
  { tag: '新品偏好', count: 180, percentage: 18, color: 'blue' },
  { tag: '大额消费', count: 120, percentage: 12, color: 'green' },
  { tag: '拼团达人', count: 100, percentage: 10, color: 'purple' },
]

const mockUserSegments: UserSegment[] = [
  { segment: '高价值用户', description: '消费金额高、复购率高', count: 150, percentage: 15, tags: ['大额消费', '高频购买'] },
  { segment: '活跃用户', description: '近期有购买行为', count: 350, percentage: 35, tags: ['新品偏好', '拼团达人'] },
  { segment: '潜力用户', description: '有购买意向但未转化', count: 200, percentage: 20, tags: ['价格敏感'] },
  { segment: '沉睡用户', description: '超过30天未活跃', count: 300, percentage: 30, tags: [] },
]

export const MerchantAnalytics: React.FC = () => {
  const { isLoading, error, fetchStats } = useMerchantStore()
  const [userStats] = useState<UserStats>(mockUserStats)
  const [regionData] = useState<RegionData[]>(mockRegionData)
  const [modelPreferences] = useState<ModelPreference[]>(mockModelPreferences)
  const [topUsers] = useState<TopUser[]>(mockTopUsers)
  const [userTags] = useState<UserTag[]>(mockUserTags)
  const [userSegments] = useState<UserSegment[]>(mockUserSegments)

  useEffect(() => {
    fetchStats()
  }, [fetchStats])

  if (error) {
    return <Empty description={`错误: ${error}`} />
  }

  return (
    <div className={styles.container}>
      <Title level={3} style={{ marginBottom: 24 }}>
        <TeamOutlined style={{ marginRight: 8 }} />
        用户分析
      </Title>

      <Spin spinning={isLoading}>
        <Row gutter={[24, 24]}>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="新客户数"
                value={userStats.new_customers}
                prefix={<UserOutlined />}
                valueStyle={{ color: '#1890ff' }}
              />
              <div style={{ marginTop: 8 }}>
                <Text type="secondary">占比 {userStats.new_customer_rate}%</Text>
              </div>
            </Card>
          </Col>

          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="复购客户数"
                value={userStats.returning_customers}
                prefix={<RiseOutlined />}
                valueStyle={{ color: '#52c41a' }}
              />
              <div style={{ marginTop: 8 }}>
                <Text type="secondary">复购率 {userStats.repeat_rate}%</Text>
              </div>
            </Card>
          </Col>

          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="客单价"
                value={userStats.avg_order_value}
                precision={2}
                prefix="¥"
                valueStyle={{ color: '#faad14' }}
              />
              <div style={{ marginTop: 8 }}>
                <Text type="secondary">平均订单金额</Text>
              </div>
            </Card>
          </Col>

          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="总用户数"
                value={userStats.new_customers + userStats.returning_customers}
                prefix={<TeamOutlined />}
                valueStyle={{ color: '#722ed1' }}
              />
              <div style={{ marginTop: 8 }}>
                <Text type="secondary">累计注册用户</Text>
              </div>
            </Card>
          </Col>
        </Row>

        <Row gutter={[24, 24]} style={{ marginTop: 24 }}>
          <Col xs={24} lg={12}>
            <Card title="用户地域分布 (Top 5)" extra={<EnvironmentOutlined />}>
              <List
                dataSource={regionData}
                renderItem={(item) => (
                  <List.Item>
                    <List.Item.Meta
                      avatar={<Avatar style={{ backgroundColor: '#1890ff' }}>{item.region[0]}</Avatar>}
                      title={item.region}
                      description={`${item.count} 位用户`}
                    />
                    <div style={{ width: 200 }}>
                      <Progress percent={item.percentage} size="small" />
                    </div>
                  </List.Item>
                )}
              />
            </Card>
          </Col>

          <Col xs={24} lg={12}>
            <Card title="模型偏好分布" extra={<ShoppingCartOutlined />}>
              {modelPreferences.map((pref) => (
                <div key={pref.model} style={{ marginBottom: 16 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                    <Text>{pref.model}</Text>
                    <Text type="secondary">{pref.percentage}%</Text>
                  </div>
                  <Progress
                    percent={pref.percentage}
                    strokeColor={{
                      '0%': '#108ee9',
                      '100%': '#87d068',
                    }}
                  />
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    {pref.count.toLocaleString()} 次购买
                  </Text>
                </div>
              ))}
            </Card>
          </Col>
        </Row>

        <Row gutter={[24, 24]} style={{ marginTop: 24 }}>
          <Col xs={24}>
            <Card title="消费TOP 5用户" extra={<TrophyOutlined />}>
              <List
                dataSource={topUsers}
                renderItem={(item, index) => (
                  <List.Item>
                    <List.Item.Meta
                      avatar={
                        <Avatar style={{ 
                          backgroundColor: index === 0 ? '#faad14' : index === 1 ? '#8c8c8c' : index === 2 ? '#cd7f32' : '#1890ff'
                        }}>
                          {index + 1}
                        </Avatar>
                      }
                      title={item.name}
                      description={`${item.order_count} 笔订单`}
                    />
                    <Space direction="vertical" align="end">
                      <Text strong style={{ color: '#f5222d' }}>
                        ¥{item.total_spent.toLocaleString()}
                      </Text>
                      <Text type="secondary" style={{ fontSize: 12 }}>
                        累计消费
                      </Text>
                    </Space>
                  </List.Item>
                )}
              />
            </Card>
          </Col>
        </Row>

        <Card style={{ marginTop: 24 }}>
          <Title level={5}>用户行为分析</Title>
          <Divider />
          <Row gutter={[24, 24]}>
            <Col xs={24} sm={8}>
              <Statistic
                title="平均访问时长"
                value={8.5}
                suffix="分钟"
              />
            </Col>
            <Col xs={24} sm={8}>
              <Statistic
                title="平均浏览商品数"
                value={12}
                suffix="件"
              />
            </Col>
            <Col xs={24} sm={8}>
              <Statistic
                title="转化率"
                value={15.8}
                suffix="%"
              />
            </Col>
          </Row>
        </Card>

        <Card style={{ marginTop: 24 }} title={<><TagOutlined style={{ marginRight: 8 }} />用户标签分析</>}>
          <Tabs defaultActiveKey="tags">
            <Tabs.TabPane tab="用户标签分布" key="tags">
              <Row gutter={[16, 16]}>
                {userTags.map((item) => (
                  <Col xs={12} sm={8} md={4} key={item.tag}>
                    <Card className={styles.tagCard}>
                      <div className={styles.tagHeader}>
                        <Tag color={item.color} style={{ fontSize: 14, padding: '4px 12px' }}>
                          {item.tag}
                        </Tag>
                      </div>
                      <Statistic
                        value={item.count}
                        suffix="人"
                        valueStyle={{ fontSize: 24 }}
                      />
                      <Progress 
                        percent={item.percentage} 
                        size="small" 
                        showInfo={false}
                        strokeColor={item.color}
                      />
                      <Text type="secondary" style={{ fontSize: 12 }}>
                        占比 {item.percentage}%
                      </Text>
                    </Card>
                  </Col>
                ))}
              </Row>
            </Tabs.TabPane>
            <Tabs.TabPane tab="用户分群" key="segments">
              <List
                dataSource={userSegments}
                renderItem={(item) => (
                  <List.Item>
                    <List.Item.Meta
                      avatar={
                        <Avatar 
                          style={{ 
                            backgroundColor: item.segment === '高价值用户' ? '#faad14' : 
                                           item.segment === '活跃用户' ? '#52c41a' :
                                           item.segment === '潜力用户' ? '#1890ff' : '#8c8c8c'
                          }}
                          icon={item.segment === '高价值用户' ? <TrophyOutlined /> : <UserOutlined />}
                        />
                      }
                      title={
                        <Space>
                          <Text strong>{item.segment}</Text>
                          {item.tags.map(tag => (
                            <Tag key={tag} color="blue">{tag}</Tag>
                          ))}
                        </Space>
                      }
                      description={item.description}
                    />
                    <Space direction="vertical" align="end">
                      <Text strong style={{ fontSize: 16 }}>{item.count} 人</Text>
                      <Progress 
                        percent={item.percentage} 
                        size="small" 
                        style={{ width: 100 }}
                        showInfo={false}
                      />
                    </Space>
                  </List.Item>
                )}
              />
            </Tabs.TabPane>
          </Tabs>
        </Card>

        <Card style={{ marginTop: 24 }} title={<><FireOutlined style={{ marginRight: 8, color: '#ff4d4f' }} />热门标签趋势</>}>
          <Row gutter={[16, 16]}>
            <Col xs={24} md={12}>
              <Card type="inner" title="本周新增标签">
                <List
                  size="small"
                  dataSource={[
                    { tag: 'AI工具爱好者', count: 156 },
                    { tag: '企业用户', count: 89 },
                    { tag: 'API重度用户', count: 67 },
                  ]}
                  renderItem={(item) => (
                    <List.Item>
                      <Tag color="blue">{item.tag}</Tag>
                      <Text type="secondary">+{item.count} 人</Text>
                    </List.Item>
                  )}
                />
              </Card>
            </Col>
            <Col xs={24} md={12}>
              <Card type="inner" title="标签转化率">
                <List
                  size="small"
                  dataSource={[
                    { tag: '高频购买', rate: 85 },
                    { tag: '新品偏好', rate: 72 },
                    { tag: '大额消费', rate: 68 },
                  ]}
                  renderItem={(item) => (
                    <List.Item>
                      <Tag color="green">{item.tag}</Tag>
                      <Progress percent={item.rate} size="small" style={{ width: 120 }} />
                    </List.Item>
                  )}
                />
              </Card>
            </Col>
          </Row>
        </Card>
      </Spin>
    </div>
  )
}

export default MerchantAnalytics
