package main

import (
	"final-project/config"
	"final-project/controller"
	_ "final-project/docs"
	"final-project/middleware"
	"final-project/repository"
	"final-project/service"
	"final-project/utils/helpers"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func setupRoutes(cfg *config.Config, db *gorm.DB) *gin.Engine {
	if cfg.IsProd {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// JWT Konfigurasi
	jwtHelper := helpers.NewJWTHelper(cfg.JWTSecret, cfg.AccessTokenExp, cfg.RefreshTokenExp, cfg.Issuer)

	// User token
	userTokenRepo := repository.NewUserTokenRepository(db)
	userTokenSvc := service.NewTokenService(userTokenRepo, *jwtHelper)

	// Users
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo, userTokenRepo, *jwtHelper)
	userController := controller.NewUserController(userSvc, userTokenSvc)

	// Toy category
	toyCategoryRepo := repository.NewToyCategoryRepository(db)
	toyCategorySvc := service.NewToyCategoryService(toyCategoryRepo)
	toyCategoryController := controller.NewToyCategoryController(toyCategorySvc)

	// Toy Images
	toyImageRepo := repository.NewToyImageRepository(db)
	toyImageSvc := service.NewToyImageService(toyImageRepo)
	toyImageController := controller.NewToyImageController(toyImageSvc)

	// Toy
	toyRepo := repository.NewToyRepository(db)
	toySvc := service.NewToyService(toyRepo, toyImageRepo, toyCategoryRepo)
	toyController := controller.NewToyController(toySvc)

	// Rental
	rentalRepo := repository.NewRentalRepository(db)

	// Payment
	paymentRepo := repository.NewPaymentRepository(db)
	midtransSvc := service.NewMidtransService(cfg)
	paymentSvc := service.NewPaymentService(paymentRepo, rentalRepo, midtransSvc)
	paymentController := controller.NewPaymentController(paymentSvc)

	rentalSvc := service.NewRentalService(rentalRepo, userRepo, toyRepo, paymentSvc)
	rentalController := controller.NewRentalController(rentalSvc)

	// Report
	businessReportRepo := repository.NewBusinessReportRepository(db)
	businessReportSvc := service.NewBusinessReportService(businessReportRepo)
	businessReportController := controller.NewBusinessReportController(businessReportSvc)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(*jwtHelper, userTokenSvc)

	// Public routes
	public := r.Group("/api")
	{
		// User routes
		auth := public.Group("/user")
		{
			auth.POST("/auth/register", userController.Insert)
			auth.POST("/auth/login", userController.Login)
		}

		// Toy category routes
		toyCategory := public.Group("/toy")
		{
			toyCategory.GET("/category", toyCategoryController.FindAll)
			toyCategory.GET("/category/:id", toyCategoryController.FinById)
		}

		// Toy image
		toyImage := public.Group("/toy")
		{
			toyImage.GET("/image", toyImageController.FindAll)
		}

		// Toy routes
		toy := public.Group("/toy")
		{
			toy.GET("", toyController.FindAll)
			toy.GET("/:id", toyController.FinById)
		}

		// Payment routes
		payment := public.Group("/payment")
		{
			payment.POST("/callback", paymentController.HandlePaymentCallback)
		}
	}

	// Protected routes
	protected := r.Group("/api")
	protected.Use(authMiddleware.AuthMiddleware())
	{
		// User routes
		auth := protected.Group("/user")
		{
			auth.PUT("/auth/:id", userController.UpdateById)
			auth.DELETE("/auth/:id", userController.DeleteById)
			auth.DELETE("/auth/logout", userController.Logout)
			auth.GET("/auth/me", userController.Me)
		}

		// Rental routes
		rental := protected.Group("/rental")
		{
			rental.POST("", rentalController.Insert)
			rental.PUT("/:id", rentalController.UpdateById)
		}

		// Payment routes
		payment := protected.Group("/payment")
		{
			payment.POST("", paymentController.CreatePayment)
			payment.GET("/:id", paymentController.GetPaymentByID)
			payment.GET("/rental/:rental_id", paymentController.GetPaymentsByRentalID)
		}
	}

	// Admin routes
	admin := r.Group("/api")
	admin.Use(authMiddleware.AdminMiddleware())
	{
		// Admin user routes
		auth := admin.Group("/admin")
		{
			auth.GET("/users", userController.FindAll)
			auth.GET("/user/:id", userController.FinById)
		}

		// Admin toy category routes
		toyCategory := admin.Group("/toy")
		{
			toyCategory.POST("/category", toyCategoryController.Insert)
			toyCategory.PUT("/category/:id", toyCategoryController.UpdateById)
			toyCategory.DELETE("/category/:id", toyCategoryController.DeleteById)
		}

		// Admin toy images routes
		toyImage := admin.Group("/toy")
		{
			toyImage.POST("/image", toyImageController.Insert)
			toyImage.DELETE("/image/:id", toyImageController.DeleteById)
		}

		// Admin toy routes
		toy := admin.Group("/toy")
		{
			toy.POST("", toyController.Insert)
			toy.PUT("/:id", toyController.UpdateById)
			toy.DELETE("/:id", toyController.DeleteById)
		}

		// Admin rental routes
		rental := admin.Group("/rental")
		{
			rental.GET("", rentalController.FindAll)
			rental.GET("/:id", rentalController.FinById)
			rental.PUT("/:id/return", rentalController.ReturnRental)
		}

		// Admin report routes
		report := admin.Group("/business-report")
		{
			report.GET("/sales", businessReportController.GetSalesReport)
			report.GET("/popular-toys", businessReportController.GetPopularToysReport)
			report.GET("/customers", businessReportController.GetTopCustomersReport)
			report.GET("/rental-status", businessReportController.GetRentalStatusReport)
		}
	}

	return r
}
