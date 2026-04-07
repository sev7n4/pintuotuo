import React, { useCallback, useEffect, useMemo, useState } from 'react';
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
  Typography,
  Card,
  Grid,
  Progress,
  Statistic,
  FloatButton,
  Badge,
} from 'antd';
import {
  SearchOutlined,
  FireOutlined,
  ClockCircleOutlined,
  ThunderboltOutlined,
  ShoppingCartOutlined,
  GiftOutlined,
} from '@ant-design/icons';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useCartStore } from '@stores/cartStore';
import { skuService } from '@/services/sku';
import api from '@/services/api';
import type { SKUWithSPU } from '@/types/sku';
import dayjs from 'dayjs';
import { ScenarioFilter } from '@/components/ScenarioFilter';

const { Option } = Select;
const { Title, Text } = Typography;
const { useBreakpoint } = Grid;

interface FlashSaleProduct {
  id: number;
  flash_sale_id: number;
  sku_id: number;
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
  skus: FlashSaleProduct[];
}

const sortTypeMap: Record<string, { title: string; icon: React.ReactNode }> = {
  hot: { title: '热销爆款', icon: <FireOutlined style={{ color: '#ff4d4f' }} /> },
  new: { title: '新品上架', icon: <ClockCircleOutlined style={{ color: '#1890ff' }} /> },
  flash: { title: '限时秒杀', icon: <ThunderboltOutlined style={{ color: '#faad14' }} /> },
  group: { title: '超值拼团', icon: <GiftOutlined style={{ color: '#52c41a' }} /> },
};

function parseNum(s: string | null): number | undefined {
  if (s == null || s === '') return undefined;
  const n = Number(s);
  return Number.isFinite(n) ? n : undefined;
}

function parseIntParam(s: string | null): number | undefined {
  if (s == null || s === '') return undefined;
  const n = parseInt(s, 10);
  return Number.isFinite(n) ? n : undefined;
}

export const ProductListPage: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const screens = useBreakpoint();
  const { items } = useCartStore();

  const [flashSales, setFlashSales] = useState<FlashSale[]>([]);
  const [flashLoading, setFlashLoading] = useState(false);
  const [countdown, setCountdown] = useState<string>('');

  const [skus, setSKUs] = useState<SKUWithSPU[]>([]);
  const [skuTotal, setSkuTotal] = useState(0);
  const [skuLoading, setSkuLoading] = useState(false);
  const [skuPage, setSkuPage] = useState(1);
  const skuPerPage = 20;

  const [skuFilters, setSkuFilters] = useState<{
    type?: string;
    provider?: string;
    tier?: string;
  }>({});

  const sortParam = searchParams.get('sort');
  const flashParam = searchParams.get('flash');
  const searchParam = searchParams.get('search') || searchParams.get('q');
  const groupCatalog =
    searchParams.get('group_enabled') === 'true' || searchParams.get('group_enabled') === '1';

  const isFlashSale = flashParam === 'true';

  const pageTitle = useMemo(() => {
    if (isFlashSale) return '限时秒杀';
    if (groupCatalog) return sortTypeMap.group.title;
    if (sortParam && sortTypeMap[sortParam]) return sortTypeMap[sortParam].title;
    return 'SKU套餐';
  }, [isFlashSale, groupCatalog, sortParam]);

  const pageIcon = useMemo(() => {
    if (isFlashSale) return sortTypeMap.flash.icon;
    if (groupCatalog) return sortTypeMap.group.icon;
    if (sortParam && sortTypeMap[sortParam]) return sortTypeMap[sortParam].icon;
    return null;
  }, [isFlashSale, groupCatalog, sortParam]);

  const loadFlashSales = async () => {
    setFlashLoading(true);
    try {
      const response = await api.get('/flash-sales/active');
      const data = (response.data as { data?: FlashSale[] })?.data || [];
      setFlashSales(data);
    } catch {
      setFlashSales([]);
    } finally {
      setFlashLoading(false);
    }
  };

  const loadSKUs = useCallback(async () => {
    if (isFlashSale) return;
    setSkuLoading(true);
    try {
      const sortRaw = searchParams.get('sort') || undefined;
      const sort =
        sortRaw && ['hot', 'new', 'price_asc', 'price_desc'].includes(sortRaw)
          ? (sortRaw as 'hot' | 'new' | 'price_asc' | 'price_desc')
          : undefined;

      const response = await skuService.getPublicSKUs({
        page: skuPage,
        per_page: skuPerPage,
        q: searchParam || undefined,
        category: searchParams.get('category') || undefined,
        provider: skuFilters.provider || searchParams.get('provider') || undefined,
        tier: skuFilters.tier || searchParams.get('tier') || undefined,
        model_name: searchParams.get('model_name') || undefined,
        type: skuFilters.type || searchParams.get('type') || undefined,
        group_enabled:
          groupCatalog || searchParams.get('group_enabled') === 'true' ? true : undefined,
        price_min: parseNum(searchParams.get('price_min')),
        price_max: parseNum(searchParams.get('price_max')),
        valid_days_min: parseIntParam(searchParams.get('valid_days_min')),
        valid_days_max: parseIntParam(searchParams.get('valid_days_max')),
        sort,
        scenario: searchParams.get('scenario') || undefined,
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
  }, [
    isFlashSale,
    searchParams,
    skuPage,
    skuFilters.provider,
    skuFilters.tier,
    skuFilters.type,
    searchParam,
    groupCatalog,
  ]);

  useEffect(() => {
    if (isFlashSale) {
      loadFlashSales();
    }
  }, [isFlashSale]);

  useEffect(() => {
    setSkuPage(1);
  }, [searchParams.toString()]);

  useEffect(() => {
    if (!isFlashSale) {
      loadSKUs();
    }
  }, [isFlashSale, loadSKUs]);

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

  const handleBuyFlashProduct = (product: FlashSaleProduct) => {
    navigate(`/catalog/${product.sku_id}?flash_sale_id=${product.flash_sale_id}`);
  };

  const displayLoading = isFlashSale ? flashLoading : skuLoading;

  const skuColumns = [
    {
      title: 'SKU / SPU',
      dataIndex: 'spu_name',
      key: 'spu_name',
      render: (text: string, record: SKUWithSPU) => (
        <a onClick={() => navigate(`/catalog/${record.id}`)}>{text}</a>
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
      render: (_: unknown, record: SKUWithSPU) => {
        if (record.sku_type === 'token_pack') {
          return <Text>{record.token_amount?.toLocaleString()} tokens</Text>;
        }
        if (record.sku_type === 'subscription') {
          const periodMap: Record<string, string> = {
            monthly: '月度',
            quarterly: '季度',
            yearly: '年度',
          };
          return <Text>{periodMap[record.subscription_period || 'monthly'] || '月度'}</Text>;
        }
        if (record.sku_type === 'concurrent') {
          return <Text>{record.concurrent_requests} 并发</Text>;
        }
        if (record.sku_type === 'trial') {
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
      title: '已售',
      key: 'sold',
      width: 90,
      render: (_: unknown, record: SKUWithSPU) => (
        <Text>{Number(record.sales_count ?? 0).toLocaleString()}</Text>
      ),
    },
    {
      title: '评分',
      key: 'rating',
      width: 100,
      render: (_: unknown, record: SKUWithSPU) => {
        const r = record.spu_average_rating;
        return r != null && r > 0 ? (
          <Text>{Number(r).toFixed(1)}/5</Text>
        ) : (
          <Text type="secondary">—</Text>
        );
      },
    },
    {
      title: '拼团',
      key: 'group',
      width: 80,
      render: (_: unknown, record: SKUWithSPU) =>
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
      render: (_: unknown, record: SKUWithSPU) => (
        <Space>
          <Button type="link" size="small" onClick={() => navigate(`/catalog/${record.id}`)}>
            详情
          </Button>
        </Space>
      ),
    },
  ];

  const handleSearch = (value: string) => {
    const q = value.trim();
    if (!q) {
      navigate('/catalog');
      return;
    }
    const next = new URLSearchParams(searchParams);
    next.set('q', q);
    next.delete('search');
    navigate(`/catalog?${next.toString()}`);
  };

  const handlePageChange = (page: number) => {
    setSkuPage(page);
  };

  const cartItemCount = items.reduce((sum, item) => sum + item.quantity, 0);

  const catalogEmptyDescription = useMemo(() => {
    if (groupCatalog) {
      return (
        <Space direction="vertical">
          <Title level={4}>暂无可拼团的 SKU</Title>
          <Text type="secondary">可调整筛选条件或稍后再试</Text>
          <Button type="link" onClick={() => navigate('/catalog')}>
            浏览全部 SKU
          </Button>
        </Space>
      );
    }
    if (sortParam === 'hot') {
      return (
        <Space direction="vertical">
          <Title level={4}>热销活动下暂无 SKU</Title>
          <Text type="secondary">敬请期待后续上架</Text>
        </Space>
      );
    }
    if (sortParam === 'new') {
      return (
        <Space direction="vertical">
          <Title level={4}>新品活动下暂无 SKU</Title>
          <Text type="secondary">敬请期待后续上架</Text>
        </Space>
      );
    }
    return '暂无 SKU 数据';
  }, [groupCatalog, sortParam, navigate]);

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
                    <Button type="primary" onClick={() => navigate('/catalog')}>
                      浏览 SKU 套餐
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
                    {sale.skus.map((product) => {
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

        <Badge count={cartItemCount} size="small">
          <FloatButton
            icon={<ShoppingCartOutlined />}
            tooltip={<div>购物车</div>}
            onClick={() => navigate('/cart')}
            style={{ right: 24, bottom: 24 }}
          />
        </Badge>
      </div>
    );
  }

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

      <ScenarioFilter />

      <Row gutter={16} style={{ marginBottom: 20 }}>
        <Col flex="auto">
          <Input.Search
            placeholder="搜索 SKU / 模型 / 关键词..."
            prefix={<SearchOutlined />}
            onSearch={handleSearch}
            defaultValue={searchParam || ''}
            allowClear
            style={{ width: '100%' }}
          />
        </Col>
        <Col>
          <Space wrap>
            <Select
              placeholder="类型"
              allowClear
              style={{ width: 120 }}
              value={skuFilters.type}
              onChange={(value) => setSkuFilters((prev) => ({ ...prev, type: value || undefined }))}
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
              value={skuFilters.provider}
              onChange={(value) =>
                setSkuFilters((prev) => ({ ...prev, provider: value || undefined }))
              }
            >
              <Option value="openai">OpenAI</Option>
              <Option value="anthropic">Anthropic</Option>
              <Option value="google">Google</Option>
              <Option value="deepseek">DeepSeek</Option>
              <Option value="zhipu">智谱</Option>
            </Select>
            <Select
              placeholder="层级"
              allowClear
              style={{ width: 100 }}
              value={skuFilters.tier}
              onChange={(value) => setSkuFilters((prev) => ({ ...prev, tier: value || undefined }))}
            >
              <Option value="pro">Pro</Option>
              <Option value="lite">Lite</Option>
              <Option value="mini">Mini</Option>
              <Option value="vision">Vision</Option>
            </Select>
            <Select
              placeholder="排序"
              allowClear
              style={{ width: 140 }}
              value={sortParam || undefined}
              onChange={(value) => {
                const next = new URLSearchParams(searchParams);
                if (value) next.set('sort', value);
                else next.delete('sort');
                navigate(`/catalog?${next.toString()}`);
              }}
              options={[
                { label: '热销', value: 'hot' },
                { label: '最新', value: 'new' },
                { label: '价格↑', value: 'price_asc' },
                { label: '价格↓', value: 'price_desc' },
              ]}
            />
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
          locale={{ emptyText: <Empty description={catalogEmptyDescription} /> }}
          size={screens.xs ? 'small' : 'middle'}
        />
      </Spin>

      {skuTotal > 0 && (
        <Pagination
          current={skuPage}
          pageSize={skuPerPage}
          total={skuTotal}
          onChange={handlePageChange}
          style={{ marginTop: 20, textAlign: 'right' }}
        />
      )}

      <Badge count={cartItemCount} size="small">
        <FloatButton
          icon={<ShoppingCartOutlined />}
          tooltip={<div>购物车</div>}
          onClick={() => navigate('/cart')}
          style={{ right: 24, bottom: 24 }}
        />
      </Badge>
    </div>
  );
};

export default ProductListPage;
