import { useState } from 'react';
import { isAxiosError } from 'axios';
import { Button, Modal, Space, Tooltip, Typography, message } from 'antd';
import { CopyOutlined, EyeInvisibleOutlined, EyeOutlined } from '@ant-design/icons';
import type { UserAPIKey } from '@/types';
import { tokenService } from '@/services/token';
import { copyToClipboard } from '@/utils/clipboard';

const { Text, Paragraph } = Typography;

type Props = {
  record: UserAPIKey;
};

/** 列表展示预览 + 眼睛查看完整密钥 + 复制完整密钥（需后端 key_encrypted） */
export function PlatformAPIKeySecretCell({ record }: Props) {
  const [open, setOpen] = useState(false);
  const [fullKey, setFullKey] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const preview =
    record.key_preview && record.key_preview.trim() !== ''
      ? record.key_preview
      : 'ptd_…（无预览）';

  const canReveal = record.can_reveal === true;

  const loadSecret = async () => {
    setLoading(true);
    setFullKey(null);
    try {
      const res = await tokenService.revealAPIKey(record.id);
      const k = res.data?.key;
      if (!k) {
        message.error('未能获取完整密钥');
        return;
      }
      setFullKey(k);
      setOpen(true);
    } catch (e: unknown) {
      let msg = '获取密钥失败';
      if (isAxiosError(e)) {
        const d = e.response?.data as { message?: string; error?: string } | undefined;
        msg = (d?.message || d?.error || msg).trim() || msg;
      }
      message.error(msg);
    } finally {
      setLoading(false);
    }
  };

  const handleCopyFull = async () => {
    if (!fullKey) return;
    const ok = await copyToClipboard(fullKey);
    if (ok) message.success('已复制完整密钥');
    else message.error('复制失败，请长按手动选择复制');
  };

  const handleCopyPreview = async () => {
    const ok = await copyToClipboard(preview);
    if (ok) message.success('已复制预览（非完整密钥）');
    else message.error('复制失败');
  };

  return (
    <>
      <Space wrap size={4}>
        <Text code>{preview}</Text>
        <Tooltip title="复制预览片段（不可用于鉴权，完整密钥请点眼睛）">
          <Button type="text" size="small" icon={<CopyOutlined />} onClick={handleCopyPreview} />
        </Tooltip>
        {canReveal ? (
          <Tooltip title="查看完整密钥">
            <Button
              type="text"
              size="small"
              icon={<EyeOutlined />}
              loading={loading}
              onClick={loadSecret}
              aria-label="查看完整密钥"
            />
          </Tooltip>
        ) : (
          <Tooltip title="该密钥创建于加密存储上线前，无法再次查看完整内容，请新建密钥">
            <EyeInvisibleOutlined style={{ color: '#bfbfbf' }} />
          </Tooltip>
        )}
      </Space>

      <Modal
        title="完整平台密钥"
        open={open}
        onCancel={() => {
          setOpen(false);
          setFullKey(null);
        }}
        footer={[
          <Button key="copy" type="primary" icon={<CopyOutlined />} onClick={handleCopyFull}>
            复制完整密钥
          </Button>,
          <Button
            key="close"
            onClick={() => {
              setOpen(false);
              setFullKey(null);
            }}
          >
            关闭
          </Button>,
        ]}
      >
        <Text type="warning" style={{ display: 'block', marginBottom: 8 }}>
          请勿泄露给他人；关闭窗口后仍可随时通过眼睛图标再次查看。
        </Text>
        <Paragraph style={{ marginBottom: 0 }}>
          <Text code style={{ wordBreak: 'break-all' }}>
            {fullKey ?? '—'}
          </Text>
        </Paragraph>
      </Modal>
    </>
  );
}
