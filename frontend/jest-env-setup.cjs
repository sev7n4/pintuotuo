/**
 * 在测试文件加载前执行；配合 babel-plugin-transform-vite-meta-env
 * 将 import.meta.env.VITE_* 转为 process.env.VITE_*。
 */
process.env.VITE_API_BASE_URL = process.env.VITE_API_BASE_URL || '/api/v1';
process.env.VITE_ALLOW_MOCK_RECHARGE = process.env.VITE_ALLOW_MOCK_RECHARGE || 'false';
