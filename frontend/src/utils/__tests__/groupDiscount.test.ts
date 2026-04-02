import { normalizeGroupDiscountRate } from '../groupDiscount';

describe('normalizeGroupDiscountRate', () => {
  it('treats values above 1 as percentage points', () => {
    expect(normalizeGroupDiscountRate(20)).toBeCloseTo(0.2);
    expect(normalizeGroupDiscountRate(25)).toBeCloseTo(0.25);
  });

  it('keeps fractions in (0, 1]', () => {
    expect(normalizeGroupDiscountRate(0.2)).toBeCloseTo(0.2);
  });

  it('returns 0 for null/undefined/non-positive', () => {
    expect(normalizeGroupDiscountRate(null)).toBe(0);
    expect(normalizeGroupDiscountRate(undefined)).toBe(0);
    expect(normalizeGroupDiscountRate(0)).toBe(0);
  });

  it('caps at 100% discount', () => {
    expect(normalizeGroupDiscountRate(200)).toBe(1);
  });
});
