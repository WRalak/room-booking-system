package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	PaystackSecretKey   string
	MpesaConsumerKey    string
	MpesaConsumerSecret string
	MpesaShortcode      string
	MpesaPasskey        string
	DBURL               string
	JWTSecret           string
}

var config *AppConfig

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	config = &AppConfig{
		PaystackSecretKey:   os.Getenv("PAYSTACK_SECRET_KEY"),
		MpesaConsumerKey:    os.Getenv("MPESA_CONSUMER_KEY"),
		MpesaConsumerSecret: os.Getenv("MPESA_CONSUMER_SECRET"),
		MpesaShortcode:      os.Getenv("MPESA_SHORTCODE"),
		MpesaPasskey:        os.Getenv("MPESA_PASSKEY"),
		DBURL:               os.Getenv("DATABASE_URL"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
	}
	if config.JWTSecret == "" {
		config.JWTSecret = "dev-secret-change-me"
	}

	// Initialize database
	db, err := initDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize Redis for notifications
	redisClient := initRedis()

	// Setup Gin router
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initialize handlers
	paymentHandler := handlers.NewPaymentHandler(db, redisClient, config)
	bookingHandler := handlers.NewBookingHandler(db, redisClient)
	notificationHandler := handlers.NewNotificationHandler(db, redisClient)
	adminHandler := handlers.NewAdminHandler(db)
	authHandler := handlers.NewAuthHandler(db, config.JWTSecret)
	reviewHandler := handlers.NewReviewHandler(db)
	invoiceHandler := handlers.NewInvoiceHandler(db)
	recurringBookingHandler := handlers.NewRecurringBookingHandler(db, redisClient)

	// API routes
	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", authHandler.AuthMiddleware(), authHandler.Me)
		}

		// Booking routes
		bookings := api.Group("/bookings")
		{
			bookings.POST("/", bookingHandler.CreateBooking)
			bookings.GET("/user/:user_id", bookingHandler.GetUserBookings)
			bookings.GET("/room/:room_id", bookingHandler.GetRoomBookings)
			bookings.PUT("/:booking_id/cancel", bookingHandler.CancelBooking)
		}

		// Payment routes
		payments := api.Group("/payments")
		{
			payments.POST("/paystack/initialize", paymentHandler.InitializePaystackPayment)
			payments.POST("/paystack/webhook", paymentHandler.PaystackWebhook)
			payments.POST("/mpesa/stkpush", paymentHandler.InitiateMpesaPayment)
			payments.POST("/mpesa/callback", paymentHandler.MpesaCallback)
			payments.GET("/status/:reference", paymentHandler.GetPaymentStatus)
		}

		// Notification routes
		notifications := api.Group("/notifications", authHandler.AuthMiddleware())
		{
			notifications.GET("/user/:user_id", notificationHandler.GetUserNotifications)
			notifications.POST("/mark-read/:notification_id", notificationHandler.MarkAsRead)
			notifications.GET("/ws/:user_id", notificationHandler.WebSocketHandler)
		}

		reviews := api.Group("/reviews")
		{
			reviews.POST("", authHandler.AuthMiddleware(), reviewHandler.CreateReview)
			reviews.GET("/room/:room_id", reviewHandler.GetRoomReviews)
			reviews.GET("/user/:user_id", authHandler.AuthMiddleware(), reviewHandler.GetUserReviews)
		}

		invoices := api.Group("/invoices", authHandler.AuthMiddleware())
		{
			invoices.GET("/booking/:booking_id", invoiceHandler.GenerateInvoice)
			invoices.GET("/:id/download", invoiceHandler.DownloadInvoice)
			invoices.GET("/user/:user_id", invoiceHandler.GetUserInvoices)
		}

		recurringBookings := api.Group("/recurring-bookings", authHandler.AuthMiddleware())
		{
			recurringBookings.POST("", recurringBookingHandler.CreateRecurringBooking)
			recurringBookings.GET("/user/:user_id", recurringBookingHandler.GetUserRecurringBookings)
			recurringBookings.PUT("/:id/cancel", recurringBookingHandler.CancelRecurringBooking)
		}

		// Room routes
		rooms := api.Group("/rooms")
		{
			rooms.GET("/", bookingHandler.GetAvailableRooms)
			rooms.GET("/:room_id", bookingHandler.GetRoomDetails)
		}

		// Admin routes
		admin := api.Group("/admin", authHandler.AuthMiddleware(), authHandler.AdminOnly())
		{
			admin.GET("/dashboard", adminHandler.GetDashboardStats)
			admin.GET("/revenue", adminHandler.GetRevenueReport)
			admin.GET("/rooms", adminHandler.GetAllRooms)
			admin.POST("/rooms", adminHandler.ManageRoom)
			admin.PUT("/rooms/:id", adminHandler.ManageRoom)
			admin.DELETE("/rooms/:id", adminHandler.DeleteRoom)
			admin.GET("/bookings", adminHandler.GetAllBookings)
			admin.GET("/users", adminHandler.GetAllUsers)
			admin.POST("/reviews/:id/verify", adminHandler.VerifyReview)
		}
	}

	// Start server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
}
