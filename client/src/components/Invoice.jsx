import { useCallback, useEffect, useState } from 'react';
import { Download } from 'lucide-react';
import invoiceService from '../services/invoiceService';
import toast from 'react-hot-toast';

const Invoice = ({ bookingId }) => {
  const [invoice, setInvoice] = useState(null);
  const [loading, setLoading] = useState(false);

  const loadInvoice = useCallback(async () => {
    setLoading(true);
    try {
      const data = await invoiceService.getInvoice(bookingId);
      setInvoice(data);
    } catch (error) {
      console.error('Error loading invoice:', error);
    } finally {
      setLoading(false);
    }
  }, [bookingId]);

  useEffect(() => {
    if (bookingId) {
      const timer = window.setTimeout(() => {
        void loadInvoice();
      }, 0);

      return () => window.clearTimeout(timer);
    }
  }, [bookingId, loadInvoice]);

  const handleDownload = async () => {
    try {
      const blob = await invoiceService.downloadInvoice(invoice.id);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `invoice_${invoice.invoice_number}.pdf`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      toast.success('Invoice downloaded');
    } catch {
      toast.error('Failed to download invoice');
    }
  };

  if (loading) return <div className="text-center py-4">Loading invoice...</div>;
  if (!invoice) return null;

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-xl font-semibold">Invoice {invoice.invoice_number}</h3>
        <button
          onClick={handleDownload}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg flex items-center gap-2"
        >
          <Download className="w-4 h-4" /> Download PDF
        </button>
      </div>
      
      <div className="space-y-2">
        <div className="flex justify-between py-2 border-b">
          <span className="font-medium">Subtotal:</span>
          <span>KES {invoice.amount.toFixed(2)}</span>
        </div>
        <div className="flex justify-between py-2 border-b">
          <span className="font-medium">Tax (16% VAT):</span>
          <span>KES {invoice.tax.toFixed(2)}</span>
        </div>
        <div className="flex justify-between py-2 border-b">
          <span className="font-medium">Total:</span>
          <span className="font-bold text-lg">KES {invoice.total_amount.toFixed(2)}</span>
        </div>
        <div className="flex justify-between py-2">
          <span className="font-medium">Status:</span>
          <span className={`px-2 py-1 rounded ${
            invoice.status === 'paid' ? 'bg-green-100 text-green-800' :
            invoice.status === 'pending' ? 'bg-yellow-100 text-yellow-800' :
            'bg-red-100 text-red-800'
          }`}>
            {invoice.status.toUpperCase()}
          </span>
        </div>
        <div className="flex justify-between py-2">
          <span className="font-medium">Due Date:</span>
          <span>{new Date(invoice.due_date).toLocaleDateString()}</span>
        </div>
      </div>
    </div>
  );
};

export default Invoice;
