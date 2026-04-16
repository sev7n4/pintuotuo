import { useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  Col,
  Divider,
  Form,
  Input,
  InputNumber,
  Row,
  Select,
  Space,
  Tag,
  Tooltip,
  Typography,
  Modal,
  message,
} from 'antd';
import { MinusCircleOutlined, PlusOutlined, SaveOutlined } from '@ant-design/icons';
import { fuelStationService } from '@/services/fuelStation';
import type { FuelStationConfig, FuelStationTemplate } from '@/types/fuelStation';
import { skuService } from '@/services/sku';
import type { SKUWithSPU } from '@/types/sku';

const { Title, Paragraph, Text } = Typography;

const statusOptions = [
  { label: '启用', value: 'active' },
  { label: '停用', value: 'inactive' },
];

type TierHealth = {
  ok: boolean;
  messages: string[];
};

const AdminFuelStation = () => {
  const [form] = Form.useForm<FuelStationConfig>();
  const [loading, setLoading] = useState(false);
  const [skuOptions, setSkuOptions] = useState<SKUWithSPU[]>([]);
  const [templateLibrary, setTemplateLibrary] = useState<FuelStationTemplate[]>([]);
  const [selectedTemplateKey, setSelectedTemplateKey] = useState<string>('coding_v1');
  const [templateJSON, setTemplateJSON] = useState<string>('');
  const [newTemplateModalOpen, setNewTemplateModalOpen] = useState(false);
  const [newTemplateKey, setNewTemplateKey] = useState('');
  const [newTemplateName, setNewTemplateName] = useState('');
  const [newTemplateDesc, setNewTemplateDesc] = useState('');
  const watchedValues = Form.useWatch([], form);

  useEffect(() => {
    if (templateLibrary.length === 0) return;
    const tpl = templateLibrary.find((t) => t.key === selectedTemplateKey) || templateLibrary[0];
    if (tpl) setTemplateJSON(JSON.stringify(tpl.payload, null, 2));
  }, [selectedTemplateKey, templateLibrary]);

  const loadSKUOptions = async (q?: string) => {
    try {
      const res = await skuService.getSKUs({
        scope: 'all',
        status: 'active',
        q,
        per_page: 100,
      });
      setSkuOptions(res.data.data || []);
    } catch {
      setSkuOptions([]);
    }
  };

  const load = async () => {
    setLoading(true);
    try {
      const res = await fuelStationService.getAdminConfig();
      form.setFieldsValue(res.data.data);
      const tplRes = await fuelStationService.getAdminTemplates();
      const tpls = tplRes.data.data || [];
      setTemplateLibrary(tpls);
      if (tpls.length > 0) {
        setSelectedTemplateKey(tpls[0].key);
      }
      await loadSKUOptions();
    } catch {
      message.error('加载加油站配置失败');
    } finally {
      setLoading(false);
    }
  };

  const skuLabelMap = useMemo(() => {
    const m = new Map<number, string>();
    for (const s of skuOptions) {
      m.set(
        Number(s.id),
        `${s.sku_code} / ${s.spu_name} / ${Number(s.token_amount || 0).toLocaleString()} Token / ¥${Number(s.retail_price || 0).toFixed(2)}`
      );
    }
    return m;
  }, [skuOptions]);

  const skuByID = useMemo(() => {
    const m = new Map<number, SKUWithSPU>();
    for (const s of skuOptions) {
      m.set(Number(s.id), s);
    }
    return m;
  }, [skuOptions]);

  const evaluateTierHealth = (skuID?: number): TierHealth => {
    if (!skuID) return { ok: false, messages: ['未选择 SKU'] };
    const sku = skuByID.get(Number(skuID));
    if (!sku) return { ok: false, messages: ['SKU 不在候选列表（可能不存在或已下架）'] };
    const issues: string[] = [];
    if (sku.status !== 'active') issues.push('SKU 非 active');
    if (sku.spu_status && sku.spu_status !== 'active') issues.push('所属 SPU 非 active');
    if (sku.sku_type !== 'token_pack') issues.push(`SKU 类型为 ${sku.sku_type}，建议使用 token_pack`);
    if (Number(sku.stock || 0) === 0) issues.push('库存为 0');
    return { ok: issues.length === 0, messages: issues.length ? issues : ['可售'] };
  };

  useEffect(() => {
    void load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const onFinish = async (values: FuelStationConfig) => {
    setLoading(true);
    try {
      await fuelStationService.updateAdminConfig(values);
      message.success('加油站配置已保存');
      await load();
    } catch {
      message.error('保存失败');
    } finally {
      setLoading(false);
    }
  };

  const applyTemplate = () => {
    try {
      const parsed = JSON.parse(templateJSON) as FuelStationConfig;
      form.setFieldsValue(parsed);
      message.success('模板已应用到当前草稿，请检查并填写 SKU 后保存');
    } catch {
      message.error('模板 JSON 解析失败，请检查格式');
    }
  };

  const saveTemplateLibrary = async () => {
    try {
      const parsed = JSON.parse(templateJSON) as FuelStationConfig;
      const next = templateLibrary.map((t) =>
        t.key === selectedTemplateKey ? { ...t, payload: parsed } : t
      );
      await fuelStationService.updateAdminTemplates(next);
      setTemplateLibrary(next);
      message.success('模板库已保存');
    } catch {
      message.error('模板 JSON 解析失败或保存失败');
    }
  };

  const createTemplateFromCurrentJSON = async () => {
    const key = newTemplateKey.trim();
    const name = newTemplateName.trim();
    if (!key || !name) {
      message.warning('模板 key 和名称必填');
      return;
    }
    if (templateLibrary.some((t) => t.key === key)) {
      message.warning('模板 key 已存在，请换一个');
      return;
    }
    try {
      const payload = JSON.parse(templateJSON) as FuelStationConfig;
      const next = [
        ...templateLibrary,
        {
          key,
          name,
          description: newTemplateDesc.trim(),
          payload,
        },
      ];
      await fuelStationService.updateAdminTemplates(next);
      setTemplateLibrary(next);
      setSelectedTemplateKey(key);
      setNewTemplateModalOpen(false);
      setNewTemplateKey('');
      setNewTemplateName('');
      setNewTemplateDesc('');
      message.success('模板已新增并保存');
    } catch {
      message.error('当前模板 JSON 非法，无法新增模板');
    }
  };

  const deleteSelectedTemplate = async () => {
    if (templateLibrary.length <= 1) {
      message.warning('至少保留一个模板');
      return;
    }
    const next = templateLibrary.filter((t) => t.key !== selectedTemplateKey);
    try {
      await fuelStationService.updateAdminTemplates(next);
      setTemplateLibrary(next);
      setSelectedTemplateKey(next[0].key);
      message.success('模板已删除');
    } catch {
      message.error('删除模板失败');
    }
  };

  const moveTemplate = async (direction: -1 | 1) => {
    const idx = templateLibrary.findIndex((t) => t.key === selectedTemplateKey);
    if (idx < 0) return;
    const target = idx + direction;
    if (target < 0 || target >= templateLibrary.length) return;
    const next = [...templateLibrary];
    const [item] = next.splice(idx, 1);
    next.splice(target, 0, item);
    try {
      await fuelStationService.updateAdminTemplates(next);
      setTemplateLibrary(next);
      message.success('模板顺序已更新');
    } catch {
      message.error('模板排序保存失败');
    }
  };

  return (
    <div>
      <Title level={2}>加油站专区运营</Title>
      <Alert
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
        message="用于运营配置加油站专区卡片"
        description="每个卡片可配置 S/M/L 三档 SKU。前台以“选档后加入购物车”方式下单，不再一键组合下单。"
      />
      <Card loading={loading}>
        <Card size="small" title="模板中心（可配置骨架）" style={{ marginBottom: 16 }}>
          <Paragraph type="secondary" style={{ marginBottom: 8 }}>
            这里是模板骨架入口：先选模板，再按需改 JSON 后应用到表单。后续可扩展为后端持久化模板。
          </Paragraph>
          <Space direction="vertical" style={{ width: '100%' }} size={10}>
            <Space wrap>
              <Select
                style={{ width: 280 }}
                value={selectedTemplateKey}
                onChange={setSelectedTemplateKey}
                options={templateLibrary.map((t) => ({ value: t.key, label: t.name }))}
              />
              <Tooltip title={templateLibrary.find((t) => t.key === selectedTemplateKey)?.description || ''}>
                <Tag color="blue">模板说明</Tag>
              </Tooltip>
              <Button onClick={applyTemplate}>应用模板到草稿</Button>
              <Button onClick={saveTemplateLibrary}>保存到模板库</Button>
              <Button onClick={() => setNewTemplateModalOpen(true)}>新增模板</Button>
              <Button danger onClick={deleteSelectedTemplate}>
                删除模板
              </Button>
              <Button onClick={() => moveTemplate(-1)}>上移</Button>
              <Button onClick={() => moveTemplate(1)}>下移</Button>
            </Space>
            <Input.TextArea
              value={templateJSON}
              onChange={(e) => setTemplateJSON(e.target.value)}
              rows={10}
              placeholder="可编辑模板 JSON"
            />
          </Space>
        </Card>

        <Form form={form} layout="vertical" onFinish={onFinish}>
          <Form.Item name="page_title" label="页面标题" rules={[{ required: true, message: '必填' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="page_subtitle" label="页面副标题">
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="rule_text" label="规则提示文案">
            <Input.TextArea rows={2} />
          </Form.Item>

          <Form.List name="sections">
            {(fields, { add, remove }) => (
              <>
                {fields.map((field) => (
                  <Card
                    key={field.key}
                    type="inner"
                    style={{ marginBottom: 12 }}
                    title={`卡片 #${field.name + 1}`}
                    extra={<MinusCircleOutlined onClick={() => remove(field.name)} />}
                  >
                    <Form.Item
                      name={[field.name, 'code']}
                      label="卡片编码"
                      rules={[{ required: true, message: '必填' }]}
                    >
                      <Input placeholder="如 coding / video / image" />
                    </Form.Item>
                    <Form.Item
                      name={[field.name, 'name']}
                      label="卡片名称"
                      rules={[{ required: true, message: '必填' }]}
                    >
                      <Input />
                    </Form.Item>
                    <Form.Item name={[field.name, 'description']} label="卡片描述">
                      <Input />
                    </Form.Item>
                    <Space style={{ width: '100%' }} size={12} wrap>
                      <Form.Item name={[field.name, 'badge']} label="角标">
                        <Input style={{ width: 180 }} />
                      </Form.Item>
                      <Form.Item name={[field.name, 'sort_order']} label="排序">
                        <InputNumber min={0} style={{ width: 120 }} />
                      </Form.Item>
                      <Form.Item name={[field.name, 'status']} label="状态">
                        <Select options={statusOptions} style={{ width: 140 }} />
                      </Form.Item>
                    </Space>

                    <Form.List name={[field.name, 'tiers']}>
                      {(tierFields, tierOps) => (
                        <>
                          {tierFields.map((tf) => (
                            <Row key={tf.key} gutter={12} align="middle">
                              <Col xs={24} md={5}>
                                <Form.Item
                                  name={[tf.name, 'label']}
                                  label="档位名"
                                  rules={[{ required: true, message: '必填' }]}
                                >
                                  <Input placeholder="S 档 / M 档 / L 档" />
                                </Form.Item>
                              </Col>
                              <Col xs={24} md={18}>
                                <Form.Item
                                  name={[tf.name, 'sku_id']}
                                  label="SKU（可搜索）"
                                  rules={[{ required: true, message: '必填' }]}
                                >
                                  <Select
                                    showSearch
                                    placeholder="输入 SKU 编码 / SPU 名称搜索"
                                    optionFilterProp="label"
                                    onSearch={(v) => {
                                      void loadSKUOptions(v);
                                    }}
                                    options={skuOptions.map((s) => ({
                                      value: s.id,
                                      label: `${s.sku_code} / ${s.spu_name} / ${Number(s.token_amount || 0).toLocaleString()} Token / ¥${Number(s.retail_price || 0).toFixed(2)}`,
                                    }))}
                                  />
                                </Form.Item>
                                <Form.Item shouldUpdate noStyle>
                                  {() => {
                                    const currentID = form.getFieldValue([
                                      'sections',
                                      field.name,
                                      'tiers',
                                      tf.name,
                                      'sku_id',
                                    ]) as number | undefined;
                                    const health = evaluateTierHealth(currentID);
                                    return (
                                      <div style={{ marginBottom: 12 }}>
                                        <Tag color={health.ok ? 'green' : 'red'}>
                                          {health.ok ? '可售校验通过' : '可售校验未通过'}
                                        </Tag>
                                        {!health.ok ? (
                                          <Text type="secondary" style={{ marginLeft: 8 }}>
                                            {health.messages.join('；')}
                                          </Text>
                                        ) : null}
                                      </div>
                                    );
                                  }}
                                </Form.Item>
                              </Col>
                              <Col xs={24} md={1} style={{ textAlign: 'right' }}>
                                <MinusCircleOutlined onClick={() => tierOps.remove(tf.name)} />
                              </Col>
                            </Row>
                          ))}
                          <Form.Item>
                            <Button type="dashed" onClick={() => tierOps.add()} icon={<PlusOutlined />}>
                              添加档位
                            </Button>
                          </Form.Item>
                        </>
                      )}
                    </Form.List>
                  </Card>
                ))}
                <Form.Item>
                  <Button type="dashed" onClick={() => add()} icon={<PlusOutlined />} block>
                    添加卡片
                  </Button>
                </Form.Item>
              </>
            )}
          </Form.List>

          <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={loading}>
            保存配置
          </Button>
        </Form>
      </Card>

      <Card style={{ marginTop: 16 }} title="前台效果预览（草稿）">
        <Paragraph type="secondary" style={{ marginBottom: 12 }}>
          该预览基于当前表单草稿实时渲染，帮助运营确认卡片布局与 S/M/L 档位展示效果。
        </Paragraph>
        <Title level={4} style={{ marginBottom: 0 }}>
          {watchedValues?.page_title || '智燃加油站'}
        </Title>
        <Paragraph type="secondary">{watchedValues?.page_subtitle || '面向已订购模型权益用户，按用途补充 Token。'}</Paragraph>
        <Alert
          type="info"
          showIcon
          style={{ marginBottom: 12 }}
          message="规则提示"
          description={watchedValues?.rule_text || '加油包不可单独购买，需与模型商品或套餐包组合下单。'}
        />
        <Alert
          type="warning"
          showIcon
          style={{ marginBottom: 12 }}
          message="运营检查建议"
          description="每个档位建议使用在售 token_pack SKU，且所属 SPU 也应为 active；下方红色标记表示当前配置可能导致用户侧不可售。"
        />
        <Row gutter={[12, 12]}>
          {(watchedValues?.sections || []).map((section, idx) => (
            <Col xs={24} md={12} key={`${section?.code || 'section'}-${idx}`}>
              <Card
                size="small"
                title={section?.name || `卡片 #${idx + 1}`}
                extra={section?.badge ? <Tag color="blue">{section.badge}</Tag> : null}
              >
                <Paragraph type="secondary">{section?.description || '未填写描述'}</Paragraph>
                <Space direction="vertical" size={4} style={{ width: '100%' }}>
                  {(section?.tiers || []).map((tier, i) => (
                    <div key={`${tier?.label || 'tier'}-${i}`}>
                      <Text strong>{tier?.label || `档位 #${i + 1}`}</Text>
                      <br />
                      <Text type="secondary">{skuLabelMap.get(Number(tier?.sku_id || 0)) || '未关联 SKU'}</Text>
                    </div>
                  ))}
                </Space>
              </Card>
            </Col>
          ))}
        </Row>
        <Divider style={{ marginTop: 16, marginBottom: 8 }} />
        <Text type="secondary">用户端路径：`/fuel-station`</Text>
      </Card>

      <Modal
        title="新增模板"
        open={newTemplateModalOpen}
        onCancel={() => setNewTemplateModalOpen(false)}
        onOk={createTemplateFromCurrentJSON}
        okText="创建并保存"
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Input
            placeholder="模板 key（唯一，如 coding_v2）"
            value={newTemplateKey}
            onChange={(e) => setNewTemplateKey(e.target.value)}
          />
          <Input
            placeholder="模板名称"
            value={newTemplateName}
            onChange={(e) => setNewTemplateName(e.target.value)}
          />
          <Input.TextArea
            rows={2}
            placeholder="模板说明（可选）"
            value={newTemplateDesc}
            onChange={(e) => setNewTemplateDesc(e.target.value)}
          />
          <Text type="secondary">将基于当前模板 JSON 创建新模板。</Text>
        </Space>
      </Modal>
    </div>
  );
};

export default AdminFuelStation;

