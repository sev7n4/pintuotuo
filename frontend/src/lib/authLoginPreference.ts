/** localStorage key：上次成功登录使用的主入口（邮箱密码/魔法链接 vs 手机验证码） */
export const AUTH_PRIMARY_LOGIN_KEY = 'pintuotuo_auth_primary_login';

export type PrimaryLoginPreference = 'email' | 'phone';

export function readPrimaryLoginPreference(): PrimaryLoginPreference {
  if (typeof window === 'undefined') return 'email';
  try {
    const v = localStorage.getItem(AUTH_PRIMARY_LOGIN_KEY);
    if (v === 'phone' || v === 'email') return v;
  } catch {
    /* quota / private mode */
  }
  return 'email';
}

export function writePrimaryLoginPreference(v: PrimaryLoginPreference): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem(AUTH_PRIMARY_LOGIN_KEY, v);
  } catch {
    /* ignore */
  }
}
