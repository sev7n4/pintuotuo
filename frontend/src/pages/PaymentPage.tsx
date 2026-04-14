import React, { useEffect, useState } from 'react';
import { Card, Button, Radio, Space, Statistic, Divider, message, Spin, Empty, Result, List, Typography } from 'antd';
import { AlipayCircleOutlined, WechatOutlined, CheckCircleOutlined } from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { useOrderStore } from '@stores/orderStore';
import { paymentService } from '@services/payment';
import type { Payment } from '@/types';

type PaymentMethod = 'alipay' | 'wechat';

const PaymentPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { currentOrder, fetchOrderByID, isLoading: orderLoading } = useOrderStore();
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('alipay');
  const [isPaying, setIsPaying] = useState(false);
  const [paymentResult, setPaymentResult] = useState<'success' | 'failed' | null>(null);

  useEffect(() => {
    if (id) {
      fetchOrderByID(parseInt(id));
    }
  }, [id]);

  const handlePayment = async () => {
    if (!currentOrder) return;

    setIsPaying(true);
    try {
      const response = await paymentService.initiatePayment({
        order_id: currentOrder.id,
        pay_method: paymentMethod,
        amount: currentOrder.total_price,
      });
      const payment = response.data.data as Payment & { pay_url?: string; qrcode_url?: string };

      if (payment.pay_url) {
        window.location.href = payment.pay_url;
        return;
      }

      if (payment.status === 'success' || payment.status === 'pending') {
        setPaymentResult('success');
        message.success('支付成功');
      } else {
        setPaymentResult('failed');
        message.error('支付失败，请重试');
      }
    } catch (error) {
      setPaymentResult('failed');
      message.error('支付失败，请稍后重试');
    } finally {
      setIsPaying(false);
    }
  };

  const itemCount =
    currentOrder?.items?.reduce((sum, item) => sum + item.quantity, 0) || currentOrder?.quantity || 0;

  if (orderLoading) {
    return <Spin />;
  }

  if (!currentOrder) {
    return <Empty description="订单不存在" />;
  }

  if (currentOrder.status === 'paid' || currentOrder.status === 'completed') {
    return (
      <Result
        status="success"
        title="订单已支付"
        subTitle={`订单号: #${currentOrder.id}`}
        extra={[
          <Button type="primary" key="orders" onClick={() => navigate('/orders')}>
            查看订单
          </Button>,
          <Button key="products" onClick={() => navigate('/catalog')}>
            继续购物
          </Button>,
        ]}
      />
    );
  }

  if (paymentResult === 'success') {
    return (
      <Result
        status="success"
        icon={<CheckCircleOutlined />}
        title="支付成功"
        subTitle={`订单号: #${currentOrder.id}`}
        extra={[
          <Button type="primary" key="orders" onClick={() => navigate('/orders')}>
            查看订单
          </Button>,
          <Button key="products" onClick={() => navigate('/catalog')}>
            继续购物
          </Button>,
        ]}
      />
    );
  }

  if (paymentResult === 'failed') {
    return (
      <Result
        status="error"
        title="支付失败"
        subTitle="请检查支付方式或稍后重试"
        extra={[
          <Button type="primary" key="retry" onClick={() => setPaymentResult(null)}>
            重新支付
          </Button>,
          <Button key="orders" onClick={() => navigate('/orders')}>
            返回订单
          </Button>,
        ]}
      />
    );
  }

  return (
    <div style={{ padding: '20px', maxWidth: 600, margin: '0 auto' }}>
      <Card title="订单支付">
        <div style={{ marginBottom: '20px' }}>
          <p>
            <strong>订单号:</strong> #{currentOrder.id}
          </p>
          <p>
            <strong>商品项数:</strong> {currentOrder.items?.length || 1}
          </p>
          <p>
            <strong>总数量:</strong> {itemCount}
          </p>
        </div>

        <Card size="small" title="订单明细" style={{ marginBottom: 16, borderRadius: 12 }}>
          <List
            size="small"
            dataSource={currentOrder.items || []}
            locale={{ emptyText: '暂无明细' }}
            renderItem={(item) => (
              <List.Item>
                <Space style={{ justifyContent: 'space-between', width: '100%' }}>
                  <Typography.Text>SKU #{item.sku_id} × {item.quantity}</Typography.Text>
                  <Typography.Text>¥{item.total_price.toFixed(2)}</Typography.Text>
                </Space>
              </List.Item>
            )}
          />
        </Card>

        <Divider />

        <div style={{ textAlign: 'center', marginBottom: '30px' }}>
          <Statistic
            title="支付金额"
            value={currentOrder.total_price}
            precision={2}
            prefix="¥"
            valueStyle={{ fontSize: '32px', color: '#f5222d' }}
          />
        </div>

        <Divider>选择支付方式</Divider>

        <Radio.Group
          value={paymentMethod}
          onChange={(e) => setPaymentMethod(e.target.value)}
          style={{ width: '100%' }}
        >
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <Radio.Button
              value="alipay"
              style={{
                width: '100%',
                height: 60,
                display: 'flex',
                alignItems: 'center',
                padding: '0 20px',
              }}
            >
              <AlipayCircleOutlined style={{ fontSize: 28, color: '#1677ff', marginRight: 12 }} />
              <span style={{ fontSize: 16 }}>支付宝</span>
            </Radio.Button>
            <Radio.Button
              value="wechat"
              style={{
                width: '100%',
                height: 60,
                display: 'flex',
                alignItems: 'center',
                padding: '0 20px',
              }}
            >
              <WechatOutlined style={{ fontSize: 28, color: '#52c41a', marginRight: 12 }} />
              <span style={{ fontSize: 16 }}>微信支付</span>
            </Radio.Button>
          </Space>
        </Radio.Group>

        <Divider />

        <Button
          type="primary"
          size="large"
          block
          loading={isPaying}
          onClick={handlePayment}
          style={{ height: 50, fontSize: 16 }}
        >
          立即支付 ¥{currentOrder.total_price.toFixed(2)}
        </Button>

        <div style={{ marginTop: 16, textAlign: 'center' }}>
          <Button type="link" onClick={() => navigate('/orders')}>
            返回订单列表
          </Button>
        </div>
      </Card>
    </div>
  );
};

export default PaymentPage;
