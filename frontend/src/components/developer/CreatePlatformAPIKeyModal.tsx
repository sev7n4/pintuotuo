import { useEffect, useState } from 'react';
import { Button, Form, Input, Modal, Typography, message } from 'antd';
import { CopyOutlined } from '@ant-design/icons';
import { useTokenStore } from '@/stores/tokenStore';
import { copyToClipboard } from '@/utils/clipboard';

const { Paragraph, Text } = Typography;

export type CreatePlatformAPIKeyModalProps = {
  open: boolean;
  onClose: () => void;
  /** 创建成功且用户点「我已保存」关闭时调用 */
  onSuccess?: () => void;
};

/**
 * 与「密钥与安全」页相同的创建流程：名称 → 创建 → 一次性展示完整 ptd_ 密钥。
 */
export function CreatePlatformAPIKeyModal({
  open,
  onClose,
  onSuccess,
}: CreatePlatformAPIKeyModalProps) {
  const { createAPIKey, error } = useTokenStore();
  const [newKeyDisplay, setNewKeyDisplay] = useState<string | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    if (!open) {
      form.resetFields();
      setNewKeyDisplay(null);
    }
  }, [open, form]);

  const handleCreate = async () => {
    try {
      const values = await form.validateFields();
      const key = await createAPIKey(values.name.trim());
      if (key) {
        message.success('API密钥创建成功');
        setNewKeyDisplay(key);
        form.resetFields();
        onSuccess?.();
      } else {
        message.error(error || '创建失败');
      }
    } catch {
      /* validation */
    }
  };

  const handleClose = () => {
    onClose();
    form.resetFields();
    setNewKeyDisplay(null);
  };

  const handleOk = () => {
    if (newKeyDisplay) {
      handleClose();
      return;
    }
    void handleCreate();
  };

  return (
    <Modal
      title="创建 API 密钥"
      open={open}
      onOk={handleOk}
      onCancel={handleClose}
      okText={newKeyDisplay ? '我已保存' : '创建'}
      cancelText="取消"
      destroyOnClose
    >
      {newKeyDisplay ? (
        <>
          <Paragraph type="warning">
            密钥已加密保存，你可关闭本窗口，稍后通过列表右侧眼睛图标随时查看完整密钥。
          </Paragraph>
          <Paragraph>
            <Text code style={{ wordBreak: 'break-all' }}>
              {newKeyDisplay}
            </Text>
          </Paragraph>
          <Button
            type="primary"
            icon={<CopyOutlined />}
            onClick={async () => {
              const ok = await copyToClipboard(newKeyDisplay);
              if (ok) message.success('已复制完整密钥');
              else message.error('复制失败，请手动选择文本复制');
            }}
          >
            复制完整密钥
          </Button>
        </>
      ) : (
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="例如：本机开发" autoFocus />
          </Form.Item>
        </Form>
      )}
    </Modal>
  );
}
