import { Typography } from 'antd';
import { buildPackageIncludeBullets } from './entitlementItemDisplay';
import styles from './EntitlementPackageCard.module.css';

import type { EntitlementPackageItem } from '@/types/entitlementPackage';

const { Text } = Typography;

type Props = {
  items: EntitlementPackageItem[];
};

/**
 * 「套餐包含」：仅展示聚合后的四条规格（与下方折叠明细不重复）。
 */
export function PackageIncludeSummary({ items }: Props) {
  const bullets = buildPackageIncludeBullets(items ?? []);
  if (bullets.length === 0) return null;

  return (
    <div className={styles.includeInner}>
      <Text strong className={styles.includeTitle}>
        套餐包含
      </Text>
      <ul className={styles.includeBulletList}>
        {bullets.map((b) => (
          <li key={b.key} className={styles.includeBulletItem}>
            <Text>· {b.text}</Text>
          </li>
        ))}
      </ul>
    </div>
  );
}
