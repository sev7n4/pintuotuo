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
  ScatterChart,
  Scatter,
  ZAxis,
  Legend,
  Cell,
} from 'recharts';
import styles from './Consumption.module.css';
import { formatLedgerUnits } from '@/utils/ledgerDisplay';

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
  created_at: string;
}

interface ConsumptionStats {
  total_requests: number;
  total_token_deduction: number;
  avg_latency_ms: number;
}

interface ProviderStats {
  provider: string;
  count: number;
  tokens: number;
}

/** 与 GET /consumption/stats 中 model_comparison 一致 */
interface ModelComparisonRow {
  provider: string;
  model: string;
  request_count: number;
  total_token_deduction: number;
  avg_token_deduction: number;
  latency_p50_ms: number;
  latency_p95_ms: number;
  success_rate: number;
}

type ModelComparePoint = ModelComparisonRow & {
  rpm: number;
  tpm: number;
  low_sample: boolean;
};

const MIN_MODEL_SAMPLES = 5;

function getProviderColorStatic(provider: string): string {
  const colors: Record<string, string> = {
    openai: 'green',
    anthropic: 'purple',
    google: 'blue',
    azure: 'cyan',
  };
  return colors[provider.toLowerCase()] || 'default';
}

function providerChartFill(provider: string): string {
  const key = (provider || '').toLowerCase();
  const map: Record<string, string> = {
    openai: '#52c41a',
    anthropic: '#722ed1',
    google: '#1677ff',
    azure: '#13c2c2',
  };
  return map[key] || '#8c8c8c';
}

function ModelCompareTooltipView({
  active,
  payload,
}: {
  active?: boolean;
  payload?: ReadonlyArray<{ payload: ModelComparePoint }>;
}) {
  if (!active || !payload?.length) return null;
  const p = payload[0].payload;
  const srPct = Number.isFinite(p.success_rate) ? (p.success_rate * 100).toFixed(1) : '—';
  return (
    <div className={styles.modelCompareTooltip}>
      <div style={{ fontWeight: 600, marginBottom: 6 }}>
        {p.model}
        <Tag color={getProviderColorStatic(p.provider)} style={{ marginLeft: 8 }}>
          {p.provider.toUpperCase()}
        </Tag>
      </div>
      <div>
        延迟 p50：{Math.round(p.latency_p50_ms)} ms · p95：{Math.round(p.latency_p95_ms)} ms
      </div>
      <div>单次平均扣减：{formatLedgerUnits(Math.round(p.avg_token_deduction))} Tokens</div>
      <div>
        请求数：{p.request_count.toLocaleString()} · 成功率：{srPct}%
      </div>
      <div>
        RPM：{p.rpm.toFixed(2)} / min · TPM：{Math.round(p.tpm).toLocaleString()} / min
      </div>
      {p.low_sample ? (
        <Paragraph type="warning" style={{ marginTop: 8, marginBottom: 0, fontSize: 11 }}>
          样本较少（少于 {MIN_MODEL_SAMPLES} 次请求），对比仅供参考
        </Paragraph>
      ) : null}
    </div>
  );
}

/** C 端扣减口径：输入 tokens + 输出 tokens（与统计合计一致，不展示内部 cost） */
function rowTokenDeduct(r: Pick<ConsumptionRecord, 'input_tokens' | 'output_tokens'>): number {
  return (Number(r.input_tokens) || 0) + (Number(r.output_tokens) || 0);
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
  const [modelComparison, setModelComparison] = useState<ModelComparisonRow[]>([]);
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
        const mc = (data.model_comparison as ModelComparisonRow[] | undefined) || [];
        setModelComparison(
          mc.map((row) => ({
            provider: String(row.provider ?? ''),
            model: String(row.model ?? ''),
            request_count: Number(row.request_count) || 0,
            total_token_deduction: Number(row.total_token_deduction) || 0,
            avg_token_deduction: Number(row.avg_token_deduction) || 0,
            latency_p50_ms: Number(row.latency_p50_ms) || 0,
            latency_p95_ms: Number(row.latency_p95_ms) || 0,
            success_rate: Number(row.success_rate) || 0,
          }))
        );
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
        '扣减(输入+输出Tokens)',
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
          rowTokenDeduct(r),
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
        title: '扣减（输入+输出）',
        key: 'token_deduct',
        width: 120,
        align: 'right',
        render: (_: unknown, record: ConsumptionRecord) => (
          <span style={{ color: '#f5222d' }}>{formatLedgerUnits(rowTokenDeduct(record))}</span>
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
      m.set(d, (m.get(d) || 0) + rowTokenDeduct(r));
    }
    return [...m.entries()]
      .map(([date, tokens]) => ({ date, tokens: Math.round(tokens) }))
      .sort((a, b) => a.date.localeCompare(b.date));
  }, [records]);

  const providerBarData = useMemo(
    () => providerStats.map((p) => ({ name: p.provider, tokens: p.tokens, count: p.count })),
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
  const tpm = stats ? stats.total_token_deduction / rangeMinutes : 0;

  const modelComparePoints = useMemo((): ModelComparePoint[] => {
    return modelComparison.map((row) => ({
      ...row,
      rpm: row.request_count / rangeMinutes,
      tpm: row.total_token_deduction / rangeMinutes,
      low_sample: row.request_count < MIN_MODEL_SAMPLES,
    }));
  }, [modelComparison, rangeMinutes]);

  const compareProvidersOrdered = useMemo(() => {
    const uniq = new Set(modelComparePoints.map((x) => x.provider).filter(Boolean));
    return [...uniq].sort((a, b) => a.localeCompare(b));
  }, [modelComparePoints]);

  const maxCompareRequests = useMemo(
    () => modelComparePoints.reduce((m, x) => Math.max(m, x.request_count), 0),
    [modelComparePoints]
  );

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
          上方筛选项对列表与统计、图表共用。图表视图中含按日扣减、按 Provider
          汇总，以及「模型对比」气泡图（延迟与单次扣减、用量节奏），便于在相同筛选条件下评估不同厂家与模型。
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
              title="合计扣减（输入+输出）"
              value={stats?.total_token_deduction || 0}
              suffix="Tokens"
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
          合计扣减为当前筛选条件下，各次请求「输入 tokens + 输出
          tokens」之和（与明细列表一致）；不展示内部计费 cost。RPM、TPM 为区间均值。
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
                    suffix={`次 / 扣减 ${formatLedgerUnits(p.tokens)} Tokens`}
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
                        dataKey="tokens"
                        name="扣减（输入+输出）"
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
                    <BarChart
                      data={providerBarData}
                      margin={{ top: 8, right: 8, left: 0, bottom: 0 }}
                    >
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="name" tick={{ fontSize: 11 }} />
                      <YAxis tick={{ fontSize: 11 }} />
                      <RechartsTooltip />
                      <Bar
                        dataKey="tokens"
                        name="扣减（输入+输出）"
                        fill="#722ed1"
                        radius={[4, 4, 0, 0]}
                      />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              )}
            </Card>
          </Col>
        </Row>
      )}

      {mainView === 'charts' && (
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col xs={24}>
            <Card
              title="模型对比（选购参考）"
              extra={
                <Text type="secondary" style={{ fontSize: 12 }}>
                  横轴越快越好 · 纵轴单次扣减越少越省 · 气泡越大用量越多
                </Text>
              }
            >
              {modelComparePoints.length === 0 ? (
                <Paragraph type="secondary" style={{ marginBottom: 0 }}>
                  当前筛选条件下暂无带模型名的请求，无法绘制对比图。请调整日期或去掉过窄的模型筛选。
                </Paragraph>
              ) : (
                <>
                  <div style={{ width: '100%', height: isMobile ? 340 : 420 }}>
                    <ResponsiveContainer>
                      <ScatterChart margin={{ top: 12, right: 12, left: 8, bottom: 12 }}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis
                          type="number"
                          dataKey="latency_p50_ms"
                          name="延迟 p50"
                          unit=" ms"
                          tick={{ fontSize: 11 }}
                          domain={[0, 'auto']}
                          label={{
                            value: '延迟 p50（ms）· 越小越快',
                            position: 'insideBottom',
                            offset: -4,
                            style: { fontSize: 11, fill: '#8c8c8c' },
                          }}
                        />
                        <YAxis
                          type="number"
                          dataKey="avg_token_deduction"
                          name="单次平均扣减"
                          tick={{ fontSize: 11 }}
                          domain={[0, 'auto']}
                          width={56}
                          label={{
                            value: '单次平均扣减（Tokens）',
                            angle: -90,
                            position: 'insideLeft',
                            style: { fontSize: 11, fill: '#8c8c8c' },
                          }}
                        />
                        <ZAxis
                          type="number"
                          dataKey="request_count"
                          range={[56, 320]}
                          domain={[0, Math.max(maxCompareRequests, 1)]}
                        />
                        <RechartsTooltip
                          cursor={{ strokeDasharray: '3 3' }}
                          content={(props) => (
                            <ModelCompareTooltipView
                              active={props.active}
                              payload={
                                props.payload as ReadonlyArray<{ payload: ModelComparePoint }>
                              }
                            />
                          )}
                        />
                        <Legend wrapperStyle={{ fontSize: 12 }} />
                        {compareProvidersOrdered.map((prov) => {
                          const series = modelComparePoints.filter((d) => d.provider === prov);
                          return (
                            <Scatter
                              key={prov}
                              name={prov.toUpperCase()}
                              data={series}
                              fill={providerChartFill(prov)}
                            >
                              {series.map((entry) => (
                                <Cell
                                  key={`${entry.provider}:${entry.model}`}
                                  fill={providerChartFill(prov)}
                                  fillOpacity={entry.low_sample ? 0.42 : 0.9}
                                />
                              ))}
                            </Scatter>
                          );
                        })}
                      </ScatterChart>
                    </ResponsiveContainer>
                  </div>
                  <Paragraph
                    type="secondary"
                    style={{ marginTop: 12, marginBottom: 0, fontSize: 12 }}
                  >
                    每个气泡为「Provider + 模型」在当前筛选时间范围内的汇总；扣减为输入+输出
                    Tokens，与明细一致。成功率按 2xx 请求占比。RPM、TPM
                    按该时间区间长度折算（与上方统计卡片同一口径），用于观察各模型的调用与消耗节奏。
                  </Paragraph>
                </>
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
                          扣减（输入+输出）：
                          <Text type="danger">{formatLedgerUnits(rowTokenDeduct(r))}</Text>
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
            <Descriptions.Item label="扣减（输入+输出）" span={2}>
              <span style={{ color: '#f5222d', fontWeight: 'bold' }}>
                {formatLedgerUnits(rowTokenDeduct(selectedRecord))} Tokens（=
                {selectedRecord.input_tokens.toLocaleString()} +{' '}
                {selectedRecord.output_tokens.toLocaleString()}）
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
