import axios from 'axios';

const API_URL = process.env.REACT_APP_API_URL;

class AdminService {
  async getDashboardStats() {
    const response = await axios.get(`${API_URL}/admin/dashboard`);
    return response.data;
  }

  async getRevenueReport(startDate, endDate) {
    const response = await axios.get(`${API_URL}/admin/revenue`, {
      params: { start_date: startDate, end_date: endDate }
    });
    return response.data;
  }

  async getAllRooms() {
    const response = await axios.get(`${API_URL}/admin/rooms`);
    return response.data;
  }

  async createRoom(roomData) {
    const response = await axios.post(`${API_URL}/admin/rooms`, roomData);
    return response.data;
  }

  async updateRoom(roomId, roomData) {
    const response = await axios.put(`${API_URL}/admin/rooms/${roomId}`, roomData);
    return response.data;
  }

  async deleteRoom(roomId) {
    const response = await axios.delete(`${API_URL}/admin/rooms/${roomId}`);
    return response.data;
  }

  async getAllBookings(filters) {
    const response = await axios.get(`${API_URL}/admin/bookings`, { params: filters });
    return response.data;
  }

  async verifyReview(reviewId) {
    const response = await axios.post(`${API_URL}/admin/reviews/${reviewId}/verify`);
    return response.data;
  }
}

export default new AdminService();