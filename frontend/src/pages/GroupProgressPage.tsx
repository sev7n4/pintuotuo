import React, { useEffect, useState } from 'react';
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
  Input,
} from 'antd';
import {
  ShareAltOutlined,
  CopyOutlined,
  CloseOutlined,
  UserOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  CrownOutlined,
} from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import { useGroupStore } from '@/stores/groupStore';
import { useAuthStore } from '@/stores/authStore';
import { groupService } from '@/services/group';
import type { APIResponse, GroupMemberPublic } from '@/types';
import { getApiErrorMessage } from '@/utils/apiError';
import { copyTextToClipboard } from '@/utils/clipboardCopy';
import {
  groupProgressBarStatus,
  groupStatusTagLabel,
  isGroupPastDeadline,
} from '@/utils/groupDisplay';
import { IconHintButton } from '@/components/IconHintButton';

const { Title, Text } = Typography;

const statusConfig: Record<string, { color: string; label: string; icon: React.ReactNode }> = {
  active: { color: 'processing', label: '进行中', icon: <ClockCircleOutlined /> },
  completed: { color: 'success', label: '已成团', icon: <CheckCircleOutlined /> },
  failed: { color: 'error', label: '已失败', icon: <ExclamationCircleOutlined /> },
};

export const GroupProgressPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const user = useAuthStore((s) => s.user);
  const { currentGroup, isLoading, error, getGroupProgress, cancelGroup } = useGroupStore();
  const [cancelModalVisible, setCancelModalVisible] = useState(false);
  const [countdown, setCountdown] = useState('');
  const [members, setMembers] = useState<GroupMemberPublic[]>([]);

  useEffect(() => {
    if (id) {
      void getGroupProgress(parseInt(id, 10));
    }
  }, [id, getGroupProgress]);

  useEffect(() => {
    if (!id) return;
    let cancelled = false;
    (async () => {
      try {
        const res = await groupService.listGroupMembers(parseInt(id, 10));
        const body = res.data as APIResponse<GroupMemberPublic[]>;
        const next = Array.isArray(body?.data) ? body.data : [];
        if (!cancelled) {
          queueMicrotask(() => setMembers(next));
        }
      } catch {
        if (!cancelled) queueMicrotask(() => setMembers([]));
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [id]);

  useEffect(() => {
    if (!currentGroup || currentGroup.status !== 'active') return;

    const updateCountdown = () => {
      const deadline = new Date(currentGroup.deadline).getTime();
      const now = Date.now();
      const diff = deadline - now;

      if (diff <= 0) {
        setCountdown('已截止');
        return;
      }

      const hours = Math.floor(diff / (1000 * 60 * 60));
      const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
      const seconds = Math.floor((diff % (1000 * 60)) / 1000);

      setCountdown(`${hours}小时${minutes}分钟${seconds}秒`);
    };

    updateCountdown();
    const timer = setInterval(updateCountdown, 1000);

    return () => clearInterval(timer);
  }, [currentGroup]);

  const shareJoinURL = () => {
    if (!currentGroup) return '';
    return `${window.location.origin}/groups/${currentGroup.id}/join`;
  };

  const handleShare = async () => {
    const shareUrl = shareJoinURL();
    if (!shareUrl) return;
    const ok = await copyTextToClipboard(shareUrl);
    if (ok) {
      message.success('分享链接已复制到剪贴板');
      return;
    }
    Modal.info({
      title: '请手动复制链接',
      width: 520,
      content: (
        <Input.TextArea
          readOnly
          autoSize={{ minRows: 2, maxRows: 4 }}
          value={shareUrl}
          style={{ marginTop: 8 }}
        />
      ),
    });
  };

  const handleCancel = async () => {
    if (!currentGroup) return;
    try {
      await cancelGroup(currentGroup.id);
      message.success('拼团已取消');
      setCancelModalVisible(false);
      navigate('/groups');
    } catch (err) {
      message.error(getApiErrorMessage(err, '取消失败，请稍后重试'));
    }
  };

  if (isLoading) {
    return (
      <div
        style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}
      >
        <Spin size="large" tip="加载中..." />
      </div>
    );
  }

  if (error) {
    return <Empty description={`错误: ${error}`} />;
  }

  if (!currentGroup) {
    return <Empty description="拼团不存在" />;
  }

  const statusInfo = statusConfig[currentGroup.status] || statusConfig.active;
  const progress = (currentGroup.current_count / currentGroup.target_count) * 100;
  const remainingCount = Math.max(0, currentGroup.target_count - currentGroup.current_count);
  const progressTagLabel = groupStatusTagLabel(currentGroup);
  const tagColor =
    currentGroup.status === 'completed'
      ? 'success'
      : currentGroup.status === 'failed'
        ? 'error'
        : progressTagLabel === '已截止'
          ? 'warning'
          : statusInfo.color;
  const tagIcon = progressTagLabel === '已截止' ? <ExclamationCircleOutlined /> : statusInfo.icon;
  const isCreator = user?.id != null && currentGroup.creator_id === user.id;

  if (currentGroup.status === 'completed') {
    return (
      <div style={{ padding: '20px', maxWidth: 600, margin: '0 auto' }}>
        <Result
          status="success"
          icon={<CheckCircleOutlined />}
          title="拼团成功！"
          subTitle={`拼团编号: #${currentGroup.id}`}
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
    );
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
            <Button type="primary" key="retry" onClick={() => navigate('/catalog')}>
              重新购买
            </Button>,
            <Button key="home" onClick={() => navigate('/')}>
              返回首页
            </Button>,
          ]}
        />
      </div>
    );
  }

  return (
    <div
      style={{
        padding: 12,
        maxWidth: 'min(600px, calc(100vw - 24px))',
        width: '100%',
        margin: '0 auto',
        boxSizing: 'border-box',
        overflowX: 'hidden',
      }}
    >
      <Card style={{ maxWidth: '100%' }}>
        <div style={{ marginBottom: 20 }}>
          <Space align="center" wrap>
            <Title level={3} style={{ margin: 0 }}>
              拼团进度
            </Title>
            <Tag color={tagColor} icon={tagIcon}>
              {progressTagLabel}
            </Tag>
          </Space>
        </div>

        <div style={{ marginBottom: 20 }}>
          <Text type="secondary">拼团编号：#{currentGroup.id}</Text>
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
            status={groupProgressBarStatus(currentGroup)}
            format={() => `还需${remainingCount}人`}
          />

          <div style={{ marginTop: 20, textAlign: 'center' }}>
            <ClockCircleOutlined style={{ marginRight: 8 }} />
            <Text>剩余时间：{countdown}</Text>
          </div>
        </Card>

        <Card title="成员列表" style={{ marginBottom: 20 }}>
          <List
            dataSource={members}
            locale={{ emptyText: '暂无成员数据' }}
            renderItem={(m) => (
              <List.Item>
                <List.Item.Meta
                  avatar={
                    <Avatar
                      icon={m.is_creator ? <CrownOutlined /> : <UserOutlined />}
                      style={{
                        backgroundColor: m.is_creator ? '#faad14' : '#1890ff',
                        color: '#fff',
                      }}
                    />
                  }
                  title={
                    <Space size={8} wrap>
                      <span>{m.display_name}</span>
                      {m.is_creator && <Tag color="gold">团长</Tag>}
                    </Space>
                  }
                  description="已参团"
                />
              </List.Item>
            )}
          />
          {remainingCount > 0 && (
            <Text type="secondary" style={{ display: 'block', marginTop: 8 }}>
              还差 {remainingCount} 个名额，邀请好友参团吧
            </Text>
          )}
          {isGroupPastDeadline(currentGroup.deadline) && currentGroup.status === 'active' && (
            <Text type="warning" style={{ display: 'block', marginTop: 8 }}>
              已超过截止时间，系统将按规则更新拼团状态；若仍显示进行中，请稍候刷新。
            </Text>
          )}
        </Card>

        <Space direction="vertical" style={{ width: '100%' }}>
          <IconHintButton
            type="primary"
            icon={<ShareAltOutlined />}
            hint="复制邀请链接，分享给好友参团"
            onClick={() => void handleShare()}
            block
            size="large"
          />
          <IconHintButton
            icon={<CopyOutlined />}
            hint="复制邀请链接"
            onClick={() => void handleShare()}
            block
            size="large"
          />
          {isCreator && (
            <Button
              danger
              icon={<CloseOutlined />}
              onClick={() => setCancelModalVisible(true)}
              block
            >
              取消拼团
            </Button>
          )}
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
        <p style={{ color: '#999', fontSize: 13 }}>
          仅发起人可以取消整团；其他成员如需退出，请在订单中取消未支付订单。
        </p>
      </Modal>
    </div>
  );
};

export default GroupProgressPage;
