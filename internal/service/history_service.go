package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/repository"
)

// HistoryService handles transaction history business logic
type HistoryService struct {
	historyRepo repository.HistoryRepository
}

// NewHistoryService creates a new history service
func NewHistoryService(historyRepo repository.HistoryRepository) *HistoryService {
	return &HistoryService{
		historyRepo: historyRepo,
	}
}

// List returns transaction history list with pagination
func (s *HistoryService) List(ctx context.Context, userID string, filter repository.TransactionFilter) (*domain.TransactionListResponse, error) {
	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}
	if filter.PerPage > 50 {
		filter.PerPage = 50
	}

	// Get transactions
	transactions, total, err := s.historyRepo.FindByUserID(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	// Convert to summary
	summaries := make([]*domain.TransactionSummary, 0, len(transactions))
	for _, tx := range transactions {
		summary := s.toTransactionSummary(tx)
		summaries = append(summaries, summary)
	}

	// Build pagination
	totalPages := int(math.Ceil(float64(total) / float64(filter.PerPage)))
	pagination := &domain.Pagination{
		CurrentPage:  filter.Page,
		TotalPages:   totalPages,
		TotalItems:   total,
		ItemsPerPage: filter.PerPage,
		HasNextPage:  filter.Page < totalPages,
		HasPrevPage:  filter.Page > 1,
	}

	return &domain.TransactionListResponse{
		Transactions: summaries,
		Pagination:   pagination,
	}, nil
}

// GetDetail returns detailed transaction information
func (s *HistoryService) GetDetail(ctx context.Context, userID, transactionID string) (*domain.TransactionDetailResponse, error) {
	// Get transaction with ownership validation
	tx, err := s.historyRepo.FindByUserAndID(ctx, userID, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	if tx == nil {
		return nil, domain.ErrValidationFailed("Transaction not found")
	}

	// Build detailed response based on transaction type
	return s.buildTransactionDetail(tx), nil
}

// Helper functions

func (s *HistoryService) toTransactionSummary(tx *domain.Transaction) *domain.TransactionSummary {
	title, description, icon := s.getTransactionDisplay(tx)
	statusLabel := s.getStatusLabel(tx.Status)

	var completedAt *string
	if tx.CompletedAt != nil {
		formatted := tx.CompletedAt.Format(time.RFC3339)
		completedAt = &formatted
	}

	serviceType := tx.ServiceType
	var serviceTypePtr *string
	if serviceType != "" {
		serviceTypePtr = &serviceType
	}

	return &domain.TransactionSummary{
		ID:                   tx.ID,
		Type:                 tx.Type,
		ServiceType:          serviceTypePtr,
		Title:                title,
		Description:          description,
		Amount:               tx.Amount,
		AmountFormatted:      formatHomeCurrency(tx.Amount),
		AdminFee:             tx.AdminFee,
		TotalAmount:          tx.TotalPayment,
		TotalAmountFormatted: formatHomeCurrency(tx.TotalPayment),
		Status:               tx.Status,
		StatusLabel:          statusLabel,
		CreatedAt:            tx.CreatedAt.Format(time.RFC3339),
		CompletedAt:          completedAt,
		Icon:                 icon,
		IconURL:              fmt.Sprintf("https://cdn.ppob.id/icons/%s.png", icon),
	}
}

func (s *HistoryService) buildTransactionDetail(tx *domain.Transaction) *domain.TransactionDetailResponse {
	title, _, _ := s.getTransactionDisplay(tx)
	statusLabel := s.getStatusLabel(tx.Status)

	serviceType := tx.ServiceType
	var serviceTypePtr *string
	if serviceType != "" {
		serviceTypePtr = &serviceType
	}

	var completedAt *string
	if tx.CompletedAt != nil {
		formatted := tx.CompletedAt.Format(time.RFC3339)
		completedAt = &formatted
	}

	response := &domain.TransactionDetailResponse{
		Transaction: &domain.HistoryTransactionInfo{
			ID:          tx.ID,
			Type:        tx.Type,
			ServiceType: serviceTypePtr,
			Title:       title,
			Status:      tx.Status,
			StatusLabel: statusLabel,
			CreatedAt:   tx.CreatedAt.Format(time.RFC3339),
			CompletedAt: completedAt,
		},
		Pricing: &domain.HistoryPricingInfo{
			AdminFee:              tx.AdminFee,
			AdminFeeFormatted:     formatHomeCurrency(tx.AdminFee),
			TotalPayment:          tx.TotalPayment,
			TotalPaymentFormatted: formatHomeCurrency(tx.TotalPayment),
		},
		Payment: &domain.HistoryPaymentInfo{
			Method:      "balance",
			MethodLabel: "Saldo PPOB.ID",
		},
		Actions: &domain.TransactionActions{
			CanShare:           tx.Status == domain.TransactionStatusSuccess,
			CanDownloadReceipt: tx.Status == domain.TransactionStatusSuccess,
			CanRepeat:          true,
		},
	}

	// Add type-specific details
	switch tx.Type {
	case domain.TransactionTypePrepaid:
		s.addPrepaidDetails(response, tx)
	case domain.TransactionTypePostpaid:
		s.addPostpaidDetails(response, tx)
	case domain.TransactionTypeTransfer:
		s.addTransferDetails(response, tx)
	}

	// Add discount if any
	if tx.Discount > 0 {
		discountFormatted := formatHomeCurrency(tx.Discount)
		response.Pricing.VoucherDiscount = &tx.Discount
		response.Pricing.VoucherDiscountFormatted = &discountFormatted
	}

	// Add receipt info if available
	if tx.SerialNumber != nil || tx.Token != nil {
		response.Receipt = &domain.HistoryReceiptInfo{
			SerialNumber: tx.SerialNumber,
			Token:        tx.Token,
		}
		if tx.ServiceType == "pln_prepaid" && tx.Token != nil {
			kwh := "35.5 kWh" // Mock KWH
			response.Receipt.KWH = &kwh
			response.Actions.CanCopyToken = true
		}
	}

	return response
}

func (s *HistoryService) addPrepaidDetails(response *domain.TransactionDetailResponse, tx *domain.Transaction) {
	productID := "prd_" + tx.ServiceType
	if tx.ProductID != nil {
		productID = *tx.ProductID
	}

	productPriceFormatted := formatHomeCurrency(tx.Amount)
	response.Pricing.ProductPrice = &tx.Amount
	response.Pricing.ProductPriceFormatted = &productPriceFormatted

	response.Product = &domain.HistoryProductInfo{
		ID:      productID,
		Name:    tx.ProductName,
		Nominal: tx.Amount,
	}

	// Add target info
	if tx.ServiceType == "pln_prepaid" {
		customerName := "BUDI SANTOSO"
		segmentPower := "R1/1300VA"
		address := "JL. MERDEKA NO 123"
		response.Target = &domain.TargetInfo{
			CustomerID:   &tx.Target,
			CustomerName: &customerName,
			SegmentPower: &segmentPower,
			Address:      &address,
		}
	} else {
		// Phone number
		name := "John Dhoe"
		response.Target = &domain.TargetInfo{
			Number: &tx.Target,
			Name:   &name,
		}
		if tx.ServiceType == "pulsa" || tx.ServiceType == "data" {
			response.Target.Operator = &domain.OperatorInfo{
				ID:   "telkomsel",
				Name: "Telkomsel",
			}
		}
	}
}

func (s *HistoryService) addPostpaidDetails(response *domain.TransactionDetailResponse, tx *domain.Transaction) {
	productPriceFormatted := formatHomeCurrency(tx.Amount)
	response.Pricing.ProductPrice = &tx.Amount
	response.Pricing.ProductPriceFormatted = &productPriceFormatted

	response.Product = &domain.HistoryProductInfo{
		ID:      "prd_" + tx.ServiceType,
		Name:    tx.ProductName,
		Nominal: tx.Amount,
	}

	// Add target info
	customerName := "BUDI SANTOSO"
	response.Target = &domain.TargetInfo{
		CustomerID:   &tx.Target,
		CustomerName: &customerName,
	}
}

func (s *HistoryService) addTransferDetails(response *domain.TransactionDetailResponse, tx *domain.Transaction) {
	transferAmountFormatted := formatHomeCurrency(tx.Amount)
	response.Pricing.TransferAmount = &tx.Amount
	response.Pricing.TransferAmountFormatted = &transferAmountFormatted

	response.Destination = &domain.DestinationInfo{
		BankCode:      "014",
		BankName:      "Bank Central Asia",
		BankShortName: "BCA",
		BankIcon:      "https://cdn.ppob.id/banks/bca.png",
		AccountNumber: tx.Target,
		AccountName:   "BUDI SANTOSO",
	}

	response.Transfer = &domain.TransferInfo{
		Amount:          tx.Amount,
		AmountFormatted: formatHomeCurrency(tx.Amount),
	}
}

func (s *HistoryService) getTransactionDisplay(tx *domain.Transaction) (title, description, icon string) {
	switch tx.Type {
	case domain.TransactionTypePrepaid:
		switch tx.ServiceType {
		case "pulsa":
			title = "Pembelian Pulsa"
			description = fmt.Sprintf("Pulsa Telkomsel %s", maskNumber(tx.Target))
			icon = "pulsa"
		case "data":
			title = "Pembelian Paket Data"
			description = fmt.Sprintf("Paket Data XL %s", maskNumber(tx.Target))
			icon = "paket_data"
		case "pln_prepaid":
			title = "Pembelian Token PLN"
			description = fmt.Sprintf("Token PLN %s", maskNumber(tx.Target))
			icon = "token_pln"
		case "ewallet":
			title = "Top Up E-Wallet"
			description = fmt.Sprintf("Top Up %s", maskNumber(tx.Target))
			icon = "ewallet"
		default:
			title = tx.ProductName
			description = fmt.Sprintf("%s %s", tx.ProductName, maskNumber(tx.Target))
			icon = tx.ServiceType
		}
	case domain.TransactionTypePostpaid:
		switch tx.ServiceType {
		case "pln_postpaid":
			title = "Pembayaran Tagihan PLN"
			description = fmt.Sprintf("Tagihan PLN %s", maskNumber(tx.Target))
			icon = "tagihan_pln"
		case "pdam":
			title = "Pembayaran PDAM"
			description = fmt.Sprintf("Tagihan PDAM %s", maskNumber(tx.Target))
			icon = "pdam"
		default:
			title = tx.ProductName
			description = fmt.Sprintf("%s %s", tx.ProductName, maskNumber(tx.Target))
			icon = tx.ServiceType
		}
	case domain.TransactionTypeTransfer:
		title = "Transfer Bank BCA"
		description = fmt.Sprintf("Transfer ke bank BCA %s", maskNumber(tx.Target))
		icon = "transfer_bank"
	default:
		title = tx.ProductName
		description = tx.Target
		icon = "default"
	}
	return
}

func (s *HistoryService) getStatusLabel(status string) string {
	labels := map[string]string{
		domain.TransactionStatusPending:    "Menunggu",
		domain.TransactionStatusProcessing: "Diproses",
		domain.TransactionStatusSuccess:    "Berhasil",
		domain.TransactionStatusFailed:     "Gagal",
		domain.TransactionStatusCancelled:  "Dibatalkan",
		domain.TransactionStatusRefunded:   "Dikembalikan",
		domain.TransactionStatusExpired:    "Expired",
	}
	if label, ok := labels[status]; ok {
		return label
	}
	return status
}

func maskNumber(number string) string {
	if len(number) <= 4 {
		return number
	}
	// Mask middle digits, keep first 4 and last 4
	visibleStart := 4
	visibleEnd := 4
	if len(number) < 8 {
		visibleStart = 2
		visibleEnd = 2
	}
	start := number[:visibleStart]
	end := number[len(number)-visibleEnd:]
	return fmt.Sprintf("%s****%s", start, end)
}

// GetReceipt returns receipt data for a transaction
func (s *HistoryService) GetReceipt(ctx context.Context, userID, transactionID string) (*domain.ReceiptResponse, error) {
	// Get transaction with ownership validation
	tx, err := s.historyRepo.FindByUserAndID(ctx, userID, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	if tx == nil {
		return nil, domain.ErrNotFound("Transaksi")
	}

	// Only success transactions can generate receipts
	if tx.Status != domain.TransactionStatusSuccess {
		return nil, domain.ErrValidationFailed("Struk hanya tersedia untuk transaksi yang berhasil")
	}

	// Build receipt response
	return &domain.ReceiptResponse{
		TransactionID: tx.ID,
		Title:         "Struk Transaksi",
		Subtitle:      "Transaksi Berhasil",
		Amount:        formatHomeCurrency(tx.TotalPayment),
		Date:          tx.CreatedAt.Format("02 Jan 2006, 15:04"),
		ReceiptURL:    fmt.Sprintf("https://receipt.ppob.id/%s", tx.ID),
	}, nil
}

// GenerateReceipt generates downloadable receipt (image or PDF)
func (s *HistoryService) GenerateReceipt(ctx context.Context, userID, transactionID, format string) ([]byte, string, error) {
	// Get transaction with ownership validation
	tx, err := s.historyRepo.FindByUserAndID(ctx, userID, transactionID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get transaction: %w", err)
	}
	if tx == nil {
		return nil, "", domain.ErrNotFound("Transaksi")
	}

	// Only success transactions can generate receipts
	if tx.Status != domain.TransactionStatusSuccess {
		return nil, "", domain.ErrValidationFailed("Struk hanya tersedia untuk transaksi yang berhasil")
	}

	// TODO: Implement actual receipt generation
	// For now, return mock data
	mockData := []byte("Receipt data for transaction " + transactionID)
	contentType := "image/png"
	if format == "pdf" {
		contentType = "application/pdf"
	}

	return mockData, contentType, nil
}

// GetShareData returns shareable receipt data
func (s *HistoryService) GetShareData(ctx context.Context, userID, transactionID string) (*domain.ShareReceiptResponse, error) {
	// Get transaction with ownership validation
	tx, err := s.historyRepo.FindByUserAndID(ctx, userID, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	if tx == nil {
		return nil, domain.ErrNotFound("Transaksi")
	}

	// Only success transactions can be shared
	if tx.Status != domain.TransactionStatusSuccess {
		return nil, domain.ErrValidationFailed("Hanya transaksi yang berhasil yang dapat dibagikan")
	}

	// Build share data
	title, _, _ := s.getTransactionDisplay(tx)
	text := fmt.Sprintf("✅ %s berhasil!\n💰 Total: %s\n🕐 %s",
		title,
		formatHomeCurrency(tx.TotalPayment),
		tx.CreatedAt.Format("02 Jan 2006, 15:04"),
	)

	return &domain.ShareReceiptResponse{
		ShareText: text,
		ShareURL:  fmt.Sprintf("https://receipt.ppob.id/%s", tx.ID),
		ImageURL:  fmt.Sprintf("https://receipt.ppob.id/%s.png", tx.ID),
		DeepLink:  fmt.Sprintf("ppobid://receipt/%s", tx.ID),
		CanShare:  true,
		CanCopy:   true,
		CanPrint:  true,
		CanEmail:  true,
	}, nil
}

// UpdateSellingPrice updates custom selling price for a transaction (for merchants)
func (s *HistoryService) UpdateSellingPrice(ctx context.Context, userID, transactionID string, sellingPrice int64) error {
	// Get transaction with ownership validation
	tx, err := s.historyRepo.FindByUserAndID(ctx, userID, transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}
	if tx == nil {
		return domain.ErrNotFound("Transaksi")
	}

	// Validate selling price
	if sellingPrice < tx.TotalPayment {
		return domain.ErrValidationFailed("Harga jual tidak boleh lebih kecil dari harga modal")
	}

	// TODO: Implement actual database update when custom pricing feature is ready
	// For now, just validate and return success

	return nil
}
