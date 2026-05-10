import { getOpenAICompatBaseURL } from '../openaiCompat';

describe('getOpenAICompatBaseURL', () => {
  it('uses window.origin for relative VITE_API_BASE_URL', () => {
    expect(getOpenAICompatBaseURL()).toMatch(/\/api\/v1\/openai\/v1$/);
  });
});
