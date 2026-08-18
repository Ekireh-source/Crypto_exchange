'use client';

import { useState, useEffect } from 'react';
import { Icon } from '@iconify/react';
import Image from 'next/image';
import AssetSelectModal from '@/components/ui/asset-select';
import { assetsService } from '@/feature/assets/assets.service';
import { transferService } from '@/feature/transfer/transfer.service';
import { type Asset } from '@/feature/assets/assets.schema';
import { type TransactionResponse } from '@/feature/transfer/transfer.schema';

interface SendModalProps {
  isOpen: boolean;
  onClose: () => void;
}

type Step = 'form' | 'review' | 'success';

export default function SendModal({ isOpen, onClose }: SendModalProps) {
  const [assets, setAssets] = useState<Asset[]>([]);
  const [selectedAsset, setSelectedAsset] = useState<Asset | null>(null);
  const [isAssetSelectOpen, setIsAssetSelectOpen] = useState(false);

  const [toAddress, setToAddress] = useState('');
  const [amount, setAmount] = useState('');
  const [note, setNote] = useState('');

  const [step, setStep] = useState<Step>('form');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [txResult, setTxResult] = useState<TransactionResponse | null>(null);

  useEffect(() => {
    if (!isOpen) return;
    const loadAssets = async () => {
      try {
        const data = await assetsService.getAssets();
        setAssets(data);
        if (data.length > 0 && !selectedAsset) {
          setSelectedAsset(data[0]);
        }
      } catch (err) {
        console.error('Failed to load assets:', err);
      }
    };
    loadAssets();
  }, [isOpen]);

  const handleReset = () => {
    setStep('form');
    setToAddress('');
    setAmount('');
    setNote('');
    setError(null);
    setTxResult(null);
  };

  const handleClose = () => {
    handleReset();
    onClose();
  };

  const handleProceedToReview = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!selectedAsset) {
      setError('Please select an asset');
      return;
    }
    if (!toAddress.trim()) {
      setError('Recipient address is required');
      return;
    }
    if (!amount.trim() || parseFloat(amount) <= 0) {
      setError('Please enter a valid amount');
      return;
    }
    setStep('review');
  };

  const handleConfirmSend = async () => {
    if (!selectedAsset) return;
    setLoading(true);
    setError(null);

    try {
      const result = await transferService.sendCrypto({
        asset_id: selectedAsset.id,
        to_address: toAddress.trim(),
        amount: amount.trim(),
        note: note.trim() || undefined,
      });
      setTxResult(result);
      setStep('success');
    } catch (err: any) {
      console.error('Send failed:', err);
      setError(err?.response?.data?.error || 'Failed to send crypto. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <>
      <div className="fixed inset-0 z-[90] flex items-center justify-center p-4 bg-black/75 backdrop-blur-md animate-in fade-in duration-200">
        <div className="w-full max-w-[480px] bg-[#13151a] border border-[#22252e] rounded-[24px] shadow-2xl p-6 flex flex-col gap-6 text-white font-sans animate-in zoom-in-95 duration-200">
          
          {/* Header */}
          <div className="flex items-center justify-between">
            <div className="flex flex-col">
              <h3 className="text-xl font-bold tracking-tight">
                {step === 'form' && 'Send Crypto'}
                {step === 'review' && 'Review Order'}
                {step === 'success' && 'Transaction Sent'}
              </h3>
              <p className="text-[14px] text-[#888c99]">
                {step === 'form' && 'Transfer funds to an external wallet'}
                {step === 'review' && 'Confirm details before broadcasting'}
                {step === 'success' && 'Your transfer request was submitted'}
              </p>
            </div>
            <button
              onClick={handleClose}
              className="size-8 rounded-full bg-[#1c1f26] text-[#888c99] hover:text-white flex items-center justify-center transition-colors"
            >
              <Icon icon="hugeicons:cancel-01" className="size-5" />
            </button>
          </div>

          {error && (
            <div className="p-3.5 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-sm font-medium flex items-center gap-2">
              <Icon icon="hugeicons:alert-circle" className="size-5 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {/* STEP 1: FORM */}
          {step === 'form' && (
            <form onSubmit={handleProceedToReview} className="flex flex-col gap-5">
              {/* Asset Select */}
              <div className="flex flex-col gap-2">
                <label className="text-[13px] font-semibold text-[#888c99]">Asset & Network</label>
                <button
                  type="button"
                  onClick={() => setIsAssetSelectOpen(true)}
                  className="flex items-center justify-between p-3.5 bg-[#1c1f26] border border-[#22252e] hover:border-[#3a3f4e] rounded-xl transition-all group"
                >
                  {selectedAsset ? (
                    <div className="flex items-center gap-3">
                      <div className="relative size-8 rounded-full bg-white/10 overflow-hidden flex items-center justify-center shrink-0">
                        {selectedAsset.logo_url ? (
                          <Image
                            src={selectedAsset.logo_url}
                            alt={selectedAsset.name}
                            fill
                            className="object-cover"
                          />
                        ) : (
                          <div className="w-full h-full bg-[#f7931a] flex items-center justify-center text-white font-bold text-xs">
                            {selectedAsset.symbol[0]}
                          </div>
                        )}
                      </div>
                      <div className="flex flex-col items-start">
                        <span className="text-[15px] font-semibold text-white">
                          {selectedAsset.name} ({selectedAsset.symbol})
                        </span>
                        <span className="text-[12px] text-[#888c99]">
                          Network: <span className="text-blue-400 font-medium">{selectedAsset.network}</span>
                        </span>
                      </div>
                    </div>
                  ) : (
                    <span className="text-sm text-[#888c99]">Choose coin</span>
                  )}
                  <Icon icon="hugeicons:arrow-down-01" className="size-5 text-[#888c99] group-hover:text-white transition-colors" />
                </button>
              </div>

              {/* Recipient Address */}
              <div className="flex flex-col gap-2">
                <label className="text-[13px] font-semibold text-[#888c99]">To Address</label>
                <input
                  type="text"
                  value={toAddress}
                  onChange={(e) => setToAddress(e.target.value)}
                  placeholder="Paste recipient wallet address"
                  className="w-full h-12 bg-[#1c1f26] border border-[#22252e] focus:border-blue-500 rounded-xl px-4 text-sm text-white font-mono placeholder-[#686d7d] outline-none transition-colors"
                />
              </div>

              {/* Amount */}
              <div className="flex flex-col gap-2">
                <div className="flex items-center justify-between">
                  <label className="text-[13px] font-semibold text-[#888c99]">Amount</label>
                  <span className="text-[12px] text-[#888c99]">
                    Balance: <strong className="text-white">0.00 {selectedAsset?.symbol}</strong>
                  </span>
                </div>
                <div className="relative flex items-center">
                  <input
                    type="number"
                    step="any"
                    value={amount}
                    onChange={(e) => setAmount(e.target.value)}
                    placeholder="0.00"
                    className="w-full h-12 bg-[#1c1f26] border border-[#22252e] focus:border-blue-500 rounded-xl pl-4 pr-16 text-sm text-white font-mono placeholder-[#686d7d] outline-none transition-colors"
                  />
                  <span className="absolute right-4 text-xs font-bold text-blue-400 uppercase">
                    {selectedAsset?.symbol}
                  </span>
                </div>
              </div>

              {/* Note (Optional) */}
              <div className="flex flex-col gap-2">
                <label className="text-[13px] font-semibold text-[#888c99]">Note (Optional)</label>
                <input
                  type="text"
                  value={note}
                  onChange={(e) => setNote(e.target.value)}
                  placeholder="e.g. Payment for invoice"
                  className="w-full h-11 bg-[#1c1f26] border border-[#22252e] focus:border-blue-500 rounded-xl px-4 text-sm text-white placeholder-[#686d7d] outline-none transition-colors"
                />
              </div>

              <button
                type="submit"
                className="w-full h-12 mt-2 bg-blue-600 hover:bg-blue-500 text-white font-semibold text-[15px] rounded-full transition-colors shadow-lg shadow-blue-600/20"
              >
                Continue to Review
              </button>
            </form>
          )}

          {/* STEP 2: REVIEW */}
          {step === 'review' && selectedAsset && (
            <div className="flex flex-col gap-5">
              <div className="p-4 bg-[#1c1f26] border border-[#22252e] rounded-2xl flex flex-col gap-4">
                <div className="flex items-center justify-between pb-3 border-b border-[#22252e]">
                  <span className="text-sm text-[#888c99]">Asset</span>
                  <span className="text-sm font-semibold text-white flex items-center gap-1.5">
                    {selectedAsset.name} ({selectedAsset.symbol})
                  </span>
                </div>

                <div className="flex items-center justify-between pb-3 border-b border-[#22252e]">
                  <span className="text-sm text-[#888c99]">Network</span>
                  <span className="text-sm font-semibold text-blue-400">{selectedAsset.network}</span>
                </div>

                <div className="flex flex-col gap-1 pb-3 border-b border-[#22252e]">
                  <span className="text-sm text-[#888c99]">Recipient</span>
                  <span className="text-xs font-mono text-white break-all">{toAddress}</span>
                </div>

                <div className="flex items-center justify-between pb-3 border-b border-[#22252e]">
                  <span className="text-sm text-[#888c99]">Send Amount</span>
                  <span className="text-sm font-bold text-white">
                    {amount} {selectedAsset.symbol}
                  </span>
                </div>

                <div className="flex items-center justify-between">
                  <span className="text-sm text-[#888c99]">Network Fee</span>
                  <span className="text-sm font-semibold text-emerald-400">0.0005 {selectedAsset.symbol}</span>
                </div>
              </div>

              <div className="flex items-center gap-3">
                <button
                  type="button"
                  onClick={() => setStep('form')}
                  className="flex-1 h-12 bg-[#1c1f26] hover:bg-[#252933] text-white font-semibold text-[15px] rounded-full transition-colors"
                >
                  Back
                </button>
                <button
                  type="button"
                  onClick={handleConfirmSend}
                  disabled={loading}
                  className="flex-1 h-12 bg-blue-600 hover:bg-blue-500 text-white font-semibold text-[15px] rounded-full flex items-center justify-center gap-2 transition-colors shadow-lg shadow-blue-600/20"
                >
                  {loading ? (
                    <Icon icon="hugeicons:loading-03" className="size-5 text-white animate-spin" />
                  ) : (
                    'Confirm & Send'
                  )}
                </button>
              </div>
            </div>
          )}

          {/* STEP 3: SUCCESS */}
          {step === 'success' && (
            <div className="flex flex-col items-center justify-center gap-5 py-4">
              <div className="size-16 rounded-full bg-emerald-500/10 text-emerald-500 flex items-center justify-center border border-emerald-500/20">
                <Icon icon="hugeicons:checkmark-circle-02" className="size-10" />
              </div>
              
              <div className="text-center flex flex-col gap-1">
                <h4 className="text-lg font-bold text-white">Transfer Initiated</h4>
                <p className="text-sm text-[#888c99]">
                  Sent <strong className="text-white">{amount} {selectedAsset?.symbol}</strong> to recipient address.
                </p>
              </div>

              {txResult?.id && (
                <div className="w-full p-3 bg-[#1c1f26] rounded-xl border border-[#22252e] text-xs font-mono text-[#888c99] text-center">
                  Tx ID: {txResult.id}
                </div>
              )}

              <button
                onClick={handleClose}
                className="w-full h-12 bg-blue-600 hover:bg-blue-500 text-white font-semibold text-[15px] rounded-full transition-colors"
              >
                Done
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Asset Select Modal */}
      <AssetSelectModal
        isOpen={isAssetSelectOpen}
        onClose={() => setIsAssetSelectOpen(false)}
        assets={assets}
        selectedAsset={selectedAsset}
        onSelect={(asset) => setSelectedAsset(asset)}
        title="Select Send Asset"
      />
    </>
  );
}
