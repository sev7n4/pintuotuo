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
} from 'antd'
import { ShoppingCartOutlined, ArrowLeftOutlined } from '@ant-design/icons'
import { useNavigate, useParams } from 'react-router-dom'
import { useProductStore } from '@stores/productStore'
import { useCartStore } from '@stores/cartStore'
import type { Product } from '@types/index'

export const ProductDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { fetchProductByID, isLoading, error } = useProductStore()
  const { addItem } = useCartStore()
  const [product, setProduct] = useState<Product | null>(null)
  const [quantity, setQuantity] = useState(1)

  useEffect(() => {
    if (id) {
      loadProduct()
    }
  }, [id])

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

  if (error) {
    return <Empty description={`错误: ${error}`} />
  }

  if (isLoading || !product) {
    return <Spin />
  }

  return (
    <div style={{ padding: '20px', maxWidth: 800, margin: '0 auto' }}>
      <Button
        type="text"
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate('/products')}
        style={{ marginBottom: '20px' }}
      >
        返回列表
      </Button>

      <Card>
        <div style={{ marginBottom: '20px' }}>
          <h1>{product.name}</h1>
          <p style={{ color: '#666' }}>{product.description}</p>
        </div>

        <Divider />

        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div>
            <Statistic title="价格" value={product.price} prefix="¥" />
          </div>

          <div>
            <Statistic title="库存" value={product.stock} suffix="件" />
          </div>

          <div>
            <span>购买数量: </span>
            <InputNumber
              min={1}
              max={product.stock}
              value={quantity}
              onChange={(val) => setQuantity(val || 1)}
              style={{ marginLeft: '10px', width: 100 }}
            />
          </div>

          <Button
            type="primary"
            size="large"
            icon={<ShoppingCartOutlined />}
            onClick={handleAddToCart}
            disabled={product.stock === 0}
            block
          >
            {product.stock === 0 ? '暂无库存' : '加入购物车'}
          </Button>
        </Space>
      </Card>
    </div>
  )
}

export default ProductDetailPage
