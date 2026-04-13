import React, { useEffect, useState, useCallback, useRef, useMemo } from 'react';
import {
  Form,
  Input,
  Button,
  Card,
  message,
  Modal,
  Checkbox,
  Alert,
  Divider,
  Typography,
  Space,
  Segmented,
} from 'antd';
import { WechatOutlined, GithubOutlined } from '@ant-design/icons';
import { useNavigate, useLocation, Link } from 'react-router-dom';
import { useAuthStore } from '@stores/authStore';
import { AuthPhoneSection } from './AuthPhoneSection';
import {
  readPrimaryLoginPreference,
  writePrimaryLoginPreference,
} from '@/lib/authLoginPreference';
import styles from './AuthPage.module.css';

export type AuthPageProps = {
  /** 无路由上下文时（单测）用于区分登录/注册默认页签 */
  defaultMode?: 'login' | 'register';
};

type AuthCapabilities = {
  sms: boolean;
  email_magic: boolean;
  wechat_oauth: boolean;
  github_oauth: boolean;
  account_linking: boolean;
};

const defaultAuthCapabilities: AuthCapabilities = {
  sms: false,
  email_magic: false,
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
  const { login, register, sendEmailMagicLink, fetchUser, isLoading, user, isAuthenticated } =
    useAuthStore();
  const [loginForm] = Form.useForm();
  const [capabilities, setCapabilities] = useState<AuthCapabilities | null>(null);
  const [sendingMagic, setSendingMagic] = useState(false);
  /** 默认邮箱；成功登录后会写入 localStorage，下次进入优先展示上次方式 */
  const [primaryLogin, setPrimaryLogin] = useState<'email' | 'phone'>(() =>
    readPrimaryLoginPreference()
  );
  /** /register 邮箱注册时的账号类型（与旧版「买家 / 商户」Tab 对齐） */
  const [registerRole, setRegisterRole] = useState<'user' | 'merchant'>('user');
  const oauthRedirectHandled = useRef(false);

  /** URL 仍带 oauth=1&token= 时，避免「已登录自动跳转首页」与 OAuth 处理竞态，导致未清 query 就离开 /login */
  const oauthCallbackPending = useMemo(() => {
    const p = new URLSearchParams(location.search);
    return p.get('oauth') === '1' && Boolean(p.get('token'));
  }, [location.search]);

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
      writePrimaryLoginPreference('email');
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
  const isRegisterRoute = authTab === 'register';

  const onEmailPasswordSubmit = async (values: {
    email: string;
    password: string;
    rememberMe?: boolean;
  }) => {
    try {
      if (isRegisterRoute) {
        await register(values.email, values.password, registerRole);
        writePrimaryLoginPreference('email');
        message.success('注册成功');
      } else {
        await login(values.email, values.password, values.rememberMe || false);
        writePrimaryLoginPreference('email');
        message.success('登录成功');
      }
    } catch (err) {
      const errorMsg =
        err instanceof Error
          ? err.message
          : isRegisterRoute
            ? '注册失败，请检查邮箱与密码'
            : '登录失败，请检查邮箱和密码';
      message.error(errorMsg);
    }
  };

  const onSendMagicLink = async (values: { email: string }) => {
    if (!capabilities?.email_magic) {
      message.warning('当前未开启邮箱魔法链接登录');
      return;
    }
    if (!sendEmailMagicLink) {
      message.warning('当前版本未接入邮箱魔法链接接口');
      return;
    }
    setSendingMagic(true);
    try {
      const res = await sendEmailMagicLink(values.email.trim());
      if (res?.debug_link) {
        // Mock：未发真实邮件；Message 内文字不可点击，用弹窗给出可跳转入口
        message.success('当前为联调模式，未发送真实邮件');
        Modal.info({
          title: '完成登录',
          width: 520,
          okText: '立即验证并登录',
          onOk: () => {
            window.location.assign(res.debug_link!);
          },
          content: (
            <div>
              <Typography.Paragraph style={{ marginBottom: 8 }}>
                真实环境会发到邮箱；Mock 模式下请点下方按钮，浏览器将打开验证链接并跳转回本站完成登录。
              </Typography.Paragraph>
              <Typography.Paragraph copyable={{ text: res.debug_link }} style={{ marginBottom: 0 }}>
                {res.debug_link}
              </Typography.Paragraph>
            </div>
          ),
        });
      } else {
        message.success('登录链接已发送，请检查邮箱');
      }
    } catch (err) {
      message.error(err instanceof Error ? err.message : '发送失败，请稍后重试');
    } finally {
      setSendingMagic(false);
    }
  };

  useEffect(() => {
    if (!isAuthenticated || !user) return;
    if (oauthCallbackPending) return;
    if (user.role === 'admin') {
      navigate('/admin', { replace: true });
    } else if (user.role === 'merchant') {
      navigate('/merchant', { replace: true });
    } else {
      navigate('/', { replace: true });
    }
  }, [isAuthenticated, user, navigate, oauthCallbackPending]);

  const cardTitle = '拼脱脱 - 登录 / 注册';

  return (
    <div className={styles.authPage}>
      <Card className={`${styles.authCard} auth-card`} title={cardTitle}>
        {isRegisterRoute && (
          <Alert
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
            message="账号体系已升级"
            description="已切换为无感注册：手机号验证码或邮箱魔法链接首次通过即自动创建账号。"
          />
        )}

        <Segmented
          block
          value={primaryLogin}
          onChange={(v) => setPrimaryLogin(v as 'email' | 'phone')}
          options={[
            {
              label: isRegisterRoute ? '邮箱注册' : '邮箱登录',
              value: 'email',
            },
            {
              label: isRegisterRoute ? '手机注册' : '手机登录',
              value: 'phone',
            },
          ]}
          style={{ marginBottom: 16 }}
        />

        {isRegisterRoute && primaryLogin === 'email' && (
          <Segmented
            block
            value={registerRole}
            onChange={(v) => setRegisterRole(v as 'user' | 'merchant')}
            options={[
              { label: '个人用户', value: 'user' },
              { label: '商户入驻', value: 'merchant' },
            ]}
            style={{ marginBottom: 16 }}
          />
        )}

        {primaryLogin === 'email' ? (
          <>
            <Form
              form={loginForm}
              layout="vertical"
              onFinish={onEmailPasswordSubmit}
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
                label={
                  isRegisterRoute ? '密码' : '密码（仅曾用邮箱注册的账号）'
                }
                name="password"
                rules={
                  isRegisterRoute
                    ? [
                        { required: true, message: '请输入密码' },
                        { min: 6, message: '密码至少 6 位' },
                      ]
                    : [{ required: true, message: '请输入密码' }]
                }
              >
                <Input.Password
                  placeholder={isRegisterRoute ? '设置密码（至少 6 位）' : '输入密码'}
                />
              </Form.Item>
              {!isRegisterRoute && (
                <Form.Item name="rememberMe" valuePropName="checked">
                  <Checkbox>记住我</Checkbox>
                </Form.Item>
              )}
              <Form.Item style={{ marginBottom: 8 }}>
                <Button type="primary" htmlType="submit" block loading={isLoading}>
                  {isRegisterRoute ? '注册并进入' : '密码登录'}
                </Button>
              </Form.Item>
              {!isRegisterRoute && (
                <Form.Item style={{ marginBottom: 0 }}>
                  <Button
                    block
                    onClick={() => void loginForm.validateFields(['email']).then(onSendMagicLink)}
                    loading={sendingMagic}
                    disabled={!capabilities?.email_magic}
                  >
                    发送邮箱魔法链接
                  </Button>
                </Form.Item>
              )}
            </Form>
            <Typography.Paragraph
              type="secondary"
              style={{ marginTop: 10, marginBottom: 0, fontSize: 12 }}
            >
              邮箱魔法链接开启后，首次验证通过会自动创建账号并登录；未开启时可继续使用上方邮箱密码登录。手机登录请切换到「手机登录」。
            </Typography.Paragraph>
          </>
        ) : (
          <AuthPhoneSection
            smsEnabled={capabilities?.sms === true}
            capabilitiesLoaded={capabilities !== null}
            embedded
            onSmsLoginSuccess={() => writePrimaryLoginPreference('phone')}
          />
        )}

        {!isRegisterRoute ? (
          <Typography.Paragraph style={{ marginTop: 16, marginBottom: 0, textAlign: 'center' }}>
            <Link to="/register">创建新账户</Link>
          </Typography.Paragraph>
        ) : (
          <Typography.Paragraph style={{ marginTop: 16, marginBottom: 0, textAlign: 'center' }}>
            <Link to="/login">立即登录</Link>
          </Typography.Paragraph>
        )}

        <Divider plain>
          <Typography.Text type="secondary">更多方式（需服务端配置）</Typography.Text>
        </Divider>
        <Space direction="vertical" style={{ width: '100%' }} size="small">
          <Typography.Paragraph type="secondary" style={{ marginBottom: 8, fontSize: 12 }}>
            邮箱魔法链接、微信扫码、GitHub 与账号绑定由后端开关控制。手机号验证码请切换到「手机登录」（未开启时会说明如何配置）。
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
