package services

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "net/smtp"
    "os"
    "text/template"
    "time"

    "github.com/go-redis/redis/v8"
    "gorm.io/gorm"
)

type NotificationService struct {
    db    *gorm.DB
    redis *redis.Client
}

type EmailData struct {
    To          string
    Subject     string
    TemplateName string
    Data        map[string]interface{}
}

type SMSData struct {
    To      string
    Message string
}

func NewNotificationService(db *gorm.DB, redis *redis.Client) *NotificationService {
    service := &NotificationService{
        db:    db,
        redis: redis,
    }
    
    // Start background workers
    go service.processEmailQueue()
    go service.processSMSQueue()
    
    return service
}

func (s *NotificationService) SendBookingConfirmation(bookingID uint) {
    var booking models.Booking
    s.db.Preload("User").Preload("Room").First(&booking, bookingID)
    
    // Email
    emailData := EmailData{
        To:      booking.User.Email,
        Subject: "Booking Confirmation - RoomBooker",
        TemplateName: "booking_confirmation",
        Data: map[string]interface{}{
            "user_name": booking.User.Name,
            "room_name": booking.Room.Name,
            "date":      booking.StartTime.Format("2006-01-02"),
            "start_time": booking.StartTime.Format("15:04"),
            "end_time":   booking.EndTime.Format("15:04"),
            "total_amount": booking.TotalAmount,
            "booking_ref":  booking.BookingRef,
        },
    }
    s.queueEmail(emailData)
    
    // SMS if phone number exists
    if booking.User.Phone != "" {
        smsData := SMSData{
            To: booking.User.Phone,
            Message: fmt.Sprintf("Booking confirmed: %s on %s at %s. Ref: %s", 
                booking.Room.Name, 
                booking.StartTime.Format("2006-01-02"),
                booking.StartTime.Format("15:04"),
                booking.BookingRef),
        }
        s.queueSMS(smsData)
    }
    
    // Database notification
    s.createDBNotification(booking.UserID, "Booking Confirmed", 
        fmt.Sprintf("Your booking for %s has been confirmed", booking.Room.Name), 
        "booking_confirmation")
}

func (s *NotificationService) SendPaymentReceipt(bookingID uint, paymentMethod string) {
    var booking models.Booking
    s.db.Preload("User").Preload("Room").First(&booking, bookingID)
    
    var payment models.Payment
    s.db.Where("booking_id = ?", bookingID).First(&payment)
    
    // Email receipt
    emailData := EmailData{
        To:      booking.User.Email,
        Subject: "Payment Receipt - RoomBooker",
        TemplateName: "payment_receipt",
        Data: map[string]interface{}{
            "user_name":    booking.User.Name,
            "amount":       payment.Amount,
            "payment_method": paymentMethod,
            "payment_date": payment.CreatedAt.Format("2006-01-02 15:04"),
            "transaction_id": payment.ProviderRef,
            "booking_ref":   booking.BookingRef,
        },
    }
    s.queueEmail(emailData)
    
    // SMS for M-Pesa transactions
    if paymentMethod == "mpesa" && booking.User.Phone != "" {
        smsData := SMSData{
            To: booking.User.Phone,
            Message: fmt.Sprintf("Payment of KES %.2f received. Transaction ID: %s", 
                payment.Amount, payment.ProviderRef),
        }
        s.queueSMS(smsData)
    }
}

func (s *NotificationService) SendBookingReminder() {
    // Find bookings starting in 24 hours
    tomorrow := time.Now().Add(24 * time.Hour)
    dayAfter := tomorrow.Add(24 * time.Hour)
    
    var bookings []models.Booking
    s.db.Preload("User").Preload("Room").
        Where("start_time BETWEEN ? AND ? AND status = 'confirmed'", tomorrow, dayAfter).
        Find(&bookings)
    
    for _, booking := range bookings {
        // Email reminder
        emailData := EmailData{
            To:      booking.User.Email,
            Subject: "Booking Reminder - Tomorrow",
            TemplateName: "booking_reminder",
            Data: map[string]interface{}{
                "user_name": booking.User.Name,
                "room_name": booking.Room.Name,
                "date":      booking.StartTime.Format("2006-01-02"),
                "start_time": booking.StartTime.Format("15:04"),
                "end_time":   booking.EndTime.Format("15:04"),
            },
        }
        s.queueEmail(emailData)
        
        // SMS reminder
        if booking.User.Phone != "" {
            smsData := SMSData{
                To: booking.User.Phone,
                Message: fmt.Sprintf("Reminder: You have a booking tomorrow at %s from %s to %s", 
                    booking.Room.Name,
                    booking.StartTime.Format("15:04"),
                    booking.EndTime.Format("15:04")),
            }
            s.queueSMS(smsData)
        }
        
        // Database notification
        s.createDBNotification(booking.UserID, "Upcoming Booking Reminder",
            fmt.Sprintf("You have a booking for %s tomorrow at %s", 
                booking.Room.Name, booking.StartTime.Format("15:04")),
            "reminder")
    }
}

func (s *NotificationService) queueEmail(emailData EmailData) {
    data, _ := json.Marshal(emailData)
    s.redis.LPush(context.Background(), "email_queue", data)
}

func (s *NotificationService) queueSMS(smsData SMSData) {
    data, _ := json.Marshal(smsData)
    s.redis.LPush(context.Background(), "sms_queue", data)
}

func (s *NotificationService) processEmailQueue() {
    for {
        result, err := s.redis.BRPop(context.Background(), 0, "email_queue").Result()
        if err != nil {
            continue
        }
        
        var emailData EmailData
        json.Unmarshal([]byte(result[1]), &emailData)
        
        s.sendEmail(emailData)
    }
}

func (s *NotificationService) sendEmail(emailData EmailData) {
    // Parse email template
    tmpl, err := template.ParseFiles(fmt.Sprintf("templates/email/%s.html", emailData.TemplateName))
    if err != nil {
        fmt.Printf("Error parsing email template: %v\n", err)
        return
    }
    
    var body bytes.Buffer
    if err := tmpl.Execute(&body, emailData.Data); err != nil {
        fmt.Printf("Error executing template: %v\n", err)
        return
    }
    
    // SMTP configuration
    from := os.Getenv("SMTP_FROM")
    password := os.Getenv("SMTP_PASSWORD")
    smtpHost := os.Getenv("SMTP_HOST")
    smtpPort := os.Getenv("SMTP_PORT")
    
    to := []string{emailData.To}
    
    msg := []byte(fmt.Sprintf("To: %s\r\n"+
        "Subject: %s\r\n"+
        "Content-Type: text/html; charset=utf-8\r\n"+
        "\r\n%s\r\n", emailData.To, emailData.Subject, body.String()))
    
    auth := smtp.PlainAuth("", from, password, smtpHost)
    
    err = smtp.SendMail(fmt.Sprintf("%s:%s", smtpHost, smtpPort), auth, from, to, msg)
    if err != nil {
        fmt.Printf("Error sending email: %v\n", err)
    }
}

func (s *NotificationService) processSMSQueue() {
    for {
        result, err := s.redis.BRPop(context.Background(), 0, "sms_queue").Result()
        if err != nil {
            continue
        }
        
        var smsData SMSData
        json.Unmarshal([]byte(result[1]), &smsData)
        
        s.sendSMS(smsData)
    }
}

func (s *NotificationService) sendSMS(smsData SMSData) {
    // Using Africa's Talking API for SMS (Kenya)
    apiKey := os.Getenv("AFRICASTALKING_API_KEY")
    username := os.Getenv("AFRICASTALKING_USERNAME")
    
    url := "https://api.africastalking.com/version1/messaging"
    
    data := map[string]string{
        "username": username,
        "to":       smsData.To,
        "message":  smsData.Message,
    }
    
    jsonData, _ := json.Marshal(data)
    
    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    req.Header.Set("apiKey", apiKey)
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf("Error sending SMS: %v\n", err)
        return
    }
    defer resp.Body.Close()
    
    // For M-Pesa specific SMS (using Safaricom's API)
    if strings.Contains(smsData.Message, "M-Pesa") {
        s.sendMpesaSpecificSMS(smsData)
    }
}

func (s *NotificationService) sendMpesaSpecificSMS(smsData SMSData) {
    // Send transaction-specific SMS via M-Pesa API
    // This would use the same M-Pesa API for sending receipts
    fmt.Printf("Sending M-Pesa specific SMS to %s: %s\n", smsData.To, smsData.Message)
}

func (s *NotificationService) createDBNotification(userID uint, title, message, notifType string) {
    notification := models.Notification{
        UserID:  userID,
        Title:   title,
        Message: message,
        Type:    notifType,
    }
    s.db.Create(&notification)
    
    // Publish to Redis for real-time notifications
    s.redis.Publish(context.Background(), "notifications", fmt.Sprintf("%d", userID))
}