import apiClient from './apiClient';
import sampleRooms from '../data/sampleRooms';

class RoomService {
  async getAvailableRooms(filters = {}) {
    try {
      const response = await apiClient.get('/rooms/', { params: filters });
      return response.data;
    } catch (error) {
      // Fallback to sample rooms when API is unreachable
      console.warn('roomService.getAvailableRooms failed, using sampleRooms fallback', error);
      return sampleRooms;
    }
  }

  async getRoomDetails(roomId) {
    try {
      const response = await apiClient.get(`/rooms/${roomId}`);
      return response.data;
    } catch (error) {
      console.warn('roomService.getRoomDetails failed, using sampleRooms fallback', error);
      return sampleRooms.find((r) => String(r.id) === String(roomId)) || null;
    }
  }
}

export default new RoomService();
