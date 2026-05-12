import { Button, Progress, Space, Tag, Typography, Alert, Grid, Spin, Segmented } from 'antd';
import { TeamOutlined, RightOutlined, LoginOutlined } from '@ant-design/icons';
import type { Group } from '@/types';
import type { GroupListScope } from '@/services/group';
import {
  groupProductSubtitle,
  groupProgressBarStatus,
  groupStatusTagLabel,
} from '@/utils/groupDisplay';
import styles from './CatalogGroupsShowcase.module.css';

const { Text, Title } = Typography;
const { useBreakpoint } = Grid;

export type CatalogGroupsLoadStatus =
  | 'idle'
  | 'loading'
  | 'ok'
  | 'empty'
  | 'unauthorized'
  | 'error';

interface CatalogGroupsShowcaseProps {
  layout: 'rail' | 'expanded';
  groups: Group[];
  total: number;
  status: CatalogGroupsLoadStatus;
  /** 与 GET /groups?scope= 对齐：全站 / 我的拼团 / 我发起 / 我跟团 */
  groupScope: GroupListScope;
  onGroupScopeChange: (scope: GroupListScope) => void;
  onOpenGroup: (id: number) => void;
  onOpenAll: () => void;
  onLogin: () => void;
}

function scopeTitle(scope: GroupListScope, layout: 'rail' | 'expanded'): string {
  if (layout === 'expanded') {
    if (scope === 'mine_involved') return '我的拼团（发起与参团）';
    if (scope === 'mine_created') return '我发起的拼团';
    if (scope === 'mine_joined') return '我跟的拼团';
    return '全站进行中的拼团';
  }
  if (scope === 'mine_involved') return '我的拼团';
  if (scope === 'mine_created') return '我发起的拼团';
  if (scope === 'mine_joined') return '我跟的拼团';
  return '全站热团';
}

function emptyHint(
  scope: GroupListScope,
  layout: 'rail' | 'expanded'
): { title: string; desc?: string } {
  if (scope === 'mine_involved') {
    return {
      title: '暂无你发起或参团的进行中拼团',
      desc: '去卖场发起新团，或到「全站热团」加入别人的团。',
    };
  }
  if (scope === 'mine_created') {
    return {
      title: layout === 'rail' ? '暂无你发起的进行中拼团' : '暂无你发起的进行中拼团',
      desc: '去卖场挑支持拼团的商品发起新团。',
    };
  }
  if (scope === 'mine_joined') {
    return {
      title: '暂无你跟团的记录',
      desc: '切换到「全站热团」发现更多可参团的团。',
    };
  }
  return {
    title: layout === 'rail' ? '暂无进行中的团' : '暂无进行中的拼团',
    desc: '先去挑支持拼团的商品开团吧。',
  };
}

function GroupMiniCard({ group, onClick }: { group: Group; onClick: () => void }) {
  const progress = Math.min(100, Math.round((group.current_count / group.target_count) * 100));
  const tagLabel = groupStatusTagLabel(group);
  const tagColor =
    group.status === 'completed'
      ? 'green'
      : group.status === 'failed' || tagLabel === '已截止'
        ? 'default'
        : 'blue';
  const deadline = new Date(group.deadline);
  const expired = deadline.getTime() <= Date.now();
  const subtitle = groupProductSubtitle(group);

  return (
    <div
      className={styles.railCard}
      role="button"
      tabIndex={0}
      onClick={onClick}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          onClick();
        }
      }}
    >
      <Space direction="vertical" size={6} style={{ width: '100%' }}>
        <Space style={{ width: '100%', justifyContent: 'space-between' }} align="start">
          <Text strong>拼团 #{group.id}</Text>
          <Tag color={tagColor}>{tagLabel}</Tag>
        </Space>
        {subtitle && (
          <Text type="secondary" ellipsis style={{ fontSize: 12 }}>
            {subtitle}
          </Text>
        )}
        <Progress percent={progress} size="small" status={groupProgressBarStatus(group)} />
        <Text type="secondary" style={{ fontSize: 12 }}>
          {group.current_count}/{group.target_count} 人
          {expired && group.status === 'active' ? ' · 已截止' : ''}
        </Text>
      </Space>
    </div>
  );
}

export function CatalogGroupsShowcase({
  layout,
  groups,
  total,
  status,
  groupScope,
  onGroupScopeChange,
  onOpenGroup,
  onOpenAll,
  onLogin,
}: CatalogGroupsShowcaseProps) {
  const screens = useBreakpoint();

  const scopeSwitcher = (
    <Segmented<GroupListScope>
      size="small"
      value={groupScope}
      onChange={(v) => onGroupScopeChange(v as GroupListScope)}
      options={[
        { label: '全站热团', value: 'all' },
        { label: '我的拼团', value: 'mine_involved' },
        { label: '我发起', value: 'mine_created' },
        { label: '我跟团', value: 'mine_joined' },
      ]}
      className={styles.scopeSegmented}
    />
  );

  if (status === 'idle' || status === 'loading') {
    return (
      <div className={styles.section}>
        {scopeSwitcher}
        <div className={styles.sectionHead}>
          <Space>
            <TeamOutlined />
            <Title level={5} style={{ margin: 0 }}>
              {scopeTitle(groupScope, layout)}
            </Title>
          </Space>
        </div>
        {status === 'loading' ? <Spin /> : null}
      </div>
    );
  }

  if (status === 'unauthorized') {
    return (
      <div className={styles.section}>
        <Alert
          type="info"
          showIcon
          message="登录后可查看拼团列表（含我发起 / 我跟团）"
          description={
            <Button type="primary" icon={<LoginOutlined />} onClick={onLogin}>
              去登录
            </Button>
          }
        />
      </div>
    );
  }

  if (status === 'error') {
    return (
      <div className={styles.section}>
        {scopeSwitcher}
        <Alert type="warning" showIcon message="拼团列表暂时无法加载，请稍后重试" />
      </div>
    );
  }

  if (layout === 'rail' && status === 'empty') {
    const hint = emptyHint(groupScope, 'rail');
    return (
      <div className={styles.section}>
        {scopeSwitcher}
        <div className={styles.sectionHead}>
          <Space>
            <TeamOutlined />
            <Title level={5} style={{ margin: 0 }}>
              {scopeTitle(groupScope, layout)}
            </Title>
          </Space>
          <Button type="link" size="small" icon={<RightOutlined />} onClick={onOpenAll}>
            拼团中心
          </Button>
        </div>
        <Text type="secondary">{hint.title}</Text>
        {hint.desc ? (
          <Text type="secondary" style={{ display: 'block', marginTop: 4 }}>
            {hint.desc}
          </Text>
        ) : null}
      </div>
    );
  }

  if (layout === 'expanded' && status === 'empty') {
    const hint = emptyHint(groupScope, 'expanded');
    return (
      <div className={styles.section}>
        {scopeSwitcher}
        <Alert
          type="info"
          showIcon
          message={hint.title}
          description={
            <Space direction="vertical">
              {hint.desc ? <Text type="secondary">{hint.desc}</Text> : null}
              <Button type="primary" onClick={() => onOpenAll()}>
                打开拼团中心
              </Button>
            </Space>
          }
        />
      </div>
    );
  }

  const list = layout === 'rail' ? groups.slice(0, 12) : groups;

  if (layout === 'rail') {
    return (
      <div className={styles.section}>
        {scopeSwitcher}
        <div className={styles.sectionHead}>
          <Space wrap>
            <TeamOutlined />
            <Title level={5} style={{ margin: 0 }}>
              {scopeTitle(groupScope, layout)}
            </Title>
            {total > list.length ? <Tag color="processing">共 {total} 个</Tag> : null}
          </Space>
          <Button type="link" size="small" icon={<RightOutlined />} onClick={onOpenAll}>
            {screens.xs ? '全部' : '拼团中心'}
          </Button>
        </div>
        <div className={styles.railScroll}>
          {list.map((g) => (
            <GroupMiniCard key={g.id} group={g} onClick={() => onOpenGroup(g.id)} />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className={styles.section}>
      {scopeSwitcher}
      <div className={styles.sectionHead}>
        <Space wrap>
          <TeamOutlined />
          <Title level={5} style={{ margin: 0 }}>
            {scopeTitle(groupScope, layout)}
          </Title>
          <Tag>{total} 个</Tag>
        </Space>
        <Button type="link" size="small" icon={<RightOutlined />} onClick={onOpenAll}>
          打开拼团中心
        </Button>
      </div>
      <div className={styles.expandedGrid}>
        {list.map((g) => (
          <div key={g.id} className={styles.expandedCard}>
            <GroupMiniCard group={g} onClick={() => onOpenGroup(g.id)} />
          </div>
        ))}
      </div>
    </div>
  );
}
