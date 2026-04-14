import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Row, Col, Input, Empty, Spin, Typography } from 'antd';
import { SearchOutlined, AppstoreOutlined, ClusterOutlined } from '@ant-design/icons';
import { productService } from '@/services/product';
import type { Category as ApiCategory } from '@/types';
import styles from './CategoryPage.module.css';

const { Title, Text } = Typography;

interface ScenarioItem {
  id: number;
  code: string;
  name: string;
  spu_count?: number;
}

const CategoryPage = () => {
  const navigate = useNavigate();
  const [searchText, setSearchText] = useState('');
  const [loading, setLoading] = useState(false);
  const [tierCategories, setTierCategories] = useState<ApiCategory[]>([]);
  const [scenarios, setScenarios] = useState<ScenarioItem[]>([]);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    Promise.all([
      productService.getCategories().then((res) => {
        const body = res.data as { data?: ApiCategory[] };
        return body?.data || [];
      }),
      productService.getCatalogScenarios().then((res) => {
        const body = res.data as { scenarios?: ScenarioItem[] };
        return body?.scenarios || [];
      }),
    ])
      .then(([tiers, scen]) => {
        if (!cancelled) {
          setTierCategories(tiers);
          setScenarios(scen);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setTierCategories([]);
          setScenarios([]);
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const q = searchText.trim().toLowerCase();
  const match = (s: string) => !q || s.toLowerCase().includes(q);

  const filteredScenarios = scenarios.filter((s) => match(s.name) || match(s.code));
  const filteredTiers = tierCategories.filter((c) => match(c.name));

  const goScenario = (code: string) => {
    navigate(`/catalog?scenario=${encodeURIComponent(code)}`);
  };

  const goTier = (tierName: string) => {
    navigate(`/catalog?tier=${encodeURIComponent(tierName)}`);
  };

  const emptyAll = !loading && filteredScenarios.length === 0 && filteredTiers.length === 0;

  return (
    <div className={`${styles.categoryPage} ${styles.pageMinimal}`}>
      <div className={styles.header}>
        <Title level={3} className={styles.title}>
          商品分类
        </Title>
        <Text type="secondary" className={styles.subtitle}>
          使用场景为主、模型层级为辅；点击跳转对应 SKU 列表。
        </Text>
        <Input
          placeholder="筛选名称..."
          prefix={<SearchOutlined />}
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          className={styles.searchInput}
          allowClear
        />
      </div>

      <Spin spinning={loading}>
        {emptyAll ? (
          <Empty description="暂无分类数据" />
        ) : (
          <>
            {filteredScenarios.length > 0 && (
              <section className={styles.section}>
                <Text className={styles.sectionLabel}>
                  <ClusterOutlined /> 使用场景
                </Text>
                <Row gutter={[12, 12]} className={styles.categoryGrid}>
                  {filteredScenarios.map((s) => (
                    <Col xs={12} sm={8} md={6} lg={4} key={s.code}>
                      <Card
                        hoverable
                        className={styles.categoryCard}
                        onClick={() => goScenario(s.code)}
                        styles={{ body: { padding: 14 } }}
                      >
                        <div className={styles.categoryContent}>
                          <div className={`${styles.categoryIcon} ${styles.categoryIconScenario}`}>
                            <AppstoreOutlined />
                          </div>
                          <Text strong className={styles.categoryName}>
                            {s.name}
                          </Text>
                          <Text type="secondary" className={styles.categoryMeta}>
                            {s.spu_count != null && s.spu_count > 0
                              ? `${s.spu_count} 款关联 SPU`
                              : '进入筛选'}
                          </Text>
                        </div>
                      </Card>
                    </Col>
                  ))}
                </Row>
              </section>
            )}

            {filteredTiers.length > 0 && (
              <section className={styles.section}>
                <Text className={styles.sectionLabel}>
                  <AppstoreOutlined /> 模型层级
                </Text>
                <Row gutter={[16, 16]} className={styles.categoryGrid}>
                  {filteredTiers.map((category) => (
                    <Col xs={12} sm={8} md={6} lg={4} key={category.name}>
                      <Card
                        hoverable
                        className={`${styles.categoryCard} ${styles.categoryCardSubtle}`}
                        onClick={() => goTier(category.name)}
                        styles={{ body: { padding: 14 } }}
                      >
                        <div className={styles.categoryContent}>
                          <div className={`${styles.categoryIcon} ${styles.categoryIconTier}`}>
                            <AppstoreOutlined />
                          </div>
                          <Text strong className={styles.categoryName}>
                            {category.name}
                          </Text>
                          <Text type="secondary" className={styles.categoryMeta}>
                            {category.count} 件在售
                          </Text>
                        </div>
                      </Card>
                    </Col>
                  ))}
                </Row>
              </section>
            )}
          </>
        )}
      </Spin>
    </div>
  );
};

export default CategoryPage;
