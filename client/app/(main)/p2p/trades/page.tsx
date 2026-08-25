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
    <div className="flex flex-col flex-1 h-full max-w-[1440px] mx-auto w-full px-4 sm:px-6 lg:px-8 py-8 gap-8 overflow-y-auto">
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
      <div className="bg-[#13151a] rounded-2xl border border-[#22252e] overflow-hidden flex flex-col flex-1">
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
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="bg-[#1c1f26] border-b border-[#22252e]">
                  <th className="py-4 px-6 text-xs font-semibold text-[#888c99] uppercase tracking-wider">Role</th>
                  <th className="py-4 px-6 text-xs font-semibold text-[#888c99] uppercase tracking-wider">Amount</th>
                  <th className="py-4 px-6 text-xs font-semibold text-[#888c99] uppercase tracking-wider">Fiat Amount</th>
                  <th className="py-4 px-6 text-xs font-semibold text-[#888c99] uppercase tracking-wider">Status</th>
                  <th className="py-4 px-6 text-xs font-semibold text-[#888c99] uppercase tracking-wider">Date</th>
                  <th className="py-4 px-6 text-right"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#22252e]">
                {trades.map((trade) => {
                  const isBuyer = trade.buyer_id === user?.id;
                  const symbol = getAssetSymbol(trade.asset_id);
                  return (
                    <tr 
                      key={trade.id} 
                      onClick={() => router.push(`/p2p/trade/${trade.id}`)}
                      className="hover:bg-[#1c1f26]/60 transition-colors cursor-pointer group"
                    >
                      <td className="py-4 px-6">
                        <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-bold border ${isBuyer ? 'bg-blue-500/10 text-blue-400 border-blue-500/20' : 'bg-purple-500/10 text-purple-400 border-purple-500/20'}`}>
                          {isBuyer ? "Buy" : "Sell"}
                        </span>
                      </td>
                      <td className="py-4 px-6">
                        <p className="text-sm font-bold text-white">{parseFloat(trade.amount).toFixed(4)} {symbol}</p>
                      </td>
                      <td className="py-4 px-6">
                        <p className="text-sm font-medium text-white">{parseFloat(trade.fiat_amount).toLocaleString()} USD</p>
                      </td>
                      <td className="py-4 px-6">
                        <span className={`inline-flex items-center px-2.5 py-1 rounded-lg text-[10px] font-bold uppercase tracking-wider border ${getStatusColor(trade.status)}`}>
                          {trade.status.replace("_", " ")}
                        </span>
                      </td>
                      <td className="py-4 px-6">
                        <span className="text-sm text-[#888c99]">
                          {new Date(trade.created_at).toLocaleString()}
                        </span>
                      </td>
                      <td className="py-4 px-6 text-right">
                        <Icon icon="hugeicons:arrow-right-01" className="size-5 text-[#888c99] group-hover:text-white transition-colors ml-auto" />
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
