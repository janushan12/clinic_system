package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"clinic-notes/config"
	"clinic-notes/models"
)

func lookupDrugPrice(drugName string) (*string, float64) {
	query := `
		SELECT id, unit_price
		FROM drug_catalogue
		WHERE LOWER(generic_name) LIKE LOWER($1)
		   OR LOWER(brand_name)   LIKE LOWER($1)
		LIMIT 1`

	searchTerm := "%" + strings.ToLower(drugName) + "%"
	var id string
	var price float64
	err := config.DB.QueryRow(query, searchTerm).Scan(&id, &price)
	if err != nil {
		return nil, 0.00 // not found — price stays 0
	}
	return &id, price
}

func lookupLabPrice(testName string) (*string, float64) {
	query := `
		SELECT id, unit_price
		FROM lab_test_catalogue
		WHERE LOWER(test_name) LIKE LOWER($1)
		LIMIT 1`

	searchTerm := "%" + strings.ToLower(testName) + "%"
	var id string
	var price float64
	err := config.DB.QueryRow(query, searchTerm).Scan(&id, &price)
	if err != nil {
		return nil, 0.00
	}
	return &id, price
}

func lookupConsultationFee() float64 {
	var amount float64
	err := config.DB.QueryRow(
		`SELECT amount FROM charge_catalogue WHERE name = 'Consultation Fee' LIMIT 1`,
	).Scan(&amount)
	if err != nil {
		return 15.00 // fallback default
	}
	return amount
}

func SaveParsedVisit(visitID string, parsed *models.ParsedResult) (*models.ParseResponse, error) {
	tx, err := config.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	var prescriptions []models.Prescription
	var labOrders []models.LabOrder
	var notes []models.ClinicNote

	drugTotal := 0.00
	labTotal := 0.00

	// ── 1. Save prescriptions ──────────────────────────────
	for _, d := range parsed.Drugs {
		catID, price := lookupDrugPrice(d.DrugName)
		if d.UnitPrice > 0 {
			price = d.UnitPrice // use AI price if provided
		}
		lineTotal := price * float64(d.Quantity)
		drugTotal += lineTotal

		var id string
		err := tx.QueryRow(`
			INSERT INTO prescriptions
				(visit_id, drug_name, catalogue_id, dosage, frequency, duration, quantity, unit_price, instructions)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			RETURNING id`,
			visitID, d.DrugName, catID, d.Dosage, d.Frequency, d.Duration,
			d.Quantity, price, d.Instructions,
		).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to insert prescription: %w", err)
		}
		prescriptions = append(prescriptions, models.Prescription{
			ID: id, VisitID: visitID, DrugName: d.DrugName,
			CatalogueID: catID, Dosage: d.Dosage, Frequency: d.Frequency,
			Duration: d.Duration, Quantity: d.Quantity,
			UnitPrice: price, Instructions: d.Instructions, CreatedAt: now,
		})
	}
	for _, lt := range parsed.LabTests {
		catID, price := lookupLabPrice(lt.TestName)
		if lt.UnitPrice > 0 {
			price = lt.UnitPrice
		}
		labTotal += price

		var id string
		err := tx.QueryRow(`
			INSERT INTO lab_orders (visit_id, test_name, catalogue_id, unit_price, notes)
			VALUES ($1,$2,$3,$4,$5)
			RETURNING id`,
			visitID, lt.TestName, catID, price, lt.Notes,
		).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to insert lab order: %w", err)
		}
		labOrders = append(labOrders, models.LabOrder{
			ID: id, VisitID: visitID, TestName: lt.TestName,
			CatalogueID: catID, UnitPrice: price, Notes: lt.Notes, CreatedAt: now,
		})
	}
	for _, n := range parsed.Notes {
		var id string
		err := tx.QueryRow(`
			INSERT INTO clinic_notes (visit_id, note_text, category)
			VALUES ($1,$2,$3)
			RETURNING id`,
			visitID, n.NoteText, n.Category,
		).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to insert clinic note: %w", err)
		}
		notes = append(notes, models.ClinicNote{
			ID: id, VisitID: visitID,
			NoteText: n.NoteText, Category: n.Category, CreatedAt: now,
		})
	}
	_, err = tx.Exec(
		`UPDATE visits SET ai_parsed_at = $1, status = 'confirmed' WHERE id = $2`,
		now, visitID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update visit: %w", err)
	}
	consultationFee := lookupConsultationFee()
	grandTotal := consultationFee + drugTotal + labTotal

	var billID string
	err = tx.QueryRow(`
		INSERT INTO bills (visit_id, consultation_fee, drug_total, lab_total, other_charges, discount)
		VALUES ($1,$2,$3,$4,0,0)
		RETURNING id, grand_total`,
		visitID, consultationFee, drugTotal, labTotal,
	).Scan(&billID, &grandTotal)
	if err != nil {
		return nil, fmt.Errorf("failed to create bill: %w", err)
	}
	_, err = tx.Exec(`
		INSERT INTO bill_line_items (bill_id, category, description, quantity, unit_price)
		VALUES ($1,'consultation','Consultation Fee',1,$2)`,
		billID, consultationFee,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert consultation line: %w", err)
	}

	for _, p := range prescriptions {
		_, err = tx.Exec(`
			INSERT INTO bill_line_items (bill_id, category, description, quantity, unit_price)
			VALUES ($1,'drug',$2,$3,$4)`,
			billID, p.DrugName, p.Quantity, p.UnitPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert drug line: %w", err)
		}
	}
	for _, lo := range labOrders {
		_, err = tx.Exec(`
			INSERT INTO bill_line_items (bill_id, category, description, quantity, unit_price)
			VALUES ($1,'lab_test',$2,1,$3)`,
			billID, lo.TestName, lo.UnitPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert lab line: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	bill := models.Bill{
		ID: billID, VisitID: visitID,
		ConsultationFee: consultationFee,
		DrugTotal:       drugTotal,
		LabTotal:        labTotal,
		GrandTotal:      grandTotal,
		Status:          "pending",
		CreatedAt:       now,
	}

	return &models.ParseResponse{
		VisitID:       visitID,
		Parsed:        *parsed,
		Prescriptions: prescriptions,
		LabOrders:     labOrders,
		Notes:         notes,
		Bill:          bill,
	}, nil
}

func GetBillWithLineItems(visitID string) (*models.Bill, []models.BillLineItem, error) {
	var bill models.Bill
	err := config.DB.QueryRow(`
		SELECT id, visit_id, consultation_fee, drug_total, lab_total,
		       other_charges, discount, grand_total, status, created_at
		FROM bills WHERE visit_id = $1`, visitID,
	).Scan(
		&bill.ID, &bill.VisitID, &bill.ConsultationFee, &bill.DrugTotal,
		&bill.LabTotal, &bill.OtherCharges, &bill.Discount, &bill.GrandTotal,
		&bill.Status, &bill.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil, fmt.Errorf("bill not found for visit %s", visitID)
	}
	if err != nil {
		return nil, nil, err
	}

	rows, err := config.DB.Query(`
		SELECT id, bill_id, category, description, quantity, unit_price, line_total, created_at
		FROM bill_line_items WHERE bill_id = $1 ORDER BY category, created_at`, bill.ID,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var items []models.BillLineItem
	for rows.Next() {
		var item models.BillLineItem
		if err := rows.Scan(
			&item.ID, &item.BillID, &item.Category, &item.Description,
			&item.Quantity, &item.UnitPrice, &item.LineTotal, &item.CreatedAt,
		); err != nil {
			return nil, nil, err
		}
		items = append(items, item)
	}
	return &bill, items, nil
}
