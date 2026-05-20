package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
    "gorm.io/gorm"
    "backend/models"
)

type RecurringBookingHandler struct {
    db    *gorm.DB
    redis *redis.Client
}

type CreateRecurringRequest struct {
    RoomID      uint      `json:"room_id" binding:"required"`
    UserID      uint      `json:"user_id" binding:"required"`
    StartTime   time.Time `json:"start_time" binding:"required"`
    EndTime     time.Time `json:"end_time" binding:"required"`
    Frequency   string    `json:"frequency" binding:"required"` // daily, weekly, monthly
    Interval    int       `json:"interval" binding:"required"`
    DayOfWeek   *int      `json:"day_of_week"`   // 0-6 for weekly
    DayOfMonth  *int      `json:"day_of_month"`  // 1-31 for monthly
    EndDate     *time.Time `json:"end_date"`
    Occurrences *int      `json:"occurrences"`   // number of occurrences
}

func NewRecurringBookingHandler(db *gorm.DB, redis *redis.Client) *RecurringBookingHandler {
    return &RecurringBookingHandler{
        db:    db,
        redis: redis,
    }
}

func (h *RecurringBookingHandler) CreateRecurringBooking(c *gin.Context) {
    var req CreateRecurringRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Calculate duration
    duration := req.EndTime.Sub(req.StartTime)
    totalAmount := duration.Hours() * h.getRoomPrice(req.RoomID) / 100
    
    // Create template booking
    booking := models.Booking{
        UserID:      req.UserID,
        RoomID:      req.RoomID,
        StartTime:   req.StartTime,
        EndTime:     req.EndTime,
        TotalAmount: totalAmount,
        Status:      "pending",
        PaymentStatus: "unpaid",
    }
    
    if err := h.db.Create(&booking).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template booking"})
        return
    }
    
    // Create recurring booking record
    recurring := models.RecurringBooking{
        BookingID:      booking.ID,
        UserID:         req.UserID,
        RoomID:         req.RoomID,
        Frequency:      req.Frequency,
        Interval:       req.Interval,
        DayOfWeek:      req.DayOfWeek,
        DayOfMonth:     req.DayOfMonth,
        StartDate:      req.StartTime,
        EndDate:        req.EndDate,
        NextOccurrence: h.calculateNextOccurrence(req.StartTime, req.Frequency, req.Interval, req.DayOfWeek, req.DayOfMonth),
        IsActive:       true,
    }
    
    h.db.Create(&recurring)

    // Generate future occurrences
    if req.Occurrences != nil {
        go h.generateFutureOccurrences(recurring.ID, *req.Occurrences)
    }
    
    c.JSON(http.StatusCreated, gin.H{
        "booking": booking,
        "recurring": recurring,
    })
}

func (h *RecurringBookingHandler) generateFutureOccurrences(recurringID uint, occurrences int) {
    var recurring models.RecurringBooking
    h.db.Preload("Booking").First(&recurring, recurringID)
    
    generated := 0
    nextDate := recurring.NextOccurrence
    
    for (occurrences == 0 || generated < occurrences) && 
          (recurring.EndDate == nil || nextDate.Before(*recurring.EndDate)) {
        
        // Check if room is available
        if h.isRoomAvailable(recurring.RoomID, nextDate, nextDate.Add(recurring.Booking.EndTime.Sub(recurring.Booking.StartTime))) {
            // Create booking occurrence
            newBooking := models.Booking{
                UserID:      recurring.UserID,
                RoomID:      recurring.RoomID,
                StartTime:   nextDate,
                EndTime:     nextDate.Add(recurring.Booking.EndTime.Sub(recurring.Booking.StartTime)),
                TotalAmount: recurring.Booking.TotalAmount,
                Status:      "pending",
                PaymentStatus: "unpaid",
            }
            
            h.db.Create(&newBooking)
            generated++
        }
        
        // Calculate next occurrence
        nextDate = h.calculateNextOccurrence(nextDate, recurring.Frequency, recurring.Interval, 
            recurring.DayOfWeek, recurring.DayOfMonth)
    }
    
    // Update next occurrence
    recurring.NextOccurrence = nextDate
    h.db.Save(&recurring)
}

func (h *RecurringBookingHandler) GetUserRecurringBookings(c *gin.Context) {
    userID := c.Param("user_id")
    
    var recurring []models.RecurringBooking
    h.db.Preload("Booking").Preload("Room").
        Where("user_id = ? AND is_active = ?", userID, true).
        Find(&recurring)
    
    c.JSON(http.StatusOK, recurring)
}

func (h *RecurringBookingHandler) CancelRecurringBooking(c *gin.Context) {
    recurringID := c.Param("id")
    
    var recurring models.RecurringBooking
    if err := h.db.First(&recurring, recurringID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Recurring booking not found"})
        return
    }
    
    recurring.IsActive = false
    h.db.Save(&recurring)
    
    // Cancel all future occurrences
    h.db.Model(&models.Booking{}).
        Where("room_id = ? AND user_id = ? AND start_time > ? AND status = 'pending'",
            recurring.RoomID, recurring.UserID, time.Now()).
        Update("status", "cancelled")
    
    c.JSON(http.StatusOK, gin.H{"message": "Recurring booking cancelled"})
}

func (h *RecurringBookingHandler) calculateNextOccurrence(current time.Time, frequency string, interval int, dayOfWeek *int, dayOfMonth *int) time.Time {
    switch frequency {
    case "daily":
        return current.AddDate(0, 0, interval)
    case "weekly":
        next := current.AddDate(0, 0, interval*7)
        if dayOfWeek != nil {
            // Adjust to specific day of week
            daysToAdd := (*dayOfWeek - int(next.Weekday()) + 7) % 7
            next = next.AddDate(0, 0, daysToAdd)
        }
        return next
    case "monthly":
        next := current.AddDate(0, interval, 0)
        if dayOfMonth != nil {
            // Adjust to specific day of month
            targetDay := *dayOfMonth
            if targetDay > 28 {
                // Handle month end
                lastDay := time.Date(next.Year(), next.Month()+1, 0, 0, 0, 0, 0, next.Location()).Day()
                if targetDay > lastDay {
                    targetDay = lastDay
                }
            }
            next = time.Date(next.Year(), next.Month(), targetDay, next.Hour(), next.Minute(), 0, 0, next.Location())
        }
        return next
    }
    return current
}

func (h *RecurringBookingHandler) getRoomPrice(roomID uint) float64 {
    var room models.Room
    h.db.Select("price_per_hour").First(&room, roomID)
    return room.PricePerHour
}

func (h *RecurringBookingHandler) isRoomAvailable(roomID uint, start, end time.Time) bool {
    var count int64
    h.db.Model(&models.Booking{}).
        Where("room_id = ? AND status != 'cancelled' AND ((start_time BETWEEN ? AND ?) OR (end_time BETWEEN ? AND ?))",
            roomID, start, end, start, end).
        Count(&count)
    return count == 0
}