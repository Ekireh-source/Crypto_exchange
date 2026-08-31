'use client';

import React, { useEffect, useState } from 'react';
import Image from 'next/image';
import { Icon } from '@iconify/react';
import { assetsService } from '@/feature/assets/assets.service';
import { walletService } from '@/feature/wallet/wallet.service';
import { type Asset } from '@/feature/assets/assets.schema';
import { usePriceWebsocket } from '@/hooks/usePriceWebsocket';

export default function AssetsPage() {
  const [assets, setAssets] = useState<Asset[]>([]);
  const [watchlist, setWatchlist] = useState<number[]>([]);
  const [loading, setLoading] = useState(true);

  const livePrices = usePriceWebsocket();

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [assetsData, watchlistData] = await Promise.all([
          assetsService.getAssets(),
          walletService.getWatchlist().catch(() => [])
        ]);
        setAssets(assetsData);
        setWatchlist(watchlistData);
      } catch (error) {
        console.error('Failed to fetch data:', error);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, []);

  const handleToggleWatchlist = async (e: React.MouseEvent, assetId: number) => {
    e.stopPropagation();
    
    // Optimistic update
    setWatchlist(prev => 
      prev.includes(assetId) ? prev.filter(id => id !== assetId) : [...prev, assetId]
    );

    try {
      await walletService.toggleWatchlist(assetId);
    } catch (err) {
      console.error('Failed to toggle watchlist:', err);
      // Revert if failed (simplistic revert)
      setWatchlist(prev => 
        prev.includes(assetId) ? prev.filter(id => id !== assetId) : [...prev, assetId]
      );
    }
  };

  return (
    <div className="flex flex-col w-full animate-in fade-in duration-500 pb-20">
      <div className="w-full">
        {/* Table Header */}
        <div className="flex md:grid md:grid-cols-12 justify-between gap-4 pb-4 border-b border-[#22252e] text-[13px] font-semibold text-[#888c99]">
          <div className="col-span-3">Name</div>
          <div className="col-span-2">Market price</div>
          <div className="hidden md:block col-span-2">Volume</div>
          <div className="hidden md:block col-span-2">Market cap</div>
          <div className="hidden md:block col-span-2">Change</div>
          <div className="col-span-1 hidden md:block"></div> {/* For Buy / Star */}
        </div>

        {/* Table Body */}
        <div className="flex flex-col">
          {loading ? (
            <div className="py-10 flex justify-center">
              <span className="text-[#888c99] animate-pulse">Loading assets...</span>
            </div>
          ) : (
            assets.map((asset) => {
              const currentPrice = livePrices[asset.symbol] || asset.current_price;
              
              return (
                <div 
                  key={asset.id} 
                  className="flex md:grid md:grid-cols-12 justify-between gap-4 py-4 items-center border-b border-transparent hover:bg-[#16181d] transition-colors group cursor-pointer"
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

                  {/* Market Price */}
                  <div className="col-span-2 flex items-center">
                    <span key={currentPrice} className="text-[15px] font-medium text-white animate-in fade-in transition-colors duration-500 flex flex-col md:block items-end">
                      <span className="md:hidden text-[11px] text-[#888c99] mb-0.5">Price</span>
                      {currentPrice ? `$${currentPrice.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 6 })}` : '—'}
                    </span>
                  </div>

                  {/* Volume */}
                  <div className="hidden md:flex col-span-2 items-center">
                    <span className="text-[15px] font-medium text-white">—</span>
                  </div>

                  {/* Market Cap */}
                  <div className="hidden md:flex col-span-2 items-center">
                    <span className="text-[15px] font-medium text-white">—</span>
                  </div>

                  {/* Change */}
                  <div className="hidden md:flex col-span-2 items-center">
                    <span className="text-[15px] font-semibold text-[#888c99]">
                      —
                    </span>
                  </div>

                  {/* Action / Watchlist */}
                  <div className="hidden md:flex col-span-1 items-center justify-end pr-2">
                    <button 
                      onClick={(e) => handleToggleWatchlist(e, asset.id)}
                      className="p-2 rounded-full hover:bg-[#22252e] transition-colors"
                    >
                      <Icon 
                        icon={watchlist.includes(asset.id) ? "hugeicons:star" : "hugeicons:star"} 
                        className={`size-5 transition-colors ${watchlist.includes(asset.id) ? 'text-yellow-400 fill-yellow-400' : 'text-[#888c99] hover:text-white'}`} 
                      />
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
