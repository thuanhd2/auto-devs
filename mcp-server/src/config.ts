import dotenv from 'dotenv';

dotenv.config();

export const config = {
  apiUrl: process.env.AUTO_DEVS_API_URL || 'http://localhost:8098',
  apiKey: process.env.AUTO_DEVS_API_KEY || '',
  debug: process.env.MCP_DEBUG === 'true',
  enableCaching: process.env.ENABLE_CACHING !== 'false',
} as const;

if (!config.apiUrl) {
  throw new Error('AUTO_DEVS_API_URL environment variable is required');
}
