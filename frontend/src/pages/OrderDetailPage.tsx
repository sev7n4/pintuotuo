import React, { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Card,
  Button,
  Descriptions,
  Tag,
  Spin,
  Empty,
  Space,
  Typography,
  List,
  Alert,
  Collapse,
  message,
} from 'antd';
import {
  ArrowLeftOutlined,
  PayCircleOutlined,
  TeamOutlined,
  ShoppingCartOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { useOrderStore } from '@/stores/orderStore';
import { canReorderFromOrder, orderItemLineTitle } from '@/utils/orderSummary';
import { getApiErrorMessage } from '@/utils/apiError';
import { IconHintButton } from '@/components/IconHintButton';

const { Title, Text } = Typography;

const statusMap: Record<string, { color: string; label: string }> = {
  pending: { color: 'orange', label: '待支付' },
  paid: { color: 'blue', label: '已支付' },
  processing: { color: 'cyan', label: '处理中' },
  completed: { color: 'green', label: '已完成' },
  failed: { color: 'red', label: '失败' },
  cancelled: { color: 'gray', label: '已取消' },
};

const groupStatusMap: Record<string, { color: string; label: string }> = {
  active: { color: 'processing', label: '拼团中' },
  completed: { color: 'success', label: '已成团' },
  failed: { color: 'error', label: '拼团失败' },
};

export const OrderDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { currentOrder, isLoading, error, fetchOrderByID, createOrder } = useOrderStore();
  const [cartLoading, setCartLoading] = useState(false);

  useEffect(() => {
    if (id) {
      void fetchOrderByID(parseInt(id, 10));
    }
  }, [id, fetchOrderByID]);

  const handleQuickReorder = useCallback(async () => {
    if (!currentOrder?.items?.length) {
      message.error('订单无明细，无法复购');
      return;
    }
    try {
      const newId = await createOrder(
        currentOrder.items.map((i) => ({ sku_id: i.sku_id, quantity: i.quantity }))
      );
      if (newId) {
        message.success('已生成新订单');
        navigate(`/payment/${newId}`);
      }
    } catch (e) {
      message.error(getApiErrorMessage(e, '下单失败'));
    }
  }, [createOrder, currentOrder, navigate]);

  const handleAddToCart = useCallback(async () => {
    if (!currentOrder?.items?.length) return;
    setCartLoading(true);
    try {
      const { useCartStore } = await import('@/stores/cartStore');
      const { useProductStore } = await import('@/stores/productStore');
      const { addItem } = useCartStore.getState();
      const { fetchProductByID } = useProductStore.getState();
      for (const line of currentOrder.items) {
        const product = await fetchProductByID(line.sku_id);
        if (product) addItem(product, line.quantity);
      }
      message.success('已加入购物车');
      navigate('/cart');
    } catch {
      message.error('加入购物车失败');
    } finally {
      setCartLoading(false);
    }
  }, [currentOrder, navigate]);

  if (isLoading) {
    return (
      <div
        style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}
      >
        <Spin size="large" tip="加载中..." />
      </div>
    );
  }

  if (error) {
    return <Empty description={`错误: ${error}`} />;
  }

  if (!currentOrder) {
    return <Empty description="订单不存在" />;
  }

  const statusInfo = statusMap[currentOrder.status] || statusMap.pending;
  const groupStatusKey = currentOrder.group_status || 'active';
  const groupStatusInfo = currentOrder.group_id ? groupStatusMap[groupStatusKey] : null;

  const handlePay = () => {
    navigate(`/payment/${currentOrder.id}`);
  };

  const handleViewGroup = () => {
    if (currentOrder.group_id) {
      navigate(`/groups/${currentOrder.group_id}`);
    }
  };

  const itemCount =
    (currentOrder.items || []).reduce((sum, item) => sum + item.quantity, 0) ||
    currentOrder.quantity;

  const summaryLine =
    currentOrder.product_id != null && Number(currentOrder.product_id) > 0
      ? `商品 #${currentOrder.product_id}`
      : (() => {
          const items = currentOrder.items || [];
          if (items.length === 0) return '—';
          const n = items[0].spu_name?.trim() || `规格 #${items[0].sku_id}`;
          return items.length === 1 ? n : `${n} 等 ${items.length} 项`;
        })();

  return (
    <div style={{ padding: '20px', maxWidth: 800, margin: '0 auto' }}>
      <Card>
        <div style={{ marginBottom: 20 }}>
          <Space align="center">
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/orders')}>
              返回订单列表
            </Button>
            <Title level={3} style={{ margin: 0 }}>
              订单详情
            </Title>
          </Space>
        </div>

        <Descriptions column={2} bordered>
          <Descriptions.Item label="订单号">#{currentOrder.id}</Descriptions.Item>
          <Descriptions.Item label="商品摘要">{summaryLine}</Descriptions.Item>
          <Descriptions.Item label="商品项数">{currentOrder.items?.length || 1}</Descriptions.Item>
          <Descriptions.Item label="总数量">{itemCount}</Descriptions.Item>
          <Descriptions.Item label="均价">
            ¥{(currentOrder.total_price / Math.max(itemCount, 1)).toFixed(2)}
          </Descriptions.Item>
          <Descriptions.Item label="总价">
            <Text strong style={{ color: '#f5222d', fontSize: 18 }}>
              ¥{currentOrder.total_price}
            </Text>
          </Descriptions.Item>
          <Descriptions.Item label="订单状态">
            <Tag color={statusInfo.color}>{statusInfo.label}</Tag>
          </Descriptions.Item>
          {currentOrder.group_id && (
            <Descriptions.Item label="拼团状态">
              <Tag color={groupStatusInfo?.color || 'default'}>
                {groupStatusInfo?.label || currentOrder.group_status || '—'}
              </Tag>
            </Descriptions.Item>
          )}
          <Descriptions.Item label="创建时间">
            {new Date(currentOrder.created_at).toLocaleString()}
          </Descriptions.Item>
        </Descriptions>

        {(currentOrder.items?.length ?? 0) > 1 && (
          <Alert
            type="info"
            showIcon
            style={{ marginTop: 16 }}
            message="多明细订单"
            description="支付成功后，系统按每条明细分别履约（如订阅、Token 等）。对账与售后以明细为准。"
          />
        )}

        <Card
          size="small"
          title="订单明细"
          style={{ marginTop: 16, borderRadius: 12 }}
          bodyStyle={{ padding: 12 }}
        >
          <Collapse
            defaultActiveKey={['lines']}
            items={[
              {
                key: 'lines',
                label: `共 ${currentOrder.items?.length || 0} 条（展开查看名称、类型、数量与金额）`,
                children: (
                  <List
                    size="small"
                    dataSource={currentOrder.items || []}
                    locale={{ emptyText: '暂无明细' }}
                    renderItem={(item) => (
                      <List.Item>
                        <Space direction="vertical" size={4} style={{ width: '100%' }}>
                          <Space style={{ justifyContent: 'space-between', width: '100%' }}>
                            <Text>{orderItemLineTitle(item)}</Text>
                            <Text strong>¥{item.total_price.toFixed(2)}</Text>
                          </Space>
                          <Text type="secondary" style={{ fontSize: 12 }}>
                            数量 {item.quantity} · 单价 ¥{item.unit_price.toFixed(2)}
                          </Text>
                        </Space>
                      </List.Item>
                    )}
                  />
                ),
              },
            ]}
          />
        </Card>

        <div style={{ marginTop: 24, textAlign: 'center' }}>
          <Space size="middle" wrap>
            {currentOrder.status === 'pending' && (
              <Button type="primary" size="large" icon={<PayCircleOutlined />} onClick={handlePay}>
                立即支付
              </Button>
            )}
            {canReorderFromOrder(currentOrder) && (
              <>
                <IconHintButton
                  size="large"
                  hint="加入购物车"
                  icon={<ShoppingCartOutlined />}
                  loading={cartLoading}
                  onClick={() => void handleAddToCart()}
                />
                <Button
                  type="primary"
                  size="large"
                  icon={<ThunderboltOutlined />}
                  onClick={() => void handleQuickReorder()}
                >
                  同配置再下一单
                </Button>
              </>
            )}
            {currentOrder.group_id && (
              <Button size="large" icon={<TeamOutlined />} onClick={handleViewGroup}>
                查看拼团进度
              </Button>
            )}
          </Space>
        </div>
      </Card>
    </div>
  );
};

export default OrderDetailPage;
