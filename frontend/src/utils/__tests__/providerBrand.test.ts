import { getProviderLogoUrl, resolveProviderIconSlug } from '../providerBrand';

describe('providerBrand', () => {
  it('resolves known provider to Lobe Icons slug', () => {
    expect(resolveProviderIconSlug('openai')).toBe('openai');
    expect(resolveProviderIconSlug('Anthropic')).toBe('anthropic');
    expect(resolveProviderIconSlug('google-ai')).toBe('google');
    expect(resolveProviderIconSlug('mistral')).toBe('mistral');
    expect(resolveProviderIconSlug('zhipu')).toBe('zhipu');
    expect(resolveProviderIconSlug('stepfun')).toBe('stepfun');
    expect(resolveProviderIconSlug('openrouter')).toBe('openrouter');
    expect(resolveProviderIconSlug('siliconflow')).toBe('siliconcloud');
    expect(resolveProviderIconSlug('bytedance')).toBe('bytedance');
    expect(resolveProviderIconSlug('volcengine')).toBe('volcengine');
    expect(resolveProviderIconSlug('doubao')).toBe('doubao');
  });

  it('returns logo URL for mapped providers (Lobe Icons CDN)', () => {
    expect(getProviderLogoUrl('openai')).toMatch(
      /@lobehub\/icons-static-svg@\d+\.\d+\.\d+\/icons\/openai\.svg$/
    );
    expect(getProviderLogoUrl('deepseek')).toMatch(/\/icons\/deepseek\.svg$/);
    expect(getProviderLogoUrl('zhipu')).toMatch(/\/icons\/zhipu\.svg$/);
    expect(getProviderLogoUrl('openrouter')).toMatch(/\/icons\/openrouter\.svg$/);
    expect(getProviderLogoUrl('siliconflow')).toMatch(/\/icons\/siliconcloud\.svg$/);
    expect(getProviderLogoUrl('bytedance')).toMatch(/\/icons\/bytedance\.svg$/);
    expect(getProviderLogoUrl('volcengine')).toMatch(/\/icons\/volcengine\.svg$/);
  });

  it('returns null for unknown provider', () => {
    expect(getProviderLogoUrl('unknown_vendor_xyz')).toBeNull();
  });
});
