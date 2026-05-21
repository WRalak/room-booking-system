import { useCallback, useEffect, useState } from 'react';
import { useAuth } from '../context/useAuth';
import notificationService from '../services/notificationService';
import Spinner from './Spinner';
import toast from 'react-hot-toast';

const NotificationPage = () => {
  const { user } = useAuth();
  const [notifications, setNotifications] = useState([]);
  const [loading, setLoading] = useState(true);

  const loadNotifications = useCallback(async () => {
    if (!user) return;
    setLoading(true);

    try {
      const response = await notificationService.getNotifications(user.id);
      setNotifications(response || []);
    } catch (error) {
      toast.error('Failed to load notifications');
      console.error(error);
    } finally {
      setLoading(false);
    }
  }, [user]);

  useEffect(() => {
    void loadNotifications();
  }, [loadNotifications]);

  const markRead = async (notificationId) => {
    try {
      await notificationService.markAsRead(notificationId);
      setNotifications((prev) => prev.map((notification) => (
        notification.id === notificationId ? { ...notification, is_read: true } : notification
      )));
    } catch {
      toast.error('Unable to mark notification as read');
    }
  };

  if (loading) {
    return (
      <div className="text-center py-10">
        <Spinner size={48} />
        <div className="mt-2 text-gray-600">Loading notifications...</div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h1 className="text-3xl font-bold mb-4">Notifications</h1>
      {notifications.length === 0 ? (
        <p className="text-gray-600">You have no notifications at the moment.</p>
      ) : (
        <div className="space-y-4">
          {notifications.map((notification) => (
            <div
              key={notification.id}
              className={`rounded-lg p-4 ${notification.is_read ? 'bg-gray-50' : 'bg-blue-50'}`}
            >
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="font-semibold">{notification.title}</p>
                  <p className="text-sm text-gray-700">{notification.message}</p>
                  <p className="text-xs text-gray-400 mt-2">{new Date(notification.created_at).toLocaleString()}</p>
                </div>
                {!notification.is_read && (
                  <button
                    type="button"
                    onClick={() => markRead(notification.id)}
                    className="rounded-md bg-blue-600 px-3 py-2 text-sm text-white hover:bg-blue-700"
                  >
                    Mark read
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default NotificationPage;
