'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAppSelector } from '@/store/hooks';
import Link from 'next/link';
import { Icon } from '@iconify/react';

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const { user, isAuthenticated } = useAppSelector((state) => state.auth);
  const router = useRouter();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (mounted) {
      if (!isAuthenticated) {
        router.push('/login');
      } else if (user?.role !== 'admin' && user?.role !== 'superadmin') {
        router.push('/dashboard');
      }
    }
  }, [isAuthenticated, user, router, mounted]);

  if (!mounted || !isAuthenticated || (user?.role !== 'admin' && user?.role !== 'superadmin')) {
    return <div className="min-h-screen bg-[#0a0b0d] flex items-center justify-center text-white">Loading...</div>;
  }

  return (
    <div className="min-h-screen bg-[#0a0b0d] text-white flex font-sans">
      {/* Sidebar */}
      <aside className="w-64 border-r border-[#22252e] bg-[#13151a] flex flex-col hidden md:flex">
        <div className="p-6 border-b border-[#22252e]">
          <h1 className="text-xl font-bold bg-gradient-to-r from-blue-400 to-indigo-500 bg-clip-text text-transparent">
            Admin Panel
          </h1>
        </div>
        <nav className="flex-1 p-4 space-y-2">
          <Link href="/admin" className="flex items-center gap-3 p-3 rounded-xl hover:bg-[#22252e] transition-colors text-gray-300 hover:text-white">
            <Icon icon="hugeicons:dashboard-square-01" className="size-5" />
            <span>Dashboard</span>
          </Link>
          <Link href="/admin/users" className="flex items-center gap-3 p-3 rounded-xl hover:bg-[#22252e] transition-colors text-gray-300 hover:text-white">
            <Icon icon="hugeicons:user-group" className="size-5" />
            <span>Users</span>
          </Link>
          <Link href="/admin/transactions" className="flex items-center gap-3 p-3 rounded-xl hover:bg-[#22252e] transition-colors text-gray-300 hover:text-white">
            <Icon icon="hugeicons:transaction" className="size-5" />
            <span>Transactions</span>
          </Link>
          <div className="pt-4 mt-4 border-t border-[#22252e]">
            <Link href="/dashboard" className="flex items-center gap-3 p-3 rounded-xl hover:bg-[#22252e] transition-colors text-gray-400">
              <Icon icon="hugeicons:arrow-left-01" className="size-5" />
              <span>Back to App</span>
            </Link>
          </div>
        </nav>
      </aside>

      {/* Main Content */}
      <main className="flex-1 w-full max-w-[1440px] mx-auto p-4 sm:p-6 lg:p-8 overflow-y-auto">
        {children}
      </main>
    </div>
  );
}
