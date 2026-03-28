import React from 'react';
import { Card, Typography, Divider, Space } from 'antd';

const { Title, Text, Paragraph } = Typography;

export const UserAgreementPage: React.FC = () => {
  return (
    <div style={{ padding: '20px', maxWidth: 900, margin: '0 auto' }}>
      <Card>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div style={{ textAlign: 'center' }}>
            <Title level={2}>用户服务协议</Title>
            <Text type="secondary">最后更新日期：2026年3月1日</Text>
          </div>

          <Divider />

          <div>
            <Title level={4}>一、服务条款的确认和接纳</Title>
            <Paragraph>
              拼团Token平台（以下简称"本平台"）的各项服务的所有权和运营权归本平台所有。用户在使用本平台提供的各项服务之前，应仔细阅读本服务协议。
            </Paragraph>
            <Paragraph>
              如用户不同意本服务协议及/或随时对其的修改，用户可以不使用或主动取消本平台提供的服务。用户一旦使用本平台服务，即视为用户已了解并完全同意本服务协议各项内容。
            </Paragraph>
          </div>

          <div>
            <Title level={4}>二、用户注册</Title>
            <Paragraph>
              1.
              用户注册成功后，本平台将给予每个用户一个用户账号及相应的密码，该用户账号和密码由用户负责保管；用户应当对以其用户账号进行的所有活动和事件负法律责任。
            </Paragraph>
            <Paragraph>
              2.
              用户对以其用户账号和密码进行的所有活动和事件负法律责任，包括但不限于数据的修改、发布的言论等。
            </Paragraph>
            <Paragraph>3. 用户发现其账号被盗用或存在安全漏洞的情况，请立即通知本平台。</Paragraph>
          </div>

          <div>
            <Title level={4}>三、服务内容</Title>
            <Paragraph>1. 本平台提供AI模型Token的购买、使用及相关服务。</Paragraph>
            <Paragraph>
              2.
              本平台有权在必要时修改服务条款，服务条款一旦发生变动，将会在重要页面上提示修改内容。
            </Paragraph>
            <Paragraph>3. 用户在使用本平台服务过程中，必须遵循以下原则：</Paragraph>
            <Paragraph>
              (1) 遵守中国有关的法律和法规；
              <br />
              (2) 不得为任何非法目的而使用网络服务系统；
              <br />
              (3) 遵守所有与网络服务有关的网络协议、规定和程序；
              <br />
              (4) 不得利用本平台服务进行任何可能对互联网正常运转造成不利影响的行为。
            </Paragraph>
          </div>

          <div>
            <Title level={4}>四、购买与退款</Title>
            <Paragraph>
              1. 用户通过本平台购买Token服务，支付成功后Token将自动充值到用户账户。
            </Paragraph>
            <Paragraph>
              2. 拼团购买模式下，如拼团未能在规定时间内成功，系统将自动取消订单并全额退款。
            </Paragraph>
            <Paragraph>3. 已使用的Token不支持退款，未使用的Token可在有效期内申请退款。</Paragraph>
            <Paragraph>4. 退款申请审核通过后，款项将在1-3个工作日内原路退回。</Paragraph>
          </div>

          <div>
            <Title level={4}>五、免责声明</Title>
            <Paragraph>
              1. 用户明确同意其使用本平台网络服务所存在的风险将完全由其自己承担。
            </Paragraph>
            <Paragraph>
              2.
              本平台不担保服务一定能满足用户的要求，也不担保服务不会中断，对服务的及时性、安全性、准确性也都不作担保。
            </Paragraph>
            <Paragraph>
              3.
              本平台不保证为向用户提供便利而设置的外部链接的准确性和完整性，同时，对于该等外部链接指向的不由本平台实际控制的任何网页上的内容，本平台不承担任何责任。
            </Paragraph>
          </div>

          <div>
            <Title level={4}>六、法律管辖</Title>
            <Paragraph>
              本协议的订立、执行和解释及争议的解决均应适用中国法律。如双方就本协议内容或其执行发生任何争议，双方应尽力友好协商解决；协商不成时，任何一方均可向本平台所在地人民法院提起诉讼。
            </Paragraph>
          </div>

          <Divider />

          <div style={{ textAlign: 'center' }}>
            <Text type="secondary">如有疑问，请联系客服：support@pintuotuo.com</Text>
          </div>
        </Space>
      </Card>
    </div>
  );
};

export default UserAgreementPage;
