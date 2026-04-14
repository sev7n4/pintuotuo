import React, { useState } from 'react';
import {
  Card,
  Table,
  Button,
  Radio,
  Space,
  Statistic,
  Divider,
  message,
  Empty,
  Row,
  Col,
  Checkbox,
  List,
  Grid,
} from 'antd';
import { AlipayCircleOutlined, WechatOutlined, ArrowLeftOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useCartStore } from '@stores/cartStore';
import { useOrderStore } from '@stores/orderStore';
import type { CartItem } from '@/types';

type PaymentMethod = 'alipay' | 'wechat';

const CheckoutPage: React.FC = () => {
  const { useBreakpoint } = Grid;
  const screens = useBreakpoint();
  const isMobile = screens.xs || (screens.sm && !screens.md);
  const navigate = useNavigate();
  const { items, clear } = useCartStore();
  const { createOrder, isLoading, fetchOrders } = useOrderStore();
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('alipay');
  const [selectedItems, setSelectedItems] = useState<string[]>(items.map((item) => item.id));

  const selectedCartItems = items.filter((item) => selectedItems.includes(item.id));
  const selectedTotal = selectedCartItems.reduce(
    (sum, item) => sum + item.product.price * item.quantity,
    0
  );

  const handleSelectItem = (itemId: string) => {
    setSelectedItems((prev) =>
      prev.includes(itemId) ? prev.filter((id) => id !== itemId) : [...prev, itemId]
    );
  };

  const handleSelectAll = () => {
    if (selectedItems.length === items.length) {
      setSelectedItems([]);
    } else {
      setSelectedItems(items.map((item) => item.id));
    }
  };

  const handleCheckout = async () => {
    if (selectedCartItems.length === 0) {
      message.warning('请选择要结算的商品');
      return;
    }

    try {
      const orderId = await createOrder(
        selectedCartItems.map((item) => ({
          sku_id: item.sku_id,
          quantity: item.quantity,
        }))
      );

      clear();
      await fetchOrders();
      if (!orderId) {
        message.success('订单创建成功');
        navigate('/orders');
        return;
      }
      message.success('订单创建成功，正在跳转到支付页面');
      navigate(`/payment/${orderId}`);
    } catch (error) {
      message.error('创建订单失败，请重试');
    }
  };

  if (items.length === 0) {
    return (
      <div style={{ marginTop: 50, textAlign: 'center' }}>
        <Empty description="购物车是空的">
          <Button type="primary" onClick={() => navigate('/catalog')}>
            去购物
          </Button>
        </Empty>
      </div>
    );
  }

  const columns = [
    {
      title: (
        <Checkbox checked={selectedItems.length === items.length} onChange={handleSelectAll}>
          全选
        </Checkbox>
      ),
      key: 'select',
      width: 80,
      render: (_: any, record: CartItem) => (
        <Checkbox
          checked={selectedItems.includes(record.id)}
          onChange={() => handleSelectItem(record.id)}
        />
      ),
    },
    {
      title: '商品',
      dataIndex: ['product', 'name'],
      key: 'name',
      render: (text: string, record: CartItem) => (
        <a onClick={() => navigate(`/catalog/${record.sku_id}`)}>{text}</a>
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
    },
    {
      title: '小计',
      key: 'subtotal',
      render: (_: any, record: CartItem) => (
        <span style={{ color: '#f5222d', fontWeight: 'bold' }}>
          ¥{(record.product.price * record.quantity).toFixed(2)}
        </span>
      ),
    },
  ];

  return (
    <div style={{ padding: '20px', maxWidth: 1000, margin: '0 auto' }}>
      <Button
        type="text"
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate('/cart')}
        style={{ marginBottom: '20px' }}
      >
        返回购物车
      </Button>

      <Row gutter={24}>
        <Col xs={24} lg={16}>
          <Card title="确认订单" style={{ borderRadius: 12 }}>
            {isMobile ? (
              <List
                dataSource={items}
                renderItem={(record) => (
                  <List.Item>
                    <Card size="small" style={{ width: '100%', borderRadius: 10 }}>
                      <Space direction="vertical" style={{ width: '100%' }} size={6}>
                        <Space style={{ justifyContent: 'space-between', width: '100%' }}>
                          <Checkbox
                            checked={selectedItems.includes(record.id)}
                            onChange={() => handleSelectItem(record.id)}
                          >
                            {record.product.name}
                          </Checkbox>
                          <span style={{ color: '#f5222d', fontWeight: 600 }}>
                            ¥{(record.product.price * record.quantity).toFixed(2)}
                          </span>
                        </Space>
                        <span>单价 ¥{record.product.price.toFixed(2)} · 数量 {record.quantity}</span>
                      </Space>
                    </Card>
                  </List.Item>
                )}
              />
            ) : (
              <Table columns={columns} dataSource={items} rowKey="id" pagination={false} />
            )}
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card title="订单结算">
            <div style={{ marginBottom: '20px' }}>
              <p>已选商品: {selectedCartItems.length} 类</p>
            </div>

            <Divider>选择支付方式</Divider>

            <Radio.Group
              value={paymentMethod}
              onChange={(e) => setPaymentMethod(e.target.value)}
              style={{ width: '100%' }}
            >
              <Space direction="vertical" style={{ width: '100%' }} size="small">
                <Radio.Button
                  value="alipay"
                  style={{
                    width: '100%',
                    height: 50,
                    display: 'flex',
                    alignItems: 'center',
                    padding: '0 16px',
                  }}
                >
                  <AlipayCircleOutlined
                    style={{ fontSize: 24, color: '#1677ff', marginRight: 10 }}
                  />
                  <span>支付宝</span>
                </Radio.Button>
                <Radio.Button
                  value="wechat"
                  style={{
                    width: '100%',
                    height: 50,
                    display: 'flex',
                    alignItems: 'center',
                    padding: '0 16px',
                  }}
                >
                  <WechatOutlined style={{ fontSize: 24, color: '#52c41a', marginRight: 10 }} />
                  <span>微信支付</span>
                </Radio.Button>
              </Space>
            </Radio.Group>

            <Divider />

            <div style={{ textAlign: 'center', marginBottom: '20px' }}>
              <Statistic
                title="应付金额"
                value={selectedTotal}
                precision={2}
                prefix="¥"
                valueStyle={{ fontSize: '24px', color: '#f5222d' }}
              />
            </div>

            <Button
              type="primary"
              size="large"
              block
              loading={isLoading}
              disabled={selectedCartItems.length === 0}
              onClick={handleCheckout}
              style={{ height: 48 }}
            >
              提交订单
            </Button>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default CheckoutPage;
