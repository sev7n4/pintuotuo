import { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Tag,
  Button,
  Modal,
  Descriptions,
  message,
  Row,
  Col,
  Statistic,
  Empty,
  Form,
  Input,
  Select,
  Space,
} from 'antd';
import {
  FileTextOutlined,
  DollarOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  UploadOutlined,
  DownloadOutlined,
  PrinterOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';
import styles from './MerchantInvoices.module.css';

interface Invoice {
  id: string;
  invoice_number: string;
  amount: number;
  tax_amount: number;
  status: 'pending' | 'submitted' | 'approved' | 'rejected';
  type: 'normal' | 'special';
  created_at: string;
  updated_at?: string;
  remark?: string;
  file_url?: string;
}

const MerchantInvoices = () => {
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [loading, setLoading] = useState(false);
  const [detailVisible, setDetailVisible] = useState(false);
  const [selectedInvoice, setSelectedInvoice] = useState<Invoice | null>(null);
  const [applyVisible, setApplyVisible] = useState(false);
  const [applyForm] = Form.useForm();

  useEffect(() => {
    fetchInvoices();
  }, []);

  const fetchInvoices = async () => {
    setLoading(true);
    await new Promise((resolve) => setTimeout(resolve, 500));
    const mockInvoices: Invoice[] = [
      {
        id: 'INV-001',
        invoice_number: 'INV-2026-03-001',
        amount: 24396.0,
        tax_amount: 0,
        status: 'approved',
        type: 'normal',
        created_at: '2026-03-25T10:00:00Z',
        file_url: '/invoices/INV-2026-03-001.pdf',
      },
      {
        id: 'INV-002',
        invoice_number: 'INV-2026-02-001',
        amount: 21232.5,
        tax_amount: 0,
        status: 'approved',
        type: 'normal',
        created_at: '2026-02-28T14:00:00Z',
        file_url: '/invoices/INV-2026-02-001.pdf',
      },
      {
        id: 'INV-003',
        invoice_number: 'INV-2026-01-001',
        amount: 17974.0,
        tax_amount: 0,
        status: 'approved',
        type: 'normal',
        created_at: '2026-01-28T10:00:00Z',
        file_url: '/invoices/INV-2026-01-001.pdf',
      },
      {
        id: 'INV-004',
        invoice_number: 'INV-2026-03-002',
        amount: 15600.0,
        tax_amount: 0,
        status: 'pending',
        type: 'special',
        created_at: '2026-03-20T09:00:00Z',
      },
    ];
    setInvoices(mockInvoices);
    setLoading(false);
  };

  const handleViewDetail = (invoice: Invoice) => {
    setSelectedInvoice(invoice);
    setDetailVisible(true);
  };

  const handleApplyInvoice = () => {
    setApplyVisible(true);
  };

  const submitApply = async () => {
    try {
      await applyForm.validateFields();
      message.loading({ content: '正在提交申请...', duration: 0 });
      await new Promise((resolve) => setTimeout(resolve, 1000));
      message.success('发票申请已提交');
      setApplyVisible(false);
      applyForm.resetFields();
      fetchInvoices();
    } catch {
      message.error('提交失败');
    }
  };

  const handleDownload = (invoice: Invoice) => {
    if (invoice.file_url) {
      message.info('正在下载发票...');
    } else {
      message.warning('发票文件暂未生成');
    }
  };

  const handlePrint = () => {
    message.info('正在打印发票...');
  };

  const statusMap: Record<string, { color: string; text: string; icon: React.ReactNode }> = {
    pending: { color: 'default', text: '待开票', icon: <ClockCircleOutlined /> },
    submitted: { color: 'processing', text: '已提交', icon: <FileTextOutlined /> },
    approved: { color: 'success', text: '已审核', icon: <CheckCircleOutlined /> },
    rejected: { color: 'error', text: '已驳回', icon: <CloseCircleOutlined /> },
  };

  const typeMap: Record<string, { color: string; text: string }> = {
    normal: { color: 'blue', text: '普通发票' },
    special: { color: 'purple', text: '专用发票' },
  };

  const columns = [
    {
      title: '发票编号',
      dataIndex: 'invoice_number',
      key: 'invoice_number',
      width: 160,
    },
    {
      title: '金额',
      dataIndex: 'amount',
      key: 'amount',
      render: (v: number) => `¥${v.toFixed(2)}`,
    },
    {
      title: '税额',
      dataIndex: 'tax_amount',
      key: 'tax_amount',
      render: (v: number) => (v > 0 ? `¥${v.toFixed(2)}` : '-'),
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 100,
      render: (type: string) => <Tag color={typeMap[type]?.color}>{typeMap[type]?.text}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => {
        const { color, text, icon } = statusMap[status] || {
          color: 'default',
          text: status,
          icon: null,
        };
        return (
          <Tag color={color} icon={icon}>
            {text}
          </Tag>
        );
      },
    },
    {
      title: '申请时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleDateString('zh-CN'),
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
      render: (_: unknown, record: Invoice) => (
        <Space>
          <Button type="link" size="small" onClick={() => handleViewDetail(record)}>
            详情
          </Button>
          {record.status === 'approved' && (
            <>
              <Button
                type="link"
                size="small"
                icon={<DownloadOutlined />}
                onClick={() => handleDownload(record)}
              >
                下载
              </Button>
              <Button type="link" size="small" icon={<PrinterOutlined />} onClick={handlePrint}>
                打印
              </Button>
            </>
          )}
        </Space>
      ),
    },
  ];

  const totalApproved = invoices
    .filter((i) => i.status === 'approved')
    .reduce((sum, i) => sum + i.amount, 0);
  const totalPending = invoices.filter((i) => i.status === 'pending').length;

  return (
    <div className={styles.invoices}>
      <h2 className={styles.pageTitle}>发票管理</h2>

      <Row gutter={[16, 16]} className={styles.statsRow}>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="累计开票金额"
              value={totalApproved}
              precision={2}
              prefix={<DollarOutlined />}
              valueStyle={{ color: '#52c41a' }}
              suffix="元"
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="待开票"
              value={totalPending}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic title="发票总数" value={invoices.length} suffix="张" />
          </Card>
        </Col>
      </Row>

      <Card className={styles.tableCard}>
        <div className={styles.cardHeader}>
          <span>发票列表</span>
          <Button type="primary" icon={<UploadOutlined />} onClick={handleApplyInvoice}>
            申请发票
          </Button>
        </div>
        <Table
          columns={columns}
          dataSource={invoices}
          rowKey="id"
          loading={loading}
          scroll={{ x: 900 }}
          pagination={{ pageSize: 10 }}
          locale={{ emptyText: <Empty description="暂无发票数据" /> }}
        />
      </Card>

      <Modal
        title="发票详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={600}
      >
        {selectedInvoice && (
          <Descriptions column={2} bordered>
            <Descriptions.Item label="发票编号">{selectedInvoice.invoice_number}</Descriptions.Item>
            <Descriptions.Item label="发票类型">
              <Tag color={typeMap[selectedInvoice.type]?.color}>
                {typeMap[selectedInvoice.type]?.text}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="金额" span={2}>
              ¥{selectedInvoice.amount.toFixed(2)}
            </Descriptions.Item>
            <Descriptions.Item label="税额">
              {selectedInvoice.tax_amount > 0 ? `¥${selectedInvoice.tax_amount.toFixed(2)}` : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag
                color={statusMap[selectedInvoice.status]?.color}
                icon={statusMap[selectedInvoice.status]?.icon}
              >
                {statusMap[selectedInvoice.status]?.text}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="申请时间">
              {new Date(selectedInvoice.created_at).toLocaleString('zh-CN')}
            </Descriptions.Item>
            <Descriptions.Item label="更新时间">
              {selectedInvoice.updated_at
                ? new Date(selectedInvoice.updated_at).toLocaleString('zh-CN')
                : '-'}
            </Descriptions.Item>
            {selectedInvoice.remark && (
              <Descriptions.Item label="备注" span={2}>
                {selectedInvoice.remark}
              </Descriptions.Item>
            )}
          </Descriptions>
        )}
      </Modal>

      <Modal
        title="申请发票"
        open={applyVisible}
        onCancel={() => setApplyVisible(false)}
        onOk={submitApply}
        okText="提交申请"
      >
        <Form form={applyForm} layout="vertical">
          <Form.Item
            name="type"
            label="发票类型"
            rules={[{ required: true, message: '请选择发票类型' }]}
          >
            <Select placeholder="请选择发票类型">
              <Select.Option value="normal">普通发票</Select.Option>
              <Select.Option value="special">专用发票</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="amount"
            label="开票金额"
            rules={[{ required: true, message: '请输入开票金额' }]}
          >
            <Input type="number" prefix="¥" placeholder="请输入金额" />
          </Form.Item>
          <Form.Item name="remark" label="备注">
            <Input.TextArea placeholder="请输入备注信息" rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default MerchantInvoices;
