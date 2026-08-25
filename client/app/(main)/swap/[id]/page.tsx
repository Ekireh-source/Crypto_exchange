'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Icon } from '@iconify/react';
import { transactionsService } from '@/feature/transactions/transactions.service';
import { assetsService } from '@/feature/assets/assets.service';
import { type TransactionItem } from '@/feature/transactions/transactions.schema';
import { type Asset } from '@/feature/assets/assets.schema';

export default function SwapDetailPage() {
  const { id } = useParams() as { id: string };
  const router = useRouter();

  const [transaction, setTransaction] = useState<TransactionItem | null>(null);
  const [asset, setAsset] = useState<Asset | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        // Fetch transaction details
        const txData = await transactionsService.getTransactionById(id);
        setTransaction(txData);

        // Fetch asset info for this transaction
        if (txData && txData.asset_id) {
          const assetsData = await assetsService.getAssets();
          const foundAsset = assetsData.find(a => a.id === txData.asset_id);
          if (foundAsset) setAsset(foundAsset);
        }
      } catch (err: any) {
        console.error('Failed to fetch transaction details:', err);
        setError(err.response?.data?.error || 'Failed to load details');
      } finally {
        setLoading(false);
      }
    };
    if (id) {
      fetchData();
    }
  }, [id]);

  if (loading) {
    return (
      <div className="flex justify-center pt-20">
        <div className="animate-spin text-[#4f7cf7]">
          <Icon icon="hugeicons:loading-03" className="size-8" />
        </div>
      </div>
    );
  }

  if (error || !transaction) {
    return (
      <div className="flex flex-col items-center justify-center pt-20 gap-4">
        <div className="text-red-400 p-4 bg-red-500/10 rounded-xl border border-red-500/20">
          {error || 'Transaction not found'}
        </div>
        <button
          onClick={() => router.back()}
          className="text-[#888c99] hover:text-white flex items-center gap-2"
        >
          <Icon icon="hugeicons:arrow-left-01" /> Back
        </button>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center w-full max-w-2xl mx-auto animate-in fade-in duration-500 pt-10 pb-20 px-4">
      
      {/* Back Button */}
      <div className="w-full mb-6">
        <button
          onClick={() => router.push('/swap')}
          className="text-[#888c99] hover:text-white flex items-center gap-2 text-sm font-medium transition-colors w-fit"
        >
          <Icon icon="hugeicons:arrow-left-01" className="size-4" />
          Back to Swaps
        </button>
      </div>

      <div className="w-full bg-[#13151a] border border-[#22252e] rounded-3xl p-6 md:p-8 shadow-2xl relative">
        <div className="flex flex-col items-center gap-4 mb-10 text-center">
          <div className="size-16 rounded-full bg-purple-500/10 border border-purple-500/20 text-purple-400 flex items-center justify-center">
            <Icon icon="hugeicons:arrow-data-transfer-horizontal" className="size-8" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-white mb-1">Swap Transaction</h1>
            <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wider border bg-emerald-500/10 text-emerald-400 border-emerald-500/20">
              {transaction.status}
            </span>
          </div>
        </div>

        <div className="flex flex-col gap-4">
          
          <div className="flex justify-between items-center p-4 bg-[#1c1f26] border border-[#22252e] rounded-2xl">
            <span className="text-[#888c99] text-sm font-semibold">Amount</span>
            <span className="text-white font-mono font-bold text-lg">
              {transaction.amount} {asset?.symbol}
            </span>
          </div>

          <div className="flex justify-between items-center p-4 bg-[#1c1f26] border border-[#22252e] rounded-2xl">
            <span className="text-[#888c99] text-sm font-semibold">Asset</span>
            <div className="flex items-center gap-2">
              <span className="text-white font-medium">{asset?.name || `Asset ID: ${transaction.asset_id}`}</span>
              <span className="text-[#888c99] text-xs px-2 py-0.5 bg-[#22252e] rounded-md">{asset?.network}</span>
            </div>
          </div>

          <div className="flex justify-between items-center p-4 bg-[#1c1f26] border border-[#22252e] rounded-2xl">
            <span className="text-[#888c99] text-sm font-semibold">Fee</span>
            <span className="text-white font-mono">{transaction.fee || '0'} {asset?.symbol}</span>
          </div>

          <div className="flex justify-between items-center p-4 bg-[#1c1f26] border border-[#22252e] rounded-2xl">
            <span className="text-[#888c99] text-sm font-semibold">Swap Hash ID</span>
            <span className="text-white font-mono text-xs max-w-[200px] truncate md:max-w-xs">{transaction.tx_hash}</span>
          </div>

          <div className="flex justify-between items-center p-4 bg-[#1c1f26] border border-[#22252e] rounded-2xl">
            <span className="text-[#888c99] text-sm font-semibold">Date & Time</span>
            <div className="flex flex-col items-end">
              <span className="text-white font-medium">
                {new Date(transaction.created_at).toLocaleDateString(undefined, {
                  month: 'long',
                  day: 'numeric',
                  year: 'numeric',
                })}
              </span>
              <span className="text-[#888c99] text-xs">
                {new Date(transaction.created_at).toLocaleTimeString(undefined, {
                  hour: '2-digit',
                  minute: '2-digit',
                  second: '2-digit'
                })}
              </span>
            </div>
          </div>

        </div>
      </div>

    </div>
  );
}
