import apiClient from './apiClient';

class AdminService {
  async getDashboardStats() {
    const response = await apiClient.get('/admin/dashboard');
    return response.data;
  }

  async getRevenueReport(startDate, endDate) {
    const response = await apiClient.get('/admin/revenue', {
      params: { start_date: startDate, end_date: endDate }
    });
    return response.data;
  }

  async getAllRooms() {
    const response = await apiClient.get('/admin/rooms');
    return response.data;
  }

  async createRoom(roomData) {
    const response = await apiClient.post('/admin/rooms', roomData);
    return response.data;
  }

  async updateRoom(roomId, roomData) {
    const response = await apiClient.put(`/admin/rooms/${roomId}`, roomData);
    return response.data;
  }

  async deleteRoom(roomId) {
    const response = await apiClient.delete(`/admin/rooms/${roomId}`);
    return response.data;
  }

  async getAllBookings(filters) {
    const response = await apiClient.get('/admin/bookings', { params: filters });
    return response.data;
  }

  async verifyReview(reviewId) {
    const response = await apiClient.post(`/admin/reviews/${reviewId}/verify`);
    return response.data;
  }
}

export default new AdminService();
