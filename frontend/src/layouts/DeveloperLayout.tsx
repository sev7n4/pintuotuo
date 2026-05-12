import { useEffect, useState } from 'react';
import { Layout, Menu, Button, Spin } from 'antd';
import {
  RocketOutlined,
  KeyOutlined,
  ReadOutlined,
  LineChartOutlined,
  BugOutlined,
  AppstoreOutlined,
  HomeOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation, Link } from 'react-router-dom';
import { useAuthStore } from '@/stores/authStore';
import styles from './DeveloperLayout.module.css';

const { Header, Sider, Content } = Layout;

const menuItems = [
  { key: '/developer/quickstart', icon: <RocketOutlined />, label: '快速开始' },
  { key: '/developer/keys', icon: <KeyOutlined />, label: '密钥与安全' },
  { key: '/developer/ide-clients', icon: <ReadOutlined />, label: 'IDE 与 CLI 接入' },
  { key: '/developer/usage', icon: <LineChartOutlined />, label: '用量与账单' },
  { key: '/developer/troubleshoot', icon: <BugOutlined />, label: '错误与排障' },
  { key: '/developer/models', icon: <AppstoreOutlined />, label: '模型与权益' },
];

export default function DeveloperLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const { isAuthenticated, fetchUser, user } = useAuthStore();
  const [ready, setReady] = useState(false);
  const [siderCollapsed, setSiderCollapsed] = useState(() =>
    typeof window !== 'undefined' ? window.matchMedia('(max-width: 991px)').matches : false
  );

  useEffect(() => {
    const mq = window.matchMedia('(max-width: 991px)');
    const sync = () => setSiderCollapsed(mq.matches);
    sync();
    mq.addEventListener('change', sync);
    return () => mq.removeEventListener('change', sync);
  }, []);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      if (!isAuthenticated) {
        navigate(`/login?redirect=${encodeURIComponent(location.pathname + location.search)}`, {
          replace: true,
        });
        return;
      }
      if (!user) {
        await fetchUser();
      }
      if (!cancelled) setReady(true);
    })();
    return () => {
      cancelled = true;
    };
  }, [isAuthenticated, user, fetchUser, navigate, location.pathname, location.search]);

  if (!isAuthenticated || !ready) {
    return (
      <div style={{ padding: 48, textAlign: 'center' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <Layout className={styles.root}>
      <Header className={styles.header}>
        <div className={styles.brand}>
          <Button
            type="text"
            className={styles.siderToggle}
            aria-label={siderCollapsed ? '展开侧栏' : '收起侧栏'}
            icon={siderCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setSiderCollapsed((v) => !v)}
          />
          <Link to="/">
            <Button type="link" icon={<HomeOutlined />}>
              返回卖场
            </Button>
          </Link>
          <h1 className={styles.title}>开发者中心</h1>
        </div>
        <Link to="/my-tokens">
          <Button type="default">我的 Token</Button>
        </Link>
      </Header>
      <Layout className={styles.innerLayout}>
        <Sider
          width={220}
          breakpoint="lg"
          collapsedWidth={0}
          collapsible
          collapsed={siderCollapsed}
          onCollapse={setSiderCollapsed}
          trigger={null}
          className={styles.sider}
          theme="light"
        >
          <Menu
            mode="inline"
            data-testid="developer-sider"
            selectedKeys={
              menuItems.some((m) => m.key === location.pathname)
                ? [location.pathname]
                : ['/developer/quickstart']
            }
            items={menuItems}
            onClick={({ key }) => navigate(key)}
            style={{ borderRight: 0 }}
          />
        </Sider>
        <Content className={styles.content}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
