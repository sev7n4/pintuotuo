import { useEffect, useMemo, useState } from 'react';
import {
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  Tooltip,
  Typography,
  message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { CopyOutlined, LinkOutlined, QrcodeOutlined, StopOutlined } from '@ant-design/icons';
import { adminMerchantService, MerchantInvite } from '@/services/adminMerchant';

const { Text } = Typography;

const INVITE_STATUS_OPTIONS = [
  { value: '', label: '全部状态' },
  { value: 'active', label: '可用' },
  { value: 'used_up', label: '已用尽' },
  { value: 'expired', label: '已过期' },
  { value: 'revoked', label: '已撤销' },
];

const statusTag = (status?: string) => {
  const map: Record<string, { color: string; text: string }> = {
    active: { color: 'green', text: '可用' },
    used_up: { color: 'orange', text: '已用尽' },
    expired: { color: 'gold', text: '已过期' },
    revoked: { color: 'red', text: '已撤销' },
  };
  const m = map[status || ''] ?? { color: 'default', text: status || '-' };
  return <Tag color={m.color}>{m.text}</Tag>;
};

type Props = { visible: boolean };

const InviteManagementPanel = ({ visible }: Props) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [invites, setInvites] = useState<MerchantInvite[]>([]);
  const [keyword, setKeyword] = useState('');
  const [status, setStatus] = useState('');
  const [qrUrl, setQrUrl] = useState('');
  const [revokeReason, setRevokeReason] = useState('');
  const [revokeTarget, setRevokeTarget] = useState<MerchantInvite | null>(null);

  const load = async () => {
    setLoading(true);
    try {
      const { data } = await adminMerchantService.listMerchantInvites({
        keyword: keyword.trim() || undefined,
        status: status || undefined,
        limit: 300,
      });
      setInvites(data.data || []);
    } catch {
      message.error('获取邀请码列表失败');
    } finally {
      setLoading(false);
    }
  };

  const onCreate = async () => {
    try {
      const values = await form.validateFields();
      setSubmitting(true);
      const payload = {
        max_uses: values.max_uses,
        expires_in: values.expires_in?.trim() || undefined,
        note: values.note?.trim() || undefined,
      };
      const { data } = await adminMerchantService.createMerchantInvite(payload);
      message.success('邀请码已创建');
      if (data.data?.register_url) {
        await navigator.clipboard.writeText(data.data.register_url);
        message.success('注册链接已复制');
      }
      form.resetFields();
      await load();
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return;
      message.error('创建邀请码失败');
    } finally {
      setSubmitting(false);
    }
  };

  const handleCopy = async (text?: string) => {
    if (!text) return;
    try {
      await navigator.clipboard.writeText(text);
      message.success('已复制');
    } catch {
      message.error('复制失败');
    }
  };

  const effectiveUrl = (invite: MerchantInvite) =>
    invite.register_url || `${window.location.origin}/register?invite=${invite.code}`;

  const columns: ColumnsType<MerchantInvite> = useMemo(
    () => [
      { title: '邀请码', dataIndex: 'code', key: 'code', width: 180, fixed: 'left' },
      { title: '状态', dataIndex: 'status', key: 'status', width: 100, render: (v) => statusTag(v) },
      {
        title: '使用次数',
        key: 'usage',
        width: 110,
        render: (_v, row) => `${row.used_count}/${row.max_uses}`,
      },
      {
        title: '过期时间',
        dataIndex: 'expires_at',
        key: 'expires_at',
        width: 180,
        render: (v?: string) => (v ? new Date(v).toLocaleString('zh-CN') : '不过期'),
      },
      {
        title: '备注',
        dataIndex: 'note',
        key: 'note',
        ellipsis: true,
        render: (v?: string) => v || '-',
      },
      {
        title: '创建人',
        dataIndex: 'creator',
        key: 'creator',
        width: 220,
        responsive: ['md'],
        render: (v?: string) => v || '-',
      },
      {
        title: '创建时间',
        dataIndex: 'created_at',
        key: 'created_at',
        width: 180,
        responsive: ['lg'],
        render: (v: string) => new Date(v).toLocaleString('zh-CN'),
      },
      {
        title: '操作',
        key: 'actions',
        width: 230,
        fixed: 'right',
        render: (_v, row) => (
          <Space wrap size="small">
            <Tooltip title="复制注册链接">
              <Button
                icon={<CopyOutlined />}
                size="small"
                onClick={() => handleCopy(effectiveUrl(row))}
              />
            </Tooltip>
            <Tooltip title="打开注册链接">
              <Button
                icon={<LinkOutlined />}
                size="small"
                onClick={() => window.open(effectiveUrl(row), '_blank', 'noopener,noreferrer')}
              />
            </Tooltip>
            <Tooltip title="查看二维码">
              <Button
                icon={<QrcodeOutlined />}
                size="small"
                onClick={() =>
                  setQrUrl(
                    `https://api.qrserver.com/v1/create-qr-code/?size=360x360&data=${encodeURIComponent(
                      effectiveUrl(row)
                    )}`
                  )
                }
              />
            </Tooltip>
            <Button
              danger
              size="small"
              icon={<StopOutlined />}
              disabled={row.status !== 'active'}
              onClick={() => {
                setRevokeReason('');
                setRevokeTarget(row);
              }}
            >
              撤销
            </Button>
          </Space>
        ),
      },
    ],
    []
  );

  useEffect(() => {
    if (visible) {
      void load();
    }
  }, [visible]);

  if (!visible) return null;

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Card title="创建邀请码">
        <Form form={form} layout="vertical">
          <Space wrap style={{ width: '100%' }}>
            <Form.Item
              label="最大使用次数"
              name="max_uses"
              initialValue={1}
              rules={[{ required: true, message: '请输入次数' }]}
            >
              <InputNumber min={1} max={9999} style={{ width: 180 }} />
            </Form.Item>
            <Form.Item
              label="过期时长"
              name="expires_in"
              tooltip="例如 168h（7天），留空表示永不过期"
            >
              <Input placeholder="如 168h / 720h" style={{ width: 220 }} />
            </Form.Item>
            <Form.Item label="备注" name="note">
              <Input placeholder="渠道 / 客户简称 / 运营备注" style={{ width: 320 }} />
            </Form.Item>
            <Form.Item label=" ">
              <Button type="primary" loading={submitting} onClick={onCreate}>
                创建并复制链接
              </Button>
            </Form.Item>
          </Space>
        </Form>
      </Card>

      <Card title="邀请码列表">
        <Space wrap style={{ marginBottom: 12 }}>
          <Select
            style={{ width: 140 }}
            value={status || undefined}
            placeholder="状态"
            allowClear
            options={INVITE_STATUS_OPTIONS.filter((v) => v.value !== '')}
            onChange={(v) => setStatus(v ?? '')}
          />
          <Input.Search
            allowClear
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            onSearch={load}
            style={{ width: 300, maxWidth: '100%' }}
            placeholder="搜索 邀请码/备注/创建人"
          />
          <Button type="primary" onClick={load}>
            查询
          </Button>
          <Button onClick={() => void load()}>刷新</Button>
        </Space>
        <Table
          rowKey="id"
          loading={loading}
          dataSource={invites}
          columns={columns}
          scroll={{ x: 1200 }}
          pagination={{ pageSize: 20, showSizeChanger: false, showTotal: (t) => `共 ${t} 条` }}
        />
      </Card>

      <Modal title="邀请链接二维码" open={!!qrUrl} footer={null} onCancel={() => setQrUrl('')}>
        {qrUrl ? (
          <Space direction="vertical" style={{ width: '100%', alignItems: 'center' }}>
            <img src={qrUrl} alt="invite-qr" style={{ width: 260, maxWidth: '100%' }} />
            <Text type="secondary">扫码后可直达商户注册页并自动带邀请码</Text>
          </Space>
        ) : null}
      </Modal>

      <Modal
        title="撤销邀请码"
        open={!!revokeTarget}
        onCancel={() => setRevokeTarget(null)}
        okText="确认撤销"
        okButtonProps={{ danger: true }}
        onOk={async () => {
          if (!revokeTarget) return;
          try {
            await adminMerchantService.revokeMerchantInvite(revokeTarget.id, revokeReason);
            message.success('邀请码已撤销');
            setRevokeTarget(null);
            await load();
          } catch {
            message.error('撤销失败');
          }
        }}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Text>将撤销邀请码：{revokeTarget?.code}</Text>
          <Input.TextArea
            rows={3}
            placeholder="撤销原因（选填）"
            value={revokeReason}
            onChange={(e) => setRevokeReason(e.target.value)}
          />
        </Space>
      </Modal>
    </Space>
  );
};

export default InviteManagementPanel;
