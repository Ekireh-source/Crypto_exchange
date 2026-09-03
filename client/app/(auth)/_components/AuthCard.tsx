'use client';

import { Icon } from '@iconify/react';

interface AuthCardProps {
  children: React.ReactNode;
  title: string;
  subtitle: string;
  leftTitle?: string;
  leftSubtitle?: string;
  leftDescription?: string;
}

export default function AuthCard({
  children,
  title,
  subtitle,
}: AuthCardProps) {
  return (
    <div className="min-h-screen flex items-center justify-center bg-[#01010d] text-slate-200 overflow-hidden font-sans relative selection:bg-pink-500/30 selection:text-pink-100 w-full p-4">
      {/* Background Grid (Matches Home) */}
      <div className="absolute inset-0 bg-grid-pattern opacity-40 pointer-events-none"></div>

      {/* Top Gradient (Matches Home) */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[1000px] h-[500px] bg-indigo-900/20 blur-[120px] rounded-full pointer-events-none"></div>

      {/* ── Form Container ── */}
      <div className="relative z-10 w-full max-w-[420px] bg-[#01010d]/80 backdrop-blur-md p-8 sm:p-10 rounded-lg border border-white/10 shadow-[0_0_50px_rgba(99,102,241,0.1)]">
        
        {/* Cyberpunk corner accents */}
        <div className="absolute -top-1 -left-1 w-2 h-2 border-t border-l border-white/30"></div>
        <div className="absolute -top-1 -right-1 w-2 h-2 border-t border-r border-white/30"></div>
        <div className="absolute -bottom-1 -left-1 w-2 h-2 border-b border-l border-white/30"></div>
        <div className="absolute -bottom-1 -right-1 w-2 h-2 border-b border-r border-white/30"></div>

        <div className="flex justify-center mb-6">
          <div className="w-14 h-14 rounded-full border border-indigo-500/30 bg-indigo-500/10 flex items-center justify-center shadow-[0_0_15px_rgba(99,102,241,0.2)]">
            <Icon icon="lucide:hexagon" className="text-indigo-400 text-2xl" />
          </div>
        </div>

        <h2 className="text-2xl font-semibold text-white text-center mb-1.5 glow-text tracking-tight">{title}</h2>
        <p className="text-slate-400 text-sm text-center mb-8">{subtitle}</p>

        {children}
      </div>
    </div>
  );
}
