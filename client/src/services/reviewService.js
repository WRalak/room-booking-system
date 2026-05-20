import axios from 'axios';

const API_URL = process.env.REACT_APP_API_URL;

class ReviewService {
  async createReview(reviewData) {
    const response = await axios.post(`${API_URL}/reviews`, reviewData);
    return response.data;
  }

  async getRoomReviews(roomId) {
    const response = await axios.get(`${API_URL}/reviews/room/${roomId}`);
    return response.data;
  }

  async getUserReviews(userId) {
    const response = await axios.get(`${API_URL}/reviews/user/${userId}`);
    return response.data;
  }
}

export default new ReviewService();