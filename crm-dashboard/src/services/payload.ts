// Forms contain display names, local row IDs and calculated fields. Only these
// explicit editable fields cross the write boundary.
export function editable(data: object, fields: readonly string[]): Record<string, unknown> {
  const source = data as Record<string, unknown>
  return Object.fromEntries(fields.filter(key => source[key] !== undefined).map(key => {
    const value = source[key]
    return [key, key.endsWith('Id') && value === '' ? null : value]
  }))
}
export function errorMessage(error: unknown): string {
  const e = error as { response?: { data?: { error?: { message?: string } } }; message?: string }
  return e.response?.data?.error?.message ?? e.message ?? 'The operation failed. Please try again.'
}
