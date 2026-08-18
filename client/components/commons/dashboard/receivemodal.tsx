'use client';

import { useState, useEffect } from 'react';
import { Icon } from '@iconify/react';
import Image from 'next/image';
import { QRCodeSVG } from 'qrcode.react';
import AssetSelectModal from '@/components/ui/asset-select';
import { assetsService } from '@/feature/assets/assets.service';
import { transferService } from '@/feature/transfer/transfer.service';
import { type Asset } from '@/feature/assets/assets.schema';
import { type DepositAddressResponse } from '@/feature/transfer/transfer.schema';

interface ReceiveModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export default function ReceiveModal({ isOpen, onClose }: ReceiveModalProps) {
  const [assets, setAssets] = useState<Asset[]>([]);
  const [selectedAsset, setSelectedAsset] = useState<Asset | null>(null);
  const [isAssetSelectOpen, setIsAssetSelectOpen] = useState(false);
  const [depositData, setDepositData] = useState<DepositAddressResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);

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

  useEffect(() => {
    if (!selectedAsset || !isOpen) return;
    const loadDepositAddress = async () => {
      setLoading(true);
      try {
        const res = await transferService.getDepositAddress(selectedAsset.id);
        setDepositData(res);
      } catch (err) {
        console.error('Failed to fetch deposit address:', err);
      } finally {
        setLoading(false);
      }
    };
    loadDepositAddress();
  }, [selectedAsset, isOpen]);

  const handleCopy = () => {
    if (!depositData?.address) return;
    navigator.clipboard.writeText(depositData.address);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (!isOpen) return null;

  return (
    <>
      <div className="fixed inset-0 z-[90] flex items-center justify-center p-4 bg-black/75 backdrop-blur-md animate-in fade-in duration-200">
        <div className="w-full max-w-[460px] bg-[#13151a] border border-[#22252e] rounded-[24px] shadow-2xl p-6 flex flex-col gap-6 text-white font-sans animate-in zoom-in-95 duration-200">
          
          {/* Header */}
          <div className="flex items-center justify-between">
            <div className="flex flex-col">
              <h3 className="text-xl font-bold tracking-tight">Receive Crypto</h3>
              <p className="text-[14px] text-[#888c99]">Deposit funds to your wallet</p>
            </div>
            <button
              onClick={onClose}
              className="size-8 rounded-full bg-[#1c1f26] text-[#888c99] hover:text-white flex items-center justify-center transition-colors"
            >
              <Icon icon="hugeicons:cancel-01" className="size-5" />
            </button>
          </div>

          {/* Asset Selector Trigger */}
          <div className="flex flex-col gap-2">
            <label className="text-[13px] font-semibold text-[#888c99]">Select Asset & Network</label>
            <button
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
                      Network: <span className="text-blue-400 font-medium">{selectedAsset.network}</span> ({selectedAsset.standard})
                    </span>
                  </div>
                </div>
              ) : (
                <span className="text-sm text-[#888c99]">Choose coin</span>
              )}
              <Icon icon="hugeicons:arrow-down-01" className="size-5 text-[#888c99] group-hover:text-white transition-colors" />
            </button>
          </div>

          {/* Warning Banner */}
          {selectedAsset && (
            <div className="flex items-start gap-3 p-3.5 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-300 text-[13px] leading-snug">
              <Icon icon="hugeicons:alert-02" className="size-5 text-amber-400 shrink-0 mt-0.5" />
              <span>
                Only send <strong className="text-white">{selectedAsset.symbol}</strong> on the{' '}
                <strong className="text-white">{selectedAsset.network}</strong> network to this address.
              </span>
            </div>
          )}

          {/* QR Code & Address Display */}
          <div className="flex flex-col items-center justify-center p-6 bg-[#1c1f26] border border-[#22252e] rounded-2xl gap-4">
            {loading ? (
              <div className="size-44 flex items-center justify-center">
                <Icon icon="hugeicons:loading-03" className="size-8 text-blue-500 animate-spin" />
              </div>
            ) : depositData?.address ? (
              <>
                <div className="p-3 bg-white rounded-xl shadow-lg">
                  <QRCodeSVG value={depositData.address} size={150} />
                </div>
                <div className="w-full flex flex-col gap-1 items-center">
                  <span className="text-[12px] text-[#888c99] font-medium">Deposit Address</span>
                  <p className="text-[13px] font-mono text-white break-all text-center px-2 py-1 bg-[#13151a] rounded-lg border border-[#22252e] w-full">
                    {depositData.address}
                  </p>
                </div>
              </>
            ) : (
              <div className="text-sm text-[#888c99] py-8">Select an asset to generate address</div>
            )}
          </div>

          {/* Action Buttons */}
          <button
            onClick={handleCopy}
            disabled={!depositData?.address || loading}
            className="w-full h-12 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white font-semibold text-[15px] rounded-full flex items-center justify-center gap-2 transition-colors shadow-lg shadow-blue-600/20"
          >
            {copied ? (
              <>
                <Icon icon="hugeicons:checkmark-circle-02" className="size-5 text-white" />
                Address Copied!
              </>
            ) : (
              <>
                <Icon icon="hugeicons:copy-01" className="size-5" />
                Copy Deposit Address
              </>
            )}
          </button>
        </div>
      </div>

      {/* Asset Select Modal */}
      <AssetSelectModal
        isOpen={isAssetSelectOpen}
        onClose={() => setIsAssetSelectOpen(false)}
        assets={assets}
        selectedAsset={selectedAsset}
        onSelect={(asset) => setSelectedAsset(asset)}
        title="Select Deposit Asset"
      />
    </>
  );
}
