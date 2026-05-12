import { useEffect, useRef } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Spin, message } from 'antd';
import { useAuthStore } from '@/stores/authStore';
import { useGroupStore } from '@/stores/groupStore';
import { getApiErrorMessage } from '@/utils/apiError';

/**
 * 参团落地页：与分享链接 /groups/:groupId/join 对应，登录后自动调 join 并跳转支付。
 */
export default function GroupJoinPage() {
  const { groupId } = useParams<{ groupId: string }>();
  const navigate = useNavigate();
  const { isAuthenticated } = useAuthStore();
  const joinGroup = useGroupStore((s) => s.joinGroup);
  const joinStarted = useRef(false);

  useEffect(() => {
    const raw = (groupId || '').trim();
    if (!raw || !/^\d+$/.test(raw)) {
      message.warning('无效的拼团链接');
      navigate('/groups', { replace: true });
      return;
    }

    if (!isAuthenticated) {
      navigate(`/login?redirect=${encodeURIComponent(`/groups/${raw}/join`)}`, { replace: true });
      return;
    }

    if (joinStarted.current) return;
    joinStarted.current = true;

    const id = parseInt(raw, 10);
    void (async () => {
      try {
        const orderId = await joinGroup(id);
        message.success('参团成功，请完成支付');
        if (orderId) {
          navigate(`/payment/${orderId}`, { replace: true });
        } else {
          navigate(`/groups/${id}`, { replace: true });
        }
      } catch (e) {
        message.error(getApiErrorMessage(e, '加入拼团失败'));
        navigate(`/groups/${id}`, { replace: true });
      }
    })();
  }, [groupId, isAuthenticated, joinGroup, navigate]);

  return (
    <div
      style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 320 }}
    >
      <Spin size="large" tip="正在处理参团…" />
    </div>
  );
}
