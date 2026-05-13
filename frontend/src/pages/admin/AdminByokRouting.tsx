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
    '# 默认非计费（与 Admin 弹窗矩阵接近）；在部署机 backend 容器内执行：',
    `docker exec pintuotuo-backend /app/capability-probe -out /tmp/cap-key-${keyID}.csv -api-key-id ${keyID} -limit 1`,
    '',
    '# 计费类全矩阵（慎用，见 documentation/capability/phase0-scope.md）：',
    `# docker exec pintuotuo-backend /app/capability-probe -out /tmp/cap-billable-${keyID}.csv -api-key-id ${keyID} -billable -limit 1`,
  ].join('\n');
}

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

  const [probeModelModalVisible, setProbeModelModalVisible] = useState(false);
  const [probeModelTarget, setProbeModelTarget] = useState<BYOKRoutingItem | null>(null);
  const [selectedProbeModel, setSelectedProbeModel] = useState<string | undefined>(undefined);
  const [cachedModels, setCachedModels] = useState<Map<number, string[]>>(new Map());

  const [capabilityModalVisible, setCapabilityModalVisible] = useState(false);
  const [capabilityTarget, setCapabilityTarget] = useState<BYOKRoutingItem | null>(null);
  const [capabilityRows, setCapabilityRows] = useState<CapabilityProbeRow[]>([]);
  const [capabilityLoading, setCapabilityLoading] = useState(false);
  const [capabilitySkipEmbeddings, setCapabilitySkipEmbeddings] = useState(false);
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

  const capabilityModelOptions = useMemo(() => {
    if (!capabilityTarget) return [];
    const fromCache = cachedModels.get(capabilityTarget.id) || [];
    const fromRow = capabilityTarget.models_supported || [];
    const merged = [...new Set([...capabilityFetchedModels, ...fromCache, ...fromRow])];
    return merged.map((m) => ({ label: m, value: m }));
  }, [capabilityFetchedModels, capabilityTarget, cachedModels]);

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

  const openProbeModelSelector = (record: BYOKRoutingItem) => {
    setProbeModelTarget(record);
    const existing = cachedModels.get(record.id) || record.models_supported || [];
    setSelectedProbeModel(existing.length > 0 ? existing[0] : undefined);
    setProbeModelModalVisible(true);
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
    setCapabilityFetchedModels([]);
    setCapabilityRows([]);
    setCapabilityModalVisible(true);
  };

  const runCapabilityFromModal = async () => {
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
    setCapabilityLoading(true);
    try {
      const body: CapabilityProbeRequest = {
        skip_embeddings: capabilitySkipEmbeddings,
        probes: probesForRequest,
      };
      if (capabilityEmbeddingModel) body.embedding_model = capabilityEmbeddingModel;
      if (capabilityModerationModel) body.moderation_model = capabilityModerationModel;
      if (capabilityResponsesModel) body.responses_model = capabilityResponsesModel;
      const res = await adminByokRoutingService.runCapabilityProbe(capabilityTarget.id, body);
      setCapabilityRows(res.data.rows || []);
      message.success('能力探测完成（非 chat 矩阵）');
    } catch {
      message.error('能力探测失败或超时，请稍后重试');
    } finally {
      setCapabilityLoading(false);
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

          if (latest.models_found && latest.models_found.length > 0) {
            setCachedModels((prev) => {
              const next = new Map(prev);
              next.set(id, latest.models_found!);
              return next;
            });
          }

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
          <Tooltip title="Phase0 能力探测（非 chat：models / embeddings / moderations / responses 等）">
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
        <Tooltip title="Phase0 能力探测（非 chat）">
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
        title="Phase0 能力探测（非 chat）"
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
              严格来说「非
              chat」还包含图生（images_*）、语音（audio_*）等；本后台探测固定为不计费模式，只会对下面三项发起真实
              POST。图/音在结果表里会始终为 skipped（billable_disabled）。完整含计费矩阵请用仓库内
              capability-probe CLI 的 -billable。每次探测仍会先跑 get_models（列表里第一行）。
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
              各端点模型（可选，不选则使用服务端默认；选项来自 /v1/models 与已缓存列表）
            </Divider>
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
                    capabilityProbeModelsLoading ? '加载中…' : '暂无列表，可完成深度验证后重试'
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
                  notFoundContent={capabilityProbeModelsLoading ? '加载中…' : '暂无列表'}
                />
              </div>
              <div>
                <div style={{ marginBottom: 4, fontSize: 12, color: 'rgba(0,0,0,0.55)' }}>
                  Responses
                </div>
                <Select
                  style={{ width: '100%', maxWidth: 480 }}
                  placeholder="默认与 chat 探测一致（gpt-4o-mini）"
                  allowClear
                  showSearch
                  value={capabilityResponsesModel}
                  onChange={(v) => setCapabilityResponsesModel(v)}
                  options={capabilityModelOptions}
                  disabled={capabilityLoading}
                  notFoundContent={capabilityProbeModelsLoading ? '加载中…' : '暂无列表'}
                />
              </div>
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
        onCancel={() => setProbeModelModalVisible(false)}
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
            options={(
              cachedModels.get(probeModelTarget?.id || 0) ||
              probeModelTarget?.models_supported ||
              []
            ).map((m) => ({
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

            {verificationResult.route_mode && (
              <Descriptions.Item label="路由模式">
                <Tag color={getRouteModeColor(verificationResult.route_mode)}>
                  {getRouteModeLabel(verificationResult.route_mode)}
                </Tag>
              </Descriptions.Item>
            )}

            {verificationResult.endpoint_used && (
              <Descriptions.Item label="使用端点">
                <span style={{ fontFamily: 'monospace', fontSize: '12px' }}>
                  {verificationResult.endpoint_used}
                </span>
              </Descriptions.Item>
            )}

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
