'use client';

import { useState, useEffect } from 'react';
import { Icon } from '@iconify/react';
import { walletService } from '@/feature/wallet/wallet.service';
import { type Transaction } from '@/feature/wallet/wallet.schema';
import { assetsService } from '@/feature/assets/assets.service';
import { type Asset } from '@/feature/assets/assets.schema';

export default function RecentActivity() {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [assets, setAssets] = useState<Asset[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      walletService.getTransactions(5),
      assetsService.getAssets()
    ]).then(([txRes, assetsRes]) => {
      setTransactions(txRes.transactions);
      setAssets(assetsRes);
    }).catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  const getAsset = (assetId: number) => assets.find(a => a.id === assetId);

  const getIcon = (type: string) => {
    switch (type) {
      case 'deposit': return 'hugeicons:arrow-down-01';
      case 'withdrawal': return 'hugeicons:arrow-up-01';
      case 'swap': return 'hugeicons:arrow-data-transfer-horizontal';
      case 'p2p_buy': return 'hugeicons:trade-up';
      case 'p2p_sell': return 'hugeicons:trade-down';
      default: return 'hugeicons:transaction';
    }
  };

  const getTitle = (type: string) => {
    switch (type) {
      case 'deposit': return 'Deposit';
      case 'withdrawal': return 'Withdrawal';
      case 'swap': return 'Swap';
      case 'p2p_buy': return 'P2P Buy';
      case 'p2p_sell': return 'P2P Sell';
      default: return 'Transaction';
    }
  };

  const getAmountColor = (type: string) => {
    if (['deposit', 'p2p_buy'].includes(type)) return 'text-emerald-500';
    if (['withdrawal', 'p2p_sell'].includes(type)) return 'text-red-500';
    return 'text-white';
  };

  const getAmountPrefix = (type: string) => {
    if (['deposit', 'p2p_buy'].includes(type)) return '+';
    if (['withdrawal', 'p2p_sell'].includes(type)) return '-';
    return '';
  };

  if (loading) {
    return (
      <div className="w-full flex flex-col gap-4 animate-pulse">
        {[1,2,3].map(i => (
          <div key={i} className="flex items-center justify-between py-3 border-b border-[#22252e]/50 last:border-0">
            <div className="flex items-center gap-3">
              <div className="size-10 rounded-full bg-[#16181d]" />
              <div className="flex flex-col gap-2">
                <div className="h-4 w-24 bg-[#16181d] rounded" />
                <div className="h-3 w-16 bg-[#16181d] rounded" />
              </div>
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (transactions.length === 0) {
    return (
      <div className="text-center py-8 bg-[#0a0b0d] rounded-[8px]">
        <Icon icon="hugeicons:transaction" className="size-10 text-[#22252e] mx-auto mb-3" />
        <p className="text-[14px] font-semibold text-[#888c99]">No recent activity</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col bg-[#0a0b0d] rounded-[8px] p-4">
      {transactions.map((tx) => {
        const asset = getAsset(tx.asset_id);
        const date = new Date(tx.created_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
        
        return (
          <div key={tx.id} className="flex items-center justify-between py-3 group cursor-pointer border-b border-[#22252e]/50 last:border-0">
            <div className="flex items-center gap-3">
              <div className="size-10 rounded-full bg-[#16181d] flex items-center justify-center shrink-0">
                <Icon icon={getIcon(tx.type)} className="size-5 text-[#888c99] group-hover:text-white transition-colors" />
              </div>
              <div className="flex flex-col">
                <span className="text-[15px] font-semibold text-white group-hover:text-blue-400 transition-colors">{getTitle(tx.type)}</span>
                <span className="text-[13px] text-[#888c99]">{date} • <span className="capitalize">{tx.status}</span></span>
              </div>
            </div>
            <div className="flex flex-col items-end">
              <span className={`text-[15px] font-bold ${getAmountColor(tx.type)}`}>
                {getAmountPrefix(tx.type)}{parseFloat(tx.amount).toLocaleString(undefined, { maximumFractionDigits: 4 })} {asset?.symbol}
              </span>
            </div>
          </div>
        );
      })}
    </div>
  );
}
