import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import roomService from '../services/roomService';
import Spinner from './Spinner';

const RoomList = ({ onSelectRoom }) => {
  const navigate = useNavigate();
  const [rooms, setRooms] = useState([]);
  const [loading, setLoading] = useState(true);

  const loadRooms = useCallback(async () => {
    try {
      const data = await roomService.getAvailableRooms();
      setRooms(data);
    } catch {
      toast.error('Failed to load rooms');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadRooms();
    }, 0);

    return () => window.clearTimeout(timer);
  }, [loadRooms]);

  const handleBook = (room) => {
    if (onSelectRoom) onSelectRoom(room);
    navigate(`/book/${room.id}`);
  };

  if (loading) {
    return (
      <div className="text-center py-10">
        <Spinner size={48} />
        <div className="mt-2 text-gray-600">Loading rooms...</div>
      </div>
    );
  }

  if (rooms.length === 0) {
    return <div className="bg-white rounded-lg shadow p-6 text-gray-600">No rooms are available.</div>;
  }

  return (
    <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3" role="list">
      {rooms.map((room) => (
        <article key={room.id} role="listitem" aria-labelledby={`room-${room.id}`} className="bg-white rounded-lg shadow p-6">
          <h2 id={`room-${room.id}`} className="text-xl font-semibold mb-2">{room.name}</h2>
          <p className="text-gray-600 mb-4">{room.description}</p>
          <div className="mb-4">
            <span className="font-medium">Capacity:</span> {room.capacity}
          </div>
          <div className="mb-4">
            <span className="font-medium">Price:</span> KES {room.price_per_hour?.toLocaleString()} / hour
          </div>
          <button
            onClick={() => handleBook(room)}
            aria-label={`Book ${room.name}`}
            className="w-full bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700"
          >
            Book this room
          </button>
        </article>
      ))}
    </div>
  );
};

export default RoomList;
