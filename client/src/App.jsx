import { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import BookingForm from './components/BookingForm';
import NotificationCenter from './components/NotificationCenter';
import RoomList from './components/RoomList';
import MyBookings from './components/MyBookings';

function App() {
  const [user] = useState(() => {
    const savedUser = localStorage.getItem('user');
    if (savedUser) {
      return JSON.parse(savedUser);
    }

    return {
      id: 1,
      email: 'user@example.com',
      name: 'John Doe',
      phone: '0712345678'
    };
  });

  const [selectedRoom, setSelectedRoom] = useState(null);

  return (
    <Router>
      <div className="min-h-screen bg-gray-100">
        <nav className="bg-white shadow-md">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between h-16">
              <div className="flex items-center">
                <Link to="/" className="text-xl font-bold text-blue-600">
                  RoomBooker
                </Link>
                <div className="ml-10 flex space-x-4">
                  <Link to="/rooms" className="text-gray-700 hover:text-blue-600">
                    Rooms
                  </Link>
                  <Link to="/my-bookings" className="text-gray-700 hover:text-blue-600">
                    My Bookings
                  </Link>
                </div>
              </div>
              <div className="flex items-center space-x-4">
                <NotificationCenter userId={user.id} />
                <div className="text-gray-700">
                  Welcome, {user.name}
                </div>
              </div>
            </div>
          </div>
        </nav>

        <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
          <Routes>
            <Route path="/" element={
              <div className="text-center">
                <h1 className="text-4xl font-bold mb-4">Book Your Perfect Room</h1>
                <p className="text-xl text-gray-600 mb-8">
                  Find and book meeting rooms, conference halls, and workspaces
                </p>
                <Link to="/rooms" className="bg-blue-600 text-white px-6 py-3 rounded-lg">
                  Browse Rooms
                </Link>
              </div>
            } />
            <Route path="/rooms" element={
              <RoomList onSelectRoom={setSelectedRoom} />
            } />
            <Route path="/book/:roomId" element={
              selectedRoom && <BookingForm room={selectedRoom} user={user} />
            } />
            <Route path="/my-bookings" element={
              <MyBookings userId={user.id} />
            } />
          </Routes>
        </main>

        <Toaster position="top-right" />
      </div>
    </Router>
  );
}

export default App;