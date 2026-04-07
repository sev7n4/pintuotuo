import React, { useEffect, useState } from 'react';
import { Card, Row, Col, Select, DatePicker, Spin, Empty, Table } from 'antd';
import dayjs from 'dayjs';
import { billingService, type BillingTrend } from '@/services/billing';

const { RangePicker } = DatePicker;

interface BillingTrendChartProps {
  merchantId?: number;
  userId?: number;
}

const BillingTrendChart: React.FC<BillingTrendChartProps> = ({ merchantId, userId }) => {
  const [trends, setTrends] = useState<BillingTrend[]>([]);
  const [loading, setLoading] = useState(false);
  const [granularity, setGranularity] = useState<'day' | 'week' | 'month'>('day');
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);

  useEffect(() => {
    fetchTrends();
  }, [merchantId, userId, granularity, dateRange]);

  const fetchTrends = async () => {
    setLoading(true);
    try {
      const params: Record<string, unknown> = { granularity };
      if (merchantId) params.merchant_id = merchantId;
      if (userId) params.user_id = userId;
      if (dateRange) {
        params.start_date = dateRange[0].format('YYYY-MM-DD');
        params.end_date = dateRange[1].format('YYYY-MM-DD');
      }
      const response = await billingService.getBillingTrends(params);
      setTrends(response.data.trends || []);
    } catch (error) {
      console.error('获取趋势数据失败', error);
    } finally {
      setLoading(false);
    }
  };

  const columns = [
    {
      title: '日期',
      dataIndex: 'date',
      key: 'date',
    },
    {
      title: '消费(USD)',
      dataIndex: 'total_cost',
      key: 'total_cost',
      render: (v: number) => `$${v.toFixed(4)}`,
    },
    {
      title: 'Tokens',
      dataIndex: 'total_tokens',
      key: 'total_tokens',
      render: (v: number) => v.toLocaleString(),
    },
    {
      title: '请求数',
      dataIndex: 'total_requests',
      key: 'total_requests',
      render: (v: number) => v.toLocaleString(),
    },
    {
      title: '平均延迟(ms)',
      dataIndex: 'avg_latency',
      key: 'avg_latency',
      render: (v: number) => v.toFixed(2),
    },
  ];

  return (
    <Card 
      title="消费趋势" 
      extra={
        <Row gutter={16}>
          <Col>
            <Select
              value={granularity}
              onChange={setGranularity}
              style={{ width: 100, marginRight: 8 }}
            >
              <Select.Option value="day">按日</Select.Option>
              <Select.Option value="week">按周</Select.Option>
              <Select.Option value="month">按月</Select.Option>
            </Select>
          </Col>
          <Col>
            <RangePicker
              value={dateRange}
              onChange={(dates) => setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs] | null)}
            />
          </Col>
        </Row>
      }
    >
      <Spin spinning={loading}>
        {trends.length === 0 ? (
          <Empty description="暂无数据" />
        ) : (
          <Table
            dataSource={trends}
            columns={columns}
            rowKey="date"
            pagination={false}
            size="small"
          />
        )}
      </Spin>
    </Card>
  );
};

export default BillingTrendChart;
