export const roles = [
  ['admin', 'Administrator'], ['sales_manager', 'Sales Manager'], ['sales_executive', 'Sales Executive'],
  ['pre_sales', 'Pre-Sales'], ['procurement_manager', 'Procurement Manager'], ['procurement_officer', 'Procurement Officer'],
  ['warehouse_manager', 'Warehouse Manager'], ['project_manager', 'Project Manager'], ['viewer', 'Viewer'],
] as const
export const departments = ['sales', 'pre-sales', 'technical', 'support', 'marketing', 'management', 'operations'] as const
export function roleLabel(role: string) { return roles.find(([value]) => value === role)?.[1] ?? role }
export function passwordError(password: string, confirmation: string): string {
  if (password.length < 8) return 'Use at least 8 characters for the password.'
  if (new TextEncoder().encode(password).length > 72) return 'The password must be at most 72 bytes.'
  if (password !== confirmation) return 'The passwords do not match.'
  return ''
}
