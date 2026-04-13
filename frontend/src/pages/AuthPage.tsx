import React, { useEffect, useState, useCallback, useRef } from 'react';
import {
  Form,
  Input,
  Button,
  Card,
  message,
  Checkbox,
  Tabs,
  Alert,
  Divider,
  Typography,
  Space,
} from 'antd';
import { ShopOutlined, WechatOutlined, GithubOutlined } from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '@stores/authStore';
import { AuthPhoneSection } from './AuthPhoneSection';
import styles from './AuthPage.module.css';

type UserRole = 'user' | 'merchant';

export type AuthPageProps = {
  /** 无路由上下文时（单测）用于区分登录/注册默认页签 */
  defaultMode?: 'login' | 'register';
};

type AuthCapabilities = {
  sms: boolean;
  wechat_oauth: boolean;
  github_oauth: boolean;
  account_linking: boolean;
};

const defaultAuthCapabilities: AuthCapabilities = {
  sms: false,
  wechat_oauth: false,
  github_oauth: false,
  account_linking: false,
};

/**
 * 登录与注册合一页：注册成功后由 authStore 写入 token，即「注册即登录」。
 * /login、/register 仍保留，便于书签与 E2E。
 */
export const AuthPage: React.FC<AuthPageProps> = ({ defaultMode = 'login' }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { login, register, fetchUser, isLoading, error, user, isAuthenticated } =
    useAuthStore();
  const [loginForm] = Form.useForm();
  const [registerForm] = Form.useForm();
  const [accountTab, setAccountTab] = useState<'buyer' | 'merchant'>('buyer');
  const [capabilities, setCapabilities] = useState<AuthCapabilities | null>(null);
  const oauthRedirectHandled = useRef(false);

  const loadCapabilities = useCallback(() => {
    fetch('/api/v1/users/auth/capabilities')
      .then((r) => (r.ok ? r.json() : null))
      .then((data) => {
        if (data && typeof data.sms === 'boolean') {
          setCapabilities(data);
        } else {
          setCapabilities(defaultAuthCapabilities);
        }
      })
      .catch(() => {
        setCapabilities(defaultAuthCapabilities);
      });
  }, []);

  useEffect(() => {
    loadCapabilities();
  }, [loadCapabilities]);

  /** OAuth 回调：后端重定向到 /login?oauth=1&token=... 或 ?oauth_error=... */
  useEffect(() => {
    const params = new URLSearchParams(location.search);
    const oauthFlag = params.get('oauth');
    const tok = params.get('token');
    const oauthErr = params.get('oauth_error');

    if (oauthErr) {
      message.error(decodeURIComponent(oauthErr.replace(/\+/g, ' ')));
      navigate({ pathname: location.pathname, search: '' }, { replace: true });
      return;
    }
    if (oauthFlag === '1' && tok) {
      if (oauthRedirectHandled.current) return;
      oauthRedirectHandled.current = true;
      localStorage.setItem('auth_token', tok);
      void fetchUser()
        .then(() => {
          message.success('登录成功');
          navigate({ pathname: location.pathname, search: '' }, { replace: true });
        })
        .catch(() => {
          oauthRedirectHandled.current = false;
          message.error('第三方登录失败，请重试');
          navigate({ pathname: location.pathname, search: '' }, { replace: true });
        });
    }
  }, [location.search, location.pathname, fetchUser, navigate]);

  const path = location.pathname;
  const authTab: 'login' | 'register' =
    path === '/register' || (path === '/' && defaultMode === 'register') ? 'register' : 'login';

  const onLoginFinish = async (values: {
    email: string;
    password: string;
    rememberMe?: boolean;
  }) => {
    try {
      await login(values.email, values.password, values.rememberMe || false);
      message.success('登录成功');
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : '登录失败，请检查邮箱和密码';
      message.error(errorMsg);
    }
  };

  const onRegisterFinish = async (values: {
    email: string;
    password: string;
    confirmPassword: string;
  }) => {
    if (values.password !== values.confirmPassword) {
      message.error('两次输入的密码不一致');
      return;
    }
    const role: UserRole = accountTab === 'merchant' ? 'merchant' : 'user';
    try {
      await register(values.email, values.password, role);
      message.success('注册成功，已自动登录');
    } catch (err) {
      message.error(error || '注册失败，请稍后重试');
    }
  };

  useEffect(() => {
    if (!isAuthenticated || !user) return;
    if (user.role === 'admin') {
      navigate('/admin', { replace: true });
    } else if (user.role === 'merchant') {
      navigate('/merchant', { replace: true });
    } else {
      navigate('/', { replace: true });
    }
  }, [isAuthenticated, user, navigate]);

  const cardTitle =
    authTab === 'login' ? '拼脱脱 - 登录' : '拼脱脱 - 注册';

  return (
    <div className={styles.authPage}>
      <Card className={`${styles.authCard} auth-card`} title={cardTitle}>
        <Tabs
          activeKey={authTab}
          onChange={(key) => {
            navigate(key === 'register' ? '/register' : '/login');
          }}
          items={[
            { key: 'login', label: '登录' },
            { key: 'register', label: '注册' },
          ]}
          style={{ marginBottom: 16 }}
        />

        {authTab === 'login' ? (
          <Form
            form={loginForm}
            layout="vertical"
            onFinish={onLoginFinish}
            autoComplete="off"
            initialValues={{ rememberMe: true }}
          >
            <Form.Item
              label="邮箱"
              name="email"
              rules={[
                { required: true, message: '请输入邮箱' },
                { type: 'email', message: '邮箱格式不正确' },
              ]}
            >
              <Input placeholder="example@email.com" />
            </Form.Item>
            <Form.Item
              label="密码"
              name="password"
              rules={[{ required: true, message: '请输入密码' }]}
            >
              <Input.Password placeholder="输入密码" />
            </Form.Item>
            <Form.Item name="rememberMe" valuePropName="checked">
              <Checkbox>记住我</Checkbox>
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" block loading={isLoading}>
                登录
              </Button>
            </Form.Item>
            <div style={{ textAlign: 'center' }}>
              <span>没有账户？ </span>
              <Button type="link" onClick={() => navigate('/register')}>
                创建新账户
              </Button>
            </div>
          </Form>
        ) : (
          <>
            <Tabs
              activeKey={accountTab}
              onChange={(k) => setAccountTab(k as 'buyer' | 'merchant')}
              items={[
                { key: 'buyer', label: '买家注册' },
                {
                  key: 'merchant',
                  label: (
                    <span>
                      <ShopOutlined /> 商户入驻
                    </span>
                  ),
                },
              ]}
              style={{ marginBottom: 16 }}
            />
            {accountTab === 'merchant' && (
              <Alert
                type="info"
                showIcon
                message="商户入驻"
                description="用于上架 SKU、管理订单与结算。若仅购买 Token，请使用「买家注册」。"
                style={{ marginBottom: 16 }}
              />
            )}
            <Form
              form={registerForm}
              layout="vertical"
              onFinish={onRegisterFinish}
              autoComplete="off"
            >
              <Form.Item
                label="邮箱"
                name="email"
                rules={[
                  { required: true, message: '请输入邮箱' },
                  { type: 'email', message: '邮箱格式不正确' },
                ]}
              >
                <Input placeholder="example@email.com" />
              </Form.Item>
              <Form.Item
                label="密码"
                name="password"
                rules={[
                  { required: true, message: '请输入密码' },
                  { min: 6, message: '密码至少6个字符' },
                ]}
              >
                <Input.Password placeholder="设置密码" />
              </Form.Item>
              <Form.Item
                label="确认密码"
                name="confirmPassword"
                rules={[{ required: true, message: '请确认密码' }]}
              >
                <Input.Password placeholder="再次输入密码" />
              </Form.Item>
              <Form.Item>
                <Button type="primary" htmlType="submit" block loading={isLoading}>
                  创建账户
                </Button>
              </Form.Item>
              <div style={{ textAlign: 'center' }}>
                <span>已有账户？ </span>
                <Button type="link" onClick={() => navigate('/login')}>
                  立即登录
                </Button>
              </div>
            </Form>
          </>
        )}

        <AuthPhoneSection
          smsEnabled={capabilities?.sms === true}
          capabilitiesLoaded={capabilities !== null}
        />

        <Divider plain>
          <Typography.Text type="secondary">更多方式（需服务端配置）</Typography.Text>
        </Divider>
        <Space direction="vertical" style={{ width: '100%' }} size="small">
          <Typography.Paragraph type="secondary" style={{ marginBottom: 8, fontSize: 12 }}>
            微信扫码、GitHub 与账号绑定由后端开关控制。手机号验证码见上方「手机号」区域（未开启时会说明如何配置）。
          </Typography.Paragraph>
          <Space wrap>
            <Button
              icon={<WechatOutlined />}
              disabled={!capabilities?.wechat_oauth}
              onClick={() => {
                if (capabilities?.wechat_oauth) window.location.assign('/api/v1/users/oauth/wechat/start');
              }}
              title={capabilities?.wechat_oauth ? '微信 OAuth' : '未配置 WECHAT_OPEN_APP_ID'}
            >
              微信登录
            </Button>
            <Button
              icon={<GithubOutlined />}
              disabled={!capabilities?.github_oauth}
              onClick={() => {
                if (capabilities?.github_oauth) window.location.assign('/api/v1/users/oauth/github/start');
              }}
              title={capabilities?.github_oauth ? 'GitHub OAuth' : '未配置 GITHUB_OAUTH_CLIENT_ID'}
            >
              GitHub
            </Button>
          </Space>
          {capabilities?.account_linking && (
            <Typography.Paragraph type="secondary" style={{ fontSize: 12, marginTop: 8 }}>
              账号绑定（邮箱 / 微信 / GitHub）已在服务端开启，请前往「个人资料」关联。
            </Typography.Paragraph>
          )}
        </Space>
      </Card>
    </div>
  );
};

export default AuthPage;
