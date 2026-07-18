import { ErrorCode, errorHttpStatus } from './error-codes.js';

export { ErrorCode } from './error-codes.js';

export interface ErrorDetails {
  code: ErrorCode;
  message: string;
  details?: Record<string, unknown>;
  retryAfter?: number;
  suggestion?: string;
}

export class AppError extends Error implements ErrorDetails {
  code: ErrorCode;
  details?: Record<string, unknown>;
  retryAfter?: number;
  suggestion?: string;

  constructor(
    code: ErrorCode,
    message: string,
    options?: {
      details?: Record<string, unknown>;
      retryAfter?: number;
      suggestion?: string;
    }
  ) {
    super(message);
    this.name = 'AppError';
    this.code = code;
    this.details = options?.details;
    this.retryAfter = options?.retryAfter;
    this.suggestion = options?.suggestion;
  }

  getHttpStatus(): number {
    return errorHttpStatus[this.code];
  }

  isRetryable(): boolean {
    return this.code === ErrorCode.RATE_LIMITED ||
           this.code === ErrorCode.SERVICE_UNAVAILABLE ||
           this.code === ErrorCode.TIMEOUT;
  }

  toJSON() {
    return {
      code: this.code,
      message: this.message,
      details: this.details,
      retryAfter: this.retryAfter,
      suggestion: this.suggestion,
    };
  }
}
