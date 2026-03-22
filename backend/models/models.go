package models

import "time"

type Patient struct {
	ID        string    `json:"id"`
	FullName  string    `json:"full_name"`
	DOB       string    `json:"dob"`
	Gender    string    `json:"gender"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
}

type Doctor struct {
	ID        string    `json:"id"`
	FullName  string    `json:"full_name"`
	Speciality string   `json:"speciality"`
	LicenseNo string    `json:"license_no"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Visit struct {
	ID          string     `json:"id"`
	PatientID   string     `json:"patient_id"`
	DoctorID    string     `json:"doctor_id"`
	VisitDate   time.Time  `json:"visit_date"`
	RawInput    string     `json:"raw_input"`
	AIParsedAt  *time.Time `json:"ai_parsed_at"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
}

type ClinicNote struct {
	ID        string    `json:"id"`
	VisitID   string    `json:"visit_id"`
	NoteText  string    `json:"note_text"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
}

type Prescription struct {
	ID           string    `json:"id"`
	VisitID      string    `json:"visit_id"`
	DrugName     string    `json:"drug_name"`
	CatalogueID  *string   `json:"catalogue_id"`
	Dosage       string    `json:"dosage"`
	Frequency    string    `json:"frequency"`
	Duration     string    `json:"duration"`
	Quantity     int       `json:"quantity"`
	UnitPrice    float64   `json:"unit_price"`
	Instructions string    `json:"instructions"`
	CreatedAt    time.Time `json:"created_at"`
}

type LabOrder struct {
	ID          string    `json:"id"`
	VisitID     string    `json:"visit_id"`
	TestName    string    `json:"test_name"`
	CatalogueID *string   `json:"catalogue_id"`
	UnitPrice   float64   `json:"unit_price"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
}

type Bill struct {
	ID              string     `json:"id"`
	VisitID         string     `json:"visit_id"`
	ConsultationFee float64    `json:"consultation_fee"`
	DrugTotal       float64    `json:"drug_total"`
	LabTotal        float64    `json:"lab_total"`
	OtherCharges    float64    `json:"other_charges"`
	Discount        float64    `json:"discount"`
	GrandTotal      float64    `json:"grand_total"`
	Status          string     `json:"status"`
	PaidAt          *time.Time `json:"paid_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

type BillLineItem struct {
	ID          string    `json:"id"`
	BillID      string    `json:"bill_id"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	LineTotal   float64   `json:"line_total"`
	CreatedAt   time.Time `json:"created_at"`
}

type ParsedDrug struct {
	DrugName     string  `json:"drug_name"`
	Dosage       string  `json:"dosage"`
	Frequency    string  `json:"frequency"`
	Duration     string  `json:"duration"`
	Quantity     int     `json:"quantity"`
	Instructions string  `json:"instructions"`
	UnitPrice    float64 `json:"unit_price"`
}

type ParsedLabTest struct {
	TestName  string  `json:"test_name"`
	Notes     string  `json:"notes"`
	UnitPrice float64 `json:"unit_price"`
}

type ParsedNote struct {
	NoteText string `json:"note_text"`
	Category string `json:"category"` // observation | diagnosis | history | allergy | other
}

type ParsedResult struct {
	Drugs    []ParsedDrug    `json:"drugs"`
	LabTests []ParsedLabTest `json:"lab_tests"`
	Notes    []ParsedNote    `json:"notes"`
}

type ParseRequest struct {
	VisitID   string `json:"visit_id"`
	RawInput  string `json:"raw_input"`
	PatientID string `json:"patient_id"`
	DoctorID  string `json:"doctor_id"`
}

type ParseResponse struct {
	VisitID       string         `json:"visit_id"`
	Parsed        ParsedResult   `json:"parsed"`
	Prescriptions []Prescription `json:"prescriptions"`
	LabOrders     []LabOrder     `json:"lab_orders"`
	Notes         []ClinicNote   `json:"clinic_notes"`
	Bill          Bill           `json:"bill"`
}