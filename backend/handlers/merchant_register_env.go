package handlers

import (
	"os"
	"strings"
)

const merchantModeInviteOnly = "invite_only"

func merchantRegisterMode() string {
	m := strings.ToLower(strings.TrimSpace(os.Getenv("MERCHANT_REGISTER_MODE")))
	switch m {
	case "open", merchantModeInviteOnly, "hidden":
		return m
	default:
		return merchantModeInviteOnly
	}
}

func merchantRegisterRequiresInvite(mode string) bool {
	return mode == merchantModeInviteOnly || mode == "hidden"
}

func adminMFARequired() bool {
	v := strings.TrimSpace(os.Getenv("ADMIN_MFA_REQUIRED"))
	return v == envTrue || v == "1"
}
