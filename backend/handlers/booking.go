package handlers

import (
	"fmt"
	"net/http"
	"time"

	"backend/models"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type BookingHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

type CreateBookingRequest struct {
	RoomID    uint      `json:"room_id" binding:"required"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
	UserID    uint      `json:"user_id" binding:"required"`
}

func NewBookingHandler(db *gorm.DB, redis *redis.Client) *BookingHandler {
	return &BookingHandler{
		db:    db,
		redis: redis,
	}
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if room exists
	var room models.Room
	if err := h.db.First(&room, req.RoomID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Check for overlapping bookings
	var existingBooking models.Booking
	err := h.db.Where("room_id = ? AND status != 'cancelled' AND ((start_time BETWEEN ? AND ?) OR (end_time BETWEEN ? AND ?))",
		req.RoomID, req.StartTime, req.EndTime, req.StartTime, req.EndTime).
		First(&existingBooking).Error

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Room already booked for this time slot"})
		return
	}

	// Calculate total amount (duration in hours)
	duration := req.EndTime.Sub(req.StartTime).Hours()
	totalAmount := float64(duration) * room.PricePerHour

	// Create booking
	booking := models.Booking{
		UserID:        req.UserID,
		RoomID:        req.RoomID,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		TotalAmount:   totalAmount,
		Status:        "pending",
		PaymentStatus: "unpaid",
	}

	if err := h.db.Create(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
		return
	}

	// Create booking confirmation notification
	notification := models.Notification{
		UserID:   req.UserID,
		Title:    "Booking Created",
		Message:  "Your booking has been created. Please complete payment to confirm.",
		Type:     "booking_confirmation",
		Metadata: fmt.Sprintf(`{"booking_id": %d}`, booking.ID),
	}
	h.db.Create(&notification)

	c.JSON(http.StatusCreated, booking)
}

func (h *BookingHandler) GetUserBookings(c *gin.Context) {
	userID := c.Param("user_id")

	var bookings []models.Booking
	h.db.Preload("Room").Where("user_id = ?", userID).Order("created_at desc").Find(&bookings)

	c.JSON(http.StatusOK, bookings)
}

func (h *BookingHandler) GetRoomBookings(c *gin.Context) {
	roomID := c.Param("room_id")

	var bookings []models.Booking
	h.db.Where("room_id = ? AND status != 'cancelled' AND start_time > ?", roomID, time.Now()).
		Order("start_time asc").Find(&bookings)

	c.JSON(http.StatusOK, bookings)
}

func (h *BookingHandler) GetAvailableRooms(c *gin.Context) {
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	var rooms []models.Room

	if startTime != "" && endTime != "" {
		// Get available rooms for specific time slot
		start, _ := time.Parse(time.RFC3339, startTime)
		end, _ := time.Parse(time.RFC3339, endTime)

		subQuery := h.db.Table("bookings").
			Select("room_id").
			Where("status != 'cancelled' AND ((start_time BETWEEN ? AND ?) OR (end_time BETWEEN ? AND ?))",
				start, end, start, end)

		h.db.Where("is_available = true").Not("id IN (?)", subQuery).Find(&rooms)
	} else {
		h.db.Where("is_available = true").Find(&rooms)
	}

	c.JSON(http.StatusOK, rooms)
}

func (h *BookingHandler) GetRoomDetails(c *gin.Context) {
	roomID := c.Param("room_id")

	var room models.Room
	if err := h.db.First(&room, roomID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	c.JSON(http.StatusOK, room)
}

func (h *BookingHandler) CancelBooking(c *gin.Context) {
	bookingID := c.Param("booking_id")

	var booking models.Booking
	if err := h.db.First(&booking, bookingID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	if booking.StartTime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot cancel past bookings"})
		return
	}

	booking.Status = "cancelled"
	h.db.Save(&booking)

	// Send cancellation notification
	notification := models.Notification{
		UserID:  booking.UserID,
		Title:   "Booking Cancelled",
		Message: "Your booking has been cancelled successfully.",
		Type:    "cancellation",
	}
	h.db.Create(&notification)

	c.JSON(http.StatusOK, gin.H{"message": "Booking cancelled successfully"})
}
