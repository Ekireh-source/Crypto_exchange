"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Icon } from "@iconify/react";
import Image from "next/image";
import { p2pService } from "@/feature/p2p/p2p.service";
import { walletService } from "@/feature/wallet/wallet.service";
import { type PortfolioResponse } from "@/feature/wallet/wallet.schema";
import { toast } from "sonner";

export default function CreateP2PAd() {
  const router = useRouter();
  const [portfolio, setPortfolio] = useState<PortfolioResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  // Form State
  const [assetId, setAssetId] = useState<number | "">("");
  const [fiatCurrency, setFiatCurrency] = useState("USD");
  const [rate, setRate] = useState("");
  const [totalAmount, setTotalAmount] = useState("");
  const [minAmount, setMinAmount] = useState("");
  const [maxAmount, setMaxAmount] = useState("");
  const [paymentMethod, setPaymentMethod] = useState("Bank Transfer");

  useEffect(() => {
    fetchPortfolio();
  }, []);

  const fetchPortfolio = async () => {
    try {
      setLoading(true);
      const res = await walletService.getPortfolio();
      setPortfolio(res);
      if (res.assets.length > 0) {
        setAssetId(res.assets[0].asset.id);
      }
    } catch (error) {
      console.error("Failed to fetch portfolio:", error);
    } finally {
      setLoading(false);
    }
  };

  const selectedAssetBalance = portfolio?.assets.find((a) => a.asset.id === assetId);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!assetId || !rate || !totalAmount || !minAmount || !maxAmount) return;

    try {
      setSubmitting(true);
      await p2pService.createOrder({
        asset_id: Number(assetId),
        fiat_currency: fiatCurrency,
        rate,
        min_amount: minAmount,
        max_amount: maxAmount,
        total_amount: totalAmount,
        payment_method: paymentMethod,
      });
      toast.success("Ad created successfully!");
      router.push("/p2p");
    } catch (error: any) {
      const msg = 
        error.response?.data?.message || 
        error.response?.data?.error || 
        (typeof error.response?.data === 'string' ? error.response.data : null) || 
        error.message || 
        "Failed to create ad";
      toast.error(msg);
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <Icon icon="svg-spinners:180-ring" className="size-8 text-blue-500" />
      </div>
    );
  }

  return (
    <div className="flex flex-col w-full animate-in fade-in duration-500 pb-20 gap-8">
      {/* Header */}
      <div className="flex items-center gap-4">
        <button
          onClick={() => router.back()}
          className="size-10 rounded-full bg-[#1c1f26] hover:bg-[#22252e] flex items-center justify-center text-white transition-colors border border-[#22252e]"
        >
          <Icon icon="hugeicons:arrow-left-01" className="size-5" />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-white">Post an Ad</h1>
          <p className="text-sm text-[#888c99]">Sell your crypto on the P2P marketplace</p>
        </div>
      </div>

      {/* Form */}
      <form onSubmit={handleSubmit} className="bg-[#13151a] rounded-[6px] p-6 flex flex-col gap-6">
        
        {/* Asset Selection */}
        <div className="flex flex-col gap-2">
          <label className="text-sm font-semibold text-[#888c99]">Select Asset to Sell</label>
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
            {portfolio?.assets.map((b) => (
              <div
                key={b.asset.id}
                onClick={() => setAssetId(b.asset.id)}
                className={`flex items-center gap-3 p-3 rounded-[8px] border cursor-pointer transition-all ${
                  assetId === b.asset.id
                    ? "bg-blue-600/10 border-blue-500"
                    : "bg-[#1c1f26] border-[#22252e] hover:bg-[#22252e]"
                }`}
              >
                <div className="relative size-8 rounded-full overflow-hidden bg-[#22252e] flex items-center justify-center shrink-0">
                  {b.asset.logo_url ? (
                    <Image src={b.asset.logo_url} alt={b.asset.symbol} fill className="object-cover" />
                  ) : (
                    <span className="text-xs font-bold text-white">{b.asset.symbol[0]}</span>
                  )}
                </div>
                <div className="flex flex-col min-w-0">
                  <span className="text-sm font-bold text-white truncate">{b.asset.symbol}</span>
                  <span className="text-xs text-[#888c99] truncate">{parseFloat(b.available).toFixed(4)} avail</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Pricing */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
          <div className="flex flex-col gap-2">
            <label className="text-sm font-semibold text-[#888c99]">Fiat Currency</label>
            <select
              value={fiatCurrency}
              onChange={(e) => setFiatCurrency(e.target.value)}
              className="w-full bg-[#1c1f26] border border-[#22252e] rounded-[8px] px-4 py-3.5 text-white focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 appearance-none font-medium"
            >
              <option value="USD">USD - US Dollar</option>
              <option value="EUR">EUR - Euro</option>
              <option value="GBP">GBP - British Pound</option>
              <option value="NGN">NGN - Nigerian Naira</option>
            </select>
          </div>
          <div className="flex flex-col gap-2">
            <label className="text-sm font-semibold text-[#888c99]">Price per {selectedAssetBalance?.asset.symbol || "Coin"}</label>
            <input
              type="number"
              step="any"
              value={rate}
              onChange={(e) => setRate(e.target.value)}
              placeholder="e.g. 1.05"
              className="w-full bg-[#1c1f26] border border-[#22252e] rounded-[8px] px-4 py-3.5 text-white placeholder-[#888c99]/50 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all font-medium"
              required
            />
          </div>
        </div>

        {/* Amounts */}
        <div className="flex flex-col gap-2">
          <label className="text-sm font-semibold text-[#888c99]">Total Amount to Sell</label>
          <div className="relative">
            <input
              type="number"
              step="any"
              value={totalAmount}
              onChange={(e) => setTotalAmount(e.target.value)}
              placeholder="Total crypto you want to lock in this ad"
              className="w-full bg-[#1c1f26] border border-[#22252e] rounded-[8px] px-4 py-3.5 text-white placeholder-[#888c99]/50 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all font-medium pr-24"
              required
            />
            <div className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-2">
              <button
                type="button"
                onClick={() => setTotalAmount(selectedAssetBalance?.available || "0")}
                className="px-2 py-1 bg-[#22252e] text-[#888c99] hover:text-white text-xs font-bold rounded-lg transition-colors uppercase"
              >
                Max
              </button>
            </div>
          </div>
          {selectedAssetBalance && (
             <p className="text-xs text-[#888c99]">
               Available Balance: {selectedAssetBalance.available} {selectedAssetBalance.asset.symbol}
             </p>
          )}
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
          <div className="flex flex-col gap-2">
            <label className="text-sm font-semibold text-[#888c99]">Order Limit (Min)</label>
            <input
              type="number"
              step="any"
              value={minAmount}
              onChange={(e) => setMinAmount(e.target.value)}
              placeholder="Minimum crypto per order"
              className="w-full bg-[#1c1f26] border border-[#22252e] rounded-[8px] px-4 py-3.5 text-white placeholder-[#888c99]/50 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all font-medium"
              required
            />
          </div>
          <div className="flex flex-col gap-2">
            <label className="text-sm font-semibold text-[#888c99]">Order Limit (Max)</label>
            <input
              type="number"
              step="any"
              value={maxAmount}
              onChange={(e) => setMaxAmount(e.target.value)}
              placeholder="Maximum crypto per order"
              className="w-full bg-[#1c1f26] border border-[#22252e] rounded-[8px] px-4 py-3.5 text-white placeholder-[#888c99]/50 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-all font-medium"
              required
            />
          </div>
        </div>

        {/* Payment Method */}
        <div className="flex flex-col gap-2">
          <label className="text-sm font-semibold text-[#888c99]">Payment Method</label>
          <select
            value={paymentMethod}
            onChange={(e) => setPaymentMethod(e.target.value)}
            className="w-full bg-[#1c1f26] border border-[#22252e] rounded-[8px] px-4 py-3.5 text-white focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 appearance-none font-medium"
          >
            <option value="Bank Transfer">Bank Transfer</option>
            <option value="Cash App">Cash App</option>
            <option value="Zelle">Zelle</option>
            <option value="PayPal">PayPal</option>
          </select>
        </div>

        {/* Note */}
        <div className="p-4 bg-blue-600/10 rounded-[8px] border border-blue-500/20">
          <div className="flex gap-3">
            <Icon icon="hugeicons:information-circle" className="size-5 text-blue-400 shrink-0 mt-0.5" />
            <p className="text-sm text-blue-400 leading-relaxed">
              When you post this ad, <span className="font-bold text-white">{totalAmount || "0"} {selectedAssetBalance?.asset.symbol}</span> will be temporarily locked in escrow from your available balance. You can cancel the ad anytime to unlock remaining funds.
            </p>
          </div>
        </div>

        {/* Submit */}
        <button
          type="submit"
          disabled={submitting || !assetId}
          className="w-full py-4 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold rounded-[8px] transition-all shadow-lg shadow-blue-600/20 flex justify-center items-center gap-2 mt-4"
        >
          {submitting ? (
            <>
              <Icon icon="svg-spinners:180-ring" className="size-5" />
              <span>Posting Ad...</span>
            </>
          ) : (
            "Post Ad"
          )}
        </button>

      </form>
    </div>
  );
}
