import { useEffect, useMemo, useState } from 'react';
import { List, Space, Spin, Typography, message, Segmented } from 'antd';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useOrderStore } from '@/stores/orderStore';
import { entitlementPackageService } from '@/services/entitlementPackage';
import type { EntitlementPackage } from '@/types/entitlementPackage';
import type { PackageSocialStats } from '@/components/entitlement/PackageSocialBar';
import { ENTITLEMENT_PACKAGE_FILTER_OPTIONS } from '@/types/entitlementPackage';
import { EntitlementPackageCard } from '@/components/entitlement/EntitlementPackageCard';
import { getApiErrorMessage } from '@/utils/apiError';
import styles from './EntitlementPackagesPage.module.css';

const { Title, Paragraph } = Typography;

export default function EntitlementPackagesPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { createOrder } = useOrderStore();
  const [loading, setLoading] = useState(true);
  const [submittingID, setSubmittingID] = useState<string>('');
  const [packages, setPackages] = useState<EntitlementPackage[]>([]);
  const [socialById, setSocialById] = useState<Record<number, PackageSocialStats>>({});
  const [category, setCategory] = useState<string>('all');

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setLoading(true);
      try {
        const res = await entitlementPackageService.listPublic();
        if (!cancelled) setPackages(res.data.data || []);
      } catch {
        if (!cancelled) setPackages([]);
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (packages.length === 0) return;
    let cancelled = false;
    const ids = packages.map((p) => p.id);
    (async () => {
      try {
        const res = await entitlementPackageService.batchStats(ids);
        const rows = res.data?.data || [];
        if (cancelled) return;
        const next: Record<number, PackageSocialStats> = {};
        for (const r of rows) {
          next[r.package_id] = {
            favoriteCount: r.favorite_count,
            likeCount: r.like_count,
            salesCount: r.sales_count,
            reviewCount: r.review_count,
            userFavorited: r.user_favorited,
            userLiked: r.user_liked,
            userReviewed: r.user_reviewed,
          };
        }
        setSocialById(next);
      } catch {
        if (!cancelled) setSocialById({});
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [packages]);

  const highlightCode = searchParams.get('pkg') || searchParams.get('highlight') || '';

  const packageView = useMemo(
    () =>
      packages.map((pkg) => {
        const totalPrice = (pkg.items || []).reduce(
          (sum, s) => sum + Number(s.retail_price || 0) * Number(s.default_quantity || 1),
          0
        );
        return { ...pkg, totalPrice };
      }),
    [packages]
  );

  const filtered = useMemo(() => {
    if (category === 'all') return packageView;
    return packageView.filter((p) => (p.category_code || 'general') === category);
  }, [packageView, category]);

  useEffect(() => {
    if (!highlightCode || filtered.length === 0) return;
    const t = window.setTimeout(() => {
      const el = document.getElementById(`pkg-card-${highlightCode}`);
      el?.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }, 300);
    return () => window.clearTimeout(t);
  }, [highlightCode, filtered]);

  const handleOneClickOrder = async (pkgID: string, pkg: EntitlementPackage) => {
    if (pkg.purchasable === false) {
      message.warning(pkg.unavailable_reason || '当前套餐包暂不可购买');
      return;
    }
    const items = (pkg.items || []).map((s) => ({
      sku_id: s.sku_id,
      quantity: s.default_quantity || 1,
    }));
    if (items.length === 0) {
      message.warning('当前套餐暂不可购买，请联系运营配置商品明细。');
      return;
    }
    setSubmittingID(pkgID);
    try {
      const orderID = await createOrder(items, { entitlement_package_id: pkg.id });
      if (!orderID) {
        message.success('套餐订单已创建');
        navigate('/orders');
        return;
      }
      message.success('套餐订单已创建，正在跳转支付');
      navigate(`/payment/${orderID}`);
    } catch (e) {
      message.error(getApiErrorMessage(e));
    } finally {
      setSubmittingID('');
    }
  };

  const copyShareLink = (code: string) => {
    const url = `${window.location.origin}/packages?pkg=${encodeURIComponent(code)}`;
    void navigator.clipboard.writeText(url).then(
      () => message.success('链接已复制'),
      () => message.error('复制失败')
    );
  };

  return (
    <div className={styles.page}>
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <div>
          <Title level={3} style={{ marginBottom: 8 }}>
            套餐包
          </Title>
          <Paragraph type="secondary" style={{ marginBottom: 0 }}>
            一次购买，组合履约（订阅、Token 等按套餐约定发放）。
          </Paragraph>
        </div>
        <div className={styles.toolbar}>
          <Segmented
            value={category}
            onChange={(v) => setCategory(String(v))}
            options={ENTITLEMENT_PACKAGE_FILTER_OPTIONS.map((o) => ({ label: o.label, value: o.value }))}
            style={{ flex: 1, maxWidth: '100%' }}
          />
        </div>
        <Paragraph type="secondary" style={{ marginBottom: 0, fontSize: 12 }}>
          套餐包为组合一口价下单；若需单品拼团，请在支持拼团的商品详情页发起。
        </Paragraph>
        <Spin spinning={loading}>
          <List
            grid={{ gutter: 16, xs: 1, sm: 1, md: 2 }}
            dataSource={filtered}
            locale={{ emptyText: '该分类下暂无套餐包' }}
            renderItem={(pkg) => (
              <List.Item>
                <div id={`pkg-card-${pkg.package_code}`}>
                  <EntitlementPackageCard
                    pkg={pkg}
                    loading={submittingID === String(pkg.id)}
                    onBuy={() => handleOneClickOrder(String(pkg.id), pkg)}
                    onCopyShareLink={() => copyShareLink(pkg.package_code)}
                    socialStats={socialById[pkg.id]}
                    onSocialPatch={(patch) =>
                      setSocialById((prev) => ({
                        ...prev,
                        [pkg.id]: { ...prev[pkg.id], ...patch },
                      }))
                    }
                  />
                </div>
              </List.Item>
            )}
          />
        </Spin>
      </Space>
    </div>
  );
}
