export enum CircuitState {
  CLOSED = 'CLOSED',
  OPEN = 'OPEN',
  HALF_OPEN = 'HALF_OPEN',
}

export interface CircuitBreakerOptions {
  failureThreshold?: number;
  resetTimeoutMs?: number;
}

export class CircuitBreaker {
  private state: CircuitState = CircuitState.CLOSED;
  private failureCount = 0;
  private lastFailureTime = 0;
  private halfOpenProbeInFlight = false;

  private readonly failureThreshold: number;
  private readonly resetTimeoutMs: number;

  constructor(options: CircuitBreakerOptions = {}) {
    this.failureThreshold = options.failureThreshold ?? 5;
    this.resetTimeoutMs = options.resetTimeoutMs ?? 30_000;
  }

  getState(): CircuitState {
    if (this.state === CircuitState.OPEN) {
      if (Date.now() - this.lastFailureTime >= this.resetTimeoutMs) {
        this.state = CircuitState.HALF_OPEN;
        this.halfOpenProbeInFlight = false;
      }
    }
    return this.state;
  }

  canExecute(): boolean {
    const state = this.getState();

    if (state === CircuitState.CLOSED) {
      return true;
    }

    if (state === CircuitState.OPEN) {
      return false;
    }

    // HALF_OPEN: allow a single probe request
    if (!this.halfOpenProbeInFlight) {
      this.halfOpenProbeInFlight = true;
      return true;
    }

    return false;
  }

  recordSuccess(): void {
    this.failureCount = 0;
    this.state = CircuitState.CLOSED;
    this.halfOpenProbeInFlight = false;
  }

  recordFailure(): void {
    this.lastFailureTime = Date.now();
    this.halfOpenProbeInFlight = false;

    if (this.state === CircuitState.HALF_OPEN) {
      this.state = CircuitState.OPEN;
      return;
    }

    this.failureCount += 1;
    if (this.failureCount >= this.failureThreshold) {
      this.state = CircuitState.OPEN;
    }
  }
}

/** Shared circuit breaker for all backend API calls */
export const backendCircuitBreaker = new CircuitBreaker({
  failureThreshold: 5,
  resetTimeoutMs: 30_000,
});
