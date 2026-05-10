import { useEffect, useState } from 'react';
import { Alert, Button, Card, Space, Table, Typography } from 'antd';
import { Link } from 'react-router-dom';
import { tokenService } from '@/services/token';
import type { APIUsageGuideItem, APIUsageGuideResponse } from '@/types';
import { modelValueFromItem } from '@/utils/apiUsageGuideModel';

const { Title, Paragraph } = Typography;

export default function DeveloperModelsPage() {
  const [guide, setGuide] = useState<APIUsageGuideResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setLoading(true);
      try {
        const res = await tokenService.getAPIUsageGuide();
        const data = (res.data as { data?: APIUsageGuideResponse })?.data ?? null;
        if (!cancelled) setGuide(data);
      } catch {
        if (!cancelled) setGuide(null);
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const items = guide?.items ?? [];

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <div>
        <Title level={3} style={{ marginTop: 0 }}>
          模型与权益
        </Title>
        <Paragraph type="secondary">
          数据来自 <Typography.Text code>GET /tokens/api-usage-guide</Typography.Text>
          ，与当前登录账号的订阅/已支付订单权益一致。
        </Paragraph>
      </div>

      <Space wrap>
        <Link to="/packages">
          <Button type="primary">购买套餐</Button>
        </Link>
        <Link to="/my/entitlements">
          <Button>我的权益</Button>
        </Link>
        <Link to="/developer/quickstart">
          <Button type="link">去快速开始试调用</Button>
        </Link>
      </Space>

      {guide?.disclaimer ? (
        <Alert type="info" showIcon message={guide.disclaimer} />
      ) : null}

      <Card title="可用模型（请求体 model）" loading={loading}>
        <Table<APIUsageGuideItem>
          rowKey={(r, i) => `${r.provider_code}-${r.model_example}-${i}`}
          pagination={false}
          dataSource={items}
          locale={{ emptyText: '暂无权益对应的模型；请先购买套餐或权益。' }}
          columns={[
            { title: '厂商', dataIndex: 'provider_code', key: 'p' },
            {
              title: 'model 推荐值',
              key: 'm',
              render: (_, r) => <Typography.Text code>{modelValueFromItem(r)}</Typography.Text>,
            },
            {
              title: '来源',
              key: 's',
              render: (_, r) => (
                <span>
                  {r.source === 'subscription' ? '订阅' : '订单'}
                  {r.spu_name ? ` · ${r.spu_name}` : ''}
                  {r.sku_code ? ` / ${r.sku_code}` : ''}
                </span>
              ),
            },
          ]}
        />
        {guide?.default_model_example ? (
          <Paragraph style={{ marginTop: 12, marginBottom: 0 }}>
            默认示例 model：<Typography.Text code>{guide.default_model_example}</Typography.Text>
          </Paragraph>
        ) : null}
      </Card>
    </Space>
  );
}
