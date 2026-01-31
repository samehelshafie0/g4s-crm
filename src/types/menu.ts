export interface MenuItem {
  id: string
  label: string
  icon: string
  route: string
  children?: MenuItem[]
  badge?: string
  description?: string
}

export const menuItems: MenuItem[] = [
  {
    id: 'dashboard',
    label: 'Dashboard & Analytics',
    icon: '📊',
    route: '/dashboard',
    description: 'KPIs, Sales Overview, Recent Updates'
  },
  {
    id: 'customers',
    label: 'Customers & Contacts',
    icon: '👥',
    route: '/customers',
    description: 'Accounts, Sites, Contacts, Client History'
  },
  {
    id: 'leads',
    label: 'Leads & Opportunities',
    icon: '🎯',
    route: '/leads',
    description: 'Lead Management, Opportunities, Workflow'
  },
  {
    id: 'products',
    label: 'Product Catalogue',
    icon: '📦',
    route: '/products',
    description: 'Materials & Services, Multi-currency Costing'
  },
  {
    id: 'manufacturers',
    label: 'Manufacturers Database',
    icon: '🏭',
    route: '/manufacturers',
    description: 'Manufacturer Catalog, Categories, Product Models'
  },
  {
    id: 'pricebooks',
    label: 'Price Books & FX Rates',
    icon: '💰',
    route: '/pricebooks',
    description: 'Price Management, Exchange Rates, Audit Trail'
  },
  {
    id: 'exchange-rates',
    label: 'Exchange Rates Settings',
    icon: '💱',
    route: '/exchange-rates',
    description: 'Live Rates, Manual Updates, Historical Tracking'
  },
  {
    id: 'inventory',
    label: 'Inventory & Warehouses',
    icon: '🏭',
    route: '/inventory',
    description: 'Multi-warehouse, Stock Ledger, Movements'
  },
  {
    id: 'procurement',
    label: 'Procurement',
    icon: '🛒',
    route: '/procurement',
    description: 'Purchase Orders, Vendors, Receipts'
  },
  {
    id: 'manpower',
    label: 'Manpower & Rates',
    icon: '👷',
    route: '/manpower',
    description: 'Roles, Departments, Rate Management'
  },
  {
    id: 'teams',
    label: 'Teams & Personnel',
    icon: '👥',
    route: '/teams',
    description: 'Sales Teams, Pre-Sales Teams, Team Members'
  },
  {
    id: 'recurring',
    label: 'Recurring Services',
    icon: '🔄',
    route: '/recurring',
    description: 'Monthly Rentals, Service Subscriptions'
  },
  {
    id: 'quoting',
    label: 'Quoting & Proposals',
    icon: '📝',
    route: '/quoting',
    description: 'CPQ, Internal/Client Views, Proposals'
  },
  {
    id: 'projects',
    label: 'Projects & Handover',
    icon: '🚀',
    route: '/projects',
    description: 'Project Management, BOM, Budget Tracking'
  },
  {
    id: 'documents',
    label: 'Terms & Documents',
    icon: '📄',
    route: '/documents',
    description: 'Templates, Terms & Conditions'
  },
  {
    id: 'workflows',
    label: 'Workflows & Approvals',
    icon: '✅',
    route: '/workflows',
    description: 'Approval Rules, Notifications, Audit Log'
  },
  {
    id: 'security',
    label: 'Security & Permissions',
    icon: '🔒',
    route: '/security',
    description: 'Role-based Access, Field Controls'
  }
]
