/** Babel 仅用于 Jest；Vite 开发与构建仍走 vite.config。 */
module.exports = function (api) {
  api.cache.using(() => process.env.NODE_ENV);
  const isTest = api.env('test');
  if (!isTest) {
    return {};
  }
  return {
    presets: [
      ['@babel/preset-env', { targets: { node: 'current' } }],
      ['@babel/preset-typescript', { isTSX: true, allExtensions: true }],
      ['@babel/preset-react', { runtime: 'automatic' }],
    ],
    plugins: [
      'babel-plugin-transform-vite-meta-env',
      'babel-plugin-transform-import-meta',
    ],
  };
};
