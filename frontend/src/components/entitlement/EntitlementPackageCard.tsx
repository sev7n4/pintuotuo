import { Button, Card, Space, Tag, Typography } from 'antd';
import { LinkOutlined } from '@ant-design/icons';
import type { EntitlementPackage } from '@/types/entitlementPackage';
import dayjs from 'dayjs';
import { PackageIncludeSummary } from './PackageIncludeSummary';
import { PackageItemsCollapse } from './PackageItemsCollapse';
import styles from './EntitlementPackageCard.module.css';

const { Paragraph, Text } = Typography;

type Props = {
  pkg: EntitlementPackage & { totalPrice: number };
  loading?: boolean;
  onBuy: () => void;
  onCopyShareLink?: () => void;
};

export function EntitlementPackageCard({ pkg, loading, onBuy, onCopyShareLink }: Props) {
  const canBuy = pkg.purchasable !== false;

  return (
    <Card
      title={pkg.name}
      className="entitlement-package-card"
      extra={
        <Space wrap size={4}>
          {pkg.is_featured ? <Tag color="gold">推荐</Tag> : null}
          {pkg.badge_text ? <Tag color="purple">{pkg.badge_text}</Tag> : null}
          {pkg.badge_text_secondary ? <Tag color="cyan">{pkg.badge_text_secondary}</Tag> : null}
        </Space>
      }
      actions={[
        <Button
          key="buy"
          type="primary"
          block
          loading={loading}
          disabled={!canBuy}
          onClick={onBuy}
        >
          {canBuy ? '一键购买' : '暂不可购买'}
        </Button>,
      ]}
    >
      {pkg.marketing_line ? (
        <Paragraph style={{ marginBottom: 8 }}>{pkg.marketing_line}</Paragraph>
      ) : null}
      <Paragraph type="secondary">{pkg.description}</Paragraph>
      {(pkg.start_at || pkg.end_at) && (
        <Paragraph type="secondary" style={{ marginBottom: 8 }}>
          有效期：
          {pkg.start_at ? dayjs(pkg.start_at).format('YYYY-MM-DD HH:mm') : '不限'} ~{' '}
          {pkg.end_at ? dayjs(pkg.end_at).format('YYYY-MM-DD HH:mm') : '不限'}
        </Paragraph>
      )}
      {pkg.promo_label && (
        <Paragraph style={{ marginBottom: 8 }}>
          <Tag color="magenta">{pkg.promo_label}</Tag>
          {pkg.promo_ends_at ? (
            <Text type="secondary"> 截止 {dayjs(pkg.promo_ends_at).format('MM-DD HH:mm')}</Text>
          ) : null}
        </Paragraph>
      )}
      <PackageIncludeSummary items={pkg.items || []} />
      <div className={styles.totalPriceBlock}>
        <div className={styles.totalPriceLabel}>组合总价</div>
        <div className={styles.totalPriceValue}>¥{pkg.totalPrice.toFixed(2)}</div>
      </div>
      {!canBuy && pkg.unavailable_reason ? (
        <Paragraph type="danger" style={{ marginBottom: 8 }}>
          {pkg.unavailable_reason}
        </Paragraph>
      ) : null}
      {onCopyShareLink ? (
        <Paragraph style={{ marginBottom: 8 }}>
          <Button type="link" size="small" icon={<LinkOutlined />} onClick={onCopyShareLink}>
            复制推广链接
          </Button>
        </Paragraph>
      ) : null}
      <PackageItemsCollapse items={pkg.items || []} mode="shop" />
    </Card>
  );
}
