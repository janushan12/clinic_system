package services

import (
	"bytes"
	"fmt"
	"time"

	"clinic-notes/models"

	"github.com/jung-kurt/gofpdf"
)

var (
	cNavy   = [3]int{15, 31, 53}
	cBlue   = [3]int{26, 111, 219}
	cGray   = [3]int{74, 85, 104}
	cGrayLt = [3]int{240, 244, 248}
	cWhite  = [3]int{255, 255, 255}
	cGreen  = [3]int{10, 138, 92}
	cPurple = [3]int{109, 40, 217}
	cTeal   = [3]int{3, 105, 161}
	cBorder = [3]int{221, 227, 236}
)

func sf(pdf *gofpdf.Fpdf, c [3]int) { pdf.SetFillColor(c[0], c[1], c[2]) }
func sd(pdf *gofpdf.Fpdf, c [3]int) { pdf.SetDrawColor(c[0], c[1], c[2]) }
func st(pdf *gofpdf.Fpdf, c [3]int) { pdf.SetTextColor(c[0], c[1], c[2]) }

func pWidth(pdf *gofpdf.Fpdf) float64 { w, _ := pdf.GetPageSize(); return w }

func GenerateVisitPDF(
	patient *models.Patient,
	doctor *models.Doctor,
	visit *models.Visit,
	prescriptions []models.Prescription,
	labOrders []models.LabOrder,
	notes []models.ClinicNote,
	bill *models.Bill,
	lineItems []models.BillLineItem,
) ([]byte, error) {

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()

	cW := pWidth(pdf) - 30 // content width

	// ── Header ───────────────────────────────────────────────
	sf(pdf, cNavy)
	pdf.Rect(0, 0, pWidth(pdf), 28, "F")
	pdf.SetFont("Helvetica", "B", 16)
	st(pdf, cWhite)
	pdf.SetXY(15, 8)
	pdf.CellFormat(120, 8, "ABC Health Clinic", "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	st(pdf, [3]int{176, 196, 222})
	pdf.SetXY(15, 17)
	pdf.CellFormat(120, 6, "AI-Powered Unified Notes & Billing System", "", 0, "L", false, 0, "")
	pdf.SetXY(pWidth(pdf)-80, 11)
	pdf.CellFormat(65, 6, time.Now().Format("02 Jan 2006  15:04"), "", 0, "R", false, 0, "")

	// ── Info boxes ────────────────────────────────────────────
	bw := cW / 3
	infoBox(pdf, 15, 33, bw-2, "PATIENT", patient.FullName, patient.Phone)
	infoBox(pdf, 15+bw, 33, bw-2, "DOCTOR", doctor.FullName, doctor.Speciality)
	infoBox(pdf, 15+bw*2, 33, bw-2, "VISIT", visit.ID[:8]+"...", visit.VisitDate.Format("02 Jan 2006"))
	pdf.SetY(60)

	// ── Clinical Notes ────────────────────────────────────────
	if len(notes) > 0 {
		sectionHdr(pdf, cW, "CLINICAL NOTES", cGreen)
		for _, n := range notes {
			noteRow(pdf, cW, n)
		}
		pdf.Ln(3)
	}

	// ── Prescription ──────────────────────────────────────────
	if len(prescriptions) > 0 {
		sectionHdr(pdf, cW, "PRESCRIPTION", cPurple)
		tblHdr(pdf, cW, []string{"Drug Name", "Dosage", "Frequency", "Duration", "Qty"},
			[]float64{0.35, 0.15, 0.25, 0.15, 0.10})
		for i, p := range prescriptions {
			drugRow(pdf, cW, p, i)
		}
		pdf.Ln(3)
	}

	// ── Lab Tests ─────────────────────────────────────────────
	if len(labOrders) > 0 {
		sectionHdr(pdf, cW, "LAB INVESTIGATIONS", cTeal)
		tblHdr(pdf, cW, []string{"Test Name", "Notes", "Price"},
			[]float64{0.55, 0.30, 0.15})
		for i, lo := range labOrders {
			labRow(pdf, cW, lo, i)
		}
		pdf.Ln(3)
	}

	// ── Bill ──────────────────────────────────────────────────
	sectionHdr(pdf, cW, "PATIENT BILL", [3]int{120, 70, 10})
	tblHdr(pdf, cW, []string{"Description", "Category", "Qty", "Unit Price", "Total"},
		[]float64{0.35, 0.20, 0.10, 0.17, 0.18})

	billRow(pdf, cW, "Consultation Fee", "consultation", 1, bill.ConsultationFee, bill.ConsultationFee, 0)
	for i, item := range lineItems {
		if item.Category == "consultation" {
			continue
		}
		billRow(pdf, cW, item.Description, item.Category, item.Quantity, item.UnitPrice, item.LineTotal, i+1)
	}

	pdf.Ln(3)
	grandTotal(pdf, cW, bill)

	// ── Footer ────────────────────────────────────────────────
	pdf.SetY(-18)
	sd(pdf, cBorder)
	pdf.Line(15, pdf.GetY(), pWidth(pdf)-15, pdf.GetY())
	pdf.Ln(2)
	pdf.SetFont("Helvetica", "I", 8)
	st(pdf, cGray)
	pdf.SetX(15)
	pdf.CellFormat(cW/2, 5, "ABC Health Clinic — Confidential Patient Record", "", 0, "L", false, 0, "")
	pdf.CellFormat(cW/2, 5, "Generated: "+time.Now().Format("02 Jan 2006 15:04"), "", 0, "R", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("PDF generation failed: %w", err)
	}
	return buf.Bytes(), nil
}

func infoBox(pdf *gofpdf.Fpdf, x, y, w float64, label, line1, line2 string) {
	sf(pdf, cGrayLt)
	sd(pdf, cBorder)
	pdf.RoundedRect(x, y, w, 22, 2, "1234", "FD")
	pdf.SetFont("Helvetica", "B", 7)
	st(pdf, [3]int{138, 150, 167})
	pdf.SetXY(x+3, y+3)
	pdf.CellFormat(w-6, 4, label, "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 9)
	st(pdf, cNavy)
	pdf.SetXY(x+3, y+8)
	pdf.CellFormat(w-6, 5, line1, "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8)
	st(pdf, cGray)
	pdf.SetXY(x+3, y+14)
	pdf.CellFormat(w-6, 5, line2, "", 0, "L", false, 0, "")
}

func sectionHdr(pdf *gofpdf.Fpdf, cW float64, title string, c [3]int) {
	sf(pdf, c)
	pdf.Rect(15, pdf.GetY(), cW, 7, "F")
	pdf.SetFont("Helvetica", "B", 9)
	st(pdf, cWhite)
	pdf.SetX(18)
	pdf.CellFormat(cW-6, 7, title, "", 1, "L", false, 0, "")
}

func tblHdr(pdf *gofpdf.Fpdf, cW float64, cols []string, ratios []float64) {
	sf(pdf, [3]int{220, 228, 240})
	sd(pdf, cBorder)
	y := pdf.GetY()
	x := 15.0
	for i, col := range cols {
		w := cW * ratios[i]
		pdf.Rect(x, y, w, 6, "FD")
		pdf.SetFont("Helvetica", "B", 8)
		st(pdf, cGray)
		pdf.SetXY(x+2, y+1)
		pdf.CellFormat(w-4, 4, col, "", 0, "L", false, 0, "")
		x += w
	}
	pdf.SetY(y + 6)
}

func rowCells(pdf *gofpdf.Fpdf, cW float64, vals []string, ratios []float64, bolds []bool, aligns []string, stripe int) {
	bg := cWhite
	if stripe%2 == 0 {
		bg = cGrayLt
	}
	sf(pdf, bg)
	sd(pdf, cBorder)
	y := pdf.GetY()
	x := 15.0
	h := 7.0
	for i, val := range vals {
		w := cW * ratios[i]
		pdf.Rect(x, y, w, h, "FD")
		if bolds[i] {
			pdf.SetFont("Helvetica", "B", 8)
			st(pdf, cNavy)
		} else {
			pdf.SetFont("Helvetica", "", 8)
			st(pdf, cGray)
		}
		pdf.SetXY(x+2, y+1)
		pdf.CellFormat(w-4, h-2, val, "", 0, aligns[i], false, 0, "")
		x += w
	}
	pdf.SetY(y + h)
}

func drugRow(pdf *gofpdf.Fpdf, cW float64, p models.Prescription, i int) {
	rowCells(pdf, cW,
		[]string{p.DrugName, p.Dosage, p.Frequency, p.Duration, fmt.Sprintf("%d", p.Quantity)},
		[]float64{0.35, 0.15, 0.25, 0.15, 0.10},
		[]bool{true, false, false, false, false},
		[]string{"L", "L", "L", "L", "C"}, i)
}

func labRow(pdf *gofpdf.Fpdf, cW float64, lo models.LabOrder, i int) {
	rowCells(pdf, cW,
		[]string{lo.TestName, lo.Notes, fmt.Sprintf("$%.2f", lo.UnitPrice)},
		[]float64{0.55, 0.30, 0.15},
		[]bool{true, false, false},
		[]string{"L", "L", "R"}, i)
}

func noteRow(pdf *gofpdf.Fpdf, cW float64, n models.ClinicNote) {
	catC := map[string][3]int{
		"observation": cGreen, "diagnosis": cTeal,
		"history": {180, 100, 10}, "allergy": {185, 28, 28}, "other": cGray,
	}
	c := catC[n.Category]
	if c == ([3]int{}) {
		c = cGray
	}
	y := pdf.GetY()
	sf(pdf, cGrayLt)
	sd(pdf, cBorder)
	pdf.Rect(15, y, cW, 8, "FD")
	sf(pdf, c)
	pdf.Rect(17, y+1.5, 22, 5, "F")
	pdf.SetFont("Helvetica", "B", 7)
	st(pdf, cWhite)
	pdf.SetXY(17, y+2.5)
	pdf.CellFormat(22, 3, n.Category, "", 0, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 8)
	st(pdf, cNavy)
	pdf.SetXY(42, y+2)
	pdf.CellFormat(cW-30, 5, n.NoteText, "", 0, "L", false, 0, "")
	pdf.SetY(y + 8)
}

func billRow(pdf *gofpdf.Fpdf, cW float64, desc, cat string, qty int, unit, total float64, i int) {
	rowCells(pdf, cW,
		[]string{desc, cat, fmt.Sprintf("%d", qty), fmt.Sprintf("$%.2f", unit), fmt.Sprintf("$%.2f", total)},
		[]float64{0.35, 0.20, 0.10, 0.17, 0.18},
		[]bool{true, false, false, false, true},
		[]string{"L", "L", "C", "R", "R"}, i)
}

func grandTotal(pdf *gofpdf.Fpdf, cW float64, bill *models.Bill) {
	right := pWidth(pdf) - 15
	summaries := []struct{ l, v string }{
		{"Consultation:", fmt.Sprintf("$%.2f", bill.ConsultationFee)},
		{"Drugs:", fmt.Sprintf("$%.2f", bill.DrugTotal)},
		{"Lab Tests:", fmt.Sprintf("$%.2f", bill.LabTotal)},
	}
	for _, s := range summaries {
		pdf.SetFont("Helvetica", "", 9)
		st(pdf, cGray)
		pdf.SetX(right - 60)
		pdf.CellFormat(35, 5, s.l, "", 0, "R", false, 0, "")
		pdf.CellFormat(25, 5, s.v, "", 1, "R", false, 0, "")
	}
	pdf.Ln(1)
	sd(pdf, cBorder)
	pdf.Line(right-65, pdf.GetY(), right, pdf.GetY())
	pdf.Ln(2)

	// Grand total box
	gtY := pdf.GetY()
	sf(pdf, cNavy)
	pdf.Rect(right-68, gtY, 53, 10, "F")
	pdf.SetFont("Helvetica", "B", 9)
	st(pdf, cWhite)
	pdf.SetXY(right-66, gtY+2)
	pdf.CellFormat(25, 6, "GRAND TOTAL", "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 11)
	st(pdf, [3]int{100, 190, 255})
	pdf.CellFormat(24, 6, fmt.Sprintf("$%.2f", bill.GrandTotal), "", 1, "R", false, 0, "")

	// Status
	statusC := map[string][3]int{
		"pending": {196, 125, 14}, "paid": cGreen, "cancelled": {185, 28, 28},
	}
	sc := statusC[bill.Status]
	if sc == ([3]int{}) {
		sc = statusC["pending"]
	}
	sf(pdf, sc)
	pdf.Rect(right-68, gtY+11, 30, 5, "F")
	pdf.SetFont("Helvetica", "B", 7)
	st(pdf, cWhite)
	pdf.SetXY(right-68, gtY+12.5)
	pdf.CellFormat(30, 3, "STATUS: "+bill.Status, "", 0, "C", false, 0, "")
}