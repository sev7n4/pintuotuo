import React, { useEffect, useState } from 'react'
import {
  Table,
  Input,
  Button,
  Space,
  Tag,
  Empty,
  Spin,
  Pagination,
  Row,
  Col,
  Select,
  Slider,
  Dropdown,
} from 'antd'
import { SearchOutlined, PlusOutlined, FilterOutlined, SortAscendingOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useProductStore } from '@stores/productStore'
import { useAuthStore } from '@stores/authStore'
import type { Product } from '@/types'

const { Option } = Select

type SortField = 'price' | 'stock' | 'created_at'
type SortOrder = 'asc' | 'desc'

interface ProductFilters {
  minPrice?: number
  maxPrice?: number
  category?: string
  status?: string
}

export const ProductListPage: React.FC = () => {
  const navigate = useNavigate()
  const {
    products,
    total,
    filters,
    isLoading,
    error,
    fetchProducts,
    setFilters,
    searchProducts,
  } = useProductStore()
  const { user } = useAuthStore()

  const [sortField, setSortField] = useState<SortField>('created_at')
  const [sortOrder, setSortOrder] = useState<SortOrder>('desc')
  const [productFilters, setProductFilters] = useState<ProductFilters>({})
  const [priceRange, setPriceRange] = useState<[number, number]>([0, 10000])

  useEffect(() => {
    fetchProducts()
  }, [])

  const columns = [
    {
      title: '产品名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Product) => (
        <a onClick={() => navigate(`/products/${record.id}`)}>{text}</a>
      ),
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      width: 200,
      ellipsis: true,
    },
    {
      title: '价格',
      dataIndex: 'price',
      key: 'price',
      render: (price: number) => `¥${price.toFixed(2)}`,
      width: 100,
    },
    {
      title: '库存',
      dataIndex: 'stock',
      key: 'stock',
      width: 80,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'active' ? 'green' : 'red'}>
          {status === 'active' ? '上架' : '下架'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Product) => (
        <Space>
          <Button type="link" onClick={() => navigate(`/products/${record.id}`)}>
            详情
          </Button>
          <Button type="link" onClick={() => navigate(`/products/${record.id}/cart`)}>
            加购
          </Button>
        </Space>
      ),
    },
  ]

  const handleSearch = (value: string) => {
    if (value.trim()) {
      searchProducts(value)
    } else {
      fetchProducts()
    }
  }

  const handlePageChange = (page: number, pageSize: number) => {
    setFilters({ page, per_page: pageSize })
    fetchProducts({ page, per_page: pageSize })
  }

  const handleSort = (field: SortField) => {
    const newOrder = sortField === field && sortOrder === 'asc' ? 'desc' : 'asc'
    setSortField(field)
    setSortOrder(newOrder)
    
    fetchProducts({
      ...filters,
      sort_field: field,
      sort_order: newOrder,
    } as any)
  }

  const handleFilterChange = (key: keyof ProductFilters, value: any) => {
    setProductFilters(prev => ({ ...prev, [key]: value }))
  }

  const applyFilters = () => {
    fetchProducts({
      ...filters,
      ...productFilters,
      sort_field: sortField,
      sort_order: sortOrder,
    } as any)
  }

  const resetFilters = () => {
    setProductFilters({})
    setPriceRange([0, 10000])
    setSortField('created_at')
    setSortOrder('desc')
    fetchProducts()
  }

  const filterDropdownItems = [
    {
      key: 'price',
      label: (
        <div style={{ padding: '8px', width: 250 }}>
          <div style={{ marginBottom: 8 }}>价格区间</div>
          <Slider
            range
            min={0}
            max={10000}
            value={priceRange}
            onChange={(value) => setPriceRange(value as [number, number])}
            onAfterChange={(value) => {
              handleFilterChange('minPrice', value[0])
              handleFilterChange('maxPrice', value[1])
            }}
          />
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <span>¥{priceRange[0]}</span>
            <span>¥{priceRange[1]}</span>
          </div>
        </div>
      ),
    },
    {
      key: 'status',
      label: (
        <div style={{ padding: '8px' }}>
          <div style={{ marginBottom: 8 }}>状态筛选</div>
          <Select
            style={{ width: '100%' }}
            placeholder="选择状态"
            allowClear
            onChange={(value) => handleFilterChange('status', value)}
          >
            <Option value="active">上架</Option>
            <Option value="inactive">下架</Option>
          </Select>
        </div>
      ),
    },
    {
      key: 'actions',
      label: (
        <Space style={{ padding: '8px' }}>
          <Button type="primary" size="small" onClick={applyFilters}>
            应用筛选
          </Button>
          <Button size="small" onClick={resetFilters}>
            重置
          </Button>
        </Space>
      ),
    },
  ]

  const sortDropdownItems = [
    {
      key: 'price_asc',
      label: '价格从低到高',
      onClick: () => handleSort('price'),
    },
    {
      key: 'price_desc',
      label: '价格从高到低',
      onClick: () => { setSortField('price'); setSortOrder('desc'); },
    },
    {
      key: 'stock_asc',
      label: '库存从少到多',
      onClick: () => handleSort('stock'),
    },
    {
      key: 'stock_desc',
      label: '库存从多到少',
      onClick: () => { setSortField('stock'); setSortOrder('desc'); },
    },
    {
      key: 'created_desc',
      label: '最新发布',
      onClick: () => { setSortField('created_at'); setSortOrder('desc'); },
    },
  ]

  if (error) {
    return <Empty description={`错误: ${error}`} />
  }

  return (
    <div style={{ padding: '20px' }}>
      <Row gutter={16} style={{ marginBottom: '20px' }}>
        <Col flex="auto">
          <Input.Search
            placeholder="搜索产品..."
            prefix={<SearchOutlined />}
            onSearch={handleSearch}
            style={{ width: '100%' }}
          />
        </Col>
        <Col>
          <Space>
            <Dropdown menu={{ items: filterDropdownItems }} trigger={['click']}>
              <Button icon={<FilterOutlined />}>
                筛选
              </Button>
            </Dropdown>
            <Dropdown menu={{ items: sortDropdownItems }} trigger={['click']}>
              <Button icon={<SortAscendingOutlined />}>
                排序
              </Button>
            </Dropdown>
            {user?.role === 'merchant' && (
              <Button type="primary" icon={<PlusOutlined />}>
                发布产品
              </Button>
            )}
          </Space>
        </Col>
      </Row>

      <Spin spinning={isLoading}>
        <Table
          columns={columns}
          dataSource={products}
          rowKey="id"
          pagination={false}
          locale={{ emptyText: '暂无数据' }}
        />
      </Spin>

      {total > 0 && (
        <Pagination
          current={filters.page}
          pageSize={filters.per_page}
          total={total}
          onChange={handlePageChange}
          style={{ marginTop: '20px', textAlign: 'right' }}
        />
      )}
    </div>
  )
}

export default ProductListPage
