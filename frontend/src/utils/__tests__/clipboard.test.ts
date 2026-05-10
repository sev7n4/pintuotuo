import { copyToClipboard } from '../clipboard';

describe('copyToClipboard', () => {
  it('returns false for empty string', async () => {
    expect(await copyToClipboard('')).toBe(false);
  });

  it('falls back to execCommand when clipboard API unavailable', async () => {
    const origClipboard = navigator.clipboard;
    Object.defineProperty(navigator, 'clipboard', { value: undefined, configurable: true });

    const execFn = jest.fn().mockReturnValue(true);
    Object.defineProperty(document, 'execCommand', { value: execFn, configurable: true });

    const ok = await copyToClipboard('hello-fallback');
    expect(ok).toBe(true);
    expect(execFn).toHaveBeenCalledWith('copy');

    Object.defineProperty(navigator, 'clipboard', { value: origClipboard, configurable: true });
  });
});
