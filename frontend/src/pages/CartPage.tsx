import React from 'react';
import {
  Table,
  Button,
  Space,
  Empty,
  InputNumber,
  Row,
  Col,
  Card,
  Statistic,
  message,
  Tag,
  Typography,
} from 'antd';
import { DeleteOutlined, TagsOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useCartStore } from '@stores/cartStore';
import type { CartItem } from '@/types';

const { Text } = Typography;

export const CartPage: React.FC = () => {
  const navigate = useNavigate();
  const { items, total, removeItem, updateQuantity } = useCartStore();

  if (items.length === 0) {
    return (
      <div style={{ marginTop: 50, textAlign: 'center' }}>
        <Empty description="购物车是空的" />
        <Button type="primary" style={{ marginTop: 16 }} onClick={() => navigate('/products')}>
          继续购物
        </Button>
      </div>
    );
  }

  const columns = [
    {
      title: '产品',
      dataIndex: ['product', 'name'],
      key: 'name',
      render: (text: string, record: CartItem) => (
        <Space direction="vertical" size={0}>
          <a onClick={() => navigate(`/products/${record.product_id}`)}>{text}</a>
          {record.sku_name && (
            <Space size={4}>
              <TagsOutlined style={{ fontSize: 12, color: '#999' }} />
              <Text type="secondary" style={{ fontSize: 12 }}>
                {record.sku_name}
              </Text>
            </Space>
          )}
        </Space>
      ),
    },
    {
      title: '规格',
      key: 'specs',
      render: (_: any, record: CartItem) => {
        if (!record.sku_type) return <Text type="secondary">-</Text>;
        const typeMap: Record<string, { color: string; text: string }> = {
          token_pack: { color: 'blue', text: 'Token包' },
          subscription: { color: 'green', text: '订阅' },
          concurrent: { color: 'orange', text: '并发' },
          trial: { color: 'purple', text: '试用' },
        };
        const config = typeMap[record.sku_type] || { color: 'default', text: record.sku_type };
        return (
          <Space direction="vertical" size={0}>
            <Tag color={config.color}>{config.text}</Tag>
            {record.sku_specs && (
              <Text type="secondary" style={{ fontSize: 12 }}>
                {record.sku_specs}
              </Text>
            )}
          </Space>
        );
      },
    },
    {
      title: '单价',
      dataIndex: ['product', 'price'],
      key: 'price',
      render: (price: number) => `¥${price.toFixed(2)}`,
    },
    {
      title: '数量',
      dataIndex: 'quantity',
      key: 'quantity',
      render: (quantity: number, record: CartItem) => (
        <InputNumber
          min={1}
          value={quantity}
          onChange={(val) => updateQuantity(record.id, val || 1)}
        />
      ),
    },
    {
      title: '小计',
      key: 'subtotal',
      render: (_: any, record: CartItem) => (
        <span>¥{(record.product.price * record.quantity).toFixed(2)}</span>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: CartItem) => (
        <Button
          danger
          icon={<DeleteOutlined />}
          onClick={() => {
            removeItem(record.id);
            message.success('已删除');
          }}
        >
          删除
        </Button>
      ),
    },
  ];

  return (
    <div style={{ padding: '20px' }}>
      <h1>购物车</h1>

      <Table
        columns={columns}
        dataSource={items}
        rowKey="id"
        pagination={false}
        style={{ marginBottom: '20px' }}
      />

      <Row gutter={16}>
        <Col span={24}>
          <Card>
            <Row gutter={16} justify="end">
              <Col>
                <Statistic title="总金额" value={total} prefix="¥" />
              </Col>
              <Col>
                <Space>
                  <Button onClick={() => navigate('/products')}>继续购物</Button>
                  <Button type="primary" size="large" onClick={() => navigate('/checkout')}>
                    去结算
                  </Button>
                </Space>
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default CartPage;
