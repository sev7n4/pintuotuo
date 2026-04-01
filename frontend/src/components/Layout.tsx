import { useEffect, useMemo } from 'react';
import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom';
import { Layout as AntLayout, Menu, Dropdown, Avatar, Space, message, Badge } from 'antd';
import {
  UserOutlined,
  LogoutOutlined,
  HomeOutlined,
  AppstoreOutlined,
  ShoppingCartOutlined,
} from '@ant-design/icons';
import { useAuthStore } from '@/stores/authStore';
import { useCartStore } from '@/stores/cartStore';
import './Layout.css';

const { Header, Content, Footer } = AntLayout;

export default function Layout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, logout, isAuthenticated, fetchUser } = useAuthStore();
  const { items } = useCartStore();

  useEffect(() => {
    if (isAuthenticated && !user) {
      fetchUser();
    }
  }, [isAuthenticated, user, fetchUser]);

  const cartItemCount = useMemo(() => items.reduce((sum, item) => sum + item.quantity, 0), [items]);

  const getSelectedTab = () => {
    const path = location.pathname;
    if (path === '/') return 'home';
    if (path === '/categories' || path.startsWith('/products')) return 'category';
    if (path === '/cart') return 'cart';
    if (
      path.startsWith('/orders') ||
      path.startsWith('/groups') ||
      path.startsWith('/payment') ||
      path === '/my-tokens' ||
      path === '/consumption'
    )
      return 'orders';
    if (
      path === '/my' ||
      path === '/profile' ||
      path === '/referral' ||
      path === '/favorites' ||
      path === '/history'
    )
      return 'my';
    return 'home';
  };

  const baseTabItems = [
    {
      key: 'home',
      label: <Link to="/">首页</Link>,
      icon: <HomeOutlined />,
    },
    {
      key: 'category',
      label: <Link to="/categories">分类</Link>,
      icon: <AppstoreOutlined />,
    },
    {
      key: 'cart',
      label: (
        <Link to="/cart">
          <Badge count={cartItemCount} size="small" offset={[6, 0]}>
            购物车
          </Badge>
        </Link>
      ),
      icon: <ShoppingCartOutlined />,
    },
  ];

  const authTabItems = [
    {
      key: 'orders',
      label: <Link to="/orders">订单</Link>,
      icon: <ShoppingCartOutlined />,
    },
  ];

  const tabItems = isAuthenticated ? [...baseTabItems, ...authTabItems] : baseTabItems;

  const userMenuItems = [
    { key: 'my', label: '我的主页', icon: <UserOutlined /> },
    { key: 'services', label: '我的服务', icon: <AppstoreOutlined /> },
    { type: 'divider' as const },
    { key: 'logout', label: '退出登录', icon: <LogoutOutlined /> },
  ];

  const handleUserMenuClick = async ({ key }: { key: string }) => {
    switch (key) {
      case 'logout':
        await logout();
        message.success('已退出登录');
        navigate('/login');
        break;
      case 'my':
        navigate('/my');
        break;
      case 'services':
        navigate('/my');
        break;
    }
  };

  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Header className="layout-header">
        <Link to="/" className="layout-logo">
          拼脱脱
        </Link>
        <Menu
          mode="horizontal"
          selectedKeys={[getSelectedTab()]}
          items={tabItems}
          className="layout-menu"
        />
        <div className="layout-user">
          {isAuthenticated && user ? (
            <Dropdown
              menu={{ items: userMenuItems, onClick: handleUserMenuClick }}
              placement="bottomRight"
            >
              <Space style={{ cursor: 'pointer' }} data-testid="user-dropdown">
                <Avatar icon={<UserOutlined />} />
                <span className="user-name">{user.name || user.email}</span>
              </Space>
            </Dropdown>
          ) : (
            <Space>
              <Link to="/login">登录</Link>
              <Link to="/register">注册</Link>
            </Space>
          )}
        </div>
      </Header>
      <Content className="layout-content">
        <Outlet />
      </Content>
      <Footer className="layout-footer">
        <p>&copy; 2026 拼脱脱 - AI Token 二级市场</p>
      </Footer>
    </AntLayout>
  );
}
