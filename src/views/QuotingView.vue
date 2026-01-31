<template>
  <div class="quoting-view">
    <div class="view-header">
      <div class="header-actions">
        <input
          v-model="searchQuery"
          type="search"
          placeholder="Search quotes by number, customer, or opportunity..."
          class="search-input"
        />
        <button class="btn btn-primary" @click="createNewQuote">
          <span>➕</span> New Quote
        </button>
      </div>
    </div>

    <div class="filters-bar">
      <div class="filter-group">
        <label>Status:</label>
        <select v-model="statusFilter" class="filter-select">
          <option value="all">All Status</option>
          <option value="draft">Draft</option>
          <option value="pending_approval">Pending Approval</option>
          <option value="approved">Approved</option>
          <option value="sent">Sent</option>
          <option value="accepted">Accepted</option>
          <option value="declined">Declined</option>
          <option value="expired">Expired</option>
        </select>
      </div>

      <div class="filter-group">
        <label>Customer:</label>
        <select v-model="customerFilter" class="filter-select">
          <option value="all">All Customers</option>
          <option v-for="customer in customersStore.customers" :key="customer.id" :value="customer.id">
            {{ customer.name }}
          </option>
        </select>
      </div>

      <div class="filter-group">
        <label>Opportunity:</label>
        <select v-model="opportunityFilter" class="filter-select">
          <option value="all">All Opportunities</option>
          <option value="no_opportunity">No Opportunity</option>
          <option v-for="opp in opportunitiesStore.activeOpportunities" :key="opp.id" :value="opp.id">
            {{ opp.opportunityNumber }} - {{ opp.name }}
          </option>
        </select>
      </div>

      <div class="filter-group">
        <label>Sort By:</label>
        <select v-model="sortBy" class="filter-select">
          <option value="date-desc">Newest First</option>
          <option value="date-asc">Oldest First</option>
          <option value="value-desc">Highest Value</option>
          <option value="value-asc">Lowest Value</option>
        </select>
      </div>

      <div class="stats-summary">
        <div class="stat-item">
          <span class="stat-label">Total Quotes:</span>
          <span class="stat-value">{{ filteredQuotes.length }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Total Value:</span>
          <span class="stat-value">{{ formatCurrency(totalValue) }}</span>
        </div>
      </div>
    </div>

    <BaseCard no-padding>
      <BaseTable
        :columns="quotesColumns"
        :data="sortedQuotes"
        :clickable="true"
        @row-click="viewQuoteDetails"
      >
        <template #cell-quoteNumber="{ row }">
          <div class="quote-number-cell">
            <div class="quote-number">{{ row.quoteNumber }}</div>
            <div class="quote-version">v{{ row.version }}</div>
          </div>
        </template>

        <template #cell-customer="{ row }">
          <div class="customer-cell">
            <div class="customer-name">{{ getCustomerName(row.companyId) }}</div>
            <div class="opportunity-name">{{ getOpportunityName(row.opportunityId) }}</div>
          </div>
        </template>

        <template #cell-dates="{ row }">
          <div class="dates-cell">
            <div class="date-row">
              <span class="date-label">Created:</span>
              <span>{{ formatDate(row.quoteDate) }}</span>
            </div>
            <div class="date-row">
              <span class="date-label">Valid Until:</span>
              <span :class="{ 'text-danger': isExpiringSoon(row.validUntil) }">
                {{ formatDate(row.validUntil) }}
              </span>
            </div>
          </div>
        </template>

        <template #cell-lineItems="{ row }">
          <div class="line-items-summary">
            <div class="items-total">{{ getLineItemsSummary(row.id).total }} items</div>
            <div class="items-breakdown">
              <span v-if="getLineItemsSummary(row.id).materials > 0" class="item-badge materials">
                📦 {{ getLineItemsSummary(row.id).materials }}
              </span>
              <span v-if="getLineItemsSummary(row.id).manpower > 0" class="item-badge manpower">
                👷 {{ getLineItemsSummary(row.id).manpower }}
              </span>
              <span v-if="getLineItemsSummary(row.id).miscellaneous > 0" class="item-badge misc">
                📋 {{ getLineItemsSummary(row.id).miscellaneous }}
              </span>
            </div>
          </div>
        </template>

        <template #cell-financial="{ row }">
          <div class="financial-cell">
            <div class="amount-row">{{ formatCurrency(row.total) }}</div>
            <div class="margin-row">
              Margin: <span :class="getMarginClass(row.marginPercent)">{{ row.marginPercent.toFixed(1) }}%</span>
            </div>
          </div>
        </template>

        <template #cell-status="{ value }">
          <BaseBadge :variant="getStatusVariant(value)" size="sm">
            {{ formatStatus(value) }}
          </BaseBadge>
        </template>

        <template #cell-approval="{ row }">
          <div v-if="row.requiresApproval" class="approval-cell">
            <span class="approval-icon">⚠️</span>
            <span class="approval-text">{{ row.approvalReason }}</span>
          </div>
          <div v-else class="approval-cell">
            <span class="approval-icon">✓</span>
            <span class="approval-text">Standard</span>
          </div>
        </template>

        <template #cell-actions="{ row }">
          <div class="action-buttons">
            <button class="action-btn" title="View Details" @click.stop="viewQuoteDetails(row)">👁️</button>
            <button class="action-btn" title="Edit" @click.stop="editQuote(row)" :disabled="!canEdit(row)">✏️</button>
            <button class="action-btn" title="Duplicate" @click.stop="duplicateQuote(row)">📋</button>
            <button class="action-btn" title="Download PDF" @click.stop="downloadPDF(row)">📄</button>
            <button
              v-if="row.status === 'approved' && row.status !== 'sent'"
              class="action-btn"
              title="Send to Customer"
              @click.stop="sendToCustomer(row)"
            >
              📧
            </button>
            <button class="action-btn action-btn-danger" title="Delete" @click.stop="deleteQuote(row)" :disabled="row.status === 'sent' || row.status === 'accepted'">🗑️</button>
          </div>
        </template>
      </BaseTable>
    </BaseCard>

    <!-- Add/Edit Quote Modal -->
    <BaseModal
      v-model="showQuoteModal"
      :title="editingQuote ? 'Edit Quote' : 'Create New Quote'"
      size="xl"
      :hideFooter="true"
      @close="resetQuoteForm"
    >
      <QuoteBuilderModern />
    </BaseModal>

    <!-- Delete Confirmation Modal -->
    <BaseModal
      v-model="showDeleteModal"
      title="Delete Quote"
      size="sm"
      @confirm="confirmDelete"
    >
      <p>Are you sure you want to delete this quote?</p>
      <p><strong>{{ quoteToDelete?.quoteNumber }}</strong></p>
      <p class="text-danger">This action cannot be undone.</p>
    </BaseModal>

    <!-- Quote Details Modal -->
    <BaseModal
      v-if="selectedQuote"
      v-model="showDetailsModal"
      :title="`${selectedQuote.quoteNumber} v${selectedQuote.version} - ${getCustomerName(selectedQuote.companyId)}`"
      size="xl"
    >
      <div class="quote-details-modern">
          <!-- Status Progress Bar -->
          <div class="modern-progress-card">
            <div class="progress-steps">
              <div
                v-for="(stage, index) in workflowStages"
                :key="stage.key"
                class="progress-step"
                :class="getStepClass(stage.key, selectedQuote.status)"
              >
                <div class="step-icon">
                  <span v-if="isStepCompleted(stage.key, selectedQuote.status)">✓</span>
                  <span v-else>{{ index + 1 }}</span>
                </div>
                <div class="step-label">{{ stage.label }}</div>
              </div>
            </div>
          </div>

          <!-- Header Information -->
          <div class="modern-info-card">
            <div class="card-header-mini">
              <h4>📋 Quote Information</h4>
              <BaseBadge :variant="getStatusVariant(selectedQuote.status)" size="sm">
                {{ formatStatus(selectedQuote.status) }}
              </BaseBadge>
            </div>
            <div class="card-content">
              <div class="info-grid">
                <div class="info-item">
                  <span class="info-label">Quote Number</span>
                  <span class="info-value highlight">{{ selectedQuote.quoteNumber }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">Version</span>
                  <span class="info-value">v{{ selectedQuote.version }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">Customer</span>
                  <span class="info-value">{{ getCustomerName(selectedQuote.companyId) }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">Contact</span>
                  <span class="info-value">{{ getContactName(selectedQuote.contactId) }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">Opportunity</span>
                  <span class="info-value">
                    <span
                      v-if="selectedQuote.opportunityId"
                      class="opportunity-link"
                      @click.stop="viewOpportunity(selectedQuote.opportunityId)"
                    >
                      {{ getOpportunityName(selectedQuote.opportunityId) }}
                    </span>
                    <span v-else class="no-opportunity">No linked opportunity</span>
                  </span>
                </div>
                <div class="info-item">
                  <span class="info-label">Currency</span>
                  <span class="info-value">{{ selectedQuote.currency }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">Quote Date</span>
                  <span class="info-value">{{ formatDate(selectedQuote.quoteDate) }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">Valid Until</span>
                  <span class="info-value">{{ formatDate(selectedQuote.validUntil) }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Line Items - Modern Card Layout -->
          <div class="modern-info-card items-card">
            <div class="card-header-mini">
              <h4>📦 Line Items</h4>
              <span class="count-badge">
                {{ getCategorizedLineItems(selectedQuote.id).materials.length +
                   getCategorizedLineItems(selectedQuote.id).manpower.length +
                   getCategorizedLineItems(selectedQuote.id).miscellaneous.length }}
              </span>
            </div>
            <div class="card-content">
              <!-- Materials & Equipment Category -->
              <div v-if="getCategorizedLineItems(selectedQuote.id).materials.length > 0" class="category-section-modern">
                <div class="category-header-modern">
                  <span class="category-icon">📦</span>
                  <h5>Materials & Equipment</h5>
                  <span class="category-count">{{ getCategorizedLineItems(selectedQuote.id).materials.length }}</span>
                </div>
                <div class="items-list-modern">
                  <div v-for="item in getCategorizedLineItems(selectedQuote.id).materials" :key="item.id" class="item-card-mini">
                    <div class="item-number-mini">{{ item.lineNumber }}</div>
                    <div class="item-details">
                      <div class="item-header-row">
                        <div class="item-name">
                          <span class="item-code">{{ item.productSku || '-' }}</span>
                          <span class="item-title">{{ item.productName }}</span>
                          <span v-if="item.description" class="item-desc">{{ item.description }}</span>
                        </div>
                        <div class="item-final-price">{{ formatCurrency(item.lineTotal) }}</div>
                      </div>
                      <div class="item-pricing-row">
                        <div class="pricing-detail">
                          <span class="pricing-label">Qty:</span>
                          <span class="pricing-value">{{ item.quantity }}</span>
                        </div>
                        <div class="pricing-detail">
                          <span class="pricing-label">Unit Price:</span>
                          <span class="pricing-value">{{ formatCurrency(item.unitPrice) }}</span>
                        </div>
                        <div class="pricing-detail">
                          <span class="pricing-label">Margin:</span>
                          <span class="pricing-value" :class="getMarginClass(item.marginPercent)">
                            {{ item.marginPercent.toFixed(1) }}%
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
                <div class="category-subtotal">
                  <span>Subtotal:</span>
                  <span class="subtotal-value">{{ formatCurrency(getCategorySubtotal(getCategorizedLineItems(selectedQuote.id).materials)) }}</span>
                </div>
              </div>

              <!-- Manpower & Services Category -->
              <div v-if="getCategorizedLineItems(selectedQuote.id).manpower.length > 0" class="category-section-modern">
                <div class="category-header-modern">
                  <span class="category-icon">👷</span>
                  <h5>Manpower & Services</h5>
                  <span class="category-count">{{ getCategorizedLineItems(selectedQuote.id).manpower.length }}</span>
                </div>
                <div class="items-list-modern">
                  <div v-for="item in getCategorizedLineItems(selectedQuote.id).manpower" :key="item.id" class="item-card-mini">
                    <div class="item-number-mini">{{ item.lineNumber }}</div>
                    <div class="item-details">
                      <div class="item-header-row">
                        <div class="item-name">
                          <span class="item-code">{{ item.laborPosition || item.productName }}</span>
                          <span class="item-title">{{ item.productName }}</span>
                          <span v-if="item.description" class="item-desc">{{ item.description }}</span>
                          <span v-if="item.hoursPerMonth" class="item-meta-tag">
                            {{ item.hoursPerMonth }} hrs/month × {{ item.months }} months
                          </span>
                          <span v-else-if="item.months && item.months > 0 && item.type === 'service'" class="item-meta-tag recurring">
                            🔄 {{ item.months }} month{{ item.months > 1 ? 's' : '' }} duration
                          </span>
                        </div>
                        <div class="item-final-price">{{ formatCurrency(item.lineTotal) }}</div>
                      </div>
                      <div class="item-pricing-row">
                        <div class="pricing-detail">
                          <span class="pricing-label">Duration/Qty:</span>
                          <span class="pricing-value">
                            <span v-if="item.months && item.months > 0 && item.type === 'service'">
                              {{ item.months }} month{{ item.months > 1 ? 's' : '' }}
                            </span>
                            <span v-else>{{ item.quantity }}</span>
                          </span>
                        </div>
                        <div class="pricing-detail">
                          <span class="pricing-label">Rate:</span>
                          <span class="pricing-value">
                            {{ formatCurrency(item.unitPrice) }}
                            <span v-if="item.type === 'service'" class="rate-suffix">/mo</span>
                          </span>
                        </div>
                        <div class="pricing-detail">
                          <span class="pricing-label">Margin:</span>
                          <span class="pricing-value" :class="getMarginClass(item.marginPercent)">
                            {{ item.marginPercent.toFixed(1) }}%
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
                <div class="category-subtotal">
                  <span>Subtotal:</span>
                  <span class="subtotal-value">{{ formatCurrency(getCategorySubtotal(getCategorizedLineItems(selectedQuote.id).manpower)) }}</span>
                </div>
              </div>

              <!-- Miscellaneous Category -->
              <div v-if="getCategorizedLineItems(selectedQuote.id).miscellaneous.length > 0" class="category-section-modern">
                <div class="category-header-modern">
                  <span class="category-icon">📋</span>
                  <h5>Miscellaneous</h5>
                  <span class="category-count">{{ getCategorizedLineItems(selectedQuote.id).miscellaneous.length }}</span>
                </div>
                <div class="items-list-modern">
                  <div v-for="item in getCategorizedLineItems(selectedQuote.id).miscellaneous" :key="item.id" class="item-card-mini">
                    <div class="item-number-mini">{{ item.lineNumber }}</div>
                    <div class="item-details">
                      <div class="item-header-row">
                        <div class="item-name">
                          <span class="item-title">{{ item.productName }}</span>
                          <span v-if="item.description" class="item-desc">{{ item.description }}</span>
                        </div>
                        <div class="item-final-price">{{ formatCurrency(item.lineTotal) }}</div>
                      </div>
                      <div class="item-pricing-row">
                        <div class="pricing-detail">
                          <span class="pricing-label">Qty:</span>
                          <span class="pricing-value">{{ item.quantity }}</span>
                        </div>
                        <div class="pricing-detail">
                          <span class="pricing-label">Unit Price:</span>
                          <span class="pricing-value">{{ formatCurrency(item.unitPrice) }}</span>
                        </div>
                        <div class="pricing-detail">
                          <span class="pricing-label">Margin:</span>
                          <span class="pricing-value" :class="getMarginClass(item.marginPercent)">
                            {{ item.marginPercent.toFixed(1) }}%
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
                <div class="category-subtotal">
                  <span>Subtotal:</span>
                  <span class="subtotal-value">{{ formatCurrency(getCategorySubtotal(getCategorizedLineItems(selectedQuote.id).miscellaneous)) }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Financial Summary -->
          <div class="modern-info-card financial-card">
            <div class="card-header-mini">
              <h4>💰 Financial Summary</h4>
            </div>
            <div class="card-content">
              <div class="financial-grid-modern">
                <div class="financial-row">
                  <span class="financial-label">Subtotal:</span>
                  <span class="financial-value">{{ formatCurrency(selectedQuote.subtotal) }}</span>
                </div>
                <div v-if="selectedQuote.discountPercent > 0" class="financial-row discount-row">
                  <span class="financial-label">Discount ({{ selectedQuote.discountPercent }}%):</span>
                  <span class="financial-value discount">-{{ formatCurrency(selectedQuote.discountAmount) }}</span>
                </div>
                <div class="financial-row">
                  <span class="financial-label">After Discount:</span>
                  <span class="financial-value">{{ formatCurrency(selectedQuote.subtotalAfterDiscount) }}</span>
                </div>
                <div class="financial-row">
                  <span class="financial-label">VAT ({{ selectedQuote.vatPercent }}%):</span>
                  <span class="financial-value">{{ formatCurrency(selectedQuote.vatAmount) }}</span>
                </div>
                <div class="financial-row total-row-modern">
                  <span class="financial-label">Grand Total:</span>
                  <span class="financial-value total">{{ formatCurrency(selectedQuote.total) }}</span>
                </div>
                <div class="divider-line"></div>
                <div class="financial-row cost-row">
                  <span class="financial-label">Total Cost:</span>
                  <span class="financial-value cost">{{ formatCurrency(selectedQuote.totalCost) }}</span>
                </div>
                <div class="financial-row profit-row">
                  <span class="financial-label">Profit:</span>
                  <span class="financial-value profit">{{ formatCurrency(selectedQuote.marginAmount) }}</span>
                </div>
                <div class="financial-row margin-row-modern">
                  <span class="financial-label">Margin:</span>
                  <span class="financial-value margin" :class="getMarginClass(selectedQuote.marginPercent)">
                    {{ selectedQuote.marginPercent.toFixed(1) }}%
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Terms & Notes -->
          <div class="modern-info-card">
            <div class="card-header-mini">
              <h4>📜 Terms & Conditions</h4>
            </div>
            <div class="card-content">
              <div class="terms-grid">
                <div class="term-box">
                  <span class="term-label">💳 Payment Terms</span>
                  <p class="term-value">{{ selectedQuote.paymentTerms }}</p>
                </div>
                <div class="term-box">
                  <span class="term-label">🚚 Delivery Terms</span>
                  <p class="term-value">{{ selectedQuote.deliveryTerms }}</p>
                </div>
                <div v-if="selectedQuote.customerNotes" class="term-box full-width">
                  <span class="term-label">📝 Customer Notes</span>
                  <p class="term-value notes">{{ selectedQuote.customerNotes }}</p>
                </div>
                <div v-if="selectedQuote.internalNotes" class="term-box full-width internal-note">
                  <span class="term-label">🔒 Internal Notes</span>
                  <p class="term-value notes">{{ selectedQuote.internalNotes }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Approval History -->
          <div v-if="selectedQuote.approvalHistory && selectedQuote.approvalHistory.length > 0" class="details-section">
            <h3>Approval History</h3>
            <div class="approval-history-table">
              <table>
                <thead>
                  <tr>
                    <th width="20%">Date</th>
                    <th width="20%">Status</th>
                    <th width="25%">User</th>
                    <th width="35%">Comments</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(entry, index) in selectedQuote.approvalHistory" :key="index">
                    <td>{{ formatDate(entry.date) }}</td>
                    <td>
                      <BaseBadge :variant="getStatusVariant(entry.status)">
                        {{ formatStatus(entry.status) }}
                      </BaseBadge>
                    </td>
                    <td>{{ entry.userName }}</td>
                    <td>{{ entry.comments }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Workflow Actions -->
          <div v-if="selectedQuote.status !== 'accepted' && selectedQuote.status !== 'declined'" class="details-section">
            <h3>Workflow Actions</h3>
            <div class="workflow-actions">
              <button
                v-if="selectedQuote.status === 'draft'"
                class="btn btn-primary"
                @click="submitForApproval(selectedQuote)"
              >
                Submit for Approval
              </button>
              <button
                v-if="selectedQuote.status === 'pending_approval'"
                class="btn btn-success"
                @click="approveQuoteAction(selectedQuote)"
              >
                Approve Quote
              </button>
              <button
                v-if="selectedQuote.status === 'pending_approval'"
                class="btn btn-danger"
                @click="rejectQuoteAction(selectedQuote)"
              >
                Reject Quote
              </button>
              <button
                v-if="selectedQuote.status === 'approved'"
                class="btn btn-primary"
                @click="sendToCustomer(selectedQuote)"
              >
                Send to Customer
              </button>
            </div>
          </div>
        </div>

      <template #footer>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="closeDetailsModal">Close</button>
          <button v-if="canEdit(selectedQuote)" class="btn btn-primary" @click="editQuote(selectedQuote)">
            Edit Quote
          </button>
          <button class="btn btn-outline" @click="printQuote(selectedQuote)" title="Print Quote">
            🖨️ Print
          </button>
          <button class="btn btn-outline" @click="exportQuoteToExcel(selectedQuote)" title="Export to Excel">
            📊 Export Excel
          </button>
          <button class="btn btn-primary" @click="downloadPDF(selectedQuote)">
            📄 Download PDF
          </button>
          <button
            v-if="selectedQuote.status === 'approved'"
            class="btn btn-success"
            @click="sendToCustomer(selectedQuote)"
          >
            📧 Send to Customer
          </button>
        </div>
      </template>
    </BaseModal>

    <!-- Add/Edit Line Item Modal -->
    <BaseModal
      v-model="showAddLineItemModal"
      :title="editingLineItem ? 'Edit Line Item' : 'Add Line Item'"
      size="lg"
    >
      <form @submit.prevent="saveLineItem" class="line-item-form">
        <!-- Two-Dropdown Item Selection -->
        <div class="form-row">
          <div class="form-group">
            <label class="required">Category</label>
            <select v-model="lineItemForm.category" class="form-control" @change="onCategoryChange">
              <option value="">-- Select Category --</option>
              <option value="product">📦 Products</option>
              <option value="recurring">🔄 Recurring Services</option>
              <option value="labor">👷 Labor Positions</option>
            </select>
          </div>
        </div>

        <div v-if="lineItemForm.category" class="form-row">
          <div class="form-group">
            <label class="required">Select {{ getCategoryLabel() }}</label>
            <select v-model="selectedItemId" class="form-control" @change="onItemSelect">
              <option value="">-- Select {{ getCategoryLabel() }} --</option>
              <option
                v-for="item in getFilteredItems()"
                :key="item.id"
                :value="`${item.type}:${item.id}`"
              >
                {{ item.icon }} {{ item.sku ? `${item.sku} - ` : '' }}{{ item.name }}{{ item.department ? ` (${item.department})` : '' }} - {{ item.priceLabel }}
              </option>
            </select>
            <small v-if="quoteForm.priceBookId && availableItems.length > 0" class="form-hint">
              Showing {{ getFilteredItems().length }} item(s) from selected price book
            </small>
            <small v-else-if="quoteForm.priceBookId && availableItems.length === 0" class="form-hint warning">
              No items available in selected price book
            </small>
            <small v-else class="form-hint">
              {{ getFilteredItems().length }} items available
            </small>
          </div>
        </div>

        <!-- Product/Service Name -->
        <div class="form-row">
          <div class="form-group">
            <label class="required">Item Name</label>
            <input
              v-model="lineItemForm.productName"
              type="text"
              class="form-control"
              placeholder="e.g., HP ProLiant Server or Senior Engineer"
              required
            />
          </div>
        </div>

        <!-- SKU (for products) -->
        <div v-if="lineItemForm.type === 'product'" class="form-row">
          <div class="form-group">
            <label>Product SKU/Code</label>
            <input
              v-model="lineItemForm.productSku"
              type="text"
              class="form-control"
              placeholder="Product SKU"
            />
          </div>
        </div>

        <!-- Description -->
        <div class="form-row">
          <div class="form-group">
            <label>Description</label>
            <textarea
              v-model="lineItemForm.description"
              class="form-control"
              rows="3"
              placeholder="Enter detailed description..."
            ></textarea>
          </div>
        </div>

        <!-- Labor-specific fields -->
        <div v-if="lineItemForm.type === 'labor'" class="form-row">
          <div class="form-group">
            <label>Position</label>
            <input
              v-model="lineItemForm.laborPosition"
              type="text"
              class="form-control"
              placeholder="e.g., Senior Security Engineer"
            />
          </div>
          <div class="form-group">
            <label>Hours/Month</label>
            <input
              v-model.number="lineItemForm.hoursPerMonth"
              type="number"
              class="form-control"
              min="0"
              @input="calculateLaborQuantity"
            />
          </div>
          <div class="form-group">
            <label>Months</label>
            <input
              v-model.number="lineItemForm.months"
              type="number"
              class="form-control"
              min="0"
              @input="calculateLaborQuantity"
            />
          </div>
        </div>

        <!-- Quantity & Pricing -->
        <div class="form-row">
          <!-- For Services (Manpower), show Duration in Months -->
          <div v-if="lineItemForm.category === 'manpower' && lineItemForm.type === 'service'" class="form-group">
            <label class="required">Duration (Months)</label>
            <input
              v-model.number="lineItemForm.months"
              type="number"
              class="form-control"
              min="1"
              step="1"
              required
              @input="lineItemForm.quantity = lineItemForm.months"
            />
            <small class="form-text">
              Monthly recurring service duration
            </small>
          </div>

          <!-- For Products, show Quantity -->
          <div v-else class="form-group">
            <label class="required">Quantity</label>
            <input
              v-model.number="lineItemForm.quantity"
              type="number"
              class="form-control"
              min="1"
              step="1"
              required
              :readonly="lineItemForm.type === 'labor' && lineItemForm.hoursPerMonth > 0"
            />
            <small v-if="lineItemForm.type === 'labor' && lineItemForm.hoursPerMonth > 0" class="form-text">
              Auto-calculated from hours/month × months
            </small>
          </div>

          <div class="form-group">
            <label class="required">Unit Cost (SAR)</label>
            <input
              v-model.number="lineItemForm.unitCost"
              type="number"
              class="form-control"
              min="0"
              step="0.01"
              required
            />
            <small v-if="selectedProductId || selectedLaborId" class="form-text">
              From catalog: {{ formatCurrency(catalogCost) }}
            </small>
          </div>

          <div class="form-group">
            <label class="required">
              Unit Price (SAR)
              <span v-if="lineItemForm.category === 'manpower' && lineItemForm.type === 'service'">/month</span>
            </label>
            <input
              v-model.number="lineItemForm.unitPrice"
              type="number"
              class="form-control"
              min="0"
              step="0.01"
              required
            />
            <small v-if="selectedProductId && catalogPrice > 0" class="form-text">
              Standard price: {{ formatCurrency(catalogPrice) }}{{ lineItemForm.category === 'manpower' && lineItemForm.type === 'service' ? '/month' : '' }}
              <span v-if="quoteForm.priceBookId && lineItemForm.unitPrice !== catalogPrice" class="price-difference">
                (Price Book: {{ formatCurrency(lineItemForm.unitPrice) }} -
                <span :class="lineItemForm.unitPrice < catalogPrice ? 'discount-text' : 'markup-text'">
                  {{ lineItemForm.unitPrice < catalogPrice ? 'Save' : '+' }}
                  {{ formatCurrency(Math.abs(catalogPrice - lineItemForm.unitPrice)) }}
                  ({{ (Math.abs(catalogPrice - lineItemForm.unitPrice) / catalogPrice * 100).toFixed(1) }}%)
                </span>)
              </span>
            </small>
            <small v-else-if="selectedLaborId" class="form-text">
              Standard rate: {{ formatCurrency(catalogPrice) }}/hour
            </small>
          </div>

          <div class="form-group">
            <label>Discount (%)</label>
            <input
              v-model.number="lineItemForm.discountPercent"
              type="number"
              class="form-control"
              min="0"
              max="100"
              step="0.1"
            />
          </div>
        </div>

        <!-- Preview Calculations -->
        <div class="line-item-preview">
          <div class="preview-row">
            <span class="preview-label">Subtotal:</span>
            <span class="preview-value">{{ formatCurrency(lineItemForm.quantity * lineItemForm.unitPrice) }}</span>
          </div>
          <div class="preview-row">
            <span class="preview-label">Discount:</span>
            <span class="preview-value">-{{ formatCurrency((lineItemForm.quantity * lineItemForm.unitPrice * lineItemForm.discountPercent) / 100) }}</span>
          </div>
          <div class="preview-row preview-total">
            <span class="preview-label">Line Total:</span>
            <span class="preview-value">{{ formatCurrency(lineItemForm.quantity * lineItemForm.unitPrice * (1 - lineItemForm.discountPercent / 100)) }}</span>
          </div>
          <div class="preview-row">
            <span class="preview-label">Total Cost:</span>
            <span class="preview-value">{{ formatCurrency(lineItemForm.quantity * lineItemForm.unitCost) }}</span>
          </div>
          <div class="preview-row">
            <span class="preview-label">Margin:</span>
            <span class="preview-value" :class="getMarginClass(lineItemMarginPercent)">
              {{ lineItemMarginPercent.toFixed(1) }}% ({{ formatCurrency(lineItemMarginAmount) }})
            </span>
          </div>
        </div>

        <!-- Notes -->
        <div class="form-row">
          <div class="form-group">
            <label>Notes</label>
            <textarea
              v-model="lineItemForm.notes"
              class="form-control"
              rows="2"
              placeholder="Internal notes (optional)..."
            ></textarea>
          </div>
        </div>
      </form>

      <template #footer>
        <div class="modal-actions">
          <button type="button" class="btn btn-secondary" @click="closeLineItemModal">Cancel</button>
          <button type="button" class="btn btn-primary" @click="saveLineItem">
            {{ editingLineItem ? 'Update Item' : 'Add Item' }}
          </button>
        </div>
      </template>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useQuotesStore } from '@/stores/quotes'
import { useCustomersStore } from '@/stores/customers'
import { useContactsStore } from '@/stores/contacts'
import { useOpportunitiesStore } from '@/stores/opportunities'
import { useProductsStore } from '@/stores/products'
import { useLaborPositionsStore } from '@/stores/laborPositions'
import { useRecurringServicesStore } from '@/stores/recurringServices'
import { usePriceBooksStore } from '@/stores/priceBooks'
import { useContractsStore } from '@/stores/contracts'
import { useDocumentsStore } from '@/stores/documents'
import { useWarehouseStockStore } from '@/stores/warehouseStock'
import { useToast } from '@/composables/useToast'
import { usePdfExport } from '@/composables/usePdfExport'
import BaseCard from '@/components/UI/BaseCard.vue'
import BaseTable from '@/components/UI/BaseTable.vue'
import BaseBadge from '@/components/UI/BaseBadge.vue'
import BaseModal from '@/components/UI/BaseModal.vue'
import QuoteBuilderModern from '@/components/QuoteBuilderModern.vue'
import type { TableColumn } from '@/components/UI/BaseTable.vue'
import type { Quote, QuoteStatus, Currency } from '@/types'

const router = useRouter()
const route = useRoute()
const quotesStore = useQuotesStore()
const customersStore = useCustomersStore()
const productsStore = useProductsStore()
const laborStore = useLaborPositionsStore()
const recurringStore = useRecurringServicesStore()
const contactsStore = useContactsStore()
const opportunitiesStore = useOpportunitiesStore()
const priceBooksStore = usePriceBooksStore()
const contractsStore = useContractsStore()
const documentsStore = useDocumentsStore()
const warehouseStockStore = useWarehouseStockStore()
const { success, info, error } = useToast()
const { generateQuotePdf, generateContractPdf } = usePdfExport()

const searchQuery = ref('')
const statusFilter = ref<QuoteStatus | 'all'>('all')
const customerFilter = ref('all')
const opportunityFilter = ref('all')
const sortBy = ref('date-desc')
const showDetailsModal = ref(false)
const selectedQuote = ref<Quote | null>(null)
const showQuoteModal = ref(false)
const showDeleteModal = ref(false)
const editingQuote = ref<Quote | null>(null)
const quoteToDelete = ref<Quote | null>(null)
const showWorkflowModal = ref(false)
const showAddLineItemModal = ref(false)
const quoteFormLineItems = ref<any[]>([])
const editingLineItem = ref<any>(null)
const lineItemForm = ref({
  category: 'materials' as 'materials' | 'manpower' | 'miscellaneous',
  type: 'product' as 'product' | 'labor' | 'service',
  productId: '',
  productSku: '',
  productName: '',
  description: '',
  laborPosition: '',
  hoursPerMonth: 0,
  months: 0,
  quantity: 1,
  unitCost: 0,
  unitPrice: 0,
  discountPercent: 0,
  notes: ''
})

// Product and Labor selection
const selectedProductId = ref('')
const selectedLaborId = ref('')
const selectedItemId = ref('')
const selectedItemType = ref<'product' | 'service' | 'labor'>('product')
const catalogCost = ref(0)
const catalogPrice = ref(0)

// Computed for line item margin
const lineItemMarginPercent = computed(() => {
  const total = lineItemForm.value.quantity * lineItemForm.value.unitPrice * (1 - lineItemForm.value.discountPercent / 100)
  const cost = lineItemForm.value.quantity * lineItemForm.value.unitCost
  if (total === 0) return 0
  return ((total - cost) / total) * 100
})

const lineItemMarginAmount = computed(() => {
  const total = lineItemForm.value.quantity * lineItemForm.value.unitPrice * (1 - lineItemForm.value.discountPercent / 100)
  const cost = lineItemForm.value.quantity * lineItemForm.value.unitCost
  return total - cost
})

// Workflow stages for progress bar
const workflowStages = [
  { key: 'draft', label: 'Draft' },
  { key: 'pending_approval', label: 'Pending Approval' },
  { key: 'approved', label: 'Approved' },
  { key: 'sent', label: 'Sent' },
  { key: 'accepted', label: 'Accepted/Declined' }
]

// Quote form
const quoteForm = ref({
  opportunityId: '',
  companyId: '',
  contactId: '',
  priceBookId: '',
  quoteDate: new Date().toISOString().split('T')[0],
  validUntil: '',
  currency: 'SAR' as Currency,
  exchangeRate: 1.0,
  status: 'draft' as QuoteStatus,
  subtotal: 0,
  discountPercent: 0,
  discountAmount: 0,
  subtotalAfterDiscount: 0,
  vatPercent: 15,
  vatAmount: 0,
  total: 0,
  totalCost: 0,
  marginAmount: 0,
  marginPercent: 0,
  paymentTerms: '',
  deliveryTerms: '',
  customerNotes: '',
  internalNotes: '',
  requiresApproval: false,
  approvalReason: ''
})

// Quote attachments
const quoteFormAttachments = ref<any[]>([])

// Available contacts based on selected company
const availableContacts = computed(() => {
  if (!quoteForm.value.companyId) return []
  return contactsStore.contacts.filter(c => c.companyId === quoteForm.value.companyId)
})

// Available price books based on selected customer
const availablePriceBooks = computed(() => {
  if (!quoteForm.value.companyId) return []
  return priceBooksStore.getPriceBooksForCustomer(quoteForm.value.companyId)
})

// Available products - filtered by price book if selected
const availableProducts = computed(() => {
  const allProducts = productsStore.getActiveProducts()

  // If no price book selected, show all products
  if (!quoteForm.value.priceBookId) {
    return allProducts
  }

  // If price book selected, only show products in that price book
  const priceBook = priceBooksStore.getPriceBookById(quoteForm.value.priceBookId)
  if (!priceBook) return allProducts

  const priceBookProductIds = priceBook.priceBookEntries.map(entry => entry.productId)
  return allProducts.filter(p => priceBookProductIds.includes(p.id))
})

// Available labor positions - filtered by price book if needed
const availableLaborPositions = computed(() => {
  // For now, return all active labor positions
  // In future, can filter by price book if labor is included in price books
  return laborStore.getActivePositions()
})

// Unified available items - combines products, recurring services, and labor positions
const availableItems = computed(() => {
  const items: Array<{
    id: string
    type: 'product' | 'recurring' | 'labor'
    name: string
    sku?: string
    price: number
    priceLabel: string
    unitOfMeasure?: string
    department?: string
    icon: string
  }> = []

  // Add products from the products store (filtered by price book if selected)
  availableProducts.value.forEach(product => {
    items.push({
      id: product.id,
      type: 'product',
      name: product.name,
      sku: product.sku,
      price: product.sellingPrice,
      priceLabel: formatCurrency(product.sellingPrice),
      unitOfMeasure: product.unitOfMeasure,
      icon: '📦'
    })
  })

  // Add recurring services (only if no price book selected)
  if (!quoteForm.value.priceBookId) {
    recurringStore.getActiveRecurringServices().forEach(service => {
      items.push({
        id: service.id,
        type: 'recurring',
        name: service.name,
        sku: service.id,
        price: service.monthlyPrice,
        priceLabel: `${formatCurrency(service.monthlyPrice)}/month`,
        unitOfMeasure: 'Month',
        icon: '🔄'
      })
    })
  }

  // Add labor positions (only if no price book selected, or if we add labor to price books in future)
  if (!quoteForm.value.priceBookId) {
    availableLaborPositions.value.forEach(position => {
      items.push({
        id: position.id,
        type: 'labor',
        name: position.name,
        price: position.sellingRatePerHour,
        priceLabel: `${formatCurrency(position.sellingRatePerHour)}/hr`,
        department: position.department,
        icon: '👷'
      })
    })
  }

  return items
})

// Grouped items for the dropdown
const groupedItems = computed(() => {
  const groups: Record<string, typeof availableItems.value> = {
    'Products': [],
    'Recurring Services': [],
    'Labor Positions': []
  }

  availableItems.value.forEach(item => {
    if (item.type === 'product') {
      groups['Products'].push(item)
    } else if (item.type === 'recurring') {
      groups['Recurring Services'].push(item)
    } else if (item.type === 'labor') {
      groups['Labor Positions'].push(item)
    }
  })

  // Remove empty groups
  Object.keys(groups).forEach(key => {
    if (groups[key].length === 0) {
      delete groups[key]
    }
  })

  return groups
})

// Table columns
const quotesColumns: TableColumn[] = [
  { key: 'quoteNumber', label: 'Quote #', sortable: true },
  { key: 'customer', label: 'Customer & Opportunity', sortable: false },
  { key: 'lineItems', label: 'Line Items', sortable: false },
  { key: 'dates', label: 'Dates', sortable: false },
  { key: 'financial', label: 'Amount & Margin', sortable: false },
  { key: 'status', label: 'Status', sortable: true },
  { key: 'approval', label: 'Approval', sortable: false },
  { key: 'actions', label: 'Actions', align: 'right' }
]

// Filtered quotes
const filteredQuotes = computed(() => {
  let quotes = quotesStore.quotes

  // Filter by status
  if (statusFilter.value !== 'all') {
    quotes = quotes.filter(q => q.status === statusFilter.value)
  }

  // Filter by customer
  if (customerFilter.value !== 'all') {
    quotes = quotes.filter(q => q.companyId === customerFilter.value)
  }

  // Filter by opportunity
  if (opportunityFilter.value !== 'all') {
    if (opportunityFilter.value === 'no_opportunity') {
      quotes = quotes.filter(q => !q.opportunityId || q.opportunityId === '')
    } else {
      quotes = quotes.filter(q => q.opportunityId === opportunityFilter.value)
    }
  }

  // Filter by search query
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    quotes = quotes.filter(q =>
      q.quoteNumber.toLowerCase().includes(query) ||
      getCustomerName(q.companyId).toLowerCase().includes(query) ||
      getOpportunityName(q.opportunityId).toLowerCase().includes(query)
    )
  }

  return quotes
})

// Sorted quotes
const sortedQuotes = computed(() => {
  const quotes = [...filteredQuotes.value]

  switch (sortBy.value) {
    case 'date-desc':
      return quotes.sort((a, b) => new Date(b.quoteDate).getTime() - new Date(a.quoteDate).getTime())
    case 'date-asc':
      return quotes.sort((a, b) => new Date(a.quoteDate).getTime() - new Date(b.quoteDate).getTime())
    case 'value-desc':
      return quotes.sort((a, b) => b.total - a.total)
    case 'value-asc':
      return quotes.sort((a, b) => a.total - b.total)
    default:
      return quotes
  }
})

// Total value
const totalValue = computed(() =>
  filteredQuotes.value.reduce((sum, q) => sum + q.total, 0)
)

// Helper functions
function getCustomerName(companyId: string) {
  const customer = customersStore.getCustomerById(companyId)
  return customer?.name || 'Unknown Customer'
}

function getContactName(contactId: string) {
  const contact = contactsStore.getContactById(contactId)
  return contact?.fullName || 'Unknown Contact'
}

function getOpportunityName(opportunityId: string) {
  const opportunity = opportunitiesStore.getOpportunityById(opportunityId)
  return opportunity?.name || 'Unknown Opportunity'
}

function getQuoteLineItems(quoteId: string) {
  return quotesStore.getLineItemsByQuoteId(quoteId)
}

function formatCurrency(amount: number) {
  return new Intl.NumberFormat('en-SA', {
    style: 'currency',
    currency: 'SAR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(amount)
}

function formatDate(dateString: string) {
  return new Date(dateString).toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric'
  })
}

function formatStatus(status: QuoteStatus) {
  return status.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())
}

function getStatusVariant(status: QuoteStatus): 'success' | 'info' | 'default' | 'primary' | 'warning' | 'danger' {
  const variants: Record<QuoteStatus, 'success' | 'info' | 'default' | 'primary' | 'warning' | 'danger'> = {
    draft: 'default',
    pending_approval: 'warning',
    approved: 'info',
    sent: 'info',
    accepted: 'success',
    declined: 'danger',
    expired: 'danger'
  }
  return variants[status] || 'default'
}

function getMarginClass(marginPercent: number) {
  if (marginPercent < 10) return 'text-danger'
  if (marginPercent < 15) return 'text-warning'
  return 'text-success'
}

function isExpiringSoon(validUntil: string) {
  const daysUntilExpiry = Math.ceil(
    (new Date(validUntil).getTime() - new Date().getTime()) / (1000 * 60 * 60 * 24)
  )
  return daysUntilExpiry <= 7 && daysUntilExpiry >= 0
}

function canEdit(quote: Quote) {
  return quote.status === 'draft' || quote.status === 'pending_approval'
}

// Workflow Progress Bar Functions
function getStepClass(stageKey: string, currentStatus: QuoteStatus) {
  const stageOrder = ['draft', 'pending_approval', 'approved', 'sent', 'accepted', 'declined']
  const currentIndex = stageOrder.indexOf(currentStatus)
  const stepIndex = stageOrder.indexOf(stageKey)

  if (currentStatus === 'declined' && stageKey === 'accepted') {
    return 'declined'
  }

  if (stepIndex < currentIndex) {
    return 'completed'
  } else if (stepIndex === currentIndex) {
    return 'active'
  }
  return ''
}

function isStepCompleted(stageKey: string, currentStatus: QuoteStatus) {
  const stageOrder = ['draft', 'pending_approval', 'approved', 'sent', 'accepted', 'declined']
  const currentIndex = stageOrder.indexOf(currentStatus)
  const stepIndex = stageOrder.indexOf(stageKey)

  return stepIndex < currentIndex
}

// Line Item Computed Properties (for quote form)
const getMaterialsItems = computed(() =>
  quoteFormLineItems.value.filter(item => item.category === 'materials')
)

const getManpowerItems = computed(() =>
  quoteFormLineItems.value.filter(item => item.category === 'manpower')
)

const getMiscellaneousItems = computed(() =>
  quoteFormLineItems.value.filter(item => item.category === 'miscellaneous')
)

const getLineItemsGrandTotal = computed(() =>
  quoteFormLineItems.value.reduce((sum, item) => sum + item.lineTotal, 0)
)

// Line Item Functions
function getCategoryTotal(category: string) {
  return quoteFormLineItems.value
    .filter(item => item.category === category)
    .reduce((sum, item) => sum + item.lineTotal, 0)
}

function addLineItem() {
  editingLineItem.value = null
  resetLineItemForm()
  showAddLineItemModal.value = true
}

function editLineItem(item: any) {
  editingLineItem.value = item
  // Populate form with item data
  lineItemForm.value = {
    category: item.category,
    type: item.type,
    productId: item.productId || '',
    productSku: item.productSku || '',
    productName: item.productName,
    description: item.description,
    laborPosition: item.laborPosition || '',
    hoursPerMonth: item.hoursPerMonth || 0,
    months: item.months || 0,
    quantity: item.quantity,
    unitCost: item.unitCost,
    unitPrice: item.unitPrice,
    discountPercent: item.discountPercent,
    notes: item.notes || ''
  }
  showAddLineItemModal.value = true
}

function removeLineItem(tempId: string) {
  const index = quoteFormLineItems.value.findIndex(item => item.tempId === tempId)
  if (index !== -1) {
    quoteFormLineItems.value.splice(index, 1)
    syncLineItemsToQuoteForm()
  }
}

function saveLineItem() {
  const lineItem = calculateLineItemTotals()

  if (editingLineItem.value) {
    // Update existing item
    const index = quoteFormLineItems.value.findIndex(item => item.tempId === editingLineItem.value.tempId)
    if (index !== -1) {
      quoteFormLineItems.value[index] = lineItem
    }
  } else {
    // Add new item
    lineItem.tempId = `temp-${Date.now()}`
    quoteFormLineItems.value.push(lineItem)
  }

  syncLineItemsToQuoteForm()
  closeLineItemModal()
}

function calculateLineItemTotals() {
  const quantity = lineItemForm.value.quantity
  const unitPrice = lineItemForm.value.unitPrice
  const unitCost = lineItemForm.value.unitCost
  const discountPercent = lineItemForm.value.discountPercent

  const subtotal = quantity * unitPrice
  const discountAmount = (subtotal * discountPercent) / 100
  const lineTotal = subtotal - discountAmount
  const totalCost = quantity * unitCost
  const marginAmount = lineTotal - totalCost
  const marginPercent = lineTotal > 0 ? (marginAmount / lineTotal) * 100 : 0

  return {
    ...lineItemForm.value,
    tempId: editingLineItem.value?.tempId || '',
    quantity,
    unitPrice,
    unitCost,
    discountPercent,
    discountAmount: Number(discountAmount.toFixed(2)),
    lineTotal: Number(lineTotal.toFixed(2)),
    totalCost: Number(totalCost.toFixed(2)),
    marginAmount: Number(marginAmount.toFixed(2)),
    marginPercent: Number(marginPercent.toFixed(2))
  }
}

function syncLineItemsToQuoteForm() {
  // Calculate subtotal from all line items
  const subtotal = quoteFormLineItems.value.reduce((sum, item) => sum + item.lineTotal, 0)
  const totalCost = quoteFormLineItems.value.reduce((sum, item) => sum + item.totalCost, 0)

  quoteForm.value.subtotal = Number(subtotal.toFixed(2))
  quoteForm.value.totalCost = Number(totalCost.toFixed(2))

  // Recalculate all totals
  calculateTotals()
}

function resetLineItemForm() {
  lineItemForm.value = {
    category: 'materials',
    type: 'product',
    productId: '',
    productSku: '',
    productName: '',
    description: '',
    laborPosition: '',
    hoursPerMonth: 0,
    months: 0,
    quantity: 1,
    unitCost: 0,
    unitPrice: 0,
    discountPercent: 0,
    notes: ''
  }
  selectedProductId.value = ''
  selectedLaborId.value = ''
  catalogCost.value = 0
  catalogPrice.value = 0
}

function closeLineItemModal() {
  showAddLineItemModal.value = false
  editingLineItem.value = null
  resetLineItemForm()
}

// Product and Labor selection handlers
function onCategoryChange() {
  // Reset selections when category changes
  selectedItemId.value = ''
  selectedProductId.value = ''
  selectedLaborId.value = ''
  lineItemForm.value.productId = ''
  lineItemForm.value.productName = ''
  lineItemForm.value.productSku = ''
  lineItemForm.value.description = ''
  lineItemForm.value.unitCost = 0
  lineItemForm.value.unitPrice = 0
  catalogCost.value = 0
  catalogPrice.value = 0

  // Auto-set type based on category (for old system compatibility)
  if (lineItemForm.value.category === 'materials') {
    lineItemForm.value.type = 'product'
  } else if (lineItemForm.value.category === 'manpower') {
    lineItemForm.value.type = 'labor'
  } else if (lineItemForm.value.category === 'recurring') {
    lineItemForm.value.type = 'service'
  } else {
    lineItemForm.value.type = 'service'
  }
}

function onTypeChange() {
  // Reset selections when type changes
  selectedProductId.value = ''
  selectedLaborId.value = ''
}

function onProductSelect() {
  if (!selectedProductId.value) return

  const product = productsStore.getProductById(selectedProductId.value)
  if (!product) return

  // Check if this is a service (manpower) based on unit of measure
  const isService = product.unitOfMeasure === 'Month'

  // Auto-populate form fields
  lineItemForm.value.productId = product.id
  lineItemForm.value.productSku = product.sku
  lineItemForm.value.productName = product.name
  lineItemForm.value.description = product.description
  lineItemForm.value.unitCost = product.landedCostSAR

  // For services, set default duration instead of quantity
  if (isService) {
    lineItemForm.value.months = lineItemForm.value.months || 1
    // Set quantity to match duration for calculation purposes
    lineItemForm.value.quantity = lineItemForm.value.months
    lineItemForm.value.category = 'manpower'
    lineItemForm.value.type = 'service'
    info('This is a recurring service - specify duration in months')
  } else {
    // For products, use normal quantity
    if (!lineItemForm.value.quantity || lineItemForm.value.quantity === 0) {
      lineItemForm.value.quantity = 1
    }
  }

  // Use price book custom pricing if selected
  let unitPrice = product.sellingPrice
  let priceSource = 'Standard pricing'

  if (quoteForm.value.priceBookId) {
    const customPrice = priceBooksStore.getProductPrice(
      product.id,
      quoteForm.value.priceBookId,
      lineItemForm.value.quantity
    )
    if (customPrice > 0) {
      unitPrice = customPrice
      const priceBook = priceBooksStore.getPriceBookById(quoteForm.value.priceBookId)
      priceSource = `Price Book: ${priceBook?.name}`

      // Calculate and show discount
      const discount = product.sellingPrice - customPrice
      const discountPercent = (discount / product.sellingPrice) * 100
      if (discount > 0) {
        info(`Custom pricing applied: ${formatCurrency(customPrice)} (${discountPercent.toFixed(1)}% discount from standard)`)
      }
    }
  }

  lineItemForm.value.unitPrice = unitPrice
  catalogCost.value = product.landedCostSAR
  catalogPrice.value = product.sellingPrice

  success(`${isService ? 'Service' : 'Product'} loaded: ${product.name} - ${priceSource}`)
}

function onLaborSelect() {
  if (!selectedLaborId.value) return

  const position = laborStore.getPositionById(selectedLaborId.value)
  if (!position) return

  // Auto-populate form fields
  lineItemForm.value.productId = position.id
  lineItemForm.value.productName = position.name
  lineItemForm.value.description = position.description
  lineItemForm.value.laborPosition = position.name
  lineItemForm.value.unitCost = position.costPerHour
  lineItemForm.value.unitPrice = position.sellingRatePerHour
  lineItemForm.value.hoursPerMonth = 160 // Default
  lineItemForm.value.months = 1 // Default
  lineItemForm.value.quantity = 160 // 160 hours

  catalogCost.value = position.costPerHour
  catalogPrice.value = position.sellingRatePerHour

  success(`Labor position loaded: ${position.name}`)
}

function calculateLaborQuantity() {
  if (lineItemForm.value.type === 'labor') {
    const hours = lineItemForm.value.hoursPerMonth || 0
    const months = lineItemForm.value.months || 0
    lineItemForm.value.quantity = hours * months
  }
}

// Two-dropdown helper functions
function getCategoryLabel(): string {
  switch (lineItemForm.value.category) {
    case 'product':
      return 'Product'
    case 'recurring':
      return 'Recurring Service'
    case 'labor':
      return 'Labor Position'
    default:
      return 'Item'
  }
}

function getFilteredItems() {
  return availableItems.value.filter(item => item.type === lineItemForm.value.category)
}

// Unified item selection handler
function onItemSelect() {
  if (!selectedItemId.value) return

  // Parse the selection: format is "type:id"
  const [type, id] = selectedItemId.value.split(':')
  selectedItemType.value = type as 'product' | 'recurring' | 'labor'

  if (type === 'recurring') {
    // Handle recurring service
    const service = recurringStore.getRecurringServiceById(id)
    if (!service) return

    // Auto-populate form fields for recurring service
    lineItemForm.value.productId = service.id
    lineItemForm.value.productSku = service.id
    lineItemForm.value.productName = service.name
    lineItemForm.value.description = service.description
    lineItemForm.value.unitCost = service.monthlyCost
    lineItemForm.value.unitPrice = service.monthlyPrice
    lineItemForm.value.quantity = service.minimumMonths || 1
    lineItemForm.value.category = 'manpower'
    lineItemForm.value.type = 'service'

    catalogCost.value = service.monthlyCost
    catalogPrice.value = service.monthlyPrice

    success(`Recurring service loaded: ${service.name} - ${formatCurrency(service.monthlyPrice)}/month`)
  } else if (type === 'labor') {
    // Handle labor position
    const position = laborStore.getPositionById(id)
    if (!position) return

    // Auto-populate form fields for labor
    lineItemForm.value.productId = position.id
    lineItemForm.value.productName = position.name
    lineItemForm.value.description = position.description
    lineItemForm.value.laborPosition = position.name
    lineItemForm.value.unitCost = position.costPerHour
    lineItemForm.value.unitPrice = position.sellingRatePerHour
    lineItemForm.value.hoursPerMonth = 160 // Default
    lineItemForm.value.months = 1 // Default
    lineItemForm.value.quantity = 160 // 160 hours
    lineItemForm.value.category = 'manpower'
    lineItemForm.value.type = 'labor'

    catalogCost.value = position.costPerHour
    catalogPrice.value = position.sellingRatePerHour

    success(`Labor position loaded: ${position.name}`)
  } else {
    // Handle product or service
    const product = productsStore.getProductById(id)
    if (!product) return

    // Check if this is a service (manpower) based on unit of measure
    const isService = product.unitOfMeasure === 'Month'

    // Auto-populate form fields
    lineItemForm.value.productId = product.id
    lineItemForm.value.productSku = product.sku
    lineItemForm.value.productName = product.name
    lineItemForm.value.description = product.description
    lineItemForm.value.unitCost = product.landedCostSAR

    // For services, set default duration instead of quantity
    if (isService) {
      lineItemForm.value.months = lineItemForm.value.months || 1
      lineItemForm.value.quantity = lineItemForm.value.months
      lineItemForm.value.category = 'manpower'
      lineItemForm.value.type = 'service'
      info('This is a recurring service - specify duration in months')
    } else {
      // For products, use normal quantity
      if (!lineItemForm.value.quantity || lineItemForm.value.quantity === 0) {
        lineItemForm.value.quantity = 1
      }
      lineItemForm.value.category = 'materials'
      lineItemForm.value.type = 'product'
    }

    // Use price book custom pricing if selected
    let unitPrice = product.sellingPrice
    let priceSource = 'Standard pricing'

    if (quoteForm.value.priceBookId) {
      const customPrice = priceBooksStore.getProductPrice(
        product.id,
        quoteForm.value.priceBookId,
        lineItemForm.value.quantity
      )
      if (customPrice > 0) {
        unitPrice = customPrice
        const priceBook = priceBooksStore.getPriceBookById(quoteForm.value.priceBookId)
        priceSource = `Price Book: ${priceBook?.name}`

        // Calculate and show discount
        const discount = product.sellingPrice - customPrice
        const discountPercent = (discount / product.sellingPrice) * 100
        if (discount > 0) {
          info(`Custom pricing applied: ${formatCurrency(customPrice)} (${discountPercent.toFixed(1)}% discount from standard)`)
        }
      }
    }

    lineItemForm.value.unitPrice = unitPrice
    catalogCost.value = product.landedCostSAR
    catalogPrice.value = product.sellingPrice

    success(`${isService ? 'Service' : 'Product'} loaded: ${product.name} - ${priceSource}`)
  }
}

// Get categorized line items (for viewing saved quotes)
function getCategorizedLineItems(quoteId: string) {
  const items = quotesStore.getLineItemsByQuoteId(quoteId)
  return {
    materials: items.filter(item => item.category === 'materials'),
    manpower: items.filter(item => item.category === 'manpower'),
    miscellaneous: items.filter(item => item.category === 'miscellaneous')
  }
}

function getCategorySubtotal(items: any[]) {
  return items.reduce((sum, item) => sum + item.lineTotal, 0)
}

// Get line item summary for quote list view
function getLineItemsSummary(quoteId: string) {
  const items = quotesStore.getLineItemsByQuoteId(quoteId)
  const materialsCount = items.filter(item => item.category === 'materials').length
  const manpowerCount = items.filter(item => item.category === 'manpower').length
  const miscCount = items.filter(item => item.category === 'miscellaneous').length

  return {
    total: items.length,
    materials: materialsCount,
    manpower: manpowerCount,
    miscellaneous: miscCount
  }
}

// Calculation Functions
function calculateTotals() {
  const subtotal = quoteForm.value.subtotal
  const discountAmount = (subtotal * quoteForm.value.discountPercent) / 100
  const subtotalAfterDiscount = subtotal - discountAmount
  const vatAmount = (subtotalAfterDiscount * quoteForm.value.vatPercent) / 100
  const total = subtotalAfterDiscount + vatAmount

  quoteForm.value.discountAmount = Number(discountAmount.toFixed(2))
  quoteForm.value.subtotalAfterDiscount = Number(subtotalAfterDiscount.toFixed(2))
  quoteForm.value.vatAmount = Number(vatAmount.toFixed(2))
  quoteForm.value.total = Number(total.toFixed(2))

  calculateMargin()
}

function calculateMargin() {
  const marginAmount = quoteForm.value.subtotalAfterDiscount - quoteForm.value.totalCost
  const marginPercent = quoteForm.value.subtotalAfterDiscount > 0
    ? (marginAmount / quoteForm.value.subtotalAfterDiscount) * 100
    : 0

  quoteForm.value.marginAmount = Number(marginAmount.toFixed(2))
  quoteForm.value.marginPercent = Number(marginPercent.toFixed(2))
}

function updateContactsList() {
  // Reset contact when company changes
  quoteForm.value.contactId = ''
}

function onOpportunityChange() {
  const selected = quoteForm.value.opportunityId

  if (!selected) {
    quoteForm.value.companyId = ''
    quoteForm.value.contactId = ''
    quoteForm.value.priceBookId = ''
    return
  }

  // Check if it's a direct customer (starts with DIRECT-)
  if (selected.startsWith('DIRECT-')) {
    const customerId = selected.replace('DIRECT-', '')
    quoteForm.value.opportunityId = '' // Clear opportunity since it's direct
    quoteForm.value.companyId = customerId
    quoteForm.value.contactId = ''

    // Auto-populate price book for direct customer
    autoPopulatePriceBook(customerId)

    info('Quote for direct customer selected')
  } else {
    // It's an opportunity
    const opportunity = opportunitiesStore.getOpportunityById(selected)
    if (opportunity) {
      quoteForm.value.companyId = opportunity.companyId
      if (opportunity.contactId) {
        quoteForm.value.contactId = opportunity.contactId
      }

      // Auto-populate price book based on opportunity's contract or customer
      autoPopulatePriceBook(opportunity.companyId, opportunity.contractId)

      info('Quote linked to opportunity - customer auto-populated')
    }
  }
}

// Auto-populate price book based on customer and optional contract
function autoPopulatePriceBook(customerId: string, contractId?: string) {
  quoteForm.value.priceBookId = ''

  // First, try to find price book linked to contract if provided
  if (contractId) {
    const contract = contractsStore.getContractById(contractId)
    if (contract?.priceBookId) {
      const priceBook = priceBooksStore.getPriceBookById(contract.priceBookId)
      if (priceBook && priceBook.status === 'active') {
        quoteForm.value.priceBookId = priceBook.id
        success(`Price book auto-selected from contract: ${priceBook.name}`)
        return
      }
    }
  }

  // Otherwise, get available price books for customer
  const customerPriceBooks = priceBooksStore.getPriceBooksForCustomer(customerId)

  if (customerPriceBooks.length > 0) {
    // Prioritize contract-specific price books
    const contractPriceBook = customerPriceBooks.find(pb => pb.type === 'contract')
    if (contractPriceBook) {
      quoteForm.value.priceBookId = contractPriceBook.id
      success(`Price book auto-selected: ${contractPriceBook.name}`)
      return
    }

    // Then customer-specific price books
    const customerSpecificPriceBook = customerPriceBooks.find(pb => pb.type === 'customer_specific')
    if (customerSpecificPriceBook) {
      quoteForm.value.priceBookId = customerSpecificPriceBook.id
      success(`Price book auto-selected: ${customerSpecificPriceBook.name}`)
      return
    }

    // Use first available price book
    quoteForm.value.priceBookId = customerPriceBooks[0].id
    info(`Price book available: ${customerPriceBooks[0].name}`)
  }
}

// Attachment Management Functions
function addNewAttachment() {
  const newAttachment = {
    id: `temp-${Date.now()}`,
    documentId: '',
    documentName: '',
    documentType: '' as any,
    includeInPdf: true,
    order: quoteFormAttachments.value.length
  }
  quoteFormAttachments.value.push(newAttachment)
}

function removeAttachment(index: number) {
  quoteFormAttachments.value.splice(index, 1)
  // Update order numbers
  quoteFormAttachments.value.forEach((att, idx) => {
    att.order = idx
  })
}

function moveAttachmentUp(index: number) {
  if (index === 0) return
  const temp = quoteFormAttachments.value[index]
  quoteFormAttachments.value[index] = quoteFormAttachments.value[index - 1]
  quoteFormAttachments.value[index - 1] = temp
  // Update order numbers
  quoteFormAttachments.value.forEach((att, idx) => {
    att.order = idx
  })
}

function moveAttachmentDown(index: number) {
  if (index === quoteFormAttachments.value.length - 1) return
  const temp = quoteFormAttachments.value[index]
  quoteFormAttachments.value[index] = quoteFormAttachments.value[index + 1]
  quoteFormAttachments.value[index + 1] = temp
  // Update order numbers
  quoteFormAttachments.value.forEach((att, idx) => {
    att.order = idx
  })
}

function onAttachmentDocumentChange(attachment: any) {
  // When document is selected, auto-populate the name and type
  if (attachment.documentId) {
    const doc = documentsStore.getDocumentById(attachment.documentId)
    if (doc) {
      if (!attachment.documentName) {
        attachment.documentName = doc.name
      }
      attachment.documentType = doc.type
    }
  }
}

function formatDocumentType(type: string): string {
  const typeMap: Record<string, string> = {
    terms: 'Terms',
    delivery: 'Delivery',
    technical: 'Technical',
    warranty: 'Warranty',
    sla: 'SLA',
    other: 'Other'
  }
  return typeMap[type] || type
}

// Modal Functions
function resetQuoteForm() {
  quoteForm.value = {
    opportunityId: '',
    companyId: '',
    contactId: '',
    quoteDate: new Date().toISOString().split('T')[0],
    validUntil: '',
    currency: 'SAR' as Currency,
    exchangeRate: 1.0,
    status: 'draft' as QuoteStatus,
    subtotal: 0,
    discountPercent: 0,
    discountAmount: 0,
    subtotalAfterDiscount: 0,
    vatPercent: 15,
    vatAmount: 0,
    total: 0,
    totalCost: 0,
    marginAmount: 0,
    marginPercent: 0,
    paymentTerms: '',
    deliveryTerms: '',
    customerNotes: '',
    internalNotes: '',
    requiresApproval: false,
    approvalReason: ''
  }
  quoteFormLineItems.value = []
  quoteFormAttachments.value = []
  editingQuote.value = null
}

function closeQuoteModal() {
  showQuoteModal.value = false
  resetQuoteForm()
}

// CRUD Actions
function createNewQuote() {
  resetQuoteForm()
  // Set default valid until date (30 days from now)
  const validUntil = new Date()
  validUntil.setDate(validUntil.getDate() + 30)
  quoteForm.value.validUntil = validUntil.toISOString().split('T')[0]

  // Check if coming from opportunity
  const opportunityId = route.query.opportunityId as string
  if (opportunityId) {
    const opportunity = opportunitiesStore.getOpportunityById(opportunityId)
    if (opportunity) {
      quoteForm.value.opportunityId = opportunityId
      quoteForm.value.companyId = opportunity.companyId
      if (opportunity.contactId) {
        quoteForm.value.contactId = opportunity.contactId
      }
      info('Quote linked to opportunity and customer auto-populated')
    }
  }

  showQuoteModal.value = true
}

function viewQuoteDetails(quote: Quote) {
  selectedQuote.value = quote
  showDetailsModal.value = true
}

function closeDetailsModal() {
  showDetailsModal.value = false
  selectedQuote.value = null
}

function viewOpportunity(opportunityId: string) {
  // Navigate to leads view with the opportunity highlighted
  router.push('/leads')
  // Could also add query param to highlight specific opportunity
  // router.push({ path: '/leads', query: { opportunityId } })
}

function editQuote(quote: Quote) {
  if (showDetailsModal.value) {
    showDetailsModal.value = false
  }

  editingQuote.value = quote
  quoteForm.value = {
    opportunityId: quote.opportunityId,
    companyId: quote.companyId,
    contactId: quote.contactId,
    priceBookId: quote.priceBookId || '',
    quoteDate: quote.quoteDate,
    validUntil: quote.validUntil,
    currency: quote.currency,
    exchangeRate: quote.exchangeRate,
    status: quote.status,
    subtotal: quote.subtotal,
    discountPercent: quote.discountPercent,
    discountAmount: quote.discountAmount,
    subtotalAfterDiscount: quote.subtotalAfterDiscount,
    vatPercent: quote.vatPercent,
    vatAmount: quote.vatAmount,
    total: quote.total,
    totalCost: quote.totalCost,
    marginAmount: quote.marginAmount,
    marginPercent: quote.marginPercent,
    paymentTerms: quote.paymentTerms,
    deliveryTerms: quote.deliveryTerms,
    customerNotes: quote.customerNotes || '',
    internalNotes: quote.internalNotes || '',
    requiresApproval: quote.requiresApproval,
    approvalReason: quote.approvalReason || ''
  }

  // Load attachments
  quoteFormAttachments.value = quote.attachments ? [...quote.attachments] : []

  showQuoteModal.value = true
}

function saveQuote() {
  // Validation
  if (!quoteForm.value.opportunityId || !quoteForm.value.companyId ||
      !quoteForm.value.contactId || !quoteForm.value.quoteDate ||
      !quoteForm.value.validUntil || !quoteForm.value.paymentTerms ||
      !quoteForm.value.deliveryTerms) {
    info('Please fill in all required fields')
    return
  }

  const quoteData = {
    opportunityId: quoteForm.value.opportunityId,
    companyId: quoteForm.value.companyId,
    contactId: quoteForm.value.contactId,
    priceBookId: quoteForm.value.priceBookId || undefined,
    quoteDate: quoteForm.value.quoteDate,
    validUntil: quoteForm.value.validUntil,
    expiryDate: quoteForm.value.validUntil,
    currency: quoteForm.value.currency,
    exchangeRate: quoteForm.value.exchangeRate,
    status: quoteForm.value.status,
    subtotal: quoteForm.value.subtotal,
    discountPercent: quoteForm.value.discountPercent,
    discountAmount: quoteForm.value.discountAmount,
    subtotalAfterDiscount: quoteForm.value.subtotalAfterDiscount,
    vatPercent: quoteForm.value.vatPercent,
    vatAmount: quoteForm.value.vatAmount,
    total: quoteForm.value.total,
    totalCost: quoteForm.value.totalCost,
    marginAmount: quoteForm.value.marginAmount,
    marginPercent: quoteForm.value.marginPercent,
    paymentTerms: quoteForm.value.paymentTerms,
    deliveryTerms: quoteForm.value.deliveryTerms,
    customerNotes: quoteForm.value.customerNotes,
    internalNotes: quoteForm.value.internalNotes,
    requiresApproval: quoteForm.value.requiresApproval,
    approvalReason: quoteForm.value.approvalReason,
    attachments: quoteFormAttachments.value.filter(att => att.documentId) // Only include attachments with documents selected
  }

  if (editingQuote.value) {
    quotesStore.updateQuote(editingQuote.value.id, quoteData)
    success('Quote updated successfully')
  } else {
    const newQuote: Quote = {
      id: `QTE-${Date.now()}`,
      quoteNumber: `Q-${new Date().getFullYear()}-${String(quotesStore.quotes.length + 1).padStart(3, '0')}`,
      version: 1,
      ...quoteData,
      customerNotes: quoteData.customerNotes || undefined,
      internalNotes: quoteData.internalNotes || undefined,
      approvalReason: quoteData.approvalReason || undefined,
      createdBy: 'USR-001',
      createdAt: new Date().toISOString()
    }
    quotesStore.addQuote(newQuote)
    success('Quote created successfully')
  }

  closeQuoteModal()
}

function deleteQuote(quote: Quote) {
  quoteToDelete.value = quote
  showDeleteModal.value = true
}

function confirmDelete() {
  if (quoteToDelete.value) {
    quotesStore.deleteQuote(quoteToDelete.value.id)
    success(`Quote ${quoteToDelete.value.quoteNumber} deleted successfully`)
    quoteToDelete.value = null
  }
  showDeleteModal.value = false
}

function duplicateQuote(quote: Quote) {
  const clonedQuote = quotesStore.cloneQuote(quote.id)
  if (clonedQuote) {
    success(`Quote duplicated as ${clonedQuote.quoteNumber}`)
  }
}

function downloadPDF(quote: Quote) {
  const lineItems = quotesStore.getLineItemsByQuoteId(quote.id)
  const customerName = getCustomerName(quote.companyId)
  const contactName = getContactName(quote.contactId)
  const attachments = quote.attachments || []

  generateQuotePdf(quote, lineItems, customerName, contactName, attachments)
  success(`Generating PDF for quote ${quote.quoteNumber}`)
}

function printQuote(quote: Quote) {
  const printWindow = window.open('', '_blank')
  if (!printWindow) {
    error('Please allow popups to print the quote')
    return
  }

  const lineItems = quotesStore.getLineItemsByQuoteId(quote.id)
  const customerName = getCustomerName(quote.companyId)
  const contactName = getContactName(quote.contactId)

  const categorized = getCategorizedLineItems(quote.id)

  // Build HTML content using string concatenation to avoid nested template literal issues
  let htmlContent = '<!DOCTYPE html><html><head>'
  htmlContent += '<title>Quote ' + quote.quoteNumber + '</title>'
  htmlContent += '<style>'
  htmlContent += 'body { font-family: Arial, sans-serif; padding: 20px; color: #333; }'
  htmlContent += '.header { text-align: center; margin-bottom: 30px; border-bottom: 3px solid #3b82f6; padding-bottom: 20px; }'
  htmlContent += '.header h1 { color: #3b82f6; margin: 0; font-size: 28px; }'
  htmlContent += '.header p { margin: 5px 0; color: #666; }'
  htmlContent += '.section { margin: 25px 0; }'
  htmlContent += '.section-title { font-size: 18px; font-weight: bold; color: #3b82f6; margin-bottom: 15px; border-bottom: 2px solid #e5e7eb; padding-bottom: 8px; }'
  htmlContent += '.info-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 15px; margin-bottom: 20px; }'
  htmlContent += '.info-item { display: flex; }'
  htmlContent += '.info-label { font-weight: 600; width: 140px; color: #666; }'
  htmlContent += '.info-value { color: #111; }'
  htmlContent += 'table { width: 100%; border-collapse: collapse; margin: 15px 0; }'
  htmlContent += 'th { background: #f3f4f6; padding: 12px; text-align: left; font-weight: 600; border-bottom: 2px solid #d1d5db; }'
  htmlContent += 'td { padding: 10px 12px; border-bottom: 1px solid #e5e7eb; }'
  htmlContent += 'tr:last-child td { border-bottom: none; }'
  htmlContent += '.category-header { background: #eff6ff; font-weight: 600; color: #3b82f6; padding: 10px 12px; margin-top: 15px; }'
  htmlContent += '.subtotal-row { font-weight: 600; background: #f9fafb; }'
  htmlContent += '.financial-summary { background: #f9fafb; padding: 20px; border-radius: 8px; margin-top: 25px; }'
  htmlContent += '.financial-row { display: flex; justify-content: space-between; padding: 8px 0; }'
  htmlContent += '.financial-label { font-weight: 600; color: #666; }'
  htmlContent += '.financial-value { font-weight: 700; }'
  htmlContent += '.total-row { font-size: 20px; color: #3b82f6; border-top: 2px solid #3b82f6; padding-top: 12px; margin-top: 12px; }'
  htmlContent += '.margin-info { display: flex; gap: 30px; margin-top: 15px; padding-top: 15px; border-top: 1px solid #d1d5db; }'
  htmlContent += '.margin-item { flex: 1; }'
  htmlContent += '.terms { margin-top: 25px; padding: 20px; background: #fffbeb; border-left: 4px solid #f59e0b; }'
  htmlContent += '.terms h4 { margin: 0 0 10px 0; color: #f59e0b; }'
  htmlContent += '@media print { body { padding: 0; } .no-print { display: none; } }'
  htmlContent += '</style></head><body>'

  // Header
  htmlContent += '<div class="header">'
  htmlContent += '<h1>QUOTATION</h1>'
  htmlContent += '<p><strong>' + quote.quoteNumber + '</strong> (Version ' + quote.version + ')</p>'
  htmlContent += '<p>Date: ' + formatDate(quote.quoteDate) + ' | Valid Until: ' + formatDate(quote.validUntil) + '</p>'
  htmlContent += '</div>'

  // Customer Information
  htmlContent += '<div class="section">'
  htmlContent += '<div class="section-title">Customer Information</div>'
  htmlContent += '<div class="info-grid">'
  htmlContent += '<div class="info-item"><span class="info-label">Customer:</span><span class="info-value">' + customerName + '</span></div>'
  htmlContent += '<div class="info-item"><span class="info-label">Contact:</span><span class="info-value">' + contactName + '</span></div>'
  htmlContent += '<div class="info-item"><span class="info-label">Currency:</span><span class="info-value">' + quote.currency + '</span></div>'
  htmlContent += '<div class="info-item"><span class="info-label">Status:</span><span class="info-value">' + formatStatus(quote.status) + '</span></div>'
  htmlContent += '</div></div>'

  // Line Items
  htmlContent += '<div class="section"><div class="section-title">Line Items</div>'

  // Materials category
  if (categorized.materials.length > 0) {
    htmlContent += '<div class="category-header">📦 Materials & Equipment (' + categorized.materials.length + ')</div>'
    htmlContent += '<table><thead><tr>'
    htmlContent += '<th style="width: 40px">#</th><th>Item</th><th style="width: 80px">Qty</th>'
    htmlContent += '<th style="width: 100px">Unit Price</th><th style="width: 100px">Total</th>'
    htmlContent += '</tr></thead><tbody>'

    categorized.materials.forEach(item => {
      htmlContent += '<tr><td>' + item.lineNumber + '</td><td><strong>' + item.productName + '</strong>'
      if (item.productSku) {
        htmlContent += '<br><small style="color: #666;">' + item.productSku + '</small>'
      }
      if (item.description) {
        htmlContent += '<br><small style="color: #666;">' + item.description + '</small>'
      }
      htmlContent += '</td><td>' + item.quantity + '</td>'
      htmlContent += '<td>' + formatCurrency(item.unitPrice) + '</td>'
      htmlContent += '<td><strong>' + formatCurrency(item.lineTotal) + '</strong></td></tr>'
    })

    htmlContent += '<tr class="subtotal-row"><td colspan="4" style="text-align: right;">Subtotal:</td>'
    htmlContent += '<td><strong>' + formatCurrency(getCategorySubtotal(categorized.materials)) + '</strong></td></tr>'
    htmlContent += '</tbody></table>'
  }

  // Manpower category
  if (categorized.manpower.length > 0) {
    htmlContent += '<div class="category-header">👷 Manpower & Services (' + categorized.manpower.length + ')</div>'
    htmlContent += '<table><thead><tr>'
    htmlContent += '<th style="width: 40px">#</th><th>Service</th><th style="width: 100px">Duration</th>'
    htmlContent += '<th style="width: 100px">Rate</th><th style="width: 100px">Total</th>'
    htmlContent += '</tr></thead><tbody>'

    categorized.manpower.forEach(item => {
      htmlContent += '<tr><td>' + item.lineNumber + '</td><td><strong>' + item.productName + '</strong>'
      if (item.description) {
        htmlContent += '<br><small style="color: #666;">' + item.description + '</small>'
      }
      if (item.months && item.months > 0) {
        htmlContent += '<br><small style="color: #f59e0b;">🔄 ' + item.months + ' month' + (item.months > 1 ? 's' : '') + '</small>'
      }
      htmlContent += '</td>'
      htmlContent += '<td>' + (item.months && item.months > 0 ? item.months + ' mo' : item.quantity) + '</td>'
      htmlContent += '<td>' + formatCurrency(item.unitPrice) + (item.type === 'service' ? '/mo' : '') + '</td>'
      htmlContent += '<td><strong>' + formatCurrency(item.lineTotal) + '</strong></td></tr>'
    })

    htmlContent += '<tr class="subtotal-row"><td colspan="4" style="text-align: right;">Subtotal:</td>'
    htmlContent += '<td><strong>' + formatCurrency(getCategorySubtotal(categorized.manpower)) + '</strong></td></tr>'
    htmlContent += '</tbody></table>'
  }

  // Miscellaneous category
  if (categorized.miscellaneous.length > 0) {
    htmlContent += '<div class="category-header">📋 Miscellaneous (' + categorized.miscellaneous.length + ')</div>'
    htmlContent += '<table><thead><tr>'
    htmlContent += '<th style="width: 40px">#</th><th>Item</th><th style="width: 80px">Qty</th>'
    htmlContent += '<th style="width: 100px">Unit Price</th><th style="width: 100px">Total</th>'
    htmlContent += '</tr></thead><tbody>'

    categorized.miscellaneous.forEach(item => {
      htmlContent += '<tr><td>' + item.lineNumber + '</td><td><strong>' + item.productName + '</strong>'
      if (item.description) {
        htmlContent += '<br><small style="color: #666;">' + item.description + '</small>'
      }
      htmlContent += '</td><td>' + item.quantity + '</td>'
      htmlContent += '<td>' + formatCurrency(item.unitPrice) + '</td>'
      htmlContent += '<td><strong>' + formatCurrency(item.lineTotal) + '</strong></td></tr>'
    })

    htmlContent += '<tr class="subtotal-row"><td colspan="4" style="text-align: right;">Subtotal:</td>'
    htmlContent += '<td><strong>' + formatCurrency(getCategorySubtotal(categorized.miscellaneous)) + '</strong></td></tr>'
    htmlContent += '</tbody></table>'
  }

  htmlContent += '</div>'

  // Financial Summary
  htmlContent += '<div class="financial-summary">'
  htmlContent += '<div class="financial-row"><span class="financial-label">Subtotal:</span>'
  htmlContent += '<span class="financial-value">' + formatCurrency(quote.subtotal) + '</span></div>'

  if (quote.discountPercent > 0) {
    htmlContent += '<div class="financial-row" style="color: #ef4444;">'
    htmlContent += '<span class="financial-label">Discount (' + quote.discountPercent + '%):</span>'
    htmlContent += '<span class="financial-value">-' + formatCurrency(quote.discountAmount) + '</span></div>'
    htmlContent += '<div class="financial-row"><span class="financial-label">After Discount:</span>'
    htmlContent += '<span class="financial-value">' + formatCurrency(quote.subtotalAfterDiscount) + '</span></div>'
  }

  htmlContent += '<div class="financial-row"><span class="financial-label">VAT (' + quote.vatPercent + '%):</span>'
  htmlContent += '<span class="financial-value">' + formatCurrency(quote.vatAmount) + '</span></div>'
  htmlContent += '<div class="financial-row total-row"><span class="financial-label">Grand Total:</span>'
  htmlContent += '<span class="financial-value">' + formatCurrency(quote.total) + '</span></div>'

  // Margin Info
  htmlContent += '<div class="margin-info">'
  htmlContent += '<div class="margin-item"><div class="financial-label">Total Cost:</div>'
  htmlContent += '<div class="financial-value" style="color: #ef4444;">' + formatCurrency(quote.totalCost) + '</div></div>'
  htmlContent += '<div class="margin-item"><div class="financial-label">Profit:</div>'
  htmlContent += '<div class="financial-value" style="color: #10b981;">' + formatCurrency(quote.marginAmount) + '</div></div>'
  htmlContent += '<div class="margin-item"><div class="financial-label">Margin:</div>'
  htmlContent += '<div class="financial-value" style="color: ' + (quote.marginPercent >= 20 ? '#10b981' : '#f59e0b') + ';">'
  htmlContent += quote.marginPercent.toFixed(1) + '%</div></div>'
  htmlContent += '</div></div>'

  // Terms & Conditions
  htmlContent += '<div class="section"><div class="section-title">Terms & Conditions</div><div class="terms">'
  htmlContent += '<h4>💳 Payment Terms</h4><p>' + quote.paymentTerms + '</p>'
  htmlContent += '<h4 style="margin-top: 15px;">🚚 Delivery Terms</h4><p>' + quote.deliveryTerms + '</p>'

  if (quote.customerNotes) {
    htmlContent += '<h4 style="margin-top: 15px;">📝 Notes</h4><p>' + quote.customerNotes + '</p>'
  }

  htmlContent += '</div></div>'

  // Footer
  htmlContent += '<div style="margin-top: 40px; text-align: center; color: #666; font-size: 14px;">'
  htmlContent += '<p>Thank you for your business!</p></div>'

  // Auto-print script
  htmlContent += '<scr' + 'ipt>window.onload = function() { window.print(); }</scr' + 'ipt>'
  htmlContent += '</body></html>'

  printWindow.document.write(htmlContent)
  printWindow.document.close()
  success('Opening print dialog...')
}

function exportQuoteToExcel(quote: Quote) {
  const lineItems = quotesStore.getLineItemsByQuoteId(quote.id)
  const customerName = getCustomerName(quote.companyId)
  const contactName = getContactName(quote.contactId)
  const categorized = getCategorizedLineItems(quote.id)

  // Create CSV content
  let csvContent = 'data:text/csv;charset=utf-8,'

  // Header information
  csvContent += 'QUOTATION\n'
  csvContent += 'Quote Number,' + quote.quoteNumber + '\n'
  csvContent += 'Version,' + quote.version + '\n'
  csvContent += 'Date,' + formatDate(quote.quoteDate) + '\n'
  csvContent += 'Valid Until,' + formatDate(quote.validUntil) + '\n'
  csvContent += 'Customer,' + customerName + '\n'
  csvContent += 'Contact,' + contactName + '\n'
  csvContent += 'Currency,' + quote.currency + '\n'
  csvContent += 'Status,' + formatStatus(quote.status) + '\n'
  csvContent += '\n'

  // Materials & Equipment Section
  if (categorized.materials.length > 0) {
    csvContent += 'MATERIALS & EQUIPMENT (' + categorized.materials.length + ' items)\n'
    csvContent += '#,SKU,Product Name,Description,Manufacturer,Stock Available,Lead Time (days),Qty,Unit Price,Discount %,Line Total,Unit Cost,Margin Amount,Margin %\n'

    categorized.materials.forEach((item) => {
      // Get product details for manufacturer and inventory info
      const product = productsStore.getProductById(item.productId || '')
      const manufacturer = product?.manufacturer || ''
      const stockAvailable = product ? warehouseStockStore.getTotalAvailableForProduct(product.id) : ''
      const leadTime = product?.leadTimeDays || ''

      const row = [
        item.lineNumber,
        '"' + (item.productSku || '') + '"',
        '"' + (item.productName || '') + '"',
        '"' + (item.description || '') + '"',
        '"' + manufacturer + '"',
        stockAvailable,
        leadTime,
        item.quantity,
        item.unitPrice.toFixed(2),
        (item.discountPercent || 0).toFixed(2),
        item.lineTotal.toFixed(2),
        item.unitCost.toFixed(2),
        item.marginAmount.toFixed(2),
        item.marginPercent.toFixed(2)
      ]
      csvContent += row.join(',') + '\n'
    })

    csvContent += 'Subtotal,' + formatCurrency(getCategorySubtotal(categorized.materials)) + '\n'
    csvContent += '\n'
  }

  // Manpower & Services Section
  if (categorized.manpower.length > 0) {
    csvContent += 'MANPOWER & SERVICES (' + categorized.manpower.length + ' items)\n'
    csvContent += '#,SKU,Service Name,Description,Type,Duration/Qty,Unit Price,Rate Type,Discount %,Line Total,Unit Cost,Margin Amount,Margin %\n'

    categorized.manpower.forEach((item) => {
      const rateType = item.type === 'service' ? '/month' : (item.type === 'labor' ? '/hour' : '')
      const duration = item.months && item.months > 0 ? item.months + ' months' : item.quantity

      const row = [
        item.lineNumber,
        '"' + (item.productSku || '') + '"',
        '"' + (item.productName || '') + '"',
        '"' + (item.description || '') + '"',
        item.type,
        duration,
        item.unitPrice.toFixed(2),
        rateType,
        (item.discountPercent || 0).toFixed(2),
        item.lineTotal.toFixed(2),
        item.unitCost.toFixed(2),
        item.marginAmount.toFixed(2),
        item.marginPercent.toFixed(2)
      ]
      csvContent += row.join(',') + '\n'
    })

    csvContent += 'Subtotal,' + formatCurrency(getCategorySubtotal(categorized.manpower)) + '\n'
    csvContent += '\n'
  }

  // Miscellaneous Section
  if (categorized.miscellaneous.length > 0) {
    csvContent += 'MISCELLANEOUS (' + categorized.miscellaneous.length + ' items)\n'
    csvContent += '#,Item Name,Description,Qty,Unit Price,Discount %,Line Total,Unit Cost,Margin Amount,Margin %\n'

    categorized.miscellaneous.forEach((item) => {
      const row = [
        item.lineNumber,
        '"' + (item.productName || '') + '"',
        '"' + (item.description || '') + '"',
        item.quantity,
        item.unitPrice.toFixed(2),
        (item.discountPercent || 0).toFixed(2),
        item.lineTotal.toFixed(2),
        item.unitCost.toFixed(2),
        item.marginAmount.toFixed(2),
        item.marginPercent.toFixed(2)
      ]
      csvContent += row.join(',') + '\n'
    })

    csvContent += 'Subtotal,' + formatCurrency(getCategorySubtotal(categorized.miscellaneous)) + '\n'
    csvContent += '\n'
  }

  // Financial summary
  csvContent += 'FINANCIAL SUMMARY\n'
  csvContent += 'Subtotal,' + quote.subtotal.toFixed(2) + '\n'
  if (quote.discountPercent > 0) {
    csvContent += 'Discount (' + quote.discountPercent + '%),-' + quote.discountAmount.toFixed(2) + '\n'
    csvContent += 'After Discount,' + quote.subtotalAfterDiscount.toFixed(2) + '\n'
  }
  csvContent += 'VAT (' + quote.vatPercent + '%),' + quote.vatAmount.toFixed(2) + '\n'
  csvContent += 'Grand Total,' + quote.total.toFixed(2) + '\n'
  csvContent += '\n'
  csvContent += 'PROFITABILITY ANALYSIS\n'
  csvContent += 'Total Cost,' + quote.totalCost.toFixed(2) + '\n'
  csvContent += 'Profit,' + quote.marginAmount.toFixed(2) + '\n'
  csvContent += 'Margin %,' + quote.marginPercent.toFixed(2) + '%\n'

  // Terms
  csvContent += '\n'
  csvContent += 'TERMS & CONDITIONS\n'
  csvContent += 'Payment Terms,"' + quote.paymentTerms + '"\n'
  csvContent += 'Delivery Terms,"' + quote.deliveryTerms + '"\n'
  if (quote.customerNotes) {
    csvContent += 'Customer Notes,"' + quote.customerNotes + '"\n'
  }

  // Create download link with proper Excel format
  const BOM = '\uFEFF' // UTF-8 BOM for Excel to recognize UTF-8 encoding
  const blob = new Blob([BOM + csvContent.replace('data:text/csv;charset=utf-8,', '')], {
    type: 'application/vnd.ms-excel;charset=utf-8;'
  })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.setAttribute('href', url)
  link.setAttribute('download', 'Quote_' + quote.quoteNumber + '_v' + quote.version + '.xls')
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)

  success('Exported quote ' + quote.quoteNumber + ' to Excel')
}

function sendToCustomer(quote: Quote) {
  if (confirm(`Send quote ${quote.quoteNumber} to customer?`)) {
    quotesStore.sendQuote(quote.id)
    success(`Quote ${quote.quoteNumber} sent successfully`)
    if (showDetailsModal.value) {
      showDetailsModal.value = false
      selectedQuote.value = null
    }
  }
}

// Workflow Actions
function submitForApproval(quote: Quote) {
  const reason = prompt('Enter reason for approval request (optional):')
  quotesStore.updateQuoteStatus(quote.id, 'pending_approval')
  success(`Quote ${quote.quoteNumber} submitted for approval`)
  if (showDetailsModal.value) {
    showDetailsModal.value = false
    selectedQuote.value = null
  }
}

function approveQuoteAction(quote: Quote) {
  const comments = prompt('Add approval comments (optional):')
  quotesStore.approveQuote(quote.id, 'USR-004') // Current user ID
  success(`Quote ${quote.quoteNumber} approved successfully`)
  if (showDetailsModal.value) {
    showDetailsModal.value = false
    selectedQuote.value = null
  }
}

function rejectQuoteAction(quote: Quote) {
  const reason = prompt('Enter rejection reason:')
  if (reason) {
    quotesStore.rejectQuote(quote.id, reason)
    success(`Quote ${quote.quoteNumber} rejected`)
    if (showDetailsModal.value) {
      showDetailsModal.value = false
      selectedQuote.value = null
    }
  }
}

// Watch price book changes
watch(() => quoteForm.value.priceBookId, (newPriceBookId, oldPriceBookId) => {
  if (oldPriceBookId && newPriceBookId !== oldPriceBookId && quoteFormLineItems.value.length > 0) {
    info('Price book changed. New line items will use the new pricing. Existing items remain unchanged.')
  }
})

// Lifecycle - Check for opportunityId in route on mount
onMounted(() => {
  const opportunityId = route.query.opportunityId as string
  if (opportunityId) {
    // Auto-open create quote modal when coming from opportunity
    createNewQuote()
  }
})
</script>

<style scoped>
.quoting-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 1rem;
  width: 100%;
}

.search-input {
  flex: 1;
  padding: 0.75rem 1rem;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text-primary);
  border-radius: 8px;
  font-size: 0.95rem;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
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

.btn-secondary {
  background: var(--color-border);
  color: var(--color-text-primary);
}

.btn-secondary:hover {
  background: var(--color-surface-hover);
}

.btn-success {
  background: var(--color-success);
  color: #fff;
}

.btn-success:hover {
  opacity: 0.9;
}

.filters-bar {
  display: flex;
  gap: 1.5rem;
  align-items: center;
  flex-wrap: wrap;
  padding: 1.25rem;
  background: var(--color-surface);
  border-radius: 12px;
  box-shadow: var(--shadow-sm);
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.filter-group label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.filter-select {
  padding: 0.5rem 1rem;
  border: 1px solid var(--color-border);
  background: var(--color-background);
  color: var(--color-text-primary);
  border-radius: 6px;
  font-size: 0.875rem;
  cursor: pointer;
  min-width: 150px;
}

.filter-select:focus {
  outline: none;
  border-color: var(--color-primary);
}

.stats-summary {
  display: flex;
  gap: 2rem;
  margin-left: auto;
  padding-left: 2rem;
  border-left: 1px solid var(--color-border);
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.stat-label {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-value {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--color-primary);
}

.quote-number-cell {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.quote-number {
  font-weight: 600;
  color: var(--color-text-primary);
}

.quote-version {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
}

.customer-cell {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.customer-name {
  font-weight: 500;
  color: var(--color-text-primary);
}

.opportunity-name {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
}

.dates-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.date-row {
  font-size: 0.85rem;
  display: flex;
  gap: 0.5rem;
}

.date-label {
  font-weight: 500;
  color: var(--color-text-secondary);
  min-width: 80px;
}

.financial-cell {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.amount-row {
  font-weight: 600;
  font-size: 1rem;
  color: var(--color-text-primary);
}

.margin-row {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
}

.approval-cell {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
}

.approval-icon {
  font-size: 1rem;
}

.approval-text {
  color: var(--color-text-secondary);
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
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

.action-btn:hover:not(:disabled) {
  transform: scale(1.2);
  opacity: 1;
}

.action-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.text-danger {
  color: var(--color-danger) !important;
}

.text-warning {
  color: var(--color-warning) !important;
}

.text-success {
  color: var(--color-success) !important;
}

.text-right {
  text-align: right;
}

/* Modal Styles */
.modal-header-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  width: 100%;
}

.modal-header-content h2 {
  margin: 0;
  font-size: 1.5rem;
  color: var(--color-text-primary);
}

.modal-subtitle {
  margin: 0.25rem 0 0 0;
  font-size: 0.95rem;
  color: var(--color-text-secondary);
}

.quote-details {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.details-section {
  border-bottom: 1px solid var(--color-border);
  padding-bottom: 1.5rem;
}

.details-section:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.details-section h3 {
  margin: 0 0 1rem 0;
  font-size: 1.125rem;
  color: var(--color-text-primary);
}

.details-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1rem;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.detail-item.full-width {
  grid-column: 1 / -1;
}

.detail-label {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.detail-value {
  font-size: 0.95rem;
  color: var(--color-text-primary);
}

.line-items-table {
  overflow-x: auto;
}

.line-items-table table {
  width: 100%;
  border-collapse: collapse;
}

.line-items-table th,
.line-items-table td {
  padding: 0.75rem;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
}

.line-items-table th {
  background: var(--color-background);
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.line-items-table td {
  font-size: 0.9rem;
  color: var(--color-text-primary);
}

.description-cell {
  max-width: 300px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.financial-summary {
  max-width: 500px;
  margin-left: auto;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.summary-row {
  display: flex;
  justify-content: space-between;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--color-border-light);
}

.summary-row:last-child {
  border-bottom: none;
}

.summary-row.total-row {
  border-top: 2px solid var(--color-border);
  padding-top: 1rem;
  margin-top: 0.5rem;
  font-size: 1.125rem;
  font-weight: 700;
}

.summary-row.margin-row {
  background: var(--color-background);
  padding: 0.75rem;
  border-radius: 6px;
  margin-top: 0.5rem;
}

.summary-label {
  color: var(--color-text-secondary);
  font-weight: 500;
}

.summary-value {
  color: var(--color-text-primary);
  font-weight: 600;
}

.modal-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
}

/* Form Styles */
.quote-form {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.form-section {
  border-bottom: 1px solid var(--color-border);
  padding-bottom: 1.5rem;
}

.form-section:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.section-title {
  margin: 0 0 1.25rem 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.25rem;
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
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-primary);
}

.form-hint {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.75rem;
  color: var(--color-primary);
  font-style: italic;
}

.form-hint.warning {
  color: var(--color-warning);
  font-weight: 500;
}

.form-input,
.form-input:focus {
  padding: 0.75rem;
  border: 1px solid var(--color-border);
  background: var(--color-background);
  color: var(--color-text-primary);
  border-radius: 8px;
  font-size: 0.95rem;
  transition: all 0.2s;
  font-family: inherit;
}

.form-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-input:read-only {
  background: var(--color-surface);
  color: var(--color-text-secondary);
  cursor: not-allowed;
}

.form-input[readonly] {
  background: var(--color-surface);
  opacity: 0.7;
}

textarea.form-input {
  resize: vertical;
  min-height: 80px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  padding: 0.5rem 0;
}

.checkbox-label input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: var(--color-primary);
}

.checkbox-label span {
  font-size: 0.95rem;
  color: var(--color-text-primary);
}

.form-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  padding-top: 1.5rem;
  border-top: 1px solid var(--color-border);
}

.action-btn-danger {
  color: var(--color-danger);
}

.action-btn-danger:hover:not(:disabled) {
  color: #dc2626;
}

/* Progress Track / Workflow Stages */
.progress-track {
  margin-bottom: 2rem;
  padding: 2rem 1rem;
  background: linear-gradient(135deg, var(--color-surface) 0%, var(--color-background) 100%);
  border-radius: 12px;
  border: 1px solid var(--color-border);
}

.progressbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  position: relative;
  max-width: 900px;
  margin: 0 auto;
}

.progressbar::before {
  content: '';
  position: absolute;
  top: 24px;
  left: 10%;
  right: 10%;
  height: 3px;
  background: var(--color-border);
  z-index: 0;
}

.progress-step {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  position: relative;
  z-index: 1;
  flex: 1;
}

.step-circle {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: var(--color-surface);
  border: 3px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 1.1rem;
  color: var(--color-text-secondary);
  transition: all 0.3s ease;
}

.progress-step.completed .step-circle {
  background: var(--color-success);
  border-color: var(--color-success);
  color: white;
}

.progress-step.active .step-circle {
  background: var(--color-primary);
  border-color: var(--color-primary);
  color: white;
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.2);
  transform: scale(1.1);
}

.progress-step.declined .step-circle {
  background: var(--color-danger);
  border-color: var(--color-danger);
  color: white;
}

.step-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-align: center;
  max-width: 120px;
}

.progress-step.completed .step-label,
.progress-step.active .step-label {
  color: var(--color-text-primary);
}

/* Category Sections */
.category-section {
  margin-bottom: 1.5rem;
}

.category-section:last-child {
  margin-bottom: 0;
}

.category-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin: 0 0 1rem 0;
  padding: 0.75rem 1rem;
  background: linear-gradient(135deg, var(--color-primary-light) 0%, var(--color-surface) 100%);
  border-left: 4px solid var(--color-primary);
  border-radius: 6px;
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--color-primary);
}

.category-icon {
  font-size: 1.3rem;
}

.line-items-table table {
  width: 100%;
  border-collapse: collapse;
  background: var(--color-surface);
  border-radius: 8px;
  overflow: hidden;
}

.line-items-table thead {
  background: var(--color-background);
}

.line-items-table thead th {
  padding: 0.875rem 1rem;
  text-align: left;
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 2px solid var(--color-border);
}

.line-items-table tbody td {
  padding: 1rem;
  border-bottom: 1px solid var(--color-border-light);
  color: var(--color-text-primary);
  font-size: 0.95rem;
}

.line-items-table tbody tr:last-child td {
  border-bottom: none;
}

.line-items-table tbody tr:hover {
  background: var(--color-background);
}

.item-name {
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: 0.25rem;
}

.item-description {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  line-height: 1.4;
}

.item-meta {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  font-style: italic;
  margin-top: 0.25rem;
}

.service-duration {
  color: var(--color-primary);
  font-weight: 600;
  font-style: normal;
}

.rate-suffix {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  font-weight: normal;
}

.category-footer {
  background: var(--color-background);
  border-top: 2px solid var(--color-border);
}

.category-footer th {
  padding: 0.875rem 1rem;
  font-weight: 700;
  color: var(--color-text-primary);
  font-size: 0.95rem;
}

/* Approval History Table */
.approval-history-table {
  background: var(--color-surface);
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--color-border);
}

.approval-history-table table {
  width: 100%;
  border-collapse: collapse;
}

.approval-history-table thead {
  background: var(--color-background);
}

.approval-history-table thead th {
  padding: 0.875rem 1rem;
  text-align: left;
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 2px solid var(--color-border);
}

.approval-history-table tbody td {
  padding: 1rem;
  border-bottom: 1px solid var(--color-border-light);
  color: var(--color-text-primary);
  font-size: 0.95rem;
  vertical-align: top;
}

.approval-history-table tbody tr:last-child td {
  border-bottom: none;
}

.approval-history-table tbody tr:hover {
  background: var(--color-background);
}

/* Workflow Actions */
.workflow-actions {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
  padding: 1rem;
  background: var(--color-surface);
  border-radius: 8px;
  border: 1px solid var(--color-border);
}

.workflow-actions .btn {
  min-width: 140px;
}

/* Line Item Form & Management */
.line-item-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.line-items-list {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  margin-top: 1rem;
}

.category-group {
  border: 1px solid var(--color-border);
  border-radius: 8px;
  overflow: hidden;
  background: var(--color-surface);
}

.category-group-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.875rem 1rem;
  background: var(--color-background);
  border-bottom: 1px solid var(--color-border);
  font-weight: 600;
  font-size: 0.95rem;
}

.category-icon {
  font-size: 1.25rem;
}

.category-title {
  color: var(--color-text-primary);
}

.category-count {
  color: var(--color-text-secondary);
  font-size: 0.875rem;
  font-weight: 400;
}

.line-items-mini-table {
  overflow-x: auto;
}

.line-items-mini-table table {
  width: 100%;
  border-collapse: collapse;
}

.line-items-mini-table thead {
  background: var(--color-background);
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.line-items-mini-table thead th {
  padding: 0.625rem 0.75rem;
  text-align: left;
  font-weight: 600;
  border-bottom: 1px solid var(--color-border);
}

.line-items-mini-table thead th.text-right {
  text-align: right;
}

.line-items-mini-table tbody td {
  padding: 0.75rem;
  font-size: 0.9rem;
  color: var(--color-text-primary);
  border-bottom: 1px solid var(--color-border-light);
}

.line-items-mini-table tbody tr:last-child td {
  border-bottom: none;
}

.line-items-mini-table tbody tr:hover {
  background: var(--color-background);
}

.item-mini-name {
  font-weight: 500;
  color: var(--color-text-primary);
  margin-bottom: 0.25rem;
}

.item-mini-desc {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  line-height: 1.3;
}

.item-mini-meta {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
  margin-top: 0.25rem;
  font-style: italic;
}

.line-items-mini-table tfoot {
  background: var(--color-background);
  border-top: 2px solid var(--color-border);
  font-weight: 600;
}

.line-items-mini-table tfoot th {
  padding: 0.75rem;
  font-size: 0.9rem;
}

.line-items-total {
  margin-top: 1rem;
  padding: 1rem;
  background: var(--color-primary);
  color: white;
  border-radius: 8px;
  display: flex;
  justify-content: flex-end;
}

.total-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  font-size: 1.1rem;
  font-weight: 600;
}

.total-label {
  opacity: 0.95;
}

.total-value {
  font-size: 1.25rem;
  font-weight: 700;
}

.empty-line-items {
  padding: 3rem 1rem;
  text-align: center;
  color: var(--color-text-secondary);
  background: var(--color-background);
  border: 2px dashed var(--color-border);
  border-radius: 8px;
  margin-top: 1rem;
}

.btn-icon {
  background: none;
  border: none;
  padding: 0.25rem 0.5rem;
  cursor: pointer;
  font-size: 1rem;
  opacity: 0.7;
  transition: opacity 0.2s;
}

.btn-icon:hover {
  opacity: 1;
}

.line-item-preview {
  background: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.preview-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.95rem;
}

.preview-label {
  color: var(--color-text-secondary);
  font-weight: 500;
}

.preview-value {
  color: var(--color-text-primary);
  font-weight: 600;
}

.preview-total {
  padding-top: 0.5rem;
  margin-top: 0.5rem;
  border-top: 2px solid var(--color-border);
  font-size: 1.05rem;
}

.preview-total .preview-label {
  color: var(--color-text-primary);
  font-weight: 600;
}

.preview-total .preview-value {
  font-weight: 700;
  color: var(--color-primary);
  font-size: 1.15rem;
}

/* Line Items Summary in Table */
.line-items-summary {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.items-total {
  font-weight: 600;
  color: var(--color-text-primary);
  font-size: 0.95rem;
}

.items-breakdown {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.item-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 500;
  white-space: nowrap;
}

.item-badge.materials {
  background: #e3f2fd;
  color: #1976d2;
}

.item-badge.manpower {
  background: #fff3e0;
  color: #f57c00;
}

.item-badge.misc {
  background: #f3e5f5;
  color: #7b1fa2;
}

/* Opportunity Link */
.opportunity-link {
  color: var(--color-primary);
  text-decoration: underline;
  cursor: pointer;
  transition: color 0.2s;
}

.opportunity-link:hover {
  color: var(--color-primary-dark);
  text-decoration-color: var(--color-primary-dark);
}

.no-opportunity {
  color: var(--color-text-secondary);
  font-style: italic;
}

/* Price Book Pricing Display */
.price-difference {
  display: inline-block;
  margin-left: 0.5rem;
  font-weight: 500;
}

.discount-text {
  color: var(--color-success);
  font-weight: 600;
}

.markup-text {
  color: var(--color-warning);
  font-weight: 600;
}

.form-text {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  line-height: 1.4;
}

.form-control {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-background);
  color: var(--color-text-primary);
  font-size: 0.95rem;
  transition: all 0.2s;
}

.form-control:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-checkbox {
  margin-right: 0.5rem;
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.form-check-label {
  display: flex;
  align-items: center;
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;
  color: var(--color-text-primary);
}

/* Attachments Section */
.attachments-container {
  margin-top: 1rem;
}

.attachments-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.attachment-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
}

.attachment-info {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex: 1;
}

.attachment-order {
  font-weight: 700;
  color: var(--color-text-secondary);
  min-width: 30px;
}

.attachment-name-input {
  flex: 1;
  min-width: 200px;
  padding: 0.5rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-background);
  color: var(--color-text-primary);
  font-size: 0.95rem;
}

.attachment-select {
  flex: 1;
  min-width: 250px;
  padding: 0.5rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-background);
  color: var(--color-text-primary);
  font-size: 0.95rem;
}

.checkbox-label-inline {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  cursor: pointer;
  white-space: nowrap;
}

.checkbox-label-inline input[type="checkbox"] {
  cursor: pointer;
}

.attachment-actions {
  display: flex;
  gap: 0.5rem;
}

.btn-icon {
  background: none;
  border: none;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  transition: background 0.2s;
}

.btn-icon:hover:not(:disabled) {
  background: var(--color-border);
}

.btn-icon:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.btn-icon.btn-danger:hover:not(:disabled) {
  background: rgba(244, 67, 54, 0.1);
}

.empty-state {
  padding: 2rem;
  text-align: center;
  color: var(--color-text-secondary);
  font-size: 0.95rem;
  background: var(--color-surface);
  border: 1px dashed var(--color-border);
  border-radius: 8px;
  margin-top: 1rem;
}

/* Modern Premium Quote Details Styles */
.quote-details-modern {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

/* Modern Progress Card */
.modern-progress-card {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.05), rgba(139, 92, 246, 0.05));
  border-radius: 16px;
  padding: 2rem;
  border: 1px solid rgba(59, 130, 246, 0.1);
}

.modern-progress-card .progress-steps {
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: relative;
  gap: 1rem;
}

.modern-progress-card .progress-step {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  flex: 0 0 auto;
  z-index: 1;
}

.modern-progress-card .step-icon {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--color-surface);
  border: 3px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
  font-weight: 700;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  color: var(--color-text-secondary);
}

.modern-progress-card .progress-step.active .step-icon,
.modern-progress-card .progress-step.completed .step-icon {
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  border-color: transparent;
  color: white;
  transform: scale(1.15);
  box-shadow: 0 8px 20px rgba(59, 130, 246, 0.4);
}

.modern-progress-card .step-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.modern-progress-card .progress-step.active .step-label,
.modern-progress-card .progress-step.completed .step-label {
  color: var(--color-primary);
  font-weight: 700;
}

/* Modern Info Card */
.modern-info-card {
  background: var(--color-surface);
  border-radius: 16px;
  border: 1px solid var(--color-border);
  overflow: hidden;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.modern-info-card:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}

.card-header-mini {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.25rem 1.5rem;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.03), rgba(139, 92, 246, 0.03));
  border-bottom: 1px solid var(--color-border);
}

.card-header-mini h4 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--color-text-primary);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.count-badge {
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
  padding: 0.375rem 0.875rem;
  border-radius: 20px;
  font-size: 0.875rem;
  font-weight: 700;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
}

.card-content {
  padding: 1.5rem;
}

/* Info Grid */
.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.25rem;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.info-label {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.info-value {
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.info-value.highlight {
  font-size: 1.125rem;
  color: var(--color-primary);
  font-weight: 700;
}

/* Category Section Modern */
.category-section-modern {
  margin-bottom: 2rem;
}

.category-section-modern:last-child {
  margin-bottom: 0;
}

.category-header-modern {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem 1.25rem;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.05), rgba(139, 92, 246, 0.05));
  border-radius: 12px;
  margin-bottom: 1rem;
  border: 1px solid rgba(59, 130, 246, 0.1);
}

.category-header-modern .category-icon {
  font-size: 1.5rem;
}

.category-header-modern h5 {
  flex: 1;
  margin: 0;
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.category-count {
  background: var(--color-primary);
  color: white;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.875rem;
  font-weight: 700;
}

/* Items List Modern */
.items-list-modern {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.item-card-mini {
  display: flex;
  gap: 1rem;
  padding: 1.25rem;
  background: var(--color-background);
  border: 2px solid var(--color-border);
  border-radius: 12px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.item-card-mini:hover {
  border-color: #3b82f6;
  box-shadow: 0 4px 16px rgba(59, 130, 246, 0.1);
  transform: translateX(4px);
}

.item-number-mini {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.875rem;
  flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
}

.item-details {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.item-header-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
}

.item-name {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  flex: 1;
}

.item-code {
  font-family: 'Courier New', monospace;
  font-size: 0.8125rem;
  font-weight: 700;
  color: var(--color-primary);
  opacity: 0.8;
}

.item-title {
  font-size: 1rem;
  font-weight: 700;
  color: var(--color-text-primary);
  line-height: 1.4;
}

.item-desc {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  line-height: 1.5;
}

.item-meta-tag {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.625rem;
  background: rgba(59, 130, 246, 0.1);
  border-radius: 6px;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-primary);
  margin-top: 0.25rem;
  width: fit-content;
}

.item-meta-tag.recurring {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.item-final-price {
  font-size: 1.25rem;
  font-weight: 800;
  color: #10b981;
  background: rgba(16, 185, 129, 0.1);
  padding: 0.5rem 1rem;
  border-radius: 10px;
  white-space: nowrap;
}

.item-pricing-row {
  display: flex;
  flex-wrap: wrap;
  gap: 1.25rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--color-border);
}

.pricing-detail {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.pricing-label {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.pricing-value {
  font-size: 0.9375rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.rate-suffix {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  font-weight: 500;
}

/* Category Subtotal */
.category-subtotal {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.05), rgba(139, 92, 246, 0.05));
  border-radius: 12px;
  margin-top: 1rem;
  font-weight: 700;
  border: 1px solid rgba(59, 130, 246, 0.1);
}

.subtotal-value {
  font-size: 1.25rem;
  color: var(--color-primary);
}

/* Financial Grid Modern */
.financial-grid-modern {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.financial-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.875rem 1rem;
  background: var(--color-background);
  border-radius: 10px;
  transition: all 0.2s;
}

.financial-row:hover {
  background: rgba(59, 130, 246, 0.03);
}

.financial-label {
  font-size: 0.9375rem;
  font-weight: 500;
  color: var(--color-text-secondary);
}

.financial-value {
  font-size: 1rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.financial-row.discount-row {
  background: rgba(239, 68, 68, 0.05);
}

.financial-value.discount {
  color: #ef4444;
}

.financial-row.total-row-modern {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.1), rgba(139, 92, 246, 0.1));
  border: 2px solid rgba(59, 130, 246, 0.2);
  padding: 1.25rem 1rem;
  margin: 0.5rem 0;
}

.financial-row.total-row-modern .financial-label {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.financial-value.total {
  font-size: 1.5rem;
  color: var(--color-primary);
  font-weight: 800;
}

.divider-line {
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--color-border), transparent);
  margin: 0.5rem 0;
}

.financial-row.cost-row {
  background: rgba(239, 68, 68, 0.05);
}

.financial-value.cost {
  color: #ef4444;
}

.financial-row.profit-row {
  background: rgba(16, 185, 129, 0.05);
}

.financial-value.profit {
  color: #10b981;
}

.financial-row.margin-row-modern {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.1), rgba(5, 150, 105, 0.1));
  border: 1px solid rgba(16, 185, 129, 0.2);
}

.financial-value.margin {
  font-size: 1.25rem;
}

/* Terms Grid */
.terms-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1.25rem;
}

.term-box {
  padding: 1.25rem;
  background: var(--color-background);
  border-radius: 12px;
  border: 1px solid var(--color-border);
  transition: all 0.3s;
}

.term-box:hover {
  border-color: rgba(59, 130, 246, 0.3);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.term-box.full-width {
  grid-column: 1 / -1;
}

.term-box.internal-note {
  background: rgba(251, 191, 36, 0.05);
  border-color: rgba(251, 191, 36, 0.2);
}

.term-label {
  display: block;
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--color-text-secondary);
  margin-bottom: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.term-value {
  margin: 0;
  font-size: 0.9375rem;
  line-height: 1.6;
  color: var(--color-text-primary);
  font-weight: 500;
}

.term-value.notes {
  font-style: italic;
  padding: 0.75rem;
  background: rgba(0, 0, 0, 0.02);
  border-radius: 8px;
  border-left: 3px solid var(--color-primary);
}

.opportunity-link {
  color: var(--color-primary);
  text-decoration: underline;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
}

.opportunity-link:hover {
  color: #2563eb;
}

.no-opportunity {
  color: var(--color-text-secondary);
  font-style: italic;
}
</style>
