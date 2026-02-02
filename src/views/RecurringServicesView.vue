<template>
  <div class="recurring-services-view">
    <div class="view-header">
      <div class="header-actions">
        <input
          v-model="searchQuery"
          type="search"
          placeholder="Search recurring services..."
          class="search-input"
        />
        <button class="btn btn-primary" @click="createNewService">
          <span>➕</span> New Recurring Service
        </button>
      </div>
    </div>

    <div class="filters-bar">
      <div class="filter-group">
        <label>Type:</label>
        <select v-model="typeFilter" class="filter-select">
          <option value="all">All Types</option>
          <option value="product">Product Rental</option>
          <option value="labor">Labor Service</option>
        </select>
      </div>

      <div class="filter-group">
        <label>Status:</label>
        <select v-model="statusFilter" class="filter-select">
          <option value="all">All Status</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
        </select>
      </div>

      <div class="stats-summary">
        <div class="stat-item">
          <span class="stat-label">Total Services:</span>
          <span class="stat-value">{{ filteredServices.length }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Active:</span>
          <span class="stat-value text-success">{{ activeServicesCount }}</span>
        </div>
      </div>
    </div>

    <BaseCard no-padding>
      <BaseTable
        :columns="servicesColumns"
        :data="filteredServices"
        :clickable="true"
        @row-click="viewServiceDetails"
      >
        <template #cell-name="{ row }">
          <div class="name-cell">
            <div class="name">{{ row.name }}</div>
            <div class="description">{{ row.description }}</div>
          </div>
        </template>

        <template #cell-type="{ value }">
          <BaseBadge
            :variant="value === 'product' ? 'info' : 'warning'"
            size="sm"
          >
            {{ value === "product" ? "📦 Product Rental" : "👷 Labor Service" }}
          </BaseBadge>
        </template>

        <template #cell-source="{ row }">
          <div class="source-cell">
            <div class="source-header">
              <span class="source-name">{{ row.sourceName }}</span>
              <BaseBadge
                :variant="row.source === 'owned' ? 'success' : 'warning'"
                size="sm"
                class="source-badge"
              >
                {{ row.source === "owned" ? "🏢 Owned" : "🏭 Rented" }}
              </BaseBadge>
            </div>
            <div v-if="row.sourceId" class="source-id">{{ row.sourceId }}</div>
            <div
              v-if="row.source === 'rented' && row.vendorQuote"
              class="vendor-info"
            >
              Vendor: {{ row.vendorQuote.vendorName }}
            </div>
          </div>
        </template>

        <template #cell-pricing="{ row }">
          <div class="pricing-cell">
            <div class="monthly-price">
              {{ formatCurrency(row.monthlyPrice) }}/month
            </div>
            <div class="min-period">Min: {{ row.minimumMonths }} months</div>
            <div v-if="row.setupFee" class="setup-fee">
              Setup: {{ formatCurrency(row.setupFee) }}
            </div>
          </div>
        </template>

        <template #cell-margin="{ row }">
          <div class="margin-cell">
            <span :class="getMarginClass(calculateMargin(row))">
              {{ calculateMargin(row).toFixed(1) }}%
            </span>
          </div>
        </template>

        <template #cell-status="{ value }">
          <BaseBadge
            :variant="value === 'active' ? 'success' : 'default'"
            size="sm"
          >
            {{ value === "active" ? "Active" : "Inactive" }}
          </BaseBadge>
        </template>

        <template #cell-actions="{ row }">
          <div class="action-buttons">
            <button
              class="action-btn"
              title="View Details"
              @click.stop="viewServiceDetails(row)"
            >
              👁️
            </button>
            <button
              class="action-btn"
              title="Edit"
              @click.stop="editService(row)"
            >
              ✏️
            </button>
            <button
              class="action-btn"
              title="Duplicate"
              @click.stop="duplicateService(row)"
            >
              📋
            </button>
            <button
              class="action-btn"
              :title="row.status === 'active' ? 'Deactivate' : 'Activate'"
              @click.stop="toggleStatus(row)"
            >
              {{ row.status === "active" ? "⏸️" : "▶️" }}
            </button>
          </div>
        </template>
      </BaseTable>
    </BaseCard>

    <!-- Create/Edit Service Modal -->
    <BaseModal
      v-model="showServiceModal"
      :title="
        editingService
          ? 'Edit Recurring Service'
          : 'Create New Recurring Service'
      "
      size="lg"
      hide-confirm
      cancel-text="Cancel"
    >
      <form @submit.prevent="saveService" class="service-form">
        <div class="form-section">
          <h4 class="section-title">Basic Information</h4>

          <div class="form-group">
            <label class="form-label">Service Type *</label>
            <select
              v-model="serviceForm.type"
              class="form-input"
              required
              @change="onTypeChange"
            >
              <option value="">-- Select Type --</option>
              <option value="product">📦 Product Rental</option>
              <option value="labor">👷 Labor Service</option>
            </select>
          </div>

          <div class="form-group">
            <label class="form-label">Source Type *</label>
            <select
              v-model="serviceForm.source"
              class="form-input"
              required
              @change="onSourceTypeChange"
            >
              <option value="">-- Select Source --</option>
              <option value="owned">🏢 Owned (Our Products/Services)</option>
              <option value="rented">🏭 Rented (From Vendor)</option>
            </select>
            <small class="form-hint">
              Owned: Your own products/services | Rented: Equipment/services
              leased from vendor
            </small>
          </div>

          <!-- Source Selection: Only for OWNED products/services -->
          <div
            v-if="serviceForm.source === 'owned' && serviceForm.type"
            class="form-group autocomplete-container"
          >
            <label class="form-label"
              >Search
              {{
                serviceForm.type === "product" ? "Product" : "Labor Position"
              }}
              *</label
            >
            <div class="autocomplete-wrapper">
              <input
                v-model="sourceSearchQuery"
                type="text"
                class="form-input"
                :placeholder="`Search by name or SKU...`"
                :required="serviceForm.source === 'owned'"
                @input="handleSourceSearchInput"
                @focus="showSourceSuggestions = true"
                @blur="handleSourceSearchBlur"
                autocomplete="off"
              />
              <div
                v-if="
                  showSourceSuggestions && filteredSourceSuggestions.length > 0
                "
                class="autocomplete-dropdown"
              >
                <div class="autocomplete-header">
                  <span class="suggestion-count"
                    >{{ filteredSourceSuggestions.length }} results found</span
                  >
                </div>
                <div
                  v-for="item in filteredSourceSuggestions"
                  :key="item.id"
                  class="autocomplete-item"
                  @mousedown.prevent="selectSourceItem(item)"
                >
                  <div class="suggestion-main">
                    <span class="suggestion-icon">{{
                      serviceForm.type === "product" ? "📦" : "👷"
                    }}</span>
                    <span class="suggestion-sku" v-if="item.sku">{{
                      item.sku
                    }}</span>
                    <span class="suggestion-name">{{ item.name }}</span>
                  </div>
                  <div class="suggestion-meta">
                    <span class="suggestion-price">{{
                      formatCurrency(item.price)
                    }}</span>
                  </div>
                </div>
              </div>
            </div>
            <small class="form-hint">
              {{ availableSources.length }} items available from your inventory
            </small>
          </div>

          <!-- Vendor Product/Service Entry: Only for RENTED -->
          <div
            v-if="serviceForm.source === 'rented'"
            class="vendor-product-section"
          >
            <h5 class="subsection-title">
              {{
                serviceForm.type === "product"
                  ? "📦 Vendor Product Details"
                  : "👷 Vendor Service Details"
              }}
            </h5>
            <div class="form-grid">
              <div class="form-group">
                <label class="form-label">
                  {{
                    serviceForm.type === "product"
                      ? "Product SKU"
                      : "Service Code"
                  }}
                </label>
                <input
                  v-model="vendorProductForm.sku"
                  type="text"
                  class="form-input"
                  :placeholder="
                    serviceForm.type === 'product'
                      ? 'e.g., HIK-CAM-8MP-001'
                      : 'e.g., GUARD-SVC-001'
                  "
                />
                <small class="form-hint">
                  {{
                    serviceForm.type === "product"
                      ? "Vendor's product SKU"
                      : "Vendor's service code"
                  }}
                </small>
              </div>

              <div class="form-group">
                <label class="form-label">
                  {{
                    serviceForm.type === "product"
                      ? "Product Name"
                      : "Service Name"
                  }}
                  *
                </label>
                <input
                  v-model="vendorProductForm.name"
                  type="text"
                  class="form-input"
                  :placeholder="
                    serviceForm.type === 'product'
                      ? 'e.g., 8MP IP Camera'
                      : 'e.g., Security Guard Service'
                  "
                  required
                />
              </div>

              <div class="form-group full-width">
                <label class="form-label">Description</label>
                <textarea
                  v-model="vendorProductForm.description"
                  class="form-input"
                  rows="2"
                  :placeholder="
                    serviceForm.type === 'product'
                      ? 'Product specifications and details...'
                      : 'Service scope and responsibilities...'
                  "
                ></textarea>
              </div>
            </div>
          </div>

          <div class="form-group full-width">
            <label class="form-label">
              {{
                serviceForm.type === "product"
                  ? "Product Rental Name"
                  : "Service Name"
              }}
              *
            </label>
            <input
              v-model="serviceForm.name"
              type="text"
              class="form-input"
              :placeholder="
                serviceForm.type === 'product'
                  ? 'e.g., CCTV Camera - Monthly Rental'
                  : 'e.g., Security Guard - Monthly Service'
              "
              required
            />
            <small class="form-hint">
              {{
                serviceForm.type === "product"
                  ? "What you call this product rental offering"
                  : "What you call this service offering"
              }}
            </small>
          </div>

          <div class="form-group full-width">
            <label class="form-label">Description *</label>
            <textarea
              v-model="serviceForm.description"
              class="form-input"
              rows="2"
              :placeholder="
                serviceForm.type === 'product'
                  ? 'Describe the product rental terms, what\'s included, specifications...'
                  : 'Describe the service scope, hours, responsibilities...'
              "
              required
            ></textarea>
          </div>
        </div>

        <div class="form-section">
          <h4 class="section-title">Pricing</h4>

          <div class="form-grid">
            <div class="form-group">
              <label class="form-label">
                {{
                  serviceForm.type === "product"
                    ? "Monthly Rental Price (SAR)"
                    : "Monthly Service Price (SAR)"
                }}
                *
              </label>
              <input
                v-model.number="serviceForm.monthlyPrice"
                type="number"
                step="0.01"
                min="0"
                class="form-input"
                required
              />
              <small class="form-hint">
                {{
                  serviceForm.type === "product"
                    ? "What you charge customer per month"
                    : "What you charge for monthly service"
                }}
              </small>
            </div>

            <div class="form-group">
              <label class="form-label">Monthly Cost (SAR) *</label>
              <input
                v-model.number="serviceForm.monthlyCost"
                type="number"
                step="0.01"
                min="0"
                class="form-input"
                required
              />
              <small class="form-hint">
                {{
                  serviceForm.source === "owned"
                    ? "Your internal cost"
                    : "Vendor charges you"
                }}
                | Margin: {{ calculateFormMargin() }}%
              </small>
            </div>

            <div class="form-group">
              <label class="form-label"
                >Minimum Contract Period (Months) *</label
              >
              <input
                v-model.number="serviceForm.minimumMonths"
                type="number"
                min="1"
                class="form-input"
                required
              />
              <small class="form-hint">
                {{
                  serviceForm.type === "product"
                    ? "Minimum rental period"
                    : "Minimum service contract"
                }}
              </small>
            </div>

            <div class="form-group">
              <label class="form-label">
                {{
                  serviceForm.type === "product"
                    ? "Installation/Setup Fee (SAR)"
                    : "Onboarding Fee (SAR)"
                }}
              </label>
              <input
                v-model.number="serviceForm.setupFee"
                type="number"
                step="0.01"
                min="0"
                class="form-input"
              />
              <small class="form-hint">
                {{
                  serviceForm.type === "product"
                    ? "One-time installation cost"
                    : "One-time onboarding/training cost"
                }}
              </small>
            </div>
          </div>
        </div>

        <!-- Vendor Quote Section (only for rented services) -->
        <div
          v-if="serviceForm.source === 'rented'"
          class="form-section vendor-quote-section"
        >
          <h4 class="section-title">
            🏭 Vendor Rental Information
            <span class="section-subtitle"
              >Quote details from the vendor we're renting from</span
            >
          </h4>

          <div class="form-grid">
            <div class="form-group">
              <label class="form-label">Vendor Name *</label>
              <input
                v-model="vendorQuoteForm.vendorName"
                type="text"
                class="form-input"
                placeholder="e.g., Hikvision MENA"
                required
              />
            </div>

            <div class="form-group">
              <label class="form-label">Vendor Quote Number *</label>
              <input
                v-model="vendorQuoteForm.quoteNumber"
                type="text"
                class="form-input"
                placeholder="e.g., HIK-Q-2024-001"
                required
              />
            </div>

            <div class="form-group">
              <label class="form-label">Quote Date *</label>
              <input
                v-model="vendorQuoteForm.quoteDate"
                type="date"
                class="form-input"
                required
              />
            </div>

            <div class="form-group">
              <label class="form-label">Valid Until *</label>
              <input
                v-model="vendorQuoteForm.validUntil"
                type="date"
                class="form-input"
                required
              />
            </div>

            <div class="form-group">
              <label class="form-label">Vendor Monthly Rate (SAR) *</label>
              <input
                v-model.number="vendorQuoteForm.monthlyRate"
                type="number"
                step="0.01"
                min="0"
                class="form-input"
                required
                @input="serviceForm.monthlyCost = vendorQuoteForm.monthlyRate"
              />
              <small class="form-hint"
                >This will auto-fill your monthly cost</small
              >
            </div>

            <div class="form-group">
              <label class="form-label">Minimum Period (Months) *</label>
              <input
                v-model.number="vendorQuoteForm.minimumPeriodMonths"
                type="number"
                min="1"
                class="form-input"
                required
              />
            </div>

            <div class="form-group">
              <label class="form-label">Vendor Setup Cost (SAR)</label>
              <input
                v-model.number="vendorQuoteForm.setupCost"
                type="number"
                step="0.01"
                min="0"
                class="form-input"
              />
            </div>

            <div class="form-group full-width">
              <label class="form-label">Quote Notes</label>
              <textarea
                v-model="vendorQuoteForm.notes"
                class="form-input"
                rows="2"
                placeholder="Any special terms or conditions from vendor..."
              ></textarea>
            </div>
          </div>

          <!-- File Upload Section -->
          <div class="file-upload-section">
            <h5 class="subsection-title">
              📎 Attach Vendor
              {{ serviceForm.type === "product" ? "Product" : "Service" }} Quote
              (Optional)
            </h5>
            <p class="upload-hint">
              {{
                serviceForm.type === "product"
                  ? "Attach vendor product quote/catalog for reference"
                  : "Attach vendor service agreement/quote for reference"
              }}
              (PDF, Excel, images, etc.)
            </p>

            <!-- Previously Uploaded Files Library -->
            <div v-if="uploadedFiles.length > 0" class="file-library">
              <label class="form-label"
                >Or select from previously uploaded files:</label
              >
              <div class="file-library-grid">
                <div
                  v-for="file in uploadedFiles"
                  :key="file.id"
                  class="library-file-item"
                  :class="{ selected: selectedLibraryFile?.id === file.id }"
                  @click="selectLibraryFile(file)"
                >
                  <span class="file-icon-small">{{
                    getFileIcon(file.name)
                  }}</span>
                  <div class="library-file-info">
                    <p class="library-file-name">{{ file.name }}</p>
                    <p class="library-file-meta">
                      {{ formatFileSize(file.size) }} • {{ file.uploadedAt }}
                    </p>
                  </div>
                  <span
                    v-if="selectedLibraryFile?.id === file.id"
                    class="check-icon"
                    >✓</span
                  >
                </div>
              </div>
            </div>

            <!-- Upload New File -->
            <div
              class="file-upload-area"
              @drop.prevent="handleFileDrop"
              @dragover.prevent
              @dragenter.prevent
            >
              <input
                ref="fileInput"
                type="file"
                accept=".pdf,.xlsx,.xls,.csv,.txt,.doc,.docx,.jpg,.jpeg,.png"
                @change="handleFileSelect"
                class="file-input-hidden"
              />
              <div
                v-if="!uploadedFile && !selectedLibraryFile"
                class="upload-placeholder"
                @click="triggerFileInput"
              >
                <span class="upload-icon">📁</span>
                <p class="upload-text">Click to upload or drag & drop</p>
                <p class="upload-subtext">
                  PDF, Excel, Word, Images, or Text files
                </p>
              </div>
              <div v-else-if="uploadedFile" class="uploaded-file-info">
                <span class="file-icon">{{
                  getFileIcon(uploadedFile.name)
                }}</span>
                <div class="file-details">
                  <p class="file-name">{{ uploadedFile.name }}</p>
                  <p class="file-size">
                    {{ formatFileSize(uploadedFile.size) }}
                  </p>
                </div>
                <button
                  type="button"
                  class="btn-remove-file"
                  @click="removeFile"
                >
                  ✕
                </button>
              </div>
              <div v-else-if="selectedLibraryFile" class="uploaded-file-info">
                <span class="file-icon">{{
                  getFileIcon(selectedLibraryFile.name)
                }}</span>
                <div class="file-details">
                  <p class="file-name">
                    {{ selectedLibraryFile.name }}
                    <span class="library-badge">(From Library)</span>
                  </p>
                  <p class="file-size">
                    {{ formatFileSize(selectedLibraryFile.size) }}
                  </p>
                </div>
                <button
                  type="button"
                  class="btn-remove-file"
                  @click="clearLibrarySelection"
                >
                  ✕
                </button>
              </div>
            </div>
          </div>

          <!-- Current Rental Period -->
          <div class="rental-period-section">
            <h5 class="subsection-title">Current Rental Period</h5>
            <div class="form-grid">
              <div class="form-group">
                <label class="form-label">Start Date *</label>
                <input
                  v-model="rentalPeriodForm.startDate"
                  type="date"
                  class="form-input"
                  required
                />
              </div>

              <div class="form-group">
                <label class="form-label">End Date *</label>
                <input
                  v-model="rentalPeriodForm.endDate"
                  type="date"
                  class="form-input"
                  required
                />
              </div>

              <div class="form-group">
                <label class="form-label">Months Remaining</label>
                <input
                  :value="calculateRemainingMonths()"
                  type="number"
                  class="form-input"
                  readonly
                />
                <small class="form-hint">Auto-calculated from dates</small>
              </div>
            </div>
          </div>
        </div>

        <div class="form-section">
          <h4 class="section-title">Additional Information</h4>

          <div class="form-group">
            <label class="form-label">Status *</label>
            <select v-model="serviceForm.status" class="form-input" required>
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
            </select>
          </div>

          <div class="form-group full-width">
            <label class="form-label">Notes</label>
            <textarea
              v-model="serviceForm.notes"
              class="form-input"
              rows="2"
              placeholder="Additional notes about this service..."
            ></textarea>
          </div>
        </div>

        <div class="form-actions">
          <button
            type="button"
            class="btn btn-secondary"
            @click="closeServiceModal"
          >
            Cancel
          </button>
          <button type="submit" class="btn btn-primary">
            {{ editingService ? "Update Service" : "Create Service" }}
          </button>
        </div>
      </form>
    </BaseModal>

    <!-- Service Details Modal -->
    <BaseModal
      v-if="selectedService"
      v-model="showDetailsModal"
      :title="selectedService.name"
      size="lg"
    >
      <div class="service-details">
        <div class="details-section">
          <h3>Service Information</h3>
          <div class="details-grid">
            <div class="detail-item">
              <span class="detail-label">Type:</span>
              <span class="detail-value">
                <BaseBadge
                  :variant="
                    selectedService.type === 'product' ? 'info' : 'warning'
                  "
                  size="sm"
                >
                  {{
                    selectedService.type === "product"
                      ? "Product Rental"
                      : "Labor Service"
                  }}
                </BaseBadge>
              </span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Source Type:</span>
              <span class="detail-value">
                <BaseBadge
                  :variant="
                    selectedService.source === 'owned' ? 'success' : 'warning'
                  "
                  size="sm"
                >
                  {{
                    selectedService.source === "owned"
                      ? "🏢 Owned"
                      : "🏭 Rented from Vendor"
                  }}
                </BaseBadge>
              </span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Source:</span>
              <span class="detail-value">{{ selectedService.sourceName }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Source ID:</span>
              <span class="detail-value">{{ selectedService.sourceId }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Status:</span>
              <span class="detail-value">
                <BaseBadge
                  :variant="
                    selectedService.status === 'active' ? 'success' : 'default'
                  "
                  size="sm"
                >
                  {{
                    selectedService.status === "active" ? "Active" : "Inactive"
                  }}
                </BaseBadge>
              </span>
            </div>
            <div class="detail-item full-width">
              <span class="detail-label">Description:</span>
              <span class="detail-value">{{
                selectedService.description
              }}</span>
            </div>
          </div>
        </div>

        <div class="details-section">
          <h3>Pricing Details</h3>
          <div class="details-grid">
            <div class="detail-item">
              <span class="detail-label">Monthly Price:</span>
              <span class="detail-value price-value"
                >{{ formatCurrency(selectedService.monthlyPrice) }}/month</span
              >
            </div>
            <div class="detail-item">
              <span class="detail-label">Monthly Cost:</span>
              <span class="detail-value">{{
                formatCurrency(selectedService.monthlyCost)
              }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Margin:</span>
              <span
                class="detail-value"
                :class="getMarginClass(calculateMargin(selectedService))"
              >
                {{ calculateMargin(selectedService).toFixed(1) }}%
              </span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Minimum Period:</span>
              <span class="detail-value"
                >{{ selectedService.minimumMonths }} months</span
              >
            </div>
            <div v-if="selectedService.setupFee" class="detail-item">
              <span class="detail-label">Setup Fee:</span>
              <span class="detail-value">{{
                formatCurrency(selectedService.setupFee)
              }}</span>
            </div>
          </div>
        </div>

        <!-- Vendor Quote Information (for rented services) -->
        <div
          v-if="
            selectedService.source === 'rented' && selectedService.vendorQuote
          "
          class="details-section vendor-quote-details"
        >
          <h3>🏭 Vendor Rental Information</h3>
          <div class="details-grid">
            <div class="detail-item">
              <span class="detail-label">Vendor:</span>
              <span class="detail-value">{{
                selectedService.vendorQuote.vendorName
              }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Quote Number:</span>
              <span class="detail-value">{{
                selectedService.vendorQuote.quoteNumber
              }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Quote Date:</span>
              <span class="detail-value">{{
                selectedService.vendorQuote.quoteDate
              }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Valid Until:</span>
              <span class="detail-value">{{
                selectedService.vendorQuote.validUntil
              }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Vendor Monthly Rate:</span>
              <span class="detail-value cost-value"
                >{{
                  formatCurrency(selectedService.vendorQuote.monthlyRate)
                }}/month</span
              >
            </div>
            <div class="detail-item">
              <span class="detail-label">Minimum Period:</span>
              <span class="detail-value"
                >{{
                  selectedService.vendorQuote.minimumPeriodMonths
                }}
                months</span
              >
            </div>
            <div
              v-if="selectedService.vendorQuote.setupCost"
              class="detail-item"
            >
              <span class="detail-label">Vendor Setup Cost:</span>
              <span class="detail-value">{{
                formatCurrency(selectedService.vendorQuote.setupCost)
              }}</span>
            </div>
            <div
              v-if="selectedService.vendorQuote.notes"
              class="detail-item full-width"
            >
              <span class="detail-label">Quote Notes:</span>
              <span class="detail-value">{{
                selectedService.vendorQuote.notes
              }}</span>
            </div>
          </div>

          <!-- Current Rental Period -->
          <div
            v-if="selectedService.currentRentalPeriod"
            class="rental-period-info"
          >
            <h4>Current Rental Period</h4>
            <div class="details-grid">
              <div class="detail-item">
                <span class="detail-label">Start Date:</span>
                <span class="detail-value">{{
                  selectedService.currentRentalPeriod.startDate
                }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">End Date:</span>
                <span class="detail-value">{{
                  selectedService.currentRentalPeriod.endDate
                }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">Months Remaining:</span>
                <span
                  class="detail-value"
                  :class="
                    selectedService.currentRentalPeriod.monthsRemaining <= 3
                      ? 'warning-text'
                      : ''
                  "
                >
                  {{
                    selectedService.currentRentalPeriod.monthsRemaining
                  }}
                  months
                  <span
                    v-if="
                      selectedService.currentRentalPeriod.monthsRemaining <= 3
                    "
                    class="warning-badge"
                    >⚠️ Expiring Soon</span
                  >
                </span>
              </div>
            </div>
          </div>
        </div>

        <div v-if="selectedService.notes" class="details-section">
          <h3>Notes</h3>
          <div class="notes-content">
            {{ selectedService.notes }}
          </div>
        </div>

        <div class="details-section">
          <h3>Cost Examples</h3>
          <table class="cost-examples-table">
            <thead>
              <tr>
                <th>Period</th>
                <th>Monthly Cost</th>
                <th>Setup Fee</th>
                <th>Total Cost</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="months in [selectedService.minimumMonths, 12, 24, 36]"
                :key="months"
              >
                <td>{{ months }} months</td>
                <td>
                  {{ formatCurrency(selectedService.monthlyPrice * months) }}
                </td>
                <td>{{ formatCurrency(selectedService.setupFee || 0) }}</td>
                <td class="total-cost">
                  {{
                    formatCurrency(
                      selectedService.monthlyPrice * months +
                        (selectedService.setupFee || 0),
                    )
                  }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <template #footer>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="closeDetailsModal">
            Close
          </button>
          <button class="btn btn-primary" @click="editService(selectedService)">
            Edit Service
          </button>
        </div>
      </template>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useRecurringServicesStore } from "@/stores/recurringServices";
import { useProductsStore } from "@/stores/products";
import { useLaborPositionsStore } from "@/stores/laborPositions";
import { useToast } from "@/composables/useToast";
import BaseCard from "@/components/UI/BaseCard.vue";
import BaseTable from "@/components/UI/BaseTable.vue";
import BaseBadge from "@/components/UI/BaseBadge.vue";
import BaseModal from "@/components/UI/BaseModal.vue";
import type { TableColumn } from "@/components/UI/BaseTable.vue";
import type {
  RecurringService,
  RecurringServiceType,
  RecurringServiceSource,
} from "@/stores/recurringServices";

const recurringStore = useRecurringServicesStore();
const productsStore = useProductsStore();
const laborStore = useLaborPositionsStore();
const { success, info } = useToast();

const searchQuery = ref("");
const typeFilter = ref<"all" | RecurringServiceType>("all");
const statusFilter = ref<"all" | "active" | "inactive">("all");
const showServiceModal = ref(false);
const showDetailsModal = ref(false);
const editingService = ref<RecurringService | null>(null);
const selectedService = ref<RecurringService | null>(null);

const serviceForm = ref({
  type: "" as RecurringServiceType | "",
  source: "" as RecurringServiceSource | "",
  sourceId: "",
  sourceName: "",
  name: "",
  description: "",
  monthlyPrice: 0,
  monthlyCost: 0,
  minimumMonths: 1,
  setupFee: 0,
  status: "active" as "active" | "inactive",
  notes: "",
});

const vendorQuoteForm = ref({
  vendorId: "",
  vendorName: "",
  quoteNumber: "",
  quoteDate: "",
  validUntil: "",
  monthlyRate: 0,
  minimumPeriodMonths: 1,
  setupCost: 0,
  notes: "",
});

const rentalPeriodForm = ref({
  startDate: "",
  endDate: "",
  monthsRemaining: 0,
});

const vendorProductForm = ref({
  sku: "",
  name: "",
  description: "",
  monthlyRate: 0,
  setupCost: 0,
});

// Autocomplete for owned products
const sourceSearchQuery = ref("");
const showSourceSuggestions = ref(false);

// File upload state - simplified to attachment only
const fileInput = ref<HTMLInputElement | null>(null);
const uploadedFile = ref<File | null>(null);

// File library for reusing uploaded files
interface UploadedFileInfo {
  id: string;
  name: string;
  size: number;
  uploadedAt: string;
  url: string;
}
const uploadedFiles = ref<UploadedFileInfo[]>([]);
const selectedLibraryFile = ref<UploadedFileInfo | null>(null);

// Table columns
const servicesColumns: TableColumn[] = [
  { key: "name", label: "Service Name", sortable: true },
  { key: "type", label: "Type", sortable: true },
  { key: "source", label: "Source", sortable: false },
  { key: "pricing", label: "Pricing", sortable: false },
  { key: "margin", label: "Margin", sortable: false },
  { key: "status", label: "Status", sortable: true },
  { key: "actions", label: "Actions", align: "right" },
];

// Computed
const filteredServices = computed(() => {
  let filtered = recurringStore.recurringServices;

  if (typeFilter.value !== "all") {
    filtered = filtered.filter((s) => s.type === typeFilter.value);
  }

  if (statusFilter.value !== "all") {
    filtered = filtered.filter((s) => s.status === statusFilter.value);
  }

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(
      (s) =>
        s.name.toLowerCase().includes(query) ||
        s.description.toLowerCase().includes(query) ||
        s.sourceName.toLowerCase().includes(query),
    );
  }

  return filtered;
});

const activeServicesCount = computed(
  () =>
    recurringStore.recurringServices.filter((s) => s.status === "active")
      .length,
);

const availableSources = computed(() => {
  if (serviceForm.value.type === "product") {
    return productsStore.getActiveProducts();
  } else if (serviceForm.value.type === "labor") {
    return laborStore.getActivePositions();
  }
  return [];
});

// Helper functions
function formatCurrency(amount: number) {
  return new Intl.NumberFormat("en-SA", {
    style: "currency",
    currency: "SAR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount);
}

function calculateMargin(service: RecurringService): number {
  if (service.monthlyPrice === 0) return 0;
  return (
    ((service.monthlyPrice - service.monthlyCost) / service.monthlyPrice) * 100
  );
}

function calculateFormMargin(): string {
  if (serviceForm.value.monthlyPrice === 0) return "0.00";
  const margin =
    ((serviceForm.value.monthlyPrice - serviceForm.value.monthlyCost) /
      serviceForm.value.monthlyPrice) *
    100;
  return margin.toFixed(2);
}

function getMarginClass(margin: number): string {
  if (margin >= 30) return "margin-good";
  if (margin >= 20) return "margin-ok";
  return "margin-low";
}

function onSourceTypeChange() {
  // Clear vendor quote when switching between owned/rented
  if (serviceForm.value.source === "owned") {
    vendorQuoteForm.value = {
      vendorId: "",
      vendorName: "",
      quoteNumber: "",
      quoteDate: "",
      validUntil: "",
      monthlyRate: 0,
      minimumPeriodMonths: 1,
      setupCost: 0,
      notes: "",
    };
    rentalPeriodForm.value = {
      startDate: "",
      endDate: "",
      monthsRemaining: 0,
    };
  }
}

function calculateRemainingMonths(): number {
  if (!rentalPeriodForm.value.startDate || !rentalPeriodForm.value.endDate)
    return 0;

  const now = new Date();
  const endDate = new Date(rentalPeriodForm.value.endDate);
  const monthsRemaining = Math.max(
    0,
    Math.ceil(
      (endDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24 * 30.44),
    ),
  );

  return monthsRemaining;
}

// Autocomplete for owned products
const filteredSourceSuggestions = computed(() => {
  if (
    !sourceSearchQuery.value ||
    sourceSearchQuery.value.length < 2 ||
    !serviceForm.value.type
  ) {
    return [];
  }

  const query = sourceSearchQuery.value.toLowerCase();
  const sources = availableSources.value;

  const filtered = sources.filter(
    (item) =>
      item.name.toLowerCase().includes(query) ||
      (item.sku && item.sku.toLowerCase().includes(query)),
  );

  return filtered.slice(0, 10);
});

function handleSourceSearchInput() {
  if (sourceSearchQuery.value.length >= 2) {
    showSourceSuggestions.value = true;
  } else {
    showSourceSuggestions.value = false;
  }
}

function handleSourceSearchBlur() {
  setTimeout(() => {
    showSourceSuggestions.value = false;
  }, 200);
}

function selectSourceItem(item: any) {
  sourceSearchQuery.value = `${item.sku ? item.sku + " - " : ""}${item.name}`;
  serviceForm.value.sourceId = item.id;
  serviceForm.value.sourceName = item.name;
  showSourceSuggestions.value = false;
  onSourceSelect();
}

// File handling - simple attachment only (no parsing)
function triggerFileInput() {
  fileInput.value?.click();
}

function handleFileSelect(event: Event) {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  if (file) {
    uploadedFile.value = file;
    selectedLibraryFile.value = null; // Clear library selection when uploading new
    success("File attached successfully");
  }
}

function handleFileDrop(event: DragEvent) {
  const file = event.dataTransfer?.files[0];
  if (file) {
    uploadedFile.value = file;
    selectedLibraryFile.value = null; // Clear library selection when uploading new
    success("File attached successfully");
  }
}

function removeFile() {
  uploadedFile.value = null;
  selectedLibraryFile.value = null;
  if (fileInput.value) {
    fileInput.value.value = "";
  }
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + " B";
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
  return (bytes / (1024 * 1024)).toFixed(1) + " MB";
}

// File library functions
function getFileIcon(fileName: string): string {
  const ext = fileName.split(".").pop()?.toLowerCase();
  if (ext === "pdf") return "📄";
  if (ext === "xlsx" || ext === "xls") return "📊";
  if (ext === "doc" || ext === "docx") return "📝";
  if (ext === "jpg" || ext === "jpeg" || ext === "png") return "🖼️";
  return "📎";
}

function selectLibraryFile(file: UploadedFileInfo) {
  if (selectedLibraryFile.value?.id === file.id) {
    selectedLibraryFile.value = null; // Deselect if clicking same file
  } else {
    selectedLibraryFile.value = file;
    uploadedFile.value = null; // Clear uploaded file when selecting from library
  }
}

function clearLibrarySelection() {
  selectedLibraryFile.value = null;
}

// Actions
function createNewService() {
  resetServiceForm();
  showServiceModal.value = true;
}

function resetServiceForm() {
  serviceForm.value = {
    type: "",
    source: "",
    sourceId: "",
    sourceName: "",
    name: "",
    description: "",
    monthlyPrice: 0,
    monthlyCost: 0,
    minimumMonths: 1,
    setupFee: 0,
    status: "active",
    notes: "",
  };
  vendorQuoteForm.value = {
    vendorId: "",
    vendorName: "",
    quoteNumber: "",
    quoteDate: "",
    validUntil: "",
    monthlyRate: 0,
    minimumPeriodMonths: 1,
    setupCost: 0,
    notes: "",
  };
  rentalPeriodForm.value = {
    startDate: "",
    endDate: "",
    monthsRemaining: 0,
  };
  vendorProductForm.value = {
    sku: "",
    name: "",
    description: "",
    monthlyRate: 0,
    setupCost: 0,
  };
  sourceSearchQuery.value = "";
  uploadedFile.value = null;
  selectedLibraryFile.value = null;
  editingService.value = null;
}

function onTypeChange() {
  serviceForm.value.sourceId = "";
  serviceForm.value.sourceName = "";
}

function onSourceSelect() {
  if (!serviceForm.value.sourceId) return;

  if (serviceForm.value.type === "product") {
    const product = productsStore.getProductById(serviceForm.value.sourceId);
    if (product) {
      serviceForm.value.sourceName = product.name;
      // Pre-fill with suggested values
      if (!editingService.value) {
        serviceForm.value.name = `${product.name} - Monthly Rental`;
        serviceForm.value.description = `Monthly rental service for ${product.name}`;
        serviceForm.value.monthlyCost = Math.round(product.landedCostSAR * 0.1); // 10% of product cost per month
        serviceForm.value.monthlyPrice = Math.round(
          product.sellingPrice * 0.15,
        ); // 15% of selling price per month
        serviceForm.value.minimumMonths = 12;
        serviceForm.value.setupFee = Math.round(product.sellingPrice * 0.5); // 50% as setup
      }
    }
  } else if (serviceForm.value.type === "labor") {
    const labor = laborStore.getPositionById(serviceForm.value.sourceId);
    if (labor) {
      serviceForm.value.sourceName = labor.name;
      // Pre-fill with suggested values
      if (!editingService.value) {
        serviceForm.value.name = `${labor.name} - Monthly Service`;
        serviceForm.value.description = `Monthly service for ${labor.name} (160 hours/month)`;
        serviceForm.value.monthlyCost = labor.costPerHour * 160; // 160 hours per month
        serviceForm.value.monthlyPrice = labor.sellingRatePerHour * 160;
        serviceForm.value.minimumMonths = 3;
        serviceForm.value.setupFee = 500;
      }
    }
  }
}

function saveService() {
  // Validation: owned services need sourceId, rented services need vendor product
  if (serviceForm.value.source === "owned") {
    if (!serviceForm.value.type || !serviceForm.value.sourceId) {
      info("Please select a product or labor position from your inventory");
      return;
    }
  } else if (serviceForm.value.source === "rented") {
    if (!vendorProductForm.value.name) {
      info("Please enter vendor product name");
      return;
    }
    // Validate vendor quote for rented services
    if (
      !vendorQuoteForm.value.vendorName ||
      !vendorQuoteForm.value.quoteNumber ||
      !vendorQuoteForm.value.monthlyRate ||
      !rentalPeriodForm.value.startDate ||
      !rentalPeriodForm.value.endDate
    ) {
      info(
        "Please fill in all vendor quote and rental period fields for rented services",
      );
      return;
    }
  } else {
    info("Please select source type (Owned or Rented)");
    return;
  }

  // Prepare vendor product data for rented services
  const vendorProduct =
    serviceForm.value.source === "rented"
      ? {
          sku: vendorProductForm.value.sku,
          name: vendorProductForm.value.name,
          description: vendorProductForm.value.description || undefined,
          monthlyRate:
            vendorProductForm.value.monthlyRate ||
            vendorQuoteForm.value.monthlyRate,
          setupCost: vendorProductForm.value.setupCost || undefined,
        }
      : undefined;

  // Prepare vendor quote data for rented services
  const vendorQuote =
    serviceForm.value.source === "rented"
      ? {
          id: `VQ-${Date.now()}`,
          vendorId: vendorQuoteForm.value.vendorId || `VEN-${Date.now()}`,
          vendorName: vendorQuoteForm.value.vendorName,
          quoteNumber: vendorQuoteForm.value.quoteNumber,
          quoteDate: vendorQuoteForm.value.quoteDate,
          validUntil: vendorQuoteForm.value.validUntil,
          monthlyRate: vendorQuoteForm.value.monthlyRate,
          minimumPeriodMonths: vendorQuoteForm.value.minimumPeriodMonths,
          setupCost: vendorQuoteForm.value.setupCost || undefined,
          notes: vendorQuoteForm.value.notes || undefined,
          attachmentFileName:
            uploadedFile.value?.name || selectedLibraryFile.value?.name,
          attachmentUrl: selectedLibraryFile.value?.url || undefined,
        }
      : undefined;

  // Add uploaded file to library if it's a new file
  if (
    uploadedFile.value &&
    !uploadedFiles.value.find((f) => f.name === uploadedFile.value!.name)
  ) {
    uploadedFiles.value.push({
      id: `FILE-${Date.now()}`,
      name: uploadedFile.value.name,
      size: uploadedFile.value.size,
      uploadedAt: new Date().toLocaleDateString(),
      url: `/uploads/vendor-quotes/${uploadedFile.value.name}`, // Simulated URL
    });
  }

  const currentRentalPeriod =
    serviceForm.value.source === "rented"
      ? {
          startDate: rentalPeriodForm.value.startDate,
          endDate: rentalPeriodForm.value.endDate,
          monthsRemaining: calculateRemainingMonths(),
        }
      : undefined;

  // Use vendor product name for sourceName if rented
  const sourceName =
    serviceForm.value.source === "rented"
      ? vendorProductForm.value.name
      : serviceForm.value.sourceName;

  if (editingService.value) {
    recurringStore.updateRecurringService(editingService.value.id, {
      type: serviceForm.value.type as RecurringServiceType,
      source: serviceForm.value.source as RecurringServiceSource,
      sourceId: serviceForm.value.sourceId,
      sourceName,
      vendorProduct,
      name: serviceForm.value.name,
      description: serviceForm.value.description,
      monthlyPrice: serviceForm.value.monthlyPrice,
      monthlyCost: serviceForm.value.monthlyCost,
      minimumMonths: serviceForm.value.minimumMonths,
      setupFee: serviceForm.value.setupFee || undefined,
      vendorQuote,
      currentRentalPeriod,
      status: serviceForm.value.status,
      notes: serviceForm.value.notes || undefined,
    });
    success("Recurring service updated successfully");
  } else {
    recurringStore.addRecurringService({
      type: serviceForm.value.type as RecurringServiceType,
      source: serviceForm.value.source as RecurringServiceSource,
      sourceId: serviceForm.value.sourceId,
      sourceName,
      vendorProduct,
      name: serviceForm.value.name,
      description: serviceForm.value.description,
      monthlyPrice: serviceForm.value.monthlyPrice,
      monthlyCost: serviceForm.value.monthlyCost,
      minimumMonths: serviceForm.value.minimumMonths,
      setupFee: serviceForm.value.setupFee || undefined,
      vendorQuote,
      currentRentalPeriod,
      status: serviceForm.value.status,
      notes: serviceForm.value.notes || undefined,
    });
    success("Recurring service created successfully");
  }

  closeServiceModal();
}

function editService(service: RecurringService) {
  editingService.value = service;
  serviceForm.value = {
    type: service.type,
    source: service.source,
    sourceId: service.sourceId,
    sourceName: service.sourceName,
    name: service.name,
    description: service.description,
    monthlyPrice: service.monthlyPrice,
    monthlyCost: service.monthlyCost,
    minimumMonths: service.minimumMonths,
    setupFee: service.setupFee || 0,
    status: service.status,
    notes: service.notes || "",
  };

  // Load vendor quote if it exists
  if (service.source === "rented" && service.vendorQuote) {
    vendorQuoteForm.value = {
      vendorId: service.vendorQuote.vendorId,
      vendorName: service.vendorQuote.vendorName,
      quoteNumber: service.vendorQuote.quoteNumber,
      quoteDate: service.vendorQuote.quoteDate,
      validUntil: service.vendorQuote.validUntil,
      monthlyRate: service.vendorQuote.monthlyRate,
      minimumPeriodMonths: service.vendorQuote.minimumPeriodMonths,
      setupCost: service.vendorQuote.setupCost || 0,
      notes: service.vendorQuote.notes || "",
    };

    // Load rental period if it exists
    if (service.currentRentalPeriod) {
      rentalPeriodForm.value = {
        startDate: service.currentRentalPeriod.startDate,
        endDate: service.currentRentalPeriod.endDate,
        monthsRemaining: service.currentRentalPeriod.monthsRemaining,
      };
    }
  }

  closeDetailsModal();
  showServiceModal.value = true;
}

function duplicateService(service: RecurringService) {
  recurringStore.addRecurringService({
    ...service,
    name: `${service.name} (Copy)`,
    status: "inactive",
  });
  success(`Service duplicated: ${service.name} (Copy)`);
}

function toggleStatus(service: RecurringService) {
  recurringStore.toggleRecurringServiceStatus(service.id);
  success(
    `Service ${service.status === "active" ? "deactivated" : "activated"}`,
  );
}

function viewServiceDetails(service: RecurringService) {
  selectedService.value = service;
  showDetailsModal.value = true;
}

function closeServiceModal() {
  showServiceModal.value = false;
  resetServiceForm();
}

function closeDetailsModal() {
  showDetailsModal.value = false;
  selectedService.value = null;
}
</script>

<style scoped>
.recurring-services-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 100%;
}

.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 1.25rem;
  width: 100%;
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
  font-size: 1.05rem;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.625rem;
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

.filters-bar {
  display: flex;
  gap: 2.5rem;
  align-items: center;
  flex-wrap: wrap;
  padding: 1.75rem 2rem;
  background: var(--color-surface);
  border-radius: 12px;
  box-shadow: var(--shadow-sm);
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.filter-group label {
  font-size: 1rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.filter-select {
  padding: 0.75rem 1.5rem;
  border: 1px solid var(--color-border);
  background: var(--color-background);
  color: var(--color-text-primary);
  border-radius: 8px;
  font-size: 1rem;
  cursor: pointer;
  min-width: 200px;
}

.filter-select:focus {
  outline: none;
  border-color: var(--color-primary);
}

.stats-summary {
  display: flex;
  gap: 2.5rem;
  margin-left: auto;
  padding-left: 2.5rem;
  border-left: 1px solid var(--color-border);
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.stat-label {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-primary);
}

.text-success {
  color: var(--color-success) !important;
}

.name-cell {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  padding: 1rem 0.75rem;
  min-width: 300px;
}

.name {
  font-weight: 600;
  font-size: 1.1rem;
  color: var(--color-text-primary);
}

.description {
  font-size: 0.95rem;
  color: var(--color-text-secondary);
  line-height: 1.4;
}

.source-cell {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.source-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.source-name {
  font-weight: 500;
  color: var(--color-text-primary);
}

.source-badge {
  flex-shrink: 0;
}

.source-id {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  font-family: monospace;
}

.vendor-info {
  font-size: 0.85rem;
  color: #f59e0b;
  font-weight: 500;
}

.pricing-cell {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.monthly-price {
  font-weight: 600;
  font-size: 1.1rem;
  color: var(--color-primary);
}

.min-period,
.setup-fee {
  font-size: 0.9rem;
  color: var(--color-text-secondary);
}

.margin-cell {
  font-weight: 600;
  font-size: 1.1rem;
}

.margin-good {
  color: var(--color-success, #4caf50);
}

.margin-ok {
  color: var(--color-warning, #ff9800);
}

.margin-low {
  color: var(--color-danger, #f44336);
}

.action-buttons {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  padding: 0.5rem;
}

.action-btn {
  background: none;
  border: none;
  font-size: 1.3rem;
  cursor: pointer;
  padding: 0.35rem 0.65rem;
  transition: transform 0.2s;
  opacity: 0.7;
}

.action-btn:hover {
  transform: scale(1.2);
  opacity: 1;
}

/* Modal Styles */
.service-form {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.section-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--color-primary);
  margin-bottom: 0.5rem;
  padding-bottom: 0.5rem;
  border-bottom: 2px solid var(--color-border);
}

.form-grid {
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
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-secondary);
}

.form-input {
  padding: 0.625rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-background);
  color: var(--color-text-primary);
  font-size: 0.95rem;
}

.form-input:focus {
  outline: none;
  border-color: var(--color-primary);
}

.form-hint {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  font-style: italic;
}

.form-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  padding-top: 1rem;
  border-top: 1px solid var(--color-border);
}

/* Service Details */
.service-details {
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
}

.details-section h3 {
  margin: 0 0 1rem 0;
  font-size: 1.2rem;
  color: var(--color-text-primary);
  font-weight: 600;
}

.details-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.detail-item.full-width {
  grid-column: 1 / -1;
}

.detail-label {
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.detail-value {
  font-size: 1.05rem;
  color: var(--color-text-primary);
}

.price-value {
  font-weight: 600;
  font-size: 1.2rem;
  color: var(--color-primary);
}

.notes-content {
  padding: 1rem;
  background: var(--color-background);
  border-radius: 8px;
  color: var(--color-text-primary);
  line-height: 1.6;
}

.cost-examples-table {
  width: 100%;
  border-collapse: collapse;
}

.cost-examples-table th,
.cost-examples-table td {
  padding: 0.75rem;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
}

.cost-examples-table th {
  background: var(--color-background);
  font-weight: 600;
  color: var(--color-text-secondary);
  font-size: 0.9rem;
  text-transform: uppercase;
}

.cost-examples-table td {
  color: var(--color-text-primary);
}

.total-cost {
  font-weight: 700;
  color: var(--color-primary);
  font-size: 1.1rem;
}

.modal-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
}

/* Vendor Quote Section Styles */
.vendor-quote-section {
  background: linear-gradient(135deg, #fef3c7 0%, #fed7aa 100%);
  padding: 1.5rem;
  border-radius: 12px;
  border: 2px solid #f59e0b;
}

.vendor-quote-section .section-title {
  color: #92400e;
  margin-bottom: 0.5rem;
}

.section-subtitle {
  display: block;
  font-size: 0.85rem;
  font-weight: 400;
  color: #78350f;
  margin-top: 0.25rem;
}

.vendor-quote-details {
  background: #fffbeb;
  border: 1px solid #fde68a;
  padding: 1.5rem;
  border-radius: 8px;
}

.vendor-quote-details h3 {
  color: #92400e;
}

.rental-period-section {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 2px dashed #f59e0b;
}

.subsection-title {
  font-size: 1rem;
  font-weight: 600;
  color: #92400e;
  margin-bottom: 1rem;
}

.rental-period-info {
  margin-top: 1rem;
  padding: 1rem;
  background: #fef3c7;
  border-radius: 8px;
}

.rental-period-info h4 {
  font-size: 0.95rem;
  font-weight: 600;
  color: #92400e;
  margin-bottom: 0.75rem;
}

.cost-value {
  color: #ef4444;
  font-weight: 700;
}

.warning-text {
  color: #dc2626;
  font-weight: 600;
}

.warning-badge {
  display: inline-block;
  margin-left: 0.5rem;
  padding: 0.2rem 0.5rem;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 4px;
  font-size: 0.75rem;
  color: #dc2626;
}

/* Autocomplete Styles */
.autocomplete-container {
  position: relative;
}

.autocomplete-wrapper {
  position: relative;
}

.autocomplete-dropdown {
  position: absolute;
  top: calc(100% + 0.5rem);
  left: 0;
  right: 0;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  box-shadow: var(--shadow-lg);
  max-height: 400px;
  overflow-y: auto;
  z-index: 1000;
}

.autocomplete-header {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--color-border);
  background: var(--color-background);
}

.suggestion-count {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.autocomplete-item {
  padding: 0.875rem 1rem;
  cursor: pointer;
  border-bottom: 1px solid var(--color-border);
  transition: all 0.15s;
}

.autocomplete-item:hover {
  background: var(--color-background);
}

.autocomplete-item:last-child {
  border-bottom: none;
}

.suggestion-main {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.25rem;
}

.suggestion-icon {
  font-size: 1.25rem;
}

.suggestion-sku {
  font-family: monospace;
  font-size: 0.9rem;
  color: var(--color-text-secondary);
  background: var(--color-background);
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
}

.suggestion-name {
  font-weight: 500;
  color: var(--color-text-primary);
}

.suggestion-meta {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-left: 2rem;
  font-size: 0.85rem;
}

.suggestion-price {
  color: var(--color-primary);
  font-weight: 600;
}

/* File Upload Styles */
.file-upload-section {
  margin-top: 1.5rem;
  padding: 1.5rem;
  background: #f0f9ff;
  border: 2px dashed #3b82f6;
  border-radius: 12px;
}

.upload-hint {
  font-size: 0.9rem;
  color: #1e40af;
  margin-bottom: 1rem;
}

.file-upload-area {
  border: 2px dashed #93c5fd;
  border-radius: 8px;
  padding: 2rem;
  text-align: center;
  background: white;
  transition: all 0.2s;
  cursor: pointer;
}

.file-upload-area:hover {
  border-color: #3b82f6;
  background: #f8fafc;
}

.file-input-hidden {
  display: none;
}

.upload-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
}

.upload-icon {
  font-size: 3rem;
}

.upload-text {
  font-size: 1.1rem;
  font-weight: 600;
  color: #1e40af;
  margin: 0;
}

.upload-subtext {
  font-size: 0.85rem;
  color: #64748b;
  margin: 0;
}

.uploaded-file-info {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  background: #f1f5f9;
  border-radius: 8px;
}

.file-icon {
  font-size: 2rem;
}

.file-details {
  flex: 1;
  text-align: left;
}

.file-name {
  font-weight: 600;
  color: #1e293b;
  margin: 0 0 0.25rem 0;
}

.file-size {
  font-size: 0.85rem;
  color: #64748b;
  margin: 0;
}

.btn-remove-file {
  background: #ef4444;
  color: white;
  border: none;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  cursor: pointer;
  font-size: 1.2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.btn-remove-file:hover {
  background: #dc2626;
  transform: scale(1.1);
}

.library-badge {
  display: inline-block;
  padding: 0.15rem 0.5rem;
  background: #dbeafe;
  border: 1px solid #3b82f6;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  color: #1e40af;
  margin-left: 0.5rem;
}

/* File Library Styles */
.file-library {
  margin-bottom: 1.5rem;
  padding: 1rem;
  background: white;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.file-library-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 0.75rem;
  margin-top: 0.75rem;
}

.library-file-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  background: #f8fafc;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.library-file-item:hover {
  border-color: #3b82f6;
  background: #eff6ff;
  transform: translateY(-2px);
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.library-file-item.selected {
  border-color: #3b82f6;
  background: #dbeafe;
  box-shadow: 0 4px 6px -1px rgba(59, 130, 246, 0.2);
}

.file-icon-small {
  font-size: 1.5rem;
  flex-shrink: 0;
}

.library-file-info {
  flex: 1;
  min-width: 0;
}

.library-file-name {
  font-weight: 600;
  color: #1e293b;
  margin: 0 0 0.25rem 0;
  font-size: 0.9rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.library-file-meta {
  font-size: 0.75rem;
  color: #64748b;
  margin: 0;
}

.check-icon {
  color: #3b82f6;
  font-size: 1.25rem;
  font-weight: bold;
  flex-shrink: 0;
}

/* Vendor Product Section */
.vendor-product-section {
  padding: 1.5rem;
  background: #fef3c7;
  border: 1px solid #fde68a;
  border-radius: 8px;
  margin-bottom: 1rem;
}

.vendor-product-section .subsection-title {
  color: #92400e;
  margin-bottom: 1rem;
}
</style>
