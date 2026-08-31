'use client';

import { useState, useEffect } from 'react';
import SupportedAssets from '@/components/commons/dashboard/supportedassets';
import QuickActions from '@/components/commons/dashboard/quickactions';
import Watchlist from '@/components/commons/dashboard/watchlist';
import QuickBuy from '@/components/commons/dashboard/quickbuy';
import RecentActivity from '@/components/commons/dashboard/recentactivity';
import PortfolioChart from '@/components/commons/dashboard/portfoliochart';
import { Icon } from '@iconify/react';
import Image from 'next/image';
import { walletService } from '@/feature/wallet/wallet.service';
import { type PortfolioResponse } from '@/feature/wallet/wallet.schema';

export default function DashboardPage() {
  const [portfolio, setPortfolio] = useState<PortfolioResponse | null>(null);
  const [loading, setLoading] = useState(true);

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

  useEffect(() => {
    fetchPortfolio();
  }, []);
  return (
    <div className="flex flex-col w-full animate-in fade-in duration-500 pb-20">

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">

        {/* Left Column */}
        <div className="lg:col-span-2 flex flex-col gap-8">

          {/* Balance & Assets */}
          <div className="relative">
            {loading ? (
              <div className="flex flex-col gap-2 mb-8">
                <div className="h-10 w-48 bg-[#16181d] rounded animate-pulse" />
                <div className="h-[120px] w-full bg-[#16181d]/50 rounded animate-pulse mt-4" />
              </div>
            ) : (
              <>
                <h2 className="text-[32px] font-medium text-white mb-2 tracking-tight">
                  ${(portfolio?.total_usd_value || 0).toLocaleString('en-US', {
                    minimumFractionDigits: 2,
                    maximumFractionDigits: 2,
                  })}
                </h2>
                <PortfolioChart currentBalance={portfolio?.total_usd_value || 0} />
              </>
            )}

            {/* Asset Allocation or Onboarding */}
            {loading ? (
              <div className="flex flex-col gap-3">
                <div className="h-12 w-full bg-[#16181d] rounded animate-pulse" />
                <div className="h-12 w-full bg-[#16181d] rounded animate-pulse" />
              </div>
            ) : portfolio?.total_usd_value === 0 ? (
              <div className="bg-[#16181d] rounded-[8px] p-8 flex flex-col items-center justify-center text-center">
                <Icon icon="hugeicons:wallet-add-01" className="size-12 text-[#4f7cf7] mb-4" />
                <h3 className="text-[20px] font-bold text-white mb-2">Welcome to your Dashboard</h3>
                <p className="text-[14px] text-[#888c99] mb-6 max-w-md">Your portfolio is currently empty. Add funds or swap assets using the Quick Trade widget to get started.</p>
                <button className="bg-[#4f7cf7] hover:bg-[#3f6be7] text-white font-semibold text-[14px] px-6 py-2.5 rounded-full transition-colors">
                  Deposit Funds
                </button>
              </div>
            ) : (
              <div className="flex flex-col gap-4">
                <h3 className="text-[17px] font-bold text-white mb-2">Asset Allocation</h3>
                {/* Allocation Progress Bar */}
                <div className="flex w-full h-2 rounded-full overflow-hidden mb-4 bg-[#16181d]">
                  {portfolio?.assets.map((a, i) => {
                    const colors = ['bg-[#f7931a]', 'bg-[#627eea]', 'bg-[#10b981]', 'bg-[#8b5cf6]', 'bg-[#ec4899]'];
                    const percentage = (a.usd_value / (portfolio.total_usd_value || 1)) * 100;
                    if (percentage === 0) return null;
                    return <div key={a.asset.id} style={{ width: `${percentage}%` }} className={colors[i % colors.length]} title={`${a.asset.symbol} ${percentage.toFixed(1)}%`} />;
                  })}
                </div>
                {/* Asset Breakdown List */}
                {portfolio?.assets.map((a, i) => {
                  const colors = ['text-[#f7931a]', 'text-[#627eea]', 'text-[#10b981]', 'text-[#8b5cf6]', 'text-[#ec4899]'];
                  const percentage = (a.usd_value / (portfolio.total_usd_value || 1)) * 100;
                  if (percentage === 0) return null;
                  return (
                    <div key={a.asset.id} className="flex items-center justify-between py-2 cursor-pointer group border-b border-[#22252e]/50 last:border-0">
                      <div className="flex items-center gap-3 min-w-0">
                        <div className="size-8 rounded-full bg-[#16181d] flex items-center justify-center overflow-hidden shrink-0">
                          {a.asset.logo_url ? <img src={a.asset.logo_url} className="size-full object-cover" alt="" /> : <Icon icon="hugeicons:bitcoin-01" className={`size-4 ${colors[i % colors.length]}`} />}
                        </div>
                        <div className="flex flex-col min-w-0">
                          <span className="text-[15px] font-semibold text-white leading-tight truncate">{a.asset.name}</span>
                          <span className="text-[13px] text-[#888c99]">{percentage.toFixed(1)}%</span>
                        </div>
                      </div>
                      <div className="flex items-center gap-3">
                        <div className="flex flex-col items-end">
                          <span className="text-[15px] font-semibold text-white">${a.usd_value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>
                          <span className="text-[13px] text-[#888c99]">{parseFloat(a.available).toLocaleString(undefined, { maximumFractionDigits: 6 })} {a.asset.symbol}</span>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          {/* Recent Activity */}
          <div className="mt-4">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-[17px] font-bold text-white">Recent Activity</h3>
            </div>
            <RecentActivity />
          </div>

          {/* Watchlist */}
          <div className="mt-4">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-[17px] font-bold text-white">Watchlist</h3>
              <button className="flex items-center justify-center size-8 rounded-full bg-[#16181d] text-[#888c99] hover:text-white transition-colors">
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
          <QuickBuy portfolio={portfolio} refreshPortfolio={fetchPortfolio} />

          {/* Quick Actions */}
          <QuickActions onActionComplete={fetchPortfolio} />

        </div>

      </div>
    </div>
  );
}
