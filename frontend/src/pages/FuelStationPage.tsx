import { useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  Col,
  Empty,
  Radio,
  Row,
  Space,
  Spin,
  Tag,
  Typography,
  message,
  Collapse,
} from 'antd';
import { RocketOutlined, ThunderboltOutlined, ShoppingCartOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { fuelStationService } from '@/services/fuelStation';
import { skuService } from '@/services/sku';
import type { FuelStationConfig } from '@/types/fuelStation';
import type { SKUWithSPU } from '@/types/sku';
import { useCartStore } from '@/stores/cartStore';
import { IconHintButton } from '@/components/IconHintButton';

const { Title, Paragraph, Text } = Typography;

export default function FuelStationPage() {
  const navigate = useNavigate();
  const { addItem } = useCartStore();
  const [loading, setLoading] = useState(true);
  const [config, setConfig] = useState<FuelStationConfig | null>(null);
  const [skuMap, setSKUMap] = useState<Record<number, SKUWithSPU>>({});
  const [selectedTier, setSelectedTier] = useState<Record<string, number>>({});

  useEffect(() => {
    let mounted = true;
    (async () => {
      setLoading(true);
      try {
        const cfgRes = await fuelStationService.getPublicConfig();
        const cfg = cfgRes.data.data;
        if (!mounted) return;
        setConfig(cfg);
        const skuIDs = Array.from(
          new Set(
            (cfg.sections || [])
              .flatMap((s) => s.tiers || [])
              .map((t) => Number(t.sku_id))
              .filter((id) => id > 0)
          )
        );
        const skuPairs = await Promise.all(
          skuIDs.map(async (id) => {
            try {
              const res = await skuService.getPublicSKU(id);
              return [id, res.data.data] as [number, SKUWithSPU];
            } catch {
              return null;
            }
          })
        );
        const nextMap: Record<number, SKUWithSPU> = {};
        for (const it of skuPairs) {
          if (it?.[1]) nextMap[it[0]] = it[1];
        }
        setSKUMap(nextMap);
      } catch {
        if (mounted) {
          setConfig(null);
          setSKUMap({});
        }
      } finally {
        if (mounted) setLoading(false);
      }
    })();
    return () => {
      mounted = false;
    };
  }, []);

  const activeSections = useMemo(
    () =>
      (config?.sections || [])
        .filter((s) => s.status === 'active')
        .sort((a, b) => Number(a.sort_order || 0) - Number(b.sort_order || 0)),
    [config]
  );

  const addSelectedTierToCart = (sectionCode: string) => {
    const skuID = selectedTier[sectionCode];
    const sku = skuID ? skuMap[skuID] : undefined;
    if (!sku) {
      message.warning('请先选择一个 S/M/L 档位');
      return;
    }
    addItem(
      {
        id: sku.id,
        name: `${sku.spu_name} · ${sku.sku_code}`,
        description: '',
        price: Number(sku.retail_price || 0),
        stock: Number(sku.stock || 0),
        sold_count: Number(sku.sales_count || 0),
        merchant_id: 0,
        status: sku.status === 'inactive' || sku.status === 'archived' ? sku.status : 'active',
        created_at: sku.created_at || '',
        updated_at: sku.updated_at || '',
      },
      1
    );
    message.success('已加入购物车，请与模型商品组合结算');
    navigate('/cart');
  };

  return (
    <div style={{ padding: '20px', maxWidth: 1080, margin: '0 auto' }}>
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <Title level={3} style={{ marginBottom: 0 }}>
          {config?.page_title || '智燃加油站'}
        </Title>
        <Paragraph type="secondary" style={{ marginBottom: 0 }}>
          {config?.page_subtitle ||
            '面向已订购模型权益的用户提供加油补充。你订购了哪些在售模型，加油包余额即可用于对应模型调用。'}
        </Paragraph>

        {!loading && !config && (
          <Alert
            type="warning"
            showIcon
            message="暂时无法加载加油站配置"
            description="请检查网络后重试，或先到卖场搭配模型商品与加油包一起结算。"
            action={
              <Button type="primary" onClick={() => navigate('/catalog')}>
                去卖场
              </Button>
            }
          />
        )}

        <Collapse
          bordered={false}
          style={{ background: 'transparent' }}
          items={[
            {
              key: 'fuel-rules',
              label: '购买规则',
              children: (
                <Paragraph type="secondary" style={{ marginBottom: 0 }}>
                  {config?.rule_text ||
                    '加油包不可单独购买，需与至少一个在售模型商品或套餐包组合下单；仅持有余额不代表自动开通新模型权限。'}
                </Paragraph>
              ),
            },
          ]}
        />

        <Spin spinning={loading}>
          {activeSections.length === 0 ? (
            <Empty description="暂无可购买加油包，请联系运营配置" />
          ) : (
            <Row gutter={[16, 16]}>
              {activeSections.map((section) => (
                <Col xs={24} md={12} key={section.code}>
                  <Card
                    title={
                      <Space>
                        <ThunderboltOutlined />
                        <span>{section.name}</span>
                      </Space>
                    }
                    extra={section.badge ? <Tag color="blue">{section.badge}</Tag> : null}
                  >
                    <Space direction="vertical" size={10} style={{ width: '100%' }}>
                      <Text type="secondary">{section.description}</Text>
                      <Radio.Group
                        value={selectedTier[section.code]}
                        onChange={(e) =>
                          setSelectedTier((prev) => ({
                            ...prev,
                            [section.code]: Number(e.target.value),
                          }))
                        }
                      >
                        <Space direction="vertical">
                          {(section.tiers || []).map((tier) => {
                            const sku = skuMap[Number(tier.sku_id)];
                            return (
                              <Radio
                                key={`${section.code}-${tier.label}`}
                                value={tier.sku_id}
                                disabled={!sku}
                              >
                                <Space size={6}>
                                  <Text strong>{tier.label}</Text>
                                  <Text>
                                    {sku ? `${sku.token_amount || 0} Token` : '未配置 SKU'}
                                  </Text>
                                  <Text type="secondary">
                                    {sku ? `¥${Number(sku.retail_price || 0).toFixed(2)}` : ''}
                                  </Text>
                                </Space>
                              </Radio>
                            );
                          })}
                        </Space>
                      </Radio.Group>
                      <IconHintButton
                        type="primary"
                        hint="将当前所选档位加入购物车"
                        icon={<ShoppingCartOutlined />}
                        onClick={() => addSelectedTierToCart(section.code)}
                      />
                    </Space>
                  </Card>
                </Col>
              ))}
            </Row>
          )}
        </Spin>

        <Card>
          <Space direction="vertical" size={8}>
            <Space>
              <RocketOutlined />
              <Text strong>后续扩展</Text>
            </Space>
            <Text>
              当前优先覆盖编程品类；后续将按用户已订购模型和用途扩展到音视频、绘画等场景加油包。
            </Text>
          </Space>
        </Card>
      </Space>
    </div>
  );
}
