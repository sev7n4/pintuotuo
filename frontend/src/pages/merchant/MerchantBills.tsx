import { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Tag,
  Button,
  Modal,
  Descriptions,
  Row,
  Col,
  Statistic,
  DatePicker,
  Tabs,
  Empty,
  message,
  Select,
} from 'antd';
import {
  DollarOutlined,
  ClockCircleOutlined,
  DownloadOutlined,
  FileTextOutlined,
  CalendarOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import styles from './MerchantBills.module.css';

const { RangePicker } = DatePicker;

interface BillDetail {
  date: string;
  orders: number;
  sales: number;
  refund: number;
  net_sales: number;
  endpoint_type?: string;
  unit_type?: string;
  unit_count?: number;
}

interface MonthlyBill {
  id: string;
  period: string;
  year: number;
  month: number;
  total_sales: number;
  total_orders: number;
  platform_fee: number;
  settlement_amount: number;
  status: 'pending' | 'confirmed' | 'settled';
  created_at: string;
  details: BillDetail[];
}

const MerchantBills = () => {
  const [bills, setBills] = useState<MonthlyBill[]>([]);
  const [loading, setLoading] = useState(false);
  const [detailVisible, setDetailVisible] = useState(false);
  const [selectedBill, setSelectedBill] = useState<MonthlyBill | null>(null);
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);

  useEffect(() => {
    fetchBills();
  }, []);

  const fetchBills = async () => {
    setLoading(true);
    await new Promise((resolve) => setTimeout(resolve, 500));
    const mockBills: MonthlyBill[] = [
      {
        id: 'BILL-2026-03',
        period: '2026年3月',
        year: 2026,
        month: 3,
        total_sales: 25680.0,
        total_orders: 156,
        platform_fee: 1284.0,
        settlement_amount: 24396.0,
        status: 'settled',
        created_at: '2026-04-01T00:00:00Z',
        details: generateBillDetails(2026, 3),
      },
      {
        id: 'BILL-2026-02',
        period: '2026年2月',
        year: 2026,
        month: 2,
        total_sales: 22350.0,
        total_orders: 132,
        platform_fee: 1117.5,
        settlement_amount: 21232.5,
        status: 'settled',
        created_at: '2026-03-01T00:00:00Z',
        details: generateBillDetails(2026, 2),
      },
      {
        id: 'BILL-2026-01',
        period: '2026年1月',
        year: 2026,
        month: 1,
        total_sales: 18920.0,
        total_orders: 98,
        platform_fee: 946.0,
        settlement_amount: 17974.0,
        status: 'settled',
        created_at: '2026-02-01T00:00:00Z',
        details: generateBillDetails(2026, 1),
      },
    ];
    setBills(mockBills);
    setLoading(false);
  };

  const generateBillDetails = (year: number, month: number): BillDetail[] => {
    const daysInMonth = new Date(year, month, 0).getDate();
    const details: BillDetail[] = [];
    for (let i = 1; i <= daysInMonth; i++) {
      const orders = Math.floor(Math.random() * 10) + 1;
      const sales = Math.random() * 2000 + 500;
      const refund = Math.random() * 100;
      details.push({
        date: `${year}-${String(month).padStart(2, '0')}-${String(i).padStart(2, '0')}`,
        orders,
        sales: Math.round(sales * 100) / 100,
        refund: Math.round(refund * 100) / 100,
        net_sales: Math.round((sales - refund) * 100) / 100,
      });
    }
    return details;
  };

  const handleViewDetail = (bill: MonthlyBill) => {
    setSelectedBill(bill);
    setDetailVisible(true);
  };

  const handleExport = (bill: MonthlyBill) => {
    const content = generateBillCSV(bill);
    const blob = new Blob([content], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = `账单_${bill.period}.csv`;
    link.click();
    message.success('账单导出成功');
  };

  const generateBillCSV = (bill: MonthlyBill): string => {
    let csv = '日期,订单数,销售额(¥),退款(¥),净销售额(¥)\n';
    bill.details.forEach((d) => {
      csv += `${d.date},${d.orders},${d.sales.toFixed(6)},${d.refund.toFixed(6)},${d.net_sales.toFixed(6)}\n`;
    });
    csv += `\n合计,,${bill.total_sales.toFixed(6)},,${bill.total_sales.toFixed(6)}\n`;
    csv += `平台费用(5%),,,${bill.platform_fee.toFixed(6)}\n`;
    csv += `结算金额,,,${bill.settlement_amount.toFixed(6)}\n`;
    return csv;
  };

  const statusMap: Record<string, { color: string; text: string; icon: React.ReactNode }> = {
    pending: { color: 'default', text: '待确认', icon: <ClockCircleOutlined /> },
    confirmed: { color: 'processing', text: '已确认', icon: <FileTextOutlined /> },
    settled: { color: 'success', text: '已结算', icon: <DollarOutlined /> },
  };

  const columns = [
    {
      title: '账单编号',
      dataIndex: 'id',
      key: 'id',
      width: 140,
    },
    {
      title: '账单周期',
      dataIndex: 'period',
      key: 'period',
      render: (period: string) => (
        <span>
          <CalendarOutlined style={{ marginRight: 8, color: '#1890ff' }} />
          {period}
        </span>
      ),
    },
    {
      title: '订单数',
      dataIndex: 'total_orders',
      key: 'total_orders',
      width: 100,
      render: (v: number) => `${v} 笔`,
    },
    {
      title: '销售总额(¥)',
      dataIndex: 'total_sales',
      key: 'total_sales',
      render: (v: number) => `¥${v.toFixed(6)}`,
    },
    {
      title: '平台费用',
      dataIndex: 'platform_fee',
      key: 'platform_fee',
      render: (v: number) => <span style={{ color: '#ff4d4f' }}>-¥{v.toFixed(6)}</span>,
    },
    {
      title: '结算金额',
      dataIndex: 'settlement_amount',
      key: 'settlement_amount',
      render: (v: number) => <span className={styles.amount}>¥{v.toFixed(6)}</span>,
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
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: unknown, record: MonthlyBill) => (
        <div className={styles.actionBtns}>
          <Button type="link" size="small" onClick={() => handleViewDetail(record)}>
            详情
          </Button>
          <Button
            type="link"
            size="small"
            icon={<DownloadOutlined />}
            onClick={() => handleExport(record)}
          >
            导出
          </Button>
        </div>
      ),
    },
  ];

  const detailColumns = [
    { title: '日期', dataIndex: 'date', key: 'date', width: 120 },
    {
      title: '订单数',
      dataIndex: 'orders',
      key: 'orders',
      width: 80,
      render: (v: number) => `${v} 笔`,
    },
    {
      title: '销售额(¥)',
      dataIndex: 'sales',
      key: 'sales',
      render: (v: number) => `¥${v.toFixed(6)}`,
    },
    {
      title: '退款',
      dataIndex: 'refund',
      key: 'refund',
      render: (v: number) => (v > 0 ? `-¥${v.toFixed(6)}` : '-'),
    },
    {
      title: '净销售额(¥)',
      dataIndex: 'net_sales',
      key: 'net_sales',
      render: (v: number) => `¥${v.toFixed(6)}`,
    },
    {
      title: '端点类型',
      dataIndex: 'endpoint_type',
      key: 'endpoint_type',
      width: 100,
      render: (v: string) => v ? <Tag color="blue">{v}</Tag> : '-',
    },
    {
      title: '计费单位',
      dataIndex: 'unit_type',
      key: 'unit_type',
      width: 80,
      render: (v: string) => v || '-',
    },
    {
      title: '单位数量',
      dataIndex: 'unit_count',
      key: 'unit_count',
      width: 80,
      render: (v: number) => v ?? '-',
    },
  ];

  const totalSettled = bills
    .filter((b) => b.status === 'settled')
    .reduce((sum, b) => sum + b.settlement_amount, 0);
  const totalPending = bills
    .filter((b) => b.status !== 'settled')
    .reduce((sum, b) => sum + b.settlement_amount, 0);

  const filteredBills = dateRange
    ? bills.filter((b) => {
        const billDate = dayjs(`${b.year}-${b.month}-01`);
        return (
          billDate.isAfter(dateRange[0].startOf('month')) &&
          billDate.isBefore(dateRange[1].endOf('month'))
        );
      })
    : bills;

  return (
    <div className={styles.bills}>
      <h2 className={styles.pageTitle}>月度账单</h2>

      <Row gutter={[16, 16]} className={styles.statsRow}>
        <Col xs={24} sm={12}>
          <Card>
            <Statistic
              title="累计结算"
              value={totalSettled}
              precision={2}
              prefix={<DollarOutlined />}
              suffix="元"
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12}>
          <Card>
            <Statistic
              title="待结算"
              value={totalPending}
              precision={2}
              prefix={<ClockCircleOutlined />}
              suffix="元"
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
      </Row>

      <Card className={styles.tableCard}>
        <div className={styles.cardHeader}>
          <RangePicker
            picker="month"
            onChange={(dates) => setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs] | null)}
            placeholder={['开始月份', '结束月份']}
          />
          <Select
            placeholder="端点类型"
            allowClear
            style={{ width: 140, marginLeft: 8 }}
            options={[
              { value: 'chat_completions', label: '对话补全' },
              { value: 'responses', label: 'Response API' },
              { value: 'embeddings', label: '嵌入' },
              { value: 'images_generations', label: '图像生成' },
              { value: 'audio_speech', label: '语音合成' },
              { value: 'moderations', label: '内容审核' },
            ]}
          />
        </div>
        <Table
          columns={columns}
          dataSource={filteredBills}
          rowKey="id"
          loading={loading}
          scroll={{ x: 900 }}
          pagination={{ pageSize: 10 }}
          locale={{ emptyText: <Empty description="暂无账单数据" /> }}
        />
      </Card>

      <Modal
        title={`账单详情 - ${selectedBill?.period}`}
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={900}
      >
        {selectedBill && (
          <div>
            <Descriptions column={4} bordered className={styles.billSummary}>
              <Descriptions.Item label="账单编号">{selectedBill.id}</Descriptions.Item>
              <Descriptions.Item label="账单周期">{selectedBill.period}</Descriptions.Item>
              <Descriptions.Item label="订单总数">{selectedBill.total_orders} 笔</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={statusMap[selectedBill.status]?.color}>
                  {statusMap[selectedBill.status]?.text}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="销售总额(¥)">
                ¥{selectedBill.total_sales.toFixed(6)}
              </Descriptions.Item>
              <Descriptions.Item label="平台费用(5%)">
                <span style={{ color: '#ff4d4f' }}>-¥{selectedBill.platform_fee.toFixed(6)}</span>
              </Descriptions.Item>
              <Descriptions.Item label="结算金额" span={2}>
                <span className={styles.amount}>¥{selectedBill.settlement_amount.toFixed(6)}</span>
              </Descriptions.Item>
            </Descriptions>

            <Tabs defaultActiveKey="daily" className={styles.detailTabs}>
              <Tabs.TabPane tab="每日明细" key="daily">
                <Table
                  columns={detailColumns}
                  dataSource={selectedBill.details}
                  rowKey="date"
                  pagination={{ pageSize: 10 }}
                  size="small"
                  scroll={{ y: 300 }}
                />
              </Tabs.TabPane>
              <Tabs.TabPane tab="汇总统计" key="summary">
                <Row gutter={[16, 16]}>
                  <Col span={8}>
                    <Card>
                      <Statistic
                        title="日均订单"
                        value={Math.round(selectedBill.total_orders / selectedBill.details.length)}
                        suffix="笔"
                      />
                    </Card>
                  </Col>
                  <Col span={8}>
                    <Card>
                      <Statistic
                        title="日均销售"
                        value={selectedBill.total_sales / selectedBill.details.length}
                        precision={2}
                        prefix="¥"
                      />
                    </Card>
                  </Col>
                  <Col span={8}>
                    <Card>
                      <Statistic
                        title="最高日销售"
                        value={Math.max(...selectedBill.details.map((d) => d.net_sales))}
                        precision={2}
                        prefix="¥"
                      />
                    </Card>
                  </Col>
                </Row>
              </Tabs.TabPane>
            </Tabs>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default MerchantBills;
