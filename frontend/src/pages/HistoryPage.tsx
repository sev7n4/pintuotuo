import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Card,
  List,
  Button,
  Empty,
  Image,
  Space,
  Typography,
  Popconfirm,
  message,
  Tag,
  Spin,
} from 'antd';
import {
  HistoryOutlined,
  DeleteOutlined,
  ShoppingCartOutlined,
  ClearOutlined,
} from '@ant-design/icons';
import { useAuthStore } from '@/stores/authStore';
import { useCartStore } from '@/stores/cartStore';
import { browseHistoryService, BrowseHistoryItem } from '@/services/favorite';
import styles from './HistoryPage.module.css';

const { Title, Text } = Typography;

export default function HistoryPage() {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();
  const { addItem } = useCartStore();
  const [history, setHistory] = useState<BrowseHistoryItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);

  useEffect(() => {
    if (isAuthenticated) {
      fetchHistory();
    }
  }, [isAuthenticated]);

  const fetchHistory = async () => {
    setLoading(true);
    try {
      const response = await browseHistoryService.getHistory();
      if (response.data?.data) {
        setHistory(response.data.data.items || []);
        setTotal(response.data.data.total || 0);
      }
    } catch {
      message.error('获取浏览历史失败');
    } finally {
      setLoading(false);
    }
  };

  const handleRemove = async (productId: number) => {
    try {
      await browseHistoryService.removeHistoryItem(productId);
      setHistory(history.filter((item) => item.product_id !== productId));
      setTotal(total - 1);
      message.success('已删除该记录');
    } catch {
      message.error('删除失败');
    }
  };

  const handleClearAll = async () => {
    try {
      await browseHistoryService.clearHistory();
      setHistory([]);
      setTotal(0);
      message.success('已清空浏览历史');
    } catch {
      message.error('清空失败');
    }
  };

  const handleAddToCart = (item: BrowseHistoryItem) => {
    addItem(item.product, 1);
    message.success('已添加到购物车');
  };

  const handleViewProduct = (productId: number) => {
    navigate(`/products/${productId}`);
  };

  if (!isAuthenticated) {
    return (
      <div className={styles.container}>
        <Card>
          <Empty
            image={<HistoryOutlined style={{ fontSize: 64, color: '#ccc' }} />}
            description="请先登录查看浏览历史"
          >
            <Button type="primary" onClick={() => navigate('/login')}>
              立即登录
            </Button>
          </Empty>
        </Card>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <Card>
        <div className={styles.header}>
          <Title level={3} className={styles.title}>
            <HistoryOutlined className={styles.icon} />
            浏览历史
          </Title>
          <Space>
            <Text type="secondary">共 {total} 件商品</Text>
            {total > 0 && (
              <Popconfirm
                title="确定清空所有浏览历史？"
                onConfirm={handleClearAll}
                okText="确定"
                cancelText="取消"
              >
                <Button danger icon={<ClearOutlined />} size="small">
                  清空全部
                </Button>
              </Popconfirm>
            )}
          </Space>
        </div>

        <Spin spinning={loading}>
          {history.length === 0 && !loading ? (
            <Empty
              image={<HistoryOutlined style={{ fontSize: 64, color: '#ccc' }} />}
              description="暂无浏览记录"
            >
              <Button type="primary" onClick={() => navigate('/categories')}>
                去逛逛
              </Button>
            </Empty>
          ) : (
            <List
              grid={{ gutter: 16, xs: 1, sm: 2, md: 3, lg: 4, xl: 4, xxl: 5 }}
              dataSource={history}
              renderItem={(item) => (
                <List.Item>
                  <Card
                    hoverable
                    className={styles.productCard}
                    cover={
                      <div className={styles.imageWrapper}>
                        <Image
                          src="/placeholder.png"
                          alt={item.product.name}
                          className={styles.image}
                          preview={false}
                          fallback="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
                        />
                      </div>
                    }
                    onClick={() => handleViewProduct(item.product_id)}
                  >
                    <Card.Meta
                      title={<div className={styles.productName}>{item.product.name}</div>}
                      description={
                        <div className={styles.productInfo}>
                          <Text type="danger" strong className={styles.price}>
                            ¥{item.product.price.toFixed(2)}
                          </Text>
                          <Text type="secondary" className={styles.date}>
                            {new Date(item.viewed_at).toLocaleDateString('zh-CN')}
                          </Text>
                        </div>
                      }
                    />
                    <div className={styles.meta}>
                      <Tag color="blue">浏览 {item.view_count} 次</Tag>
                    </div>
                    <div className={styles.actions}>
                      <Space>
                        <Button
                          type="primary"
                          size="small"
                          icon={<ShoppingCartOutlined />}
                          onClick={(e) => {
                            e.stopPropagation();
                            handleAddToCart(item);
                          }}
                        >
                          加入购物车
                        </Button>
                        <Popconfirm
                          title="确定删除该记录？"
                          onConfirm={(e) => {
                            e?.stopPropagation();
                            handleRemove(item.product_id);
                          }}
                          onCancel={(e) => e?.stopPropagation()}
                          okText="确定"
                          cancelText="取消"
                        >
                          <Button
                            danger
                            size="small"
                            icon={<DeleteOutlined />}
                            onClick={(e) => e.stopPropagation()}
                          >
                            删除
                          </Button>
                        </Popconfirm>
                      </Space>
                    </div>
                  </Card>
                </List.Item>
              )}
            />
          )}
        </Spin>
      </Card>
    </div>
  );
}
