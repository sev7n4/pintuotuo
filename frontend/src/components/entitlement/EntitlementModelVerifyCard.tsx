import { useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  Divider,
  Input,
  Select,
  Space,
  Tooltip,
  Typography,
  message,
} from 'antd';
import {
  ApiOutlined,
  CommentOutlined,
  ExperimentOutlined,
  InfoCircleOutlined,
  MessageOutlined,
  PieChartOutlined,
  ReloadOutlined,
  SendOutlined,
  WalletOutlined,
} from '@ant-design/icons';
import api from '@/services/api';
import type { APIUsageGuideResponse } from '@/types';
import { modelValueFromItem } from '@/utils/apiUsageGuideModel';

const { Paragraph, Text } = Typography;

type ChatCompletionResponse = {
  choices?: Array<{ message?: { content?: string } }>;
  usage?: {
    prompt_tokens?: number;
    completion_tokens?: number;
    total_tokens?: number;
  };
  error?: { message?: string };
};

type Props = {
  usageGuide: APIUsageGuideResponse | null;
  loadingGuide: boolean;
  /** 与页头统计同源，便于本卡片内「所见即所得」对照余额 */
  tokenBalance: number | null;
  balanceLoading: boolean;
  onRefreshBalance: () => Promise<void>;
  /** 验证成功并已刷新余额后回调（如开发者中心首调埋点） */
  onVerifySuccess?: () => void;
};

/** 登录态 JWT 调用 OpenAI 兼容接口，仅用于权益页连通性验证（真实计费）。 */
export function EntitlementModelVerifyCard({
  usageGuide,
  loadingGuide,
  tokenBalance,
  balanceLoading,
  onRefreshBalance,
  onVerifySuccess,
}: Props) {
  const items = useMemo(() => usageGuide?.items ?? [], [usageGuide]);
  const [model, setModel] = useState<string>('');
  const [prompt, setPrompt] = useState('请回复：ok');
  const [submitting, setSubmitting] = useState(false);
  const [refreshingBalance, setRefreshingBalance] = useState(false);
  const [lastReply, setLastReply] = useState<string | null>(null);
  const [lastError, setLastError] = useState<string | null>(null);
  const [lastUsage, setLastUsage] = useState<ChatCompletionResponse['usage'] | null>(null);

  const options = useMemo(
    () =>
      items.map((it) => ({
        value: modelValueFromItem(it),
        label: `${it.provider_code} · ${modelValueFromItem(it)}`,
      })),
    [items]
  );

  useEffect(() => {
    if (items.length === 0) return;
    setModel((prev) => {
      if (prev.trim()) return prev;
      const def = usageGuide?.default_model_example?.trim();
      const first = modelValueFromItem(items[0]);
      if (def && items.some((it) => modelValueFromItem(it) === def)) return def;
      return first;
    });
  }, [items, usageGuide?.default_model_example]);

  if (loadingGuide) {
    return null;
  }
  if (items.length === 0) {
    return null;
  }

  const handleRefreshBalance = async () => {
    setRefreshingBalance(true);
    try {
      await onRefreshBalance();
      message.success('余额已更新');
    } catch {
      message.error('刷新失败');
    } finally {
      setRefreshingBalance(false);
    }
  };

  const onVerify = async () => {
    const m = model.trim();
    const content = prompt.trim();
    if (!m) {
      message.warning('请选择要验证的模型');
      return;
    }
    if (!content) {
      message.warning('请输入验证内容');
      return;
    }
    setSubmitting(true);
    setLastReply(null);
    setLastError(null);
    setLastUsage(null);
    try {
      const res = await api.post<ChatCompletionResponse>(
        '/openai/v1/chat/completions',
        {
          model: m,
          messages: [{ role: 'user', content }],
          max_tokens: 64,
        },
        { timeout: 120000 }
      );
      const data = res.data;
      if (data.error?.message) {
        setLastError(data.error.message);
        message.error('验证失败');
        return;
      }
      const text = data.choices?.[0]?.message?.content ?? '';
      setLastReply(text || '（空回复）');
      if (data.usage) {
        setLastUsage(data.usage);
      }
      message.success('验证成功');
      await onRefreshBalance();
      onVerifySuccess?.();
    } catch (e: unknown) {
      let msg = '请求失败';
      if (typeof e === 'object' && e !== null && 'response' in e) {
        const data = (e as { response?: { data?: { error?: { message?: string } } } }).response
          ?.data;
        msg = data?.error?.message ?? msg;
      } else if (e instanceof Error) {
        msg = e.message;
      }
      setLastError(msg);
      message.error('验证失败');
    } finally {
      setSubmitting(false);
    }
  };

  const balanceDisplay =
    balanceLoading && tokenBalance == null ? (
      <Text type="secondary">加载中…</Text>
    ) : (
      <Space size={6}>
        <Text strong>{tokenBalance != null ? tokenBalance.toLocaleString('zh-CN') : '—'}</Text>
        <Text type="secondary">Token</Text>
      </Space>
    );

  return (
    <Card
      title={
        <Space size={8}>
          <ExperimentOutlined style={{ color: '#722ed1' }} />
          <span>权益连通性验证</span>
        </Space>
      }
    >
      <Alert
        type="info"
        icon={<InfoCircleOutlined />}
        style={{ marginBottom: 12 }}
        message="使用当前登录账号（JWT）发起真实调用，将按平台规则扣减 Token。"
      />

      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          flexWrap: 'wrap',
          gap: 8,
          marginBottom: 12,
          padding: '10px 12px',
          background: 'var(--ant-color-fill-alter, #fafafa)',
          borderRadius: 8,
          border: '1px solid var(--ant-color-border-secondary, #f0f0f0)',
        }}
      >
        <Space size="middle" wrap>
          <Space size={6}>
            <WalletOutlined style={{ color: '#52c41a', fontSize: 16 }} />
            <Text type="secondary">当前余额</Text>
          </Space>
          {balanceDisplay}
        </Space>
        <Tooltip title="重新拉取余额，调用成功后也会自动刷新">
          <Button
            type="default"
            size="small"
            icon={<ReloadOutlined />}
            loading={refreshingBalance || balanceLoading}
            onClick={handleRefreshBalance}
            aria-label="刷新余额"
          >
            刷新
          </Button>
        </Tooltip>
      </div>

      <Space direction="vertical" size="middle" style={{ width: '100%' }}>
        <div>
          <Space size={6} style={{ marginBottom: 4 }}>
            <ApiOutlined style={{ color: '#1677ff' }} />
            <Text type="secondary">模型（来自当前权益）</Text>
          </Space>
          <Select
            style={{ width: '100%' }}
            options={options}
            value={model || undefined}
            onChange={setModel}
            showSearch
            optionFilterProp="label"
          />
        </div>

        <div>
          <Space size={6} style={{ marginBottom: 4 }}>
            <MessageOutlined style={{ color: '#fa8c16' }} />
            <Text type="secondary">验证内容</Text>
          </Space>
          <Input.TextArea
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            rows={3}
            placeholder="短句即可，例如：请用一句话自我介绍"
          />
        </div>

        <Button
          type="primary"
          icon={<SendOutlined />}
          onClick={onVerify}
          loading={submitting}
          block
        >
          发送验证
        </Button>

        {lastUsage != null &&
          (lastUsage.prompt_tokens != null ||
            lastUsage.completion_tokens != null ||
            lastUsage.total_tokens != null) && (
            <>
              <Divider plain style={{ margin: '4px 0' }}>
                <Space size={4}>
                  <PieChartOutlined />
                  <Text type="secondary">本次用量</Text>
                </Space>
              </Divider>
              <Paragraph style={{ marginBottom: 0 }}>
                <Space size="large" wrap>
                  {lastUsage.prompt_tokens != null && (
                    <Text>
                      <PieChartOutlined /> 输入 {lastUsage.prompt_tokens.toLocaleString('zh-CN')}
                    </Text>
                  )}
                  {lastUsage.completion_tokens != null && (
                    <Text>
                      <PieChartOutlined /> 输出{' '}
                      {lastUsage.completion_tokens.toLocaleString('zh-CN')}
                    </Text>
                  )}
                  {lastUsage.total_tokens != null && (
                    <Text type="secondary">
                      合计 {lastUsage.total_tokens.toLocaleString('zh-CN')} tokens
                    </Text>
                  )}
                </Space>
              </Paragraph>
            </>
          )}

        {lastReply != null && (
          <>
            <Divider plain style={{ margin: '4px 0' }}>
              <Space size={4}>
                <CommentOutlined />
                <Text type="secondary">模型回复</Text>
              </Space>
            </Divider>
            <Paragraph style={{ marginBottom: 0 }}>
              <Text>{lastReply}</Text>
            </Paragraph>
          </>
        )}

        {lastError != null && <Alert type="error" message={lastError} showIcon />}
      </Space>
    </Card>
  );
}
