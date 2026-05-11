import React, { useCallback, useEffect, useState } from 'react';
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
  Segmented,
} from 'antd';
import { UserAddOutlined, ShoppingOutlined, TagsOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useGroupStore } from '@stores/groupStore';
import { useAuthStore } from '@/stores/authStore';
import type { Group } from '@/types';
import type { GroupListScope } from '@/services/group';
import { getApiErrorMessage } from '@/utils/apiError';
import {
  groupProductSubtitle,
  groupProgressBarStatus,
  groupStatusTagLabel,
  isGroupJoinableByState,
  isGroupPastDeadline,
} from '@/utils/groupDisplay';

const { Title, Text } = Typography;
const { useBreakpoint } = Grid;

interface GroupWithStore extends Group {
  onJoin?: () => void;
  isLoading?: boolean;
}

export const GroupListPage: React.FC = () => {
  const navigate = useNavigate();
  const screens = useBreakpoint();
  const user = useAuthStore((s) => s.user);
  const { isLoading, error, fetchGroups, joinGroup } = useGroupStore();
  const [groups, setGroups] = useState<GroupWithStore[]>([]);
  const [joiningId, setJoiningId] = useState<number | null>(null);
  const [listScope, setListScope] = useState<GroupListScope>('all');

  const loadGroups = useCallback(async () => {
    const result = await fetchGroups(1, 40, { scope: listScope, status: 'active' });
    if (result) {
      setGroups(result);
    } else {
      setGroups([]);
    }
  }, [listScope, fetchGroups]);

  useEffect(() => {
    void loadGroups();
  }, [loadGroups]);

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
      message.error(getApiErrorMessage(err, '加入失败，请稍后重试'));
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

  if (isLoading && groups.length === 0) {
    return (
      <div style={{ padding: screens.xs ? 12 : 24, textAlign: 'center', marginTop: 48 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (groups.length === 0) {
    let emptyPrimary = '暂无进行中的拼团';
    let emptySecondary = '去商品页面发起拼团吧！';
    if (listScope === 'mine_involved') {
      emptyPrimary = '暂无你发起或参团的进行中拼团';
      emptySecondary = '去卖场开团或到「全站热团」加入别人的团。';
    } else if (listScope === 'mine_created') {
      emptyPrimary = '暂无你发起的进行中拼团';
      emptySecondary = '去卖场发起一个新团，或切换到「全站热团」看看。';
    } else if (listScope === 'mine_joined') {
      emptyPrimary = '暂无你跟团的记录';
      emptySecondary = '切换到「全站热团」发现更多可参团的团。';
    }
    return (
      <div style={{ padding: screens.xs ? 12 : 24, marginTop: 50, textAlign: 'center' }}>
        <Segmented<GroupListScope>
          value={listScope}
          onChange={(v) => setListScope(v as GroupListScope)}
          options={[
            { label: '全站热团', value: 'all' },
            { label: '我的拼团', value: 'mine_involved' },
            { label: '我发起', value: 'mine_created' },
            { label: '我跟团', value: 'mine_joined' },
          ]}
          style={{ display: 'block', maxWidth: 520, margin: '0 auto 24px' }}
        />
        <Empty
          description={
            <Space direction="vertical">
              <span>{emptyPrimary}</span>
              <span style={{ color: '#999', fontSize: 14 }}>{emptySecondary}</span>
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

      <Segmented<GroupListScope>
        value={listScope}
        onChange={(v) => setListScope(v as GroupListScope)}
        options={[
          { label: '全站热团', value: 'all' },
          { label: '我的拼团', value: 'mine_involved' },
          { label: '我发起', value: 'mine_created' },
          { label: '我跟团', value: 'mine_joined' },
        ]}
        style={{ marginBottom: 16, maxWidth: 520 }}
      />

      <Spin spinning={isLoading}>
        <List
          grid={{ gutter: 16, xs: 1, sm: 2, md: 2, lg: 3, xl: 3, xxl: 4 }}
          dataSource={groups}
          renderItem={(group) => {
            const progress = (group.current_count / group.target_count) * 100;
            const deadline = new Date(group.deadline);
            const isExpired = isGroupPastDeadline(group.deadline);
            const tagLabel = groupStatusTagLabel(group);
            const tagColor =
              group.status === 'completed'
                ? 'green'
                : group.status === 'failed' || tagLabel === '已截止'
                  ? 'default'
                  : 'blue';
            const subtitle = groupProductSubtitle(group);
            const imCreator = user?.id != null && group.creator_id === user.id;
            const canJoin = isGroupJoinableByState(group) && !imCreator;
            let joinLabel = '加入拼团';
            if (canJoin) {
              joinLabel = '加入拼团';
            } else if (imCreator && group.status === 'active' && !isExpired) {
              joinLabel = '我发起的';
            } else {
              joinLabel = tagLabel;
            }

            return (
              <List.Item key={group.id}>
                <Card
                  hoverable
                  onClick={() => handleViewGroupDetail(group.id)}
                  actions={[
                    <Button
                      type={canJoin ? 'primary' : 'default'}
                      icon={<UserAddOutlined />}
                      onClick={(e) => {
                        e.stopPropagation();
                        handleJoinGroup(group.id);
                      }}
                      loading={joiningId === group.id}
                      disabled={!canJoin}
                    >
                      {joinLabel}
                    </Button>,
                  ]}
                >
                  <Card.Meta
                    title={
                      <Space direction="vertical" size={0}>
                        <span>拼团 #{group.id}</span>
                        {subtitle && (
                          <Space size={4}>
                            <TagsOutlined style={{ fontSize: 12, color: '#999' }} />
                            <Text type="secondary" style={{ fontSize: 12 }}>
                              {subtitle}
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
                          <Tag color={tagColor} style={{ marginLeft: 10 }}>
                            {tagLabel}
                          </Tag>
                        </p>
                        <Progress
                          percent={Math.round(progress)}
                          status={groupProgressBarStatus(group)}
                          size="small"
                        />
                        <p style={{ marginTop: 10, fontSize: 12, color: '#999' }}>
                          截止时间: {deadline.toLocaleString('zh-CN')}
                        </p>
                        {isExpired && group.status === 'active' && (
                          <Tag color="warning" style={{ marginTop: 8 }}>
                            已超过截止时间
                          </Tag>
                        )}
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
