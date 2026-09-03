'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useAppDispatch } from '@/store/hooks';
import { setCredentials } from '@/store/authSlice';
import { authService } from '@/feature/auth/auth.service';
import { registerSchema, type RegisterDto } from '@/feature/auth/auth.schema';
import AuthCard from '../_components/AuthCard';
import Link from 'next/link';
import { Icon } from '@iconify/react';
import { toast } from 'sonner';

export default function SignupPage() {
  const router = useRouter();
  const dispatch = useAppDispatch();
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterDto>({
    resolver: zodResolver(registerSchema),
  });

  const onSubmit = async (data: RegisterDto) => {
    try {
      const res = await authService.register(data);
      router.push('/login');
    } catch (err: any) {
      const data = err.response?.data;
      const message = data?.message || data?.error || 'Registration failed. Please try again.';
      toast.error(message);
    }
  };

  const inputBase =
    'w-full h-[52px] pl-11 pr-4 rounded-md bg-[#01010d]/50 border text-slate-200 text-[0.95rem] placeholder:text-slate-500 outline-none transition-all duration-200 focus:bg-[#01010d]';
  const inputOk = 'border-white/10 focus:border-indigo-500/50';
  const inputErr = 'border-pink-500/60 focus:border-pink-500';

  return (
    <AuthCard
      title="Create Account"
      subtitle="Join the future of Web3 trading"
      leftTitle="Start Your"
      leftSubtitle="Journey"
      leftDescription="Create an account and start trading&#10;with zero hidden fees."
    >
      <form className="w-full flex flex-col gap-4 mt-2" onSubmit={handleSubmit(onSubmit)} noValidate>

        {/* ── Email ── */}
        <div className="flex flex-col gap-1.5">
          <label htmlFor="signup-email" className="text-[#a0a0b0] text-[0.85rem] font-medium pl-1">
            Email Address
          </label>
          <div className="relative">
            <Icon icon="lucide:mail" className="absolute left-4 top-1/2 -translate-y-1/2 text-[#555570] text-lg pointer-events-none" />
            <input
              id="signup-email"
              type="email"
              placeholder="you@example.com"
              autoComplete="email"
              {...register('email')}
              className={`${inputBase} ${errors.email ? inputErr : inputOk}`}
            />
          </div>
          {errors.email && (
            <p className="text-red-400 text-xs mt-0.5 ml-1">{errors.email.message}</p>
          )}
        </div>

        {/* ── Password ── */}
        <div className="flex flex-col gap-1.5">
          <label htmlFor="signup-password" className="text-[#a0a0b0] text-[0.85rem] font-medium pl-1">
            Password
          </label>
          <div className="relative">
            <Icon icon="lucide:lock" className="absolute left-4 top-1/2 -translate-y-1/2 text-[#555570] text-lg pointer-events-none" />
            <input
              id="signup-password"
              type={showPassword ? 'text' : 'password'}
              placeholder="••••••••"
              autoComplete="new-password"
              {...register('password')}
              className={`${inputBase} pr-[52px] ${errors.password ? inputErr : inputOk}`}
            />
            <button
              type="button"
              onClick={() => setShowPassword((v) => !v)}
              aria-label={showPassword ? 'Hide password' : 'Show password'}
              className="absolute right-4 top-1/2 -translate-y-1/2 text-[#555570] hover:text-[#f0c78a] transition-colors flex items-center justify-center w-6 h-6"
            >
              <Icon icon={showPassword ? 'lucide:eye-off' : 'lucide:eye'} className="text-lg" />
            </button>
          </div>
          {errors.password && (
            <p className="text-red-400 text-xs mt-0.5 ml-1">{errors.password.message}</p>
          )}
        </div>

        {/* ── Confirm Password ── */}
        <div className="flex flex-col gap-1.5">
          <label htmlFor="signup-confirm-password" className="text-[#a0a0b0] text-[0.85rem] font-medium pl-1">
            Confirm Password
          </label>
          <div className="relative">
            <Icon icon="lucide:lock" className="absolute left-4 top-1/2 -translate-y-1/2 text-[#555570] text-lg pointer-events-none" />
            <input
              id="signup-confirm-password"
              type={showConfirmPassword ? 'text' : 'password'}
              placeholder="••••••••"
              autoComplete="new-password"
              {...register('confirm_password')}
              className={`${inputBase} pr-[52px] ${errors.confirm_password ? inputErr : inputOk}`}
            />
            <button
              type="button"
              onClick={() => setShowConfirmPassword((v) => !v)}
              aria-label={showConfirmPassword ? 'Hide password' : 'Show password'}
              className="absolute right-4 top-1/2 -translate-y-1/2 text-[#555570] hover:text-[#f0c78a] transition-colors flex items-center justify-center w-6 h-6"
            >
              <Icon icon={showConfirmPassword ? 'lucide:eye-off' : 'lucide:eye'} className="text-lg" />
            </button>
          </div>
          {errors.confirm_password && (
            <p className="text-red-400 text-xs mt-0.5 ml-1">{errors.confirm_password.message}</p>
          )}
        </div>

        {/* ── Referral code (optional) ── */}
        <div className="flex flex-col gap-1.5">
          <label htmlFor="signup-referral" className="text-[#a0a0b0] text-[0.85rem] font-medium pl-1">
            Referral Code (Optional)
          </label>
          <div className="relative">
            <Icon icon="lucide:users" className="absolute left-4 top-1/2 -translate-y-1/2 text-[#555570] text-lg pointer-events-none" />
            <input
              id="signup-referral"
              type="text"
              placeholder="Code"
              autoComplete="off"
              {...register('referral_code')}
              className={`${inputBase} ${errors.referral_code ? inputErr : inputOk}`}
            />
          </div>
          {errors.referral_code && (
            <p className="text-red-400 text-xs mt-0.5 ml-1">{errors.referral_code.message}</p>
          )}
        </div>

        {/* ── Submit ── */}
        <button
          id="signup-submit"
          type="submit"
          disabled={isSubmitting}
          className="w-full h-[52px] mt-2 rounded-md bg-indigo-600 hover:bg-indigo-500 text-white text-[0.95rem] font-medium tracking-wide flex items-center justify-center gap-2 shadow-[0_4px_24px_rgba(79,70,229,0.15)] transition-all duration-200 active:scale-[0.99] disabled:opacity-60 disabled:cursor-not-allowed"
        >
          {isSubmitting ? (
            <span className="inline-block w-5 h-5 rounded-full border-[2.5px] border-white/20 border-t-white animate-spin" />
          ) : (
            <>
              Create Account
              <Icon icon="lucide:arrow-right" className="text-lg ml-1" />
            </>
          )}
        </button>

        {/* ── Divider ── */}
        <div className="flex items-center my-4">
          <div className="flex-1 h-px bg-[#2c2d3a]"></div>
          <span className="px-4 text-[#555570] text-[0.8rem]">or continue with</span>
          <div className="flex-1 h-px bg-[#2c2d3a]"></div>
        </div>

        {/* ── Social Login ── */}
        <div className="flex justify-center gap-4">
          <button type="button" className="w-[52px] h-[52px] rounded-full bg-[#14151a] border border-[#2c2d3a] flex items-center justify-center text-white hover:bg-[#1a1b24] hover:border-[#f0c78a] transition-all">
            <Icon icon="logos:google-icon" className="text-xl" />
          </button>
          <button type="button" className="w-[52px] h-[52px] rounded-full bg-[#14151a] border border-[#2c2d3a] flex items-center justify-center text-white hover:bg-[#1a1b24] hover:border-[#f0c78a] transition-all">
            <Icon icon="logos:apple" className="text-[22px] mb-0.5" />
          </button>
          <button type="button" className="w-[52px] h-[52px] rounded-full bg-[#14151a] border border-[#2c2d3a] flex items-center justify-center text-white hover:bg-[#1a1b24] hover:border-[#f0c78a] transition-all">
            <Icon icon="lucide:github" className="text-[22px]" />
          </button>
        </div>

        {/* ── Login link ── */}
        <p className="text-center text-[#7070a0] text-[0.85rem] mt-4">
          Already have an account?{' '}
          <Link href="/login" className="text-[#f0c78a] font-semibold hover:text-[#ffd699] transition-colors ml-1">
            Log In
          </Link>
        </p>
      </form>
    </AuthCard>
  );
}
