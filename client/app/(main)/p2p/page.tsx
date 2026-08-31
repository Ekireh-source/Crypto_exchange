"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Icon } from "@iconify/react";
import Image from "next/image";
import { p2pService } from "@/feature/p2p/p2p.service";
import { type P2POrder } from "@/feature/p2p/p2p.schema";
import { walletService } from "@/feature/wallet/wallet.service";
import { type PortfolioResponse } from "@/feature/wallet/wallet.schema";
import { toast } from "sonner";

export default function P2PMarketplace() {
  const router = useRouter();
  const [orders, setOrders] = useState<P2POrder[]>([]);
  const [portfolio, setPortfolio] = useState<PortfolioResponse | null>(null);
  const [loading, setLoading] = useState(true);

  // Buy Modal State
  const [selectedOrder, setSelectedOrder] = useState<P2POrder | null>(null);
  const [buyAmount, setBuyAmount] = useState("");
  const [processing, setProcessing] = useState(false);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [ordersRes, portfolioRes] = await Promise.all([
        p2pService.listOrders(),
        walletService.getPortfolio(),
      ]);
      setOrders(ordersRes);
      setPortfolio(portfolioRes);
    } catch (error) {
      console.error("Failed to fetch P2P data:", error);
    } finally {
      setLoading(false);
    }
  };

  const getAssetDetails = (assetId: number) => {
    return portfolio?.assets.find((a) => a.asset.id === assetId)?.asset;
  };

  const handleBuy = async () => {
    if (!selectedOrder || !buyAmount) return;
    try {
      setProcessing(true);
      const rate = parseFloat(selectedOrder.rate);
      const amountFloat = parseFloat(buyAmount);
      const fiatAmount = (amountFloat * rate).toString();

      const trade = await p2pService.createTrade(selectedOrder.id, {
        amount: buyAmount,
        fiat_amount: fiatAmount,
      });
      toast.success("Trade initiated successfully!");
      router.push(`/p2p/trade/${trade.id}`);
    } catch (error: any) {
      console.error("Trade creation error:", error);
      const msg = 
        error.response?.data?.message || 
        error.response?.data?.error || 
        (typeof error.response?.data === 'string' ? error.response.data : null) || 
        error.message || 
        "Failed to create trade";
      toast.error(msg);
    } finally {
      setProcessing(false);
    }
  };

  return (
    <div className="flex flex-col w-full animate-in fade-in duration-500 pb-20 gap-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-[#13151a] p-6 rounded-[8px]">
        <div className="flex items-center gap-4">
          <div className="size-12 rounded-xl bg-[#1c1f26] flex items-center justify-center border border-blue-500/30">
            <Icon icon="hugeicons:trade-up" className="size-6 text-blue-500" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-white">P2P Marketplace</h1>
            <p className="text-sm text-[#888c99]">Trade crypto directly with other users</p>
          </div>
        </div>
        <div className="flex flex-row items-stretch sm:items-center gap-3 w-full sm:w-auto">
          <button
            onClick={() => router.push("/p2p/trades")}
            className="w-full sm:w-auto px-4 py-3 sm:py-2 bg-[#1c1f26] hover:bg-[#22252e] text-white text-sm font-semibold rounded-[8px] border border-[#22252e] transition-colors"
          >
            My Trades
          </button>
          <button
            onClick={() => router.push("/p2p/create")}
            className="w-full sm:w-auto px-4 py-3 sm:py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-semibold rounded-[8px] transition-colors shadow-lg shadow-blue-600/20"
          >
            Create Ad
          </button>
        </div>
      </div>

      {/* Orders List */}
      <div className="bg-[#13151a] rounded-[8px] flex flex-col">
        <div className="p-6 border-b border-[#22252e]">
          <h2 className="text-lg font-bold text-white">Active Sell Orders</h2>
        </div>
        
        {loading ? (
          <div className="p-12 flex items-center justify-center">
            <Icon icon="svg-spinners:180-ring" className="size-8 text-blue-500" />
          </div>
        ) : orders.length === 0 ? (
          <div className="p-12 flex flex-col items-center justify-center gap-3">
            <Icon icon="hugeicons:file-not-found" className="size-12 text-[#888c99]/50" />
            <p className="text-[#888c99] font-medium text-center max-w-sm">
              No active sell orders available right now. Be the first to create an ad!
            </p>
          </div>
        ) : (
          <div className="flex flex-col w-full">
            {/* Table Header */}
            <div className="flex md:grid md:grid-cols-12 justify-between gap-4 p-6 border-b border-[#22252e] text-[13px] font-semibold text-[#888c99] uppercase tracking-wider bg-[#1c1f26]/30">
              <div className="col-span-3">Asset</div>
              <div className="col-span-3">Price / Rate</div>
              <div className="col-span-3 hidden md:block">Limits / Available</div>
              <div className="col-span-2 hidden md:block">Payment</div>
              <div className="col-span-1 hidden md:block"></div>
            </div>

            {/* Table Body */}
            <div className="flex flex-col divide-y divide-[#22252e]">
              {orders.map((order) => {
                const asset = getAssetDetails(order.asset_id);
                return (
                  <div key={order.id} className="flex flex-wrap md:flex-nowrap md:grid md:grid-cols-12 justify-between items-center gap-y-4 md:gap-4 p-6 hover:bg-[#1c1f26]/60 transition-colors">
                    {/* Asset */}
                    <div className="w-1/2 md:w-auto col-span-3 flex items-center gap-3">
                      <div className="relative size-8 rounded-full overflow-hidden bg-[#22252e] flex flex-shrink-0 items-center justify-center">
                        {asset?.logo_url ? (
                          <Image src={asset.logo_url} alt={asset.symbol} fill className="object-cover" />
                        ) : (
                          <span className="text-xs font-bold text-white">{asset?.symbol?.[0] || "?"}</span>
                        )}
                      </div>
                      <div>
                        <p className="text-sm font-bold text-white">{asset?.symbol || "Unknown"}</p>
                        <p className="text-xs text-[#888c99]">{asset?.name || "---"}</p>
                      </div>
                    </div>
                    
                    {/* Rate */}
                    <div className="w-1/2 md:w-auto col-span-3 flex flex-col items-end md:items-start md:block">
                      <span className="md:hidden text-[11px] text-[#888c99] mb-0.5">Rate</span>
                      <p className="text-lg font-bold text-white">{parseFloat(order.rate).toLocaleString(undefined, { minimumFractionDigits: 2 })} <span className="text-sm font-medium text-[#888c99]">{order.fiat_currency}</span></p>
                    </div>

                    {/* Limits / Available (Hidden on Mobile) */}
                    <div className="col-span-3 hidden md:flex flex-col gap-1 text-sm">
                      <div className="flex justify-between gap-4">
                        <span className="text-[#888c99]">Available:</span>
                        <span className="text-white font-medium">{parseFloat(order.available_amount)} {asset?.symbol}</span>
                      </div>
                      <div className="flex justify-between gap-4">
                        <span className="text-[#888c99]">Limits:</span>
                        <span className="text-white">{parseFloat(order.min_amount)} - {parseFloat(order.max_amount)} {asset?.symbol}</span>
                      </div>
                    </div>

                    {/* Payment (Hidden on Mobile) */}
                    <div className="col-span-2 hidden md:block">
                      <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-emerald-500/10 text-emerald-400 text-xs font-bold border border-emerald-500/20">
                        <Icon icon="hugeicons:bank" className="size-3.5" />
                        {order.payment_method}
                      </span>
                    </div>

                    {/* Action (Desktop) */}
                    <div className="col-span-1 hidden md:flex justify-end">
                      <button
                        onClick={() => { setSelectedOrder(order); setBuyAmount(""); }}
                        className="px-5 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-bold rounded-[8px] transition-colors shadow-lg shadow-blue-600/20"
                      >
                        Buy {asset?.symbol}
                      </button>
                    </div>

                    {/* Action (Mobile) */}
                    <div className="w-full md:hidden flex mt-2 border-t border-[#22252e]/50 pt-4">
                      <button
                        onClick={() => { setSelectedOrder(order); setBuyAmount(""); }}
                        className="w-full py-3 bg-blue-600 hover:bg-blue-500 text-white text-sm font-bold rounded-[8px] transition-colors shadow-lg shadow-blue-600/20"
                      >
                        Buy {asset?.symbol}
                      </button>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </div>

      {/* Buy Modal */}
      {selectedOrder && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in duration-200">
          <div className="bg-[#13151a] rounded-[6px] w-full max-w-md overflow-hidden shadow-2xl flex flex-col">
            <div className="p-6 border-b border-[#22252e] flex items-center justify-between">
              <h3 className="text-xl font-bold text-white">
                Buy {getAssetDetails(selectedOrder.asset_id)?.symbol}
              </h3>
              <button
                onClick={() => setSelectedOrder(null)}
                className="size-8 rounded-full bg-[#1c1f26] hover:bg-[#22252e] flex items-center justify-center text-[#888c99] hover:text-white transition-colors"
              >
                <Icon icon="hugeicons:cancel-01" className="size-5" />
              </button>
            </div>
            
            <div className="p-6 flex flex-col gap-5">
              <div className="p-4 bg-[#1c1f26] rounded-xl border border-[#22252e] flex flex-col gap-2">
                <div className="flex justify-between text-sm">
                  <span className="text-[#888c99]">Price</span>
                  <span className="font-bold text-white">{parseFloat(selectedOrder.rate).toLocaleString()} {selectedOrder.fiat_currency}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-[#888c99]">Available</span>
                  <span className="font-bold text-white">{parseFloat(selectedOrder.available_amount)}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-[#888c99]">Limits</span>
                  <span className="font-bold text-white">{parseFloat(selectedOrder.min_amount)} - {parseFloat(selectedOrder.max_amount)}</span>
                </div>
              </div>

              <div className="flex flex-col gap-2">
                <label className="text-sm font-semibold text-[#888c99]">I want to buy</label>
                <div className="relative">
                  <input
                    type="number"
                    value={buyAmount}
                    onChange={(e) => setBuyAmount(e.target.value)}
                    placeholder={`e.g. ${parseFloat(selectedOrder.min_amount)}`}
                    className="w-full bg-[#1c1f26] border border-[#22252e] rounded-xl px-4 py-3.5 text-white placeholder-[#888c99]/50 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all font-medium pr-20"
                  />
                  <div className="absolute right-4 top-1/2 -translate-y-1/2 flex items-center gap-2">
                    <span className="text-sm font-bold text-[#888c99]">{getAssetDetails(selectedOrder.asset_id)?.symbol}</span>
                  </div>
                </div>
              </div>

              {buyAmount && !isNaN(parseFloat(buyAmount)) && (
                <div className="flex justify-between items-center px-4 py-3 bg-blue-600/10 rounded-xl border border-blue-500/20">
                  <span className="text-sm font-medium text-blue-400">You will pay</span>
                  <span className="text-lg font-bold text-blue-500">
                    {(parseFloat(buyAmount) * parseFloat(selectedOrder.rate)).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })} {selectedOrder.fiat_currency}
                  </span>
                </div>
              )}

              <button
                onClick={handleBuy}
                disabled={!buyAmount || isNaN(parseFloat(buyAmount)) || processing}
                className="w-full py-3.5 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold rounded-[8px] transition-all shadow-lg shadow-blue-600/20 flex justify-center items-center gap-2 mt-2"
              >
                {processing ? (
                  <Icon icon="svg-spinners:180-ring" className="size-5" />
                ) : (
                  "Confirm Buy"
                )}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
