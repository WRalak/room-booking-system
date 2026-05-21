import { useState, useEffect, useCallback } from 'react';
import { Bell, Check } from 'lucide-react';
import { useAuth } from '../context/useAuth';
import notificationService from '../services/notificationService';

const NotificationCenter = () => {
  const { user, isAuthenticated } = useAuth();
  const [notifications, setNotifications] = useState([]);
  const [isOpen, setIsOpen] = useState(false);
  const [unreadCount, setUnreadCount] = useState(0);

  const loadNotifications = useCallback(async () => {
    if (!isAuthenticated || !user) return;

    try {
      const data = await notificationService.getNotifications(user.id);
      setNotifications(data);
      setUnreadCount(data.filter((n) => !n.is_read).length);
    } catch (error) {
      console.error('Error loading notifications:', error);
    }
  }, [isAuthenticated, user]);

  useEffect(() => {
    if (!isAuthenticated || !user) return undefined;

    const loadTimer = window.setTimeout(() => {
      void loadNotifications();
    }, 0);

    const ws = notificationService.connectWebSocket(user.id, (notification) => {
      setNotifications((prev) => [notification, ...prev]);
      setUnreadCount((prev) => prev + 1);

      if (Notification.permission === 'granted') {
        new Notification(notification.title, {
          body: notification.message,
          icon: '/favicon.ico',
        });
      }
    });

    if (Notification.permission === 'default') {
      Notification.requestPermission();
    }

    return () => {
      window.clearTimeout(loadTimer);
      if (ws) ws.close();
    };
  }, [isAuthenticated, user, loadNotifications]);

  const markAsRead = async (notificationId) => {
    await notificationService.markAsRead(notificationId);
    setNotifications((prev) =>
      prev.map((n) =>
        n.id === notificationId ? { ...n, is_read: true } : n
      )
    );
    setUnreadCount((prev) => Math.max(0, prev - 1));
  };

  const markAllAsRead = async () => {
    await Promise.all(
      notifications.filter((n) => !n.is_read).map((n) => markAsRead(n.id))
    );
  };

  const getNotificationIcon = (type) => {
    switch (type) {
      case 'payment_success':
        return '💰';
      case 'booking_confirmation':
        return '✅';
      case 'reminder':
        return '⏰';
      case 'cancellation':
        return '❌';
      default:
        return '📢';
    }
  };

  if (!isAuthenticated || !user) {
    return null;
  }

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="relative p-2 rounded-full hover:bg-gray-100"
      >
        <Bell className="w-6 h-6" />
        {unreadCount > 0 && (
          <span className="absolute top-0 right-0 inline-flex items-center justify-center px-2 py-1 text-xs font-bold leading-none text-white transform translate-x-1/2 -translate-y-1/2 bg-red-600 rounded-full">
            {unreadCount}
          </span>
        )}
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-96 bg-white rounded-lg shadow-xl z-50">
          <div className="p-4 border-b flex justify-between items-center">
            <h3 className="text-lg font-semibold">Notifications</h3>
            {unreadCount > 0 && (
              <button
                type="button"
                onClick={markAllAsRead}
                className="text-sm text-blue-600 hover:text-blue-800"
              >
                Mark all as read
              </button>
            )}
          </div>

          <div className="max-h-96 overflow-y-auto">
            {notifications.length === 0 ? (
              <div className="p-8 text-center text-gray-500">No notifications yet</div>
            ) : (
              notifications.map((notification) => (
                <div
                  key={notification.id}
                  className={`p-4 border-b hover:bg-gray-50 cursor-pointer ${
                    !notification.is_read ? 'bg-blue-50' : ''
                  }`}
                  onClick={() => markAsRead(notification.id)}
                >
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <div className="flex items-center mb-1">
                        <span className="text-xl mr-2">{getNotificationIcon(notification.type)}</span>
                        <span className="font-semibold">{notification.title}</span>
                      </div>
                      <p className="text-gray-600 text-sm">{notification.message}</p>
                      <span className="text-xs text-gray-400 mt-1">{new Date(notification.created_at).toLocaleString()}</span>
                    </div>
                    {!notification.is_read && <Check className="w-4 h-4 text-blue-600" />}
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default NotificationCenter;
