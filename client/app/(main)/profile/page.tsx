'use client';

import { useState, useEffect } from 'react';
import { Icon } from '@iconify/react';
import { userService } from '@/feature/user/user.service';
import { type UserProfileResponse } from '@/feature/user/user.schema';

type Tab = 'user_info' | 'referrals';

export default function ProfilePage() {
  const [profileData, setProfileData] = useState<UserProfileResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<Tab>('user_info');
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        const data = await userService.getProfile();
        setProfileData(data);
      } catch (err) {
        console.error('Failed to load profile:', err);
      } finally {
        setLoading(false);
      }
    };
    fetchProfile();
  }, []);

  const handleCopyLink = () => {
    if (profileData?.referral_stats?.referral_link) {
      navigator.clipboard.writeText(profileData.referral_stats.referral_link);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center pt-20">
        <div className="animate-spin text-[#4f7cf7]">
          <Icon icon="hugeicons:loading-03" className="size-8" />
        </div>
      </div>
    );
  }

  if (!profileData) {
    return (
      <div className="flex flex-col items-center justify-center pt-20 gap-4">
        <div className="text-red-400 p-4 bg-red-500/10 rounded-xl border border-red-500/20">
          Failed to load profile
        </div>
      </div>
    );
  }

  const { user, referral_stats } = profileData;
  const initial = user.email.charAt(0).toUpperCase();

  return (
    <div className="flex flex-col w-full max-w-[800px] mx-auto animate-in fade-in duration-500 pb-20 font-sans gap-8">
      
      {/* Profile Header Card */}
      <div className="w-full bg-[#13151a] border border-[#22252e] rounded-[24px] p-8 shadow-2xl flex flex-col items-center gap-4 relative overflow-hidden">
        {/* Subtle background glow */}
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[300px] h-[300px] bg-blue-600/10 rounded-full blur-[80px] pointer-events-none" />
        
        <div className="relative size-24 rounded-full bg-blue-600 border-[4px] border-[#13151a] shadow-xl flex items-center justify-center text-4xl font-bold text-white z-10">
          {initial}
        </div>
        
        <div className="flex flex-col items-center gap-1 z-10 text-center">
          <h1 className="text-2xl font-bold text-white tracking-tight">{user.email}</h1>
          <div className="flex items-center gap-2 mt-2">
            <span
              className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-bold uppercase tracking-wider border ${
                user.kyc_status === 'verified'
                  ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                  : user.kyc_status === 'pending'
                  ? 'bg-amber-500/10 text-amber-400 border-amber-500/20'
                  : 'bg-red-500/10 text-red-400 border-red-500/20'
              }`}
            >
              KYC: {user.kyc_status}
            </span>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-2 border-b border-[#22252e]">
        <button
          onClick={() => setActiveTab('user_info')}
          className={`pb-4 px-4 text-[15px] font-semibold transition-all relative ${
            activeTab === 'user_info' ? 'text-white' : 'text-[#888c99] hover:text-white'
          }`}
        >
          User Info
          {activeTab === 'user_info' && (
            <div className="absolute bottom-0 left-0 w-full h-[2px] bg-blue-600 rounded-t-full" />
          )}
        </button>
        <button
          onClick={() => setActiveTab('referrals')}
          className={`pb-4 px-4 text-[15px] font-semibold transition-all relative ${
            activeTab === 'referrals' ? 'text-white' : 'text-[#888c99] hover:text-white'
          }`}
        >
          Referrals
          {activeTab === 'referrals' && (
            <div className="absolute bottom-0 left-0 w-full h-[2px] bg-blue-600 rounded-t-full" />
          )}
        </button>
      </div>

      {/* Tab Content */}
      <div className="w-full bg-[#13151a] border border-[#22252e] rounded-[24px] p-6 sm:p-8 shadow-xl">
        
        {activeTab === 'user_info' && (
          <div className="flex flex-col gap-6 animate-in fade-in zoom-in-95 duration-300">
            <h2 className="text-xl font-bold text-white mb-2">Account Details</h2>
            
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="p-4 bg-[#1c1f26] border border-[#22252e] rounded-2xl flex flex-col gap-1">
                <span className="text-sm font-semibold text-[#888c99]">Email Address</span>
                <span className="text-white font-medium">{user.email}</span>
              </div>
              <div className="p-4 bg-[#1c1f26] border border-[#22252e] rounded-2xl flex flex-col gap-1">
                <span className="text-sm font-semibold text-[#888c99]">Phone Number</span>
                <span className="text-white font-medium">{user.phone || 'Not provided'}</span>
              </div>
              <div className="p-4 bg-[#1c1f26] border border-[#22252e] rounded-2xl flex flex-col gap-1">
                <span className="text-sm font-semibold text-[#888c99]">User ID</span>
                <span className="text-white font-mono text-sm truncate">{user.id}</span>
              </div>
              <div className="p-4 bg-[#1c1f26] border border-[#22252e] rounded-2xl flex flex-col gap-1">
                <span className="text-sm font-semibold text-[#888c99]">Joined Date</span>
                <span className="text-white font-medium">
                  {new Date(user.created_at).toLocaleDateString(undefined, {
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric'
                  })}
                </span>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'referrals' && (
          <div className="flex flex-col gap-8 animate-in fade-in zoom-in-95 duration-300">
            <div className="flex flex-col gap-2">
              <h2 className="text-xl font-bold text-white">Invite Friends</h2>
              <p className="text-sm text-[#888c99]">
                Share your referral link to build your team.
              </p>
            </div>

            <div className="flex flex-col sm:flex-row gap-4">
              {/* Stats Cards */}
              <div className="flex-1 p-5 bg-gradient-to-br from-blue-600/20 to-purple-600/20 border border-blue-500/20 rounded-2xl flex flex-col items-center justify-center gap-2 text-center">
                <Icon icon="hugeicons:user-group" className="size-8 text-blue-400" />
                <span className="text-sm font-semibold text-blue-200">Team Count</span>
                <span className="text-3xl font-bold text-white">{referral_stats.team_count}</span>
              </div>
              
              <div className="flex-1 p-5 bg-[#1c1f26] border border-[#22252e] rounded-2xl flex flex-col items-center justify-center gap-2 text-center">
                <Icon icon="hugeicons:money-bag-02" className="size-8 text-emerald-400" />
                <span className="text-sm font-semibold text-[#888c99]">Fee Earnings</span>
                <span className="text-3xl font-bold text-white">${referral_stats.fee_earnings.toFixed(2)}</span>
              </div>
            </div>

            <div className="flex flex-col gap-4">
              <div className="flex flex-col gap-2">
                <label className="text-sm font-semibold text-[#888c99]">Your Referral Code</label>
                <div className="flex items-center justify-between p-4 bg-[#1c1f26] border border-[#22252e] rounded-xl">
                  <span className="font-mono text-lg font-bold text-white tracking-widest">
                    {referral_stats.referral_code}
                  </span>
                </div>
              </div>

              <div className="flex flex-col gap-2">
                <label className="text-sm font-semibold text-[#888c99]">Your Referral Link</label>
                <div className="flex items-center gap-2">
                  <div className="flex-1 px-4 py-3 bg-[#1c1f26] border border-[#22252e] rounded-xl overflow-hidden">
                    <span className="font-mono text-sm text-white truncate block">
                      {referral_stats.referral_link}
                    </span>
                  </div>
                  <button
                    onClick={handleCopyLink}
                    className="flex items-center justify-center size-12 shrink-0 bg-blue-600 hover:bg-blue-500 text-white rounded-xl transition-colors shadow-lg shadow-blue-600/20"
                    title="Copy Link"
                  >
                    <Icon icon={copied ? "hugeicons:checkmark-circle-02" : "hugeicons:copy-01"} className="size-5" />
                  </button>
                </div>
              </div>
            </div>

            {/* Referred Users List */}
            <div className="flex flex-col gap-4 mt-4">
              <h3 className="text-lg font-bold text-white">Your Team</h3>
              
              {(!profileData.referred_users || profileData.referred_users.length === 0) ? (
                <div className="p-8 bg-[#1c1f26] border border-[#22252e] rounded-xl flex flex-col items-center justify-center gap-3 text-center">
                  <Icon icon="hugeicons:user-group" className="size-12 text-[#888c99]/50" />
                  <p className="text-[#888c99] text-sm">You haven't referred anyone yet.</p>
                </div>
              ) : (
                <div className="w-full overflow-x-auto rounded-xl border border-[#22252e]">
                  <table className="w-full text-left border-collapse">
                    <thead>
                      <tr className="bg-[#1c1f26] border-b border-[#22252e]">
                        <th className="py-3 px-4 text-xs font-semibold text-[#888c99] uppercase tracking-wider">Email</th>
                        <th className="py-3 px-4 text-xs font-semibold text-[#888c99] uppercase tracking-wider">Joined Date</th>
                        <th className="py-3 px-4 text-xs font-semibold text-[#888c99] uppercase tracking-wider text-right">Status</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-[#22252e] bg-[#13151a]">
                      {profileData.referred_users.map((ru) => (
                        <tr key={ru.id} className="hover:bg-[#1c1f26]/60 transition-colors">
                          <td className="py-3 px-4">
                            <span className="text-sm font-medium text-white">
                              {ru.email.replace(/(.{2})(.*)(@.*)/, '$1***$3')}
                            </span>
                          </td>
                          <td className="py-3 px-4">
                            <span className="text-sm text-[#888c99]">
                              {new Date(ru.created_at).toLocaleDateString()}
                            </span>
                          </td>
                          <td className="py-3 px-4 text-right">
                            <span className={`inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider ${
                                ru.kyc_status === 'verified'
                                  ? 'bg-emerald-500/10 text-emerald-400'
                                  : ru.kyc_status === 'pending'
                                  ? 'bg-amber-500/10 text-amber-400'
                                  : 'bg-red-500/10 text-red-400'
                              }`}
                            >
                              {ru.kyc_status}
                            </span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>

          </div>
        )}

      </div>
    </div>
  );
}
