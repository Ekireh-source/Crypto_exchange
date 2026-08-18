'use client';

import FloatingNavbar from '@/components/navigation/floatingnav';

export default function MainLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-[#0a0b0d] text-white flex flex-col font-sans">
      <FloatingNavbar />
      <main className="flex-1 w-full max-w-[1440px] mx-auto p-4 sm:p-6 lg:p-8">
        {children}
      </main>
    </div>
  );
}
