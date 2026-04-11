import { useEffect, useState, useCallback } from 'react';
import { Card, Select, Spin, Empty, Table, Space } from 'antd';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import dayjs from 'dayjs';
import { reconciliationService, type GMVTrendPoint } from '@/services/reconciliation';

export type GMVGranularity = 'day' | 'week' | 'month';

interface GMVTrendChartProps {
  /** 与上方 GMV 汇总共用；为 null 时曲线请求使用服务端默认近 90 日 */
  dateRange: [dayjs.Dayjs, dayjs.Dayjs] | null;
}

const GMVTrendChart: React.FC<GMVTrendChartProps> = ({ dateRange }) => {
  const [granularity, setGranularity] = useState<GMVGranularity>('day');
  const [points, setPoints] = useState<GMVTrendPoint[]>([]);
  const [loading, setLoading] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const params: {
        granularity: GMVGranularity;
        start_date?: string;
        end_date?: string;
      } = { granularity };
      if (dateRange) {
        params.start_date = dateRange[0].format('YYYY-MM-DD');
        params.end_date = dateRange[1].format('YYYY-MM-DD');
      }
      const { data } = await reconciliationService.getGMVTrends(params);
      setPoints(data.trends || []);
    } catch {
      setPoints([]);
    } finally {
      setLoading(false);
    }
  }, [granularity, dateRange]);

  useEffect(() => {
    void load();
  }, [load]);

  const chartData = points.map((p) => ({
    ...p,
    label: p.period,
  }));

  const columns = [
    { title: '周期', dataIndex: 'period', key: 'period' },
    {
      title: 'GMV（¥）',
      dataIndex: 'gmv_cny',
      key: 'gmv_cny',
      render: (v: number) =>
        v.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }),
    },
    { title: '订单数', dataIndex: 'order_count', key: 'order_count' },
  ];

  return (
    <Card
      title="GMV 曲线"
      extra={
        <Space>
          <Select<GMVGranularity>
            value={granularity}
            onChange={setGranularity}
            style={{ width: 100 }}
            options={[
              { value: 'day', label: '按日' },
              { value: 'week', label: '按周' },
              { value: 'month', label: '按月' },
            ]}
          />
        </Space>
      }
    >
      <Spin spinning={loading}>
        {chartData.length === 0 ? (
          <Empty description="暂无趋势数据" />
        ) : (
          <>
            <div style={{ width: '100%', height: 320 }}>
              <ResponsiveContainer>
                <LineChart data={chartData} margin={{ top: 8, right: 12, left: 8, bottom: 8 }}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="label" tick={{ fontSize: 11 }} interval="preserveStartEnd" />
                  <YAxis tickFormatter={(v) => `¥${v}`} width={72} tick={{ fontSize: 11 }} />
                  <Tooltip
                    formatter={(value) => {
                      const n = Number(value ?? 0);
                      return [
                        `¥${n.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`,
                        'GMV',
                      ];
                    }}
                    labelFormatter={(l) => `周期: ${l}`}
                  />
                  <Line
                    type="monotone"
                    dataKey="gmv_cny"
                    stroke="#1677ff"
                    strokeWidth={2}
                    dot={{ r: 3 }}
                    name="GMV"
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
            <Table
              style={{ marginTop: 16 }}
              size="small"
              rowKey="period"
              pagination={false}
              dataSource={points}
              columns={columns}
            />
          </>
        )}
      </Spin>
    </Card>
  );
};

export default GMVTrendChart;
