package services

import (
	"context"
	"testing"
	"time"

	"github.com/pintuotuo/backend/billing"
)

func TestSubscriptionEndOnOrBeforeToday(t *testing.T) {
	today := time.Date(2026, 4, 3, 15, 0, 0, 0, time.UTC)
	past := time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC)
	same := time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC)
	future := time.Date(2026, 4, 4, 0, 0, 0, 0, time.UTC)

	restore := subscriptionRenewalNowForTest
	defer func() { subscriptionRenewalNowForTest = restore }()
	subscriptionRenewalNowForTest = func() time.Time { return today }

	if !subscriptionEndOnOrBeforeToday(past) {
		t.Error("past end should be on or before today")
	}
	if !subscriptionEndOnOrBeforeToday(same) {
		t.Error("same calendar day should match")
	}
	if subscriptionEndOnOrBeforeToday(future) {
		t.Error("future end should not match")
	}
}

func TestProcessDueTokenAutoRenewals_nilDeps(t *testing.T) {
	ctx := context.Background()
	_, _, _, err := ProcessDueTokenAutoRenewals(ctx, nil, billing.GetBillingEngine())
	if err == nil {
		t.Fatal("expected error for nil db")
	}
	_, _, _, err = ProcessDueTokenAutoRenewals(ctx, nil, nil)
	if err == nil {
		t.Fatal("expected error for nil engine")
	}
}
