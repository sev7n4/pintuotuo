import { Collapse, Space, Tag, Typography } from 'antd';
import { UnorderedListOutlined } from '@ant-design/icons';
import type { EntitlementPackageItem } from '@/types/entitlementPackage';
import { lineDisplayName, lineValueHint, skuTypeLabel } from './entitlementItemDisplay';
import styles from './EntitlementPackageCard.module.css';

const { Paragraph, Text } = Typography;

type Mode = 'shop' | 'mine';

type Props = {
  items: EntitlementPackageItem[];
  mode: Mode;
};

function rightTag(it: EntitlementPackageItem, mode: Mode) {
  if (mode === 'mine') {
    if (it.line_covered === true) return <Tag color="success">已具备</Tag>;
    if (it.line_covered === false) return <Tag>待开通</Tag>;
    return null;
  }
  if (it.line_purchasable === false) return <Tag color="error">不可售</Tag>;
  return <Tag color="success">可售</Tag>;
}

export function PackageItemsCollapse({ items, mode }: Props) {
  const n = items?.length ?? 0;
  if (n === 0) return null;

  return (
    <Collapse
      bordered={false}
      className={styles.collapse}
      defaultActiveKey={[]}
      items={[
        {
          key: 'lines',
          label: (
            <Space size={6}>
              <UnorderedListOutlined />
              <Text type="secondary">
                包内模型明细（{n} 项）<Text strong>点击展开</Text>
              </Text>
            </Space>
          ),
          children: (
            <div>
              {items.map((it) => (
                <div key={it.id} className={styles.itemDetail}>
                  <div className={styles.itemRow}>
                    <div>
                      <Text strong>{lineDisplayName(it)}</Text>
                      <div className={styles.itemMeta}>
                        类型：{skuTypeLabel(it.sku_type)} · 数量：{it.default_quantity}
                        {it.sku_code ? (
                          <>
                            {' '}
                            · 规格 <Text code>{it.sku_code}</Text>
                          </>
                        ) : null}
                      </div>
                    </div>
                    {rightTag(it, mode)}
                  </div>
                  <Paragraph type="secondary" style={{ margin: '8px 0 0', fontSize: 13 }}>
                    价值说明：{lineValueHint(it)}
                    {it.line_issue && mode === 'shop' ? (
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
  );
}
