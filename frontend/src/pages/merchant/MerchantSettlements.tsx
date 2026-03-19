import { useEffect, useState } from 'react'
import { Card, Table, Tag, Button, Modal, Descriptions, message, Statistic, Row, Col } from 'antd'
import { DollarOutlined, CheckCircleOutlined, ClockCircleOutlined, SyncOutlined } from '@ant-design/icons'
import { useMerchantStore } from '@/stores/merchantStore'
import { MerchantSettlement } from '@/types'
import styles from './MerchantSettlements.module.css'

const MerchantSettlements = () => {
  const { settlements, fetchSettlements, requestSettlement, isLoading } = useMerchantStore()
  const [detailVisible, setDetailVisible] = useState(false)
  const [selectedSettlement, setSelectedSettlement] = useState<MerchantSettlement | null>(null)

  useEffect(() => {
    fetchSettlements()
  }, [fetchSettlements])

  const handleRequestSettlement = async () => {
    const success = await requestSettlement()
    if (success) {
      message.success('结算申请已提交')
      fetchSettlements()
    }
  }

  const handleViewDetail = (record: MerchantSettlement) => {
    setSelectedSettlement(record)
    setDetailVisible(true)
  }

  const statusMap: Record<string, { color: string; text: string; icon: React.ReactNode }> = {
    pending: { color: 'default', text: '待处理', icon: <ClockCircleOutlined /> },
    processing: { color: 'processing', text: '处理中', icon: <SyncOutlined spin /> },
    completed: { color: 'success', text: '已完成', icon: <CheckCircleOutlined /> },
  }

  const columns = [
    {
      title: '结算ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '结算周期',
      key: 'period',
      render: (_: unknown, record: MerchantSettlement) => (
        <span>
          {new Date(record.period_start).toLocaleDateString('zh-CN')} - {new Date(record.period_end).toLocaleDateString('zh-CN')}
        </span>
      ),
    },
    {
      title: '销售总额',
      dataIndex: 'total_sales',
      key: 'total_sales',
      render: (amount: number) => `¥${amount.toFixed(2)}`,
    },
    {
      title: '平台费用',
      dataIndex: 'platform_fee',
      key: 'platform_fee',
      render: (fee: number) => `¥${fee.toFixed(2)}`,
    },
    {
      title: '结算金额',
      dataIndex: 'settlement_amount',
      key: 'settlement_amount',
      render: (amount: number) => (
        <span className={styles.amount}>¥{amount.toFixed(2)}</span>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const { color, text, icon } = statusMap[status] || { color: 'default', text: status, icon: null }
        return (
          <Tag color={color} icon={icon}>
            {text}
          </Tag>
        )
      },
    },
    {
      title: '结算时间',
      dataIndex: 'settled_at',
      key: 'settled_at',
      render: (date: string) => date ? new Date(date).toLocaleString('zh-CN') : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_: unknown, record: MerchantSettlement) => (
        <Button type="link" size="small" onClick={() => handleViewDetail(record)}>
          查看详情
        </Button>
      ),
    },
  ]

  const totalSettlements = settlements.filter(s => s.status === 'completed').reduce((sum, s) => sum + s.settlement_amount, 0)
  const pendingSettlements = settlements.filter(s => s.status === 'pending' || s.status === 'processing').reduce((sum, s) => sum + s.settlement_amount, 0)

  return (
    <div className={styles.settlements}>
      <h2 className={styles.pageTitle}>结算管理</h2>

      <Row gutter={[16, 16]} className={styles.statsRow}>
        <Col xs={24} sm={12}>
          <Card>
            <Statistic
              title="累计结算"
              value={totalSettlements}
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
              value={pendingSettlements}
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
          <span>结算记录</span>
          <Button type="primary" onClick={handleRequestSettlement} loading={isLoading}>
            申请结算
          </Button>
        </div>
        <Table
          columns={columns}
          dataSource={settlements}
          rowKey="id"
          loading={isLoading}
          pagination={false}
        />
      </Card>

      <Modal
        title="结算详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={600}
      >
        {selectedSettlement && (
          <Descriptions column={2} bordered>
            <Descriptions.Item label="结算ID">{selectedSettlement.id}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={statusMap[selectedSettlement.status]?.color}>
                {statusMap[selectedSettlement.status]?.text}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="结算周期" span={2}>
              {new Date(selectedSettlement.period_start).toLocaleDateString('zh-CN')} - {new Date(selectedSettlement.period_end).toLocaleDateString('zh-CN')}
            </Descriptions.Item>
            <Descriptions.Item label="销售总额">
              ¥{selectedSettlement.total_sales.toFixed(2)}
            </Descriptions.Item>
            <Descriptions.Item label="平台费用（5%）">
              ¥{selectedSettlement.platform_fee.toFixed(2)}
            </Descriptions.Item>
            <Descriptions.Item label="结算金额" span={2}>
              <span className={styles.amount}>¥{selectedSettlement.settlement_amount.toFixed(2)}</span>
            </Descriptions.Item>
            <Descriptions.Item label="创建时间">
              {new Date(selectedSettlement.created_at).toLocaleString('zh-CN')}
            </Descriptions.Item>
            <Descriptions.Item label="结算时间">
              {selectedSettlement.settled_at ? new Date(selectedSettlement.settled_at).toLocaleString('zh-CN') : '-'}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  )
}

export default MerchantSettlements
