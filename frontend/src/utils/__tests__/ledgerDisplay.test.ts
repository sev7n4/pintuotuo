import { formatLedgerUnits, ledgerUnitColumnTitle } from '../ledgerDisplay';

describe('ledgerDisplay', () => {
  it('formats numbers without implying fiat currency', () => {
    const rounded = formatLedgerUnits(1.234567);
    expect(rounded).not.toMatch(/¥|\$|CNY|元/);
    expect(Number(String(rounded).replace(/[,\s]/g, ''))).toBe(1);
    expect(formatLedgerUnits(0)).toBe('0');
  });

  it('exposes a stable column title for usage tables', () => {
    expect(ledgerUnitColumnTitle).toContain('Token');
  });
});
