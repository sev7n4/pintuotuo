import { Typography } from 'antd';
import {
  AppstoreOutlined,
  CalendarOutlined,
  ExperimentOutlined,
  GiftOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import type { EntitlementPackageItem } from '@/types/entitlementPackage';
import {
  lineDisplayName,
  packageIncludeHeadline,
  packageItemSpecParts,
  packageModelComboSummary,
} from './entitlementItemDisplay';
import styles from './EntitlementPackageCard.module.css';

const { Text } = Typography;

function iconForSkuType(t: string) {
  switch (t) {
    case 'subscription':
      return <CalendarOutlined />;
    case 'token_pack':
      return <GiftOutlined />;
    case 'concurrent':
      return <ThunderboltOutlined />;
    case 'trial':
      return <ExperimentOutlined />;
    default:
      return <AppstoreOutlined />;
  }
}

type Props = {
  items: EntitlementPackageItem[];
};

/** 套餐包含：文案取自包内各 SKU 实际字段（类型、周期、Token、有效期、路由等） */
export function PackageIncludeSummary({ items }: Props) {
  const list = items ?? [];
  if (list.length === 0) return null;

  const headline = packageIncludeHeadline(list);
  const comboLine = packageModelComboSummary(list);

  return (
    <div className={styles.includeBlock}>
      <Text strong className={styles.includeTitle}>
        套餐包含
      </Text>
      <Text className={styles.includeCombo}>{comboLine}</Text>
      {headline ? (
        <Text type="secondary" className={styles.includeHeadline}>
          {headline}
        </Text>
      ) : null}
      <div className={styles.includeRows}>
        {list.map((it) => (
          <div key={it.id} className={styles.includeRow}>
            <span className={styles.includeIcon}>{iconForSkuType(it.sku_type)}</span>
            <div className={styles.includeText}>
              <div>
                <Text strong>
                  模型：{lineDisplayName(it)}
                </Text>
                {it.default_quantity > 1 ? (
                  <Text type="secondary"> ×{it.default_quantity}</Text>
                ) : null}
              </div>
              <Text type="secondary" className={styles.includeSpec}>
                {packageItemSpecParts(it).join(' · ')}
              </Text>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
