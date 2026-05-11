import React, { useEffect, useState } from 'react';
import {
  List,
  Card,
  Button,
  Tag,
  Progress,
  Spin,
  Empty,
  message,
  Typography,
  Space,
  Grid,
} from 'antd';
import { UserAddOutlined, ShoppingOutlined, TagsOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useGroupStore } from '@stores/groupStore';
import type { Group } from '@/types';

const { Title, Text } = Typography;
const { useBreakpoint } = Grid;

const statusMap: Record<string, { color: string; label: string }> = {
  active: { color: 'blue', label: '进行中' },
  completed: { color: 'green', label: '已成团' },
  failed: { color: 'red', label: '已失败' },
};

interface GroupWithStore extends Group {
  onJoin?: () => void;
  isLoading?: boolean;
}

export const GroupListPage: React.FC = () => {
  const navigate = useNavigate();
  const screens = useBreakpoint();
  const { isLoading, error, fetchGroups, joinGroup } = useGroupStore();
  const [groups, setGroups] = useState<GroupWithStore[]>([]);
  const [joiningId, setJoiningId] = useState<number | null>(null);

  useEffect(() => {
    loadGroups();
  }, []);

  const loadGroups = async () => {
    const result = await fetchGroups();
    if (result) {
      setGroups(result);
    }
  };

  const handleJoinGroup = async (groupId: number) => {
    setJoiningId(groupId);
    try {
      const orderId = await joinGroup(groupId);
      message.success('加入拼团成功！请完成支付');
      if (orderId) {
        navigate(`/payment/${orderId}`);
      } else {
        await loadGroups();
      }
    } catch (err) {
      message.error('加入失败，请稍后重试');
    } finally {
      setJoiningId(null);
    }
  };

  const handleViewGroupDetail = (groupId: number) => {
    navigate(`/groups/${groupId}`);
  };

  if (error) {
    return <Empty description={`错误: ${error}`} />;
  }

  if (groups.length === 0) {
    return (
      <div style={{ padding: screens.xs ? 12 : 24, marginTop: 50, textAlign: 'center' }}>
        <Empty
          description={
            <Space direction="vertical">
              <span>暂无进行中的拼团</span>
              <span style={{ color: '#999', fontSize: 14 }}>去商品页面发起拼团吧！</span>
            </Space>
          }
        />
        <Button
          type="primary"
          icon={<ShoppingOutlined />}
          style={{ marginTop: 16 }}
          onClick={() => navigate('/catalog?group_enabled=true')}
        >
          浏览可拼团商品
        </Button>
      </div>
    );
  }

  return (
    <div style={{ padding: screens.xs ? 12 : 24 }}>
      <div
        style={{
          marginBottom: 20,
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <Title level={screens.xs ? 4 : 3} style={{ margin: 0 }}>
          拼团中心
        </Title>
        <Button
          type="primary"
          icon={<ShoppingOutlined />}
          onClick={() => navigate('/catalog?group_enabled=true')}
        >
          {screens.xs ? '' : '去拼团'}
        </Button>
      </div>

      <Spin spinning={isLoading}>
        <List
          grid={{ gutter: 16, xs: 1, sm: 2, md: 2, lg: 3, xl: 3, xxl: 4 }}
          dataSource={groups}
          renderItem={(group) => {
            const progress = (group.current_count / group.target_count) * 100;
            const deadline = new Date(group.deadline);
            const isExpired = deadline < new Date();
            const s = statusMap[group.status] || { color: 'default', label: group.status };

            return (
              <List.Item key={group.id}>
                <Card
                  hoverable
                  onClick={() => handleViewGroupDetail(group.id)}
                  actions={[
                    <Button
                      type={group.status === 'active' ? 'primary' : 'default'}
                      icon={<UserAddOutlined />}
                      onClick={(e) => {
                        e.stopPropagation();
                        handleJoinGroup(group.id);
                      }}
                      loading={joiningId === group.id}
                      disabled={group.status !== 'active' || isExpired}
                    >
                      {group.status === 'active' ? '加入拼团' : '已' + s.label}
                    </Button>,
                  ]}
                >
                  <Card.Meta
                    title={
                      <Space direction="vertical" size={0}>
                        <span>拼团 #{group.id}</span>
                        {group.sku_name && (
                          <Space size={4}>
                            <TagsOutlined style={{ fontSize: 12, color: '#999' }} />
                            <Text type="secondary" style={{ fontSize: 12 }}>
                              {group.sku_name}
                            </Text>
                          </Space>
                        )}
                      </Space>
                    }
                    description={
                      <div>
                        {group.sku_type && (
                          <p style={{ margin: '8px 0' }}>
                            <Space>
                              <Text type="secondary">规格:</Text>
                              <Tag
                                color={
                                  group.sku_type === 'token_pack'
                                    ? 'blue'
                                    : group.sku_type === 'subscription'
                                      ? 'green'
                                      : 'orange'
                                }
                              >
                                {group.sku_type === 'token_pack'
                                  ? 'Token包'
                                  : group.sku_type === 'subscription'
                                    ? '订阅'
                                    : '并发'}
                              </Tag>
                              {group.sku_specs && (
                                <Text type="secondary" style={{ fontSize: 12 }}>
                                  {group.sku_specs}
                                </Text>
                              )}
                            </Space>
                          </p>
                        )}
                        {group.group_discount_rate && (
                          <p style={{ margin: '8px 0' }}>
                            <Tag color="red">
                              拼团折扣 {(group.group_discount_rate * 100).toFixed(0)}%
                            </Tag>
                          </p>
                        )}
                        <p style={{ margin: '8px 0' }}>目标人数: {group.target_count}人</p>
                        <p style={{ margin: '8px 0' }}>
                          当前人数: {group.current_count}人
                          <Tag color={s.color} style={{ marginLeft: 10 }}>
                            {s.label}
                          </Tag>
                        </p>
                        <Progress
                          percent={Math.round(progress)}
                          status={group.status === 'active' ? 'active' : 'success'}
                          size="small"
                        />
                        <p style={{ marginTop: 10, fontSize: 12, color: '#999' }}>
                          截止时间: {deadline.toLocaleString('zh-CN')}
                        </p>
                        {isExpired && <Tag color="red">已过期</Tag>}
                      </div>
                    }
                  />
                </Card>
              </List.Item>
            );
          }}
        />
      </Spin>
    </div>
  );
};

export default GroupListPage;
