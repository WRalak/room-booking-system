import { useCallback, useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import toast from 'react-hot-toast';
import { useAuth } from '../context/useAuth';
import bookingService from '../services/bookingService';
import roomService from '../services/roomService';
import Spinner from './Spinner';
import gamificationService from '../services/gamificationService';

const BookingForm = ({ room: selectedRoom }) => {
  const { user } = useAuth();
  const { roomId } = useParams();
  const navigate = useNavigate();
  const [room, setRoom] = useState(selectedRoom || null);
  const [loading, setLoading] = useState(!selectedRoom);
  const [startTime, setStartTime] = useState('');
  const [durationHours, setDurationHours] = useState(1);
  const [notes, setNotes] = useState('');
  const [errors, setErrors] = useState({});

  const loadRoom = useCallback(async () => {
    try {
      const data = await roomService.getRoomDetails(roomId);
      setRoom(data);
    } catch {
      toast.error('Room not found');
    } finally {
      setLoading(false);
    }
  }, [roomId]);

  useEffect(() => {
    if (!selectedRoom) {
      const timer = window.setTimeout(() => {
        void loadRoom();
      }, 0);

      return () => window.clearTimeout(timer);
    }
  }, [loadRoom, selectedRoom]);

  if (loading) {
    return (
      <div className="text-center py-10">
        <Spinner size={48} />
        <div className="mt-2 text-gray-600">Loading room...</div>
      </div>
    );
  }

  if (!room) {
    return (
      <div className="bg-white rounded-lg shadow p-6 text-center">
        <h2 className="text-2xl font-semibold mb-4">Room not found</h2>
        <p className="text-gray-600 mb-6">Please choose a room from the rooms page.</p>
        <button
          onClick={() => navigate('/rooms')}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg"
        >
          Back to Rooms
        </button>
      </div>
    );
  }

  const totalAmount = durationHours * room.price_per_hour;

  const handleSubmit = async (event) => {
    event.preventDefault();

    const nextErrors = {};
    if (!startTime) nextErrors.startTime = 'Please select a start time.';
    if (!durationHours || durationHours < 1) nextErrors.durationHours = 'Duration must be at least 1 hour.';

    setErrors(nextErrors);
    if (Object.keys(nextErrors).length > 0) return;

    const startDate = new Date(startTime);
    const endDate = new Date(startDate.getTime() + durationHours * 60 * 60 * 1000);

    try {
      const booking = await bookingService.createBooking({
        user_id: user?.id,
        room_id: room.id,
        start_time: startDate.toISOString(),
        end_time: endDate.toISOString(),
        notes,
      });

      // Award gamification points: 1 point per 100 KES (minimum 1)
      const points = Math.max(1, Math.floor(totalAmount / 100));
      const uid = user?.id || localStorage.getItem('rb_user') || 'guest';
      gamificationService.addPoints(String(uid), points);
      gamificationService.grantBadge(String(uid), 'First Booking');

      toast.success('Booking created! Redirecting to payment...');
      // Navigate to payment page for this booking
      navigate(`/pay/${booking.id}`);
    } catch (error) {
      toast.error(error?.response?.data?.error || 'Failed to create booking');
    }
  };

  return (
    <div className="bg-white rounded-lg shadow p-6 max-w-3xl mx-auto">
      <h2 className="text-2xl font-bold mb-4">Book {room.name}</h2>
      <p className="text-gray-600 mb-6">{room.description}</p>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Start time</label>
          <input
            type="datetime-local"
            value={startTime}
            onChange={(event) => setStartTime(event.target.value)}
            className="w-full border rounded-lg px-3 py-2"
          />
          {errors.startTime && <p className="text-red-600 text-sm mt-1">{errors.startTime}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Duration (hours)</label>
          <input
            type="number"
            min="1"
            max="24"
            value={durationHours}
            onChange={(event) => setDurationHours(Number(event.target.value))}
            className="w-full border rounded-lg px-3 py-2"
          />
          {errors.durationHours && <p className="text-red-600 text-sm mt-1">{errors.durationHours}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Notes</label>
          <textarea
            value={notes}
            onChange={(event) => setNotes(event.target.value)}
            className="w-full border rounded-lg px-3 py-2"
            rows="4"
            placeholder="Optional notes for the booking"
          />
        </div>

        <div className="rounded-lg bg-gray-50 p-4">
          <p className="text-sm text-gray-600">Price per hour: KES {room.price_per_hour?.toLocaleString()}</p>
          <p className="text-lg font-semibold mt-2">Total: KES {totalAmount.toLocaleString()}</p>
        </div>

        <button
          type="submit"
          className="w-full bg-blue-600 text-white px-4 py-3 rounded-lg hover:bg-blue-700"
        >
          Confirm Booking
        </button>
      </form>
    </div>
  );
};

export default BookingForm;
