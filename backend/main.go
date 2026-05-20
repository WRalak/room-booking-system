package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    "backend/handlers"
)

type AppConfig struct {
    PaystackSecretKey   string
    MpesaConsumerKey    string
    MpesaConsumerSecret string
    MpesaShortcode      string
    MpesaPasskey        string
    DBURL               string
}

var config *AppConfig

func main() {
    // Load environment variables
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    config = &AppConfig{
        PaystackSecretKey: os.Getenv("PAYSTACK_SECRET_KEY"),
        MpesaConsumerKey:  os.Getenv("MPESA_CONSUMER_KEY"),
        MpesaConsumerSecret: os.Getenv("MPESA_CONSUMER_SECRET"),
        MpesaShortcode:    os.Getenv("MPESA_SHORTCODE"),
        MpesaPasskey:      os.Getenv("MPESA_PASSKEY"),
        DBURL:             os.Getenv("DATABASE_URL"),
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
        AllowOrigins:     []string{"http://localhost:3000"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))

    // Initialize handlers
    paymentHandler := handlers.NewPaymentHandler(db, redisClient, config)
    bookingHandler := handlers.NewBookingHandler(db, redisClient)
    notificationHandler := handlers.NewNotificationHandler(redisClient)

    // API routes
    api := router.Group("/api/v1")
    {
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
        notifications := api.Group("/notifications")
        {
            notifications.GET("/user/:user_id", notificationHandler.GetUserNotifications)
            notifications.POST("/mark-read/:notification_id", notificationHandler.MarkAsRead)
            notifications.GET("/ws/:user_id", notificationHandler.WebSocketHandler)
        }

        // Room routes
        rooms := api.Group("/rooms")
        {
            rooms.GET("/", bookingHandler.GetAvailableRooms)
            rooms.GET("/:room_id", bookingHandler.GetRoomDetails)
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