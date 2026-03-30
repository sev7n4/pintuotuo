import React, { useEffect, useState } from 'react';
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
  Typography,
  Card,
  Grid,
  Progress,
  Statistic,
  Tabs,
} from 'antd';
import {
  SearchOutlined,
  PlusOutlined,
  FilterOutlined,
  SortAscendingOutlined,
  FireOutlined,
  ClockCircleOutlined,
  ThunderboltOutlined,
  ShoppingCartOutlined,
  AppstoreOutlined,
  TagsOutlined,
} from '@ant-design/icons';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useProductStore } from '@stores/productStore';
import { useAuthStore } from '@stores/authStore';
import { productService } from '@/services/product';
import { skuService } from '@/services/sku';
import api from '@/services/api';
import type { Product } from '@/types';
import type { SKUWithSPU } from '@/types/sku';
import dayjs from 'dayjs';

const { Option } = Select;
const { Title, Text } = Typography;
const { useBreakpoint } = Grid;
const { TabPane } = Tabs;

type SortField = 'price' | 'stock' | 'created_at';
type SortOrder = 'asc' | 'desc';

interface ProductFilters {
  minPrice?: number;
  maxPrice?: number;
  category?: string;
  status?: string;
}

interface FlashSaleProduct {
  id: number;
  flash_sale_id: number;
  product_id: number;
  product_name: string;
  flash_price: number;
  original_price: number;
  stock_limit: number;
  stock_sold: number;
  per_user_limit: number;
  discount: number;
}

interface FlashSale {
  id: number;
  name: string;
  description: string;
  start_time: string;
  end_time: string;
  status: string;
  products: FlashSaleProduct[];
}

const sortTypeMap: Record<string, { title: string; icon: React.ReactNode }> = {
  hot: { title: '热销爆款', icon: <FireOutlined style={{ color: '#ff4d4f' }} /> },
  new: { title: '新品上架', icon: <ClockCircleOutlined style={{ color: '#1890ff' }} /> },
  flash: { title: '限时秒杀', icon: <ThunderboltOutlined style={{ color: '#faad14' }} /> },
};

export const ProductListPage: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const screens = useBreakpoint();
  const { products, total, filters, isLoading, error, fetchProducts, setFilters, searchProducts } =
    useProductStore();
  const { user } = useAuthStore();

  const [sortField, setSortField] = useState<SortField>('created_at');
  const [sortOrder, setSortOrder] = useState<SortOrder>('desc');
  const [productFilters, setProductFilters] = useState<ProductFilters>({});
  const [priceRange, setPriceRange] = useState<[number, number]>([0, 10000]);
  const [localLoading, setLocalLoading] = useState(false);
  const [localProducts, setLocalProducts] = useState<Product[]>([]);
  const [flashSales, setFlashSales] = useState<FlashSale[]>([]);
  const [flashLoading, setFlashLoading] = useState(false);
  const [countdown, setCountdown] = useState<string>('');
  const [activeTab, setActiveTab] = useState<'products' | 'skus'>('products');
  const [skus, setSKUs] = useState<SKUWithSPU[]>([]);
  const [skuTotal, setSkuTotal] = useState(0);
  const [skuLoading, setSkuLoading] = useState(false);
  const [skuFilters, setSkuFilters] = useState<{
    type?: string;
    provider?: string;
    tier?: string;
  }>({});

  const sortParam = searchParams.get('sort');
  const flashParam = searchParams.get('flash');
  const categoryParam = searchParams.get('category');
  const searchParam = searchParams.get('search');

  const isSpecialSort = sortParam === 'hot' || sortParam === 'new';
  const isFlashSale = flashParam === 'true';
  const pageTitle = isFlashSale
    ? '限时秒杀'
    : sortParam
      ? sortTypeMap[sortParam]?.title
      : '商品列表';
  const pageIcon = isFlashSale
    ? sortTypeMap['flash'].icon
    : sortParam
      ? sortTypeMap[sortParam]?.icon
      : null;

  useEffect(() => {
    if (isFlashSale) {
      loadFlashSales();
      return;
    }

    if (sortParam === 'hot') {
      loadHotProducts();
    } else if (sortParam === 'new') {
      loadNewProducts();
    } else if (searchParam) {
      searchProducts(searchParam);
    } else {
      fetchProducts();
    }
  }, [sortParam, flashParam, searchParam]);

  useEffect(() => {
    if (activeTab === 'skus') {
      loadSKUs();
    }
  }, [activeTab, skuFilters]);

  const loadSKUs = async () => {
    setSkuLoading(true);
    try {
      const response = await skuService.getPublicSKUs({
        page: filters.page,
        per_page: filters.per_page,
        ...skuFilters,
      });
      const apiResponse = response.data;
      setSKUs(apiResponse.data || []);
      setSkuTotal(apiResponse.total || 0);
    } catch {
      setSKUs([]);
      setSkuTotal(0);
    } finally {
      setSkuLoading(false);
    }
  };

  useEffect(() => {
    if (categoryParam) {
      setProductFilters((prev) => ({ ...prev, category: categoryParam }));
      fetchProducts({ category: categoryParam } as any);
    }
  }, [categoryParam]);

  useEffect(() => {
    if (!isFlashSale || flashSales.length === 0) return;

    const timer = setInterval(() => {
      const now = dayjs();
      const endTime = dayjs(flashSales[0]?.end_time);
      const diff = endTime.diff(now, 'second');

      if (diff <= 0) {
        setCountdown('已结束');
        loadFlashSales();
      } else {
        const hours = Math.floor(diff / 3600);
        const minutes = Math.floor((diff % 3600) / 60);
        const seconds = diff % 60;
        setCountdown(
          `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`
        );
      }
    }, 1000);

    return () => clearInterval(timer);
  }, [isFlashSale, flashSales]);

  const loadFlashSales = async () => {
    setFlashLoading(true);
    try {
      const response = await api.get('/flash-sales/active');
      const data = (response.data as any)?.data || [];
      setFlashSales(data);
    } catch {
      setFlashSales([]);
    } finally {
      setFlashLoading(false);
    }
  };

  const loadHotProducts = async () => {
    setLocalLoading(true);
    try {
      const response = await productService.getHotProducts(50);
      const apiResponse = response.data;
      setLocalProducts(apiResponse.data || []);
    } catch {
      setLocalProducts([]);
    } finally {
      setLocalLoading(false);
    }
  };

  const loadNewProducts = async () => {
    setLocalLoading(true);
    try {
      const response = await productService.getNewProducts(50);
      const apiResponse = response.data;
      setLocalProducts(apiResponse.data || []);
    } catch {
      setLocalProducts([]);
    } finally {
      setLocalLoading(false);
    }
  };

  const handleBuyFlashProduct = (product: FlashSaleProduct) => {
    navigate(`/products/${product.product_id}?flash_sale_id=${product.flash_sale_id}`);
  };

  const displayProducts = isSpecialSort ? localProducts : products;
  const displayLoading = isFlashSale ? flashLoading : isSpecialSort ? localLoading : isLoading;

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
      width: screens.md ? 200 : 100,
      ellipsis: true,
    },
    {
      title: '价格',
      dataIndex: 'price',
      key: 'price',
      render: (price: number) => <Text type="danger">¥{price.toFixed(2)}</Text>,
      width: 100,
    },
    ...(screens.md
      ? [
          {
            title: '库存',
            dataIndex: 'stock',
            key: 'stock',
            width: 80,
          },
        ]
      : []),
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
          <Button type="link" size="small" onClick={() => navigate(`/products/${record.id}`)}>
            详情
          </Button>
          <Button type="link" size="small" onClick={() => navigate(`/products/${record.id}/cart`)}>
            加购
          </Button>
        </Space>
      ),
    },
  ];

  const skuColumns = [
    {
      title: 'SKU名称',
      dataIndex: 'spu_name',
      key: 'spu_name',
      render: (text: string, record: SKUWithSPU) => (
        <a onClick={() => navigate(`/skus/${record.id}`)}>{text}</a>
      ),
    },
    {
      title: '类型',
      dataIndex: 'sku_type',
      key: 'sku_type',
      width: 100,
      render: (type: string) => {
        const typeMap: Record<string, { color: string; text: string }> = {
          token_pack: { color: 'blue', text: 'Token包' },
          subscription: { color: 'green', text: '订阅' },
          concurrent: { color: 'orange', text: '并发' },
          trial: { color: 'purple', text: '试用' },
        };
        const config = typeMap[type] || { color: 'default', text: type };
        return <Tag color={config.color}>{config.text}</Tag>;
      },
    },
    {
      title: '模型',
      dataIndex: 'model_name',
      key: 'model_name',
      width: 120,
      render: (name: string, record: SKUWithSPU) => (
        <Space direction="vertical" size={0}>
          <Text strong>{name}</Text>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {record.model_provider} · {record.model_tier}
          </Text>
        </Space>
      ),
    },
    {
      title: '规格',
      key: 'specs',
      width: 150,
      render: (_: any, record: SKUWithSPU) => {
        if (record.sku_type === 'token_pack') {
          return <Text>{record.token_amount?.toLocaleString()} tokens</Text>;
        } else if (record.sku_type === 'subscription') {
          const periodMap: Record<string, string> = {
            monthly: '月度',
            quarterly: '季度',
            yearly: '年度',
          };
          return <Text>{periodMap[record.subscription_period || 'monthly'] || '月度'}</Text>;
        } else if (record.sku_type === 'concurrent') {
          return <Text>{record.concurrent_requests} 并发</Text>;
        } else if (record.sku_type === 'trial') {
          return <Text>{record.trial_duration_days}天试用</Text>;
        }
        return null;
      },
    },
    {
      title: '价格',
      dataIndex: 'retail_price',
      key: 'retail_price',
      width: 100,
      render: (price: number, record: SKUWithSPU) => (
        <Space direction="vertical" size={0}>
          <Text type="danger" strong>
            ¥{price.toFixed(2)}
          </Text>
          {record.original_price && record.original_price > price && (
            <Text delete type="secondary" style={{ fontSize: 12 }}>
              ¥{record.original_price.toFixed(2)}
            </Text>
          )}
        </Space>
      ),
    },
    {
      title: '拼团',
      key: 'group',
      width: 80,
      render: (_: any, record: SKUWithSPU) =>
        record.group_enabled ? (
          <Tag color="red">
            {record.min_group_size}-{record.max_group_size}人
          </Tag>
        ) : (
          <Text type="secondary">-</Text>
        ),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_: any, record: SKUWithSPU) => (
        <Space>
          <Button type="link" size="small" onClick={() => navigate(`/skus/${record.id}`)}>
            详情
          </Button>
          <Button type="link" size="small" onClick={() => navigate(`/skus/${record.id}/cart`)}>
            加购
          </Button>
        </Space>
      ),
    },
  ];

  const handleSearch = (value: string) => {
    if (value.trim()) {
      navigate(`/products?search=${encodeURIComponent(value.trim())}`);
    }
  };

  const handlePageChange = (page: number, pageSize: number) => {
    setFilters({ page, per_page: pageSize });
    fetchProducts({ page, per_page: pageSize });
  };

  const handleSort = (field: SortField) => {
    const newOrder = sortField === field && sortOrder === 'asc' ? 'desc' : 'asc';
    setSortField(field);
    setSortOrder(newOrder);

    fetchProducts({
      ...filters,
      sort_field: field,
      sort_order: newOrder,
    } as any);
  };

  const handleFilterChange = (key: keyof ProductFilters, value: any) => {
    setProductFilters((prev) => ({ ...prev, [key]: value }));
  };

  const applyFilters = () => {
    fetchProducts({
      ...filters,
      ...productFilters,
      sort_field: sortField,
      sort_order: sortOrder,
    } as any);
  };

  const resetFilters = () => {
    setProductFilters({});
    setPriceRange([0, 10000]);
    setSortField('created_at');
    setSortOrder('desc');
    fetchProducts();
  };

  if (error) {
    return <Empty description={`错误: ${error}`} />;
  }

  if (isFlashSale) {
    return (
      <div style={{ padding: screens.xs ? 12 : 24 }}>
        <Row gutter={16} style={{ marginBottom: 16 }}>
          <Col flex="auto">
            <Space>
              {pageIcon}
              <Title level={screens.xs ? 4 : 3} style={{ margin: 0 }}>
                {pageTitle}
              </Title>
            </Space>
          </Col>
          {flashSales.length > 0 && (
            <Col>
              <Card size="small" style={{ background: '#fff7e6', borderColor: '#faad14' }}>
                <Statistic
                  title="距离结束"
                  value={countdown}
                  valueStyle={{ color: '#faad14', fontSize: screens.xs ? 18 : 24 }}
                />
              </Card>
            </Col>
          )}
        </Row>

        <Spin spinning={displayLoading}>
          {flashSales.length === 0 ? (
            <Card>
              <Empty
                description={
                  <Space direction="vertical">
                    <Title level={4}>暂无进行中的秒杀活动</Title>
                    <Text type="secondary">敬请期待下一场秒杀！</Text>
                    <Button type="primary" onClick={() => navigate('/products')}>
                      浏览商品
                    </Button>
                  </Space>
                }
              />
            </Card>
          ) : (
            flashSales.map((sale) => (
              <div key={sale.id} style={{ marginBottom: 24 }}>
                <Card
                  title={
                    <Space>
                      <ThunderboltOutlined style={{ color: '#faad14' }} />
                      <span>{sale.name}</span>
                      {sale.description && <Text type="secondary">({sale.description})</Text>}
                    </Space>
                  }
                  style={{ marginBottom: 16 }}
                >
                  <Row gutter={[16, 16]}>
                    {sale.products.map((product) => {
                      const stockPercent =
                        ((product.stock_limit - product.stock_sold) / product.stock_limit) * 100;
                      return (
                        <Col xs={24} sm={12} md={8} lg={6} key={product.id}>
                          <Card
                            hoverable
                            cover={
                              <div
                                style={{
                                  height: 120,
                                  background: '#f5f5f5',
                                  display: 'flex',
                                  alignItems: 'center',
                                  justifyContent: 'center',
                                }}
                              >
                                <Text style={{ fontSize: 32 }}>
                                  {product.product_name.substring(0, 2)}
                                </Text>
                              </div>
                            }
                            onClick={() => handleBuyFlashProduct(product)}
                          >
                            <Card.Meta
                              title={product.product_name}
                              description={
                                <Space direction="vertical" style={{ width: '100%' }}>
                                  <Space>
                                    <Text type="danger" strong style={{ fontSize: 18 }}>
                                      ¥{product.flash_price.toFixed(2)}
                                    </Text>
                                    <Text delete type="secondary">
                                      ¥{product.original_price.toFixed(2)}
                                    </Text>
                                  </Space>
                                  <Tag color="red">-{product.discount}%</Tag>
                                  <Progress
                                    percent={100 - stockPercent}
                                    size="small"
                                    strokeColor="#ff4d4f"
                                    format={() =>
                                      `已抢${product.stock_sold}/${product.stock_limit}`
                                    }
                                  />
                                  <Button
                                    type="primary"
                                    danger
                                    block
                                    icon={<ShoppingCartOutlined />}
                                    disabled={stockPercent <= 0}
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      handleBuyFlashProduct(product);
                                    }}
                                  >
                                    {stockPercent <= 0 ? '已抢光' : '立即抢购'}
                                  </Button>
                                </Space>
                              }
                            />
                          </Card>
                        </Col>
                      );
                    })}
                  </Row>
                </Card>
              </div>
            ))
          )}
        </Spin>
      </div>
    );
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
              handleFilterChange('minPrice', value[0]);
              handleFilterChange('maxPrice', value[1]);
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
  ];

  const sortDropdownItems = [
    {
      key: 'price_asc',
      label: '价格从低到高',
      onClick: () => handleSort('price'),
    },
    {
      key: 'price_desc',
      label: '价格从高到低',
      onClick: () => {
        setSortField('price');
        setSortOrder('desc');
      },
    },
    {
      key: 'stock_asc',
      label: '库存从少到多',
      onClick: () => handleSort('stock'),
    },
    {
      key: 'stock_desc',
      label: '库存从多到少',
      onClick: () => {
        setSortField('stock');
        setSortOrder('desc');
      },
    },
    {
      key: 'created_desc',
      label: '最新发布',
      onClick: () => {
        setSortField('created_at');
        setSortOrder('desc');
      },
    },
  ];

  return (
    <div style={{ padding: screens.xs ? 12 : 24 }}>
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col flex="auto">
          <Space>
            {pageIcon}
            <Title level={screens.xs ? 4 : 3} style={{ margin: 0 }}>
              {pageTitle}
            </Title>
          </Space>
        </Col>
      </Row>

      <Tabs
        activeKey={activeTab}
        onChange={(key) => setActiveTab(key as 'products' | 'skus')}
        style={{ marginBottom: 16 }}
      >
        <TabPane
          tab={
            <span>
              <AppstoreOutlined />
              商品
            </span>
          }
          key="products"
        >
          <Row gutter={16} style={{ marginBottom: '20px' }}>
            <Col flex="auto">
              <Input.Search
                placeholder="搜索产品..."
                prefix={<SearchOutlined />}
                onSearch={handleSearch}
                defaultValue={searchParam || ''}
                style={{ width: '100%' }}
              />
            </Col>
            {!isSpecialSort && (
              <Col>
                <Space>
                  <Dropdown menu={{ items: filterDropdownItems }} trigger={['click']}>
                    <Button icon={<FilterOutlined />}>筛选</Button>
                  </Dropdown>
                  <Dropdown menu={{ items: sortDropdownItems }} trigger={['click']}>
                    <Button icon={<SortAscendingOutlined />}>排序</Button>
                  </Dropdown>
                  {user?.role === 'merchant' && (
                    <Button type="primary" icon={<PlusOutlined />}>
                      发布产品
                    </Button>
                  )}
                </Space>
              </Col>
            )}
          </Row>

          <Spin spinning={displayLoading}>
            <Table
              columns={columns}
              dataSource={displayProducts}
              rowKey="id"
              pagination={false}
              scroll={{ x: 600 }}
              locale={{ emptyText: '暂无数据' }}
              size={screens.xs ? 'small' : 'middle'}
            />
          </Spin>

          {!isSpecialSort && total > 0 && (
            <Pagination
              current={filters.page}
              pageSize={filters.per_page}
              total={total}
              onChange={handlePageChange}
              style={{ marginTop: '20px', textAlign: 'right' }}
            />
          )}
        </TabPane>

        <TabPane
          tab={
            <span>
              <TagsOutlined />
              SKU套餐
            </span>
          }
          key="skus"
        >
          <Row gutter={16} style={{ marginBottom: '20px' }}>
            <Col flex="auto">
              <Space>
                <Select
                  placeholder="类型"
                  allowClear
                  style={{ width: 120 }}
                  onChange={(value) => setSkuFilters((prev) => ({ ...prev, type: value }))}
                >
                  <Option value="token_pack">Token包</Option>
                  <Option value="subscription">订阅</Option>
                  <Option value="concurrent">并发</Option>
                  <Option value="trial">试用</Option>
                </Select>
                <Select
                  placeholder="厂商"
                  allowClear
                  style={{ width: 120 }}
                  onChange={(value) => setSkuFilters((prev) => ({ ...prev, provider: value }))}
                >
                  <Option value="openai">OpenAI</Option>
                  <Option value="anthropic">Anthropic</Option>
                  <Option value="google">Google</Option>
                  <Option value="deepseek">DeepSeek</Option>
                </Select>
                <Select
                  placeholder="层级"
                  allowClear
                  style={{ width: 100 }}
                  onChange={(value) => setSkuFilters((prev) => ({ ...prev, tier: value }))}
                >
                  <Option value="pro">Pro</Option>
                  <Option value="lite">Lite</Option>
                  <Option value="mini">Mini</Option>
                  <Option value="vision">Vision</Option>
                </Select>
              </Space>
            </Col>
          </Row>

          <Spin spinning={skuLoading}>
            <Table
              columns={skuColumns}
              dataSource={skus}
              rowKey="id"
              pagination={false}
              scroll={{ x: 800 }}
              locale={{ emptyText: '暂无SKU数据' }}
              size={screens.xs ? 'small' : 'middle'}
            />
          </Spin>

          {skuTotal > 0 && (
            <Pagination
              current={filters.page}
              pageSize={filters.per_page}
              total={skuTotal}
              onChange={handlePageChange}
              style={{ marginTop: '20px', textAlign: 'right' }}
            />
          )}
        </TabPane>
      </Tabs>
    </div>
  );
};

export default ProductListPage;
