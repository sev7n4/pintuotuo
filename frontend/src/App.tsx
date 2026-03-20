import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import Layout from '@components/Layout'
import MerchantLayout from '@layouts/MerchantLayout'
import AdminLayout from '@layouts/AdminLayout'

// Pages
import LoginPage from '@pages/LoginPage'
import RegisterPage from '@pages/RegisterPage'
import HomePage from '@pages/HomePage'
import ProductListPage from '@pages/ProductListPage'
import ProductDetailPage from '@pages/ProductDetailPage'
import CartPage from '@pages/CartPage'
import CheckoutPage from '@pages/CheckoutPage'
import OrderListPage from '@pages/OrderListPage'
import PaymentPage from '@pages/PaymentPage'
import GroupListPage from '@pages/GroupListPage'
import ReferralPage from '@pages/ReferralPage'
import MyToken from '@pages/MyToken'
import Profile from '@pages/Profile'
import Consumption from '@pages/Consumption'

// Merchant Pages
import MerchantDashboard from '@pages/merchant/MerchantDashboard'
import MerchantProducts from '@pages/merchant/MerchantProducts'
import MerchantOrders from '@pages/merchant/MerchantOrders'
import MerchantSettings from '@pages/merchant/MerchantSettings'
import MerchantAPIKeys from '@pages/merchant/MerchantAPIKeys'
import MerchantSettlements from '@pages/merchant/MerchantSettlements'

// Admin Pages
import AdminDashboard from '@pages/admin/AdminDashboard'
import AdminUsers from '@pages/admin/AdminUsers'

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
            <Route index element={<HomePage />} />

            {/* Products */}
            <Route path="/products" element={<ProductListPage />} />
            <Route path="/products/:id" element={<ProductDetailPage />} />

            {/* Shopping */}
            <Route path="/cart" element={<CartPage />} />
            <Route path="/checkout" element={<CheckoutPage />} />

            {/* Orders */}
            <Route path="/orders" element={<OrderListPage />} />
            <Route path="/payment/:id" element={<PaymentPage />} />

            {/* Groups */}
            <Route path="/groups" element={<GroupListPage />} />

            {/* Referral */}
            <Route path="/referral" element={<ReferralPage />} />

            {/* My Token */}
            <Route path="/my-tokens" element={<MyToken />} />

            {/* Consumption */}
            <Route path="/consumption" element={<Consumption />} />

            {/* Profile */}
            <Route path="/profile" element={<Profile />} />

            {/* Catch all */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Route>

          {/* Merchant routes */}
          <Route path="/merchant" element={<MerchantLayout />}>
            <Route index element={<MerchantDashboard />} />
            <Route path="products" element={<MerchantProducts />} />
            <Route path="orders" element={<MerchantOrders />} />
            <Route path="settlements" element={<MerchantSettlements />} />
            <Route path="api-keys" element={<MerchantAPIKeys />} />
            <Route path="settings" element={<MerchantSettings />} />
          </Route>

          {/* Admin routes */}
          <Route path="/admin" element={<AdminLayout />}>
            <Route index element={<AdminDashboard />} />
            <Route path="users" element={<AdminUsers />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </ConfigProvider>
  )
}

export default App
