'use client';

import { useState, useEffect } from 'react';
import SupportedAssets from '@/components/commons/dashboard/supportedassets';
import QuickActions from '@/components/commons/dashboard/quickactions';
import Watchlist from '@/components/commons/dashboard/watchlist';
import { Icon } from '@iconify/react';
import Image from 'next/image';
import { walletService } from '@/feature/wallet/wallet.service';
import { type PortfolioResponse } from '@/feature/wallet/wallet.schema';

export default function DashboardPage() {
  const [portfolio, setPortfolio] = useState<PortfolioResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchPortfolio = async () => {
      try {
        const data = await walletService.getPortfolio();
        setPortfolio(data);
      } catch (err) {
        console.error('Failed to fetch portfolio balances:', err);
      } finally {
        setLoading(false);
      }
    };
    fetchPortfolio();
  }, []);
  return (
    <div className="flex flex-col w-full animate-in fade-in duration-500 pb-20">

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">

        {/* Left Column */}
        <div className="lg:col-span-2 flex flex-col gap-8">

          {/* Balance & Assets */}
          <div>
            <h2 className="text-[32px] font-medium text-white mb-8 tracking-tight">
              {loading ? (
                <span className="text-[#888c99] animate-pulse">$0.00</span>
              ) : (
                `$${(portfolio?.total_usd_value || 0).toLocaleString('en-US', {
                  minimumFractionDigits: 2,
                  maximumFractionDigits: 2,
                })}`
              )}
            </h2>

            <div className="flex flex-col gap-2">
              {/* Asset Item: Crypto */}
              <div className="flex items-center justify-between py-3 cursor-pointer group">
                <div className="flex items-center gap-3">
                  <div className="size-8 rounded-full bg-[#16181d] border border-[#22252e] flex items-center justify-center">
                    <Icon icon="hugeicons:bitcoin-01" className="size-4 text-[#888c99]" />
                  </div>
                  <span className="text-[15px] font-semibold text-white">Crypto</span>
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-[15px] font-semibold text-white">
                    ${(portfolio?.total_usd_value || 0).toFixed(2)}
                  </span>
                  <Icon icon="hugeicons:arrow-right-01" className="size-5 text-[#888c99] group-hover:text-white transition-colors" />
                </div>
              </div>

              {/* Asset Item: Cash */}
              <div className="flex items-center justify-between py-3 cursor-pointer group">
                <div className="flex items-center gap-3">
                  <div className="size-8 rounded-full bg-[#16181d] border border-[#22252e] flex items-center justify-center">
                    <Icon icon="hugeicons:cash-01" className="size-4 text-[#888c99]" />
                  </div>
                  <span className="text-[15px] font-semibold text-white">Cash <span className="text-emerald-500 font-medium">· 4.64% APY</span></span>
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-[15px] font-semibold text-blue-500">Deposit</span>
                  <Icon icon="hugeicons:arrow-right-01" className="size-5 text-[#888c99] group-hover:text-white transition-colors" />
                </div>
              </div>

              {/* Asset Item: Derivatives */}
              <div className="flex items-center justify-between py-3 cursor-pointer group">
                <div className="flex items-center gap-3">
                  <div className="size-8 rounded-full bg-[#16181d] border border-[#22252e] flex items-center justify-center">
                    <Icon icon="hugeicons:chart-line-up-01" className="size-4 text-[#888c99]" />
                  </div>
                  <span className="text-[15px] font-semibold text-white">Derivatives</span>
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-[15px] font-semibold text-[#888c99]">0 positions</span>
                  <Icon icon="hugeicons:arrow-right-01" className="size-5 text-[#888c99] group-hover:text-white transition-colors" />
                </div>
              </div>
            </div>
          </div>

          {/* Earn Banner */}
          <div className="mt-2 bg-[#16181d] border border-[#22252e] rounded-[16px] p-4 flex items-center justify-between cursor-pointer hover:bg-[#1a1c23] transition-colors">
            <div className="flex items-center gap-3">
              <div className="size-10 rounded-full border-4 border-[#1e2029] bg-[#2d313f]" />
              <span className="text-[15px] font-semibold text-white">Earn</span>
            </div>
            <span className="text-[15px] font-semibold text-[#888c99]">up to 13.38% APY</span>
          </div>

          {/* Watchlist */}
          <div className="mt-4">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-[17px] font-bold text-white">Watchlist</h3>
              <button className="flex items-center justify-center size-8 rounded-full bg-[#16181d] border border-[#22252e] text-[#888c99] hover:text-white transition-colors">
                <Icon icon="hugeicons:arrow-right-01" className="size-4" />
              </button>
            </div>

            <Watchlist />
          </div>

          <SupportedAssets />

        </div>

        {/* Right Column */}
        <div className="lg:col-span-1 flex flex-col gap-6">

          {/* Quick Buy Widget */}
          <div className="bg-[#0a0b0d] border border-[#22252e] rounded-[16px] p-6">

            {/* Tabs */}
            <div className="flex items-center justify-between mb-8">
              <div className="flex items-center gap-1 bg-[#16181d] p-1 rounded-full border border-[#22252e]">
                <button className="px-4 py-1.5 rounded-full bg-white text-[#0a0b0d] text-[14px] font-bold">Buy</button>
                <button className="px-4 py-1.5 rounded-full text-[#888c99] hover:text-white text-[14px] font-bold transition-colors">Sell</button>
                <button className="px-4 py-1.5 rounded-full text-[#888c99] hover:text-white text-[14px] font-bold transition-colors">Convert</button>
              </div>
              <button className="flex items-center gap-1 text-[13px] font-bold text-white bg-[#16181d] hover:bg-[#1a1c23] border border-[#22252e] px-3 py-2 rounded-full transition-colors">
                Quick buy
                <Icon icon="hugeicons:arrow-down-01" className="size-3" />
              </button>
            </div>

            {/* Input Area */}
            <div className="flex items-center justify-between mb-2">
              <div className="text-[48px] font-medium text-white tracking-tight leading-none">0<span className="text-[#888c99]">UGX</span></div>
              <button className="bg-[#16181d] hover:bg-[#1a1c23] border border-[#22252e] text-white text-[13px] font-bold px-3 py-1.5 rounded-full transition-colors">
                Max
              </button>
            </div>
            <div className="text-[13px] font-medium text-[#4f7cf7] flex items-center gap-1 mb-8">
              <Icon icon="hugeicons:arrow-data-transfer-horizontal" className="size-3" />
              0 BTC
            </div>

            {/* Selection Items */}
            <div className="flex flex-col gap-2 mb-6">

              <div className="flex items-center justify-between p-3 rounded-[12px] hover:bg-[#16181d] cursor-pointer transition-colors border border-transparent hover:border-[#22252e] group">
                <div className="flex items-center gap-3">
                  <div className="size-8 rounded-full bg-[#1e2029] flex items-center justify-center p-1.5">
                    <Icon icon="hugeicons:ethereum" className="size-full text-[#627eea]" />
                  </div>
                  <div className="flex flex-col">
                    <span className="text-[14px] font-bold text-white leading-tight">Pay with</span>
                    <span className="text-[13px] text-[#888c99]">Ethereum</span>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <div className="flex flex-col items-end">
                    <span className="text-[14px] font-bold text-white leading-tight">UGX 0</span>
                    <span className="text-[12px] font-semibold text-[#4f7cf7] flex items-center gap-1">Available <Icon icon="hugeicons:information-circle" className="size-3" /></span>
                  </div>
                  <Icon icon="hugeicons:arrow-right-01" className="size-4 text-[#888c99] group-hover:text-white" />
                </div>
              </div>

              <div className="flex items-center justify-between p-3 rounded-[12px] hover:bg-[#16181d] cursor-pointer transition-colors border border-transparent hover:border-[#22252e] group">
                <div className="flex items-center gap-3">
                  <div className="size-8 rounded-full bg-[#f7931a] flex items-center justify-center p-1.5">
                    <Icon icon="hugeicons:bitcoin-01" className="size-full text-white" />
                  </div>
                  <div className="flex flex-col">
                    <span className="text-[14px] font-bold text-white leading-tight">Buy</span>
                    <span className="text-[13px] text-[#888c99]">Bitcoin</span>
                  </div>
                </div>
                <Icon icon="hugeicons:arrow-right-01" className="size-4 text-[#888c99] group-hover:text-white" />
              </div>

            </div>

            {/* CTA Button */}
            <button className="w-full bg-[#4f7cf7] hover:bg-[#3f6be7] text-white font-semibold text-[15px] rounded-full h-[52px] transition-colors">
              Review order
            </button>
          </div>

          {/* Quick Actions */}
          <QuickActions />

        </div>

      </div>
    </div>
  );
}
