import { getAnthropicCompatBaseURL, getOpenAICompatBaseURL } from '../openaiCompat';

describe('getOpenAICompatBaseURL', () => {
  it('uses window.origin for relative VITE_API_BASE_URL', () => {
    expect(getOpenAICompatBaseURL()).toMatch(/\/api\/v1\/openai\/v1$/);
  });
});

describe('getAnthropicCompatBaseURL', () => {
  it('ends with anthropic v1 path', () => {
    expect(getAnthropicCompatBaseURL()).toMatch(/\/api\/v1\/anthropic\/v1$/);
  });
});
