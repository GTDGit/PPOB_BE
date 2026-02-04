package domain

import (
	"fmt"
	"net/http"
)

// AppError represents an application error
type AppError struct {
	Code              string      `json:"code"`
	Message           string      `json:"message"`
	Details           interface{} `json:"details,omitempty"`
	RemainingAttempts *int        `json:"remainingAttempts,omitempty"`
	LockUntil         *string     `json:"lockUntil,omitempty"`
	RetryAfter        *int        `json:"retryAfter,omitempty"`
	HTTPStatus        int         `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Error code strings
const (
	// 400 Bad Request
	CodeInvalidRequest     = "INVALID_REQUEST"
	CodeInvalidToken       = "INVALID_TOKEN"
	CodeInvalidOTPMethod   = "INVALID_OTP_METHOD"
	CodeInvalidPhoneFormat = "INVALID_PHONE_FORMAT"
	CodePINMismatch        = "PIN_MISMATCH"
	CodeWeakPIN            = "WEAK_PIN"
	CodeInvalidDeviceID    = "INVALID_DEVICE_ID"

	// 401 Unauthorized
	CodeInvalidOTP          = "INVALID_OTP"
	CodeExpiredOTP          = "EXPIRED_OTP"
	CodeInvalidPIN          = "INVALID_PIN"
	CodeExpiredToken        = "EXPIRED_TOKEN"
	CodeTempTokenExpired    = "TEMP_TOKEN_EXPIRED"
	CodeInvalidRefreshToken = "INVALID_REFRESH_TOKEN"
	CodeUnauthorized        = "UNAUTHORIZED"

	// 403 Forbidden
	CodePINLocked                 = "PIN_LOCKED"
	CodeDeviceNotRecognized       = "DEVICE_NOT_RECOGNIZED"
	CodeCannotRemoveCurrentDevice = "CANNOT_REMOVE_CURRENT_DEVICE"

	// 404 Not Found
	CodeUserNotFound    = "USER_NOT_FOUND"
	CodeDeviceNotFound  = "DEVICE_NOT_FOUND"
	CodeSessionNotFound = "SESSION_NOT_FOUND"

	// 409 Conflict
	CodePhoneAlreadyRegistered = "PHONE_ALREADY_REGISTERED"

	// 423 Locked
	CodeAccountLocked = "ACCOUNT_LOCKED"

	// 429 Too Many Requests
	CodeOTPResendLimit  = "OTP_RESEND_LIMIT"
	CodeTooManyAttempts = "TOO_MANY_ATTEMPTS"
	CodeRateLimited     = "RATE_LIMITED"
	CodeOTPMaxAttempts  = "OTP_MAX_ATTEMPTS"
	CodeOTPRateLimited  = "OTP_RATE_LIMITED"

	// 500 Internal Server Error
	CodeInternalError = "INTERNAL_ERROR"
	CodeOTPSendFailed = "OTP_SEND_FAILED"

	// Transaction Errors - 400 Bad Request
	CodeInvalidTarget        = "INVALID_TARGET"
	CodeInvalidProduct       = "INVALID_PRODUCT"
	CodeInvalidServiceType   = "INVALID_SERVICE_TYPE"
	CodeInquiryExpired       = "INQUIRY_EXPIRED"
	CodeInquiryNotFound      = "INQUIRY_NOT_FOUND"
	CodeOrderExpired         = "ORDER_EXPIRED"
	CodeInvalidVoucher       = "INVALID_VOUCHER"
	CodeVoucherExpired       = "VOUCHER_EXPIRED"
	CodeVoucherUsed          = "VOUCHER_USED"
	CodeVoucherNotApplicable = "VOUCHER_NOT_APPLICABLE"
	CodeMinTransactionNotMet = "MIN_TRANSACTION_NOT_MET"
	CodeMaxVouchersExceeded  = "MAX_VOUCHERS_EXCEEDED"
	CodePINRequired          = "PIN_REQUIRED"

	// Transaction Errors - 402 Payment Required
	CodeInsufficientBalance = "INSUFFICIENT_BALANCE"

	// Transaction Errors - 409 Conflict
	CodeDuplicateTransaction = "DUPLICATE_TRANSACTION"

	// Transaction Errors - 422 Unprocessable Entity
	CodeProductUnavailable = "PRODUCT_UNAVAILABLE"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	CodeNoBill             = "NO_BILL"

	// KYC Errors - 400 Bad Request
	CodeKYCInvalidFile     = "KYC_INVALID_FILE"
	CodeKYCInvalidFileSize = "KYC_INVALID_FILE_SIZE"
	CodeKYCInvalidFileType = "KYC_INVALID_FILE_TYPE"
	CodeKYCOCRFailed       = "KYC_OCR_FAILED"
	CodeKYCFaceNoMatch     = "KYC_FACE_NO_MATCH"
	CodeKYCLivenessFailed  = "KYC_LIVENESS_FAILED"
	CodeKYCSessionNotFound = "KYC_SESSION_NOT_FOUND"
	CodeKYCSessionExpired  = "KYC_SESSION_EXPIRED"

	// KYC Errors - 403 Forbidden
	CodeKYCRequired        = "KYC_REQUIRED"
	CodeKYCAlreadyVerified = "KYC_ALREADY_VERIFIED"

	// KYC Errors - 409 Conflict
	CodeKYCNIKAlreadyUsed = "KYC_NIK_ALREADY_USED"

	// Deposit Errors - 400 Bad Request
	CodeInvalidAmount   = "INVALID_AMOUNT"
	CodeAmountTooLow    = "AMOUNT_TOO_LOW"
	CodeAmountTooHigh   = "AMOUNT_TOO_HIGH"
	CodeInvalidProvider = "INVALID_PROVIDER"
	CodeInvalidBank     = "INVALID_BANK"

	// Deposit Errors - 404 Not Found
	CodeDepositNotFound = "DEPOSIT_NOT_FOUND"

	// Deposit Errors - 409 Conflict
	CodeDepositAlreadyPaid = "DEPOSIT_ALREADY_PAID"

	// Deposit Errors - 410 Gone
	CodeDepositExpired = "DEPOSIT_EXPIRED"

	// Deposit Errors - 422 Unprocessable Entity
	CodeMethodUnavailable = "METHOD_UNAVAILABLE"

	// Deposit Errors - 429 Too Many Requests
	CodeTooManyPending = "TOO_MANY_PENDING"
)

// NewError creates a new AppError
func NewError(code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Common AppError variables for use in service layer
var (
	ErrInvalidRequestError = &AppError{
		Code:       CodeInvalidRequest,
		Message:    "Request tidak valid",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidTokenError = &AppError{
		Code:       CodeInvalidToken,
		Message:    "Token tidak valid",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidOTPMethodError = &AppError{
		Code:       CodeInvalidOTPMethod,
		Message:    "Metode OTP tidak valid",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidPhoneFormatError = &AppError{
		Code:       CodeInvalidPhoneFormat,
		Message:    "Format nomor HP tidak valid",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrPINMismatchError = &AppError{
		Code:       CodePINMismatch,
		Message:    "PIN dan konfirmasi tidak sama",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrWeakPINError = &AppError{
		Code:       CodeWeakPIN,
		Message:    "PIN terlalu lemah, hindari kombinasi mudah ditebak",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidOTPError = &AppError{
		Code:       CodeInvalidOTP,
		Message:    "Kode OTP salah",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrExpiredOTPError = &AppError{
		Code:       CodeExpiredOTP,
		Message:    "Kode OTP sudah kadaluarsa",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrInvalidPINError = &AppError{
		Code:       CodeInvalidPIN,
		Message:    "PIN yang Anda masukkan salah",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrExpiredTokenError = &AppError{
		Code:       CodeExpiredToken,
		Message:    "Token sudah kadaluarsa",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrTempTokenExpiredError = &AppError{
		Code:       CodeTempTokenExpired,
		Message:    "Sesi sudah kadaluarsa, silakan mulai ulang",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrInvalidRefreshTokenError = &AppError{
		Code:       CodeInvalidRefreshToken,
		Message:    "Refresh token tidak valid",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrUnauthorizedError = &AppError{
		Code:       CodeUnauthorized,
		Message:    "Akses tidak diizinkan",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrPINLockedError = &AppError{
		Code:       CodePINLocked,
		Message:    "PIN terkunci karena terlalu banyak percobaan salah",
		HTTPStatus: http.StatusForbidden,
	}

	ErrDeviceNotRecognizedError = &AppError{
		Code:       CodeDeviceNotRecognized,
		Message:    "Perangkat tidak dikenali, verifikasi diperlukan",
		HTTPStatus: http.StatusForbidden,
	}

	ErrCannotRemoveCurrentDeviceError = &AppError{
		Code:       CodeCannotRemoveCurrentDevice,
		Message:    "Tidak dapat menghapus perangkat yang sedang aktif",
		HTTPStatus: http.StatusForbidden,
	}

	ErrUserNotFoundError = &AppError{
		Code:       CodeUserNotFound,
		Message:    "Pengguna tidak ditemukan",
		HTTPStatus: http.StatusNotFound,
	}

	ErrDeviceNotFoundError = &AppError{
		Code:       CodeDeviceNotFound,
		Message:    "Perangkat tidak ditemukan",
		HTTPStatus: http.StatusNotFound,
	}

	ErrPhoneAlreadyRegisteredError = &AppError{
		Code:       CodePhoneAlreadyRegistered,
		Message:    "Nomor HP sudah terdaftar",
		HTTPStatus: http.StatusConflict,
	}

	ErrAccountLockedError = &AppError{
		Code:       CodeAccountLocked,
		Message:    "Akun terkunci, silakan hubungi customer service",
		HTTPStatus: http.StatusLocked,
	}

	ErrOTPResendLimitError = &AppError{
		Code:       CodeOTPResendLimit,
		Message:    "Batas pengiriman ulang OTP tercapai",
		HTTPStatus: http.StatusTooManyRequests,
	}

	ErrTooManyAttemptsError = &AppError{
		Code:       CodeTooManyAttempts,
		Message:    "Terlalu banyak percobaan, silakan tunggu beberapa saat",
		HTTPStatus: http.StatusTooManyRequests,
	}

	ErrRateLimitedError = &AppError{
		Code:       CodeRateLimited,
		Message:    "Terlalu banyak request, silakan tunggu",
		HTTPStatus: http.StatusTooManyRequests,
	}

	ErrOTPMaxAttemptsError = &AppError{
		Code:       CodeOTPMaxAttempts,
		Message:    "Batas percobaan OTP tercapai",
		HTTPStatus: http.StatusTooManyRequests,
	}

	ErrOTPRateLimitedError = &AppError{
		Code:       CodeOTPRateLimited,
		Message:    "Terlalu cepat, silakan tunggu sebelum meminta OTP kembali",
		HTTPStatus: http.StatusTooManyRequests,
	}

	ErrInternalServerError = &AppError{
		Code:       CodeInternalError,
		Message:    "Terjadi kesalahan sistem",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrOTPSendFailedError = &AppError{
		Code:       CodeOTPSendFailed,
		Message:    "Gagal mengirim OTP, silakan coba lagi",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrInvalidDeviceIDError = &AppError{
		Code:       CodeInvalidDeviceID,
		Message:    "Device ID tidak valid",
		HTTPStatus: http.StatusBadRequest,
	}

	// Transaction Errors
	ErrInvalidTarget = &AppError{
		Code:       CodeInvalidTarget,
		Message:    "Nomor tujuan tidak valid",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidProduct = &AppError{
		Code:       CodeInvalidProduct,
		Message:    "Produk tidak valid",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidServiceType = &AppError{
		Code:       CodeInvalidServiceType,
		Message:    "Jenis layanan tidak valid",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInquiryExpired = &AppError{
		Code:       CodeInquiryExpired,
		Message:    "Inquiry sudah expired",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInquiryNotFound = &AppError{
		Code:       CodeInquiryNotFound,
		Message:    "Inquiry tidak ditemukan",
		HTTPStatus: http.StatusNotFound,
	}

	ErrOrderExpired = &AppError{
		Code:       CodeOrderExpired,
		Message:    "Order sudah expired",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInsufficientBalance = &AppError{
		Code:       CodeInsufficientBalance,
		Message:    "Saldo tidak mencukupi",
		HTTPStatus: http.StatusPaymentRequired,
	}

	ErrDuplicateTransaction = &AppError{
		Code:       CodeDuplicateTransaction,
		Message:    "Transaksi duplikat",
		HTTPStatus: http.StatusConflict,
	}

	ErrProductUnavailable = &AppError{
		Code:       CodeProductUnavailable,
		Message:    "Produk tidak tersedia",
		HTTPStatus: http.StatusUnprocessableEntity,
	}

	ErrServiceUnavailable = &AppError{
		Code:       CodeServiceUnavailable,
		Message:    "Layanan sedang gangguan",
		HTTPStatus: http.StatusUnprocessableEntity,
	}

	ErrNoBill = &AppError{
		Code:       CodeNoBill,
		Message:    "Tidak ada tagihan",
		HTTPStatus: http.StatusUnprocessableEntity,
	}

	ErrKYCRequired = &AppError{
		Code:       CodeKYCRequired,
		Message:    "Verifikasi identitas diperlukan untuk transfer bank",
		HTTPStatus: http.StatusForbidden,
	}

	ErrInvalidVoucher = &AppError{
		Code:       CodeInvalidVoucher,
		Message:    "Kode voucher tidak valid",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrVoucherExpired = &AppError{
		Code:       CodeVoucherExpired,
		Message:    "Voucher sudah expired",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrVoucherUsed = &AppError{
		Code:       CodeVoucherUsed,
		Message:    "Voucher sudah digunakan",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrVoucherNotApplicable = &AppError{
		Code:       CodeVoucherNotApplicable,
		Message:    "Voucher tidak berlaku untuk transaksi ini",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrMinTransactionNotMet = &AppError{
		Code:       CodeMinTransactionNotMet,
		Message:    "Minimum transaksi tidak terpenuhi",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrMaxVouchersExceeded = &AppError{
		Code:       CodeMaxVouchersExceeded,
		Message:    "Melebihi batas voucher per transaksi",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrPINRequiredError = &AppError{
		Code:       CodePINRequired,
		Message:    "PIN wajib diisi",
		HTTPStatus: http.StatusBadRequest,
	}

	// Deposit Errors
	ErrInvalidAmount = &AppError{
		Code:       CodeInvalidAmount,
		Message:    "Nominal tidak valid",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrAmountTooLow = &AppError{
		Code:       CodeAmountTooLow,
		Message:    "Nominal di bawah minimum",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrAmountTooHigh = &AppError{
		Code:       CodeAmountTooHigh,
		Message:    "Nominal melebihi maksimum",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidProvider = &AppError{
		Code:       CodeInvalidProvider,
		Message:    "Provider tidak valid",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidBank = &AppError{
		Code:       CodeInvalidBank,
		Message:    "Bank tidak valid",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrDepositNotFound = &AppError{
		Code:       CodeDepositNotFound,
		Message:    "Deposit tidak ditemukan",
		HTTPStatus: http.StatusNotFound,
	}

	ErrDepositAlreadyPaid = &AppError{
		Code:       CodeDepositAlreadyPaid,
		Message:    "Deposit sudah dibayar",
		HTTPStatus: http.StatusConflict,
	}

	ErrDepositExpired = &AppError{
		Code:       CodeDepositExpired,
		Message:    "Deposit sudah expired",
		HTTPStatus: http.StatusGone,
	}

	ErrMethodUnavailable = &AppError{
		Code:       CodeMethodUnavailable,
		Message:    "Metode tidak tersedia saat ini",
		HTTPStatus: http.StatusUnprocessableEntity,
	}

	ErrTooManyPending = &AppError{
		Code:       CodeTooManyPending,
		Message:    "Terlalu banyak deposit pending. Selesaikan atau batalkan deposit sebelumnya",
		HTTPStatus: http.StatusTooManyRequests,
	}
)

// Aliases for easier use in service layer
var (
	ErrInvalidPhone    = ErrInvalidPhoneFormatError
	ErrInvalidDeviceID = ErrInvalidDeviceIDError
	ErrOTPInvalid      = ErrInvalidOTPError
	ErrOTPExpired      = ErrExpiredOTPError
	ErrOTPMaxAttempts  = ErrOTPMaxAttemptsError
	ErrOTPRateLimited  = ErrOTPRateLimitedError
	ErrOTPResendLimit  = ErrOTPResendLimitError
	ErrOTPSendFailed   = ErrOTPSendFailedError
	ErrPINLocked       = ErrPINLockedError
	ErrWrongPIN        = ErrInvalidPINError
	ErrTokenExpired    = ErrExpiredTokenError
	ErrInvalidToken    = ErrInvalidTokenError
	ErrUnauthorized    = ErrUnauthorizedError
	ErrUserNotFound    = ErrUserNotFoundError
)

// ErrValidationFailed creates a validation error with custom message
func ErrValidationFailed(message string) *AppError {
	return &AppError{
		Code:       "VALIDATION_FAILED",
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// ErrNotFound creates a NOT_FOUND error for a resource
func ErrNotFound(resource string) *AppError {
	return &AppError{
		Code:       "NOT_FOUND",
		Message:    fmt.Sprintf("%s tidak ditemukan", resource),
		HTTPStatus: http.StatusNotFound,
	}
}

// ErrWithRemainingAttempts creates an error with remaining attempts
func ErrWithRemainingAttempts(baseErr *AppError, remaining int) *AppError {
	return &AppError{
		Code:              baseErr.Code,
		Message:           baseErr.Message,
		HTTPStatus:        baseErr.HTTPStatus,
		RemainingAttempts: &remaining,
	}
}

// ErrWithLockUntil creates an error with lock until timestamp
func ErrWithLockUntil(baseErr *AppError, lockUntil string) *AppError {
	zero := 0
	return &AppError{
		Code:              baseErr.Code,
		Message:           baseErr.Message,
		HTTPStatus:        baseErr.HTTPStatus,
		RemainingAttempts: &zero,
		LockUntil:         &lockUntil,
	}
}

// ErrWithRetryAfter creates an error with retry after seconds
func ErrWithRetryAfter(baseErr *AppError, retryAfter int) *AppError {
	return &AppError{
		Code:       baseErr.Code,
		Message:    baseErr.Message,
		HTTPStatus: baseErr.HTTPStatus,
		RetryAfter: &retryAfter,
	}
}
