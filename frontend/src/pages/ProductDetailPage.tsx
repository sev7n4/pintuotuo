import React, { useEffect, useState } from 'react'
import {
  Card,
  Button,
  Space,
  Statistic,
  Divider,
  InputNumber,
  message,
  Spin,
  Empty,
  Tabs,
  Tag,
  Rate,
  List,
  Avatar,
  Typography,
  Row,
  Col,
} from 'antd'
import {
  ShoppingCartOutlined,
  ArrowLeftOutlined,
  ShareAltOutlined,
  TeamOutlined,
  UserOutlined,
} from '@ant-design/icons'
import { useNavigate, useParams } from 'react-router-dom'
import { useProductStore } from '@stores/productStore'
import { useCartStore } from '@stores/cartStore'
import { useGroupStore } from '@stores/groupStore'
import type { Product, GroupPrice, ProductReview } from '@/types'

const { Title, Text } = Typography
const { TabPane } = Tabs

const mockReviews: ProductReview[] = [
  {
    id: 1,
    user_id: 101,
    user_name: '张三',
    rating: 5,
    content: '非常便宜，支持拼团！Token到账很快，模型效果很好。',
    created_at: '2026-03-13',
  },
  {
    id: 2,
    user_id: 102,
    user_name: '李四',
    rating: 4,
    content: '性价比很高，拼团省了不少钱。客服响应也很快。',
    created_at: '2026-03-12',
  },
  {
    id: 3,
    user_id: 103,
    user_name: '王五',
    rating: 5,
    content: '第二次购买了，一直很稳定，推荐！',
    created_at: '2026-03-10',
  },
]

const defaultGroupPrices: GroupPrice[] = [
  { min_members: 2, price_per_person: 60, discount_percent: 40 },
  { min_members: 5, price_per_person: 50, discount_percent: 50 },
]

export const ProductDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { fetchProductByID, isLoading, error } = useProductStore()
  const { addItem } = useCartStore()
  const { createGroup } = useGroupStore()
  const [product, setProduct] = useState<Product | null>(null)
  const [quantity, setQuantity] = useState(1)
  const [purchaseMode, setPurchaseMode] = useState<'single' | 'group'>('group')
  const [selectedGroupPrice, setSelectedGroupPrice] = useState<GroupPrice | null>(null)

  useEffect(() => {
    if (id) {
      loadProduct()
    }
  }, [id])

  useEffect(() => {
    if (product?.group_prices?.length) {
      setSelectedGroupPrice(product.group_prices[0])
    }
  }, [product])

  const loadProduct = async () => {
    if (!id) return
    const result = await fetchProductByID(parseInt(id))
    if (result) {
      setProduct(result)
    }
  }

  const handleAddToCart = () => {
    if (!product) return
    addItem(product, quantity)
    message.success(`已添加 ${quantity} 件到购物车`)
    setTimeout(() => navigate('/cart'), 1000)
  }

  const handleGroupPurchase = async () => {
    if (!product) {
      message.error('商品信息加载失败')
      return
    }
    const currentGroupPrice = selectedGroupPrice || groupPrices[0]
    if (!currentGroupPrice) {
      message.error('请选择拼团规则')
      return
    }
    try {
      const deadline = new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString()
      const group = await createGroup(product.id, currentGroupPrice.min_members, deadline)
      if (group) {
        message.success('拼团已创建，快去邀请好友吧！')
        navigate(`/groups/${group.id}`)
      } else {
        message.error('创建拼团失败，请重试')
      }
    } catch {
      message.error('创建拼团失败，请重试')
    }
  }

  const handleShare = () => {
    if (!product) return
    const shareUrl = `${window.location.origin}/products/${product.id}`
    if (navigator.clipboard && window.isSecureContext) {
      navigator.clipboard.writeText(shareUrl)
        .then(() => {
          message.success('链接已复制到剪贴板')
        })
        .catch(() => {
          message.error('复制失败，请手动复制链接')
        })
    } else {
      // 降级方案：创建一个临时输入框来复制
      const textArea = document.createElement('textarea')
      textArea.value = shareUrl
      textArea.style.position = 'fixed'
      textArea.style.left = '-999999px'
      textArea.style.top = '-999999px'
      document.body.appendChild(textArea)
      textArea.focus()
      textArea.select()
      try {
        document.execCommand('copy')
        message.success('链接已复制到剪贴板')
      } catch {
        message.error('复制失败，请手动复制链接')
      } finally {
        document.body.removeChild(textArea)
      }
    }
  }

  const calculateDiscount = () => {
    if (!product || !selectedGroupPrice) return 0
    return Math.round((1 - selectedGroupPrice.price_per_person / product.price) * 100)
  }

  if (error) {
    return <Empty description={`错误: ${error}`} />
  }

  if (isLoading || !product) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}>
        <Spin size="large" />
      </div>
    )
  }

  const groupPrices = product.group_prices || defaultGroupPrices
  const rating = product.rating || 4.8
  const reviewCount = product.review_count || 1000

  return (
    <div style={{ padding: '20px', maxWidth: 900, margin: '0 auto' }}>
      <Space style={{ marginBottom: '20px' }}>
        <Button type="text" icon={<ArrowLeftOutlined />} onClick={() => navigate('/products')}>
          返回列表
        </Button>
      </Space>

      <Card>
        <Row gutter={[24, 24]}>
          <Col xs={24} md={12}>
            <div style={{ 
              background: '#f5f5f5', 
              borderRadius: 8, 
              height: 300, 
              display: 'flex', 
              alignItems: 'center', 
              justifyContent: 'center',
              marginBottom: 16,
            }}>
              <Text type="secondary" style={{ fontSize: 48 }}>📦</Text>
            </div>
          </Col>

          <Col xs={24} md={12}>
            <Title level={3}>{product.name}</Title>
            <Space style={{ marginBottom: 16 }}>
              <Tag color="blue">已拼 {(product.sold_count || 100000).toLocaleString()} 件</Tag>
              <Tag color="gold">⭐ {rating}/5.0 ({reviewCount}+ 评)</Tag>
            </Space>

            <div style={{ marginBottom: 16 }}>
              <Text type="secondary">{product.description}</Text>
            </div>

            <Divider style={{ margin: '12px 0' }} />

            <div style={{ marginBottom: 16 }}>
              <Space direction="vertical" size="small">
                {product.token_count && (
                  <Text>包含：<Text strong>{(product.token_count / 10000).toFixed(0)}万 Token</Text></Text>
                )}
                {product.models && product.models.length > 0 && (
                  <Text>模型：<Text strong>{product.models.join(', ')}</Text></Text>
                )}
                {product.validity_period && (
                  <Text>有效期：<Text strong>{product.validity_period}</Text></Text>
                )}
                {product.context_length && (
                  <Text>上下文：<Text strong>{product.context_length}</Text></Text>
                )}
              </Space>
            </div>
          </Col>
        </Row>

        <Divider />

        <div style={{ marginBottom: 24 }}>
          <Title level={4}>定价信息</Title>
          <Row gutter={[16, 16]}>
            <Col xs={24} sm={12}>
              <Card 
                size="small" 
                hoverable
                style={{ 
                  border: purchaseMode === 'single' ? '2px solid #1890ff' : '1px solid #d9d9d9',
                  cursor: 'pointer',
                }}
                onClick={() => setPurchaseMode('single')}
              >
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Text type="secondary">单独购买</Text>
                  <Statistic 
                    value={product.price} 
                    prefix="¥" 
                    valueStyle={{ color: '#333', fontSize: 24 }}
                  />
                  <Text type="secondary">原价购买，立即发货</Text>
                </Space>
              </Card>
            </Col>

            <Col xs={24} sm={12}>
              <Card 
                size="small" 
                hoverable
                style={{ 
                  border: purchaseMode === 'group' ? '2px solid #52c41a' : '1px solid #d9d9d9',
                  cursor: 'pointer',
                  background: '#f6ffed',
                }}
                onClick={() => setPurchaseMode('group')}
              >
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Space>
                    <Text type="secondary">拼团购买</Text>
                    <Tag color="green">推荐</Tag>
                  </Space>
                  <Statistic 
                    value={selectedGroupPrice?.price_per_person || groupPrices[0].price_per_person} 
                    prefix="¥" 
                    valueStyle={{ color: '#52c41a', fontSize: 24 }}
                    suffix={<Text type="secondary" style={{ fontSize: 14 }}>/人</Text>}
                  />
                  <Text type="success">节省 {calculateDiscount()}%</Text>
                </Space>
              </Card>
            </Col>
          </Row>
        </div>

        {purchaseMode === 'group' && (
          <div style={{ marginBottom: 24 }}>
            <Title level={5}>选择拼团规则</Title>
            <Space direction="vertical" style={{ width: '100%' }}>
              {groupPrices.map((gp) => (
                <Card 
                  key={gp.min_members}
                  size="small"
                  hoverable
                  style={{ 
                    border: selectedGroupPrice?.min_members === gp.min_members ? '2px solid #52c41a' : '1px solid #d9d9d9',
                    cursor: 'pointer',
                  }}
                  onClick={() => setSelectedGroupPrice(gp)}
                >
                  <Row justify="space-between" align="middle">
                    <Col>
                      <Space>
                        <TeamOutlined />
                        <Text strong>{gp.min_members}人团</Text>
                        <Tag color="green">省{gp.discount_percent}%</Tag>
                      </Space>
                    </Col>
                    <Col>
                      <Space>
                        <Text delete type="secondary">¥{product.price}</Text>
                        <Text strong style={{ color: '#52c41a', fontSize: 18 }}>¥{gp.price_per_person}/人</Text>
                      </Space>
                    </Col>
                  </Row>
                </Card>
              ))}
            </Space>
          </div>
        )}

        <Divider />

        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div>
            <span>购买数量: </span>
            <InputNumber
              min={1}
              max={product.stock}
              value={quantity}
              onChange={(val) => setQuantity(val || 1)}
              style={{ marginLeft: '10px', width: 100 }}
            />
            <Text type="secondary" style={{ marginLeft: 16 }}>库存: {product.stock} 件</Text>
          </div>

          <Space style={{ width: '100%' }} size="middle">
            {purchaseMode === 'single' ? (
              <Button
                type="primary"
                size="large"
                icon={<ShoppingCartOutlined />}
                onClick={handleAddToCart}
                disabled={product.stock === 0}
                style={{ flex: 1 }}
              >
                {product.stock === 0 ? '暂无库存' : '加入购物车'}
              </Button>
            ) : (
              <Button
                type="primary"
                size="large"
                icon={<TeamOutlined />}
                onClick={handleGroupPurchase}
                disabled={product.stock === 0}
                style={{ flex: 1, background: '#52c41a', borderColor: '#52c41a' }}
              >
                {product.stock === 0 ? '暂无库存' : `立即拼团 (¥${selectedGroupPrice?.price_per_person || groupPrices[0].price_per_person}/人)`}
              </Button>
            )}
            <Button
              size="large"
              icon={<ShareAltOutlined />}
              onClick={handleShare}
            >
              分享
            </Button>
          </Space>
        </Space>
      </Card>

      <Card style={{ marginTop: 16 }}>
        <Tabs defaultActiveKey="detail">
          <TabPane tab="商品详情" key="detail">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Title level={5}>商品信息</Title>
              <ul style={{ paddingLeft: 20 }}>
                <li>包含Token数量和类型：{(product.token_count || 1000000).toLocaleString()} Token</li>
                <li>支持模型：{product.models?.join('、') || 'GLM-5, K2.5'}</li>
                <li>有效期：{product.validity_period || '1年'}</li>
                <li>上下文长度：{product.context_length || '128K'}</li>
              </ul>

              <Divider />

              <Title level={5}>使用指南</Title>
              <ol style={{ paddingLeft: 20 }}>
                <li>购买成功后，Token将自动充值到您的账户</li>
                <li>在"我的Token"页面可以查看余额和使用记录</li>
                <li>支持通过API调用使用Token</li>
                <li>可在"API密钥管理"创建和管理API密钥</li>
              </ol>

              <Divider />

              <Title level={5}>常见问题</Title>
              <Space direction="vertical" style={{ width: '100%' }}>
                <div>
                  <Text strong>Q: Token有效期多久？</Text>
                  <br />
                  <Text type="secondary">A: Token有效期为1年，从购买之日起计算。</Text>
                </div>
                <div>
                  <Text strong>Q: 拼团失败怎么办？</Text>
                  <br />
                  <Text type="secondary">A: 拼团失败后，系统会自动退款，您可以选择重新发起拼团或单独购买。</Text>
                </div>
                <div>
                  <Text strong>Q: Token可以转让吗？</Text>
                  <br />
                  <Text type="secondary">A: Token暂不支持转让，但可以通过API为其他项目提供服务。</Text>
                </div>
              </Space>
            </Space>
          </TabPane>

          <TabPane tab={`用户评价 (${reviewCount})`} key="reviews">
            <Space direction="vertical" style={{ width: '100%' }} size="large">
              <Space>
                <Statistic title="综合评分" value={rating} suffix="/ 5.0" />
                <Rate disabled defaultValue={rating} allowHalf />
              </Space>

              <Divider />

              <List
                itemLayout="horizontal"
                dataSource={mockReviews}
                renderItem={(review) => (
                  <List.Item>
                    <List.Item.Meta
                      avatar={<Avatar icon={<UserOutlined />} />}
                      title={
                        <Space>
                          <Text strong>{review.user_name}</Text>
                          <Rate disabled defaultValue={review.rating} style={{ fontSize: 12 }} />
                        </Space>
                      }
                      description={
                        <Space direction="vertical" size="small">
                          <Text>{review.content}</Text>
                          <Text type="secondary" style={{ fontSize: 12 }}>{review.created_at}</Text>
                        </Space>
                      }
                    />
                  </List.Item>
                )}
              />
            </Space>
          </TabPane>
        </Tabs>
      </Card>
    </div>
  )
}

export default ProductDetailPage
