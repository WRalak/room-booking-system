package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "backend/models"
)

type ReviewHandler struct {
    db *gorm.DB
}

type CreateReviewRequest struct {
    UserID    uint   `json:"user_id" binding:"required"`
    RoomID    uint   `json:"room_id" binding:"required"`
    BookingID uint   `json:"booking_id" binding:"required"`
    Rating    int    `json:"rating" binding:"required,min=1,max=5"`
    Comment   string `json:"comment" binding:"required"`
}

func NewReviewHandler(db *gorm.DB) *ReviewHandler {
    return &ReviewHandler{db: db}
}

func (h *ReviewHandler) CreateReview(c *gin.Context) {
    var req CreateReviewRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Verify user has completed booking
    var booking models.Booking
    if err := h.db.Where("id = ? AND user_id = ? AND status = 'confirmed' AND end_time < NOW()", 
        req.BookingID, req.UserID).First(&booking).Error; err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "You can only review completed bookings"})
        return
    }
    
    // Check if already reviewed
    var existingReview models.Review
    if err := h.db.Where("booking_id = ?", req.BookingID).First(&existingReview).Error; err == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "You have already reviewed this booking"})
        return
    }
    
    review := models.Review{
        UserID:    req.UserID,
        RoomID:    req.RoomID,
        BookingID: req.BookingID,
        Rating:    req.Rating,
        Comment:   req.Comment,
        IsVerified: false,
    }
    
    if err := h.db.Create(&review).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review"})
        return
    }
    
    c.JSON(http.StatusCreated, review)
}

func (h *ReviewHandler) GetRoomReviews(c *gin.Context) {
    roomID := c.Param("room_id")
    
    var reviews []models.Review
    h.db.Preload("User").Where("room_id = ? AND is_verified = ?", roomID, true).
        Order("created_at DESC").Find(&reviews)
    
    // Calculate average rating
    var avgRating float64
    h.db.Model(&models.Review{}).
        Where("room_id = ? AND is_verified = ?", roomID, true).
        Select("COALESCE(AVG(rating), 0)").
        Scan(&avgRating)
    
    c.JSON(http.StatusOK, gin.H{
        "reviews": reviews,
        "average_rating": avgRating,
        "total_reviews": len(reviews),
    })
}

func (h *ReviewHandler) GetUserReviews(c *gin.Context) {
    userID := c.Param("user_id")
    
    var reviews []models.Review
    h.db.Preload("Room").Where("user_id = ?", userID).
        Order("created_at DESC").Find(&reviews)
    
    c.JSON(http.StatusOK, reviews)
}