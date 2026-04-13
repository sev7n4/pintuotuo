import { AuthPage } from './AuthPage';

/** @deprecated 使用 AuthPage；保留默认导出以兼容路由与测试 */
const LoginPage = () => <AuthPage defaultMode="login" />;
export default LoginPage;
