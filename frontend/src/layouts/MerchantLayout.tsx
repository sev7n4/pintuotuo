import { useState, useEffect, useRef } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Avatar, Dropdown, message, Spin, Drawer, Button } from 'antd';
import {
  ShopOutlined,
  AppstoreOutlined,
  ShoppingCartOutlined,
  WalletOutlined,
  SettingOutlined,
  UserOutlined,
  LogoutOutlined,
  DashboardOutlined,
  KeyOutlined,
  GiftOutlined,
  FileTextOutlined,
  TeamOutlined,
  MenuOutlined,
  CloseOutlined,
} from '@ant-design/icons';
import { useAuthStore } from '@/stores/authStore';
import { useMerchantStore } from '@/stores/merchantStore';
import styles from './MerchantLayout.module.css';

const { Header, Sider, Content } = Layout;

const menuItems = [
  {
    key: '/merchant',
    icon: <DashboardOutlined />,
    label: '数据概览',
  },
  {
    key: '/merchant/products',
    icon: <AppstoreOutlined />,
    label: '商品管理',
  },
  {
    key: '/merchant/skus',
    icon: <ShopOutlined />,
    label: 'SKU管理',
  },
  {
    key: '/merchant/orders',
    icon: <ShoppingCartOutlined />,
    label: '订单管理',
  },
  {
    key: '/merchant/settlements',
    icon: <WalletOutlined />,
    label: '结算管理',
  },
  {
    key: '/merchant/bills',
    icon: <FileTextOutlined />,
    label: '月度账单',
  },
  {
    key: '/merchant/invoices',
    icon: <FileTextOutlined />,
    label: '发票管理',
  },
  {
    key: '/merchant/analytics',
    icon: <TeamOutlined />,
    label: '用户分析',
  },
  {
    key: '/merchant/marketing',
    icon: <GiftOutlined />,
    label: '营销工具',
  },
  {
    key: '/merchant/api-keys',
    icon: <KeyOutlined />,
    label: 'API密钥',
  },
  {
    key: '/merchant/settings',
    icon: <SettingOutlined />,
    label: '店铺设置',
  },
];

const MerchantLayout = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout, isAuthenticated, fetchUser } = useAuthStore();
  const { profile: merchantProfile, fetchProfile } = useMerchantStore();
  const [collapsed, setCollapsed] = useState(false);
  const [checkingAuth, setCheckingAuth] = useState(true);
  const [isMobile, setIsMobile] = useState(false);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const profileFetchedRef = useRef(false);
  const statusCheckedRef = useRef(false);

  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 992);
    };
    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  useEffect(() => {
    const checkAuth = async () => {
      const hasToken = !!(
        localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token')
      );
      if (!hasToken) {
        navigate('/login', { state: { from: location.pathname } });
        return;
      }

      if (!user) {
        try {
          await fetchUser();
        } catch {
          localStorage.removeItem('auth_token');
          sessionStorage.removeItem('auth_token');
          navigate('/login', { state: { from: location.pathname } });
          return;
        }
      }

      setCheckingAuth(false);
    };

    checkAuth();
  }, [isAuthenticated, user, fetchUser, navigate, location.pathname]);

  useEffect(() => {
    if (!checkingAuth && user && user.role === 'merchant' && !profileFetchedRef.current) {
      profileFetchedRef.current = true;
      fetchProfile().catch(() => {
        message.error('获取商户信息失败');
      });
    }
  }, [checkingAuth, user, fetchProfile]);

  useEffect(() => {
    if (
      !checkingAuth &&
      merchantProfile &&
      !statusCheckedRef.current &&
      location.pathname !== '/merchant/settings'
    ) {
      statusCheckedRef.current = true;
      if (merchantProfile.status === 'pending' || merchantProfile.status === 'reviewing') {
        message.warning('您的商户申请正在审核中，请先提交资料');
        navigate('/merchant/settings');
      }
    }
  }, [checkingAuth, merchantProfile, navigate, location.pathname]);

  useEffect(() => {
    if (!checkingAuth && user && user.role !== 'merchant') {
      message.error('无权限访问商户后台');
      navigate('/');
    }
  }, [checkingAuth, user, navigate]);

  if (checkingAuth) {
    return (
      <div
        style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}
      >
        <Spin size="large" />
      </div>
    );
  }

  if (!isAuthenticated || (user && user.role !== 'merchant')) {
    return null;
  }

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key);
    if (isMobile) {
      setDrawerVisible(false);
    }
  };

  const handleLogout = () => {
    logout();
    message.success('已退出登录');
    navigate('/login');
  };

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人中心',
      onClick: () => navigate('/'),
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
    },
  ];

  const getSelectedKey = () => {
    const path = location.pathname;
    if (path === '/merchant') return '/merchant';
    return path;
  };

  const toggleDrawer = () => {
    setDrawerVisible(!drawerVisible);
  };

  const menuContent = (
    <Menu
      mode="inline"
      selectedKeys={[getSelectedKey()]}
      items={menuItems}
      onClick={handleMenuClick}
      style={{ borderRight: 0 }}
    />
  );

  return (
    <Layout className={styles.layout}>
      {isMobile ? (
        <>
          <Drawer
            placement="left"
            closable={false}
            onClose={() => setDrawerVisible(false)}
            open={drawerVisible}
            className={styles.drawer}
            width={250}
          >
            <div className={styles.drawerHeader}>
              <ShopOutlined className={styles.logoIcon} />
              <span>商家后台</span>
              <Button
                type="text"
                icon={<CloseOutlined />}
                onClick={() => setDrawerVisible(false)}
                className={styles.closeBtn}
              />
            </div>
            {menuContent}
          </Drawer>
        </>
      ) : (
        <Sider
          collapsible
          collapsed={collapsed}
          onCollapse={setCollapsed}
          className={styles.sider}
          theme="light"
          breakpoint="lg"
          collapsedWidth={80}
        >
          <div className={styles.logo}>
            <ShopOutlined className={styles.logoIcon} />
            {!collapsed && <span>商家后台</span>}
          </div>
          {menuContent}
        </Sider>
      )}
      <Layout>
        <Header className={styles.header}>
          {isMobile && (
            <Button
              type="text"
              icon={<MenuOutlined />}
              onClick={toggleDrawer}
              className={styles.menuBtn}
            />
          )}
          <div className={styles.headerTitle}>{isMobile && <span>商家后台</span>}</div>
          <div className={styles.headerRight}>
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <div className={styles.userInfo}>
                <Avatar icon={<UserOutlined />} size={isMobile ? 'small' : 'default'} />
                {!isMobile && <span className={styles.userName}>{user?.name || '商家'}</span>}
              </div>
            </Dropdown>
          </div>
        </Header>
        <Content className={styles.content}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default MerchantLayout;
