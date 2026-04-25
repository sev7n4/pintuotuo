import { useEffect, useState } from 'react';
import { Button, Card, Col, Row, Space, Statistic, Tag, Tooltip, Typography } from 'antd';
import { ApiOutlined, DownOutlined, UpOutlined } from '@ant-design/icons';
import type { BYOKSummaryResponse } from '@/services/adminMerchant';

const { Text } = Typography;

type Props = {
  data: BYOKSummaryResponse | null;
  loading?: boolean;
};

const STORAGE_KEY = 'admin_byok_summary_collapsed';

export default function AdminBYOKSummaryCard({ data, loading }: Props) {
  const [collapsed, setCollapsed] = useState(false);

  useEffect(() => {
    try {
      const saved = window.localStorage.getItem(STORAGE_KEY);
      if (saved === 'true') setCollapsed(true);
      if (saved === 'false') setCollapsed(false);
    } catch {
      // ignore storage errors
    }
  }, []);

  const toggleCollapsed = () => {
    setCollapsed((prev) => {
      const next = !prev;
      try {
        window.localStorage.setItem(STORAGE_KEY, String(next));
      } catch {
        // ignore storage errors
      }
      return next;
    });
  };
  const s = data?.summary;
  return (
    <Card
      size="small"
      loading={loading}
      style={{ marginBottom: 16 }}
      title={
        <Space>
          <ApiOutlined />
          <span>全平台 BYOK 概览</span>
          <Tooltip
            title={
              <div>
                <div>仅统计启用（active）密钥参与「可路由」判断。</div>
                <div>
                  绿：至少一把 active 满足 strict 白名单；黄：无一可路由且存在不健康或验证失败；
                </div>
                <div>灰：有启用 Key 但无一可路由且无上述坏信号。</div>
              </div>
            }
          >
            <Text type="secondary" style={{ cursor: 'help', fontSize: 12 }}>
              规则
            </Text>
          </Tooltip>
        </Space>
      }
      extra={
        <Button
          type="link"
          size="small"
          onClick={toggleCollapsed}
          icon={collapsed ? <DownOutlined /> : <UpOutlined />}
        >
          {collapsed ? '展开' : '收起'}
        </Button>
      }
    >
      {s && collapsed && (
        <Space wrap size="small">
          <Text type="secondary">启用密钥</Text>
          <Tag>{s.active_keys_total}</Tag>
          <Text type="secondary">可路由商户</Text>
          <Tag color="success">{s.merchants_has_routable}</Tag>
          <Text type="secondary">需关注商户</Text>
          <Tag color="error">{s.merchants_need_attention}</Tag>
          <Text type="secondary">无密钥商户</Text>
          <Tag>{s.merchants_with_no_keys}</Tag>
        </Space>
      )}
      {s && !collapsed && (
        <>
          <Row gutter={[16, 8]}>
            <Col xs={12} sm={6}>
              <Statistic title="启用密钥总数" value={s.active_keys_total} />
            </Col>
            <Col xs={12} sm={6}>
              <Statistic title="有启用密钥的商户" value={s.merchants_with_active_keys} />
            </Col>
            <Col xs={12} sm={6}>
              <Statistic title="存在可路由(strict)的商户" value={s.merchants_has_routable} />
            </Col>
            <Col xs={12} sm={6}>
              <Statistic title="需关注商户" value={s.merchants_need_attention} />
            </Col>
          </Row>
          <div style={{ marginTop: 8 }}>
            <Space wrap size="small">
              <Text type="secondary">无任何密钥记录的商户：</Text>
              <Tag>{s.merchants_with_no_keys}</Tag>
              <Tag color="success">绿=可路由</Tag>
              <Tag color="warning">黄=异常信号</Tag>
              <Tag>灰=未就绪</Tag>
            </Space>
          </div>
        </>
      )}
    </Card>
  );
}
