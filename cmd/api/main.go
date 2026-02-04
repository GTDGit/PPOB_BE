package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/config"
	"github.com/GTDGit/PPOB_BE/internal/external/gerbang"
	"github.com/GTDGit/PPOB_BE/internal/external/s3"
	"github.com/GTDGit/PPOB_BE/internal/handler"
	"github.com/GTDGit/PPOB_BE/internal/job"
	"github.com/GTDGit/PPOB_BE/internal/middleware"
	"github.com/GTDGit/PPOB_BE/internal/repository"
	"github.com/GTDGit/PPOB_BE/internal/service"
	"github.com/GTDGit/PPOB_BE/pkg/database"
	"github.com/GTDGit/PPOB_BE/pkg/redis"
	"github.com/gin-gonic/gin"
)

func main() {
	// Setup logger early for migrations
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup database
	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	migrator := database.NewMigrator(db, logger)
	if err := migrator.RunMigrations("./migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	logger.Info("database migrations completed")

	// Setup Redis
	redisClient, err := redis.NewClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	rdb := redisClient // Keep backward compatibility for existing code

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	balanceRepo := repository.NewBalanceRepository(db)
	settingsRepo := repository.NewUserSettingsRepository(db)
	prepaidRepo := repository.NewPrepaidRepository(db)
	postpaidRepo := repository.NewPostpaidRepository(db)
	transferRepo := repository.NewTransferRepository(db)
	productRepo := repository.NewProductRepository(db)
	voucherRepo := repository.NewVoucherRepository(db)
	contactRepo := repository.NewContactRepository(db)
	homeRepo := repository.NewHomeRepository() // No DB needed - in-memory data
	historyRepo := repository.NewHistoryRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	depositRepo := repository.NewDepositRepository(db)
	territoryRepo := repository.NewTerritoryRepository(db)
	kycRepo := repository.NewKYCRepository(db)

	// Initialize external clients
	gerbangClient := gerbang.NewClient(gerbang.Config{
		BaseURL:      cfg.Gerbang.BaseURL,
		ClientID:     cfg.Gerbang.ClientID,
		ClientSecret: cfg.Gerbang.ClientSecret,
		Timeout:      cfg.Gerbang.Timeout,
	})

	// Initialize S3 client for KYC files (KTP + face photos)
	s3Client, err := s3.NewClient(s3.Config{
		Region:          cfg.S3.Region,
		Bucket:          cfg.S3.Bucket,
		AccessKeyID:     cfg.S3.AccessKey,
		SecretAccessKey: cfg.S3.SecretKey,
		PublicURL:       cfg.S3.BaseURL,
	})
	if err != nil {
		log.Fatalf("Failed to initialize S3 client: %v", err)
	}

	// Initialize services
	otpService := service.NewOTPService(rdb, cfg.OTP, cfg.WhatsApp, cfg.Fazpass)
	emailService := service.NewEmailService(cfg.Brevo)
	authService := service.NewAuthService(
		userRepo,
		deviceRepo,
		sessionRepo,
		balanceRepo,
		settingsRepo,
		otpService,
		emailService,
		rdb,
		cfg.JWT,
	)
	prepaidService := service.NewPrepaidService(
		prepaidRepo,
		balanceRepo,
		userRepo,
		gerbangClient,
	)
	postpaidService := service.NewPostpaidService(
		postpaidRepo,
		balanceRepo,
		voucherRepo,
		userRepo,
		gerbangClient,
	)
	transferService := service.NewTransferService(
		transferRepo,
		balanceRepo,
		userRepo,
		productRepo,
		gerbangClient,
	)
	productService := service.NewProductService(productRepo, redisClient)
	voucherService := service.NewVoucherService(voucherRepo)
	contactService := service.NewContactService(contactRepo, productRepo)
	homeService := service.NewHomeService(homeRepo, userRepo, balanceRepo)
	userService := service.NewUserService(userRepo, balanceRepo, settingsRepo)
	historyService := service.NewHistoryService(historyRepo)
	notificationService := service.NewNotificationService(notificationRepo)
	depositService := service.NewDepositService(depositRepo, balanceRepo, userRepo, gerbangClient)
	territoryService := service.NewTerritoryService(territoryRepo)
	kycService := service.NewKYCService(kycRepo, userRepo, gerbangClient, s3Client)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	prepaidHandler := handler.NewPrepaidHandler(prepaidService)
	postpaidHandler := handler.NewPostpaidHandler(postpaidService)
	transferHandler := handler.NewTransferHandler(transferService)
	productHandler := handler.NewProductHandler(productService)
	voucherHandler := handler.NewVoucherHandler(voucherService)
	contactHandler := handler.NewContactHandler(contactService)
	homeHandler := handler.NewHomeHandler(homeService)
	userHandler := handler.NewUserHandler(userService)
	historyHandler := handler.NewHistoryHandler(historyService, depositService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	territoryHandler := handler.NewTerritoryHandler(territoryService)
	kycHandler := handler.NewKYCHandler(kycService)
	depositHandler := handler.NewDepositHandler(depositService, cfg.Gerbang.CallbackSecret)

	// Setup Gin
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize product sync job
	productSyncJob := job.NewProductSyncJob(
		gerbangClient,
		productRepo,
		redisClient,
		logger,
		cfg.ProductSync.Interval,
	)

	// Start product sync job in background (if enabled)
	if cfg.ProductSync.Enabled {
		go productSyncJob.Start(context.Background())
		logger.Info("product sync job enabled", "interval", cfg.ProductSync.Interval.String())
	} else {
		logger.Info("product sync job disabled")
	}

	// Initialize bank sync job
	bankSyncJob := job.NewBankSyncJob(
		gerbangClient,
		productRepo,
		logger,
		cfg.BankCodeSync.Interval,
		cfg.BankCodeSync.EnableOnStart,
	)

	// Start bank sync job in background (if enabled)
	if cfg.BankCodeSync.Enabled {
		go bankSyncJob.Start(context.Background())
		logger.Info("bank sync job enabled", "interval", cfg.BankCodeSync.Interval.String())
	} else {
		logger.Info("bank sync job disabled")
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS(middleware.DefaultCORSConfig()))
	router.Use(middleware.RequestID())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Home routes (protected)
		v1.GET("/home", middleware.JWTAuth(cfg.JWT.Secret), homeHandler.GetHome)

		// Services routes (protected)
		v1.GET("/services", middleware.JWTAuth(cfg.JWT.Secret), homeHandler.GetServices)

		// Banners routes (protected)
		v1.GET("/banners", middleware.JWTAuth(cfg.JWT.Secret), homeHandler.GetBanners)

		// User routes (protected)
		user := v1.Group("/user")
		user.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			user.GET("/balance", homeHandler.GetBalance)
			user.GET("/profile", userHandler.GetProfile)
			user.PUT("/profile", userHandler.UpdateProfile)
			user.POST("/avatar", userHandler.UploadAvatar)
			user.DELETE("/avatar", userHandler.DeleteAvatar)
			user.GET("/settings", userHandler.GetSettings)
			user.PUT("/settings", userHandler.UpdateSettings)
			user.GET("/referral", userHandler.GetReferralInfo)
			user.GET("/referral/history", userHandler.GetReferralHistory)
		}

		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/start", authHandler.StartAuth)
			auth.POST("/verify-otp", authHandler.VerifyOTP)
			auth.POST("/resend-otp", authHandler.ResendOTP)
			auth.POST("/pin-login", authHandler.PINLogin)
			auth.POST("/refresh-token", authHandler.RefreshToken)

			// Protected with temp token
			auth.POST("/complete-profile", middleware.TempTokenAuth(rdb), authHandler.CompleteProfile)
			auth.POST("/set-pin", middleware.TempTokenAuth(rdb), authHandler.SetPIN)

			// Protected with access token
			protected := auth.Group("")
			protected.Use(middleware.JWTAuth(cfg.JWT.Secret))
			{
				protected.POST("/verify-pin-only", authHandler.VerifyPINOnly)
				protected.POST("/logout", authHandler.Logout)
				protected.GET("/devices", authHandler.ListDevices)
				protected.DELETE("/devices/:deviceId", authHandler.RemoveDevice)

				// Change PIN
				protected.POST("/change-pin/verify-current", authHandler.ChangePINVerifyCurrent)
			}

			// Change PIN confirm (temp token)
			auth.POST("/change-pin/confirm", middleware.TempTokenAuth(rdb), authHandler.ChangePINConfirm)

			// Change Phone (access token)
			changePhone := auth.Group("/change-phone")
			changePhone.Use(middleware.JWTAuth(cfg.JWT.Secret))
			{
				changePhone.POST("/verify-old/request-otp", authHandler.ChangePhoneRequestOTPOld)
				changePhone.POST("/verify-old", authHandler.ChangePhoneVerifyOld)
			}

			// Change Phone new (temp token)
			auth.POST("/change-phone/new/request-otp", middleware.TempTokenAuth(rdb), authHandler.ChangePhoneRequestOTPNew)
			auth.POST("/change-phone/new/verify-otp", middleware.TempTokenAuth(rdb), authHandler.ChangePhoneVerifyNew)

			// Email verification (protected)
			email := auth.Group("/email")
			email.Use(middleware.JWTAuth(cfg.JWT.Secret))
			{
				email.POST("/request-verification", authHandler.RequestEmailVerification)
			}

			// Email verification (public - for verification link)
			auth.GET("/email/verify", authHandler.VerifyEmail)
		}

		// Prepaid routes (protected)
		prepaid := v1.Group("/prepaid")
		prepaid.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			prepaid.POST("/inquiry", prepaidHandler.Inquiry)
			prepaid.POST("/order", prepaidHandler.CreateOrder)
			prepaid.POST("/pay", prepaidHandler.Pay)
		}

		// Postpaid routes (protected)
		postpaid := v1.Group("/postpaid")
		postpaid.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			postpaid.POST("/inquiry", postpaidHandler.Inquiry)
			postpaid.POST("/pay", postpaidHandler.Pay)
		}

		// Transfer routes (protected)
		transfer := v1.Group("/transfer")
		transfer.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			transfer.POST("/inquiry", transferHandler.Inquiry)
			transfer.POST("/execute", transferHandler.Execute)
		}

		// Products routes (PUBLIC - no auth)
		products := v1.Group("/products")
		{
			products.GET("/operators", productHandler.GetOperators)
			products.GET("/ewallet/providers", productHandler.GetEwalletProviders)
			products.GET("/pdam/regions", productHandler.GetPDAMRegions)
			products.GET("/banks", productHandler.GetBanks)
			products.GET("/tv/providers", productHandler.GetTVProviders)
		}

		// Vouchers routes (protected)
		vouchers := v1.Group("/vouchers")
		vouchers.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			vouchers.GET("", voucherHandler.List)
			vouchers.GET("/applicable", voucherHandler.GetApplicable)
			vouchers.POST("/validate", voucherHandler.Validate)
		}

		// Contacts routes (protected)
		contacts := v1.Group("/contacts")
		contacts.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			contacts.GET("", contactHandler.List)
			contacts.POST("", contactHandler.Create)
			contacts.PUT("/:contactId", contactHandler.Update)
			contacts.DELETE("/:contactId", contactHandler.Delete)
		}

		// Territory routes (public, no auth - static reference data)
		territory := v1.Group("/territory")
		{
			territory.GET("/provinces", territoryHandler.GetProvinces)
			territory.GET("/cities/:provinceCode", territoryHandler.GetCities)
			territory.GET("/districts/:cityCode", territoryHandler.GetDistricts)
			territory.GET("/sub-districts/:districtCode", territoryHandler.GetSubDistricts)
			territory.GET("/postal-codes/:subDistrictCode", territoryHandler.GetPostalCodes)
			territory.GET("/search/postal-code/:postalCode", territoryHandler.SearchByPostalCode)
		}

		// History routes (protected)
		history := v1.Group("/history")
		history.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			history.GET("/transactions", historyHandler.List)
			history.GET("/transactions/:transactionId", historyHandler.GetDetail)
			history.GET("/deposits", historyHandler.ListDeposits)
			history.GET("/qris", historyHandler.ListQRISIncome)
			history.GET("/transactions/:transactionId/receipt", historyHandler.GetReceipt)
			history.GET("/transactions/:transactionId/receipt/download", historyHandler.DownloadReceipt)
			history.GET("/transactions/:transactionId/receipt/share", historyHandler.ShareReceipt)
			history.PUT("/transactions/:transactionId/selling-price", historyHandler.UpdateSellingPrice)
		}

		// Notification routes (protected)
		notifications := v1.Group("/notifications")
		notifications.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			notifications.GET("", notificationHandler.List)
			notifications.GET("/unread-count", notificationHandler.GetUnreadCount)
			notifications.GET("/:id", notificationHandler.GetDetail)
			notifications.PUT("/:id/read", notificationHandler.MarkAsRead)
			notifications.PUT("/read-all", notificationHandler.MarkAllAsRead)
			notifications.DELETE("/:id", notificationHandler.Delete)
		}

		// KYC routes (protected)
		kyc := v1.Group("/kyc")
		kyc.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			kyc.GET("/status", kycHandler.GetStatus)
			kyc.GET("/session", kycHandler.GetSession)
			kyc.POST("/start", kycHandler.StartVerification)
			kyc.POST("/cancel", kycHandler.CancelVerification)
			kyc.POST("/ktp", kycHandler.UploadKTP)
			kyc.POST("/face", kycHandler.UploadFacePhotos)
			kyc.POST("/liveness/session", kycHandler.CreateLivenessSession) // Create session for frontend SDK
			kyc.POST("/liveness/verify", kycHandler.VerifyLiveness)         // Verify after frontend done
			kyc.POST("/submit", kycHandler.Submit)
		}

		// Deposit routes (protected)
		deposit := v1.Group("/deposit")
		deposit.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			deposit.GET("/methods", depositHandler.GetMethods)
			deposit.POST("/bank-transfer", depositHandler.CreateBankTransfer)
			deposit.POST("/qris", depositHandler.CreateQRIS)
			deposit.GET("/retail/providers", depositHandler.GetRetailProviders)
			deposit.POST("/retail", depositHandler.CreateRetail)
			deposit.GET("/va/banks", depositHandler.GetVABanks)
			deposit.POST("/va", depositHandler.CreateVA)
			deposit.GET("/history", depositHandler.GetHistory)
			deposit.GET("/:depositId", depositHandler.GetStatus)
		}
	}

	// Internal routes (no JWT, signature verification in handler)
	internal := router.Group("/internal")
	// Add rate limiting for webhook protection (100 requests/minute)
	internal.Use(middleware.RateLimit(rdb, middleware.RateLimitConfig{
		Limit:  100,
		Window: time.Minute,
	}))
	{
		internal.POST("/webhook/deposit", depositHandler.HandleWebhook)
		internal.POST("/webhook/transfer", transferHandler.HandleWebhook)
		internal.POST("/webhook/prepaid", prepaidHandler.HandleWebhook)
		internal.POST("/webhook/postpaid", postpaidHandler.HandleWebhook)
	}

	// Create server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("Server starting on port %d...", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
