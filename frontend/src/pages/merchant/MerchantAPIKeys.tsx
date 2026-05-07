import { useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Card,
  Table,
  Button,
  Tag,
  Space,
  Modal,
  Form,
  Input,
  InputNumber,
  Select,
  Segmented,
  Switch,
  message,
  Popconfirm,
  Progress,
  Tooltip,
  Descriptions,
  Divider,
  Spin,
  Checkbox,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  InfoCircleOutlined,
  SyncOutlined,
  ApiOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import { useMerchantStore } from '@/stores/merchantStore';
import { MerchantAPIKey, VerificationResult } from '@/types';
import type { ModelProvider } from '@/types/sku';
import api from '@/services/api';
import { merchantService } from '@/services/merchant';
import { merchantSkuService } from '@/services/merchantSku';
import styles from './MerchantAPIKeys.module.css';
import type { MerchantSKUDetail } from '@/types/merchantSku';
import { isStrictEntitlementEligible } from '@/utils/byokStatus';
import MerchantBYOKOverviewCard from '@/components/merchant/MerchantBYOKOverviewCard';

/** GET /merchants/api-keys/:id/verification response shape */
interface VerificationPollResponse {
  api_key: MerchantAPIKey;
  history: VerificationResult[];
}

function normalizeVerificationStatus(raw: string | undefined): VerificationResult['status'] {
  if (raw === 'verified') return 'success';
  if (raw === 'success') return 'success';
  if (raw === 'failed') return 'failed';
  if (raw === 'in_progress') return 'in_progress';
  return 'pending';
}

function buildVerificationView(
  keyId: number,
  payload: VerificationPollResponse
): VerificationResult {
  const latest = payload.history?.[0];
  if (latest) {
    return {
      ...latest,
      status: normalizeVerificationStatus(latest.status),
    };
  }
  const keyStatus = normalizeVerificationStatus(payload.api_key?.verification_result);
  if (keyStatus === 'success' || keyStatus === 'failed') {
    return {
      id: 0,
      api_key_id: keyId,
      verification_type: 'manual',
      status: keyStatus,
      connection_test: keyStatus === 'success',
      models_found: payload.api_key?.models_supported,
      models_count: payload.api_key?.models_supported?.length || 0,
      pricing_verified: false,
      error_message: payload.api_key?.verification_message,
      started_at: new Date().toISOString(),
      retry_count: 0,
    };
  }
  return {
    id: 0,
    api_key_id: keyId,
    verification_type: 'manual',
    status: 'in_progress',
    connection_test: false,
    models_count: 0,
    pricing_verified: false,
    started_at: new Date().toISOString(),
    retry_count: 0,
  };
}

function healthDotClass(status?: string): string {
  const s = (status || 'unknown').toLowerCase();
  if (s === 'healthy') return styles.statusDotHealthy;
  if (s === 'degraded') return styles.statusDotDegraded;
  if (s === 'unhealthy') return styles.statusDotUnhealthy;
  return styles.statusDotUnknown;
}

function healthLabel(status?: string): string {
  const s = (status || 'unknown').toLowerCase();
  if (s === 'healthy') return '健康';
  if (s === 'degraded') return '降级';
  if (s === 'unhealthy') return '不健康';
  return '未知';
}

function healthTooltipDesc(status?: string): string {
  const s = (status || 'unknown').toLowerCase();
  const base = `当前健康状态：${healthLabel(status)}。`;
  if (s === 'healthy' || s === 'degraded') {
    return `${base}在平台开启 strict 权益时，可作为路由候选（与验证条件同时满足）。真实调用成功也可能将「未知」更新为健康（被动健康）。`;
  }
  if (s === 'unhealthy') {
    return `${base}请检查上游 Key、网络或「立即探测」结果；strict 下通常不可进入权益白名单。`;
  }
  return `${base}尚未探测或仍为初始值；请使用「立即探测」或等待主动探测，strict 下需为健康或降级才可进白名单。`;
}

function verificationDotClass(result?: string): string {
  const r = (result || '').toLowerCase();
  if (r === 'verified' || r === 'success') return styles.statusDotVerified;
  if (r === 'failed') return styles.statusDotVerifyFailed;
  return styles.statusDotVerifyPending;
}

function verificationLabel(result?: string): string {
  const r = (result || '').toLowerCase();
  if (r === 'verified' || r === 'success') return '已验证';
  if (r === 'failed') return '验证失败';
  if (r === 'pending' || r === 'in_progress') return '验证中';
  return '未验证';
}

function verificationTooltipDesc(k: MerchantAPIKey): string {
  const r = (k.verification_result || '').toLowerCase();
  const base = `验证结果：${verificationLabel(k.verification_result)}。`;
  if (r === 'verified' || r === 'success') {
    return `${base}已通过上游 OpenAI 兼容 /models（及深度验证时的额外检查）。`;
  }
  if (r === 'failed') {
    return `${base}请修正 Key 或厂商配置后重新「轻量/深度验证」。注意：若仅有验证时间但结果为失败，strict 仍可能因其它条件不通过。`;
  }
  return `${base}请完成「轻量验证」或「深度验证」；通过后会写入 verified。`;
}

function formatHealthError(record: MerchantAPIKey): string {
  const parts: string[] = [];
  if (record.health_error_category)
    parts.push(`分类: ${toHealthCategoryCN(record.health_error_category)}`);
  if (record.health_error_code) parts.push(`上游码: ${record.health_error_code}`);
  if (record.health_provider_request_id) parts.push(`请求ID: ${record.health_provider_request_id}`);
  if (record.health_error_message) parts.push(`信息: ${record.health_error_message}`);
  return parts.join(' | ');
}

function toHealthCategoryCN(category?: string): string {
  const c = (category || '').toUpperCase();
  const map: Record<string, string> = {
    AUTH_INVALID_KEY: '鉴权失败（Key无效）',
    AUTH_PERMISSION_DENIED: '鉴权失败（权限不足）',
    QUOTA_INSUFFICIENT: '额度不足',
    RATE_LIMITED: '触发限流',
    MODEL_NOT_FOUND: '模型不存在',
    CONTEXT_WINDOW_EXCEEDED: '上下文超限',
    SERVICE_UNAVAILABLE: '上游服务不可用',
    NETWORK_TIMEOUT: '网络超时',
    NETWORK_DNS: 'DNS/域名解析失败',
    UPSTREAM_BAD_REQUEST: '请求参数错误',
    UNKNOWN: '未知错误',
  };
  return map[c] || c || '未知错误';
}

const MerchantAPIKeys = () => {
  const {
    apiKeys,
    apiKeyUsage,
    fetchAPIKeys,
    fetchAPIKeyUsage,
    createAPIKey,
    updateAPIKey,
    deleteAPIKey,
    isLoading,
  } = useMerchantStore();
  const [modalVisible, setModalVisible] = useState(false);
  const [editingKey, setEditingKey] = useState<MerchantAPIKey | null>(null);
  const [form] = Form.useForm();
  const [verificationModalVisible, setVerificationModalVisible] = useState(false);
  const [verificationResult, setVerificationResult] = useState<VerificationResult | null>(null);
  const [verificationLoading, setVerificationLoading] = useState(false);

  const [probeModelModalVisible, setProbeModelModalVisible] = useState(false);
  const [probeModelTargetId, setProbeModelTargetId] = useState<number | null>(null);
  const [selectedProbeModel, setSelectedProbeModel] = useState<string | undefined>(undefined);
  const [cachedModels, setCachedModels] = useState<Map<number, string[]>>(new Map());
  const [modelProviders, setModelProviders] = useState<ModelProvider[]>([]);
  const [providersLoading, setProvidersLoading] = useState(false);
  const [merchantSKUs, setMerchantSKUs] = useState<MerchantSKUDetail[]>([]);
  const [keyword, setKeyword] = useState('');
  const [providerFilter, setProviderFilter] = useState<string>('all');
  const [byokTypeFilter, setByokTypeFilter] = useState<string>('all');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [healthFilter, setHealthFilter] = useState<string>('all');
  const [verifyFilter, setVerifyFilter] = useState<string>('all');
  const [strictFilter, setStrictFilter] = useState<string>('all');
  const [quickFilter, setQuickFilter] = useState<'all' | 'recent' | 'attention' | 'verifying'>(
    'all'
  );
  const [recentMinutes, setRecentMinutes] = useState<number>(10);
  const [autoRefresh, setAutoRefresh] = useState(false);

  useEffect(() => {
    fetchAPIKeys();
    fetchAPIKeyUsage();
  }, [fetchAPIKeys, fetchAPIKeyUsage]);

  useEffect(() => {
    if (!autoRefresh) return;
    const timer = window.setInterval(() => {
      fetchAPIKeys();
      fetchAPIKeyUsage();
    }, 8000);
    return () => window.clearInterval(timer);
  }, [autoRefresh, fetchAPIKeys, fetchAPIKeyUsage]);

  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      setProvidersLoading(true);
      try {
        const res = await merchantService.getMerchantModelProviders();
        if (!cancelled && res.data?.data) {
          setModelProviders(res.data.data);
        }
      } catch {
        if (!cancelled) {
          message.error('加载提供商列表失败');
          setModelProviders([]);
        }
      } finally {
        if (!cancelled) setProvidersLoading(false);
      }
    };
    load();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    merchantSkuService
      .getMerchantSKUs('active')
      .then((data) => setMerchantSKUs(data || []))
      .catch(() => setMerchantSKUs([]));
  }, []);

  const getCostSource = (key: MerchantAPIKey): { label: string; color: string; hint: string } => {
    if ((key.cost_input_rate ?? 0) > 0 || (key.cost_output_rate ?? 0) > 0) {
      return { label: 'Key自定义', color: 'blue', hint: '已在 API Key 直接配置' };
    }
    const linked = merchantSKUs.find((s) => s.api_key_id === key.id);
    if (linked?.custom_pricing_enabled) {
      return { label: 'SKU自定义', color: 'gold', hint: '来自商户上架SKU自定义成本' };
    }
    if (linked) {
      return { label: 'SPU继承', color: 'green', hint: '来自SKU继承的SPU参考价' };
    }
    return { label: '未绑定', color: 'default', hint: '建议先绑定SKU继承默认成本' };
  };

  const handleAdd = () => {
    setEditingKey(null);
    form.resetFields();
    form.setFieldsValue({ unlimited_quota: true, health_check_level: 'medium', byok_type: 'official' });
    setModalVisible(true);
  };

  const handleEdit = (record: MerchantAPIKey) => {
    setEditingKey(record);
    const linked = merchantSKUs.find((s) => s.api_key_id === record.id);
    const fallbackInput = linked
      ? linked.custom_pricing_enabled
        ? linked.cost_input_rate
        : linked.spu_input_rate
      : undefined;
    const fallbackOutput = linked
      ? linked.custom_pricing_enabled
        ? linked.cost_output_rate
        : linked.spu_output_rate
      : undefined;
    const fallbackMargin = linked?.profit_margin;
    const unlimited = record.quota_limit == null || record.quota_limit === 0;
    form.setFieldsValue({
      name: record.name,
      unlimited_quota: unlimited,
      quota_limit: unlimited ? undefined : record.quota_limit,
      status: record.status,
      endpoint_url: record.endpoint_url,
      health_check_level: record.health_check_level,
      cost_input_rate: record.cost_input_rate ?? fallbackInput,
      cost_output_rate: record.cost_output_rate ?? fallbackOutput,
      profit_margin: record.profit_margin ?? fallbackMargin,
      region: record.region || 'domestic',
      security_level: record.security_level || 'standard',
      byok_type: record.byok_type || 'official',
    });
    setModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    const success = await deleteAPIKey(id);
    if (success) {
      message.success('API密钥已删除');
      fetchAPIKeys();
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (editingKey) {
        const unlimited = Boolean(values.unlimited_quota);
        const patch: Partial<MerchantAPIKey> = {
          name: values.name as string,
          status: values.status as MerchantAPIKey['status'],
          endpoint_url: values.endpoint_url as string | undefined,
          health_check_level: values.health_check_level as MerchantAPIKey['health_check_level'],
          cost_input_rate: values.cost_input_rate as number | undefined,
          cost_output_rate: values.cost_output_rate as number | undefined,
          profit_margin: values.profit_margin as number | undefined,
          quota_limit: unlimited ? null : (values.quota_limit as number),
          region: values.region as 'domestic' | 'overseas' | undefined,
          security_level: values.security_level as 'standard' | 'high' | undefined,
          byok_type: values.byok_type as 'official' | 'reseller' | 'self_hosted' | undefined,
        };
        const success = await updateAPIKey(editingKey.id, patch);
        if (!success) {
          const msg = useMerchantStore.getState().error || '更新失败';
          message.error(msg);
          throw new Error(msg);
        }
        message.success('API密钥已更新');
        setModalVisible(false);
        fetchAPIKeys();
        return;
      }
      const payload = {
        name: values.name as string,
        provider: values.provider as string,
        api_key: values.api_key as string,
        api_secret: values.api_secret as string | undefined,
        quota_limit: values.unlimited_quota ? null : (values.quota_limit as number),
        health_check_level:
          (values.health_check_level as MerchantAPIKey['health_check_level']) || 'medium',
        endpoint_url: (values.endpoint_url as string | undefined)?.trim() || undefined,
        region: (values.region as 'domestic' | 'overseas') || 'domestic',
        security_level: (values.security_level as 'standard' | 'high') || 'standard',
        byok_type: (values.byok_type as 'official' | 'reseller' | 'self_hosted') || 'official',
      };
      const success = await createAPIKey(payload);
      if (!success) {
        const msg = useMerchantStore.getState().error || '创建失败';
        message.error(msg);
        throw new Error(msg);
      }
      message.success('API密钥已创建');
      setModalVisible(false);
      fetchAPIKeys();
    } catch (e: unknown) {
      if (e && typeof e === 'object' && 'errorFields' in e) {
        throw e;
      }
      if (!(e instanceof Error)) {
        message.error('操作失败');
      }
      throw e instanceof Error ? e : new Error('操作失败');
    }
  };

  const handleVerify = async (id: number, mode: 'light' | 'deep' = 'light', probeModel?: string) => {
    setVerificationModalVisible(true);
    setVerificationLoading(true);
    setVerificationResult(null);

    try {
      const body: Record<string, string> = { verification_mode: mode };
      if (probeModel) body.probe_model = probeModel;
      await api.post(`/merchants/api-keys/${id}/verify`, body);
      message.success(
        mode === 'deep'
          ? '深度验证已启动（包含配额探测，支持的提供商会执行）'
          : '轻量验证已启动，正在后台执行...'
      );

      await pollVerificationResult(id);
    } catch (error) {
      message.error('启动验证失败');
      setVerificationLoading(false);
    }
  };

  const openProbeModelSelector = (id: number) => {
    setProbeModelTargetId(id);
    const existing = cachedModels.get(id) || [];
    setSelectedProbeModel(existing.length > 0 ? existing[0] : undefined);
    setProbeModelModalVisible(true);
  };

  const confirmProbeModel = () => {
    setProbeModelModalVisible(false);
    if (probeModelTargetId !== null) {
      handleVerify(probeModelTargetId, 'deep', selectedProbeModel);
    }
  };

  const skipProbeModel = () => {
    setProbeModelModalVisible(false);
    if (probeModelTargetId !== null) {
      handleVerify(probeModelTargetId, 'deep');
    }
  };

  const handleImmediateHealthCheck = async (id: number) => {
    try {
      await api.post(`/merchants/api-keys/${id}/health-check`);
      message.success('已触发立即健康探测，正在获取结果...');
      fetchAPIKeys();
      void pollHealthCheckResult(id);
    } catch {
      message.error('触发健康探测失败');
    }
  };

  const pollHealthCheckResult = async (id: number) => {
    const maxAttempts = 8;
    const interval = 2000;
    for (let i = 0; i < maxAttempts; i++) {
      await new Promise((resolve) => setTimeout(resolve, interval));
      await fetchAPIKeys();
      const latest = useMerchantStore.getState().apiKeys.find((k) => k.id === id);
      if (!latest?.last_health_check_at) continue;
      const health = (latest.health_status || 'unknown').toLowerCase();
      if (health === 'healthy' || health === 'degraded') {
        message.success(`探测完成：当前状态 ${health === 'healthy' ? '健康' : '降级'}`);
        return;
      }
      if (health === 'unhealthy') {
        const reason = formatHealthError(latest).trim();
        message.error(reason ? `探测失败：${reason}` : '探测失败：状态不健康');
        return;
      }
    }
    message.warning('探测已触发，但结果尚未返回，请稍后手动刷新查看');
  };

  const pollVerificationResult = async (id: number) => {
    const maxAttempts = 30;
    const interval = 2000;
    let attempts = 0;

    const poll = async (): Promise<void> => {
      try {
        const response = await api.get<VerificationPollResponse>(
          `/merchants/api-keys/${id}/verification`
        );
        const result = buildVerificationView(id, response.data);

        setVerificationResult(result);

        if (result.models_found && result.models_found.length > 0) {
          setCachedModels(prev => {
            const next = new Map(prev);
            next.set(id, result.models_found!);
            return next;
          });
        }

        if (result.status === 'pending' || result.status === 'in_progress') {
          attempts++;
          if (attempts < maxAttempts) {
            await new Promise((resolve) => setTimeout(resolve, interval));
            await poll();
          } else {
            message.warning('验证超时，请稍后查看结果');
            setVerificationLoading(false);
          }
        } else {
          setVerificationLoading(false);
          fetchAPIKeys();
        }
      } catch (error) {
        attempts++;
        if (attempts < maxAttempts) {
          await new Promise((resolve) => setTimeout(resolve, interval));
          await poll();
        } else {
          message.error('获取验证结果失败');
          setVerificationLoading(false);
        }
      }
    };

    await poll();
  };

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '提供商',
      dataIndex: 'provider',
      key: 'provider',
      render: (provider: string) => <Tag color="blue">{provider.toUpperCase()}</Tag>,
    },
    {
      title: 'BYOK类型',
      dataIndex: 'byok_type',
      key: 'byok_type',
      render: (byokType?: string) => {
        const type = byokType || 'official';
        const colorMap: Record<string, string> = {
          official: 'blue',
          reseller: 'orange',
          self_hosted: 'purple',
        };
        const labelMap: Record<string, string> = {
          official: '官方',
          reseller: '代理商',
          self_hosted: '自建商',
        };
        return <Tag color={colorMap[type] || 'default'}>{labelMap[type] || type}</Tag>;
      },
    },
    {
      title: '区域',
      dataIndex: 'region',
      key: 'region',
      render: (region?: string) => (
        <Tag color={region === 'overseas' ? 'blue' : 'green'}>
          {region === 'overseas' ? '海外' : '国内'}
        </Tag>
      ),
    },
    {
      title: '安全等级',
      dataIndex: 'security_level',
      key: 'security_level',
      render: (level?: string) => (
        <Tag color={level === 'high' ? 'red' : 'default'}>
          {level === 'high' ? '高安全' : '标准'}
        </Tag>
      ),
    },
    {
      title: '端点URL',
      dataIndex: 'endpoint_url',
      key: 'endpoint_url',
      render: (url?: string) => (
        <Tooltip title={url || '使用默认端点'}>
          <span>{url ? url.substring(0, 30) + '...' : '默认'}</span>
        </Tooltip>
      ),
    },
    {
      title: '配额',
      dataIndex: 'quota_limit',
      key: 'quota_limit',
      render: (_: unknown, record: MerchantAPIKey) => {
        const usage = apiKeyUsage.find((u) => u.id === record.id);
        if (
          !usage ||
          usage.quota_limit === null ||
          usage.quota_limit === undefined ||
          usage.quota_limit === 0
        ) {
          return '无限制';
        }
        const percent = Math.min(usage.usage_percentage, 100);
        return (
          <div className={styles.quotaCell}>
            <Progress percent={percent} size="small" />
            <span className={styles.quotaText}>
              {usage.quota_used.toFixed(2)} / {usage.quota_limit.toFixed(2)} Token
            </span>
          </div>
        );
      },
    },
    {
      title: (
        <span>
          健康状态{' '}
          <Tooltip title="绿灯=健康，黄灯=降级，红灯=不健康，灰灯=未知。悬停查看说明。">
            <InfoCircleOutlined style={{ color: '#8c8c8c' }} />
          </Tooltip>
        </span>
      ),
      dataIndex: 'health_status',
      key: 'health_status',
      width: 168,
      render: (status: string, record: MerchantAPIKey) => (
        <Space direction="vertical" size={4}>
          <Tooltip title={healthTooltipDesc(status)}>
            <span className={styles.statusLightRow}>
              <span className={`${styles.statusDot} ${healthDotClass(status)}`} aria-hidden />
              <span className={styles.statusLightLabel}>{healthLabel(status)}</span>
            </span>
          </Tooltip>
          {record.last_health_check_at && (
            <span style={{ fontSize: '12px', color: '#999' }}>
              {new Date(record.last_health_check_at).toLocaleString('zh-CN')}
            </span>
          )}
          {record.health_status === 'unhealthy' && (
            <Tooltip title={formatHealthError(record) || '暂无结构化错误信息'}>
              <Tag color="error" style={{ marginRight: 0 }}>
                {toHealthCategoryCN(record.health_error_category) || '探测失败'}
              </Tag>
            </Tooltip>
          )}
        </Space>
      ),
    },
    {
      title: (
        <span>
          验证状态{' '}
          <Tooltip title="绿灯=已通过验证，红灯=失败，灰灯=未验证/进行中。与轻量/深度验证结果一致。">
            <InfoCircleOutlined style={{ color: '#8c8c8c' }} />
          </Tooltip>
        </span>
      ),
      dataIndex: 'verification_result',
      key: 'verification_result',
      width: 168,
      render: (_: string, record: MerchantAPIKey) => (
        <Space direction="vertical" size={4}>
          <Tooltip title={verificationTooltipDesc(record)}>
            <span className={styles.statusLightRow}>
              <span
                className={`${styles.statusDot} ${verificationDotClass(record.verification_result)}`}
                aria-hidden
              />
              <span className={styles.statusLightLabel}>
                {verificationLabel(record.verification_result)}
              </span>
            </span>
          </Tooltip>
          {record.verified_at && (
            <span style={{ fontSize: '12px', color: '#999' }}>
              最近验证：{new Date(record.verified_at).toLocaleString('zh-CN')}
            </span>
          )}
        </Space>
      ),
    },
    {
      title: (
        <span>
          Strict 权益{' '}
          <Tooltip title="与后端 strict 路由白名单一致：需「验证条件 + 健康为健康/降级」同时满足。">
            <InfoCircleOutlined style={{ color: '#8c8c8c' }} />
          </Tooltip>
        </span>
      ),
      key: 'strict_entitlement',
      width: 112,
      render: (_: unknown, record: MerchantAPIKey) => {
        const ok = isStrictEntitlementEligible(record);
        return (
          <Tooltip
            title={
              ok
                ? '当前满足 strict 权益路由对密钥的筛选条件（仍要求 SKU 承接等数据完整）。'
                : '未同时满足：需 (已有验证记录或 verification=verified) 且 健康为「健康」或「降级」。请完成验证并「立即探测」。'
            }
          >
            <Tag color={ok ? 'success' : 'warning'}>{ok ? '可路由' : '未满足'}</Tag>
          </Tooltip>
        );
      },
    },
    {
      title: '成本来源',
      key: 'cost_source',
      render: (_: unknown, record: MerchantAPIKey) => {
        const source = getCostSource(record);
        return (
          <Space direction="vertical" size={0}>
            <Tag color={source.color}>{source.label}</Tag>
            <span style={{ fontSize: 12, color: '#999' }}>{source.hint}</span>
          </Space>
        );
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'active' ? 'success' : 'default'}>
          {status === 'active' ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 220,
      render: (_: unknown, record: MerchantAPIKey) => (
        <Space size="small">
          <Tooltip title="轻量验证（连通性/鉴权）">
            <Button
              type="link"
              size="small"
              icon={<ApiOutlined />}
              onClick={() => handleVerify(record.id, 'light')}
            >
              轻量验证
            </Button>
          </Tooltip>
          <Tooltip title="深度验证（api_format=openai 的厂商会探测 /chat/completions 是否可用，含上游余额类错误）">
            <Button type="link" size="small" onClick={() => openProbeModelSelector(record.id)}>
              深度验证
            </Button>
          </Tooltip>
          <Tooltip title="立即健康探测（绕过周期节流，按该 Key 的探测策略执行一次）">
            <Button type="link" size="small" onClick={() => handleImmediateHealthCheck(record.id)}>
              立即探测
            </Button>
          </Tooltip>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这个API密钥吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const providerOptions = useMemo(() => {
    const set = new Set<string>();
    for (const k of apiKeys) {
      if (k.provider) set.add(k.provider.toLowerCase());
    }
    return Array.from(set).sort();
  }, [apiKeys]);

  const filteredApiKeys = useMemo(() => {
    const kw = keyword.trim().toLowerCase();
    return apiKeys.filter((k) => {
      if (providerFilter !== 'all' && (k.provider || '').toLowerCase() !== providerFilter) {
        return false;
      }
      if (byokTypeFilter !== 'all' && (k.byok_type || 'official') !== byokTypeFilter) {
        return false;
      }
      if (statusFilter !== 'all' && (k.status || '') !== statusFilter) {
        return false;
      }
      const health = (k.health_status || 'unknown').toLowerCase();
      if (healthFilter !== 'all' && health !== healthFilter) {
        return false;
      }
      const vr = normalizeVerificationStatus(k.verification_result);
      if (verifyFilter !== 'all' && vr !== verifyFilter) {
        return false;
      }
      const strictOk = isStrictEntitlementEligible(k);
      if (strictFilter === 'routable' && !strictOk) {
        return false;
      }
      if (strictFilter === 'unmet' && strictOk) {
        return false;
      }
      if (kw) {
        const hay =
          `${k.name || ''} ${(k.provider || '').toLowerCase()} ${k.endpoint_url || ''}`.toLowerCase();
        if (!hay.includes(kw)) {
          return false;
        }
      }

      if (quickFilter === 'attention') {
        if (!((k.status || '') === 'active' && !isStrictEntitlementEligible(k))) return false;
      }
      if (quickFilter === 'verifying') {
        if (normalizeVerificationStatus(k.verification_result) !== 'in_progress') return false;
      }
      if (quickFilter === 'recent') {
        if (!k.updated_at) return false;
        const updated = new Date(k.updated_at).getTime();
        if (Number.isNaN(updated)) return false;
        if (Date.now() - updated > recentMinutes * 60 * 1000) return false;
      }
      return true;
    });
  }, [
    apiKeys,
    byokTypeFilter,
    healthFilter,
    keyword,
    providerFilter,
    quickFilter,
    recentMinutes,
    statusFilter,
    strictFilter,
    verifyFilter,
  ]);

  const resetFilters = () => {
    setKeyword('');
    setProviderFilter('all');
    setByokTypeFilter('all');
    setStatusFilter('all');
    setHealthFilter('all');
    setVerifyFilter('all');
    setStrictFilter('all');
    setQuickFilter('all');
    setRecentMinutes(10);
  };

  return (
    <div className={styles.apiKeys}>
      <div className={styles.header}>
        <h2 className={styles.pageTitle}>API密钥管理</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
          添加密钥
        </Button>
      </div>

      <Alert
        type="info"
        showIcon
        icon={<InfoCircleOutlined />}
        style={{ marginBottom: 16 }}
        message="BYOK 与 strict 权益路由"
        description={
          <ol style={{ margin: 0, paddingLeft: 20 }}>
            <li>上传 Key 后请完成「轻量验证」或「深度验证」，使验证灯为绿色（已验证）。</li>
            <li>
              点击「立即探测」（或等待主动探测），使健康灯为绿（健康）或黄（降级）；避免长期停留在灰灯（未知）。
            </li>
            <li>
              「Strict 权益」列为「可路由」时，表示与后端白名单条件一致；仍需 SKU
              承接等数据完整，最终以接口与日志为准。
            </li>
          </ol>
        }
      />

      <MerchantBYOKOverviewCard apiKeys={apiKeys} />

      <Card>
        <Space wrap size={12} style={{ marginBottom: 12 }}>
          <Segmented
            value={quickFilter}
            onChange={(v) => setQuickFilter(v as typeof quickFilter)}
            options={[
              { label: '全部', value: 'all' },
              { label: '最近变化', value: 'recent' },
              { label: '待处理', value: 'attention' },
              { label: '验证中', value: 'verifying' },
            ]}
          />
          <Select
            style={{ width: 150 }}
            value={recentMinutes}
            onChange={(v) => setRecentMinutes(Number(v))}
            disabled={quickFilter !== 'recent'}
            options={[
              { value: 5, label: '最近5分钟' },
              { value: 10, label: '最近10分钟' },
              { value: 30, label: '最近30分钟' },
            ]}
          />
          <Space>
            <span style={{ fontSize: 12, color: '#666' }}>自动刷新</span>
            <Switch checked={autoRefresh} onChange={setAutoRefresh} />
          </Space>
        </Space>
        <Space wrap size={12} style={{ marginBottom: 12 }}>
          <Input
            allowClear
            style={{ width: 240 }}
            placeholder="搜索名称/供应商/端点"
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
          />
          <Select
            style={{ width: 140 }}
            value={providerFilter}
            onChange={setProviderFilter}
            options={[
              { value: 'all', label: '供应商：全部' },
              ...providerOptions.map((p) => ({ value: p, label: `供应商：${p.toUpperCase()}` })),
            ]}
          />
          <Select
            style={{ width: 130 }}
            value={byokTypeFilter}
            onChange={setByokTypeFilter}
            options={[
              { value: 'all', label: 'BYOK类型：全部' },
              { value: 'official', label: 'BYOK类型：官方' },
              { value: 'reseller', label: 'BYOK类型：代理商' },
              { value: 'self_hosted', label: 'BYOK类型：自建商' },
            ]}
          />
          <Select
            style={{ width: 120 }}
            value={statusFilter}
            onChange={setStatusFilter}
            options={[
              { value: 'all', label: '状态：全部' },
              { value: 'active', label: '状态：启用' },
              { value: 'inactive', label: '状态：禁用' },
            ]}
          />
          <Select
            style={{ width: 140 }}
            value={healthFilter}
            onChange={setHealthFilter}
            options={[
              { value: 'all', label: '健康灯：全部' },
              { value: 'healthy', label: '健康灯：健康' },
              { value: 'degraded', label: '健康灯：降级' },
              { value: 'unhealthy', label: '健康灯：不健康' },
              { value: 'unknown', label: '健康灯：未知' },
            ]}
          />
          <Select
            style={{ width: 160 }}
            value={verifyFilter}
            onChange={setVerifyFilter}
            options={[
              { value: 'all', label: '验证状态：全部' },
              { value: 'success', label: '验证状态：已验证' },
              { value: 'failed', label: '验证状态：失败' },
              { value: 'in_progress', label: '验证状态：进行中' },
              { value: 'pending', label: '验证状态：未验证' },
            ]}
          />
          <Select
            style={{ width: 150 }}
            value={strictFilter}
            onChange={setStrictFilter}
            options={[
              { value: 'all', label: 'Strict：全部' },
              { value: 'routable', label: 'Strict：可路由' },
              { value: 'unmet', label: 'Strict：未满足' },
            ]}
          />
          <Button onClick={resetFilters}>重置筛选</Button>
        </Space>
        <Table
          columns={columns}
          dataSource={filteredApiKeys}
          rowKey="id"
          loading={isLoading}
          pagination={false}
          scroll={{ x: 'max-content' }}
        />
      </Card>

      <Modal
        title={editingKey ? '编辑API密钥' : '添加API密钥'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        okText="保存"
        cancelText="取消"
        width={700}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="name"
            label="密钥名称"
            rules={[{ required: true, message: '请输入密钥名称' }]}
          >
            <Input placeholder="例如：生产环境密钥" disabled={!!editingKey} />
          </Form.Item>

          {!editingKey && (
            <>
              <Spin spinning={providersLoading}>
                <Form.Item
                  name="provider"
                  label="提供商"
                  rules={[{ required: true, message: '请选择提供商' }]}
                >
                  <Select
                    placeholder="请选择提供商"
                    allowClear
                    showSearch
                    optionFilterProp="label"
                    options={modelProviders.map((p) => ({
                      value: p.code,
                      label: p.name,
                    }))}
                    notFoundContent={providersLoading ? '加载中…' : '暂无可用提供商'}
                  />
                </Form.Item>
              </Spin>
              <Form.Item
                name="api_key"
                label="API Key"
                rules={[{ required: true, message: '请输入API Key' }]}
              >
                <Input.Password placeholder="请输入API Key" />
              </Form.Item>
              <Form.Item name="api_secret" label="API Secret">
                <Input.Password placeholder="请输入API Secret（可选）" />
              </Form.Item>
            </>
          )}

          <Form.Item name="endpoint_url" label="端点URL">
            <Input placeholder="自定义端点URL（可选，留空使用默认）" />
          </Form.Item>

          <Form.Item name="unlimited_quota" valuePropName="checked" initialValue={true}>
            <Checkbox>无配额上限（quota_limit 为空；代理仍要求上游 Key 有效）</Checkbox>
          </Form.Item>
          <Form.Item
            noStyle
            shouldUpdate={(prev, cur) => prev.unlimited_quota !== cur.unlimited_quota}
          >
            {({ getFieldValue }) =>
              !getFieldValue('unlimited_quota') ? (
                <Form.Item
                  name="quota_limit"
                  label="配额上限（元）"
                  rules={[
                    { required: true, message: '请填写限额' },
                    { type: 'number', min: 0.01, message: '须大于 0' },
                  ]}
                >
                  <InputNumber
                    min={0.01}
                    precision={2}
                    style={{ width: '100%' }}
                    placeholder="平台侧该商户 Key 用量上限"
                  />
                </Form.Item>
              ) : null
            }
          </Form.Item>

          <Form.Item name="health_check_level" label="主动探测频率">
            <Select placeholder="选择主动探测频率（影响调度器与非强制探测）">
              <Select.Option value="high">高频（约每1分钟）</Select.Option>
              <Select.Option value="medium">中频（约每5分钟）</Select.Option>
              <Select.Option value="low">低频（约每30分钟）</Select.Option>
              <Select.Option value="daily">每日一次</Select.Option>
            </Select>
          </Form.Item>

          <Divider>智能路由配置</Divider>

          <Form.Item name="byok_type" label="BYOK类型">
            <Select placeholder="选择 API Key 来源类型">
              <Select.Option value="official">官方（官方渠道获取）</Select.Option>
              <Select.Option value="reseller">代理商（代理商渠道获取）</Select.Option>
              <Select.Option value="self_hosted">自建商（自建服务）</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item name="region" label="区域">
            <Select placeholder="选择 API Key 区域（用于智能路由）">
              <Select.Option value="domestic">国内</Select.Option>
              <Select.Option value="overseas">海外</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item name="security_level" label="安全等级">
            <Select placeholder="选择安全等级（用于智能路由）">
              <Select.Option value="standard">标准</Select.Option>
              <Select.Option value="high">高安全</Select.Option>
            </Select>
          </Form.Item>

          <Divider>成本定价配置</Divider>

          <Form.Item name="cost_input_rate" label="输入成本率（元/1K tokens）">
            <InputNumber
              min={0}
              precision={6}
              style={{ width: '100%' }}
              placeholder="输入token成本"
            />
          </Form.Item>

          <Form.Item name="cost_output_rate" label="输出成本率（元/1K tokens）">
            <InputNumber
              min={0}
              precision={6}
              style={{ width: '100%' }}
              placeholder="输出token成本"
            />
          </Form.Item>

          <Form.Item name="profit_margin" label="利润率（%）">
            <InputNumber
              min={0}
              max={100}
              precision={2}
              style={{ width: '100%' }}
              placeholder="利润率百分比"
            />
          </Form.Item>

          {editingKey && (
            <Form.Item name="status" label="状态">
              <Select>
                <Select.Option value="active">启用</Select.Option>
                <Select.Option value="inactive">禁用</Select.Option>
              </Select>
            </Form.Item>
          )}
        </Form>
      </Modal>

      <Modal
        title="深度验证 - 探测模型选择"
        open={probeModelModalVisible}
        onCancel={() => setProbeModelModalVisible(false)}
        width={480}
        footer={[
          <Button key="default" onClick={skipProbeModel}>
            使用默认模型
          </Button>,
          <Button key="confirm" type="primary" onClick={confirmProbeModel}>
            开始深度验证
          </Button>,
        ]}
      >
        <div style={{ marginBottom: 16 }}>
          <p style={{ marginBottom: 8, color: 'rgba(0,0,0,0.65)' }}>
            选择用于配额探测的模型，不选择将使用默认模型。
          </p>
          <Select
            style={{ width: '100%' }}
            placeholder="选择探测模型（可选）"
            allowClear
            showSearch
            value={selectedProbeModel}
            onChange={(val) => setSelectedProbeModel(val)}
            options={(cachedModels.get(probeModelTargetId || 0) || []).map(m => ({
              label: m,
              value: m,
            }))}
            notFoundContent="暂无模型列表，将使用默认模型"
          />
        </div>
      </Modal>

      <Modal
        title="API Key 验证"
        open={verificationModalVisible}
        onCancel={() => {
          setVerificationModalVisible(false);
          setVerificationResult(null);
        }}
        footer={null}
        width={600}
      >
        {verificationLoading && (
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <Spin size="large" />
            <p style={{ marginTop: 16 }}>正在验证 API Key...</p>
          </div>
        )}

        {verificationResult && (
          <Descriptions bordered column={1}>
            <Descriptions.Item label="验证状态">
              {verificationResult.status === 'success' ? (
                <Tag color="success" icon={<CheckCircleOutlined />}>
                  验证成功
                </Tag>
              ) : verificationResult.status === 'failed' ? (
                <Tag color="error" icon={<ExclamationCircleOutlined />}>
                  验证失败
                </Tag>
              ) : (
                <Tag color="processing" icon={<SyncOutlined spin />}>
                  验证中
                </Tag>
              )}
            </Descriptions.Item>

            <Descriptions.Item label="连接测试">
              {verificationResult.connection_test ? (
                <Tag color="success">成功 ({verificationResult.connection_latency_ms}ms)</Tag>
              ) : (
                <Tag color="error">失败</Tag>
              )}
            </Descriptions.Item>

            {verificationResult.models_found && verificationResult.models_found.length > 0 && (
              <Descriptions.Item label="支持的模型">
                <Space wrap>
                  {verificationResult.models_found.map((model) => (
                    <Tag key={model}>{model}</Tag>
                  ))}
                </Space>
              </Descriptions.Item>
            )}

            <Descriptions.Item label="定价验证">
              {verificationResult.pricing_verified ? (
                <Tag color="success">已验证</Tag>
              ) : (
                <Tag color="warning">未验证</Tag>
              )}
            </Descriptions.Item>

            {verificationResult.error_code && (
              <Descriptions.Item label="错误码">
                <Tag color="volcano">{verificationResult.error_code}</Tag>
              </Descriptions.Item>
            )}

            {verificationResult.error_message && (
              <Descriptions.Item label="错误信息">
                <span style={{ color: '#ff4d4f' }}>{verificationResult.error_message}</span>
              </Descriptions.Item>
            )}

            <Descriptions.Item label="重试次数">{verificationResult.retry_count}</Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  );
};

export default MerchantAPIKeys;
