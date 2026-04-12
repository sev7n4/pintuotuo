package services

import "testing"

func TestProcurementCostCNY(t *testing.T) {
	got := ProcurementCostCNY(0.01, 0.03, 1000, 500)
	want := CostFromPer1KRates(0.01, 0.03, 1000, 500)
	if got != want {
		t.Fatalf("got %v want %v", got, want)
	}
}
