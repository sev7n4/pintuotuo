import React, { useEffect } from 'react'
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
} from 'antd'
import { SearchOutlined, PlusOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useProductStore } from '@stores/productStore'
import type { Product } from '@/types'

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
          <Button type="primary" icon={<PlusOutlined />}>
            发布产品
          </Button>
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
