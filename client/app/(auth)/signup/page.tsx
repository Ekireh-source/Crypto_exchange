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

export default function SignupPage() {
  const router = useRouter();
  const dispatch = useAppDispatch();
  const [serverError, setServerError] = useState('');
  const [showPassword, setShowPassword] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterDto>({
    resolver: zodResolver(registerSchema),
  });

  const onSubmit = async (data: RegisterDto) => {
    setServerError('');
    try {
      const res = await authService.register(data);
      dispatch(
        setCredentials({
          user: res.user,
          accessToken: res.access_token,
          refreshToken: res.refresh_token,
        })
      );
      router.push('/dashboard');
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ??
        'Registration failed. Please try again.';
      setServerError(message);
    }
  };

  const inputBase =
    'w-full h-[54px] px-[18px] rounded-[14px] bg-[#1a1a24] border text-[#e8e8f0] text-[0.95rem] placeholder:text-[#555570] outline-none transition-all duration-200 focus:bg-[#1e1e2d] focus:shadow-[0_0_0_3px_rgba(108,99,255,0.15)]';
  const inputOk = 'border-[#2a2a3a] focus:border-[#6c63ff]';
  const inputErr = 'border-red-500/60 focus:border-red-500';

  return (
    <AuthCard title="CREATE ACCOUNT" subtitle="Join the future of Web3">
      <form className="w-full flex flex-col gap-3 mt-5" onSubmit={handleSubmit(onSubmit)} noValidate>

        {/* ── Server error banner ── */}
        {serverError && (
          <div className="flex items-center gap-2 px-4 py-3 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="shrink-0">
              <circle cx="12" cy="12" r="10" /><line x1="12" y1="8" x2="12" y2="12" /><line x1="12" y1="16" x2="12.01" y2="16" />
            </svg>
            {serverError}
          </div>
        )}

        {/* ── Email ── */}
        <div>
          <input
            id="signup-email"
            type="email"
            placeholder="Email"
            autoComplete="email"
            {...register('email')}
            className={`${inputBase} ${errors.email ? inputErr : inputOk}`}
          />
          {errors.email && (
            <p className="text-red-400 text-xs mt-1 ml-1">{errors.email.message}</p>
          )}
        </div>

        {/* ── Password ── */}
        <div>
          <div className="relative">
            <input
              id="signup-password"
              type={showPassword ? 'text' : 'password'}
              placeholder="Password"
              autoComplete="new-password"
              {...register('password')}
              className={`${inputBase} pr-[52px] ${errors.password ? inputErr : inputOk}`}
            />
            <button
              type="button"
              onClick={() => setShowPassword((v) => !v)}
              aria-label={showPassword ? 'Hide password' : 'Show password'}
              className="absolute right-4 top-1/2 -translate-y-1/2 text-[#555570] hover:text-[#7c6bff] transition-colors flex items-center"
            >
              {showPassword ? (
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94" />
                  <path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19" />
                  <line x1="1" y1="1" x2="23" y2="23" />
                </svg>
              ) : (
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                  <circle cx="12" cy="12" r="3" />
                </svg>
              )}
            </button>
          </div>
          {errors.password && (
            <p className="text-red-400 text-xs mt-1 ml-1">{errors.password.message}</p>
          )}
        </div>

        {/* ── Referral code (optional) ── */}
        <div>
          <input
            id="signup-referral"
            type="text"
            placeholder="Referral Code (optional)"
            autoComplete="off"
            {...register('referral_code')}
            className={`${inputBase} ${errors.referral_code ? inputErr : inputOk}`}
          />
          {errors.referral_code && (
            <p className="text-red-400 text-xs mt-1 ml-1">{errors.referral_code.message}</p>
          )}
        </div>

        {/* ── Submit ── */}
        <button
          id="signup-submit"
          type="submit"
          disabled={isSubmitting}
          className="w-full h-[54px] mt-1 rounded-[14px] bg-white text-[#0a0a0f] text-base font-bold tracking-wide flex items-center justify-center gap-2 shadow-[0_4px_24px_rgba(255,255,255,0.08)] transition-all duration-200 hover:bg-[#e8e8ff] hover:-translate-y-px hover:shadow-[0_8px_32px_rgba(255,255,255,0.14)] active:translate-y-0 disabled:opacity-60 disabled:cursor-not-allowed"
        >
          {isSubmitting ? (
            <span className="inline-block w-5 h-5 rounded-full border-[2.5px] border-black/20 border-t-black animate-spin" />
          ) : (
            'Create Account'
          )}
        </button>

        {/* ── Login link ── */}
        <p className="text-center text-[#7070a0] text-sm mt-1">
          Already have an account?{' '}
          <Link href="/login" className="text-[#7c6bff] font-semibold hover:text-[#9b8eff] transition-colors">
            Log In
          </Link>
        </p>
      </form>
    </AuthCard>
  );
}
