'use client';

import { useEffect, useState } from 'react';
import Image from 'next/image';
import { Icon } from '@iconify/react';
import Link from 'next/link';
import { cn } from '@/lib/utils';
import { assetsService } from '@/feature/assets/assets.service';
import { walletService } from '@/feature/wallet/wallet.service';
import { type Asset } from '@/feature/assets/assets.schema';
import { type AssetBalance } from '@/feature/wallet/wallet.schema';

export default function SupportedAssets() {
  const [assets, setAssets] = useState<Asset[]>([]);
  const [balances, setBalances] = useState<Map<number, AssetBalance>>(new Map());
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [assetsData, portfolioData] = await Promise.all([
          assetsService.getAssets(),
          walletService.getPortfolio().catch(() => null),
        ]);
        setAssets(assetsData);

        if (portfolioData?.assets) {
          const map = new Map<number, AssetBalance>();
          portfolioData.assets.forEach((ab) => map.set(ab.asset.id, ab));
          setBalances(map);
        }
      } catch (err) {
        console.error('Failed to fetch assets or portfolio:', err);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, []);

  // Filter out a few to show as highlighted crypto, like BTC, ETH (fallback to what we have if BTC/ETH aren't there)
  // Or just slice the first 3 active assets
  const highlightedCrypto = assets;

  return (
    <div className="w-full rounded-[24px] bg-[#13151a] border border-[#22252e] overflow-hidden flex flex-col font-sans">
      {/* ── Crypto Section ── */}
      <div className="p-6 pb-5 border-b border-[#22252e]">
        <div className="flex items-center justify-between mb-5">
          <div className="flex flex-col gap-0.5">
            <h2 className="text-xl font-bold text-white tracking-tight">Crypto</h2>
            <p className="text-[15px] text-[#888c99]">Trade millions of assets</p>
          </div>
          <button className="flex items-center justify-center size-9 rounded-full bg-[#22252e] text-[#888c99] hover:text-white transition-colors">
            <Icon icon="hugeicons:arrow-right-01" className="size-5" />
          </button>
        </div>

        <div className="flex flex-col gap-4 mb-6">
          {loading ? (
            <div className="animate-pulse space-y-4">
              <div className="h-12 bg-[#22252e] rounded-xl w-full" />
              <div className="h-12 bg-[#22252e] rounded-xl w-full" />
            </div>
          ) : (
            highlightedCrypto.map((asset) => {
              const bal = balances.get(asset.id);
              const available = bal ? bal.available : '0';
              return (
                <div key={asset.id} className="flex items-center justify-between group cursor-pointer">
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
                      <div className="flex items-center gap-2">
                        <span className="text-[16px] font-semibold text-white leading-tight">
                          {asset.name}
                        </span>
                        <span className="text-[11px] font-bold px-1.5 py-0.5 rounded bg-[#22252e] text-[#888c99]">
                          {asset.network}
                        </span>
                      </div>
                      <span className="text-[14px] text-[#888c99]">
                        {available} {asset.symbol}
                      </span>
                    </div>
                  </div>
                  <button className="px-5 py-2 rounded-full bg-[#22252e] text-white text-[15px] font-semibold hover:bg-[#2a2d36] transition-colors">
                    Buy
                  </button>
                </div>
              );
            })
          )}
        </div>

        <Link href="/assets" className="block w-full">
          <button className="w-full py-3.5 rounded-full bg-[#22252e] text-white text-[15px] font-semibold hover:bg-[#2a2d36] transition-colors">
            Explore all crypto
          </button>
        </Link>
      </div>

    </div>
  );
}
