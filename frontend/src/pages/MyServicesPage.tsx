import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Row, Col, Button } from 'antd';
import {
  UserOutlined,
  WalletOutlined,
  HeartOutlined,
  HistoryOutlined,
  BarChartOutlined,
  TeamOutlined,
  ShoppingCartOutlined,
} from '@ant-design/icons';
import { useAuthStore } from '@/stores/authStore';
import styles from './MyPage.module.css';

/** 与「我的主页」同级的功能入口：订单、Token、收藏、历史、消费、邀请 */
const MyServicesPage = () => {
  const { user, isAuthenticated, fetchUser } = useAuthStore();
  const navigate = useNavigate();

  useEffect(() => {
    if (isAuthenticated && !user) {
      fetchUser();
    }
  }, [isAuthenticated, user, fetchUser]);

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
      title: '我的收藏',
      icon: <HeartOutlined />,
      link: '/favorites',
      color: '#ff4d4f',
    },
    {
      title: '浏览历史',
      icon: <HistoryOutlined />,
      link: '/history',
      color: '#722ed1',
    },
    {
      title: '消费明细',
      icon: <BarChartOutlined />,
      link: '/consumption',
      color: '#13c2c2',
    },
    {
      title: '邀请好友',
      icon: <TeamOutlined />,
      link: '/referral',
      color: '#faad14',
    },
  ];

  const handleMenuClick = (link: string) => {
    navigate(link);
  };

  if (!isAuthenticated) {
    return (
      <div className={styles.myPage}>
        <Card className={styles.loginCard}>
          <div className={styles.loginPrompt}>
            <UserOutlined className={styles.loginIcon} />
            <p>请先登录查看我的服务</p>
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
      <Card className={styles.menuCard} title="我的服务">
        <Row gutter={[16, 16]}>
          {menuItems.map((item) => (
            <Col xs={12} sm={6} key={item.title}>
              <div className={styles.menuItem} onClick={() => handleMenuClick(item.link)}>
                <div
                  className={styles.menuIcon}
                  style={{ backgroundColor: item.color + '20', color: item.color }}
                >
                  {item.icon}
                </div>
                <span className={styles.menuTitle}>{item.title}</span>
              </div>
            </Col>
          ))}
        </Row>
      </Card>
    </div>
  );
};

export default MyServicesPage;
