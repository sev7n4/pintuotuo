import React, { useEffect, useState } from 'react';
import { Card, Row, Col, Spin, Tag } from 'antd';
import {
  CodeOutlined,
  EditOutlined,
  BarChartOutlined,
  MessageOutlined,
  EyeOutlined,
  ThunderboltOutlined,
  SoundOutlined,
  BulbOutlined,
} from '@ant-design/icons';
import { useNavigate, useSearchParams } from 'react-router-dom';
import api from '@/services/api';
import styles from './ScenarioFilter.module.css';

interface UsageScenario {
  id: number;
  code: string;
  name: string;
  description?: string;
  icon_url?: string;
  sort_order: number;
  status: string;
  spu_count?: number;
}

const scenarioIconMap: Record<string, React.ReactNode> = {
  coding: <CodeOutlined />,
  writing: <EditOutlined />,
  analysis: <BarChartOutlined />,
  chat: <MessageOutlined />,
  vision: <EyeOutlined />,
  embedding: <ThunderboltOutlined />,
  audio: <SoundOutlined />,
  reasoning: <BulbOutlined />,
};

const scenarioColorMap: Record<string, string> = {
  coding: '#1890ff',
  writing: '#52c41a',
  analysis: '#722ed1',
  chat: '#fa8c16',
  vision: '#eb2f96',
  embedding: '#13c2c2',
  audio: '#2f54eb',
  reasoning: '#faad14',
};

export type ScenarioFilterVariant = 'panel' | 'rail';

interface ScenarioFilterProps {
  variant?: ScenarioFilterVariant;
}

export const ScenarioFilter: React.FC<ScenarioFilterProps> = ({ variant = 'panel' }) => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [scenarios, setScenarios] = useState<UsageScenario[]>([]);
  const [loading, setLoading] = useState(false);
  const [activeScenario, setActiveScenario] = useState<string | null>(null);

  useEffect(() => {
    const scenarioParam = searchParams.get('scenario');
    setActiveScenario(scenarioParam);
  }, [searchParams]);

  useEffect(() => {
    loadScenarios();
  }, []);

  const loadScenarios = async () => {
    setLoading(true);
    try {
      const response = await api.get('/catalog/scenarios');
      const data = (response.data as { scenarios?: UsageScenario[] }).scenarios || [];
      setScenarios(data);
    } catch {
      setScenarios([]);
    } finally {
      setLoading(false);
    }
  };

  const handleScenarioClick = (scenarioCode: string) => {
    if (activeScenario === scenarioCode) {
      const newParams = new URLSearchParams(searchParams);
      newParams.delete('scenario');
      navigate({ search: newParams.toString() });
    } else {
      const newParams = new URLSearchParams(searchParams);
      newParams.set('scenario', scenarioCode);
      navigate({ search: newParams.toString() });
    }
  };

  if (loading && variant === 'rail') {
    return (
      <div className={styles.railWrap}>
        <Spin size="small" />
      </div>
    );
  }

  if (loading) {
    return (
      <Card className={styles.panelCard}>
        <Spin tip="加载场景分类...">
          <div style={{ height: 80 }} />
        </Spin>
      </Card>
    );
  }

  if (scenarios.length === 0) {
    return null;
  }

  if (variant === 'rail') {
    return (
      <div className={styles.railWrap} role="toolbar" aria-label="使用场景筛选">
        <div className={styles.railScroll}>
          {scenarios.map((scenario) => {
            const isActive = activeScenario === scenario.code;
            const icon = scenarioIconMap[scenario.code] || <BulbOutlined />;
            return (
              <button
                key={scenario.id}
                type="button"
                className={`${styles.railChip} ${isActive ? styles.railChipActive : ''}`}
                onClick={() => handleScenarioClick(scenario.code)}
                title={
                  scenario.spu_count != null && scenario.spu_count > 0
                    ? `${scenario.name}（${scenario.spu_count} 款 SPU）`
                    : scenario.name
                }
              >
                <span className={styles.railIcon}>{icon}</span>
                <span>{scenario.name}</span>
              </button>
            );
          })}
        </div>
      </div>
    );
  }

  return (
    <Card title="使用场景" className={styles.panelCard} bodyStyle={{ padding: '12px 16px' }}>
      <Row gutter={[12, 12]}>
        {scenarios.map((scenario) => {
          const isActive = activeScenario === scenario.code;
          const icon = scenarioIconMap[scenario.code] || <BulbOutlined />;
          const color = scenarioColorMap[scenario.code] || '#1890ff';

          return (
            <Col key={scenario.id} xs={12} sm={8} md={6} lg={4}>
              <Card
                hoverable
                onClick={() => handleScenarioClick(scenario.code)}
                style={{
                  textAlign: 'center',
                  borderColor: isActive ? color : undefined,
                  borderWidth: isActive ? 2 : 1,
                  backgroundColor: isActive ? `${color}10` : undefined,
                  transition: 'all 0.3s',
                }}
                bodyStyle={{ padding: '12px 8px' }}
              >
                <div style={{ fontSize: 28, color: isActive ? color : '#666', marginBottom: 4 }}>
                  {icon}
                </div>
                <div style={{ fontWeight: isActive ? 600 : 400, color: isActive ? color : '#333' }}>
                  {scenario.name}
                </div>
                {scenario.spu_count !== undefined && scenario.spu_count > 0 && (
                  <Tag color={isActive ? color : 'default'} style={{ marginTop: 4, fontSize: 10 }}>
                    {scenario.spu_count} 款
                  </Tag>
                )}
              </Card>
            </Col>
          );
        })}
      </Row>
    </Card>
  );
};
