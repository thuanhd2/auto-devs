import { z } from 'zod';
import { AppError, ErrorCode } from '../errors/app-error.js';

export function validateInput<T extends z.ZodType>(
  schema: T,
  input: unknown
): z.infer<T> {
  const result = schema.safeParse(input);

  if (!result.success) {
    const fieldErrors = result.error.flatten().fieldErrors;
    const messages = Object.entries(fieldErrors)
      .map(([field, errors]) => `${field}: ${(errors ?? []).join(', ')}`)
      .join('; ');

    throw new AppError(
      ErrorCode.INVALID_INPUT,
      messages ? `Invalid tool input: ${messages}` : 'Invalid tool input',
      {
        details: { fields: fieldErrors },
        suggestion: 'Verify required fields and types match the tool schema',
      }
    );
  }

  return result.data;
}
