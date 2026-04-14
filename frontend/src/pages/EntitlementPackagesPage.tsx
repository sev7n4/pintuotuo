import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, List, Space, Spin, Tag, Typography, message } from 'antd';
import { useNavigate } from 'react-router-dom';
import { useOrderStore } from '@/stores/orderStore';
import { entitlementPackageService } from '@/services/entitlementPackage';
import type { EntitlementPackage } from '@/types/entitlementPackage';
import dayjs from 'dayjs';

const { Title, Paragraph, Text } = Typography;

export default function EntitlementPackagesPage() {
  const navigate = useNavigate();
  const { createOrder } = useOrderStore();
  const [loading, setLoading] = useState(true);
  const [submittingID, setSubmittingID] = useState<string>('');
  const [packages, setPackages] = useState<EntitlementPackage[]>([]);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setLoading(true);
      try {
        const res = await entitlementPackageService.listPublic();
        if (!cancelled) setPackages(res.data.data || []);
      } catch {
        if (!cancelled) setPackages([]);
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const packageView = useMemo(
    () =>
      packages.map((pkg) => {
        const totalPrice = (pkg.items || []).reduce(
          (sum, s) => sum + Number(s.retail_price || 0) * Number(s.default_quantity || 1),
          0
        );
        return { ...pkg, totalPrice };
      }),
    [packages]
  );

  const handleOneClickOrder = async (pkgID: string, pkg: EntitlementPackage) => {
    const items = (pkg.items || []).map((s) => ({
      sku_id: s.sku_id,
      quantity: s.default_quantity || 1,
    }));
    if (items.length === 0) {
      message.warning('当前权益包暂不可购买，请联系运营配置 SKU。');
      return;
    }
    setSubmittingID(pkgID);
    try {
      const orderID = await createOrder(items);
      if (!orderID) {
        message.success('权益包订单已创建');
        navigate('/orders');
        return;
      }
      message.success('权益包订单已创建，正在跳转支付');
      navigate(`/payment/${orderID}`);
    } catch {
      message.error('权益包下单失败，请稍后重试');
    } finally {
      setSubmittingID('');
    }
  };

  return (
    <div style={{ maxWidth: 1080, margin: '0 auto', padding: 16 }}>
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <div>
          <Title level={3} style={{ marginBottom: 8 }}>
            权益包
          </Title>
          <Paragraph type="secondary" style={{ marginBottom: 0 }}>
            将多个 SKU 作为一个“权益包”理解与购买，避免分散下单。
          </Paragraph>
        </div>
        <Alert
          type="info"
          showIcon
          message="权益包下单说明"
          description="点击「一键组合下单」会生成一个多明细订单，支付后按每个明细履约（订阅、Token赠送等）。"
        />
        <Spin spinning={loading}>
          <List
            grid={{ gutter: 16, xs: 1, sm: 1, md: 2 }}
            dataSource={packageView}
            renderItem={(pkg) => (
              <List.Item>
                <Card
                  title={pkg.name}
                  extra={
                    <Space>
                      {pkg.is_featured ? <Tag color="gold">推荐</Tag> : null}
                      {pkg.badge_text ? (
                        <Tag color="purple">{pkg.badge_text}</Tag>
                      ) : (
                        <Tag color="blue">权益包</Tag>
                      )}
                    </Space>
                  }
                  actions={[
                    <Button
                      key="buy"
                      type="primary"
                      loading={submittingID === String(pkg.id)}
                      onClick={() => handleOneClickOrder(String(pkg.id), pkg)}
                    >
                      一键组合下单
                    </Button>,
                  ]}
                >
                  <Paragraph type="secondary">{pkg.description}</Paragraph>
                  {(pkg.start_at || pkg.end_at) && (
                    <Paragraph type="secondary" style={{ marginBottom: 8 }}>
                      有效期：
                      {pkg.start_at
                        ? dayjs(pkg.start_at).format('YYYY-MM-DD HH:mm')
                        : '不限'} ~{' '}
                      {pkg.end_at ? dayjs(pkg.end_at).format('YYYY-MM-DD HH:mm') : '不限'}
                    </Paragraph>
                  )}
                  <Paragraph style={{ marginBottom: 8 }}>
                    组合总价：<Text strong>¥{pkg.totalPrice.toFixed(2)}</Text>
                  </Paragraph>
                  <Space wrap style={{ marginBottom: 8 }}>
                    {(pkg.items || []).map((s) => (
                      <Tag key={s.id} color="green">
                        {s.spu_name} / {s.sku_code} x{s.default_quantity}
                      </Tag>
                    ))}
                  </Space>
                </Card>
              </List.Item>
            )}
          />
        </Spin>
      </Space>
    </div>
  );
}
