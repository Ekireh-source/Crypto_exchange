import { z } from 'zod';

// ─── Login ────────────────────────────────────────────────────
export const loginSchema = z.object({
  email: z
    .string()
    .min(1, 'Email is required')
    .email('Enter a valid email address'),
  password: z
    .string()
    .min(1, 'Password is required')
    .min(8, 'Password must be at least 8 characters'),
});

export type LoginDto = z.infer<typeof loginSchema>;

// ─── Register ─────────────────────────────────────────────────
export const registerSchema = z.object({
  email: z
    .string()
    .min(1, 'Email is required')
    .email('Enter a valid email address'),
  password: z
    .string()
    .min(1, 'Password is required')
    .min(8, 'Password must be at least 8 characters')
    .regex(/[A-Z]/, 'Must contain at least one uppercase letter')
    .regex(/[0-9]/, 'Must contain at least one number'),
  referral_code: z
    .string()
    .optional(),
});

export type RegisterDto = z.infer<typeof registerSchema>;

// ─── Auth API response ────────────────────────────────────────
export const authUserSchema = z.object({
  id: z.string(),
  email: z.string().email(),
  referral_code: z.string(),
});

export const authResponseSchema = z.object({
  user: authUserSchema,
  access_token: z.string(),
  refresh_token: z.string(),
});

export type AuthUser = z.infer<typeof authUserSchema>;
export type AuthResponse = z.infer<typeof authResponseSchema>;
