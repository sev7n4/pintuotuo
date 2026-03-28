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
  Spin,
} from 'antd';
import {
  HeartOutlined,
  HeartFilled,
  DeleteOutlined,
  ShoppingCartOutlined,
} from '@ant-design/icons';
import { useAuthStore } from '@/stores/authStore';
import { useCartStore } from '@/stores/cartStore';
import { favoriteService, FavoriteItem } from '@/services/favorite';
import styles from './FavoritesPage.module.css';

const { Title, Text } = Typography;

export default function FavoritesPage() {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();
  const { addItem } = useCartStore();
  const [favorites, setFavorites] = useState<FavoriteItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);

  useEffect(() => {
    if (isAuthenticated) {
      fetchFavorites();
    }
  }, [isAuthenticated]);

  const fetchFavorites = async () => {
    setLoading(true);
    try {
      const response = await favoriteService.getFavorites();
      if (response.data?.data) {
        setFavorites(response.data.data.items || []);
        setTotal(response.data.data.total || 0);
      }
    } catch {
      message.error('获取收藏列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleRemove = async (productId: number) => {
    try {
      await favoriteService.removeFavorite(productId);
      setFavorites(favorites.filter((item) => item.product_id !== productId));
      setTotal(total - 1);
      message.success('已取消收藏');
    } catch {
      message.error('取消收藏失败');
    }
  };

  const handleAddToCart = (item: FavoriteItem) => {
    addItem(
      {
        id: item.product_id,
        name: item.product.name,
        price: item.product.price,
        image: '',
        description: item.product.description || '',
        stock: item.product.stock,
        merchant_id: item.product.merchant_id,
        category_id: 0,
        status: item.product.status,
        created_at: '',
        updated_at: '',
      },
      1
    );
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
            image={<HeartOutlined style={{ fontSize: 64, color: '#ccc' }} />}
            description="请先登录查看收藏"
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
            <HeartFilled className={styles.icon} />
            我的收藏
          </Title>
          <Text type="secondary">共 {total} 件商品</Text>
        </div>

        <Spin spinning={loading}>
          {favorites.length === 0 && !loading ? (
            <Empty
              image={<HeartOutlined style={{ fontSize: 64, color: '#ccc' }} />}
              description="暂无收藏商品"
            >
              <Button type="primary" onClick={() => navigate('/categories')}>
                去逛逛
              </Button>
            </Empty>
          ) : (
            <List
              grid={{ gutter: 16, xs: 1, sm: 2, md: 3, lg: 4, xl: 4, xxl: 5 }}
              dataSource={favorites}
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
                            {new Date(item.created_at).toLocaleDateString('zh-CN')}
                          </Text>
                        </div>
                      }
                    />
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
                          title="确定取消收藏？"
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
                            移除
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
