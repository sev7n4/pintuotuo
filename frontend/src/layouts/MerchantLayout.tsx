import { useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Layout, Menu, Avatar, Dropdown, message } from 'antd'
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
} from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'
import styles from './MerchantLayout.module.css'

const { Header, Sider, Content } = Layout

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
    key: '/merchant/api-keys',
    icon: <KeyOutlined />,
    label: 'API密钥',
  },
  {
    key: '/merchant/settings',
    icon: <SettingOutlined />,
    label: '店铺设置',
  },
]

const MerchantLayout = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuthStore()
  const [collapsed, setCollapsed] = useState(false)

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key)
  }

  const handleLogout = () => {
    logout()
    message.success('已退出登录')
    navigate('/login')
  }

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
  ]

  const getSelectedKey = () => {
    const path = location.pathname
    if (path === '/merchant') return '/merchant'
    return path
  }

  return (
    <Layout className={styles.layout}>
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        className={styles.sider}
        theme="light"
      >
        <div className={styles.logo}>
          <ShopOutlined className={styles.logoIcon} />
          {!collapsed && <span>商家后台</span>}
        </div>
        <Menu
          mode="inline"
          selectedKeys={[getSelectedKey()]}
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Sider>
      <Layout>
        <Header className={styles.header}>
          <div className={styles.headerRight}>
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <div className={styles.userInfo}>
                <Avatar icon={<UserOutlined />} />
                <span className={styles.userName}>{user?.name || '商家'}</span>
              </div>
            </Dropdown>
          </div>
        </Header>
        <Content className={styles.content}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}

export default MerchantLayout
