import axios from 'axios';

const API_URL = process.env.REACT_APP_API_URL;

class InvoiceService {
  async getInvoice(bookingId) {
    const response = await axios.get(`${API_URL}/invoices/booking/${bookingId}`);
    return response.data;
  }

  async downloadInvoice(invoiceId) {
    const response = await axios.get(`${API_URL}/invoices/${invoiceId}/download`, {
      responseType: 'blob'
    });
    return response.data;
  }

  async getUserInvoices(userId) {
    const response = await axios.get(`${API_URL}/invoices/user/${userId}`);
    return response.data;
  }
}

export default new InvoiceService();