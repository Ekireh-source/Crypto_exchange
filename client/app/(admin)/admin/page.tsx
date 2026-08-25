'use client';

import { Icon } from '@iconify/react';
import Link from 'next/link';

export default function AdminDashboardPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Admin Dashboard</h1>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-[#13151a] border border-[#22252e] rounded-2xl p-6 shadow-xl">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-gray-400 font-medium">Total Users</h3>
            <div className="size-10 rounded-xl bg-blue-500/10 flex items-center justify-center text-blue-500">
              <Icon icon="hugeicons:user-group" className="size-6" />
            </div>
          </div>
          <p className="text-3xl font-bold">Manage</p>
          <Link href="/admin/users" className="text-sm text-blue-400 hover:text-blue-300 mt-4 inline-block">
            View all users →
          </Link>
        </div>

        <div className="bg-[#13151a] border border-[#22252e] rounded-2xl p-6 shadow-xl">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-gray-400 font-medium">Transactions</h3>
            <div className="size-10 rounded-xl bg-green-500/10 flex items-center justify-center text-green-500">
              <Icon icon="hugeicons:transaction" className="size-6" />
            </div>
          </div>
          <p className="text-3xl font-bold">Monitor</p>
          <Link href="/admin/transactions" className="text-sm text-green-400 hover:text-green-300 mt-4 inline-block">
            View all transactions →
          </Link>
        </div>
      </div>
    </div>
  );
}
