import React from 'react'
import { Card, Row, Col, Statistic } from 'antd'
import { UserOutlined, ShopOutlined, ShoppingCartOutlined, DollarOutlined } from '@ant-design/icons'

const AdminDashboard: React.FC = () => {
  return (
    <div>
      <h2 style={{ marginBottom: 24 }}>平台运营概览</h2>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总用户数"
              value={1234}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="商户数量"
              value={56}
              prefix={<ShopOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总订单数"
              value={892}
              prefix={<ShoppingCartOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="平台总收入"
              value={125680}
              precision={2}
              prefix={<DollarOutlined />}
              valueStyle={{ color: '#f5222d' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} lg={12}>
          <Card title="最近注册用户">
            <p style={{ color: '#999' }}>暂无数据</p>
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="最近订单">
            <p style={{ color: '#999' }}>暂无数据</p>
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default AdminDashboard
