/**
 * 环境变量配置
 */

export const isDevelopment = import.meta.env.MODE === 'development'
export const isProduction = import.meta.env.MODE === 'production'

const runtimeApiBaseUrl =
  typeof window !== 'undefined'
    ? ((window as any).__VITE_API_BASE_URL__ || (window as any).__VITE_RUNTIME_CONFIG__?.VITE_API_BASE_URL || '')
    : ''

export const API_BASE_URL = (typeof runtimeApiBaseUrl === 'string' && runtimeApiBaseUrl.trim() !== ''
  ? runtimeApiBaseUrl.trim()
  : import.meta.env.VITE_API_BASE_URL) || '/api/v1'

export const DEBUG_CONFIG = {
  enableNetworkLogging: isDevelopment,
} as const
