package models

import (
    "fmt"
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type User struct {
    ID        uint      `gorm:"primaryKey"`
    Email     string    `gorm:"unique;not null"`
    Password  string    `gorm:"not null"`
    FirstName string
    LastName  string
    Name      string    // Full name
    Phone     string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Room struct {
    ID          uint      `gorm:"primaryKey"`
    Name        string    `gorm:"not null"`
    Description string    `gorm:"type:text"`
    Capacity    int
    PricePerNight float64
    PricePerHour float64
    IsAvailable bool      `gorm:"default:true"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Booking struct {
    ID             uint      `gorm:"primaryKey"`
    UserID         uint      `gorm:"not null"`
    RoomID         uint      `gorm:"not null"`
    CheckIn        time.Time `gorm:"not null"`
    CheckOut       time.Time `gorm:"not null"`
    StartTime      time.Time // alias for CheckIn
    EndTime        time.Time // alias for CheckOut
    Status         string    `gorm:"default:'pending'"` // pending, confirmed, cancelled
    TotalCost      float64
    TotalAmount    float64   // alias for TotalCost
    PaymentStatus  string
    CreatedAt      time.Time
    UpdatedAt      time.Time

    User User `gorm:"foreignKey:UserID"`
    Room Room `gorm:"foreignKey:RoomID"`
}

type Payment struct {
    ID            uint      `gorm:"primaryKey"`
    BookingID     uint      `gorm:"not null"`
    Amount        float64   `gorm:"not null"`
    Currency      string    `gorm:"default:'KES'"`
    Status        string    `gorm:"default:'pending'"` // pending, completed, failed
    PaymentMethod string
    Reference     string    `gorm:"unique"`
    ProviderRef   string    // Reference from payment provider (Paystack, M-Pesa)
    CreatedAt     time.Time
    UpdatedAt     time.Time

    Booking Booking `gorm:"foreignKey:BookingID"`
}

type Notification struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    uint      `gorm:"not null"`
    Title     string
    Message   string    `gorm:"type:text"`
    Type      string    // email, sms, push
    Metadata  string    `gorm:"type:json"` // Additional data as JSON
    IsRead    bool      `gorm:"default:false"`
    CreatedAt time.Time
    UpdatedAt time.Time

    User User `gorm:"foreignKey:UserID"`
}

type Review struct {
    ID          uint      `gorm:"primaryKey"`
    UserID      uint
    RoomID      uint
    BookingID   uint
    Rating      int       `gorm:"check:rating >= 1 AND rating <= 5"`
    Comment     string    `gorm:"type:text"`
    IsVerified  bool      `gorm:"default:false"`
    CreatedAt   time.Time
    UpdatedAt   time.Time

    User    User    `gorm:"foreignKey:UserID"`
    Room    Room    `gorm:"foreignKey:RoomID"`
    Booking Booking `gorm:"foreignKey:BookingID"`
}

type RecurringBooking struct {
    ID              uint      `gorm:"primaryKey"`
    BookingID       uint      // Original booking template
    UserID          uint
    RoomID          uint
    Frequency       string    // daily, weekly, monthly
    Interval        int       // every X days/weeks/months
    DayOfWeek       *int      // for weekly: 0-6 (Sunday-Saturday)
    DayOfMonth      *int      // for monthly: 1-31
    StartDate       time.Time
    EndDate         *time.Time
    NextOccurrence  time.Time
    IsActive        bool      `gorm:"default:true"`
    CreatedAt       time.Time
    UpdatedAt       time.Time

    Booking Booking `gorm:"foreignKey:BookingID"`
    User    User    `gorm:"foreignKey:UserID"`
    Room    Room    `gorm:"foreignKey:RoomID"`
}

type Invoice struct {
    ID             uint      `gorm:"primaryKey"`
    InvoiceNumber  string    `gorm:"unique;not null"`
    BookingID      uint
    UserID         uint
    Amount         float64
    Tax            float64
    TotalAmount    float64
    Currency       string    `gorm:"default:'KES'"`
    Status         string    // paid, pending, overdue, cancelled
    PaymentDate    *time.Time
    DueDate        time.Time
    InvoicePDF     string    // URL to PDF file
    CreatedAt      time.Time
    UpdatedAt      time.Time

    Booking Booking `gorm:"foreignKey:BookingID"`
    User    User    `gorm:"foreignKey:UserID"`
}

type AuditLog struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    *uint     // NULL for system actions
    Action    string
    Entity    string    // booking, user, room, payment
    EntityID  string
    OldValue  string    `gorm:"type:json"`
    NewValue  string    `gorm:"type:json"`
    IPAddress string
    UserAgent string
    CreatedAt time.Time
}

func (i *Invoice) BeforeCreate(tx *gorm.DB) error {
    i.InvoiceNumber = fmt.Sprintf("INV-%d-%s", time.Now().Unix(), uuid.New().String()[:8])
    return nil
}