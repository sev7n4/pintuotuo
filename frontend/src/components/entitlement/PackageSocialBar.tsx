import { useCallback, useEffect, useState } from 'react';
import {
  Button,
  Divider,
  Input,
  Modal,
  Rate,
  Space,
  Tooltip,
  Typography,
  message,
} from 'antd';
import {
  CommentOutlined,
  HeartFilled,
  HeartOutlined,
  LikeOutlined,
  ShoppingOutlined,
  StarOutlined,
} from '@ant-design/icons';

import { entitlementPackageService } from '@/services/entitlementPackage';
import { useAuthStore } from '@/stores/authStore';
import { getApiErrorMessage } from '@/utils/apiError';
import styles from './EntitlementPackageCard.module.css';

const { Text } = Typography;

export type PackageSocialStats = {
  favoriteCount?: number;
  likeCount?: number;
  salesCount?: number;
  reviewCount?: number;
  userFavorited?: boolean;
  userLiked?: boolean;
  userReviewed?: boolean;
};

type Props = {
  packageId: number;
  packageCode: string;
  stats?: PackageSocialStats;
  onSocialPatch?: (patch: Partial<PackageSocialStats>) => void;
};

/**
 * 收藏 / 点赞 / 销量 / 评价：数据来自后端；登录后可收藏、点赞、发表评价。
 */
const LEGACY_LS_PREFIX = 'ptd_ent_pkg_fav_';

export function PackageSocialBar({ packageId, packageCode, stats, onSocialPatch }: Props) {
  const { isAuthenticated } = useAuthStore();
  const [reviewOpen, setReviewOpen] = useState(false);
  const [reviewRating, setReviewRating] = useState(5);
  const [reviewComment, setReviewComment] = useState('');
  const [reviewSubmitting, setReviewSubmitting] = useState(false);
  const [favLoading, setFavLoading] = useState(false);

  const [local, setLocal] = useState<PackageSocialStats>({});

  /** 历史版本曾用 localStorage 存收藏态；账号收藏已接后端，清理旧键避免与「我的收藏」不一致 */
  useEffect(() => {
    try {
      const drop: string[] = [];
      for (let i = 0; i < localStorage.length; i++) {
        const k = localStorage.key(i);
        if (k?.startsWith(LEGACY_LS_PREFIX)) drop.push(k);
      }
      drop.forEach((k) => localStorage.removeItem(k));
    } catch {
      /* ignore */
    }
  }, []);

  useEffect(() => {
    setLocal(stats || {});
  }, [stats]);

  const favN = local.favoriteCount ?? 0;
  const likeN = local.likeCount ?? 0;
  const salesN = local.salesCount ?? 0;
  const reviewN = local.reviewCount ?? 0;
  const favorited = !!local.userFavorited;
  const liked = !!local.userLiked;

  const applyPatch = useCallback(
    (patch: Partial<PackageSocialStats>) => {
      setLocal((prev) => ({ ...prev, ...patch }));
      onSocialPatch?.(patch);
    },
    [onSocialPatch]
  );

  const toggleFav = useCallback(async () => {
    if (!isAuthenticated) {
      message.info('请先登录后再收藏');
      return;
    }
    setFavLoading(true);
    try {
      if (favorited) {
        const res = await entitlementPackageService.removeFavorite(packageId);
        const d = res.data?.data;
        applyPatch({
          userFavorited: d?.favorited ?? false,
          favoriteCount: d?.favorite_count ?? Math.max(0, favN - 1),
        });
        message.success('已取消收藏');
      } else {
        const res = await entitlementPackageService.addFavorite(packageId);
        const d = res.data?.data;
        applyPatch({
          userFavorited: d?.favorited ?? true,
          favoriteCount: d?.favorite_count ?? favN + 1,
        });
        message.success('已收藏，可在「我的收藏」查看');
      }
    } catch (e) {
      message.error(getApiErrorMessage(e));
    } finally {
      setFavLoading(false);
    }
  }, [applyPatch, favN, favorited, isAuthenticated, packageId]);

  const toggleLike = useCallback(async () => {
    if (!isAuthenticated) {
      message.info('请先登录后再点赞');
      return;
    }
    try {
      const res = await entitlementPackageService.toggleLike(packageId);
      const d = res.data?.data;
      applyPatch({
        userLiked: d?.liked,
        likeCount: d?.like_count,
      });
    } catch (e) {
      message.error(getApiErrorMessage(e));
    }
  }, [applyPatch, isAuthenticated, packageId]);

  const submitReview = useCallback(async () => {
    if (!isAuthenticated) {
      message.info('请先登录后再评价');
      return;
    }
    setReviewSubmitting(true);
    try {
      const res = await entitlementPackageService.upsertReview(packageId, {
        rating: reviewRating,
        comment: reviewComment.trim() || undefined,
      });
      const d = res.data?.data;
      applyPatch({
        reviewCount: d?.review_count,
        userReviewed: true,
      });
      setReviewOpen(false);
      setReviewComment('');
      message.success('评价已提交');
    } catch (e) {
      message.error(getApiErrorMessage(e));
    } finally {
      setReviewSubmitting(false);
    }
  }, [applyPatch, isAuthenticated, packageId, reviewComment, reviewRating]);

  return (
    <div className={styles.socialBar}>
      <Space size="middle" wrap split={<Divider type="vertical" />}>
        <Tooltip title={isAuthenticated ? '收藏（与「我的收藏」同步）' : '登录后可收藏'}>
          <Button
            type="text"
            size="small"
            className={styles.socialBtn}
            loading={favLoading}
            icon={favorited ? <HeartFilled style={{ color: '#eb2f96' }} /> : <HeartOutlined />}
            onClick={() => void toggleFav()}
          >
            收藏 <Text type="secondary">{favN}</Text>
          </Button>
        </Tooltip>
        <Tooltip title={isAuthenticated ? '点赞' : '登录后可点赞'}>
          <Button
            type="text"
            size="small"
            className={styles.socialBtn}
            icon={<LikeOutlined style={liked ? { color: '#1677ff' } : undefined} />}
            onClick={() => void toggleLike()}
          >
            点赞 <Text type="secondary">{likeN}</Text>
          </Button>
        </Tooltip>
        <span className={styles.socialItem}>
          <ShoppingOutlined /> 销量 <Text type="secondary">{salesN}</Text>
        </span>
        <span className={styles.socialItem}>
          <StarOutlined /> 评价 <Text type="secondary">{reviewN}</Text>
        </span>
        <Tooltip title="发表或更新评价">
          <Button
            type="text"
            size="small"
            className={styles.socialBtn}
            icon={<CommentOutlined />}
            onClick={() => {
              if (!isAuthenticated) {
                message.info('请先登录后再评价');
                return;
              }
              setReviewOpen(true);
            }}
          >
            用户评价
          </Button>
        </Tooltip>
      </Space>

      <Modal
        title={`评价套餐 ${packageCode}`}
        open={reviewOpen}
        onOk={() => void submitReview()}
        onCancel={() => setReviewOpen(false)}
        confirmLoading={reviewSubmitting}
        okText="提交"
      >
        <Space direction="vertical" style={{ width: '100%' }} size="middle">
          <div>
            <Text type="secondary">评分</Text>
            <div>
              <Rate value={reviewRating} onChange={setReviewRating} />
            </div>
          </div>
          <div>
            <Text type="secondary">评语（可选）</Text>
            <Input.TextArea
              rows={3}
              value={reviewComment}
              onChange={(e) => setReviewComment(e.target.value)}
              placeholder="说说使用体验"
              maxLength={2000}
              showCount
            />
          </div>
        </Space>
      </Modal>
    </div>
  );
}
