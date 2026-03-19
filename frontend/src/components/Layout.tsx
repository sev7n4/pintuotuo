import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom'
import { Layout as AntLayout, Menu, Dropdown, Avatar, Space, message } from 'antd'
import { UserOutlined, GiftOutlined, WalletOutlined, LogoutOutlined, BarChartOutlined, ShopOutlined } from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'
import './Layout.css'

const { Header, Content, Footer } = AntLayout

export default function Layout() {
  const location = useLocation()
  const navigate = useNavigate()
  const { user, logout, isAuthenticated } = useAuthStore()

  const menuItems = [
    { key: '/', label: <Link to="/">首页</Link> },
    { key: '/products', label: <Link to="/products">商品</Link> },
    { key: '/orders', label: <Link to="/orders">订单</Link> },
    { key: '/groups', label: <Link to="/groups">拼团</Link> },
    { key: '/my-tokens', label: <Link to="/my-tokens">我的Token</Link>, icon: <WalletOutlined /> },
    { key: '/consumption', label: <Link to="/consumption">消费明细</Link>, icon: <BarChartOutlined /> },
    { key: '/referral', label: <Link to="/referral">邀请返利</Link>, icon: <GiftOutlined /> },
  ]

  const userMenuItems = [
    { key: 'profile', label: '个人中心' },
    { key: 'consumption', label: '消费明细' },
    { key: 'logout', label: '退出登录', icon: <LogoutOutlined /> },
  ]

  const handleUserMenuClick = async ({ key }: { key: string }) => {
    if (key === 'logout') {
      await logout()
      message.success('已退出登录')
      navigate('/login')
    } else if (key === 'profile') {
      navigate('/profile')
    } else if (key === 'consumption') {
      navigate('/consumption')
    }
  }

  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Header className="layout-header">
        <div className="layout-logo">拼脱脱</div>
        <Menu
          mode="horizontal"
          selectedKeys={[location.pathname]}
          items={menuItems}
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
                <span>{user.name || user.email}</span>
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
