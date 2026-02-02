import { ref, computed } from "vue";
import { defineStore } from "pinia";
import type { User, UserRole } from "@/types";

export const useUsersStore = defineStore("users", () => {
  const users = ref<User[]>([
    {
      id: "USR-001",
      email: "ahmed.manager@company.com",
      firstName: "Ahmed",
      lastName: "Al-Rashid",
      fullName: "Ahmed Al-Rashid",
      phone: "+966 50 111 2222",
      role: "presales_manager",
      department: "Presales",
      region: "Riyadh",
      status: "active",
      lastLogin: "2024-01-25T09:30:00Z",
      createdAt: "2023-01-01T00:00:00Z",
    },
    {
      id: "USR-002",
      email: "sara.sales@company.com",
      firstName: "Sara",
      lastName: "Al-Mutairi",
      fullName: "Sara Al-Mutairi",
      phone: "+966 50 222 3333",
      role: "sales_executive",
      department: "Sales",
      region: "Jeddah",
      status: "active",
      lastLogin: "2024-01-25T10:15:00Z",
      createdAt: "2023-02-15T00:00:00Z",
    },
    {
      id: "USR-003",
      email: "mohammed.presales@company.com",
      firstName: "Mohammed",
      lastName: "Al-Qahtani",
      fullName: "Mohammed Al-Qahtani",
      phone: "+966 50 333 4444",
      role: "presales",
      department: "Presales",
      region: "Riyadh",
      status: "active",
      lastLogin: "2024-01-25T08:45:00Z",
      createdAt: "2023-03-10T00:00:00Z",
    },
    {
      id: "USR-004",
      email: "fatima.finance@company.com",
      firstName: "Fatima",
      lastName: "Al-Dosari",
      fullName: "Fatima Al-Dosari",
      phone: "+966 50 444 5555",
      role: "finance_reviewer",
      department: "Finance",
      status: "active",
      lastLogin: "2024-01-24T16:30:00Z",
      createdAt: "2023-01-15T00:00:00Z",
    },
    {
      id: "USR-005",
      email: "admin@company.com",
      firstName: "System",
      lastName: "Administrator",
      fullName: "System Administrator",
      phone: "+966 50 555 6666",
      role: "admin",
      department: "IT",
      status: "active",
      lastLogin: "2024-01-25T07:00:00Z",
      createdAt: "2023-01-01T00:00:00Z",
    },
    {
      id: "USR-006",
      email: "khalid.warehouse@company.com",
      firstName: "Khalid",
      lastName: "Al-Harbi",
      fullName: "Khalid Al-Harbi",
      phone: "+966 50 666 7777",
      role: "presales",
      department: "Warehouse",
      region: "Riyadh",
      status: "active",
      lastLogin: "2024-01-25T06:30:00Z",
      createdAt: "2023-04-01T00:00:00Z",
    },
  ]);

  const currentUser = ref<User | null>(users.value[0]); // Default to first user for demo

  // Computed
  const activeUsers = computed(() => {
    return users.value.filter((u) => u.status === "active");
  });

  const usersByRole = computed(() => {
    const byRole: Record<UserRole, User[]> = {
      admin: [],
      presales_manager: [],
      sales_executive: [],
      presales: [],
      finance_reviewer: [],
    };

    users.value.forEach((user) => {
      byRole[user.role].push(user);
    });

    return byRole;
  });

  const salesExecutives = computed(() => {
    return users.value.filter(
      (u) => u.role === "sales_executive" && u.status === "active",
    );
  });

  const presalesTeam = computed(() => {
    return users.value.filter(
      (u) =>
        (u.role === "presales" || u.role === "presales_manager") &&
        u.status === "active",
    );
  });

  const isAdmin = computed(() => {
    return currentUser.value?.role === "admin";
  });

  const canApproveQuotes = computed(() => {
    return (
      currentUser.value?.role === "presales_manager" ||
      currentUser.value?.role === "finance_reviewer" ||
      currentUser.value?.role === "admin"
    );
  });

  const canManageProducts = computed(() => {
    return (
      currentUser.value?.role === "admin" ||
      currentUser.value?.role === "presales_manager" ||
      currentUser.value?.role === "presales"
    );
  });

  // Actions
  function addUser(user: User) {
    users.value.push(user);
  }

  function updateUser(id: string, updates: Partial<User>) {
    const index = users.value.findIndex((u) => u.id === id);
    if (index !== -1) {
      users.value[index] = { ...users.value[index], ...updates };
    }
  }

  function deleteUser(id: string) {
    const index = users.value.findIndex((u) => u.id === id);
    if (index !== -1) {
      users.value.splice(index, 1);
    }
  }

  function getUserById(id: string) {
    return users.value.find((u) => u.id === id);
  }

  function getUsersByDepartment(department: string) {
    return users.value.filter((u) => u.department === department);
  }

  function getUsersByRegion(region: string) {
    return users.value.filter((u) => u.region === region);
  }

  function setCurrentUser(user: User) {
    currentUser.value = user;
  }

  function updateLastLogin(id: string) {
    updateUser(id, { lastLogin: new Date().toISOString() });
  }

  function hasPermission(permission: string): boolean {
    if (!currentUser.value) return false;

    const rolePermissions: Record<UserRole, string[]> = {
      admin: ["all"],
      presales_manager: [
        "view_dashboard",
        "manage_opportunities",
        "manage_quotes",
        "approve_quotes",
        "manage_products",
        "view_customers",
        "manage_pricing",
        "view_reports",
      ],
      sales_executive: [
        "view_dashboard",
        "manage_opportunities",
        "create_quotes",
        "view_customers",
        "view_products",
      ],
      presales: [
        "view_dashboard",
        "view_opportunities",
        "create_quotes",
        "manage_products",
        "view_customers",
        "manage_pricing",
      ],
      finance_reviewer: [
        "view_dashboard",
        "view_quotes",
        "approve_quotes",
        "view_reports",
        "view_pricing_history",
      ],
    };

    const userPermissions = rolePermissions[currentUser.value.role];
    return (
      userPermissions.includes("all") || userPermissions.includes(permission)
    );
  }

  return {
    users,
    currentUser,
    activeUsers,
    usersByRole,
    salesExecutives,
    presalesTeam,
    isAdmin,
    canApproveQuotes,
    canManageProducts,
    addUser,
    updateUser,
    deleteUser,
    getUserById,
    getUsersByDepartment,
    getUsersByRegion,
    setCurrentUser,
    updateLastLogin,
    hasPermission,
  };
});
