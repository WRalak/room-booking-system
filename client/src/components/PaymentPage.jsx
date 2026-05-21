import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import paymentService from '../services/paymentService';
// bookingService not required here; payments initiated via paymentService
import gamificationService from '../services/gamificationService';
import Spinner from './Spinner';

export default function PaymentPage() {
  const { bookingId } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = React.useState(true);
  const [booking, setBooking] = React.useState(null);
  const [phone, setPhone] = React.useState('');

  React.useEffect(() => {
    let mounted = true;
    (async () => {
      try {
        // Try to fetch booking from user bookings or local fallback
        // We will attempt to find the booking in local storage if API fails
        const stored = JSON.parse(localStorage.getItem('local_bookings') || '[]');
        const found = stored.find(b => String(b.id) === String(bookingId));
        if (found && mounted) setBooking(found);
      } finally {
        if (mounted) setLoading(false);
      }
    })();
    return () => { mounted = false; };
  }, [bookingId]);

  if (loading) return <div className="text-center py-10"><Spinner size={48} /><div className="mt-2">Loading payment...</div></div>;

  if (!booking) {
    return (
      <div className="bg-white rounded shadow p-6 text-center">
        <h2 className="text-xl font-semibold mb-2">Booking not found</h2>
        <p className="text-gray-600 mb-4">Cannot locate booking to pay for.</p>
      </div>
    );
  }

  const amount = booking.amount || booking.total_amount || booking.TotalAmount ||
    (booking.end_time && booking.start_time ? Math.max(1, Math.round((new Date(booking.end_time) - new Date(booking.start_time)) / 3600000) * (booking.price_per_hour || booking.pricePerHour || 0)) : 0);

  const handlePaystack = async () => {
    try {
      await paymentService.initializePaystackPayment(booking.user_email || booking.email || 'guest@example.com', amount, booking.id,
        async () => {
          toast.success('Payment successful');
          // award bonus points
          const uid = booking.user_id || localStorage.getItem('rb_user') || 'guest';
          gamificationService.addPoints(String(uid), 5);
          navigate('/my-bookings');
        },
        () => toast('Payment cancelled')
      );
    } catch (error) {
      console.error(error);
      toast.error('Failed to initialize Paystack payment');
    }
  };

  const handleMpesa = async () => {
    if (!phone) return toast.error('Please enter phone number');
    try {
      const resp = await paymentService.initiateMpesaPayment(phone, amount, booking.id);
      console.log('M-Pesa response', resp);
      toast.success('M-Pesa STK Push initiated. Check your phone.');
      // optionally poll for result
      // award small points for completing the payment later on verification
      // for now assume success and award points
      const uid = booking.user_id || localStorage.getItem('rb_user') || 'guest';
      gamificationService.addPoints(String(uid), 5);
      navigate('/my-bookings');
    } catch (error) {
      console.error(error);
      toast.error('Failed to initiate M-Pesa payment');
    }
  };

  return (
    <div className="max-w-2xl mx-auto bg-white rounded shadow p-6">
      <h2 className="text-2xl font-bold mb-4">Pay for Booking</h2>
      <p className="mb-2">Booking ID: <span className="font-mono">{bookingId}</span></p>
      <p className="mb-4">Amount: KES {amount.toLocaleString()}</p>

      <div className="space-y-4">
        <div>
          <button onClick={handlePaystack} className="w-full bg-yellow-500 text-black px-4 py-3 rounded">Pay with Paystack</button>
        </div>

        <div className="p-4 border rounded">
          <label className="block text-sm font-medium text-gray-700 mb-2">Phone for M-Pesa (e.g. 2547xxxxxxxx)</label>
          <input value={phone} onChange={(e) => setPhone(e.target.value)} className="w-full border rounded px-3 py-2" placeholder="2547..." />
          <button onClick={handleMpesa} className="mt-3 w-full bg-green-600 text-white px-4 py-2 rounded">Pay with M-Pesa</button>
        </div>
      </div>
    </div>
  );
}
