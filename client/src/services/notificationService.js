class NotificationService {
  async getNotifications(userId) {
    return Promise.resolve([
      {
        id: 1,
        userId,
        title: 'Welcome to RoomBooker',
        message: 'Your notification center is set up and ready.',
        type: 'info',
        is_read: false,
        created_at: new Date().toISOString(),
      },
    ]);
  }

  async markAsRead(notificationId) {
    return Promise.resolve({ success: true });
  }

  connectWebSocket(userId, onMessage) {
    return {
      close: () => {},
    };
  }
}

export default new NotificationService();
