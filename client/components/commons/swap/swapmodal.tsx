'use client';

import { useState, useEffect } from 'react';
import { Icon } from '@iconify/react';
import Image from 'next/image';
import { walletService } from '@/feature/wallet/wallet.service';
import { assetsService } from '@/feature/assets/assets.service';
import { type PortfolioResponse } from '@/feature/wallet/wallet.schema';
import { type Asset } from '@/feature/assets/assets.schema';

interface SwapModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess?: () => void;
}

export default function SwapModal({ isOpen, onClose, onSuccess }: SwapModalProps) {
  const [portfolio, setPortfolio] = useState<PortfolioResponse | null>(null);
  const [assets, setAssets] = useState<Asset[]>([]);
  const [loading, setLoading] = useState(true);

  const [fromAssetId, setFromAssetId] = useState<number | null>(null);
  const [toAssetId, setToAssetId] = useState<number | null>(null);
  const [amount, setAmount] = useState('');

  const [isSwapping, setIsSwapping] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // Dropdown states
  const [isFromDropdownOpen, setIsFromDropdownOpen] = useState(false);
  const [isToDropdownOpen, setIsToDropdownOpen] = useState(false);

  useEffect(() => {
    if (!isOpen) return;
    
    const fetchData = async () => {
      try {
        setLoading(true);
        const [portData, assetsData] = await Promise.all([
          walletService.getPortfolio(),
          assetsService.getAssets()
        ]);
        setPortfolio(portData);
        setAssets(assetsData);

        // Auto-select first two assets if available
        if (assetsData.length >= 2) {
          setFromAssetId(assetsData[0].id);
          setToAssetId(assetsData[1].id);
        }
      } catch (err) {
        console.error('Failed to fetch data:', err);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, [isOpen]);

  if (!isOpen) return null;

  const handleReset = () => {
    setAmount('');
    setError('');
    setSuccess('');
    setIsFromDropdownOpen(false);
    setIsToDropdownOpen(false);
  };

  const handleClose = () => {
    handleReset();
    onClose();
  };

  const fromAsset = assets.find(a => a.id === fromAssetId);
  const toAsset = assets.find(a => a.id === toAssetId);

  // Get available balance for "From" asset
  const fromBalance = portfolio?.assets.find(b => b.asset.id === fromAssetId)?.available || '0';

  const fromPrice = portfolio?.assets.find(b => b.asset.id === fromAssetId)?.usd_value ?
    (portfolio?.assets.find(b => b.asset.id === fromAssetId)!.usd_value / parseFloat(fromBalance || '1')) : 0;

  const toPrice = portfolio?.assets.find(b => b.asset.id === toAssetId)?.usd_value ?
    (portfolio?.assets.find(b => b.asset.id === toAssetId)!.usd_value / parseFloat(portfolio?.assets.find(b => b.asset.id === toAssetId)?.available || '1')) : 0;

  const estimatedReceive = (parseFloat(amount || '0') * (fromPrice && toPrice ? (fromPrice / toPrice) : 1) * 0.999).toFixed(6);

  const handleSwapClick = () => {
    if (fromAssetId === toAssetId) return;
    const temp = fromAssetId;
    setFromAssetId(toAssetId);
    setToAssetId(temp);
  };

  const handleSwap = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!fromAssetId || !toAssetId || !amount || parseFloat(amount) <= 0) return;

    setError('');
    setSuccess('');
    setIsSwapping(true);

    try {
      await walletService.swap({
        from_asset_id: fromAssetId,
        to_asset_id: toAssetId,
        amount: amount,
      });
      setSuccess('Swap successful!');
      setAmount('');
      
      // Refresh portfolio
      const portData = await walletService.getPortfolio();
      setPortfolio(portData);
      
      if (onSuccess) onSuccess();
      
      // Close modal after a short delay so user sees success message
      setTimeout(() => {
        handleClose();
      }, 1500);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Swap failed');
    } finally {
      setIsSwapping(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[90] flex items-center justify-center p-4 bg-black/75 backdrop-blur-md animate-in fade-in duration-200">
      <div className="w-full max-w-[700px] bg-[#13151a] border border-[#22252e] rounded-[24px] shadow-2xl p-6 flex flex-col gap-6 text-white font-sans animate-in zoom-in-95 duration-200">
        
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex flex-col">
            <h3 className="text-xl font-bold tracking-tight">Swap Assets</h3>
            <p className="text-[14px] text-[#888c99]">Exchange tokens instantly</p>
          </div>
          <button
            onClick={handleClose}
            className="size-8 rounded-full bg-[#1c1f26] text-[#888c99] hover:text-white flex items-center justify-center transition-colors"
          >
            <Icon icon="hugeicons:cancel-01" className="size-5" />
          </button>
        </div>

        {loading ? (
          <div className="flex justify-center py-10">
            <div className="animate-spin text-[#4f7cf7]">
              <Icon icon="hugeicons:loading-03" className="size-8" />
            </div>
          </div>
        ) : (
          <>
            {error && (
              <div className="p-3.5 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-sm font-medium flex items-center gap-2">
                <Icon icon="hugeicons:alert-circle" className="size-5 shrink-0" />
                <span>{error}</span>
              </div>
            )}

            {success && (
              <div className="p-3.5 rounded-xl bg-green-500/10 border border-green-500/20 text-green-400 text-sm font-medium flex items-center gap-2">
                <Icon icon="hugeicons:check-circle" className="size-5 shrink-0" />
                <span>{success}</span>
              </div>
            )}

            <form onSubmit={handleSwap} className="flex flex-col gap-2 relative">
              {/* FROM BOX */}
              <div className={`bg-[#1c1f26] border border-[#22252e] hover:border-[#3a3f4e] transition-colors rounded-2xl p-4 flex flex-col gap-3 relative ${isFromDropdownOpen ? 'z-40' : 'z-10'}`}>
                <div className="flex justify-between items-center text-[#888c99] text-[13px] font-semibold">
                  <span>Pay</span>
                  <span>Available: {parseFloat(fromBalance).toLocaleString('en-US', { maximumFractionDigits: 6 })}</span>
                </div>
                <div className="flex items-center gap-3">
                  <input
                    type="number"
                    step="any"
                    value={amount}
                    onChange={(e) => setAmount(e.target.value)}
                    placeholder="0.00"
                    className="w-full bg-transparent text-[28px] font-medium text-white placeholder-[#3a3f4e] outline-none"
                  />

                  <div className="relative">
                    <button
                      type="button"
                      onClick={() => { setIsFromDropdownOpen(!isFromDropdownOpen); setIsToDropdownOpen(false); }}
                      className="flex items-center justify-between gap-2 bg-[#22252e] hover:bg-[#2c303b] transition-colors rounded-full py-1.5 pl-1.5 pr-3 shrink-0 min-w-[160px]"
                    >
                      <div className="flex items-center gap-2">
                        <div className="relative size-7 rounded-full bg-white/10 overflow-hidden flex items-center justify-center shrink-0">
                          {fromAsset?.logo_url ? (
                            <Image src={fromAsset.logo_url} alt={fromAsset.symbol} fill className="object-cover" />
                          ) : (
                            <span className="text-white text-[10px] font-bold">{fromAsset?.symbol[0]}</span>
                          )}
                        </div>
                        <span className="text-white font-semibold text-sm truncate max-w-[90px] text-left">{fromAsset?.name}</span>
                      </div>
                      <Icon icon="hugeicons:arrow-down-01" className="size-4 text-[#888c99] shrink-0" />
                    </button>

                    {isFromDropdownOpen && (
                      <div className="absolute right-0 top-12 w-64 bg-[#1c1f26] border border-[#22252e] rounded-xl overflow-hidden shadow-xl z-50">
                        {assets.map(a => (
                          <div
                            key={a.id}
                            onClick={() => { setFromAssetId(a.id); setIsFromDropdownOpen(false); }}
                            className="flex items-center gap-3 p-3 hover:bg-[#22252e] cursor-pointer transition-colors"
                          >
                            <div className="relative size-6 rounded-full bg-white/10 overflow-hidden shrink-0">
                              {a.logo_url && <Image src={a.logo_url} alt={a.symbol} fill className="object-cover" />}
                            </div>
                            <div className="flex flex-col">
                              <span className="text-white text-sm font-semibold leading-tight">{a.name} ({a.symbol})</span>
                              <span className="text-[#888c99] text-[10px] uppercase font-medium">{a.network}</span>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              </div>

              {/* SWAP ICON BUTTON */}
              <div className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 z-20">
                <button
                  type="button"
                  onClick={handleSwapClick}
                  className="size-10 rounded-xl bg-[#22252e] border-4 border-[#13151a] hover:bg-[#2c303b] text-white flex items-center justify-center transition-transform hover:rotate-180 duration-300"
                >
                  <Icon icon="hugeicons:arrow-up-down" className="size-5" />
                </button>
              </div>

              {/* TO BOX */}
              <div className={`bg-[#1c1f26] border border-[#22252e] hover:border-[#3a3f4e] transition-colors rounded-2xl p-4 flex flex-col gap-3 relative ${isToDropdownOpen ? 'z-40' : 'z-10'}`}>
                <div className="flex justify-between items-center text-[#888c99] text-[13px] font-semibold">
                  <span>Receive (Estimated)</span>
                </div>
                <div className="flex items-center gap-3">
                  <input
                    type="text"
                    readOnly
                    value={amount ? estimatedReceive : ''}
                    placeholder="0.00"
                    className="w-full bg-transparent text-[28px] font-medium text-white placeholder-[#3a3f4e] outline-none cursor-not-allowed opacity-80"
                  />

                  <div className="relative">
                    <button
                      type="button"
                      onClick={() => { setIsToDropdownOpen(!isToDropdownOpen); setIsFromDropdownOpen(false); }}
                      className="flex items-center justify-between gap-2 bg-[#22252e] hover:bg-[#2c303b] transition-colors rounded-full py-1.5 pl-1.5 pr-3 shrink-0 min-w-[160px]"
                    >
                      <div className="flex items-center gap-2">
                        <div className="relative size-7 rounded-full bg-white/10 overflow-hidden flex items-center justify-center shrink-0">
                          {toAsset?.logo_url ? (
                            <Image src={toAsset.logo_url} alt={toAsset.symbol} fill className="object-cover" />
                          ) : (
                            <span className="text-white text-[10px] font-bold">{toAsset?.symbol[0]}</span>
                          )}
                        </div>
                        <span className="text-white font-semibold text-sm truncate max-w-[90px] text-left">{toAsset?.name}</span>
                      </div>
                      <Icon icon="hugeicons:arrow-down-01" className="size-4 text-[#888c99] shrink-0" />
                    </button>

                    {isToDropdownOpen && (
                      <div className="absolute right-0 top-12 w-64 bg-[#1c1f26] border border-[#22252e] rounded-xl overflow-hidden shadow-xl z-50">
                        {assets.map(a => (
                          <div
                            key={a.id}
                            onClick={() => { setToAssetId(a.id); setIsToDropdownOpen(false); }}
                            className="flex items-center gap-3 p-3 hover:bg-[#22252e] cursor-pointer transition-colors"
                          >
                            <div className="relative size-6 rounded-full bg-white/10 overflow-hidden shrink-0">
                              {a.logo_url && <Image src={a.logo_url} alt={a.symbol} fill className="object-cover" />}
                            </div>
                            <div className="flex flex-col">
                              <span className="text-white text-sm font-semibold leading-tight">{a.name} ({a.symbol})</span>
                              <span className="text-[#888c99] text-[10px] uppercase font-medium">{a.network}</span>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              </div>

              <button
                type="submit"
                disabled={isSwapping || !amount || parseFloat(amount) <= 0 || fromAssetId === toAssetId}
                className="w-full h-14 mt-4 bg-blue-600 hover:bg-blue-500 disabled:bg-[#22252e] disabled:text-[#888c99] text-white font-semibold text-[16px] rounded-2xl transition-colors flex items-center justify-center gap-2 shadow-lg shadow-blue-600/20 disabled:shadow-none"
              >
                {isSwapping ? (
                  <><Icon icon="hugeicons:loading-03" className="size-5 animate-spin" /> Swapping...</>
                ) : (
                  'Confirm Swap'
                )}
              </button>
            </form>
          </>
        )}
      </div>
    </div>
  );
}
