import apiClient from './apiClient';

class InvoiceService {
  async getInvoice(bookingId) {
    const response = await apiClient.get(`/invoices/booking/${bookingId}`);
    return response.data;
  }

  async downloadInvoice(invoiceId) {
    const response = await apiClient.get(`/invoices/${invoiceId}/download`, {
      responseType: 'blob'
    });
    return response.data;
  }

  async getUserInvoices(userId) {
    const response = await apiClient.get(`/invoices/user/${userId}`);
    return response.data;
  }
}

export default new InvoiceService();
