import React, { useEffect, useState } from 'react'
import {
  Card,
  Button,
  Tag,
  Progress,
  Space,
  Spin,
  Empty,
  Result,
  Avatar,
  List,
  Typography,
  message,
  Modal,
} from 'antd'
import {
  ShareAltOutlined,
  CopyOutlined,
  CloseOutlined,
  UserOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons'
import { useParams, useNavigate } from 'react-router-dom'
import { useGroupStore } from '@/stores/groupStore'

const { Title, Text } = Typography

const statusConfig: Record<string, { color: string; label: string; icon: React.ReactNode }> = {
  active: { color: 'processing', label: '进行中', icon: <ClockCircleOutlined /> },
  completed: { color: 'success', label: '已成团', icon: <CheckCircleOutlined /> },
  failed: { color: 'error', label: '已失败', icon: <ExclamationCircleOutlined /> },
}

export const GroupProgressPage: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { currentGroup, isLoading, error, getGroupProgress, cancelGroup } = useGroupStore()
  const [cancelModalVisible, setCancelModalVisible] = useState(false)
  const [countdown, setCountdown] = useState('')

  useEffect(() => {
    if (id) {
      getGroupProgress(parseInt(id))
    }
  }, [id, getGroupProgress])

  useEffect(() => {
    if (!currentGroup || currentGroup.status !== 'active') return

    const updateCountdown = () => {
      const deadline = new Date(currentGroup.deadline).getTime()
      const now = Date.now()
      const diff = deadline - now

      if (diff <= 0) {
        setCountdown('已过期')
        return
      }

      const hours = Math.floor(diff / (1000 * 60 * 60))
      const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
      const seconds = Math.floor((diff % (1000 * 60)) / 1000)

      setCountdown(`${hours}小时${minutes}分钟${seconds}秒`)
    }

    updateCountdown()
    const timer = setInterval(updateCountdown, 1000)

    return () => clearInterval(timer)
  }, [currentGroup])

  const handleShare = () => {
    if (!currentGroup) return
    const shareUrl = `${window.location.origin}/groups/${currentGroup.id}/join`
    navigator.clipboard.writeText(shareUrl)
    message.success('分享链接已复制到剪贴板')
  }

  const handleCopyLink = () => {
    handleShare()
  }

  const handleCancel = async () => {
    if (!currentGroup) return
    try {
      await cancelGroup(currentGroup.id)
      message.success('拼团已取消')
      setCancelModalVisible(false)
      navigate('/groups')
    } catch {
      message.error('取消失败，请重试')
    }
  }

  if (isLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}>
        <Spin size="large" tip="加载中..." />
      </div>
    )
  }

  if (error) {
    return <Empty description={`错误: ${error}`} />
  }

  if (!currentGroup) {
    return <Empty description="拼团不存在" />
  }

  const statusInfo = statusConfig[currentGroup.status] || statusConfig.active
  const progress = (currentGroup.current_count / currentGroup.target_count) * 100
  const remainingCount = currentGroup.target_count - currentGroup.current_count

  if (currentGroup.status === 'completed') {
    return (
      <div style={{ padding: '20px', maxWidth: 600, margin: '0 auto' }}>
        <Result
          status="success"
          icon={<CheckCircleOutlined />}
          title="拼团成功！"
          subTitle={`订单号: #${currentGroup.id}`}
          extra={[
            <Button type="primary" key="pay" onClick={() => navigate('/orders')}>
              查看订单
            </Button>,
            <Button key="home" onClick={() => navigate('/')}>
              返回首页
            </Button>,
          ]}
        />
      </div>
    )
  }

  if (currentGroup.status === 'failed') {
    return (
      <div style={{ padding: '20px', maxWidth: 600, margin: '0 auto' }}>
        <Result
          status="error"
          icon={<ExclamationCircleOutlined />}
          title="拼团失败"
          subTitle="拼团时间已过，未能成功成团"
          extra={[
            <Button type="primary" key="retry" onClick={() => navigate('/products')}>
              重新购买
            </Button>,
            <Button key="home" onClick={() => navigate('/')}>
              返回首页
            </Button>,
          ]}
        />
      </div>
    )
  }

  return (
    <div style={{ padding: '20px', maxWidth: 600, margin: '0 auto' }}>
      <Card>
        <div style={{ marginBottom: 20 }}>
          <Space align="center">
            <Title level={3}>拼团进度</Title>
            <Tag color={statusInfo.color} icon={statusInfo.icon}>
              {statusInfo.label}
            </Tag>
          </Space>
        </div>

        <div style={{ marginBottom: 20 }}>
          <Text type="secondary">订单号：#{currentGroup.id}</Text>
        </div>

        <Card style={{ marginBottom: 20, background: '#fafafa' }}>
          <div style={{ textAlign: 'center', marginBottom: 20 }}>
            <Title level={2}>
              {currentGroup.current_count}/{currentGroup.target_count}
            </Title>
            <Text type="secondary">人成团</Text>
          </div>

          <Progress
            percent={progress}
            status={currentGroup.status === 'active' ? 'active' : 'success'}
            format={() => `还需${remainingCount}人`}
          />

          <div style={{ marginTop: 20, textAlign: 'center' }}>
            <ClockCircleOutlined style={{ marginRight: 8 }} />
            <Text>剩余时间：{countdown}</Text>
          </div>
        </Card>

        <Card title="成员列表" style={{ marginBottom: 20 }}>
          <List
            dataSource={Array.from({ length: currentGroup.target_count }, (_, i) => ({
              id: i,
              joined: i < currentGroup.current_count,
            }))}
            renderItem={(item: { id: number; joined: boolean }) => (
              <List.Item>
                <List.Item.Meta
                  avatar={
                    <Avatar
                      icon={item.joined ? <UserOutlined /> : <CloseOutlined />}
                      style={{ 
                        backgroundColor: item.joined ? '#1890ff' : '#f0f0f0',
                        color: item.joined ? '#fff' : '#999',
                      }}
                    />
                  }
                  title={item.joined ? `成员 ${(item.id as number) + 1}` : '等待加入...'}
                  description={item.joined ? '已加入' : '等待中'}
                />
              </List.Item>
            )}
          />
        </Card>

        <Space direction="vertical" style={{ width: '100%' }}>
          <Button
            type="primary"
            icon={<ShareAltOutlined />}
            onClick={handleShare}
            block
            size="large"
          >
            分享邀请
          </Button>
          <Button
            icon={<CopyOutlined />}
            onClick={handleCopyLink}
            block
          >
            复制链接
          </Button>
          <Button
            danger
            icon={<CloseOutlined />}
            onClick={() => setCancelModalVisible(true)}
            block
          >
            取消拼团
          </Button>
        </Space>

        <div style={{ marginTop: 16, textAlign: 'center' }}>
          <Button type="link" onClick={() => navigate('/groups')}>
            返回拼团列表
          </Button>
        </div>
      </Card>

      <Modal
        title="确认取消拼团"
        open={cancelModalVisible}
        onOk={handleCancel}
        onCancel={() => setCancelModalVisible(false)}
        okText="确认取消"
        cancelText="返回"
        okButtonProps={{ danger: true }}
      >
        <p>确定要取消这个拼团吗？取消后需要重新发起拼团。</p>
      </Modal>
    </div>
  )
}

export default GroupProgressPage
