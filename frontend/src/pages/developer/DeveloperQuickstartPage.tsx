import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Card, Checkbox, Space, Steps, Typography, message } from 'antd';
import { CopyOutlined, KeyOutlined, LinkOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';
import { tokenService } from '@/services/token';
import type { APIUsageGuideResponse } from '@/types';
import { getOpenAICompatBaseURL } from '@/utils/openaiCompat';
import { useTokenStore } from '@/stores/tokenStore';
import { EntitlementModelVerifyCard } from '@/components/entitlement/EntitlementModelVerifyCard';
import { trackDevCenter } from '@/utils/devCenterAnalytics';
import { copyToClipboard } from '@/utils/clipboard';
import { CreatePlatformAPIKeyModal } from '@/components/developer/CreatePlatformAPIKeyModal';

const { Title, Paragraph, Text } = Typography;

const LS_QUICKSTART_DONE = 'dev_center_quickstart_done';

export default function DeveloperQuickstartPage() {
  const openaiBase = useMemo(() => getOpenAICompatBaseURL(), []);
  const [usageGuide, setUsageGuide] = useState<APIUsageGuideResponse | null>(null);
  const [usageGuideLoading, setUsageGuideLoading] = useState(true);
  const [doneLocal, setDoneLocal] = useState(
    () => localStorage.getItem(LS_QUICKSTART_DONE) === '1'
  );
  const [createKeyModalOpen, setCreateKeyModalOpen] = useState(false);

  const { balance, fetchBalance, fetchAPIKeys, isLoading } = useTokenStore();

  useEffect(() => {
    fetchBalance();
    fetchAPIKeys();
  }, [fetchBalance, fetchAPIKeys]);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setUsageGuideLoading(true);
      try {
        const res = await tokenService.getAPIUsageGuide();
        const payload = (res.data as { data?: APIUsageGuideResponse })?.data;
        if (!cancelled) setUsageGuide(payload ?? null);
      } catch {
        if (!cancelled) setUsageGuide(null);
      } finally {
        if (!cancelled) setUsageGuideLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const copy = async (text: string, label: string) => {
    const ok = await copyToClipboard(text);
    if (ok) message.success(`已复制${label}`);
    else message.error('复制失败，请长按或手动选择文本复制');
  };

  const tokenBalance = balance?.balance ?? null;
  const balanceLoading = balance === null && isLoading;

  const onVerifySuccess = () => {
    trackDevCenter('quickstart_complete', { path: '/developer/quickstart' });
    if (!doneLocal) {
      localStorage.setItem(LS_QUICKSTART_DONE, '1');
      setDoneLocal(true);
    }
  };

  const exampleKeyPlaceholder = 'YOUR_PTD_KEY';
  const curlExample = `curl -sS "${openaiBase}/chat/completions" \\
  -H "Authorization: Bearer ${exampleKeyPlaceholder}" \\
  -H "Content-Type: application/json" \\
  -d '{"model":"${usageGuide?.default_model_example || 'provider/model'}","messages":[{"role":"user","content":"hi"}],"max_tokens":32}'`;

  const openCreateKeyModal = () => {
    trackDevCenter('quickstart_cta_create_key', { path: '/developer/quickstart' });
    setCreateKeyModalOpen(true);
  };

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <div>
        <Title level={3} style={{ marginTop: 0 }}>
          快速开始
        </Title>
        <Paragraph type="secondary">
          按步骤完成平台密钥（<Text code>ptd_</Text> 前缀）、Base URL 与一次试调用。与 OpenAI SDK
          兼容：将 <Text code>baseURL</Text> 指向下方地址即可。
        </Paragraph>
      </div>

      <Card
        styles={{ body: { padding: '20px 24px' } }}
        style={{
          background:
            'linear-gradient(135deg, rgba(22, 119, 255, 0.1) 0%, rgba(114, 46, 209, 0.07) 50%, rgba(22, 119, 255, 0.05) 100%)',
          borderColor: 'rgba(22, 119, 255, 0.35)',
        }}
      >
        <Space
          direction="vertical"
          size="middle"
          style={{ width: '100%' }}
          styles={{ item: { width: '100%' } }}
        >
          <div>
            <Title level={4} style={{ marginTop: 0, marginBottom: 8 }}>
              还没有平台密钥？
            </Title>
            <Paragraph type="secondary" style={{ marginBottom: 0 }}>
              建议<strong>先创建</strong> <Text code>ptd_</Text> Bearer Token，再复制 Base URL 与
              curl，避免粘了命令才发现没有 Key。
            </Paragraph>
          </div>
          <Space wrap size="middle">
            <Button type="primary" size="large" icon={<KeyOutlined />} onClick={openCreateKeyModal}>
              创建平台密钥（ptd_）
            </Button>
            <Link to="/developer/keys">
              <Button size="large">管理全部密钥</Button>
            </Link>
          </Space>
        </Space>
      </Card>

      <CreatePlatformAPIKeyModal
        open={createKeyModalOpen}
        onClose={() => setCreateKeyModalOpen(false)}
        onSuccess={() => fetchAPIKeys()}
      />

      <Steps
        direction="vertical"
        current={-1}
        items={[
          {
            title: '平台 API Key',
            description: (
              <Card size="small" styles={{ body: { padding: 12 } }}>
                <Paragraph style={{ marginBottom: 8 }}>
                  将完整密钥填入请求头 <Text code>Authorization: Bearer &lt;ptd_…&gt;</Text>。
                </Paragraph>
                <Space wrap>
                  <Button type="primary" icon={<KeyOutlined />} onClick={openCreateKeyModal}>
                    创建平台密钥
                  </Button>
                  <Link to="/developer/keys">
                    <Button icon={<LinkOutlined />}>密钥与安全（列表 / 删除）</Button>
                  </Link>
                </Space>
              </Card>
            ),
          },
          {
            title: 'Base URL',
            description: (
              <Card size="small" styles={{ body: { padding: 12 } }}>
                <Space wrap>
                  <Text code>{openaiBase}</Text>
                  <Button
                    type="primary"
                    size="small"
                    icon={<CopyOutlined />}
                    onClick={() => copy(openaiBase, ' Base URL')}
                  >
                    复制
                  </Button>
                </Space>
                <Paragraph type="secondary" style={{ marginTop: 8, marginBottom: 0 }}>
                  请求路径：<Text code>POST …/chat/completions</Text>。支持{' '}
                  <Text code>stream: true</Text>（OpenAI 兼容厂商路径；Anthropic
                  原生格式走单独端点）。
                </Paragraph>
              </Card>
            ),
          },
          {
            title: '复制 curl 或 SDK',
            description: (
              <Card size="small" styles={{ body: { padding: 12 } }}>
                <Paragraph type="secondary" style={{ marginBottom: 8 }}>
                  将下面命令中的 <Text code>{exampleKeyPlaceholder}</Text> 换成你的平台密钥。
                </Paragraph>
                <pre
                  style={{
                    margin: 0,
                    padding: 12,
                    background: '#f5f5f5',
                    borderRadius: 8,
                    fontSize: 12,
                    overflow: 'auto',
                  }}
                >
                  {curlExample}
                </pre>
                <Button
                  style={{ marginTop: 8 }}
                  size="small"
                  icon={<CopyOutlined />}
                  onClick={() => copy(curlExample, ' curl 示例')}
                >
                  复制 curl 示例
                </Button>
                <Paragraph type="secondary" style={{ marginTop: 12, marginBottom: 0 }}>
                  TypeScript（OpenAI SDK）：<Text code>baseURL: &apos;{openaiBase}&apos;</Text>，{' '}
                  <Text code>apiKey: process.env.PTD_KEY</Text>。
                </Paragraph>
              </Card>
            ),
          },
          {
            title: '浏览器内试打一条（真实扣费）',
            description: (
              <>
                {!usageGuideLoading && (usageGuide?.items?.length ?? 0) === 0 ? (
                  <Alert
                    type="warning"
                    showIcon
                    message="暂无权益对应的模型"
                    description={
                      <Space direction="vertical">
                        <span>购买套餐或权益后，此处将显示可用 model。</span>
                        <Link to="/packages">
                          <Button type="primary" size="small">
                            去套餐包
                          </Button>
                        </Link>
                        <Link to="/my/entitlements">
                          <Button size="small">我的权益</Button>
                        </Link>
                      </Space>
                    }
                  />
                ) : (
                  <EntitlementModelVerifyCard
                    usageGuide={usageGuide}
                    loadingGuide={usageGuideLoading}
                    tokenBalance={tokenBalance}
                    balanceLoading={balanceLoading}
                    onRefreshBalance={fetchBalance}
                    onVerifySuccess={onVerifySuccess}
                  />
                )}
              </>
            ),
          },
        ]}
      />

      <Card size="small" title="首调进度（本地）">
        <Checkbox
          checked={doneLocal}
          onChange={(e) => {
            const v = e.target.checked;
            setDoneLocal(v);
            if (v) localStorage.setItem(LS_QUICKSTART_DONE, '1');
            else localStorage.removeItem(LS_QUICKSTART_DONE);
          }}
        >
          我已完成首次成功调用
        </Checkbox>
      </Card>

      <Alert
        type="info"
        showIcon
        message="更多说明"
        description={
          <a href="/docs/developer/openai.md" target="_blank" rel="noreferrer">
            OpenAI 兼容接入（Markdown）
          </a>
        }
      />
    </Space>
  );
}
