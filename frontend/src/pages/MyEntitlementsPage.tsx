import { useEffect, useState } from 'react';
import { Alert, Card, Col, Empty, List, Row, Space, Spin, Statistic, Tag, Typography } from 'antd';
import { WalletOutlined, ApiOutlined, SafetyCertificateOutlined } from '@ant-design/icons';
import { skuService } from '@/services/sku';
import { tokenService } from '@/services/token';
import { entitlementPackageService } from '@/services/entitlementPackage';
import type { APIUsageGuideResponse } from '@/types';
import type { UserSubscriptionWithSKU } from '@/types/sku';
import type { EntitlementPackageUserView } from '@/types/entitlementPackage';

const { Title, Text, Paragraph } = Typography;

type TokenBalanceResp = {
  balance: number;
  total_earned: number;
  total_used: number;
};

export default function MyEntitlementsPage() {
  const [loading, setLoading] = useState(true);
  const [subs, setSubs] = useState<UserSubscriptionWithSKU[]>([]);
  const [packages, setPackages] = useState<EntitlementPackageUserView[]>([]);
  const [usageGuide, setUsageGuide] = useState<APIUsageGuideResponse | null>(null);
  const [balance, setBalance] = useState<TokenBalanceResp | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setLoading(true);
      try {
        const [subRes, guideRes, balRes, pkgRes] = await Promise.all([
          skuService.getUserSubscriptions(),
          tokenService.getAPIUsageGuide(),
          tokenService.getBalance(),
          entitlementPackageService.listMine(),
        ]);
        if (cancelled) return;
        setSubs(subRes.data.data || []);
        setUsageGuide(guideRes.data.data || null);
        setPackages(pkgRes.data.data || []);
        setBalance({
          balance: Number(balRes.data?.balance || 0),
          total_earned: Number(balRes.data?.total_earned || 0),
          total_used: Number(balRes.data?.total_used || 0),
        });
      } catch {
        if (!cancelled) {
          setSubs([]);
          setUsageGuide(null);
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div style={{ maxWidth: 1100, margin: '0 auto', padding: 16 }}>
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <div>
          <Title level={3} style={{ marginBottom: 8 }}>
            我的权益
          </Title>
          <Paragraph type="secondary" style={{ marginBottom: 0 }}>
            统一查看可用模型权益、订阅有效期与可扣费余额。
          </Paragraph>
        </div>

        <Row gutter={[16, 16]}>
          <Col xs={24} md={8}>
            <Card>
              <Statistic
                title="当前余额"
                value={balance?.balance || 0}
                suffix="Token"
                prefix={<WalletOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} md={8}>
            <Card>
              <Statistic title="累计获得" value={balance?.total_earned || 0} suffix="Token" />
            </Card>
          </Col>
          <Col xs={24} md={8}>
            <Card>
              <Statistic
                title="有效订阅"
                value={subs.length}
                suffix="项"
                prefix={<SafetyCertificateOutlined />}
              />
            </Card>
          </Col>
        </Row>

        <Alert
          type="info"
          showIcon
          message="计费提醒"
          description="即便已订阅模型，调用时仍会按 token 用量扣减余额；若订阅 SKU 配置了赠送 token，会在支付后自动入账。"
        />

        <Spin spinning={loading}>
          <Card title="权益包状态">
            {packages.length === 0 ? (
              <Empty description="暂无权益包配置" />
            ) : (
              <List
                dataSource={packages}
                renderItem={(p) => (
                  <List.Item>
                    <Space direction="vertical" size={2}>
                      <Text strong>{p.name}</Text>
                      <Text type="secondary">
                        覆盖进度：{p.covered_items}/{p.total_items}
                      </Text>
                      <Space wrap>
                        {(p.items || []).map((it) => (
                          <Tag key={`${p.id}-${it.id}`} color={p.is_active ? 'green' : 'default'}>
                            {it.spu_name}
                          </Tag>
                        ))}
                      </Space>
                    </Space>
                    <Tag color={p.is_active ? 'success' : 'warning'}>
                      {p.is_active ? '已激活' : '未覆盖完整'}
                    </Tag>
                  </List.Item>
                )}
              />
            )}
          </Card>

          <Card title="订阅权益">
            {subs.length === 0 ? (
              <Empty description="暂无有效订阅" />
            ) : (
              <List
                dataSource={subs}
                renderItem={(s) => (
                  <List.Item>
                    <Space direction="vertical" size={2}>
                      <Text strong>{s.spu_name}</Text>
                      <Text type="secondary">
                        SKU: {s.sku_code} · 到期：{new Date(s.end_date).toLocaleDateString('zh-CN')}
                      </Text>
                    </Space>
                    <Tag color="green">active</Tag>
                  </List.Item>
                )}
              />
            )}
          </Card>

          <Card
            title={
              <span>
                <ApiOutlined /> 模型调用权益
              </span>
            }
          >
            {(usageGuide?.items?.length || 0) === 0 ? (
              <Empty description="暂无可调用模型，请先购买权益包或订阅" />
            ) : (
              <List
                dataSource={usageGuide?.items || []}
                renderItem={(it) => (
                  <List.Item>
                    <Space direction="vertical" size={2}>
                      <Text>
                        <Tag color="blue">{it.provider_code}</Tag>
                        <Text code>{it.provider_slash_example || it.model_example}</Text>
                      </Text>
                      <Text type="secondary">
                        来源：{it.source === 'subscription' ? '订阅' : '订单'} {it.spu_name || ''}
                      </Text>
                    </Space>
                  </List.Item>
                )}
              />
            )}
          </Card>
        </Spin>
      </Space>
    </div>
  );
}
