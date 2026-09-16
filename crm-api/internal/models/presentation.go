package models

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// JSON responses expose the names/arrays used by the frontend while retaining
// canonical relationship IDs and nested records for clients that need them.
func withFields(value any, fields map[string]any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var object map[string]any
	if err := json.Unmarshal(data, &object); err != nil {
		return nil, err
	}
	for key, value := range fields {
		object[key] = value
	}
	return json.Marshal(object)
}
func rows[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}
func userName(user *User) string {
	if user == nil {
		return ""
	}
	return user.FullName()
}
func customerName(customer *Customer) string {
	if customer == nil {
		return ""
	}
	return customer.CompanyName
}
func (o Opportunity) MarshalJSON() ([]byte, error) {
	type plain Opportunity
	ids := []string{}
	for _, quote := range o.Quotes {
		ids = append(ids, quote.ID.String())
	}
	return withFields(plain(o), map[string]any{"customerName": customerName(o.Customer), "salesExecutiveName": userName(o.SalesExecutive), "preSalesName": userName(o.PreSales), "serviceTypes": rows(o.ServiceTypes), "quoteIds": ids})
}
func (p Product) MarshalJSON() ([]byte, error) {
	type plain Product
	mfr, category := "", ""
	if p.Manufacturer != nil {
		mfr = p.Manufacturer.Name
	}
	if p.Category != nil {
		category = p.Category.Name
	}
	return withFields(plain(p), map[string]any{"manufacturerName": mfr, "categoryName": category, "vendorEntries": rows(p.VendorEntries), "priceHistory": rows(p.PriceHistory), "documents": rows(p.Documents)})
}
func (m Manufacturer) MarshalJSON() ([]byte, error) {
	type plain Manufacturer
	return withFields(plain(m), map[string]any{"categories": rows(m.Categories)})
}
func (q Quote) MarshalJSON() ([]byte, error) {
	type plain Quote
	return withFields(plain(q), map[string]any{"approvedBy": userName(q.ApprovedBy), "customerName": customerName(q.Customer), "lineItems": rows(q.LineItems)})
}
func (c Contract) MarshalJSON() ([]byte, error) {
	type plain Contract
	return withFields(plain(c), map[string]any{"customerName": customerName(c.Customer)})
}
func (p Project) MarshalJSON() ([]byte, error) {
	type plain Project
	number := ""
	items := []QuoteLineItem{}
	if p.Quote != nil {
		number = p.Quote.QuoteNumber
		items = rows(p.Quote.LineItems)
	}
	return withFields(plain(p), map[string]any{"customerName": customerName(p.Customer), "quoteNumber": number, "projectManager": userName(p.ProjectManager), "lineItems": items})
}
func (t Team) MarshalJSON() ([]byte, error) {
	type plain Team
	members := []map[string]any{}
	for _, u := range t.Members {
		members = append(members, map[string]any{"id": u.ID, "name": u.FullName(), "email": u.Email, "phone": u.Phone, "role": u.Role, "department": u.Department, "isLeader": t.LeaderID != nil && *t.LeaderID == u.ID})
	}
	return withFields(plain(t), map[string]any{"leaderName": userName(t.Leader), "members": members})
}
func (p PriceBook) MarshalJSON() ([]byte, error) {
	type plain PriceBook
	return withFields(plain(p), map[string]any{"customerName": customerName(p.Customer), "entries": rows(p.Entries)})
}
func (e PriceBookEntry) MarshalJSON() ([]byte, error) {
	type entry PriceBookEntry
	sku, name := "", ""
	if e.Product != nil {
		sku = e.Product.SKU
		name = e.Product.Name
	}
	id := e.ProductID
	if e.Service != nil {
		sku = e.Service.SKU
		name = e.Service.Name
		id = e.ServiceID
	}
	if e.RecurringService != nil {
		sku = "REC-" + e.RecurringService.ID.String()[:8]
		name = e.RecurringService.Name
		id = e.RecurringServiceID
	}
	return withFields(entry(e), map[string]any{"productId": id, "productSku": sku, "productName": name})
}
func (e ExchangeRate) MarshalJSON() ([]byte, error) {
	type plain ExchangeRate
	return withFields(plain(e), map[string]any{"history": rows(e.History)})
}
func (w WarehouseStock) MarshalJSON() ([]byte, error) {
	type plain WarehouseStock
	sku, name := "", ""
	if w.Product != nil {
		sku = w.Product.SKU
		name = w.Product.Name
	}
	return withFields(plain(w), map[string]any{"productSku": sku, "productName": name})
}
func (r StockReservation) MarshalJSON() ([]byte, error) {
	type plain StockReservation
	sku, name := "", ""
	if r.Product != nil {
		sku = r.Product.SKU
		name = r.Product.Name
	}
	return withFields(plain(r), map[string]any{"productSku": sku, "productName": name, "reservedBy": userName(r.ReservedBy)})
}
func (m InventoryMovement) MarshalJSON() ([]byte, error) {
	type plain InventoryMovement
	sku, name := "", ""
	if m.Product != nil {
		sku = m.Product.SKU
		name = m.Product.Name
	}
	return withFields(plain(m), map[string]any{"productSku": sku, "productName": name, "performedBy": userName(m.PerformedBy)})
}

func (d Document) MarshalJSON() ([]byte, error) {
	type plain Document
	return withFields(plain(d), map[string]any{"tags": rows(d.Tags), "links": rows(d.Links), "linkedEntities": documentLinks(d.Links), "fileSizeBytes": d.FileSize, "fileSize": fileSize(d.FileSize), "mimeType": d.FileType, "fileType": strings.TrimPrefix(filepath.Ext(d.FileName), "."), "uploadedBy": userName(d.UploadedBy)})
}
func (p PurchaseOrder) MarshalJSON() ([]byte, error) {
	type plain PurchaseOrder
	return withFields(plain(p), map[string]any{"items": rows(p.Items)})
}
func (s SupplierQuote) MarshalJSON() ([]byte, error) {
	type plain SupplierQuote
	return withFields(plain(s), map[string]any{"items": rows(s.Items)})
}
func (g GoodsReceipt) MarshalJSON() ([]byte, error) {
	type plain GoodsReceipt
	return withFields(plain(g), map[string]any{"items": rows(g.Items)})
}
func (p ProductDocument) MarshalJSON() ([]byte, error) {
	type plain ProductDocument
	return withFields(plain(p), map[string]any{"fileSizeBytes": p.FileSize, "fileSize": fileSize(p.FileSize), "uploadedBy": userName(p.UploadedBy), "uploadedAt": p.CreatedAt})
}
func (s SupplierItemCatalog) MarshalJSON() ([]byte, error) {
	type plain SupplierItemCatalog
	name, sku, mfr, supplier := "", "", "", ""
	if s.Product != nil {
		name = s.Product.Name
		sku = s.Product.SKU
		if s.Product.Manufacturer != nil {
			mfr = s.Product.Manufacturer.Name
		}
	}
	if s.Supplier != nil {
		supplier = s.Supplier.Name
	}
	return withFields(plain(s), map[string]any{"productName": name, "productSku": sku, "manufacturerName": mfr, "supplierName": supplier})
}

func fileSize(size int64) string {
	if size >= 1<<20 {
		return fmt.Sprintf("%.1f MB", float64(size)/(1<<20))
	}
	if size >= 1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	return fmt.Sprintf("%d B", size)
}
func documentLinks(links []DocumentLink) []map[string]any {
	result := []map[string]any{}
	for _, link := range links {
		result = append(result, map[string]any{"id": link.EntityID, "type": link.EntityType, "name": link.EntityName})
	}
	return result
}
