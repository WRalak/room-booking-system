package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "strconv"
    "sync"

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
    "github.com/gorilla/websocket"
    "gorm.io/gorm"
    "backend/models"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

type NotificationHandler struct {
    db         *gorm.DB
    redis      *redis.Client
    clients    map[uint]*websocket.Conn
    clientsMu  sync.RWMutex
}

func NewNotificationHandler(redis *redis.Client) *NotificationHandler {
    handler := &NotificationHandler{
        redis:   redis,
        clients: make(map[uint]*websocket.Conn),
    }
    
    // Start Redis subscriber for real-time notifications
    go handler.subscribeToNotifications()
    
    return handler
}

func (h *NotificationHandler) WebSocketHandler(c *gin.Context) {
    userIDStr := c.Param("user_id")
    userID, err := strconv.ParseUint(userIDStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    defer conn.Close()

    h.clientsMu.Lock()
    h.clients[uint(userID)] = conn
    h.clientsMu.Unlock()

    defer func() {
        h.clientsMu.Lock()
        delete(h.clients, uint(userID))
        h.clientsMu.Unlock()
    }()

    // Send unread notifications on connection
    var notifications []models.Notification
    h.db.Where("user_id = ? AND is_read = ?", uint(userID), false).Order("created_at desc").Find(&notifications)

    for _, notif := range notifications {
        h.sendToClient(uint(userID), notif)
    }

    // Keep connection alive
    for {
        if _, _, err := conn.ReadMessage(); err != nil {
            break
        }
    }
}

func (h *NotificationHandler) GetUserNotifications(c *gin.Context) {
    userID := c.Param("user_id")
    
    var notifications []models.Notification
    h.db.Where("user_id = ?", userID).Order("created_at desc").Limit(50).Find(&notifications)
    
    c.JSON(http.StatusOK, notifications)
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
    notificationID := c.Param("notification_id")
    
    h.db.Model(&models.Notification{}).Where("id = ?", notificationID).Update("is_read", true)
    
    c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

func (h *NotificationHandler) subscribeToNotifications() {
    pubsub := h.redis.Subscribe(context.Background(), "notifications")
    defer pubsub.Close()
    
    for {
        msg, err := pubsub.ReceiveMessage(context.Background())
        if err != nil {
            continue
        }
        
        var userID uint
        json.Unmarshal([]byte(msg.Payload), &userID)
        
        // Fetch new notifications and send to client
        var notifications []models.Notification
        h.db.Where("user_id = ? AND is_read = ?", userID, false).Order("created_at desc").Limit(1).Find(&notifications)
        
        if len(notifications) > 0 {
            h.sendToClient(userID, notifications[0])
        }
    }
}

func (h *NotificationHandler) sendToClient(userID uint, notification models.Notification) {
    h.clientsMu.RLock()
    conn, exists := h.clients[userID]
    h.clientsMu.RUnlock()
    
    if exists {
        conn.WriteJSON(notification)
    }
}