<template>
  <div class="opportunities-view">
    <!-- Header with filters and actions -->
    <div class="view-header">
      <div class="header-actions">
        <input
          v-model="searchQuery"
          type="search"
          placeholder="Search opportunities..."
          class="search-input"
        />
        <select v-model="filterStage" class="filter-select">
          <option value="">All Stages</option>
          <option value="qualification">Qualification</option>
          <option value="proposal">Proposal</option>
          <option value="negotiation">Negotiation</option>
          <option value="closed_won">Closed Won</option>
          <option value="closed_lost">Closed Lost</option>
        </select>
        <select v-model="filterPriority" class="filter-select">
          <option value="">All Priorities</option>
          <option value="low">Low</option>
          <option value="medium">Medium</option>
          <option value="high">High</option>
          <option value="critical">Critical</option>
        </select>
        <button class="btn btn-primary" @click="createOpportunity">
          <span>➕</span> New Opportunity
        </button>
      </div>
    </div>

    <!-- View toggle -->
    <div class="view-toggle">
      <button
        class="toggle-btn"
        :class="{ active: viewMode === 'pipeline' }"
        @click="viewMode = 'pipeline'"
      >
        📊 Pipeline View
      </button>
      <button
        class="toggle-btn"
        :class="{ active: viewMode === 'list' }"
        @click="viewMode = 'list'"
      >
        📋 List View
      </button>
    </div>

    <!-- Pipeline View -->
    <div v-if="viewMode === 'pipeline'" class="pipeline-view">
      <div
        v-for="stage in pipelineStages"
        :key="stage.key"
        class="pipeline-column"
      >
        <div class="column-header">
          <h3 class="column-title">{{ stage.label }}</h3>
          <span class="column-count">{{ getOpportunitiesByStage(stage.key).length }}</span>
          <span class="column-value">
            {{ formatCurrency(getTotalValueByStage(stage.key)) }}
          </span>
        </div>

        <div class="column-cards">
          <div
            v-for="opp in getOpportunitiesByStage(stage.key)"
            :key="opp.id"
            class="opportunity-card"
            :class="`priority-${opp.priority}`"
            @click="viewOpportunity(opp)"
          >
            <div class="card-header">
              <BaseBadge :variant="getPriorityVariant(opp.priority)" size="sm">
                {{ opp.priority }}
              </BaseBadge>
              <span class="win-probability">{{ opp.winProbability }}%</span>
            </div>

            <h4 class="card-title">{{ opp.name }}</h4>

            <div class="card-meta">
              <div class="meta-row">
                <span class="meta-label">Company:</span>
                <span class="meta-value">{{ getCompanyName(opp.companyId) }}</span>
              </div>
              <div class="meta-row">
                <span class="meta-label">Value:</span>
                <span class="meta-value value-highlight">
                  {{ formatCurrency(opp.estimatedValue) }}
                </span>
              </div>
              <div class="meta-row">
                <span class="meta-label">Margin:</span>
                <span
                  class="meta-value"
                  :class="{ 'margin-warning': opp.targetMargin < 20 }"
                >
                  {{ opp.targetMargin }}%
                </span>
              </div>
              <div class="meta-row" v-if="opp.teamId">
                <span class="meta-label">Lead Source:</span>
                <span class="meta-value">{{ getTeamName(opp.teamId) }}</span>
              </div>
            </div>

            <div class="card-footer">
              <div class="footer-team-info">
                <span class="team-label" v-if="opp.quoteSalesPersonId && opp.quotePresalesPersonId">
                  📊 Quote Team: {{ getTeamMemberName(opp.quoteSalesPersonId) }} & {{ getTeamMemberName(opp.quotePresalesPersonId) }}
                </span>
                <span class="team-label warning" v-else-if="opp.stage !== 'closed_won' && opp.stage !== 'closed_lost'">
                  ⚠️ Quote team not assigned
                </span>
                <span class="team-label" v-else-if="opp.salesExecutiveId">
                  👤 {{ getUserName(opp.salesExecutiveId) }}
                </span>
              </div>
            </div>
          </div>

          <div v-if="getOpportunitiesByStage(stage.key).length === 0" class="empty-column">
            No opportunities in this stage
          </div>
        </div>
      </div>
    </div>

    <!-- List View -->
    <BaseCard v-if="viewMode === 'list'" no-padding>
      <BaseTable
        :columns="tableColumns"
        :data="filteredOpportunities"
        :clickable="true"
        :pagination="true"
        :page-size="10"
        @row-click="viewOpportunity"
      >
        <template #cell-name="{ row }">
          <div class="opp-name">
            <div class="name-text">{{ row.name }}</div>
            <div class="name-subtitle">{{ row.opportunityNumber }}</div>
          </div>
        </template>

        <template #cell-companyId="{ value }">
          {{ getCompanyName(value) }}
        </template>

        <template #cell-stage="{ value }">
          <BaseBadge :variant="getStageVariant(value)" size="sm">
            {{ formatStageLabel(value) }}
          </BaseBadge>
        </template>

        <template #cell-priority="{ value }">
          <BaseBadge :variant="getPriorityVariant(value)" size="sm">
            {{ value }}
          </BaseBadge>
        </template>

        <template #cell-estimatedValue="{ value }">
          <span class="value-highlight">{{ formatCurrency(value) }}</span>
        </template>

        <template #cell-targetMargin="{ value }">
          <span :class="{ 'margin-warning': value < 20 }">{{ value }}%</span>
        </template>

        <template #cell-winProbability="{ value }">
          <div class="probability-bar">
            <div class="probability-fill" :style="{ width: value + '%' }"></div>
            <span class="probability-text">{{ value }}%</span>
          </div>
        </template>

        <template #cell-actions="{ row }">
          <div class="action-buttons">
            <button class="action-btn" title="View" @click.stop="viewOpportunity(row)">👁️</button>
            <button class="action-btn" title="Edit" @click.stop="editOpportunity(row)">✏️</button>
            <button class="action-btn" title="Create Quote" @click.stop="createQuote(row)">📝</button>
            <button class="action-btn" title="Delete" @click.stop="deleteOpportunity(row)">🗑️</button>
          </div>
        </template>
      </BaseTable>
    </BaseCard>

    <!-- Add/Edit Opportunity Modal -->
    <BaseModal
      v-model="showOpportunityModal"
      :title="editingOpportunity ? 'Edit Opportunity' : 'Create New Opportunity'"
      size="xl"
      @confirm="saveOpportunity"
    >
      <div class="modal-form">
        <div class="form-section">
          <h4>Basic Information</h4>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Opportunity Name *</label>
              <input v-model="opportunityForm.name" type="text" class="form-input" placeholder="Enter opportunity name" />
            </div>
            <div class="form-group">
              <label class="form-label">Company *</label>
              <select v-model="opportunityForm.companyId" class="form-input">
                <option value="">Select company</option>
                <option v-for="company in customersStore.customers" :key="company.id" :value="company.id">
                  {{ company.name }}
                </option>
              </select>
            </div>
          </div>

          <div class="form-row" v-if="opportunityForm.companyId && customerContracts.length > 0">
            <div class="form-group full-width">
              <label class="form-label">Link to Existing Contract/Price Book (Optional)</label>
              <select v-model="opportunityForm.contractId" class="form-input">
                <option value="">-- No Price Book (New Quote) --</option>
                <option v-for="contract in customerContracts" :key="contract.id" :value="contract.id">
                  {{ contract.contractNumber }} - {{ contract.name }}
                  <template v-if="contract.priceBookId">
                    (Has Price Book)
                  </template>
                  - {{ formatCurrency(contract.totalValue) }}
                </option>
              </select>
              <small class="form-hint">
                Select if this opportunity should use pricing from an existing contract. The contract's price book will be auto-applied when creating quotes.
              </small>
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">Description *</label>
            <textarea v-model="opportunityForm.description" class="form-input" rows="3" placeholder="Describe the opportunity"></textarea>
          </div>

          <div class="form-group full-width">
            <label class="form-label">Service Types * (Select one or more)</label>
            <div class="service-types-grid">
              <div class="service-type-category">
                <div class="category-header">Security Systems</div>
                <label class="checkbox-label">
                  <input type="checkbox" value="Security Systems" v-model="opportunityForm.serviceTypes" />
                  <span>Security Systems (General)</span>
                </label>
                <label class="checkbox-label">
                  <input type="checkbox" value="CCTV & Surveillance" v-model="opportunityForm.serviceTypes" />
                  <span>CCTV & Surveillance</span>
                </label>
                <label class="checkbox-label">
                  <input type="checkbox" value="Access Control Systems" v-model="opportunityForm.serviceTypes" />
                  <span>Access Control Systems</span>
                </label>
                <label class="checkbox-label">
                  <input type="checkbox" value="Alarm Systems" v-model="opportunityForm.serviceTypes" />
                  <span>Alarm Systems</span>
                </label>
                <label class="checkbox-label">
                  <input type="checkbox" value="Fire Safety Systems" v-model="opportunityForm.serviceTypes" />
                  <span>Fire Safety Systems</span>
                </label>
              </div>

              <div class="service-type-category">
                <div class="category-header">Guarding Services</div>
                <label class="checkbox-label">
                  <input type="checkbox" value="Security Guards" v-model="opportunityForm.serviceTypes" />
                  <span>Security Guards</span>
                </label>
                <label class="checkbox-label">
                  <input type="checkbox" value="Mobile Patrols" v-model="opportunityForm.serviceTypes" />
                  <span>Mobile Patrols</span>
                </label>
                <label class="checkbox-label">
                  <input type="checkbox" value="Event Security" v-model="opportunityForm.serviceTypes" />
                  <span>Event Security</span>
                </label>
              </div>

              <div class="service-type-category">
                <div class="category-header">Professional Services</div>
                <label class="checkbox-label">
                  <input type="checkbox" value="Facility Management" v-model="opportunityForm.serviceTypes" />
                  <span>Facility Management</span>
                </label>
                <label class="checkbox-label">
                  <input type="checkbox" value="Maintenance Services" v-model="opportunityForm.serviceTypes" />
                  <span>Maintenance Services</span>
                </label>
                <label class="checkbox-label">
                  <input type="checkbox" value="Installation Services" v-model="opportunityForm.serviceTypes" />
                  <span>Installation Services</span>
                </label>
                <label class="checkbox-label">
                  <input type="checkbox" value="Consulting Services" v-model="opportunityForm.serviceTypes" />
                  <span>Consulting Services</span>
                </label>
                <label class="checkbox-label">
                  <input type="checkbox" value="Training Services" v-model="opportunityForm.serviceTypes" />
                  <span>Training Services</span>
                </label>
                <label class="checkbox-label">
                  <input type="checkbox" value="Integration Services" v-model="opportunityForm.serviceTypes" />
                  <span>Integration Services</span>
                </label>
                <label class="checkbox-label">
                  <input type="checkbox" value="Monitoring Services" v-model="opportunityForm.serviceTypes" />
                  <span>Monitoring Services</span>
                </label>
                <label class="checkbox-label">
                  <input type="checkbox" value="Risk Assessment" v-model="opportunityForm.serviceTypes" />
                  <span>Risk Assessment</span>
                </label>
              </div>

              <div class="service-type-category">
                <div class="category-header">Other</div>
                <label class="checkbox-label">
                  <input type="checkbox" value="Rent Service" v-model="opportunityForm.serviceTypes" />
                  <span>Rent Service</span>
                </label>
                <label class="checkbox-label">
                  <input type="checkbox" value="Other" v-model="opportunityForm.serviceTypes" />
                  <span>Other</span>
                </label>
              </div>
            </div>
            <div class="selected-services" v-if="opportunityForm.serviceTypes.length > 0">
              <span class="selected-label">Selected ({{ opportunityForm.serviceTypes.length }}):</span>
              <span class="selected-tags">
                <span v-for="service in opportunityForm.serviceTypes" :key="service" class="service-tag">
                  {{ service }}
                  <button type="button" @click="removeServiceType(service)" class="remove-tag">×</button>
                </span>
              </span>
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">Priority *</label>
            <select v-model="opportunityForm.priority" class="form-input">
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
              <option value="critical">Critical</option>
            </select>
          </div>
        </div>

        <div class="form-section">
          <h4>Financial Details</h4>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Estimated Value (SAR) *</label>
              <input v-model.number="opportunityForm.estimatedValue" type="number" step="0.01" class="form-input" placeholder="0.00" />
            </div>
            <div class="form-group">
              <label class="form-label">Target Margin (%) *</label>
              <input v-model.number="opportunityForm.targetMargin" type="number" step="0.1" class="form-input" placeholder="20" min="0" max="100" />
            </div>
            <div class="form-group">
              <label class="form-label">Estimated Cost (Calculated)</label>
              <input :value="formatCurrency(calculatedCost)" type="text" class="form-input" disabled />
            </div>
          </div>
        </div>

        <div class="form-section">
          <h4>Sales Pipeline</h4>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Stage *</label>
              <select v-model="opportunityForm.stage" class="form-input">
                <option value="qualification">Qualification</option>
                <option value="proposal">Proposal</option>
                <option value="negotiation">Negotiation</option>
                <option value="closed_won">Closed Won</option>
                <option value="closed_lost">Closed Lost</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">Win Probability (%) *</label>
              <input v-model.number="opportunityForm.winProbability" type="number" class="form-input" placeholder="50" min="0" max="100" />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Request Date</label>
              <input v-model="opportunityForm.requestDate" type="date" class="form-input" />
            </div>
          </div>
        </div>

        <div class="form-section">
          <h4>Lead Source - Team That Got This Opportunity</h4>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Team That Got Lead *</label>
              <select v-model="opportunityForm.teamId" class="form-input">
                <option value="">Select team</option>
                <option v-for="team in activeTeams" :key="team.id" :value="team.id">
                  {{ team.name }} ({{ team.department }})
                </option>
              </select>
              <small class="form-hint">Which team acquired this opportunity (Marketing, Sales, etc.)</small>
            </div>
            <div class="form-group">
              <label class="form-label">Person Who Got Lead</label>
              <select v-model="opportunityForm.createdByPersonId" class="form-input">
                <option value="">Select person (optional)</option>
                <option v-for="member in allActiveTeamMembers" :key="member.id" :value="member.id">
                  {{ member.name }} - {{ member.role }}
                </option>
              </select>
              <small class="form-hint">Specific person from any team who created/got this lead</small>
            </div>
          </div>
        </div>

        <div class="form-section">
          <h4>Quote Team Assignment - Required for Creating Quotes</h4>
          <p class="section-description">
            To create quotes for this opportunity, you must assign both a Sales person and a Pre-Sales person.
          </p>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Sales Person (for Quotes)</label>
              <select v-model="opportunityForm.quoteSalesPersonId" class="form-input">
                <option value="">Select sales person</option>
                <option v-for="member in activeSalesPeople" :key="member.id" :value="member.id">
                  {{ member.name }} - {{ member.role }}
                </option>
              </select>
              <small class="form-hint">Required before creating quotes - must be from Sales team</small>
            </div>
            <div class="form-group">
              <label class="form-label">Pre-Sales Person (for Quotes)</label>
              <select v-model="opportunityForm.quotePresalesPersonId" class="form-input">
                <option value="">Select pre-sales person</option>
                <option v-for="member in activePresalesPeople" :key="member.id" :value="member.id">
                  {{ member.name }} - {{ member.role }}
                </option>
              </select>
              <small class="form-hint">Required before creating quotes - must be from Pre-Sales team</small>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Sales Executive (Legacy)</label>
              <select v-model="opportunityForm.salesExecutiveId" class="form-input">
                <option value="">Select sales executive</option>
                <option v-for="user in salesUsers" :key="user.id" :value="user.id">
                  {{ user.fullName }}
                </option>
              </select>
              <small class="form-hint">Optional: Legacy field for backward compatibility</small>
            </div>
            <div class="form-group">
              <label class="form-label">Presales Support (Legacy)</label>
              <select v-model="opportunityForm.presalesId" class="form-input">
                <option value="">Select presales support</option>
                <option v-for="user in presalesUsers" :key="user.id" :value="user.id">
                  {{ user.fullName }}
                </option>
              </select>
              <small class="form-hint">Optional: Legacy field for backward compatibility</small>
            </div>
          </div>
        </div>

        <div class="form-section">
          <h4>Additional Information</h4>
          <div class="form-group">
            <label class="form-label">Notes</label>
            <textarea v-model="opportunityForm.notes" class="form-input" rows="3" placeholder="Additional notes..."></textarea>
          </div>
        </div>
      </div>
    </BaseModal>

    <!-- View Opportunity Detail Modal -->
    <BaseModal
      v-model="showDetailModal"
      :title="viewingOpportunity?.name || 'Opportunity Details'"
      size="xl"
      hide-footer
    >
      <div class="detail-view" v-if="viewingOpportunity">
        <div class="detail-header">
          <div class="detail-badges">
            <BaseBadge :variant="getStageVariant(viewingOpportunity.stage)">
              {{ formatStageLabel(viewingOpportunity.stage) }}
            </BaseBadge>
            <BaseBadge :variant="getPriorityVariant(viewingOpportunity.priority)">
              {{ viewingOpportunity.priority }}
            </BaseBadge>
          </div>
          <div class="detail-number">{{ viewingOpportunity.opportunityNumber }}</div>
        </div>

        <div class="detail-section">
          <h4>Basic Information</h4>
          <div class="detail-row">
            <div class="detail-label">Company:</div>
            <div class="detail-value">{{ getCompanyName(viewingOpportunity.companyId) }}</div>
          </div>
          <div class="detail-row">
            <div class="detail-label">Description:</div>
            <div class="detail-value">{{ viewingOpportunity.description }}</div>
          </div>
          <div class="detail-row">
            <div class="detail-label">Service Type:</div>
            <div class="detail-value">{{ viewingOpportunity.serviceType }}</div>
          </div>
        </div>

        <div class="detail-section">
          <h4>Financial Details</h4>
          <div class="detail-row">
            <div class="detail-label">Estimated Value:</div>
            <div class="detail-value value-highlight">{{ formatCurrency(viewingOpportunity.estimatedValue) }}</div>
          </div>
          <div class="detail-row">
            <div class="detail-label">Target Margin:</div>
            <div class="detail-value" :class="{ 'margin-warning': viewingOpportunity.targetMargin < 20 }">
              {{ viewingOpportunity.targetMargin }}%
            </div>
          </div>
          <div class="detail-row">
            <div class="detail-label">Estimated Cost:</div>
            <div class="detail-value">{{ formatCurrency(viewingOpportunity.estimatedCost) }}</div>
          </div>
          <div class="detail-row">
            <div class="detail-label">Win Probability:</div>
            <div class="detail-value">{{ viewingOpportunity.winProbability }}%</div>
          </div>
        </div>

        <div class="detail-section">
          <h4>Timeline</h4>
          <div class="detail-row">
            <div class="detail-label">Request Date:</div>
            <div class="detail-value">{{ formatDate(viewingOpportunity.requestDate) }}</div>
          </div>
        </div>

        <div class="detail-section">
          <h4>Lead Source</h4>
          <div class="detail-row" v-if="viewingOpportunity.teamId">
            <div class="detail-label">Team That Got Lead:</div>
            <div class="detail-value">{{ getTeamName(viewingOpportunity.teamId) }}</div>
          </div>
          <div class="detail-row" v-if="viewingOpportunity.createdByPersonId">
            <div class="detail-label">Person Who Got Lead:</div>
            <div class="detail-value">{{ getTeamMemberName(viewingOpportunity.createdByPersonId) }}</div>
          </div>
        </div>

        <div class="detail-section">
          <h4>Quote Team Assignment</h4>
          <div class="detail-row" v-if="viewingOpportunity.quoteSalesPersonId">
            <div class="detail-label">Sales Person (Quotes):</div>
            <div class="detail-value">{{ getTeamMemberName(viewingOpportunity.quoteSalesPersonId) }}</div>
          </div>
          <div class="detail-row" v-if="viewingOpportunity.quotePresalesPersonId">
            <div class="detail-label">Pre-Sales Person (Quotes):</div>
            <div class="detail-value">{{ getTeamMemberName(viewingOpportunity.quotePresalesPersonId) }}</div>
          </div>
          <div class="detail-row" v-if="!viewingOpportunity.quoteSalesPersonId || !viewingOpportunity.quotePresalesPersonId">
            <div class="detail-label">Status:</div>
            <div class="detail-value warning-text">⚠️ Quote team not fully assigned - cannot create quotes until both Sales and Pre-Sales persons are assigned</div>
          </div>
          <div class="detail-row" v-if="viewingOpportunity.salesExecutiveId">
            <div class="detail-label">Sales Executive (Legacy):</div>
            <div class="detail-value">{{ getUserName(viewingOpportunity.salesExecutiveId) }}</div>
          </div>
          <div class="detail-row" v-if="viewingOpportunity.presalesId">
            <div class="detail-label">Presales Support (Legacy):</div>
            <div class="detail-value">{{ getUserName(viewingOpportunity.presalesId) }}</div>
          </div>
        </div>

        <div class="detail-section" v-if="viewingOpportunity.quoteDates && viewingOpportunity.quoteDates.length > 0">
          <h4>Quote Timeline</h4>
          <div v-for="quoteDate in viewingOpportunity.quoteDates" :key="quoteDate.quoteId" class="detail-row">
            <div class="detail-label">Quote {{ quoteDate.quoteId }}:</div>
            <div class="detail-value">
              Sent: {{ formatDate(quoteDate.sentDate) }}
              <span v-if="quoteDate.respondedDate"> | Responded: {{ formatDate(quoteDate.respondedDate) }}</span>
            </div>
          </div>
        </div>

        <div class="detail-section" v-if="viewingOpportunity.notes">
          <h4>Notes</h4>
          <div class="detail-value">{{ viewingOpportunity.notes }}</div>
        </div>

        <!-- Linked Quotes Section -->
        <div class="detail-section">
          <div class="section-header">
            <h4>Linked Quotes ({{ opportunityQuotes.length }})</h4>
            <button class="btn btn-sm btn-primary" @click="createQuote(viewingOpportunity)">
              ➕ New Quote
            </button>
          </div>

          <div v-if="opportunityQuotes.length === 0" class="empty-state">
            <p>No quotes created yet for this opportunity</p>
            <button class="btn btn-primary" @click="createQuote(viewingOpportunity)">
              Create First Quote
            </button>
          </div>

          <div v-else class="quotes-list">
            <div
              v-for="quote in opportunityQuotes"
              :key="quote.id"
              class="quote-card"
              @click="viewQuote(quote)"
            >
              <div class="quote-header">
                <div class="quote-info">
                  <span class="quote-number">{{ quote.quoteNumber }}</span>
                  <span class="quote-version">v{{ quote.version }}</span>
                </div>
                <BaseBadge :variant="getQuoteStatusVariant(quote.status)" size="sm">
                  {{ formatQuoteStatus(quote.status) }}
                </BaseBadge>
              </div>

              <div class="quote-details">
                <div class="quote-row">
                  <span class="label">Date:</span>
                  <span class="value">{{ formatDate(quote.quoteDate) }}</span>
                </div>
                <div class="quote-row">
                  <span class="label">Total:</span>
                  <span class="value value-highlight">{{ formatCurrency(quote.total) }}</span>
                </div>
                <div class="quote-row">
                  <span class="label">Margin:</span>
                  <span class="value" :class="{ 'margin-warning': quote.marginPercent < 20 }">
                    {{ quote.marginPercent.toFixed(1) }}%
                  </span>
                </div>
                <div class="quote-row">
                  <span class="label">Valid Until:</span>
                  <span class="value">{{ formatDate(quote.validUntil) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="detail-actions">
          <button class="btn btn-primary" @click="editOpportunity(viewingOpportunity)">
            Edit Opportunity
          </button>
          <button class="btn btn-secondary" @click="showDetailModal = false">
            Close
          </button>
        </div>
      </div>
    </BaseModal>

    <!-- Delete Confirmation Modal -->
    <BaseModal
      v-model="showDeleteModal"
      title="Confirm Delete"
      size="sm"
      confirm-text="Delete"
      @confirm="confirmDelete"
    >
      <div class="delete-modal-content">
        <div class="delete-icon">🗑️</div>
        <p class="delete-message">Are you sure you want to delete <strong>{{ deleteItem?.name }}</strong>?</p>
        <p class="delete-warning">This action cannot be undone.</p>
      </div>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useOpportunitiesStore } from '@/stores/opportunities'
import { useCustomersStore } from '@/stores/customers'
import { useUsersStore } from '@/stores/users'
import { useQuotesStore } from '@/stores/quotes'
import { useContractsStore } from '@/stores/contracts'
import { usePriceBooksStore } from '@/stores/priceBooks'
import { useTeamsStore } from '@/stores/teams'
import { useToast } from '@/composables/useToast'
import BaseCard from '@/components/UI/BaseCard.vue'
import BaseTable from '@/components/UI/BaseTable.vue'
import BaseBadge from '@/components/UI/BaseBadge.vue'
import BaseModal from '@/components/UI/BaseModal.vue'
import type { TableColumn } from '@/components/UI/BaseTable.vue'
import type { OpportunityStage, Priority, ServiceType } from '@/types'

const router = useRouter()
const opportunitiesStore = useOpportunitiesStore()
const customersStore = useCustomersStore()
const usersStore = useUsersStore()
const quotesStore = useQuotesStore()
const contractsStore = useContractsStore()
const priceBooksStore = usePriceBooksStore()
const teamsStore = useTeamsStore()
const { success, info } = useToast()

const viewMode = ref<'pipeline' | 'list'>('pipeline')
const searchQuery = ref('')
const filterStage = ref<OpportunityStage | ''>('')
const filterPriority = ref<Priority | ''>('')

// Modal states
const showOpportunityModal = ref(false)
const editingOpportunity = ref<any>(null)
const showDetailModal = ref(false)
const viewingOpportunity = ref<any>(null)
const showDeleteModal = ref(false)
const deleteItem = ref<any>(null)

// Form data
const opportunityForm = ref({
  name: '',
  companyId: '',
  contractId: '',
  description: '',
  serviceTypes: [] as ServiceType[],
  priority: 'medium' as Priority,
  estimatedValue: 0,
  targetMargin: 20,
  stage: 'qualification' as OpportunityStage,
  winProbability: 50,
  requestDate: '',
  teamId: '',
  createdByPersonId: '',
  quoteSalesPersonId: '',
  quotePresalesPersonId: '',
  salesExecutiveId: '',
  presalesId: '',
  notes: ''
})

const pipelineStages = [
  { key: 'qualification' as OpportunityStage, label: 'Qualification' },
  { key: 'proposal' as OpportunityStage, label: 'Proposal' },
  { key: 'negotiation' as OpportunityStage, label: 'Negotiation' },
  { key: 'closed_won' as OpportunityStage, label: 'Closed Won' },
  { key: 'closed_lost' as OpportunityStage, label: 'Closed Lost' }
]

const tableColumns: TableColumn[] = [
  { key: 'name', label: 'Opportunity', sortable: true },
  { key: 'companyId', label: 'Company', sortable: true },
  { key: 'stage', label: 'Stage', sortable: true },
  { key: 'priority', label: 'Priority', sortable: true },
  { key: 'estimatedValue', label: 'Value', sortable: true },
  { key: 'targetMargin', label: 'Margin', sortable: true },
  { key: 'winProbability', label: 'Win %', sortable: true },
  { key: 'actions', label: 'Actions', align: 'right' }
]

// Computed properties
const calculatedCost = computed(() => {
  const margin = opportunityForm.value.targetMargin / 100
  return opportunityForm.value.estimatedValue * (1 - margin)
})

const salesUsers = computed(() => {
  return usersStore.users.filter(u => u.role === 'sales_executive' || u.role === 'presales_manager')
})

const presalesUsers = computed(() => {
  return usersStore.users.filter(u => u.role === 'presales')
})

// Teams for assignment
const activeTeams = computed(() => {
  return teamsStore.activeTeams
})

// Active sales people from Sales team only
const activeSalesPeople = computed(() => {
  const salesTeam = teamsStore.teams.find(t => t.department === 'Sales' && t.status === 'active')
  return salesTeam ? salesTeam.members.filter(m => m.status === 'active') : []
})

// Active presales people from Pre-Sales team only
const activePresalesPeople = computed(() => {
  const presalesTeam = teamsStore.teams.find(t => t.department === 'Pre-Sales' && t.status === 'active')
  return presalesTeam ? presalesTeam.members.filter(m => m.status === 'active') : []
})

// All active team members (for lead creator selection)
const allActiveTeamMembers = computed(() => {
  return teamsStore.getAllTeamMembers().filter(m => m.status === 'active')
})

// Quotes for viewing opportunity
const opportunityQuotes = computed(() => {
  if (!viewingOpportunity.value) return []
  return quotesStore.getQuotesByOpportunity(viewingOpportunity.value.id)
})

// Active contracts for selected customer
const customerContracts = computed(() => {
  if (!opportunityForm.value.companyId) return []
  return contractsStore.contracts.filter(c =>
    c.customerId === opportunityForm.value.companyId &&
    c.status === 'active'
  )
})

// Filtered opportunities
const filteredOpportunities = computed(() => {
  let opps = opportunitiesStore.opportunities

  // Filter by stage
  if (filterStage.value) {
    opps = opps.filter(o => o.stage === filterStage.value)
  }

  // Filter by priority
  if (filterPriority.value) {
    opps = opps.filter(o => o.priority === filterPriority.value)
  }

  // Search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    opps = opps.filter(o =>
      o.name.toLowerCase().includes(query) ||
      o.opportunityNumber.toLowerCase().includes(query) ||
      o.description.toLowerCase().includes(query) ||
      getCompanyName(o.companyId).toLowerCase().includes(query)
    )
  }

  return opps
})

// Helper functions
const getOpportunitiesByStage = (stage: OpportunityStage) => {
  return filteredOpportunities.value.filter(o => o.stage === stage)
}

const getTotalValueByStage = (stage: OpportunityStage) => {
  return getOpportunitiesByStage(stage).reduce((sum, o) => sum + o.estimatedValue, 0)
}

const getCompanyName = (companyId: string) => {
  const company = customersStore.getCustomerById(companyId)
  return company?.name || 'Unknown'
}

const getUserName = (userId: string) => {
  const user = usersStore.getUserById(userId)
  return user?.fullName || 'Unknown'
}

const getTeamName = (teamId: string) => {
  const team = teamsStore.getTeamById(teamId)
  return team?.name || 'Unknown'
}

const getTeamMemberName = (memberId: string) => {
  const allMembers = teamsStore.getAllTeamMembers()
  const member = allMembers.find(m => m.id === memberId)
  return member?.name || 'Unknown'
}

const formatCurrency = (value: number) => {
  return new Intl.NumberFormat('en-SA', {
    style: 'currency',
    currency: 'SAR',
    minimumFractionDigits: 0
  }).format(value)
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric'
  })
}

const formatStageLabel = (stage: OpportunityStage) => {
  return stage.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())
}

const getStageVariant = (stage: OpportunityStage) => {
  const variants: Record<OpportunityStage, 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
    qualification: 'info',
    proposal: 'primary',
    negotiation: 'warning',
    closed_won: 'success',
    closed_lost: 'danger'
  }
  return variants[stage]
}

const getPriorityVariant = (priority: Priority) => {
  const variants: Record<Priority, 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
    low: 'default',
    medium: 'info',
    high: 'warning',
    critical: 'danger'
  }
  return variants[priority]
}

// Helper function to remove service type
const removeServiceType = (service: ServiceType) => {
  const index = opportunityForm.value.serviceTypes.indexOf(service)
  if (index > -1) {
    opportunityForm.value.serviceTypes.splice(index, 1)
  }
}

// Form reset function
const resetOpportunityForm = () => {
  opportunityForm.value = {
    name: '',
    companyId: '',
    contractId: '',
    description: '',
    serviceTypes: [],
    priority: 'medium',
    estimatedValue: 0,
    targetMargin: 20,
    stage: 'qualification',
    winProbability: 50,
    requestDate: '',
    teamId: '',
    createdByPersonId: '',
    quoteSalesPersonId: '',
    quotePresalesPersonId: '',
    salesExecutiveId: '',
    presalesId: '',
    notes: ''
  }
}

// CRUD Actions
const createOpportunity = () => {
  editingOpportunity.value = null
  resetOpportunityForm()
  showOpportunityModal.value = true
}

const viewOpportunity = (opp: any) => {
  viewingOpportunity.value = opp
  showDetailModal.value = true
}

const editOpportunity = (opp: any) => {
  if (showDetailModal.value) {
    showDetailModal.value = false
  }
  editingOpportunity.value = opp
  opportunityForm.value = {
    name: opp.name,
    companyId: opp.companyId,
    contractId: opp.contractId || '',
    description: opp.description,
    serviceTypes: Array.isArray(opp.serviceType) ? opp.serviceType : [opp.serviceType],
    priority: opp.priority,
    estimatedValue: opp.estimatedValue,
    targetMargin: opp.targetMargin,
    stage: opp.stage,
    winProbability: opp.winProbability,
    requestDate: opp.requestDate,
    teamId: opp.teamId || '',
    createdByPersonId: opp.createdByPersonId || '',
    quoteSalesPersonId: opp.quoteSalesPersonId || '',
    quotePresalesPersonId: opp.quotePresalesPersonId || '',
    salesExecutiveId: opp.salesExecutiveId || '',
    presalesId: opp.presalesId || '',
    notes: opp.notes || ''
  }
  showOpportunityModal.value = true
}

const saveOpportunity = () => {
  // Validation
  if (!opportunityForm.value.name || !opportunityForm.value.companyId ||
      !opportunityForm.value.description || opportunityForm.value.serviceTypes.length === 0 ||
      !opportunityForm.value.teamId) {
    info('Please fill in all required fields including the team that got this lead')
    return
  }

  // Calculate estimated cost
  const estimatedCost = opportunityForm.value.estimatedValue *
    (1 - opportunityForm.value.targetMargin / 100)

  const oppData = {
    name: opportunityForm.value.name,
    companyId: opportunityForm.value.companyId,
    contractId: opportunityForm.value.contractId || undefined,
    description: opportunityForm.value.description,
    serviceType: opportunityForm.value.serviceTypes.length === 1
      ? opportunityForm.value.serviceTypes[0]
      : opportunityForm.value.serviceTypes,
    priority: opportunityForm.value.priority,
    estimatedValue: opportunityForm.value.estimatedValue,
    targetMargin: opportunityForm.value.targetMargin,
    estimatedCost,
    stage: opportunityForm.value.stage,
    winProbability: opportunityForm.value.winProbability,
    requestDate: opportunityForm.value.requestDate,
    teamId: opportunityForm.value.teamId,
    createdByPersonId: opportunityForm.value.createdByPersonId || undefined,
    quoteSalesPersonId: opportunityForm.value.quoteSalesPersonId || undefined,
    quotePresalesPersonId: opportunityForm.value.quotePresalesPersonId || undefined,
    salesExecutiveId: opportunityForm.value.salesExecutiveId || undefined,
    presalesId: opportunityForm.value.presalesId || undefined,
    notes: opportunityForm.value.notes || undefined
  }

  if (editingOpportunity.value) {
    opportunitiesStore.updateOpportunity(editingOpportunity.value.id, oppData)
    success('Opportunity updated successfully')
  } else {
    opportunitiesStore.addOpportunity(oppData as any)
    success('Opportunity created successfully')
  }

  showOpportunityModal.value = false
  editingOpportunity.value = null
  resetOpportunityForm()
}

const deleteOpportunity = (opp: any) => {
  deleteItem.value = opp
  showDeleteModal.value = true
}

const confirmDelete = () => {
  if (!deleteItem.value) return
  opportunitiesStore.deleteOpportunity(deleteItem.value.id)
  success('Opportunity deleted successfully')
  showDeleteModal.value = false
  deleteItem.value = null
}

const createQuote = (opp: any) => {
  // Validate that both sales and presales are assigned before creating quote
  if (!opp.quoteSalesPersonId || !opp.quotePresalesPersonId) {
    info('Cannot create quote: Both Sales and Pre-Sales persons must be assigned to this opportunity first. Please edit the opportunity and assign the quote team.')
    return
  }

  // Navigate to quote builder with opportunity pre-filled
  router.push({
    path: '/quoting',
    query: { opportunityId: opp.id }
  })
}

const viewQuote = (quote: any) => {
  // Navigate to quote view
  router.push(`/quoting?quoteId=${quote.id}`)
}

// Quote status formatting
function formatQuoteStatus(status: string): string {
  const statusMap: Record<string, string> = {
    draft: 'Draft',
    pending_approval: 'Pending Approval',
    approved: 'Approved',
    sent: 'Sent',
    accepted: 'Accepted',
    declined: 'Declined',
    expired: 'Expired'
  }
  return statusMap[status] || status
}

function getQuoteStatusVariant(status: string): 'default' | 'success' | 'warning' | 'danger' | 'info' {
  const variantMap: Record<string, 'default' | 'success' | 'warning' | 'danger' | 'info'> = {
    draft: 'default',
    pending_approval: 'warning',
    approved: 'info',
    sent: 'info',
    accepted: 'success',
    declined: 'danger',
    expired: 'danger'
  }
  return variantMap[status] || 'default'
}
</script>

<style scoped>
.opportunities-view {
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
  flex-wrap: wrap;
}

.search-input {
  flex: 1;
  min-width: 250px;
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

.filter-select {
  padding: 0.75rem 1rem;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text-primary);
  border-radius: 8px;
  font-size: 0.95rem;
  cursor: pointer;
  transition: all 0.2s;
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

.view-toggle {
  display: flex;
  gap: 0.5rem;
  background: var(--color-surface);
  padding: 0.25rem;
  border-radius: 8px;
  width: fit-content;
}

.toggle-btn {
  padding: 0.5rem 1rem;
  background: none;
  border: none;
  border-radius: 6px;
  font-weight: 500;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.toggle-btn:hover {
  background: var(--color-surface-hover);
}

.toggle-btn.active {
  background: var(--color-primary);
  color: #fff;
}

/* Pipeline View */
.pipeline-view {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 1rem;
  overflow-x: auto;
  padding-bottom: 1rem;
}

.pipeline-column {
  min-width: 280px;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.column-header {
  background: var(--color-surface);
  padding: 1rem;
  border-radius: 8px;
  box-shadow: var(--shadow-sm);
}

.column-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 0.5rem 0;
}

.column-count {
  display: inline-block;
  background: var(--color-primary-light);
  color: var(--color-primary);
  padding: 0.125rem 0.5rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
  margin-right: 0.5rem;
}

.column-value {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  font-weight: 600;
}

.column-cards {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  min-height: 200px;
}

.opportunity-card {
  background: var(--color-surface);
  padding: 1rem;
  border-radius: 8px;
  box-shadow: var(--shadow-sm);
  cursor: pointer;
  transition: all 0.2s;
  border-left: 3px solid transparent;
}

.opportunity-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.opportunity-card.priority-critical {
  border-left-color: var(--color-danger);
}

.opportunity-card.priority-high {
  border-left-color: var(--color-warning);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.win-probability {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-success);
}

.card-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 0.75rem 0;
  line-height: 1.4;
}

.card-meta {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.meta-row {
  display: flex;
  justify-content: space-between;
  font-size: 0.8rem;
}

.meta-label {
  color: var(--color-text-secondary);
}

.meta-value {
  font-weight: 500;
  color: var(--color-text-primary);
}

.value-highlight {
  color: var(--color-primary);
  font-weight: 600;
}

.margin-warning {
  color: var(--color-danger);
  font-weight: 600;
}

.card-footer {
  padding-top: 0.75rem;
  border-top: 1px solid var(--color-border);
}

.sales-exec {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
}

.footer-team-info {
  width: 100%;
}

.team-label {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
}

.team-label.warning {
  color: var(--color-warning);
  font-weight: 600;
}

.warning-text {
  color: var(--color-warning);
  font-weight: 600;
}

.empty-column {
  text-align: center;
  padding: 2rem 1rem;
  color: var(--color-text-tertiary);
  font-size: 0.875rem;
}

/* List View */
.opp-name {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.name-text {
  font-weight: 500;
  color: var(--color-text-primary);
}

.name-subtitle {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
}

.probability-bar {
  position: relative;
  width: 100%;
  height: 24px;
  background: var(--color-border-light);
  border-radius: 4px;
  overflow: hidden;
}

.probability-fill {
  position: absolute;
  top: 0;
  left: 0;
  height: 100%;
  background: linear-gradient(90deg, var(--color-success) 0%, var(--color-success-light) 100%);
  transition: width 0.3s ease;
}

.probability-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--color-text-primary);
  z-index: 1;
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

.action-btn:hover {
  transform: scale(1.2);
  opacity: 1;
}

/* Responsive */
@media (max-width: 1400px) {
  .pipeline-view {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 900px) {
  .pipeline-view {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 600px) {
  .pipeline-view {
    grid-template-columns: 1fr;
  }
}

/* Modal Form Styles */
.modal-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid var(--color-border);
}

.form-section:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.form-section h4 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.section-description {
  margin: 0.5rem 0 1rem 0;
  padding: 0.75rem;
  background: var(--color-info-light);
  border-left: 3px solid var(--color-info);
  color: var(--color-text-secondary);
  font-size: 0.9rem;
  border-radius: 4px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-label {
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--color-text-primary);
}

.form-hint {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  font-style: italic;
  margin-top: 0.25rem;
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

.form-input::placeholder {
  color: var(--color-text-secondary);
}

.form-input:disabled {
  background: var(--color-background);
  opacity: 0.6;
  cursor: not-allowed;
}

textarea.form-input {
  resize: vertical;
  font-family: inherit;
}

.full-width {
  grid-column: 1 / -1;
}

/* Service Types Multi-Select Styles */
.service-types-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
  padding: 1rem;
  background: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  max-height: 400px;
  overflow-y: auto;
}

.service-type-category {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.category-header {
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--color-primary);
  padding-bottom: 0.5rem;
  border-bottom: 2px solid var(--color-primary);
  margin-bottom: 0.25rem;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem;
  cursor: pointer;
  border-radius: 4px;
  transition: background 0.2s;
  font-size: 0.9rem;
}

.checkbox-label:hover {
  background: var(--color-surface-hover);
}

.checkbox-label input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: var(--color-primary);
}

.checkbox-label span {
  user-select: none;
}

.selected-services {
  margin-top: 1rem;
  padding: 1rem;
  background: var(--color-surface);
  border-radius: 8px;
  border: 1px solid var(--color-border);
}

.selected-label {
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--color-text-primary);
  margin-right: 0.5rem;
}

.selected-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.75rem;
}

.service-tag {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.75rem;
  background: var(--color-primary-light);
  color: var(--color-primary);
  border-radius: 16px;
  font-size: 0.85rem;
  font-weight: 500;
}

.remove-tag {
  background: none;
  border: none;
  color: var(--color-primary);
  font-size: 1.2rem;
  line-height: 1;
  cursor: pointer;
  padding: 0;
  width: 18px;
  height: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.2s;
}

.remove-tag:hover {
  background: var(--color-primary);
  color: white;
}

/* Delete Modal Styles */
.delete-modal-content {
  text-align: center;
  padding: 1rem;
}

.delete-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.delete-message {
  font-size: 1.05rem;
  color: var(--color-text-primary);
  margin-bottom: 0.75rem;
}

.delete-warning {
  font-size: 0.9rem;
  color: var(--color-danger);
}

/* Detail View Styles */
.detail-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 1rem;
  border-bottom: 2px solid var(--color-border);
  margin-bottom: 0.5rem;
}

.detail-badges {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.detail-number {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  font-family: 'Courier New', monospace;
}

.detail-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.detail-section h4 {
  margin: 0 0 0.5rem 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text-primary);
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--color-border);
}

.detail-row {
  display: grid;
  grid-template-columns: 180px 1fr;
  gap: 1rem;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--color-border);
}

.detail-row:last-child {
  border-bottom: none;
}

.detail-label {
  font-weight: 600;
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}

.detail-value {
  color: var(--color-text-primary);
  font-size: 0.95rem;
}

.detail-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--color-border);
}

/* Quotes Section */
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.section-header h4 {
  margin: 0;
}

.empty-state {
  text-align: center;
  padding: 2rem;
  background: var(--color-background);
  border-radius: 10px;
  border: 2px dashed var(--color-border);
}

.empty-state p {
  margin-bottom: 1rem;
  color: var(--color-text-secondary);
}

.quotes-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.quote-card {
  background: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  padding: 1rem;
  cursor: pointer;
  transition: all 0.2s;
}

.quote-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-sm);
  transform: translateY(-2px);
}

.quote-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--color-border);
}

.quote-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.quote-number {
  font-weight: 600;
  color: var(--color-primary);
  font-size: 1rem;
}

.quote-version {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  padding: 0.25rem 0.5rem;
  background: var(--color-surface);
  border-radius: 4px;
}

.quote-details {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.5rem;
}

.quote-row {
  display: flex;
  justify-content: space-between;
  font-size: 0.9rem;
}

.quote-row .label {
  color: var(--color-text-secondary);
  font-weight: 500;
}

.quote-row .value {
  font-weight: 600;
  color: var(--color-text-primary);
}

.btn-sm {
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
}

.btn-secondary {
  background: var(--color-surface);
  color: var(--color-text-primary);
  border: 1px solid var(--color-border);
}

.btn-secondary:hover {
  background: var(--color-background);
  border-color: var(--color-text-secondary);
}
</style>
