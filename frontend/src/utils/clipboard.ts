/**
 * 复制到剪贴板：优先 Clipboard API；非 HTTPS 或不支持时回退到 textarea + execCommand。
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  if (text === undefined || text === null) return false;
  const s = String(text);
  if (s.length === 0) return false;

  try {
    if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(s);
      return true;
    }
  } catch {
    /* fallback below */
  }

  try {
    const ta = document.createElement('textarea');
    ta.value = s;
    ta.setAttribute('readonly', '');
    ta.style.position = 'fixed';
    ta.style.left = '-9999px';
    ta.style.top = '0';
    document.body.appendChild(ta);
    ta.focus();
    ta.select();
    ta.setSelectionRange(0, s.length);
    const ok = document.execCommand('copy');
    document.body.removeChild(ta);
    return ok;
  } catch {
    return false;
  }
}
