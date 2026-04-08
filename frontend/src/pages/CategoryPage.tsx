import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Row, Col, Input, Empty, Spin, Typography } from 'antd';
import { SearchOutlined, AppstoreOutlined } from '@ant-design/icons';
import { productService } from '@/services/product';
import type { Category as ApiCategory } from '@/types';
import styles from './CategoryPage.module.css';

const { Title, Text } = Typography;

const CategoryPage = () => {
  const navigate = useNavigate();
  const [searchText, setSearchText] = useState('');
  const [loading, setLoading] = useState(false);
  const [categories, setCategories] = useState<ApiCategory[]>([]);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    productService
      .getCategories()
      .then((res) => {
        const body = res.data as { data?: ApiCategory[] };
        if (!cancelled) setCategories(body?.data || []);
      })
      .catch(() => {
        if (!cancelled) setCategories([]);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const filteredCategories = categories.filter((c) => {
    if (!searchText.trim()) return true;
    const q = searchText.toLowerCase();
    return c.name.toLowerCase().includes(q);
  });

  const goCatalog = (categoryName: string) => {
    navigate(`/catalog?category=${encodeURIComponent(categoryName)}`);
  };

  return (
    <div className={styles.categoryPage}>
      <div className={styles.header}>
        <Title level={3} className={styles.title}>
          商品分类
        </Title>
        <Text type="secondary" className={styles.subtitle}>
          与首页、卖场使用同一套分类数据；点选进入 SKU 列表。
        </Text>
        <Input
          placeholder="筛选分类名称..."
          prefix={<SearchOutlined />}
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          className={styles.searchInput}
          allowClear
        />
      </div>

      <Spin spinning={loading}>
        {filteredCategories.length === 0 && !loading ? (
          <Empty description="暂无分类数据" />
        ) : (
          <Row gutter={[16, 16]} className={styles.categoryGrid}>
            {filteredCategories.map((category) => (
              <Col xs={12} sm={8} md={6} lg={4} key={category.name}>
                <Card
                  hoverable
                  className={styles.categoryCard}
                  onClick={() => goCatalog(category.name)}
                >
                  <div className={styles.categoryContent}>
                    <div className={styles.categoryIcon}>
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
        )}
      </Spin>
    </div>
  );
};

export default CategoryPage;
