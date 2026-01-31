<template>
  <BaseModal
    v-model="isOpen"
    :title="isEditMode ? `Edit Quote ${quoteData.quoteNumber}` : 'Create New Quote'"
    size="xl"
    @close="handleClose"
  >
    <div class="quote-builder">
      <!-- Quote Header Section -->
      <div class="builder-section">
        <h3 class="section-title">Quote Information</h3>
        <div class="form-grid">
          <div class="form-group">
            <label class="form-label required">Customer</label>
            <select
              v-model="quoteData.companyId"
              class="form-select"
              @change="onCustomerChange"
              required
            >
              <option value="">Select Customer...</option>
              <option
                v-for="customer in customersStore.customers"
                :key="customer.id"
                :value="customer.id"
              >
                {{ customer.name }}
              </option>
            </select>
          </div>

          <div class="form-group">
            <label class="form-label required">Contact</label>
            <select
              v-model="quoteData.contactId"
              class="form-select"
              :disabled="!quoteData.companyId"
              required
            >
              <option value="">Select Contact...</option>
              <option
                v-for="contact in availableContacts"
                :key="contact.id"
                :value="contact.id"
              >
                {{ contact.fullName }} - {{ contact.email }}
              </option>
            </select>
          </div>

          <div class="form-group">
            <label class="form-label">Opportunity</label>
            <select
              v-model="quoteData.opportunityId"
              class="form-select"
              :disabled="!quoteData.companyId"
            >
              <option value="">Select Opportunity (Optional)...</option>
              <option
                v-for="opportunity in availableOpportunities"
                :key="opportunity.id"
                :value="opportunity.id"
              >
                {{ opportunity.name }} - {{ formatCurrency(opportunity.value) }}
              </option>
            </select>
          </div>

          <div class="form-group">
            <label class="form-label required">Quote Date</label>
            <input
              v-model="quoteData.quoteDate"
              type="date"
              class="form-input"
              required
            />
          </div>

          <div class="form-group">
            <label class="form-label required">Valid Until</label>
            <input
              v-model="quoteData.validUntil"
              type="date"
              class="form-input"
              :min="quoteData.quoteDate"
              required
            />
          </div>

          <div class="form-group">
            <label class="form-label">Currency</label>
            <select v-model="quoteData.currency" class="form-select">
              <option value="SAR">SAR - Saudi Riyal</option>
              <option value="USD">USD - US Dollar</option>
              <option value="EUR">EUR - Euro</option>
            </select>
          </div>
        </div>
      </div>

      <!-- Line Items Section -->
      <div class="builder-section">
        <div class="section-header">
          <h3 class="section-title">Line Items</h3>
          <button class="btn btn-sm btn-primary" @click="addLineItem">
            + Add Line Item
          </button>
        </div>

        <div v-if="lineItems.length === 0" class="empty-state">
          <p>No line items added yet. Click "Add Line Item" to get started.</p>
        </div>

        <div v-else class="line-items-container">
          <table class="line-items-table">
            <thead>
              <tr>
                <th style="width: 40px">#</th>
                <th style="width: 250px">Product/Service</th>
                <th style="width: 80px">Qty</th>
                <th style="width: 120px">Unit Price</th>
                <th style="width: 80px">Disc %</th>
                <th style="width: 120px">Line Total</th>
                <th style="width: 80px">Cost</th>
                <th style="width: 80px">Margin %</th>
                <th style="width: 80px">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(item, index) in lineItems" :key="item.tempId" class="line-item-row">
                <td class="line-number">{{ index + 1 }}</td>
                <td>
                  <div class="product-cell">
                    <select
                      v-model="item.productId"
                      class="form-select-sm"
                      @change="onProductChange(item)"
                    >
                      <option value="">Select Product...</option>
                      <option
                        v-for="product in productsStore.products"
                        :key="product.id"
                        :value="product.id"
                      >
                        {{ product.code }} - {{ product.name }}
                      </option>
                    </select>
                    <input
                      v-model="item.description"
                      type="text"
                      class="form-input-sm"
                      placeholder="Description..."
                    />
                  </div>
                </td>
                <td>
                  <input
                    v-model.number="item.quantity"
                    type="number"
                    min="1"
                    class="form-input-sm text-right"
                    @input="calculateLineItem(item)"
                  />
                </td>
                <td>
                  <input
                    v-model.number="item.unitPrice"
                    type="number"
                    min="0"
                    step="0.01"
                    class="form-input-sm text-right"
                    @input="calculateLineItem(item)"
                  />
                </td>
                <td>
                  <input
                    v-model.number="item.discountPercent"
                    type="number"
                    min="0"
                    max="100"
                    step="0.1"
                    class="form-input-sm text-right"
                    @input="calculateLineItem(item)"
                  />
                </td>
                <td class="line-total">
                  {{ formatCurrency(item.lineTotal) }}d
                </td>
                <td>
                  <input
                    v-model.number="item.unitCost"
                    type="number"
                    min="0"
                    step="0.01"
                    class="form-input-sm text-right"
                    @input="calculateLineItem(item)"
                  />
                </td>
                <td class="margin-cell" :class="getMarginClass(item.marginPercent)">
                  {{ item.marginPercent.toFixed(1) }}%
                </td>
                <td>
                  <button
                    class="btn-icon btn-danger-icon"
                    title="Remove Line"
                    @click="removeLineItem(index)"
                  >
                    🗑️
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Financial Summary Section -->
      <div class="builder-section">
        <h3 class="section-title">Financial Summary</h3>
        <div class="summary-layout">
          <div class="summary-left">
            <div class="form-group">
              <label class="form-label">Additional Discount %</label>
              <input
                v-model.number="quoteData.discountPercent"
                type="number"
                min="0"
                max="100"
                step="0.1"
                class="form-input"
                style="max-width: 150px"
                @input="calculateTotals"
              />
            </div>
            <div class="form-group">
              <label class="form-label">VAT %</label>
              <input
                v-model.number="quoteData.vatPercent"
                type="number"
                min="0"
                max="100"
                step="0.1"
                class="form-input"
                style="max-width: 150px"
                @input="calculateTotals"
              />
            </div>
          </div>

          <div class="summary-right">
            <div class="summary-row">
              <span class="summary-label">Subtotal:</span>
              <span class="summary-value">{{ formatCurrency(calculatedSubtotal) }}</span>
            </div>
            <div v-if="quoteData.discountPercent > 0" class="summary-row">
              <span class="summary-label">Discount ({{ quoteData.discountPercent }}%):</span>
              <span class="summary-value text-danger">-{{ formatCurrency(calculatedDiscountAmount) }}</span>
            </div>
            <div class="summary-row">
              <span class="summary-label">Subtotal After Discount:</span>
              <span class="summary-value">{{ formatCurrency(calculatedSubtotalAfterDiscount) }}</span>
            </div>
            <div class="summary-row">
              <span class="summary-label">VAT ({{ quoteData.vatPercent }}%):</span>
              <span class="summary-value">{{ formatCurrency(calculatedVatAmount) }}</span>
            </div>
            <div class="summary-row total-row">
              <span class="summary-label">Total:</span>
              <span class="summary-value">{{ formatCurrency(calculatedTotal) }}</span>
            </div>
            <div class="summary-row margin-row">
              <span class="summary-label">Total Cost:</span>
              <span class="summary-value">{{ formatCurrency(calculatedTotalCost) }}</span>
            </div>
            <div class="summary-row margin-row">
              <span class="summary-label">Total Margin:</span>
              <span class="summary-value" :class="getMarginClass(calculatedMarginPercent)">
                {{ formatCurrency(calculatedMarginAmount) }} ({{ calculatedMarginPercent.toFixed(1) }}%)
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- Terms & Notes Section -->
      <div class="builder-section">
        <h3 class="section-title">Terms & Notes</h3>
        <div class="form-grid">
          <div class="form-group full-width">
            <label class="form-label">Payment Terms</label>
            <select v-model="quoteData.paymentTerms" class="form-select">
              <option value="Net 30">Net 30 Days</option>
              <option value="Net 60">Net 60 Days</option>
              <option value="50% Advance, 50% on Delivery">50% Advance, 50% on Delivery</option>
              <option value="100% Advance">100% Advance</option>
              <option value="Letter of Credit">Letter of Credit</option>
            </select>
          </div>

          <div class="form-group full-width">
            <label class="form-label">Delivery Terms</label>
            <select v-model="quoteData.deliveryTerms" class="form-select">
              <option value="Ex-Works">Ex-Works</option>
              <option value="FOB">FOB (Free on Board)</option>
              <option value="CIF">CIF (Cost, Insurance & Freight)</option>
              <option value="DDP">DDP (Delivered Duty Paid)</option>
              <option value="4-6 Weeks">4-6 Weeks from Order</option>
              <option value="8-10 Weeks">8-10 Weeks from Order</option>
            </select>
          </div>

          <div class="form-group full-width">
            <label class="form-label">Customer Notes (Visible to Customer)</label>
            <textarea
              v-model="quoteData.customerNotes"
              class="form-textarea"
              rows="3"
              placeholder="Enter any notes or special terms for the customer..."
            ></textarea>
          </div>

          <div class="form-group full-width">
            <label class="form-label">Internal Notes (Internal Only)</label>
            <textarea
              v-model="quoteData.internalNotes"
              class="form-textarea"
              rows="3"
              placeholder="Enter internal notes or comments..."
            ></textarea>
          </div>
        </div>
      </div>

      <!-- Approval Warning -->
      <div v-if="requiresApproval" class="approval-warning">
        <div class="warning-icon">⚠️</div>
        <div class="warning-content">
          <strong>Approval Required</strong>
          <p>{{ approvalReason }}</p>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="modal-footer-actions">
        <button class="btn btn-secondary" @click="handleClose">Cancel</button>
        <button
          class="btn btn-outline"
          @click="saveAsDraft"
          :disabled="!canSave"
        >
          Save as Draft
        </button>
        <button
          class="btn btn-primary"
          @click="submitForApproval"
          :disabled="!canSave"
        >
          {{ requiresApproval ? 'Submit for Approval' : 'Create Quote' }}
        </button>
      </div>
    </template>
  </BaseModal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useCustomersStore } from '@/stores/customers'
import { useContactsStore } from '@/stores/contacts'
import { useOpportunitiesStore } from '@/stores/opportunities'
import { useProductsStore } from '@/stores/products'
import { useQuotesStore } from '@/stores/quotes'
import { useToast } from '@/composables/useToast'
import BaseModal from '@/components/UI/BaseModal.vue'
import type { Quote } from '@/types'

interface LineItem {
  tempId: string
  productId: string
  productName: string
  description: string
  quantity: number
  unitPrice: number
  unitCost: number
  discountPercent: number
  lineTotal: number
  lineCost: number
  marginAmount: number
  marginPercent: number
}

interface Props {
  show: boolean
  quoteToEdit?: Quote | null
}

interface Emits {
  (e: 'close'): void
  (e: 'saved', quote: Quote): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const customersStore = useCustomersStore()
const contactsStore = useContactsStore()
const opportunitiesStore = useOpportunitiesStore()
const productsStore = useProductsStore()
const quotesStore = useQuotesStore()
const { success, error: showError } = useToast()

const isOpen = computed({
  get: () => props.show,
  set: (value) => {
    if (!value) emit('close')
  }
})

const isEditMode = computed(() => !!props.quoteToEdit)

// Quote data
const quoteData = ref({
  id: '',
  quoteNumber: '',
  version: 1,
  companyId: '',
  contactId: '',
  opportunityId: '',
  quoteDate: new Date().toISOString().split('T')[0],
  validUntil: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
  currency: 'SAR',
  discountPercent: 0,
  vatPercent: 15,
  paymentTerms: 'Net 30',
  deliveryTerms: 'Ex-Works',
  customerNotes: '',
  internalNotes: '',
  status: 'draft' as const
})

const lineItems = ref<LineItem[]>([])
let lineItemCounter = 0

// Available options based on selections
const availableContacts = computed(() => {
  if (!quoteData.value.companyId) return []
  return contactsStore.getContactsByCompanyId(quoteData.value.companyId)
})

const availableOpportunities = computed(() => {
  if (!quoteData.value.companyId) return []
  return opportunitiesStore.getOpportunitiesByCompanyId(quoteData.value.companyId)
})

// Calculations
const calculatedSubtotal = computed(() =>
  lineItems.value.reduce((sum, item) => sum + item.lineTotal, 0)
)

const calculatedDiscountAmount = computed(() =>
  calculatedSubtotal.value * (quoteData.value.discountPercent / 100)
)

const calculatedSubtotalAfterDiscount = computed(() =>
  calculatedSubtotal.value - calculatedDiscountAmount.value
)

const calculatedVatAmount = computed(() =>
  calculatedSubtotalAfterDiscount.value * (quoteData.value.vatPercent / 100)
)

const calculatedTotal = computed(() =>
  calculatedSubtotalAfterDiscount.value + calculatedVatAmount.value
)

const calculatedTotalCost = computed(() =>
  lineItems.value.reduce((sum, item) => sum + item.lineCost, 0)
)

const calculatedMarginAmount = computed(() =>
  calculatedSubtotalAfterDiscount.value - calculatedTotalCost.value
)

const calculatedMarginPercent = computed(() => {
  if (calculatedSubtotalAfterDiscount.value === 0) return 0
  return (calculatedMarginAmount.value / calculatedSubtotalAfterDiscount.value) * 100
})

// Approval logic
const requiresApproval = computed(() => {
  // Requires approval if margin < 10% or discount > 20%
  return calculatedMarginPercent.value < 10 || quoteData.value.discountPercent > 20
})

const approvalReason = computed(() => {
  const reasons: string[] = []
  if (calculatedMarginPercent.value < 10) {
    reasons.push(`Low margin (${calculatedMarginPercent.value.toFixed(1)}%)`)
  }
  if (quoteData.value.discountPercent > 20) {
    reasons.push(`High discount (${quoteData.value.discountPercent}%)`)
  }
  return reasons.join(', ')
})

const canSave = computed(() => {
  return (
    quoteData.value.companyId &&
    quoteData.value.contactId &&
    quoteData.value.quoteDate &&
    quoteData.value.validUntil &&
    lineItems.value.length > 0 &&
    lineItems.value.every(item => item.productId && item.quantity > 0 && item.unitPrice >= 0)
  )
})

// Initialize data when editing
watch(() => props.quoteToEdit, (quote) => {
  if (quote) {
    quoteData.value = {
      id: quote.id,
      quoteNumber: quote.quoteNumber,
      version: quote.version,
      companyId: quote.companyId,
      contactId: quote.contactId,
      opportunityId: quote.opportunityId || '',
      quoteDate: quote.quoteDate,
      validUntil: quote.validUntil,
      currency: quote.currency,
      discountPercent: quote.discountPercent,
      vatPercent: quote.vatPercent,
      paymentTerms: quote.paymentTerms,
      deliveryTerms: quote.deliveryTerms,
      customerNotes: quote.customerNotes || '',
      internalNotes: quote.internalNotes || '',
      status: quote.status
    }

    // Load line items
    const quoteLineItems = quotesStore.getLineItemsByQuoteId(quote.id)
    lineItems.value = quoteLineItems.map(item => ({
      tempId: `line-${lineItemCounter++}`,
      productId: item.productId,
      productName: item.productName,
      description: item.description,
      quantity: item.quantity,
      unitPrice: item.unitPrice,
      unitCost: item.unitCost,
      discountPercent: item.discountPercent,
      lineTotal: item.lineTotal,
      lineCost: item.lineCost,
      marginAmount: item.marginAmount,
      marginPercent: item.marginPercent
    }))
  } else {
    resetForm()
  }
}, { immediate: true })

// Line item functions
function addLineItem() {
  lineItems.value.push({
    tempId: `line-${lineItemCounter++}`,
    productId: '',
    productName: '',
    description: '',
    quantity: 1,
    unitPrice: 0,
    unitCost: 0,
    discountPercent: 0,
    lineTotal: 0,
    lineCost: 0,
    marginAmount: 0,
    marginPercent: 0
  })
}

function removeLineItem(index: number) {
  lineItems.value.splice(index, 1)
  calculateTotals()
}

function onProductChange(item: LineItem) {
  const product = productsStore.getProductById(item.productId)
  if (product) {
    item.productName = product.name
    item.description = product.description
    item.unitPrice = product.sellingPrice
    item.unitCost = product.sarCost
    calculateLineItem(item)
  }
}

function calculateLineItem(item: LineItem) {
  const subtotal = item.quantity * item.unitPrice
  const discount = subtotal * (item.discountPercent / 100)
  item.lineTotal = subtotal - discount
  item.lineCost = item.quantity * item.unitCost
  item.marginAmount = item.lineTotal - item.lineCost
  item.marginPercent = item.lineTotal > 0 ? (item.marginAmount / item.lineTotal) * 100 : 0

  calculateTotals()
}

function calculateTotals() {
  // Totals are computed automatically via computed properties
  // This function is called to trigger reactivity
}

function onCustomerChange() {
  // Reset dependent fields when customer changes
  quoteData.value.contactId = ''
  quoteData.value.opportunityId = ''
}

// Save functions
async function saveAsDraft() {
  if (!canSave.value) {
    showError('Please fill in all required fields')
    return
  }

  try {
    const quote = buildQuoteObject('draft')

    if (isEditMode.value) {
      // Update existing quote
      success('Quote saved as draft')
    } else {
      // Create new quote
      success('Quote created as draft')
    }

    emit('saved', quote)
    emit('close')
  } catch (err) {
    showError('Failed to save quote')
    console.error(err)
  }
}

async function submitForApproval() {
  if (!canSave.value) {
    showError('Please fill in all required fields')
    return
  }

  try {
    const status = requiresApproval.value ? 'pending_approval' : 'approved'
    const quote = buildQuoteObject(status)

    if (isEditMode.value) {
      success(`Quote ${quote.quoteNumber} updated`)
    } else {
      if (requiresApproval.value) {
        success('Quote submitted for approval')
      } else {
        success('Quote created and approved')
      }
    }

    emit('saved', quote)
    emit('close')
  } catch (err) {
    showError('Failed to submit quote')
    console.error(err)
  }
}

function buildQuoteObject(status: Quote['status']): Quote {
  return {
    id: quoteData.value.id || `Q${Date.now()}`,
    quoteNumber: quoteData.value.quoteNumber || `QT-${Date.now().toString().slice(-6)}`,
    version: quoteData.value.version,
    companyId: quoteData.value.companyId,
    contactId: quoteData.value.contactId,
    opportunityId: quoteData.value.opportunityId || undefined,
    quoteDate: quoteData.value.quoteDate,
    validUntil: quoteData.value.validUntil,
    currency: quoteData.value.currency,
    subtotal: calculatedSubtotal.value,
    discountPercent: quoteData.value.discountPercent,
    discountAmount: calculatedDiscountAmount.value,
    subtotalAfterDiscount: calculatedSubtotalAfterDiscount.value,
    vatPercent: quoteData.value.vatPercent,
    vatAmount: calculatedVatAmount.value,
    total: calculatedTotal.value,
    totalCost: calculatedTotalCost.value,
    marginAmount: calculatedMarginAmount.value,
    marginPercent: calculatedMarginPercent.value,
    paymentTerms: quoteData.value.paymentTerms,
    deliveryTerms: quoteData.value.deliveryTerms,
    customerNotes: quoteData.value.customerNotes,
    internalNotes: quoteData.value.internalNotes,
    status,
    requiresApproval: requiresApproval.value,
    approvalReason: requiresApproval.value ? approvalReason.value : undefined,
    approvedBy: status === 'approved' && !requiresApproval.value ? 'System Auto-Approved' : undefined,
    approvedAt: status === 'approved' && !requiresApproval.value ? new Date().toISOString() : undefined,
    createdBy: 'Current User',
    createdAt: quoteData.value.id ? undefined : new Date().toISOString()
  } as Quote
}

function resetForm() {
  quoteData.value = {
    id: '',
    quoteNumber: '',
    version: 1,
    companyId: '',
    contactId: '',
    opportunityId: '',
    quoteDate: new Date().toISOString().split('T')[0],
    validUntil: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
    currency: 'SAR',
    discountPercent: 0,
    vatPercent: 15,
    paymentTerms: 'Net 30',
    deliveryTerms: 'Ex-Works',
    customerNotes: '',
    internalNotes: '',
    status: 'draft'
  }
  lineItems.value = []
  lineItemCounter = 0
}

function handleClose() {
  if (confirm('Are you sure you want to close? Any unsaved changes will be lost.')) {
    emit('close')
  }
}

// Formatting functions
function formatCurrency(amount: number) {
  return new Intl.NumberFormat('en-SA', {
    style: 'currency',
    currency: 'SAR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(amount)
}

function getMarginClass(marginPercent: number) {
  if (marginPercent < 10) return 'text-danger'
  if (marginPercent < 15) return 'text-warning'
  return 'text-success'
}
</script>

<style scoped>
.quote-builder {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  max-height: 70vh;
  overflow-y: auto;
  padding-right: 0.5rem;
}

.builder-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid var(--color-border);
}

.builder-section:last-of-type {
  border-bottom: none;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.section-title {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
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
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-secondary);
}

.form-label.required::after {
  content: ' *';
  color: var(--color-danger);
}

.form-input,
.form-select,
.form-textarea {
  padding: 0.625rem 0.875rem;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text-primary);
  border-radius: 6px;
  font-size: 0.9375rem;
  transition: all 0.2s;
}

.form-input:focus,
.form-select:focus,
.form-textarea:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-input:disabled,
.form-select:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  background: var(--color-background);
}

.form-textarea {
  resize: vertical;
  font-family: inherit;
}

.btn {
  padding: 0.625rem 1.25rem;
  border: none;
  border-radius: 6px;
  font-weight: 500;
  font-size: 0.9375rem;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-sm {
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
}

.btn-primary {
  background: var(--color-primary);
  color: #fff;
}

.btn-primary:hover:not(:disabled) {
  background: var(--color-primary-hover);
  transform: translateY(-1px);
}

.btn-secondary {
  background: var(--color-border);
  color: var(--color-text-primary);
}

.btn-secondary:hover:not(:disabled) {
  background: var(--color-surface-hover);
}

.btn-outline {
  background: transparent;
  border: 1px solid var(--color-border);
  color: var(--color-text-primary);
}

.btn-outline:hover:not(:disabled) {
  background: var(--color-surface);
  border-color: var(--color-primary);
}

.empty-state {
  padding: 3rem;
  text-align: center;
  color: var(--color-text-secondary);
  background: var(--color-background);
  border-radius: 8px;
  border: 2px dashed var(--color-border);
}

.line-items-container {
  overflow-x: auto;
  border: 1px solid var(--color-border);
  border-radius: 8px;
}

.line-items-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 1000px;
}

.line-items-table th {
  background: var(--color-background);
  padding: 0.75rem 0.5rem;
  text-align: left;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 2px solid var(--color-border);
  white-space: nowrap;
}

.line-items-table td {
  padding: 0.75rem 0.5rem;
  border-bottom: 1px solid var(--color-border);
}

.line-item-row:hover {
  background: var(--color-background);
}

.line-number {
  text-align: center;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.product-cell {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-input-sm,
.form-select-sm {
  padding: 0.375rem 0.5rem;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text-primary);
  border-radius: 4px;
  font-size: 0.875rem;
  width: 100%;
}

.form-input-sm:focus,
.form-select-sm:focus {
  outline: none;
  border-color: var(--color-primary);
}

.text-right {
  text-align: right;
}

.line-total {
  font-weight: 600;
  color: var(--color-text-primary);
  text-align: right;
}

.margin-cell {
  text-align: right;
  font-weight: 600;
}

.btn-icon {
  background: none;
  border: none;
  font-size: 1.125rem;
  cursor: pointer;
  padding: 0.25rem;
  transition: transform 0.2s;
  opacity: 0.7;
}

.btn-icon:hover {
  transform: scale(1.2);
  opacity: 1;
}

.btn-danger-icon:hover {
  filter: brightness(1.2);
}

.summary-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  align-items: start;
}

.summary-right {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1.5rem;
  background: var(--color-background);
  border-radius: 8px;
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
  background: var(--color-surface);
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

.text-danger {
  color: var(--color-danger) !important;
}

.text-warning {
  color: var(--color-warning) !important;
}

.text-success {
  color: var(--color-success) !important;
}

.approval-warning {
  display: flex;
  gap: 1rem;
  padding: 1rem 1.25rem;
  background: rgba(234, 179, 8, 0.1);
  border: 1px solid var(--color-warning);
  border-radius: 8px;
}

.warning-icon {
  font-size: 1.5rem;
}

.warning-content {
  flex: 1;
}

.warning-content strong {
  display: block;
  color: var(--color-warning);
  margin-bottom: 0.25rem;
}

.warning-content p {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 0.9375rem;
}

.modal-footer-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  width: 100%;
}

@media (max-width: 768px) {
  .summary-layout {
    grid-template-columns: 1fr;
  }
}
</style>
