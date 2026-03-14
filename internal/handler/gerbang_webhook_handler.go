package handler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/external/gerbang"
	"github.com/GTDGit/PPOB_BE/internal/service"
)

// GerbangWebhookHandler handles all GTD callbacks via a single client callback URL.
type GerbangWebhookHandler struct {
	prepaidService *service.PrepaidService
	postpaidService *service.PostpaidService
	transferService *service.TransferService
	depositService *service.DepositService
	callbackSecret string
}

// NewGerbangWebhookHandler creates a generic Gerbang webhook handler.
func NewGerbangWebhookHandler(
	prepaidService *service.PrepaidService,
	postpaidService *service.PostpaidService,
	transferService *service.TransferService,
	depositService *service.DepositService,
	callbackSecret string,
) *GerbangWebhookHandler {
	return &GerbangWebhookHandler{
		prepaidService: prepaidService,
		postpaidService: postpaidService,
		transferService: transferService,
		depositService: depositService,
		callbackSecret: callbackSecret,
	}
}

// HandleWebhook accepts GTD callbacks for deposit, prepaid, postpaid, and transfer flows.
func (h *GerbangWebhookHandler) HandleWebhook(c *gin.Context) {
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		slog.Error("failed to read gerbang webhook body", slog.String("error", err.Error()))
		respondWithError(c, domain.ErrInvalidRequestError)
		return
	}

	signature := getGerbangSignature(c)
	if signature == "" {
		slog.Error("gerbang webhook signature missing")
		respondWithError(c, domain.NewError(domain.CodeUnauthorized, "Signature tidak valid", http.StatusUnauthorized))
		return
	}

	if h.callbackSecret != "" && !gerbang.VerifySignature(rawBody, signature, h.callbackSecret) {
		slog.Error("gerbang webhook signature mismatch")
		respondWithError(c, domain.NewError(domain.CodeUnauthorized, "Signature tidak valid", http.StatusUnauthorized))
		return
	}

	webhook, err := gerbang.ParseWebhook(rawBody)
	if err != nil {
		slog.Error("failed to parse gerbang webhook", slog.String("error", err.Error()))
		respondWithError(c, domain.ErrInvalidRequestError)
		return
	}

	if err := h.dispatch(c, webhook); err != nil {
		slog.Error("failed to process gerbang webhook",
			slog.String("event", webhook.Event),
			slog.String("error", err.Error()),
		)
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Webhook received",
		"event":   webhook.Event,
	})
}

func (h *GerbangWebhookHandler) dispatch(c *gin.Context, webhook *gerbang.WebhookPayload) error {
	switch {
	case webhook.IsPaymentEvent():
		paymentData, err := webhook.ParsePaymentData()
		if err != nil {
			return domain.ErrInvalidRequestError
		}

		slog.Info("gerbang payment webhook received",
			slog.String("event", webhook.Event),
			slog.String("reference_id", paymentData.ReferenceID),
			slog.String("status", paymentData.Status),
		)

		return h.depositService.HandleWebhook(c.Request.Context(), paymentData.ReferenceID)

	case webhook.IsTransferEvent():
		transferData, err := webhook.ParseTransferData()
		if err != nil {
			return domain.ErrInvalidRequestError
		}

		slog.Info("gerbang transfer webhook received",
			slog.String("event", webhook.Event),
			slog.String("reference_id", transferData.ReferenceID),
			slog.String("status", transferData.Status),
		)

		return h.transferService.HandleWebhook(c.Request.Context(), transferData)

	case webhook.IsTransactionEvent():
		transactionData, err := webhook.ParseTransactionData()
		if err != nil {
			return domain.ErrInvalidRequestError
		}

		slog.Info("gerbang transaction webhook received",
			slog.String("event", webhook.Event),
			slog.String("reference_id", transactionData.ReferenceID),
			slog.String("status", transactionData.Status),
		)

		return h.dispatchTransactionWebhook(c, transactionData)

	default:
		return domain.ErrValidationFailed("Event webhook tidak dikenali")
	}
}

func (h *GerbangWebhookHandler) dispatchTransactionWebhook(c *gin.Context, data *gerbang.TransactionWebhookData) error {
	referenceID := data.ReferenceID
	switch {
	case strings.HasPrefix(referenceID, "ord_"):
		return h.prepaidService.HandleWebhook(c.Request.Context(), data)
	case strings.HasPrefix(referenceID, "trx_post_"):
		return h.postpaidService.HandleWebhook(c.Request.Context(), data)
	default:
		if err := h.prepaidService.HandleWebhook(c.Request.Context(), data); err == nil {
			return nil
		} else if !isNotFoundError(err) {
			return err
		}

		return h.postpaidService.HandleWebhook(c.Request.Context(), data)
	}
}

func isNotFoundError(err error) bool {
	var appErr *domain.AppError
	if !errors.As(err, &appErr) {
		return false
	}
	return appErr.HTTPStatus == http.StatusNotFound
}
