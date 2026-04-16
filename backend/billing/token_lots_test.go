package billing

import "testing"

func TestDefaultTokenLotValidDays(t *testing.T) {
	if DefaultTokenLotValidDays != 365 {
		t.Fatalf("DefaultTokenLotValidDays = %d, want 365", DefaultTokenLotValidDays)
	}
}

func TestRound2(t *testing.T) {
	tests := []struct {
		in   float64
		want float64
	}{
		{1.234, 1.23},
		{10.005, 10.01},
		{0, 0},
	}
	for _, tt := range tests {
		if g := round2(tt.in); g != tt.want {
			t.Errorf("round2(%v) = %v, want %v", tt.in, g, tt.want)
		}
	}
}
