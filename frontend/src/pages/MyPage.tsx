import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, Row, Col, Avatar, List, Button, Divider, Tag, Space, Statistic } from 'antd'
import {
  UserOutlined,
  WalletOutlined,
  HeartOutlined,
  HistoryOutlined,
  SettingOutlined,
  SafetyOutlined,
  RightOutlined,
  TrophyOutlined,
  BarChartOutlined,
  ShoppingCartOutlined,
  TeamOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'
import styles from './MyPage.module.css'

const MyPage = () => {
  const { user, isAuthenticated, fetchUser } = useAuthStore()
  const navigate = useNavigate()

  useEffect(() => {
    if (isAuthenticated && !user) {
      fetchUser()
    }
  }, [isAuthenticated, user, fetchUser])

  const getRoleTag = (role: string) => {
    const roleMap: Record<string, { color: string; text: string }> = {
      user: { color: 'blue', text: '普通用户' },
      merchant: { color: 'green', text: '商家' },
      admin: { color: 'red', text: '管理员' },
    }
    const { color, text } = roleMap[role] || { color: 'default', text: role }
    return <Tag color={color}>{text}</Tag>
  }

  const getUserLevel = (createdAt: string) => {
    const days = Math.floor((Date.now() - new Date(createdAt).getTime()) / (1000 * 60 * 60 * 24))
    if (days < 30) return { level: 1, name: '新用户', progress: days / 30 * 100 }
    if (days < 90) return { level: 2, name: '活跃用户', progress: (days - 30) / 60 * 100 }
    if (days < 180) return { level: 3, name: '忠诚用户', progress: (days - 90) / 90 * 100 }
    return { level: 4, name: '资深用户', progress: 100 }
  }

  const userLevel = user ? getUserLevel(user.created_at) : { level: 1, name: '新用户', progress: 0 }

  const menuItems = [
    {
      title: '我的订单',
      icon: <ShoppingCartOutlined />,
      link: '/orders',
      color: '#1890ff',
    },
    {
      title: '我的Token',
      icon: <WalletOutlined />,
      link: '/my-tokens',
      color: '#52c41a',
    },
    {
      title: '消费明细',
      icon: <BarChartOutlined />,
      link: '/consumption',
      color: '#722ed1',
    },
    {
      title: '邀请好友',
      icon: <TeamOutlined />,
      link: '/referral',
      color: '#faad14',
    },
  ]

  const settingItems = [
    {
      title: '个人资料',
      icon: <UserOutlined />,
      link: '/profile',
    },
    {
      title: '我的收藏',
      icon: <HeartOutlined />,
      link: '/favorites',
    },
    {
      title: '浏览历史',
      icon: <HistoryOutlined />,
      link: '/history',
    },
    {
      title: '账户安全',
      icon: <SafetyOutlined />,
      link: '/profile',
      tab: 'security',
    },
    {
      title: '帮助中心',
      icon: <SettingOutlined />,
      link: '/help',
    },
  ]

  const handleMenuClick = (link: string, tab?: string) => {
    if (tab) {
      navigate(`${link}?tab=${tab}`)
    } else {
      navigate(link)
    }
  }

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
    )
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
              value={Math.floor((Date.now() - new Date(user?.created_at || Date.now()).getTime()) / (1000 * 60 * 60 * 24))} 
              suffix="天" 
            />
          </Col>
          <Col span={8}>
            <Statistic title="用户等级" value={userLevel.level} prefix="Lv." />
          </Col>
          <Col span={8}>
            <Statistic title="邀请人数" value={0} suffix="人" />
          </Col>
        </Row>
      </Card>

      <Card className={styles.menuCard} title="我的服务">
        <Row gutter={[16, 16]}>
          {menuItems.map((item) => (
            <Col xs={12} sm={6} key={item.title}>
              <div 
                className={styles.menuItem} 
                onClick={() => handleMenuClick(item.link)}
              >
                <div className={styles.menuIcon} style={{ backgroundColor: item.color + '20', color: item.color }}>
                  {item.icon}
                </div>
                <span className={styles.menuTitle}>{item.title}</span>
              </div>
            </Col>
          ))}
        </Row>
      </Card>

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
  )
}

export default MyPage
