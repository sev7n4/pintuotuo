import { BrowserRouter, Routes, Route, Navigate, useParams } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import Layout from '@components/Layout';
import MerchantLayout from '@layouts/MerchantLayout';
import AdminLayout from '@layouts/AdminLayout';

// Pages
import LoginPage from '@pages/LoginPage';
import RegisterPage from '@pages/RegisterPage';
import HomePage from '@pages/HomePage';
import ProductListPage from '@pages/ProductListPage';
import ProductDetailPage from '@pages/ProductDetailPage';
import CartPage from '@pages/CartPage';
import CheckoutPage from '@pages/CheckoutPage';
import OrderListPage from '@pages/OrderListPage';
import OrderDetailPage from '@pages/OrderDetailPage';
import PaymentPage from '@pages/PaymentPage';
import GroupListPage from '@pages/GroupListPage';
import GroupProgressPage from '@pages/GroupProgressPage';
import ReferralPage from '@pages/ReferralPage';
import MyToken from '@pages/MyToken';
import Profile from '@pages/Profile';
import Consumption from '@pages/Consumption';
import HelpCenterPage from '@pages/HelpCenterPage';
import AboutPage from '@pages/AboutPage';
import UserAgreementPage from '@pages/UserAgreementPage';
import PrivacyPolicyPage from '@pages/PrivacyPolicyPage';
import MyPage from '@pages/MyPage';
import MyServicesPage from '@pages/MyServicesPage';
import CategoryPage from '@pages/CategoryPage';
import FavoritesPage from '@pages/FavoritesPage';
import HistoryPage from '@pages/HistoryPage';
import EntitlementPackagesPage from '@pages/EntitlementPackagesPage';
import MyEntitlementsPage from '@pages/MyEntitlementsPage';
import FuelStationPage from '@pages/FuelStationPage';

// Merchant Pages
import MerchantDashboard from '@pages/merchant/MerchantDashboard';
import MerchantSKUs from '@pages/merchant/MerchantSKUs';
import MerchantOrders from '@pages/merchant/MerchantOrders';
import MerchantSettings from '@pages/merchant/MerchantSettings';
import MerchantAPIKeys from '@pages/merchant/MerchantAPIKeys';
import MerchantSettlements from '@pages/merchant/MerchantSettlements';
import MerchantAnalytics from '@pages/merchant/MerchantAnalytics';
import MerchantMarketing from '@pages/merchant/MerchantMarketing';
import MerchantBills from '@pages/merchant/MerchantBills';
import MerchantInvoices from '@pages/merchant/MerchantInvoices';

// Admin Pages
import AdminDashboard from '@pages/admin/AdminDashboard';
import AdminUsers from '@pages/admin/AdminUsers';
import AdminMerchants from '@pages/admin/AdminMerchants';
import AdminProducts from '@pages/admin/AdminProducts';
import AdminOrders from '@pages/admin/AdminOrders';
import AdminSettings from '@pages/admin/AdminSettings';
import AdminSPUs from '@pages/admin/AdminSPUs';
import AdminSKUs from '@pages/admin/AdminSKUs';
import AdminModelProviders from '@pages/admin/AdminModelProviders';
import AdminModelFallback from '@pages/admin/AdminModelFallback';
import AdminFlashSales from '@pages/admin/AdminFlashSales';
import AdminFuelStation from '@pages/admin/AdminFuelStation';
import AdminBillings from '@pages/admin/AdminBillings';
import AdminUserBillings from '@pages/admin/AdminUserBillings';
import AdminSettlements from '@pages/admin/AdminSettlements';
import AdminRoutingStrategies from '@pages/admin/AdminRoutingStrategies';
import AdminReconciliation from '@pages/admin/AdminReconciliation';
import AdminEntitlementPackages from '@pages/admin/AdminEntitlementPackages';
import AdminAPIKeyStatus from '@pages/admin/AdminAPIKeyStatus';
import AdminEndpointPricing from '@pages/admin/AdminEndpointPricing';
import AdminResponseStorage from '@pages/admin/AdminResponseStorage';
import AdminRouteDecisionLogs from '@pages/admin/AdminRouteDecisionLogs';
import AdminGatewayStats from '@pages/admin/AdminGatewayStats';
import AdminRateLimiter from '@pages/admin/AdminRateLimiter';
import AdminQueueConfig from '@pages/admin/AdminQueueConfig';
import AdminByokRouting from '@pages/admin/AdminByokRouting';

function LegacyProductDetailRedirect() {
  const { id } = useParams();
  return <Navigate to={`/catalog/${id ?? ''}`} replace />;
}

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

            {/* Catalog (卖场 SKU 列表/详情)；旧路径 /products 重定向 */}
            <Route path="/products" element={<Navigate to="/catalog" replace />} />
            <Route path="/products/:id" element={<LegacyProductDetailRedirect />} />
            <Route path="/catalog" element={<ProductListPage />} />
            <Route path="/catalog/:id" element={<ProductDetailPage />} />
            <Route path="/packages" element={<EntitlementPackagesPage />} />
            <Route path="/fuel-station" element={<FuelStationPage />} />

            {/* Categories */}
            <Route path="/categories" element={<CategoryPage />} />

            {/* Shopping */}
            <Route path="/cart" element={<CartPage />} />
            <Route path="/checkout" element={<CheckoutPage />} />

            {/* Orders */}
            <Route path="/orders" element={<OrderListPage />} />
            <Route path="/orders/:id" element={<OrderDetailPage />} />
            <Route path="/payment/:id" element={<PaymentPage />} />

            {/* Groups */}
            <Route path="/groups" element={<GroupListPage />} />
            <Route path="/groups/:id" element={<GroupProgressPage />} />

            {/* Referral */}
            <Route path="/referral" element={<ReferralPage />} />

            {/* My Token */}
            <Route path="/my-tokens" element={<MyToken />} />

            {/* Consumption */}
            <Route path="/consumption" element={<Consumption />} />

            {/* Profile */}
            <Route path="/profile" element={<Profile />} />

            {/* My Page（个人主页）；我的服务为同级独立页 */}
            <Route path="/my" element={<MyPage />} />
            <Route path="/my/services" element={<MyServicesPage />} />
            <Route path="/my/entitlements" element={<MyEntitlementsPage />} />

            {/* Favorites */}
            <Route path="/favorites" element={<FavoritesPage />} />

            {/* Browse History */}
            <Route path="/history" element={<HistoryPage />} />

            {/* Static Pages */}
            <Route path="/help" element={<HelpCenterPage />} />
            <Route path="/about" element={<AboutPage />} />
            <Route path="/agreement" element={<UserAgreementPage />} />
            <Route path="/privacy" element={<PrivacyPolicyPage />} />

            {/* Catch all */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Route>

          {/* Merchant routes */}
          <Route path="/merchant" element={<MerchantLayout />}>
            <Route index element={<MerchantDashboard />} />
            <Route path="products" element={<Navigate to="/merchant/skus" replace />} />
            <Route path="skus" element={<MerchantSKUs />} />
            <Route path="orders" element={<MerchantOrders />} />
            <Route path="settlements" element={<MerchantSettlements />} />
            <Route path="bills" element={<MerchantBills />} />
            <Route path="invoices" element={<MerchantInvoices />} />
            <Route path="analytics" element={<MerchantAnalytics />} />
            <Route path="marketing" element={<MerchantMarketing />} />
            <Route path="api-keys" element={<MerchantAPIKeys />} />
            <Route path="settings" element={<MerchantSettings />} />
          </Route>

          {/* Admin routes */}
          <Route path="/admin" element={<AdminLayout />}>
            <Route index element={<AdminDashboard />} />
            <Route path="users" element={<AdminUsers />} />
            <Route path="merchants" element={<AdminMerchants />} />
            <Route path="products" element={<AdminProducts />} />
            <Route path="orders" element={<AdminOrders />} />
            <Route path="settlements" element={<AdminSettlements />} />
            <Route path="billings" element={<AdminBillings />} />
            <Route path="user-billings" element={<AdminUserBillings />} />
            <Route path="settings" element={<AdminSettings />} />
            <Route path="spus" element={<AdminSPUs />} />
            <Route path="skus" element={<AdminSKUs />} />
            <Route path="model-providers" element={<AdminModelProviders />} />
            <Route path="model-fallback" element={<AdminModelFallback />} />
            <Route path="routing-strategies" element={<AdminRoutingStrategies />} />
            <Route path="api-key-status" element={<AdminAPIKeyStatus />} />
            <Route path="route-decision-logs" element={<AdminRouteDecisionLogs />} />
            <Route path="gateway-stats" element={<AdminGatewayStats />} />
            <Route path="rate-limiter" element={<AdminRateLimiter />} />
            <Route path="queue-config" element={<AdminQueueConfig />} />
            <Route path="byok-routing" element={<AdminByokRouting />} />
            <Route path="reconciliation" element={<AdminReconciliation />} />
            <Route path="flash-sales" element={<AdminFlashSales />} />
            <Route path="fuel-station" element={<AdminFuelStation />} />
            <Route path="entitlement-packages" element={<AdminEntitlementPackages />} />
            <Route path="endpoint-pricing" element={<AdminEndpointPricing />} />
            <Route path="response-storage" element={<AdminResponseStorage />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </ConfigProvider>
  );
}

export default App;
