import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  Carousel,
  Card,
  Row,
  Col,
  Spin,
  Tag,
  Input,
  Typography,
  Space,
  Button,
  Skeleton,
} from 'antd';
import {
  FireOutlined,
  ClockCircleOutlined,
  SearchOutlined,
  RightOutlined,
  ThunderboltOutlined,
  GiftOutlined,
  FilterOutlined,
} from '@ant-design/icons';
import { useHomeStore } from '@/stores/homeStore';
import { Product } from '@/types';
import {
  getProductCardSubtitle,
  pushRecentSearch,
  readRecentSearches,
} from '@/utils/productDisplay';
import { getProviderCardSurfaceStyle, getProviderLogoUrl } from '@/utils/providerBrand';
import styles from './HomePage.module.css';

const { Title, Text } = Typography;
const { Search } = Input;

interface QuickNav {
  key: string;
  name: string;
  icon: React.ReactNode;
  link: string;
}

const quickNavItems: QuickNav[] = [
  {
    key: 'hot',
    name: '热销',
    icon: <FireOutlined />,
    link: '/catalog?sort=hot',
  },
  {
    key: 'group',
    name: '拼团',
    icon: <GiftOutlined />,
    link: '/catalog?group_enabled=true',
  },
  {
    key: 'flash',
    name: '秒杀',
    icon: <ThunderboltOutlined />,
    link: '/catalog?flash=true',
  },
  {
    key: 'new',
    name: '新品',
    icon: <ClockCircleOutlined />,
    link: '/catalog?sort=new',
  },
  {
    key: 'fuel',
    name: '加油站',
    icon: <ThunderboltOutlined />,
    link: '/fuel-station',
  },
];

function coverImageUrl(p: Product) {
  return p.image_url || p.thumbnail_url;
}

function HomeProductCardCover({
  product,
  showGroupTag,
  discount,
}: {
  product: Product;
  showGroupTag: boolean;
  discount: number;
}) {
  const imgUrl = coverImageUrl(product);
  const brandUrl =
    !imgUrl && product.model_provider ? getProviderLogoUrl(product.model_provider) : null;
  const surface =
    !imgUrl && product.model_provider
      ? getProviderCardSurfaceStyle(product.model_provider)
      : undefined;
  const [brandBroken, setBrandBroken] = useState(false);

  return (
    <div className={styles.productImage} style={surface ? { background: surface } : undefined}>
      {imgUrl ? (
        <img
          src={imgUrl}
          alt=""
          className={styles.productCoverImg}
          loading="lazy"
          decoding="async"
        />
      ) : brandUrl && !brandBroken ? (
        <img
          src={brandUrl}
          alt=""
          className={styles.productBrandLogo}
          loading="lazy"
          decoding="async"
          onError={() => setBrandBroken(true)}
        />
      ) : (
        <div className={styles.productPlaceholder}>
          <Text type="secondary">{product.name.substring(0, 2)}</Text>
        </div>
      )}
      {discount > 0 && (
        <Tag className={styles.discountTag} bordered={false}>
          -{discount}%
        </Tag>
      )}
      {showGroupTag && (
        <Tag color="#52c41a" className={styles.groupTag}>
          拼团
        </Tag>
      )}
    </div>
  );
}

const HomePage = () => {
  const navigate = useNavigate();
  const {
    banners,
    hotProducts,
    newProducts,
    categories,
    scenarioCategories,
    isLoading,
    error,
    fetchHomeData,
  } = useHomeStore();

  const [recommendedProducts, setRecommendedProducts] = useState<Product[]>([]);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [recentSearches, setRecentSearches] = useState<string[]>(() => readRecentSearches());

  useEffect(() => {
    fetchHomeData();
  }, [fetchHomeData]);

  useEffect(() => {
    const seen = new Set<number>();
    const mixed: Product[] = [];
    for (const p of [...hotProducts, ...newProducts]) {
      if (!seen.has(p.id)) {
        seen.add(p.id);
        mixed.push(p);
      }
    }
    setRecommendedProducts(mixed.slice(0, 12));
  }, [hotProducts, newProducts]);

  const handleSearch = (value: string) => {
    const q = value.trim();
    if (q) {
      pushRecentSearch(q);
      setRecentSearches(readRecentSearches());
      navigate(`/catalog?q=${encodeURIComponent(q)}`);
    }
  };

  const handleProductClick = (productId: number) => {
    navigate(`/catalog/${productId}`);
  };

  const handleScenarioNavigate = (code: string) => {
    navigate(`/catalog?scenario=${encodeURIComponent(code)}`);
  };

  const handleTierNavigate = (tierName: string) => {
    navigate(`/catalog?tier=${encodeURIComponent(tierName)}`);
  };

  const formatPrice = (price: number) => {
    return `¥${price.toFixed(2)}`;
  };

  const renderProductCard = (product: Product, showGroupTag = false) => {
    const discount =
      product.original_price && product.original_price > product.price
        ? Math.round((1 - product.price / product.original_price) * 100)
        : 0;
    const subtitle = getProductCardSubtitle(product);

    return (
      <Card
        hoverable
        className={styles.productCard}
        cover={
          <HomeProductCardCover product={product} showGroupTag={showGroupTag} discount={discount} />
        }
        onClick={() => handleProductClick(product.id)}
      >
        <div className={styles.productInfo}>
          <Text className={styles.productName} ellipsis>
            {product.name}
          </Text>
          <Text type="secondary" className={styles.productSubtitle} ellipsis>
            {subtitle}
          </Text>
          <div className={styles.priceRow}>
            <Text type="danger" strong className={styles.price}>
              {formatPrice(product.price)}
            </Text>
            {product.original_price && product.original_price > product.price && (
              <Text delete type="secondary" className={styles.originalPrice}>
                {formatPrice(product.original_price)}
              </Text>
            )}
          </div>
          <div className={styles.productMeta}>
            <Text type="secondary" className={styles.soldCount}>
              已售 {product.sold_count || 0}
            </Text>
            <Text type="secondary">
              {product.rating != null && product.rating > 0
                ? `${product.rating.toFixed(1)}分`
                : `库存 ${product.stock}`}
            </Text>
          </div>
        </div>
      </Card>
    );
  };

  const hasFeed = recommendedProducts.length > 0;

  if (error) {
    return (
      <div className={styles.errorContainer}>
        <Text type="danger">{error}</Text>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <div className={styles.searchSection}>
        <Space.Compact style={{ width: '100%', maxWidth: 720 }}>
          <Search
            placeholder="搜索模型或关键词"
            allowClear
            enterButton={<SearchOutlined />}
            size="large"
            value={searchKeyword}
            onChange={(e) => setSearchKeyword(e.target.value)}
            onSearch={handleSearch}
            className={styles.searchInput}
            style={{ flex: 1, minWidth: 0 }}
          />
          <Button
            size="large"
            icon={<FilterOutlined />}
            onClick={() => navigate('/catalog?filters=1')}
          >
            筛选
          </Button>
        </Space.Compact>
        {recentSearches.length > 0 && (
          <div className={styles.recentSearch}>
            <Text type="secondary" className={styles.recentLabel}>
              最近搜索：
            </Text>
            <Space size={[8, 8]} wrap>
              {recentSearches.slice(0, 6).map((q) => (
                <Tag
                  key={q}
                  className={styles.recentTag}
                  onClick={() => {
                    setSearchKeyword(q);
                    handleSearch(q);
                  }}
                >
                  {q}
                </Tag>
              ))}
            </Space>
          </div>
        )}
      </div>

      <div className={styles.quickNavSection}>
        <div className={styles.quickNavScroll}>
          {quickNavItems.map((item) => (
            <button
              type="button"
              key={item.key}
              className={styles.quickNavPill}
              onClick={() => navigate(item.link)}
            >
              <span className={styles.quickNavIconMuted}>{item.icon}</span>
              <span className={styles.quickNavName}>{item.name}</span>
            </button>
          ))}
        </div>
      </div>

      <Card style={{ marginBottom: 16, borderRadius: 12 }}>
        <Space direction="vertical" size={6}>
          <Text strong>套餐包专区</Text>
          <Text type="secondary">多模型与赠送 Token 组合定价，一步开通可用权益。</Text>
          <Button type="primary" onClick={() => navigate('/packages')}>
            查看套餐包
          </Button>
        </Space>
      </Card>

      {banners.length > 0 && (
        <Carousel autoplay className={styles.bannerCarousel}>
          {banners.map((banner) => (
            <div key={banner.id} className={styles.bannerItem}>
              <div className={styles.bannerContent} onClick={() => navigate(banner.link)}>
                <Text className={styles.bannerTitle}>{banner.title}</Text>
              </div>
            </div>
          ))}
        </Carousel>
      )}

      {(scenarioCategories.length > 0 || categories.length > 0) && (
        <div className={styles.categorySection}>
          <div className={styles.categoryHeaderRow}>
            <Title level={5} className={styles.categoryTitle}>
              分类
            </Title>
            <Link to="/categories" className={styles.allCategoriesLink}>
              全部分类 <RightOutlined />
            </Link>
          </div>
          {scenarioCategories.length > 0 && (
            <>
              <Text type="secondary" className={styles.categoryGroupLabel}>
                使用场景
              </Text>
              <div className={styles.categoryChips}>
                {scenarioCategories.map((s) => (
                  <button
                    type="button"
                    key={s.code}
                    className={styles.categoryChip}
                    onClick={() => handleScenarioNavigate(s.code)}
                  >
                    <span className={styles.categoryChipName}>{s.name}</span>
                    <span className={styles.categoryChipCount}>{s.count}</span>
                  </button>
                ))}
              </div>
            </>
          )}
          {categories.length > 0 && (
            <>
              <Text type="secondary" className={styles.categoryGroupLabel}>
                模型层级
              </Text>
              <div className={styles.categoryChips}>
                {categories.map((category) => (
                  <button
                    type="button"
                    key={category.name}
                    className={`${styles.categoryChip} ${styles.categoryChipSubtle}`}
                    onClick={() => handleTierNavigate(category.name)}
                  >
                    <span className={styles.categoryChipName}>{category.name}</span>
                    <span className={styles.categoryChipCount}>{category.count}</span>
                  </button>
                ))}
              </div>
            </>
          )}
        </div>
      )}

      {isLoading && !hasFeed ? (
        <div className={styles.skeletonWrap}>
          <Skeleton active paragraph={{ rows: 1 }} title={{ width: '40%' }} />
          <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
            {[1, 2, 3, 4].map((k) => (
              <Col xs={12} sm={8} md={6} lg={4} key={k}>
                <Card>
                  <Skeleton.Image active className={styles.skeletonImg} />
                  <Skeleton active paragraph={{ rows: 2 }} style={{ marginTop: 12 }} />
                </Card>
              </Col>
            ))}
          </Row>
        </div>
      ) : (
        <Spin spinning={isLoading}>
          <div className={styles.section}>
            <div className={styles.sectionHeader}>
              <Space>
                <FireOutlined className={styles.sectionIconMuted} />
                <Title level={4} className={styles.sectionTitle}>
                  精选推荐
                </Title>
              </Space>
              <Text
                type="secondary"
                className={styles.viewAll}
                onClick={() => navigate('/catalog')}
              >
                查看全部 <RightOutlined />
              </Text>
            </div>
            {hasFeed ? (
              <Row gutter={[16, 16]}>
                {recommendedProducts.map((product) => (
                  <Col xs={12} sm={8} md={6} lg={4} key={product.id}>
                    {renderProductCard(product)}
                  </Col>
                ))}
              </Row>
            ) : (
              <Text type="secondary">暂无推荐，请前往卖场浏览全部 SKU。</Text>
            )}
          </div>
        </Spin>
      )}
    </div>
  );
};

export default HomePage;
