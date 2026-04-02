import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
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
  Drawer,
  Form,
  Select,
  InputNumber,
  Switch,
} from 'antd';
import {
  FireOutlined,
  ClockCircleOutlined,
  SearchOutlined,
  RightOutlined,
  ThunderboltOutlined,
  GiftOutlined,
  StarOutlined,
  FilterOutlined,
} from '@ant-design/icons';
import { useHomeStore } from '@/stores/homeStore';
import { Product } from '@/types';
import styles from './HomePage.module.css';

const { Title, Text } = Typography;
const { Search } = Input;

interface QuickNav {
  key: string;
  name: string;
  icon: React.ReactNode;
  color: string;
  link: string;
}

const quickNavItems: QuickNav[] = [
  {
    key: 'hot',
    name: '热销爆款',
    icon: <FireOutlined />,
    color: '#ff4d4f',
    link: '/catalog?sort=hot',
  },
  {
    key: 'group',
    name: '超值拼团',
    icon: <GiftOutlined />,
    color: '#52c41a',
    link: '/catalog?group_enabled=true',
  },
  {
    key: 'flash',
    name: '限时秒杀',
    icon: <ThunderboltOutlined />,
    color: '#faad14',
    link: '/catalog?flash=true',
  },
  {
    key: 'new',
    name: '新品上架',
    icon: <ClockCircleOutlined />,
    color: '#1890ff',
    link: '/catalog?sort=new',
  },
];

const HomePage = () => {
  const navigate = useNavigate();
  const { banners, hotProducts, newProducts, categories, isLoading, error, fetchHomeData } =
    useHomeStore();

  const [recommendedProducts, setRecommendedProducts] = useState<Product[]>([]);
  const [filterDrawerOpen, setFilterDrawerOpen] = useState(false);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [filterForm] = Form.useForm();

  useEffect(() => {
    fetchHomeData();
  }, [fetchHomeData]);

  useEffect(() => {
    if (hotProducts.length > 0 && newProducts.length > 0) {
      const mixed = [...hotProducts.slice(0, 2), ...newProducts.slice(0, 2)];
      setRecommendedProducts(mixed);
    }
  }, [hotProducts, newProducts]);

  useEffect(() => {
    if (filterDrawerOpen) {
      filterForm.setFieldsValue({ q: searchKeyword });
    }
  }, [filterDrawerOpen, filterForm, searchKeyword]);

  const handleSearch = (value: string) => {
    const q = value.trim();
    if (q) {
      navigate(`/catalog?q=${encodeURIComponent(q)}`);
    }
  };

  const applyCatalogFilters = (values: Record<string, unknown>) => {
    const params = new URLSearchParams();
    const q = String(values.q ?? searchKeyword ?? '').trim();
    if (q) params.set('q', q);
    if (values.category) params.set('category', String(values.category));
    if (values.model_name) params.set('model_name', String(values.model_name).trim());
    if (values.provider) params.set('provider', String(values.provider));
    if (values.tier) params.set('tier', String(values.tier));
    if (values.sku_type) params.set('type', String(values.sku_type));
    if (values.group_enabled) params.set('group_enabled', 'true');
    if (values.price_min != null && values.price_min !== '')
      params.set('price_min', String(values.price_min));
    if (values.price_max != null && values.price_max !== '')
      params.set('price_max', String(values.price_max));
    if (values.valid_days_min != null && values.valid_days_min !== '')
      params.set('valid_days_min', String(values.valid_days_min));
    if (values.valid_days_max != null && values.valid_days_max !== '')
      params.set('valid_days_max', String(values.valid_days_max));
    if (values.sort) params.set('sort', String(values.sort));
    const qs = params.toString();
    navigate(qs ? `/catalog?${qs}` : '/catalog');
    setFilterDrawerOpen(false);
  };

  const handleProductClick = (productId: number) => {
    navigate(`/catalog/${productId}`);
  };

  const handleCategoryClick = (category: string) => {
    navigate(`/catalog?category=${encodeURIComponent(category)}`);
  };

  const handleQuickNavClick = (link: string) => {
    navigate(link);
  };

  const formatPrice = (price: number) => {
    return `¥${price.toFixed(2)}`;
  };

  const renderProductCard = (product: Product, showGroupTag = false) => {
    const discount =
      product.original_price && product.original_price > product.price
        ? Math.round((1 - product.price / product.original_price) * 100)
        : 0;

    return (
      <Card
        hoverable
        className={styles.productCard}
        cover={
          <div className={styles.productImage}>
            <div className={styles.productPlaceholder}>
              <Text type="secondary">{product.name.substring(0, 2)}</Text>
            </div>
            {discount > 0 && (
              <Tag color="#ff4d4f" className={styles.discountTag}>
                -{discount}%
              </Tag>
            )}
            {showGroupTag && (
              <Tag color="#52c41a" className={styles.groupTag}>
                拼团
              </Tag>
            )}
          </div>
        }
        onClick={() => handleProductClick(product.id)}
      >
        <div className={styles.productInfo}>
          <Text className={styles.productName} ellipsis>
            {product.name}
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
            <Text type="secondary">库存 {product.stock}</Text>
          </div>
        </div>
      </Card>
    );
  };

  const renderSection = (
    title: string,
    icon: React.ReactNode,
    products: Product[],
    viewAllLink?: string,
    showGroupTag = false
  ) => (
    <div className={styles.section}>
      <div className={styles.sectionHeader}>
        <Space>
          {icon}
          <Title level={4} className={styles.sectionTitle}>
            {title}
          </Title>
        </Space>
        {viewAllLink && (
          <Text type="secondary" className={styles.viewAll} onClick={() => navigate(viewAllLink)}>
            查看全部 <RightOutlined />
          </Text>
        )}
      </div>
      <Row gutter={[16, 16]}>
        {products.map((product) => (
          <Col xs={12} sm={8} md={6} lg={4} key={product.id}>
            {renderProductCard(product, showGroupTag)}
          </Col>
        ))}
      </Row>
    </div>
  );

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
            style={{ flex: 1 }}
          />
          <Button size="large" icon={<FilterOutlined />} onClick={() => setFilterDrawerOpen(true)}>
            筛选
          </Button>
        </Space.Compact>
      </div>

      <Drawer
        title="多维筛选"
        placement="right"
        width={360}
        open={filterDrawerOpen}
        onClose={() => setFilterDrawerOpen(false)}
        destroyOnClose
      >
        <Form
          form={filterForm}
          layout="vertical"
          onFinish={applyCatalogFilters}
          initialValues={{ group_enabled: false }}
        >
          <Form.Item name="q" label="关键词（可与上方搜索框一致）">
            <Input placeholder="模型名、套餐名、SKU 编码" allowClear />
          </Form.Item>
          <Form.Item name="category" label="品类 / SPU 名称">
            <Select
              allowClear
              placeholder="选择分类"
              options={categories.map((c) => ({ label: c.name, value: c.name }))}
            />
          </Form.Item>
          <Space style={{ width: '100%' }} size="middle">
            <Form.Item name="price_min" label="最低价">
              <InputNumber min={0} style={{ width: 140 }} placeholder="元" />
            </Form.Item>
            <Form.Item name="price_max" label="最高价">
              <InputNumber min={0} style={{ width: 140 }} placeholder="元" />
            </Form.Item>
          </Space>
          <Form.Item name="model_name" label="模型名称">
            <Input placeholder="如 GLM、DeepSeek" allowClear />
          </Form.Item>
          <Form.Item name="provider" label="厂商">
            <Select
              allowClear
              placeholder="厂商"
              options={[
                { label: 'OpenAI', value: 'openai' },
                { label: 'Anthropic', value: 'anthropic' },
                { label: 'Google', value: 'google' },
                { label: 'DeepSeek', value: 'deepseek' },
                { label: '智谱', value: 'zhipu' },
              ]}
            />
          </Form.Item>
          <Form.Item name="tier" label="模型层级">
            <Select
              allowClear
              options={[
                { label: 'Pro', value: 'pro' },
                { label: 'Lite', value: 'lite' },
                { label: 'Mini', value: 'mini' },
                { label: 'Vision', value: 'vision' },
              ]}
            />
          </Form.Item>
          <Form.Item name="sku_type" label="套餐类型">
            <Select
              allowClear
              options={[
                { label: 'Token包', value: 'token_pack' },
                { label: '订阅', value: 'subscription' },
                { label: '并发', value: 'concurrent' },
                { label: '试用', value: 'trial' },
              ]}
            />
          </Form.Item>
          <Form.Item name="group_enabled" label="仅看支持拼团" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Space style={{ width: '100%' }} size="middle">
            <Form.Item name="valid_days_min" label="有效期≥(天)">
              <InputNumber min={0} style={{ width: 140 }} />
            </Form.Item>
            <Form.Item name="valid_days_max" label="有效期≤(天)">
              <InputNumber min={0} style={{ width: 140 }} />
            </Form.Item>
          </Space>
          <Form.Item name="sort" label="排序">
            <Select
              allowClear
              placeholder="默认推荐"
              options={[
                { label: '热销', value: 'hot' },
                { label: '最新', value: 'new' },
                { label: '价格从低到高', value: 'price_asc' },
                { label: '价格从高到低', value: 'price_desc' },
              ]}
            />
          </Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">
              应用并搜索
            </Button>
            <Button
              onClick={() => {
                filterForm.resetFields();
                setSearchKeyword('');
              }}
            >
              重置
            </Button>
          </Space>
        </Form>
      </Drawer>

      <div className={styles.quickNavSection}>
        <Row gutter={[12, 12]}>
          {quickNavItems.map((item) => (
            <Col span={6} key={item.key}>
              <div className={styles.quickNavItem} onClick={() => handleQuickNavClick(item.link)}>
                <div className={styles.quickNavIcon} style={{ background: item.color }}>
                  {item.icon}
                </div>
                <Text className={styles.quickNavName}>{item.name}</Text>
              </div>
            </Col>
          ))}
        </Row>
      </div>

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

      {categories.length > 0 && (
        <div className={styles.categorySection}>
          <div className={styles.categoryHeader}>
            <Title level={5} className={styles.categoryTitle}>
              商品分类
            </Title>
          </div>
          <Row gutter={[12, 12]}>
            {categories.slice(0, 8).map((category) => (
              <Col span={6} key={category.name}>
                <div
                  className={styles.categoryItem}
                  onClick={() => handleCategoryClick(category.name)}
                >
                  <Text className={styles.categoryName}>{category.name}</Text>
                  <Text type="secondary" className={styles.categoryCount}>
                    {category.count}件
                  </Text>
                </div>
              </Col>
            ))}
          </Row>
        </div>
      )}

      <Spin spinning={isLoading}>
        {renderSection(
          '热门推荐',
          <FireOutlined style={{ color: '#ff4d4f' }} />,
          hotProducts,
          '/catalog?sort=hot'
        )}

        {renderSection(
          '超值拼团',
          <GiftOutlined style={{ color: '#52c41a' }} />,
          hotProducts.slice(0, 4),
          '/catalog?group_enabled=true',
          true
        )}

        {renderSection(
          '新品上架',
          <ClockCircleOutlined style={{ color: '#1890ff' }} />,
          newProducts,
          '/catalog?sort=new'
        )}

        {recommendedProducts.length > 0 && (
          <div className={styles.recommendedSection}>
            <div className={styles.sectionHeader}>
              <Space>
                <StarOutlined style={{ color: '#faad14' }} />
                <Title level={4} className={styles.sectionTitle}>
                  猜你喜欢
                </Title>
              </Space>
            </div>
            <Row gutter={[16, 16]}>
              {recommendedProducts.map((product) => (
                <Col xs={12} sm={8} md={6} lg={4} key={`rec-${product.id}`}>
                  {renderProductCard(product)}
                </Col>
              ))}
            </Row>
          </div>
        )}
      </Spin>
    </div>
  );
};

export default HomePage;
