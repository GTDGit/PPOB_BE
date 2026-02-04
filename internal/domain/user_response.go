package domain

// ProfileResponse represents user profile response
type ProfileResponse struct {
	User     *ProfileUser     `json:"user"`
	Stats    *ProfileStats    `json:"stats"`
	TierInfo *ProfileTierInfo `json:"tierInfo"`
}

// ProfileUser represents user profile data
type ProfileUser struct {
	ID           string  `json:"id"`
	MIC          string  `json:"mic"`
	Phone        string  `json:"phone"`
	FullName     string  `json:"fullName"`
	Email        *string `json:"email,omitempty"`
	Gender       *string `json:"gender,omitempty"`
	Tier         string  `json:"tier"`
	AvatarURL    *string `json:"avatarUrl,omitempty"`
	KYCStatus    string  `json:"kycStatus"`
	BusinessType *string `json:"businessType,omitempty"`
	CreatedAt    string  `json:"createdAt"`
}

// ProfileStats represents user statistics
type ProfileStats struct {
	TotalTransactions               int    `json:"totalTransactions"`
	TotalTransactionAmount          int64  `json:"totalTransactionAmount"`
	TotalTransactionAmountFormatted string `json:"totalTransactionAmountFormatted"`
	MemberSince                     string `json:"memberSince"`
	MemberDays                      int    `json:"memberDays"`
}

// ProfileTierInfo represents tier information and progress
type ProfileTierInfo struct {
	CurrentTier      string   `json:"currentTier"`
	NextTier         *string  `json:"nextTier,omitempty"`
	Progress         int      `json:"progress"`
	PointsToNextTier *int     `json:"pointsToNextTier,omitempty"`
	Benefits         []string `json:"benefits"`
}

// UpdateProfileResponse represents update profile response
type UpdateProfileResponse struct {
	Updated bool         `json:"updated"`
	User    *ProfileUser `json:"user"`
}

// AvatarUploadResponse represents avatar upload response
type AvatarUploadResponse struct {
	Uploaded  bool   `json:"uploaded"`
	AvatarURL string `json:"avatarUrl"`
}

// AvatarDeleteResponse represents avatar delete response
type AvatarDeleteResponse struct {
	Deleted bool `json:"deleted"`
}

// SettingsResponse represents user settings response
type SettingsResponse struct {
	Settings *UserSettingsDetail `json:"settings"`
}

// UserSettingsDetail represents detailed user settings
type UserSettingsDetail struct {
	PINRequiredForTransaction     bool   `json:"pinRequiredForTransaction"`
	PINRequiredMinAmount          int64  `json:"pinRequiredMinAmount"`
	PINRequiredMinAmountFormatted string `json:"pinRequiredMinAmountFormatted"`
	BiometricEnabled              bool   `json:"biometricEnabled"`
	DefaultSellingPriceMarkup     int    `json:"defaultSellingPriceMarkup"`
	NotificationEnabled           bool   `json:"notificationEnabled"`
	EmailNotificationEnabled      bool   `json:"emailNotificationEnabled"`
	WhatsAppNotificationEnabled   bool   `json:"whatsappNotificationEnabled"`
	Language                      string `json:"language"`
	Theme                         string `json:"theme"`
}

// UpdateSettingsResponse represents update settings response
type UpdateSettingsResponse struct {
	Updated  bool                `json:"updated"`
	Settings *UserSettingsDetail `json:"settings"`
}

// ReferralInfoResponse represents referral info response
type ReferralInfoResponse struct {
	ReferralCode           string         `json:"referralCode"`
	ReferralCount          int            `json:"referralCount"`
	TotalEarnings          int64          `json:"totalEarnings"`
	TotalEarningsFormatted string         `json:"totalEarningsFormatted"`
	ShareLink              string         `json:"shareLink"`
	Stats                  *ReferralStats `json:"stats"`
}

// ReferralStats represents referral statistics
type ReferralStats struct {
	ThisMonth              int    `json:"thisMonth"`
	LastMonth              int    `json:"lastMonth"`
	EarningsToday          int64  `json:"earningsToday"`
	EarningsTodayFormatted string `json:"earningsTodayFormatted"`
}

// ReferralHistoryResponse represents referral history response
type ReferralHistoryResponse struct {
	Referrals []*ReferralDetail `json:"referrals"`
	Total     int               `json:"total"`
}

// ReferralDetail represents detailed referral information
type ReferralDetail struct {
	ID                    string `json:"id"`
	FullName              string `json:"fullName"`
	Phone                 string `json:"phone"`
	JoinedAt              string `json:"joinedAt"`
	Status                string `json:"status"` // active, inactive
	RewardEarned          int64  `json:"rewardEarned"`
	RewardEarnedFormatted string `json:"rewardEarnedFormatted"`
	FirstTransaction      bool   `json:"firstTransaction"`
}
