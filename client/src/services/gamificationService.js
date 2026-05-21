const POINTS_KEY = 'rb_points';
const BADGES_KEY = 'rb_badges';

class GamificationService {
  getPoints(userId) {
    const all = JSON.parse(localStorage.getItem(POINTS_KEY) || '{}');
    return all[userId] || 0;
  }

  addPoints(userId, points) {
    const all = JSON.parse(localStorage.getItem(POINTS_KEY) || '{}');
    all[userId] = (all[userId] || 0) + points;
    localStorage.setItem(POINTS_KEY, JSON.stringify(all));
    return all[userId];
  }

  grantBadge(userId, badge) {
    const all = JSON.parse(localStorage.getItem(BADGES_KEY) || '{}');
    all[userId] = all[userId] || [];
    if (!all[userId].includes(badge)) all[userId].push(badge);
    localStorage.setItem(BADGES_KEY, JSON.stringify(all));
    return all[userId];
  }

  getBadges(userId) {
    const all = JSON.parse(localStorage.getItem(BADGES_KEY) || '{}');
    return all[userId] || [];
  }

  // Simple leaderboard computed from local storage
  getLeaderboard() {
    const all = JSON.parse(localStorage.getItem(POINTS_KEY) || '{}');
    return Object.entries(all)
      .map(([userId, points]) => ({ userId, points }))
      .sort((a, b) => b.points - a.points)
      .slice(0, 10);
  }
}

export default new GamificationService();
