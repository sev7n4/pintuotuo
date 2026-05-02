package services

const (
	VerificationStatusPending     = "pending"
	VerificationStatusVerified    = "verified"
	VerificationStatusSuspend     = "suspend"
	VerificationStatusUnreachable = "unreachable"
	VerificationStatusInvalid     = "invalid"
	VerificationStatusFailed      = "failed"
)

func MapErrorCategoryToVerificationStatus(errorCategory string, currentStatus string) string {
	switch errorCategory {
	case errorCategoryAuthInvalidKey, errorCategoryAuthPermissionDenied:
		return VerificationStatusInvalid

	case errorCategoryQuotaInsufficient:
		return VerificationStatusSuspend

	case errorCategoryNetworkTimeout, errorCategoryNetworkDNS, errorCategoryServiceUnavailable:
		return VerificationStatusUnreachable

	case errorCategoryRateLimited:
		return currentStatus

	case errorCategoryModelNotFound, errorCategoryContextTooLong, errorCategoryUpstreamBadRequest:
		return VerificationStatusFailed

	default:
		return VerificationStatusFailed
	}
}

func IsValidVerificationStatus(status string) bool {
	return status == VerificationStatusVerified
}
