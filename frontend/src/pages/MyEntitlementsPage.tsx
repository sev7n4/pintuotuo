import { useEffect, useState } from 'react';
import { Card, Col, Empty, List, Row, Space, Skeleton, Spin, Statistic, Tag, Typography } from 'antd';
import { WalletOutlined, ApiOutlined, SafetyCertificateOutlined } from '@ant-design/icons';
import { skuService } from '@/services/sku';
import { tokenService } from '@/services/token';
import { entitlementPackageService } from '@/services/entitlementPackage';
import type { APIUsageGuideResponse } from '@/types';
import type { UserSubscriptionWithSKU } from '@/types/sku';
import type { EntitlementPackageUserView } from '@/types/entitlementPackage';
import { PackageItemsCollapse } from '@/components/entitlement/PackageItemsCollapse';

const { Title, Text, Paragraph } = Typography;

type TokenBalanceResp = {
  balance: number;
  total_earned: number;
  total_used: number;
};

export default function MyEntitlementsPage() {
  const [loadingMain, setLoadingMain] = useState(true);
  const [loadingExtras, setLoadingExtras] = useState(true);
  const [subs, setSubs] = useState<UserSubscriptionWithSKU[]>([]);
  const [packages, setPackages] = useState<EntitlementPackageUserView[]>([]);
  const [usageGuide, setUsageGuide] = useState<APIUsageGuideResponse | null>(null);
  const [balance, setBalance] = useState<TokenBalanceResp | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setLoadingMain(true);
      try {
        const [subRes, pkgRes] = await Promise.all([
          skuService.getUserSubscriptions(),
          entitlementPackageService.listMine(),
        ]);
        if (!cancelled) {
          setSubs(subRes.data.data || []);
          setPackages(pkgRes.data.data || []);
        }
      } catch {
        if (!cancelled) {
          setSubs([]);
          setPackages([]);
        }
      } finally {
        if (!cancelled) setLoadingMain(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setLoadingExtras(true);
      try {
        const [guideRes, balRes] = await Promise.all([
          tokenService.getAPIUsageGuide(),
          tokenService.getBalance(),
        ]);
        if (!cancelled) {
          setUsageGuide(guideRes.data.data || null);
          setBalance({
            balance: Number(balRes.data?.balance || 0),
            total_earned: Number(balRes.data?.total_earned || 0),
            total_used: Number(balRes.data?.total_used || 0),
          });
        }
      } catch {
        if (!cancelled) {
          setUsageGuide(null);
          setBalance(null);
        }
      } finally {
        if (!cancelled) setLoadingExtras(false);
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
            查看模型与套餐权益、订阅有效期及可扣费余额。
          </Paragraph>
        </div>

        <Row gutter={[16, 16]}>
          <Col xs={24} md={8}>
            <Card>
              <Skeleton loading={loadingExtras} active paragraph={{ rows: 1 }}>
                <Statistic
                  title="当前余额"
                  value={balance?.balance ?? 0}
                  suffix="Token"
                  prefix={<WalletOutlined />}
                />
                <Paragraph type="secondary" style={{ marginTop: 8, marginBottom: 0, fontSize: 12 }}>
                  支付成功后，若套餐含赠送 Token，将并入此处余额。
                </Paragraph>
              </Skeleton>
            </Card>
          </Col>
          <Col xs={24} md={8}>
            <Card>
              <Skeleton loading={loadingExtras} active paragraph={{ rows: 1 }}>
                <Statistic title="累计获得" value={balance?.total_earned ?? 0} suffix="Token" />
              </Skeleton>
            </Card>
          </Col>
          <Col xs={24} md={8}>
            <Card>
              <Spin spinning={loadingMain}>
                <Statistic
                  title="有效订阅"
                  value={subs.length}
                  suffix="项"
                  prefix={<SafetyCertificateOutlined />}
                />
              </Spin>
            </Card>
          </Col>
        </Row>

        <Spin spinning={loadingMain}>
          <Card title="套餐包状态">
            {packages.length === 0 ? (
              <Empty description="暂无套餐包配置" />
            ) : (
              <List
                dataSource={packages}
                renderItem={(p) => (
                  <List.Item>
                    <Space direction="vertical" size={8} style={{ width: '100%' }}>
                      <Space align="start" style={{ width: '100%', justifyContent: 'space-between' }}>
                        <Space direction="vertical" size={2}>
                          <Text strong>{p.name}</Text>
                          <Text type="secondary">
                            覆盖进度：{p.covered_items}/{p.total_items}
                          </Text>
                        </Space>
                        <Tag color={p.is_active ? 'success' : 'warning'}>
                          {p.is_active ? '已激活' : '未覆盖完整'}
                        </Tag>
                      </Space>
                      <PackageItemsCollapse items={p.items || []} mode="mine" />
                    </Space>
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
                        规格：{s.sku_code} · 到期：
                        {new Date(s.end_date).toLocaleDateString('zh-CN')}
                      </Text>
                    </Space>
                    <Tag color="green">生效中</Tag>
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
            <Skeleton loading={loadingExtras} active paragraph={{ rows: 3 }}>
              {(usageGuide?.items?.length || 0) === 0 ? (
                <Empty description="暂无可调用模型，请先购买套餐或订阅" />
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
                          来源：{it.source === 'subscription' ? '订阅' : '订单'}{' '}
                          {it.spu_name || ''}
                        </Text>
                      </Space>
                    </List.Item>
                  )}
                />
              )}
            </Skeleton>
          </Card>
        </Spin>
      </Space>
    </div>
  );
}
