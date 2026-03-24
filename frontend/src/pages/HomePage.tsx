import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Carousel, Card, Row, Col, Spin, Tag, Input, Typography, Space } from 'antd'
import { FireOutlined, ClockCircleOutlined, SearchOutlined, RightOutlined, ThunderboltOutlined, GiftOutlined, StarOutlined } from '@ant-design/icons'
import { useHomeStore } from '@/stores/homeStore'
import { Product } from '@/types'
import styles from './HomePage.module.css'

const { Title, Text } = Typography
const { Search } = Input

interface QuickNav {
  key: string
  name: string
  icon: React.ReactNode
  color: string
  link: string
}

const quickNavItems: QuickNav[] = [
  { key: 'hot', name: '热销爆款', icon: <FireOutlined />, color: '#ff4d4f', link: '/products?sort=hot' },
  { key: 'group', name: '超值拼团', icon: <GiftOutlined />, color: '#52c41a', link: '/groups' },
  { key: 'flash', name: '限时秒杀', icon: <ThunderboltOutlined />, color: '#faad14', link: '/products?flash=true' },
  { key: 'new', name: '新品上架', icon: <ClockCircleOutlined />, color: '#1890ff', link: '/products?sort=new' },
]

const HomePage = () => {
  const navigate = useNavigate()
  const { 
    banners, 
    hotProducts, 
    newProducts, 
    categories, 
    isLoading, 
    error,
    fetchHomeData 
  } = useHomeStore()

  const [recommendedProducts, setRecommendedProducts] = useState<Product[]>([])

  useEffect(() => {
    fetchHomeData()
  }, [fetchHomeData])

  useEffect(() => {
    if (hotProducts.length > 0 && newProducts.length > 0) {
      const mixed = [...hotProducts.slice(0, 2), ...newProducts.slice(0, 2)]
      setRecommendedProducts(mixed)
    }
  }, [hotProducts, newProducts])

  const handleSearch = (value: string) => {
    if (value.trim()) {
      navigate(`/products?search=${encodeURIComponent(value.trim())}`)
    }
  }

  const handleProductClick = (productId: number) => {
    navigate(`/products/${productId}`)
  }

  const handleCategoryClick = (category: string) => {
    navigate(`/products?category=${encodeURIComponent(category)}`)
  }

  const handleQuickNavClick = (link: string) => {
    navigate(link)
  }

  const formatPrice = (price: number) => {
    return `¥${price.toFixed(2)}`
  }

  const renderProductCard = (product: Product, showGroupTag = false) => {
    const discount = product.original_price && product.original_price > product.price
      ? Math.round((1 - product.price / product.original_price) * 100)
      : 0

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
    )
  }

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
          <Title level={4} className={styles.sectionTitle}>{title}</Title>
        </Space>
        {viewAllLink && (
          <Text 
            type="secondary" 
            className={styles.viewAll}
            onClick={() => navigate(viewAllLink)}
          >
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
  )

  if (error) {
    return (
      <div className={styles.errorContainer}>
        <Text type="danger">{error}</Text>
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <div className={styles.searchSection}>
        <Search
          placeholder="搜索模型或关键词"
          allowClear
          enterButton={<SearchOutlined />}
          size="large"
          onSearch={handleSearch}
          className={styles.searchInput}
        />
      </div>

      <div className={styles.quickNavSection}>
        <Row gutter={[12, 12]}>
          {quickNavItems.map((item) => (
            <Col span={6} key={item.key}>
              <div 
                className={styles.quickNavItem}
                onClick={() => handleQuickNavClick(item.link)}
              >
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
              <div 
                className={styles.bannerContent}
                onClick={() => navigate(banner.link)}
              >
                <Text className={styles.bannerTitle}>{banner.title}</Text>
              </div>
            </div>
          ))}
        </Carousel>
      )}

      {categories.length > 0 && (
        <div className={styles.categorySection}>
          <div className={styles.categoryHeader}>
            <Title level={5} className={styles.categoryTitle}>商品分类</Title>
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
          '/products?sort=hot'
        )}

        {renderSection(
          '超值拼团',
          <GiftOutlined style={{ color: '#52c41a' }} />,
          hotProducts.slice(0, 4),
          '/groups',
          true
        )}

        {renderSection(
          '新品上架',
          <ClockCircleOutlined style={{ color: '#1890ff' }} />,
          newProducts,
          '/products?sort=new'
        )}

        {recommendedProducts.length > 0 && (
          <div className={styles.recommendedSection}>
            <div className={styles.sectionHeader}>
              <Space>
                <StarOutlined style={{ color: '#faad14' }} />
                <Title level={4} className={styles.sectionTitle}>猜你喜欢</Title>
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
  )
}

export default HomePage
