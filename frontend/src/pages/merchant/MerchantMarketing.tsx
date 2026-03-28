import React, { useState } from 'react';
import {
  Card,
  Row,
  Col,
  Typography,
  Button,
  Table,
  Tag,
  Space,
  Modal,
  Form,
  Input,
  InputNumber,
  DatePicker,
  Select,
  Divider,
  Tabs,
  List,
  Switch,
  message,
} from 'antd';
import {
  GiftOutlined,
  PlusOutlined,
  ThunderboltOutlined,
  EditOutlined,
  DeleteOutlined,
} from '@ant-design/icons';
import styles from './Merchant.module.css';

const { Title, Text } = Typography;
const { RangePicker } = DatePicker;
const { Option } = Select;
const { TabPane } = Tabs;

interface Coupon {
  id: number;
  name: string;
  type: 'discount' | 'cash';
  value: number;
  min_purchase: number;
  used_count: number;
  total_count: number;
  status: 'active' | 'inactive';
  expire_at: string;
}

interface FlashSale {
  id: number;
  product_id: number;
  product_name: string;
  original_price: number;
  flash_price: number;
  stock: number;
  sold: number;
  start_time: string;
  end_time: string;
  status: 'pending' | 'active' | 'ended';
}

const mockCoupons: Coupon[] = [
  {
    id: 1,
    name: '新用户专享券',
    type: 'cash',
    value: 10,
    min_purchase: 50,
    used_count: 156,
    total_count: 500,
    status: 'active',
    expire_at: '2026-04-30',
  },
  {
    id: 2,
    name: '满100减20',
    type: 'cash',
    value: 20,
    min_purchase: 100,
    used_count: 89,
    total_count: 200,
    status: 'active',
    expire_at: '2026-03-31',
  },
  {
    id: 3,
    name: '全场9折券',
    type: 'discount',
    value: 10,
    min_purchase: 0,
    used_count: 234,
    total_count: 300,
    status: 'active',
    expire_at: '2026-04-15',
  },
];

const mockFlashSales: FlashSale[] = [
  {
    id: 1,
    product_id: 101,
    product_name: 'GLM-5 Token包',
    original_price: 100,
    flash_price: 59,
    stock: 100,
    sold: 78,
    start_time: '2026-03-24 10:00',
    end_time: '2026-03-24 12:00',
    status: 'active',
  },
  {
    id: 2,
    product_id: 102,
    product_name: 'K2.5 Token包',
    original_price: 150,
    flash_price: 99,
    stock: 50,
    sold: 50,
    start_time: '2026-03-24 14:00',
    end_time: '2026-03-24 16:00',
    status: 'ended',
  },
];

export const MerchantMarketing: React.FC = () => {
  const [coupons, setCoupons] = useState<Coupon[]>(mockCoupons);
  const [flashSales, setFlashSales] = useState<FlashSale[]>(mockFlashSales);
  const [createCouponModalVisible, setCreateCouponModalVisible] = useState(false);
  const [createFlashSaleModalVisible, setCreateFlashSaleModalVisible] = useState(false);
  const [form] = Form.useForm();
  const [flashSaleForm] = Form.useForm();

  const handleCreateCoupon = async () => {
    try {
      const values = await form.validateFields();
      const newCoupon: Coupon = {
        id: coupons.length + 1,
        name: values.name,
        type: values.type,
        value: values.value,
        min_purchase: values.min_purchase || 0,
        used_count: 0,
        total_count: values.total_count,
        status: 'active',
        expire_at: values.expire_at.format('YYYY-MM-DD'),
      };
      setCoupons([...coupons, newCoupon]);
      message.success('优惠券创建成功');
      setCreateCouponModalVisible(false);
      form.resetFields();
    } catch {
      message.error('请填写完整信息');
    }
  };

  const handleCreateFlashSale = async () => {
    try {
      const values = await flashSaleForm.validateFields();
      const newFlashSale: FlashSale = {
        id: flashSales.length + 1,
        product_id: values.product_id,
        product_name: values.product_name,
        original_price: values.original_price,
        flash_price: values.flash_price,
        stock: values.stock,
        sold: 0,
        start_time: values.time_range[0].format('YYYY-MM-DD HH:mm'),
        end_time: values.time_range[1].format('YYYY-MM-DD HH:mm'),
        status: 'pending',
      };
      setFlashSales([...flashSales, newFlashSale]);
      message.success('秒杀活动创建成功');
      setCreateFlashSaleModalVisible(false);
      flashSaleForm.resetFields();
    } catch {
      message.error('请填写完整信息');
    }
  };

  const couponColumns = [
    { title: '优惠券名称', dataIndex: 'name', key: 'name' },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) =>
        type === 'cash' ? <Tag color="blue">满减券</Tag> : <Tag color="green">折扣券</Tag>,
    },
    {
      title: '优惠内容',
      key: 'value',
      render: (_: unknown, record: Coupon) =>
        record.type === 'cash' ? `减¥${record.value}` : `${100 - record.value}折`,
    },
    {
      title: '使用门槛',
      dataIndex: 'min_purchase',
      key: 'min_purchase',
      render: (v: number) => (v > 0 ? `满¥${v}` : '无门槛'),
    },
    {
      title: '已用/总量',
      key: 'usage',
      render: (_: unknown, record: Coupon) => `${record.used_count}/${record.total_count}`,
    },
    { title: '有效期', dataIndex: 'expire_at', key: 'expire_at' },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) =>
        status === 'active' ? <Tag color="success">生效中</Tag> : <Tag color="default">已停用</Tag>,
    },
    {
      title: '操作',
      key: 'action',
      render: () => (
        <Space>
          <Button type="link" icon={<EditOutlined />}>
            编辑
          </Button>
          <Button type="link" danger icon={<DeleteOutlined />}>
            删除
          </Button>
        </Space>
      ),
    },
  ];

  const flashSaleColumns = [
    { title: '商品ID', dataIndex: 'product_id', key: 'product_id' },
    { title: '商品名称', dataIndex: 'product_name', key: 'product_name' },
    {
      title: '价格',
      key: 'price',
      render: (_: unknown, record: FlashSale) => (
        <Space>
          <Text delete type="secondary">
            ¥{record.original_price}
          </Text>
          <Text type="danger" strong>
            ¥{record.flash_price}
          </Text>
        </Space>
      ),
    },
    { title: '库存', dataIndex: 'stock', key: 'stock' },
    { title: '已售', dataIndex: 'sold', key: 'sold' },
    { title: '开始时间', dataIndex: 'start_time', key: 'start_time' },
    { title: '结束时间', dataIndex: 'end_time', key: 'end_time' },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const config: Record<string, { color: string; label: string }> = {
          pending: { color: 'processing', label: '待开始' },
          active: { color: 'success', label: '进行中' },
          ended: { color: 'default', label: '已结束' },
        };
        return <Tag color={config[status].color}>{config[status].label}</Tag>;
      },
    },
    {
      title: '操作',
      key: 'action',
      render: () => (
        <Space>
          <Button type="link" icon={<EditOutlined />}>
            编辑
          </Button>
          <Button type="link" danger icon={<DeleteOutlined />}>
            取消
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div className={styles.container}>
      <Title level={3} style={{ marginBottom: 24 }}>
        <GiftOutlined style={{ marginRight: 8 }} />
        营销工具
      </Title>

      <Tabs defaultActiveKey="coupons">
        <TabPane tab="优惠券管理" key="coupons">
          <Card>
            <div style={{ marginBottom: 16 }}>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => setCreateCouponModalVisible(true)}
              >
                创建优惠券
              </Button>
            </div>
            <Table
              columns={couponColumns}
              dataSource={coupons}
              rowKey="id"
              pagination={false}
              scroll={{ x: 'max-content' }}
            />
          </Card>
        </TabPane>

        <TabPane tab="限时秒杀" key="flash-sale">
          <Card>
            <div style={{ marginBottom: 16 }}>
              <Button
                type="primary"
                icon={<ThunderboltOutlined />}
                onClick={() => setCreateFlashSaleModalVisible(true)}
              >
                创建秒杀活动
              </Button>
            </div>
            <Table
              columns={flashSaleColumns}
              dataSource={flashSales}
              rowKey="id"
              pagination={false}
              scroll={{ x: 'max-content' }}
            />
          </Card>
        </TabPane>

        <TabPane tab="新客优惠" key="new-customer">
          <Card title="新客礼包设置">
            <Row gutter={[24, 24]}>
              <Col span={24}>
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    padding: '16px 0',
                  }}
                >
                  <div>
                    <Text strong>首单折扣</Text>
                    <br />
                    <Text type="secondary">新用户首次下单享受折扣优惠</Text>
                  </div>
                  <Space>
                    <InputNumber defaultValue={10} min={0} max={50} addonAfter="%" />
                    <Switch defaultChecked />
                  </Space>
                </div>
                <Divider />
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    padding: '16px 0',
                  }}
                >
                  <div>
                    <Text strong>新人优惠券</Text>
                    <br />
                    <Text type="secondary">注册即送优惠券</Text>
                  </div>
                  <Space>
                    <Select defaultValue="10" style={{ width: 120 }}>
                      <Option value="10">满50减10</Option>
                      <Option value="20">满100减20</Option>
                      <Option value="50">满200减50</Option>
                    </Select>
                    <Switch defaultChecked />
                  </Space>
                </div>
                <Divider />
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    padding: '16px 0',
                  }}
                >
                  <div>
                    <Text strong>新人专享价</Text>
                    <br />
                    <Text type="secondary">部分商品对新用户显示特价</Text>
                  </div>
                  <Switch defaultChecked />
                </div>
              </Col>
            </Row>
          </Card>
        </TabPane>

        <TabPane tab="平台活动" key="platform">
          <Card title="平台大促活动">
            <List
              dataSource={[
                { name: '618大促', status: '报名中', deadline: '2026-05-31' },
                { name: '双11狂欢节', status: '即将开始', deadline: '2026-10-15' },
                { name: '年终盛典', status: '筹备中', deadline: '2026-11-30' },
              ]}
              renderItem={(item) => (
                <List.Item
                  actions={[
                    <Button type="link" key="join">
                      立即报名
                    </Button>,
                  ]}
                >
                  <List.Item.Meta title={item.name} description={`截止日期: ${item.deadline}`} />
                  <Tag color="blue">{item.status}</Tag>
                </List.Item>
              )}
            />
          </Card>
        </TabPane>
      </Tabs>

      <Modal
        title="创建优惠券"
        open={createCouponModalVisible}
        onCancel={() => setCreateCouponModalVisible(false)}
        onOk={handleCreateCoupon}
        okText="创建"
        cancelText="取消"
        width={600}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="优惠券名称" rules={[{ required: true }]}>
            <Input placeholder="请输入优惠券名称" />
          </Form.Item>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Select placeholder="请选择类型">
              <Option value="cash">满减券</Option>
              <Option value="discount">折扣券</Option>
            </Select>
          </Form.Item>
          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <Form.Item name="value" label="优惠额度" rules={[{ required: true }]}>
                <InputNumber style={{ width: '100%' }} min={1} placeholder="满减金额或折扣百分比" />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="min_purchase" label="使用门槛">
                <InputNumber
                  style={{ width: '100%' }}
                  min={0}
                  placeholder="最低消费金额，0表示无门槛"
                />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="total_count" label="发放数量" rules={[{ required: true }]}>
            <InputNumber style={{ width: '100%' }} min={1} placeholder="发放总数量" />
          </Form.Item>
          <Form.Item name="expire_at" label="有效期" rules={[{ required: true }]}>
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="创建秒杀活动"
        open={createFlashSaleModalVisible}
        onCancel={() => setCreateFlashSaleModalVisible(false)}
        onOk={handleCreateFlashSale}
        okText="创建"
        cancelText="取消"
        width={600}
      >
        <Form form={flashSaleForm} layout="vertical">
          <Form.Item name="product_id" label="商品ID" rules={[{ required: true }]}>
            <InputNumber style={{ width: '100%' }} placeholder="请输入商品ID" />
          </Form.Item>
          <Form.Item name="product_name" label="商品名称" rules={[{ required: true }]}>
            <Input placeholder="请输入商品名称" />
          </Form.Item>
          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <Form.Item name="original_price" label="原价" rules={[{ required: true }]}>
                <InputNumber style={{ width: '100%' }} min={0} prefix="¥" />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="flash_price" label="秒杀价" rules={[{ required: true }]}>
                <InputNumber style={{ width: '100%' }} min={0} prefix="¥" />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <Form.Item name="stock" label="秒杀库存" rules={[{ required: true }]}>
                <InputNumber style={{ width: '100%' }} min={1} placeholder="秒杀库存数量" />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="time_range" label="活动时间" rules={[{ required: true }]}>
                <RangePicker showTime style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </div>
  );
};

export default MerchantMarketing;
