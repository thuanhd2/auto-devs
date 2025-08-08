export const API_CONFIG = {
  BASE_URL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8098/api/v1',
  WS_URL: import.meta.env.VITE_WS_BASE_URL || 'ws://localhost:8098/ws',
  TIMEOUT: 10000,
} as const

export const API_ENDPOINTS = {
  PROJECTS: '/projects',
  TASKS: '/tasks',
  EXECUTIONS: '/executions',
  PULL_REQUESTS: '/pull-requests',
} as const
