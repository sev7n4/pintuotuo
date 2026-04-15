package handlers

import (
	"os"
	"strings"
)

func merchantRegisterMode() string {
	m := strings.ToLower(strings.TrimSpace(os.Getenv("MERCHANT_REGISTER_MODE")))
	switch m {
	case "open", "invite_only", "hidden":
		return m
	default:
		return "invite_only"
	}
}

func merchantRegisterRequiresInvite(mode string) bool {
	return mode == "invite_only" || mode == "hidden"
}

func adminMFARequired() bool {
	return strings.TrimSpace(os.Getenv("ADMIN_MFA_REQUIRED")) == "true" || strings.TrimSpace(os.Getenv("ADMIN_MFA_REQUIRED")) == "1"
}
