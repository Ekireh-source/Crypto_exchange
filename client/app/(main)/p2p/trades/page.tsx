"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Icon } from "@iconify/react";
import { p2pService } from "@/feature/p2p/p2p.service";
import { type P2PTrade } from "@/feature/p2p/p2p.schema";
import { walletService } from "@/feature/wallet/wallet.service";
import { type PortfolioResponse } from "@/feature/wallet/wallet.schema";
import { useSelector } from "react-redux";
import { selectUser } from "@/store/authSlice";

export default function MyTrades() {
  const router = useRouter();
  const user = useSelector(selectUser);
  const [trades, setTrades] = useState<P2PTrade[]>([]);
  const [portfolio, setPortfolio] = useState<PortfolioResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [tradesRes, portfolioRes] = await Promise.all([
        p2pService.listMyTrades(),
        walletService.getPortfolio(),
      ]);
      setTrades(tradesRes);
      setPortfolio(portfolioRes);
    } catch (error) {
      console.error("Failed to fetch trades:", error);
    } finally {
      setLoading(false);
    }
  };

  const getAssetSymbol = (assetId: number) => {
    return portfolio?.assets.find((a) => a.asset.id === assetId)?.asset.symbol || "Asset";
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "waiting_payment": return "bg-amber-500/10 text-amber-400 border-amber-500/20";
      case "paid": return "bg-blue-500/10 text-blue-400 border-blue-500/20";
      case "released": return "bg-emerald-500/10 text-emerald-400 border-emerald-500/20";
      case "cancelled": return "bg-red-500/10 text-red-400 border-red-500/20";
      default: return "bg-[#22252e] text-[#888c99] border-[#22252e]";
    }
  };

  return (
    <div className="flex flex-col w-full animate-in fade-in duration-500 pb-20 gap-8">
      {/* Header */}
      <div className="flex items-center gap-4">
        <button
          onClick={() => router.push("/p2p")}
          className="size-10 rounded-full bg-[#1c1f26] hover:bg-[#22252e] flex items-center justify-center text-white transition-colors border border-[#22252e]"
        >
          <Icon icon="hugeicons:arrow-left-01" className="size-5" />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-white">My Trades</h1>
          <p className="text-sm text-[#888c99]">View your active and past P2P transactions</p>
        </div>
      </div>

      {/* Trades List */}
      <div className="bg-[#13151a] rounded-[6px] flex flex-col">
        {loading ? (
          <div className="p-12 flex items-center justify-center">
            <Icon icon="svg-spinners:180-ring" className="size-8 text-blue-500" />
          </div>
        ) : trades.length === 0 ? (
          <div className="p-12 flex flex-col items-center justify-center gap-3">
            <Icon icon="hugeicons:trade-down" className="size-12 text-[#888c99]/50" />
            <p className="text-[#888c99] font-medium text-center max-w-sm">
              You don't have any trades yet.
            </p>
          </div>
        ) : (
          <div className="flex flex-col w-full">
            {/* Table Header */}
            <div className="flex md:grid md:grid-cols-12 justify-between gap-4 p-6 border-b border-[#22252e] text-[13px] font-semibold text-[#888c99] uppercase tracking-wider bg-[#1c1f26]/30">
              <div className="col-span-2">Role</div>
              <div className="col-span-3">Amount</div>
              <div className="col-span-3 hidden md:block">Fiat Amount</div>
              <div className="col-span-2 hidden md:block">Status</div>
              <div className="col-span-2 hidden md:block">Date</div>
            </div>

            {/* Table Body */}
            <div className="flex flex-col divide-y divide-[#22252e]">
              {trades.map((trade) => {
                const isBuyer = trade.buyer_id === user?.id;
                const symbol = getAssetSymbol(trade.asset_id);
                return (
                  <div 
                    key={trade.id} 
                    onClick={() => router.push(`/p2p/trade/${trade.id}`)}
                    className="flex flex-wrap md:flex-nowrap md:grid md:grid-cols-12 justify-between items-center gap-y-4 md:gap-4 p-6 hover:bg-[#1c1f26]/60 transition-colors cursor-pointer group"
                  >
                    {/* Role */}
                    <div className="w-1/2 md:w-auto col-span-2">
                      <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-[6px] text-xs font-bold border ${isBuyer ? 'bg-blue-500/10 text-blue-400 border-blue-500/20' : 'bg-sky-500/10 text-sky-400 border-sky-500/20'}`}>
                        {isBuyer ? "Buy" : "Sell"}
                      </span>
                    </div>

                    {/* Amount */}
                    <div className="w-1/2 md:w-auto col-span-3 flex flex-col items-end md:items-start md:block">
                      <span className="md:hidden text-[11px] text-[#888c99] mb-0.5">Amount</span>
                      <p className="text-sm font-bold text-white">{parseFloat(trade.amount).toFixed(4)} {symbol}</p>
                    </div>

                    {/* Fiat Amount */}
                    <div className="w-1/2 md:w-auto col-span-3 flex flex-col md:block mt-2 md:mt-0 pt-2 md:pt-0 border-t border-[#22252e]/50 md:border-0">
                      <span className="md:hidden text-[11px] text-[#888c99] mb-0.5">Fiat Amount</span>
                      <p className="text-sm font-medium text-white">{parseFloat(trade.fiat_amount).toLocaleString()} USD</p>
                    </div>

                    {/* Status */}
                    <div className="w-1/2 md:w-auto col-span-2 flex flex-col items-end md:items-start md:block mt-2 md:mt-0 pt-2 md:pt-0 border-t border-[#22252e]/50 md:border-0">
                      <span className="md:hidden text-[11px] text-[#888c99] mb-0.5">Status</span>
                      <span className={`inline-flex items-center px-2.5 py-1 rounded-[6px] text-[10px] font-bold uppercase tracking-wider border ${getStatusColor(trade.status)}`}>
                        {trade.status.replace("_", " ")}
                      </span>
                    </div>

                    {/* Date */}
                    <div className="col-span-1 hidden md:block">
                      <span className="text-xs text-[#888c99]">
                        {new Date(trade.created_at).toLocaleDateString()}
                      </span>
                    </div>

                    {/* Arrow */}
                    <div className="col-span-1 hidden md:flex justify-end">
                      <Icon icon="hugeicons:arrow-right-01" className="size-5 text-[#888c99] group-hover:text-white transition-colors" />
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
