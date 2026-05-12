import React, { useEffect, useMemo, useState } from 'react';
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
import styles from './CartPage.module.css';

const { Text } = Typography;

const typeMap: Record<string, { color: string; text: string }> = {
  token_pack: { color: 'blue', text: 'Token包' },
  subscription: { color: 'green', text: '订阅' },
  concurrent: { color: 'orange', text: '并发' },
  trial: { color: 'purple', text: '试用' },
};

export const CartPage: React.FC = () => {
  const navigate = useNavigate();
  const { items, total, removeItem, updateQuantity } = useCartStore();
  const [isNarrow, setIsNarrow] = useState(false);

  useEffect(() => {
    const mq = window.matchMedia('(max-width: 768px)');
    const apply = () => setIsNarrow(mq.matches);
    apply();
    mq.addEventListener('change', apply);
    return () => mq.removeEventListener('change', apply);
  }, []);

  const columns = useMemo(
    () => [
      {
        title: '产品',
        dataIndex: ['product', 'name'],
        key: 'name',
        render: (text: string, record: CartItem) => (
          <Space direction="vertical" size={0}>
            <a onClick={() => navigate(`/catalog/${record.sku_id}`)}>{text}</a>
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
        render: (_: unknown, record: CartItem) => {
          if (!record.sku_type) return <Text type="secondary">-</Text>;
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
        render: (_: unknown, record: CartItem) => (
          <span>¥{(record.product.price * record.quantity).toFixed(2)}</span>
        ),
      },
      {
        title: '操作',
        key: 'action',
        width: 88,
        render: (_: unknown, record: CartItem) => (
          <Button
            danger
            type="text"
            size="small"
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
    ],
    [navigate, removeItem, updateQuantity]
  );

  if (items.length === 0) {
    return (
      <div style={{ marginTop: 50, textAlign: 'center' }}>
        <Empty description="购物车是空的" />
        <Button type="primary" style={{ marginTop: 16 }} onClick={() => navigate('/catalog')}>
          继续购物
        </Button>
      </div>
    );
  }

  const renderSpec = (record: CartItem) => {
    if (!record.sku_type) return <Text type="secondary">-</Text>;
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
  };

  return (
    <div className={styles.wrap}>
      <h1 className={styles.title}>购物车</h1>

      {isNarrow ? (
        <Space direction="vertical" size={12} className={styles.mobileList}>
          {items.map((record) => (
            <Card key={record.id} size="small" className={styles.mobileCard}>
              <div className={styles.mobileRow}>
                <div className={styles.mobileMain}>
                  <Text strong ellipsis>
                    <a onClick={() => navigate(`/catalog/${record.sku_id}`)}>
                      {record.product.name}
                    </a>
                  </Text>
                  {record.sku_name && (
                    <Text type="secondary" className={styles.mobileSub} ellipsis>
                      {record.sku_name}
                    </Text>
                  )}
                  <div className={styles.mobileMeta}>{renderSpec(record)}</div>
                </div>
                <Button
                  danger
                  type="text"
                  size="small"
                  icon={<DeleteOutlined />}
                  aria-label="删除"
                  onClick={() => {
                    removeItem(record.id);
                    message.success('已删除');
                  }}
                />
              </div>
              <Row gutter={8} align="middle" className={styles.mobileQtyRow}>
                <Col span={12}>
                  <Text type="secondary">单价 ¥{record.product.price.toFixed(2)}</Text>
                </Col>
                <Col span={12} style={{ textAlign: 'right' }}>
                  <Space size={8}>
                    <Text type="secondary">数量</Text>
                    <InputNumber
                      min={1}
                      size="small"
                      value={record.quantity}
                      onChange={(val) => updateQuantity(record.id, val || 1)}
                    />
                  </Space>
                </Col>
              </Row>
              <div className={styles.mobileSubtotal}>
                小计 <Text strong>¥{(record.product.price * record.quantity).toFixed(2)}</Text>
              </div>
            </Card>
          ))}
        </Space>
      ) : (
        <Table
          columns={columns}
          dataSource={items}
          rowKey="id"
          pagination={false}
          scroll={{ x: 'max-content' }}
          className={styles.table}
        />
      )}

      <Card className={styles.summaryCard}>
        <Row gutter={16} justify="space-between" align="middle">
          <Col xs={24} sm={8} style={{ marginBottom: isNarrow ? 12 : 0 }}>
            <Statistic title="总金额" value={total} prefix="¥" />
          </Col>
          <Col xs={24} sm={16} style={{ textAlign: isNarrow ? 'left' : 'right' }}>
            <Space wrap>
              <Button onClick={() => navigate('/catalog')}>继续购物</Button>
              <Button type="primary" size="large" onClick={() => navigate('/checkout')}>
                去结算
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>
    </div>
  );
};

export default CartPage;
