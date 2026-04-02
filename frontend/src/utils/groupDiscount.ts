/**
 * Aligns with backend utils.NormalizeGroupDiscountRate:
 * - raw > 1 → percentage points (20 → 0.2)
 * - raw in (0, 1] → fraction (0.2 → 0.2)
 */
export function normalizeGroupDiscountRate(raw: number | null | undefined): number {
  if (raw == null || raw <= 0) {
    return 0;
  }
  let r = raw;
  if (r > 1) {
    r = r / 100;
  }
  if (r > 1) {
    return 1;
  }
  return r;
}
