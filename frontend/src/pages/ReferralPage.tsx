import { useEffect, useState } from 'react'
import { Card, Row, Col, Typography, Button, Statistic, Table, Tabs, message, Input, Space, Spin } from 'antd'
import { CopyOutlined, ShareAltOutlined, GiftOutlined, TeamOutlined, DollarOutlined } from '@ant-design/icons'
import { useReferralStore } from '@/stores/referralStore'
import { Referral, ReferralReward } from '@/types'
import type { ColumnsType } from 'antd/es/table'
import styles from './ReferralPage.module.css'

const { Title, Text, Paragraph } = Typography
const { TabPane } = Tabs

const ReferralPage = () => {
  const {
    referralCode,
    stats,
    referrals,
    rewards,
    isLoading,
    fetchReferralCode,
    fetchStats,
    fetchReferrals,
    fetchRewards,
  } = useReferralStore()

  const [bindCode, setBindCode] = useState('')

  useEffect(() => {
    fetchReferralCode()
    fetchStats()
    fetchReferrals()
    fetchRewards()
  }, [fetchReferralCode, fetchStats, fetchReferrals, fetchRewards])

  const handleCopyCode = () => {
    navigator.clipboard.writeText(referralCode)
    message.success('邀请码已复制到剪贴板')
  }

  const handleShare = () => {
    const shareUrl = `${window.location.origin}/register?code=${referralCode}`
    navigator.clipboard.writeText(shareUrl)
    message.success('分享链接已复制到剪贴板')
  }

  const referralColumns: ColumnsType<Referral> = [
    {
      title: '用户',
      dataIndex: 'referee_name',
      key: 'referee_name',
    },
    {
      title: '使用邀请码',
      dataIndex: 'code_used',
      key: 'code_used',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <span className={status === 'active' ? styles.statusActive : styles.statusCancelled}>
          {status === 'active' ? '有效' : '已取消'}
        </span>
      ),
    },
    {
      title: '注册时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleDateString('zh-CN'),
    },
  ]

  const rewardColumns: ColumnsType<ReferralReward> = [
    {
      title: '来源用户',
      dataIndex: 'referee_name',
      key: 'referee_name',
    },
    {
      title: '返利金额',
      dataIndex: 'amount',
      key: 'amount',
      render: (amount: number) => `¥${amount.toFixed(2)}`,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const statusMap: Record<string, { text: string; className: string }> = {
          pending: { text: '待发放', className: styles.statusPending },
          paid: { text: '已发放', className: styles.statusPaid },
          cancelled: { text: '已取消', className: styles.statusCancelled },
        }
        const { text, className } = statusMap[status] || { text: status, className: '' }
        return <span className={className}>{text}</span>
      },
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleDateString('zh-CN'),
    },
  ]

  return (
    <div className={styles.container}>
      <Title level={2} className={styles.pageTitle}>
        <GiftOutlined /> 邀请好友
      </Title>

      <Row gutter={[24, 24]}>
        <Col xs={24} md={12}>
          <Card className={styles.codeCard}>
            <div className={styles.codeSection}>
              <Text type="secondary">我的邀请码</Text>
              <Title level={1} className={styles.codeText}>
                {isLoading ? <Spin /> : referralCode || '------'}
              </Title>
              <Space>
                <Button type="primary" icon={<CopyOutlined />} onClick={handleCopyCode}>
                  复制邀请码
                </Button>
                <Button icon={<ShareAltOutlined />} onClick={handleShare}>
                  分享链接
                </Button>
              </Space>
            </div>
            <div className={styles.tips}>
              <Paragraph type="secondary">
                分享您的邀请码给好友，好友注册成功后，您将获得其首单消费金额 5% 的返利奖励！
              </Paragraph>
            </div>
          </Card>
        </Col>

        <Col xs={24} md={12}>
          <Card className={styles.statsCard}>
            <Row gutter={16}>
              <Col span={12}>
                <Statistic
                  title="邀请人数"
                  value={stats?.total_referrals || 0}
                  prefix={<TeamOutlined />}
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="累计返利"
                  value={stats?.total_rewards || 0}
                  precision={2}
                  prefix={<DollarOutlined />}
                  suffix="元"
                />
              </Col>
            </Row>
            <Row gutter={16} style={{ marginTop: 24 }}>
              <Col span={12}>
                <Statistic
                  title="待发放返利"
                  value={stats?.pending_rewards || 0}
                  precision={2}
                  valueStyle={{ color: '#faad14' }}
                  suffix="元"
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="已发放返利"
                  value={stats?.paid_rewards || 0}
                  precision={2}
                  valueStyle={{ color: '#52c41a' }}
                  suffix="元"
                />
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>

      <Card style={{ marginTop: 24 }}>
        <Tabs defaultActiveKey="referrals">
          <TabPane tab="邀请记录" key="referrals">
            <Table
              columns={referralColumns}
              dataSource={referrals}
              rowKey="id"
              pagination={{ pageSize: 10 }}
              loading={isLoading}
              locale={{ emptyText: '暂无邀请记录' }}
            />
          </TabPane>
          <TabPane tab="返利明细" key="rewards">
            <Table
              columns={rewardColumns}
              dataSource={rewards}
              rowKey="id"
              pagination={{ pageSize: 10 }}
              loading={isLoading}
              locale={{ emptyText: '暂无返利记录' }}
            />
          </TabPane>
        </Tabs>
      </Card>

      <Card style={{ marginTop: 24 }} title="绑定邀请码">
        <Space.Compact style={{ width: '100%', maxWidth: 400 }}>
          <Input
            placeholder="输入好友的邀请码"
            value={bindCode}
            onChange={(e) => setBindCode(e.target.value.toUpperCase())}
            maxLength={8}
          />
          <Button type="primary" disabled={!bindCode || bindCode.length !== 8}>
            绑定
          </Button>
        </Space.Compact>
        <Paragraph type="secondary" style={{ marginTop: 12 }}>
          如果您有好友的邀请码，可以在此绑定，绑定后不可更改。
        </Paragraph>
      </Card>
    </div>
  )
}

export default ReferralPage
