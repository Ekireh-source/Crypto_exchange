'use client';

import { useState, useEffect } from 'react';
import { Icon } from '@iconify/react';
import { walletService } from '@/feature/wallet/wallet.service';
import { assetsService } from '@/feature/assets/assets.service';
import { type Asset } from '@/feature/assets/assets.schema';
import { type PortfolioResponse } from '@/feature/wallet/wallet.schema';
import { toast } from 'sonner';

export default function QuickBuy({ portfolio, refreshPortfolio }: { portfolio: PortfolioResponse | null, refreshPortfolio: () => void }) {
  const [assets, setAssets] = useState<Asset[]>([]);
  const [fromAssetId, setFromAssetId] = useState<number | null>(null);
  const [toAssetId, setToAssetId] = useState<number | null>(null);
  const [amount, setAmount] = useState<string>('');
  const [isSwapping, setIsSwapping] = useState(false);

  // Fetch all assets
  useEffect(() => {
    assetsService.getAssets().then(res => {
      setAssets(res);
      if (res.length >= 2) {
        setFromAssetId(res[0].id);
        setToAssetId(res[1].id);
      }
    }).catch(console.error);
  }, []);

  const handleSwap = async () => {
    if (!fromAssetId || !toAssetId || !amount) return;
    try {
      setIsSwapping(true);
      await walletService.swap({
        from_asset_id: fromAssetId,
        to_asset_id: toAssetId,
        amount: amount
      });
      toast.success("Swap successful!");
      setAmount('');
      refreshPortfolio();
    } catch (error: any) {
      console.error(error);
      const msg = error.response?.data?.message || "Swap failed";
      toast.error(msg);
    } finally {
      setIsSwapping(false);
    }
  };

  const fromAsset = assets.find(a => a.id === fromAssetId);
  const toAsset = assets.find(a => a.id === toAssetId);
  
  // Find balance from portfolio
  const fromBalance = portfolio?.assets.find(a => a.asset.id === fromAssetId)?.available || "0";

  return (
    <div className="bg-[#0a0b0d] rounded-[8px] p-6">
      {/* Tabs */}
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center gap-1 bg-[#16181d] p-1 rounded-full">
          <button className="px-4 py-1.5 rounded-full bg-white text-[#0a0b0d] text-[14px] font-bold shadow-sm">Swap</button>
        </div>
        <button className="flex items-center gap-1 text-[13px] font-bold text-white bg-[#16181d] hover:bg-[#1a1c23] px-3 py-2 rounded-full transition-colors">
          Quick Trade
          <Icon icon="hugeicons:arrow-down-01" className="size-3" />
        </button>
      </div>

      {/* Input Area */}
      <div className="flex items-center justify-between mb-2 border-b border-[#22252e] pb-4 focus-within:border-white transition-colors">
        <input 
          type="number"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          placeholder="0.00"
          className="bg-transparent text-[32px] md:text-[40px] font-medium text-white tracking-tight leading-none outline-none w-[70%] placeholder:text-[#333] [-moz-appearance:_textfield] [&::-webkit-outer-spin-button]:m-0 [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:m-0 [&::-webkit-inner-spin-button]:appearance-none"
        />
        <button 
          onClick={() => setAmount(fromBalance)}
          className="bg-[#16181d] hover:bg-[#1a1c23] text-white text-[13px] font-bold px-3 py-1.5 rounded-full transition-colors shrink-0"
        >
          Max
        </button>
      </div>
      <div className="text-[13px] font-medium text-[#4f7cf7] flex items-center gap-1 mb-8 mt-2">
        <Icon icon="hugeicons:wallet-01" className="size-3" />
        Available: {fromBalance} {fromAsset?.symbol}
      </div>

      {/* Selection Items */}
      <div className="flex flex-col gap-2 mb-6">

        {/* From Asset */}
        <div className="flex items-center justify-between p-3 border border-[#16181d] rounded-[12px] hover:bg-[#16181d] transition-colors group">
          <div className="flex items-center gap-3">
            <div className="size-8 rounded-full bg-[#1e2029] flex items-center justify-center p-1.5 overflow-hidden shrink-0">
              {fromAsset?.logo_url ? <img src={fromAsset.logo_url} className="size-full object-cover" alt="" /> : <Icon icon="hugeicons:bitcoin-01" className="size-full text-[#627eea]" />}
            </div>
            <div className="flex flex-col">
              <span className="text-[14px] font-bold text-white leading-tight">Pay with</span>
              <span className="text-[13px] text-[#888c99] line-clamp-1">{fromAsset?.name || 'Loading...'}</span>
            </div>
          </div>
          <div className="flex items-center gap-2 bg-[#0a0b0d] rounded-lg px-2 py-1 border border-transparent group-hover:border-[#333] transition-colors">
            <select 
              className="bg-transparent text-white font-bold text-[14px] outline-none cursor-pointer appearance-none pr-1"
              value={fromAssetId || ""}
              onChange={(e) => setFromAssetId(Number(e.target.value))}
            >
              {assets.map(a => <option key={a.id} value={a.id} className="bg-[#16181d]">{a.symbol}</option>)}
            </select>
            <Icon icon="hugeicons:arrow-down-01" className="size-4 text-[#888c99] group-hover:text-white" />
          </div>
        </div>

        {/* To Asset */}
        <div className="flex items-center justify-between p-3 border border-[#16181d] rounded-[12px] hover:bg-[#16181d] transition-colors group">
          <div className="flex items-center gap-3">
            <div className="size-8 rounded-full bg-[#1e2029] flex items-center justify-center p-1.5 overflow-hidden shrink-0">
              {toAsset?.logo_url ? <img src={toAsset.logo_url} className="size-full object-cover" alt="" /> : <Icon icon="hugeicons:bitcoin-01" className="size-full text-white" />}
            </div>
            <div className="flex flex-col">
              <span className="text-[14px] font-bold text-white leading-tight">Receive</span>
              <span className="text-[13px] text-[#888c99] line-clamp-1">{toAsset?.name || 'Loading...'}</span>
            </div>
          </div>
          <div className="flex items-center gap-2 bg-[#0a0b0d] rounded-lg px-2 py-1 border border-transparent group-hover:border-[#333] transition-colors">
            <select 
              className="bg-transparent text-white font-bold text-[14px] outline-none cursor-pointer appearance-none pr-1"
              value={toAssetId || ""}
              onChange={(e) => setToAssetId(Number(e.target.value))}
            >
              {assets.map(a => <option key={a.id} value={a.id} className="bg-[#16181d]">{a.symbol}</option>)}
            </select>
            <Icon icon="hugeicons:arrow-down-01" className="size-4 text-[#888c99] group-hover:text-white" />
          </div>
        </div>

      </div>

      {/* CTA Button */}
      <button 
        onClick={handleSwap}
        disabled={!amount || isSwapping || Number(amount) <= 0 || fromAssetId === toAssetId}
        className="w-full bg-[#4f7cf7] hover:bg-[#3f6be7] disabled:bg-[#4f7cf7]/50 disabled:cursor-not-allowed text-white font-semibold text-[15px] rounded-full h-[52px] transition-colors flex items-center justify-center gap-2"
      >
        {isSwapping ? <Icon icon="hugeicons:loading-03" className="animate-spin size-5" /> : "Confirm Swap"}
      </button>
    </div>
  );
}
