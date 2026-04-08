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
  Segmented,
  Tooltip,
} from 'antd';
import {
  SearchOutlined,
  FireOutlined,
  ClockCircleOutlined,
  ThunderboltOutlined,
  ShoppingCartOutlined,
  GiftOutlined,
  AppstoreOutlined,
  UnorderedListOutlined,
  FilterOutlined,
} from '@ant-design/icons';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useCartStore } from '@stores/cartStore';
import { skuService } from '@/services/sku';
import api from '@/services/api';
import type { SKUWithSPU } from '@/types/sku';
import dayjs from 'dayjs';
import { ScenarioFilter } from '@/components/ScenarioFilter';
import { CatalogFilterDrawer, type CatalogFilterValues } from '@/components/CatalogFilterDrawer';
import { productService } from '@/services/product';
import type { Category } from '@/types';
import { getSkuCardSubtitle, pushRecentSearch } from '@/utils/productDisplay';
import styles from './ProductListPage.module.css';

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

  const [searchInput, setSearchInput] = useState('');
  const [viewMode, setViewMode] = useState<'table' | 'grid'>('table');
  const [filterDrawerOpen, setFilterDrawerOpen] = useState(false);
  const [catalogCategories, setCatalogCategories] = useState<Category[]>([]);

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

  useEffect(() => {
    setSearchInput(searchParam || '');
  }, [searchParam]);

  useEffect(() => {
    productService
      .getCategories()
      .then((res) => {
        const body = res.data as { data?: Category[] };
        setCatalogCategories(body?.data || []);
      })
      .catch(() => setCatalogCategories([]));
  }, []);

  useEffect(() => {
    if (searchParams.get('filters') === '1') {
      setFilterDrawerOpen(true);
      const next = new URLSearchParams(searchParams);
      next.delete('filters');
      const qs = next.toString();
      navigate(qs ? `/catalog?${qs}` : '/catalog', { replace: true });
    }
  }, [searchParams, navigate]);

  const filterInitialValues = useMemo((): CatalogFilterValues => {
    const ge = searchParams.get('group_enabled');
    return {
      q: searchParam || '',
      category: searchParams.get('category') || undefined,
      model_name: searchParams.get('model_name') || undefined,
      provider: searchParams.get('provider') || undefined,
      tier: searchParams.get('tier') || undefined,
      sku_type: searchParams.get('type') || undefined,
      group_enabled: ge === 'true' || ge === '1',
      price_min: parseNum(searchParams.get('price_min')) ?? null,
      price_max: parseNum(searchParams.get('price_max')) ?? null,
      valid_days_min: parseIntParam(searchParams.get('valid_days_min')) ?? null,
      valid_days_max: parseIntParam(searchParams.get('valid_days_max')) ?? null,
      sort: searchParams.get('sort') || undefined,
    };
  }, [searchParams, searchParam]);

  const applyCatalogFilters = useCallback(
    (values: CatalogFilterValues) => {
      const params = new URLSearchParams();
      const q = String(values.q ?? '').trim();
      if (q) params.set('q', q);
      if (values.category) params.set('category', String(values.category));
      if (values.model_name) params.set('model_name', String(values.model_name).trim());
      if (values.provider) params.set('provider', String(values.provider));
      if (values.tier) params.set('tier', String(values.tier));
      if (values.sku_type) params.set('type', String(values.sku_type));
      if (values.group_enabled) params.set('group_enabled', 'true');
      if (values.price_min != null && Number.isFinite(Number(values.price_min)))
        params.set('price_min', String(values.price_min));
      if (values.price_max != null && Number.isFinite(Number(values.price_max)))
        params.set('price_max', String(values.price_max));
      if (values.valid_days_min != null && Number.isFinite(Number(values.valid_days_min)))
        params.set('valid_days_min', String(values.valid_days_min));
      if (values.valid_days_max != null && Number.isFinite(Number(values.valid_days_max)))
        params.set('valid_days_max', String(values.valid_days_max));
      if (values.sort) params.set('sort', String(values.sort));
      const qs = params.toString();
      navigate(qs ? `/catalog?${qs}` : '/catalog');
    },
    [navigate]
  );

  const effectiveView = screens.xs ? 'grid' : viewMode;

  const filterChips = useMemo(() => {
    const chips: { key: string; label: string }[] = [];
    const q = (searchParam || '').trim();
    if (q) chips.push({ key: 'q', label: `搜索：${q}` });

    const sort = searchParams.get('sort');
    if (sort) {
      const map: Record<string, string> = {
        hot: '热销',
        new: '最新',
        price_asc: '价格从低到高',
        price_desc: '价格从高到低',
      };
      chips.push({ key: 'sort', label: `排序：${map[sort] || sort}` });
    }

    const type = searchParams.get('type');
    if (type) {
      const map: Record<string, string> = {
        token_pack: 'Token 包',
        subscription: '订阅',
        concurrent: '并发',
        trial: '试用',
      };
      chips.push({ key: 'type', label: `类型：${map[type] || type}` });
    }

    const provider = searchParams.get('provider');
    if (provider) {
      const map: Record<string, string> = {
        openai: 'OpenAI',
        anthropic: 'Anthropic',
        google: 'Google',
        deepseek: 'DeepSeek',
        zhipu: '智谱',
      };
      chips.push({ key: 'provider', label: `厂商：${map[provider] || provider}` });
    }

    const tier = searchParams.get('tier');
    if (tier) chips.push({ key: 'tier', label: `层级：${tier}` });

    const cat = searchParams.get('category');
    if (cat) chips.push({ key: 'category', label: `分类：${cat}` });

    const scenario = searchParams.get('scenario');
    if (scenario) chips.push({ key: 'scenario', label: `场景：${scenario}` });

    if (searchParams.get('group_enabled') === 'true' || searchParams.get('group_enabled') === '1') {
      chips.push({ key: 'group_enabled', label: '仅拼团' });
    }

    const pmin = searchParams.get('price_min');
    const pmax = searchParams.get('price_max');
    if (pmin != null || pmax != null) {
      chips.push({
        key: 'price_range',
        label: `价格：¥${pmin ?? '—'} ~ ¥${pmax ?? '—'}`,
      });
    }

    const modelName = searchParams.get('model_name');
    if (modelName) chips.push({ key: 'model_name', label: `模型：${modelName}` });

    const vdMin = searchParams.get('valid_days_min');
    const vdMax = searchParams.get('valid_days_max');
    if (vdMin != null || vdMax != null) {
      chips.push({
        key: 'valid_days',
        label: `有效期：${vdMin ?? '—'} ~ ${vdMax ?? '—'} 天`,
      });
    }

    return chips;
  }, [searchParams, searchParam]);

  const setSearchParamsNavigate = useCallback(
    (mutate: (next: URLSearchParams) => void) => {
      const next = new URLSearchParams(searchParams);
      mutate(next);
      navigate(`/catalog?${next.toString()}`);
    },
    [navigate, searchParams]
  );

  const removeFilterChip = (key: string) => {
    setSearchParamsNavigate((next) => {
      if (key === 'q') {
        next.delete('q');
        next.delete('search');
      } else if (key === 'price_range') {
        next.delete('price_min');
        next.delete('price_max');
      } else if (key === 'valid_days') {
        next.delete('valid_days_min');
        next.delete('valid_days_max');
      } else {
        next.delete(key);
      }
    });
  };

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
        provider: searchParams.get('provider') || undefined,
        tier: searchParams.get('tier') || undefined,
        model_name: searchParams.get('model_name') || undefined,
        type: searchParams.get('type') || undefined,
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
  }, [isFlashSale, searchParams, skuPage, searchParam, groupCatalog]);

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
    if (q) pushRecentSearch(q);
    if (!q) {
      const next = new URLSearchParams(searchParams);
      next.delete('q');
      next.delete('search');
      navigate(next.toString() ? `/catalog?${next.toString()}` : '/catalog');
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

  const renderSkuTypeTag = (type: string) => {
    const typeMap: Record<string, { color: string; text: string }> = {
      token_pack: { color: 'blue', text: 'Token包' },
      subscription: { color: 'green', text: '订阅' },
      concurrent: { color: 'orange', text: '并发' },
      trial: { color: 'purple', text: '试用' },
    };
    const config = typeMap[type] || { color: 'default', text: type };
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  return (
    <div className={styles.wrap} style={{ padding: screens.xs ? 12 : 24 }}>
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

      <ScenarioFilter variant={sortParam === 'hot' ? 'rail' : 'panel'} />

      {filterChips.length > 0 && (
        <div className={styles.chipRow}>
          <span className={styles.chipLabel}>当前筛选</span>
          {filterChips.map((c) => (
            <Tag key={c.key} closable onClose={() => removeFilterChip(c.key)}>
              {c.label}
            </Tag>
          ))}
          <Button type="link" size="small" onClick={() => navigate('/catalog')}>
            清除全部
          </Button>
        </div>
      )}

      <Row gutter={[16, 16]} style={{ marginBottom: 20 }}>
        <Col xs={24} lg={10}>
          <Space.Compact style={{ width: '100%', display: 'flex' }}>
            <Input.Search
              placeholder="搜索 SKU / 模型 / 关键词..."
              prefix={<SearchOutlined />}
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              onSearch={handleSearch}
              allowClear
              enterButton
              style={{ flex: 1, minWidth: 0 }}
            />
            <Button
              type="default"
              icon={<FilterOutlined />}
              onClick={() => setFilterDrawerOpen(true)}
              aria-label="打开筛选抽屉"
            >
              {screens.xs ? '' : '筛选'}
            </Button>
          </Space.Compact>
        </Col>
        <Col xs={24} lg={14}>
          <Space wrap size="middle" style={{ width: '100%', justifyContent: 'flex-end' }}>
            <Select
              placeholder="类型"
              allowClear
              style={{ width: screens.xs ? '100%' : 120 }}
              value={searchParams.get('type') || undefined}
              onChange={(value) => {
                const next = new URLSearchParams(searchParams);
                if (value) next.set('type', value);
                else next.delete('type');
                navigate(`/catalog?${next.toString()}`);
              }}
            >
              <Option value="token_pack">Token包</Option>
              <Option value="subscription">订阅</Option>
              <Option value="concurrent">并发</Option>
              <Option value="trial">试用</Option>
            </Select>
            <Select
              placeholder="厂商"
              allowClear
              style={{ width: screens.xs ? '100%' : 120 }}
              value={searchParams.get('provider') || undefined}
              onChange={(value) => {
                const next = new URLSearchParams(searchParams);
                if (value) next.set('provider', value);
                else next.delete('provider');
                navigate(`/catalog?${next.toString()}`);
              }}
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
              style={{ width: screens.xs ? '100%' : 100 }}
              value={searchParams.get('tier') || undefined}
              onChange={(value) => {
                const next = new URLSearchParams(searchParams);
                if (value) next.set('tier', value);
                else next.delete('tier');
                navigate(`/catalog?${next.toString()}`);
              }}
            >
              <Option value="pro">Pro</Option>
              <Option value="lite">Lite</Option>
              <Option value="mini">Mini</Option>
              <Option value="vision">Vision</Option>
            </Select>
            <Select
              placeholder="排序"
              allowClear
              style={{ width: screens.xs ? '100%' : 140 }}
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
            {!screens.xs && (
              <Tooltip title="移动端默认使用卡片视图，便于浏览">
                <Segmented
                  value={viewMode}
                  onChange={(v) => setViewMode(v as 'table' | 'grid')}
                  options={[
                    { value: 'table', icon: <UnorderedListOutlined />, label: '表格' },
                    { value: 'grid', icon: <AppstoreOutlined />, label: '卡片' },
                  ]}
                />
              </Tooltip>
            )}
          </Space>
        </Col>
      </Row>

      <Spin spinning={skuLoading}>
        {effectiveView === 'grid' ? (
          skus.length === 0 ? (
            <Empty description={catalogEmptyDescription} />
          ) : (
            <Row gutter={[16, 16]}>
              {skus.map((record) => (
                <Col xs={24} sm={12} md={8} lg={6} key={record.id}>
                  <Card
                    hoverable
                    styles={{ body: { padding: 12 } }}
                    cover={
                      <div className={styles.cardCover}>
                        {record.thumbnail_url ? (
                          <img
                            className={styles.cardCoverImg}
                            src={record.thumbnail_url}
                            alt=""
                            loading="lazy"
                          />
                        ) : (
                          <div className={styles.cardCoverPlaceholder}>
                            {record.spu_name?.slice(0, 2) || 'SKU'}
                          </div>
                        )}
                      </div>
                    }
                    onClick={() => navigate(`/catalog/${record.id}`)}
                  >
                    <Space direction="vertical" size={4} style={{ width: '100%' }}>
                      <Space wrap size={4}>
                        {renderSkuTypeTag(record.sku_type)}
                        {record.group_enabled && (
                          <Tag color="red">
                            {record.min_group_size}-{record.max_group_size}人团
                          </Tag>
                        )}
                      </Space>
                      <Tooltip title={record.spu_name}>
                        <Text strong ellipsis style={{ width: '100%' }}>
                          {record.spu_name}
                        </Text>
                      </Tooltip>
                      <div className={styles.cardMeta}>{getSkuCardSubtitle(record)}</div>
                      <Space
                        align="baseline"
                        style={{ width: '100%', justifyContent: 'space-between' }}
                      >
                        <Text type="danger" strong style={{ fontSize: 18 }}>
                          ¥{record.retail_price.toFixed(2)}
                        </Text>
                        <Button
                          type="link"
                          size="small"
                          onClick={(e) => {
                            e.stopPropagation();
                            navigate(`/catalog/${record.id}`);
                          }}
                        >
                          详情
                        </Button>
                      </Space>
                    </Space>
                  </Card>
                </Col>
              ))}
            </Row>
          )
        ) : (
          <Table
            columns={skuColumns}
            dataSource={skus}
            rowKey="id"
            pagination={false}
            scroll={{ x: 800 }}
            locale={{ emptyText: <Empty description={catalogEmptyDescription} /> }}
            size={screens.xs ? 'small' : 'middle'}
          />
        )}
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

      <CatalogFilterDrawer
        open={filterDrawerOpen}
        onClose={() => setFilterDrawerOpen(false)}
        categories={catalogCategories}
        initialValues={filterInitialValues}
        onApply={applyCatalogFilters}
      />
    </div>
  );
};

export default ProductListPage;
