'use client';

import { useState, useMemo } from 'react';
import { Icon } from '@iconify/react';
import Image from 'next/image';
import { type Asset } from '@/feature/assets/assets.schema';

interface AssetSelectModalProps {
  isOpen: boolean;
  onClose: () => void;
  assets: Asset[];
  selectedAsset: Asset | null;
  onSelect: (asset: Asset) => void;
  title?: string;
}

export default function AssetSelectModal({
  isOpen,
  onClose,
  assets,
  selectedAsset,
  onSelect,
  title = 'Select Asset',
}: AssetSelectModalProps) {
  const [search, setSearch] = useState('');

  const filteredAssets = useMemo(() => {
    if (!search.trim()) return assets;
    const q = search.toLowerCase();
    return assets.filter(
      (a) =>
        a.name.toLowerCase().includes(q) ||
        a.symbol.toLowerCase().includes(q) ||
        a.network.toLowerCase().includes(q)
    );
  }, [assets, search]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/70 backdrop-blur-md animate-in fade-in duration-200">
      <div className="w-full max-w-[440px] bg-[#13151a] border border-[#22252e] rounded-[24px] shadow-2xl p-6 flex flex-col gap-5 text-white font-sans animate-in zoom-in-95 duration-200">
        
        {/* Header */}
        <div className="flex items-center justify-between">
          <h3 className="text-xl font-bold tracking-tight">{title}</h3>
          <button
            onClick={onClose}
            className="size-8 rounded-full bg-[#1c1f26] text-[#888c99] hover:text-white flex items-center justify-center transition-colors"
          >
            <Icon icon="hugeicons:cancel-01" className="size-5" />
          </button>
        </div>

        {/* Search Input */}
        <div className="relative flex items-center">
          <Icon
            icon="hugeicons:search-01"
            className="absolute left-3.5 size-5 text-[#888c99]"
          />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search coin or network..."
            className="w-full h-11 bg-[#1c1f26] border border-[#22252e] focus:border-blue-500 rounded-xl pl-11 pr-9 text-sm text-white placeholder-[#686d7d] outline-none transition-colors"
          />
          {search && (
            <button
              onClick={() => setSearch('')}
              className="absolute right-3 text-[#888c99] hover:text-white"
            >
              <Icon icon="hugeicons:cancel-circle" className="size-4" />
            </button>
          )}
        </div>

        {/* Assets List */}
        <div className="flex flex-col gap-1.5 max-h-[340px] overflow-y-auto pr-1">
          {filteredAssets.length === 0 ? (
            <div className="py-10 text-center text-[#888c99] text-sm font-medium">
              No assets found
            </div>
          ) : (
            filteredAssets.map((asset) => {
              const isSelected = selectedAsset?.id === asset.id;
              return (
                <div
                  key={asset.id}
                  onClick={() => {
                    onSelect(asset);
                    onClose();
                  }}
                  className={`flex items-center justify-between p-3 rounded-xl cursor-pointer transition-all border ${
                    isSelected
                      ? 'bg-blue-600/10 border-blue-500/40 text-white'
                      : 'border-transparent hover:bg-[#1c1f26] hover:border-[#22252e]'
                  }`}
                >
                  <div className="flex items-center gap-3.5">
                    <div className="relative size-10 rounded-full bg-white/10 overflow-hidden flex items-center justify-center shrink-0">
                      {asset.logo_url ? (
                        <Image
                          src={asset.logo_url}
                          alt={asset.name}
                          fill
                          className="object-cover"
                        />
                      ) : (
                        <div className="w-full h-full bg-[#f7931a] flex items-center justify-center text-white font-bold text-sm">
                          {asset.symbol[0]}
                        </div>
                      )}
                    </div>
                    <div className="flex flex-col">
                      <div className="flex items-center gap-2">
                        <span className="text-[15px] font-semibold text-white leading-tight">
                          {asset.name}
                        </span>
                        <span className="text-[11px] font-bold px-1.5 py-0.5 rounded bg-[#22252e] text-[#888c99]">
                          {asset.network}
                        </span>
                      </div>
                      <span className="text-[13px] text-[#888c99]">
                        {asset.symbol} · {asset.standard}
                      </span>
                    </div>
                  </div>

                  {isSelected && (
                    <Icon
                      icon="hugeicons:checkmark-circle-02"
                      className="size-6 text-blue-500 shrink-0"
                    />
                  )}
                </div>
              );
            })
          )}
        </div>
      </div>
    </div>
  );
}
