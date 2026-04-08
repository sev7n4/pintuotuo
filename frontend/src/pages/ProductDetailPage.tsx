import React, { useCallback, useEffect, useState } from 'react';
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
  Badge,
  Alert,
  Progress,
  Collapse,
  Tooltip,
  Table,
  Grid,
} from 'antd';
import {
  ShoppingCartOutlined,
  ArrowLeftOutlined,
  ShareAltOutlined,
  TeamOutlined,
  UserOutlined,
  ClockCircleOutlined,
  TagsOutlined,
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { useProductStore } from '@stores/productStore';
import { useCartStore } from '@stores/cartStore';
import { useGroupStore } from '@stores/groupStore';
import { productService } from '@services/product';
import { skuService } from '@services/sku';
import type { Product, GroupPrice, Group } from '@/types';
import type { SKUWithSPU } from '@/types/sku';
import { normalizeGroupDiscountRate } from '@/utils/groupDiscount';

const { Title, Text } = Typography;
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
  const [skus, setSKUs] = useState<SKUWithSPU[]>([]);
  const [selectedSKU, setSelectedSKU] = useState<SKUWithSPU | null>(null);
  const [skuLoading, setSkuLoading] = useState(false);
  const [groupModalTick, setGroupModalTick] = useState(0);

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
    addItem(product, quantity);
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

  const buildSkuSummary = (sku: SKUWithSPU) => {
    const parts: string[] = [];
    if (sku.token_amount) parts.push(`${sku.token_amount.toLocaleString()} tokens`);
    if (sku.subscription_period) {
      const periodMap: Record<string, string> = {
        monthly: '月度',
        quarterly: '季度',
        yearly: '年度',
      };
      parts.push(periodMap[sku.subscription_period] || sku.subscription_period);
    }
    if (sku.concurrent_requests) parts.push(`${sku.concurrent_requests} 并发`);
    if (sku.valid_days) parts.push(`${sku.valid_days}天有效期`);
    return parts.join(' · ');
  };

  const getPromotionHighlights = () => {
    const sku = selectedSKU;
    if (!sku) return [];
    const promoLabels: string[] = [];
    const dynamic = sku as SKUWithSPU & {
      promotion_labels?: string[];
      coupons?: Array<{ name?: string }>;
      new_user_offer?: string;
      full_reduction?: string;
    };
    if (Array.isArray(dynamic.promotion_labels)) promoLabels.push(...dynamic.promotion_labels);
    if (Array.isArray(dynamic.coupons)) {
      dynamic.coupons.forEach((coupon) => {
        if (coupon?.name) promoLabels.push(coupon.name);
      });
    }
    if (dynamic.new_user_offer) promoLabels.push(dynamic.new_user_offer);
    if (dynamic.full_reduction) promoLabels.push(dynamic.full_reduction);
    if (sku.original_price && sku.original_price > sku.retail_price) {
      promoLabels.push(`直降 ¥${(sku.original_price - sku.retail_price).toFixed(2)}`);
    }
    if (sku.group_enabled) {
      promoLabels.push(`支持${sku.min_group_size}-${sku.max_group_size}人拼团`);
    }
    return Array.from(new Set(promoLabels)).slice(0, 4);
  };

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

  const estimatedFinalPrice = () => {
    const skuPrice = selectedSKU?.retail_price || product.price || 0;
    if (!skuPrice) return 0;
    if (purchaseMode === 'group') {
      return effectiveGroupTier?.price_per_person ?? skuPrice;
    }
    return skuPrice;
  };

  const calculateDiscount = () => {
    const basePrice = selectedSKU?.retail_price || product.price;
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

  return (
    <div
      style={{
        padding: pagePadding,
        maxWidth: 900,
        margin: '0 auto',
        boxSizing: 'border-box',
      }}
    >
      <Space style={{ marginBottom: '20px' }}>
        <Button type="text" icon={<ArrowLeftOutlined />} onClick={() => navigate('/catalog')}>
          返回列表
        </Button>
      </Space>

      <Card>
        <Row gutter={[24, 24]}>
          <Col xs={24} md={12}>
            <div
              style={{
                background: '#f5f5f5',
                borderRadius: 8,
                height: screens.xs ? 220 : 300,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                marginBottom: 16,
                overflow: 'hidden',
                position: 'relative',
              }}
            >
              {coverSrc ? (
                <img
                  src={coverSrc}
                  alt=""
                  style={{ width: '100%', height: '100%', objectFit: 'cover' }}
                  loading="lazy"
                />
              ) : (
                <Text type="secondary" style={{ fontSize: 48 }}>
                  📦
                </Text>
              )}
            </div>
          </Col>

          <Col xs={24} md={12}>
            <Title level={3}>{product.name}</Title>
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

            <div style={{ marginBottom: 16 }}>
              <Text type="secondary">{product.description}</Text>
            </div>

            <Divider style={{ margin: '12px 0' }} />

            <div style={{ marginBottom: 16 }}>
              <Space direction="vertical" size="small">
                {product.token_count && (
                  <Text>
                    包含：<Text strong>{(product.token_count / 10000).toFixed(0)}万 Token</Text>
                  </Text>
                )}
                {product.models && product.models.length > 0 && (
                  <Text>
                    模型：<Text strong>{product.models.join(', ')}</Text>
                  </Text>
                )}
                {product.validity_period && (
                  <Text>
                    有效期：<Text strong>{product.validity_period}</Text>
                  </Text>
                )}
                {product.context_length && (
                  <Text>
                    上下文：<Text strong>{product.context_length}</Text>
                  </Text>
                )}
              </Space>
            </div>
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
                    style={{
                      border:
                        selectedSKU?.id === sku.id ? '2px solid #1890ff' : '1px solid #d9d9d9',
                      cursor: 'pointer',
                      background: selectedSKU?.id === sku.id ? '#e6f7ff' : 'white',
                    }}
                    onClick={() => setSelectedSKU(sku)}
                  >
                    <Row justify="space-between" align="middle">
                      <Col flex="auto">
                        <Space direction="vertical" size={0}>
                          <Space>
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
                            <Text strong>{sku.spu_name}</Text>
                            <Tag color="geekblue">{sku.sku_code}</Tag>
                          </Space>
                          <Space size="small">
                            <Text type="secondary" style={{ fontSize: 12 }}>
                              {buildSkuSummary(sku)}
                            </Text>
                          </Space>
                        </Space>
                      </Col>
                      <Col>
                        <Space direction="vertical" size={0} align="end">
                          <Text strong style={{ fontSize: 18, color: '#1890ff' }}>
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
                        </Space>
                      </Col>
                    </Row>
                  </Card>
                ))}
              </Space>
            </Spin>
          </div>
        )}

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
                    value={selectedSKU?.retail_price || product.price}
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
        </div>

        {purchaseMode === 'group' && (
          <div style={{ marginBottom: 24 }}>
            {groupPrices.length === 0 && (
              <Alert
                type="info"
                showIcon
                message="当前规格不支持拼团"
                description="请选择支持拼团的 SKU 规格，或使用单独购买。"
                style={{ marginBottom: 16 }}
              />
            )}
            {(selectedSKU?.is_promoted || selectedSKU?.group_enabled) && groupPrices.length > 0 && (
              <Alert
                type="warning"
                showIcon
                message="大促特惠 · 拼团省更多"
                description="以下人数为常用成团档位，具体以支付页与 SKU 配置为准。"
                style={{ marginBottom: 16 }}
              />
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

        <Divider />

        <Alert
          type="info"
          showIcon
          message="优惠权益"
          description={
            <Space direction="vertical" style={{ width: '100%' }} size={8}>
              <Space wrap>
                {getPromotionHighlights().length > 0 ? (
                  getPromotionHighlights().map((label) => (
                    <Tag key={label} color="magenta">
                      {label}
                    </Tag>
                  ))
                ) : (
                  <Text type="secondary">当前套餐暂无活动优惠，后续可关注商家券与拼团优惠。</Text>
                )}
              </Space>
              <Text strong>
                预计到手价：
                <span style={{ color: '#cf1322' }}>¥{estimatedFinalPrice().toFixed(2)}</span>
              </Text>
            </Space>
          }
          style={{ marginBottom: 24 }}
        />

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
            <Text type="secondary" style={{ marginLeft: 16 }}>
              库存: {product.stock} 件
            </Text>
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
              <Space style={{ flex: 1, display: 'flex', gap: 8 }} direction="vertical">
                <Card
                  size="small"
                  style={{ width: '100%', background: '#f6ffed', borderColor: '#b7eb8f' }}
                >
                  <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                    <Text>当前可加入团数：{activeGroups.length}</Text>
                    <Text type="success">
                      拼团每人最高省 ¥
                      {(
                        (selectedSKU?.retail_price || product.price) - estimatedFinalPrice()
                      ).toFixed(2)}
                    </Text>
                  </Space>
                </Card>
                <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                  <Badge count={activeGroups.length} size="small" offset={[5, 0]}>
                    <Button
                      type="primary"
                      size="large"
                      icon={<TeamOutlined />}
                      onClick={() => {
                        loadActiveGroups();
                        setShowGroupsModal(true);
                      }}
                      disabled={product.stock === 0}
                      style={{ background: '#52c41a', borderColor: '#52c41a' }}
                    >
                      立即加入团
                    </Button>
                  </Badge>
                  <Button
                    type="primary"
                    size="large"
                    icon={<TeamOutlined />}
                    onClick={handleGroupPurchase}
                    disabled={product.stock === 0 || groupPrices.length === 0}
                    style={{ flex: 1, background: '#1890ff', borderColor: '#1890ff' }}
                  >
                    {product.stock === 0
                      ? '暂无库存'
                      : groupPrices.length === 0
                        ? '当前规格不可拼团'
                        : '发起拼团并支付'}
                  </Button>
                </Space>
              </Space>
            )}
            <Button size="large" icon={<ShareAltOutlined />} onClick={handleShare}>
              分享
            </Button>
            <Button size="large" onClick={() => navigate('/cart')}>
              查看购物车
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
                <li>
                  包含Token数量和类型：{(product.token_count || 1000000).toLocaleString()} Token
                </li>
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

              {skus.length > 1 && (
                <>
                  <Divider />
                  <Title level={5}>同系列规格对比</Title>
                  <Table<SKUWithSPU>
                    size="small"
                    pagination={false}
                    scroll={{ x: screens.xs ? 520 : undefined }}
                    rowKey="id"
                    dataSource={skus}
                    columns={[
                      {
                        title: '套餐',
                        dataIndex: 'spu_name',
                        key: 'spu_name',
                        ellipsis: true,
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
                        title: '规格摘要',
                        key: 'spec',
                        render: (_: unknown, r: SKUWithSPU) => (
                          <Text type="secondary" style={{ fontSize: 12 }}>
                            {buildSkuSummary(r)}
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
                      {
                        title: '操作',
                        key: 'pick',
                        width: 80,
                        render: (_: unknown, r: SKUWithSPU) => (
                          <Button type="link" size="small" onClick={() => setSelectedSKU(r)}>
                            {selectedSKU?.id === r.id ? '当前' : '选择'}
                          </Button>
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
                <Text type="secondary">暂无综合评分数据（以商品与 SPU 汇总为准）。</Text>
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
      </Card>
    </div>
  );
};

export default ProductDetailPage;
