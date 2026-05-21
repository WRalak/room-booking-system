import apiClient from './apiClient';

class ReviewService {
  async createReview(reviewData) {
    const response = await apiClient.post('/reviews', reviewData);
    return response.data;
  }

  async getRoomReviews(roomId) {
    const response = await apiClient.get(`/reviews/room/${roomId}`);
    return response.data;
  }

  async getUserReviews(userId) {
    const response = await apiClient.get(`/reviews/user/${userId}`);
    return response.data;
  }
}

export default new ReviewService();
