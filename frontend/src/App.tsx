import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import Layout from '@components/Layout'

// Pages
import LoginPage from '@pages/LoginPage'
import RegisterPage from '@pages/RegisterPage'
import ProductListPage from '@pages/ProductListPage'
import ProductDetailPage from '@pages/ProductDetailPage'
import CartPage from '@pages/CartPage'
import OrderListPage from '@pages/OrderListPage'
import GroupListPage from '@pages/GroupListPage'

function App() {
  return (
    <ConfigProvider locale={zhCN}>
      <BrowserRouter>
        <Routes>
          {/* Auth routes */}
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />

          {/* App routes */}
          <Route element={<Layout />}>
            <Route index element={<Navigate to="/products" replace />} />

            {/* Products */}
            <Route path="/products" element={<ProductListPage />} />
            <Route path="/products/:id" element={<ProductDetailPage />} />

            {/* Shopping */}
            <Route path="/cart" element={<CartPage />} />

            {/* Orders */}
            <Route path="/orders" element={<OrderListPage />} />

            {/* Groups */}
            <Route path="/groups" element={<GroupListPage />} />

            {/* Catch all */}
            <Route path="*" element={<Navigate to="/products" replace />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </ConfigProvider>
  )
}

export default App
