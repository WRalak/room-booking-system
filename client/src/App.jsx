import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import BookingForm from './components/BookingForm';
import RoomList from './components/RoomList';
import MyBookings from './components/MyBookings';
import Layout from './components/Layout';
import Leaderboard from './components/Leaderboard';
import PaymentPage from './components/PaymentPage';
import UserDashboard from './components/UserDashboard';
import NotificationPage from './components/NotificationPage';
import AdminDashboard from './components/Admin/Dashboard';
import AuthPage from './components/AuthPage';
import ProtectedRoute from './components/ProtectedRoute';

function App() {
  return (
    <Router>
      <Layout>
        <div className="space-y-6">
          <Routes>
            <Route
              path="/"
              element={
                <div className="text-center">
                  <h2 className="text-4xl font-bold mb-4">Book Your Perfect Room</h2>
                  <p className="text-xl text-gray-600 mb-8">Find and reserve meeting rooms, conference halls, and workspaces with ease.</p>
                  <Link to="/rooms" className="bg-blue-600 text-white px-6 py-3 rounded-lg">Browse Rooms</Link>
                </div>
              }
            />
            <Route path="/auth" element={<AuthPage />} />
            <Route path="/rooms" element={<RoomList />} />
            <Route path="/book/:roomId" element={<ProtectedRoute><BookingForm /></ProtectedRoute>} />
            <Route path="/my-bookings" element={<ProtectedRoute><MyBookings /></ProtectedRoute>} />
            <Route path="/dashboard" element={<ProtectedRoute><UserDashboard /></ProtectedRoute>} />
            <Route path="/notifications" element={<ProtectedRoute><NotificationPage /></ProtectedRoute>} />
            <Route path="/leaderboard" element={<Leaderboard />} />
            <Route path="/pay/:bookingId" element={<ProtectedRoute><PaymentPage /></ProtectedRoute>} />
            <Route path="/admin/dashboard" element={<ProtectedRoute requireAdmin><AdminDashboard /></ProtectedRoute>} />
          </Routes>
        </div>
      </Layout>
    </Router>
  );
}

export default App;
