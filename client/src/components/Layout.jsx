import React from 'react';
import { Link } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import { useAuth } from '../context/useAuth';
import gamificationService from '../services/gamificationService';
import NotificationCenter from './NotificationCenter';

const Layout = ({ children }) => {
  const { user, isAuthenticated, logout } = useAuth();
  const userId = user ? String(user.id) : 'guest';
  const points = gamificationService.getPoints(userId);
  const badges = user ? gamificationService.getBadges(userId) : [];

  return (
    <div className="min-h-screen flex flex-col bg-gray-100">
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center gap-6">
              <Link to="/" className="text-2xl font-bold text-blue-600" aria-label="Home">
                RoomBooker
              </Link>
              <nav aria-label="Main navigation">
                <ul className="flex gap-4 items-center">
                  <li><Link to="/rooms" className="text-gray-700 hover:text-blue-600">Rooms</Link></li>
                  {isAuthenticated && <li><Link to="/dashboard" className="text-gray-700 hover:text-blue-600">Dashboard</Link></li>}
                  <li><Link to="/leaderboard" className="text-gray-700 hover:text-blue-600">Leaderboard</Link></li>
                </ul>
              </nav>
            </div>
            <div className="flex items-center gap-4">
              {isAuthenticated ? (
                <>
                  <NotificationCenter />
                  <div className="text-sm text-gray-700">Points: <span className="font-semibold">{points}</span></div>
                  <div className="hidden sm:block text-sm text-gray-700">Badges: {badges.length}</div>
                  <button
                    type="button"
                    onClick={logout}
                    className="rounded-lg border border-gray-300 px-3 py-1 text-sm text-gray-700 hover:bg-gray-50"
                  >
                    Sign out
                  </button>
                  {user?.is_admin && (
                    <Link to="/admin/dashboard" className="rounded-lg bg-blue-600 px-3 py-1 text-sm text-white hover:bg-blue-700">Admin</Link>
                  )}
                </>
              ) : (
                <Link to="/auth" className="rounded-lg bg-blue-600 px-3 py-1 text-sm text-white hover:bg-blue-700">Sign in</Link>
              )}
            </div>
          </div>
        </div>
      </header>

      <main className="flex-1 max-w-7xl mx-auto py-6 sm:px-6 lg:px-8 w-full">{children}</main>

      <footer className="bg-white border-t">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4 text-sm text-gray-500">
          © {new Date().getFullYear()} RoomBooker — Built for simplicity
        </div>
      </footer>
      <Toaster />
    </div>
  );
};

export default Layout;
