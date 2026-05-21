import apiClient from './apiClient';

const SESSION_KEY = 'auth_session';

class AuthService {
  getStoredSession() {
    const session = localStorage.getItem(SESSION_KEY);
    return session ? JSON.parse(session) : null;
  }

  storeSession(session) {
    localStorage.setItem(SESSION_KEY, JSON.stringify(session));
  }

  clearSession() {
    localStorage.removeItem(SESSION_KEY);
  }

  async login(credentials) {
    const response = await apiClient.post('/auth/login', credentials);
    this.storeSession(response.data);
    return response.data;
  }

  async register(userData) {
    const response = await apiClient.post('/auth/register', userData);
    this.storeSession(response.data);
    return response.data;
  }

  async me() {
    const response = await apiClient.get('/auth/me');
    return response.data.user;
  }
}

export default new AuthService();
