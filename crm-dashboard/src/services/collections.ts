import type { ApiResponse, PaginationParams } from './http'
// Existing tables filter locally. Fetch all pages until server-driven table
// pagination replaces them, rather than silently hiding records after page 1.
export async function allPages<T>(list: (params?: PaginationParams) => Promise<ApiResponse<T[]>>, params: PaginationParams = {}): Promise<T[]> {
  const result: T[] = []
  for (let page = 1; ; page++) {
    const response = await list({ ...params, page, limit: 100 })
    if (!response.success) throw new Error(response.error?.message ?? 'Unable to load records')
    result.push(...(response.data ?? []))
    if (!response.meta || page >= response.meta.totalPages) return result
  }
}
