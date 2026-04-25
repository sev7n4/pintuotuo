import { Card, Col, Progress, Row, Space, Statistic, Tag, Tooltip, Typography } from 'antd';
import {
  ApiOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  KeyOutlined,
  RadarChartOutlined,
} from '@ant-design/icons';
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
  const activeRatio = agg.totalCount > 0 ? Math.round((agg.activeCount / agg.totalCount) * 100) : 0;
  const routableCount = agg.hasRoutable ? 1 : 0;
  const healthColor =
    agg.level === 'green' ? '#52c41a' : agg.level === 'yellow' ? '#faad14' : '#8c8c8c';

  return (
    <Card
      size="small"
      style={{
        marginBottom: 16,
        border: '1px solid #e6f4ff',
        background: 'linear-gradient(180deg, #fcfdff 0%, #f7fbff 100%)',
      }}
    >
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 12,
        }}
      >
        <Space>
          <ApiOutlined />
          <span style={{ fontWeight: 600 }}>秘钥链路概览（BYOK）</span>
          <Tag color={color}>{title}</Tag>
        </Space>
      </div>

      <Row gutter={[12, 12]}>
        <Col xs={24} sm={12} md={8}>
          <div
            style={{
              border: '1px solid #edf2ff',
              borderRadius: 10,
              background: '#fff',
              padding: 12,
              minHeight: 102,
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'space-between',
            }}
          >
            <Statistic
              title="启用秘钥条数"
              value={agg.activeCount}
              suffix={`/ ${agg.totalCount}`}
              prefix={<KeyOutlined />}
            />
            <Progress
              percent={activeRatio}
              showInfo={false}
              strokeColor={healthColor}
              size="small"
            />
          </div>
        </Col>
        <Col xs={24} sm={12} md={8}>
          <div
            style={{
              border: '1px solid #edf2ff',
              borderRadius: 10,
              background: '#fff',
              padding: 12,
              minHeight: 102,
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'space-between',
            }}
          >
            <Statistic
              title="Strict 可路由"
              value={routableCount}
              suffix={agg.hasRoutable ? '把（至少）' : '把'}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: agg.hasRoutable ? '#389e0d' : '#8c8c8c' }}
            />
            <Text type="secondary">满足验证 + 健康双条件</Text>
          </div>
        </Col>
        <Col xs={24} sm={24} md={8}>
          <div
            style={{
              border: '1px solid #edf2ff',
              borderRadius: 10,
              background: '#fff',
              padding: 12,
              minHeight: 102,
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'space-between',
            }}
          >
            <Statistic
              title="需立即关注"
              value={agg.needAttentionActive}
              suffix="把"
              prefix={<ExclamationCircleOutlined />}
              valueStyle={{ color: agg.needAttentionActive > 0 ? '#cf1322' : '#8c8c8c' }}
            />
            <Tooltip title="与表格「Strict 权益」列“未满足”口径一致。">
              <Text type="secondary">
                定位入口：列表筛选 + 立即探测 <RadarChartOutlined style={{ marginLeft: 4 }} />
              </Text>
            </Tooltip>
          </div>
        </Col>
      </Row>
    </Card>
  );
}
