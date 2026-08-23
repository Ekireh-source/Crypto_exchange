'use client';

import { useEffect, useState } from 'react';
import Image from 'next/image';
import { Icon } from '@iconify/react';
import Link from 'next/link';
import { assetsService } from '@/feature/assets/assets.service';
import { walletService } from '@/feature/wallet/wallet.service';
import { type Asset } from '@/feature/assets/assets.schema';

export default function Watchlist() {
  const [assets, setAssets] = useState<Asset[]>([]);
  const [watchlistIds, setWatchlistIds] = useState<number[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [assetsData, watchlistData] = await Promise.all([
          assetsService.getAssets(),
          walletService.getWatchlist().catch(() => []),
        ]);
        setAssets(assetsData);
        setWatchlistIds(watchlistData);
      } catch (err) {
        console.error('Failed to fetch watchlist data:', err);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, []);

  const handleToggleWatchlist = async (e: React.MouseEvent, assetId: number) => {
    e.stopPropagation();
    setWatchlistIds((prev) =>
      prev.includes(assetId) ? prev.filter((id) => id !== assetId) : [...prev, assetId]
    );
    try {
      await walletService.toggleWatchlist(assetId);
    } catch (err) {
      console.error('Failed to toggle watchlist:', err);
      // Revert if failed
      setWatchlistIds((prev) =>
        prev.includes(assetId) ? prev.filter((id) => id !== assetId) : [...prev, assetId]
      );
    }
  };

  const watchedAssets = assets.filter((a) => watchlistIds.includes(a.id));

  if (loading) {
    return (
      <div className="bg-[#0a0b0d] border border-[#22252e] rounded-[16px] p-6 flex items-center justify-center h-32">
        <span className="text-[#888c99] animate-pulse">Loading watchlist...</span>
      </div>
    );
  }

  if (watchedAssets.length === 0) {
    return (
      <div className="bg-[#0a0b0d] border border-[#22252e] rounded-[16px] p-6 flex flex-col items-center justify-center">
        <div className="relative mb-6">
          <div className="size-16 bg-[#16181d] rounded-full flex items-center justify-center">
            <Icon icon="hugeicons:add-01" className="size-6 text-[#888c99]" />
          </div>
          <div className="absolute top-0 right-0 size-4 rounded-full bg-blue-500/20" />
          <div className="absolute bottom-0 left-0 size-3 rounded-full bg-emerald-500/20" />
        </div>
        <h4 className="text-[17px] font-bold text-white mb-2">Build your watchlist</h4>
        <p className="text-[14px] text-[#888c99] mb-8 text-center">Keep track of crypto prices by adding assets to your watchlist</p>
        <Link href="/assets" className="w-full">
          <button className="w-full bg-[#16181d] hover:bg-[#1a1c23] border border-[#22252e] text-white font-semibold text-[15px] rounded-[12px] h-[48px] transition-colors">
            Add to watchlist
          </button>
        </Link>
      </div>
    );
  }

  return (
    <div className="bg-[#0a0b0d] border border-[#22252e] rounded-[16px] p-6 flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h3 className="text-[17px] font-bold text-white">Your Watchlist</h3>
        <Link href="/assets" className="text-[#4f7cf7] text-[13px] font-bold hover:underline">
          Manage
        </Link>
      </div>

      <div className="flex flex-col gap-4">
        {watchedAssets.map((asset) => (
          <div key={asset.id} className="flex items-center justify-between group">
            <div className="flex items-center gap-4">
              <div className="relative size-10 rounded-full overflow-hidden bg-white/10 flex items-center justify-center shrink-0">
                {asset.logo_url ? (
                  <Image src={asset.logo_url} alt={asset.name} fill className="object-cover" />
                ) : (
                  <div className="w-full h-full bg-[#f7931a] flex items-center justify-center text-white font-bold text-sm">
                    {asset.symbol[0]}
                  </div>
                )}
              </div>
              <div className="flex flex-col">
                <span className="text-[15px] font-bold text-white leading-tight">
                  {asset.symbol}
                </span>
                <span className="text-[13px] text-[#888c99]">
                  {asset.name}
                </span>
              </div>
            </div>
            
            <div className="flex items-center gap-4">
              <span className="text-[15px] font-medium text-white">
                {asset.current_price ? `$${asset.current_price.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 6 })}` : '—'}
              </span>
              <button 
                onClick={(e) => handleToggleWatchlist(e, asset.id)}
                className="p-1.5 rounded-full hover:bg-[#22252e] transition-colors"
              >
                <Icon 
                  icon="hugeicons:star" 
                  className="size-5 text-yellow-400 fill-yellow-400" 
                />
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
