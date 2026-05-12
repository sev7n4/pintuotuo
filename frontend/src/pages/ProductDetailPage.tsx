import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
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
  Modal,
  Alert,
  Progress,
  Collapse,
  Tooltip,
  Table,
  Grid,
  Drawer,
} from 'antd';
import {
  ShoppingCartOutlined,
  ArrowLeftOutlined,
  ShareAltOutlined,
  TeamOutlined,
  UserOutlined,
  ClockCircleOutlined,
  TagsOutlined,
  UsergroupAddOutlined,
  PlusCircleOutlined,
} from '@ant-design/icons';
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { useProductStore } from '@stores/productStore';
import { useCartStore } from '@stores/cartStore';
import { useGroupStore } from '@stores/groupStore';
import { productService } from '@services/product';
import api from '@/services/api';
import { skuService } from '@services/sku';
import type { Product, GroupPrice, Group } from '@/types';
import type { SKUWithSPU } from '@/types/sku';
import { normalizeGroupDiscountRate } from '@/utils/groupDiscount';
import { getSkuCardSubtitle } from '@/utils/productDisplay';
import styles from './ProductDetailPage.module.css';
import { ProductCoverMedia } from '@/components/ProductCoverMedia';
import { IconHintButton } from '@/components/IconHintButton';

const { Title, Text, Paragraph } = Typography;
const { TabPane } = Tabs;
const { useBreakpoint } = Grid;

/** 加载失败或未返回 SKU 时：用零售价占位各档位，避免写死 demo 价；折扣为 0。 */
function buildRetailOnlyGroupTiers(
  retail: number,
  minGroupSize = 2,
  maxGroupSize = 10
): GroupPrice[] {
  if (retail <= 0) return [];
  const pick: number[] = [];
  for (const n of [2, 5]) {
    if (n >= minGroupSize && n <= maxGroupSize) pick.push(n);
  }
  if (pick.length === 0) {
    pick.push(minGroupSize);
    if (maxGroupSize !== minGroupSize) pick.push(maxGroupSize);
  }
  return [...new Set(pick)]
    .sort((a, b) => a - b)
    .map((min_members) => ({
      min_members,
      price_per_person: Number(retail.toFixed(2)),
      discount_percent: 0,
    }));
}

function buildGroupPricesFromSku(sku: SKUWithSPU | null): GroupPrice[] {
  if (!sku?.group_enabled) return [];
  const retail = sku.retail_price;
  const rate = normalizeGroupDiscountRate(sku.group_discount_rate);
  const perPerson = retail * (1 - rate);
  const discountPct = retail > 0 ? Math.max(0, Math.round((1 - perPerson / retail) * 100)) : 0;
  const pick: number[] = [];
  for (const n of [2, 5]) {
    if (n >= sku.min_group_size && n <= sku.max_group_size) pick.push(n);
  }
  if (pick.length === 0) {
    pick.push(sku.min_group_size);
    if (sku.max_group_size !== sku.min_group_size) pick.push(sku.max_group_size);
  }
  return [...new Set(pick)]
    .sort((a, b) => a - b)
    .map((min_members) => ({
      min_members,
      price_per_person: Number(perPerson.toFixed(2)),
      discount_percent: discountPct,
    }));
}

function resolveGroupPrices(sku: SKUWithSPU | null, product: Product): GroupPrice[] {
  const skuGroupTiers = buildGroupPricesFromSku(sku);
  if (skuGroupTiers.length > 0) return skuGroupTiers;
  if (product.group_prices?.length) return product.group_prices;
  if (sku && !sku.group_enabled) return [];
  const retail = sku?.retail_price ?? product.price ?? 0;
  return buildRetailOnlyGroupTiers(retail);
}

function formatCountdown(deadline: Date): string {
  const now = Date.now();
  const ms = Math.max(0, deadline.getTime() - now);
  const totalSec = Math.floor(ms / 1000);
  const h = Math.floor(totalSec / 3600);
  const m = Math.floor((totalSec % 3600) / 60);
  const s = totalSec % 60;
  return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
}

export const ProductDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const screens = useBreakpoint();
  const { fetchProductByID, isLoading, error } = useProductStore();
  const { addItem } = useCartStore();
  const { createGroup, joinGroup } = useGroupStore();
  const [product, setProduct] = useState<Product | null>(null);
  const [quantity, setQuantity] = useState(1);
  const [purchaseMode, setPurchaseMode] = useState<'single' | 'group'>('group');
  const [selectedGroupPrice, setSelectedGroupPrice] = useState<GroupPrice | null>(null);
  const [activeGroups, setActiveGroups] = useState<Group[]>([]);
  const [showGroupsModal, setShowGroupsModal] = useState(false);
  const [groupsLoading, setGroupsLoading] = useState(false);
  const [joiningGroupId, setJoiningGroupId] = useState<number | null>(null);
  const [detailDrawer, setDetailDrawer] = useState<'group' | null>(null);
  const [skus, setSKUs] = useState<SKUWithSPU[]>([]);
  const [selectedSKU, setSelectedSKU] = useState<SKUWithSPU | null>(null);
  const [skuLoading, setSkuLoading] = useState(false);
  const [groupModalTick, setGroupModalTick] = useState(0);
  const [detailTabKey, setDetailTabKey] = useState('detail');
  const detailTabsRef = useRef<HTMLDivElement>(null);

  const [searchParams] = useSearchParams();
  const flashSaleIdQuery = useMemo(() => {
    const raw = searchParams.get('flash_sale_id');
    if (!raw) return undefined;
    const n = Number(raw);
    return Number.isFinite(n) && n > 0 ? n : undefined;
  }, [searchParams]);

  type FlashSkuLine = {
    flash_price: number;
    sku_id: number;
    stock_limit: number;
    stock_sold: number;
  };
  const [flashSkuLine, setFlashSkuLine] = useState<FlashSkuLine | null>(null);

  useEffect(() => {
    if (!flashSaleIdQuery || !selectedSKU?.id) {
      setFlashSkuLine(null);
      return;
    }
    let cancelled = false;
    void (async () => {
      try {
        const res = await api.get(`/flash-sales/${flashSaleIdQuery}/skus`);
        const body = res.data as { data?: FlashSkuLine[] };
        const rows = body?.data || [];
        const hit = rows.find((r) => r.sku_id === selectedSKU.id);
        if (!cancelled) {
          setFlashSkuLine(hit ?? null);
        }
      } catch {
        if (!cancelled) {
          setFlashSkuLine(null);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [flashSaleIdQuery, selectedSKU?.id]);

  const scrollToProductDetail = useCallback(() => {
    setDetailTabKey('detail');
    requestAnimationFrame(() => {
      detailTabsRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' });
    });
  }, []);

  useEffect(() => {
    if (id) {
      loadProduct();
    }
  }, [id]);

  useEffect(() => {
    if (product?.spu_id) {
      loadSKUs(product.spu_id, product.id);
    }
  }, [product]);

  useEffect(() => {
    if (!product) return;
    const tiers = resolveGroupPrices(selectedSKU, product);
    if (tiers.length > 0) {
      setSelectedGroupPrice((prev) => {
        const match = tiers.find((t) => t.min_members === prev?.min_members);
        return match ?? tiers[0];
      });
    } else {
      setSelectedGroupPrice(null);
    }
  }, [
    selectedSKU?.id,
    selectedSKU?.group_enabled,
    selectedSKU?.retail_price,
    selectedSKU?.group_discount_rate,
    selectedSKU?.min_group_size,
    selectedSKU?.max_group_size,
    product?.id,
    product?.price,
    product?.group_prices,
  ]);

  const loadSKUs = async (spuId: number, currentSkuId: number) => {
    setSkuLoading(true);
    let primary: SKUWithSPU | null = null;
    try {
      try {
        const one = await skuService.getPublicSKU(currentSkuId);
        const body = one.data as { data?: SKUWithSPU };
        if (body?.data) {
          primary = body.data;
          setSelectedSKU(primary);
        }
      } catch {
        /* 单 SKU 预取失败时仍尝试列表接口 */
      }

      const response = await skuService.getPublicSKUs({
        page: 1,
        per_page: 50,
        spu_id: spuId,
      });
      const apiResponse = response.data;
      const productSKUs = apiResponse.data || [];
      setSKUs(productSKUs);
      const match = productSKUs.find((sku: SKUWithSPU) => sku.id === currentSkuId);
      setSelectedSKU(match || primary || productSKUs[0] || null);
    } catch {
      setSKUs([]);
      if (!primary) {
        setSelectedSKU(null);
      }
    } finally {
      setSkuLoading(false);
    }
  };

  const loadProduct = async () => {
    if (!id) return;
    const result = await fetchProductByID(parseInt(id));
    if (result) {
      setProduct(result);
    }
  };

  const loadActiveGroups = useCallback(async () => {
    if (!id) return;
    setGroupsLoading(true);
    try {
      const response = await productService.getProductGroups(parseInt(id, 10));
      if (response.data.code === 0 && response.data.data) {
        setActiveGroups(response.data.data);
      }
    } catch {
      console.error('Failed to load active groups');
    } finally {
      setGroupsLoading(false);
    }
  }, [id]);

  useEffect(() => {
    if (!id || purchaseMode !== 'group') return undefined;
    void loadActiveGroups();
    const interval = setInterval(() => void loadActiveGroups(), 10000);
    return () => clearInterval(interval);
  }, [id, purchaseMode, loadActiveGroups]);

  useEffect(() => {
    if (!showGroupsModal) return undefined;
    const t = setInterval(() => setGroupModalTick((x) => x + 1), 1000);
    return () => clearInterval(t);
  }, [showGroupsModal]);

  const handleJoinGroup = async (group: Group) => {
    setJoiningGroupId(group.id);
    try {
      const orderId = await joinGroup(group.id);
      message.success('加入拼团成功！');
      if (orderId) {
        navigate(`/payment/${orderId}`);
      } else {
        navigate('/orders');
      }
    } catch {
      message.error('加入拼团失败，请重试');
    } finally {
      setJoiningGroupId(null);
    }
  };

  const handleAddToCart = () => {
    if (!product) return;
    const retail = selectedSKU?.retail_price ?? product.price ?? 0;
    const cartPrice = flashSkuLine ? flashSkuLine.flash_price : retail;
    const cap =
      flashSkuLine != null
        ? Math.max(0, flashSkuLine.stock_limit - flashSkuLine.stock_sold)
        : product.stock;
    const cartProduct = { ...product, price: cartPrice, stock: Math.min(product.stock, cap) };
    addItem(cartProduct, quantity, undefined, flashSaleIdQuery);
    message.success(`已添加 ${quantity} 件到购物车`);
    setTimeout(() => navigate('/cart'), 1000);
  };

  const handleGroupPurchase = async () => {
    if (!product) {
      message.error('商品信息加载失败');
      return;
    }
    const currentGroupPrice = selectedGroupPrice || groupPrices[0];
    if (!currentGroupPrice) {
      message.error('请选择拼团规则');
      return;
    }
    try {
      const deadline = new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString();
      const skuId = selectedSKU?.id ?? product.id;
      const orderId = await createGroup(skuId, currentGroupPrice.min_members, deadline);
      if (orderId) {
        message.success('拼团已创建，请完成支付！');
        navigate(`/payment/${orderId}`);
      } else {
        message.error('创建拼团失败，请重试');
      }
    } catch {
      message.error('创建拼团失败，请重试');
    }
  };

  const handleShare = () => {
    if (!product) return;
    const shareUrl = `${window.location.origin}/catalog/${product.id}`;
    if (navigator.clipboard && window.isSecureContext) {
      navigator.clipboard
        .writeText(shareUrl)
        .then(() => {
          message.success('链接已复制到剪贴板');
        })
        .catch(() => {
          message.error('复制失败，请手动复制链接');
        });
    } else {
      // 降级方案：创建一个临时输入框来复制
      const textArea = document.createElement('textarea');
      textArea.value = shareUrl;
      textArea.style.position = 'fixed';
      textArea.style.left = '-999999px';
      textArea.style.top = '-999999px';
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();
      try {
        document.execCommand('copy');
        message.success('链接已复制到剪贴板');
      } catch {
        message.error('复制失败，请手动复制链接');
      } finally {
        document.body.removeChild(textArea);
      }
    }
  };

  const spuProductInfoRows = useMemo(() => {
    if (!product) return [];
    const rows: { key: string; label: string; content: React.ReactNode }[] = [];
    if (product.token_count != null && product.token_count > 0) {
      rows.push({
        key: 'token',
        label: '包含 Token',
        content: (
          <Text strong>
            {(product.token_count / 10000).toFixed(0)}万（{product.token_count.toLocaleString()}）
          </Text>
        ),
      });
    }
    if (product.models && product.models.length > 0) {
      rows.push({
        key: 'models',
        label: '支持模型',
        content: <Text strong>{product.models.join('、')}</Text>,
      });
    }
    if (product.validity_period) {
      rows.push({
        key: 'validity',
        label: '有效期',
        content: <Text strong>{product.validity_period}</Text>,
      });
    }
    if (product.context_length) {
      rows.push({
        key: 'context',
        label: '上下文长度',
        content: <Text strong>{product.context_length}</Text>,
      });
    }
    return rows;
  }, [product]);

  if (error) {
    return <Empty description={`错误: ${error}`} />;
  }

  if (isLoading || !product) {
    return (
      <div
        style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}
      >
        <Spin size="large" />
      </div>
    );
  }

  const groupPrices = resolveGroupPrices(selectedSKU, product);
  const effectiveGroupTier = selectedGroupPrice ?? groupPrices[0] ?? null;

  const baseRetail = selectedSKU?.retail_price ?? product.price ?? 0;
  const displayRetail = flashSkuLine ? flashSkuLine.flash_price : baseRetail;
  const flashQtyCap =
    flashSkuLine != null
      ? Math.max(0, flashSkuLine.stock_limit - flashSkuLine.stock_sold)
      : undefined;
  const maxBuyQty =
    flashQtyCap != null ? Math.min(product.stock, flashQtyCap) : product.stock;

  const estimatedFinalPrice = () => {
    if (purchaseMode === 'group') {
      return effectiveGroupTier?.price_per_person ?? baseRetail;
    }
    return displayRetail;
  };

  const calculateDiscount = () => {
    const basePrice = baseRetail;
    if (!basePrice || !effectiveGroupTier) return 0;
    return Math.max(0, Math.round((1 - effectiveGroupTier.price_per_person / basePrice) * 100));
  };

  const displaySoldCount = Number(selectedSKU?.sales_count ?? product?.sold_count ?? 0);
  const rawRating = selectedSKU?.spu_average_rating ?? product?.rating;
  const hasRating = rawRating != null && Number(rawRating) > 0;
  const rating = hasRating ? Number(rawRating) : null;
  const reviewCount = product?.review_count ?? 0;

  const coverSrc =
    selectedSKU?.thumbnail_url ?? product.thumbnail_url ?? product.image_url ?? undefined;

  const pagePadding = screens.xs ? 12 : 20;

  const primaryTitle = selectedSKU?.spu_name || product.name;
  const subtitle =
    selectedSKU?.spu_name && product.name !== selectedSKU.spu_name ? product.name : null;

  return (
    <div className={styles.page} style={{ padding: pagePadding }}>
      <Space style={{ marginBottom: '20px' }}>
        <Button type="text" icon={<ArrowLeftOutlined />} onClick={() => navigate('/catalog')}>
          返回列表
        </Button>
      </Space>

      <Card>
        <Row gutter={[24, 24]}>
          <Col xs={24} md={12}>
            <div
              className={styles.heroMedia}
              style={{
                height: screens.xs ? 220 : 300,
              }}
            >
              <ProductCoverMedia
                variant="hero"
                coverUrl={coverSrc}
                modelProvider={selectedSKU?.model_provider ?? product.model_provider}
                fallbackTitle={primaryTitle}
                resetKey={selectedSKU?.id ?? product.id}
              />
            </div>
          </Col>

          <Col xs={24} md={12}>
            <Title level={3} style={{ marginBottom: 8 }}>
              {primaryTitle}
            </Title>
            {subtitle ? (
              <Text type="secondary" style={{ display: 'block', marginBottom: 12 }}>
                {subtitle}
              </Text>
            ) : null}
            {flashSkuLine && flashSaleIdQuery ? (
              <Alert
                type="warning"
                showIcon
                style={{ marginBottom: 12 }}
                message="限时秒杀价"
                description={`当前以秒杀价 ¥${displayRetail.toFixed(2)} 加入购物车；结算时将校验场次与库存。`}
              />
            ) : null}
            {selectedSKU?.sku_code ? (
              <Tag style={{ marginBottom: 12 }}>{selectedSKU.sku_code}</Tag>
            ) : null}
            <Space style={{ marginBottom: 16 }} wrap>
              <Tag color="blue">已售 {displaySoldCount.toLocaleString()} 件</Tag>
              <Tag color="gold">
                {hasRating ? (
                  <>
                    ⭐ {rating!.toFixed(1)}/5.0
                    {reviewCount > 0 ? `（${reviewCount} 条评价）` : '（暂无评价）'}
                  </>
                ) : (
                  <>⭐ 暂无评分</>
                )}
              </Tag>
            </Space>

            {selectedSKU ? (
              <Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
                已选套餐：
                <Tooltip title={selectedSKU.spu_name}>
                  <Text strong>{getSkuCardSubtitle(selectedSKU) || selectedSKU.spu_name}</Text>
                </Tooltip>
                <Text type="secondary">
                  （完整套餐名见气泡；用量与价格见下方「选择套餐」，商品级说明见「商品详情」）
                </Text>
              </Text>
            ) : null}
            <Button
              type="link"
              onClick={scrollToProductDetail}
              style={{ padding: 0, height: 'auto' }}
            >
              查看完整规格、商品介绍与常见问题
            </Button>
          </Col>
        </Row>

        <Divider />

        {skus.length > 0 && (
          <div style={{ marginBottom: 24 }}>
            <Title level={4}>
              <TagsOutlined /> 选择套餐
            </Title>
            <Spin spinning={skuLoading}>
              <Space direction="vertical" style={{ width: '100%' }} size="middle">
                {skus.map((sku) => (
                  <Card
                    key={sku.id}
                    size="small"
                    hoverable
                    className={styles.skuPickCard}
                    style={{
                      border:
                        selectedSKU?.id === sku.id ? '2px solid #1890ff' : '1px solid #d9d9d9',
                      background: selectedSKU?.id === sku.id ? '#e6f7ff' : 'white',
                    }}
                    onClick={() => setSelectedSKU(sku)}
                  >
                    <div className={styles.skuPickRow}>
                      <div className={styles.skuPickMain}>
                        <div className={styles.skuPickTypeRow}>
                          <Tag
                            color={
                              sku.sku_type === 'token_pack'
                                ? 'blue'
                                : sku.sku_type === 'subscription'
                                  ? 'green'
                                  : 'orange'
                            }
                          >
                            {sku.sku_type === 'token_pack'
                              ? 'Token包'
                              : sku.sku_type === 'subscription'
                                ? '订阅'
                                : '并发'}
                          </Tag>
                          <Tag color="geekblue">{sku.sku_code}</Tag>
                        </div>
                        <Text
                          strong
                          className={styles.skuPickName}
                          ellipsis
                          style={{ display: 'block' }}
                        >
                          {getSkuCardSubtitle(sku) || sku.spu_name}
                        </Text>
                        <Tooltip title={sku.spu_name}>
                          <div className={styles.skuPickSummary}>
                            <Text type="secondary" ellipsis style={{ fontSize: 12 }}>
                              {sku.spu_name}
                            </Text>
                          </div>
                        </Tooltip>
                      </div>
                      <div className={styles.skuPickPriceCol}>
                        <Text strong style={{ fontSize: 18, color: '#1890ff', display: 'block' }}>
                          ¥{sku.retail_price.toFixed(2)}
                        </Text>
                        {sku.original_price && sku.original_price > sku.retail_price && (
                          <Text delete type="secondary" style={{ fontSize: 12 }}>
                            ¥{sku.original_price.toFixed(2)}
                          </Text>
                        )}
                        {sku.group_enabled && (
                          <Tag color="red" style={{ marginTop: 4 }}>
                            <TeamOutlined /> {sku.min_group_size}-{sku.max_group_size}人团
                          </Tag>
                        )}
                      </div>
                    </div>
                  </Card>
                ))}
              </Space>
            </Spin>
          </div>
        )}

        <Divider />

        <div style={{ marginBottom: 24 }}>
          <Title level={4}>定价信息</Title>
          {screens.xs ? (
            <div className={styles.pricingRowMobile}>
              <div
                role="button"
                tabIndex={0}
                className={`${styles.pricingCell} ${purchaseMode === 'single' ? styles.pricingCellActive : ''}`}
                onClick={() => setPurchaseMode('single')}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    setPurchaseMode('single');
                  }
                }}
              >
                <Space direction="vertical" style={{ width: '100%' }} size={4}>
                  <Text type="secondary">单独购买</Text>
                  <Statistic
                    value={displayRetail}
                    prefix="¥"
                    valueStyle={{ color: '#333', fontSize: 22 }}
                  />
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    到手约 ¥{estimatedFinalPrice().toFixed(2)}
                  </Text>
                </Space>
              </div>
              <div
                role="button"
                tabIndex={0}
                className={`${styles.pricingCell} ${purchaseMode === 'group' ? styles.pricingCellGroupActive : ''}`}
                onClick={() => setPurchaseMode('group')}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    setPurchaseMode('group');
                  }
                }}
              >
                <Space direction="vertical" style={{ width: '100%' }} size={4}>
                  <Space size={4} wrap>
                    <Text type="secondary">拼团</Text>
                    <Tag color="green">推荐</Tag>
                  </Space>
                  <Statistic
                    value={
                      effectiveGroupTier?.price_per_person ??
                      selectedSKU?.retail_price ??
                      product.price
                    }
                    prefix="¥"
                    valueStyle={{ color: '#52c41a', fontSize: 22 }}
                    suffix={
                      <Text type="secondary" style={{ fontSize: 12 }}>
                        /人
                      </Text>
                    }
                  />
                  <Text type="success" style={{ fontSize: 12 }}>
                    到手 ¥{estimatedFinalPrice().toFixed(2)}
                    {calculateDiscount() > 0 ? ` · 省${calculateDiscount()}%` : ''}
                  </Text>
                </Space>
              </div>
            </div>
          ) : (
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
                      value={displayRetail}
                      prefix="¥"
                      valueStyle={{ color: '#333', fontSize: 24 }}
                    />
                    <Text type="secondary">
                      直接下单，预计到手 ¥{estimatedFinalPrice().toFixed(2)}
                    </Text>
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
                      value={
                        effectiveGroupTier?.price_per_person ??
                        selectedSKU?.retail_price ??
                        product.price
                      }
                      prefix="¥"
                      valueStyle={{ color: '#52c41a', fontSize: 24 }}
                      suffix={
                        <Text type="secondary" style={{ fontSize: 14 }}>
                          /人
                        </Text>
                      }
                    />
                    <Text type="success">
                      预计到手 ¥{estimatedFinalPrice().toFixed(2)}
                      {calculateDiscount() > 0 ? `，立省 ${calculateDiscount()}%` : ''}
                    </Text>
                  </Space>
                </Card>
              </Col>
            </Row>
          )}
        </div>

        {purchaseMode === 'group' && (
          <div style={{ marginBottom: 24 }}>
            {groupPrices.length === 0 && (
              <div style={{ marginBottom: 16 }}>
                <Text type="secondary">当前规格不可发起拼团。</Text>{' '}
                <Button
                  type="link"
                  size="small"
                  style={{ padding: 0 }}
                  onClick={() => setDetailDrawer('group')}
                >
                  说明
                </Button>
              </div>
            )}
            {groupPrices.length > 0 && <Title level={5}>选择拼团规则</Title>}
            <Space direction="vertical" style={{ width: '100%' }}>
              {groupPrices.map((gp) => (
                <Card
                  key={gp.min_members}
                  size="small"
                  hoverable
                  style={{
                    border:
                      selectedGroupPrice?.min_members === gp.min_members
                        ? '2px solid #52c41a'
                        : '1px solid #d9d9d9',
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
                        <Text delete type="secondary">
                          ¥{(selectedSKU?.retail_price || product.price).toFixed(2)}
                        </Text>
                        <Text strong style={{ color: '#52c41a', fontSize: 18 }}>
                          ¥{gp.price_per_person}/人
                        </Text>
                      </Space>
                    </Col>
                  </Row>
                </Card>
              ))}
            </Space>
          </div>
        )}

        <Space direction="vertical" size="large" style={{ width: '100%', maxWidth: '100%' }}>
          <div>
            <span>购买数量: </span>
            <InputNumber
              min={1}
              max={Math.max(1, maxBuyQty)}
              value={quantity}
              onChange={(val) => setQuantity(val || 1)}
              style={{ marginLeft: '10px', width: 100 }}
            />
            <Text type="secondary" style={{ marginLeft: 16 }}>
              库存: {product.stock} 件
            </Text>
          </div>

          <div style={{ width: '100%', maxWidth: '100%', boxSizing: 'border-box' }}>
            {purchaseMode === 'single' ? (
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
            ) : (
              <Space style={{ width: '100%' }} direction="vertical" size="middle">
                <Card
                  size="small"
                  style={{ width: '100%', background: '#f6ffed', borderColor: '#b7eb8f' }}
                >
                  <Space style={{ width: '100%', justifyContent: 'space-between' }} wrap>
                    <Text>当前可加入团数：{activeGroups.length}</Text>
                    <Text type="success">
                      拼团每人最高省 ¥
                      {(
                        (selectedSKU?.retail_price || product.price) - estimatedFinalPrice()
                      ).toFixed(2)}
                    </Text>
                  </Space>
                </Card>
                <div className={styles.groupActionsRow}>
                  <div className={styles.groupActionCell}>
                    <span
                      style={{
                        display: 'flex',
                        width: '100%',
                        minWidth: 0,
                        cursor: product.stock === 0 ? 'not-allowed' : undefined,
                      }}
                    >
                      <Button
                        type="primary"
                        size="large"
                        block
                        icon={<UsergroupAddOutlined />}
                        aria-label={
                          product.stock === 0 ? '暂无库存' : '加入拼团，打开可加入的团列表'
                        }
                        className={styles.groupPairBtn}
                        onClick={() => {
                          loadActiveGroups();
                          setShowGroupsModal(true);
                        }}
                        disabled={product.stock === 0}
                        style={{ background: '#52c41a', borderColor: '#52c41a' }}
                      >
                        {product.stock === 0 ? '暂无库存' : '加入拼团'}
                      </Button>
                    </span>
                  </div>
                  <div className={styles.groupActionCell}>
                    <span
                      style={{
                        display: 'flex',
                        width: '100%',
                        minWidth: 0,
                        cursor:
                          product.stock === 0 || groupPrices.length === 0
                            ? 'not-allowed'
                            : undefined,
                      }}
                    >
                      <Button
                        type="primary"
                        size="large"
                        block
                        icon={<PlusCircleOutlined />}
                        aria-label={
                          product.stock === 0
                            ? '暂无库存'
                            : groupPrices.length === 0
                              ? '当前规格不可拼团'
                              : '发起拼团并前往支付'
                        }
                        className={styles.groupPairBtn}
                        onClick={handleGroupPurchase}
                        disabled={product.stock === 0 || groupPrices.length === 0}
                        style={{
                          background: '#1890ff',
                          borderColor: '#1890ff',
                        }}
                      >
                        {product.stock === 0
                          ? '暂无库存'
                          : groupPrices.length === 0
                            ? '不可拼团'
                            : '发起拼团'}
                      </Button>
                    </span>
                  </div>
                </div>
              </Space>
            )}

            <div className={styles.secondaryRow}>
              <div className={styles.secondaryCell}>
                <IconHintButton
                  hint="复制商品分享链接"
                  type="default"
                  size="large"
                  block
                  icon={<ShareAltOutlined />}
                  onClick={handleShare}
                  className={styles.iconToolbarBtn}
                >
                  {screens.xs ? '分享' : null}
                </IconHintButton>
              </div>
              <div className={styles.secondaryCell}>
                <IconHintButton
                  hint="打开购物车"
                  type="default"
                  size="large"
                  block
                  icon={<ShoppingCartOutlined />}
                  onClick={() => navigate('/cart')}
                  className={styles.iconToolbarBtn}
                >
                  {screens.xs ? '购物车' : null}
                </IconHintButton>
              </div>
            </div>
          </div>
        </Space>
      </Card>

      <Card ref={detailTabsRef} className={styles.tabsCard}>
        <Tabs activeKey={detailTabKey} onChange={setDetailTabKey}>
          <TabPane tab="商品详情" key="detail">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Title level={5}>商品说明</Title>
              <Paragraph type="secondary" style={{ marginBottom: 12, fontSize: 13 }}>
                以下为与目录同步的概览信息；具体 Token 量、价格与拼团规则以「选择套餐」卡片为准。
              </Paragraph>
              {spuProductInfoRows.length > 0 ? (
                <ul style={{ paddingLeft: 20 }}>
                  {spuProductInfoRows.map((r) => (
                    <li key={r.key}>
                      {r.label}：{r.content}
                    </li>
                  ))}
                </ul>
              ) : (
                <Alert
                  type="info"
                  showIcon
                  message="暂无商品级概要"
                  description={
                    selectedSKU ? (
                      <>
                        当前所选套餐摘要：
                        <Text strong> {getSkuCardSubtitle(selectedSKU)}</Text>
                        。未配置商品级 token/模型/上下文等时，以套餐卡与订单为准。
                      </>
                    ) : (
                      '请先在上方「选择套餐」；关键用量与价格以所选 SKU 为准。'
                    )
                  }
                />
              )}

              {product.description?.trim() ? (
                <>
                  <Divider />
                  <Title level={5}>商品介绍</Title>
                  <Paragraph type="secondary" style={{ marginBottom: 0, whiteSpace: 'pre-wrap' }}>
                    {product.description.trim()}
                  </Paragraph>
                </>
              ) : null}

              <Divider />

              <Title level={5}>使用指南（商品级）</Title>
              <Paragraph type="secondary" style={{ marginBottom: 12 }}>
                以下为通用说明；您实际享有的模型名与调用示例以购买后「我的服务 → 我的
                Token」中「本接口调用说明」为准。
              </Paragraph>
              <ol style={{ paddingLeft: 20 }}>
                <li>购买成功后，Token 将按规则充值到您的账户（以订单为准）</li>
                <li>在「我的 Token」可查看余额、创建平台 API 密钥（ptd_ 前缀）</li>
                <li>
                  支持 OpenAI 兼容路径：将 Base URL 设为平台提供的地址，请求体中填写
                  model（可用「厂商/模型」或兼容前缀匹配）
                </li>
                <li>
                  流式输出（请求体 <Text code>stream: true</Text>
                  ）平台已支持；客户端/SDK 接入示例见{' '}
                  <Link to="/developer/quickstart">开发者中心 → 快速开始</Link>。
                </li>
                <li>
                  更多接口说明、模型列表、用量与排错请前往{' '}
                  <Link to="/developer/quickstart">开发者中心</Link>。
                </li>
              </ol>
              {selectedSKU && (
                <>
                  <Divider />
                  <Title level={5}>规格与计费参考</Title>
                  <Text
                    type="secondary"
                    style={{ display: 'block', marginBottom: 8, fontSize: 13 }}
                  >
                    以下为便于核对的参数摘要；实际扣费以调用用量与账单为准。
                  </Text>
                  <Collapse
                    bordered={false}
                    defaultActiveKey={[]}
                    items={[
                      {
                        key: 'billing',
                        label: '计费说明（技术细节，默认收起）',
                        children: (
                          <>
                            {selectedSKU.catalog_pricing_version_id != null ? (
                              <Alert
                                type="info"
                                showIcon
                                style={{ marginBottom: 8 }}
                                message={
                                  <>
                                    「参考输入/输出成本」按您<strong>最近已支付订单</strong>
                                    锁定的定价版本（
                                    <Text code>
                                      pricing_version_id = {selectedSKU.catalog_pricing_version_id}
                                    </Text>
                                    ）展示；未登录用户为 baseline 版本。
                                  </>
                                }
                              />
                            ) : (
                              <Paragraph type="secondary" style={{ marginBottom: 8, fontSize: 13 }}>
                                「参考输入/输出成本」优先取自 baseline 定价版本在{' '}
                                <Text code>pricing_version_spu_rates</Text>{' '}
                                的快照；若无快照再回落目录当前列。
                              </Paragraph>
                            )}
                            <ul style={{ paddingLeft: 20 }}>
                              <li>
                                厂商代码（provider）：<Text code>{selectedSKU.model_provider}</Text>
                              </li>
                              <li>
                                模型名参考：{' '}
                                <Text code>
                                  {selectedSKU.provider_model_id?.trim() || selectedSKU.model_name}
                                </Text>
                                {selectedSKU.provider_model_id?.trim() ? (
                                  <Text type="secondary">（目录配置的 provider_model_id）</Text>
                                ) : null}
                              </li>
                              <li>
                                OpenAI 兼容示例：
                                <Text
                                  code
                                >{`${selectedSKU.model_provider}/${selectedSKU.provider_model_id?.trim() || selectedSKU.model_name}`}</Text>
                              </li>
                              {(Number(selectedSKU.provider_input_rate) > 0 ||
                                Number(selectedSKU.provider_output_rate) > 0) && (
                                <>
                                  <li>
                                    参考输入成本（元/1K tokens）：
                                    <Text strong>
                                      {Number(selectedSKU.provider_input_rate) > 0
                                        ? `¥${Number(selectedSKU.provider_input_rate).toFixed(6)}`
                                        : '—'}
                                    </Text>
                                  </li>
                                  <li>
                                    参考输出成本（元/1K tokens）：
                                    <Text strong>
                                      {Number(selectedSKU.provider_output_rate) > 0
                                        ? `¥${Number(selectedSKU.provider_output_rate).toFixed(6)}`
                                        : '—'}
                                    </Text>
                                  </li>
                                </>
                              )}
                            </ul>
                          </>
                        ),
                      },
                    ]}
                  />
                </>
              )}

              {skus.length > 1 && (
                <>
                  <Divider />
                  <Title level={5}>同系列规格对比</Title>
                  <Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
                    请在上方「选择套餐」中切换规格；本表仅作对比参考，数据与套餐卡片一致。
                  </Text>
                  <Table<SKUWithSPU>
                    size="small"
                    pagination={false}
                    scroll={{ x: screens.xs ? 520 : undefined }}
                    rowKey="id"
                    dataSource={skus}
                    columns={[
                      {
                        title: '套餐',
                        key: 'spu_name',
                        ellipsis: true,
                        render: (_: unknown, r: SKUWithSPU) => (
                          <Tooltip title={r.spu_name}>
                            <Space direction="vertical" size={0}>
                              <Text strong ellipsis style={{ maxWidth: 260 }}>
                                {getSkuCardSubtitle(r) || r.spu_name}
                              </Text>
                              <Text
                                type="secondary"
                                ellipsis
                                style={{ maxWidth: 260, fontSize: 12 }}
                              >
                                {r.sku_code}
                              </Text>
                            </Space>
                          </Tooltip>
                        ),
                      },
                      {
                        title: '类型',
                        key: 'sku_type',
                        width: 88,
                        render: (_: unknown, r: SKUWithSPU) => {
                          const map: Record<string, string> = {
                            token_pack: 'Token包',
                            subscription: '订阅',
                            concurrent: '并发',
                            trial: '试用',
                          };
                          return map[r.sku_type] || r.sku_type;
                        },
                      },
                      {
                        title: '套餐全名',
                        key: 'spu_full',
                        ellipsis: true,
                        render: (_: unknown, r: SKUWithSPU) => (
                          <Text type="secondary" style={{ fontSize: 12 }}>
                            {r.spu_name}
                          </Text>
                        ),
                      },
                      {
                        title: '价格',
                        key: 'price',
                        width: 96,
                        render: (_: unknown, r: SKUWithSPU) => (
                          <Text strong>¥{r.retail_price.toFixed(2)}</Text>
                        ),
                      },
                      {
                        title: '拼团',
                        key: 'group',
                        width: 88,
                        render: (_: unknown, r: SKUWithSPU) =>
                          r.group_enabled ? (
                            <Tag color="red">
                              {r.min_group_size}-{r.max_group_size}人
                            </Tag>
                          ) : (
                            <Text type="secondary">—</Text>
                          ),
                      },
                    ]}
                  />
                </>
              )}

              <Divider />

              <Title level={5}>常见问题</Title>
              <Collapse
                bordered={false}
                defaultActiveKey={[]}
                items={[
                  {
                    key: '1',
                    label: (
                      <Text strong>
                        <Tooltip title="按量计费的模型调用额度单位">Token</Tooltip> 有效期多久？
                      </Text>
                    ),
                    children: (
                      <Text type="secondary">
                        套餐标注的有效期内可用；具体以订单与商品说明为准。
                      </Text>
                    ),
                  },
                  {
                    key: '2',
                    label: (
                      <Text strong>
                        <Tooltip title="多人成团享受团购价">拼团</Tooltip> 失败怎么办？
                      </Text>
                    ),
                    children: (
                      <Text type="secondary">
                        拼团未成团时，请以订单状态与平台规则为准；可联系客服或选择单独购买。
                      </Text>
                    ),
                  },
                  {
                    key: '3',
                    label: <Text strong>额度可以转让吗？</Text>,
                    children: (
                      <Text type="secondary">
                        数字商品一般绑定账户使用，不支持转让；详见服务条款。
                      </Text>
                    ),
                  },
                ]}
              />
            </Space>
          </TabPane>

          <TabPane tab={`用户评价 (${reviewCount})`} key="reviews">
            <Space direction="vertical" style={{ width: '100%' }} size="large">
              {hasRating && rating != null ? (
                <Space wrap>
                  <Statistic title="综合评分" value={rating} suffix="/ 5.0" precision={1} />
                  <Rate disabled value={rating} allowHalf />
                </Space>
              ) : (
                <Text type="secondary">暂无综合评分数据（以商品与目录汇总为准）。</Text>
              )}

              <Divider />

              <Empty description="暂无用户评价，接口接入后将在此展示真实评价。" />
            </Space>
          </TabPane>
        </Tabs>

        <Modal
          title="加入现有拼团"
          open={showGroupsModal}
          onCancel={() => setShowGroupsModal(false)}
          footer={null}
          width={500}
        >
          <Spin spinning={groupsLoading}>
            {activeGroups.length === 0 ? (
              <Empty description="暂无进行中的拼团" style={{ margin: '40px 0' }}>
                <Button type="primary" onClick={() => setShowGroupsModal(false)}>
                  发起新拼团
                </Button>
              </Empty>
            ) : (
              <List
                itemLayout="horizontal"
                dataSource={activeGroups}
                renderItem={(group) => {
                  void groupModalTick;
                  const remainingSlots = group.target_count - group.current_count;
                  const deadline = new Date(group.deadline);
                  const pct = Math.round((100 * group.current_count) / group.target_count);

                  return (
                    <List.Item
                      actions={[
                        <Button
                          type="primary"
                          key="join"
                          loading={joiningGroupId === group.id}
                          onClick={() => handleJoinGroup(group)}
                        >
                          立即参团 (差{remainingSlots}人)
                        </Button>,
                      ]}
                    >
                      <List.Item.Meta
                        avatar={
                          <Avatar.Group maxCount={4} size="small">
                            {Array.from({ length: group.current_count }).map((_, i) => (
                              <Avatar
                                key={i}
                                style={{ backgroundColor: '#1890ff' }}
                                icon={<UserOutlined />}
                              />
                            ))}
                            {Array.from({ length: remainingSlots }).map((_, i) => (
                              <Avatar
                                key={`e-${i}`}
                                style={{ backgroundColor: '#f0f0f0', color: '#bbb' }}
                              >
                                +
                              </Avatar>
                            ))}
                          </Avatar.Group>
                        }
                        title={
                          <Space>
                            <Text strong>{group.target_count}人团</Text>
                            <Tag color="green">进行中</Tag>
                          </Space>
                        }
                        description={
                          <Space direction="vertical" size="small" style={{ width: '100%' }}>
                            <Progress percent={pct} size="small" status="active" />
                            <Text type="secondary">
                              已拼 {group.current_count}/{group.target_count}，成团价每人再省 ¥
                              {(
                                (selectedSKU?.retail_price || product.price) -
                                (selectedGroupPrice?.price_per_person ||
                                  groupPrices[0]?.price_per_person ||
                                  0)
                              ).toFixed(2)}
                            </Text>
                            <Text type="secondary">
                              <ClockCircleOutlined /> 剩余 {formatCountdown(deadline)}
                            </Text>
                          </Space>
                        }
                      />
                    </List.Item>
                  );
                }}
              />
            )}
          </Spin>
        </Modal>
        <Drawer
          title="拼团说明"
          placement="right"
          width={400}
          onClose={() => setDetailDrawer(null)}
          open={detailDrawer === 'group'}
        >
          <Paragraph>
            当前所选套餐规格未开启拼团或未配置成团档位时，无法发起或加入拼团。请在上方「选择套餐」中切换到支持拼团的规格，或改用单独购买。
          </Paragraph>
        </Drawer>
      </Card>
    </div>
  );
};

export default ProductDetailPage;
