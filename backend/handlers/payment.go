package handlers

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
    "gorm.io/gorm"
    "backend/models"
)

type AppConfig struct {
    PaystackSecretKey   string
    MpesaConsumerKey    string
    MpesaConsumerSecret string
    MpesaShortcode      string
    MpesaPasskey        string
    DBURL               string
}

type PaymentHandler struct {
    db     *gorm.DB
    redis  *redis.Client
    config interface{}
}

type PaystackInitializeRequest struct {
    Email     string  `json:"email"`
    Amount    float64 `json:"amount"`
    BookingID uint    `json:"booking_id"`
}

type MpesaSTKPushRequest struct {
    PhoneNumber string  `json:"phone_number"`
    Amount      float64 `json:"amount"`
    BookingID   uint    `json:"booking_id"`
}

func NewPaymentHandler(db *gorm.DB, redis *redis.Client, cfg interface{}) *PaymentHandler {
    return &PaymentHandler{
        db:     db,
        redis:  redis,
        config: cfg,
    }
}

func (h *PaymentHandler) InitializePaystackPayment(c *gin.Context) {
    var req PaystackInitializeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Get booking details
    var booking models.Booking
    if err := h.db.Preload("Room").First(&booking, req.BookingID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
        return
    }

    // Initialize Paystack payment
    paystackReq := map[string]interface{}{
        "email": req.Email,
        "amount": int(req.Amount * 100), // Paystack uses kobo/cents
        "reference": fmt.Sprintf("BOOK-%d-%d", booking.ID, time.Now().Unix()),
        "callback_url": "https://yourdomain.com/payment/callback",
        "metadata": map[string]interface{}{
            "booking_id": booking.ID,
            "user_id":    booking.UserID,
        },
    }

    jsonData, _ := json.Marshal(paystackReq)
    
    paystackURL := "https://api.paystack.co/transaction/initialize"
    request, _ := http.NewRequest("POST", paystackURL, bytes.NewBuffer(jsonData))
    request.Header.Set("Authorization", "Bearer "+h.config.(*AppConfig).PaystackSecretKey)
    request.Header.Set("Content-Type", "application/json")

    client := &http.Client{Timeout: 30 * time.Second}
    response, err := client.Do(request)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment initialization failed"})
        return
    }
    defer response.Body.Close()

    body, _ := io.ReadAll(response.Body)
    
    var paystackResp map[string]interface{}
    json.Unmarshal(body, &paystackResp)

    // Store payment reference
    payment := models.Payment{
        BookingID:     booking.ID,
        Amount:        req.Amount,
        Status:        "pending",
        PaymentMethod: "paystack",
        ProviderRef:   paystackReq["reference"].(string),
    }
    h.db.Create(&payment)

    c.JSON(http.StatusOK, paystackResp)
}

func (h *PaymentHandler) InitiateMpesaPayment(c *gin.Context) {
    var req MpesaSTKPushRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Get access token
    token, err := h.getMpesaAccessToken()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get M-Pesa token"})
        return
    }

    // Format phone number (remove leading 0 or +254)
    phoneNumber := req.PhoneNumber
    if phoneNumber[0] == '0' {
        phoneNumber = "254" + phoneNumber[1:]
    } else if phoneNumber[0:3] == "+254" {
        phoneNumber = phoneNumber[1:]
    }

    timestamp := time.Now().Format("20060102150405")
    password := h.generateMpesaPassword(timestamp)

    stkPushReq := map[string]interface{}{
        "BusinessShortCode": h.config.(*AppConfig).MpesaShortcode,
        "Password":          password,
        "Timestamp":         timestamp,
        "TransactionType":   "CustomerPayBillOnline",
        "Amount":            int(req.Amount),
        "PartyA":            phoneNumber,
        "PartyB":            h.config.(*AppConfig).MpesaShortcode,
        "PhoneNumber":       phoneNumber,
        "CallBackURL":       "https://yourdomain.com/api/v1/payments/mpesa/callback",
        "AccountReference":  fmt.Sprintf("BOOK-%d", req.BookingID),
        "TransactionDesc":   "Room Booking Payment",
    }

    jsonData, _ := json.Marshal(stkPushReq)
    
    mpesaURL := "https://sandbox.safaricom.co.ke/mpesa/stkpush/v1/processrequest"
    request, _ := http.NewRequest("POST", mpesaURL, bytes.NewBuffer(jsonData))
    request.Header.Set("Authorization", "Bearer "+token)
    request.Header.Set("Content-Type", "application/json")

    client := &http.Client{Timeout: 30 * time.Second}
    response, err := client.Do(request)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "M-Pesa request failed"})
        return
    }
    defer response.Body.Close()

    body, _ := io.ReadAll(response.Body)
    
    var mpesaResp map[string]interface{}
    json.Unmarshal(body, &mpesaResp)

    // Store payment record
    payment := models.Payment{
        BookingID:     req.BookingID,
        Amount:        req.Amount,
        Status:        "pending",
        PaymentMethod: "mpesa",
        ProviderRef:   fmt.Sprintf("%s-%d", mpesaResp["CheckoutRequestID"], time.Now().Unix()),
    }
    h.db.Create(&payment)

    c.JSON(http.StatusOK, mpesaResp)
}

func (h *PaymentHandler) PaystackWebhook(c *gin.Context) {
    var webhookData map[string]interface{}
    if err := c.ShouldBindJSON(&webhookData); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook data"})
        return
    }

    event := webhookData["event"].(string)
    if event == "charge.success" {
        data := webhookData["data"].(map[string]interface{})
        reference := data["reference"].(string)
        
        // Update payment status
        var payment models.Payment
        if err := h.db.Where("provider_ref = ?", reference).First(&payment).Error; err == nil {
            payment.Status = "success"
            h.db.Save(&payment)

            // Update booking status
            var booking models.Booking
            h.db.First(&booking, payment.BookingID)
            booking.PaymentStatus = "paid"
            booking.Status = "confirmed"
            h.db.Save(&booking)

            // Send notification
            h.sendPaymentNotification(booking.UserID, "paystack", payment.Amount)
        }
    }

    c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *PaymentHandler) MpesaCallback(c *gin.Context) {
    var callbackData map[string]interface{}
    if err := c.ShouldBindJSON(&callbackData); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid callback data"})
        return
    }

    // Extract result code
    body := callbackData["Body"].(map[string]interface{})
    stkCallback := body["stkCallback"].(map[string]interface{})
    resultCode := stkCallback["ResultCode"].(float64)

    if resultCode == 0 {
        // Payment successful
        checkoutRequestID := stkCallback["CheckoutRequestID"].(string)
        
        // Update payment status
        var payment models.Payment
        if err := h.db.Where("provider_ref LIKE ?", "%"+checkoutRequestID+"%").First(&payment).Error; err == nil {
            payment.Status = "success"
            h.db.Save(&payment)

            // Update booking status
            var booking models.Booking
            h.db.First(&booking, payment.BookingID)
            booking.PaymentStatus = "paid"
            booking.Status = "confirmed"
            h.db.Save(&booking)

            // Send notification
            h.sendPaymentNotification(booking.UserID, "mpesa", payment.Amount)
        }
    }

    c.JSON(http.StatusOK, gin.H{"ResultCode": 0, "ResultDesc": "Success"})
}

func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
    reference := c.Param("reference")
    
    var payment models.Payment
    if err := h.db.Where("payment_ref = ? OR provider_ref = ?", reference, reference).First(&payment).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
        return
    }

    c.JSON(http.StatusOK, payment)
}

// Helper methods
func (h *PaymentHandler) getMpesaAccessToken() (string, error) {
    url := "https://sandbox.safaricom.co.ke/oauth/v1/generate?grant_type=client_credentials"
    request, _ := http.NewRequest("GET", url, nil)
    request.SetBasicAuth(h.config.(*AppConfig).MpesaConsumerKey, h.config.(*AppConfig).MpesaConsumerSecret)

    client := &http.Client{Timeout: 30 * time.Second}
    response, err := client.Do(request)
    if err != nil {
        return "", err
    }
    defer response.Body.Close()

    body, _ := io.ReadAll(response.Body)
    var tokenResp map[string]interface{}
    json.Unmarshal(body, &tokenResp)

    return tokenResp["access_token"].(string), nil
}

func (h *PaymentHandler) generateMpesaPassword(timestamp string) string {
    data := h.config.(*AppConfig).MpesaShortcode + h.config.(*AppConfig).MpesaPasskey + timestamp
    // Return base64 encoded string
    return data // You'll need to implement proper base64 encoding
}

func (h *PaymentHandler) sendPaymentNotification(userID uint, method string, amount float64) {
    notification := models.Notification{
        UserID:  userID,
        Title:   "Payment Successful",
        Message: fmt.Sprintf("Your payment of %.2f KES via %s has been confirmed", amount, method),
        Type:    "payment_success",
    }
    h.db.Create(&notification)

    // Publish to Redis for real-time notification
    h.redis.Publish(context.Background(), "notifications", fmt.Sprintf("%d", userID))
}