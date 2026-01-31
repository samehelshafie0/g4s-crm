<template>
  <div class="procurement-view">
    <!-- Stats Bar -->
    <div class="stats-bar">
      <div class="stat-card">
        <div class="stat-label">Total Orders</div>
        <div class="stat-value">{{ poStore.orderStats.total }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Pending Orders</div>
        <div class="stat-value warning">{{ poStore.orderStats.pending }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Partial Received</div>
        <div class="stat-value info">{{ poStore.orderStats.partial }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Total Value</div>
        <div class="stat-value">{{ formatCurrency(poStore.orderStats.totalValue) }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Pending Value</div>
        <div class="stat-value warning">{{ formatCurrency(poStore.orderStats.pendingValue) }}</div>
      </div>
    </div>

    <!-- Header -->
    <div class="view-header">
      <div class="header-actions">
        <input
          v-model="searchQuery"
          type="search"
          placeholder="Search purchase orders..."
          class="search-input"
        />
        <button class="btn btn-primary" @click="showCreatePOModal = true">
          New Purchase Order
        </button>
      </div>
    </div>

    <!-- Tabs -->
    <div class="tabs">
      <button
        class="tab"
        :class="{ active: activeTab === 'all' }"
        @click="activeTab = 'all'"
      >
        All Orders
      </button>
      <button
        class="tab"
        :class="{ active: activeTab === 'pending' }"
        @click="activeTab = 'pending'"
      >
        Pending ({{ poStore.pendingOrders.length }})
      </button>
      <button
        class="tab"
        :class="{ active: activeTab === 'partial' }"
        @click="activeTab = 'partial'"
      >
        Partial ({{ poStore.partialOrders.length }})
      </button>
      <button
        class="tab"
        :class="{ active: activeTab === 'upcoming' }"
        @click="activeTab = 'upcoming'"
      >
        Upcoming Deliveries ({{ poStore.upcomingDeliveries.length }})
      </button>
      <button
        class="tab"
        :class="{ active: activeTab === 'overdue' }"
        @click="activeTab = 'overdue'"
      >
        Overdue ({{ poStore.overdueOrders.length }})
      </button>
    </div>

    <!-- Filters -->
    <div class="filters-bar">
      <div class="filter-group">
        <label>Project:</label>
        <select v-model="projectFilter" class="filter-select">
          <option value="all">All Projects</option>
          <option value="no_project">No Project</option>
          <option v-for="project in wonProjects" :key="project.id" :value="project.opportunityNumber">
            {{ project.opportunityNumber }} - {{ project.name }}
          </option>
        </select>
      </div>

      <div class="filter-group">
        <label>Vendor:</label>
        <select v-model="vendorFilter" class="filter-select">
          <option value="all">All Vendors</option>
          <option v-for="vendor in vendorsStore.activeVendors" :key="vendor.id" :value="vendor.id">
            {{ vendor.name }}
          </option>
        </select>
      </div>

      <div class="filter-group">
        <label>Status:</label>
        <select v-model="statusFilter" class="filter-select">
          <option value="all">All Status</option>
          <option value="draft">Draft</option>
          <option value="pending_approval">Pending Approval</option>
          <option value="approved">Approved</option>
          <option value="sent">Sent</option>
          <option value="partial">Partial</option>
          <option value="received">Received</option>
          <option value="cancelled">Cancelled</option>
        </select>
      </div>

      <div class="filter-group">
        <label>Warehouse:</label>
        <select v-model="warehouseFilter" class="filter-select">
          <option value="all">All Warehouses</option>
          <option value="Riyadh Main">Riyadh Main</option>
          <option value="Jeddah Branch">Jeddah Branch</option>
          <option value="Dammam Branch">Dammam Branch</option>
        </select>
      </div>
    </div>

    <!-- Orders Table -->
    <BaseCard no-padding>
      <div class="table-container">
        <table class="orders-table">
          <thead>
            <tr>
              <th>PO Number</th>
              <th>Vendor</th>
              <th>Order Date</th>
              <th>Expected Delivery</th>
              <th>Warehouse</th>
              <th class="text-right">Total Amount</th>
              <th class="text-right">Received %</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="order in filteredOrders" :key="order.id" class="order-row">
              <td>
                <div class="po-number-cell">
                  <div class="po-number">{{ order.poNumber }}</div>
                  <div class="po-items">{{ order.items.length }} items</div>
                </div>
              </td>
              <td>
                <div class="vendor-cell">
                  <div class="vendor-name">{{ order.vendorName }}</div>
                  <div class="vendor-code">{{ order.vendorCode }}</div>
                </div>
              </td>
              <td>{{ formatDate(order.orderDate) }}</td>
              <td>
                <div class="delivery-cell">
                  <div>{{ formatDate(order.expectedDeliveryDate) }}</div>
                  <div v-if="isOverdue(order)" class="overdue-badge">Overdue!</div>
                </div>
              </td>
              <td>
                <BaseBadge variant="info" size="sm">{{ order.deliveryWarehouse }}</BaseBadge>
              </td>
              <td class="text-right">{{ formatCurrency(order.totalAmount) }}</td>
              <td class="text-right">
                <div class="progress-cell">
                  <div class="progress-bar-container">
                    <div class="progress-bar" :style="{ width: order.receivedPercentage + '%' }"></div>
                  </div>
                  <span class="progress-text">{{ order.receivedPercentage }}%</span>
                </div>
              </td>
              <td>
                <BaseBadge :variant="getStatusVariant(order.status)" size="sm">
                  {{ formatStatus(order.status) }}
                </BaseBadge>
              </td>
              <td>
                <div class="action-buttons">
                  <button class="action-btn" title="View Details" @click="viewOrderDetails(order)">👁️</button>
                  <button
                    v-if="order.status === 'draft' || order.status === 'pending_approval'"
                    class="action-btn"
                    title="Edit"
                    @click="editOrder(order)"
                  >✏️</button>
                  <button
                    v-if="order.status === 'pending_approval'"
                    class="action-btn"
                    title="Approve"
                    @click="approveOrder(order)"
                  >✓</button>
                  <button
                    v-if="order.status === 'approved' || order.status === 'sent' || order.status === 'partial'"
                    class="action-btn"
                    title="Receive Goods"
                    @click="receiveGoods(order)"
                  >📦</button>
                </div>
              </td>
            </tr>
            <tr v-if="filteredOrders.length === 0">
              <td colspan="9" class="empty-row">
                No purchase orders found matching your criteria
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </BaseCard>

    <!-- Create PO Modal -->
    <BaseModal
      v-model="showCreatePOModal"
      title="Create Purchase Order"
      size="xl"
      @confirm="savePurchaseOrder"
    >
      <div class="po-form">
        <!-- Project Selection -->
        <div class="form-section">
          <h3>Project Information</h3>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Link to Won Project</label>
              <select v-model="poForm.projectId" class="form-input" @change="onProjectSelect">
                <option value="">-- No Project (Direct Purchase) --</option>
                <option v-for="project in wonProjects" :key="project.id" :value="project.id">
                  {{ project.opportunityNumber }} - {{ project.name }} ({{ formatCurrency(project.estimatedValue) }})
                </option>
              </select>
              <small v-if="selectedProject" class="form-hint success">
                Project: {{ selectedProject.name }} | Customer: {{ selectedProject.companyId }} | Value: {{ formatCurrency(selectedProject.estimatedValue) }}
              </small>
            </div>

            <div class="form-group">
              <label class="form-label">PO Reference Number</label>
              <input v-model="poForm.referenceNumber" type="text" class="form-input" placeholder="e.g., RFQ-2024-001" />
            </div>
          </div>
        </div>

        <!-- Vendor Selection -->
        <div class="form-section">
          <h3>Vendor Information</h3>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Select Vendor *</label>
              <select v-model="poForm.vendorId" class="form-input" required @change="onVendorChange">
                <option value="">-- Select Vendor --</option>
                <option v-for="vendor in vendorsStore.activeVendors" :key="vendor.id" :value="vendor.id">
                  {{ vendor.name }} ({{ vendor.defaultLeadTimeDays }} days lead time)
                </option>
              </select>
            </div>

            <div class="form-group">
              <label class="form-label">Delivery Warehouse *</label>
              <select v-model="poForm.deliveryWarehouse" class="form-input" required>
                <option value="Riyadh Main">Riyadh Main</option>
                <option value="Jeddah Branch">Jeddah Branch</option>
                <option value="Dammam Branch">Dammam Branch</option>
              </select>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Order Date *</label>
              <input v-model="poForm.orderDate" type="date" class="form-input" required />
            </div>

            <div class="form-group">
              <label class="form-label">Expected Delivery Date</label>
              <input v-model="poForm.expectedDeliveryDate" type="date" class="form-input" :disabled="!poForm.vendorId" />
              <small v-if="selectedVendor" class="form-hint">
                Auto-calculated based on {{ selectedVendor.defaultLeadTimeDays }} days lead time
              </small>
            </div>
          </div>
        </div>

        <!-- Line Items -->
        <div class="form-section">
          <div class="section-header">
            <h3>Line Items</h3>
            <button type="button" class="btn btn-small btn-primary" @click="addLineItem">
              Add Item
            </button>
          </div>

          <div class="line-items-container">
            <div v-for="(item, index) in poForm.items" :key="index" class="line-item">
              <div class="line-item-header">
                <span class="line-item-number">#{{ index + 1 }}</span>
                <button type="button" class="btn-remove" @click="removeLineItem(index)">×</button>
              </div>

              <div class="form-row">
                <div class="form-group" style="flex: 2">
                  <label class="form-label">Product *</label>
                  <select v-model="item.productId" class="form-input" @change="onProductSelect(index)" required>
                    <option value="">-- Select Product --</option>
                    <option v-for="product in productsStore.getActiveProducts()" :key="product.id" :value="product.id">
                      {{ product.sku }} - {{ product.name }} ({{ formatCurrency(product.landedCostSAR) }})
                    </option>
                  </select>
                </div>

                <div class="form-group">
                  <label class="form-label">Quantity *</label>
                  <input v-model.number="item.quantity" type="number" class="form-input" min="1" required @input="calculateLineTotal(index)" />
                </div>

                <div class="form-group">
                  <label class="form-label">Unit Price *</label>
                  <input v-model.number="item.unitPrice" type="number" class="form-input" min="0" step="0.01" required @input="calculateLineTotal(index)" />
                </div>

                <div class="form-group">
                  <label class="form-label">Discount %</label>
                  <input v-model.number="item.discount" type="number" class="form-input" min="0" max="100" step="0.1" @input="calculateLineTotal(index)" />
                </div>

                <div class="form-group">
                  <label class="form-label">Total</label>
                  <input :value="formatCurrency(item.total)" class="form-input" readonly />
                </div>
              </div>

              <div class="form-row">
                <div class="form-group full-width">
                  <label class="form-label">Description</label>
                  <input v-model="item.description" type="text" class="form-input" />
                </div>
              </div>
            </div>

            <div v-if="poForm.items.length === 0" class="empty-items">
              No items added yet. Click "Add Item" to add products to this PO.
            </div>
          </div>
        </div>

        <!-- Financial Summary -->
        <div class="form-section">
          <h3>Financial Summary</h3>
          <div class="financial-grid">
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Freight Cost</label>
                <input v-model.number="poForm.freightCost" type="number" class="form-input" min="0" step="0.01" @input="calculateTotals" />
              </div>

              <div class="form-group">
                <label class="form-label">Other Charges</label>
                <input v-model.number="poForm.otherCharges" type="number" class="form-input" min="0" step="0.01" @input="calculateTotals" />
              </div>
            </div>

            <div class="summary-box">
              <div class="summary-row">
                <span>Subtotal:</span>
                <span>{{ formatCurrency(poTotals.subtotal) }}</span>
              </div>
              <div class="summary-row">
                <span>Discount:</span>
                <span>{{ formatCurrency(poTotals.discountAmount) }}</span>
              </div>
              <div class="summary-row">
                <span>Tax (15%):</span>
                <span>{{ formatCurrency(poTotals.taxAmount) }}</span>
              </div>
              <div class="summary-row">
                <span>Freight:</span>
                <span>{{ formatCurrency(poForm.freightCost) }}</span>
              </div>
              <div class="summary-row">
                <span>Other Charges:</span>
                <span>{{ formatCurrency(poForm.otherCharges) }}</span>
              </div>
              <div class="summary-row total-row">
                <span>Total Amount:</span>
                <span>{{ formatCurrency(poTotals.totalAmount) }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Notes -->
        <div class="form-section">
          <h3>Additional Information</h3>
          <div class="form-group">
            <label class="form-label">Notes</label>
            <textarea v-model="poForm.notes" class="form-input" rows="3" placeholder="Any special instructions or notes..."></textarea>
          </div>
        </div>
      </div>
    </BaseModal>

    <!-- Order Details Modal -->
    <BaseModal
      v-if="selectedOrder"
      v-model="showDetailsModal"
      :title="`Purchase Order ${selectedOrder.poNumber}`"
      size="xl"
      hide-confirm
      cancel-text="Close"
    >
      <div class="order-details">
        <!-- Header Info -->
        <div class="details-header">
          <div v-if="selectedOrder.projectReference" class="project-banner">
            <span class="project-icon">📋</span>
            <div class="project-info-banner">
              <span class="project-label">Linked to Project:</span>
              <span class="project-name">{{ selectedOrder.projectReference }}</span>
            </div>
          </div>

          <div class="details-row">
            <div class="detail-item">
              <span class="detail-label">PO Number:</span>
              <span class="detail-value">{{ selectedOrder.poNumber }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Status:</span>
              <BaseBadge :variant="getStatusVariant(selectedOrder.status)" size="sm">
                {{ formatStatus(selectedOrder.status) }}
              </BaseBadge>
            </div>
            <div class="detail-item">
              <span class="detail-label">Vendor:</span>
              <span class="detail-value">{{ selectedOrder.vendorName }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Order Date:</span>
              <span class="detail-value">{{ formatDate(selectedOrder.orderDate) }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Expected Delivery:</span>
              <span class="detail-value">{{ formatDate(selectedOrder.expectedDeliveryDate) }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Warehouse:</span>
              <span class="detail-value">{{ selectedOrder.deliveryWarehouse }}</span>
            </div>
          </div>
        </div>

        <!-- Line Items -->
        <div class="details-section">
          <h3>Line Items</h3>
          <table class="details-table">
            <thead>
              <tr>
                <th>#</th>
                <th>Product</th>
                <th>Description</th>
                <th class="text-right">Qty</th>
                <th class="text-right">Unit Price</th>
                <th class="text-right">Total</th>
                <th class="text-right">Received</th>
                <th class="text-right">Remaining</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(item, index) in selectedOrder.items" :key="item.id">
                <td>{{ index + 1 }}</td>
                <td>
                  <div class="product-cell">
                    <div>{{ item.productName }}</div>
                    <div class="product-sku">{{ item.productSku }}</div>
                  </div>
                </td>
                <td>{{ item.description }}</td>
                <td class="text-right">{{ item.quantity }}</td>
                <td class="text-right">{{ formatCurrency(item.unitPrice) }}</td>
                <td class="text-right">{{ formatCurrency(item.total) }}</td>
                <td class="text-right">
                  <span class="received-qty">{{ item.receivedQuantity }}</span>
                </td>
                <td class="text-right">
                  <span class="remaining-qty">{{ item.remainingQuantity }}</span>
                </td>
                <td>
                  <BaseBadge :variant="getItemStatusVariant(item.status)" size="sm">
                    {{ formatItemStatus(item.status) }}
                  </BaseBadge>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Financial Summary -->
        <div class="details-section">
          <div class="financial-summary">
            <div class="summary-row">
              <span>Subtotal:</span>
              <span>{{ formatCurrency(selectedOrder.subtotal) }}</span>
            </div>
            <div class="summary-row">
              <span>Discount:</span>
              <span>{{ formatCurrency(selectedOrder.discountAmount) }}</span>
            </div>
            <div class="summary-row">
              <span>Tax:</span>
              <span>{{ formatCurrency(selectedOrder.taxAmount) }}</span>
            </div>
            <div class="summary-row">
              <span>Freight:</span>
              <span>{{ formatCurrency(selectedOrder.freightCost) }}</span>
            </div>
            <div class="summary-row">
              <span>Other Charges:</span>
              <span>{{ formatCurrency(selectedOrder.otherCharges) }}</span>
            </div>
            <div class="summary-row total-row">
              <span>Total Amount:</span>
              <span>{{ formatCurrency(selectedOrder.totalAmount) }}</span>
            </div>
            <div class="summary-row received-row">
              <span>Received Value:</span>
              <span>{{ formatCurrency(selectedOrder.receivedValue) }} ({{ selectedOrder.receivedPercentage }}%)</span>
            </div>
          </div>
        </div>

        <!-- Goods Receipts History -->
        <div v-if="orderReceipts.length > 0" class="details-section">
          <h3>Goods Receipts History</h3>
          <div class="receipts-list">
            <div v-for="receipt in orderReceipts" :key="receipt.id" class="receipt-card">
              <div class="receipt-header">
                <span class="receipt-number">{{ receipt.grNumber }}</span>
                <BaseBadge :variant="receipt.status === 'completed' ? 'success' : 'warning'" size="sm">
                  {{ receipt.status }}
                </BaseBadge>
              </div>
              <div class="receipt-info">
                <span>Date: {{ formatDate(receipt.receiptDate) }}</span>
                <span>Type: {{ receipt.receiptType }}</span>
                <span>Value: {{ formatCurrency(receipt.totalValue) }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Notes -->
        <div v-if="selectedOrder.notes" class="details-section">
          <h3>Notes</h3>
          <div class="notes-box">{{ selectedOrder.notes }}</div>
        </div>
      </div>
    </BaseModal>

    <!-- Receive Goods Modal -->
    <BaseModal
      v-if="receivingOrder"
      v-model="showReceiveModal"
      :title="`Receive Goods - ${receivingOrder.poNumber}`"
      size="xl"
      @confirm="saveGoodsReceipt"
    >
      <div class="receive-form">
        <div class="form-section">
          <h3>Receipt Information</h3>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Receipt Date *</label>
              <input v-model="receiptForm.receiptDate" type="date" class="form-input" required />
            </div>

            <div class="form-group">
              <label class="form-label">Receipt Type *</label>
              <select v-model="receiptForm.receiptType" class="form-input" required>
                <option value="partial">Partial Receipt</option>
                <option value="full">Full Receipt</option>
                <option value="final">Final Receipt</option>
              </select>
            </div>

            <div class="form-group">
              <label class="form-label">Delivery Note Number</label>
              <input v-model="receiptForm.deliveryNoteNumber" type="text" class="form-input" />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Carrier Name</label>
              <input v-model="receiptForm.carrierName" type="text" class="form-input" />
            </div>

            <div class="form-group">
              <label class="form-label">Tracking Number</label>
              <input v-model="receiptForm.trackingNumber" type="text" class="form-input" />
            </div>
          </div>
        </div>

        <div class="form-section">
          <h3>Items to Receive</h3>
          <div class="receive-items">
            <div v-for="(item, index) in receiptForm.items" :key="item.poLineItemId" class="receive-item">
              <div class="receive-item-header">
                <div class="product-info">
                  <div class="product-name">{{ item.productName }}</div>
                  <div class="product-sku">{{ item.productSku }}</div>
                </div>
                <div class="ordered-info">
                  Ordered: <strong>{{ item.orderedQuantity }}</strong> |
                  Already Received: <strong>{{ getAlreadyReceived(item.poLineItemId) }}</strong> |
                  Remaining: <strong>{{ item.orderedQuantity - getAlreadyReceived(item.poLineItemId) }}</strong>
                </div>
              </div>

              <div class="form-row">
                <div class="form-group">
                  <label class="form-label">Received Quantity *</label>
                  <input
                    v-model.number="item.receivedQuantity"
                    type="number"
                    class="form-input"
                    min="0"
                    :max="item.orderedQuantity - getAlreadyReceived(item.poLineItemId)"
                    @input="calculateReceivedTotals(index)"
                  />
                </div>

                <div class="form-group">
                  <label class="form-label">Accepted Quantity *</label>
                  <input
                    v-model.number="item.acceptedQuantity"
                    type="number"
                    class="form-input"
                    min="0"
                    :max="item.receivedQuantity"
                  />
                </div>

                <div class="form-group">
                  <label class="form-label">Rejected Quantity</label>
                  <input
                    v-model.number="item.rejectedQuantity"
                    type="number"
                    class="form-input"
                    min="0"
                    readonly
                  />
                </div>

                <div class="form-group">
                  <label class="form-label">Batch Number</label>
                  <input v-model="item.batchNumber" type="text" class="form-input" />
                </div>
              </div>

              <div v-if="item.rejectedQuantity > 0" class="form-row">
                <div class="form-group full-width">
                  <label class="form-label">Rejection Reason *</label>
                  <input v-model="item.rejectionReason" type="text" class="form-input" required />
                </div>
              </div>

              <div class="form-row">
                <div class="form-group full-width">
                  <label class="form-label">Notes</label>
                  <input v-model="item.notes" type="text" class="form-input" placeholder="Any notes about this item..." />
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="form-section">
          <h3>Quality Control</h3>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Quality Check Status</label>
              <select v-model="receiptForm.qualityCheckStatus" class="form-input">
                <option value="pending">Pending</option>
                <option value="passed">Passed</option>
                <option value="failed">Failed</option>
                <option value="partial">Partial</option>
              </select>
            </div>

            <div class="form-group" style="flex: 2">
              <label class="form-label">Quality Notes</label>
              <input v-model="receiptForm.qualityNotes" type="text" class="form-input" />
            </div>
          </div>
        </div>

        <div class="form-section">
          <h3>General Notes</h3>
          <div class="form-group">
            <textarea v-model="receiptForm.notes" class="form-input" rows="3" placeholder="Any general notes or observations..."></textarea>
          </div>
        </div>
      </div>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { usePurchaseOrdersStore } from '@/stores/purchaseOrders'
import { useVendorsStore } from '@/stores/vendors'
import { useProductsStore } from '@/stores/products'
import { useGoodsReceiptsStore } from '@/stores/goodsReceipts'
import { useWarehouseStockStore } from '@/stores/warehouseStock'
import { useStockMovementsStore } from '@/stores/stockMovements'
import { useOpportunitiesStore } from '@/stores/opportunities'
import { useToast } from '@/composables/useToast'
import BaseCard from '@/components/UI/BaseCard.vue'
import BaseBadge from '@/components/UI/BaseBadge.vue'
import BaseModal from '@/components/UI/BaseModal.vue'

const poStore = usePurchaseOrdersStore()
const vendorsStore = useVendorsStore()
const productsStore = useProductsStore()
const receiptsStore = useGoodsReceiptsStore()
const warehouseStore = useWarehouseStockStore()
const movementsStore = useStockMovementsStore()
const opportunitiesStore = useOpportunitiesStore()
const { success, info } = useToast()

const searchQuery = ref('')
const activeTab = ref('all')
const vendorFilter = ref('all')
const statusFilter = ref('all')
const warehouseFilter = ref('all')
const projectFilter = ref('all')

const showCreatePOModal = ref(false)
const showDetailsModal = ref(false)
const showReceiveModal = ref(false)
const selectedOrder = ref<any>(null)
const receivingOrder = ref<any>(null)

// Won projects (opportunities that are closed_won)
const wonProjects = computed(() => opportunitiesStore.wonOpportunities)

// PO Form
const poForm = ref({
  vendorId: '',
  orderDate: new Date().toISOString().split('T')[0] || '',
  expectedDeliveryDate: '',
  deliveryWarehouse: 'Riyadh Main',
  items: [] as any[],
  freightCost: 0,
  otherCharges: 0,
  referenceNumber: '',
  projectReference: '',
  projectId: '',
  notes: ''
})

// Receipt Form
const receiptForm = ref({
  receiptDate: new Date().toISOString().split('T')[0] || '',
  receiptType: 'partial' as 'full' | 'partial' | 'final',
  deliveryNoteNumber: '',
  carrierName: '',
  trackingNumber: '',
  items: [] as any[],
  qualityCheckStatus: 'pending' as 'pending' | 'passed' | 'failed' | 'partial',
  qualityNotes: '',
  notes: ''
})

// Selected vendor
const selectedVendor = computed(() => {
  if (!poForm.value.vendorId) return null
  return vendorsStore.getVendorById(poForm.value.vendorId)
})

// Selected project
const selectedProject = computed(() => {
  if (!poForm.value.projectId) return null
  return opportunitiesStore.getOpportunityById(poForm.value.projectId)
})

// Calculate PO totals
const poTotals = computed(() => {
  const subtotal = poForm.value.items.reduce((sum, item) => sum + (item.total || 0), 0)
  const discountAmount = poForm.value.items.reduce((sum, item) => {
    const itemSubtotal = item.quantity * item.unitPrice
    return sum + (itemSubtotal * (item.discount || 0) / 100)
  }, 0)
  const taxableAmount = subtotal - discountAmount
  const taxAmount = taxableAmount * 0.15
  const totalAmount = taxableAmount + taxAmount + poForm.value.freightCost + poForm.value.otherCharges

  return {
    subtotal,
    discountAmount,
    taxAmount,
    totalAmount
  }
})

// Filtered orders
const filteredOrders = computed(() => {
  let orders = poStore.orders

  // Filter by tab
  if (activeTab.value === 'pending') {
    orders = poStore.pendingOrders
  } else if (activeTab.value === 'partial') {
    orders = poStore.partialOrders
  } else if (activeTab.value === 'upcoming') {
    orders = poStore.upcomingDeliveries
  } else if (activeTab.value === 'overdue') {
    orders = poStore.overdueOrders
  }

  // Filter by vendor
  if (vendorFilter.value !== 'all') {
    orders = orders.filter(o => o.vendorId === vendorFilter.value)
  }

  // Filter by status
  if (statusFilter.value !== 'all') {
    orders = orders.filter(o => o.status === statusFilter.value)
  }

  // Filter by warehouse
  if (warehouseFilter.value !== 'all') {
    orders = orders.filter(o => o.deliveryWarehouse === warehouseFilter.value)
  }

  // Filter by project
  if (projectFilter.value !== 'all') {
    if (projectFilter.value === 'no_project') {
      orders = orders.filter(o => !o.projectReference)
    } else {
      orders = orders.filter(o => o.projectReference === projectFilter.value)
    }
  }

  // Search
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    orders = orders.filter(o =>
      o.poNumber.toLowerCase().includes(query) ||
      o.vendorName.toLowerCase().includes(query) ||
      o.items.some(i => i.productName.toLowerCase().includes(query))
    )
  }

  return orders
})

// Order receipts
const orderReceipts = computed(() => {
  if (!selectedOrder.value) return []
  return receiptsStore.getReceiptsByPO(selectedOrder.value.id)
})

// Watch vendor change to calculate delivery date
watch(() => poForm.value.vendorId, (vendorId) => {
  if (vendorId && selectedVendor.value && poForm.value.orderDate) {
    const orderDate = new Date(poForm.value.orderDate)
    orderDate.setDate(orderDate.getDate() + selectedVendor.value.defaultLeadTimeDays)
    const dateStr = orderDate.toISOString().split('T')[0]
    if (dateStr) poForm.value.expectedDeliveryDate = dateStr
  }
})

// Watch order date to recalculate delivery date
watch(() => poForm.value.orderDate, (orderDate) => {
  if (selectedVendor.value && orderDate) {
    const date = new Date(orderDate)
    date.setDate(date.getDate() + selectedVendor.value.defaultLeadTimeDays)
    const dateStr = date.toISOString().split('T')[0]
    if (dateStr) poForm.value.expectedDeliveryDate = dateStr
  }
})

// Functions
function formatCurrency(amount: number) {
  return new Intl.NumberFormat('en-SA', {
    style: 'currency',
    currency: 'SAR',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(amount)
}

function formatDate(dateString: string) {
  return new Date(dateString).toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric'
  })
}

function formatStatus(status: string) {
  const statuses: Record<string, string> = {
    draft: 'Draft',
    pending_approval: 'Pending Approval',
    approved: 'Approved',
    sent: 'Sent',
    partial: 'Partial',
    received: 'Received',
    cancelled: 'Cancelled',
    closed: 'Closed'
  }
  return statuses[status] || status
}

function formatItemStatus(status: string) {
  const statuses: Record<string, string> = {
    pending: 'Pending',
    partial: 'Partial',
    received: 'Received',
    cancelled: 'Cancelled'
  }
  return statuses[status] || status
}

function getStatusVariant(status: string): 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const variants: Record<string, any> = {
    draft: 'default',
    pending_approval: 'warning',
    approved: 'info',
    sent: 'primary',
    partial: 'warning',
    received: 'success',
    cancelled: 'danger',
    closed: 'default'
  }
  return variants[status] || 'default'
}

function getItemStatusVariant(status: string): 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const variants: Record<string, any> = {
    pending: 'warning',
    partial: 'info',
    received: 'success',
    cancelled: 'danger'
  }
  return variants[status] || 'default'
}

function isOverdue(order: any) {
  if (order.status === 'received' || order.status === 'cancelled') return false
  return new Date(order.expectedDeliveryDate) < new Date()
}

function onVendorChange() {
  if (selectedVendor.value && poForm.value.orderDate) {
    const orderDate = new Date(poForm.value.orderDate)
    orderDate.setDate(orderDate.getDate() + selectedVendor.value.defaultLeadTimeDays)
    const dateStr = orderDate.toISOString().split('T')[0]
    if (dateStr) poForm.value.expectedDeliveryDate = dateStr
  }
}

function onProjectSelect() {
  if (selectedProject.value) {
    // Auto-populate project reference
    poForm.value.projectReference = selectedProject.value.opportunityNumber
    success(`Linked to project: ${selectedProject.value.name}`)
  } else {
    poForm.value.projectReference = ''
  }
}

function addLineItem() {
  poForm.value.items.push({
    productId: '',
    productSku: '',
    productName: '',
    description: '',
    quantity: 1,
    unitPrice: 0,
    discount: 0,
    subtotal: 0,
    taxPercent: 15,
    taxAmount: 0,
    total: 0,
    expectedDeliveryDate: poForm.value.expectedDeliveryDate,
    receivedQuantity: 0,
    remainingQuantity: 0,
    status: 'pending'
  })
}

function removeLineItem(index: number) {
  poForm.value.items.splice(index, 1)
  calculateTotals()
}

function onProductSelect(index: number) {
  const item = poForm.value.items[index]
  const product = productsStore.getProductById(item.productId)
  if (product) {
    item.productSku = product.sku
    item.productName = product.name
    item.description = product.description
    item.unitPrice = product.landedCostSAR
    calculateLineTotal(index)
  }
}

function calculateLineTotal(index: number) {
  const item = poForm.value.items[index]
  const subtotal = item.quantity * item.unitPrice
  const discountAmount = subtotal * (item.discount || 0) / 100
  item.subtotal = subtotal - discountAmount
  item.taxAmount = item.subtotal * 0.15
  item.total = item.subtotal + item.taxAmount
  item.remainingQuantity = item.quantity
  calculateTotals()
}

function calculateTotals() {
  // Totals are computed in poTotals
}

function savePurchaseOrder() {
  if (!poForm.value.vendorId || poForm.value.items.length === 0) {
    info('Please select a vendor and add at least one item')
    return
  }

  const vendor = selectedVendor.value!

  poStore.addOrder({
    vendorId: poForm.value.vendorId,
    vendorName: vendor.name,
    vendorCode: vendor.vendorCode,
    orderDate: poForm.value.orderDate,
    expectedDeliveryDate: poForm.value.expectedDeliveryDate,
    deliveryWarehouse: poForm.value.deliveryWarehouse,
    status: 'draft',
    approvalStatus: 'pending',
    items: poForm.value.items,
    subtotal: poTotals.value.subtotal,
    discountAmount: poTotals.value.discountAmount,
    taxAmount: poTotals.value.taxAmount,
    freightCost: poForm.value.freightCost,
    otherCharges: poForm.value.otherCharges,
    totalAmount: poTotals.value.totalAmount,
    currency: 'SAR',
    paymentTerms: vendor.paymentTerms,
    paymentStatus: 'unpaid',
    referenceNumber: poForm.value.referenceNumber,
    projectReference: poForm.value.projectReference,
    notes: poForm.value.notes,
    createdBy: 'USR-001'
  })

  success('Purchase Order created successfully')
  showCreatePOModal.value = false
  resetPOForm()
}

function resetPOForm() {
  poForm.value = {
    vendorId: '',
    orderDate: new Date().toISOString().split('T')[0] || '',
    expectedDeliveryDate: '',
    deliveryWarehouse: 'Riyadh Main',
    items: [],
    freightCost: 0,
    otherCharges: 0,
    referenceNumber: '',
    projectReference: '',
    projectId: '',
    notes: ''
  }
}

function viewOrderDetails(order: any) {
  selectedOrder.value = order
  showDetailsModal.value = true
}

function editOrder(order: any) {
  info(`Editing order ${order.poNumber}`)
}

function approveOrder(order: any) {
  poStore.approveOrder(order.id, 'USR-001')
  success(`Purchase Order ${order.poNumber} approved`)
}

function receiveGoods(order: any) {
  receivingOrder.value = order

  // Initialize receipt form with PO items
  receiptForm.value.items = order.items.map((item: any) => ({
    poLineItemId: item.id,
    productId: item.productId,
    productSku: item.productSku,
    productName: item.productName,
    orderedQuantity: item.quantity,
    receivedQuantity: 0,
    acceptedQuantity: 0,
    rejectedQuantity: 0,
    rejectionReason: '',
    unitPrice: item.unitPrice,
    totalValue: 0,
    batchNumber: '',
    notes: ''
  }))

  showReceiveModal.value = true
}

function getAlreadyReceived(poLineItemId: string) {
  const item = receivingOrder.value?.items.find((i: any) => i.id === poLineItemId)
  return item?.receivedQuantity || 0
}

function calculateReceivedTotals(index: number) {
  const item = receiptForm.value.items[index]
  item.rejectedQuantity = Math.max(0, item.receivedQuantity - item.acceptedQuantity)
  item.totalValue = item.acceptedQuantity * item.unitPrice
}

function saveGoodsReceipt() {
  if (!receivingOrder.value) return

  const totalReceiving = receiptForm.value.items.reduce((sum, item) => sum + item.receivedQuantity, 0)
  if (totalReceiving === 0) {
    info('Please enter quantities to receive')
    return
  }

  // Create goods receipt
  receiptsStore.addReceipt({
    purchaseOrderId: receivingOrder.value!.id,
    purchaseOrderNumber: receivingOrder.value!.poNumber,
    vendorId: receivingOrder.value!.vendorId,
    vendorName: receivingOrder.value!.vendorName,
    receiptDate: receiptForm.value.receiptDate,
    receiptType: receiptForm.value.receiptType,
    warehouseLocation: receivingOrder.value!.deliveryWarehouse,
    deliveryNoteNumber: receiptForm.value.deliveryNoteNumber,
    carrierName: receiptForm.value.carrierName,
    trackingNumber: receiptForm.value.trackingNumber,
    items: receiptForm.value.items,
    qualityCheckStatus: receiptForm.value.qualityCheckStatus,
    status: 'completed',
    totalValue: receiptForm.value.items.reduce((sum, item) => sum + item.totalValue, 0),
    currency: 'SAR',
    hasDiscrepancy: receiptForm.value.items.some(item => item.rejectedQuantity > 0),
    receivedBy: 'USR-001',
    qualityCheckedBy: 'USR-003',
    qualityCheckedAt: new Date().toISOString(),
    qualityNotes: receiptForm.value.qualityNotes,
    notes: receiptForm.value.notes
  })

  // Update PO with received quantities
  receiptForm.value.items.forEach(item => {
    if (item.receivedQuantity > 0) {
      poStore.receiveOrderItem(receivingOrder.value!.id, item.poLineItemId, item.receivedQuantity)

      // Update warehouse stock
      const existingStock = warehouseStore.getStockByProductAndLocation(
        item.productId,
        receivingOrder.value!.deliveryWarehouse as any
      )

      if (existingStock) {
        // Adjust existing stock
        warehouseStore.adjustQuantity(
          existingStock.id,
          item.acceptedQuantity,
          `Received from PO ${receivingOrder.value!.poNumber}`
        )
      } else {
        // Add new stock record
        const product = productsStore.getProductById(item.productId)
        if (product) {
          warehouseStore.addStock({
            warehouseLocation: receivingOrder.value!.deliveryWarehouse as any,
            productId: item.productId,
            quantityOnHand: item.acceptedQuantity,
            quantityReserved: 0,
            quantityAvailable: item.acceptedQuantity,
            unitCost: item.unitPrice,
            totalValue: item.acceptedQuantity * item.unitPrice,
            batchNumber: item.batchNumber,
            supplierId: product.vendorId,
            supplierSku: item.productSku,
            lastPurchaseDate: new Date().toISOString(),
            lastPurchasePrice: item.unitPrice
          })
        }
      }

      // Record stock movement
      movementsStore.addMovement({
        productId: item.productId,
        warehouseLocation: receivingOrder.value!.deliveryWarehouse as any,
        movementType: 'receipt',
        quantity: item.acceptedQuantity,
        referenceType: 'purchase_order',
        referenceNumber: receivingOrder.value!.poNumber,
        performedBy: 'USR-001',
        notes: `Received from PO ${receivingOrder.value!.poNumber}`
      })
    }
  })

  success(`Goods received successfully for PO ${receivingOrder.value.poNumber}`)
  showReceiveModal.value = false
  receivingOrder.value = null
  resetReceiptForm()
}

function resetReceiptForm() {
  receiptForm.value = {
    receiptDate: new Date().toISOString().split('T')[0] || '',
    receiptType: 'partial',
    deliveryNoteNumber: '',
    carrierName: '',
    trackingNumber: '',
    items: [],
    qualityCheckStatus: 'pending',
    qualityNotes: '',
    notes: ''
  }
}
</script>

<style scoped>
.procurement-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 100%;
}

/* Stats Bar */
.stats-bar {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
  margin-bottom: 0.5rem;
}

.stat-card {
  background: var(--color-surface);
  padding: 1.5rem;
  border-radius: 12px;
  box-shadow: var(--shadow-sm);
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.stat-label {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-value {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.stat-value.warning {
  color: var(--color-warning);
}

.stat-value.info {
  color: var(--color-info);
}

/* Header */
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 1rem;
  width: 100%;
  align-items: center;
}

.search-input {
  flex: 1;
  padding: 1rem 1.5rem;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text-primary);
  border-radius: 10px;
  font-size: 1.05rem;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.btn {
  padding: 1rem 2rem;
  border: none;
  border-radius: 10px;
  font-weight: 500;
  font-size: 1.05rem;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-primary {
  background: var(--color-primary);
  color: #fff;
}

.btn-primary:hover {
  background: var(--color-primary-hover);
  transform: translateY(-2px);
}

.btn-small {
  padding: 0.5rem 1rem;
  font-size: 0.9rem;
}

/* Tabs */
.tabs {
  display: flex;
  gap: 0.5rem;
  border-bottom: 2px solid var(--color-border);
  padding: 0 0.5rem;
}

.tab {
  padding: 1rem 1.5rem;
  background: none;
  border: none;
  border-bottom: 3px solid transparent;
  color: var(--color-text-secondary);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  margin-bottom: -2px;
}

.tab:hover {
  color: var(--color-text-primary);
  background: var(--color-surface-hover);
}

.tab.active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
}

/* Filters */
.filters-bar {
  display: flex;
  gap: 1.5rem;
  align-items: center;
  flex-wrap: wrap;
  padding: 1rem;
  background: var(--color-surface);
  border-radius: 10px;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.filter-group label {
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--color-text-secondary);
}

.filter-select {
  padding: 0.5rem 1rem;
  border: 1px solid var(--color-border);
  background: var(--color-background);
  color: var(--color-text-primary);
  border-radius: 8px;
  cursor: pointer;
  min-width: 150px;
}

/* Table */
.table-container {
  overflow-x: auto;
}

.orders-table {
  width: 100%;
  border-collapse: collapse;
}

.orders-table th,
.orders-table td {
  padding: 1rem;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
}

.orders-table th {
  background: var(--color-background);
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.orders-table td {
  font-size: 0.95rem;
  color: var(--color-text-primary);
}

.order-row:hover {
  background: var(--color-surface-hover);
}

.po-number-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.po-number {
  font-weight: 600;
  font-family: monospace;
}

.po-items {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
}

.vendor-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.vendor-name {
  font-weight: 500;
}

.vendor-code {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  font-family: monospace;
}

.delivery-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.overdue-badge {
  font-size: 0.75rem;
  color: var(--color-danger);
  font-weight: 600;
}

.progress-cell {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  align-items: flex-end;
}

.progress-bar-container {
  width: 100px;
  height: 8px;
  background: var(--color-border);
  border-radius: 4px;
  overflow: hidden;
}

.progress-bar {
  height: 100%;
  background: var(--color-success);
  transition: width 0.3s;
}

.progress-text {
  font-size: 0.85rem;
  font-weight: 600;
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-start;
}

.action-btn {
  background: none;
  border: none;
  font-size: 1.1rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  transition: transform 0.2s;
  opacity: 0.7;
}

.action-btn:hover {
  transform: scale(1.2);
  opacity: 1;
}

.empty-row {
  text-align: center;
  padding: 3rem;
  color: var(--color-text-secondary);
  font-size: 1.05rem;
}

.text-right {
  text-align: right !important;
}

/* Forms */
.po-form,
.receive-form {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-section h3 {
  margin: 0;
  font-size: 1.125rem;
  color: var(--color-text-primary);
  padding-bottom: 0.5rem;
  border-bottom: 2px solid var(--color-border);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 0;
  font-size: 1.125rem;
  color: var(--color-text-primary);
  padding-bottom: 0.5rem;
  border-bottom: 2px solid var(--color-border);
}

.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group.full-width {
  grid-column: 1 / -1;
}

.form-label {
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--color-text-primary);
}

.form-input {
  padding: 0.75rem;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text-primary);
  border-radius: 8px;
  font-size: 0.95rem;
  transition: all 0.2s;
}

.form-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.form-hint {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  font-style: italic;
}

.form-hint.success {
  color: var(--color-success);
  font-weight: 500;
  font-style: normal;
}

textarea.form-input {
  resize: vertical;
  font-family: inherit;
}

/* Line Items */
.line-items-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.line-item {
  padding: 1.5rem;
  background: var(--color-background);
  border-radius: 10px;
  border: 1px solid var(--color-border);
}

.line-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.line-item-number {
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-primary);
}

.btn-remove {
  background: var(--color-danger);
  color: white;
  border: none;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  font-size: 1.5rem;
  line-height: 1;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-remove:hover {
  transform: scale(1.1);
}

.empty-items {
  padding: 3rem;
  text-align: center;
  color: var(--color-text-secondary);
  border: 2px dashed var(--color-border);
  border-radius: 10px;
}

/* Financial Summary */
.financial-grid {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.summary-box {
  background: var(--color-background);
  padding: 1.5rem;
  border-radius: 10px;
  border: 1px solid var(--color-border);
}

.summary-row {
  display: flex;
  justify-content: space-between;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--color-border);
  font-size: 0.95rem;
}

.summary-row:last-child {
  border-bottom: none;
}

.summary-row.total-row {
  margin-top: 0.5rem;
  padding-top: 1rem;
  border-top: 2px solid var(--color-border);
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--color-primary);
}

/* Order Details */
.order-details {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.details-header {
  background: var(--color-background);
  padding: 1.5rem;
  border-radius: 10px;
}

/* Project Banner */
.project-banner {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.25rem;
  margin-bottom: 1.5rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 10px;
  color: white;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.project-icon {
  font-size: 1.5rem;
}

.project-info-banner {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.project-label {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  opacity: 0.9;
  font-weight: 600;
}

.project-name {
  font-size: 1rem;
  font-weight: 700;
  letter-spacing: 0.3px;
}

.details-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.detail-label {
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.detail-value {
  font-size: 0.95rem;
  color: var(--color-text-primary);
  font-weight: 500;
}

.details-section {
  padding-bottom: 1.5rem;
  border-bottom: 1px solid var(--color-border);
}

.details-section:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.details-table {
  width: 100%;
  border-collapse: collapse;
}

.details-table th,
.details-table td {
  padding: 0.75rem;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
}

.details-table th {
  background: var(--color-background);
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
}

.product-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.product-sku {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  font-family: monospace;
}

.received-qty {
  color: var(--color-success);
  font-weight: 600;
}

.remaining-qty {
  color: var(--color-warning);
  font-weight: 600;
}

.financial-summary {
  max-width: 400px;
  margin-left: auto;
  background: var(--color-background);
  padding: 1.5rem;
  border-radius: 10px;
}

.financial-summary .received-row {
  color: var(--color-success);
  font-weight: 600;
}

.receipts-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.receipt-card {
  padding: 1rem;
  background: var(--color-background);
  border-radius: 8px;
  border: 1px solid var(--color-border);
}

.receipt-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.receipt-number {
  font-weight: 600;
  font-family: monospace;
}

.receipt-info {
  display: flex;
  gap: 1.5rem;
  font-size: 0.85rem;
  color: var(--color-text-secondary);
}

.notes-box {
  padding: 1rem;
  background: var(--color-background);
  border-radius: 8px;
  color: var(--color-text-primary);
  line-height: 1.6;
}

/* Receive Items */
.receive-items {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.receive-item {
  padding: 1.5rem;
  background: var(--color-background);
  border-radius: 10px;
  border: 1px solid var(--color-border);
}

.receive-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--color-border);
}

.product-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.product-name {
  font-weight: 600;
  font-size: 1.05rem;
}

.ordered-info {
  font-size: 0.9rem;
  color: var(--color-text-secondary);
}
</style>
