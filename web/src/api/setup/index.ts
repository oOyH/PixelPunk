import { get, post } from '@/utils/network/http'
import type { ApiResult } from '@/utils/network/http-types'
import type { DatabaseTestRequest, InstallRequest, InstallStatus, RedisTestRequest } from '../types'

export function getInstallStatus(): Promise<ApiResult<InstallStatus>> {
  return get<InstallStatus>('/setup/status')
}

export function testDatabaseConnection(data: DatabaseTestRequest): Promise<ApiResult<null>> {
  return post<null>('/setup/test-connection', data)
}

export function testRedisConnection(data: RedisTestRequest): Promise<ApiResult<null>> {
  return post<null>('/setup/test-redis', data)
}

export function installSystem(data: InstallRequest): Promise<ApiResult<null>> {
  return post<null>('/setup/install', data)
}

export default {
  getInstallStatus,
  testDatabaseConnection,
  testRedisConnection,
  installSystem,
}
