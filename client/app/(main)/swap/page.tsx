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
    <div className="flex flex-col w-full animate-in fade-in duration-500 pb-20 font-sans gap-8">
      
      {/* Page Header */}
      <div className="flex flex-row items-center justify-between gap-4">
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">Swaps</h1>
          <p className="hidden sm:block text-[15px] text-[#888c99]">
            Exchange tokens instantly and view your swap history
          </p>
        </div>
        
        <button
          onClick={() => setIsSwapModalOpen(true)}
          className="bg-blue-600 hover:bg-blue-500 text-white font-semibold px-4 sm:px-6 py-2 sm:py-2.5 rounded-[8px] sm:rounded-full transition-colors flex items-center gap-2 shadow-lg shadow-blue-600/20"
        >
          <Icon icon="hugeicons:arrow-up-down" className="size-5" />
          <span>New Swap</span>
        </button>
      </div>

      {/* Main Container */}
      <div className="w-full bg-[#13151a] rounded-[6px] overflow-hidden flex flex-col">
        
        {/* Header Bar */}
        <div className="p-4 sm:p-6 border-b border-[#22252e] flex items-center justify-between bg-[#16181d]/50">
          <h2 className="text-lg font-bold text-white">Swap History</h2>
          <div className="text-xs text-[#888c99] font-medium">
            Showing {transactions.length} of {total} records
          </div>
        </div>

        {/* Transactions Table / List */}
        <div className="w-full flex flex-col">
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
            <div className="flex flex-col w-full">
              {/* Table Header */}
              <div className="hidden md:grid md:grid-cols-12 gap-4 p-6 border-b border-[#22252e] text-[13px] font-semibold text-[#888c99] uppercase tracking-wider bg-[#1c1f26]/30">
                <div className="col-span-3">Asset</div>
                <div className="col-span-3">Amount</div>
                <div className="col-span-2">Status</div>
                <div className="col-span-2">Tx Hash</div>
                <div className="col-span-2 text-right">Date</div>
              </div>

              {/* Table Body */}
              <div className="flex flex-col divide-y divide-[#22252e]">
                {transactions.map((tx) => {
                  const asset = assetMap.get(tx.asset_id);
                  return (
                    <div
                      key={tx.id}
                      onClick={() => router.push(`/swap/${tx.id}`)}
                      className="flex flex-wrap md:flex-nowrap md:grid md:grid-cols-12 justify-between items-center gap-y-4 md:gap-4 p-6 hover:bg-[#1c1f26]/60 transition-colors cursor-pointer group text-sm"
                    >
                      {/* Asset */}
                      <div className="w-1/2 md:w-auto col-span-3 flex items-center gap-3.5">
                        <div className="size-10 rounded-full flex items-center justify-center border shrink-0 bg-sky-500/10 text-sky-400 border-sky-500/20">
                          <Icon icon="hugeicons:arrow-data-transfer-horizontal" className="size-5" />
                        </div>
                        <div className="flex flex-col">
                          <span className="font-bold text-white capitalize leading-tight">Swap</span>
                          <span className="text-xs text-[#888c99]">
                            {asset ? `${asset.name} (${asset.symbol})` : `Asset #${tx.asset_id}`}
                          </span>
                        </div>
                      </div>

                      {/* Amount */}
                      <div className="w-1/2 md:w-auto col-span-3 flex flex-col items-end md:items-start md:block">
                        <span className="md:hidden text-[11px] text-[#888c99] mb-0.5">Amount</span>
                        <span className="font-mono font-bold text-white block">
                          {parseFloat(tx.amount || '0').toLocaleString(undefined, { maximumFractionDigits: 6 })} {asset?.symbol || ''}
                        </span>
                        {/* Show Status under Amount on mobile instead of Fee */}
                        <div className="mt-1 md:hidden">
                          <span className={`inline-flex items-center px-2.5 py-0.5 rounded-[6px] text-[10px] font-bold uppercase tracking-wider border ${getStatusBadge(tx.status)}`}>
                            {tx.status}
                          </span>
                        </div>
                      </div>

                      {/* Status (Hidden on Mobile, as it's shown under Amount) */}
                      <div className="hidden md:flex w-full md:w-auto col-span-2 justify-between md:flex-col md:block mt-4 md:mt-0 pt-4 md:pt-0 border-t border-[#22252e]/50 md:border-0 items-center md:items-start">
                        <span className="md:hidden text-[11px] text-[#888c99] block">Status</span>
                        <span className={`inline-flex items-center px-2.5 py-1 rounded-[6px] text-xs font-bold uppercase tracking-wider border ${getStatusBadge(tx.status)}`}>
                          {tx.status}
                        </span>
                      </div>

                      {/* Tx Hash (Hidden on Mobile) */}
                      <div className="hidden md:flex w-1/2 md:w-auto col-span-2 flex-col items-end md:items-start md:block mt-2 md:mt-0 pt-2 md:pt-0 border-t border-[#22252e]/50 md:border-0">
                        <span className="md:hidden text-[11px] text-[#888c99] mb-0.5 block">Tx Hash</span>
                        <span className="font-mono text-xs text-blue-400 hover:text-blue-300 max-w-[140px] truncate underline underline-offset-2 block">
                          {tx.id}
                        </span>
                      </div>

                      {/* Date (Hidden on Mobile) */}
                      <div className="hidden md:flex w-full md:w-auto col-span-2 justify-between md:flex-col items-center md:items-end mt-4 md:mt-0 pt-4 md:pt-0 border-t border-[#22252e]/50 md:border-0">
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
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </div>

        {/* Pagination Footer */}
        {totalPages > 1 && (
          <div className="p-4 sm:p-6 border-t border-[#22252e] flex items-center justify-between bg-[#16181d]/50">
            <button
              disabled={page === 1}
              onClick={() => setPage(page - 1)}
              className="px-4 py-2 rounded-[8px] bg-[#1c1f26] text-white text-sm font-semibold hover:bg-[#22252e] disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              Previous
            </button>
            <span className="text-sm font-medium text-[#888c99]">
              Page <span className="text-white">{page}</span> of {totalPages}
            </span>
            <button
              disabled={page === totalPages}
              onClick={() => setPage(page + 1)}
              className="px-4 py-2 rounded-[8px] bg-[#1c1f26] text-white text-sm font-semibold hover:bg-[#22252e] disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
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
