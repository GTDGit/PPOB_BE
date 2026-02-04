package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/external/gerbang"
	"github.com/GTDGit/PPOB_BE/internal/external/s3"
	"github.com/GTDGit/PPOB_BE/internal/repository"
)

// KYCService handles KYC verification business logic
type KYCService struct {
	kycRepo       repository.KYCRepository
	userRepo      repository.UserRepository
	gerbangClient *gerbang.Client
	s3Client      *s3.Client // HANYA untuk KTP + face photos, BUKAN liveness
}

// NewKYCService creates a new KYC service
func NewKYCService(
	kycRepo repository.KYCRepository,
	userRepo repository.UserRepository,
	gerbangClient *gerbang.Client,
	s3Client *s3.Client,
) *KYCService {
	return &KYCService{
		kycRepo:       kycRepo,
		userRepo:      userRepo,
		gerbangClient: gerbangClient,
		s3Client:      s3Client,
	}
}

// GetStatus returns KYC verification status for a user
func (s *KYCService) GetStatus(ctx context.Context, userID string) (*domain.KYCVerification, error) {
	// Check if user has completed verification
	verification, err := s.kycRepo.FindVerificationByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get verification: %w", err)
	}

	return verification, nil
}

// GetActiveSession returns active KYC session for a user
func (s *KYCService) GetActiveSession(ctx context.Context, userID string) (*domain.KYCSession, error) {
	session, err := s.kycRepo.FindActiveSessionByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

// StartVerification creates a new KYC session
func (s *KYCService) StartVerification(ctx context.Context, userID string) (*domain.KYCSession, error) {
	// Check if user already verified
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrNotFound("User")
	}

	if user.KYCStatus == domain.KYCStatusVerified {
		return nil, domain.ErrValidationFailed("User sudah terverifikasi")
	}

	// Delete any existing session
	if err := s.kycRepo.DeleteSessionByUserID(ctx, userID); err != nil {
		return nil, fmt.Errorf("failed to delete old session: %w", err)
	}

	// Create new session
	session := &domain.KYCSession{
		ID:          uuid.New().String(),
		UserID:      userID,
		Status:      domain.KYCSessionPending,
		CurrentStep: 1,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.kycRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Update user KYC status to pending
	user.KYCStatus = domain.KYCStatusPending
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user status: %w", err)
	}

	return session, nil
}

// CancelVerification cancels an active KYC session
func (s *KYCService) CancelVerification(ctx context.Context, userID string) error {
	// Delete session
	if err := s.kycRepo.DeleteSessionByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Update user KYC status back to unverified
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return domain.ErrNotFound("User")
	}

	user.KYCStatus = domain.KYCStatusUnverified
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	return nil
}

// UploadKTP uploads KTP image, runs OCR, and updates session
func (s *KYCService) UploadKTP(ctx context.Context, userID string, sessionID string, fileBytes []byte, filename string) error {
	// 1. Find session
	session, err := s.kycRepo.FindSessionByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return domain.NewError(domain.CodeKYCSessionNotFound, "Sesi verifikasi tidak ditemukan", 404)
	}
	if session.UserID != userID {
		return domain.ErrUnauthorized
	}

	// Check if already uploaded
	if session.CurrentStep >= 1 {
		return domain.NewError(domain.CodeKYCInvalidFile, "KTP sudah di-upload", 400)
	}

	// Validate user status
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user.KYCStatus == domain.KYCStatusVerified {
		return domain.NewError(domain.CodeKYCAlreadyVerified, "Akun sudah terverifikasi", 400)
	}

	// 2. Upload to S3 (ap-southeast-3 Jakarta)
	ktpURL, err := s.s3Client.UploadBytes(ctx, fileBytes, fmt.Sprintf("kyc/%s", userID), filename, "image/jpeg")
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	// 3. Run OCR via Gerbang API
	ocrResult, err := s.gerbangClient.KTPOCR(ctx, ktpURL)
	if err != nil {
		return domain.NewError(domain.CodeKYCOCRFailed, "Gagal membaca KTP", 400)
	}

	// 4. Check NIK uniqueness
	isUsed, err := s.kycRepo.IsNIKVerified(ctx, ocrResult.NIK)
	if err != nil {
		return fmt.Errorf("failed to check NIK: %w", err)
	}
	if isUsed {
		return domain.NewError(domain.CodeKYCNIKAlreadyUsed, "NIK sudah terdaftar", 409)
	}

	// 5. Update session with OCR results
	session.NIK = &ocrResult.NIK
	session.OCRData = map[string]interface{}{
		"fullName":           ocrResult.FullName,
		"placeOfBirth":       ocrResult.PlaceOfBirth,
		"dateOfBirth":        ocrResult.DateOfBirth,
		"gender":             ocrResult.Gender,
		"religion":           ocrResult.Religion,
		"addressStreet":      ocrResult.Address.Street,
		"addressRt":          ocrResult.Address.RT,
		"addressRw":          ocrResult.Address.RW,
		"addressSubDistrict": ocrResult.Address.SubDistrict,
		"addressDistrict":    ocrResult.Address.District,
		"addressCity":        ocrResult.Address.City,
		"addressProvince":    ocrResult.Address.Province,
		"administrativeCode": ocrResult.AdministrativeCode,
	}
	session.FaceUrls = map[string]string{
		"ktp": ktpURL,
	}
	session.CurrentStep = 1
	session.UpdatedAt = time.Now()

	if err := s.kycRepo.UpdateSession(ctx, session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	// Update user status to pending
	user.KYCStatus = domain.KYCStatusPending
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	return nil
}

// UploadFacePhotos uploads face photos (selfie, face with KTP)
func (s *KYCService) UploadFacePhotos(ctx context.Context, userID string, sessionID string, faceBytes []byte, faceFilename string, fullImageBytes []byte, fullImageFilename string) error {
	// 1. Find session
	session, err := s.kycRepo.FindSessionByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return domain.NewError(domain.CodeKYCSessionNotFound, "Sesi verifikasi tidak ditemukan", 404)
	}
	if session.UserID != userID {
		return domain.ErrUnauthorized
	}

	// Validate step (must have uploaded KTP first)
	if session.CurrentStep < 1 {
		return domain.NewError(domain.CodeKYCInvalidFile, "Harap upload KTP terlebih dahulu", 400)
	}

	// Check if already uploaded
	if session.CurrentStep >= 2 {
		return domain.NewError(domain.CodeKYCInvalidFile, "Foto wajah sudah di-upload", 400)
	}

	// 2. Upload face photo (selfie) to S3
	faceURL, err := s.s3Client.UploadBytes(ctx, faceBytes, fmt.Sprintf("kyc/%s", userID), faceFilename, "image/jpeg")
	if err != nil {
		return fmt.Errorf("failed to upload face photo: %w", err)
	}

	// 3. Upload full image (face + KTP) to S3
	fullImageURL, err := s.s3Client.UploadBytes(ctx, fullImageBytes, fmt.Sprintf("kyc/%s", userID), fullImageFilename, "image/jpeg")
	if err != nil {
		return fmt.Errorf("failed to upload full image: %w", err)
	}

	// 4. Get KTP face URL from session
	ktpFaceURL, exists := session.FaceUrls["ktp"]
	if !exists || ktpFaceURL == "" {
		return domain.NewError(domain.CodeKYCInvalidFile, "URL foto KTP tidak ditemukan", 400)
	}

	// 5. Compare faces (KTP photo vs selfie) via Gerbang API
	compareResult, err := s.gerbangClient.CompareFaces(ctx, ktpFaceURL, faceURL)
	if err != nil {
		return domain.NewError(domain.CodeKYCFaceNoMatch, "Gagal membandingkan wajah", 400)
	}

	// 6. Check face similarity threshold (e.g., >70%)
	if compareResult.Similarity < 70.0 {
		errMsg := "Wajah tidak cocok dengan KTP"
		session.Status = domain.KYCSessionFailed
		session.ErrorMessage = &errMsg
		s.kycRepo.UpdateSession(ctx, session)
		return domain.NewError(domain.CodeKYCFaceNoMatch, errMsg, 400)
	}

	// 7. Update session with face URLs and comparison results
	if session.FaceUrls == nil {
		session.FaceUrls = make(map[string]string)
	}
	session.FaceUrls["face"] = faceURL
	session.FaceUrls["fullImage"] = fullImageURL
	session.FaceComparison = map[string]interface{}{
		"similarity": compareResult.Similarity,
		"matched":    compareResult.Matched,
		"threshold":  compareResult.Threshold,
	}
	session.CurrentStep = 2
	session.UpdatedAt = time.Now()

	if err := s.kycRepo.UpdateSession(ctx, session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// CreateLivenessSession creates liveness session for frontend SDK
// POST /v1/kyc/liveness/session
func (s *KYCService) CreateLivenessSession(ctx context.Context, userID string, kycSessionID string) (map[string]interface{}, error) {
	// 1. Find KYC session
	session, err := s.kycRepo.FindSessionByID(ctx, kycSessionID)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}
	if session == nil {
		return nil, domain.NewError(domain.CodeKYCSessionNotFound, "Sesi verifikasi tidak ditemukan", 404)
	}

	// 2. Validate session ownership
	if session.UserID != userID {
		return nil, domain.ErrUnauthorized
	}

	// 3. Check current step (must have uploaded face photos first)
	if session.CurrentStep < 2 {
		return nil, domain.NewError(domain.CodeKYCInvalidFile, "Harap upload foto wajah terlebih dahulu", 400)
	}

	// 4. Call Gerbang API to create liveness session
	livenessResp, err := s.gerbangClient.CreateLivenessSession(ctx, *session.NIK)
	if err != nil {
		return nil, fmt.Errorf("create liveness session: %w", err)
	}

	// 5. Update KYC session with liveness session ID
	livenessSessID := livenessResp.SessionID
	session.LivenessData = map[string]interface{}{
		"sessionId": livenessSessID,
		"expiresAt": livenessResp.ExpiresAt,
	}
	session.UpdatedAt = time.Now()

	if err := s.kycRepo.UpdateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("update session: %w", err)
	}

	// 6. Return session ID for frontend FaceLivenessDetector
	return map[string]interface{}{
		"livenessSessionId": livenessSessID,
		"expiresAt":         livenessResp.ExpiresAt,
	}, nil
}

// VerifyLiveness verifies liveness after frontend completion
// POST /v1/kyc/liveness/verify
func (s *KYCService) VerifyLiveness(ctx context.Context, userID string, kycSessionID string) (map[string]interface{}, error) {
	// 1. Find KYC session
	session, err := s.kycRepo.FindSessionByID(ctx, kycSessionID)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}
	if session == nil {
		return nil, domain.NewError(domain.CodeKYCSessionNotFound, "Sesi verifikasi tidak ditemukan", 404)
	}

	// 2. Validate session ownership
	if session.UserID != userID {
		return nil, domain.ErrUnauthorized
	}

	// 3. Get liveness session ID
	if session.LivenessData == nil {
		return nil, domain.NewError(domain.CodeKYCInvalidFile, "Session liveness belum dibuat", 400)
	}

	livenessSessionID, ok := session.LivenessData["sessionId"].(string)
	if !ok || livenessSessionID == "" {
		return nil, domain.NewError(domain.CodeKYCInvalidFile, "Session liveness tidak valid", 400)
	}

	// 4. Call Gerbang API to verify liveness
	livenessResult, err := s.gerbangClient.VerifyLiveness(ctx, livenessSessionID)
	if err != nil {
		return nil, fmt.Errorf("verify liveness: %w", err)
	}

	// 5. Check liveness result
	if !livenessResult.IsLive {
		errMsg := "Liveness check gagal"
		if livenessResult.FailureReason != nil {
			errMsg = *livenessResult.FailureReason
		}
		session.Status = domain.KYCSessionFailed
		session.ErrorMessage = &errMsg
		s.kycRepo.UpdateSession(ctx, session)
		return nil, domain.NewError(domain.CodeKYCLivenessFailed, errMsg, 400)
	}

	// 6. Get liveness face URL from Gerbang response (NOT uploaded by us)
	var livenessFaceURL string
	if livenessResult.File != nil {
		livenessFaceURL = livenessResult.File.Face
	}

	// 7. Update session with liveness results
	session.LivenessData["confidence"] = livenessResult.Confidence
	session.LivenessData["faceUrl"] = livenessFaceURL
	session.LivenessData["isLive"] = livenessResult.IsLive
	session.CurrentStep = 3
	session.Status = domain.KYCSessionCompleted
	session.UpdatedAt = time.Now()

	if err := s.kycRepo.UpdateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("update session: %w", err)
	}

	// 8. Update user KYC status to verified (simplified - no face comparison for now)
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	user.KYCStatus = domain.KYCStatusVerified
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	// 9. Create verification record
	verification := &domain.KYCVerification{
		ID:                 uuid.New().String(),
		UserID:             userID,
		NIK:                *session.NIK,
		FullName:           session.OCRData["fullName"].(string),
		PlaceOfBirth:       session.OCRData["placeOfBirth"].(string),
		DateOfBirth:        time.Now(), // TODO: Parse from OCR
		Gender:             session.OCRData["gender"].(string),
		FaceSimilarity:     nil, // TODO: Add face comparison
		LivenessConfidence: &livenessResult.Confidence,
		VerifiedAt:         time.Now(),
	}

	if err := s.kycRepo.CreateVerification(ctx, verification); err != nil {
		return nil, fmt.Errorf("create verification: %w", err)
	}

	// 10. Clean up session
	s.kycRepo.DeleteSessionByUserID(ctx, userID)

	// 11. Return result
	return map[string]interface{}{
		"status":             domain.KYCStatusVerified,
		"nik":                maskNIK(*session.NIK),
		"livenessConfidence": livenessResult.Confidence,
		"verifiedAt":         time.Now(),
	}, nil
}

// Helper function to mask NIK
func maskNIK(nik string) string {
	if len(nik) < 16 {
		return nik
	}
	return nik[:6] + "******" + nik[12:]
}

// SubmitForReview submits KYC session for final review/approval
// TODO: Implement when all steps are complete
func (s *KYCService) SubmitForReview(ctx context.Context, userID string, sessionID string) error {
	// 1. Find session
	// 2. Validate all steps complete
	// 3. Create KYCVerification record
	// 4. Update user.kyc_status = verified
	// 5. Delete session
	return fmt.Errorf("KYC SubmitForReview: not yet implemented - requires complete flow")
}
