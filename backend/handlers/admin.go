package handlers

import (
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "backend/models"
)

type AdminHandler struct {
    db *gorm.DB
}

type DashboardStats struct {
    TotalUsers       int64                   `json:"total_users"`
    TotalBookings    int64                   `json:"total_bookings"`
    TotalRevenue     float64                 `json:"total_revenue"`
    MonthlyRevenue   float64                 `json:"monthly_revenue"`
    OccupancyRate    float64                 `json:"occupancy_rate"`
    PopularRooms     []PopularRoom           `json:"popular_rooms"`
    RecentBookings   []models.Booking        `json:"recent_bookings"`
    PendingReviews   int64                   `json:"pending_reviews"`
    RevenueByPayment map[string]float64      `json:"revenue_by_payment"`
    BookingStatus    map[string]int64        `json:"booking_status"`
}

type PopularRoom struct {
    RoomID      uint    `json:"room_id"`
    RoomName    string  `json:"room_name"`
    BookingCount int64  `json:"booking_count"`
    Revenue     float64 `json:"revenue"`
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
    return &AdminHandler{db: db}
}

func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
    stats := DashboardStats{}
    
    // Total users
    h.db.Model(&models.User{}).Count(&stats.TotalUsers)
    
    // Total bookings
    h.db.Model(&models.Booking{}).Where("status != 'cancelled'").Count(&stats.TotalBookings)
    
    // Total revenue
    h.db.Model(&models.Payment{}).Where("status = 'success'").Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalRevenue)
    
    // Monthly revenue
    startOfMonth := time.Now().Truncate(24 * time.Hour).AddDate(0, 0, -time.Now().Day()+1)
    h.db.Model(&models.Payment{}).
        Where("status = 'success' AND created_at >= ?", startOfMonth).
        Select("COALESCE(SUM(amount), 0)").
        Scan(&stats.MonthlyRevenue)
    
    // Popular rooms
    h.db.Table("bookings").
        Select("room_id, rooms.name as room_name, COUNT(*) as booking_count, SUM(total_amount) as revenue").
        Joins("JOIN rooms ON rooms.id = bookings.room_id").
        Where("bookings.status != 'cancelled'").
        Group("room_id, rooms.name").
        Order("booking_count DESC").
        Limit(5).
        Scan(&stats.PopularRooms)
    
    // Recent bookings
    h.db.Preload("User").Preload("Room").
        Order("created_at DESC").
        Limit(10).
        Find(&stats.RecentBookings)
    
    // Pending reviews
    h.db.Model(&models.Review{}).Where("is_verified = ?", false).Count(&stats.PendingReviews)
    
    // Revenue by payment method
    var paymentMethods []struct {
        PaymentMethod string
        Total         float64
    }
    h.db.Model(&models.Payment{}).
        Select("payment_method, SUM(amount) as total").
        Where("status = 'success'").
        Group("payment_method").
        Scan(&paymentMethods)
    
    stats.RevenueByPayment = make(map[string]float64)
    for _, pm := range paymentMethods {
        stats.RevenueByPayment[pm.PaymentMethod] = pm.Total
    }
    
    // Booking status distribution
    var statusCounts []struct {
        Status string
        Count  int64
    }
    h.db.Model(&models.Booking{}).
        Select("status, COUNT(*) as count").
        Group("status").
        Scan(&statusCounts)
    
    stats.BookingStatus = make(map[string]int64)
    for _, sc := range statusCounts {
        stats.BookingStatus[sc.Status] = sc.Count
    }
    
    c.JSON(http.StatusOK, stats)
}

func (h *AdminHandler) GetAllBookings(c *gin.Context) {
    page := c.DefaultQuery("page", "1")
    limit := c.DefaultQuery("limit", "20")
    status := c.Query("status")
    
    var bookings []models.Booking
    query := h.db.Preload("User").Preload("Room")
    
    if status != "" {
        query = query.Where("status = ?", status)
    }
    
    offset, _ := strconv.Atoi(page)
    limitInt, _ := strconv.Atoi(limit)
    
    query.Offset((offset - 1) * limitInt).Limit(limitInt).
        Order("created_at DESC").
        Find(&bookings)
    
    var total int64
    h.db.Model(&models.Booking{}).Count(&total)
    
    c.JSON(http.StatusOK, gin.H{
        "bookings": bookings,
        "total": total,
        "page": offset,
        "limit": limitInt,
    })
}

func (h *AdminHandler) GetAllUsers(c *gin.Context) {
    var users []models.User
    h.db.Order("created_at DESC").Find(&users)
    
    c.JSON(http.StatusOK, users)
}

func (h *AdminHandler) ManageRoom(c *gin.Context) {
    var room models.Room
    
    if c.Request.Method == "POST" {
        // Create new room
        if err := c.ShouldBindJSON(&room); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        
        h.db.Create(&room)
        c.JSON(http.StatusCreated, room)
    } else if c.Request.Method == "PUT" {
        // Update existing room
        roomID := c.Param("id")
        if err := c.ShouldBindJSON(&room); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        
        h.db.Model(&models.Room{}).Where("id = ?", roomID).Updates(room)
        c.JSON(http.StatusOK, room)
    }
}

func (h *AdminHandler) DeleteRoom(c *gin.Context) {
    roomID := c.Param("id")
    
    // Check if room has future bookings
    var count int64
    h.db.Model(&models.Booking{}).
        Where("room_id = ? AND start_time > ? AND status != 'cancelled'", roomID, time.Now()).
        Count(&count)
    
    if count > 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete room with future bookings"})
        return
    }
    
    h.db.Delete(&models.Room{}, roomID)
    c.JSON(http.StatusOK, gin.H{"message": "Room deleted successfully"})
}

func (h *AdminHandler) VerifyReview(c *gin.Context) {
    reviewID := c.Param("id")
    
    var review models.Review
    if err := h.db.First(&review, reviewID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
        return
    }
    
    review.IsVerified = true
    h.db.Save(&review)
    
    c.JSON(http.StatusOK, gin.H{"message": "Review verified successfully"})
}

func (h *AdminHandler) GetRevenueReport(c *gin.Context) {
    startDate := c.Query("start_date")
    endDate := c.Query("end_date")
    
    start, _ := time.Parse("2006-01-02", startDate)
    end, _ := time.Parse("2006-01-02", endDate)
    
    type DailyRevenue struct {
        Date  string  `json:"date"`
        Total float64 `json:"total"`
    }
    
    var revenues []DailyRevenue
    h.db.Table("payments").
        Select("DATE(created_at) as date, SUM(amount) as total").
        Where("status = 'success' AND created_at BETWEEN ? AND ?", start, end).
        Group("DATE(created_at)").
        Order("date ASC").
        Scan(&revenues)
    
    c.JSON(http.StatusOK, revenues)
}