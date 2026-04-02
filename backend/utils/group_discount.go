package utils

import "database/sql"

// NormalizeGroupDiscountRate converts sku.group_discount_rate from DB/API to a fraction in [0, 1].
// Values greater than 1 are treated as percentage points (e.g. 20 means 20% off).
// Values in (0, 1] are treated as already-normalized fractions (e.g. 0.2 means 20% off).
func NormalizeGroupDiscountRate(raw float64) float64 {
	if raw <= 0 {
		return 0
	}
	r := raw
	if r > 1 {
		r = r / 100
	}
	if r > 1 {
		return 1
	}
	return r
}

// NormalizeGroupDiscountRateNull applies NormalizeGroupDiscountRate to a nullable SQL float.
func NormalizeGroupDiscountRateNull(n sql.NullFloat64) float64 {
	if !n.Valid {
		return 0
	}
	return NormalizeGroupDiscountRate(n.Float64)
}
