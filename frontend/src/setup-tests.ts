import '@testing-library/jest-dom';

// import.meta.env 由 babel-plugin-transform-vite-meta-env + jest-env-setup.cjs 处理
jest.setTimeout(20000);

/** AuthPage 等组件在 Jest 中会请求 capabilities；统一 mock，避免 jsdom 无 fetch 或网络报错 */
const defaultCapabilities = {
  sms: false,
  wechat_oauth: false,
  github_oauth: false,
  account_linking: false,
};

globalThis.fetch = jest.fn((input: RequestInfo | URL) => {
  const url = typeof input === 'string' ? input : String(input);
  if (url.includes('/users/auth/capabilities')) {
    return Promise.resolve({
      ok: true,
      json: () => Promise.resolve(defaultCapabilities),
    } as Response);
  }
  return Promise.reject(new Error(`Unmocked fetch: ${url}`));
}) as unknown as typeof fetch;

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation((query) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(),
    removeListener: jest.fn(),
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
});

// Suppress console errors in tests (optional)
const originalError = console.error;
const originalWarn = console.warn;
beforeAll(() => {
  console.error = (...args: any[]) => {
    if (
      typeof args[0] === 'string' &&
      args[0].includes('Not implemented: HTMLFormElement.prototype.submit')
    ) {
      return;
    }
    originalError.call(console, ...args);
  };
  console.warn = (...args: any[]) => {
    const first = typeof args[0] === 'string' ? args[0] : '';
    if (
      first.includes('React Router Future Flag Warning') &&
      (first.includes('v7_startTransition') || first.includes('v7_relativeSplatPath'))
    ) {
      return;
    }
    originalWarn.call(console, ...args);
  };
});

afterAll(() => {
  console.error = originalError;
  console.warn = originalWarn;
});
