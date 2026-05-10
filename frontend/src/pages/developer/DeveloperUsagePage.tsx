import { Space, Typography } from 'antd';
import { Link } from 'react-router-dom';
import Consumption from '@/pages/Consumption';

const { Title, Paragraph } = Typography;

/** 与「消费明细」页相同的数据视图，嵌入开发者中心任务流 */
export default function DeveloperUsagePage() {
  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <div>
        <Title level={3} style={{ marginTop: 0 }}>
          用量与账单
        </Title>
        <Paragraph type="secondary">
          以下为账户维度的调用与 Token 扣减明细。也可在{' '}
          <Link to="/consumption">独立页面</Link> 打开同一视图。
        </Paragraph>
      </div>
      <Consumption />
    </Space>
  );
}
