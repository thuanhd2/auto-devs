import axios from 'axios';
import { config } from '../config.js';
import { sleep } from './retry.js';

export interface HealthCheckResult {
  status: string;
  endpoint: string;
}

const HEALTH_PATHS = ['/health', '/api/v1/health'];
const MAX_ATTEMPTS = 5;
const RETRY_DELAY_MS = 1000;

export async function checkBackendHealth(): Promise<HealthCheckResult> {
  let lastError: Error | undefined;

  for (let attempt = 1; attempt <= MAX_ATTEMPTS; attempt++) {
    for (const path of HEALTH_PATHS) {
      const endpoint = `${config.apiUrl}${path}`;

      try {
        const response = await axios.get(endpoint, {
          timeout: 5000,
          headers: config.apiKey
            ? { Authorization: `Bearer ${config.apiKey}` }
            : undefined,
          validateStatus: (status) => status === 200 || status === 503,
        });

        const status =
          typeof response.data?.status === 'string'
            ? response.data.status
            : 'ok';

        console.error(
          `[MCP] Backend health check passed (${endpoint}, status=${status})`
        );

        return { status, endpoint };
      } catch (error) {
        lastError = error instanceof Error ? error : new Error(String(error));
        console.error(
          `[MCP] Health check attempt ${attempt}/${MAX_ATTEMPTS} failed for ${endpoint}: ${lastError.message}`
        );
      }
    }

    if (attempt < MAX_ATTEMPTS) {
      await sleep(RETRY_DELAY_MS);
    }
  }

  throw new Error(
    `Backend API at ${config.apiUrl} is not reachable. ` +
      `Ensure the Auto-Devs server is running and accessible. ` +
      `Last error: ${lastError?.message ?? 'unknown'}`
  );
}
