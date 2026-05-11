import { useCallback, useEffect, useMemo, useState } from 'react';
import { Typography } from 'antd';
import { getProviderCardSurfaceStyle, getProviderLogoUrl } from '@/utils/providerBrand';
import styles from './ProductCoverMedia.module.css';

const { Text } = Typography;

export type ProductCoverVariant = 'grid' | 'wide' | 'home' | 'hero';

export interface ProductCoverMediaProps {
  /** 已解析的主图 URL（若传入则优先于 imageUrl/thumbnailUrl） */
  coverUrl?: string | null;
  imageUrl?: string | null;
  thumbnailUrl?: string | null;
  modelProvider?: string | null;
  /** 无图且无厂商标时的占位文案来源（取前两字） */
  fallbackTitle: string;
  /** 切换 SKU/商品时重置裂图状态 */
  resetKey?: string | number;
  variant?: ProductCoverVariant;
  className?: string;
}

function resolveCoverUrl(
  coverUrl?: string | null,
  imageUrl?: string | null,
  thumbnailUrl?: string | null
): string | undefined {
  const u = (coverUrl ?? imageUrl ?? thumbnailUrl)?.trim();
  return u || undefined;
}

/**
 * 卖场统一封面：主图（含 onError）→ Lobe 厂商徽标 → 灰底两字占位。
 * 用于首页、列表、详情、收藏、浏览历史等。
 */
export function ProductCoverMedia({
  coverUrl,
  imageUrl,
  thumbnailUrl,
  modelProvider,
  fallbackTitle,
  resetKey,
  variant = 'grid',
  className,
}: ProductCoverMediaProps) {
  const resolved = useMemo(
    () => resolveCoverUrl(coverUrl, imageUrl, thumbnailUrl),
    [coverUrl, imageUrl, thumbnailUrl]
  );

  const [coverBroken, setCoverBroken] = useState(false);
  const [brandBroken, setBrandBroken] = useState(false);

  useEffect(() => {
    setCoverBroken(false);
    setBrandBroken(false);
  }, [resetKey, resolved]);

  const showCoverImg = Boolean(resolved) && !coverBroken;
  const brandUrl =
    !showCoverImg && modelProvider ? getProviderLogoUrl(modelProvider) : null;
  const surface =
    !showCoverImg && modelProvider ? getProviderCardSurfaceStyle(modelProvider) : undefined;

  const onCoverError = useCallback(() => {
    setCoverBroken(true);
  }, []);

  const rootClass = [
    styles.root,
    variant === 'grid' && styles.rootGrid,
    variant === 'wide' && styles.rootWide,
    variant === 'home' && styles.rootHome,
    variant === 'hero' && styles.rootHero,
    className,
  ]
    .filter(Boolean)
    .join(' ');

  const badgeClass = [
    styles.productBrandBadge,
    variant === 'grid' && styles.badgeGrid,
    variant === 'wide' && styles.badgeWide,
    variant === 'hero' && styles.badgeHero,
  ]
    .filter(Boolean)
    .join(' ');

  const logoClass = [
    styles.productBrandLogo,
    variant === 'grid' && styles.logoGrid,
    variant === 'wide' && styles.logoWide,
    variant === 'hero' && styles.logoHero,
  ]
    .filter(Boolean)
    .join(' ');

  const placeholderBoxClass = [
    styles.placeholder,
    variant === 'grid' && styles.placeholderCompact,
    variant === 'hero' && styles.placeholderHero,
  ]
    .filter(Boolean)
    .join(' ');

  const t = fallbackTitle.trim();
  const fallbackChars = t.length >= 2 ? t.slice(0, 2) : t || '—';

  return (
    <div className={rootClass} style={surface ? { background: surface } : undefined}>
      {showCoverImg ? (
        <img
          src={resolved}
          alt=""
          className={styles.coverImg}
          loading="lazy"
          decoding="async"
          onError={onCoverError}
        />
      ) : brandUrl && !brandBroken ? (
        <div className={badgeClass} aria-hidden>
          <img
            src={brandUrl}
            alt=""
            className={logoClass}
            loading="lazy"
            decoding="async"
            onError={() => setBrandBroken(true)}
          />
        </div>
      ) : (
        <div className={placeholderBoxClass}>
          <Text type="secondary" className={styles.placeholderText}>
            {fallbackChars}
          </Text>
        </div>
      )}
    </div>
  );
}
