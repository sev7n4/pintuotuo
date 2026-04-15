import { Button, Card, Collapse, Space, Tag, Typography } from 'antd';
import { LinkOutlined } from '@ant-design/icons';
import type { EntitlementPackage, EntitlementPackageItem } from '@/types/entitlementPackage';
import dayjs from 'dayjs';
import styles from './EntitlementPackageCard.module.css';

const { Paragraph, Text } = Typography;

function skuTypeLabel(skuType: string): string {
  const m: Record<string, string> = {
    subscription: '订阅',
    token_pack: 'Token',
    trial: '试用',
    concurrent: '并发',
  };
  return m[skuType] || skuType || '—';
}

function lineDisplayName(it: EntitlementPackageItem): string {
  const d = it.display_name?.trim();
  if (d) return d;
  return it.spu_name || it.sku_code || '—';
}

function lineValueHint(it: EntitlementPackageItem): string {
  const note = it.value_note?.trim();
  if (note) return note;
  const unit = Number(it.retail_price || 0);
  const q = Number(it.default_quantity || 1);
  const sub = unit * q;
  return `单价 ¥${unit.toFixed(2)} × ${q} = ¥${sub.toFixed(2)}`;
}

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
      <Collapse
        bordered={false}
        className={styles.collapse}
        defaultActiveKey={[]}
        items={[
          {
            key: 'lines',
            label: (
              <Text type="secondary">
                包内明细（{pkg.items?.length ?? 0} 项）<Text strong>点击展开</Text>
              </Text>
            ),
            children: (
              <div>
                {(pkg.items || []).map((it) => (
                  <div key={it.id} className={styles.itemDetail}>
                    <div className={styles.itemRow}>
                      <div>
                        <Text strong>{lineDisplayName(it)}</Text>
                        <div className={styles.itemMeta}>
                          类型：{skuTypeLabel(it.sku_type)} · 数量：{it.default_quantity}
                          {it.sku_code ? (
                            <>
                              {' '}
                              · 编码 <Text code>{it.sku_code}</Text>
                            </>
                          ) : null}
                        </div>
                      </div>
                      {it.line_purchasable === false ? (
                        <Tag color="error">不可售</Tag>
                      ) : (
                        <Tag color="success">可售</Tag>
                      )}
                    </div>
                    <Paragraph type="secondary" style={{ margin: '8px 0 0', fontSize: 13 }}>
                      价值说明：{lineValueHint(it)}
                      {it.line_issue ? (
                        <Text type="danger"> （{it.line_issue}）</Text>
                      ) : null}
                    </Paragraph>
                  </div>
                ))}
              </div>
            ),
          },
        ]}
      />
    </Card>
  );
}
