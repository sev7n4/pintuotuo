import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Carousel, Card, Row, Col, Spin, Tag, Input, Typography, Space } from 'antd'
import { FireOutlined, ClockCircleOutlined, SearchOutlined, RightOutlined } from '@ant-design/icons'
import { useHomeStore } from '@/stores/homeStore'
import { Product } from '@/types'
import styles from './HomePage.module.css'

const { Title, Text } = Typography
const { Search } = Input

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

  useEffect(() => {
    fetchHomeData()
  }, [fetchHomeData])

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

  const formatPrice = (price: number) => {
    return `¥${price.toFixed(2)}`
  }

  const renderProductCard = (product: Product) => {
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
    viewAllLink?: string
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
            {renderProductCard(product)}
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
          placeholder="搜索商品"
          allowClear
          enterButton={<SearchOutlined />}
          size="large"
          onSearch={handleSearch}
          className={styles.searchInput}
        />
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
          '新品上架',
          <ClockCircleOutlined style={{ color: '#1890ff' }} />,
          newProducts,
          '/products?sort=new'
        )}
      </Spin>
    </div>
  )
}

export default HomePage
