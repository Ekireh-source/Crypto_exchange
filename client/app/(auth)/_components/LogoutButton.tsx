'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAppDispatch } from '@/store/hooks';
import { logout } from '@/store/authSlice';

interface LogoutButtonProps {
  /** Render as icon-only button (compact, for navbars) */
  iconOnly?: boolean;
  className?: string;
}

export default function LogoutButton({ iconOnly = false, className = '' }: LogoutButtonProps) {
  const dispatch = useAppDispatch();
  const router = useRouter();
  const [loading, setLoading] = useState(false);

  const handleLogout = async () => {
    setLoading(true);
    dispatch(logout());
    await new Promise((r) => setTimeout(r, 300));
    router.push('/login');
  };

  return (
    <button
      id="logout-btn"
      onClick={handleLogout}
      disabled={loading}
      aria-label="Log out"
      title="Log out"
      className={[
        'inline-flex items-center gap-2 rounded-[10px] border border-[#2a2a3a] bg-transparent',
        'text-[#7070a0] text-sm font-medium font-[inherit] cursor-pointer',
        'transition-all duration-200',
        'hover:bg-red-500/[0.08] hover:text-red-400 hover:border-red-500/30 hover:shadow-[0_0_16px_rgba(220,38,38,0.08)]',
        'disabled:opacity-50 disabled:cursor-not-allowed',
        iconOnly ? 'p-[10px]' : 'px-5 py-[10px]',
        className,
      ].join(' ')}
    >
      {loading ? (
        <span className="inline-block w-4 h-4 rounded-full border-2 border-white/25 border-t-white animate-spin" />
      ) : (
        <>
          <svg
            width="18" height="18" viewBox="0 0 24 24"
            fill="none" stroke="currentColor" strokeWidth="2"
            strokeLinecap="round" strokeLinejoin="round"
          >
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
            <polyline points="16 17 21 12 16 7" />
            <line x1="21" y1="12" x2="9" y2="12" />
          </svg>
          {!iconOnly && <span>Log Out</span>}
        </>
      )}
    </button>
  );
}
