'use client';

import Image from 'next/image';
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
  leftTitle = 'Welcome',
  leftSubtitle = 'Back',
  leftDescription = 'Glad to see you again.\nLet\'s continue where you left off.',
}: AuthCardProps) {
  return (
    <div className="min-h-screen flex w-full bg-[#121217]">
      {/* ── Left side: Image and Text overlay ── */}
      <div className="hidden lg:flex flex-1 relative flex-col justify-end p-16">
        <Image
          src="/eth-hero.png"
          alt="Ethereum Web3"
          fill
          priority
          className="object-cover"
        />
        {/* Gradient overlay for better text readability */}
        <div className="absolute inset-0 bg-gradient-to-t from-[#000000] via-[#000000]/60 to-transparent" />

        <div className="relative z-10 max-w-md mt-auto">
          <div className="flex border-l-[3px] border-[#f0c78a] pl-5 flex-col">
            <h1 className="text-[2.5rem] font-bold text-white leading-[1.1]">
              {leftTitle} <br /> {leftSubtitle}
            </h1>
          </div>
          <p className="text-[#a0a0b0] whitespace-pre-line text-sm mt-6 leading-relaxed">
            {leftDescription}
          </p>
        </div>
      </div>

      {/* ── Right side: Form Container ── */}
      <div className="flex-1 flex items-center justify-center p-6 bg-[#0e0e11]">
        <div className="w-full max-w-[420px] bg-[#1a1b24] p-8 sm:p-10 rounded-[32px] shadow-2xl relative">
          <div className="flex justify-center mb-6">
            <div className="w-14 h-14 rounded-full border border-white/5 bg-white/5 flex items-center justify-center">
              <Icon icon="lucide:hexagon" className="text-[#f0c78a] text-2xl" />
            </div>
          </div>

          <h2 className="text-2xl font-bold text-white text-center mb-1.5">{title}</h2>
          <p className="text-[#7070a0] text-sm text-center mb-8">{subtitle}</p>

          {children}
        </div>
      </div>
    </div>
  );
}
