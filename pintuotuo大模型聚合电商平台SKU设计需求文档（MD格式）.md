# pintuotuo大模型聚合电商平台SKU设计需求文档（MD格式）

|文档版本|修改日期|修改人|修改内容|

|---|---|---|---|

|V1.0|2026-03-29|AI PM|初始版本创建，基于2026年行业调研|

# 1. 行业背景与核心战略

根据2026年3月的最新行业动态，云计算行业已告别“只降不涨”的历史。阿里云、腾讯云、百度智能云在近期集体上调AI算力和Token价格，涨幅最高达450% 。核心动因在于：

#供应链成本飙升：HBM内存短缺，服务器级DRAM合约价单季翻倍，数据中心电力成本占比高达60% 。

#需求结构转变：推理需求占总算力需求首次突破70%，进入Agent时代（如OpenClaw），单个任务的Token消耗量是传统问答的数十倍 。

#商业模式迁移：头部厂商正从“出租算力”向“出售Token”迁移，模型能力（而非纯低价）决定定价权 。

#产品战略定位：作为平台方，我们的核心价值在于对冲供应链波动和降低B端商户接入复杂度。通过“拼团”模式聚合C端长尾流量，利用规模效应获取上游更低折扣（或应对涨价），同时为B端商户提供标准化的API分销渠道。

# 2. 核心产品架构：统一接入与SKU映射

为了满足B端商户“不改变当前API计费模式”的核心诉求，我们需要构建一个 “统一计量与映射层” 。

## 2.1 适配器模型设计

我们不能强制要求商户（如阿里云、火山引擎）改变他们的API规范，而是由平台主动适配。采用类似“AI API Aggregation”的架构 。

|上游厂商|计费特征|平台适配策略|

|---|---|---|

|火山引擎|分段计价（按输入长度32k/128k/256k区间）、缓存存储费、缓存命中优惠 |智能路由：将C端用户的请求，根据其购买的套餐类型（如“短文本专用”），自动路由到对应的价格区间，避免因输入长度超限导致B端商户成本溢出。|

|百度千帆|区分在线推理/批量推理、深度思考模式单独计费、搜索增强服务费 |功能封装：将“搜索增强”、“深度思考”包装为独立的“增值服务SKU”，而非基础Token包。|

|阿里/腾讯|阶梯定价、高峰时段（隐性）溢价、自研芯片成本对冲 |预付费缓冲：通过平台预购大量Token池，对冲上游涨价风险，向B端商户提供稳定的结算汇率。|

## 2.2 SKU标准化映射

平台的SKU设计不能照搬上游的复杂技术参数，必须转化为电商语言。

基础单位：定义平台内部通用积分 —— “算力点”。

规则：1 算力点 = 1K Token（标准上下文）。

汇率机制：针对不同模型的成本差异，设定消耗系数。例如：调用DeepSeek-V3.2消耗1.0系数；调用ERNIE 5.0消耗3.0系数 。

# 3. 商品SKU体系设计

SKU设计分为三层：模型分类层、套餐计费层、增值服务层。

## 3.1 模型分类（满足B端/主流厂商）

参考火山引擎和百度千帆的分类逻辑 ，将模型分为四个象限：

旗舰版（Pro）：对应 doubao-seed-2.0-pro、ERNIE 5.0。适用于复杂推理、代码生成。定价策略：高单价，主推按量付费或高额包月。

标准版（Lite）：对应 DeepSeek-V3、GLM-4。适用于日常对话、内容创作。定价策略：主力走量SKU，适合拼团。

轻量版（Mini/Flash）：对应 doubao-seed-1.6-flash、ERNIE Speed。适用于实时交互、分类、提取。定价策略：极低价，甚至免费试用，用于引流。

多模态/视觉版（Vision）：对应 Nano Banana（图像生成）、doubao-vision。定价策略：按张计费或按Token混合计费 。

## 3.2 套餐计费方式（满足C端体验）

结合“拼团”电商属性，设计以下五种计费模式：

### A. 按量付费（Pay-as-you-go）

产品形态：账户余额直接扣费。

B端适配：直接映射上游的API按量计费模式，对于B端商户而言，这等同于他们现有的模式，无缝衔接。

C端体验：适合开发测试，随用随付。

### B. 包月/季/年套餐（Subscription）

这是电商平台的核心利润来源。基于“限量/限额”原则设计，避免被C端用户滥用造成B端成本失控。

入门型（Token包）：

规格：100万Tokens/月（约等于基础对话5000-10000次）。

定价：9.9元/月（通过拼团价降至5.9元）。

逻辑：限量模式，超出后自动转为按量付费或限速。参考腾讯云涨价前的逻辑，这种门槛极低的方式易转化 。

专业型（无限量/软限制）：

产品形态：“拼团无限卡”。

机制：“公平使用原则”。在高峰时段，无限量用户可能会被降权至标准队列，优先保障按量付费高价值用户的质量。这对应了推理需求爆发后，厂商保障核心生态的做法 。

注意：真正的无限量在供应链成本暴涨的今天几乎不可持续，必须在产品文案中明确“正常使用”范围。

企业型（预留并发）：

针对SaaS商家或开发者。

支持购买 TPM（Tokens Per Minute） 或 RPM（Requests Per Minute） 配额 。

B端价值：这模拟了华为云MaaS的自定义接入点功能，让B端商户可以为其下游客户（C端用户）锁定算力资源。

### C. 混合计费（阶梯+拼团）

机制：用户购买“拼团主卡”后，可以邀请好友（拼团）。每成功邀请一人，所有成员获得额外的Token额度（如每人加赠10万Token）。

实现：通过智能合约（或后台逻辑）动态修改用户的Token余额。

### D. 免费试用

策略：新用户注册送1000 Tokens（约等价于0.8-3元）。

风控：利用“接入点流量控制”逻辑，对免费用户进行严格的TPM限流，防止被刷单（参考火山引擎限流逻辑）。

### E. 未来计费方式（前瞻性）

按效付费：参考蚂蚁数科的模式，未来可尝试“按生成成功的营销文案数量”、“按完成的任务节点”付费，而非单纯按Token消耗 。这在SKU设计中预留字段：billing_type: token | success_rate | task。

# 4. 产品功能设计细节

## 4.1 管理运营端设计

成本中心：

实时监控各上游厂商（B端商户）的API调用成本。

利用阿里/腾讯等厂商的“批量推理”低价通道（百度千帆批量推理价格仅为在线推理的40%），通过异步任务队列降低平台成本，包装成“夜间特惠包”卖给C端。

定价引擎配置：

支持按“模型 + 输入长度区间 + 时间段”配置动态价格。

例如：针对doubao-seed-2.0-pro，输入>128K的部分，平台加收30%的调度费，以覆盖B端商户的高额成本 。

## 4.2 商户端设计（B端大模型厂家视角）

自定义接入点：

允许B端商户在平台创建专属接入点。平台提供类似华为云的API Key管理和限流策略 。

透明的账单系统：

虽然B端商户自己已有账单，但平台需要提供“分账账单”。展示：C端用户A购买了9.9元套餐 -> 调用B端商户模型X -> 平台消耗成本Y元 -> 平台利润Z元。这能极大增强B端商户对平台的信任。

## 4.3 用户端设计（C端体验）

可视化Token消耗：

将抽象的Token转化为具体的使用次数。例如：“剩余Token可生成20篇小红书文案”或“可进行50次长文档翻译”。

智能路由选择：

在用户发起请求时，系统自动根据“任务类型”推荐模型。

例如：用户上传一张图片，系统检测到是发票，自动路由到doubao-vision或Qianfan-VL进行OCR识别，而不是走昂贵的ERNIE 5.0 。这需要后端具备类似API Aggregator的“意图识别”能力 。

# 5. 技术实现关键点（指导开发）

统一计量模块：

必须实现精确的Token计数（Token Counter）。由于不同厂商（如OpenAI格式 vs 百度格式）对Token的定义略有差异，平台需统一按tiktoken或厂商官方SDK统计，作为对账依据。

并发与限流：

实现分布式限流器。支持多维度限流：用户级、API Key级、接入点级。参考华为云MaaS的设计，防止一个C端用户的死循环代码打爆整个B端商户的配额 。

缓存优化：

利用“缓存存储”机制。对于频繁请求的System Prompt或长文档（如法律条文），利用上游厂商（如火山引擎）的缓存特性，将成本从输入单价降低到缓存命中单价，这部分利润可以作为平台的收益，也可以返还给用户作为折扣 。

订单与SKU引擎：

采用“SPU（标准产品单元）+ SKU（库存量单位）”模型。

SPU：DeepSeek-V3.2 模型服务。

SKU：DeepSeek-V3.2 100万Token包；DeepSeek-V3.2 包月无限量版；DeepSeek-V3.2 专属并发版。

# 6. 实施路线图

Phase 1 (MVP)：对接2-3家主流厂商（如阿里、字节），上线按量付费和基础Token包。核心逻辑：平台作为代理，赚取上游折扣与零售价的差价。重点完成统一计量与对账系统。

Phase 2 (规模化)：上线拼团功能和订阅制（包月） 。引入限流引擎，防止无限量套餐被滥用。增加“多模态”SKU。

Phase 3 (生态化)：开放“自定义接入点”给B端商户，允许他们在平台上架自己的私有模型或微调模型。引入“按效付费”等创新计费模式，并利用大数据分析帮助B端商户优化模型成本（FinOps）。

# 结语

在2026年这个“Token即石油”的时代，作为平台方，我们不再是简单的API转发器，而是AI资源的调度师。通过上述SKU设计，我们将复杂的供应链成本结构封装成标准、易用的电商商品，既满足了B端厂商回血/盈利的需求，也降低了C端用户拥抱AI的门槛。


附录：

1. SKU 与产品的概念界定
在电商领域，通常使用 SPU（标准产品单元） 和 SKU（库存量单位） 两层模型来描述商品。

SPU（产品）：是“商品”的抽象概念，代表用户“想要买什么”。它聚合了功能相同、仅规格有差异的一组商品。

例如：DeepSeek-V3.2 模型服务 是一个 SPU。它描述了模型的能力、厂商、技术参数等通用属性。

SKU（规格/库存单位）：是“商品”的具体可售卖变体，代表用户“具体买哪一个配置”。它定义了价格、计费周期、权益限制等交易属性。

例如：DeepSeek-V3.2 100万Tokens包（一次性）、DeepSeek-V3.2 包月无限量版、DeepSeek-V3.2 企业专属并发包 分别是三个不同的 SKU。

核心关系：一个 SPU 可以拥有多个 SKU。SKU 继承了 SPU 的基础信息（如模型 ID、厂商），但附加了定价、计费模式、有效期、限量/无限量等具体销售属性。

2. 为什么必须独立建设两张数据表？
答案是：必须独立设计 SPU 表和 SKU 表，两者缺一不可。 原因如下：

2.1 维护成本与复用性
SPU 是稳定的基础信息：模型厂商、模型版本、技术文档、接入地址、计费系数等，属于一次性配置，所有关联 SKU 共享。

SKU 是动态的交易配置：价格、促销活动、上下架状态、库存（配额）经常变动。如果混在一张表里，每次调整 SKU 都需要复制或更新大量冗余信息，极易出错。

2.2 支撑多变的计费模式
同一个模型 SPU，可以通过不同的 SKU 实现完全不同的计费逻辑：

按量付费 SKU（无有效期，余额消耗）

包月订阅 SKU（有效期 30 天，固定额度）

拼团卡 SKU（需要关联拼团活动 ID）

这种多态性只能通过独立的 SKU 表来承载。

2.3 电商运营灵活性
运营需要单独控制某个 SKU 的库存（如限量 1000 份）、限购（每人限购 1 件）、促销价格，而不影响其他 SKU。

数据分析时，需要按 SPU 维度统计模型受欢迎程度，按 SKU 维度分析套餐销售情况，分层级管理。

3. 数据模型设计建议（核心表结构）
以下是一个简洁但可扩展的设计，可直接指导后端开发。

3.1 产品表（spu）
存储模型基础信息，不包含任何价格或库存字段。

字段名	类型	说明
spu_id	string	主键，如 spu_deepseek_v3
name	string	产品名称，如 “DeepSeek-V3.2 模型服务”
model_code	string	上游模型标识，用于路由调用，如 deepseek-v3.2
provider	string	厂商，如 deepseek、baidu
description	string	产品介绍，技术文档链接
icon_url	string	图标
billing_coefficient	decimal	平台内部成本系数（1 Token 消耗多少算力点）
status	int	上下架（0下架 1上架）
create_time	datetime	创建时间
3.2 销售属性表（sale_attr）
定义该 SPU 下允许的销售属性维度（可选，用于动态组合 SKU）。例如：

billing_mode: [按量, 包月, 包年, 无限量, 拼团]

quota_limit: [无限制, 100万, 500万, 1000万]

period: [一次性, 1个月, 3个月, 12个月]

注：如果 SKU 类型固定且较少，可以跳过此表，直接在 SKU 表用枚举字段表示。

3.3 SKU 表（sku）
存储具体的售卖规格。关键字段如下：

字段名	类型	说明
sku_id	string	主键，如 sku_dsv3_1m_onetime
spu_id	string	外键，关联 spu
name	string	销售名称，如 “DeepSeek-V3.2 100万Tokens包”
billing_type	enum	prepaid_token（Token包）、subscription（订阅）、payg（按量）、reserved_concurrency（预留并发）
period_type	enum	one_time、monthly、quarterly、yearly
period_value	int	与 period_type 配合，如 1 表示 1 个月
token_amount	bigint	若为 Token 包模式，填入总 Token 数；若为无限量，填 -1 或 NULL
concurrency_limit	int	若为并发包，填入最大并发数
price	decimal	售价（单位：分）
original_price	decimal	原价（用于展示折扣）
quota_cycle	enum	配额重置周期：per_period（每个周期重置）、lifetime（总生命周期）
max_purchase_per_user	int	限购数量
sort_order	int	排序权重
status	int	上下架
start_time	datetime	售卖开始时间
end_time	datetime	售卖结束时间
3.4 用户购买记录表（user_asset）
记录用户购买的 SKU 实例，用于计量和权益判断。

字段名	类型	说明
asset_id	string	主键
user_id	string	用户 ID
sku_id	string	关联 SKU
order_id	string	关联订单
remaining_balance	bigint	Token 包剩余额度（若为按量或无限量，可无此字段）
expire_at	datetime	权益过期时间
status	enum	active、exhausted、expired、refunded
created_at	datetime	购买时间
3.5 计费引擎配置表（可选）
用于灵活配置不同 SKU 的计费规则（如超出套餐后如何扣费、是否自动续费等）。

4. 独立设计的好处与落地要点
解耦：当上游厂商（如阿里云）调整模型计费系数时，只需修改 spu.billing_coefficient，所有关联 SKU 自动生效，无需逐一修改。

扩展性：未来若要新增“按效付费”SKU，只需在 sku.billing_type 增加枚举值，并在计费引擎中实现对应逻辑，不影响现有 SKU。

运营易用性：运营人员可以在后台独立管理 SPU（上架新模型）和 SKU（配置不同套餐），互不干扰。

总结：SKU 是产品（SPU）在电商销售场景下的具体实现。数据模型必须分开设计，SPU 表描述“卖什么”，SKU 表描述“怎么卖”。二者通过外键关联，共同构成完整的商品体系。

在开发时，请务必让前端在商品详情页展示 SPU 信息（模型介绍），并通过下拉/选项卡选择不同 SKU（套餐规格）进行购买。这样既符合用户认知，也便于后续扩展。




SPU/SKU 与大模型厂商对接：具体实现方案
一、整体架构：统一接入层设计
核心设计思路是借鉴OpenRouter、Grab AI Gateway等行业标杆实践，构建一个“统一计量与映射层”。


flowchart TB
    subgraph C端用户
        A[电商前台]
    end
    
    subgraph 平台层
        B[统一API网关]
        C[SPU/SKU引擎]
        D[计量与计费引擎]
        E[适配器层]
    end
    
    subgraph B端厂商
        F[阿里云/通义千问]
        G[百度千帆/文心]
        H[火山引擎/豆包]
        I[其他厂商API]
    end
    
    A --> B
    B --> C
    C --> D
    D --> E
    E --> F
    E --> G
    E --> H
    E --> I
    
    F -.->|回调/账单| D
    G -.->|回调/账单| D
    H -.->|回调/账单| D
    I -.->|回调/账单| D

关键设计原则：平台充当“智能代理”，B端厂商无需任何改造，保持其现有API规范和计费模式不变。

二、SPU/SKU 与厂商映射的具体实现
2.1 映射关系表
平台层	映射对象	映射字段/逻辑	示例
SPU	厂商模型ID	model_code 直接存储厂商模型标识	ERNIE-4.0-8K
SKU	厂商计费模式	billing_strategy 存储计费适配策略	TOKEN_BASED / HYBRID
价格	厂商Token单价	存储基础成本价，平台加成后形成零售价	厂商价: ¥0.008/1K → 平台价: ¥0.012/1K
配额	厂商TPM限制	聚合多个厂商/密钥实现更高并发	单一厂商500 QPM → 平台聚合2000 QPM
2.2 数据模型扩展字段
在之前设计的基础上，为SPU表增加厂商对接字段：

sql
-- SPU表新增厂商对接字段
ALTER TABLE spu ADD COLUMN provider_model_id VARCHAR(128) COMMENT '厂商侧模型标识，如 ernie-4.0-8k';
ALTER TABLE spu ADD COLUMN provider_api_endpoint VARCHAR(512) COMMENT '厂商API地址';
ALTER TABLE spu ADD COLUMN provider_auth_type VARCHAR(32) DEFAULT 'API_KEY' COMMENT '认证方式: API_KEY/AK_SK/OAUTH';
ALTER TABLE spu ADD COLUMN provider_billing_type VARCHAR(32) COMMENT '厂商计费类型: INPUT_OUTPUT/MIXED/FLAT';
ALTER TABLE spu ADD COLUMN provider_input_rate DECIMAL(10,6) COMMENT '厂商输入Token单价(分/1K)';
ALTER TABLE spu ADD COLUMN provider_output_rate DECIMAL(10,6) COMMENT '厂商输出Token单价(分/1K)';
ALTER TABLE spu ADD COLUMN billing_coefficient DECIMAL(5,2) DEFAULT 1.0 COMMENT '平台成本系数，用于利润计算';
2.3 厂商计费模式适配策略
不同厂商的计费方式差异很大，需要在适配器中处理：

厂商	计费特征	平台适配策略	实现方式
阿里云/通义千问	输入输出分开计价	分别统计input/output tokens，按厂商单价扣费	从响应头提取input_tokens/output_tokens字段
百度千帆	模型版本不同计价	在SPU级别区分版本	不同版本对应不同SPU，各自关联SKU
火山引擎/豆包	分段计价（输入长度区间）	请求前计算输入长度，动态路由到对应区间	在适配器中实现长度判断，选择对应API端点
DeepSeek	统一单价，缓存优惠	平台侧做缓存命中判断，按优惠价结算	维护请求指纹，相同请求从缓存返回
华为云	用户自行配置鉴权	平台提供密钥托管，统一调用	加密存储用户API Key，调用时动态注入
三、API 调用与计量流程
3.1 完整调用链路


sequenceDiagram
    participant C as C端用户
    participant G as 平台API网关
    participant S as SPU/SKU引擎
    participant M as 计量引擎
    participant A as 适配器
    participant P as B端厂商API
    
    C->>G: 1. API请求 + API Key
    G->>S: 2. 验证用户权益<br/>查询购买的SKU
    S-->>G: 3. 返回用户配额/余额
    
    alt 余额不足
        G-->>C: 返回余额不足错误
    else 余额充足
        G->>A: 4. 转发请求+用户SKU信息
        A->>A: 5. 协议转换<br/>(统一格式→厂商格式)
        A->>P: 6. 调用厂商API
        P-->>A: 7. 返回响应+token用量
        A->>M: 8. 上报用量(异步)
        M->>M: 9. 扣减用户余额<br/>记录厂商成本
        A-->>G: 10. 转换响应<br/>(厂商格式→统一格式)
        G-->>C: 11. 返回结果
    end

3.2 用量上报与计费解耦
参考Stripe AI SDK的meter设计，用量上报采用异步非阻塞模式：

python
# 异步计量中间件示例
class MeterMiddleware:
    async def after_request(self, request, response, provider_response):
        # 从厂商响应中提取token用量
        usage = extract_token_usage(provider_response)
        
        # 异步上报，不阻塞响应返回
        asyncio.create_task(
            self.meter_report(
                user_id=request.user_id,
                sku_id=request.sku_id,
                input_tokens=usage.input_tokens,
                output_tokens=usage.output_tokens,
                provider_cost=usage.cost  # 厂商侧成本
            )
        )
        
        return response
3.3 缓存策略优化成本
借鉴OpenRouter的缓存方案，对重复请求进行缓存，降低厂商调用成本：

缓存命中：直接返回缓存结果，不调用厂商API，成本为0

缓存未命中：调用厂商API，成本计入用户消费

利润空间：缓存命中部分的利润可返还给用户作为折扣

四、厂商接入的具体流程
4.1 新厂商接入标准流程
参考华为云和百度智能云的接入模式：

阶段	任务	产出物	负责人
1. 调研	分析厂商API规范、计费模式、认证方式	厂商对接文档	产品/技术
2. 配置SPU	在平台录入模型信息、厂商参数	SPU数据记录	运营
3. 配置SKU	设计套餐规格、定价策略	SKU数据记录	运营/产品
4. 适配器开发	实现协议转换、认证适配	适配器代码	开发
5. 测试验证	端到端调用测试、计费验证	测试报告	QA
6. 上线发布	开放SKU售卖、监控告警	上线记录	运维
4.2 适配器开发规范
每个厂商需要实现统一的适配器接口：

python
class ProviderAdapter(ABC):
    @abstractmethod
    def transform_request(self, unified_request, user_sku):
        """将平台统一请求转换为厂商格式"""
        pass
    
    @abstractmethod
    def transform_response(self, provider_response):
        """将厂商响应转换为平台统一格式"""
        pass
    
    @abstractmethod
    def extract_usage(self, provider_response):
        """从厂商响应中提取token用量"""
        pass
    
    @abstractmethod
    def calculate_cost(self, usage, spu_config):
        """根据厂商计费规则计算成本"""
        pass
    
    @abstractmethod
    def get_auth_headers(self, user_credentials):
        """生成厂商认证头"""
        pass
五、关键难点与解决方案
5.1 成本系数的动态计算
问题：不同厂商、不同模型的Token单价差异巨大，平台需要精确计算成本和利润。

解决方案：

SPU表中存储厂商的input/output单价

用户购买SKU时，锁定当时的成本系数（对冲上游涨价）

每日对账时，按实际用量乘以厂商单价计算平台应付款

5.2 限流与配额管理
问题：多个C端用户共享B端厂商的API配额，单个用户可能耗尽所有配额。

解决方案（参考Grab AI Gateway经验）：

多层限流：用户级、SKU级、厂商级三层限流

公平队列：按用户权重分配厂商配额，避免单个用户霸占

动态路由：当某厂商限流时，自动切换到备用厂商（如果模型能力相当）

5.3 厂商账单对账
问题：平台需与多家厂商分别对账，确保账单准确。

解决方案：

平台侧记录每笔调用的厂商成本

每日生成厂商维度的对账单

与厂商官方账单定期比对，差异超过阈值告警

技术债务预留：

厂商API版本升级的兼容性设计

厂商计费模式变更的热更新机制

多厂商密钥的加密存储与轮换

通过以上设计，平台实现了：

B端厂商零改造：保持现有API和计费模式不变

C端用户统一体验：一个API Key访问所有模型，统一账单

未来扩展性：新增厂商只需开发适配器，SPU/SKU配置即可上线

核心数据模型关系：

SPU = 厂商模型（技术实体）

SKU = 售卖套餐（商业实体）

适配器 = 连接桥梁（协议转换）

计量引擎 = 成本核算（利润管理）