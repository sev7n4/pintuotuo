import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import Layout from '@components/Layout'

function App() {
  return (
    <ConfigProvider locale={zhCN}>
      <BrowserRouter>
        <Routes>
          <Route element={<Layout />}>
            {/* Routes will be added here */}
            <Route index element={<Navigate to="/products" replace />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </ConfigProvider>
  )
}

export default App
