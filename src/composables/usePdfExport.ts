import type { Quote, Contract, QuoteAttachment } from '@/types'

export function usePdfExport() {
  function generateQuotePdf(quote: Quote, lineItems: any[], customerName: string, contactName: string, attachments: QuoteAttachment[] = []) {
    // Create a hidden iframe for printing
    const iframe = document.createElement('iframe')
    iframe.style.position = 'absolute'
    iframe.style.width = '0'
    iframe.style.height = '0'
    iframe.style.border = 'none'
    document.body.appendChild(iframe)

    const iframeDoc = iframe.contentWindow?.document
    if (!iframeDoc) return

    // Generate HTML content
    const html = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Quote ${quote.quoteNumber}</title>
  <style>
    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }

    body {
      font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
      padding: 40px;
      color: #333;
      line-height: 1.6;
    }

    .header {
      display: flex;
      justify-content: space-between;
      align-items: start;
      margin-bottom: 40px;
      padding-bottom: 20px;
      border-bottom: 3px solid #2563eb;
    }

    .company-info h1 {
      color: #2563eb;
      font-size: 28px;
      margin-bottom: 5px;
    }

    .company-info p {
      color: #666;
      font-size: 14px;
    }

    .quote-info {
      text-align: right;
    }

    .quote-info h2 {
      font-size: 32px;
      color: #2563eb;
      margin-bottom: 10px;
    }

    .quote-info p {
      font-size: 14px;
      margin: 3px 0;
    }

    .section {
      margin-bottom: 30px;
    }

    .section-title {
      font-size: 18px;
      color: #2563eb;
      margin-bottom: 15px;
      padding-bottom: 5px;
      border-bottom: 2px solid #e5e7eb;
    }

    .customer-details {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 20px;
      margin-bottom: 30px;
    }

    .detail-box {
      background: #f9fafb;
      padding: 20px;
      border-radius: 8px;
    }

    .detail-box h3 {
      font-size: 14px;
      color: #666;
      margin-bottom: 10px;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }

    .detail-box p {
      margin: 5px 0;
      font-size: 14px;
    }

    table {
      width: 100%;
      border-collapse: collapse;
      margin: 20px 0;
    }

    thead {
      background: #2563eb;
      color: white;
    }

    th {
      padding: 12px;
      text-align: left;
      font-size: 13px;
      font-weight: 600;
    }

    td {
      padding: 12px;
      border-bottom: 1px solid #e5e7eb;
      font-size: 13px;
    }

    tbody tr:hover {
      background: #f9fafb;
    }

    .text-right {
      text-align: right;
    }

    .totals {
      margin-top: 30px;
      display: flex;
      justify-content: flex-end;
    }

    .totals-box {
      width: 350px;
      border: 2px solid #e5e7eb;
      border-radius: 8px;
      overflow: hidden;
    }

    .total-row {
      display: flex;
      justify-content: space-between;
      padding: 12px 20px;
      border-bottom: 1px solid #e5e7eb;
    }

    .total-row:last-child {
      border-bottom: none;
    }

    .total-row.grand-total {
      background: #2563eb;
      color: white;
      font-weight: 700;
      font-size: 18px;
    }

    .total-row.margin {
      background: #10b981;
      color: white;
      font-weight: 600;
    }

    .terms {
      margin-top: 40px;
      padding: 20px;
      background: #f9fafb;
      border-radius: 8px;
    }

    .terms h3 {
      color: #2563eb;
      margin-bottom: 10px;
    }

    .terms p {
      font-size: 13px;
      line-height: 1.8;
      color: #666;
    }

    .footer {
      margin-top: 50px;
      padding-top: 20px;
      border-top: 2px solid #e5e7eb;
      text-align: center;
      color: #999;
      font-size: 12px;
    }

    @media print {
      body {
        padding: 20px;
      }

      .header {
        page-break-after: avoid;
      }

      table {
        page-break-inside: avoid;
      }
    }
  </style>
</head>
<body>
  <div class="header">
    <div class="company-info">
      <h1>Your Company Name</h1>
      <p>Security Solutions & Services</p>
      <p>Riyadh, Saudi Arabia</p>
      <p>VAT: 300000000000003</p>
    </div>
    <div class="quote-info">
      <h2>QUOTATION</h2>
      <p><strong>Quote #:</strong> ${quote.quoteNumber}</p>
      <p><strong>Version:</strong> ${quote.version}</p>
      <p><strong>Date:</strong> ${formatDate(quote.quoteDate)}</p>
      <p><strong>Valid Until:</strong> ${formatDate(quote.validUntil)}</p>
      <p><strong>Status:</strong> ${formatStatus(quote.status)}</p>
    </div>
  </div>

  <div class="customer-details">
    <div class="detail-box">
      <h3>Bill To</h3>
      <p><strong>${customerName}</strong></p>
      <p>${contactName}</p>
    </div>
    <div class="detail-box">
      <h3>Quote Details</h3>
      <p><strong>Currency:</strong> ${quote.currency}</p>
      <p><strong>Exchange Rate:</strong> ${quote.exchangeRate}</p>
      ${quote.isRecurring ? `<p><strong>Contract Duration:</strong> ${quote.contractDurationYears} Year(s)</p>` : ''}
    </div>
  </div>

  <div class="section">
    <h3 class="section-title">Line Items</h3>
    <table>
      <thead>
        <tr>
          <th>#</th>
          <th>Description</th>
          <th class="text-right">Qty</th>
          <th class="text-right">Unit Price</th>
          <th class="text-right">Discount</th>
          <th class="text-right">Total</th>
        </tr>
      </thead>
      <tbody>
        ${lineItems.map((item, index) => `
          <tr>
            <td>${index + 1}</td>
            <td>
              <strong>${item.productName}</strong>
              ${item.description ? `<br><small style="color: #666;">${item.description}</small>` : ''}
              ${item.productSku ? `<br><small style="color: #999;">SKU: ${item.productSku}</small>` : ''}
            </td>
            <td class="text-right">${item.quantity}</td>
            <td class="text-right">${formatCurrency(item.unitPrice)}</td>
            <td class="text-right">${item.discountPercent}%</td>
            <td class="text-right"><strong>${formatCurrency(item.lineTotal)}</strong></td>
          </tr>
        `).join('')}
      </tbody>
    </table>
  </div>

  <div class="totals">
    <div class="totals-box">
      <div class="total-row">
        <span>Subtotal:</span>
        <span><strong>${formatCurrency(quote.subtotal)}</strong></span>
      </div>
      ${quote.discountPercent > 0 ? `
      <div class="total-row">
        <span>Discount (${quote.discountPercent}%):</span>
        <span><strong>-${formatCurrency(quote.discountAmount)}</strong></span>
      </div>
      <div class="total-row">
        <span>Subtotal After Discount:</span>
        <span><strong>${formatCurrency(quote.subtotalAfterDiscount)}</strong></span>
      </div>
      ` : ''}
      <div class="total-row">
        <span>VAT (${quote.vatPercent}%):</span>
        <span><strong>${formatCurrency(quote.vatAmount)}</strong></span>
      </div>
      <div class="total-row grand-total">
        <span>TOTAL:</span>
        <span>${formatCurrency(quote.total)}</span>
      </div>
      <div class="total-row margin">
        <span>Margin (${quote.marginPercent.toFixed(1)}%):</span>
        <span>${formatCurrency(quote.marginAmount)}</span>
      </div>
    </div>
  </div>

  ${quote.paymentTerms || quote.deliveryTerms ? `
  <div class="terms">
    ${quote.paymentTerms ? `
    <div style="margin-bottom: 20px;">
      <h3>Payment Terms</h3>
      <p>${quote.paymentTerms}</p>
    </div>
    ` : ''}
    ${quote.deliveryTerms ? `
    <div>
      <h3>Delivery Terms</h3>
      <p>${quote.deliveryTerms}</p>
    </div>
    ` : ''}
  </div>
  ` : ''}

  ${quote.customerNotes ? `
  <div class="terms">
    <h3>Notes</h3>
    <p>${quote.customerNotes}</p>
  </div>
  ` : ''}

  ${attachments.length > 0 ? `
  <div class="section">
    <h3 class="section-title">📎 Attached Documents</h3>
    <div style="background: #f9fafb; padding: 20px; border-radius: 8px; border-left: 4px solid #2563eb;">
      <p style="margin-bottom: 15px; color: #666; font-size: 13px;">
        The following documents are attached to this quote:
      </p>
      <table style="background: white;">
        <thead>
          <tr>
            <th style="width: 50px;">#</th>
            <th>Document Name</th>
            <th>Type</th>
            <th style="width: 120px;">Include in PDF</th>
          </tr>
        </thead>
        <tbody>
          ${attachments.map((att, index) => `
            <tr>
              <td>${index + 1}</td>
              <td><strong>${att.documentName}</strong></td>
              <td>${formatDocumentType(att.documentType)}</td>
              <td style="text-align: center;">${att.includeInPdf ? '✓ Yes' : '✗ No'}</td>
            </tr>
          `).join('')}
        </tbody>
      </table>
      <p style="margin-top: 15px; color: #999; font-size: 12px; font-style: italic;">
        Note: To append actual PDF documents, please print each document separately or use a PDF merging tool.
      </p>
    </div>
  </div>
  ` : ''}

  <div class="footer">
    <p>This quote is valid until ${formatDate(quote.validUntil)}</p>
    <p>Generated on ${formatDate(new Date().toISOString())}</p>
  </div>
</body>
</html>
    `

    // Write HTML to iframe
    iframeDoc.open()
    iframeDoc.write(html)
    iframeDoc.close()

    // Wait for content to load, then print
    iframe.onload = () => {
      setTimeout(() => {
        iframe.contentWindow?.print()
        // Remove iframe after printing
        setTimeout(() => {
          document.body.removeChild(iframe)
        }, 100)
      }, 250)
    }
  }

  function generateContractPdf(contract: Contract, customerName: string) {
    const iframe = document.createElement('iframe')
    iframe.style.position = 'absolute'
    iframe.style.width = '0'
    iframe.style.height = '0'
    iframe.style.border = 'none'
    document.body.appendChild(iframe)

    const iframeDoc = iframe.contentWindow?.document
    if (!iframeDoc) return

    const html = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Contract ${contract.contractNumber}</title>
  <style>
    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }

    body {
      font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
      padding: 40px;
      color: #333;
      line-height: 1.8;
    }

    .header {
      text-align: center;
      margin-bottom: 40px;
      padding-bottom: 20px;
      border-bottom: 3px solid #7c3aed;
    }

    .header h1 {
      color: #7c3aed;
      font-size: 36px;
      margin-bottom: 10px;
    }

    .header h2 {
      color: #666;
      font-size: 24px;
      font-weight: 400;
    }

    .contract-info {
      background: #f9fafb;
      padding: 30px;
      border-radius: 8px;
      margin-bottom: 30px;
      border-left: 5px solid #7c3aed;
    }

    .contract-info h3 {
      color: #7c3aed;
      margin-bottom: 15px;
      font-size: 18px;
    }

    .info-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 15px;
    }

    .info-item {
      display: flex;
      justify-content: space-between;
      padding: 8px 0;
      border-bottom: 1px solid #e5e7eb;
    }

    .info-label {
      color: #666;
      font-weight: 600;
    }

    .info-value {
      color: #333;
      font-weight: 500;
    }

    .section {
      margin: 30px 0;
    }

    .section-title {
      color: #7c3aed;
      font-size: 20px;
      margin-bottom: 15px;
      padding-bottom: 10px;
      border-bottom: 2px solid #e5e7eb;
    }

    .section-content {
      padding: 15px;
      background: #f9fafb;
      border-radius: 8px;
      font-size: 14px;
      line-height: 1.8;
    }

    .signature-section {
      margin-top: 60px;
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 40px;
    }

    .signature-box {
      text-align: center;
    }

    .signature-line {
      border-top: 2px solid #333;
      margin: 60px 20px 10px;
    }

    .signature-label {
      font-size: 14px;
      color: #666;
    }

    .footer {
      margin-top: 60px;
      padding-top: 20px;
      border-top: 2px solid #e5e7eb;
      text-align: center;
      color: #999;
      font-size: 12px;
    }

    .highlight-box {
      background: #7c3aed;
      color: white;
      padding: 20px;
      border-radius: 8px;
      margin: 20px 0;
    }

    .highlight-box h3 {
      margin-bottom: 10px;
    }

    @media print {
      body {
        padding: 20px;
      }

      .section {
        page-break-inside: avoid;
      }
    }
  </style>
</head>
<body>
  <div class="header">
    <h1>SERVICE CONTRACT</h1>
    <h2>${contract.contractNumber}</h2>
  </div>

  <div class="contract-info">
    <h3>Contract Information</h3>
    <div class="info-grid">
      <div class="info-item">
        <span class="info-label">Contract Number:</span>
        <span class="info-value">${contract.contractNumber}</span>
      </div>
      <div class="info-item">
        <span class="info-label">Contract Type:</span>
        <span class="info-value">${formatContractType(contract.type)}</span>
      </div>
      <div class="info-item">
        <span class="info-label">Customer:</span>
        <span class="info-value">${customerName}</span>
      </div>
      <div class="info-item">
        <span class="info-label">Status:</span>
        <span class="info-value">${formatContractStatus(contract.status)}</span>
      </div>
      <div class="info-item">
        <span class="info-label">Start Date:</span>
        <span class="info-value">${formatDate(contract.startDate)}</span>
      </div>
      <div class="info-item">
        <span class="info-label">End Date:</span>
        <span class="info-value">${formatDate(contract.endDate)}</span>
      </div>
      <div class="info-item">
        <span class="info-label">Total Value:</span>
        <span class="info-value">${formatCurrency(contract.totalValue)} ${contract.currency}</span>
      </div>
      <div class="info-item">
        <span class="info-label">Auto Renewal:</span>
        <span class="info-value">${contract.autoRenewal ? 'Yes' : 'No'}</span>
      </div>
    </div>
  </div>

  <div class="highlight-box">
    <h3>Contract Period</h3>
    <p>This contract is effective from <strong>${formatDate(contract.startDate)}</strong> to <strong>${formatDate(contract.endDate)}</strong></p>
    ${contract.autoRenewal ? `<p style="margin-top: 10px;">This contract will automatically renew unless notice is provided ${contract.renewalNoticeDays} days prior to expiration.</p>` : ''}
  </div>

  ${contract.terms ? `
  <div class="section">
    <h3 class="section-title">Terms & Conditions</h3>
    <div class="section-content">
      ${contract.terms}
    </div>
  </div>
  ` : ''}

  ${contract.deliverables ? `
  <div class="section">
    <h3 class="section-title">Deliverables</h3>
    <div class="section-content">
      ${contract.deliverables}
    </div>
  </div>
  ` : ''}

  ${contract.sla ? `
  <div class="section">
    <h3 class="section-title">Service Level Agreement (SLA)</h3>
    <div class="section-content">
      ${contract.sla}
    </div>
  </div>
  ` : ''}

  ${contract.paymentTerms ? `
  <div class="section">
    <h3 class="section-title">Payment Terms</h3>
    <div class="section-content">
      ${contract.paymentTerms}
    </div>
  </div>
  ` : ''}

  <div class="signature-section">
    <div class="signature-box">
      <div class="signature-line"></div>
      <p class="signature-label">Company Representative</p>
      <p class="signature-label">Date: _____________</p>
    </div>
    <div class="signature-box">
      <div class="signature-line"></div>
      <p class="signature-label">Customer Representative</p>
      ${contract.signedBy ? `<p class="signature-label"><strong>${contract.signedBy}</strong></p>` : ''}
      <p class="signature-label">Date: ${contract.signedDate ? formatDate(contract.signedDate) : '_____________'}</p>
    </div>
  </div>

  <div class="footer">
    <p><strong>Contract ${contract.contractNumber}</strong></p>
    <p>Generated on ${formatDate(new Date().toISOString())}</p>
  </div>
</body>
</html>
    `

    iframeDoc.open()
    iframeDoc.write(html)
    iframeDoc.close()

    iframe.onload = () => {
      setTimeout(() => {
        iframe.contentWindow?.print()
        setTimeout(() => {
          document.body.removeChild(iframe)
        }, 100)
      }, 250)
    }
  }

  // Helper functions
  function formatCurrency(amount: number): string {
    return new Intl.NumberFormat('en-SA', {
      style: 'currency',
      currency: 'SAR',
      minimumFractionDigits: 2
    }).format(amount)
  }

  function formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString('en-GB', {
      day: '2-digit',
      month: 'short',
      year: 'numeric'
    })
  }

  function formatStatus(status: string): string {
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

  function formatContractType(type: string): string {
    const typeMap: Record<string, string> = {
      sales: 'Sales Contract',
      maintenance: 'Maintenance Contract',
      service: 'Service Contract',
      project: 'Project Contract',
      subscription: 'Subscription Contract'
    }
    return typeMap[type] || type
  }

  function formatContractStatus(status: string): string {
    const statusMap: Record<string, string> = {
      draft: 'Draft',
      pending_approval: 'Pending Approval',
      active: 'Active',
      expired: 'Expired',
      terminated: 'Terminated',
      renewed: 'Renewed'
    }
    return statusMap[status] || status
  }

  function formatDocumentType(type: string): string {
    const typeMap: Record<string, string> = {
      terms: 'Terms & Conditions',
      delivery: 'Delivery Policy',
      technical: 'Technical Specifications',
      warranty: 'Warranty Information',
      sla: 'Service Level Agreement',
      other: 'Other'
    }
    return typeMap[type] || type
  }

  return {
    generateQuotePdf,
    generateContractPdf
  }
}
