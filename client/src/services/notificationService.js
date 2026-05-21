import apiClient, { getWebSocketUrl } from './apiClient';

class NotificationService {
  async getNotifications(userId) {
    const response = await apiClient.get(`/notifications/user/${userId}`);
    return response.data;
  }

  async markAsRead(notificationId) {
    const response = await apiClient.post(`/notifications/mark-read/${notificationId}`);
    return response.data;
  }

  connectWebSocket(userId, onMessage) {
    const socket = new WebSocket(getWebSocketUrl(`/notifications/ws/${userId}`));
    socket.onmessage = (event) => {
      onMessage(JSON.parse(event.data));
    };

    return {
      close: () => socket.close(),
    };
  }
}

export default new NotificationService();
