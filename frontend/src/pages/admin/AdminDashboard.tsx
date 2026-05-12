import React, { useEffect, useState } from 'react';
import { Card, Row, Col, Statistic, Spin, Typography, Alert, Tag, Space } from 'antd';
import {
  UserOutlined,
  ShopOutlined,
  ShoppingCartOutlined,
  DollarOutlined,
  LineChartOutlined,
  CheckCircleOutlined,
  StopOutlined,
  ApiOutlined,
} from '@ant-design/icons';
import { adminService, AdminStats } from '@/services/admin';
import { ENDPOINT_TYPE_LABELS, ENDPOINT_TYPE_COLORS } from '@/types/sku';

const AdminDashboard: React.FC = () => {
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      try {
        const resp = await adminService.getStats();
        setStats(resp.data.data || null);
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  if (loading) {
    return <Spin />;
  }

  if (!stats) {
    return <Alert type="warning" message="暂无监控数据" showIcon />;
  }

  return (
    <div>
      <Typography.Title level={3} style={{ marginBottom: 24 }}>
        平台运营概览
      </Typography.Title>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card style={{ borderRadius: 12 }}>
            <Statistic
              title="总用户数"
              value={stats.total_users}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card style={{ borderRadius: 12 }}>
            <Statistic
              title="商户数量"
              value={stats.total_merchants}
              prefix={<ShopOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card style={{ borderRadius: 12 }}>
            <Statistic
              title="总订单数"
              value={stats.total_orders}
              prefix={<ShoppingCartOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card style={{ borderRadius: 12 }}>
            <Statistic
              title="平台总收入"
              value={stats.total_revenue}
              precision={2}
              prefix={<DollarOutlined />}
              valueStyle={{ color: '#f5222d' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card style={{ borderRadius: 12 }}>
            <Statistic
              title="订单转化率"
              value={stats.order_conversion_rate * 100}
              precision={2}
              suffix="%"
              prefix={<LineChartOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card style={{ borderRadius: 12 }}>
            <Statistic
              title="支付成功率"
              value={stats.payment_success_rate * 100}
              precision={2}
              suffix="%"
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card style={{ borderRadius: 12 }}>
            <Statistic
              title="订单取消率"
              value={stats.cancellation_rate * 100}
              precision={2}
              suffix="%"
              prefix={<StopOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card style={{ borderRadius: 12 }}>
            <Statistic
              title="多明细订单占比"
              value={stats.multi_item_order_ratio * 100}
              precision={2}
              suffix="%"
              prefix={<ShoppingCartOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Typography.Title level={4} style={{ marginTop: 32, marginBottom: 16 }}>
        端点类型调用分布
      </Typography.Title>
      <Row gutter={[16, 16]}>
        {stats.by_endpoint_type && stats.by_endpoint_type.length > 0 ? (
          stats.by_endpoint_type.map(
            (et: { endpoint_type: string; count: number; tokens: number }) => (
              <Col xs={24} sm={12} lg={6} key={et.endpoint_type}>
                <Card style={{ borderRadius: 12 }}>
                  <Statistic
                    title={
                      <Space>
                        <Tag color={ENDPOINT_TYPE_COLORS[et.endpoint_type] || 'default'}>
                          {ENDPOINT_TYPE_LABELS[et.endpoint_type] || et.endpoint_type}
                        </Tag>
                        调用量
                      </Space>
                    }
                    value={et.count}
                    prefix={<ApiOutlined />}
                  />
                  <div style={{ marginTop: 8, color: '#8c8c8c', fontSize: 12 }}>
                    Token 消耗: {et.tokens.toLocaleString()}
                  </div>
                </Card>
              </Col>
            )
          )
        ) : (
          <Col span={24}>
            <Card style={{ borderRadius: 12, textAlign: 'center', color: '#8c8c8c' }}>
              暂无端点类型统计数据
            </Card>
          </Col>
        )}
      </Row>
    </div>
  );
};

export default AdminDashboard;
