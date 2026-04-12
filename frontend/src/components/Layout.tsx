import { useEffect, useMemo } from 'react';
import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom';
import { Layout as AntLayout, Menu, Dropdown, Avatar, Space, message } from 'antd';
import {
  UserOutlined,
  LogoutOutlined,
  HomeOutlined,
  CustomerServiceOutlined,
  AppstoreOutlined,
  ShoppingCartOutlined,
  UnorderedListOutlined,
} from '@ant-design/icons';
import { useAuthStore } from '@/stores/authStore';
import './Layout.css';

const { Header, Content, Footer } = AntLayout;

function getMainNavSelectedKey(pathname: string): string[] {
  if (pathname === '/') return ['home'];
  if (pathname.startsWith('/catalog')) return ['catalog'];
  if (pathname === '/cart' || pathname === '/checkout') return ['cart'];
  if (pathname.startsWith('/orders') || pathname.startsWith('/payment')) return ['orders'];
  return [];
}

export default function Layout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, logout, isAuthenticated, fetchUser } = useAuthStore();

  useEffect(() => {
    if (isAuthenticated && !user) {
      fetchUser();
    }
  }, [isAuthenticated, user, fetchUser]);

  const tabItems = useMemo(
    () => [
      {
        key: 'home',
        label: <Link to="/">首页</Link>,
        icon: <HomeOutlined />,
      },
      {
        key: 'catalog',
        label: <Link to="/catalog">卖场</Link>,
        icon: <AppstoreOutlined />,
      },
      {
        key: 'cart',
        label: <Link to="/cart">购物车</Link>,
        icon: <ShoppingCartOutlined />,
      },
      ...(isAuthenticated
        ? [
            {
              key: 'orders',
              label: <Link to="/orders">我的订单</Link>,
              icon: <UnorderedListOutlined />,
            },
          ]
        : []),
    ],
    [isAuthenticated]
  );

  const userMenuItems = [
    { key: 'my', label: '我的主页', icon: <UserOutlined /> },
    { key: 'services', label: '我的服务', icon: <CustomerServiceOutlined /> },
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
        navigate('/my/services');
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
          selectedKeys={getMainNavSelectedKey(location.pathname)}
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
