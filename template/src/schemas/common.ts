import { z } from 'zod';

export const ErrorResponseSchema = z.object({
  error: z.string(),
  details: z.any().optional()
});

export const HealthResponseSchema = z.object({
  status: z.enum(['healthy', 'unhealthy']),
  timestamp: z.string()
});