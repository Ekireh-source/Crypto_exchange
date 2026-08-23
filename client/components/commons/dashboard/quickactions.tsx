'use client';

import { useState } from 'react';
import { Icon } from '@iconify/react';
import SendModal from './sendmodal';
import ReceiveModal from './receivemodal';

interface QuickActionItem {
  id: string;
  label: string;
  icon: string;
  onClick: () => void;
}

export default function QuickActions() {
  const [isSendOpen, setIsSendOpen] = useState(false);
  const [isReceiveOpen, setIsReceiveOpen] = useState(false);

  const actions: QuickActionItem[] = [
    {
      id: 'send',
      label: 'Send crypto',
      icon: 'hugeicons:arrow-up-01',
      onClick: () => setIsSendOpen(true),
    },
    {
      id: 'receive',
      label: 'Receive crypto',
      icon: 'hugeicons:arrow-down-01',
      onClick: () => setIsReceiveOpen(true),
    },

  ];

  return (
    <>
      <div className="w-full bg-[#0a0b0d] border border-[#22252e] rounded-[20px] p-3 flex flex-col gap-1 font-sans">
        {actions.map((action) => (
          <div
            key={action.id}
            onClick={action.onClick}
            className="flex items-center gap-4 p-3.5 hover:bg-[#16181d] rounded-[14px] cursor-pointer transition-colors group"
          >
            <div className="size-10 rounded-full bg-[#4f7cf7] flex items-center justify-center shrink-0 shadow-sm">
              <Icon icon={action.icon} className="size-5 text-white stroke-[2.5]" />
            </div>
            <span className="text-[16px] font-bold text-white tracking-tight">
              {action.label}
            </span>
          </div>
        ))}
      </div>

      {/* Modals */}
      <SendModal isOpen={isSendOpen} onClose={() => setIsSendOpen(false)} />
      <ReceiveModal isOpen={isReceiveOpen} onClose={() => setIsReceiveOpen(false)} />
    </>
  );
}
