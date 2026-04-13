import { useState } from 'react';
import {
  Form,
  Input,
  Button,
  Card,
  Tabs,
  message,
  Space,
  Typography,
  Grid,
  Spin,
  Alert,
} from 'antd';
import { MobileOutlined } from '@ant-design/icons';
import { useAuthStore } from '@/stores/authStore';
import styles from './AuthPhoneSection.module.css';

const { useBreakpoint } = Grid;

const phoneRules = [
  { required: true, message: '请输入手机号' },
  { pattern: /^1[3-9]\d{9}$/, message: '请输入 11 位中国大陆手机号' },
];

const codeRules = [{ required: true, message: '请输入验证码' }];

function SmsCodeField({
  stack,
  sending,
  onSendCode,
}: {
  stack: boolean;
  sending: boolean;
  onSendCode: () => void;
}) {
  if (stack) {
    return (
      <div className={styles.codeRowStack}>
        <Form.Item name="code" rules={codeRules}>
          <Input placeholder="验证码" maxLength={6} />
        </Form.Item>
        <Button loading={sending} block onClick={onSendCode}>
          获取验证码
        </Button>
      </div>
    );
  }
  return (
    <Space.Compact style={{ width: '100%', marginBottom: 12 }}>
      <Form.Item name="code" noStyle rules={codeRules}>
        <Input placeholder="验证码" maxLength={6} style={{ width: '100%' }} />
      </Form.Item>
      <Button loading={sending} onClick={onSendCode}>
        获取验证码
      </Button>
    </Space.Compact>
  );
}

export type AuthPhoneSectionProps = {
  /** 与 GET /users/auth/capabilities 的 sms 一致 */
  smsEnabled: boolean;
  /** 已完成 capabilities 请求（含失败时的降级） */
  capabilitiesLoaded: boolean;
};

/**
 * 手机号验证码登录 / 注册（需后端 MOCK_SMS 或 SMS_PROVIDER）
 * 未开启时仍展示说明，避免用户误以为「没有手机登录入口」。
 */
export function AuthPhoneSection({ smsEnabled, capabilitiesLoaded }: AuthPhoneSectionProps) {
  const { sendSmsCode, loginWithSms, registerWithSms, isLoading } = useAuthStore();
  const [loginForm] = Form.useForm();
  const [regForm] = Form.useForm();
  const [sending, setSending] = useState(false);
  const screens = useBreakpoint();
  /** 与 Layout 一致：未达 md 断点时验证码行纵向排列，避免按钮挤压 */
  const stackCodeRow = screens.md === false;

  const handleSend = async (phone: string) => {
    if (!smsEnabled) return;
    if (!/^1[3-9]\d{9}$/.test(phone)) {
      message.warning('请先填写有效手机号');
      return;
    }
    setSending(true);
    try {
      const res = await sendSmsCode(phone);
      message.success('验证码已发送');
      if (res?.debug_code) {
        message.info(`[开发] 验证码：${res.debug_code}`);
      }
    } catch {
      message.error('发送失败，请稍后重试');
    } finally {
      setSending(false);
    }
  };

  if (!capabilitiesLoaded) {
    return (
      <Card size="small" title={<><MobileOutlined /> 手机号</>} style={{ marginTop: 16 }}>
        <div style={{ textAlign: 'center', padding: '16px 0' }}>
          <Spin tip="加载登录方式…" />
        </div>
      </Card>
    );
  }

  if (!smsEnabled) {
    return (
      <Card size="small" title={<><MobileOutlined /> 手机号</>} style={{ marginTop: 16 }}>
        <Alert
          type="info"
          showIcon
          message="暂未开启手机号验证码"
          description={
            <Typography.Paragraph style={{ marginBottom: 0, fontSize: 13 }}>
              服务端需在环境变量中开启短信能力后重启后端，本区域才会出现「验证码登录 / 手机号注册」：
              <ul style={{ margin: '8px 0 0', paddingLeft: 18 }}>
                <li>
                  开发/联调：设置 <Typography.Text code>MOCK_SMS=true</Typography.Text>，验证码多为{' '}
                  <Typography.Text code>123456</Typography.Text>
                </li>
                <li>
                  生产：配置真实短信网关对应的 <Typography.Text code>SMS_PROVIDER</Typography.Text> 等变量
                </li>
              </ul>
            </Typography.Paragraph>
          }
        />
      </Card>
    );
  }

  return (
    <Card size="small" title={<><MobileOutlined /> 手机号</>} style={{ marginTop: 16 }}>
      <Tabs
        items={[
          {
            key: 'plogin',
            label: '验证码登录',
            children: (
              <Form form={loginForm} layout="vertical" onFinish={onSmsLogin} size="small">
                <Form.Item label="手机号" name="phone" rules={phoneRules}>
                  <Input placeholder="11 位手机号" maxLength={11} />
                </Form.Item>
                <SmsCodeField
                  stack={stackCodeRow}
                  sending={sending}
                  onSendCode={() => {
                    const p = loginForm.getFieldValue('phone');
                    void handleSend(p);
                  }}
                />
                <Form.Item style={{ marginBottom: 0 }}>
                  <Button type="primary" htmlType="submit" block loading={isLoading}>
                    验证码登录
                  </Button>
                </Form.Item>
              </Form>
            ),
          },
          {
            key: 'preg',
            label: '手机号注册',
            children: (
              <Form form={regForm} layout="vertical" onFinish={onSmsRegister} size="small">
                <Form.Item label="手机号" name="phone" rules={phoneRules}>
                  <Input placeholder="11 位手机号" maxLength={11} />
                </Form.Item>
                <SmsCodeField
                  stack={stackCodeRow}
                  sending={sending}
                  onSendCode={() => {
                    const p = regForm.getFieldValue('phone');
                    void handleSend(p);
                  }}
                />
                <Form.Item
                  label="密码"
                  name="password"
                  rules={[{ required: true, min: 6, message: '至少 6 位密码' }]}
                >
                  <Input.Password placeholder="设置密码" />
                </Form.Item>
                <Form.Item
                  label="确认密码"
                  name="confirmPassword"
                  rules={[{ required: true, message: '请确认密码' }]}
                >
                  <Input.Password placeholder="再次输入" />
                </Form.Item>
                <Form.Item style={{ marginBottom: 0 }}>
                  <Button type="primary" htmlType="submit" block loading={isLoading}>
                    注册并登录
                  </Button>
                </Form.Item>
              </Form>
            ),
          },
        ]}
      />
      <Typography.Paragraph type="secondary" style={{ fontSize: 12, marginBottom: 0, marginTop: 8 }}>
        开发环境可设置 <code>MOCK_SMS=true</code>，验证码固定为 <code>123456</code>。
      </Typography.Paragraph>
    </Card>
  );

  async function onSmsLogin(values: { phone: string; code: string }) {
    try {
      await loginWithSms(values.phone.trim(), values.code.trim());
      message.success('登录成功');
    } catch (e) {
      message.error(e instanceof Error ? e.message : '登录失败');
    }
  }

  async function onSmsRegister(values: {
    phone: string;
    code: string;
    password: string;
    confirmPassword: string;
  }) {
    if (values.password !== values.confirmPassword) {
      message.error('两次密码不一致');
      return;
    }
    try {
      await registerWithSms(values.phone.trim(), values.code.trim(), values.password, 'user');
      message.success('注册成功，已自动登录');
    } catch (e) {
      message.error(e instanceof Error ? e.message : '注册失败');
    }
  }
}
