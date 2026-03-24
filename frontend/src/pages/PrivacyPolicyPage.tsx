import React from 'react'
import { Card, Typography, Divider, Space } from 'antd'

const { Title, Text, Paragraph } = Typography

export const PrivacyPolicyPage: React.FC = () => {
  return (
    <div style={{ padding: '20px', maxWidth: 900, margin: '0 auto' }}>
      <Card>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div style={{ textAlign: 'center' }}>
            <Title level={2}>隐私政策</Title>
            <Text type="secondary">最后更新日期：2026年3月1日</Text>
          </div>

          <Divider />

          <div>
            <Title level={4}>引言</Title>
            <Paragraph>
              拼团Token平台（以下简称"我们"）非常重视用户的隐私和个人信息保护。本隐私政策将向您说明我们如何收集、使用、存储和保护您的个人信息，以及您享有的相关权利。
            </Paragraph>
            <Paragraph>
              请您在使用我们的服务前，仔细阅读并理解本隐私政策的全部内容。
            </Paragraph>
          </div>

          <div>
            <Title level={4}>一、我们收集的信息</Title>
            <Paragraph>
              为了向您提供服务，我们可能收集以下类型的信息：
            </Paragraph>
            <Paragraph>
              <strong>1. 账户信息</strong><br />
              当您注册账户时，我们会收集您的手机号码、电子邮箱地址等联系信息。
            </Paragraph>
            <Paragraph>
              <strong>2. 交易信息</strong><br />
              当您购买Token服务时，我们会收集订单信息、支付记录、Token使用记录等。
            </Paragraph>
            <Paragraph>
              <strong>3. 使用信息</strong><br />
              我们会收集您使用服务的相关信息，如访问日志、API调用记录等，用于改进服务质量和用户体验。
            </Paragraph>
          </div>

          <div>
            <Title level={4}>二、信息的使用</Title>
            <Paragraph>
              我们收集的信息将用于：
            </Paragraph>
            <Paragraph>
              1. 提供和维护我们的服务；<br />
              2. 处理您的订单和支付；<br />
              3. 向您发送服务相关通知；<br />
              4. 改进我们的产品和服务；<br />
              5. 遵守法律法规的要求。
            </Paragraph>
          </div>

          <div>
            <Title level={4}>三、信息的存储和保护</Title>
            <Paragraph>
              <strong>1. 数据存储</strong><br />
              您的个人信息存储在位于中国境内的服务器上。未经您的同意，我们不会将您的个人信息传输至境外。
            </Paragraph>
            <Paragraph>
              <strong>2. 安全措施</strong><br />
              我们采用业界标准的安全措施保护您的个人信息，包括但不限于：<br />
              - 数据加密传输（HTTPS）<br />
              - 敏感信息加密存储<br />
              - 访问控制和权限管理<br />
              - 安全审计和监控
            </Paragraph>
            <Paragraph>
              <strong>3. 数据保留</strong><br />
              我们仅在实现本政策所述目的所需的期限内保留您的个人信息，法律法规另有规定的除外。
            </Paragraph>
          </div>

          <div>
            <Title level={4}>四、信息的共享</Title>
            <Paragraph>
              我们不会向第三方出售您的个人信息。我们仅在以下情况下共享您的信息：
            </Paragraph>
            <Paragraph>
              1. <strong>服务提供商</strong>：我们可能与服务提供商共享信息，以完成支付处理、数据分析等服务；<br />
              2. <strong>法律要求</strong>：根据法律法规、法律程序或政府要求，我们可能需要披露您的信息；<br />
              3. <strong>业务转让</strong>：如发生合并、收购或资产出售，您的信息可能作为交易的一部分被转让。
            </Paragraph>
          </div>

          <div>
            <Title level={4}>五、您的权利</Title>
            <Paragraph>
              您对您的个人信息享有以下权利：
            </Paragraph>
            <Paragraph>
              1. <strong>访问权</strong>：您有权访问我们持有的您的个人信息；<br />
              2. <strong>更正权</strong>：您有权要求更正不准确的个人信息；<br />
              3. <strong>删除权</strong>：您有权要求删除您的个人信息；<br />
              4. <strong>撤回同意权</strong>：您有权撤回之前给予我们的同意。
            </Paragraph>
          </div>

          <div>
            <Title level={4}>六、Cookie政策</Title>
            <Paragraph>
              我们使用Cookie和类似技术来改善您的使用体验。您可以通过浏览器设置管理Cookie。
            </Paragraph>
          </div>

          <div>
            <Title level={4}>七、未成年人保护</Title>
            <Paragraph>
              我们的服务面向成年人。如果您是未成年人，请在监护人的陪同下阅读本政策，并在取得监护人同意后使用我们的服务。
            </Paragraph>
          </div>

          <div>
            <Title level={4}>八、政策更新</Title>
            <Paragraph>
              我们可能会不时更新本隐私政策。更新后的政策将在本页面发布，重大变更时我们会通过站内信或邮件通知您。
            </Paragraph>
          </div>

          <div>
            <Title level={4}>九、联系我们</Title>
            <Paragraph>
              如果您对本隐私政策有任何疑问或建议，请通过以下方式联系我们：
            </Paragraph>
            <Paragraph>
              📧 邮箱：privacy@pintuotuo.com<br />
              📱 客服电话：400-888-8888<br />
              📪 地址：北京市海淀区中关村科技园
            </Paragraph>
            <Paragraph>
              我们将在15个工作日内回复您的请求。
            </Paragraph>
          </div>

          <Divider />

          <div style={{ textAlign: 'center' }}>
            <Text type="secondary">© 2026 拼团Token平台 版权所有</Text>
          </div>
        </Space>
      </Card>
    </div>
  )
}

export default PrivacyPolicyPage
