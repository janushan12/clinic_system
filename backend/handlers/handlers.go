package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"clinic-notes/config"
	"clinic-notes/models"
	"clinic-notes/services"

	"github.com/gin-gonic/gin"
)

// ── POST /api/parse ───────────────────────────────────────────
// Main endpoint: receives raw text, calls Claude, saves everything

func ParseAndSave(c *gin.Context) {
	var req models.ParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}
	if req.RawInput == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "raw_input is required"})
		return
	}

	// If no visit_id provided, create a new visit
	if req.VisitID == "" {
		if req.PatientID == "" || req.DoctorID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "patient_id and doctor_id are required when visit_id is not provided"})
			return
		}
		var visitID string
		err := config.DB.QueryRow(`
			INSERT INTO visits (patient_id, doctor_id, raw_input, status)
			VALUES ($1, $2, $3, 'draft')
			RETURNING id`,
			req.PatientID, req.DoctorID, req.RawInput,
		).Scan(&visitID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create visit: " + err.Error()})
			return
		}
		req.VisitID = visitID
	} else {
		// Update raw_input on existing visit
		_, err := config.DB.Exec(
			`UPDATE visits SET raw_input = $1 WHERE id = $2`,
			req.RawInput, req.VisitID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update visit"})
			return
		}
	}

	// Call Claude AI parser
	parsed, err := services.ParseWithClaude(req.RawInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI parsing failed: " + err.Error()})
		return
	}

	// Save everything and generate bill
	result, err := services.SaveParsedVisit(req.VisitID, parsed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save parsed data: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ── GET /api/visits/:id ───────────────────────────────────────
// Fetch a full visit with all parsed data

func GetVisit(c *gin.Context) {
	visitID := c.Param("id")

	var visit models.Visit
	err := config.DB.QueryRow(`
		SELECT id, patient_id, doctor_id, visit_date, raw_input, ai_parsed_at, status, created_at
		FROM visits WHERE id = $1`, visitID,
	).Scan(&visit.ID, &visit.PatientID, &visit.DoctorID, &visit.VisitDate,
		&visit.RawInput, &visit.AIParsedAt, &visit.Status, &visit.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Visit not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch prescriptions
	prescRows, _ := config.DB.Query(
		`SELECT id, visit_id, drug_name, dosage, frequency, duration, quantity, unit_price, instructions, created_at
		 FROM prescriptions WHERE visit_id = $1`, visitID)
	defer prescRows.Close()
	var prescriptions []models.Prescription
	for prescRows.Next() {
		var p models.Prescription
		prescRows.Scan(&p.ID, &p.VisitID, &p.DrugName, &p.Dosage, &p.Frequency,
			&p.Duration, &p.Quantity, &p.UnitPrice, &p.Instructions, &p.CreatedAt)
		prescriptions = append(prescriptions, p)
	}

	// Fetch lab orders
	labRows, _ := config.DB.Query(
		`SELECT id, visit_id, test_name, unit_price, notes, created_at
		 FROM lab_orders WHERE visit_id = $1`, visitID)
	defer labRows.Close()
	var labOrders []models.LabOrder
	for labRows.Next() {
		var lo models.LabOrder
		labRows.Scan(&lo.ID, &lo.VisitID, &lo.TestName, &lo.UnitPrice, &lo.Notes, &lo.CreatedAt)
		labOrders = append(labOrders, lo)
	}

	// Fetch clinic notes
	noteRows, _ := config.DB.Query(
		`SELECT id, visit_id, note_text, category, created_at
		 FROM clinic_notes WHERE visit_id = $1`, visitID)
	defer noteRows.Close()
	var clinicNotes []models.ClinicNote
	for noteRows.Next() {
		var n models.ClinicNote
		noteRows.Scan(&n.ID, &n.VisitID, &n.NoteText, &n.Category, &n.CreatedAt)
		clinicNotes = append(clinicNotes, n)
	}

	c.JSON(http.StatusOK, gin.H{
		"visit":         visit,
		"prescriptions": prescriptions,
		"lab_orders":    labOrders,
		"clinic_notes":  clinicNotes,
	})
}

// ── GET /api/visits/:id/bill ──────────────────────────────────
// Fetch bill with line items for a visit

func GetBill(c *gin.Context) {
	visitID := c.Param("id")
	bill, items, err := services.GetBillWithLineItems(visitID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bill": bill, "line_items": items})
}

// ── GET /api/patients ─────────────────────────────────────────
func ListPatients(c *gin.Context) {
	rows, err := config.DB.Query(
		`SELECT id, full_name, dob, gender, phone, email FROM patients ORDER BY full_name`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var patients []models.Patient
	for rows.Next() {
		var p models.Patient
		rows.Scan(&p.ID, &p.FullName, &p.DOB, &p.Gender, &p.Phone, &p.Email)
		patients = append(patients, p)
	}
	c.JSON(http.StatusOK, patients)
}

// ── POST /api/patients ────────────────────────────────────────
func CreatePatient(c *gin.Context) {
	var p models.Patient
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := config.DB.QueryRow(`
		INSERT INTO patients (full_name, dob, gender, phone, email, address)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, created_at`,
		p.FullName, p.DOB, p.Gender, p.Phone, p.Email, p.Address,
	).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// ── GET /api/doctors ──────────────────────────────────────────
func ListDoctors(c *gin.Context) {
	rows, err := config.DB.Query(
		`SELECT id, full_name, speciality, license_no FROM doctors ORDER BY full_name`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var doctors []models.Doctor
	for rows.Next() {
		var d models.Doctor
		rows.Scan(&d.ID, &d.FullName, &d.Speciality, &d.LicenseNo)
		doctors = append(doctors, d)
	}
	c.JSON(http.StatusOK, doctors)
}

// ── GET /api/visits/:id/bill/pay ─────────────────────────────
// Mark a bill as paid

func MarkBillPaid(c *gin.Context) {
	visitID := c.Param("id")
	now := time.Now()
	res, err := config.DB.Exec(`
		UPDATE bills SET status = 'paid', paid_at = $1
		WHERE visit_id = $2 AND status = 'pending'`, now, visitID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bill not found or already paid"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Bill marked as paid", "paid_at": now})
}

// ── GET /api/visits/:id/pdf ───────────────────────────────────
// Generate and download a PDF report for a visit

func DownloadPDF(c *gin.Context) {
	visitID := c.Param("id")

	// Fetch visit
	var visit models.Visit
	err := config.DB.QueryRow(
		`SELECT id, patient_id, doctor_id, visit_date, raw_input, status FROM visits WHERE id = $1`, visitID,
	).Scan(&visit.ID, &visit.PatientID, &visit.DoctorID, &visit.VisitDate, &visit.RawInput, &visit.Status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Visit not found"})
		return
	}

	// Fetch patient
	var patient models.Patient
	config.DB.QueryRow(
		`SELECT id, full_name, phone FROM patients WHERE id = $1`, visit.PatientID,
	).Scan(&patient.ID, &patient.FullName, &patient.Phone)

	// Fetch doctor
	var doctor models.Doctor
	config.DB.QueryRow(
		`SELECT id, full_name, speciality FROM doctors WHERE id = $1`, visit.DoctorID,
	).Scan(&doctor.ID, &doctor.FullName, &doctor.Speciality)

	// Fetch prescriptions
	rows, _ := config.DB.Query(
		`SELECT id, visit_id, drug_name, dosage, frequency, duration, quantity, unit_price, instructions, created_at
		 FROM prescriptions WHERE visit_id = $1`, visitID)
	defer rows.Close()
	var prescriptions []models.Prescription
	for rows.Next() {
		var p models.Prescription
		rows.Scan(&p.ID, &p.VisitID, &p.DrugName, &p.Dosage, &p.Frequency,
			&p.Duration, &p.Quantity, &p.UnitPrice, &p.Instructions, &p.CreatedAt)
		prescriptions = append(prescriptions, p)
	}

	// Fetch lab orders
	labRows, _ := config.DB.Query(
		`SELECT id, visit_id, test_name, unit_price, notes, created_at FROM lab_orders WHERE visit_id = $1`, visitID)
	defer labRows.Close()
	var labOrders []models.LabOrder
	for labRows.Next() {
		var lo models.LabOrder
		labRows.Scan(&lo.ID, &lo.VisitID, &lo.TestName, &lo.UnitPrice, &lo.Notes, &lo.CreatedAt)
		labOrders = append(labOrders, lo)
	}

	// Fetch clinic notes
	noteRows, _ := config.DB.Query(
		`SELECT id, visit_id, note_text, category, created_at FROM clinic_notes WHERE visit_id = $1`, visitID)
	defer noteRows.Close()
	var clinicNotes []models.ClinicNote
	for noteRows.Next() {
		var n models.ClinicNote
		noteRows.Scan(&n.ID, &n.VisitID, &n.NoteText, &n.Category, &n.CreatedAt)
		clinicNotes = append(clinicNotes, n)
	}

	// Fetch bill + line items
	bill, lineItems, err := services.GetBillWithLineItems(visitID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bill not found"})
		return
	}

	// Generate PDF
	pdfBytes, err := services.GenerateVisitPDF(
		&patient, &doctor, &visit,
		prescriptions, labOrders, clinicNotes,
		bill, lineItems,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF generation failed: " + err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=visit-"+visitID[:8]+".pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}