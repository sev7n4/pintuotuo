import { useCallback, useEffect, useState } from 'react';
import {
  Card,
  Tabs,
  Table,
  Statistic,
  Row,
  Col,
  Button,
  DatePicker,
  Space,
  Tag,
  message,
  Typography,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  AccountBookOutlined,
  ReloadOutlined,
  PlayCircleOutlined,
  LineChartOutlined,
  DownloadOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import {
  reconciliationService,
  type LedgerReconciliation,
  type UsageDriftRow,
  type GMVReport,
} from '@/services/reconciliation';
import { formatLedgerUnits } from '@/utils/ledgerDisplay';
import GMVTrendChart from '@/components/GMVTrendChart';

const { RangePicker } = DatePicker;
const { Paragraph, Text } = Typography;

const AdminReconciliation = () => {
  const [ledger, setLedger] = useState<LedgerReconciliation | null>(null);
  const [ledgerLoading, setLedgerLoading] = useState(false);
  const [drift, setDrift] = useState<UsageDriftRow[]>([]);
  const [driftTotal, setDriftTotal] = useState(0);
  const [driftPage, setDriftPage] = useState(1);
  const [driftPageSize, setDriftPageSize] = useState(20);
  const [driftLoading, setDriftLoading] = useState(false);
  const [gmv, setGmv] = useState<GMVReport | null>(null);
  const [gmvLoading, setGmvLoading] = useState(false);
  const [gmvRange, setGmvRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>([
    dayjs().subtract(89, 'day'),
    dayjs(),
  ]);

  const fetchGMV = useCallback(async (range: [dayjs.Dayjs, dayjs.Dayjs] | null) => {
    setGmvLoading(true);
    try {
      const params: { start_date?: string; end_date?: string } = {};
      if (range) {
        params.start_date = range[0].format('YYYY-MM-DD');
        params.end_date = range[1].format('YYYY-MM-DD');
      }
      const { data } = await reconciliationService.getGMV(
        Object.keys(params).length ? params : undefined
      );
      setGmv(data);
    } catch {
      message.error('加载 GMV 报表失败');
    } finally {
      setGmvLoading(false);
    }
  }, []);

  const loadLedger = useCallback(async () => {
    setLedgerLoading(true);
    try {
      const { data } = await reconciliationService.getLedger();
      setLedger(data);
    } catch {
      message.error('加载全库用量对账失败');
    } finally {
      setLedgerLoading(false);
    }
  }, []);

  const loadDrift = useCallback(async (page: number, pageSize: number) => {
    setDriftLoading(true);
    try {
      const { data } = await reconciliationService.getDrift({ page, page_size: pageSize });
      setDrift(data.data || []);
      setDriftTotal(data.total);
      setDriftPage(data.page);
      setDriftPageSize(data.page_size);
    } catch {
      message.error('加载差异用户失败');
    } finally {
      setDriftLoading(false);
    }
  }, []);

  const exportDrift = async () => {
    try {
      const res = await reconciliationService.exportDriftCSV();
      const blob = res.data;
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `ledger_drift_${dayjs().format('YYYYMMDD_HHmmss')}.csv`;
      a.click();
      window.URL.revokeObjectURL(url);
      message.success('已导出（最多 5 万行）');
    } catch {
      message.error('导出失败');
    }
  };

  const runJob = async () => {
    setLedgerLoading(true);
    try {
      const { data } = await reconciliationService.postLedgerCheck();
      setLedger(data);
      message.success('对账任务已执行（结果见上方）');
      await loadDrift(driftPage, driftPageSize);
    } catch {
      message.error('对账任务失败');
    } finally {
      setLedgerLoading(false);
    }
  };

  useEffect(() => {
    void loadLedger();
  }, [loadLedger]);

  useEffect(() => {
    void loadDrift(1, 20);
  }, [loadDrift]);

  useEffect(() => {
    void fetchGMV(gmvRange);
  }, [fetchGMV, gmvRange]);

  const driftColumns: ColumnsType<UsageDriftRow> = [
    { title: '用户 ID', dataIndex: 'user_id', width: 100 },
    {
      title: '用量日志合计',
      dataIndex: 'log_sum',
      render: (v: number) => formatLedgerUnits(v),
    },
    {
      title: '流水合计',
      dataIndex: 'tx_sum',
      render: (v: number) => formatLedgerUnits(v),
    },
    {
      title: '差值',
      dataIndex: 'delta',
      render: (v: number) => (
        <Text type={Math.abs(v) > 1e-3 ? 'danger' : undefined}>{formatLedgerUnits(v)}</Text>
      ),
    },
  ];

  const ledgerTab = (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Paragraph type="secondary">
        全库汇总：<Text code>api_usage_logs</Text> 中每行计费 Token 量（
        <Text code>COALESCE(token_usage, input_tokens + output_tokens)</Text>
        ）之和，应与 <Text code>token_transactions</Text> 中 <Text code>type=usage</Text>{' '}
        扣减绝对值之和一致；单位为平台 Token，与「我的 Token」扣费同口径（非人民币 cost）。
      </Paragraph>
      <Row gutter={16}>
        <Col xs={24} sm={8}>
          <Card loading={ledgerLoading}>
            <Statistic
              title="用量日志合计（Token）"
              value={ledger?.usage_log_total ?? 0}
              precision={6}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card loading={ledgerLoading}>
            <Statistic
              title="流水扣减合计（Token）"
              value={ledger?.usage_tx_total ?? 0}
              precision={6}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card loading={ledgerLoading}>
            <Statistic
              title="差值"
              value={ledger?.delta ?? 0}
              precision={6}
              valueStyle={{
                color: ledger?.matched === false ? '#cf1322' : undefined,
              }}
              prefix={
                ledger?.matched === false ? (
                  <Tag color="error">不一致</Tag>
                ) : (
                  <Tag color="success">一致</Tag>
                )
              }
            />
          </Card>
        </Col>
      </Row>
      <Space wrap>
        <Button icon={<ReloadOutlined />} onClick={() => void loadLedger()} loading={ledgerLoading}>
          刷新汇总
        </Button>
        <Button
          type="primary"
          icon={<PlayCircleOutlined />}
          onClick={() => void runJob()}
          loading={ledgerLoading}
        >
          执行对账任务
        </Button>
        <Button icon={<DownloadOutlined />} onClick={() => void exportDrift()}>
          导出差异用户 CSV
        </Button>
        <Text type="secondary">{ledger?.checked_at ? `上次检查：${ledger.checked_at}` : ''}</Text>
      </Space>

      <Card title="存在差异的用户（按 |差值| 降序）">
        <Table
          rowKey="user_id"
          loading={driftLoading}
          columns={driftColumns}
          dataSource={drift}
          pagination={{
            current: driftPage,
            pageSize: driftPageSize,
            total: driftTotal,
            showSizeChanger: true,
            showTotal: (t) => `共 ${t} 个用户`,
            onChange: (p, ps) => void loadDrift(p, ps || 20),
          }}
          scroll={{ x: 520 }}
        />
      </Card>
    </Space>
  );

  const gmvTab = (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Paragraph type="secondary">
        以下为订单维度 <Text code>orders.total_price</Text> 汇总（状态
        paid/completed），单位人民币（CNY），与内部 Token
        用量报表相互独立。未选日期区间时汇总为全量；曲线未选区间时默认近 90 日（与后端一致）。
      </Paragraph>
      <Space wrap>
        <RangePicker
          value={gmvRange}
          onChange={(d) => setGmvRange(d as [dayjs.Dayjs, dayjs.Dayjs] | null)}
          allowClear
        />
        <Button
          icon={<ReloadOutlined />}
          onClick={() => void fetchGMV(gmvRange)}
          loading={gmvLoading}
        >
          查询
        </Button>
      </Space>
      <Row gutter={16}>
        <Col xs={24} sm={12}>
          <Card loading={gmvLoading}>
            <Statistic title="订单笔数" value={gmv?.order_count ?? 0} />
          </Card>
        </Col>
        <Col xs={24} sm={12}>
          <Card loading={gmvLoading}>
            <Statistic
              title="GMV（人民币）"
              value={gmv?.gmv_cny ?? 0}
              precision={2}
              prefix="¥"
              suffix={gmv?.currency ? ` ${gmv.currency}` : ''}
            />
          </Card>
        </Col>
      </Row>
      <GMVTrendChart dateRange={gmvRange} />
    </Space>
  );

  return (
    <div style={{ padding: 24 }}>
      <Typography.Title level={4} style={{ marginBottom: 16 }}>
        <AccountBookOutlined /> 对账与 GMV
      </Typography.Title>
      <Tabs
        defaultActiveKey="ledger"
        items={[
          {
            key: 'ledger',
            label: (
              <>
                <LineChartOutlined /> 全库用量对账
              </>
            ),
            children: ledgerTab,
          },
          {
            key: 'gmv',
            label: (
              <>
                <AccountBookOutlined /> 人民币 GMV
              </>
            ),
            children: gmvTab,
          },
        ]}
      />
    </div>
  );
};

export default AdminReconciliation;
