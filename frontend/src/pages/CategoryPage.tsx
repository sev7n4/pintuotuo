import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Card, Row, Col, Input, Tag, List, Empty, Spin } from 'antd'
import { SearchOutlined, AppstoreOutlined, CodeOutlined, EyeOutlined, ThunderboltOutlined, ApiOutlined } from '@ant-design/icons'
import { productService } from '@/services/product'
import styles from './CategoryPage.module.css'

interface Category {
  id: string
  name: string
  icon: React.ReactNode
  color: string
  description: string
  count: number
}

const CategoryPage = () => {
  const [searchText, setSearchText] = useState('')
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [products, setProducts] = useState<any[]>([])

  const categories: Category[] = [
    {
      id: 'llm',
      name: '大模型',
      icon: <AppstoreOutlined />,
      color: '#1890ff',
      description: 'GPT、Claude、Gemini等大语言模型',
      count: 15,
    },
    {
      id: 'code',
      name: '编码模型',
      icon: <CodeOutlined />,
      color: '#52c41a',
      description: '代码生成、代码补全模型',
      count: 8,
    },
    {
      id: 'vision',
      name: '视觉模型',
      icon: <EyeOutlined />,
      color: '#722ed1',
      description: '图像生成、图像识别模型',
      count: 12,
    },
    {
      id: 'audio',
      name: '音频模型',
      icon: <ThunderboltOutlined />,
      color: '#faad14',
      description: '语音合成、语音识别模型',
      count: 6,
    },
    {
      id: 'embedding',
      name: '嵌入模型',
      icon: <ApiOutlined />,
      color: '#eb2f96',
      description: '文本嵌入、向量检索模型',
      count: 10,
    },
  ]

  useEffect(() => {
    const fetchProducts = async () => {
      setLoading(true)
      try {
        const response = await productService.listProducts({
          category: selectedCategory || undefined,
        })
        setProducts(response.data?.data?.data || [])
      } catch {
        setProducts([])
      } finally {
        setLoading(false)
      }
    }
    fetchProducts()
  }, [selectedCategory])

  const filteredProducts = products.filter((product: any) => {
    const matchesSearch = !searchText || 
      product.name?.toLowerCase().includes(searchText.toLowerCase()) ||
      product.description?.toLowerCase().includes(searchText.toLowerCase())
    return matchesSearch
  })

  return (
    <div className={styles.categoryPage}>
      <div className={styles.header}>
        <h1 className={styles.title}>商品分类</h1>
        <Input
          placeholder="搜索商品..."
          prefix={<SearchOutlined />}
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          className={styles.searchInput}
          allowClear
        />
      </div>

      <Row gutter={[16, 16]} className={styles.categoryGrid}>
        {categories.map((category) => (
          <Col xs={12} sm={8} md={6} lg={4} key={category.id}>
            <Card
              hoverable
              className={`${styles.categoryCard} ${selectedCategory === category.id ? styles.categoryCardActive : ''}`}
              onClick={() => setSelectedCategory(selectedCategory === category.id ? null : category.id)}
            >
              <div className={styles.categoryContent}>
                <div 
                  className={styles.categoryIcon}
                  style={{ backgroundColor: category.color + '20', color: category.color }}
                >
                  {category.icon}
                </div>
                <h3 className={styles.categoryName}>{category.name}</h3>
                <p className={styles.categoryDesc}>{category.description}</p>
                <Tag color={category.color}>{category.count} 款</Tag>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      <Card className={styles.productSection} title={selectedCategory ? `${categories.find(c => c.id === selectedCategory)?.name || ''}商品` : '全部商品'}>
        {loading ? (
          <div className={styles.loading}>
            <Spin />
          </div>
        ) : filteredProducts.length > 0 ? (
          <List
            grid={{ gutter: 16, xs: 1, sm: 2, md: 3, lg: 4, xl: 4, xxl: 6 }}
            dataSource={filteredProducts}
            renderItem={(product: any) => (
              <List.Item>
                <Link to={`/products/${product.id}`}>
                  <Card hoverable className={styles.productCard}>
                    <div className={styles.productImage}>
                      {product.image_url ? (
                        <img src={product.image_url} alt={product.name} />
                      ) : (
                        <div className={styles.productPlaceholder}>
                          <AppstoreOutlined />
                        </div>
                      )}
                    </div>
                    <div className={styles.productInfo}>
                      <h4 className={styles.productName}>{product.name}</h4>
                      <p className={styles.productPrice}>¥{product.price?.toFixed(2) || '0.00'}</p>
                    </div>
                  </Card>
                </Link>
              </List.Item>
            )}
          />
        ) : (
          <Empty description="暂无商品" />
        )}
      </Card>
    </div>
  )
}

export default CategoryPage
