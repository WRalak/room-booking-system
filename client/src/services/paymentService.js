import axios from 'axios';
import PaystackPop from '@paystack/inline-js';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

class PaymentService {
  constructor() {
    this.paystack = null;
  }

  async initializePaystackPayment(email, amount, bookingId, onSuccess, onCancel) {
    try {
      const response = await axios.post(`${API_BASE_URL}/payments/paystack/initialize`, {
        email,
        amount,
        booking_id: bookingId
      });

      const { data } = response;
      
      this.paystack = new PaystackPop();
      this.paystack.newTransaction({
        key: process.env.REACT_APP_PAYSTACK_PUBLIC_KEY,
        email: email,
        amount: amount * 100,
        reference: data.data.reference,
        onSuccess: () => {
          onSuccess();
          this.verifyPayment(data.data.reference);
        },
        onCancel: () => {
          onCancel();
        }
      });
    } catch (error) {
      console.error('Error initializing Paystack payment:', error);
      throw error;
    }
  }

  async initiateMpesaPayment(phoneNumber, amount, bookingId) {
    try {
      const response = await axios.post(`${API_BASE_URL}/payments/mpesa/stkpush`, {
        phone_number: phoneNumber,
        amount: amount,
        booking_id: bookingId
      });
      
      return response.data;
    } catch (error) {
      console.error('Error initiating M-Pesa payment:', error);
      throw error;
    }
  }

  async verifyPayment(reference) {
    try {
      const response = await axios.get(`${API_BASE_URL}/payments/status/${reference}`);
      return response.data;
    } catch (error) {
      console.error('Error verifying payment:', error);
      throw error;
    }
  }

  async checkPaymentStatus(reference) {
    try {
      const response = await axios.get(`${API_BASE_URL}/payments/status/${reference}`);
      return response.data.status === 'success';
    } catch (error) {
      return false;
    }
  }
}

export default new PaymentService();