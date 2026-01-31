<template>
  <div class="table-wrapper">
    <div class="table-container">
      <table class="base-table">
        <thead>
          <tr>
            <th
              v-for="column in columns"
              :key="column.key"
              @click="handleSort(column)"
              :class="{ sortable: column.sortable, [column.align || 'left']: true }"
            >
              <div class="th-content">
                {{ column.label }}
                <span v-if="column.sortable" class="sort-icon">
                  <span v-if="sortKey === column.key">
                    {{ sortOrder === 'asc' ? '▲' : '▼' }}
                  </span>
                  <span v-else class="sort-inactive">⇅</span>
                </span>
              </div>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td :colspan="columns.length" class="loading-cell">
              <div class="loading-spinner"></div>
              Loading...
            </td>
          </tr>
          <tr v-else-if="sortedData.length === 0">
            <td :colspan="columns.length" class="empty-cell">
              {{ emptyText }}
            </td>
          </tr>
          <tr
            v-else
            v-for="(row, index) in sortedData"
            :key="getRowKey(row, index)"
            @click="handleRowClick(row)"
            :class="{ clickable: clickable }"
          >
            <td
              v-for="column in columns"
              :key="column.key"
              :class="[column.align || 'left']"
            >
              <slot :name="`cell-${column.key}`" :row="row" :value="row[column.key]">
                {{ row[column.key] }}
              </slot>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="pagination && totalPages > 1" class="table-pagination">
      <button
        class="pagination-btn"
        :disabled="currentPage === 1"
        @click="changePage(currentPage - 1)"
      >
        Previous
      </button>
      <span class="pagination-info">
        Page {{ currentPage }} of {{ totalPages }}
      </span>
      <button
        class="pagination-btn"
        :disabled="currentPage === totalPages"
        @click="changePage(currentPage + 1)"
      >
        Next
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

export interface TableColumn {
  key: string
  label: string
  sortable?: boolean
  align?: 'left' | 'center' | 'right'
}

interface Props {
  columns: TableColumn[]
  data: any[]
  rowKey?: string
  loading?: boolean
  emptyText?: string
  clickable?: boolean
  pagination?: boolean
  pageSize?: number
}

const props = withDefaults(defineProps<Props>(), {
  rowKey: 'id',
  loading: false,
  emptyText: 'No data available',
  clickable: false,
  pagination: false,
  pageSize: 10
})

const emit = defineEmits<{
  'row-click': [row: any]
}>()

const sortKey = ref<string>('')
const sortOrder = ref<'asc' | 'desc'>('asc')
const currentPage = ref(1)

const sortedData = computed(() => {
  let data = [...props.data]

  // Apply sorting
  if (sortKey.value) {
    data.sort((a, b) => {
      const aVal = a[sortKey.value]
      const bVal = b[sortKey.value]

      if (aVal === bVal) return 0

      const comparison = aVal > bVal ? 1 : -1
      return sortOrder.value === 'asc' ? comparison : -comparison
    })
  }

  // Apply pagination
  if (props.pagination) {
    const start = (currentPage.value - 1) * props.pageSize
    const end = start + props.pageSize
    data = data.slice(start, end)
  }

  return data
})

const totalPages = computed(() => {
  return Math.ceil(props.data.length / props.pageSize)
})

const handleSort = (column: TableColumn) => {
  if (!column.sortable) return

  if (sortKey.value === column.key) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = column.key
    sortOrder.value = 'asc'
  }
}

const handleRowClick = (row: any) => {
  if (props.clickable) {
    emit('row-click', row)
  }
}

const getRowKey = (row: any, index: number) => {
  return row[props.rowKey] || index
}

const changePage = (page: number) => {
  currentPage.value = page
}
</script>

<style scoped>
.table-wrapper {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.table-container {
  overflow-x: auto;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.base-table {
  width: 100%;
  border-collapse: collapse;
  background: #fff;
}

.base-table thead {
  background: #f8fafc;
}

.base-table th {
  padding: 1rem;
  text-align: left;
  font-weight: 600;
  color: #475569;
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid #e2e8f0;
}

.base-table th.sortable {
  cursor: pointer;
  user-select: none;
}

.base-table th.sortable:hover {
  background: #f1f5f9;
}

.th-content {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.sort-icon {
  font-size: 0.7rem;
  color: #3b82f6;
}

.sort-inactive {
  color: #cbd5e1;
}

.base-table td {
  padding: 1rem;
  border-bottom: 1px solid #e2e8f0;
  color: #1e293b;
}

.base-table tbody tr:hover {
  background: #f8fafc;
}

.base-table tbody tr.clickable {
  cursor: pointer;
}

.base-table .left {
  text-align: left;
}

.base-table .center {
  text-align: center;
}

.base-table .right {
  text-align: right;
}

.loading-cell,
.empty-cell {
  text-align: center;
  padding: 3rem 1rem !important;
  color: #64748b;
}

.loading-spinner {
  width: 2rem;
  height: 2rem;
  border: 3px solid #e2e8f0;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 1rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.table-pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
}

.pagination-btn {
  padding: 0.5rem 1rem;
  border: 1px solid #e2e8f0;
  background: #fff;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.875rem;
  font-weight: 500;
  color: #475569;
}

.pagination-btn:hover:not(:disabled) {
  background: #f8fafc;
  border-color: #3b82f6;
  color: #3b82f6;
}

.pagination-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.pagination-info {
  font-size: 0.875rem;
  color: #64748b;
}
</style>
