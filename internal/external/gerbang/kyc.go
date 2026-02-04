package gerbang

import (
	"context"
	"mime/multipart"

	"github.com/GTDGit/PPOB_BE/internal/domain"
)

// ========== KYC API Methods ==========

// KTPOCRRequest represents KTP OCR request
type KTPOCRRequest struct {
	ImageURL string `json:"imageUrl"` // URL to KTP image (S3 URL)
}

// KTPOCRResponse represents KTP OCR response from Gerbang
type KTPOCRResponse struct {
	NIK                string                    `json:"nik"`
	FullName           string                    `json:"fullName"`
	PlaceOfBirth       string                    `json:"placeOfBirth"`
	DateOfBirth        string                    `json:"dateOfBirth"`
	Gender             string                    `json:"gender"`
	Religion           string                    `json:"religion"`
	Address            domain.KTPAddress         `json:"address"`
	AdministrativeCode domain.AdministrativeCode `json:"administrativeCode"`
}

// KTPOCR performs KTP OCR using Gerbang API
func (c *Client) KTPOCR(ctx context.Context, imageURL string) (*domain.KTPOCRResult, error) {
	req := KTPOCRRequest{
		ImageURL: imageURL,
	}

	resp, err := c.doRequestWithRetry(ctx, "POST", "/v1/verify/ktp-ocr", req)
	if err != nil {
		return nil, err
	}

	var ocrResp KTPOCRResponse
	if err := c.parseData(resp, &ocrResp); err != nil {
		return nil, err
	}

	// Convert to domain.KTPOCRResult
	result := &domain.KTPOCRResult{
		NIK:                ocrResp.NIK,
		FullName:           ocrResp.FullName,
		PlaceOfBirth:       ocrResp.PlaceOfBirth,
		DateOfBirth:        ocrResp.DateOfBirth,
		Gender:             ocrResp.Gender,
		Religion:           ocrResp.Religion,
		Address:            ocrResp.Address,
		AdministrativeCode: ocrResp.AdministrativeCode,
	}

	return result, nil
}

// CompareFacesRequest represents face comparison request
type CompareFacesRequest struct {
	KTPFaceURL    string `json:"ktpFaceUrl"`    // URL to face extracted from KTP
	SelfieFaceURL string `json:"selfieFaceUrl"` // URL to selfie face photo
}

// CompareFacesResponse represents face comparison response
type CompareFacesResponse struct {
	Matched    bool    `json:"matched"`
	Similarity float64 `json:"similarity"` // 0-100
	Threshold  float64 `json:"threshold"`  // Minimum threshold for match
}

// CompareFaces performs face comparison using Gerbang API
func (c *Client) CompareFaces(ctx context.Context, ktpFaceURL, selfieFaceURL string) (*domain.FaceCompareResult, error) {
	req := CompareFacesRequest{
		KTPFaceURL:    ktpFaceURL,
		SelfieFaceURL: selfieFaceURL,
	}

	resp, err := c.doRequestWithRetry(ctx, "POST", "/v1/verify/face-compare", req)
	if err != nil {
		return nil, err
	}

	var compareResp CompareFacesResponse
	if err := c.parseData(resp, &compareResp); err != nil {
		return nil, err
	}

	// Convert to domain.FaceCompareResult
	result := &domain.FaceCompareResult{
		Matched:    compareResp.Matched,
		Similarity: compareResp.Similarity,
		Threshold:  compareResp.Threshold,
	}

	return result, nil
}

// UploadImage uploads an image file to Gerbang and returns the URL
// This is a helper method for KYC flow
func (c *Client) UploadImage(ctx context.Context, file *multipart.FileHeader, imageType string) (string, error) {
	// TODO: Implement multipart file upload to Gerbang
	// For now, return placeholder
	// In production, this would:
	// 1. Create multipart/form-data request
	// 2. Upload file to Gerbang storage
	// 3. Return public URL
	return "", nil
}

// ========== LIVENESS API METHODS ==========

// LivenessSessionRequest untuk membuat session liveness
type LivenessSessionRequest struct {
	NIK string `json:"nik"`
}

// LivenessSessionResponse dari Gerbang API
type LivenessSessionResponse struct {
	SessionID string `json:"sessionId"`
	NIK       string `json:"nik"`
	ExpiresAt string `json:"expiresAt"`
}

// LivenessVerifyRequest untuk verifikasi liveness
type LivenessVerifyRequest struct {
	SessionID string `json:"sessionId"`
}

// LivenessVerifyResponse dari Gerbang API
type LivenessVerifyResponse struct {
	SessionID     string  `json:"sessionId"`
	NIK           string  `json:"nik"`
	IsLive        bool    `json:"isLive"`
	Confidence    float64 `json:"confidence"`
	FailureReason *string `json:"failureReason,omitempty"`
	ErrorCode     *string `json:"errorCode,omitempty"`
	File          *struct {
		Face string `json:"face"` // S3 URL dari Gerbang
	} `json:"file,omitempty"`
}

// LivenessSessionStatus dari GET session
type LivenessSessionStatus struct {
	SessionID  string  `json:"sessionId"`
	NIK        string  `json:"nik"`
	Status     string  `json:"status"` // Created, InProgress, Passed, Failed, Expired
	IsLive     bool    `json:"isLive"`
	Confidence float64 `json:"confidence"`
	File       *struct {
		Face string `json:"face"`
	} `json:"file,omitempty"`
	ExpiresAt   string  `json:"expiresAt"`
	CreatedAt   string  `json:"createdAt"`
	CompletedAt *string `json:"completedAt,omitempty"`
}

// CreateLivenessSession membuat session untuk liveness detection
// POST /v1/identity/liveness/session
func (c *Client) CreateLivenessSession(ctx context.Context, nik string) (*LivenessSessionResponse, error) {
	req := LivenessSessionRequest{NIK: nik}

	resp, err := c.doRequestWithRetry(ctx, "POST", "/v1/identity/liveness/session", req)
	if err != nil {
		return nil, err
	}

	var result LivenessSessionResponse
	if err := c.parseData(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// VerifyLiveness verifikasi hasil liveness setelah user complete di frontend
// POST /v1/identity/liveness/verify
func (c *Client) VerifyLiveness(ctx context.Context, sessionID string) (*LivenessVerifyResponse, error) {
	req := LivenessVerifyRequest{SessionID: sessionID}

	resp, err := c.doRequestWithRetry(ctx, "POST", "/v1/identity/liveness/verify", req)
	if err != nil {
		return nil, err
	}

	var result LivenessVerifyResponse
	if err := c.parseData(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetLivenessSession mendapatkan status session liveness
// GET /v1/identity/liveness/session/{sessionId}
func (c *Client) GetLivenessSession(ctx context.Context, sessionID string) (*LivenessSessionStatus, error) {
	resp, err := c.doRequestWithRetry(ctx, "GET", "/v1/identity/liveness/session/"+sessionID, nil)
	if err != nil {
		return nil, err
	}

	var result LivenessSessionStatus
	if err := c.parseData(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
