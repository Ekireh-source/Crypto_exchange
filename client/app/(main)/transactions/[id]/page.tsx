'use client';

import React, { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { Icon } from '@iconify/react';
import { transactionsService } from '@/feature/transactions/transactions.service';
import { type TransactionItem } from '@/feature/transactions/transactions.schema';
import { assetsService } from '@/feature/assets/assets.service';
import { type Asset } from '@/feature/assets/assets.schema';

export default function TransactionDetailPage() {
  const router = useRouter();
  const params = useParams();
  const id = params.id as string;
  
  const [transaction, setTransaction] = useState<TransactionItem | null>(null);
  const [asset, setAsset] = useState<Asset | null>(null);
  const [loading, setLoading] = useState(true);
  const [copied, setCopied] = useState<string | null>(null);

  useEffect(() => {
    const fetchDetail = async () => {
      try {
        const tx = await transactionsService.getTransactionById(id);
        setTransaction(tx);
        
        const assets = await assetsService.getAssets();
        const found = assets.find(a => a.id === tx.asset_id);
        if (found) setAsset(found);
      } catch (err) {
        console.error('Failed to load transaction detail', err);
      } finally {
        setLoading(false);
      }
    };
    if (id) fetchDetail();
  }, [id]);

  const handleCopy = (text: string, field: string) => {
    navigator.clipboard.writeText(text);
    setCopied(field);
    setTimeout(() => setCopied(null), 2000);
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'deposit': return { icon: 'hugeicons:arrow-down-01', color: 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20' };
      case 'withdrawal': return { icon: 'hugeicons:arrow-up-01', color: 'text-blue-400 bg-blue-500/10 border-blue-500/20' };
      case 'swap': return { icon: 'hugeicons:arrow-data-transfer-horizontal', color: 'text-purple-400 bg-purple-500/10 border-purple-500/20' };
      default: return { icon: 'hugeicons:exchange-01', color: 'text-gray-400 bg-gray-500/10 border-gray-500/20' };
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

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center py-40">
        <Icon icon="hugeicons:loading-03" className="size-8 animate-spin text-[#4f7cf7]" />
      </div>
    );
  }

  if (!transaction) {
    return (
      <div className="flex flex-col items-center justify-center py-40 text-[#888c99] animate-in fade-in duration-500">
        <Icon icon="hugeicons:alert-circle" className="size-12 mb-4" />
        <h2 className="text-xl font-bold text-white mb-2">Transaction not found</h2>
        <button onClick={() => router.back()} className="text-blue-500 hover:underline">Go Back</button>
      </div>
    );
  }

  const iconInfo = getTypeIcon(transaction.type);
  const isDeposit = transaction.type === 'deposit';

  return (
    <div className="flex flex-col w-full max-w-[600px] mx-auto animate-in fade-in duration-500 pb-20 font-sans gap-6 mt-6 px-4 md:px-0">
      <button onClick={() => router.back()} className="flex items-center gap-2 text-[#888c99] hover:text-white transition-colors w-fit font-semibold text-sm">
        <Icon icon="hugeicons:arrow-left-01" className="size-5" />
        <span>Back to transactions</span>
      </button>

      <div className="bg-[#13151a] rounded-[8px] overflow-hidden  flex flex-col items-center text-center">
        <div className="w-full flex flex-col items-center pt-10 pb-8 px-6">
          <div className={`size-16 rounded-full flex items-center justify-center mb-6 border ${iconInfo.color}`}>
            <Icon icon={iconInfo.icon} className="size-8" />
          </div>
          
          <div className="flex flex-col gap-1 mb-8 w-full items-center">
            <span className="text-sm font-semibold text-[#888c99] uppercase tracking-wider">{transaction.type.replace('_', ' ')}</span>
            <h1 className={`text-3xl md:text-4xl font-bold font-mono my-2 break-all px-4 ${isDeposit ? 'text-emerald-400' : 'text-white'}`}>
              {isDeposit ? '+' : '-'}{transaction.amount} {asset?.symbol}
            </h1>
            <span className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wider border ${getStatusBadge(transaction.status)}`}>
              {transaction.status}
            </span>
          </div>
        </div>

        <div className="w-full flex flex-col gap-0 text-left bg-[#16181d] border-t border-[#22252e]">
          <div className="flex justify-between items-center py-5 px-6 border-b border-[#22252e]/50">
            <span className="text-[#888c99] text-sm font-medium">Asset Type</span>
            <span className="text-white font-semibold text-sm">{asset ? `${asset.name} (${asset.symbol})` : 'Unknown Asset'}</span>
          </div>
          
          <div className="flex justify-between items-center py-5 px-6 border-b border-[#22252e]/50">
            <span className="text-[#888c99] text-sm font-medium">Date & Time</span>
            <span className="text-white font-semibold text-sm text-right">
              {new Date(transaction.created_at).toLocaleString()}
            </span>
          </div>

          {(transaction.tx_hash || transaction.to_address) && (
            <div className="flex justify-between items-center py-5 px-6 border-b border-[#22252e]/50">
              <span className="text-[#888c99] text-sm font-medium">Hash / Address</span>
              <div className="flex items-center gap-2">
                <span className="text-white font-mono text-sm max-w-[150px] sm:max-w-[200px] truncate">
                  {transaction.tx_hash || transaction.to_address || transaction.id}
                </span>
                <button 
                  onClick={() => handleCopy(transaction.tx_hash || transaction.to_address || transaction.id, 'hash')}
                  className="text-[#888c99] hover:text-white transition-colors"
                >
                  <Icon icon={copied === 'hash' ? 'hugeicons:checkmark-circle-02' : 'hugeicons:copy-01'} className={`size-4 ${copied === 'hash' ? 'text-emerald-400' : ''}`} />
                </button>
              </div>
            </div>
          )}

          {transaction.fee && parseFloat(transaction.fee) > 0 && (
            <div className="flex justify-between items-center py-5 px-6">
              <span className="text-[#888c99] text-sm font-medium">Network Fee</span>
              <span className="text-white font-mono text-sm">{transaction.fee} {asset?.symbol}</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
