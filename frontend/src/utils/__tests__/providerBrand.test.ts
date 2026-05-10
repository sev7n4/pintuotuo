import { getProviderLogoUrl, resolveProviderIconSlug } from '../providerBrand';

describe('providerBrand', () => {
  it('resolves known provider to simple-icons slug', () => {
    expect(resolveProviderIconSlug('openai')).toBe('openai');
    expect(resolveProviderIconSlug('Anthropic')).toBe('anthropic');
    expect(resolveProviderIconSlug('google-ai')).toBe('google');
  });

  it('returns logo URL for mapped providers', () => {
    expect(getProviderLogoUrl('openai')).toContain('/openai.svg');
    expect(getProviderLogoUrl('deepseek')).toContain('/deepseek.svg');
  });

  it('returns null for unknown provider', () => {
    expect(getProviderLogoUrl('unknown_vendor_xyz')).toBeNull();
  });
});
