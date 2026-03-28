import React from 'react';
import { Table, Button, Space, Empty, InputNumber, Row, Col, Card, Statistic, message } from 'antd';
import { DeleteOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useCartStore } from '@stores/cartStore';
import type { CartItem } from '@/types';

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
        <a onClick={() => navigate(`/products/${record.product_id}`)}>{text}</a>
      ),
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
