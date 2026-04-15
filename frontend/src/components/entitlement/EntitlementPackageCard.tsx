import { Button, Card, Space, Tag, Typography } from 'antd';
import { LinkOutlined } from '@ant-design/icons';
import type { EntitlementPackage } from '@/types/entitlementPackage';
import dayjs from 'dayjs';

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
          {pkg.badge_text_secondary ? (
            <Tag color="cyan">{pkg.badge_text_secondary}</Tag>
          ) : (
            <Tag color="blue">权益包</Tag>
          )}
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
          {canBuy ? '一键组合下单' : '暂不可购买'}
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
      <Paragraph style={{ marginBottom: 8 }}>
        组合总价：<Text strong>¥{pkg.totalPrice.toFixed(2)}</Text>
      </Paragraph>
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
      <Space wrap style={{ marginBottom: 4 }}>
        {(pkg.items || []).map((s) => (
          <Tag key={s.id} color={s.line_purchasable === false ? 'error' : 'green'}>
            {s.spu_name} / {s.sku_code} ×{s.default_quantity}
            {s.line_issue ? `（${s.line_issue}）` : ''}
          </Tag>
        ))}
      </Space>
    </Card>
  );
}
