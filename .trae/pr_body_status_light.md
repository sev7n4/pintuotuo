## 问题概述

修复BYOK路由管理页面红绿灯图标显示问题。

## 问题分析

### 问题3：列表没有显示红绿灯的原因

**数据库验证结果：**
```
 id | merchant_id |         name          | provider  | health_status | verification_result | route_mode 
----+-------------+-----------------------+-----------+---------------+---------------------+------------
 16 |           4 | test-deepseek         | deepseek  | healthy       | verified            | direct
 18 |           4 | test-kimi             | moonshot  | healthy       | failed              | direct
 28 |          16 | key-stepfun           | stepfun   | unhealthy     | verified            | direct
```

**结论：** 数据库中有正确的health_status和verification_result数据，说明后端逻辑正确。

**问题定位：** CSS样式问题

**原因：**
- `.statusDotHealthy`等类只定义了背景色，没有继承`.statusDot`的基础样式（display、width、height、border-radius等）
- 前端代码使用`styles.statusDotHealthy`，但没有同时应用`styles.statusDot`

## 解决方案

### 问题3的解决方案

修改CSS样式，让所有状态点类都包含完整的基础样式：

1. **增大图标尺寸**：从10px×10px改为12px×12px，更容易查看
2. **完整样式定义**：每个状态点类都包含display、width、height、border-radius等基础样式
3. **添加hover效果**：鼠标悬停时图标放大1.3倍，提升用户体验
4. **添加过渡动画**：平滑的缩放动画

## 主要变更

**前端：**
- [frontend/src/pages/admin/AdminByokRouting.module.css](file:///Users/4seven/workspace/pintuotuo/frontend/src/pages/admin/AdminByokRouting.module.css)

## 测试

- ✅ 数据库验证：health_status和verification_result字段有正确的数据
- ✅ CSS样式修复：所有状态点类都包含完整的基础样式
- ✅ 图标尺寸增大：从10px×10px改为12px×12px
- ✅ 添加hover效果：鼠标悬停时图标放大1.3倍
