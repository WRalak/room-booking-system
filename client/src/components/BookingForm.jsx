import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import toast from 'react-hot-toast';

const rooms = [
  {
    id: 1,
    name: 'Executive Meeting Room',
    description: 'A fully equipped meeting room for up to 12 people.',
    pricePerHour: 2500,
  },
  {
    id: 2,
    name: 'Conference Hall',
    description: 'Spacious hall with projector, sound system and seating for 30.',
    pricePerHour: 4000,
  },
  {
    id: 3,
    name: 'Quiet Workspace',
    description: 'Individual work pods with high-speed Wi-Fi and power outlets.',
    pricePerHour: 1200,
  },
];

const BookingForm = ({ room: selectedRoom, user }) => {
  const { roomId } = useParams();
  const navigate = useNavigate();
  const [room] = useState(() => selectedRoom || rooms.find((item) => String(item.id) === roomId) || null);
  const [startTime, setStartTime] = useState('');
  const [durationHours, setDurationHours] = useState(1);
  const [notes, setNotes] = useState('');

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

  const totalAmount = durationHours * room.pricePerHour;

  const handleSubmit = (event) => {
    event.preventDefault();

    if (!startTime) {
      toast.error('Please select a start time.');
      return;
    }

    const booking = {
      id: Date.now(),
      userId: user?.id,
      roomId: room.id,
      roomName: room.name,
      startTime,
      durationHours,
      totalAmount,
      notes,
    };

    const storedBookings = JSON.parse(localStorage.getItem('bookings') || '[]');
    localStorage.setItem('bookings', JSON.stringify([...storedBookings, booking]));

    toast.success('Booking created successfully!');
    navigate('/my-bookings');
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
          <p className="text-sm text-gray-600">Price per hour: KES {room.pricePerHour.toLocaleString()}</p>
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
