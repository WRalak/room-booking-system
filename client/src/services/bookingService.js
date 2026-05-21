import apiClient from './apiClient';

class BookingService {
  async createBooking(bookingData) {
    try {
      const response = await apiClient.post('/bookings/', bookingData);
      return response.data;
    } catch (error) {
      // Fallback: persist booking locally
      console.warn('bookingService.createBooking failed, saving locally', error);
      const stored = JSON.parse(localStorage.getItem('local_bookings') || '[]');
      const fallbackAmount = bookingData.amount || (bookingData.end_time && bookingData.start_time
        ? (new Date(bookingData.end_time) - new Date(bookingData.start_time)) / 3600000 * (bookingData.price_per_hour || 0)
        : 0);
      const fallback = {
        id: Date.now(),
        ...bookingData,
        status: 'confirmed',
        amount: fallbackAmount,
        total_amount: fallbackAmount,
      };
      localStorage.setItem('local_bookings', JSON.stringify([...stored, fallback]));
      return fallback;
    }
  }

  async getUserBookings(userId) {
    try {
      const response = await apiClient.get(`/bookings/user/${userId}`);
      return response.data;
    } catch (error) {
      console.warn('bookingService.getUserBookings failed, using local_bookings', error);
      const stored = JSON.parse(localStorage.getItem('local_bookings') || '[]');
      return stored.filter(b => String(b.user_id) === String(userId));
    }
  }

  async getRoomBookings(roomId) {
    const response = await apiClient.get(`/bookings/room/${roomId}`);
    return response.data;
  }

  async cancelBooking(bookingId) {
    try {
      const response = await apiClient.put(`/bookings/${bookingId}/cancel`);
      return response.data;
    } catch (error) {
      console.warn('bookingService.cancelBooking failed, updating local_bookings', error);
      // Fallback: update local storage booking status
      const stored = JSON.parse(localStorage.getItem('local_bookings') || '[]');
      const updated = stored.map(b => (b.id === bookingId ? { ...b, status: 'cancelled' } : b));
      localStorage.setItem('local_bookings', JSON.stringify(updated));
      return { success: true };
    }
  }

  async createRecurringBooking(bookingData) {
    const response = await apiClient.post('/recurring-bookings', bookingData);
    return response.data;
  }

  async getUserRecurringBookings(userId) {
    const response = await apiClient.get(`/recurring-bookings/user/${userId}`);
    return response.data;
  }

  async cancelRecurringBooking(recurringBookingId) {
    const response = await apiClient.put(`/recurring-bookings/${recurringBookingId}/cancel`);
    return response.data;
  }
}

export default new BookingService();
