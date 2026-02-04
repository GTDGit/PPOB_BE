package validator

import (
	"regexp"
	"strings"
)

var (
	phoneRegex = regexp.MustCompile(`^(\+62|62|0)8[1-9][0-9]{7,11}$`)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// ValidatePhone validates Indonesian phone number
func ValidatePhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

// NormalizePhone normalizes phone number to 08xxx format
func NormalizePhone(phone string) string {
	// Remove spaces and dashes
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	// Convert +62 to 0
	if strings.HasPrefix(phone, "+62") {
		phone = "0" + phone[3:]
	}

	// Convert 62 to 0
	if strings.HasPrefix(phone, "62") && !strings.HasPrefix(phone, "0") {
		phone = "0" + phone[2:]
	}

	return phone
}

// FormatPhoneInternational formats phone to international format (628xxx)
func FormatPhoneInternational(phone string) string {
	phone = NormalizePhone(phone)
	if strings.HasPrefix(phone, "0") {
		return "62" + phone[1:]
	}
	return phone
}

// MaskPhone masks a phone number for display (0812****6789)
func MaskPhone(phone string) string {
	phone = NormalizePhone(phone)
	if len(phone) < 8 {
		return phone
	}
	return phone[:4] + "****" + phone[len(phone)-4:]
}

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// MaskEmail masks an email for display (bu**@mail.com)
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	local := parts[0]
	domain := parts[1]

	if len(local) <= 2 {
		return local + "**@" + domain
	}

	return local[:2] + "**@" + domain
}

// ValidatePIN validates PIN format (6 digits)
func ValidatePIN(pin string) bool {
	if len(pin) != 6 {
		return false
	}

	for _, c := range pin {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

// ValidateOTP validates OTP format (4 digits as per spec)
func ValidateOTP(otp string) bool {
	if len(otp) != 4 {
		return false
	}

	for _, c := range otp {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

// ValidateDeviceID validates device ID (UUID format or similar)
func ValidateDeviceID(deviceID string) bool {
	// Basic validation - at least 8 characters
	return len(deviceID) >= 8
}

// ValidateName validates name (min 3 chars, max 100)
func ValidateName(name string) bool {
	name = strings.TrimSpace(name)
	return len(name) >= 3 && len(name) <= 100
}

// SanitizeName sanitizes name input
func SanitizeName(name string) string {
	name = strings.TrimSpace(name)
	// Remove extra spaces
	space := regexp.MustCompile(`\s+`)
	name = space.ReplaceAllString(name, " ")
	return name
}
