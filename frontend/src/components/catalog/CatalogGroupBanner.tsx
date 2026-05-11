import { useEffect, useState } from 'react';
import { Button, Card, Space, Tag, Typography } from 'antd';
import { RightOutlined, ThunderboltOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import type { CatalogGroupBannerConfig } from './catalogGroupBannerConfig';
import { defaultCatalogGroupBanner } from './catalogGroupBannerConfig';
import styles from './CatalogGroupBanner.module.css';

const { Text, Title } = Typography;

interface CatalogGroupBannerProps {
  /** 当前是否为「仅拼团」卖场 */
  groupCatalog: boolean;
}

function resolveHref(href: string | undefined): string {
  if (!href) return '/catalog?group_enabled=true';
  if (href.startsWith('http://') || href.startsWith('https://')) return href;
  return href.startsWith('/') ? href : `/${href}`;
}

export function CatalogGroupBanner({ groupCatalog }: CatalogGroupBannerProps) {
  const navigate = useNavigate();
  const [cfg, setCfg] = useState<CatalogGroupBannerConfig | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const res = await fetch('/marketing/catalog-group-banner.json', { cache: 'no-store' });
        if (!res.ok) {
          if (!cancelled) setCfg(defaultCatalogGroupBanner());
          return;
        }
        const json = (await res.json()) as CatalogGroupBannerConfig;
        if (!cancelled) setCfg(json);
      } catch {
        if (!cancelled) setCfg(defaultCatalogGroupBanner());
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  if (!cfg?.enabled || !cfg.title) return null;
  const visible = cfg.show_when === 'always' || (cfg.show_when === 'group_only' && groupCatalog);
  if (!visible) return null;

  const bg = cfg.background || 'linear-gradient(100deg, #fff7e6 0%, #ffe7ba 50%, #f6ffed 100%)';

  return (
    <Card
      className={styles.bannerCard}
      style={{ marginBottom: 16, background: bg, borderColor: 'rgba(0, 0, 0, 0.06)' }}
    >
      <div className={styles.bannerInner}>
        <Space direction="vertical" size={4} className={styles.bannerText}>
          <Space size={8} wrap>
            {cfg.tag && (
              <Tag icon={<ThunderboltOutlined />} color="volcano">
                {cfg.tag}
              </Tag>
            )}
          </Space>
          <Title level={5} style={{ margin: 0 }}>
            {cfg.title}
          </Title>
          {cfg.subtitle && (
            <Text type="secondary" className={styles.subtitle}>
              {cfg.subtitle}
            </Text>
          )}
          {cfg.cta_label && (
            <Button
              type="primary"
              icon={<RightOutlined />}
              onClick={() => navigate(resolveHref(cfg.cta_href))}
              style={{ marginTop: 8, alignSelf: 'flex-start' }}
            >
              {cfg.cta_label}
            </Button>
          )}
        </Space>
        {cfg.image && (
          <div className={styles.bannerImgWrap}>
            <img src={cfg.image} alt="" className={styles.bannerImg} />
          </div>
        )}
      </div>
    </Card>
  );
}
