import { Card, Space, Tag, Tooltip, Typography } from 'antd';
import { ApiOutlined } from '@ant-design/icons';
import type { MerchantAPIKey } from '@/types';
import { aggregateMerchantBYOK } from '@/utils/byokStatus';

const { Text } = Typography;

function levelLabel(level: string, activeCount: number, totalCount: number): string {
  if (totalCount === 0) return '无密钥';
  if (activeCount === 0) return '无启用密钥';
  switch (level) {
    case 'green':
      return '存在可路由 Key';
    case 'yellow':
      return '无可路由 · 有异常信号';
    case 'gray':
      return '无可路由 · 待探测/待验证';
    default:
      return level;
  }
}

function levelTagColor(level: string, activeCount: number, totalCount: number): string {
  if (totalCount === 0) return 'default';
  if (activeCount === 0) return 'default';
  if (level === 'green') return 'success';
  if (level === 'yellow') return 'warning';
  return 'default';
}

type Props = {
  apiKeys: MerchantAPIKey[];
};

export default function MerchantBYOKOverviewCard({ apiKeys }: Props) {
  const agg = aggregateMerchantBYOK(apiKeys);
  const title = levelLabel(agg.level, agg.activeCount, agg.totalCount);
  const color = levelTagColor(agg.level, agg.activeCount, agg.totalCount);

  return (
    <Card
      size="small"
      style={{ marginBottom: 16 }}
      title={
        <Space>
          <ApiOutlined />
          <span>秘钥链路概览（BYOK）</span>
          <Tag color={color}>{title}</Tag>
        </Space>
      }
    >
      <Space direction="vertical" size={4} style={{ width: '100%' }}>
        <Text type="secondary">
          启用密钥 {agg.activeCount} / 共 {agg.totalCount} 条；可进入 strict 白名单的启用 Key：{' '}
          {agg.hasRoutable ? '至少 1 把' : '无'}
        </Text>
        <Text type={agg.needAttentionActive > 0 ? 'danger' : 'secondary'}>
          需立即关注（启用但未满足 Strict 权益条件）：{agg.needAttentionActive} 把
          <Tooltip title="与表格「Strict 权益」列「未满足」口径一致。">
            <span style={{ marginLeft: 4, cursor: 'help', borderBottom: '1px dotted' }}>说明</span>
          </Tooltip>
        </Text>
      </Space>
    </Card>
  );
}
