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
  confirm_password: z
    .string()
    .min(1, 'Please confirm your password'),
  referral_code: z
    .string()
    .optional(),
}).refine((data) => data.password === data.confirm_password, {
  message: "Passwords do not match",
  path: ['confirm_password'],
});

export type RegisterDto = z.infer<typeof registerSchema>;

// ─── Auth API response ────────────────────────────────────────
export const authUserSchema = z.object({
  id: z.string(),
  email: z.string().email(),
  referral_code: z.string(),
  role: z.string(),
});

export const authResponseSchema = z.object({
  user: authUserSchema,
  access_token: z.string(),
  refresh_token: z.string(),
});

export const registerResponseSchema = z.object({
  user: authUserSchema,
  message: z.string().optional(),
});

export type AuthUser = z.infer<typeof authUserSchema>;
export type AuthResponse = z.infer<typeof authResponseSchema>;
export type RegisterResponse = z.infer<typeof registerResponseSchema>;
