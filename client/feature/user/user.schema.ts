import { z } from 'zod';

export const userSchema = z.object({
  id: z.string(),
  email: z.string(),
  phone: z.string().nullable().optional(),
  kyc_status: z.string(),
  created_at: z.string(),
  referral_code: z.string(),
});

export const referralStatsSchema = z.object({
  referral_code: z.string(),
  referral_link: z.string(),
  team_count: z.number(),
  fee_earnings: z.number(),
  total_volume: z.number(),
});

export const userProfileResponseSchema = z.object({
  user: userSchema,
  referral_stats: referralStatsSchema,
  referred_users: z.array(userSchema).optional(),
});

export type UserProfile = z.infer<typeof userSchema>;
export type ReferralStats = z.infer<typeof referralStatsSchema>;
export type UserProfileResponse = z.infer<typeof userProfileResponseSchema>;
