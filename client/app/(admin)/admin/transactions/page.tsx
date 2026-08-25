'use client';

import { useEffect, useState } from 'react';
import { adminService } from '@/feature/admin/admin.service';
import { toast } from 'sonner';
import { Icon } from '@iconify/react';

export default function AdminTransactionsPage() {
  const [transactions, setTransactions] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchTransactions();
  }, []);

  const fetchTransactions = async () => {
    try {
      setLoading(true);
      const data = await adminService.getTransactions();
      setTransactions(data || []);
    } catch (error) {
      console.error('Failed to fetch transactions:', error);
      toast.error('Failed to load transactions');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <div className="text-gray-400">Loading transactions...</div>;
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Monitor Transactions</h1>
      
      <div className="bg-[#13151a] border border-[#22252e] rounded-2xl overflow-hidden shadow-xl">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead className="bg-[#1a1d24] text-gray-400">
              <tr>
                <th className="px-6 py-4 font-medium">TxID / Hash</th>
                <th className="px-6 py-4 font-medium">Type</th>
                <th className="px-6 py-4 font-medium">Status</th>
                <th className="px-6 py-4 font-medium text-right">Amount</th>
                <th className="px-6 py-4 font-medium text-right">Date</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#22252e]">
              {transactions.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-6 py-8 text-center text-gray-500">
                    No transactions found.
                  </td>
                </tr>
              ) : (
                transactions.map((t) => (
                  <tr key={t.id} className="hover:bg-[#1a1d24]/50 transition-colors">
                    <td className="px-6 py-4 text-white">
                      <div className="text-xs text-gray-400" title={t.id}>{t.id.substring(0,8)}...</div>
                      <div className="text-xs text-blue-400 truncate max-w-[200px]" title={t.tx_hash}>{t.tx_hash || '-'}</div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        {t.type === 'deposit' ? (
                          <Icon icon="hugeicons:arrow-down-left-01" className="text-green-400 size-4" />
                        ) : (
                          <Icon icon="hugeicons:arrow-up-right-01" className="text-red-400 size-4" />
                        )}
                        <span className="capitalize text-gray-300">{t.type}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${
                        t.status === 'confirmed' ? 'bg-green-500/10 text-green-400' :
                        t.status === 'failed' ? 'bg-red-500/10 text-red-400' :
                        'bg-yellow-500/10 text-yellow-400'
                      }`}>
                        {t.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-right font-mono text-white">
                      {parseFloat(t.amount).toLocaleString(undefined, { maximumFractionDigits: 6 })}
                    </td>
                    <td className="px-6 py-4 text-right text-gray-400 text-xs">
                      {new Date(t.created_at).toLocaleString()}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
