import React from 'react';
import { Card, Typography, Space, Divider, Row, Col, Avatar } from 'antd';
import { TeamOutlined, TrophyOutlined, SafetyOutlined, GlobalOutlined } from '@ant-design/icons';

const { Title, Text, Paragraph } = Typography;

const teamMembers = [
  { name: '张三', role: '创始人 & CEO', avatar: 'Z' },
  { name: '李四', role: '技术负责人', avatar: 'L' },
  { name: '王五', role: '产品负责人', avatar: 'W' },
  { name: '赵六', role: '运营负责人', avatar: 'Z' },
];

export const AboutPage: React.FC = () => {
  return (
    <div style={{ padding: '20px', maxWidth: 900, margin: '0 auto' }}>
      <Card>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div style={{ textAlign: 'center' }}>
            <Title level={2}>关于拼团Token</Title>
            <Text type="secondary">让AI服务触手可及</Text>
          </div>

          <Divider />

          <div>
            <Title level={4}>我们的使命</Title>
            <Paragraph>
              拼团Token致力于让优质的AI服务更加普惠。通过创新的拼团模式，我们让更多用户能够以更低的价格使用到高质量的AI模型服务。
            </Paragraph>
            <Paragraph>
              我们相信AI技术正在改变世界，而我们的使命是让这种改变惠及每一个人。无论您是开发者、创业者还是企业用户，都能在这里找到适合您的AI解决方案。
            </Paragraph>
          </div>

          <Divider />

          <Row gutter={[24, 24]}>
            <Col xs={24} sm={12} md={6}>
              <Card style={{ textAlign: 'center', height: '100%' }}>
                <GlobalOutlined style={{ fontSize: 32, color: '#1890ff', marginBottom: 16 }} />
                <Title level={5}>全球领先</Title>
                <Text type="secondary">接入全球顶尖AI模型</Text>
              </Card>
            </Col>
            <Col xs={24} sm={12} md={6}>
              <Card style={{ textAlign: 'center', height: '100%' }}>
                <TrophyOutlined style={{ fontSize: 32, color: '#52c41a', marginBottom: 16 }} />
                <Title level={5}>价格优势</Title>
                <Text type="secondary">拼团模式节省30%-50%</Text>
              </Card>
            </Col>
            <Col xs={24} sm={12} md={6}>
              <Card style={{ textAlign: 'center', height: '100%' }}>
                <SafetyOutlined style={{ fontSize: 32, color: '#faad14', marginBottom: 16 }} />
                <Title level={5}>安全可靠</Title>
                <Text type="secondary">企业级安全保障</Text>
              </Card>
            </Col>
            <Col xs={24} sm={12} md={6}>
              <Card style={{ textAlign: 'center', height: '100%' }}>
                <TeamOutlined style={{ fontSize: 32, color: '#722ed1', marginBottom: 16 }} />
                <Title level={5}>专业团队</Title>
                <Text type="secondary">7x24小时技术支持</Text>
              </Card>
            </Col>
          </Row>

          <Divider />

          <div>
            <Title level={4}>核心团队</Title>
            <Row gutter={[16, 16]}>
              {teamMembers.map((member) => (
                <Col xs={12} sm={6} key={member.name}>
                  <Card style={{ textAlign: 'center' }}>
                    <Avatar size={64} style={{ backgroundColor: '#1890ff', marginBottom: 8 }}>
                      {member.avatar}
                    </Avatar>
                    <div>
                      <Text strong>{member.name}</Text>
                    </div>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {member.role}
                    </Text>
                  </Card>
                </Col>
              ))}
            </Row>
          </div>

          <Divider />

          <div>
            <Title level={4}>联系我们</Title>
            <Space direction="vertical">
              <Text>📧 商务合作：business@pintuotuo.com</Text>
              <Text>📧 技术支持：support@pintuotuo.com</Text>
              <Text>📱 客服电话：400-888-8888</Text>
              <Text>📍 公司地址：北京市海淀区中关村科技园</Text>
            </Space>
          </div>
        </Space>
      </Card>
    </div>
  );
};

export default AboutPage;
