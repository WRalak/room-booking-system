import apiClient from './apiClient';

class PaymentService {
  constructor() {
    this.paystack = null;
  }

  async initializePaystackPayment(email, amount, bookingId, onSuccess, onCancel) {
    try {
      const response = await apiClient.post('/payments/paystack/initialize', {
        email,
        amount,
        booking_id: bookingId
      });

      const { data } = response;
      // Load Paystack inline script dynamically if needed
      await new Promise((resolve, reject) => {
        if (window.PaystackPop) return resolve();
        const s = document.createElement('script');
        s.src = 'https://js.paystack.co/v1/inline.js';
        s.onload = () => resolve();
        s.onerror = reject;
        document.body.appendChild(s);
      });

      const Paystack = window.PaystackPop;
      if (!Paystack) throw new Error('Paystack script failed to load');

      // Try setup API (common pattern) then open iframe
      const handler = Paystack.setup
        ? Paystack.setup({
            key: import.meta.env.VITE_PAYSTACK_PUBLIC_KEY,
            email,
            amount: amount * 100,
            ref: data.data.reference,
            callback: (res) => {
              onSuccess();
              this.verifyPayment(res.reference || data.data.reference);
            },
            onClose: onCancel,
          })
        : null;

      if (handler && typeof handler.openIframe === 'function') {
        handler.openIframe();
      } else if (Paystack.newTransaction) {
        // fallback to older/newTransaction API
        const p = new Paystack();
        if (p.newTransaction) {
          p.newTransaction({
            key: import.meta.env.VITE_PAYSTACK_PUBLIC_KEY,
            email,
            amount: amount * 100,
            reference: data.data.reference,
            onSuccess: () => {
              onSuccess();
              this.verifyPayment(data.data.reference);
            },
            onCancel,
          });
        }
      } else {
        // As a last resort, open the authorization_url if provided
        const url = data?.data?.authorization_url || data?.data?.checkout_url;
        if (url) window.location.href = url;
        else throw new Error('Unable to start Paystack payment');
      }
    } catch (error) {
      console.error('Error initializing Paystack payment:', error);
      throw error;
    }
  }

  async initiateMpesaPayment(phoneNumber, amount, bookingId) {
    try {
      const response = await apiClient.post('/payments/mpesa/stkpush', {
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
      const response = await apiClient.get(`/payments/status/${reference}`);
      return response.data;
    } catch (error) {
      console.error('Error verifying payment:', error);
      throw error;
    }
  }

  async checkPaymentStatus(reference) {
    try {
      const response = await apiClient.get(`/payments/status/${reference}`);
      return response.data.status === 'success';
    } catch {
      return false;
    }
  }
}

export default new PaymentService();
