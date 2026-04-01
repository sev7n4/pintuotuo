import { useEffect, useMemo, useState } from 'react';
import {
  Card,
  Tabs,
  Table,
  Button,
  Tag,
  Space,
  Modal,
  Form,
  Input,
  InputNumber,
  Select,
  message,
  Statistic,
  Row,
  Col,
  Popconfirm,
  Empty,
  Typography,
  Alert,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  WalletOutlined,
  ApiOutlined,
  HistoryOutlined,
  PlusOutlined,
  DeleteOutlined,
  SendOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { useTokenStore } from '@/stores/tokenStore';
import { UserAPIKey, RechargeOrder } from '@/types';
import styles from './MyToken.module.css';

const { TabPane } = Tabs;
const { Text, Paragraph } = Typography;

function openAICompatBaseURL(): string {
  const base = (import.meta.env.VITE_API_BASE_URL as string | undefined)?.trim() || '/api/v1';
  const normalized = base.endsWith('/') ? base.slice(0, -1) : base;
  if (normalized.startsWith('http://') || normalized.startsWith('https://')) {
    return `${normalized}/openai/v1`;
  }
  if (typeof window !== 'undefined') {
    return `${window.location.origin}${normalized}/openai/v1`;
  }
  return `${normalized}/openai/v1`;
}

const MyToken = () => {
  const allowMockRecharge = import.meta.env.VITE_ALLOW_MOCK_RECHARGE === 'true';
  const openAICompatBase = useMemo(() => openAICompatBaseURL(), []);

  const {
    balance,
    transactions,
    apiKeys,
    rechargeOrders,
    fetchBalance,
    fetchTransactions,
    fetchAPIKeys,
    fetchRechargeOrders,
    createAPIKey,
    deleteAPIKey,
    createRechargeOrder,
    mockCompleteRechargeOrder,
    transfer,
    isLoading,
    error,
    clearError,
  } = useTokenStore();

  const [createKeyModalVisible, setCreateKeyModalVisible] = useState(false);
  const [transferModalVisible, setTransferModalVisible] = useState(false);
  const [rechargeModalVisible, setRechargeModalVisible] = useState(false);
  const [keyForm] = Form.useForm();
  const [transferForm] = Form.useForm();
  const [rechargeForm] = Form.useForm();
  const [newKeyDisplay, setNewKeyDisplay] = useState<string | null>(null);

  useEffect(() => {
    fetchBalance();
    fetchTransactions();
    fetchAPIKeys();
    fetchRechargeOrders();
  }, [fetchBalance, fetchTransactions, fetchAPIKeys, fetchRechargeOrders]);

  const handleCreateKey = async () => {
    try {
      const values = await keyForm.validateFields();
      const key = await createAPIKey(values.name.trim());
      if (key) {
        message.success('API密钥创建成功');
        setNewKeyDisplay(key);
        keyForm.resetFields();
        fetchAPIKeys();
      } else {
        message.error(error || '创建失败');
      }
    } catch {
      // no-op: 表单校验错误不弹全局错误
    }
  };

  const handleDeleteKey = async (id: number) => {
    const success = await deleteAPIKey(id);
    if (success) {
      message.success('API密钥已删除');
      fetchAPIKeys();
    }
  };

  const handleTransfer = async () => {
    try {
      const values = await transferForm.validateFields();
      const raw = values.recipient.trim();
      const isEmail = raw.includes('@');
      let success: boolean;
      if (isEmail) {
        success = await transfer(values.amount, { recipientEmail: raw });
      } else {
        const id = parseInt(raw, 10);
        if (Number.isNaN(id) || id < 1) {
          message.error('请输入有效的数字用户ID或完整注册邮箱');
          return;
        }
        success = await transfer(values.amount, { recipientId: id });
      }
      if (success) {
        message.success('转账成功');
        setTransferModalVisible(false);
        transferForm.resetFields();
        fetchBalance();
        fetchTransactions();
      } else {
        message.error(error || '转账失败');
      }
    } catch {
      // no-op: 表单校验错误不弹全局错误
    }
  };

  const handleMockPayOrder = async (order: RechargeOrder) => {
    if (!allowMockRecharge) return;
    const ok = await mockCompleteRechargeOrder(order.id);
    if (ok) {
      message.success('模拟支付完成，余额已更新');
      fetchBalance();
      fetchTransactions();
      fetchRechargeOrders();
    } else {
      message.error(error || '模拟支付失败');
    }
  };

  const handleCreateRechargeOrder = async () => {
    try {
      const values = await rechargeForm.validateFields();
      const order = await createRechargeOrder(values.amount, values.method);
      if (order) {
        message.success(`充值订单创建成功，订单号：${order.out_trade_no}`);
        setRechargeModalVisible(false);
        rechargeForm.resetFields();
        fetchRechargeOrders();
      } else {
        message.error(error || '创建充值订单失败');
      }
    } catch {
      // no-op: 表单校验错误不弹全局错误
    }
  };

  const handleCopyKey = (key: string) => {
    navigator.clipboard.writeText(key);
    message.success('已复制到剪贴板');
  };

  const transactionTypeMap: Record<string, { color: string; text: string }> = {
    purchase: { color: 'green', text: '购买' },
    usage: { color: 'orange', text: '使用' },
    transfer: { color: 'blue', text: '转账' },
    reward: { color: 'purple', text: '奖励' },
    refund: { color: 'cyan', text: '退款' },
    recharge: { color: 'geekblue', text: '充值' },
  };

  const transactionColumns = [
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 100,
      render: (type: string) => {
        const { color, text } = transactionTypeMap[type] || { color: 'default', text: type };
        return <Tag color={color}>{text}</Tag>;
      },
    },
    {
      title: '金额',
      dataIndex: 'amount',
      key: 'amount',
      width: 120,
      render: (amount: number) => (
        <span style={{ color: amount > 0 ? '#52c41a' : '#ff4d4f' }}>
          {amount > 0 ? '+' : ''}
          {amount.toLocaleString()}
        </span>
      ),
    },
    {
      title: '说明',
      dataIndex: 'reason',
      key: 'reason',
      render: (reason: string) => reason || '-',
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date: string) => new Date(date).toLocaleString('zh-CN'),
    },
  ];

  const apiKeyColumns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '密钥',
      dataIndex: 'key_preview',
      key: 'key_preview',
      render: (key: string) => (
        <Text code copyable={{ text: key, onCopy: () => message.success('已复制') }}>
          {key || '••••••••••••'}
        </Text>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={status === 'active' ? 'success' : 'default'}>
          {status === 'active' ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '最后使用',
      dataIndex: 'last_used_at',
      key: 'last_used_at',
      width: 180,
      render: (date: string) => (date ? new Date(date).toLocaleString('zh-CN') : '从未使用'),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date: string) => new Date(date).toLocaleDateString('zh-CN'),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_: unknown, record: UserAPIKey) => (
        <Popconfirm
          title="确定要删除这个API密钥吗？"
          onConfirm={() => handleDeleteKey(record.id)}
          okText="确定"
          cancelText="取消"
        >
          <Button type="link" size="small" danger icon={<DeleteOutlined />}>
            删除
          </Button>
        </Popconfirm>
      ),
    },
  ];

  const rechargeOrderColumns: ColumnsType<RechargeOrder> = [
    {
      title: '订单号',
      dataIndex: 'out_trade_no',
      key: 'out_trade_no',
      ellipsis: true,
    },
    {
      title: '金额',
      dataIndex: 'amount',
      key: 'amount',
      width: 120,
      render: (amount: number) => <span>{amount.toLocaleString()}</span>,
    },
    {
      title: '支付方式',
      dataIndex: 'payment_method',
      key: 'payment_method',
      width: 120,
      render: (method: string) => {
        const methodMap: Record<string, string> = { alipay: '支付宝', wechat: '微信', balance: '余额' };
        return methodMap[method] || method;
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (status: string) => {
        const color = status === 'success' ? 'success' : status === 'failed' ? 'error' : 'processing';
        const label = status === 'success' ? '成功' : status === 'failed' ? '失败' : '待支付';
        return <Tag color={color}>{label}</Tag>;
      },
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date: string) => new Date(date).toLocaleString('zh-CN'),
    },
  ];

  if (allowMockRecharge) {
    rechargeOrderColumns.push({
      title: '操作',
      key: 'actions',
      width: 140,
      fixed: 'right',
      render: (_: unknown, record) =>
        record.status === 'pending' ? (
          <Button type="link" size="small" onClick={() => handleMockPayOrder(record)}>
            模拟支付完成
          </Button>
        ) : (
          <Text type="secondary">—</Text>
        ),
    });
  }

  return (
    <div className={styles.myToken}>
      <h2 className={styles.pageTitle}>我的Token</h2>

      <Row gutter={[16, 16]} className={styles.statsRow}>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="当前余额"
              value={balance?.balance || 0}
              precision={2}
              prefix={<WalletOutlined />}
              suffix="Token"
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="累计使用"
              value={balance?.total_used || 0}
              precision={2}
              prefix={<HistoryOutlined />}
              suffix="Token"
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="累计获得"
              value={balance?.total_earned || 0}
              precision={2}
              prefix={<WalletOutlined />}
              suffix="Token"
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
      </Row>

      <Card className={styles.actionCard}>
        <Space wrap size="middle" className={styles.actionBar}>
          <Button
            type="primary"
            icon={<WalletOutlined />}
            onClick={() => {
              clearError();
              setRechargeModalVisible(true);
            }}
          >
            充值
          </Button>
          <Button
            type="primary"
            icon={<SendOutlined />}
            onClick={() => setTransferModalVisible(true)}
          >
            转账
          </Button>
          <Button
            icon={<ReloadOutlined />}
            onClick={() => {
              fetchBalance();
              fetchTransactions();
              fetchAPIKeys();
              fetchRechargeOrders();
            }}
          >
            刷新
          </Button>
        </Space>
      </Card>

      <Card className={styles.tabsCard}>
        <Tabs defaultActiveKey="transactions">
          <TabPane
            tab={
              <span>
                <HistoryOutlined /> 交易记录
              </span>
            }
            key="transactions"
          >
            <div className={styles.tableScroll}>
              <Table
                columns={transactionColumns}
                dataSource={transactions}
                rowKey="id"
                loading={isLoading}
                pagination={{ pageSize: 10 }}
                scroll={{ x: 'max-content' }}
                locale={{ emptyText: <Empty description="暂无交易记录" /> }}
              />
            </div>
          </TabPane>

          <TabPane
            tab={
              <span>
                <WalletOutlined /> 充值订单
              </span>
            }
            key="recharge"
          >
            {allowMockRecharge ? (
              <Alert
                type="info"
                showIcon
                style={{ marginBottom: 12 }}
                message="当前为测试/模拟支付模式：待支付订单可使用「模拟支付完成」（需后端 ALLOW_TEST_RECHARGE=true）。"
              />
            ) : (
              <Alert
                type="warning"
                showIcon
                style={{ marginBottom: 12 }}
                message="正式支付需对接支付宝/微信等渠道。本地或测试环境可在服务端设置 ALLOW_TEST_RECHARGE=true、构建前端时设置 VITE_ALLOW_MOCK_RECHARGE=true 后使用模拟支付。"
              />
            )}
            <div className={styles.tableScroll}>
              <Table
                columns={rechargeOrderColumns}
                dataSource={rechargeOrders}
                rowKey="id"
                loading={isLoading}
                pagination={{ pageSize: 10 }}
                scroll={{ x: 'max-content' }}
                locale={{ emptyText: <Empty description="暂无充值订单" /> }}
              />
            </div>
          </TabPane>

          <TabPane
            tab={
              <span>
                <ApiOutlined /> API密钥
              </span>
            }
            key="apikeys"
          >
            <Paragraph type="secondary" className={styles.hintParagraph}>
              此处生成的是本平台调用接口用的访问密钥（前缀一般为 <Text code>ptd_</Text>），用于请求拼坨陀 API（含模型代理等），
              <strong>不是</strong> OpenAI 等厂商控制台里的 <Text code>sk-...</Text>。调厂商接口由平台侧配置。
            </Paragraph>
            <Paragraph type="secondary" className={styles.hintParagraph}>
              <strong>OpenAI 兼容：</strong>在支持自定义 Base URL 的客户端中，将 Base URL 设为{' '}
              <Text code>{openAICompatBase}</Text>，API Key 填本平台密钥；请求 <Text code>POST /chat/completions</Text>。
              模型可写 <Text code>厂商/模型</Text>（如 <Text code>zhipu/glm-4-flash</Text>）或常见模型名（如{' '}
              <Text code>gpt-3.5-turbo</Text>，将按名称推断厂商）。流式（stream）暂未支持。
            </Paragraph>
            <div className={styles.tabHeader}>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => setCreateKeyModalVisible(true)}
              >
                创建密钥
              </Button>
            </div>
            <div className={styles.tableScroll}>
              <Table
                columns={apiKeyColumns}
                dataSource={apiKeys}
                rowKey="id"
                loading={isLoading}
                pagination={false}
                scroll={{ x: 'max-content' }}
                locale={{ emptyText: <Empty description="暂无API密钥" /> }}
              />
            </div>
          </TabPane>
        </Tabs>
      </Card>

      <Modal
        title="创建API密钥"
        open={createKeyModalVisible}
        onOk={() => {
          if (newKeyDisplay) {
            setCreateKeyModalVisible(false);
            setNewKeyDisplay(null);
            return;
          }
          handleCreateKey();
        }}
        onCancel={() => {
          setCreateKeyModalVisible(false);
          keyForm.resetFields();
          setNewKeyDisplay(null);
        }}
        okText={newKeyDisplay ? '我已保存' : '创建'}
        cancelText="取消"
      >
        {newKeyDisplay ? (
          <div className={styles.newKeyDisplay}>
            <Paragraph type="warning">请妥善保存以下密钥，关闭后将无法再次查看完整密钥：</Paragraph>
            <Paragraph copyable={{ onCopy: () => handleCopyKey(newKeyDisplay) }}>
              <Text code className={styles.keyText}>
                {newKeyDisplay}
              </Text>
            </Paragraph>
          </div>
        ) : (
          <Form form={keyForm} layout="vertical">
            <Form.Item
              name="name"
              label="密钥名称"
              rules={[{ required: true, message: '请输入密钥名称' }]}
            >
              <Input placeholder="例如：开发环境密钥" />
            </Form.Item>
          </Form>
        )}
      </Modal>

      <Modal
        title="Token充值"
        open={rechargeModalVisible}
        width={520}
        centered
        styles={{ body: { maxHeight: '70vh', overflowY: 'auto' } }}
        wrapClassName={styles.modalWrap}
        onOk={handleCreateRechargeOrder}
        onCancel={() => {
          setRechargeModalVisible(false);
          rechargeForm.resetFields();
        }}
        okText="创建充值订单"
        cancelText="取消"
      >
        <Form form={rechargeForm} layout="vertical" initialValues={{ method: 'alipay' }}>
          <Form.Item
            name="amount"
            label="充值金额"
            rules={[
              { required: true, message: '请输入充值金额' },
              { type: 'number', min: 0.01, message: '金额必须大于0' },
            ]}
          >
            <InputNumber min={0.01} precision={2} style={{ width: '100%' }} placeholder="请输入充值金额" />
          </Form.Item>
          <Form.Item name="method" label="支付方式" rules={[{ required: true, message: '请选择支付方式' }]}>
            <Select
              options={[
                { label: '支付宝', value: 'alipay' },
                { label: '微信', value: 'wechat' },
                { label: '余额', value: 'balance' },
              ]}
            />
          </Form.Item>
          <Paragraph type="secondary">
            创建订单后会出现在「充值订单」列表。正式环境需跳转第三方支付；测试环境可同时开启服务端 ALLOW_TEST_RECHARGE 与前端
            VITE_ALLOW_MOCK_RECHARGE 后在列表中使用「模拟支付完成」。
          </Paragraph>
        </Form>
      </Modal>

      <Modal
        title="转账"
        open={transferModalVisible}
        width={520}
        centered
        styles={{ body: { maxHeight: '70vh', overflowY: 'auto' } }}
        wrapClassName={styles.modalWrap}
        onOk={handleTransfer}
        onCancel={() => {
          setTransferModalVisible(false);
          transferForm.resetFields();
        }}
        okText="确认转账"
        cancelText="取消"
      >
        <Form form={transferForm} layout="vertical">
          <Form.Item
            name="recipient"
            label="接收方"
            rules={[{ required: true, message: '请输入对方注册用邮箱或数字用户ID' }]}
          >
            <Input placeholder="例：user@example.com 或 数字用户ID" allowClear autoComplete="off" />
          </Form.Item>
          <Form.Item
            name="amount"
            label="转账金额"
            rules={[
              { required: true, message: '请输入转账金额' },
              { type: 'number', min: 0.01, message: '金额必须大于0' },
            ]}
          >
            <InputNumber
              min={0.01}
              precision={2}
              style={{ width: '100%' }}
              placeholder="请输入转账金额"
              suffix="Token"
            />
          </Form.Item>
          <Paragraph type="secondary">
            当前余额：{balance?.balance?.toLocaleString() || 0} Token
          </Paragraph>
        </Form>
      </Modal>
    </div>
  );
};

export default MyToken;
