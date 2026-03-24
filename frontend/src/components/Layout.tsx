import { useEffect } from 'react'
import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom'
import { Layout as AntLayout, Menu, Dropdown, Avatar, Space, message } from 'antd'
import { 
  UserOutlined, 
  LogoutOutlined, 
  HomeOutlined,
  AppstoreOutlined,
  ShoppingCartOutlined,
  HeartOutlined,
  HistoryOutlined,
  SettingOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'
import './Layout.css'

const { Header, Content, Footer } = AntLayout

export default function Layout() {
  const location = useLocation()
  const navigate = useNavigate()
  const { user, logout, isAuthenticated, fetchUser } = useAuthStore()

  useEffect(() => {
    if (isAuthenticated && !user) {
      fetchUser()
    }
  }, [isAuthenticated, user, fetchUser])

  const getSelectedTab = () => {
    const path = location.pathname
    if (path === '/') return 'home'
    if (path === '/categories' || path.startsWith('/products')) return 'category'
    if (path.startsWith('/orders') || path.startsWith('/groups') || path.startsWith('/payment') || path === '/my-tokens' || path === '/consumption') return 'orders'
    if (path === '/my' || path === '/profile' || path === '/referral' || path === '/favorites' || path === '/history') return 'my'
    return 'home'
  }

  const tabItems = [
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
      key: 'orders', 
      label: <Link to="/orders">订单</Link>,
      icon: <ShoppingCartOutlined />,
    },
    { 
      key: 'my', 
      label: <Link to="/my">我的</Link>,
      icon: <UserOutlined />,
    },
  ]

  const userMenuItems = [
    { key: 'profile', label: '个人中心', icon: <UserOutlined /> },
    { key: 'favorites', label: '我的收藏', icon: <HeartOutlined /> },
    { key: 'history', label: '浏览历史', icon: <HistoryOutlined /> },
    { key: 'settings', label: '账户设置', icon: <SettingOutlined /> },
    { type: 'divider' as const },
    { key: 'logout', label: '退出登录', icon: <LogoutOutlined /> },
  ]

  const handleUserMenuClick = async ({ key }: { key: string }) => {
    switch (key) {
      case 'logout':
        await logout()
        message.success('已退出登录')
        navigate('/login')
        break
      case 'profile':
        navigate('/profile')
        break
      case 'favorites':
        navigate('/favorites')
        break
      case 'history':
        navigate('/history')
        break
      case 'settings':
        navigate('/my')
        break
    }
  }

  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Header className="layout-header">
        <div className="layout-logo">拼脱脱</div>
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
              <Space 
                style={{ cursor: 'pointer' }} 
                data-testid="user-dropdown"
              >
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
  )
}
