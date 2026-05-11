import React, { useEffect, useRef, useState } from 'react';
import {
  Card,
  Button,
  Radio,
  Space,
  Statistic,
  Divider,
  message,
  Spin,
  Empty,
  Result,
  List,
  Typography,
  Alert,
} from 'antd';
import { AlipayCircleOutlined, WechatOutlined, CheckCircleOutlined } from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { useOrderStore } from '@stores/orderStore';
import { paymentService } from '@services/payment';
import type { Payment } from '@/types';
import { orderItemLineTitle, getOrderProductSummary } from '@/utils/orderSummary';
import { getApiErrorMessage } from '@/utils/apiError';

type PaymentMethod = 'alipay' | 'wechat';

const paidSuccessExtras = (navigate: (path: string) => void) => [
  <Button type="primary" key="entitlements" onClick={() => navigate('/my/entitlements')}>
    我的权益
  </Button>,
  <Button key="orders" onClick={() => navigate('/orders')}>
    查看订单
  </Button>,
  <Button key="developer" onClick={() => navigate('/developer/quickstart')}>
    开发者中心
  </Button>,
  <Button key="products" onClick={() => navigate('/catalog')}>
    继续购物
  </Button>,
];

const PaymentPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { currentOrder, fetchOrderByID, isLoading: orderLoading } = useOrderStore();
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('alipay');
  const [isPaying, setIsPaying] = useState(false);
  const [paymentResult, setPaymentResult] = useState<'success' | 'failed' | null>(null);
  const pollTicks = useRef(0);

  useEffect(() => {
    if (id) {
      void fetchOrderByID(parseInt(id, 10));
    }
  }, [id, fetchOrderByID]);

  /** 第三方支付返回后，轮询订单状态（单测环境关闭，避免定时器噪音） */
  useEffect(() => {
    const isTestEnv = typeof process !== 'undefined' && process.env.NODE_ENV === 'test';
    if (isTestEnv || !id || paymentResult || !currentOrder || currentOrder.status !== 'pending')
      return;
    pollTicks.current = 0;
    const orderId = parseInt(id, 10);
    const timer = window.setInterval(() => {
      pollTicks.current += 1;
      if (pollTicks.current > 24) {
        window.clearInterval(timer);
        return;
      }
      void fetchOrderByID(orderId);
    }, 4000);
    return () => window.clearInterval(timer);
  }, [id, currentOrder?.id, currentOrder?.status, paymentResult, fetchOrderByID]);

  useEffect(() => {
    const onVis = () => {
      if (document.visibilityState !== 'visible' || !id) return;
      const orderId = parseInt(id, 10);
      if (!Number.isFinite(orderId)) return;
      if (currentOrder?.status === 'pending') {
        void fetchOrderByID(orderId);
      }
    };
    document.addEventListener('visibilitychange', onVis);
    return () => document.removeEventListener('visibilitychange', onVis);
  }, [id, currentOrder?.status, fetchOrderByID]);

  const handleRefreshOrder = () => {
    if (!id) return;
    void fetchOrderByID(parseInt(id, 10));
    message.info('已刷新订单状态');
  };

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
      message.error(getApiErrorMessage(error, '支付失败，请稍后重试'));
    } finally {
      setIsPaying(false);
    }
  };

  const itemCount =
    currentOrder?.items?.reduce((sum, item) => sum + item.quantity, 0) ||
    currentOrder?.quantity ||
    0;

  if (orderLoading) {
    return <Spin style={{ margin: '80px auto', display: 'block' }} size="large" tip="加载订单…" />;
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
        extra={paidSuccessExtras(navigate)}
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
        extra={paidSuccessExtras(navigate)}
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

  const listItems = currentOrder.items && currentOrder.items.length > 0 ? currentOrder.items : null;

  return (
    <div
      style={{
        padding: 12,
        maxWidth: 'min(600px, calc(100vw - 24px))',
        width: '100%',
        margin: '0 auto',
        boxSizing: 'border-box',
        overflowX: 'hidden',
      }}
    >
      <Card title="订单支付" style={{ maxWidth: '100%' }}>
        <Alert
          type="info"
          showIcon
          closable
          style={{ marginBottom: 16 }}
          message="若已在支付应用内完成扣款"
          description="返回本页后若仍显示待支付，请稍等几秒或点击下方「刷新订单状态」；系统确认收款后会自动更新。"
        />

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
          {listItems ? (
            <List
              size="small"
              dataSource={listItems}
              locale={{ emptyText: '暂无明细' }}
              renderItem={(item) => (
                <List.Item>
                  <Space style={{ justifyContent: 'space-between', width: '100%' }} align="start">
                    <Typography.Text ellipsis style={{ flex: 1, marginRight: 8 }}>
                      {orderItemLineTitle(item)} × {item.quantity}
                    </Typography.Text>
                    <Typography.Text>¥{item.total_price.toFixed(2)}</Typography.Text>
                  </Space>
                </List.Item>
              )}
            />
          ) : (
            <Typography.Paragraph style={{ marginBottom: 0 }}>
              {getOrderProductSummary(currentOrder)} · 应付{' '}
              <Typography.Text strong>¥{currentOrder.total_price.toFixed(2)}</Typography.Text>
            </Typography.Paragraph>
          )}
        </Card>

        <Space style={{ marginBottom: 16 }}>
          <Button onClick={handleRefreshOrder}>刷新订单状态</Button>
        </Space>

        <Divider />

        <div style={{ textAlign: 'center', marginBottom: '30px' }}>
          <Statistic
            title="支付金额"
            value={currentOrder.total_price}
            precision={2}
            prefix="¥"
            valueStyle={{
              fontSize: 'clamp(22px, 6vw, 32px)',
              color: '#f5222d',
              wordBreak: 'break-all',
            }}
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
                maxWidth: '100%',
                height: 60,
                display: 'flex',
                alignItems: 'center',
                padding: '0 12px',
                boxSizing: 'border-box',
              }}
            >
              <AlipayCircleOutlined style={{ fontSize: 28, color: '#1677ff', marginRight: 12 }} />
              <span style={{ fontSize: 16 }}>支付宝</span>
            </Radio.Button>
            <Radio.Button
              value="wechat"
              style={{
                width: '100%',
                maxWidth: '100%',
                height: 60,
                display: 'flex',
                alignItems: 'center',
                padding: '0 12px',
                boxSizing: 'border-box',
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
