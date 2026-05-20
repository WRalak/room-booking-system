package handlers

import (
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jung-kurt/gofpdf"
    "gorm.io/gorm"
    "backend/models"
)

type InvoiceHandler struct {
    db *gorm.DB
}

func NewInvoiceHandler(db *gorm.DB) *InvoiceHandler {
    return &InvoiceHandler{db: db}
}

func (h *InvoiceHandler) GenerateInvoice(c *gin.Context) {
    bookingID := c.Param("booking_id")
    
    // Get booking details
    var booking models.Booking
    if err := h.db.Preload("User").Preload("Room").First(&booking, bookingID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
        return
    }
    
    // Get payment details
    var payment models.Payment
    h.db.Where("booking_id = ? AND status = ?", bookingID, "success").First(&payment)
    
    // Check if invoice already exists
    var existingInvoice models.Invoice
    if err := h.db.Where("booking_id = ?", bookingID).First(&existingInvoice).Error; err == nil {
        c.JSON(http.StatusOK, existingInvoice)
        return
    }
    
    // Calculate tax (16% VAT for Kenya)
    taxRate := 0.16
    tax := booking.TotalAmount * taxRate
    totalAmount := booking.TotalAmount + tax
    
    // Create invoice record
    invoice := models.Invoice{
        BookingID:   booking.ID,
        UserID:      booking.UserID,
        Amount:      booking.TotalAmount,
        Tax:         tax,
        TotalAmount: totalAmount,
        Currency:    "KES",
        Status:      "pending",
        DueDate:     booking.StartTime.AddDate(0, 0, -7), // Due 7 days before booking
    }
    
    h.db.Create(&invoice)
    
    // Generate PDF
    pdfPath, err := h.generatePDFInvoice(invoice, booking)
    if err == nil {
        invoice.InvoicePDF = pdfPath
        h.db.Save(&invoice)
    }
    
    c.JSON(http.StatusCreated, invoice)
}

func (h *InvoiceHandler) generatePDFInvoice(invoice models.Invoice, booking models.Booking) (string, error) {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    
    // Company header
    pdf.SetFont("Arial", "B", 20)
    pdf.Cell(0, 10, "RoomBooker Invoice")
    pdf.Ln(10)
    
    pdf.SetFont("Arial", "", 12)
    pdf.Cell(0, 10, fmt.Sprintf("Invoice Number: %s", invoice.InvoiceNumber))
    pdf.Ln(6)
    pdf.Cell(0, 10, fmt.Sprintf("Date: %s", time.Now().Format("2006-01-02")))
    pdf.Ln(6)
    pdf.Cell(0, 10, fmt.Sprintf("Due Date: %s", invoice.DueDate.Format("2006-01-02")))
    pdf.Ln(15)
    
    // Customer details
    pdf.SetFont("Arial", "B", 14)
    pdf.Cell(0, 10, "Bill To:")
    pdf.Ln(8)
    pdf.SetFont("Arial", "", 12)
    pdf.Cell(0, 10, booking.User.Name)
    pdf.Ln(6)
    pdf.Cell(0, 10, booking.User.Email)
    pdf.Ln(6)
    if booking.User.Phone != "" {
        pdf.Cell(0, 10, booking.User.Phone)
        pdf.Ln(6)
    }
    pdf.Ln(10)
    
    // Booking details
    pdf.SetFont("Arial", "B", 14)
    pdf.Cell(0, 10, "Booking Details:")
    pdf.Ln(8)
    pdf.SetFont("Arial", "", 12)
    pdf.Cell(0, 10, fmt.Sprintf("Room: %s", booking.Room.Name))
    pdf.Ln(6)
    pdf.Cell(0, 10, fmt.Sprintf("Date: %s", booking.StartTime.Format("2006-01-02")))
    pdf.Ln(6)
    pdf.Cell(0, 10, fmt.Sprintf("Time: %s - %s", 
        booking.StartTime.Format("15:04"), 
        booking.EndTime.Format("15:04")))
    pdf.Ln(6)
    duration := booking.EndTime.Sub(booking.StartTime).Hours()
    pdf.Cell(0, 10, fmt.Sprintf("Duration: %.1f hours", duration))
    pdf.Ln(15)
    
    // Amount breakdown
    pdf.SetFont("Arial", "B", 14)
    pdf.Cell(0, 10, "Amount Breakdown:")
    pdf.Ln(8)
    pdf.SetFont("Arial", "", 12)
    pdf.Cell(100, 10, "Subtotal:")
    pdf.Cell(0, 10, fmt.Sprintf("KES %.2f", invoice.Amount))
    pdf.Ln(8)
    pdf.Cell(100, 10, "Tax (16% VAT):")
    pdf.Cell(0, 10, fmt.Sprintf("KES %.2f", invoice.Tax))
    pdf.Ln(8)
    pdf.SetFont("Arial", "B", 12)
    pdf.Cell(100, 10, "Total:")
    pdf.Cell(0, 10, fmt.Sprintf("KES %.2f", invoice.TotalAmount))
    pdf.Ln(15)
    
    // Payment instructions
    pdf.SetFont("Arial", "B", 12)
    pdf.Cell(0, 10, "Payment Instructions:")
    pdf.Ln(8)
    pdf.SetFont("Arial", "", 10)
    pdf.MultiCell(0, 5,
        "Please make payment via M-Pesa Paybill or Credit Card through our payment portal.\n"+
        "M-Pesa Paybill Number: 123456\n"+
        "Account Number: "+invoice.InvoiceNumber, "", "", false)
    
    // Save PDF
    filename := fmt.Sprintf("invoices/invoice_%s.pdf", invoice.InvoiceNumber)
    err := pdf.OutputFileAndClose(filename)
    if err != nil {
        return "", err
    }
    
    return filename, nil
}

func (h *InvoiceHandler) DownloadInvoice(c *gin.Context) {
    invoiceID := c.Param("id")
    
    var invoice models.Invoice
    if err := h.db.First(&invoice, invoiceID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
        return
    }
    
    c.File(invoice.InvoicePDF)
}

func (h *InvoiceHandler) GetUserInvoices(c *gin.Context) {
    userID := c.Param("user_id")
    
    var invoices []models.Invoice
    h.db.Preload("Booking").Preload("Booking.Room").
        Where("user_id = ?", userID).
        Order("created_at DESC").
        Find(&invoices)
    
    c.JSON(http.StatusOK, invoices)
}