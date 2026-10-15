package handlers

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"g4s-crm/api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

const templateRows = 500

var templateHeaders = []string{"Part Number", "Description", "Unit Cost", "Vendor Name", "Manufacturer", "Category", "Currency", "Rate to SAR", "Lead Time (Days)"}

var templateInstructions = [][2]string{
	{"Start here", "Fill the Products sheet, one product per row. Keep the headers."},
	{"Required", "Part Number, Description and Unit Cost."},
	{"Manufacturer", "Pick from the drop-down: it lists the manufacturers in the CRM today, so each row links to the right record."},
	{"New manufacturer", "If it is not in the list, type the name anyway. Excel warns; continue. The import screen lists new names and offers to add them (or ask an administrator)."},
	{"Category", "Pick from the drop-down after choosing the manufacturer: it shows that manufacturer's categories. A new category can be typed the same way."},
	{"Vendor Name", "The company the price came from, e.g. Ingram Micro. Blank uses the default chosen during import."},
	{"Currency", "SAR, USD, EUR, GBP, AED or CNY. Blank uses the currency chosen during import."},
	{"Rate to SAR", "SAR uses 1. For another currency enter its rate to SAR, e.g. USD 3.75. Blank uses the rate shown during import for the default currency."},
	{"Unit Cost", "Vendor price in the row currency, before freight, customs, clearance and margin."},
	{"Lead Time (Days)", "Optional whole days. Blank keeps an existing product's lead time."},
	{"Part numbers", "Cells are text, so leading zeros are kept. One row per part number; duplicates are skipped."},
	{"Import", "Products → Import from vendor file → choose this workbook → review → Import."},
	{"Prices", "Freight, customs, clearance and margin are set once for the file during import."},
	{"Inventory", "This adds catalog products and prices only. Stock comes from purchasing and receiving."},
	{"Lists", fmt.Sprintf("Drop-downs reflect the CRM on the day this file was downloaded. Download a fresh copy after adding manufacturers. %d rows are formatted; the app accepts up to 2,000 per import.", templateRows)},
}

// manufacturerLabels names each manufacturer the way the importer resolves it:
// by name, or by its unique code when the name alone would be ambiguous.
func manufacturerLabels(manufacturers []models.Manufacturer) map[string]string {
	names, codes := map[string]int{}, map[string]bool{}
	for _, m := range manufacturers {
		names[strings.ToLower(strings.TrimSpace(m.Name))]++
		codes[strings.ToLower(strings.TrimSpace(m.Code))] = true
	}
	labels := map[string]string{}
	for _, m := range manufacturers {
		key := strings.ToLower(strings.TrimSpace(m.Name))
		if names[key] == 1 && (!codes[key] || strings.EqualFold(m.Code, m.Name)) {
			labels[m.ID.String()] = strings.TrimSpace(m.Name)
		} else {
			labels[m.ID.String()] = strings.TrimSpace(m.Code)
		}
	}
	return labels
}

// ImportTemplate builds the product import workbook from the live catalog, so
// the manufacturer and category drop-downs hold exactly the names the import
// will match.
func (h *ProductHandler) ImportTemplate(c *gin.Context) {
	var manufacturers []models.Manufacturer
	if err := h.db.Preload("Categories").Where("is_active").Order("name").Find(&manufacturers).Error; err != nil {
		apiError(c, err)
		return
	}
	file, err := buildImportTemplate(manufacturers)
	if err != nil {
		apiError(c, err)
		return
	}
	defer file.Close()
	name := "g4s-product-import-" + time.Now().Format("2006-01-02") + ".xlsx"
	c.Header("Content-Disposition", `attachment; filename="`+name+`"`)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err := file.Write(c.Writer); err != nil {
		_ = c.Error(err)
	}
}

func buildImportTemplate(manufacturers []models.Manufacturer) (*excelize.File, error) {
	f := excelize.NewFile()
	const products, instructions, lists = "Products", "Instructions", "Lists"
	if err := f.SetSheetName("Sheet1", products); err != nil {
		return nil, err
	}
	for _, sheet := range []string{instructions, lists} {
		if _, err := f.NewSheet(sheet); err != nil {
			return nil, err
		}
	}

	// Lists: one row per manufacturer/category pair, grouped by manufacturer so
	// the category drop-down can take the manufacturer's contiguous block.
	labels := manufacturerLabels(manufacturers)
	type pair struct{ manufacturer, category string }
	pairs := []pair{}
	unique := []string{}
	for _, m := range manufacturers {
		label := labels[m.ID.String()]
		unique = append(unique, label)
		cats := []string{}
		for _, cat := range m.Categories {
			cats = append(cats, strings.TrimSpace(cat.Name))
		}
		sort.Strings(cats)
		if len(cats) == 0 {
			cats = []string{""}
		}
		for _, cat := range cats {
			pairs = append(pairs, pair{label, cat})
		}
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		return strings.ToLower(pairs[i].manufacturer) < strings.ToLower(pairs[j].manufacturer)
	})
	sort.Slice(unique, func(i, j int) bool { return strings.ToLower(unique[i]) < strings.ToLower(unique[j]) })
	if err := f.SetSheetRow(lists, "A1", &[]any{"Manufacturer", "Category", "", "Manufacturers", "", "Currencies"}); err != nil {
		return nil, err
	}
	for i, p := range pairs {
		if err := f.SetSheetRow(lists, fmt.Sprintf("A%d", i+2), &[]any{p.manufacturer, p.category}); err != nil {
			return nil, err
		}
	}
	for i, label := range unique {
		if err := f.SetCellStr(lists, fmt.Sprintf("D%d", i+2), label); err != nil {
			return nil, err
		}
	}
	currencies := []string{"SAR", "USD", "EUR", "GBP", "AED", "CNY"}
	for i, cur := range currencies {
		if err := f.SetCellStr(lists, fmt.Sprintf("F%d", i+2), cur); err != nil {
			return nil, err
		}
	}
	if err := f.SetSheetVisible(lists, false); err != nil {
		return nil, err
	}

	// Products: header, text part numbers, numeric checks and drop-downs.
	header, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"1F2937"}, Pattern: 1}, Alignment: &excelize.Alignment{Vertical: "center", WrapText: true}})
	text, _ := f.NewStyle(&excelize.Style{NumFmt: 49})
	input, _ := f.NewStyle(&excelize.Style{Fill: excelize.Fill{Type: "pattern", Color: []string{"FFF7E0"}, Pattern: 1}})
	if err := f.SetSheetRow(products, "A1", &templateHeaders); err != nil {
		return nil, err
	}
	last := templateRows + 1
	steps := []func() error{
		func() error { return f.SetCellStyle(products, "A1", "I1", header) },
		func() error { return f.SetCellStyle(products, "A2", fmt.Sprintf("I%d", last), input) },
		func() error { return f.SetCellStyle(products, "A2", fmt.Sprintf("A%d", last), text) },
		func() error { return f.SetRowHeight(products, 1, 30) },
		func() error {
			return f.SetPanes(products, &excelize.Panes{Freeze: true, Split: false, XSplit: 1, YSplit: 1, TopLeftCell: "B2", ActivePane: "bottomRight"})
		},
	}
	for col, width := range map[string]float64{"A": 22, "B": 48, "C": 12, "D": 26, "E": 28, "F": 28, "G": 11, "H": 12, "I": 16} {
		col, width := col, width
		steps = append(steps, func() error { return f.SetColWidth(products, col, col, width) })
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return nil, err
		}
	}

	validations := []*excelize.DataValidation{}
	number := func(cols, title string, decimal bool) *excelize.DataValidation {
		dv := excelize.NewDataValidation(true)
		dv.Sqref = cols
		kind := excelize.DataValidationTypeWhole
		if decimal {
			kind = excelize.DataValidationTypeDecimal
		}
		_ = dv.SetRange(0, 1000000000, kind, excelize.DataValidationOperatorBetween)
		dv.SetError(excelize.DataValidationErrorStyleStop, title, "Enter a number of 0 or more.")
		return dv
	}
	validations = append(validations,
		number(fmt.Sprintf("C2:C%d", last), "Unit Cost", true),
		number(fmt.Sprintf("H2:H%d", last), "Rate to SAR", true),
		number(fmt.Sprintf("I2:I%d", last), "Lead Time (Days)", false),
	)

	currency := excelize.NewDataValidation(true)
	currency.Sqref = fmt.Sprintf("G2:G%d", last)
	currency.SetSqrefDropList(fmt.Sprintf("Lists!$F$2:$F$%d", len(currencies)+1))
	currency.SetError(excelize.DataValidationErrorStyleStop, "Currency", "Choose SAR, USD, EUR, GBP, AED or CNY.")
	validations = append(validations, currency)

	if len(unique) > 0 {
		mfr := excelize.NewDataValidation(true)
		mfr.Sqref = fmt.Sprintf("E2:E%d", last)
		mfr.SetSqrefDropList(fmt.Sprintf("Lists!$D$2:$D$%d", len(unique)+1))
		// Warning, not stop: a new manufacturer can be typed and is offered for
		// creation during import.
		mfr.SetError(excelize.DataValidationErrorStyleWarning, "New manufacturer?", "This name is not in the CRM's manufacturer list. Choose Yes to keep it; the import screen will offer to add it.")
		validations = append(validations, mfr)

		// The chosen manufacturer's categories: its block in Lists columns A:B.
		cat := excelize.NewDataValidation(true)
		cat.Sqref = fmt.Sprintf("F2:F%d", last)
		cat.SetSqrefDropList("OFFSET(Lists!$B$1,MATCH($E2,Lists!$A:$A,0)-1,0,MAX(1,COUNTIF(Lists!$A:$A,$E2)),1)")
		cat.SetError(excelize.DataValidationErrorStyleWarning, "New category?", "This category is not listed for the chosen manufacturer. Choose Yes to keep it; the import screen will offer to add it.")
		validations = append(validations, cat)
	}
	for _, dv := range validations {
		if err := f.AddDataValidation(products, dv); err != nil {
			return nil, err
		}
	}

	// Instructions.
	bold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Alignment: &excelize.Alignment{Vertical: "top"}})
	wrap, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{WrapText: true, Vertical: "top"}})
	title, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 14}})
	_ = f.SetCellStr(instructions, "A1", "G4S product import")
	_ = f.SetCellStyle(instructions, "A1", "A1", title)
	for i, line := range templateInstructions {
		row := i + 3
		_ = f.SetCellStr(instructions, fmt.Sprintf("A%d", row), line[0])
		_ = f.SetCellStr(instructions, fmt.Sprintf("B%d", row), line[1])
	}
	end := len(templateInstructions) + 2
	_ = f.SetCellStyle(instructions, "A3", fmt.Sprintf("A%d", end), bold)
	_ = f.SetCellStyle(instructions, "B3", fmt.Sprintf("B%d", end), wrap)
	_ = f.SetColWidth(instructions, "A", "A", 20)
	_ = f.SetColWidth(instructions, "B", "B", 100)

	f.SetActiveSheet(0)
	return f, nil
}
