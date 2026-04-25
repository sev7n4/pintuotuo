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
  Tag,
} from 'antd';
import {
  HeartOutlined,
  HeartFilled,
  DeleteOutlined,
  ShoppingCartOutlined,
  AppstoreOutlined,
} from '@ant-design/icons';
import { useAuthStore } from '@/stores/authStore';
import { useCartStore } from '@/stores/cartStore';
import { favoriteService, type FavoriteListItem, type FavoriteSKUItem } from '@/services/favorite';
import { entitlementPackageService } from '@/services/entitlementPackage';
import styles from './FavoritesPage.module.css';

const { Title, Text } = Typography;

function isSkuFavorite(item: FavoriteListItem): item is FavoriteSKUItem {
  return item.item_type === 'sku';
}

export default function FavoritesPage() {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();
  const { addItem } = useCartStore();
  const [favorites, setFavorites] = useState<FavoriteListItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);

  useEffect(() => {
    if (isAuthenticated) {
      void fetchFavorites();
    }
  }, [isAuthenticated]);

  const fetchFavorites = async () => {
    setLoading(true);
    try {
      const response = await favoriteService.getFavorites();
      if (response.data?.data) {
        const raw = response.data.data.items || [];
        setFavorites(raw as FavoriteListItem[]);
        setTotal(response.data.data.total || 0);
      }
    } catch {
      message.error('获取收藏列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleRemoveSku = async (skuId: number) => {
    try {
      await favoriteService.removeFavorite(skuId);
      setFavorites((prev) => prev.filter((item) => !isSkuFavorite(item) || item.sku_id !== skuId));
      setTotal((t) => Math.max(0, t - 1));
      message.success('已取消收藏');
    } catch {
      message.error('取消收藏失败');
    }
  };

  const handleRemovePackage = async (packageId: number) => {
    try {
      await entitlementPackageService.removeFavorite(packageId);
      setFavorites((prev) =>
        prev.filter((item) =>
          isSkuFavorite(item) ? true : item.entitlement_package_id !== packageId
        )
      );
      setTotal((t) => Math.max(0, t - 1));
      message.success('已取消收藏');
    } catch {
      message.error('取消收藏失败');
    }
  };

  const handleAddToCart = (item: FavoriteSKUItem) => {
    addItem(item.product, 1);
    message.success('已添加到购物车');
  };

  const handleViewSku = (skuId: number) => {
    navigate(`/catalog/${skuId}`);
  };

  const handleViewPackage = (packageCode: string) => {
    navigate(`/packages?pkg=${encodeURIComponent(packageCode)}`);
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
          <Text type="secondary">共 {total} 项</Text>
        </div>

        <Spin spinning={loading}>
          {favorites.length === 0 && !loading ? (
            <Empty
              image={<HeartOutlined style={{ fontSize: 64, color: '#ccc' }} />}
              description="暂无收藏"
            >
              <Button type="primary" onClick={() => navigate('/categories')}>
                去逛逛
              </Button>
            </Empty>
          ) : (
            <List
              grid={{ gutter: 16, xs: 1, sm: 2, md: 3, lg: 4, xl: 4, xxl: 5 }}
              dataSource={favorites}
              renderItem={(item) => {
                if (isSkuFavorite(item)) {
                  return (
                    <List.Item key={`sku-${item.id}`}>
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
                        onClick={() => handleViewSku(item.sku_id)}
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
                                void handleRemoveSku(item.sku_id);
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
                  );
                }
                const ep = item.entitlement_package;
                return (
                  <List.Item key={`pkg-${item.id}`}>
                    <Card
                      hoverable
                      className={styles.productCard}
                      cover={
                        <div className={styles.imageWrapper}>
                          <div
                            style={{
                              height: 160,
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              background: 'linear-gradient(135deg, #f0f5ff 0%, #fff 100%)',
                            }}
                          >
                            <AppstoreOutlined style={{ fontSize: 48, color: '#1677ff' }} />
                          </div>
                        </div>
                      }
                      onClick={() => handleViewPackage(ep.package_code)}
                    >
                      <Card.Meta
                        title={
                          <div className={styles.productName}>
                            <Tag color="blue" style={{ marginRight: 8 }}>
                              套餐包
                            </Tag>
                            {ep.name}
                          </div>
                        }
                        description={
                          <div className={styles.productInfo}>
                            <Text type="secondary" ellipsis>
                              {ep.marketing_line || '组合一口价 · 一键购买'}
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
                            onClick={(e) => {
                              e.stopPropagation();
                              handleViewPackage(ep.package_code);
                            }}
                          >
                            去购买
                          </Button>
                          <Popconfirm
                            title="确定取消收藏？"
                            onConfirm={(e) => {
                              e?.stopPropagation();
                              void handleRemovePackage(item.entitlement_package_id);
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
                );
              }}
            />
          )}
        </Spin>
      </Card>
    </div>
  );
}
