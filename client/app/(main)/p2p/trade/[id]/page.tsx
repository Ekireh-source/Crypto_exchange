"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Icon } from "@iconify/react";
import { p2pService } from "@/feature/p2p/p2p.service";
import { type P2PTrade } from "@/feature/p2p/p2p.schema";
import { useSelector } from "react-redux";
import { selectUser } from "@/store/authSlice";
import { toast } from "sonner";

export default function TradeRoom() {
  const { id } = useParams();
  const router = useRouter();
  const user = useSelector(selectUser);
  const [trade, setTrade] = useState<P2PTrade | null>(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);

  useEffect(() => {
    if (id) {
      fetchTrade();
      // Simple polling for updates
      const interval = setInterval(fetchTrade, 5000);
      return () => clearInterval(interval);
    }
  }, [id]);

  const fetchTrade = async () => {
    try {
      const res = await p2pService.getTrade(id as string);
      setTrade(res);
    } catch (error) {
      console.error("Failed to fetch trade:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleAction = async (action: 'pay' | 'release' | 'cancel') => {
    if (!trade) return;
    try {
      setActionLoading(true);
      if (action === 'pay') {
        await p2pService.markTradePaid(trade.id);
      } else if (action === 'release') {
        await p2pService.releaseTrade(trade.id);
      } else if (action === 'cancel') {
        await p2pService.cancelTrade(trade.id);
      }
      toast.success(`Trade ${action} action successful`);
      await fetchTrade();
    } catch (error: any) {
      toast.error(error.response?.data?.message || `Failed to ${action} trade`);
    } finally {
      setActionLoading(false);
    }
  };

  if (loading && !trade) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <Icon icon="svg-spinners:180-ring" className="size-8 text-blue-500" />
      </div>
    );
  }

  if (!trade) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center gap-4">
        <Icon icon="hugeicons:file-not-found" className="size-12 text-[#888c99]" />
        <h2 className="text-xl font-bold text-white">Trade Not Found</h2>
        <button onClick={() => router.push("/p2p")} className="text-blue-500 hover:underline">Back to P2P</button>
      </div>
    );
  }

  const isBuyer = user?.id === trade.buyer_id;

  const getStatusDisplay = () => {
    switch (trade.status) {
      case "waiting_payment":
        return { text: "Waiting for Payment", color: "text-amber-500", icon: "hugeicons:time-02", bg: "bg-amber-500/10 border-amber-500/20" };
      case "paid":
        return { text: "Paid - Waiting Release", color: "text-blue-500", icon: "hugeicons:coins-swap", bg: "bg-blue-500/10 border-blue-500/20" };
      case "released":
        return { text: "Completed", color: "text-emerald-500", icon: "hugeicons:checkmark-circle-02", bg: "bg-emerald-500/10 border-emerald-500/20" };
      case "cancelled":
        return { text: "Cancelled", color: "text-red-500", icon: "hugeicons:cancel-circle", bg: "bg-red-500/10 border-red-500/20" };
      default:
        return { text: trade.status, color: "text-[#888c99]", icon: "hugeicons:information-circle", bg: "bg-[#22252e] border-[#22252e]" };
    }
  };

  const statusInfo = getStatusDisplay();

  return (
    <div className="flex flex-col w-full animate-in fade-in duration-500 pb-20 gap-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <button
            onClick={() => router.push("/p2p/trades")}
            className="size-10 rounded-full bg-[#1c1f26] hover:bg-[#22252e] flex items-center justify-center text-white transition-colors border border-[#22252e]"
          >
            <Icon icon="hugeicons:arrow-left-01" className="size-5" />
          </button>
          <div>
            <h1 className="text-2xl font-bold text-white">Trade #{trade.id.substring(0, 8)}</h1>
            <p className="text-sm text-[#888c99]">You are the {isBuyer ? "Buyer" : "Seller"}</p>
          </div>
        </div>
        <div className={`flex items-center gap-2 px-4 py-2 rounded-[8px] border ${statusInfo.bg}`}>
          <Icon icon={statusInfo.icon} className={`size-5 ${statusInfo.color}`} />
          <span className={`text-sm font-bold ${statusInfo.color}`}>{statusInfo.text}</span>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        
        {/* Trade Details Panel */}
        <div className="bg-[#13151a] rounded-[6px] p-6 flex flex-col gap-6 h-fit">
          <h2 className="text-lg font-bold text-white border-b border-[#22252e] pb-4">Trade Summary</h2>
          
          <div className="flex flex-col gap-4">
            <div className="flex justify-between items-center">
              <span className="text-[#888c99] text-sm">Crypto Amount</span>
              <span className="text-white font-bold">{parseFloat(trade.amount).toFixed(4)}</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-[#888c99] text-sm">Fiat Amount</span>
              <span className="text-blue-500 text-lg font-bold">{parseFloat(trade.fiat_amount).toLocaleString()} USD</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-[#888c99] text-sm">Exchange Rate</span>
              <span className="text-white font-medium">{parseFloat(trade.rate).toLocaleString()} USD</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-[#888c99] text-sm">Created At</span>
              <span className="text-white font-medium">{new Date(trade.created_at).toLocaleString()}</span>
            </div>
          </div>

          {trade.escrow_locked && (
            <div className="mt-2 p-3 bg-blue-600/10 rounded-[8px] border border-blue-500/20 flex items-start gap-3">
              <Icon icon="hugeicons:lock-key" className="size-5 text-blue-400 shrink-0 mt-0.5" />
              <p className="text-sm text-blue-400">Crypto is securely locked in escrow. It will be released when the seller confirms payment.</p>
            </div>
          )}
        </div>

        {/* Action Panel */}
        <div className="bg-[#13151a] rounded-[6px] p-6 flex flex-col gap-6">
          <h2 className="text-lg font-bold text-white border-b border-[#22252e] pb-4">Action Required</h2>
          
          {isBuyer ? (
            // BUYER VIEW
            <div className="flex flex-col gap-6">
              {trade.status === "waiting_payment" ? (
                <>
                  <p className="text-sm text-[#888c99]">
                    Please pay <span className="font-bold text-white">{parseFloat(trade.fiat_amount).toLocaleString()} USD</span> to the seller using the agreed payment method outside of the exchange, then click the button below.
                  </p>
                  <div className="flex flex-col gap-3 mt-4">
                    <button
                      onClick={() => handleAction("pay")}
                      disabled={actionLoading}
                      className="w-full py-4 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white font-bold rounded-[8px] transition-all shadow-lg shadow-blue-600/20 flex justify-center items-center gap-2"
                    >
                      {actionLoading ? <Icon icon="svg-spinners:180-ring" className="size-5" /> : "I Have Paid"}
                    </button>
                    <button
                      onClick={() => handleAction("cancel")}
                      disabled={actionLoading}
                      className="w-full py-4 bg-[#1c1f26] hover:bg-red-500/10 hover:text-red-400 text-[#888c99] font-bold rounded-[8px] transition-all flex justify-center items-center gap-2 border border-[#22252e] hover:border-red-500/30"
                    >
                      Cancel Trade
                    </button>
                  </div>
                </>
              ) : trade.status === "paid" ? (
                <div className="flex flex-col items-center justify-center gap-4 py-8 text-center">
                  <div className="size-16 rounded-full bg-blue-500/10 flex items-center justify-center border border-blue-500/30 relative">
                    <Icon icon="hugeicons:time-02" className="size-8 text-blue-500" />
                    <span className="absolute top-0 right-0 flex h-3 w-3">
                      <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
                      <span className="relative inline-flex rounded-full h-3 w-3 bg-blue-500"></span>
                    </span>
                  </div>
                  <h3 className="text-lg font-bold text-white">Waiting for Seller</h3>
                  <p className="text-sm text-[#888c99]">You have marked this trade as paid. The seller is verifying the payment and will release the crypto shortly.</p>
                </div>
              ) : trade.status === "released" ? (
                <div className="flex flex-col items-center justify-center gap-4 py-8 text-center">
                  <div className="size-16 rounded-full bg-emerald-500/10 flex items-center justify-center border border-emerald-500/30">
                    <Icon icon="hugeicons:checkmark-circle-02" className="size-8 text-emerald-500" />
                  </div>
                  <h3 className="text-lg font-bold text-white">Trade Completed!</h3>
                  <p className="text-sm text-[#888c99]">The crypto has been released to your available balance.</p>
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center gap-4 py-8 text-center">
                  <Icon icon="hugeicons:cancel-circle" className="size-12 text-red-500" />
                  <h3 className="text-lg font-bold text-white">Trade Cancelled</h3>
                  <p className="text-sm text-[#888c99]">This trade was cancelled and the escrow has been returned.</p>
                </div>
              )}
            </div>
          ) : (
            // SELLER VIEW
            <div className="flex flex-col gap-6">
              {trade.status === "waiting_payment" ? (
                <div className="flex flex-col items-center justify-center gap-4 py-8 text-center">
                  <div className="size-16 rounded-full bg-amber-500/10 flex items-center justify-center border border-amber-500/30 relative">
                    <Icon icon="hugeicons:time-02" className="size-8 text-amber-500" />
                    <span className="absolute top-0 right-0 flex h-3 w-3">
                      <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-amber-400 opacity-75"></span>
                      <span className="relative inline-flex rounded-full h-3 w-3 bg-amber-500"></span>
                    </span>
                  </div>
                  <h3 className="text-lg font-bold text-white">Waiting for Buyer</h3>
                  <p className="text-sm text-[#888c99]">The buyer is sending <span className="font-bold text-white">{parseFloat(trade.fiat_amount).toLocaleString()} USD</span> to you. Wait for them to mark the trade as paid.</p>
                </div>
              ) : trade.status === "paid" ? (
                <>
                  <div className="p-4 bg-amber-500/10 rounded-[8px] border border-amber-500/20 flex flex-col gap-2">
                    <h4 className="font-bold text-amber-500 flex items-center gap-2">
                      <Icon icon="hugeicons:alert-02" className="size-5" />
                      Verify Payment First!
                    </h4>
                    <p className="text-sm text-amber-400/80">
                      Please check your bank account or payment app to confirm you have actually received <span className="font-bold text-amber-500">{parseFloat(trade.fiat_amount).toLocaleString()} USD</span> from the buyer. Do not release crypto until you verify the funds.
                    </p>
                  </div>
                  <div className="flex flex-col gap-3 mt-4">
                    <button
                      onClick={() => handleAction("release")}
                      disabled={actionLoading}
                      className="w-full py-4 bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white font-bold rounded-[8px] transition-all shadow-lg shadow-emerald-600/20 flex justify-center items-center gap-2"
                    >
                      {actionLoading ? <Icon icon="svg-spinners:180-ring" className="size-5" /> : "I Have Received Payment - Release Crypto"}
                    </button>
                    <button
                      onClick={() => handleAction("cancel")}
                      disabled={actionLoading}
                      className="w-full py-4 bg-[#1c1f26] hover:bg-red-500/10 hover:text-red-400 text-[#888c99] font-bold rounded-[8px] transition-all flex justify-center items-center gap-2 border border-[#22252e] hover:border-red-500/30"
                    >
                      Cancel Trade (Not Received)
                    </button>
                  </div>
                </>
              ) : trade.status === "released" ? (
                <div className="flex flex-col items-center justify-center gap-4 py-8 text-center">
                  <div className="size-16 rounded-full bg-emerald-500/10 flex items-center justify-center border border-emerald-500/30">
                    <Icon icon="hugeicons:checkmark-circle-02" className="size-8 text-emerald-500" />
                  </div>
                  <h3 className="text-lg font-bold text-white">Trade Completed!</h3>
                  <p className="text-sm text-[#888c99]">You successfully released the crypto to the buyer.</p>
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center gap-4 py-8 text-center">
                  <Icon icon="hugeicons:cancel-circle" className="size-12 text-red-500" />
                  <h3 className="text-lg font-bold text-white">Trade Cancelled</h3>
                  <p className="text-sm text-[#888c99]">This trade was cancelled and the escrow has been returned to your available balance.</p>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
