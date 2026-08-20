'use client';

import React, { useEffect, useState } from 'react';
import Image from 'next/image';
import { Icon } from '@iconify/react';
import { assetsService } from '@/feature/assets/assets.service';
import { type Asset } from '@/feature/assets/assets.schema';

export default function AssetsPage() {
  const [assets, setAssets] = useState<Asset[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchAssets = async () => {
      try {
        const data = await assetsService.getAssets();
        setAssets(data);
      } catch (error) {
        console.error('Failed to fetch assets:', error);
      } finally {
        setLoading(false);
      }
    };
    fetchAssets();
  }, []);

  return (
    <div className="flex flex-col w-full animate-in fade-in duration-500 pb-20">
      <div className="w-full">
        {/* Table Header */}
        <div className="grid grid-cols-12 gap-4 pb-4 border-b border-[#22252e] text-[13px] font-semibold text-[#888c99]">
          <div className="col-span-3">Name</div>
          <div className="col-span-2">Market price</div>
          <div className="col-span-2">Volume</div>
          <div className="col-span-2">Market cap</div>
          <div className="col-span-2">Change</div>
          <div className="col-span-1"></div> {/* For Buy / Star */}
        </div>

        {/* Table Body */}
        <div className="flex flex-col">
          {loading ? (
            <div className="py-10 flex justify-center">
              <span className="text-[#888c99] animate-pulse">Loading assets...</span>
            </div>
          ) : (
            assets.map((asset) => {
              return (
                <div 
                  key={asset.id} 
                  className="grid grid-cols-12 gap-4 py-4 items-center border-b border-transparent hover:bg-[#16181d] transition-colors group cursor-pointer"
                >
                  {/* Name & Logo */}
                  <div className="col-span-3 flex items-center gap-4 px-2">
                    <div className="relative size-10 rounded-full bg-[#16181d] border border-[#22252e] overflow-hidden flex items-center justify-center shrink-0">
                      {asset.logo_url ? (
                        <Image src={asset.logo_url} alt={asset.name} fill className="object-cover" />
                      ) : (
                        <div className="text-[13px] font-bold text-white uppercase">{asset.symbol[0]}</div>
                      )}
                    </div>
                    <div className="flex flex-col">
                      <span className="text-[15px] font-bold text-white leading-tight">{asset.name}</span>
                      <span className="text-[13px] text-[#888c99]">{asset.symbol}</span>
                    </div>
                  </div>

                  {/* Market Price (Mocked for now since DB lacks live prices) */}
                  <div className="col-span-2 flex items-center">
                    <span className="text-[15px] font-medium text-white">—</span>
                  </div>

                  {/* Volume */}
                  <div className="col-span-2 flex items-center">
                    <span className="text-[15px] font-medium text-white">—</span>
                  </div>

                  {/* Market Cap */}
                  <div className="col-span-2 flex items-center">
                    <span className="text-[15px] font-medium text-white">—</span>
                  </div>

                  {/* Change */}
                  <div className="col-span-2 flex items-center">
                    <span className="text-[15px] font-semibold text-[#888c99]">
                      —
                    </span>
                  </div>

                  {/* Actions (Buy & Star) */}
                  <div className="col-span-1 flex items-center justify-end gap-6 pr-4">
                    <button className="text-[15px] font-semibold text-[#4f7cf7] hover:text-[#3f6be7] transition-colors">
                      Buy
                    </button>
                    <button className="text-[#888c99] hover:text-white transition-colors">
                      <Icon icon="hugeicons:star" className="size-5" />
                    </button>
                  </div>
                </div>
              );
            })
          )}
        </div>
      </div>
    </div>
  );
}
