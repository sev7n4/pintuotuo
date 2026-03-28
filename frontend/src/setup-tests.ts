import '@testing-library/jest-dom';

// Mock import.meta.env
const mockEnv = {
  VITE_API_BASE_URL: '/api/v1',
  MODE: 'test',
  DEV: false,
  PROD: false,
  SSR: false,
};

(globalThis as any).importMeta = {
  env: mockEnv,
  glob: jest.fn(),
};

Object.defineProperty(globalThis, 'importMeta', {
  value: {
    env: mockEnv,
    glob: jest.fn(),
  },
  writable: true,
});

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
});

afterAll(() => {
  console.error = originalError;
});
