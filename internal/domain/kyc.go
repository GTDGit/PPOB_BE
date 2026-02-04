package domain

import "time"

// KYC Session Status (for session tracking)
const (
	KYCSessionPending    = "pending"
	KYCSessionProcessing = "processing"
	KYCSessionCompleted  = "completed"
	KYCSessionFailed     = "failed"
)

// Note: KYCStatus constants (unverified, pending, verified, rejected) are defined in user.go

// Gender constants
const (
	GenderMale   = "MALE"
	GenderFemale = "FEMALE"
)

// Religion constants
const (
	ReligionIslam     = "ISLAM"
	ReligionChristian = "CHRISTIAN"
	ReligionCatholic  = "CATHOLIC"
	ReligionHindu     = "HINDU"
	ReligionBuddhist  = "BUDDHIST"
	ReligionConfucian = "CONFUCIAN"
	ReligionOther     = "OTHER"
)

// KYCSession represents temporary KYC verification session
type KYCSession struct {
	ID             string                 `json:"id" db:"id"`
	UserID         string                 `json:"userId" db:"user_id"`
	NIK            *string                `json:"nik" db:"nik"`
	Status         string                 `json:"status" db:"status"`
	CurrentStep    int                    `json:"currentStep" db:"current_step"`
	OCRData        map[string]interface{} `json:"ocrData" db:"ocr_data"`
	FaceUrls       map[string]string      `json:"faceUrls" db:"face_urls"`
	LivenessData   map[string]interface{} `json:"livenessData" db:"liveness_data"`
	FaceComparison map[string]interface{} `json:"faceComparison" db:"face_comparison"`
	ErrorMessage   *string                `json:"errorMessage" db:"error_message"`
	ExpiresAt      time.Time              `json:"expiresAt" db:"expires_at"`
	CreatedAt      time.Time              `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time              `json:"updatedAt" db:"updated_at"`
}

// KYCVerification represents permanent KYC verification data
type KYCVerification struct {
	ID                 string            `json:"id" db:"id"`
	UserID             string            `json:"userId" db:"user_id"`
	NIK                string            `json:"nik" db:"nik"`
	FullName           string            `json:"fullName" db:"full_name"`
	PlaceOfBirth       string            `json:"placeOfBirth" db:"place_of_birth"`
	DateOfBirth        time.Time         `json:"dateOfBirth" db:"date_of_birth"`
	Gender             string            `json:"gender" db:"gender"`
	Religion           string            `json:"religion" db:"religion"`
	AddressStreet      string            `json:"addressStreet" db:"address_street"`
	AddressRT          *string           `json:"addressRt" db:"address_rt"`
	AddressRW          *string           `json:"addressRw" db:"address_rw"`
	AddressSubDistrict string            `json:"addressSubDistrict" db:"address_sub_district"`
	AddressDistrict    string            `json:"addressDistrict" db:"address_district"`
	AddressCity        string            `json:"addressCity" db:"address_city"`
	AddressProvince    string            `json:"addressProvince" db:"address_province"`
	AdministrativeCode map[string]string `json:"administrativeCode" db:"administrative_code"`
	KTPUrl             *string           `json:"ktpUrl" db:"ktp_url"`
	FaceUrl            *string           `json:"faceUrl" db:"face_url"`
	FaceWithKTPUrl     *string           `json:"faceWithKtpUrl" db:"face_with_ktp_url"`
	LivenessUrl        *string           `json:"livenessUrl" db:"liveness_url"`
	FaceSimilarity     *float64          `json:"faceSimilarity" db:"face_similarity"`
	LivenessConfidence *float64          `json:"livenessConfidence" db:"liveness_confidence"`
	VerifiedAt         time.Time         `json:"verifiedAt" db:"verified_at"`
	CreatedAt          time.Time         `json:"createdAt" db:"created_at"`
}

// KTPOCRResult represents OCR result from Gerbang API
type KTPOCRResult struct {
	NIK                string             `json:"nik"`
	FullName           string             `json:"fullName"`
	PlaceOfBirth       string             `json:"placeOfBirth"`
	DateOfBirth        string             `json:"dateOfBirth"`
	Gender             string             `json:"gender"`
	Religion           string             `json:"religion"`
	Address            KTPAddress         `json:"address"`
	AdministrativeCode AdministrativeCode `json:"administrativeCode"`
}

// KTPAddress represents address from KTP
type KTPAddress struct {
	Street      string `json:"street"`
	RT          string `json:"rt"`
	RW          string `json:"rw"`
	SubDistrict string `json:"subDistrict"`
	District    string `json:"district"`
	City        string `json:"city"`
	Province    string `json:"province"`
}

// AdministrativeCode represents administrative code from KTP
type AdministrativeCode struct {
	Province    string `json:"province"`
	City        string `json:"city"`
	District    string `json:"district"`
	SubDistrict string `json:"subDistrict"`
}

// FaceCompareResult represents face comparison result
type FaceCompareResult struct {
	Matched    bool    `json:"matched"`
	Similarity float64 `json:"similarity"`
	Threshold  float64 `json:"threshold"`
}

// KYCHistory represents audit log for KYC activities
type KYCHistory struct {
	ID           int                    `db:"id"`
	UserID       string                 `db:"user_id"`
	SessionID    *string                `db:"session_id"`
	Action       string                 `db:"action"`
	Status       string                 `db:"status"`
	Metadata     map[string]interface{} `db:"metadata"`
	ErrorMessage *string                `db:"error_message"`
	CreatedAt    time.Time              `db:"created_at"`
}

// KYC History actions
const (
	KYCActionSessionCreated       = "session_created"
	KYCActionOCRCompleted         = "ocr_completed"
	KYCActionFaceCaptured         = "face_captured"
	KYCActionLivenessChecked      = "liveness_checked"
	KYCActionVerificationApproved = "verification_approved"
	KYCActionVerificationRejected = "verification_rejected"
)
