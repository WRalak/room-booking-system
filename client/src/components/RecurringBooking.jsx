import React, { useState } from 'react';
import DatePicker from 'react-datepicker';
import { Calendar, Repeat, Clock } from 'lucide-react';
import bookingService from '../services/bookingService';
import toast from 'react-hot-toast';

const RecurringBooking = ({ room, user }) => {
  const [startDate, setStartDate] = useState(new Date());
  const [endDate, setEndDate] = useState(new Date());
  const [frequency, setFrequency] = useState('weekly');
  const [interval, setInterval] = useState(1);
  const [dayOfWeek, setDayOfWeek] = useState(null);
  const [occurrences, setOccurrences] = useState(4);
  const [loading, setLoading] = useState(false);

  const calculateTotal = () => {
    const hours = (endDate - startDate) / (1000 * 60 * 60);
    return hours * room.pricePerHour * occurrences;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    
    try {
      await bookingService.createRecurringBooking({
        room_id: room.id,
        user_id: user.id,
        start_time: startDate.toISOString(),
        end_time: endDate.toISOString(),
        frequency: frequency,
        interval: interval,
        day_of_week: frequency === 'weekly' ? dayOfWeek : null,
        occurrences: occurrences
      });
      
      toast.success(`Recurring booking created for ${occurrences} ${frequency} sessions`);
      // Redirect or reset form
    } catch (error) {
      toast.error('Failed to create recurring booking');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h2 className="text-2xl font-bold mb-6">Recurring Booking</h2>
      <form onSubmit={handleSubmit}>
        <div className="mb-4">
          <label className="block text-gray-700 text-sm font-bold mb-2">
            Start Time
          </label>
          <DatePicker
            selected={startDate}
            onChange={setStartDate}
            showTimeSelect
            dateFormat="MMMM d, yyyy h:mm aa"
            className="w-full px-3 py-2 border rounded-lg"
            minDate={new Date()}
          />
        </div>

        <div className="mb-4">
          <label className="block text-gray-700 text-sm font-bold mb-2">
            End Time
          </label>
          <DatePicker
            selected={endDate}
            onChange={setEndDate}
            showTimeSelect
            dateFormat="MMMM d, yyyy h:mm aa"
            className="w-full px-3 py-2 border rounded-lg"
            minDate={startDate}
          />
        </div>

        <div className="mb-4">
          <label className="block text-gray-700 text-sm font-bold mb-2">
            Frequency
          </label>
          <select
            value={frequency}
            onChange={(e) => setFrequency(e.target.value)}
            className="w-full px-3 py-2 border rounded-lg"
          >
            <option value="daily">Daily</option>
            <option value="weekly">Weekly</option>
            <option value="monthly">Monthly</option>
          </select>
        </div>

        <div className="mb-4">
          <label className="block text-gray-700 text-sm font-bold mb-2">
            Every {interval} {frequency}(s)
          </label>
          <input
            type="number"
            min="1"
            max="12"
            value={interval}
            onChange={(e) => setInterval(parseInt(e.target.value))}
            className="w-full px-3 py-2 border rounded-lg"
          />
        </div>

        {frequency === 'weekly' && (
          <div className="mb-4">
            <label className="block text-gray-700 text-sm font-bold mb-2">
              Day of Week
            </label>
            <select
              value={dayOfWeek}
              onChange={(e) => setDayOfWeek(parseInt(e.target.value))}
              className="w-full px-3 py-2 border rounded-lg"
            >
              <option value="">Select day</option>
              <option value="0">Sunday</option>
              <option value="1">Monday</option>
              <option value="2">Tuesday</option>
              <option value="3">Wednesday</option>
              <option value="4">Thursday</option>
              <option value="5">Friday</option>
              <option value="6">Saturday</option>
            </select>
          </div>
        )}

        <div className="mb-4">
          <label className="block text-gray-700 text-sm font-bold mb-2">
            Number of Occurrences
          </label>
          <input
            type="number"
            min="1"
            max="52"
            value={occurrences}
            onChange={(e) => setOccurrences(parseInt(e.target.value))}
            className="w-full px-3 py-2 border rounded-lg"
          />
        </div>

        <div className="mb-6 p-4 bg-gray-100 rounded-lg">
          <h3 className="font-semibold mb-2">Summary</h3>
          <p>Room: {room.name}</p>
          <p>Total Sessions: {occurrences}</p>
          <p className="text-xl font-bold mt-2">Total Amount: KES {calculateTotal().toFixed(2)}</p>
          <p className="text-sm text-gray-600 mt-1">
            You will be charged per occurrence as each booking is confirmed
          </p>
        </div>

        <button
          type="submit"
          disabled={loading}
          className="w-full bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 transition disabled:opacity-50"
        >
          {loading ? 'Creating...' : 'Create Recurring Booking'}
        </button>
      </form>
    </div>
  );
};

export default RecurringBooking;