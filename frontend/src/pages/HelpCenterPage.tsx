import React from 'react'
import { Card, Typography, Collapse, Input, Space, Divider } from 'antd'
import { SearchOutlined, QuestionCircleOutlined, BookOutlined, CustomerServiceOutlined } from '@ant-design/icons'

const { Title, Text, Paragraph } = Typography
const { Panel } = Collapse

const faqCategories = [
  {
    category: '购买相关',
    icon: <BookOutlined />,
    faqs: [
      {
        question: '如何购买Token？',
        answer: '选择商品后，可以选择单独购买或拼团购买。拼团购买可以享受更优惠的价格，但需要等待成团成功后才能发货。支付成功后，Token将自动充值到您的账户。',
      },
      {
        question: '拼团购买有什么优势？',
        answer: '拼团购买可以享受比单独购买更优惠的价格，通常可以节省30%-50%。您可以选择2人团或5人团，人数越多价格越优惠。成团成功后Token立即到账。',
      },
      {
        question: 'Token有效期多久？',
        answer: 'Token有效期为1年，从购买之日起计算。过期后未使用的Token将自动作废，请及时使用。',
      },
      {
        question: '支持哪些支付方式？',
        answer: '目前支持支付宝和微信支付两种主流支付方式。更多支付方式正在接入中。',
      },
    ],
  },
  {
    category: '使用相关',
    icon: <CustomerServiceOutlined />,
    faqs: [
      {
        question: '如何使用Token？',
        answer: '购买成功后，Token会自动充值到您的账户。您可以通过API调用使用Token，支持多种主流AI模型。在"我的Token"页面可以查看余额和使用记录。',
      },
      {
        question: '支持哪些模型？',
        answer: '目前支持GLM-5、K2.5等多种主流AI模型，涵盖编码类、文本处理、多模态等场景。具体支持的模型列表请参考商品详情页。',
      },
      {
        question: '如何获取API密钥？',
        answer: '登录后，进入"我的Token"页面，点击"API密钥管理"，即可创建和管理您的API密钥。请妥善保管您的密钥，不要泄露给他人。',
      },
      {
        question: 'Token可以转让吗？',
        answer: 'Token暂不支持直接转让，但您可以通过API为其他项目或团队成员提供服务。',
      },
    ],
  },
  {
    category: '账户相关',
    icon: <QuestionCircleOutlined />,
    faqs: [
      {
        question: '如何注册账户？',
        answer: '点击页面右上角的"注册"按钮，填写手机号和验证码即可完成注册。注册成功后即可开始购买和使用Token。',
      },
      {
        question: '忘记密码怎么办？',
        answer: '点击登录页面的"忘记密码"，通过手机验证码重置密码。如果还有问题，请联系客服。',
      },
      {
        question: '如何成为商家？',
        answer: '如果您是AI服务提供商，可以申请成为商家。请联系我们的商务团队，提交相关资质材料进行审核。',
      },
    ],
  },
  {
    category: '退款相关',
    icon: <QuestionCircleOutlined />,
    faqs: [
      {
        question: '拼团失败怎么办？',
        answer: '如果拼团在规定时间内未能成功，系统会自动取消订单并全额退款。退款将在1-3个工作日内原路返回到您的支付账户。',
      },
      {
        question: '如何申请退款？',
        answer: '在订单列表页面，找到需要退款的订单，点击"退款"按钮，填写退款原因后提交申请。退款审核通过后，款项将在1-3个工作日内退还。',
      },
      {
        question: '退款多久到账？',
        answer: '退款申请审核通过后，款项将在1-3个工作日内原路退回到您的支付账户。具体到账时间取决于您的支付方式。',
      },
    ],
  },
]

export const HelpCenterPage: React.FC = () => {
  return (
    <div style={{ padding: '20px', maxWidth: 900, margin: '0 auto' }}>
      <Card>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div style={{ textAlign: 'center' }}>
            <Title level={2}>
              <QuestionCircleOutlined style={{ marginRight: 8 }} />
              帮助中心
            </Title>
            <Text type="secondary">常见问题解答与使用指南</Text>
          </div>

          <Input
            placeholder="搜索问题..."
            prefix={<SearchOutlined />}
            size="large"
            style={{ maxWidth: 500, margin: '0 auto', display: 'block' }}
          />

          <Divider />

          {faqCategories.map((category) => (
            <div key={category.category}>
              <Title level={4}>
                {category.icon}
                <span style={{ marginLeft: 8 }}>{category.category}</span>
              </Title>
              <Collapse accordion>
                {category.faqs.map((faq, index) => (
                  <Panel header={faq.question} key={index}>
                    <Paragraph>{faq.answer}</Paragraph>
                  </Panel>
                ))}
              </Collapse>
            </div>
          ))}

          <Divider />

          <Card style={{ background: '#f5f5f5' }}>
            <Title level={5}>没有找到答案？</Title>
            <Paragraph>
              如果您的问题没有在上面的列表中找到，请通过以下方式联系我们：
            </Paragraph>
            <Space direction="vertical">
              <Text>📧 邮箱：support@pintuotuo.com</Text>
              <Text>📱 客服电话：400-888-8888</Text>
              <Text>⏰ 工作时间：周一至周五 9:00-18:00</Text>
            </Space>
          </Card>
        </Space>
      </Card>
    </div>
  )
}

export default HelpCenterPage
