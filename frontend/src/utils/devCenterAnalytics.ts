const PREFIX = 'dev_center_';

/**
 * 开发者中心埋点（可接埋点 SDK；开发环境打 console）。
 * 约定事件名：dev_center_quickstart_complete 等。
 */
export function trackDevCenter(event: string, detail?: Record<string, unknown>): void {
  const name = event.startsWith(PREFIX) ? event : `${PREFIX}${event}`;
  if (import.meta.env.DEV) {
    // eslint-disable-next-line no-console
    console.debug('[dev-center]', name, detail ?? {});
  }
  try {
    window.dispatchEvent(
      new CustomEvent('pintuotuo_analytics', { detail: { name, ...(detail ?? {}) } })
    );
  } catch {
    /* ignore */
  }
}
