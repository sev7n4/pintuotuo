import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Row, Col, Avatar, List, Button, Divider, Tag, Space, Statistic, Spin } from 'antd';
import {
  UserOutlined,
  QuestionCircleOutlined,
  SafetyOutlined,
  RightOutlined,
  TrophyOutlined,
  CodeOutlined,
  GiftOutlined,
} from '@ant-design/icons';
import { useAuthStore } from '@/stores/authStore';
import { referralService } from '@/services/referral';
import styles from './MyPage.module.css';

const MyPage = () => {
  const { user, isAuthenticated, fetchUser } = useAuthStore();
  const navigate = useNavigate();
  const [inviteCount, setInviteCount] = useState<number | null>(null);
  const [inviteCountError, setInviteCountError] = useState(false);

  useEffect(() => {
    if (isAuthenticated && !user) {
      fetchUser();
    }
  }, [isAuthenticated, user, fetchUser]);

  useEffect(() => {
    if (!isAuthenticated) {
      setInviteCount(null);
      setInviteCountError(false);
      return;
    }
    let cancelled = false;
    setInviteCount(null);
    setInviteCountError(false);
    referralService
      .getReferralStats()
      .then((res) => {
        const n = res.data?.total_referrals;
        if (!cancelled) {
          setInviteCount(typeof n === 'number' ? n : 0);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setInviteCountError(true);
          setInviteCount(0);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [isAuthenticated]);

  const getRoleTag = (role: string) => {
    const roleMap: Record<string, { color: string; text: string }> = {
      user: { color: 'blue', text: '普通用户' },
      merchant: { color: 'green', text: '商家' },
      admin: { color: 'red', text: '管理员' },
    };
    const { color, text } = roleMap[role] || { color: 'default', text: role };
    return <Tag color={color}>{text}</Tag>;
  };

  const getUserLevel = (createdAt: string) => {
    const days = Math.floor((Date.now() - new Date(createdAt).getTime()) / (1000 * 60 * 60 * 24));
    if (days < 30) return { level: 1, name: '新用户', progress: (days / 30) * 100 };
    if (days < 90) return { level: 2, name: '活跃用户', progress: ((days - 30) / 60) * 100 };
    if (days < 180) return { level: 3, name: '忠诚用户', progress: ((days - 90) / 90) * 100 };
    return { level: 4, name: '资深用户', progress: 100 };
  };

  const userLevel = user
    ? getUserLevel(user.created_at)
    : { level: 1, name: '新用户', progress: 0 };

  const settingItems = [
    {
      title: '个人资料',
      icon: <UserOutlined />,
      link: '/profile',
    },
    {
      title: '账户安全',
      icon: <SafetyOutlined />,
      link: '/profile',
      tab: 'security',
    },
    {
      title: '帮助中心',
      icon: <QuestionCircleOutlined />,
      link: '/help',
    },
  ];

  const handleMenuClick = (link: string, tab?: string) => {
    if (tab) {
      navigate(`${link}?tab=${tab}`);
    } else {
      navigate(link);
    }
  };

  if (!isAuthenticated) {
    return (
      <div className={styles.myPage}>
        <Card className={styles.loginCard}>
          <div className={styles.loginPrompt}>
            <UserOutlined className={styles.loginIcon} />
            <p>请先登录查看个人信息</p>
            <Button type="primary" onClick={() => navigate('/login')}>
              立即登录
            </Button>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div className={styles.myPage}>
      <Card className={styles.userCard}>
        <div className={styles.userInfo}>
          <Avatar size={80} icon={<UserOutlined />} className={styles.avatar} />
          <div className={styles.userDetail}>
            <h2 className={styles.userName}>{user?.name || '用户'}</h2>
            <Space>
              {user && getRoleTag(user.role)}
              <span className={styles.userEmail}>{user?.email}</span>
            </Space>
            <div className={styles.levelSection}>
              <TrophyOutlined className={styles.levelIcon} />
              <span className={styles.levelName}>{userLevel.name}</span>
              <span className={styles.levelBadge}>Lv.{userLevel.level}</span>
            </div>
          </div>
        </div>
        <Divider />
        <Row gutter={16}>
          <Col span={8}>
            <Statistic
              title="注册天数"
              value={Math.floor(
                (Date.now() - new Date(user?.created_at || Date.now()).getTime()) /
                  (1000 * 60 * 60 * 24)
              )}
              suffix="天"
            />
          </Col>
          <Col span={8}>
            <Statistic title="用户等级" value={userLevel.level} prefix="Lv." />
          </Col>
          <Col span={8}>
            {inviteCount === null ? (
              <div style={{ paddingTop: 8 }}>
                <Spin size="small" />
                <div style={{ fontSize: 12, color: '#999', marginTop: 8 }}>邀请人数</div>
              </div>
            ) : (
              <Statistic
                title="邀请人数"
                value={inviteCountError ? '—' : inviteCount}
                suffix={inviteCountError ? '' : '人'}
              />
            )}
          </Col>
        </Row>
      </Card>

      <Card size="small" style={{ marginBottom: 16 }} bodyStyle={{ padding: '12px 16px' }}>
        <Space align="center" style={{ width: '100%', justifyContent: 'space-between' }} wrap>
          <span style={{ color: 'rgba(0,0,0,0.65)' }}>邀请好友首单消费，你可获得奖励</span>
          <Button
            type="link"
            icon={<GiftOutlined />}
            onClick={() => navigate('/referral')}
            style={{ padding: 0 }}
          >
            邀请得奖励
          </Button>
        </Space>
      </Card>

      <div className={styles.servicesEntry}>
        <Space wrap>
          <Button type="primary" ghost onClick={() => navigate('/my/services')}>
            进入我的服务
          </Button>
          <Button
            type="default"
            icon={<CodeOutlined />}
            onClick={() => navigate('/developer/quickstart')}
          >
            开发者中心
          </Button>
        </Space>
      </div>

      <Card className={styles.settingCard} title="设置">
        <List
          dataSource={settingItems}
          renderItem={(item) => (
            <List.Item
              className={styles.settingItem}
              onClick={() => handleMenuClick(item.link, item.tab)}
            >
              <div className={styles.settingLeft}>
                <span className={styles.settingIcon}>{item.icon}</span>
                <span>{item.title}</span>
              </div>
              <RightOutlined className={styles.settingArrow} />
            </List.Item>
          )}
        />
      </Card>
    </div>
  );
};

export default MyPage;
