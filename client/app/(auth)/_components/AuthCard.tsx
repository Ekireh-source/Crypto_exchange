'use client';

import Image from 'next/image';

interface AuthCardProps {
  children: React.ReactNode;
  title?: string;
  subtitle?: string;
}

export default function AuthCard({ children, title = 'WELCOME TO WEB3', subtitle }: AuthCardProps) {
  return (
    <div className="relative min-h-dvh flex flex-col items-center bg-[#0a0a0f] overflow-hidden">

      {/* ── Hero Image ── */}
      <div className="relative w-full flex-shrink-0" style={{ height: '55dvh', minHeight: 280, maxHeight: 500 }}>
        <Image
          src="/eth-hero.png"
          alt="Ethereum Web3"
          fill
          priority
          className="object-cover object-[center_20%]"
        />
        {/* Gradient fade to background colour */}
        <div className="absolute inset-0 bg-gradient-to-b from-transparent via-[rgba(10,10,15,0.3)] to-[#0a0a0f]" />
      </div>

      {/* ── Card ── */}
      <div className="relative z-10 w-full max-w-[420px] px-6 pb-12 -mt-10 flex flex-col items-center animate-[fadeUp_0.5s_ease_both]">
        <h1
          className="text-white font-black tracking-widest text-center leading-none mb-1.5"
          style={{ fontSize: 'clamp(1.6rem, 5vw, 2.2rem)', textShadow: '0 0 40px rgba(124,107,255,0.4)' }}
        >
          {title}
        </h1>
        {subtitle && (
          <p className="text-[#7070a0] text-sm text-center tracking-wide mb-6">{subtitle}</p>
        )}
        {children}
      </div>
    </div>
  );
}
