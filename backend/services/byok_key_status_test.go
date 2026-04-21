package services

import (
	"database/sql"
	"testing"
)

func TestKeyMeetsStrictAllowlist(t *testing.T) {
	if !KeyMeetsStrictAllowlist("healthy", "verified", sql.NullTime{}) {
		t.Fatal("expected healthy+verified ok without timestamp")
	}
	if KeyMeetsStrictAllowlist("unhealthy", "verified", sql.NullTime{}) {
		t.Fatal("unhealthy should fail")
	}
	if !KeyMeetsStrictAllowlist("degraded", "success", sql.NullTime{Valid: true}) {
		t.Fatal("degraded + success should pass")
	}
}

func TestAggregateMerchantBYOK(t *testing.T) {
	lv, has, need, n := AggregateMerchantBYOK(nil)
	if lv != MerchantBYOKLevelNone || has || need != 0 || n != 0 {
		t.Fatalf("empty: got %s %v %d %d", lv, has, need, n)
	}

	// inactive only -> gray
	lv, _, _, ac := AggregateMerchantBYOK([]KeyRowLite{{Status: "inactive", Health: "healthy", Verification: "verified"}})
	if lv != MerchantBYOKLevelGray || ac != 0 {
		t.Fatalf("inactive only: %s active=%d", lv, ac)
	}

	// one active strict-ok -> green
	lv, has, _, ac = AggregateMerchantBYOK([]KeyRowLite{{
		Status: "active", Health: "healthy", Verification: "verified",
	}})
	if lv != MerchantBYOKLevelGreen || !has || ac != 1 {
		t.Fatalf("green: %s has=%v n=%d", lv, has, ac)
	}

	// active not routable + failed -> yellow
	lv, has, _, _ = AggregateMerchantBYOK([]KeyRowLite{{
		Status: "active", Health: "unknown", Verification: "failed",
	}})
	if lv != MerchantBYOKLevelYellow || has {
		t.Fatalf("yellow: %s", lv)
	}

	// active not routable + unhealthy -> yellow
	lv, _, _, _ = AggregateMerchantBYOK([]KeyRowLite{{
		Status: "active", Health: "unhealthy", Verification: "pending",
	}})
	if lv != MerchantBYOKLevelYellow {
		t.Fatalf("want yellow got %s", lv)
	}

	// active unknown only -> gray
	lv, has, need, _ = AggregateMerchantBYOK([]KeyRowLite{{
		Status: "active", Health: "unknown", Verification: "pending",
	}})
	if lv != MerchantBYOKLevelGray || has || need != 1 {
		t.Fatalf("gray: %s need=%d", lv, need)
	}
}
