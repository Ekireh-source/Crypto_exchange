'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useAppDispatch } from '@/store/hooks';
import { setCredentials } from '@/store/authSlice';
import { authService } from '@/feature/auth/auth.service';
import { loginSchema, type LoginDto } from '@/feature/auth/auth.schema';
import AuthCard from '../_components/AuthCard';
import Link from 'next/link';
import { Icon } from '@iconify/react';

export default function LoginPage() {
  const router = useRouter();
  const dispatch = useAppDispatch();
  const [serverError, setServerError] = useState('');
  const [showPassword, setShowPassword] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginDto>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (data: LoginDto) => {
    setServerError('');
    try {
      const res = await authService.login(data);
      dispatch(
        setCredentials({
          user: res.user,
          accessToken: res.access_token,
          refreshToken: res.refresh_token,
        })
      );
      if (res.user.role === 'admin' || res.user.role === 'superadmin') {
        router.push('/admin');
      } else {
        router.push('/dashboard');
      }
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ??
        'Login failed. Please check your credentials.';
      setServerError(message);
    }
  };

  const inputBase =
    'w-full h-[52px] pl-11 pr-4 rounded-xl bg-[#14151a] border text-[#e8e8f0] text-[0.95rem] placeholder:text-[#555570] outline-none transition-all duration-200 focus:bg-[#1a1b24]';
  const inputOk = 'border-[#2c2d3a] focus:border-[#f0c78a]';
  const inputErr = 'border-red-500/60 focus:border-red-500';

  return (
    <AuthCard 
      title="Login" 
      subtitle="Login to your account to continue"
      leftTitle="Welcome"
      leftSubtitle="Back"
      leftDescription="Glad to see you again.&#10;Let's continue where you left off."
    >
      <form className="w-full flex flex-col gap-4 mt-2" onSubmit={handleSubmit(onSubmit)} noValidate>

        {/* ── Server error banner ── */}
        {serverError && (
          <div className="flex items-center gap-2 px-4 py-3 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm">
            <Icon icon="lucide:alert-circle" className="text-lg shrink-0" />
            {serverError}
          </div>
        )}

        {/* ── Email ── */}
        <div className="flex flex-col gap-1.5">
          <label htmlFor="login-email" className="text-[#a0a0b0] text-[0.85rem] font-medium pl-1">
            Email Address
          </label>
          <div className="relative">
            <Icon icon="lucide:mail" className="absolute left-4 top-1/2 -translate-y-1/2 text-[#555570] text-lg pointer-events-none" />
            <input
              id="login-email"
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
          <label htmlFor="login-password" className="text-[#a0a0b0] text-[0.85rem] font-medium pl-1">
            Password
          </label>
          <div className="relative">
            <Icon icon="lucide:lock" className="absolute left-4 top-1/2 -translate-y-1/2 text-[#555570] text-lg pointer-events-none" />
            <input
              id="login-password"
              type={showPassword ? 'text' : 'password'}
              placeholder="••••••••"
              autoComplete="current-password"
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

        {/* ── Forgot password ── */}
        <div className="text-right -mt-2">
          <Link href="/forgot-password" className="text-[0.8rem] font-medium text-[#f0c78a] hover:text-[#ffd699] transition-colors">
            Forgot Password?
          </Link>
        </div>

        {/* ── Submit ── */}
        <button
          id="login-submit"
          type="submit"
          disabled={isSubmitting}
          className="w-full h-[52px] mt-2 rounded-xl bg-gradient-to-r from-[#eec18d] to-[#d6a571] text-[#14151a] text-[0.95rem] font-bold tracking-wide flex items-center justify-center gap-2 shadow-[0_4px_24px_rgba(238,193,141,0.15)] transition-all duration-200 hover:scale-[1.01] hover:shadow-[0_6px_32px_rgba(238,193,141,0.25)] active:scale-[0.99] disabled:opacity-60 disabled:cursor-not-allowed disabled:hover:scale-100"
        >
          {isSubmitting ? (
            <span className="inline-block w-5 h-5 rounded-full border-[2.5px] border-[#14151a]/20 border-t-[#14151a] animate-spin" />
          ) : (
            <>
              Login
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

        {/* ── Sign-up link ── */}
        <p className="text-center text-[#7070a0] text-[0.85rem] mt-6">
          Don't have an account?{' '}
          <Link href="/signup" className="text-[#f0c78a] font-semibold hover:text-[#ffd699] transition-colors ml-1">
            Sign up
          </Link>
        </p>
      </form>
    </AuthCard>
  );
}
