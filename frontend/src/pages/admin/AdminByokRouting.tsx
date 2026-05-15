import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Card,
  Table,
  Button,
  Tag,
  Space,
  Modal,
  message,
  Typography,
  Form,
  Select,
  Input,
  Tooltip,
  Spin,
  Switch,
  Checkbox,
  Radio,
  Descriptions,
  Result,
  Divider,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  ReloadOutlined,
  SettingOutlined,
  ThunderboltOutlined,
  SafetyCertificateOutlined,
  ApiOutlined,
  SearchOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  SyncOutlined,
  CopyOutlined,
} from '@ant-design/icons';
import {
  adminByokRoutingService,
  BYOKRoutingItem,
  UpdateRouteConfigRequest,
  VerificationResult,
  CapabilityProbeRow,
  CapabilityProbeRequest,
} from '@/services/adminByokRouting';
import {
  getRouteModeLabel,
  getRouteModeColor,
  getErrorCategoryLabel,
  getErrorCategoryColor,
} from '@/utils/byokRouteMode';
import { copyToClipboard } from '@/utils/clipboard';
import styles from './AdminByokRouting.module.css';

const { Title, Text } = Typography;

/** 部署机 docker 内跑 capability-probe 的示例命令（与 documentation/capability/README.md 一致）。 */
function buildCapabilityProbeCLI(keyID: number): string {
  return [
    '# 默认非计费（与 Admin 弹窗「仅非 chat 三项」接近）；在部署机 backend 容器内执行：',
    `docker exec pintuotuo-backend /app/capability-probe -out /tmp/cap-key-${keyID}.csv -api-key-id ${keyID} -limit 1`,
    '',
    '# 计费类（等同 Admin 弹窗打开「计费类探测」或手动执行）：',
    `# docker exec pintuotuo-backend /app/capability-probe -out /tmp/cap-billable-${keyID}.csv -api-key-id ${keyID} -billable -limit 1`,
  ].join('\n');
}

const ANTHROPIC_SIBLING_SETUP_DOC =
  'https://github.com/sev7n4/pintuotuo/blob/main/documentation/capability/anthropic-sibling-provider-setup.md';

/** Admin 各处理模型列表时的统一说明（与后端 probe-models / FullVerification 一致） */
const BYOK_PROBE_MODELS_DESCRIPTION =
  '与轻量验证、能力探测、深度验证模型选择同源：GET /admin/byok-routing/:id/probe-models。OpenAI 格式拉取 /models；api_format=anthropic 或 provider 以 _anthropic 结尾时走 Messages 探测（与商户端验证相同）。';

/** 验证结果里「使用端点」与模型目录 probe 路径的区别说明 */
const BYOK_ENDPOINT_USED_NOTE =
  '为本次验证连接/配额阶段记录的实际请求端点；与上栏「发现的模型」所用 probe-models 解析路径可能不同（例如 LiteLLM 流量走网关、模型目录走已配置的直连 upstream）。';

const byokTypeTag = (byokType: string) => {
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
  return <Tag color={colorMap[byokType] || 'default'}>{labelMap[byokType] || byokType}</Tag>;
};

const routeModeTag = (routeMode: string) => {
  const colorMap: Record<string, string> = {
    auto: 'cyan',
    direct: 'green',
    litellm: 'blue',
    proxy: 'purple',
  };
  const labelMap: Record<string, string> = {
    auto: '自动',
    direct: '直连',
    litellm: 'LiteLLM',
    proxy: '代理',
  };
  return <Tag color={colorMap[routeMode] || 'default'}>{labelMap[routeMode] || routeMode}</Tag>;
};

const regionTag = (region: string) => {
  const colorMap: Record<string, string> = {
    domestic: 'green',
    overseas: 'blue',
  };
  const labelMap: Record<string, string> = {
    domestic: '国内',
    overseas: '海外',
  };
  return <Tag color={colorMap[region] || 'default'}>{labelMap[region] || region}</Tag>;
};

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
    return `${base}可作为路由候选。`;
  }
  if (s === 'unhealthy') {
    return `${base}请检查上游 Key 或网络。`;
  }
  return `${base}尚未探测或仍为初始值。`;
}

function verificationDotClass(result?: string): string {
  const r = (result || '').toLowerCase();
  if (r === 'verified') return styles.statusDotVerified;
  if (r === 'suspend') return styles.statusDotSuspend;
  if (r === 'unreachable') return styles.statusDotUnreachable;
  if (r === 'invalid') return styles.statusDotInvalid;
  if (r === 'failed') return styles.statusDotVerifyFailed;
  if (r === 'in_progress') return styles.statusDotInProgress;
  return styles.statusDotVerifyPending;
}

function verificationLabel(result?: string): string {
  const r = (result || '').toLowerCase();
  if (r === 'verified') return '验证通过';
  if (r === 'suspend') return '余额不足';
  if (r === 'unreachable') return '连接失败';
  if (r === 'invalid') return '认证失败';
  if (r === 'failed') return '验证失败';
  if (r === 'in_progress') return '验证中';
  return '待验证';
}

function verificationTooltipDesc(result?: string): string {
  const r = (result || '').toLowerCase();
  const base = `验证结果：${verificationLabel(result)}。`;
  if (r === 'verified') {
    return `${base}已通过深度验证，可作为路由候选。`;
  }
  if (r === 'suspend') {
    return `${base}请充值后重新验证。`;
  }
  if (r === 'unreachable') {
    return `${base}请检查网络或端点配置。`;
  }
  if (r === 'invalid') {
    return `${base}请更换 API Key。`;
  }
  if (r === 'failed') {
    return `${base}请查看详情并修正问题。`;
  }
  if (r === 'in_progress') {
    return `${base}正在验证中，请稍后刷新。`;
  }
  return `${base}请完成深度验证。`;
}

interface OperationResult {
  type: 'probe' | 'verify' | 'config';
  status: 'success' | 'failed' | 'loading';
  message: string;
  details?: string;
  timestamp: Date;
}

const AdminByokRouting = () => {
  const [data, setData] = useState<BYOKRoutingItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);

  const [byokTypeFilter, setByokTypeFilter] = useState<string>('');
  const [providerFilter, setProviderFilter] = useState<string>('');
  const [regionFilter, setRegionFilter] = useState<string>('');
  const [routeModeFilter, setRouteModeFilter] = useState<string>('');
  const [healthFilter, setHealthFilter] = useState<string>('');
  const [keywordFilter, setKeywordFilter] = useState<string>('');

  const [configModalVisible, setConfigModalVisible] = useState(false);
  const [configLoading, setConfigLoading] = useState(false);
  const [selectedItem, setSelectedItem] = useState<BYOKRoutingItem | null>(null);
  const [configForm] = Form.useForm();

  const [resultModalVisible, setResultModalVisible] = useState(false);
  const [operationResults, setOperationResults] = useState<Map<number, OperationResult>>(new Map());

  const [verificationModalVisible, setVerificationModalVisible] = useState(false);
  const [verificationResult, setVerificationResult] = useState<VerificationResult | null>(null);
  const [verificationLoading, setVerificationLoading] = useState(false);
  const [verificationContextKeyId, setVerificationContextKeyId] = useState<number | null>(null);
  const [verificationProbeModels, setVerificationProbeModels] = useState<string[] | null>(null);
  const [verificationProbeModelsLoading, setVerificationProbeModelsLoading] = useState(false);
  const [verificationProbeModelsError, setVerificationProbeModelsError] = useState<
    string | undefined
  >(undefined);

  const [probeModelModalVisible, setProbeModelModalVisible] = useState(false);
  const [probeModelTarget, setProbeModelTarget] = useState<BYOKRoutingItem | null>(null);
  const [selectedProbeModel, setSelectedProbeModel] = useState<string | undefined>(undefined);
  const [probeModalModels, setProbeModalModels] = useState<string[]>([]);
  const [probeModalModelsLoading, setProbeModalModelsLoading] = useState(false);

  const [capabilityModalVisible, setCapabilityModalVisible] = useState(false);
  const [capabilityTarget, setCapabilityTarget] = useState<BYOKRoutingItem | null>(null);
  const [capabilityRows, setCapabilityRows] = useState<CapabilityProbeRow[]>([]);
  const [capabilityLoading, setCapabilityLoading] = useState(false);
  const [capabilitySkipEmbeddings, setCapabilitySkipEmbeddings] = useState(false);
  const [capabilityBillable, setCapabilityBillable] = useState(false);
  const [capabilityEndpointPickMode, setCapabilityEndpointPickMode] = useState<
    'single' | 'multiple'
  >('multiple');
  const [capabilitySingleProbeId, setCapabilitySingleProbeId] = useState<string>('embeddings');
  const [capabilityProbeIds, setCapabilityProbeIds] = useState<string[]>([
    'embeddings',
    'moderations',
    'responses',
  ]);
  const [capabilityEmbeddingModel, setCapabilityEmbeddingModel] = useState<string | undefined>();
  const [capabilityModerationModel, setCapabilityModerationModel] = useState<string | undefined>();
  const [capabilityResponsesModel, setCapabilityResponsesModel] = useState<string | undefined>();
  const [capabilityChatModel, setCapabilityChatModel] = useState<string | undefined>();
  const [capabilityFetchedModels, setCapabilityFetchedModels] = useState<string[]>([]);
  const [capabilityProbeModelsLoading, setCapabilityProbeModelsLoading] = useState(false);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const response = await adminByokRoutingService.getByokRoutingList({
        byok_type: byokTypeFilter || undefined,
        provider: providerFilter || undefined,
        region: regionFilter || undefined,
        route_mode: routeModeFilter || undefined,
        health_status: healthFilter || undefined,
      });
      let items = response.data.data || [];
      if (keywordFilter.trim()) {
        const kw = keywordFilter.trim().toLowerCase();
        items = items.filter(
          (item) =>
            item.name.toLowerCase().includes(kw) ||
            item.company_name.toLowerCase().includes(kw) ||
            item.provider.toLowerCase().includes(kw)
        );
      }
      setData(items);
      setTotal(response.data.total);
    } catch {
      message.error('获取BYOK路由列表失败');
    } finally {
      setLoading(false);
    }
  }, [byokTypeFilter, providerFilter, regionFilter, routeModeFilter, healthFilter, keywordFilter]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  /** 与后端 GET .../probe-models（同轻量验证模型目录请求）一致，不合并列表缓存或 models_supported */
  const capabilityModelOptions = useMemo(() => {
    if (!capabilityTarget) return [];
    const uniq = [
      ...new Set(capabilityFetchedModels.filter((m) => Boolean(m && String(m).trim()))),
    ].sort((a, b) => a.localeCompare(b));
    return uniq.map((m) => ({ label: m, value: m }));
  }, [capabilityFetchedModels, capabilityTarget]);

  useEffect(() => {
    if (!capabilityModalVisible || !capabilityTarget) {
      return;
    }
    let cancelled = false;
    (async () => {
      setCapabilityProbeModelsLoading(true);
      try {
        const res = await adminByokRoutingService.getProbeModels(capabilityTarget.id);
        if (cancelled) return;
        setCapabilityFetchedModels(res.data.models || []);
        if (!res.data.success && (res.data.hint || res.data.error_message)) {
          message.warning((res.data.hint || res.data.error_message) as string);
        }
      } catch {
        if (!cancelled) {
          message.error('拉取 /v1/models 模型列表失败');
          setCapabilityFetchedModels([]);
        }
      } finally {
        if (!cancelled) {
          setCapabilityProbeModelsLoading(false);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [capabilityModalVisible, capabilityTarget]);

  useEffect(() => {
    if (!verificationModalVisible || verificationContextKeyId == null || !verificationResult) {
      return;
    }
    if (verificationResult.status === 'pending' || verificationResult.status === 'in_progress') {
      setVerificationProbeModels(null);
      setVerificationProbeModelsError(undefined);
      setVerificationProbeModelsLoading(false);
      return;
    }
    let cancelled = false;
    (async () => {
      setVerificationProbeModelsLoading(true);
      setVerificationProbeModelsError(undefined);
      try {
        const res = await adminByokRoutingService.getProbeModels(verificationContextKeyId);
        if (cancelled) return;
        const raw = res.data.models || [];
        const list = [...new Set(raw.filter((m) => Boolean(m && String(m).trim())))].sort((a, b) =>
          a.localeCompare(b)
        );
        setVerificationProbeModels(list);
        if (!res.data.success && (res.data.hint || res.data.error_message)) {
          message.warning((res.data.hint || res.data.error_message) as string);
        }
      } catch {
        if (!cancelled) {
          setVerificationProbeModels([]);
          setVerificationProbeModelsError(
            '拉取 probe-models 失败，无法展示与能力探测一致的模型列表'
          );
        }
      } finally {
        if (!cancelled) {
          setVerificationProbeModelsLoading(false);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [verificationModalVisible, verificationContextKeyId, verificationResult]);

  const providerOptions = useMemo(() => {
    const providers = new Set(data.map((item) => item.provider));
    return Array.from(providers).sort();
  }, [data]);

  const setOperationResult = (id: number, result: OperationResult) => {
    setOperationResults((prev) => {
      const newMap = new Map(prev);
      newMap.set(id, result);
      return newMap;
    });
  };

  const clearOperationResult = (id: number) => {
    setOperationResults((prev) => {
      const newMap = new Map(prev);
      newMap.delete(id);
      return newMap;
    });
  };

  const handleOpenConfig = (record: BYOKRoutingItem) => {
    setSelectedItem(record);
    const routeConfig = record.route_config || {};
    const endpoints = (routeConfig.endpoints || {}) as Record<string, unknown>;

    const litellmEndpoints = (endpoints.litellm || {}) as Record<string, unknown>;
    const proxyEndpoints = (endpoints.proxy || {}) as Record<string, unknown>;

    configForm.setFieldsValue({
      route_mode: record.route_mode || 'auto',
      endpoint_url: record.endpoint_url || '',
      fallback_endpoint_url: record.fallback_endpoint_url || '',
      litellm_domestic: (litellmEndpoints.domestic as string) || '',
      litellm_overseas: (litellmEndpoints.overseas as string) || '',
      proxy_url: (routeConfig.proxy_url as string) || '',
      proxy_gaap: (proxyEndpoints.gaap as string) || '',
    });
    setConfigModalVisible(true);
  };

  const handleSaveConfig = async () => {
    if (!selectedItem) return;
    const values = await configForm.validateFields();
    setConfigLoading(true);
    setOperationResult(selectedItem.id, {
      type: 'config',
      status: 'loading',
      message: '正在保存配置...',
      timestamp: new Date(),
    });
    try {
      const routeConfig: Record<string, unknown> = {};
      const endpoints: Record<string, unknown> = {};

      if (values.litellm_domestic || values.litellm_overseas) {
        const litellmEndpoints: Record<string, string> = {};
        if (values.litellm_domestic?.trim()) {
          litellmEndpoints.domestic = values.litellm_domestic.trim();
        }
        if (values.litellm_overseas?.trim()) {
          litellmEndpoints.overseas = values.litellm_overseas.trim();
        }
        if (Object.keys(litellmEndpoints).length > 0) {
          endpoints.litellm = litellmEndpoints;
        }
      }

      if (values.proxy_gaap?.trim()) {
        endpoints.proxy = { gaap: values.proxy_gaap.trim() };
      }

      if (Object.keys(endpoints).length > 0) {
        routeConfig.endpoints = endpoints;
      }

      if (values.proxy_url?.trim()) {
        routeConfig.proxy_url = values.proxy_url.trim();
      }

      const payload: UpdateRouteConfigRequest = {
        route_mode: values.route_mode,
        endpoint_url: values.endpoint_url?.trim() || '',
        fallback_endpoint_url: values.fallback_endpoint_url?.trim() || '',
        route_config: Object.keys(routeConfig).length > 0 ? routeConfig : undefined,
      };
      await adminByokRoutingService.updateRouteConfig(selectedItem.id, payload);
      setOperationResult(selectedItem.id, {
        type: 'config',
        status: 'success',
        message: '路由配置更新成功',
        timestamp: new Date(),
      });
      setConfigModalVisible(false);
      fetchData();
      setTimeout(() => clearOperationResult(selectedItem.id), 5000);
    } catch (err) {
      setOperationResult(selectedItem.id, {
        type: 'config',
        status: 'failed',
        message: '更新路由配置失败',
        details: String(err),
        timestamp: new Date(),
      });
    } finally {
      setConfigLoading(false);
    }
  };

  const handleTriggerProbe = async (record: BYOKRoutingItem) => {
    setOperationResult(record.id, {
      type: 'probe',
      status: 'loading',
      message: '正在触发探测...',
      timestamp: new Date(),
    });
    try {
      await adminByokRoutingService.triggerProbe(record.id);
      setOperationResult(record.id, {
        type: 'probe',
        status: 'success',
        message: '探测已触发',
        details: '请稍后刷新查看健康状态更新',
        timestamp: new Date(),
      });
      setSelectedItem(record);
      setResultModalVisible(true);
      setTimeout(() => fetchData(), 3000);
    } catch (err) {
      setOperationResult(record.id, {
        type: 'probe',
        status: 'failed',
        message: '触发探测失败',
        details: String(err),
        timestamp: new Date(),
      });
      setSelectedItem(record);
      setResultModalVisible(true);
    }
  };

  const handleLightVerify = async (record: BYOKRoutingItem) => {
    setVerificationContextKeyId(record.id);
    setVerificationProbeModels(null);
    setVerificationProbeModelsError(undefined);
    setVerificationModalVisible(true);
    setVerificationLoading(true);
    setVerificationResult(null);

    try {
      await adminByokRoutingService.triggerLightVerify(record.id);
      message.success('轻量验证已启动，正在后台执行...');
      await pollVerificationResult(record.id);
    } catch (error) {
      message.error('启动轻量验证失败');
      setVerificationLoading(false);
    }
  };

  const handleDeepVerify = async (record: BYOKRoutingItem, probeModel?: string) => {
    setVerificationContextKeyId(record.id);
    setVerificationProbeModels(null);
    setVerificationProbeModelsError(undefined);
    setVerificationModalVisible(true);
    setVerificationLoading(true);
    setVerificationResult(null);

    try {
      await adminByokRoutingService.triggerDeepVerify(record.id, probeModel);
      message.success('深度验证已启动，正在后台执行...');
      await pollVerificationResult(record.id);
    } catch (error) {
      message.error('启动深度验证失败');
      setVerificationLoading(false);
    }
  };

  const fetchProbeModelsForModal = async (record: BYOKRoutingItem) => {
    setProbeModalModelsLoading(true);
    try {
      const res = await adminByokRoutingService.getProbeModels(record.id);
      const list = res.data.models || [];
      setProbeModalModels(list);
      setSelectedProbeModel(list.length > 0 ? list[0] : undefined);
      if (!res.data.success && (res.data.hint || res.data.error_message)) {
        message.warning((res.data.hint || res.data.error_message) as string);
      }
    } catch {
      message.error('拉取模型列表失败');
      setProbeModalModels([]);
      setSelectedProbeModel(undefined);
    } finally {
      setProbeModalModelsLoading(false);
    }
  };

  const openProbeModelSelector = (record: BYOKRoutingItem) => {
    setProbeModelTarget(record);
    setProbeModalModels([]);
    setSelectedProbeModel(undefined);
    setProbeModelModalVisible(true);
    void fetchProbeModelsForModal(record);
  };

  const confirmProbeModel = () => {
    setProbeModelModalVisible(false);
    if (probeModelTarget) {
      handleDeepVerify(probeModelTarget, selectedProbeModel);
    }
  };

  const skipProbeModel = () => {
    setProbeModelModalVisible(false);
    if (probeModelTarget) {
      handleDeepVerify(probeModelTarget);
    }
  };

  const openCapabilityModal = (record: BYOKRoutingItem) => {
    setCapabilityTarget(record);
    setCapabilityEndpointPickMode('multiple');
    setCapabilitySingleProbeId('embeddings');
    setCapabilitySkipEmbeddings(false);
    setCapabilityProbeIds(['embeddings', 'moderations', 'responses']);
    setCapabilityEmbeddingModel(undefined);
    setCapabilityModerationModel(undefined);
    setCapabilityResponsesModel(undefined);
    setCapabilityChatModel(undefined);
    setCapabilityFetchedModels([]);
    setCapabilityBillable(false);
    setCapabilityRows([]);
    setCapabilityModalVisible(true);
  };

  const runCapabilityFromModal = () => {
    if (!capabilityTarget) return;
    const probesForRequest =
      capabilityEndpointPickMode === 'single' ? [capabilitySingleProbeId] : capabilityProbeIds;
    if (probesForRequest.length === 0) {
      message.warning('请至少选择一个非 chat 端点');
      return;
    }
    if (
      capabilityEndpointPickMode === 'single' &&
      capabilitySkipEmbeddings &&
      capabilitySingleProbeId === 'embeddings'
    ) {
      message.warning('已开启「跳过 embeddings」，请改选 moderations 或 responses');
      return;
    }

    const execute = async () => {
      setCapabilityLoading(true);
      try {
        const body: CapabilityProbeRequest = {
          skip_embeddings: capabilitySkipEmbeddings,
          probes: probesForRequest,
        };
        if (capabilityBillable) body.billable = true;
        if (capabilityEmbeddingModel) body.embedding_model = capabilityEmbeddingModel;
        if (capabilityModerationModel) body.moderation_model = capabilityModerationModel;
        if (capabilityResponsesModel) body.responses_model = capabilityResponsesModel;
        if (capabilityChatModel) body.chat_model = capabilityChatModel;
        const res = await adminByokRoutingService.runCapabilityProbe(capabilityTarget.id, body);
        setCapabilityRows(res.data.rows || []);
        message.success(
          capabilityBillable ? '能力探测完成（含计费类极小请求）' : '能力探测完成（非 chat 矩阵）'
        );
      } catch {
        message.error('能力探测失败或超时，请稍后重试');
      } finally {
        setCapabilityLoading(false);
      }
    };

    if (capabilityBillable) {
      Modal.confirm({
        title: '确认计费类探测',
        content:
          '将额外发起 chat completions、图生、语音/转写等极小 POST（与 CLI -billable 同一路径），可能产生上游费用；与深度验证类似，请仅在环境可接受时执行。可在下方「Chat completions」下拉中指定 chat 模型（可选，不选则默认 gpt-4o-mini）。',
        okText: '确认执行',
        cancelText: '取消',
        onOk: () => execute(),
      });
    } else {
      void execute();
    }
  };

  const pollVerificationResult = async (id: number) => {
    const maxAttempts = 30;
    const interval = 2000;
    let attempts = 0;

    const poll = async (): Promise<void> => {
      try {
        const response = await adminByokRoutingService.getVerificationDetails(id);
        const history = response.data.history;
        const latest = history?.[0];

        if (latest) {
          setVerificationResult(latest);

          if (latest.status === 'pending' || latest.status === 'in_progress') {
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
            fetchData();
          }
        } else {
          attempts++;
          if (attempts < maxAttempts) {
            await new Promise((resolve) => setTimeout(resolve, interval));
            await poll();
          } else {
            message.error('获取验证结果失败');
            setVerificationLoading(false);
          }
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

  const resetFilters = () => {
    setByokTypeFilter('');
    setProviderFilter('');
    setRegionFilter('');
    setRouteModeFilter('');
    setHealthFilter('');
    setKeywordFilter('');
  };

  const renderOperationButton = (
    record: BYOKRoutingItem,
    type: 'probe' | 'verify' | 'deep-verify'
  ) => {
    const result = operationResults.get(record.id);
    const isLoading = result?.status === 'loading' && result?.type === type;
    const isSuccess = result?.status === 'success' && result?.type === type;
    const isFailed = result?.status === 'failed' && result?.type === type;

    let btnClass = '';
    if (isSuccess) btnClass = styles.actionBtnSuccess;
    if (isFailed) btnClass = styles.actionBtnError;

    if (type === 'probe') {
      return (
        <Tooltip
          title={
            result ? `${result.message} (${result.timestamp.toLocaleTimeString()})` : '立即探测'
          }
        >
          <Button
            size="small"
            icon={isLoading ? <SyncOutlined spin /> : <ThunderboltOutlined />}
            onClick={() => handleTriggerProbe(record)}
            className={btnClass}
            loading={isLoading}
          />
        </Tooltip>
      );
    }

    if (type === 'deep-verify') {
      return (
        <Tooltip
          title={
            result
              ? `${result.message} (${result.timestamp.toLocaleTimeString()})`
              : '深度验证（包含配额探测）'
          }
        >
          <Button
            size="small"
            icon={isLoading ? <SyncOutlined spin /> : <SafetyCertificateOutlined />}
            onClick={() => openProbeModelSelector(record)}
            loading={isLoading}
          />
        </Tooltip>
      );
    }

    return (
      <Tooltip
        title={result ? `${result.message} (${result.timestamp.toLocaleTimeString()})` : '轻量验证'}
      >
        <Button
          size="small"
          icon={isLoading ? <SyncOutlined spin /> : <SafetyCertificateOutlined />}
          onClick={() => handleLightVerify(record)}
          className={btnClass}
          loading={isLoading}
        />
      </Tooltip>
    );
  };

  const columns: ColumnsType<BYOKRoutingItem> = [
    {
      title: '商户',
      dataIndex: 'company_name',
      key: 'company_name',
      width: 120,
      ellipsis: true,
      responsive: ['md'],
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 100,
      ellipsis: true,
    },
    {
      title: '提供商',
      dataIndex: 'provider',
      key: 'provider',
      width: 70,
      render: (provider: string) => <Tag color="blue">{provider.toUpperCase()}</Tag>,
      responsive: ['lg'],
    },
    {
      title: 'BYOK',
      dataIndex: 'byok_type',
      key: 'byok_type',
      width: 80,
      render: (byokType: string) => byokTypeTag(byokType),
      responsive: ['md'],
    },
    {
      title: '区域',
      dataIndex: 'region',
      key: 'region',
      width: 60,
      render: (region: string) => regionTag(region),
      responsive: ['lg'],
    },
    {
      title: '路由模式',
      dataIndex: 'route_mode',
      key: 'route_mode',
      width: 70,
      render: (routeMode: string) => routeModeTag(routeMode),
      responsive: ['md'],
    },
    {
      title: '健康',
      dataIndex: 'health_status',
      key: 'health_status',
      width: 60,
      render: (status: string) => (
        <Tooltip title={healthTooltipDesc(status)}>
          <span className={healthDotClass(status)} />
        </Tooltip>
      ),
    },
    {
      title: '验证',
      dataIndex: 'verification_result',
      key: 'verification_result',
      width: 60,
      render: (result: string) => (
        <Tooltip title={verificationTooltipDesc(result)}>
          <span className={verificationDotClass(result)} />
        </Tooltip>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 60,
      render: (status: string) => (
        <Tag color={status === 'active' ? 'green' : 'red'}>
          {status === 'active' ? '启用' : '禁用'}
        </Tag>
      ),
      responsive: ['lg'],
    },
    {
      title: '操作',
      key: 'actions',
      width: 280,
      fixed: 'right',
      render: (_, record) => (
        <Space size="small" wrap>
          <Tooltip title="路由配置">
            <Button
              size="small"
              icon={<SettingOutlined />}
              onClick={() => handleOpenConfig(record)}
            />
          </Tooltip>
          {renderOperationButton(record, 'probe')}
          {renderOperationButton(record, 'verify')}
          {renderOperationButton(record, 'deep-verify')}
          <Tooltip title="Phase0 能力探测：models + 可选非 chat 三项；可选计费类（chat/图/音，与 CLI -billable 同路径）">
            <Button
              size="small"
              icon={<ApiOutlined />}
              onClick={() => openCapabilityModal(record)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  const renderMobileItem = (record: BYOKRoutingItem) => (
    <div className={styles.mobileItem} key={record.id}>
      <div className={styles.mobileRow}>
        <span className={styles.mobileLabel}>商户</span>
        <span className={styles.mobileValue}>{record.company_name}</span>
      </div>
      <div className={styles.mobileRow}>
        <span className={styles.mobileLabel}>名称</span>
        <span className={styles.mobileValue}>{record.name}</span>
      </div>
      <div className={styles.mobileRow}>
        <span className={styles.mobileLabel}>提供商</span>
        <Tag color="blue">{record.provider.toUpperCase()}</Tag>
      </div>
      <div className={styles.mobileRow}>
        <span className={styles.mobileLabel}>BYOK</span>
        {byokTypeTag(record.byok_type)}
      </div>
      <div className={styles.mobileRow}>
        <span className={styles.mobileLabel}>路由</span>
        {routeModeTag(record.route_mode)}
      </div>
      <div className={styles.mobileRow}>
        <span className={styles.mobileLabel}>健康</span>
        <Tooltip title={healthTooltipDesc(record.health_status)}>
          <span className={healthDotClass(record.health_status)} />
        </Tooltip>
      </div>
      <div className={styles.mobileRow}>
        <span className={styles.mobileLabel}>验证</span>
        <Tooltip title={verificationTooltipDesc(record.verification_result)}>
          <span className={verificationDotClass(record.verification_result)} />
        </Tooltip>
      </div>
      <div className={styles.mobileActions}>
        <Tooltip title="路由配置">
          <Button
            size="small"
            icon={<SettingOutlined />}
            onClick={() => handleOpenConfig(record)}
          />
        </Tooltip>
        {renderOperationButton(record, 'probe')}
        {renderOperationButton(record, 'verify')}
        {renderOperationButton(record, 'deep-verify')}
        <Tooltip title="Phase0 能力探测（可选计费类）">
          <Button size="small" icon={<ApiOutlined />} onClick={() => openCapabilityModal(record)} />
        </Tooltip>
      </div>
    </div>
  );

  return (
    <div className={styles.byokRouting} style={{ padding: 24 }}>
      <Card>
        <div className={styles.header}>
          <Title level={4} className={styles.pageTitle}>
            BYOK 路由管理
          </Title>
          <Button icon={<ReloadOutlined />} onClick={fetchData} loading={loading}>
            刷新
          </Button>
        </div>

        <div className={styles.filterRow}>
          <Space wrap size={12}>
            <Input
              allowClear
              style={{ width: 160 }}
              placeholder="搜索商户/名称/提供商"
              prefix={<SearchOutlined />}
              value={keywordFilter}
              onChange={(e) => setKeywordFilter(e.target.value)}
            />
            <Select
              style={{ width: 110 }}
              value={byokTypeFilter}
              onChange={setByokTypeFilter}
              allowClear
              placeholder="BYOK类型"
              options={[
                { value: 'official', label: '官方' },
                { value: 'reseller', label: '代理商' },
                { value: 'self_hosted', label: '自建商' },
              ]}
            />
            <Select
              style={{ width: 100 }}
              value={providerFilter}
              onChange={setProviderFilter}
              allowClear
              placeholder="提供商"
              options={providerOptions.map((p) => ({
                value: p.toLowerCase(),
                label: p.toUpperCase(),
              }))}
            />
            <Select
              style={{ width: 90 }}
              value={regionFilter}
              onChange={setRegionFilter}
              allowClear
              placeholder="区域"
              options={[
                { value: 'domestic', label: '国内' },
                { value: 'overseas', label: '海外' },
              ]}
            />
            <Select
              style={{ width: 100 }}
              value={healthFilter}
              onChange={setHealthFilter}
              allowClear
              placeholder="健康状态"
              options={[
                { value: 'healthy', label: '健康' },
                { value: 'degraded', label: '降级' },
                { value: 'unhealthy', label: '不健康' },
                { value: 'unknown', label: '未知' },
              ]}
            />
            <Button onClick={resetFilters}>重置</Button>
          </Space>
        </div>

        <div className={styles.desktopTable}>
          <Table
            columns={columns}
            dataSource={data}
            rowKey="id"
            loading={loading}
            scroll={{ x: 1000 }}
            size="small"
            pagination={{
              total,
              pageSize: 20,
              showSizeChanger: false,
              showTotal: (t) => `共 ${t} 条`,
            }}
          />
        </div>

        <div className={styles.mobileCard}>
          {loading ? <Spin /> : data.map((item) => renderMobileItem(item))}
        </div>
      </Card>

      <Modal
        title="路由配置"
        open={configModalVisible}
        onCancel={() => setConfigModalVisible(false)}
        onOk={handleSaveConfig}
        confirmLoading={configLoading}
        width={600}
      >
        <Spin spinning={configLoading}>
          <Form form={configForm} layout="vertical">
            <Form.Item name="route_mode" label="路由模式">
              <Select placeholder="选择路由模式">
                <Select.Option value="auto">自动（系统决策）</Select.Option>
                <Select.Option value="direct">直连（直接访问上游）</Select.Option>
                <Select.Option value="litellm">LiteLLM（通过LiteLLM网关）</Select.Option>
                <Select.Option value="proxy">代理（通过代理访问）</Select.Option>
              </Select>
            </Form.Item>
            <Form.Item name="endpoint_url" label="端点URL" extra="直连模式使用的端点地址">
              <Input placeholder="自定义端点URL（可选）" />
            </Form.Item>
            <Form.Item
              name="fallback_endpoint_url"
              label="备用端点URL"
              extra="主端点不可用时的备用地址"
            >
              <Input placeholder="备用端点URL（可选）" />
            </Form.Item>

            <Divider orientation="left" plain>
              LiteLLM 配置
            </Divider>
            <Form.Item
              name="litellm_domestic"
              label="LiteLLM 国内端点"
              extra="LiteLLM模式国内区域使用的端点"
            >
              <Input placeholder="https://litellm-cn.example.com/v1" />
            </Form.Item>
            <Form.Item
              name="litellm_overseas"
              label="LiteLLM 海外端点"
              extra="LiteLLM模式海外区域使用的端点"
            >
              <Input placeholder="https://litellm-global.example.com/v1" />
            </Form.Item>

            <Divider orientation="left" plain>
              代理配置
            </Divider>
            <Form.Item name="proxy_url" label="代理URL" extra="代理模式使用的通用代理地址">
              <Input placeholder="https://proxy.example.com" />
            </Form.Item>
            <Form.Item name="proxy_gaap" label="GAAP代理端点" extra="代理模式使用的GAAP加速端点">
              <Input placeholder="https://gaap.example.com" />
            </Form.Item>
          </Form>
        </Spin>
      </Modal>

      <Modal
        title="操作结果"
        open={resultModalVisible}
        onCancel={() => setResultModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setResultModalVisible(false)}>
            关闭
          </Button>,
          <Button
            key="refresh"
            type="primary"
            onClick={() => {
              setResultModalVisible(false);
              fetchData();
            }}
          >
            刷新数据
          </Button>,
        ]}
        width={500}
      >
        {selectedItem && operationResults.get(selectedItem.id) && (
          <div className={styles.resultCard}>
            <Result
              icon={
                operationResults.get(selectedItem.id)?.status === 'success' ? (
                  <CheckCircleOutlined style={{ color: '#52c41a' }} />
                ) : (
                  <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />
                )
              }
              title={operationResults.get(selectedItem.id)?.message}
              subTitle={operationResults.get(selectedItem.id)?.details}
            />
            <Divider />
            <Descriptions column={1} size="small">
              <Descriptions.Item label="商户">{selectedItem.company_name}</Descriptions.Item>
              <Descriptions.Item label="API Key">{selectedItem.name}</Descriptions.Item>
              <Descriptions.Item label="提供商">
                {selectedItem.provider.toUpperCase()}
              </Descriptions.Item>
              <Descriptions.Item label="当前健康状态">
                <span className={styles.statusLightRow}>
                  <span className={healthDotClass(selectedItem.health_status)} />
                  <span className={styles.statusLightLabel}>
                    {healthLabel(selectedItem.health_status)}
                  </span>
                </span>
              </Descriptions.Item>
              <Descriptions.Item label="当前验证状态">
                <span className={styles.statusLightRow}>
                  <span className={verificationDotClass(selectedItem.verification_result)} />
                  <span className={styles.statusLightLabel}>
                    {verificationLabel(selectedItem.verification_result)}
                  </span>
                </span>
              </Descriptions.Item>
              {selectedItem.health_error_category && (
                <Descriptions.Item label="健康错误分类">
                  <Tag color={getErrorCategoryColor(selectedItem.health_error_category)}>
                    {getErrorCategoryLabel(selectedItem.health_error_category)}
                  </Tag>
                </Descriptions.Item>
              )}
              {selectedItem.health_error_code && (
                <Descriptions.Item label="健康错误码">
                  <Tag color="volcano">{selectedItem.health_error_code}</Tag>
                </Descriptions.Item>
              )}
              {selectedItem.health_error_message && (
                <Descriptions.Item label="健康错误信息">
                  <span style={{ color: '#ff4d4f' }}>{selectedItem.health_error_message}</span>
                </Descriptions.Item>
              )}
            </Descriptions>
          </div>
        )}
      </Modal>

      <Modal
        title="Phase0 能力探测"
        open={capabilityModalVisible}
        onCancel={() => {
          setCapabilityModalVisible(false);
          setCapabilityTarget(null);
          setCapabilityRows([]);
          setCapabilityFetchedModels([]);
        }}
        footer={null}
        width={920}
        destroyOnClose
      >
        {capabilityTarget && (
          <div>
            <p style={{ marginBottom: 12, color: 'rgba(0,0,0,0.65)' }}>
              商户：{capabilityTarget.company_name} · Key：{capabilityTarget.name}（
              {capabilityTarget.provider.toUpperCase()}）
            </p>
            <Divider orientation="left" plain style={{ margin: '8px 0 12px' }}>
              探测端点
            </Divider>
            <Text
              type="secondary"
              style={{ display: 'block', marginBottom: 8, fontSize: 12, lineHeight: 1.6 }}
            >
              默认仅对下方勾选的 embeddings / moderations / responses 发起非计费
              POST。打开「计费类探测」并确认后，会额外发起与 CLI -billable 同路径的极小请求（chat
              completions、图生、语音/转写等），可能产生上游费用，行为与深度验证类似。每次仍会先跑
              get_models（结果表第一行）。
            </Text>
            <Text type="secondary" style={{ display: 'block', marginBottom: 12, fontSize: 12 }}>
              完整对照表（Admin / CLI / 图音等）：{' '}
              <a
                href="https://github.com/sev7n4/pintuotuo/blob/main/documentation/capability/endpoint-coverage-matrix.md"
                target="_blank"
                rel="noopener noreferrer"
              >
                endpoint-coverage-matrix.md
              </a>
            </Text>
            <Space direction="vertical" size="middle" style={{ marginBottom: 12 }}>
              <div>
                <span style={{ marginRight: 12, color: 'rgba(0,0,0,0.65)' }}>选择方式</span>
                <Radio.Group
                  optionType="button"
                  buttonStyle="solid"
                  value={capabilityEndpointPickMode}
                  onChange={(e) => {
                    const mode = e.target.value as 'single' | 'multiple';
                    setCapabilityEndpointPickMode(mode);
                    if (mode === 'single') {
                      const valid = capabilityProbeIds.filter(
                        (p) => !(capabilitySkipEmbeddings && p === 'embeddings')
                      );
                      const next = valid[0] || 'moderations';
                      setCapabilitySingleProbeId(next);
                    } else {
                      const cur = capabilitySingleProbeId;
                      const all = ['embeddings', 'moderations', 'responses'].filter(
                        (p) => !(capabilitySkipEmbeddings && p === 'embeddings')
                      );
                      setCapabilityProbeIds((prev) => {
                        const merged = new Set<string>(prev.length > 0 ? prev : all);
                        merged.add(cur);
                        return Array.from(merged).filter(
                          (p) => !(capabilitySkipEmbeddings && p === 'embeddings')
                        );
                      });
                    }
                  }}
                  disabled={capabilityLoading}
                >
                  <Radio.Button value="single">单选</Radio.Button>
                  <Radio.Button value="multiple">多选</Radio.Button>
                </Radio.Group>
              </div>
              {capabilityEndpointPickMode === 'single' ? (
                <Radio.Group
                  value={capabilitySingleProbeId}
                  onChange={(e) => setCapabilitySingleProbeId(e.target.value)}
                  disabled={capabilityLoading}
                >
                  <Radio value="embeddings" disabled={capabilitySkipEmbeddings}>
                    embeddings
                  </Radio>
                  <Radio value="moderations">moderations</Radio>
                  <Radio value="responses">responses</Radio>
                </Radio.Group>
              ) : (
                <Checkbox.Group
                  value={capabilityProbeIds}
                  onChange={(v) => setCapabilityProbeIds(v as string[])}
                  disabled={capabilityLoading}
                  options={[
                    {
                      label: 'embeddings',
                      value: 'embeddings',
                      disabled: capabilitySkipEmbeddings,
                    },
                    { label: 'moderations', value: 'moderations' },
                    { label: 'responses', value: 'responses' },
                  ]}
                />
              )}
            </Space>
            <Space wrap style={{ marginBottom: 16 }}>
              <span>跳过 embeddings POST</span>
              <Switch
                checked={capabilitySkipEmbeddings}
                onChange={(checked) => {
                  setCapabilitySkipEmbeddings(checked);
                  if (checked) {
                    setCapabilityProbeIds((prev) => prev.filter((p) => p !== 'embeddings'));
                    setCapabilitySingleProbeId((cur) =>
                      cur === 'embeddings' ? 'moderations' : cur
                    );
                  }
                }}
                disabled={capabilityLoading}
              />
              <Tooltip title="开启后请求 billable=true：额外发起 chat / 图 / 音等极小计费类 POST（与 CLI -billable 同路径）；开始探测前会二次确认。">
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
                  <span>计费类探测</span>
                  <Switch
                    checked={capabilityBillable}
                    onChange={setCapabilityBillable}
                    disabled={capabilityLoading}
                  />
                </span>
              </Tooltip>
              <Button
                type="default"
                onClick={() => {
                  if (!capabilityTarget) return;
                  setCapabilityProbeModelsLoading(true);
                  adminByokRoutingService
                    .getProbeModels(capabilityTarget.id)
                    .then((res) => {
                      setCapabilityFetchedModels(res.data.models || []);
                      if (!res.data.success && (res.data.hint || res.data.error_message)) {
                        message.warning((res.data.hint || res.data.error_message) as string);
                      } else {
                        message.success('已刷新模型列表');
                      }
                    })
                    .catch(() => message.error('刷新失败'))
                    .finally(() => setCapabilityProbeModelsLoading(false));
                }}
                loading={capabilityProbeModelsLoading}
                disabled={capabilityLoading}
              >
                刷新 /v1/models
              </Button>
              <Tooltip title="复制可在部署机执行的 docker exec 命令（含单 Key -api-key-id；计费行为见注释）">
                <Button
                  type="default"
                  icon={<CopyOutlined />}
                  disabled={capabilityLoading || !capabilityTarget}
                  onClick={async () => {
                    if (!capabilityTarget) return;
                    const ok = await copyToClipboard(buildCapabilityProbeCLI(capabilityTarget.id));
                    message[ok ? 'success' : 'error'](ok ? '已复制 CLI 命令' : '复制失败');
                  }}
                >
                  复制 CLI 命令
                </Button>
              </Tooltip>
              <Button type="primary" onClick={runCapabilityFromModal} loading={capabilityLoading}>
                开始探测
              </Button>
              {capabilityRows.length > 0 && (
                <Button
                  onClick={() => {
                    setCapabilityRows([]);
                  }}
                  disabled={capabilityLoading}
                >
                  清空结果
                </Button>
              )}
            </Space>
            <Divider orientation="left" plain style={{ margin: '8px 0 12px' }}>
              各端点模型
            </Divider>
            <Text
              type="secondary"
              style={{ display: 'block', marginBottom: 12, fontSize: 12, lineHeight: 1.6 }}
            >
              {BYOK_PROBE_MODELS_DESCRIPTION}
            </Text>
            <Space direction="vertical" style={{ width: '100%', marginBottom: 16 }} size="middle">
              <div>
                <div style={{ marginBottom: 4, fontSize: 12, color: 'rgba(0,0,0,0.55)' }}>
                  Embeddings
                </div>
                <Select
                  style={{ width: '100%', maxWidth: 480 }}
                  placeholder="默认 text-embedding-3-small"
                  allowClear
                  showSearch
                  value={capabilityEmbeddingModel}
                  onChange={(v) => setCapabilityEmbeddingModel(v)}
                  options={capabilityModelOptions}
                  disabled={capabilityLoading}
                  notFoundContent={
                    capabilityProbeModelsLoading
                      ? '加载中…'
                      : '暂无，请先点「刷新 /v1/models」或检查上游'
                  }
                />
              </div>
              <div>
                <div style={{ marginBottom: 4, fontSize: 12, color: 'rgba(0,0,0,0.55)' }}>
                  Moderations
                </div>
                <Select
                  style={{ width: '100%', maxWidth: 480 }}
                  placeholder="默认 omni-moderation-latest"
                  allowClear
                  showSearch
                  value={capabilityModerationModel}
                  onChange={(v) => setCapabilityModerationModel(v)}
                  options={capabilityModelOptions}
                  disabled={capabilityLoading}
                  notFoundContent={
                    capabilityProbeModelsLoading
                      ? '加载中…'
                      : '暂无，请先点「刷新 /v1/models」或检查上游'
                  }
                />
              </div>
              <div>
                <div style={{ marginBottom: 4, fontSize: 12, color: 'rgba(0,0,0,0.55)' }}>
                  Responses
                </div>
                <Select
                  style={{ width: '100%', maxWidth: 480 }}
                  placeholder="未填时与 Chat 模型或默认 gpt-4o-mini 一致"
                  allowClear
                  showSearch
                  value={capabilityResponsesModel}
                  onChange={(v) => setCapabilityResponsesModel(v)}
                  options={capabilityModelOptions}
                  disabled={capabilityLoading}
                  notFoundContent={
                    capabilityProbeModelsLoading
                      ? '加载中…'
                      : '暂无，请先点「刷新 /v1/models」或检查上游'
                  }
                />
              </div>
              {capabilityBillable && (
                <div>
                  <div style={{ marginBottom: 4, fontSize: 12, color: 'rgba(0,0,0,0.55)' }}>
                    Chat completions（计费类）
                  </div>
                  <Select
                    style={{ width: '100%', maxWidth: 480 }}
                    placeholder="默认 gpt-4o-mini；与深度验证探测模型用法相同"
                    allowClear
                    showSearch
                    value={capabilityChatModel}
                    onChange={(v) => setCapabilityChatModel(v)}
                    options={capabilityModelOptions}
                    disabled={capabilityLoading}
                    notFoundContent={
                      capabilityProbeModelsLoading
                        ? '加载中…'
                        : '暂无，请先点「刷新 /v1/models」或检查上游'
                    }
                  />
                </div>
              )}
            </Space>
            {capabilityLoading && (
              <div style={{ textAlign: 'center', padding: 24 }}>
                <Spin size="large" />
                <p style={{ marginTop: 12 }}>正在请求上游（可能需 1–3 分钟）…</p>
              </div>
            )}
            {capabilityRows.length > 0 && (
              <Table
                size="small"
                pagination={false}
                dataSource={capabilityRows.map((r, i) => ({ ...r, key: i }))}
                columns={[
                  { title: 'probe', dataIndex: 'probe', key: 'probe', width: 220 },
                  { title: 'HTTP', dataIndex: 'http_code', key: 'http_code', width: 72 },
                  { title: 'ok', dataIndex: 'ok', key: 'ok', width: 56 },
                  { title: 'api_format', dataIndex: 'api_format', key: 'api_format', width: 96 },
                  { title: 'route_mode', dataIndex: 'route_mode', key: 'route_mode', width: 100 },
                  { title: 'note', dataIndex: 'note', key: 'note', ellipsis: true },
                ]}
              />
            )}
          </div>
        )}
      </Modal>

      <Modal
        title="深度验证 - 探测模型选择"
        open={probeModelModalVisible}
        onCancel={() => {
          setProbeModelModalVisible(false);
          setProbeModelTarget(null);
          setProbeModalModels([]);
          setSelectedProbeModel(undefined);
        }}
        onOk={confirmProbeModel}
        width={480}
        okText="开始深度验证"
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
          <Text
            type="secondary"
            style={{ display: 'block', marginBottom: 12, fontSize: 12, lineHeight: 1.6 }}
          >
            {BYOK_PROBE_MODELS_DESCRIPTION}{' '}
            <a href={ANTHROPIC_SIBLING_SETUP_DOC} target="_blank" rel="noopener noreferrer">
              XX_anthropic 接入文档
            </a>
          </Text>
          <Space.Compact style={{ width: '100%', marginBottom: 8 }}>
            <Select
              style={{ width: '100%' }}
              placeholder={probeModalModelsLoading ? '正在拉取模型列表…' : '选择探测模型（可选）'}
              allowClear
              showSearch
              loading={probeModalModelsLoading}
              value={selectedProbeModel}
              onChange={(val) => setSelectedProbeModel(val)}
              options={probeModalModels.map((m) => ({ label: m, value: m }))}
              notFoundContent={
                probeModalModelsLoading ? '加载中…' : '暂无，可点击右侧刷新或检查上游'
              }
            />
            <Button
              type="default"
              icon={<SyncOutlined spin={probeModalModelsLoading} />}
              disabled={!probeModelTarget || probeModalModelsLoading}
              onClick={() => {
                if (probeModelTarget) void fetchProbeModelsForModal(probeModelTarget);
              }}
            >
              刷新
            </Button>
          </Space.Compact>
          <Text type="secondary" style={{ fontSize: 12 }}>
            不选择将使用默认模型发起深度验证。
          </Text>
        </div>
      </Modal>

      <Modal
        title="API Key 验证"
        open={verificationModalVisible}
        onCancel={() => {
          setVerificationModalVisible(false);
          setVerificationResult(null);
          setVerificationContextKeyId(null);
          setVerificationProbeModels(null);
          setVerificationProbeModelsError(undefined);
          setVerificationProbeModelsLoading(false);
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

            {verificationResult.route_mode && (
              <Descriptions.Item label="路由模式">
                <Tag color={getRouteModeColor(verificationResult.route_mode)}>
                  {getRouteModeLabel(verificationResult.route_mode)}
                </Tag>
              </Descriptions.Item>
            )}

            {verificationResult.endpoint_used && (
              <Descriptions.Item label="使用端点">
                <Text
                  type="secondary"
                  style={{ display: 'block', marginBottom: 8, fontSize: 12, lineHeight: 1.6 }}
                >
                  {BYOK_ENDPOINT_USED_NOTE}
                </Text>
                <span style={{ fontFamily: 'monospace', fontSize: '12px' }}>
                  {verificationResult.endpoint_used}
                </span>
              </Descriptions.Item>
            )}

            <Descriptions.Item label="发现的模型">
              <Text
                type="secondary"
                style={{ display: 'block', marginBottom: 8, fontSize: 12, lineHeight: 1.6 }}
              >
                {BYOK_PROBE_MODELS_DESCRIPTION}
                {verificationResult.status === 'pending' ||
                verificationResult.status === 'in_progress'
                  ? ' 验证结束后将自动拉取并展示。'
                  : ''}
              </Text>
              {verificationProbeModelsLoading && (
                <div style={{ padding: '8px 0' }}>
                  <Spin size="small" /> <Text type="secondary">正在拉取 probe-models…</Text>
                </div>
              )}
              {verificationProbeModelsError && (
                <Text type="danger" style={{ display: 'block', marginBottom: 8 }}>
                  {verificationProbeModelsError}
                </Text>
              )}
              {!verificationProbeModelsLoading &&
                verificationProbeModels !== null &&
                verificationResult.status !== 'pending' &&
                verificationResult.status !== 'in_progress' && (
                  <>
                    {verificationProbeModels.length > 0 ? (
                      <Space wrap>
                        {verificationProbeModels.map((model) => (
                          <Tag key={model}>{model}</Tag>
                        ))}
                      </Space>
                    ) : (
                      <Text type="secondary">
                        未从上游返回模型 ID（可关闭后使用能力探测「刷新 /v1/models」重试）
                      </Text>
                    )}
                  </>
                )}
            </Descriptions.Item>

            <Descriptions.Item label="定价验证">
              {verificationResult.pricing_verified ? (
                <Tag color="success">已验证</Tag>
              ) : (
                <Tag color="warning">未验证</Tag>
              )}
            </Descriptions.Item>

            {verificationResult.error_category && (
              <Descriptions.Item label="错误分类">
                <Tag color={getErrorCategoryColor(verificationResult.error_category)}>
                  {getErrorCategoryLabel(verificationResult.error_category)}
                </Tag>
              </Descriptions.Item>
            )}

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

export default AdminByokRouting;
