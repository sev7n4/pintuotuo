import { useEffect, useState } from 'react';
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
} from 'antd';
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
import { UserAPIKey } from '@/types';
import styles from './MyToken.module.css';

const { TabPane } = Tabs;
const { Text, Paragraph } = Typography;

const MyToken = () => {
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
      const success = await transfer(values.recipient_id, values.amount);
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

  const rechargeOrderColumns = [
    {
      title: '订单号',
      dataIndex: 'out_trade_no',
      key: 'out_trade_no',
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
        <Space>
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
            <Table
              columns={transactionColumns}
              dataSource={transactions}
              rowKey="id"
              loading={isLoading}
              pagination={{ pageSize: 10 }}
              locale={{ emptyText: <Empty description="暂无交易记录" /> }}
            />
          </TabPane>

          <TabPane
            tab={
              <span>
                <WalletOutlined /> 充值订单
              </span>
            }
            key="recharge"
          >
            <Table
              columns={rechargeOrderColumns}
              dataSource={rechargeOrders}
              rowKey="id"
              loading={isLoading}
              pagination={{ pageSize: 10 }}
              locale={{ emptyText: <Empty description="暂无充值订单" /> }}
            />
          </TabPane>

          <TabPane
            tab={
              <span>
                <ApiOutlined /> API密钥
              </span>
            }
            key="apikeys"
          >
            <div className={styles.tabHeader}>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => setCreateKeyModalVisible(true)}
              >
                创建密钥
              </Button>
            </div>
            <Table
              columns={apiKeyColumns}
              dataSource={apiKeys}
              rowKey="id"
              loading={isLoading}
              pagination={false}
              locale={{ emptyText: <Empty description="暂无API密钥" /> }}
            />
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
            创建订单后会进入“充值订单”列表。支付成功后系统会通过回调更新余额。
          </Paragraph>
        </Form>
      </Modal>

      <Modal
        title="转账"
        open={transferModalVisible}
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
            name="recipient_id"
            label="接收方用户ID"
            rules={[{ required: true, message: '请输入接收方用户ID' }]}
          >
            <InputNumber min={1} style={{ width: '100%' }} placeholder="请输入用户ID" />
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
