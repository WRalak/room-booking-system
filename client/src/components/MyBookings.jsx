import { useCallback, useEffect, useState } from 'react';
import { useAuth } from '../context/useAuth';
import toast from 'react-hot-toast';
import bookingService from '../services/bookingService';
import Spinner from './Spinner';

const MyBookings = () => {
  const { user } = useAuth();
  const [bookings, setBookings] = useState([]);
  const [loading, setLoading] = useState(true);

  const loadBookings = useCallback(async () => {
    if (!user) return;

    try {
      const data = await bookingService.getUserBookings(user.id);
      setBookings(data);
    } catch {
      toast.error('Failed to load bookings');
    } finally {
      setLoading(false);
    }
  }, [user]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadBookings();
    }, 0);

    return () => window.clearTimeout(timer);
  }, [loadBookings]);

  const cancelBooking = async (bookingId) => {
    try {
      await bookingService.cancelBooking(bookingId);
      toast.success('Booking cancelled');
      await loadBookings();
    } catch (error) {
      toast.error(error.response?.data?.error || 'Failed to cancel booking');
    }
  };

  if (loading) {
    return (
      <div className="text-center py-10">
        <Spinner size={48} />
        <div className="mt-2 text-gray-600">Loading bookings...</div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h1 className="text-3xl font-bold mb-6">My Bookings</h1>
      {bookings.length === 0 ? (
        <div className="text-gray-600">You have no bookings yet.</div>
      ) : (
        <div className="space-y-4">
          {bookings.map((booking) => (
            <div key={booking.id} className="border rounded-lg p-4">
              <div className="flex justify-between items-center mb-3">
                <div>
                  <h2 className="text-xl font-semibold">{booking.room?.name}</h2>
                  <p className="text-sm text-gray-500">Booking ID: {booking.id}</p>
                </div>
                <span className="text-sm text-gray-600">{new Date(booking.start_time).toLocaleString()}</span>
              </div>
              <div className="mt-3 text-sm text-gray-700">
                <span className="font-medium">Status:</span> {booking.status}
              </div>
              <div className="mt-2 text-sm text-gray-700">
                <span className="font-medium">Total:</span> KES {booking.total_amount?.toLocaleString()}
              </div>
              {booking.status !== 'cancelled' && (
                <button
                  type="button"
                  onClick={() => cancelBooking(booking.id)}
                  className="mt-4 bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700"
                >
                  Cancel booking
                </button>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default MyBookings;
