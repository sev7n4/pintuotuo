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
  Alert,
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
  TeamOutlined,
} from '@ant-design/icons';
import { useNavigate, useSearchParams, useLocation } from 'react-router-dom';
import { isAxiosError } from 'axios';
import { useCartStore } from '@stores/cartStore';
import { skuService } from '@/services/sku';
import api from '@/services/api';
import type { SKUWithSPU } from '@/types/sku';
import { ENDPOINT_TYPE_LABELS, ENDPOINT_TYPE_COLORS } from '@/types/sku';
import dayjs from 'dayjs';
import { ScenarioFilter } from '@/components/ScenarioFilter';
import { ProductCoverMedia } from '@/components/ProductCoverMedia';
import { IconHintButton } from '@/components/IconHintButton';
import { CatalogFilterDrawer, type CatalogFilterValues } from '@/components/CatalogFilterDrawer';
import { productService } from '@/services/product';
import { groupService, type GroupListScope } from '@/services/group';
import type { Category, Group } from '@/types';
import { parseGroupsListPayload } from '@/utils/groupListPayload';
import { CatalogGroupBanner } from '@/components/catalog/CatalogGroupBanner';
import {
  CatalogGroupsShowcase,
  type CatalogGroupsLoadStatus,
} from '@/components/catalog/CatalogGroupsShowcase';
import { getSkuCardSubtitle, pushRecentSearch } from '@/utils/productDisplay';
import styles from './ProductListPage.module.css';

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

/** 与筛选抽屉一致的 query 键；应用抽屉时先清空再写入，避免残留旧条件 */
const FILTER_DRAWER_PARAM_KEYS = [
  'q',
  'search',
  'category',
  'model_name',
  'provider',
  'tier',
  'type',
  'endpoint_type',
  'group_enabled',
  'price_min',
  'price_max',
  'valid_days_min',
  'valid_days_max',
  'sort',
] as const;

export const ProductListPage: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
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
  /** 卖场 SKU 列表（含拼团入口、热销/新品等）默认卡片，便于浏览；桌面端可切换表格对比字段 */
  const [viewMode, setViewMode] = useState<'table' | 'grid'>('grid');
  const [filterDrawerOpen, setFilterDrawerOpen] = useState(false);
  const [catalogCategories, setCatalogCategories] = useState<Category[]>([]);

  const [catalogView, setCatalogView] = useState<'skus' | 'groups'>('skus');
  const [groupListScope, setGroupListScope] = useState<GroupListScope>('all');
  const [catalogGroupsBlock, setCatalogGroupsBlock] = useState<{
    list: Group[];
    total: number;
    status: CatalogGroupsLoadStatus;
  }>({ list: [], total: 0, status: 'idle' });

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
    return '模型广场';
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
    if (isFlashSale) return;
    let cancelled = false;
    setCatalogGroupsBlock((s) => ({ ...s, status: 'loading' }));
    (async () => {
      try {
        const res = await groupService.listGroups(1, 24, {
          scope: groupListScope,
          status: 'active',
        });
        if (cancelled) return;
        const { list, total } = parseGroupsListPayload(res.data);
        setCatalogGroupsBlock({
          list,
          total,
          status: list.length ? 'ok' : 'empty',
        });
      } catch (e) {
        if (cancelled) return;
        if (isAxiosError(e) && e.response?.status === 401) {
          setCatalogGroupsBlock({ list: [], total: 0, status: 'unauthorized' });
        } else {
          setCatalogGroupsBlock({ list: [], total: 0, status: 'error' });
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [isFlashSale, groupListScope]);

  useEffect(() => {
    if (isFlashSale) setCatalogView('skus');
  }, [isFlashSale]);

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
      endpoint_type: searchParams.get('endpoint_type') || undefined,
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
      const params = new URLSearchParams(searchParams);
      FILTER_DRAWER_PARAM_KEYS.forEach((k) => params.delete(k));

      const q = String(values.q ?? '').trim();
      if (q) params.set('q', q);
      if (values.category) params.set('category', String(values.category));
      if (values.model_name) params.set('model_name', String(values.model_name).trim());
      if (values.provider) params.set('provider', String(values.provider));
      if (values.tier) params.set('tier', String(values.tier));
      if (values.sku_type) params.set('type', String(values.sku_type));
      if (values.endpoint_type) params.set('endpoint_type', String(values.endpoint_type));
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
    [navigate, searchParams]
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

    const endpointType = searchParams.get('endpoint_type');
    if (endpointType) {
      chips.push({
        key: 'endpoint_type',
        label: `端点：${ENDPOINT_TYPE_LABELS[endpointType] || endpointType}`,
      });
    }

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

  const fromCategories = searchParams.get('from') === 'categories';

  const showEntryPresetBanner = useMemo(
    () =>
      !isFlashSale &&
      catalogView === 'skus' &&
      (sortParam === 'hot' || sortParam === 'new' || groupCatalog),
    [isFlashSale, catalogView, sortParam, groupCatalog]
  );

  const entryPresetDescription = useMemo(() => {
    if (groupCatalog) return '当前为「超值拼团」视图，仅展示支持拼团的 SKU。';
    if (sortParam === 'hot') return '当前为热销排序入口预设，仍可修改排序与筛选抽屉中的条件。';
    if (sortParam === 'new') return '当前为新品排序入口预设，仍可修改排序与筛选抽屉中的条件。';
    return '';
  }, [groupCatalog, sortParam]);

  const clearEntryPreset = useCallback(() => {
    const next = new URLSearchParams(searchParams);
    next.delete('sort');
    next.delete('group_enabled');
    const qs = next.toString();
    navigate(qs ? `/catalog?${qs}` : '/catalog');
  }, [navigate, searchParams]);

  const clearCategoriesFromHint = useCallback(() => {
    const next = new URLSearchParams(searchParams);
    next.delete('from');
    const qs = next.toString();
    navigate(qs ? `/catalog?${qs}` : '/catalog', { replace: true });
  }, [navigate, searchParams]);

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
        endpoint_type: searchParams.get('endpoint_type') || undefined,
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
      title: '套餐名称',
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
          <Title level={4}>暂无可拼团的商品</Title>
          <Text type="secondary">可调整筛选条件或稍后再试</Text>
          <Button type="link" onClick={() => navigate('/catalog')}>
            浏览全部商品
          </Button>
        </Space>
      );
    }
    if (sortParam === 'hot') {
      return (
        <Space direction="vertical">
          <Title level={4}>热销活动下暂无商品</Title>
          <Text type="secondary">敬请期待后续上架</Text>
        </Space>
      );
    }
    if (sortParam === 'new') {
      return (
        <Space direction="vertical">
          <Title level={4}>新品活动下暂无商品</Title>
          <Text type="secondary">敬请期待后续上架</Text>
        </Space>
      );
    }
    return '暂无商品数据';
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
                      浏览模型广场
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
                                  <IconHintButton
                                    type="primary"
                                    danger
                                    block
                                    hint={stockPercent <= 0 ? '已抢光' : '立即抢购（按秒杀价下单）'}
                                    icon={<ThunderboltOutlined />}
                                    disabled={stockPercent <= 0}
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      handleBuyFlashProduct(product);
                                    }}
                                  />
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
            tooltip={
              <div>
                {cartItemCount > 0 ? `购物车（${cartItemCount} 件）` : '购物车'}
              </div>
            }
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
      <Row gutter={16} style={{ marginBottom: 16 }} align="middle">
        <Col flex="auto">
          <Space>
            {pageIcon}
            <Title level={screens.xs ? 4 : 3} style={{ margin: 0 }}>
              {pageTitle}
            </Title>
          </Space>
        </Col>
        {groupCatalog && (
          <Col>
            <Button type="default" icon={<TeamOutlined />} onClick={() => navigate('/groups')}>
              {screens.xs ? '参团' : '进行中的拼团'}
            </Button>
          </Col>
        )}
      </Row>

      {!isFlashSale && (
        <>
          <CatalogGroupBanner groupCatalog={groupCatalog} />
          <Segmented
            value={catalogView}
            onChange={(v) => setCatalogView(v as 'skus' | 'groups')}
            options={[
              { label: '挑商品', value: 'skus' },
              { label: '参团中', value: 'groups' },
            ]}
            style={{ marginBottom: 16, maxWidth: screens.xs ? '100%' : 320 }}
          />
        </>
      )}

      {!isFlashSale && catalogView === 'skus' && (
        <CatalogGroupsShowcase
          layout="rail"
          groups={catalogGroupsBlock.list}
          total={catalogGroupsBlock.total}
          status={catalogGroupsBlock.status}
          groupScope={groupListScope}
          onGroupScopeChange={setGroupListScope}
          onOpenGroup={(id) => navigate(`/groups/${id}`)}
          onOpenAll={() => navigate('/groups')}
          onLogin={() =>
            navigate(
              `/login?redirect=${encodeURIComponent(`${location.pathname}${location.search}`)}`
            )
          }
        />
      )}

      {!isFlashSale && catalogView === 'groups' && (
        <CatalogGroupsShowcase
          layout="expanded"
          groups={catalogGroupsBlock.list}
          total={catalogGroupsBlock.total}
          status={catalogGroupsBlock.status}
          groupScope={groupListScope}
          onGroupScopeChange={setGroupListScope}
          onOpenGroup={(id) => navigate(`/groups/${id}`)}
          onOpenAll={() => navigate('/groups')}
          onLogin={() =>
            navigate(
              `/login?redirect=${encodeURIComponent(`${location.pathname}${location.search}`)}`
            )
          }
        />
      )}

      {catalogView === 'skus' && (
        <>
          <ScenarioFilter variant="rail" />

          {fromCategories && (
            <Alert
              type="info"
              showIcon
              closable
              onClose={clearCategoriesFromHint}
              message="已从「浏览场景与层级」进入"
              description="当前 URL 已带上场景或层级筛选；可在下方「当前筛选」标签中逐项关闭，或点「清除全部」。"
              style={{ marginBottom: 12 }}
            />
          )}

          {showEntryPresetBanner && (
            <Alert
              type="info"
              showIcon
              message={entryPresetDescription}
              action={
                <Button size="small" type="link" onClick={clearEntryPreset}>
                  清除活动筛选
                </Button>
              }
              style={{ marginBottom: 12 }}
            />
          )}

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
                  placeholder="搜索模型 / 套餐 / 关键词..."
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
                <Text type="secondary" style={{ maxWidth: screens.xs ? '100%' : 200, fontSize: 12 }}>
                  类型、厂商、层级、端点、价格与有效期在「筛选」抽屉
                </Text>
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
                  <Tooltip title="默认卡片视图便于浏览；切换表格可横向对比多字段">
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
                          <ProductCoverMedia
                            variant="grid"
                            thumbnailUrl={record.thumbnail_url}
                            modelProvider={record.model_provider}
                            fallbackTitle={record.spu_name || record.sku_code || 'SKU'}
                            resetKey={record.id}
                          />
                        }
                        onClick={() => navigate(`/catalog/${record.id}`)}
                      >
                        <Space direction="vertical" size={4} style={{ width: '100%' }}>
                          <Space wrap size={4}>
                            {renderSkuTypeTag(record.sku_type)}
                            {record.endpoint_type &&
                              record.endpoint_type !== 'chat_completions' && (
                                <Tag
                                  color={ENDPOINT_TYPE_COLORS[record.endpoint_type] || 'default'}
                                >
                                  {ENDPOINT_TYPE_LABELS[record.endpoint_type] ||
                                    record.endpoint_type}
                                </Tag>
                              )}
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
        </>
      )}

      <Badge count={cartItemCount} size="small">
        <FloatButton
          icon={<ShoppingCartOutlined />}
          tooltip={
            <div>
              {cartItemCount > 0 ? `购物车（${cartItemCount} 件）` : '购物车'}
            </div>
          }
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
