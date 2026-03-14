import { Outlet } from 'react-router-dom'
import { Layout as AntLayout } from 'antd'
import './Layout.css'

const { Header, Content, Footer } = AntLayout

export default function Layout() {
  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Header className="layout-header">
        <div className="layout-logo">拼脱脱</div>
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
