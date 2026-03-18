import React, { useEffect } from 'react'
import {
  List,
  Card,
  Button,
  Tag,
  Progress,
  Spin,
  Empty,
  message,
} from 'antd'
import { UserAddOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useGroupStore } from '@stores/groupStore'
import type { Group } from '@/types'

const statusMap: Record<string, { color: string; label: string }> = {
  active: { color: 'blue', label: '进行中' },
  completed: { color: 'green', label: '已成团' },
  failed: { color: 'red', label: '已失败' },
}

interface GroupWithStore extends Group {
  onJoin?: () => void
  isLoading?: boolean
}

export const GroupListPage: React.FC = () => {
  const navigate = useNavigate()
  const { isLoading, error, fetchGroups, joinGroup } = useGroupStore()
  const [groups, setGroups] = React.useState<GroupWithStore[]>([])
  const [joiningId, setJoiningId] = React.useState<number | null>(null)

  useEffect(() => {
    loadGroups()
  }, [])

  const loadGroups = async () => {
    const result = await fetchGroups()
    if (result) {
      setGroups(result)
    }
  }

  const handleJoinGroup = async (groupId: number) => {
    setJoiningId(groupId)
    try {
      await joinGroup(groupId)
      message.success('加入分组成功！')
      await loadGroups()
    } catch (err) {
      message.error('加入失败，请稍后重试')
    } finally {
      setJoiningId(null)
    }
  }

  if (error) {
    return <Empty description={`错误: ${error}`} />
  }

  if (groups.length === 0) {
    return (
      <div style={{ marginTop: 50, textAlign: 'center' }}>
        <Empty description="暂无分组" />
        <Button type="primary" style={{ marginTop: 16 }} onClick={() => navigate('/create-group')}>
          创建分组
        </Button>
      </div>
    )
  }

  return (
    <div style={{ padding: '20px' }}>
      <div style={{ marginBottom: '20px', display: 'flex', justifyContent: 'space-between' }}>
        <h1>拼团中心</h1>
        <Button type="primary" onClick={() => navigate('/create-group')}>
          创建分组
        </Button>
      </div>

      <Spin spinning={isLoading}>
        <List
          grid={{ gutter: 16, column: 2 }}
          dataSource={groups}
          renderItem={(group) => {
            const progress = (group.current_count / group.target_count) * 100
            const deadline = new Date(group.deadline)
            const isExpired = deadline < new Date()
            const s = statusMap[group.status] || { color: 'default', label: group.status }

            return (
              <List.Item key={group.id}>
                <Card
                  hoverable
                  actions={[
                    <Button
                      type={group.status === 'active' ? 'primary' : 'default'}
                      icon={<UserAddOutlined />}
                      onClick={() => handleJoinGroup(group.id)}
                      loading={joiningId === group.id}
                      disabled={group.status !== 'active'}
                    >
                      {group.status === 'active' ? '加入拼团' : '已' + s.label}
                    </Button>,
                  ]}
                >
                  <Card.Meta
                    title={`分组 #${group.id}`}
                    description={
                      <div>
                        <p>目标人数: {group.target_count}人</p>
                        <p>
                          当前人数: {group.current_count}人
                          <Tag color={s.color} style={{ marginLeft: 10 }}>
                            {s.label}
                          </Tag>
                        </p>
                        <Progress percent={progress} status={group.status === 'active' ? 'active' : 'success'} />
                        <p style={{ marginTop: 10, fontSize: 12, color: '#999' }}>
                          截止时间: {deadline.toLocaleString('zh-CN')}
                        </p>
                        {isExpired && (
                          <Tag color="red">已过期</Tag>
                        )}
                      </div>
                    }
                  />
                </Card>
              </List.Item>
            )
          }}
        />
      </Spin>
    </div>
  )
}

export default GroupListPage
