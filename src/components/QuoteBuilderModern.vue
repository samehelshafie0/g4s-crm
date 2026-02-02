<template>
  <div class="quote-builder-modern">
    <!-- Step 1: Client Selection -->
    <div class="modern-card">
      <div class="card-header">
        <div>
          <h3 class="card-title">🏢 Client & Pricing</h3>
          <p class="card-subtitle">Select client and optional price book</p>
        </div>
        <span class="step-badge">1</span>
      </div>

      <div class="card-body">
        <!-- Client Autocomplete -->
        <div class="form-field full-width">
          <label class="field-label required">Client</label>
          <div class="autocomplete-box">
            <input
              v-model="clientSearch"
              type="text"
              class="modern-input"
              placeholder="🔍 Search for a client..."
              @input="showClientSuggestions = true"
              @focus="showClientSuggestions = true"
              @blur="handleClientBlur"
            />
            <div
              v-if="showClientSuggestions && filteredClients.length > 0"
              class="suggestions-dropdown"
            >
              <div
                v-for="client in filteredClients"
                :key="client.id"
                class="suggestion-item"
                @mousedown.prevent="selectClient(client)"
              >
                <div class="suggestion-main">
                  <div class="avatar">{{ client.name[0] }}</div>
                  <div>
                    <div class="suggestion-name">{{ client.name }}</div>
                    <div class="suggestion-meta">{{ client.email }}</div>
                  </div>
                </div>
                <div
                  v-if="getClientPriceBooks(client.id).length > 0"
                  class="price-book-count"
                >
                  {{ getClientPriceBooks(client.id).length }} Price Books
                </div>
              </div>
            </div>
          </div>

          <!-- Selected Client Card -->
          <div v-if="selectedClient" class="selected-card">
            <div class="selected-avatar">{{ selectedClient.name[0] }}</div>
            <div class="selected-info">
              <strong>{{ selectedClient.name }}</strong>
              <span>{{ selectedClient.email }}</span>
            </div>
            <button class="clear-btn" @click="clearClient" title="Clear">
              ✕
            </button>
          </div>
        </div>

        <div class="form-row">
          <div class="form-field">
            <label class="field-label required">Contact</label>
            <select
              v-model="quoteData.contactId"
              class="modern-select"
              :disabled="!selectedClient"
              required
            >
              <option value="">Choose contact...</option>
              <option
                v-for="contact in availableContacts"
                :key="contact.id"
                :value="contact.id"
              >
                {{ contact.fullName }} - {{ contact.email }}
              </option>
            </select>
          </div>

          <div class="form-field">
            <label class="field-label">Price Book</label>
            <select
              v-model="selectedPriceBookId"
              class="modern-select"
              :disabled="!selectedClient"
              @change="onPriceBookChange"
            >
              <option value="">Manual Pricing</option>
              <option
                v-for="pb in availablePriceBooks"
                :key="pb.id"
                :value="pb.id"
              >
                {{ pb.name }}
              </option>
            </select>
            <p v-if="selectedPriceBook" class="field-hint">
              📋 {{ selectedPriceBook.description }}
            </p>
          </div>
        </div>

        <div class="form-row compact">
          <div class="form-field">
            <label class="field-label required">Quote Date</label>
            <input
              v-model="quoteData.quoteDate"
              type="date"
              class="modern-input"
              required
            />
          </div>
          <div class="form-field">
            <label class="field-label required">Valid Until</label>
            <input
              v-model="quoteData.validUntil"
              type="date"
              class="modern-input"
              :min="quoteData.quoteDate"
              required
            />
          </div>
          <div class="form-field">
            <label class="field-label">Currency</label>
            <select v-model="quoteData.currency" class="modern-select">
              <option value="SAR">SAR</option>
              <option value="USD">USD</option>
              <option value="EUR">EUR</option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <!-- Step 2: Line Items -->
    <div class="modern-card">
      <div class="card-header">
        <div>
          <h3 class="card-title">📦 Products & Services</h3>
          <p class="card-subtitle">Add items to quote</p>
        </div>
        <div class="header-actions">
          <span class="step-badge">2</span>
          <button class="add-btn" @click="addLineItem">+ Add Item</button>
        </div>
      </div>

      <div class="card-body">
        <div v-if="lineItems.length === 0" class="empty-box">
          <div class="empty-icon">📦</div>
          <p class="empty-text">No items added yet</p>
          <p class="empty-hint">
            Click "Add Item" to start building your quote
          </p>
        </div>

        <div v-else class="items-list">
          <div
            v-for="(item, index) in lineItems"
            :key="item.tempId"
            class="item-card"
          >
            <div class="item-number">{{ index + 1 }}</div>

            <div class="item-body">
              <!-- Item Type Selector -->
              <div class="type-row">
                <div class="type-tabs">
                  <label
                    class="type-tab"
                    :class="{ active: item.itemType === 'product' }"
                  >
                    <input
                      type="radio"
                      v-model="item.itemType"
                      value="product"
                      @change="onItemTypeChange(item)"
                    />
                    <span>Product</span>
                  </label>
                  <label
                    class="type-tab"
                    :class="{ active: item.itemType === 'service' }"
                  >
                    <input
                      type="radio"
                      v-model="item.itemType"
                      value="service"
                      @change="onItemTypeChange(item)"
                    />
                    <span>Service</span>
                  </label>
                  <label
                    class="type-tab"
                    :class="{ active: item.itemType === 'recurring' }"
                  >
                    <input
                      type="radio"
                      v-model="item.itemType"
                      value="recurring"
                      @change="onItemTypeChange(item)"
                    />
                    <span>Recurring</span>
                  </label>
                </div>

                <!-- Autocomplete Product Search -->
                <div class="autocomplete-container">
                  <input
                    v-model="item.searchQuery"
                    type="text"
                    class="item-select"
                    :placeholder="`🔍 Search ${item.itemType === 'recurring' ? 'recurring service' : item.itemType === 'service' ? 'labor/service' : 'product'} by name or SKU...`"
                    @input="handleProductSearch(item)"
                    @focus="handleProductSearchFocus(item)"
                    @blur="handleProductSearchBlur(item)"
                    autocomplete="off"
                  />
                  <div
                    v-if="
                      item.showSuggestions &&
                      getProductSuggestions(item).length > 0
                    "
                    class="autocomplete-dropdown"
                  >
                    <div class="autocomplete-header">
                      <span class="suggestion-count"
                        >{{ getProductSuggestions(item).length }} results</span
                      >
                    </div>
                    <div
                      v-for="suggestion in getProductSuggestions(item)"
                      :key="suggestion.id"
                      class="autocomplete-item"
                      @mousedown.prevent="selectProduct(item, suggestion)"
                    >
                      <div class="suggestion-main">
                        <span class="suggestion-icon">{{
                          suggestion.icon
                        }}</span>
                        <div class="suggestion-details">
                          <div class="suggestion-title">
                            <span
                              v-if="suggestion.sku"
                              class="suggestion-sku"
                              >{{ suggestion.sku }}</span
                            >
                            <span class="suggestion-name">{{
                              suggestion.name
                            }}</span>
                          </div>
                          <div
                            v-if="suggestion.manufacturer"
                            class="suggestion-manufacturer"
                          >
                            🏭 {{ suggestion.manufacturer }}
                          </div>
                          <div
                            v-if="item.itemType === 'product'"
                            class="suggestion-inventory"
                          >
                            <span
                              v-if="suggestion.hasStock"
                              class="stock-badge in-stock"
                            >
                              ✓ {{ suggestion.stockAvailable }} in stock
                            </span>
                            <span v-else class="stock-badge out-stock">
                              ⏳ Lead time: {{ suggestion.leadTime }} days
                            </span>
                          </div>
                          <div
                            v-if="suggestion.department"
                            class="suggestion-department"
                          >
                            {{ suggestion.department }}
                          </div>
                        </div>
                      </div>
                      <div class="suggestion-meta">
                        <span class="suggestion-price">{{
                          suggestion.priceLabel
                        }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <input
                v-model="item.description"
                type="text"
                class="item-input"
                placeholder="Description..."
              />

              <!-- Product Information Display -->
              <div
                v-if="item.productId && item.itemType === 'product'"
                class="product-info-bar"
              >
                <div v-if="item.manufacturer" class="info-item">
                  <span class="info-icon">🏭</span>
                  <span class="info-label">Manufacturer:</span>
                  <span class="info-value">{{ item.manufacturer }}</span>
                </div>
                <div v-if="item.stockAvailable !== undefined" class="info-item">
                  <span v-if="item.hasStock" class="info-icon">✓</span>
                  <span v-else class="info-icon">⏳</span>
                  <span class="info-label">{{
                    item.hasStock ? "In Stock:" : "Lead Time:"
                  }}</span>
                  <span
                    class="info-value"
                    :class="item.hasStock ? 'text-success' : 'text-warning'"
                  >
                    {{
                      item.hasStock
                        ? `${item.stockAvailable} units`
                        : `${item.leadTime} days`
                    }}
                  </span>
                </div>
              </div>

              <div class="details-grid">
                <div class="detail-box">
                  <label>Qty</label>
                  <input
                    v-model.number="item.quantity"
                    type="number"
                    min="1"
                    class="detail-input"
                    @input="calculateLineItem(item)"
                  />
                </div>
                <div class="detail-box">
                  <label>Unit Price</label>
                  <input
                    v-model.number="item.unitPrice"
                    type="number"
                    min="0"
                    step="0.01"
                    class="detail-input"
                    @input="calculateLineItem(item)"
                  />
                  <span v-if="item.priceBookPrice" class="pb-hint"
                    >PB: {{ formatCurrency(item.priceBookPrice) }}</span
                  >
                </div>
                <div class="detail-box">
                  <label>Disc %</label>
                  <input
                    v-model.number="item.discountPercent"
                    type="number"
                    min="0"
                    max="100"
                    class="detail-input"
                    @input="calculateLineItem(item)"
                  />
                </div>
                <div class="detail-box">
                  <label>Total</label>
                  <div class="total-display">
                    {{ formatCurrency(item.lineTotal) }}
                  </div>
                </div>
                <div class="detail-box">
                  <label>Cost</label>
                  <input
                    v-model.number="item.unitCost"
                    type="number"
                    min="0"
                    step="0.01"
                    class="detail-input cost"
                    @input="calculateLineItem(item)"
                  />
                </div>
                <div class="detail-box">
                  <label>Margin</label>
                  <div
                    class="margin-display"
                    :class="getMarginClass(item.marginPercent)"
                  >
                    {{ item.marginPercent.toFixed(1) }}%
                  </div>
                </div>
              </div>
            </div>

            <button class="remove-btn" @click="removeLineItem(index)">
              🗑️
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Step 3: Summary -->
    <div class="modern-card summary-card">
      <div class="card-header">
        <div>
          <h3 class="card-title">💰 Financial Summary</h3>
          <p class="card-subtitle">Totals and margins</p>
        </div>
        <span class="step-badge">3</span>
      </div>

      <div class="card-body">
        <div class="summary-layout">
          <div class="controls-col">
            <div class="control-box">
              <label>Additional Discount %</label>
              <input
                v-model.number="quoteData.discountPercent"
                type="number"
                min="0"
                max="100"
                class="modern-input"
                @input="calculateTotals"
              />
            </div>
            <div class="control-box">
              <label>VAT %</label>
              <input
                v-model.number="quoteData.vatPercent"
                type="number"
                min="0"
                max="100"
                class="modern-input"
                @input="calculateTotals"
              />
            </div>
          </div>

          <div class="totals-col">
            <div class="total-line">
              <span>Subtotal:</span>
              <span>{{ formatCurrency(calculatedSubtotal) }}</span>
            </div>
            <div
              v-if="quoteData.discountPercent > 0"
              class="total-line discount"
            >
              <span>Discount ({{ quoteData.discountPercent }}%):</span>
              <span>-{{ formatCurrency(calculatedDiscountAmount) }}</span>
            </div>
            <div class="total-line">
              <span>After Discount:</span>
              <span>{{ formatCurrency(calculatedSubtotalAfterDiscount) }}</span>
            </div>
            <div class="total-line">
              <span>VAT ({{ quoteData.vatPercent }}%):</span>
              <span>{{ formatCurrency(calculatedVatAmount) }}</span>
            </div>
            <div class="total-line grand">
              <span>Grand Total:</span>
              <span>{{ formatCurrency(calculatedTotal) }}</span>
            </div>
          </div>

          <div class="margin-col">
            <div class="margin-box">
              <span class="margin-label">Cost</span>
              <span class="margin-value cost">{{
                formatCurrency(calculatedTotalCost)
              }}</span>
            </div>
            <div class="margin-box">
              <span class="margin-label">Profit</span>
              <span class="margin-value profit">{{
                formatCurrency(calculatedMarginAmount)
              }}</span>
            </div>
            <div class="margin-box highlight">
              <span class="margin-label">Margin</span>
              <span
                class="margin-value"
                :class="getMarginClass(calculatedMarginPercent)"
                >{{ calculatedMarginPercent.toFixed(1) }}%</span
              >
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Step 4: Terms -->
    <div class="modern-card">
      <div class="card-header">
        <div>
          <h3 class="card-title">📝 Terms & Notes</h3>
          <p class="card-subtitle">Payment and delivery terms</p>
        </div>
        <span class="step-badge">4</span>
      </div>

      <div class="card-body">
        <div class="form-row">
          <div class="form-field">
            <label class="field-label">Payment Terms</label>
            <select v-model="quoteData.paymentTerms" class="modern-select">
              <option value="Net 30">Net 30 Days</option>
              <option value="Net 60">Net 60 Days</option>
              <option value="50% Advance, 50% on Delivery">
                50% Advance, 50% on Delivery
              </option>
              <option value="100% Advance">100% Advance</option>
            </select>
          </div>
          <div class="form-field">
            <label class="field-label">Delivery Terms</label>
            <select v-model="quoteData.deliveryTerms" class="modern-select">
              <option value="Ex-Works">Ex-Works</option>
              <option value="FOB">FOB</option>
              <option value="CIF">CIF</option>
              <option value="DDP">DDP</option>
            </select>
          </div>
        </div>

        <div class="form-field full-width">
          <label class="field-label">Customer Notes</label>
          <textarea
            v-model="quoteData.customerNotes"
            class="modern-textarea"
            placeholder="Notes visible to customer..."
          ></textarea>
        </div>

        <div class="form-field full-width">
          <label class="field-label">Internal Notes</label>
          <textarea
            v-model="quoteData.internalNotes"
            class="modern-textarea"
            placeholder="Internal notes..."
          ></textarea>
        </div>
      </div>
    </div>

    <!-- Step 5: Terms & Documents -->
    <div class="modern-card">
      <div class="card-header">
        <div>
          <h3 class="card-title">📄 Terms & Documents</h3>
          <p class="card-subtitle">
            Attach T&C and other documents to this quote
          </p>
        </div>
        <div class="header-actions">
          <span class="step-badge">5</span>
          <button class="add-btn" @click="showDocumentPicker = true">
            + Attach Document
          </button>
        </div>
      </div>

      <div class="card-body">
        <div v-if="attachedDocuments.length === 0" class="empty-box">
          <div class="empty-icon">📄</div>
          <p class="empty-text">No documents attached</p>
          <p class="empty-hint">
            Click "Attach Document" to add terms, conditions, or other documents
          </p>
        </div>

        <div v-else class="documents-list">
          <div
            v-for="doc in attachedDocuments"
            :key="doc.id"
            class="document-card"
          >
            <div class="document-icon">📄</div>
            <div class="document-info">
              <div class="document-name">{{ doc.name }}</div>
              <div class="document-type">{{ doc.type }}</div>
            </div>
            <div class="document-actions">
              <label class="include-checkbox">
                <input
                  type="checkbox"
                  :checked="doc.includeInPdf"
                  @change="toggleIncludeInPdf(doc.id)"
                />
                <span>Include in PDF</span>
              </label>
              <button
                class="remove-doc-btn"
                @click="removeAttachedDocument(doc.id)"
                title="Remove"
              >
                ✕
              </button>
            </div>
          </div>
        </div>

        <!-- Document Picker Modal -->
        <div
          v-if="showDocumentPicker"
          class="document-picker-overlay"
          @click="showDocumentPicker = false"
        >
          <div class="document-picker-modal" @click.stop>
            <div class="picker-header">
              <h4>Select Documents</h4>
              <button class="close-picker" @click="showDocumentPicker = false">
                ✕
              </button>
            </div>
            <div class="picker-search">
              <input
                v-model="documentSearchQuery"
                type="text"
                class="modern-input"
                placeholder="🔍 Search documents..."
              />
            </div>
            <div class="picker-list">
              <div
                v-for="doc in availableDocuments"
                :key="doc.id"
                class="picker-item"
                @click="attachDocument(doc)"
              >
                <div class="picker-icon">📄</div>
                <div class="picker-info">
                  <div class="picker-name">{{ doc.name }}</div>
                  <div class="picker-meta">
                    <span class="picker-type">{{ doc.type }}</span>
                    <span class="picker-category">{{ doc.category }}</span>
                  </div>
                </div>
              </div>
              <div v-if="availableDocuments.length === 0" class="picker-empty">
                No documents available
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Approval Warning -->
    <!-- Approval Warning -->
    <div v-if="requiresApproval" class="warning-banner">
      <div class="warning-icon">⚠️</div>
      <div>
        <strong>Approval Required</strong>
        <p>{{ approvalReason }}</p>
      </div>
    </div>

    <!-- Action Buttons -->
    <div class="footer-actions">
      <button class="btn-modern btn-ghost" @click="handleClose">Cancel</button>
      <button
        class="btn-modern btn-outline"
        @click="exportToExcel"
        :disabled="lineItems.length === 0"
        title="Export line items to Excel"
      >
        📊 Export Excel
      </button>
      <button
        class="btn-modern btn-outline"
        @click="exportToPdf"
        :disabled="!canSave"
        title="Export quote as PDF"
      >
        📄 Export PDF
      </button>
      <button
        class="btn-modern btn-outline"
        @click="saveAsDraft"
        :disabled="!canSave"
      >
        Save Draft
      </button>
      <button
        class="btn-modern btn-primary"
        @click="submitForApproval"
        :disabled="!canSave"
      >
        {{ requiresApproval ? "Submit for Approval" : "Create Quote" }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useCustomersStore } from "@/stores/customers";
import { useContactsStore } from "@/stores/contacts";
import { useProductsStore } from "@/stores/products";
import { useRecurringServicesStore } from "@/stores/recurringServices";
import { useLaborPositionsStore } from "@/stores/laborPositions";
import { usePriceBooksStore } from "@/stores/priceBooks";
import { useDocumentsStore } from "@/stores/documents";
import { useWarehouseStockStore } from "@/stores/warehouseStock";
import { useToast } from "@/composables/useToast";
import type { Quote } from "@/types";

interface LineItem {
  tempId: string;
  itemType: "product" | "recurring" | "service";
  productId: string;
  productName: string;
  description: string;
  quantity: number;
  unitPrice: number;
  unitCost: number;
  discountPercent: number;
  lineTotal: number;
  lineCost: number;
  marginAmount: number;
  marginPercent: number;
  priceBookPrice?: number;
  searchQuery?: string;
  showSuggestions?: boolean;
}

interface Props {
  show: boolean;
  quoteToEdit?: Quote | null;
}

interface Emits {
  (e: "close"): void;
  (e: "saved", quote: Quote): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

const customersStore = useCustomersStore();
const contactsStore = useContactsStore();
const productsStore = useProductsStore();
const recurringStore = useRecurringServicesStore();
const laborStore = useLaborPositionsStore();
const priceBooksStore = usePriceBooksStore();
const documentsStore = useDocumentsStore();
const warehouseStockStore = useWarehouseStockStore();
const { success, error: showError } = useToast();

const isEditMode = computed(() => !!props.quoteToEdit);

// Client
const clientSearch = ref("");
const showClientSuggestions = ref(false);
const selectedClient = ref<any>(null);

// Price Book
const selectedPriceBookId = ref("");
const selectedPriceBook = computed(() =>
  selectedPriceBookId.value
    ? priceBooksStore.getPriceBookById(selectedPriceBookId.value)
    : null,
);

const quoteData = ref({
  id: "",
  quoteNumber: "",
  version: 1,
  companyId: "",
  contactId: "",
  opportunityId: "",
  quoteDate: new Date().toISOString().split("T")[0],
  validUntil: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000)
    .toISOString()
    .split("T")[0],
  currency: "SAR",
  discountPercent: 0,
  vatPercent: 15,
  paymentTerms: "Net 30",
  deliveryTerms: "Ex-Works",
  customerNotes: "",
  internalNotes: "",
  status: "draft" as const,
  priceBookId: "",
});

const lineItems = ref<LineItem[]>([]);
let lineItemCounter = 0;

// Attached documents for Terms & Documents section
interface AttachedDocument {
  id: string;
  name: string;
  includeInPdf: boolean;
  type: string;
}

const attachedDocuments = ref<AttachedDocument[]>([]);
const showDocumentPicker = ref(false);
const documentSearchQuery = ref("");

const filteredClients = computed(() => {
  if (!clientSearch.value || clientSearch.value.length < 2) return [];
  return customersStore.customers
    .filter((c) =>
      c.name.toLowerCase().includes(clientSearch.value.toLowerCase()),
    )
    .slice(0, 10);
});

function handleClientBlur() {
  setTimeout(() => (showClientSuggestions.value = false), 200);
}

function selectClient(client: any) {
  selectedClient.value = client;
  quoteData.value.companyId = client.id;
  clientSearch.value = client.name;
  showClientSuggestions.value = false;
  quoteData.value.contactId = "";
  selectedPriceBookId.value = "";
}

function clearClient() {
  selectedClient.value = null;
  quoteData.value.companyId = "";
  clientSearch.value = "";
  quoteData.value.contactId = "";
  selectedPriceBookId.value = "";
}

function getClientPriceBooks(clientId: string) {
  return priceBooksStore.getPriceBooksByCustomerId(clientId);
}

const availableContacts = computed(() =>
  selectedClient.value
    ? contactsStore.getContactsByCompanyId(selectedClient.value.id)
    : [],
);
const availablePriceBooks = computed(() =>
  selectedClient.value ? getClientPriceBooks(selectedClient.value.id) : [],
);

const availableDocuments = computed(() => {
  if (!documentSearchQuery.value) {
    return documentsStore.documents.filter(
      (doc) =>
        !attachedDocuments.value.some((attached) => attached.id === doc.id),
    );
  }
  const query = documentSearchQuery.value.toLowerCase();
  return documentsStore.documents.filter(
    (doc) =>
      !attachedDocuments.value.some((attached) => attached.id === doc.id) &&
      (doc.name.toLowerCase().includes(query) ||
        doc.type.toLowerCase().includes(query) ||
        doc.category.toLowerCase().includes(query)),
  );
});

function attachDocument(doc: any) {
  attachedDocuments.value.push({
    id: doc.id,
    name: doc.name,
    includeInPdf: true,
    type: doc.type,
  });
  documentSearchQuery.value = "";
  showDocumentPicker.value = false;
}

function removeAttachedDocument(id: string) {
  const index = attachedDocuments.value.findIndex((d) => d.id === id);
  if (index !== -1) {
    attachedDocuments.value.splice(index, 1);
  }
}

function toggleIncludeInPdf(id: string) {
  const doc = attachedDocuments.value.find((d) => d.id === id);
  if (doc) {
    doc.includeInPdf = !doc.includeInPdf;
  }
}

function onPriceBookChange() {
  quoteData.value.priceBookId = selectedPriceBookId.value;
  lineItems.value.forEach((item) => {
    if (item.productId && selectedPriceBook.value) {
      const entry = priceBooksStore.getPriceBookEntry(
        selectedPriceBook.value.id,
        item.productId,
      );
      if (entry) {
        item.priceBookPrice = entry.customPrice;
        item.unitPrice = entry.customPrice;
        calculateLineItem(item);
      }
    }
  });
}

const calculatedSubtotal = computed(() =>
  lineItems.value.reduce((sum, item) => sum + item.lineTotal, 0),
);
const calculatedDiscountAmount = computed(
  () => calculatedSubtotal.value * (quoteData.value.discountPercent / 100),
);
const calculatedSubtotalAfterDiscount = computed(
  () => calculatedSubtotal.value - calculatedDiscountAmount.value,
);
const calculatedVatAmount = computed(
  () =>
    calculatedSubtotalAfterDiscount.value * (quoteData.value.vatPercent / 100),
);
const calculatedTotal = computed(
  () => calculatedSubtotalAfterDiscount.value + calculatedVatAmount.value,
);
const calculatedTotalCost = computed(() =>
  lineItems.value.reduce((sum, item) => sum + item.lineCost, 0),
);
const calculatedMarginAmount = computed(
  () => calculatedSubtotalAfterDiscount.value - calculatedTotalCost.value,
);
const calculatedMarginPercent = computed(() =>
  calculatedSubtotalAfterDiscount.value === 0
    ? 0
    : (calculatedMarginAmount.value / calculatedSubtotalAfterDiscount.value) *
      100,
);

const requiresApproval = computed(
  () =>
    calculatedMarginPercent.value < 10 || quoteData.value.discountPercent > 20,
);
const approvalReason = computed(() => {
  const reasons: string[] = [];
  if (calculatedMarginPercent.value < 10)
    reasons.push(`Low margin (${calculatedMarginPercent.value.toFixed(1)}%)`);
  if (quoteData.value.discountPercent > 20)
    reasons.push(`High discount (${quoteData.value.discountPercent}%)`);
  return reasons.join(", ");
});

const canSave = computed(() => {
  return (
    quoteData.value.companyId &&
    quoteData.value.contactId &&
    lineItems.value.length > 0 &&
    lineItems.value.every((item) => item.productId && item.quantity > 0)
  );
});

function addLineItem() {
  lineItems.value.push({
    tempId: `line-${lineItemCounter++}`,
    itemType: "product",
    productId: "",
    productName: "",
    description: "",
    quantity: 1,
    unitPrice: 0,
    unitCost: 0,
    discountPercent: 0,
    lineTotal: 0,
    lineCost: 0,
    marginAmount: 0,
    marginPercent: 0,
    searchQuery: "",
    showSuggestions: false,
  });
}

function removeLineItem(index: number) {
  lineItems.value.splice(index, 1);
}

function onItemTypeChange(item: LineItem) {
  item.productId = "";
  item.productName = "";
  item.unitPrice = 0;
  item.unitCost = 0;
  item.searchQuery = "";
  item.showSuggestions = false;
  calculateLineItem(item);
}

function onProductChange(item: LineItem) {
  if (item.itemType === "product") {
    const product = productsStore.getProductById(item.productId);
    if (product) {
      item.productName = product.name;
      item.description = product.description;
      item.unitCost = product.landedCostSAR;
      if (selectedPriceBook.value) {
        const entry = priceBooksStore.getPriceBookEntry(
          selectedPriceBook.value.id,
          item.productId,
        );
        if (entry) {
          item.priceBookPrice = entry.customPrice;
          item.unitPrice = entry.customPrice;
        } else {
          item.unitPrice = product.sellingPrice;
        }
      } else {
        item.unitPrice = product.sellingPrice;
      }
      calculateLineItem(item);
    }
  } else if (item.itemType === "service") {
    const position = laborStore.getPositionById(item.productId);
    if (position) {
      item.productName = position.name;
      item.description = position.description;
      item.unitPrice = position.sellingRatePerHour;
      item.unitCost = position.costPerHour;
      calculateLineItem(item);
    }
  } else {
    const service = recurringStore.getRecurringServiceById(item.productId);
    if (service) {
      item.productName = service.name;
      item.description = service.description;
      item.unitPrice = service.monthlyPrice;
      item.unitCost = service.monthlyCost;
      calculateLineItem(item);
    }
  }
}

// Autocomplete handlers for product search
function handleProductSearch(item: LineItem) {
  if (item.searchQuery && item.searchQuery.length >= 2) {
    item.showSuggestions = true;
  } else {
    item.showSuggestions = false;
  }
  // Clear selection if user types after selecting
  if (item.productId) {
    item.productId = "";
    item.productName = "";
  }
}

function handleProductSearchFocus(item: LineItem) {
  if (item.searchQuery && item.searchQuery.length >= 2) {
    item.showSuggestions = true;
  }
}

function handleProductSearchBlur(item: LineItem) {
  setTimeout(() => {
    item.showSuggestions = false;
  }, 200);
}

function getProductSuggestions(item: LineItem) {
  if (!item.searchQuery || item.searchQuery.length < 2) return [];

  const query = item.searchQuery.toLowerCase();
  let suggestions: any[] = [];

  if (item.itemType === "product") {
    suggestions = productsStore.products
      .filter(
        (p) =>
          p.name.toLowerCase().includes(query) ||
          p.sku?.toLowerCase().includes(query),
      )
      .map((p) => {
        // Get inventory information
        const totalAvailable = warehouseStockStore.getTotalAvailableForProduct(
          p.id,
        );
        const stockInfo = warehouseStockStore.getStockByProductId(p.id);
        const hasStock = totalAvailable > 0;

        return {
          id: p.id,
          name: p.name,
          sku: p.sku,
          icon: "📦",
          priceLabel: formatCurrency(p.sellingPrice),
          manufacturer: p.manufacturer,
          stockAvailable: totalAvailable,
          leadTime: p.leadTimeDays,
          hasStock: hasStock,
          stockLocations: stockInfo.length,
        };
      });
  } else if (item.itemType === "service") {
    suggestions = laborStore
      .getActivePositions()
      .filter(
        (pos) =>
          pos.name.toLowerCase().includes(query) ||
          pos.description.toLowerCase().includes(query) ||
          pos.department.toLowerCase().includes(query),
      )
      .map((pos) => ({
        id: pos.id,
        name: pos.name,
        sku: pos.id,
        icon: "👷",
        priceLabel: `${formatCurrency(pos.sellingRatePerHour)}/hr`,
        department: pos.department,
      }));
  } else {
    suggestions = recurringStore
      .getActiveRecurringServices()
      .filter((s) => s.name.toLowerCase().includes(query))
      .map((s) => ({
        id: s.id,
        name: s.name,
        sku: "",
        icon: "🔄",
        priceLabel: `${formatCurrency(s.monthlyPrice)}/mo`,
      }));
  }

  return suggestions.slice(0, 10);
}

function selectProduct(item: any, suggestion: any) {
  item.productId = suggestion.id;
  item.searchQuery = `${suggestion.sku ? suggestion.sku + " - " : ""}${suggestion.name}`;
  item.showSuggestions = false;

  // Store additional product information
  if (suggestion.manufacturer) {
    item.manufacturer = suggestion.manufacturer;
  }
  if (suggestion.stockAvailable !== undefined) {
    item.stockAvailable = suggestion.stockAvailable;
    item.hasStock = suggestion.hasStock;
  }
  if (suggestion.leadTime !== undefined) {
    item.leadTime = suggestion.leadTime;
  }

  onProductChange(item);
}

function calculateLineItem(item: LineItem) {
  const subtotal = item.quantity * item.unitPrice;
  const discount = subtotal * (item.discountPercent / 100);
  item.lineTotal = subtotal - discount;
  item.lineCost = item.quantity * item.unitCost;
  item.marginAmount = item.lineTotal - item.lineCost;
  item.marginPercent =
    item.lineTotal > 0 ? (item.marginAmount / item.lineTotal) * 100 : 0;
}

function calculateTotals() {
  // Computed properties handle this
}

async function saveAsDraft() {
  if (!canSave.value) return showError("Please fill required fields");
  success("Quote saved as draft");
  emit("close");
}

async function submitForApproval() {
  if (!canSave.value) return showError("Please fill required fields");
  const msg = requiresApproval.value
    ? "Quote submitted for approval"
    : "Quote created and approved";
  success(msg);
  emit("close");
}

function handleClose() {
  if (confirm("Close without saving?")) emit("close");
}

function formatCurrency(amount: number) {
  return new Intl.NumberFormat("en-SA", {
    style: "currency",
    currency: "SAR",
    minimumFractionDigits: 0,
  }).format(amount);
}

function getMarginClass(marginPercent: number) {
  if (marginPercent < 10) return "low";
  if (marginPercent < 15) return "ok";
  return "good";
}

function exportToExcel() {
  if (lineItems.value.length === 0) {
    showError("No line items to export");
    return;
  }

  // Create CSV content
  let csvContent = "data:text/csv;charset=utf-8,";

  // Headers
  const headers = [
    "Item #",
    "Type",
    "Product/Service",
    "Description",
    "Quantity",
    "Unit Price",
    "Discount %",
    "Line Total",
    "Unit Cost",
    "Line Cost",
    "Margin Amount",
    "Margin %",
  ];
  csvContent += headers.join(",") + "\n";

  // Data rows
  lineItems.value.forEach((item, index) => {
    const row = [
      index + 1,
      item.itemType,
      `"${item.productName || ""}"`,
      `"${item.description || ""}"`,
      item.quantity,
      item.unitPrice.toFixed(2),
      item.discountPercent.toFixed(2),
      item.lineTotal.toFixed(2),
      item.unitCost.toFixed(2),
      item.lineCost.toFixed(2),
      item.marginAmount.toFixed(2),
      item.marginPercent.toFixed(2),
    ];
    csvContent += row.join(",") + "\n";
  });

  // Add summary
  csvContent += "\n";
  csvContent += `Subtotal,,,,,,,${calculatedSubtotal.value.toFixed(2)}\n`;
  if (quoteData.value.discountPercent > 0) {
    csvContent += `Discount (${quoteData.value.discountPercent}%),,,,,,,${calculatedDiscountAmount.value.toFixed(2)}\n`;
  }
  csvContent += `After Discount,,,,,,,${calculatedSubtotalAfterDiscount.value.toFixed(2)}\n`;
  csvContent += `VAT (${quoteData.value.vatPercent}%),,,,,,,${calculatedVatAmount.value.toFixed(2)}\n`;
  csvContent += `Grand Total,,,,,,,${calculatedTotal.value.toFixed(2)}\n`;
  csvContent += `\n`;
  csvContent += `Total Cost,,,,,,,${calculatedTotalCost.value.toFixed(2)}\n`;
  csvContent += `Total Profit,,,,,,,${calculatedMarginAmount.value.toFixed(2)}\n`;
  csvContent += `Margin %,,,,,,,${calculatedMarginPercent.value.toFixed(2)}%\n`;

  // Create download
  const encodedUri = encodeURI(csvContent);
  const link = document.createElement("a");
  link.setAttribute("href", encodedUri);
  const fileName = `quote-${selectedClient.value?.name || "draft"}-${new Date().toISOString().split("T")[0]}.csv`;
  link.setAttribute("download", fileName);
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  success("Excel export downloaded successfully");
}

async function exportToPdf() {
  if (!canSave.value) {
    showError("Please complete all required fields");
    return;
  }

  try {
    // Dynamic import to avoid bundling if not used
    const { jsPDF } = await import("jspdf");
    const { default: autoTable } = await import("jspdf-autotable");

    const doc = new jsPDF();
    let yPosition = 20;

    // Header
    doc.setFontSize(20);
    doc.setFont("helvetica", "bold");
    doc.text("QUOTATION", 105, yPosition, { align: "center" });
    yPosition += 15;

    // Client Information
    doc.setFontSize(12);
    doc.setFont("helvetica", "bold");
    doc.text("Client Information:", 20, yPosition);
    yPosition += 7;

    doc.setFontSize(10);
    doc.setFont("helvetica", "normal");
    doc.text(`Client: ${selectedClient.value?.name || "N/A"}`, 20, yPosition);
    yPosition += 5;
    doc.text(
      `Contact: ${availableContacts.value.find((c) => c.id === quoteData.value.contactId)?.fullName || "N/A"}`,
      20,
      yPosition,
    );
    yPosition += 5;
    doc.text(`Date: ${quoteData.value.quoteDate}`, 20, yPosition);
    yPosition += 5;
    doc.text(`Valid Until: ${quoteData.value.validUntil}`, 20, yPosition);
    yPosition += 5;
    doc.text(`Currency: ${quoteData.value.currency}`, 20, yPosition);
    yPosition += 10;

    // Line Items Table
    const tableData = lineItems.value.map((item, index) => [
      index + 1,
      item.itemType.toUpperCase(),
      item.productName,
      item.quantity,
      formatCurrency(item.unitPrice),
      `${item.discountPercent}%`,
      formatCurrency(item.lineTotal),
      formatCurrency(item.marginAmount),
      `${item.marginPercent.toFixed(1)}%`,
    ]);

    autoTable(doc, {
      startY: yPosition,
      head: [
        [
          "#",
          "Type",
          "Item",
          "Qty",
          "Unit Price",
          "Disc",
          "Total",
          "Margin",
          "Margin %",
        ],
      ],
      body: tableData,
      theme: "grid",
      headStyles: {
        fillColor: [59, 130, 246],
        textColor: 255,
        fontStyle: "bold",
      },
      styles: { fontSize: 9 },
      columnStyles: {
        0: { cellWidth: 10 },
        1: { cellWidth: 20 },
        2: { cellWidth: 60 },
        3: { cellWidth: 15 },
        4: { cellWidth: 25 },
        5: { cellWidth: 15 },
        6: { cellWidth: 25 },
        7: { cellWidth: 25 },
        8: { cellWidth: 20 },
      },
    });

    yPosition = (doc as any).lastAutoTable.finalY + 10;

    // Summary
    doc.setFontSize(10);
    doc.setFont("helvetica", "bold");
    const summaryX = 140;

    doc.text("Subtotal:", summaryX, yPosition);
    doc.text(formatCurrency(calculatedSubtotal.value), 185, yPosition, {
      align: "right",
    });
    yPosition += 6;

    if (quoteData.value.discountPercent > 0) {
      doc.text(
        `Discount (${quoteData.value.discountPercent}%):`,
        summaryX,
        yPosition,
      );
      doc.text(
        `-${formatCurrency(calculatedDiscountAmount.value)}`,
        185,
        yPosition,
        { align: "right" },
      );
      yPosition += 6;
    }

    doc.text("After Discount:", summaryX, yPosition);
    doc.text(
      formatCurrency(calculatedSubtotalAfterDiscount.value),
      185,
      yPosition,
      { align: "right" },
    );
    yPosition += 6;

    doc.text(`VAT (${quoteData.value.vatPercent}%):`, summaryX, yPosition);
    doc.text(formatCurrency(calculatedVatAmount.value), 185, yPosition, {
      align: "right",
    });
    yPosition += 6;

    doc.setFontSize(12);
    doc.text("Grand Total:", summaryX, yPosition);
    doc.text(formatCurrency(calculatedTotal.value), 185, yPosition, {
      align: "right",
    });
    yPosition += 10;

    // Terms & Conditions
    if (
      quoteData.value.paymentTerms ||
      quoteData.value.deliveryTerms ||
      quoteData.value.customerNotes
    ) {
      yPosition += 5;
      if (yPosition > 250) {
        doc.addPage();
        yPosition = 20;
      }

      doc.setFontSize(12);
      doc.setFont("helvetica", "bold");
      doc.text("Terms & Conditions:", 20, yPosition);
      yPosition += 7;

      doc.setFontSize(10);
      doc.setFont("helvetica", "normal");

      if (quoteData.value.paymentTerms) {
        doc.text(
          `Payment Terms: ${quoteData.value.paymentTerms}`,
          20,
          yPosition,
        );
        yPosition += 5;
      }

      if (quoteData.value.deliveryTerms) {
        doc.text(
          `Delivery Terms: ${quoteData.value.deliveryTerms}`,
          20,
          yPosition,
        );
        yPosition += 5;
      }

      if (quoteData.value.customerNotes) {
        yPosition += 3;
        doc.text("Notes:", 20, yPosition);
        yPosition += 5;
        const splitNotes = doc.splitTextToSize(
          quoteData.value.customerNotes,
          170,
        );
        doc.text(splitNotes, 20, yPosition);
        yPosition += splitNotes.length * 5;
      }
    }

    // Attached Documents Notice
    const documentsToInclude = attachedDocuments.value.filter(
      (d) => d.includeInPdf,
    );
    if (documentsToInclude.length > 0) {
      yPosition += 10;
      if (yPosition > 250) {
        doc.addPage();
        yPosition = 20;
      }

      doc.setFontSize(12);
      doc.setFont("helvetica", "bold");
      doc.text("Attached Documents:", 20, yPosition);
      yPosition += 7;

      doc.setFontSize(10);
      doc.setFont("helvetica", "normal");
      documentsToInclude.forEach((doc) => {
        doc.text(`• ${doc.name} (${doc.type})`, 25, yPosition);
        yPosition += 5;
      });

      // Note about document appending
      yPosition += 5;
      doc.setFontSize(9);
      doc.setFont("helvetica", "italic");
      doc.text(
        "Note: Full document content would be appended in production implementation.",
        20,
        yPosition,
      );
      doc.text(
        "This requires PDF merging capabilities with actual document files.",
        20,
        yPosition + 4,
      );
    }

    // Save PDF
    const fileName = `Quote-${selectedClient.value?.name.replace(/\s+/g, "-") || "Draft"}-${new Date().toISOString().split("T")[0]}.pdf`;
    doc.save(fileName);

    success("PDF exported successfully");
  } catch (error) {
    console.error("PDF export error:", error);
    showError("Failed to export PDF");
  }
}

watch(
  () => props.show,
  (show) => {
    if (show && !props.quoteToEdit) {
      selectedClient.value = null;
      clientSearch.value = "";
      selectedPriceBookId.value = "";
      lineItems.value = [];
    }
  },
);
</script>

<style scoped>
.quote-builder-modern {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  padding: 0.5rem;
}

/* Modern Card */
.modern-card {
  background: var(--color-surface);
  border-radius: 20px;
  overflow: hidden;
  box-shadow:
    0 4px 6px -1px rgba(0, 0, 0, 0.03),
    0 2px 4px -1px rgba(0, 0, 0, 0.06);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.modern-card:hover {
  box-shadow:
    0 10px 15px -3px rgba(0, 0, 0, 0.07),
    0 4px 6px -2px rgba(0, 0, 0, 0.05);
}

.summary-card {
  background: linear-gradient(
    135deg,
    rgba(59, 130, 246, 0.03) 0%,
    rgba(147, 51, 234, 0.03) 100%
  );
  border: 1px solid rgba(59, 130, 246, 0.08);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 1.5rem 2rem;
  border-bottom: 1px solid var(--color-border);
  background: linear-gradient(to bottom, rgba(255, 255, 255, 0.5), transparent);
}

.card-title {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: -0.025em;
}

.card-subtitle {
  margin: 0.25rem 0 0;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  font-weight: 400;
}

.step-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
  font-size: 0.875rem;
  font-weight: 700;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.header-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.add-btn {
  padding: 0.625rem 1.25rem;
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
  border: none;
  border-radius: 10px;
  font-weight: 600;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.25);
}

.add-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.35);
}

.card-body {
  padding: 2rem;
}

/* Form Fields */
.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.25rem;
  margin-bottom: 1.25rem;
}

.form-row.compact {
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-field.full-width {
  grid-column: 1 / -1;
}

.field-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.field-label.required::after {
  content: " *";
  color: #ef4444;
}

.modern-input,
.modern-select,
.modern-textarea {
  padding: 0.875rem 1rem;
  border: 2px solid var(--color-border);
  background: var(--color-background);
  color: var(--color-text-primary);
  border-radius: 12px;
  font-size: 0.9375rem;
  transition: all 0.2s;
  font-family: inherit;
}

.modern-input:focus,
.modern-select:focus,
.modern-textarea:focus {
  outline: none;
  border-color: #3b82f6;
  background: var(--color-surface);
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.08);
}

.modern-textarea {
  resize: vertical;
  min-height: 80px;
}

.field-hint {
  margin-top: 0.5rem;
  padding: 0.75rem;
  background: rgba(59, 130, 246, 0.05);
  border-left: 3px solid #3b82f6;
  border-radius: 6px;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

/* Autocomplete */
.autocomplete-box {
  position: relative;
}

.suggestions-dropdown {
  position: absolute;
  top: calc(100% + 0.5rem);
  left: 0;
  right: 0;
  background: var(--color-surface);
  border: 2px solid var(--color-border);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  max-height: 300px;
  overflow-y: auto;
  z-index: 1000;
}

.suggestion-item {
  padding: 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  border-bottom: 1px solid var(--color-border);
  transition: all 0.2s;
}

.suggestion-item:last-child {
  border-bottom: none;
}

.suggestion-item:hover {
  background: var(--color-background);
  padding-left: 1.25rem;
}

.suggestion-main {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.avatar {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 1.125rem;
}

.suggestion-name {
  font-weight: 600;
  color: var(--color-text-primary);
}

.suggestion-meta {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.price-book-count {
  padding: 0.375rem 0.75rem;
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
  border-radius: 8px;
  font-size: 0.75rem;
  font-weight: 600;
}

/* Selected Client */
.selected-card {
  margin-top: 0.75rem;
  padding: 1rem;
  background: linear-gradient(
    135deg,
    rgba(59, 130, 246, 0.08),
    rgba(147, 51, 234, 0.08)
  );
  border: 2px solid rgba(59, 130, 246, 0.15);
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.selected-avatar {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 1.5rem;
  flex-shrink: 0;
}

.selected-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.selected-info strong {
  font-size: 1rem;
  color: var(--color-text-primary);
}

.selected-info span {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.clear-btn {
  padding: 0.5rem;
  background: rgba(239, 68, 68, 0.1);
  border: none;
  border-radius: 8px;
  color: #ef4444;
  cursor: pointer;
  font-size: 1.125rem;
  transition: all 0.2s;
}

.clear-btn:hover {
  background: rgba(239, 68, 68, 0.2);
  transform: scale(1.1);
}

/* Empty State */
.empty-box {
  padding: 4rem 2rem;
  text-align: center;
  background: var(--color-background);
  border-radius: 16px;
  border: 2px dashed var(--color-border);
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}

.empty-text {
  margin: 0 0 0.5rem;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.empty-hint {
  margin: 0;
  color: var(--color-text-secondary);
}

/* Items List */
.items-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.item-card {
  position: relative;
  display: flex;
  gap: 1rem;
  padding: 1.5rem;
  background: var(--color-background);
  border: 2px solid var(--color-border);
  border-radius: 16px;
  transition: all 0.3s;
}

.item-card:hover {
  border-color: #3b82f6;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.06);
}

.item-number {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  flex-shrink: 0;
}

.item-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  min-width: 0;
}

/* Product Info Bar */
.product-info-bar {
  display: flex;
  gap: 1.5rem;
  padding: 0.75rem 1rem;
  background: linear-gradient(
    135deg,
    rgba(59, 130, 246, 0.05),
    rgba(139, 92, 246, 0.05)
  );
  border: 1px solid rgba(59, 130, 246, 0.15);
  border-radius: 10px;
  flex-wrap: wrap;
}

.product-info-bar .info-item {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.875rem;
}

.product-info-bar .info-icon {
  font-size: 1rem;
}

.product-info-bar .info-label {
  font-weight: 600;
  color: var(--color-text-secondary);
}

.product-info-bar .info-value {
  font-weight: 600;
  color: var(--color-text-primary);
}

.product-info-bar .info-value.text-success {
  color: #10b981;
}

.product-info-bar .info-value.text-warning {
  color: #f59e0b;
}

.type-row {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.type-tabs {
  display: flex;
  gap: 0;
  background: var(--color-surface);
  border-radius: 10px;
  padding: 0.25rem;
}

.type-tab {
  padding: 0.5rem 1.25rem;
  cursor: pointer;
  border-radius: 8px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 500;
  font-size: 0.875rem;
}

.type-tab input {
  display: none;
}

.type-tab.active {
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
}

.item-select {
  flex: 1;
  min-width: 250px;
  padding: 0.75rem 1rem;
  border: 2px solid var(--color-border);
  background: var(--color-surface);
  border-radius: 10px;
  font-size: 0.9375rem;
  transition: all 0.2s;
}

.item-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.08);
}

/* Autocomplete for product selection */
.autocomplete-container {
  position: relative;
  flex: 1;
  min-width: 250px;
}

.autocomplete-dropdown {
  position: absolute;
  top: calc(100% + 0.5rem);
  left: 0;
  right: 0;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
  max-height: 300px;
  overflow-y: auto;
  z-index: 1000;
  animation: slideDown 0.2s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.autocomplete-header {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--color-border);
  background: var(--color-background);
  border-radius: 10px 10px 0 0;
}

.suggestion-count {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.autocomplete-item {
  padding: 0.875rem 1rem;
  cursor: pointer;
  border-bottom: 1px solid var(--color-border);
  transition: background-color 0.15s;
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.autocomplete-item:last-child {
  border-bottom: none;
  border-radius: 0 0 10px 10px;
}

.autocomplete-item:hover {
  background: rgba(59, 130, 246, 0.05);
}

.suggestion-main {
  display: flex;
  align-items: flex-start;
  gap: 0.625rem;
  width: 100%;
  justify-content: space-between;
}

.suggestion-icon {
  font-size: 1.125rem;
  flex-shrink: 0;
  margin-top: 2px;
}

.suggestion-details {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.suggestion-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.suggestion-sku {
  font-family: "Courier New", monospace;
  font-weight: 700;
  color: var(--color-primary);
  font-size: 0.8125rem;
  padding: 0.125rem 0.375rem;
  background: rgba(59, 130, 246, 0.1);
  border-radius: 4px;
}

.suggestion-name {
  color: var(--color-text-primary);
  font-weight: 500;
  font-size: 0.9375rem;
}

.suggestion-manufacturer {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.suggestion-inventory {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.stock-badge {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.stock-badge.in-stock {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

.stock-badge.out-stock {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
}

.suggestion-department {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.suggestion-meta {
  display: flex;
  gap: 0.75rem;
  align-items: flex-start;
  flex-shrink: 0;
  margin-top: 2px;
}

.suggestion-price {
  font-size: 0.875rem;
  color: #10b981;
  font-weight: 700;
  white-space: nowrap;
}

.item-input {
  padding: 0.75rem 1rem;
  border: 2px solid var(--color-border);
  background: var(--color-surface);
  border-radius: 10px;
  transition: all 0.2s;
}

.item-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.08);
}

.details-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 0.75rem;
}

.detail-box {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.detail-box label {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.detail-input {
  padding: 0.625rem 0.75rem;
  border: 2px solid var(--color-border);
  background: var(--color-surface);
  border-radius: 8px;
  text-align: right;
  font-weight: 600;
  transition: all 0.2s;
}

.detail-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.08);
}

.detail-input.cost {
  background: rgba(239, 68, 68, 0.04);
  border-color: rgba(239, 68, 68, 0.15);
}

.pb-hint {
  font-size: 0.7rem;
  color: var(--color-text-secondary);
  font-style: italic;
}

.total-display {
  padding: 0.625rem 0.75rem;
  background: linear-gradient(
    135deg,
    rgba(16, 185, 129, 0.08),
    rgba(59, 130, 246, 0.08)
  );
  border: 2px solid rgba(16, 185, 129, 0.15);
  border-radius: 8px;
  text-align: right;
  font-weight: 700;
  color: #10b981;
}

.margin-display {
  padding: 0.625rem 0.75rem;
  border-radius: 8px;
  text-align: right;
  font-weight: 700;
  border: 2px solid;
}

.margin-display.low {
  background: rgba(239, 68, 68, 0.08);
  color: #ef4444;
  border-color: rgba(239, 68, 68, 0.15);
}

.margin-display.ok {
  background: rgba(245, 158, 11, 0.08);
  color: #f59e0b;
  border-color: rgba(245, 158, 11, 0.15);
}

.margin-display.good {
  background: rgba(16, 185, 129, 0.08);
  color: #10b981;
  border-color: rgba(16, 185, 129, 0.15);
}

.remove-btn {
  position: absolute;
  top: 1rem;
  right: 1rem;
  padding: 0.5rem;
  background: rgba(239, 68, 68, 0.08);
  border: none;
  border-radius: 8px;
  color: #ef4444;
  cursor: pointer;
  font-size: 1.125rem;
  transition: all 0.2s;
}

.remove-btn:hover {
  background: rgba(239, 68, 68, 0.15);
  transform: scale(1.1);
}

/* Summary */
.summary-layout {
  display: grid;
  grid-template-columns: 1fr 1.5fr 1fr;
  gap: 1.5rem;
}

.controls-col {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.control-box {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.control-box label {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.totals-col {
  padding: 1.5rem;
  background: var(--color-background);
  border-radius: 16px;
  border: 2px solid var(--color-border);
}

.total-line {
  display: flex;
  justify-content: space-between;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--color-border);
  font-size: 0.9375rem;
}

.total-line.discount {
  color: #ef4444;
}

.total-line.grand {
  border-top: 3px solid var(--color-border);
  padding-top: 1rem;
  margin-top: 0.5rem;
  font-size: 1.25rem;
  font-weight: 700;
}

.margin-col {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.margin-box {
  padding: 1rem;
  background: var(--color-background);
  border-radius: 12px;
  border: 2px solid var(--color-border);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.margin-box.highlight {
  background: linear-gradient(
    135deg,
    rgba(16, 185, 129, 0.08),
    rgba(59, 130, 246, 0.08)
  );
  border-color: rgba(16, 185, 129, 0.2);
}

.margin-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.margin-value {
  font-size: 1.25rem;
  font-weight: 700;
}

.margin-value.cost {
  color: #ef4444;
}

.margin-value.profit {
  color: #10b981;
}

/* Warning Banner */
.warning-banner {
  display: flex;
  gap: 1rem;
  padding: 1.25rem;
  background: rgba(245, 158, 11, 0.08);
  border: 2px solid rgba(245, 158, 11, 0.2);
  border-radius: 16px;
}

.warning-icon {
  font-size: 1.5rem;
}

.warning-banner strong {
  display: block;
  color: #f59e0b;
  margin-bottom: 0.25rem;
}

.warning-banner p {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 0.9375rem;
}

/* Footer */
.footer-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
}

.btn-modern {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 12px;
  font-weight: 600;
  font-size: 0.9375rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-modern:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-modern.btn-ghost {
  background: transparent;
  color: var(--color-text-secondary);
}

.btn-modern.btn-ghost:hover {
  background: var(--color-background);
}

.btn-modern.btn-outline {
  background: transparent;
  border: 2px solid var(--color-border);
  color: var(--color-text-primary);
}

.btn-modern.btn-outline:hover {
  background: var(--color-surface);
  border-color: #3b82f6;
}

.btn-modern.btn-primary {
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.btn-modern.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(59, 130, 246, 0.4);
}

/* Terms & Documents Section */
.documents-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.document-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.25rem;
  background: var(--color-background);
  border: 2px solid var(--color-border);
  border-radius: 12px;
  transition: all 0.2s;
}

.document-card:hover {
  border-color: #3b82f6;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.06);
}

.document-icon {
  font-size: 2rem;
  flex-shrink: 0;
}

.document-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.document-name {
  font-weight: 600;
  color: var(--color-text-primary);
  font-size: 0.9375rem;
}

.document-type {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  text-transform: capitalize;
}

.document-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.include-checkbox {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.include-checkbox input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.remove-doc-btn {
  padding: 0.5rem;
  background: rgba(239, 68, 68, 0.08);
  border: none;
  border-radius: 8px;
  color: #ef4444;
  cursor: pointer;
  font-size: 1rem;
  transition: all 0.2s;
}

.remove-doc-btn:hover {
  background: rgba(239, 68, 68, 0.15);
  transform: scale(1.1);
}

/* Document Picker Modal */
.document-picker-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  animation: fadeIn 0.2s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.document-picker-modal {
  width: 90%;
  max-width: 600px;
  max-height: 80vh;
  background: var(--color-surface);
  border-radius: 16px;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
  animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.picker-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid var(--color-border);
}

.picker-header h4 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.close-picker {
  padding: 0.5rem;
  background: rgba(239, 68, 68, 0.08);
  border: none;
  border-radius: 8px;
  color: #ef4444;
  cursor: pointer;
  font-size: 1.25rem;
  transition: all 0.2s;
}

.close-picker:hover {
  background: rgba(239, 68, 68, 0.15);
  transform: scale(1.1);
}

.picker-search {
  padding: 1.25rem;
  border-bottom: 1px solid var(--color-border);
}

.picker-list {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
}

.picker-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  background: var(--color-background);
  border: 2px solid var(--color-border);
  border-radius: 12px;
  margin-bottom: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.picker-item:hover {
  border-color: #3b82f6;
  background: var(--color-surface);
  transform: translateX(4px);
}

.picker-icon {
  font-size: 1.75rem;
  flex-shrink: 0;
}

.picker-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.picker-name {
  font-weight: 600;
  color: var(--color-text-primary);
  font-size: 0.9375rem;
}

.picker-meta {
  display: flex;
  gap: 0.75rem;
  font-size: 0.75rem;
}

.picker-type,
.picker-category {
  padding: 0.25rem 0.5rem;
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
  border-radius: 4px;
  text-transform: capitalize;
  font-weight: 500;
}

.picker-category {
  background: rgba(147, 51, 234, 0.1);
  color: #9333ea;
}

.picker-empty {
  text-align: center;
  padding: 3rem 1rem;
  color: var(--color-text-secondary);
  font-size: 0.9375rem;
}

@media (max-width: 1024px) {
  .summary-layout {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .form-row {
    grid-template-columns: 1fr;
  }

  .details-grid {
    grid-template-columns: 1fr 1fr;
  }

  .document-picker-modal {
    width: 95%;
    max-height: 90vh;
  }
}
</style>
