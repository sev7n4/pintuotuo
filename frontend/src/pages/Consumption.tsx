import React, { useState, useEffect, useMemo } from 'react';
import {
  Table,
  Card,
  DatePicker,
  Select,
  Button,
  Tag,
  Space,
  Statistic,
  Row,
  Col,
  Modal,
  Descriptions,
  Spin,
  message,
  Grid,
  Segmented,
  Typography,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { EyeOutlined, ReloadOutlined, DownloadOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip as RechartsTooltip,
  BarChart,
  Bar,
} from 'recharts';
import styles from './Consumption.module.css';
import { formatLedgerUnits, ledgerUnitColumnTitle, ledgerUnitShort } from '@/utils/ledgerDisplay';

const { useBreakpoint } = Grid;
const { RangePicker } = DatePicker;

interface ConsumptionRecord {
  id: number;
  request_id: string;
  provider: string;
  model: string;
  method: string;
  path: string;
  status_code: number;
  latency_ms: number;
  input_tokens: number;
  output_tokens: number;
  cost: number;
  created_at: string;
}

interface ConsumptionStats {
  total_requests: number;
  total_tokens: number;
  total_cost: number;
  avg_latency_ms: number;
}

interface ProviderStats {
  provider: string;
  count: number;
  cost: number;
}

const { Paragraph, Text } = Typography;

type MainView = 'records' | 'charts';

const Consumption: React.FC = () => {
  const [mainView, setMainView] = useState<MainView>('records');
  const [recordsLayout, setRecordsLayout] = useState<'table' | 'cards'>('table');
  const [loading, setLoading] = useState(false);
  const [records, setRecords] = useState<ConsumptionRecord[]>([]);
  const [stats, setStats] = useState<ConsumptionStats | null>(null);
  const [providerStats, setProviderStats] = useState<ProviderStats[]>([]);
  const [modelOptions, setModelOptions] = useState<{ value: string; label: string }[]>([]);
  const [detailVisible, setDetailVisible] = useState(false);
  const [selectedRecord, setSelectedRecord] = useState<ConsumptionRecord | null>(null);
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs]>([
    dayjs().subtract(30, 'days'),
    dayjs(),
  ]);
  const [provider, setProvider] = useState<string>('all');
  const [model, setModel] = useState<string>('');
  const screens = useBreakpoint();

  const isMobile = screens.xs || (screens.sm && !screens.md);

  useEffect(() => {
    fetchConsumptionData();
  }, [dateRange, provider, model]);

  const fetchConsumptionData = async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token');
      const params = new URLSearchParams({
        start_date: dateRange[0].format('YYYY-MM-DD'),
        end_date: dateRange[1].format('YYYY-MM-DD'),
        provider,
      });
      if (model.trim()) {
        params.set('model', model.trim());
      }

      const [recordsRes, statsRes] = await Promise.all([
        fetch(`/api/v1/consumption/records?${params}`, {
          headers: { Authorization: `Bearer ${token}` },
        }),
        fetch(`/api/v1/consumption/stats?${params}`, {
          headers: { Authorization: `Bearer ${token}` },
        }),
      ]);

      if (recordsRes.ok) {
        const data = await recordsRes.json();
        setRecords(data.data || []);
      }

      if (statsRes.ok) {
        const data = await statsRes.json();
        setStats(data.stats || null);
        setProviderStats(data.by_provider || []);
        const models = (data.models_in_range as string[] | undefined) || [];
        setModelOptions(models.map((m) => ({ value: m, label: m })));
      }
    } catch {
      message.error('获取消费数据失败');
    } finally {
      setLoading(false);
    }
  };

  const showDetail = (record: ConsumptionRecord) => {
    setSelectedRecord(record);
    setDetailVisible(true);
  };

  const exportData = () => {
    const csv = [
      [
        '请求ID',
        'Provider',
        'Model',
        '输入Tokens',
        '输出Tokens',
        '扣减(Token)',
        '延迟(ms)',
        '状态码',
        '时间',
      ].join(','),
      ...records.map((r) =>
        [
          r.request_id,
          r.provider,
          r.model,
          r.input_tokens,
          r.output_tokens,
          r.cost.toFixed(6),
          r.latency_ms,
          r.status_code,
          r.created_at,
        ].join(',')
      ),
    ].join('\n');

    const blob = new Blob([csv], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `consumption_${dayjs().format('YYYYMMDD_HHmmss')}.csv`;
    a.click();
    window.URL.revokeObjectURL(url);
    message.success('导出成功');
  };

  const getProviderColor = (provider: string) => {
    const colors: Record<string, string> = {
      openai: 'green',
      anthropic: 'purple',
      google: 'blue',
      azure: 'cyan',
    };
    return colors[provider] || 'default';
  };

  const columns: ColumnsType<ConsumptionRecord> = useMemo(
    () => [
      {
        title: '请求ID',
        dataIndex: 'request_id',
        key: 'request_id',
        width: 120,
        fixed: 'left',
        ellipsis: true,
        render: (text: string) => (
          <span style={{ fontFamily: 'monospace', fontSize: 12 }}>{text.slice(0, 8)}...</span>
        ),
      },
      {
        title: 'Provider',
        dataIndex: 'provider',
        key: 'provider',
        width: 100,
        render: (provider: string) => (
          <Tag color={getProviderColor(provider)}>{provider.toUpperCase()}</Tag>
        ),
      },
      ...(screens.md
        ? [
            {
              title: 'Model',
              dataIndex: 'model',
              key: 'model',
              width: 150,
              ellipsis: true,
            },
          ]
        : []),
      ...(screens.lg
        ? [
            {
              title: '输入',
              dataIndex: 'input_tokens',
              key: 'input_tokens',
              width: 80,
              align: 'right' as const,
              render: (v: number) => v.toLocaleString(),
            },
          ]
        : []),
      ...(screens.lg
        ? [
            {
              title: '输出',
              dataIndex: 'output_tokens',
              key: 'output_tokens',
              width: 80,
              align: 'right' as const,
              render: (v: number) => v.toLocaleString(),
            },
          ]
        : []),
      {
        title: ledgerUnitColumnTitle,
        dataIndex: 'cost',
        key: 'cost',
        width: 110,
        align: 'right',
        render: (cost: number) => (
          <span style={{ color: '#f5222d' }}>{formatLedgerUnits(cost)}</span>
        ),
      },
      ...(screens.sm
        ? [
            {
              title: '延迟',
              dataIndex: 'latency_ms',
              key: 'latency_ms',
              width: 70,
              align: 'right' as const,
              render: (v: number) => {
                const color = v < 1000 ? '#52c41a' : v < 3000 ? '#faad14' : '#f5222d';
                return <span style={{ color }}>{v}</span>;
              },
            },
          ]
        : []),
      {
        title: '状态',
        dataIndex: 'status_code',
        key: 'status_code',
        width: 70,
        align: 'center',
        render: (code: number) => {
          const color = code >= 200 && code < 300 ? 'success' : 'error';
          return <Tag color={color}>{code}</Tag>;
        },
      },
      ...(screens.xl
        ? [
            {
              title: '时间',
              dataIndex: 'created_at',
              key: 'created_at',
              width: 100,
              render: (time: string) => dayjs(time).format('MM-DD HH:mm'),
            },
          ]
        : []),
      {
        title: '操作',
        key: 'action',
        width: 60,
        fixed: 'right',
        render: (_, record) => (
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => showDetail(record)}
          >
            {!isMobile && '详情'}
          </Button>
        ),
      },
    ],
    [screens, isMobile]
  );

  const dailySeries = useMemo(() => {
    const m = new Map<string, number>();
    for (const r of records) {
      const d = dayjs(r.created_at).format('YYYY-MM-DD');
      m.set(d, (m.get(d) || 0) + r.cost);
    }
    return [...m.entries()]
      .map(([date, cost]) => ({ date, cost: Number(cost.toFixed(6)) }))
      .sort((a, b) => a.date.localeCompare(b.date));
  }, [records]);

  const providerBarData = useMemo(
    () => providerStats.map((p) => ({ name: p.provider, cost: p.cost, count: p.count })),
    [providerStats]
  );

  const providerSelectOptions = useMemo(() => {
    const base = [{ value: 'all', label: '全部 Provider' }];
    const seen = new Set<string>();
    const fromStats: { value: string; label: string }[] = [];
    for (const p of providerStats) {
      if (p.provider && !seen.has(p.provider)) {
        seen.add(p.provider);
        fromStats.push({ value: p.provider, label: p.provider.toUpperCase() });
      }
    }
    if (fromStats.length > 0) {
      return [...base, ...fromStats];
    }
    return [
      ...base,
      { value: 'openai', label: 'OpenAI' },
      { value: 'anthropic', label: 'Anthropic' },
      { value: 'google', label: 'Google' },
      { value: 'azure', label: 'Azure' },
    ];
  }, [providerStats]);

  const rangeMinutes = useMemo(() => {
    const end = dateRange[1];
    const start = dateRange[0];
    if (!end?.diff || !start) {
      return 1;
    }
    const m = end.diff(start, 'minute', true);
    return Math.max(m, 1 / 60);
  }, [dateRange]);

  const rpm = stats ? stats.total_requests / rangeMinutes : 0;
  const tpm = stats ? stats.total_tokens / rangeMinutes : 0;

  return (
    <div className={styles.container} style={{ padding: isMobile ? 12 : 24 }}>
      <div
        style={{
          marginBottom: 16,
          display: 'flex',
          flexWrap: 'wrap',
          gap: 12,
          alignItems: 'center',
        }}
      >
        <Segmented
          value={mainView}
          onChange={(v) => setMainView(v as MainView)}
          options={[
            { label: '明细列表', value: 'records' },
            { label: '图表视图', value: 'charts' },
          ]}
        />
        <Paragraph type="secondary" style={{ margin: 0, flex: '1 1 200px' }}>
          上方筛选项对列表与统计、图表共用；明细列表可切换表格/卡片，图表按日汇总扣减与按 Provider 对比。
        </Paragraph>
      </div>
      {mainView === 'records' && (
        <div style={{ marginBottom: 12 }}>
          <Segmented
            value={recordsLayout}
            onChange={(v) => setRecordsLayout(v as 'table' | 'cards')}
            options={[
              { label: '表格', value: 'table' },
              { label: '卡片', value: 'cards' },
            ]}
          />
        </div>
      )}
      <Card size="small" style={{ marginBottom: 16 }} title="筛选">
        <Space size="small" wrap>
          <RangePicker
            value={dateRange}
            onChange={(dates) => dates && setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs])}
            allowClear={false}
            size={isMobile ? 'small' : 'middle'}
          />
          <Select
            value={provider}
            onChange={setProvider}
            style={{ width: isMobile ? 120 : 140 }}
            size={isMobile ? 'small' : 'middle'}
            options={providerSelectOptions}
            showSearch
            optionFilterProp="label"
            placeholder="Provider"
          />
          <Select
            value={model || undefined}
            onChange={(v) => setModel(v ?? '')}
            allowClear
            showSearch
            placeholder="模型"
            style={{ width: isMobile ? 140 : 200 }}
            size={isMobile ? 'small' : 'middle'}
            options={modelOptions}
            notFoundContent={modelOptions.length ? undefined : '先选日期并刷新'}
          />
          <Button
            icon={<ReloadOutlined />}
            onClick={fetchConsumptionData}
            size={isMobile ? 'small' : 'middle'}
          >
            刷新
          </Button>
        </Space>
      </Card>
      <Card className={styles.statsCard}>
        <Row gutter={[16, 16]}>
          <Col xs={12} sm={12} md={6} lg={4}>
            <Statistic
              title="总请求数"
              value={stats?.total_requests || 0}
              suffix="次"
              valueStyle={{ fontSize: isMobile ? 18 : 24 }}
            />
          </Col>
          <Col xs={12} sm={12} md={6} lg={4}>
            <Statistic
              title="总 Tokens"
              value={stats?.total_tokens || 0}
              valueStyle={{ fontSize: isMobile ? 18 : 24 }}
            />
          </Col>
          <Col xs={12} sm={12} md={6} lg={4}>
            <Statistic
              title="合计扣减"
              value={stats?.total_cost || 0}
              suffix={ledgerUnitShort}
              valueStyle={{ color: '#f5222d', fontSize: isMobile ? 18 : 24 }}
            />
          </Col>
          <Col xs={12} sm={12} md={6} lg={4}>
            <Statistic
              title="平均延迟"
              value={stats?.avg_latency_ms || 0}
              suffix="ms"
              valueStyle={{ fontSize: isMobile ? 18 : 24 }}
            />
          </Col>
          <Col xs={12} sm={12} md={6} lg={4}>
            <Statistic
              title="RPM（均）"
              value={rpm}
              precision={2}
              suffix="/min"
              valueStyle={{ fontSize: isMobile ? 16 : 20 }}
            />
          </Col>
          <Col xs={12} sm={12} md={6} lg={4}>
            <Statistic
              title="TPM（均）"
              value={tpm}
              precision={0}
              suffix="/min"
              valueStyle={{ fontSize: isMobile ? 16 : 20 }}
            />
          </Col>
        </Row>
        <Paragraph type="secondary" style={{ marginTop: 12, marginBottom: 0, fontSize: 12 }}>
          总请求数/总 Tokens/合计扣减/平均延迟来自后端对 api_usage_logs 的聚合；RPM、TPM
          为当前日期范围内总请求数、总 Tokens 除以区间分钟数（均值，非峰值）。
        </Paragraph>
      </Card>

      {mainView !== 'charts' && providerStats.length > 0 && (
        <Card className={styles.providerCard} title="按Provider统计">
          <Row gutter={[16, 16]}>
            {providerStats.map((p) => (
              <Col xs={12} sm={12} md={6} key={p.provider}>
                <Card size="small">
                  <Statistic
                    title={
                      <Tag color={getProviderColor(p.provider)}>{p.provider.toUpperCase()}</Tag>
                    }
                    value={p.count}
                    suffix={`次 / ${formatLedgerUnits(p.cost)} Token`}
                    valueStyle={{ fontSize: isMobile ? 14 : 16 }}
                  />
                </Card>
              </Col>
            ))}
          </Row>
        </Card>
      )}

      {mainView === 'charts' && (
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col xs={24} lg={14}>
            <Card title="扣减趋势（按日）">
              {dailySeries.length === 0 ? (
                <Paragraph type="secondary">当前筛选条件下暂无数据</Paragraph>
              ) : (
                <div style={{ width: '100%', height: 280 }}>
                  <ResponsiveContainer>
                    <AreaChart data={dailySeries} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="date" tick={{ fontSize: 11 }} />
                      <YAxis tick={{ fontSize: 11 }} />
                      <RechartsTooltip />
                      <Area
                        type="monotone"
                        dataKey="cost"
                        name="扣减(Token)"
                        stroke="#1677ff"
                        fill="#1677ff33"
                      />
                    </AreaChart>
                  </ResponsiveContainer>
                </div>
              )}
            </Card>
          </Col>
          <Col xs={24} lg={10}>
            <Card title="按 Provider 扣减">
              {providerBarData.length === 0 ? (
                <Paragraph type="secondary">暂无 Provider 汇总</Paragraph>
              ) : (
                <div style={{ width: '100%', height: 280 }}>
                  <ResponsiveContainer>
                    <BarChart data={providerBarData} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="name" tick={{ fontSize: 11 }} />
                      <YAxis tick={{ fontSize: 11 }} />
                      <RechartsTooltip />
                      <Bar dataKey="cost" name="扣减(Token)" fill="#722ed1" radius={[4, 4, 0, 0]} />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              )}
            </Card>
          </Col>
        </Row>
      )}

      {mainView === 'records' && (
        <Card
          className={styles.tableCard}
          title="消费明细"
          extra={
            <Button
              icon={<DownloadOutlined />}
              onClick={exportData}
              size={isMobile ? 'small' : 'middle'}
            >
              {!isMobile && '导出'}
            </Button>
          }
        >
          <Spin spinning={loading}>
            {recordsLayout === 'table' ? (
              <Table
                columns={columns}
                dataSource={records}
                rowKey="id"
                scroll={{ x: 600 }}
                size={isMobile ? 'small' : 'middle'}
                pagination={{
                  pageSize: 20,
                  size: isMobile ? 'small' : 'default',
                  showSizeChanger: !isMobile,
                  showQuickJumper: !isMobile,
                  showTotal: isMobile ? undefined : (total) => `共 ${total} 条记录`,
                }}
              />
            ) : (
              <Row gutter={[12, 12]}>
                {records.map((r) => (
                  <Col xs={24} sm={12} lg={8} key={r.id}>
                    <Card
                      size="small"
                      hoverable
                      onClick={() => showDetail(r)}
                      styles={{ body: { cursor: 'pointer' } }}
                      title={
                        <Tag color={getProviderColor(r.provider)}>{r.provider.toUpperCase()}</Tag>
                      }
                      extra={
                        <Button
                          type="link"
                          size="small"
                          onClick={(e) => {
                            e.stopPropagation();
                            showDetail(r);
                          }}
                        >
                          详情
                        </Button>
                      }
                    >
                      <Space direction="vertical" size={4} style={{ width: '100%' }}>
                        <span style={{ fontSize: 12, color: '#666' }}>{r.model}</span>
                        <span>
                          扣减：<Text type="danger">{formatLedgerUnits(r.cost)}</Text>
                        </span>
                        <span style={{ fontSize: 12 }}>
                          {dayjs(r.created_at).format('MM-DD HH:mm')} · {r.status_code}
                        </span>
                      </Space>
                    </Card>
                  </Col>
                ))}
              </Row>
            )}
          </Spin>
        </Card>
      )}

      <Modal
        title="请求详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={isMobile ? '95%' : 600}
      >
        {selectedRecord && (
          <Descriptions column={{ xs: 1, sm: 2 }} bordered size="small">
            <Descriptions.Item label="请求ID" span={2}>
              <code style={{ fontSize: 12 }}>{selectedRecord.request_id}</code>
            </Descriptions.Item>
            <Descriptions.Item label="Provider">
              <Tag color={getProviderColor(selectedRecord.provider)}>
                {selectedRecord.provider.toUpperCase()}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Model">{selectedRecord.model}</Descriptions.Item>
            <Descriptions.Item label="Method">
              <Tag>{selectedRecord.method}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Path">{selectedRecord.path}</Descriptions.Item>
            <Descriptions.Item label="输入Tokens">
              {selectedRecord.input_tokens.toLocaleString()}
            </Descriptions.Item>
            <Descriptions.Item label="输出Tokens">
              {selectedRecord.output_tokens.toLocaleString()}
            </Descriptions.Item>
            <Descriptions.Item label="扣减（Token）">
              <span style={{ color: '#f5222d', fontWeight: 'bold' }}>
                {formatLedgerUnits(selectedRecord.cost)}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label="延迟">{selectedRecord.latency_ms} ms</Descriptions.Item>
            <Descriptions.Item label="状态码">
              <Tag color={selectedRecord.status_code < 300 ? 'success' : 'error'}>
                {selectedRecord.status_code}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="请求时间" span={2}>
              {dayjs(selectedRecord.created_at).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  );
};

export default Consumption;
