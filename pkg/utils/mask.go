package utils

import "strings"

// MaskPhone masks phone number: 081234567890 -> 0812****7890
func MaskPhone(phone string) string {
	if len(phone) < 8 {
		return "****"
	}
	return phone[:4] + "****" + phone[len(phone)-4:]
}

// MaskEmail masks email: john@example.com -> j***@example.com
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 || len(parts[0]) == 0 {
		return "****@****"
	}
	return parts[0][:1] + "***@" + parts[1]
}

// MaskToken masks sensitive tokens: abc123def456 -> abc***456
func MaskToken(token string) string {
	if len(token) < 6 {
		return "***"
	}
	return token[:3] + "***" + token[len(token)-3:]
}
