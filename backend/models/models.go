package models

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"unique;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Name      string    `json:"name"` // Full name
	Phone     string    `json:"phone"`
	IsAdmin   bool      `gorm:"default:false" json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Room struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"not null" json:"name"`
	Description   string    `gorm:"type:text" json:"description"`
	Capacity      int       `json:"capacity"`
	PricePerNight float64   `json:"price_per_night"`
	PricePerHour  float64   `json:"price_per_hour"`
	IsAvailable   bool      `gorm:"default:true" json:"is_available"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Booking struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null" json:"user_id"`
	RoomID        uint      `gorm:"not null" json:"room_id"`
	CheckIn       time.Time `json:"check_in"`
	CheckOut      time.Time `json:"check_out"`
	StartTime     time.Time `json:"start_time"`                      // alias for CheckIn
	EndTime       time.Time `json:"end_time"`                        // alias for CheckOut
	Status        string    `gorm:"default:'pending'" json:"status"` // pending, confirmed, cancelled
	TotalCost     float64   `json:"total_cost"`
	TotalAmount   float64   `json:"total_amount"` // alias for TotalCost
	PaymentStatus string    `json:"payment_status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user"`
	Room Room `gorm:"foreignKey:RoomID" json:"room"`
}

type Payment struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	BookingID     uint      `gorm:"not null" json:"booking_id"`
	Amount        float64   `gorm:"not null" json:"amount"`
	Currency      string    `gorm:"default:'KES'" json:"currency"`
	Status        string    `gorm:"default:'pending'" json:"status"` // pending, completed, failed
	PaymentMethod string    `json:"payment_method"`
	Reference     string    `gorm:"unique" json:"reference"`
	ProviderRef   string    `json:"provider_ref"` // Reference from payment provider (Paystack, M-Pesa)
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	Booking Booking `gorm:"foreignKey:BookingID" json:"booking"`
}

type Notification struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"not null"`
	Title     string
	Message   string `gorm:"type:text"`
	Type      string // email, sms, push
	Metadata  string `gorm:"type:json"` // Additional data as JSON
	IsRead    bool   `gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time

	User User `gorm:"foreignKey:UserID"`
}

type Review struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `json:"user_id"`
	RoomID     uint      `json:"room_id"`
	BookingID  uint      `json:"booking_id"`
	Rating     int       `gorm:"check:rating >= 1 AND rating <= 5" json:"rating"`
	Comment    string    `gorm:"type:text" json:"comment"`
	IsVerified bool      `gorm:"default:false" json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	User    User    `gorm:"foreignKey:UserID" json:"user"`
	Room    Room    `gorm:"foreignKey:RoomID" json:"room"`
	Booking Booking `gorm:"foreignKey:BookingID" json:"booking"`
}

type RecurringBooking struct {
	ID             uint `gorm:"primaryKey"`
	BookingID      uint // Original booking template
	UserID         uint
	RoomID         uint
	Frequency      string // daily, weekly, monthly
	Interval       int    // every X days/weeks/months
	DayOfWeek      *int   // for weekly: 0-6 (Sunday-Saturday)
	DayOfMonth     *int   // for monthly: 1-31
	StartDate      time.Time
	EndDate        *time.Time
	NextOccurrence time.Time
	IsActive       bool `gorm:"default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Booking Booking `gorm:"foreignKey:BookingID"`
	User    User    `gorm:"foreignKey:UserID"`
	Room    Room    `gorm:"foreignKey:RoomID"`
}

type Invoice struct {
	ID            uint   `gorm:"primaryKey"`
	InvoiceNumber string `gorm:"unique;not null"`
	BookingID     uint
	UserID        uint
	Amount        float64
	Tax           float64
	TotalAmount   float64
	Currency      string `gorm:"default:'KES'"`
	Status        string // paid, pending, overdue, cancelled
	PaymentDate   *time.Time
	DueDate       time.Time
	InvoicePDF    string // URL to PDF file
	CreatedAt     time.Time
	UpdatedAt     time.Time

	Booking Booking `gorm:"foreignKey:BookingID"`
	User    User    `gorm:"foreignKey:UserID"`
}

type AuditLog struct {
	ID        uint  `gorm:"primaryKey"`
	UserID    *uint // NULL for system actions
	Action    string
	Entity    string // booking, user, room, payment
	EntityID  string
	OldValue  string `gorm:"type:json"`
	NewValue  string `gorm:"type:json"`
	IPAddress string
	UserAgent string
	CreatedAt time.Time
}

func (i *Invoice) BeforeCreate(tx *gorm.DB) error {
	i.InvoiceNumber = fmt.Sprintf("INV-%d-%s", time.Now().Unix(), uuid.New().String()[:8])
	return nil
}
