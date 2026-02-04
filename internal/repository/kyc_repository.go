package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/GTDGit/PPOB_BE/internal/domain"
)

// KYCRepository defines the interface for KYC data operations
type KYCRepository interface {
	// Session methods
	CreateSession(ctx context.Context, session *domain.KYCSession) error
	FindSessionByID(ctx context.Context, id string) (*domain.KYCSession, error)
	FindActiveSessionByUserID(ctx context.Context, userID string) (*domain.KYCSession, error)
	UpdateSession(ctx context.Context, session *domain.KYCSession) error
	DeleteExpiredSessions(ctx context.Context) error
	DeleteSessionByUserID(ctx context.Context, userID string) error

	// Verification methods
	CreateVerification(ctx context.Context, v *domain.KYCVerification) error
	FindVerificationByUserID(ctx context.Context, userID string) (*domain.KYCVerification, error)
	IsNIKVerified(ctx context.Context, nik string) (bool, error)
}

// kycRepository implements KYCRepository
type kycRepository struct {
	db *sqlx.DB
}

// NewKYCRepository creates a new KYC repository
func NewKYCRepository(db *sqlx.DB) KYCRepository {
	return &kycRepository{db: db}
}

// ========== Session Methods ==========

func (r *kycRepository) CreateSession(ctx context.Context, session *domain.KYCSession) error {
	query := `
		INSERT INTO kyc_sessions (id, user_id, nik, status, current_step, ocr_data, face_urls, 
								 liveness_data, face_comparison, error_message, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	ocrData, _ := json.Marshal(session.OCRData)
	faceUrls, _ := json.Marshal(session.FaceUrls)
	livenessData, _ := json.Marshal(session.LivenessData)
	faceComparison, _ := json.Marshal(session.FaceComparison)

	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.NIK, session.Status, session.CurrentStep,
		ocrData, faceUrls, livenessData, faceComparison, session.ErrorMessage,
		session.ExpiresAt, session.CreatedAt, session.UpdatedAt,
	)
	return err
}

func (r *kycRepository) FindSessionByID(ctx context.Context, id string) (*domain.KYCSession, error) {
	query := `SELECT id, user_id, nik, status, current_step, ocr_data, face_urls, liveness_data, 
					 face_comparison, error_message, expires_at, created_at, updated_at 
			  FROM kyc_sessions WHERE id = $1`

	var session domain.KYCSession
	var ocrData, faceUrls, livenessData, faceComparison []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID, &session.UserID, &session.NIK, &session.Status, &session.CurrentStep,
		&ocrData, &faceUrls, &livenessData, &faceComparison, &session.ErrorMessage,
		&session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(ocrData, &session.OCRData)
	json.Unmarshal(faceUrls, &session.FaceUrls)
	json.Unmarshal(livenessData, &session.LivenessData)
	json.Unmarshal(faceComparison, &session.FaceComparison)

	return &session, nil
}

func (r *kycRepository) FindActiveSessionByUserID(ctx context.Context, userID string) (*domain.KYCSession, error) {
	query := `SELECT id, user_id, nik, status, current_step, ocr_data, face_urls, liveness_data, 
					 face_comparison, error_message, expires_at, created_at, updated_at 
			  FROM kyc_sessions 
			  WHERE user_id = $1 AND expires_at > NOW() 
			  ORDER BY created_at DESC LIMIT 1`

	var session domain.KYCSession
	var ocrData, faceUrls, livenessData, faceComparison []byte

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&session.ID, &session.UserID, &session.NIK, &session.Status, &session.CurrentStep,
		&ocrData, &faceUrls, &livenessData, &faceComparison, &session.ErrorMessage,
		&session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(ocrData, &session.OCRData)
	json.Unmarshal(faceUrls, &session.FaceUrls)
	json.Unmarshal(livenessData, &session.LivenessData)
	json.Unmarshal(faceComparison, &session.FaceComparison)

	return &session, nil
}

func (r *kycRepository) UpdateSession(ctx context.Context, session *domain.KYCSession) error {
	query := `
		UPDATE kyc_sessions 
		SET nik = $1, status = $2, current_step = $3, ocr_data = $4, face_urls = $5,
			liveness_data = $6, face_comparison = $7, error_message = $8, updated_at = $9
		WHERE id = $10
	`

	ocrData, _ := json.Marshal(session.OCRData)
	faceUrls, _ := json.Marshal(session.FaceUrls)
	livenessData, _ := json.Marshal(session.LivenessData)
	faceComparison, _ := json.Marshal(session.FaceComparison)

	_, err := r.db.ExecContext(ctx, query,
		session.NIK, session.Status, session.CurrentStep,
		ocrData, faceUrls, livenessData, faceComparison,
		session.ErrorMessage, session.UpdatedAt, session.ID,
	)
	return err
}

func (r *kycRepository) DeleteExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM kyc_sessions WHERE expires_at < NOW()`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *kycRepository) DeleteSessionByUserID(ctx context.Context, userID string) error {
	query := `DELETE FROM kyc_sessions WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// ========== Verification Methods ==========

func (r *kycRepository) CreateVerification(ctx context.Context, v *domain.KYCVerification) error {
	query := `
		INSERT INTO kyc_verifications (
			id, user_id, nik, full_name, place_of_birth, date_of_birth, gender, religion,
			address_street, address_rt, address_rw, address_sub_district, address_district, 
			address_city, address_province, administrative_code, ktp_url, face_url, 
			face_with_ktp_url, liveness_url, face_similarity, liveness_confidence, verified_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23
		)
	`

	adminCode, _ := json.Marshal(v.AdministrativeCode)

	_, err := r.db.ExecContext(ctx, query,
		v.ID, v.UserID, v.NIK, v.FullName, v.PlaceOfBirth, v.DateOfBirth, v.Gender, v.Religion,
		v.AddressStreet, v.AddressRT, v.AddressRW, v.AddressSubDistrict, v.AddressDistrict,
		v.AddressCity, v.AddressProvince, adminCode, v.KTPUrl, v.FaceUrl,
		v.FaceWithKTPUrl, v.LivenessUrl, v.FaceSimilarity, v.LivenessConfidence, v.VerifiedAt,
	)
	return err
}

func (r *kycRepository) FindVerificationByUserID(ctx context.Context, userID string) (*domain.KYCVerification, error) {
	query := `
		SELECT id, user_id, nik, full_name, place_of_birth, date_of_birth, gender, religion,
			   address_street, address_rt, address_rw, address_sub_district, address_district,
			   address_city, address_province, administrative_code, ktp_url, face_url,
			   face_with_ktp_url, liveness_url, face_similarity, liveness_confidence, 
			   verified_at, created_at
		FROM kyc_verifications
		WHERE user_id = $1
	`

	var v domain.KYCVerification
	var adminCode []byte

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&v.ID, &v.UserID, &v.NIK, &v.FullName, &v.PlaceOfBirth, &v.DateOfBirth, &v.Gender, &v.Religion,
		&v.AddressStreet, &v.AddressRT, &v.AddressRW, &v.AddressSubDistrict, &v.AddressDistrict,
		&v.AddressCity, &v.AddressProvince, &adminCode, &v.KTPUrl, &v.FaceUrl,
		&v.FaceWithKTPUrl, &v.LivenessUrl, &v.FaceSimilarity, &v.LivenessConfidence,
		&v.VerifiedAt, &v.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(adminCode, &v.AdministrativeCode)
	return &v, nil
}

func (r *kycRepository) IsNIKVerified(ctx context.Context, nik string) (bool, error) {
	query := `SELECT COUNT(*) FROM kyc_verifications WHERE nik = $1`
	var count int
	err := r.db.QueryRowContext(ctx, query, nik).Scan(&count)
	return count > 0, err
}
