-- 统一配置化路由系统 - 数据模型扩展
-- Version: 062
-- 目标：
-- 1) 扩展 model_providers 表，支持多端点配置和路由策略
-- 2) 扩展 merchant_api_keys 表，支持商户区域和路由偏好
-- 3) 扩展 merchants 表，支持商户类型和区域

-- ============================================================================
-- 1. 扩展 model_providers 表
-- ============================================================================

ALTER TABLE model_providers ADD COLUMN IF NOT EXISTS provider_region VARCHAR(20) DEFAULT 'domestic';
ALTER TABLE model_providers ADD COLUMN IF NOT EXISTS route_strategy JSONB DEFAULT '{}'::jsonb;
ALTER TABLE model_providers ADD COLUMN IF NOT EXISTS endpoints JSONB DEFAULT '{}'::jsonb;

COMMENT ON COLUMN model_providers.provider_region IS '厂商区域：domestic(国内), overseas(海外)';
COMMENT ON COLUMN model_providers.route_strategy IS '路由策略配置（JSONB）：包含不同用户类型的路由策略';
COMMENT ON COLUMN model_providers.endpoints IS '多端点配置（JSONB）：包含直连、LiteLLM、代理等多种端点';

-- ============================================================================
-- 2. 扩展 merchant_api_keys 表
-- ============================================================================

ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS merchant_region VARCHAR(20) DEFAULT 'domestic';
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS route_preference JSONB DEFAULT '{}'::jsonb;

COMMENT ON COLUMN merchant_api_keys.merchant_region IS '商户区域：domestic(国内), overseas(海外)';
COMMENT ON COLUMN merchant_api_keys.route_preference IS '路由偏好配置（JSONB）：包含优先模式、端点、降级设置';

-- ============================================================================
-- 3. 扩展 merchants 表
-- ============================================================================

ALTER TABLE merchants ADD COLUMN IF NOT EXISTS merchant_type VARCHAR(20) DEFAULT 'standard';
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS region VARCHAR(20) DEFAULT 'domestic';

COMMENT ON COLUMN merchants.merchant_type IS '商户类型：standard(普通), enterprise(企业), premium(高级)';
COMMENT ON COLUMN merchants.region IS '商户所在区域：domestic(国内), overseas(海外)';

-- ============================================================================
-- 4. 创建索引
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_model_providers_provider_region ON model_providers(provider_region);
CREATE INDEX IF NOT EXISTS idx_model_providers_route_strategy ON model_providers USING GIN(route_strategy);
CREATE INDEX IF NOT EXISTS idx_model_providers_endpoints ON model_providers USING GIN(endpoints);

CREATE INDEX IF NOT EXISTS idx_merchant_api_keys_merchant_region ON merchant_api_keys(merchant_region);
CREATE INDEX IF NOT EXISTS idx_merchant_api_keys_route_preference ON merchant_api_keys USING GIN(route_preference);

CREATE INDEX IF NOT EXISTS idx_merchants_merchant_type ON merchants(merchant_type);
CREATE INDEX IF NOT EXISTS idx_merchants_region ON merchants(region);

-- ============================================================================
-- 5. 初始化数据 - 海外厂商配置
-- ============================================================================

-- OpenAI
UPDATE model_providers 
SET provider_region = 'overseas',
    route_strategy = '{
      "default_mode": "auto",
      "domestic_users": {
        "mode": "litellm",
        "fallback_mode": "proxy",
        "proxy_endpoint": "gaap"
      },
      "overseas_users": {
        "mode": "direct"
      },
      "enterprise_users": {
        "mode": "litellm",
        "fallback_mode": "proxy"
      }
    }'::jsonb,
    endpoints = '{
      "direct": {
        "domestic": "",
        "overseas": "https://api.openai.com/v1"
      },
      "litellm": {
        "domestic": "http://litellm-overseas:4000/v1",
        "overseas": "http://litellm-overseas:4000/v1"
      },
      "proxy": {
        "gaap": "https://openai-gaap.example.com",
        "nginx_hk": "https://openai-proxy-hk.example.com"
      }
    }'::jsonb,
    updated_at = CURRENT_TIMESTAMP
WHERE code = 'openai';

-- Anthropic
UPDATE model_providers 
SET provider_region = 'overseas',
    route_strategy = '{
      "default_mode": "auto",
      "domestic_users": {
        "mode": "litellm",
        "fallback_mode": "proxy",
        "proxy_endpoint": "gaap"
      },
      "overseas_users": {
        "mode": "direct"
      },
      "enterprise_users": {
        "mode": "litellm",
        "fallback_mode": "proxy"
      }
    }'::jsonb,
    endpoints = '{
      "direct": {
        "domestic": "",
        "overseas": "https://api.anthropic.com/v1"
      },
      "litellm": {
        "domestic": "http://litellm-overseas:4000/v1",
        "overseas": "http://litellm-overseas:4000/v1"
      },
      "proxy": {
        "gaap": "https://anthropic-gaap.example.com",
        "nginx_hk": "https://anthropic-proxy-hk.example.com"
      }
    }'::jsonb,
    updated_at = CURRENT_TIMESTAMP
WHERE code = 'anthropic';

-- Google
UPDATE model_providers 
SET provider_region = 'overseas',
    route_strategy = '{
      "default_mode": "auto",
      "domestic_users": {
        "mode": "litellm",
        "fallback_mode": "proxy",
        "proxy_endpoint": "gaap"
      },
      "overseas_users": {
        "mode": "direct"
      },
      "enterprise_users": {
        "mode": "litellm",
        "fallback_mode": "proxy"
      }
    }'::jsonb,
    endpoints = '{
      "direct": {
        "domestic": "",
        "overseas": "https://generativelanguage.googleapis.com/v1"
      },
      "litellm": {
        "domestic": "http://litellm-overseas:4000/v1",
        "overseas": "http://litellm-overseas:4000/v1"
      },
      "proxy": {
        "gaap": "https://google-gaap.example.com",
        "nginx_hk": "https://google-proxy-hk.example.com"
      }
    }'::jsonb,
    updated_at = CURRENT_TIMESTAMP
WHERE code = 'google';

-- ============================================================================
-- 6. 初始化数据 - 国内厂商配置
-- ============================================================================

-- DeepSeek
UPDATE model_providers 
SET provider_region = 'domestic',
    route_strategy = '{
      "default_mode": "auto",
      "domestic_users": {
        "mode": "direct"
      },
      "overseas_users": {
        "mode": "litellm"
      },
      "enterprise_users": {
        "mode": "litellm",
        "fallback_mode": "direct"
      }
    }'::jsonb,
    endpoints = '{
      "direct": {
        "domestic": "https://api.deepseek.com/v1",
        "overseas": "https://api.deepseek.com/v1"
      },
      "litellm": {
        "domestic": "http://litellm-domestic:4000/v1",
        "overseas": "http://litellm-domestic:4000/v1"
      }
    }'::jsonb,
    updated_at = CURRENT_TIMESTAMP
WHERE code = 'deepseek';

-- 智谱
UPDATE model_providers 
SET provider_region = 'domestic',
    route_strategy = '{
      "default_mode": "auto",
      "domestic_users": {
        "mode": "direct"
      },
      "overseas_users": {
        "mode": "litellm"
      },
      "enterprise_users": {
        "mode": "litellm",
        "fallback_mode": "direct"
      }
    }'::jsonb,
    endpoints = '{
      "direct": {
        "domestic": "https://open.bigmodel.cn/api/paas/v4",
        "overseas": "https://open.bigmodel.cn/api/paas/v4"
      },
      "litellm": {
        "domestic": "http://litellm-domestic:4000/v1",
        "overseas": "http://litellm-domestic:4000/v1"
      }
    }'::jsonb,
    updated_at = CURRENT_TIMESTAMP
WHERE code = 'zhipu';

-- 阿里云
UPDATE model_providers 
SET provider_region = 'domestic',
    route_strategy = '{
      "default_mode": "auto",
      "domestic_users": {
        "mode": "direct"
      },
      "overseas_users": {
        "mode": "litellm"
      },
      "enterprise_users": {
        "mode": "litellm",
        "fallback_mode": "direct"
      }
    }'::jsonb,
    endpoints = '{
      "direct": {
        "domestic": "https://dashscope.aliyuncs.com/api/v1",
        "overseas": "https://dashscope.aliyuncs.com/api/v1"
      },
      "litellm": {
        "domestic": "http://litellm-domestic:4000/v1",
        "overseas": "http://litellm-domestic:4000/v1"
      }
    }'::jsonb,
    updated_at = CURRENT_TIMESTAMP
WHERE code = 'alibaba';

-- Moonshot
UPDATE model_providers 
SET provider_region = 'domestic',
    route_strategy = '{
      "default_mode": "auto",
      "domestic_users": {
        "mode": "direct"
      },
      "overseas_users": {
        "mode": "litellm"
      },
      "enterprise_users": {
        "mode": "litellm",
        "fallback_mode": "direct"
      }
    }'::jsonb,
    endpoints = '{
      "direct": {
        "domestic": "https://api.moonshot.cn/v1",
        "overseas": "https://api.moonshot.cn/v1"
      },
      "litellm": {
        "domestic": "http://litellm-domestic:4000/v1",
        "overseas": "http://litellm-domestic:4000/v1"
      }
    }'::jsonb,
    updated_at = CURRENT_TIMESTAMP
WHERE code = 'moonshot';

-- MiniMax
UPDATE model_providers 
SET provider_region = 'domestic',
    route_strategy = '{
      "default_mode": "auto",
      "domestic_users": {
        "mode": "direct"
      },
      "overseas_users": {
        "mode": "litellm"
      },
      "enterprise_users": {
        "mode": "litellm",
        "fallback_mode": "direct"
      }
    }'::jsonb,
    endpoints = '{
      "direct": {
        "domestic": "https://api.minimax.chat/v1",
        "overseas": "https://api.minimax.chat/v1"
      },
      "litellm": {
        "domestic": "http://litellm-domestic:4000/v1",
        "overseas": "http://litellm-domestic:4000/v1"
      }
    }'::jsonb,
    updated_at = CURRENT_TIMESTAMP
WHERE code = 'minimax';

-- StepFun
UPDATE model_providers 
SET provider_region = 'domestic',
    route_strategy = '{
      "default_mode": "auto",
      "domestic_users": {
        "mode": "direct"
      },
      "overseas_users": {
        "mode": "litellm"
      },
      "enterprise_users": {
        "mode": "litellm",
        "fallback_mode": "direct"
      }
    }'::jsonb,
    endpoints = '{
      "direct": {
        "domestic": "https://api.stepfun.com/v1",
        "overseas": "https://api.stepfun.com/v1"
      },
      "litellm": {
        "domestic": "http://litellm-domestic:4000/v1",
        "overseas": "http://litellm-domestic:4000/v1"
      }
    }'::jsonb,
    updated_at = CURRENT_TIMESTAMP
WHERE code = 'stepfun';

-- ============================================================================
-- 7. 初始化商户数据
-- ============================================================================

-- 将所有现有商户设置为普通类型和国内区域
UPDATE merchants 
SET merchant_type = 'standard',
    region = 'domestic'
WHERE merchant_type IS NULL OR region IS NULL;
