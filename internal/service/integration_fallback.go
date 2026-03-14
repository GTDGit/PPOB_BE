package service

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/domain"
)

func buildDummyPaymentID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func buildDummyReference(prefix string) string {
	return fmt.Sprintf("%s%s", prefix, strings.ToUpper(strconv.FormatInt(time.Now().UnixNano()%1_000_000_000, 36)))
}

func buildDummyQRISString(referenceID string, amount int64) string {
	return fmt.Sprintf("00020101021226670016COM.PPOB.ID01189360000000000000202%s520458145303360540%d5802ID5912PPOB%%20ID6007JAKARTA6105123406304%s",
		normalizeDigits(referenceID, 12),
		amount,
		strings.ToUpper(normalizeDigits(referenceID, 4)),
	)
}

func buildDummyVANumber(bankCode, userID string) string {
	return normalizeDigits(bankCode+userID, 16)
}

func buildDummyPaymentCode(providerCode, depositID string) string {
	return strings.ToUpper(fmt.Sprintf("%s%s", sanitizeAlpha(providerCode, 4), normalizeDigits(depositID, 10)))
}

func buildDummyNIK(seed string) string {
	base := normalizeDigits(seed, 16)
	if len(base) == 16 {
		return base
	}

	hash := sha1.Sum([]byte(seed + time.Now().Format(time.RFC3339Nano)))
	digits := normalizeDigits(hex.EncodeToString(hash[:]), 16)
	if len(digits) == 16 {
		return digits
	}

	return "3175000101010001"
}

func buildDummyOCRResult(user *domain.User) *domain.KTPOCRResult {
	fullName := strings.TrimSpace(user.FullName)
	if fullName == "" {
		fullName = "PENGGUNA PPOB"
	}

	gender := domain.GenderMale
	if user.Gender != nil && strings.TrimSpace(*user.Gender) != "" {
		gender = strings.ToUpper(strings.TrimSpace(*user.Gender))
	}

	return &domain.KTPOCRResult{
		NIK:          buildDummyNIK(user.ID + user.Phone),
		FullName:     fullName,
		PlaceOfBirth: "JAKARTA",
		DateOfBirth:  "1990-01-01",
		Gender:       gender,
		Religion:     domain.ReligionIslam,
		Address: domain.KTPAddress{
			Street:      "JL. CONTOH RAYA NO. 1",
			RT:          "001",
			RW:          "002",
			SubDistrict: "GAMBIR",
			District:    "GAMBIR",
			City:        "JAKARTA PUSAT",
			Province:    "DKI JAKARTA",
		},
		AdministrativeCode: domain.AdministrativeCode{
			Province:    "31",
			City:        "3171",
			District:    "317101",
			SubDistrict: "3171011001",
		},
	}
}

func parseFlexibleDate(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02",
		"02-01-2006",
		"02/01/2006",
		"2 January 2006",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed
		}
	}

	return time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)
}

func stringFromAny(value interface{}) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typed)
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	case map[string]interface{}:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", typed))
	}
}

func floatFromAny(value interface{}) float64 {
	switch typed := value.(type) {
	case nil:
		return 0
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case string:
		parsed, _ := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		return parsed
	default:
		return 0
	}
}

func mapStringFromAny(value interface{}) map[string]string {
	switch typed := value.(type) {
	case nil:
		return map[string]string{}
	case map[string]string:
		return typed
	case map[string]interface{}:
		result := make(map[string]string, len(typed))
		for key, item := range typed {
			result[key] = stringFromAny(item)
		}
		return result
	case domain.AdministrativeCode:
		return map[string]string{
			"province":    typed.Province,
			"city":        typed.City,
			"district":    typed.District,
			"subDistrict": typed.SubDistrict,
		}
	default:
		return map[string]string{}
	}
}

func pointerIfNotEmpty(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func normalizeDigits(seed string, length int) string {
	var digits strings.Builder
	for _, char := range seed {
		if char >= '0' && char <= '9' {
			digits.WriteRune(char)
		}
	}

	result := digits.String()
	if result == "" {
		result = "1234567890"
	}
	for len(result) < length {
		result += result
	}

	return result[:length]
}

func sanitizeAlpha(seed string, length int) string {
	var builder strings.Builder
	for _, char := range strings.ToUpper(seed) {
		if char >= 'A' && char <= 'Z' {
			builder.WriteRune(char)
		}
	}

	result := builder.String()
	if result == "" {
		result = "PPOB"
	}
	for len(result) < length {
		result += "X"
	}
	return result[:length]
}
