import { formatLedgerUnits, ledgerUnitColumnTitle } from '../ledgerDisplay';

describe('ledgerDisplay', () => {
  it('formats numbers without implying fiat currency', () => {
    expect(formatLedgerUnits(1.234567)).toMatch(/1\.234567/);
    expect(formatLedgerUnits(0)).toBe('0');
  });

  it('exposes a stable column title for usage tables', () => {
    expect(ledgerUnitColumnTitle).toContain('Token');
  });
});
