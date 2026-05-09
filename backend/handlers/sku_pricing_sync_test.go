package handlers

import "testing"

func TestNormalizePricingVersionIDs(t *testing.T) {
	got := normalizePricingVersionIDs([]int{0, -1, 1, 2, 2, 3, 1})
	want := []int{1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("len mismatch: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("normalizePricingVersionIDs[%d]=%d, want %d", i, got[i], want[i])
		}
	}
}

func TestResolveSPUSyncPricingVersionIDs(t *testing.T) {
	t.Run("explicit version ids win", func(t *testing.T) {
		v := true
		got := resolveSPUSyncPricingVersionIDs(&v, []int{3, 0, 3, 2})
		want := []int{3, 2}
		if len(got) != len(want) {
			t.Fatalf("len mismatch: got %v want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("got %v want %v", got, want)
			}
		}
	})

	t.Run("baseline true defaults to id 1", func(t *testing.T) {
		v := true
		got := resolveSPUSyncPricingVersionIDs(&v, nil)
		if len(got) != 1 || got[0] != baselinePricingVersionID {
			t.Fatalf("got %v want [%d]", got, baselinePricingVersionID)
		}
	})

	t.Run("baseline false returns empty", func(t *testing.T) {
		v := false
		got := resolveSPUSyncPricingVersionIDs(&v, nil)
		if len(got) != 0 {
			t.Fatalf("got %v want empty", got)
		}
	})
}
