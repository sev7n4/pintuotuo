import { useEffect, useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Layout, Menu, Avatar, Dropdown, message, Spin } from 'antd'
import {
  DashboardOutlined,
  UserOutlined,
  ShopOutlined,
  ShoppingCartOutlined,
  SettingOutlined,
  LogoutOutlined,
  AppstoreOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'
import styles from './MerchantLayout.module.css'

const { Header, Sider, Content } = Layout

const menuItems = [
  {
    key: '/admin',
    icon: <DashboardOutlined />,
    label: '数据概览',
  },
  {
    key: '/admin/users',
    icon: <UserOutlined />,
    label: '用户管理',
  },
  {
    key: '/admin/merchants',
    icon: <ShopOutlined />,
    label: '商户管理',
  },
  {
    key: '/admin/products',
    icon: <AppstoreOutlined />,
    label: '商品管理',
  },
  {
    key: '/admin/orders',
    icon: <ShoppingCartOutlined />,
    label: '订单管理',
  },
  {
    key: '/admin/settings',
    icon: <SettingOutlined />,
    label: '系统设置',
  },
]

const AdminLayout = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout, isAuthenticated, fetchUser } = useAuthStore()
  const [collapsed, setCollapsed] = useState(false)
  const [checkingAuth, setCheckingAuth] = useState(true)

  useEffect(() => {
    const checkAuth = async () => {
      const hasToken = !!(localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token'))
      if (!hasToken) {
        navigate('/login', { state: { from: location.pathname } })
        return
      }

      if (!user) {
        try {
          await fetchUser()
        } catch (error) {
          localStorage.removeItem('auth_token')
          sessionStorage.removeItem('auth_token')
          navigate('/login', { state: { from: location.pathname } })
          return
        }
      }

      setCheckingAuth(false)
    }

    checkAuth()
  }, [isAuthenticated, user, fetchUser, navigate, location.pathname])

  useEffect(() => {
    if (!checkingAuth && user && user.role !== 'admin') {
      message.error('无权限访问管理后台')
      navigate('/')
    }
  }, [checkingAuth, user, navigate])

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
      key: 'home',
      icon: <UserOutlined />,
      label: '返回前台',
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
    if (path === '/admin') return '/admin'
    return path
  }

  if (checkingAuth) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <Spin size="large" tip="加载中..." />
      </div>
    )
  }

  if (!isAuthenticated || (user && user.role !== 'admin')) {
    return null
  }

  return (
    <Layout className={styles.layout}>
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        className={styles.sider}
        theme="dark"
      >
        <div className={styles.logo}>
          <DashboardOutlined className={styles.logoIcon} />
          {!collapsed && <span>运营管理</span>}
        </div>
        <Menu
          mode="inline"
          selectedKeys={[getSelectedKey()]}
          items={menuItems}
          onClick={handleMenuClick}
          theme="dark"
        />
      </Sider>
      <Layout>
        <Header className={styles.header}>
          <div className={styles.headerRight}>
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <div className={styles.userInfo}>
                <Avatar icon={<UserOutlined />} />
                <span className={styles.userName}>{user?.name || '管理员'}</span>
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

export default AdminLayout
