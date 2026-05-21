import { useCallback, useEffect, useState } from 'react';
import { Plus, Edit, Trash2 } from 'lucide-react';
import adminService from '../../services/adminService';
import toast from 'react-hot-toast';

const RoomManagement = () => {
  const [rooms, setRooms] = useState([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingRoom, setEditingRoom] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    price_per_hour: '',
    capacity: '',
    amenities: [],
    images: []
  });

  const loadRooms = useCallback(async () => {
    try {
      const data = await adminService.getAllRooms();
      setRooms(data);
    } catch {
      toast.error('Failed to load rooms');
    }
  }, []);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadRooms();
    }, 0);

    return () => window.clearTimeout(timer);
  }, [loadRooms]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      if (editingRoom) {
        await adminService.updateRoom(editingRoom.id, formData);
        toast.success('Room updated successfully');
      } else {
        await adminService.createRoom(formData);
        toast.success('Room created successfully');
      }
      loadRooms();
      setIsModalOpen(false);
      resetForm();
    } catch {
      toast.error('Operation failed');
    }
  };

  const handleDelete = async (roomId) => {
    if (window.confirm('Are you sure you want to delete this room?')) {
      try {
        await adminService.deleteRoom(roomId);
        toast.success('Room deleted successfully');
        loadRooms();
      } catch (error) {
        toast.error(error.response?.data?.error || 'Failed to delete room');
      }
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      price_per_hour: '',
      capacity: '',
      amenities: [],
      images: []
    });
    setEditingRoom(null);
  };

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Room Management</h1>
        <button
          onClick={() => {
            resetForm();
            setIsModalOpen(true);
          }}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg flex items-center gap-2"
        >
          <Plus className="w-4 h-4" /> Add Room
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {rooms.map(room => (
          <div key={room.id} className="bg-white rounded-lg shadow overflow-hidden">
            <div className="p-4">
              <h3 className="text-xl font-semibold mb-2">{room.name}</h3>
              <p className="text-gray-600 mb-2">{room.description}</p>
              <p className="text-blue-600 font-bold">KES {room.price_per_hour}/hour</p>
              <p className="text-gray-500">Capacity: {room.capacity} people</p>
              <div className="flex gap-2 mt-4">
                <button
                  onClick={() => {
                    setEditingRoom(room);
                    setFormData(room);
                    setIsModalOpen(true);
                  }}
                  className="flex-1 bg-yellow-500 text-white py-1 rounded flex items-center justify-center gap-1"
                >
                  <Edit className="w-4 h-4" /> Edit
                </button>
                <button
                  onClick={() => handleDelete(room.id)}
                  className="flex-1 bg-red-500 text-white py-1 rounded flex items-center justify-center gap-1"
                >
                  <Trash2 className="w-4 h-4" /> Delete
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Modal for Add/Edit Room */}
      {isModalOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h2 className="text-xl font-bold mb-4">
              {editingRoom ? 'Edit Room' : 'Add New Room'}
            </h2>
            <form onSubmit={handleSubmit}>
              <div className="mb-4">
                <label className="block text-sm font-medium mb-1">Room Name</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({...formData, name: e.target.value})}
                  className="w-full border rounded-lg px-3 py-2"
                  required
                />
              </div>
              <div className="mb-4">
                <label className="block text-sm font-medium mb-1">Description</label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({...formData, description: e.target.value})}
                  className="w-full border rounded-lg px-3 py-2"
                  rows="3"
                />
              </div>
              <div className="mb-4">
                <label className="block text-sm font-medium mb-1">Price per Hour (KES)</label>
                <input
                  type="number"
                  value={formData.price_per_hour}
                  onChange={(e) => setFormData({...formData, price_per_hour: e.target.value})}
                  className="w-full border rounded-lg px-3 py-2"
                  required
                />
              </div>
              <div className="mb-4">
                <label className="block text-sm font-medium mb-1">Capacity</label>
                <input
                  type="number"
                  value={formData.capacity}
                  onChange={(e) => setFormData({...formData, capacity: e.target.value})}
                  className="w-full border rounded-lg px-3 py-2"
                  required
                />
              </div>
              <div className="flex gap-2">
                <button type="submit" className="flex-1 bg-blue-600 text-white py-2 rounded">
                  {editingRoom ? 'Update' : 'Create'}
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setIsModalOpen(false);
                    resetForm();
                  }}
                  className="flex-1 bg-gray-300 text-gray-700 py-2 rounded"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default RoomManagement;
