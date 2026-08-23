'use client';

import { useState, useEffect, useMemo } from 'react';
import { Icon } from '@iconify/react';
import { useRouter } from 'next/navigation';
import { transactionsService } from '@/feature/transactions/transactions.service';
import { assetsService } from '@/feature/assets/assets.service';
import { type TransactionItem } from '@/feature/transactions/transactions.schema';
import { type Asset } from '@/feature/assets/assets.schema';
import SwapModal from '@/components/commons/swap/swapmodal';

export default function SwapPage() {
  const router = useRouter();
  const [transactions, setTransactions] = useState<TransactionItem[]>([]);
  const [assets, setAssets] = useState<Asset[]>([]);
  const [loading, setLoading] = useState(true);
  const [isSwapModalOpen, setIsSwapModalOpen] = useState(false);

  // Pagination
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const limit = 10;

  useEffect(() => {
    const fetchAssets = async () => {
      try {
        const data = await assetsService.getAssets();
        setAssets(data);
      } catch (err) {
        console.error('Failed to load assets mapping:', err);
      }
    };
    fetchAssets();
  }, []);

  const fetchTxs = async () => {
    setLoading(true);
    try {
      const res = await transactionsService.getTransactions({
        page,
        limit,
        type: 'swap',
      });
      setTransactions(res.transactions || []);
      setTotal(res.total || 0);
    } catch (err) {
      console.error('Failed to fetch swap transactions:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTxs();
  }, [page]);

  const assetMap = useMemo(() => {
    const map = new Map<number, Asset>();
    assets.forEach((a) => map.set(a.id, a));
    return map;
  }, [assets]);

  const totalPages = Math.ceil(total / limit) || 1;

  const getStatusBadge = (status: string) => {
    switch (status.toLowerCase()) {
      case 'confirmed':
        return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
      case 'pending':
        return 'bg-amber-500/10 text-amber-400 border-amber-500/20';
      case 'failed':
        return 'bg-red-500/10 text-red-400 border-red-500/20';
      default:
        return 'bg-[#22252e] text-[#888c99] border-[#2a2d36]';
    }
  };

  return (
    <div className="flex flex-col w-full max-w-[1440px] mx-auto animate-in fade-in duration-500 pb-20 font-sans gap-8">
      
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex flex-col gap-1">
          <h1 className="text-3xl font-bold text-white tracking-tight">Swaps</h1>
          <p className="text-[15px] text-[#888c99]">
            Exchange tokens instantly and view your swap history
          </p>
        </div>
        
        <button
          onClick={() => setIsSwapModalOpen(true)}
          className="bg-blue-600 hover:bg-blue-500 text-white font-semibold px-6 py-2.5 rounded-full transition-colors flex items-center gap-2 shadow-lg shadow-blue-600/20 self-start sm:self-auto"
        >
          <Icon icon="hugeicons:arrow-up-down" className="size-5" />
          <span>New Swap</span>
        </button>
      </div>

      {/* Main Container */}
      <div className="w-full bg-[#13151a] border border-[#22252e] rounded-[24px] overflow-hidden shadow-2xl flex flex-col">
        
        {/* Header Bar */}
        <div className="p-4 sm:p-6 border-b border-[#22252e] flex items-center justify-between bg-[#16181d]/50">
          <h2 className="text-lg font-bold text-white">Swap History</h2>
          <div className="text-xs text-[#888c99] font-medium">
            Showing {transactions.length} of {total} records
          </div>
        </div>

        {/* Transactions Table / List */}
        <div className="w-full overflow-x-auto">
          {loading ? (
            <div className="p-8 flex flex-col gap-4">
              {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="h-16 bg-[#1c1f26] animate-pulse rounded-xl w-full" />
              ))}
            </div>
          ) : transactions.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 px-4 gap-4 text-center">
              <div className="size-16 rounded-full bg-[#1c1f26] border border-[#22252e] flex items-center justify-center text-[#888c99]">
                <Icon icon="hugeicons:arrow-data-transfer-horizontal" className="size-8" />
              </div>
              <div className="flex flex-col gap-1">
                <h3 className="text-lg font-bold text-white">No swaps found</h3>
                <p className="text-sm text-[#888c99] max-w-sm">
                  You haven't made any swaps yet. Click "New Swap" to get started.
                </p>
              </div>
            </div>
          ) : (
            <table className="w-full text-left border-collapse min-w-[700px]">
              <thead>
                <tr className="border-b border-[#22252e] text-[13px] font-semibold text-[#888c99] bg-[#16181d]/30">
                  <th className="py-4 px-6">Asset</th>
                  <th className="py-4 px-6">Amount</th>
                  <th className="py-4 px-6">Status</th>
                  <th className="py-4 px-6">Tx Hash</th>
                  <th className="py-4 px-6 text-right">Date</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#22252e]">
                {transactions.map((tx) => {
                  const asset = assetMap.get(tx.asset_id);
                  const isDebit = tx.amount && tx.amount.toString().startsWith('-'); // Since it's a swap, one side is out, one side is in. But actually, in our new code, from is positive number in amount? Let's check: fromAmount is passed, we insert it.
                  // Wait, earlier we just put amount. Let's rely on standard UI display.

                  return (
                    <tr
                      key={tx.id}
                      onClick={() => router.push(`/swap/${tx.id}`)}
                      className="hover:bg-[#1c1f26]/60 transition-colors group text-sm cursor-pointer"
                    >
                      {/* Asset */}
                      <td className="py-4 px-6">
                        <div className="flex items-center gap-3.5">
                          <div
                            className="size-10 rounded-full flex items-center justify-center border shrink-0 bg-purple-500/10 text-purple-400 border-purple-500/20"
                          >
                            <Icon icon="hugeicons:arrow-data-transfer-horizontal" className="size-5" />
                          </div>
                          <div className="flex flex-col">
                            <span className="font-bold text-white capitalize leading-tight">
                              Swap
                            </span>
                            <span className="text-xs text-[#888c99]">
                              {asset ? `${asset.name} (${asset.symbol})` : `Asset #${tx.asset_id}`}
                            </span>
                          </div>
                        </div>
                      </td>

                      {/* Amount */}
                      <td className="py-4 px-6">
                        <div className="flex flex-col">
                          <span className="font-mono font-bold text-white">
                            {tx.amount} {asset?.symbol || ''}
                          </span>
                          {tx.fee && tx.fee !== '0' && (
                            <span className="text-xs text-[#888c99]">Fee: {tx.fee}</span>
                          )}
                        </div>
                      </td>

                      {/* Status */}
                      <td className="py-4 px-6">
                        <span
                          className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-bold uppercase tracking-wider border ${getStatusBadge(
                            tx.status
                          )}`}
                        >
                          {tx.status}
                        </span>
                      </td>

                      {/* Hash */}
                      <td className="py-4 px-6">
                        <span className="font-mono text-xs text-blue-400 hover:text-blue-300 max-w-[140px] truncate underline underline-offset-2">
                          {tx.id}
                        </span>
                      </td>

                      {/* Date */}
                      <td className="py-4 px-6 text-right">
                        <div className="flex flex-col items-end">
                          <span className="text-white font-medium">
                            {new Date(tx.created_at).toLocaleDateString(undefined, {
                              month: 'short',
                              day: 'numeric',
                              year: 'numeric',
                            })}
                          </span>
                          <span className="text-xs text-[#888c99]">
                            {new Date(tx.created_at).toLocaleTimeString(undefined, {
                              hour: '2-digit',
                              minute: '2-digit',
                            })}
                          </span>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          )}
        </div>

        {/* Pagination Footer */}
        {totalPages > 1 && (
          <div className="p-4 sm:p-6 border-t border-[#22252e] flex items-center justify-between bg-[#16181d]/50">
            <button
              disabled={page === 1}
              onClick={() => setPage(page - 1)}
              className="px-4 py-2 rounded-xl bg-[#1c1f26] text-white text-sm font-semibold hover:bg-[#22252e] disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              Previous
            </button>
            <span className="text-sm font-medium text-[#888c99]">
              Page <span className="text-white">{page}</span> of {totalPages}
            </span>
            <button
              disabled={page === totalPages}
              onClick={() => setPage(page + 1)}
              className="px-4 py-2 rounded-xl bg-[#1c1f26] text-white text-sm font-semibold hover:bg-[#22252e] disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              Next
            </button>
          </div>
        )}
      </div>

      <SwapModal 
        isOpen={isSwapModalOpen} 
        onClose={() => setIsSwapModalOpen(false)}
        onSuccess={() => fetchTxs()} 
      />
    </div>
  );
}
