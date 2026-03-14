# 拼脱脱 技术架构与技术栈选型

**文档用途**：为技术团队提供完整的架构设计和技术选型指导

---

## 目录

1. [架构设计原则](#架构设计原则)
2. [系统整体架构](#系统整体架构)
3. [技术栈选型](#技术栈选型)
4. [核心模块设计](#核心模块设计)
5. [数据流与交互](#数据流与交互)
6. [高可用与性能](#高可用与性能)
7. [安全设计](#安全设计)
8. [部署架构](#部署架构)

---

## 架构设计原则

### 1. 高并发支持

**需求**：秒杀、拼团成团时可能瞬间涌入大量请求

**设计方案**：
- 使用Redis缓存实时数据（拼团进度、用户余额）
- Kafka消息队列异步处理非关键任务
- CDN加速静态资源，减轻服务器压力
- 多个API网关实例进行负载均衡

**目标指标**：
- 支持10,000+ QPS
- P99延迟 < 100ms

### 2. 数据一致性

**需求**：订单、支付、拼团等关键业务数据必须一致

**设计方案**：
- 使用ACID数据库事务保证强一致性
- 关键操作使用分布式事务（如支付后订单创建）
- 使用幂等性设计防止重复处理（如支付重复回调）
- 定期数据对账和审计

### 3. 可扩展性

**需求**：快速迭代产品、支持模型快速接入

**设计方案**：
- 微服务架构，模块独立部署和扩展
- 配置驱动，无需改代码支持新模型接入
- 使用消息队列解耦服务间依赖

### 4. 可维护性

**需求**：代码清晰、易于理解和修改

**设计方案**：
- 代码分层，业务逻辑与基础设施分离
- 清晰的接口定义和错误处理
- 完整的日志和监控，便于问题排查
- 编写自动化测试保证代码质量

---

## 系统整体架构

### 高层架构图

```
┌────────────────────────────────────────────────────────────────┐
│                     客户端层 (Client Layer)                     │
│  Web | iOS App | Android App | 商家后台 | Admin管理平台        │
└──┬───────────────────────────────────────┬──────────────────────┘
   │                                        │
   ↓                                        ↓
┌────────────────────────────────────────────────────────────────┐
│                   CDN & 负载均衡 (LB)                           │
│  Cloudflare/阿里云CDN | 负载均衡器 (Nginx/Ingress Controller)  │
└───────────────────────┬──────────────────────────────────────────┘
                        │
                        ↓
┌────────────────────────────────────────────────────────────────┐
│                   API网关层 (Gateway Layer)                     │
│ ┌──────────────────────────────────────────────────────────┐   │
│ │ API Gateway (Kong/Nginx)                                 │   │
│ │ - 路由转发                                               │   │
│ │ - 认证授权 (JWT验证)                                     │   │
│ │ - 限流 (Rate Limiting)                                   │   │
│ │ - 请求转换                                               │   │
│ │ - CORS处理                                               │   │
│ └──────────────────────────────────────────────────────────┘   │
└───────────────┬────────────────────────────────────────────────┘
                │
    ┌───────────┼───────────┬───────────┬───────────┐
    ↓           ↓           ↓           ↓           ↓
┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
│  用户   │ │  商家  │ │  订单  │ │  支付  │ │  API   │
│  服务  │ │  服务  │ │  服务  │ │  服务  │ │  网关  │
└────────┘ └────────┘ └────────┘ └────────┘ └────────┘
    │           │           │           │           │
    └───────────┼───────────┼───────────┼───────────┘
                │
                ↓
    ┌───────────────────────────────────────┐
    │  业务中间件层 (Middleware & Cache)    │
    │ ┌─────────────────────────────────┐   │
    │ │ Redis (缓存层)                  │   │
    │ │ - 用户余额缓存                  │   │
    │ │ - 拼团进度缓存                  │   │
    │ │ - 热点商品缓存                  │   │
    │ │ - Session存储                   │   │
    │ └─────────────────────────────────┘   │
    │ ┌─────────────────────────────────┐   │
    │ │ Kafka (消息队列)                │   │
    │ │ - 异步订单处理                  │   │
    │ │ - 消费记录积累                  │   │
    │ │ - 用户推送通知                  │   │
    │ │ - 数据分析上报                  │   │
    │ └─────────────────────────────────┘   │
    │ ┌─────────────────────────────────┐   │
    │ │ Elasticsearch (搜索)            │   │
    │ │ - 商品搜索                      │   │
    │ │ - 用户行为搜索                  │   │
    │ └─────────────────────────────────┘   │
    └──────┬──────────────────────────────────┘
           │
           ↓
┌────────────────────────────────────────────────────────────────┐
│                    数据存储层 (Data Layer)                      │
│                                                                │
│ ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐   │
│ │ 主数据库         │  │ 从数据库(副本)   │  │ 时间序列DB  │   │
│ │ PostgreSQL      │◄─┤ PostgreSQL      │  │ InfluxDB    │   │
│ │ (master-slave)   │  │ Read Replica    │  │ (监控数据)  │   │
│ │                  │  │                 │  │             │   │
│ │ - 用户信息       │  │ - 报表查询      │  │ - 性能指标  │   │
│ │ - 订单数据       │  │ - 分析查询      │  │ - 成本数据  │   │
│ │ - 消费记录       │  │                 │  │             │   │
│ │ - 拼团数据       │  │                 │  │             │   │
│ └─────────────────┘  │ 分析数据库      │  │             │   │
│                      │ ClickHouse      │  │             │   │
│ ┌─────────────────┐  └─────────────────┘  │             │   │
│ │ 文件存储        │                        │             │   │
│ │ S3/阿里OSS      │                        │             │   │
│ │ - 图片          │                        │             │   │
│ │ - 账单文件      │                        │             │   │
│ │ - 备份          │                        │             │   │
│ └─────────────────┘                        └──────────────┘   │
└────────────────────────────────────────────────────────────────┘
           │
           ↓
┌────────────────────────────────────────────────────────────────┐
│              外部服务集成 (External Services)                   │
│                                                                │
│ ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│ │ 支付网关      │  │ 短信/邮件    │  │ B端API服务   │         │
│ │ 支付宝/微信  │  │ 阿里云SMS    │  │ 模型厂商API  │         │
│ │            │  │ SendGrid     │  │ 智谱/Kimi   │         │
│ │ 同步回调    │  │            │  │ Claude/GPT │         │
│ └──────────────┘  └──────────────┘  │ MiniMax等   │         │
│                                      └──────────────┘         │
│ ┌──────────────────────────────────────────────────────┐     │
│ │ 监控告警 (Datadog/Prometheus + AlertManager)        │     │
│ │ 日志系统 (ELK Stack)                                │     │
│ │ 链路追踪 (Jaeger)                                   │     │
│ └──────────────────────────────────────────────────────┘     │
└────────────────────────────────────────────────────────────────┘
```

### 架构分层说明

**第一层：客户端层**
- Web前端 (React/Vue)
- iOS/Android App
- B端商家后台
- 运营/管理后台

**第二层：接入层**
- CDN (加速静态资源)
- 负载均衡 (分散流量)

**第三层：API网关层**
- 统一入口，进行认证、授权、限流
- 请求路由到不同的后端服务
- 实现跨域处理

**第四层：业务服务层**
- 独立的微服务（用户服务、订单服务、支付服务等）
- 每个服务有清晰的职责

**第五层：中间件层**
- 缓存 (Redis) - 提升性能
- 消息队列 (Kafka) - 异步处理
- 搜索引擎 (Elasticsearch) - 快速搜索

**第六层：数据存储层**
- 关系数据库 (PostgreSQL)
- 时间序列数据库 (InfluxDB)
- 文件存储 (S3/OSS)

**第七层：外部服务集成**
- 支付网关
- 第三方通知服务
- 模型厂商API

---

## 技术栈选型

### 后端开发

| 层级 | 技术选型 | 说明 |
|------|----------|------|
| **编程语言** | Go + Node.js | Go用于高性能服务，Node.js用于快速迭代的服务 |
| **框架** | Gin (Go) / Express (Node.js) | 轻量级、高性能 |
| **ORM** | GORM (Go) / Sequelize (Node.js) | 便于数据库操作 |
| **日志** | Zap (Go) / Winston (Node.js) | 结构化日志 |
| **监控** | Prometheus | 指标收集和监控 |
| **配置管理** | Consul / ETCD | 分布式配置管理 |

### 前端开发

| 类型 | 技术选型 | 说明 |
|------|----------|------|
| **Web前端** | React + TypeScript | 类型安全，组件丰富 |
| **状态管理** | Zustand / Redux | 简单易用的状态管理 |
| **HTTP客户端** | Axios | 请求库 |
| **UI框架** | Material-UI / Ant Design | 企业级UI组件库 |
| **App前端** | React Native / Flutter | 跨平台App开发 |

### 数据库

| 类型 | 技术选型 | 说明 |
|------|----------|------|
| **关系数据库** | PostgreSQL | 开源、稳定、功能完整 |
| **缓存** | Redis | 高性能缓存 |
| **消息队列** | Apache Kafka | 高吞吐量，支持大规模消费者 |
| **搜索引擎** | Elasticsearch | 全文搜索 |
| **时间序列DB** | InfluxDB / Prometheus | 性能监控数据 |
| **数据仓库** | ClickHouse | 分析型数据库，支持大规模数据分析 |

### 基础设施

| 组件 | 技术选型 | 说明 |
|------|----------|------|
| **容器化** | Docker | 应用容器化 |
| **编排** | Kubernetes (K8s) | 容器编排和自动扩展 |
| **服务网格** | Istio (可选) | 灰度发布、流量管理 |
| **CI/CD** | GitLab CI / GitHub Actions | 自动化构建和部署 |
| **API网关** | Kong / Nginx | 请求路由和限流 |
| **CDN** | Cloudflare / 阿里云CDN | 加速静态资源 |
| **对象存储** | AWS S3 / 阿里云OSS | 图片、文件存储 |

---

## 核心模块设计

### 1. 用户服务 (User Service)

**职责**：
- 用户注册、登录、认证
- 用户信息管理
- 用户信用评分计算
- 用户标签维护

**核心API**：
```
POST /users/register           # 注册
POST /users/login              # 登录
GET  /users/me                 # 获取当前用户信息
PUT  /users/me                 # 修改用户信息
GET  /users/{id}               # 获取指定用户信息
PUT  /users/{id}/credit-score  # 更新信用评分
```

**核心数据**：
- users表：用户基本信息
- user_tags表：用户标签
- user_sessions表：用户会话

**依赖服务**：
- 无

### 2. 商品服务 (Product Service)

**职责**：
- 商品SKU管理
- 拼团规则配置
- 商品搜索和推荐
- 库存管理

**核心API**：
```
GET    /products                      # 获取商品列表
GET    /products/{id}                 # 获取商品详情
POST   /merchants/products            # 创建SKU(B端)
PUT    /merchants/products/{id}       # 编辑SKU(B端)
POST   /merchants/products/{id}/rules # 创建拼团规则(B端)
```

**核心数据**：
- products表：SKU信息
- group_rules表：拼团规则
- product_inventory表：库存

**依赖服务**：
- 无

### 3. 订单服务 (Order Service)

**职责**：
- 订单创建和管理
- 订单状态流转
- 订单查询和统计

**核心API**：
```
POST   /orders                     # 创建订单
GET    /orders/{id}                # 获取订单详情
GET    /orders                     # 获取订单列表
DELETE /orders/{id}                # 取消订单
```

**核心数据**：
- orders表：订单主表
- order_items表：订单项

**依赖服务**：
- 支付服务（支付回调）
- 拼团服务（检查拼团状态）
- Token服务（Token分配）

**关键流程**：
```
创建订单 → 选择支付方式 → 调用支付服务 → 支付成功回调
         ↓
      如果是拼团：检查拼团是否成团
         ↓
      如果成团或单独购买：分配Token，更新订单状态为已完成
```

### 4. 拼团服务 (Group Service)

**职责**：
- 拼团创建和管理
- 拼团成团判断
- 拼团自动补贴逻辑
- 拼团成员管理

**核心API**：
```
POST   /groups                           # 创建拼团
GET    /groups/{id}                      # 获取拼团详情
GET    /groups/{id}/progress             # 获取拼团进度
POST   /groups/{id}/join                 # 加入拼团
DELETE /groups/{id}                      # 取消拼团
POST   /groups/{id}/check-completion     # 检查是否成团(内部接口)
```

**核心数据**：
- groups表：拼团主表
- group_members表：拼团成员
- group_rules表：拼团规则

**依赖服务**：
- 无（异步调用订单服务更新订单状态）

**关键实现**：
```go
// 拼团成团检查逻辑
func (s *GroupService) CheckCompletion(groupID int64) {
    group := s.GetGroup(groupID)

    // 条件1：人数满足
    if group.CurrentCount >= group.TargetCount {
        s.MarkAsSuccess(groupID)
        s.NotifyMembers(groupID, "成团成功")
        return
    }

    // 条件2：时间到期
    if time.Now() > group.Deadline {
        if group.CurrentCount >= 2 { // 最少成团人数
            // 自动成团，触发补贴逻辑
            s.MarkAsSuccess(groupID)
            s.ProcessSubsidy(groupID) // 平台或B端补贴
        } else {
            // 拼团失败，处理退款
            s.MarkAsFailed(groupID)
            s.ProcessRefunds(groupID)
        }
    }
}
```

### 5. Token服务 (Token Service)

**职责**：
- Token充值和余额管理
- 实时扣费
- API Key管理
- 消费记录记录

**核心API**：
```
GET    /tokens/balance              # 获取用户Token余额
POST   /tokens/transfer             # Token转账(内部接口)
GET    /tokens/keys                 # 获取API Key列表
POST   /tokens/keys                 # 创建API Key(B端)
PUT    /tokens/keys/{id}            # 编辑API Key
DELETE /tokens/keys/{id}            # 删除API Key
GET    /tokens/consumption          # 获取消费明细
```

**核心数据**：
- user_balances表：用户余额
- api_keys表：API密钥
- consumption_records表：消费记录

**依赖服务**：
- 无

**关键流程（API调用时的扣费）**：
```
用户发起API请求 → API网关验证 Token余额充足?
                          ↓ (是)
                    转发请求到B端API
                          ↓
                    B端API返回结果 (包含消耗Token)
                          ↓
                    实时扣费（Token Service）
                          ↓
                    异步记录消费明细（Kafka消息）
```

### 6. 支付服务 (Payment Service)

**职责**：
- 支付网关集成
- 支付状态管理
- 支付结果回调处理
- 退款处理

**核心API**：
```
POST   /payments/initiate           # 发起支付
POST   /payments/callback/alipay    # 支付宝回调
POST   /payments/callback/wechat    # 微信回调
POST   /payments/refund/{id}        # 退款
```

**核心数据**：
- payments表：支付记录
- refunds表：退款记录

**依赖服务**：
- 订单服务（更新订单状态）
- Token服务（扣费或退款）

**关键实现**：
```
幂等性设计：每笔支付都有唯一的payment_no
  → 支付成功回调来多次，系统自动去重，避免重复扣费

状态机设计：
  pending → processing → success / failed
  success → refunded (退款)
```

### 7. API网关 (API Gateway Service)

**职责**：
- 请求路由
- JWT认证
- 限流
- 请求/响应转换
- B端API密钥管理和调用转发

**核心流程**：
```
C端请求 (带JWT Token)
    ↓
网关验证JWT是否合法
    ↓
检查请求限流是否超出
    ↓
路由到对应的后端服务
    ↓
返回响应

B端API请求 (用户API调用)
    ↓
网关验证API Key合法性
    ↓
查询用户Token余额是否充足
    ↓
识别使用的模型和B端API Key
    ↓
通过负载均衡器选择可用的API Key
    ↓
转发请求到B端API
    ↓
接收B端响应，统计消耗的Token
    ↓
异步调用Token Service进行扣费
    ↓
返回给用户
```

---

## 数据流与交互

### 用户购买Token的数据流

```
[C端App] 用户点击"拼团购买"
    ↓
[API Gateway] 请求进入
    ├─ 验证JWT Token
    ├─ 限流检查
    └─ 路由到订单服务
    ↓
[Order Service] 创建订单
    ├─ 保存到orders表
    ├─ 订单状态：pending_payment
    └─ 返回订单ID给前端
    ↓
[C端App] 用户选择支付方式（支付宝/微信）
    ↓
[Payment Service] 发起支付
    ├─ 调用支付宝/微信API
    ├─ 返回支付链接/二维码给用户
    └─ 创建payment记录
    ↓
[用户] 在支付宝/微信完成支付
    ↓
[支付网关] 支付成功，回调通知
    ↓
[Payment Service] 处理回调
    ├─ 验证签名合法性
    ├─ 幂等性检查（防重复）
    ├─ 更新payment状态为success
    ├─ 发消息到Kafka（payment_completed事件）
    └─ 返回200给支付网关
    ↓
[订单服务] 监听Kafka事件，处理支付成功
    ├─ 如果是单独购买：
    │  ├─ 直接调用Token Service分配API Key
    │  └─ 更新订单状态为completed
    │
    └─ 如果是拼团购买：
       ├─ 检查拼团是否成团（调用Group Service）
       ├─ 如果已成团：
       │  ├─ 为所有成员分配API Key
       │  └─ 更新所有订单状态为completed
       └─ 如果未成团：
          └─ 等待（不动作）
    ↓
[Group Service] 拼团监听器
    ├─ 定时检查是否成团（每30秒）
    ├─ 如果人数满足或时间到期：
    │  ├─ 成团成功：通知所有成员
    │  └─ 成团失败：触发补贴逻辑或退款逻辑
    └─ 发消息到Kafka（group_success或group_failed）
    ↓
[用户] 在App中看到
    ├─ 订单完成
    ├─ 收到通知（邮件、App推送）
    └─ 在"我的Token"中看到新的API Key
```

### 用户调用API消耗Token的数据流

```
[User App] 调用API
GET /v1/chat/completions
Headers: Authorization: Bearer sk_****XXXX
    ↓
[API Gateway] 请求进入
    ├─ 验证API Key合法性 (查Redis缓存或DB)
    ├─ 检查用户余额是否充足 (查Redis缓存)
    ├─ 限流检查 (单用户QPS)
    └─ 路由到API网关模块处理
    ↓
[API Gateway Service] 处理API调用
    ├─ 识别模型和单价
    ├─ 通过负载均衡选择可用的B端API Key
    ├─ 转发请求到B端API
    │  (如智谱、Kimi、Claude等)
    └─ 等待B端API返回
    ↓
[B端API] 返回结果
    包含：
    ├─ 输入Token数量
    ├─ 输出Token数量
    └─ 总消耗Token数
    ↓
[API Gateway Service] 接收返回
    ├─ 计算费用：消耗Token × 单价
    ├─ 实时更新Redis中的用户余额
    ├─ 发消息到Kafka：token_consumed事件
    └─ 返回结果给用户
    ↓
[Kafka Consumer] 异步处理
    ├─ Token Service处理token_consumed事件
    │  ├─ 从用户余额中扣费
    │  └─ 更新数据库user_balances表
    │
    └─ Consumption Service处理
       ├─ 记录消费明细到consumption_records表
       ├─ 累计B端的营收（用于月度结算）
       └─ 发送数据到ClickHouse（分析）
    ↓
[用户] 实时看到
    ├─ Token余额实时更新（从Redis读）
    └─ 在消费明细中看到新的消费记录（次日可见）
```

---

## 高可用与性能

### 1. 缓存策略

**多层缓存设计**：

```
第1层（浏览器缓存）
    ↓
第2层（CDN缓存）：静态资源 (图片、JS/CSS)
    ↓
第3层（Redis缓存）：热数据
    ├─ 用户余额 (user:{user_id}:balance)
    │  TTL: 实时更新，不设TTL
    │  更新时机: Token消耗时实时写入
    ├─ 拼团进度 (group:{group_id}:progress)
    │  TTL: 5分钟或成团时自动更新
    ├─ 商品信息缓存 (product:{product_id}:info)
    │  TTL: 1小时
    ├─ 热销商品列表 (products:hot)
    │  TTL: 5分钟
    └─ 用户会话 (session:{session_id})
       TTL: 7天
    ↓
第4层（数据库）：冷数据
    └─ 历史数据、不常访问的数据
```

**缓存更新策略**：
- **主动更新**（Write-Through）：订单创建时立即更新缓存
- **被动更新**（Write-Behind）：异步更新缓存
- **缓存失效**（TTL）：设置过期时间自动失效

### 2. 数据库优化

**读写分离**：
```
写入：路由到Master (PostgreSQL)
读取（实时性要求高）：Master
读取（实时性要求低）：Read Replica
```

**分表分库** (当数据量大时)：
```
订单表：按日期分表
  orders_202603, orders_202604, ...

消费记录表：按日期分表
  consumption_202603, consumption_202604, ...

优势：
  ├─ 减少单表数据量，加快查询速度
  ├─ 便于历史数据归档
  └─ 支持并行处理
```

**索引优化**：
- 在WHERE条件和JOIN条件的字段上创建索引
- 避免在索引字段上使用函数
- 定期分析和优化查询执行计划

### 3. 异步处理

**使用Kafka处理非实时任务**：

```
实时任务（同步）：
  ├─ 订单创建
  ├─ 支付处理
  ├─ 拼团成团判断
  └─ Token余额扣费

非实时任务（异步 via Kafka）：
  ├─ 消费记录积累 (可延迟1分钟)
  ├─ 用户通知发送 (可延迟5分钟)
  ├─ 数据分析上报 (可延迟1小时)
  ├─ 邮件发送 (可延迟30分钟)
  └─ 月度账单生成 (后台定时任务)
```

**Kafka Topic设计**：
```
payment_completed    # 支付完成
order_created        # 订单创建
group_success        # 拼团成功
token_consumed       # Token消耗
user_notification    # 用户通知
analytics_event      # 分析事件
```

### 4. 并发控制

**在拼团成团时的并发控制**：

```
场景：拼团人数满足，多个请求同时到达

问题：如果没有并发控制，可能会重复处理成团逻辑

解决方案：使用Redis分布式锁

实现代码伪代码：
def check_group_completion(group_id):
    lock_key = f"group:{group_id}:lock"

    # 尝试获取分布式锁（10秒超时）
    if redis.set(lock_key, 1, ex=10, nx=True):
        try:
            # 获取锁成功，执行成团处理
            group = db.get_group(group_id)
            if group.current_count >= group.target_count:
                mark_as_success(group_id)
                notify_members(group_id)
        finally:
            # 释放锁
            redis.delete(lock_key)
    else:
        # 没有获取到锁，说明其他线程正在处理，直接返回
        return
```

### 5. 监控和告警

**关键监控指标**：

```
系统级指标：
  ├─ CPU使用率 > 80% → 告警
  ├─ 内存使用率 > 85% → 告警
  ├─ 磁盘使用率 > 90% → 告警
  └─ 网络带宽使用 > 80% → 告警

应用级指标：
  ├─ API响应时间 P99 > 100ms → 告警
  ├─ 错误率 > 0.1% → 告警
  ├─ 缓存命中率 < 80% → 告警
  └─ 数据库连接池占用 > 90% → 告警

业务级指标：
  ├─ 订单成功率 < 99% → 告警
  ├─ 支付失败率 > 1% → 告警
  ├─ 拼团成团率 < 90% → 告警
  └─ Token余额异常 → 告警

日志监控：
  └─ 关键错误日志 → 立即通知
```

---

## 安全设计

### 1. 认证和授权

**用户认证**：
```
POST /login → 获得JWT Token
  JWT包含：
  ├─ user_id
  ├─ user_email
  ├─ exp (过期时间，7天)
  └─ 签名 (HMAC-SHA256)

后续请求带上Token：
  Authorization: Bearer {jwt_token}

API Gateway验证：
  ├─ 验证签名有效性
  ├─ 检查是否过期
  └─ 路由到服务
```

**B端API Key认证**：
```
用户调用API时：
  Authorization: Bearer sk_****XXXX

API Gateway验证：
  ├─ 查询api_keys表，验证key_hash是否匹配
  ├─ 检查key是否被禁用
  ├─ 检查月度配额是否超出
  └─ 路由到API网关
```

**授权控制**：
```
用户只能查看自己的订单、消费记录等
  → 在所有查询中加上 user_id = current_user_id

商家只能查看自己的商品、订单等
  → 在所有查询中加上 merchant_id = current_merchant_id
```

### 2. 数据安全

**密码存储**：
```
使用bcrypt加密存储，salt长度至少10轮
  password_hash = bcrypt(password, salt_rounds=10)
```

**API Key存储**：
```
不存储明文API Key，只存储加密后的哈希值
  key_hash = sha256(api_key)

用户复制Key时，需要再次输入密码验证
```

**敏感数据加密**：
```
银行账号：AES-256加密存储
身份证号：AES-256加密存储
```

### 3. 防作弊机制

**邀请防刷**：
```
检查：
  ├─ 同一用户短时间内不能邀请太多人（如5分钟内最多10人）
  ├─ 同一设备、IP不能邀请自己
  ├─ 邀请链接有时间限制（7天）
  └─ 邀请返利金额有上限（月度累计最多¥500）

处理：
  └─ 超过阈值的账户 → 标记为异常 → 冻结返利 → 人工审核
```

**支付防刷**：
```
检查：
  ├─ 同一用户在短时间内不能支付多个拼团订单（除非分别来自不同商品）
  ├─ 同一张卡在短时间内支付金额上限
  └─ 支付金额异常波动检测

处理：
  └─ 触发异常 → 增加验证难度（如短信验证码）
```

**API调用防刷**：
```
限流规则：
  ├─ 全局限流：全平台最多50,000 QPS
  ├─ 单用户限流：单用户最多100 QPS
  ├─ 单IP限流：单IP最多1,000 QPS
  └─ 单API限流：某些敏感API限流更严格

策略：
  └─ 超限 → 返回429 Too Many Requests → 客户端重试
```

### 4. 密钥管理

**API Key托管**：
```
密钥存储位置：
  ├─ 不存储在代码中
  ├─ 不存储在配置文件中（如果服务器泄露）
  └─ 存储在专用的密钥管理系统 (如HashiCorp Vault)

访问控制：
  ├─ 只有API网关可以访问密钥存储
  ├─ 所有访问操作都被记录
  └─ 定期轮换密钥（如每90天）
```

---

## 部署架构

### 本地开发环境

```
开发机 → Docker Compose
  ├─ API服务（Go/Node.js）
  ├─ PostgreSQL
  ├─ Redis
  └─ Kafka

使用docker-compose.yml一键启动所有服务
```

### 测试环境

```
云平台（如阿里云ECS）
  ├─ 1台API Server (2核4GB)
  ├─ 1台PostgreSQL (4核8GB)
  ├─ 1台Redis (2核4GB)
  ├─ 1台Kafka (4核8GB)
  └─ Nginx负载均衡

或使用K8s集群（推荐）
  └─ 多个Pod副本，自动扩展
```

### 生产环境

**高可用部署架构**：

```
┌─────────────────────────────────────────────────────────┐
│                  用户请求                                │
└──────────────┬──────────────────────────────────────────┘
               │
               ↓
        ┌──────────────┐
        │ 负载均衡 (LB) │
        │ (Nginx ALB)  │
        └────┬─────┬──┘
             │     │
    ┌────────┘     └────────┐
    ↓                       ↓
┌────────────────┐    ┌────────────────┐
│ K8s Cluster 1  │    │ K8s Cluster 2  │  (多区域)
│ (Region A)     │    │ (Region B)     │
│                │    │                │
│ ┌────────────┐ │    │ ┌────────────┐ │
│ │ Pod Pod    │ │    │ │ Pod Pod    │ │
│ │ API API    │ │    │ │ API API    │ │
│ │ Service    │ │    │ │ Service    │ │
│ └────────────┘ │    │ └────────────┘ │
└────────────────┘    └────────────────┘
         │                     │
         └──────────┬──────────┘
                    │
                    ↓
         ┌──────────────────────┐
         │   数据层（区域内）    │
         │                      │
         │ Master PostgreSQL    │
         │ └─ Streaming Replica │
         │    (另一区域)        │
         │                      │
         │ Redis Cluster        │
         │ └─ 3 Master + 3 Slave│
         │                      │
         │ Kafka Cluster        │
         │ └─ 3 Broker + HA     │
         └──────────────────────┘
```

**部署流程**：

```
1. 代码提交到GitHub/GitLab
   ↓
2. CI/CD Pipeline触发
   ├─ 运行单元测试
   ├─ 运行集成测试
   ├─ 代码质量检查
   └─ 构建Docker镜像
   ↓
3. 推送镜像到Docker Registry
   ↓
4. Kubernetes自动部署
   ├─ 更新Deployment
   ├─ 启动新Pod
   ├─ 等待就绪
   └─ 逐个关闭旧Pod (滚动更新)
   ↓
5. 健康检查
   ├─ 检查新Pod是否健康
   ├─ 检查服务是否可用
   └─ 监控关键指标
```

**灾难恢复 (DR)**：

```
备份策略：
  ├─ 数据库每日备份（保留30天）
  ├─ 配置文件备份到Git
  └─ 关键数据同步到异地存储

故障恢复：
  ├─ 自动故障转移（Failover）
  ├─ 同区域副本切换（秒级）
  ├─ 异区域恢复（分钟级）
  └─ 人工干预决策时间 < 5分钟
```

---

**END OF DOCUMENT**
