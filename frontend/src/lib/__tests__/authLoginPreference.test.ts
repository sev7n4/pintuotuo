import {
  AUTH_PRIMARY_LOGIN_KEY,
  readPrimaryLoginPreference,
  writePrimaryLoginPreference,
} from '../authLoginPreference';

describe('authLoginPreference', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('defaults to email when unset', () => {
    expect(readPrimaryLoginPreference()).toBe('email');
  });

  it('reads phone when stored', () => {
    localStorage.setItem(AUTH_PRIMARY_LOGIN_KEY, 'phone');
    expect(readPrimaryLoginPreference()).toBe('phone');
  });

  it('ignores invalid values', () => {
    localStorage.setItem(AUTH_PRIMARY_LOGIN_KEY, 'weird');
    expect(readPrimaryLoginPreference()).toBe('email');
  });

  it('persists email and phone', () => {
    writePrimaryLoginPreference('phone');
    expect(localStorage.getItem(AUTH_PRIMARY_LOGIN_KEY)).toBe('phone');
    writePrimaryLoginPreference('email');
    expect(localStorage.getItem(AUTH_PRIMARY_LOGIN_KEY)).toBe('email');
  });
});
