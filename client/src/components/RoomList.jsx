import React from 'react';
import { useNavigate } from 'react-router-dom';

const rooms = [
  {
    id: 1,
    name: 'Executive Meeting Room',
    description: 'A fully equipped meeting room for up to 12 people.',
    pricePerHour: 2500,
    capacity: 12,
  },
  {
    id: 2,
    name: 'Conference Hall',
    description: 'Spacious hall with projector, sound system and seating for 30.',
    pricePerHour: 4000,
    capacity: 30,
  },
  {
    id: 3,
    name: 'Quiet Workspace',
    description: 'Individual work pods with high-speed Wi-Fi and power outlets.',
    pricePerHour: 1200,
    capacity: 4,
  },
];

const RoomList = ({ onSelectRoom }) => {
  const navigate = useNavigate();

  const handleBook = (room) => {
    if (onSelectRoom) onSelectRoom(room);
    navigate(`/book/${room.id}`);
  };

  return (
    <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      {rooms.map((room) => (
        <div key={room.id} className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-2">{room.name}</h2>
          <p className="text-gray-600 mb-4">{room.description}</p>
          <div className="mb-4">
            <span className="font-medium">Capacity:</span> {room.capacity}
          </div>
          <div className="mb-4">
            <span className="font-medium">Price:</span> KES {room.pricePerHour.toLocaleString()} / hour
          </div>
          <button
            onClick={() => handleBook(room)}
            className="w-full bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700"
          >
            Book this room
          </button>
        </div>
      ))}
    </div>
  );
};

export default RoomList;
