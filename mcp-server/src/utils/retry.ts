import { AppError, ErrorCode } from '../errors/app-error.js';
import { config } from '../config.js';
import { backendCircuitBreaker } from './circuit-breaker.js';

export interface RetryOptions {
  maxRetries?: number;
  initialDelayMs?: number;
  maxDelayMs?: number;
}

const DEFAULT_OPTIONS: Required<RetryOptions> = {
  maxRetries: 3,
  initialDelayMs: 100,
  maxDelayMs: 5000,
};

export async function withRetry<T>(
  fn: () => Promise<T>,
  options: RetryOptions = {}
): Promise<T> {
  if (!backendCircuitBreaker.canExecute()) {
    const state = backendCircuitBreaker.getState();
    throw new AppError(
      ErrorCode.SERVICE_UNAVAILABLE,
      `Backend API circuit breaker is ${state}. Failing fast to prevent request storm.`,
      {
        suggestion:
          'Wait for the backend to recover. The circuit will retry automatically after 30 seconds.',
        details: { circuitState: state },
      }
    );
  }

  const opts = { ...DEFAULT_OPTIONS, ...options };
  let lastError: Error | undefined;

  try {
    for (let attempt = 0; attempt < opts.maxRetries; attempt++) {
      try {
        const result = await fn();
        backendCircuitBreaker.recordSuccess();
        return result;
      } catch (error) {
        lastError = error instanceof Error ? error : new Error(String(error));

        const isAppError = error instanceof AppError;
        const isRetryable = isAppError && error.isRetryable();

        if (!isRetryable || attempt === opts.maxRetries - 1) {
          throw error;
        }

        const delayMs = Math.min(
          opts.initialDelayMs * Math.pow(2, attempt),
          opts.maxDelayMs
        );

        if (config.debug) {
          console.error(
            `[RETRY] Attempt ${attempt + 1}/${opts.maxRetries} failed, retrying in ${delayMs}ms`
          );
        }

        await new Promise((resolve) => setTimeout(resolve, delayMs));
      }
    }

    throw lastError || new Error('Max retries exceeded');
  } catch (error) {
    backendCircuitBreaker.recordFailure();
    throw error;
  }
}

export function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
