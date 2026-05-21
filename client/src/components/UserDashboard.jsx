import { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/useAuth';
import bookingService from '../services/bookingService';
import notificationService from '../services/notificationService';
import gamificationService from '../services/gamificationService';
import Spinner from './Spinner';

const UserDashboard = () => {
  const { user } = useAuth();
  const [bookings, setBookings] = useState([]);
  const [notifications, setNotifications] = useState([]);
  const [loading, setLoading] = useState(true);

  const loadData = useCallback(async () => {
    if (!user) return;
    setLoading(true);

    try {
      const [bookingData, notificationData] = await Promise.all([
        bookingService.getUserBookings(user.id),
        notificationService.getNotifications(user.id),
      ]);
      setBookings(bookingData || []);
      setNotifications(notificationData || []);
    } catch (error) {
      console.error('Failed to load dashboard data', error);
    } finally {
      setLoading(false);
    }
  }, [user]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  const points = user ? gamificationService.getPoints(String(user.id)) : 0;
  const badges = user ? gamificationService.getBadges(String(user.id)) : [];

  if (loading) {
    return (
      <div className="text-center py-10">
        <Spinner size={48} />
        <div className="mt-2 text-gray-600">Loading your dashboard...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <section className="bg-white rounded-lg shadow p-6">
        <h1 className="text-3xl font-bold mb-2">Welcome back, {user?.name || 'Guest'}</h1>
        <p className="text-gray-600">This is your personal dashboard for bookings, notifications, and rewards.</p>
      </section>

      <section className="grid gap-4 md:grid-cols-3">
        <div className="bg-white rounded-lg shadow p-6">
          <p className="text-sm text-gray-500">Points</p>
          <p className="text-4xl font-bold">{points}</p>
        </div>
        <div className="bg-white rounded-lg shadow p-6">
          <p className="text-sm text-gray-500">Badges</p>
          <div className="mt-3 space-y-2">
            {badges.length === 0 ? (
              <p className="text-gray-500">No badges earned yet.</p>
            ) : (
              badges.map((badge) => (
                <div key={badge} className="inline-block rounded-full bg-blue-50 px-3 py-1 text-sm text-blue-700">
                  {badge}
                </div>
              ))
            )}
          </div>
        </div>
        <div className="bg-white rounded-lg shadow p-6">
          <p className="text-sm text-gray-500">Quick Actions</p>
          <div className="mt-4 space-y-3">
            <Link to="/my-bookings" className="block bg-blue-600 text-white px-4 py-3 rounded-lg text-center">My Bookings</Link>
            <Link to="/notifications" className="block bg-gray-100 text-gray-700 px-4 py-3 rounded-lg text-center">View Notifications</Link>
            <Link to="/leaderboard" className="block bg-gray-100 text-gray-700 px-4 py-3 rounded-lg text-center">Leaderboard</Link>
          </div>
        </div>
      </section>

      <section className="grid gap-6 lg:grid-cols-2">
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-semibold">Upcoming Bookings</h2>
            <Link to="/my-bookings" className="text-blue-600 hover:text-blue-800 text-sm">View all</Link>
          </div>
          {bookings.length === 0 ? (
            <p className="text-gray-500">You have no active bookings.</p>
          ) : (
            <ul className="space-y-3">
              {bookings.slice(0, 5).map((booking) => (
                <li key={booking.id} className="border rounded-lg p-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="font-semibold">{booking.room?.name || booking.room_name || 'Room'}</p>
                      <p className="text-sm text-gray-500">{new Date(booking.start_time).toLocaleString()}</p>
                    </div>
                    <span className="text-sm text-gray-700">KES {booking.total_amount?.toLocaleString()}</span>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-semibold">Recent Notifications</h2>
            <Link to="/notifications" className="text-blue-600 hover:text-blue-800 text-sm">View all</Link>
          </div>
          {notifications.length === 0 ? (
            <p className="text-gray-500">No notifications yet.</p>
          ) : (
            <ul className="space-y-3">
              {notifications.slice(0, 5).map((notification) => (
                <li key={notification.id} className={`rounded-lg p-3 ${notification.is_read ? 'bg-gray-50' : 'bg-blue-50'}`}>
                  <p className="font-semibold">{notification.title}</p>
                  <p className="text-sm text-gray-600">{notification.message}</p>
                  <p className="text-xs text-gray-400 mt-2">{new Date(notification.created_at).toLocaleString()}</p>
                </li>
              ))}
            </ul>
          )}
        </div>
      </section>
    </div>
  );
};

export default UserDashboard;
