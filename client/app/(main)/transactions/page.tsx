'use client';

import { useState, useEffect, useMemo } from 'react';
import { Icon } from '@iconify/react';
import { transactionsService } from '@/feature/transactions/transactions.service';
import { assetsService } from '@/feature/assets/assets.service';
import { type TransactionItem } from '@/feature/transactions/transactions.schema';
import { type Asset } from '@/feature/assets/assets.schema';

type FilterType = '' | 'deposit' | 'withdrawal' | 'swap' | 'p2p';

export default function TransactionsPage() {
  const [transactions, setTransactions] = useState<TransactionItem[]>([]);
  const [assets, setAssets] = useState<Asset[]>([]);
  const [activeFilter, setActiveFilter] = useState<FilterType>('');
  const [loading, setLoading] = useState(true);
  const [copiedId, setCopiedId] = useState<string | null>(null);

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

  useEffect(() => {
    const fetchTxs = async () => {
      setLoading(true);
      try {
        const res = await transactionsService.getTransactions({
          page,
          limit,
          type: activeFilter || undefined,
        });
        setTransactions(res.transactions || []);
        setTotal(res.total || 0);
      } catch (err) {
        console.error('Failed to fetch transactions:', err);
      } finally {
        setLoading(false);
      }
    };
    fetchTxs();
  }, [page, activeFilter]);

  const assetMap = useMemo(() => {
    const map = new Map<number, Asset>();
    assets.forEach((a) => map.set(a.id, a));
    return map;
  }, [assets]);

  const filterTabs: { label: string; value: FilterType }[] = [
    { label: 'All Transactions', value: '' },
    { label: 'Deposits', value: 'deposit' },
    { label: 'Withdrawals', value: 'withdrawal' },
    { label: 'Swaps', value: 'swap' },
    { label: 'P2P', value: 'p2p' },
  ];

  const handleCopy = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const totalPages = Math.ceil(total / limit) || 1;

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'deposit':
        return { icon: 'hugeicons:arrow-down-01', color: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' };
      case 'withdrawal':
        return { icon: 'hugeicons:arrow-up-01', color: 'bg-blue-500/10 text-blue-400 border-blue-500/20' };
      case 'swap':
        return { icon: 'hugeicons:arrow-data-transfer-horizontal', color: 'bg-purple-500/10 text-purple-400 border-purple-500/20' };
      case 'p2p_buy':
      case 'p2p_sell':
      case 'p2p':
        return { icon: 'hugeicons:user-group', color: 'bg-amber-500/10 text-amber-400 border-amber-500/20' };
      default:
        return { icon: 'hugeicons:exchange-01', color: 'bg-gray-500/10 text-gray-400 border-gray-500/20' };
    }
  };

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
      <div className="flex flex-col gap-1">
        <h1 className="text-3xl font-bold text-white tracking-tight">Transactions</h1>
        <p className="text-[15px] text-[#888c99]">
          View and manage your full activity history across all assets
        </p>
      </div>

      {/* Main Container */}
      <div className="w-full bg-[#13151a] border border-[#22252e] rounded-[24px] overflow-hidden shadow-2xl flex flex-col">
        
        {/* Filters Bar */}
        <div className="p-4 sm:p-6 border-b border-[#22252e] flex items-center justify-between flex-wrap gap-4 bg-[#16181d]/50">
          <div className="flex items-center gap-2 overflow-x-auto pb-1 sm:pb-0 scrollbar-none">
            {filterTabs.map((tab) => (
              <button
                key={tab.value}
                onClick={() => {
                  setActiveFilter(tab.value);
                  setPage(1);
                }}
                className={`px-4 py-2 rounded-full text-sm font-semibold whitespace-nowrap transition-all ${
                  activeFilter === tab.value
                    ? 'bg-blue-600 text-white shadow-lg shadow-blue-600/20'
                    : 'bg-[#1c1f26] text-[#888c99] hover:text-white hover:bg-[#252933]'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>

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
                <Icon icon="hugeicons:bitcoin-transaction" className="size-8" />
              </div>
              <div className="flex flex-col gap-1">
                <h3 className="text-lg font-bold text-white">No transactions found</h3>
                <p className="text-sm text-[#888c99] max-w-sm">
                  You don't have any {activeFilter ? activeFilter : ''} transactions recorded yet.
                </p>
              </div>
            </div>
          ) : (
            <table className="w-full text-left border-collapse min-w-[700px]">
              <thead>
                <tr className="border-b border-[#22252e] text-[13px] font-semibold text-[#888c99] bg-[#16181d]/30">
                  <th className="py-4 px-6">Type & Asset</th>
                  <th className="py-4 px-6">Amount</th>
                  <th className="py-4 px-6">Status</th>
                  <th className="py-4 px-6">Tx Hash / Address</th>
                  <th className="py-4 px-6 text-right">Date</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#22252e]">
                {transactions.map((tx) => {
                  const asset = assetMap.get(tx.asset_id);
                  const iconInfo = getTypeIcon(tx.type);
                  const hashOrAddr = tx.tx_hash || tx.to_address || tx.id;

                  return (
                    <tr
                      key={tx.id}
                      className="hover:bg-[#1c1f26]/60 transition-colors group text-sm"
                    >
                      {/* Type & Asset */}
                      <td className="py-4 px-6">
                        <div className="flex items-center gap-3.5">
                          <div
                            className={`size-10 rounded-full flex items-center justify-center border shrink-0 ${iconInfo.color}`}
                          >
                            <Icon icon={iconInfo.icon} className="size-5" />
                          </div>
                          <div className="flex flex-col">
                            <span className="font-bold text-white capitalize leading-tight">
                              {tx.type.replace('_', ' ')}
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
                          <span
                            className={`font-mono font-bold ${
                              tx.type === 'deposit' ? 'text-emerald-400' : 'text-white'
                            }`}
                          >
                            {tx.type === 'deposit' ? '+' : '-'}{tx.amount} {asset?.symbol || ''}
                          </span>
                          {tx.fee && (
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

                      {/* Hash / Address */}
                      <td className="py-4 px-6">
                        <div className="flex items-center gap-2">
                          <span className="font-mono text-xs text-[#888c99] max-w-[140px] truncate">
                            {hashOrAddr}
                          </span>
                          <button
                            onClick={() => handleCopy(hashOrAddr, tx.id)}
                            className="text-[#888c99] hover:text-white transition-colors"
                            title="Copy string"
                          >
                            <Icon
                              icon={
                                copiedId === tx.id
                                  ? 'hugeicons:checkmark-circle-02'
                                  : 'hugeicons:copy-01'
                              }
                              className={`size-4 ${copiedId === tx.id ? 'text-emerald-400' : ''}`}
                            />
                          </button>
                        </div>
                      </td>

                      {/* Date */}
                      <td className="py-4 px-6 text-right">
                        <div className="flex flex-col items-end">
                          <span className="text-xs font-semibold text-white">
                            {new Date(tx.created_at).toLocaleDateString()}
                          </span>
                          <span className="text-[11px] text-[#888c99]">
                            {new Date(tx.created_at).toLocaleTimeString([], {
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

        {/* Footer / Pagination */}
        {!loading && total > 0 && (
          <div className="p-4 sm:p-6 border-t border-[#22252e] flex items-center justify-between flex-wrap gap-4 bg-[#16181d]/50">
            <span className="text-xs font-medium text-[#888c99]">
              Page {page} of {totalPages}
            </span>

            <div className="flex items-center gap-2">
              <button
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                className="px-4 py-2 rounded-xl bg-[#1c1f26] hover:bg-[#252933] disabled:opacity-40 disabled:hover:bg-[#1c1f26] text-white text-xs font-semibold transition-colors flex items-center gap-1"
              >
                <Icon icon="hugeicons:arrow-left-01" className="size-4" />
                Previous
              </button>
              <button
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
                className="px-4 py-2 rounded-xl bg-[#1c1f26] hover:bg-[#252933] disabled:opacity-40 disabled:hover:bg-[#1c1f26] text-white text-xs font-semibold transition-colors flex items-center gap-1"
              >
                Next
                <Icon icon="hugeicons:arrow-right-01" className="size-4" />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
