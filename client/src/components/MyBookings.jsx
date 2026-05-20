import { useMemo } from 'react';

const MyBookings = ({ userId }) => {
  const bookings = useMemo(() => {
    const stored = localStorage.getItem('bookings');
    if (!stored) return [];
    return JSON.parse(stored).filter((booking) => booking.userId === userId);
  }, [userId]);

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
                  <h2 className="text-xl font-semibold">{booking.roomName}</h2>
                  <p className="text-sm text-gray-500">Booking ID: {booking.id}</p>
                </div>
                <span className="text-sm text-gray-600">{new Date(booking.startTime).toLocaleString()}</span>
              </div>
              <p>{booking.notes || 'No additional notes.'}</p>
              <div className="mt-3 text-sm text-gray-700">
                <span className="font-medium">Duration:</span> {booking.durationHours} hour(s)
              </div>
              <div className="mt-2 text-sm text-gray-700">
                <span className="font-medium">Total:</span> KES {booking.totalAmount.toLocaleString()}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default MyBookings;
