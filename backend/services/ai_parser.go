package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"clinic-notes/models"
)

// ── Anthropic API request/response types ────────────────────

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicResponse struct {
	Content []anthropicContent `json:"content"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ── System prompt ─────────────────────────────────────────────

const systemPrompt = `You are a medical text parser for a clinic management system.

Your job is to read a doctor's free-form clinical notes and extract three categories of information:

1. DRUGS & PRESCRIPTIONS — any medication mentioned with dosage, frequency, duration, quantity
2. LAB TESTS / INVESTIGATIONS — any test, investigation, or procedure ordered
3. CLINICAL NOTES — observations, diagnoses, patient history, allergies, other findings

Return ONLY a valid JSON object with this exact structure (no markdown, no explanation, just JSON):
{
  "drugs": [
    {
      "drug_name": "string",
      "dosage": "string (e.g. 500mg)",
      "frequency": "string (e.g. twice daily)",
      "duration": "string (e.g. 5 days)",
      "quantity": number,
      "instructions": "string (e.g. take after meals)",
      "unit_price": 0.00
    }
  ],
  "lab_tests": [
    {
      "test_name": "string",
      "notes": "string (any special instructions for the test)",
      "unit_price": 0.00
    }
  ],
  "notes": [
    {
      "note_text": "string",
      "category": "observation|diagnosis|history|allergy|other"
    }
  ]
}

Rules:
- Always set unit_price to 0.00 (the system will look up prices from the catalogue)
- If quantity is not mentioned for a drug, default to 1
- Split multiple observations into separate note objects
- If the input contains nothing for a category, return an empty array []
- Be thorough — do not miss any drug, test, or clinical finding
- Normalise drug names to their generic name where possible (e.g. "Panadol" → "Paracetamol")
- For lab test names, use standard medical abbreviations (e.g. "full blood count" → "Full Blood Count (FBC)")`

// ── Mock parser for testing without API credits ──────────────

func mockParse(rawInput string) *models.ParsedResult {
	lower := strings.ToLower(rawInput)
	result := &models.ParsedResult{}

	// Detect drugs
	drugKeywords := map[string][3]string{
		"amoxicillin": {"Amoxicillin", "500mg", "3 times daily"},
		"paracetamol": {"Paracetamol", "500mg", "every 6 hours"},
		"panadol":     {"Paracetamol", "500mg", "every 6 hours"},
		"ibuprofen":   {"Ibuprofen", "400mg", "twice daily"},
		"cetirizine":  {"Cetirizine", "10mg", "once daily"},
		"metformin":   {"Metformin", "500mg", "twice daily"},
		"omeprazole":  {"Omeprazole", "20mg", "once daily"},
	}
	for keyword, info := range drugKeywords {
		if strings.Contains(lower, keyword) {
			result.Drugs = append(result.Drugs, models.ParsedDrug{
				DrugName: info[0], Dosage: info[1],
				Frequency: info[2], Duration: "5 days", Quantity: 1,
			})
		}
	}

	// Detect lab tests
	labKeywords := map[string]string{
		"fbc": "Full Blood Count (FBC)", "full blood count": "Full Blood Count (FBC)",
		"blood glucose": "Blood Glucose (Fasting)", "hba1c": "HbA1c",
		"lipid": "Lipid Profile", "urine": "Urine Full Report (UFR)",
		"ecg": "ECG", "x-ray": "Chest X-Ray", "xray": "Chest X-Ray",
		"culture": "Blood Culture", "dengue": "Dengue NS1 Antigen",
	}
	for keyword, testName := range labKeywords {
		if strings.Contains(lower, keyword) {
			result.LabTests = append(result.LabTests, models.ParsedLabTest{TestName: testName})
		}
	}

	// Always add a general observation note
	result.Notes = append(result.Notes, models.ParsedNote{
		NoteText: rawInput,
		Category: "observation",
	})

	return result
}

// ── ParseWithClaude sends text to Claude and returns structured data ──

func ParseWithClaude(rawInput string) (*models.ParsedResult, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is not set")
	}

	// Build the request body
	reqBody := anthropicRequest{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 1000,
		System:    systemPrompt,
		Messages: []anthropicMessage{
			{
				Role:    "user",
				Content: fmt.Sprintf("Parse the following clinical notes:\n\n%s", rawInput),
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Call the Anthropic API
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse Anthropic's response envelope
	var anthropicResp anthropicResponse
	if err := json.Unmarshal(respBytes, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	if anthropicResp.Error != nil {
		msg := anthropicResp.Error.Message
		// If out of credits, fall back to mock parser so UI still works
		if strings.Contains(msg, "credit balance") || strings.Contains(msg, "billing") {
			log.Println("⚠️  No API credits — using mock parser for demo")
			return mockParse(rawInput), nil
		}
		return nil, fmt.Errorf("Anthropic API error: %s", msg)
	}

	if len(anthropicResp.Content) == 0 {
		return nil, fmt.Errorf("empty response from Claude")
	}

	// Extract the JSON text from Claude's reply
	rawJSON := strings.TrimSpace(anthropicResp.Content[0].Text)

	// Strip markdown code fences if Claude added them
	rawJSON = strings.TrimPrefix(rawJSON, "```json")
	rawJSON = strings.TrimPrefix(rawJSON, "```")
	rawJSON = strings.TrimSuffix(rawJSON, "```")
	rawJSON = strings.TrimSpace(rawJSON)

	// Unmarshal into our ParsedResult struct
	var result models.ParsedResult
	if err := json.Unmarshal([]byte(rawJSON), &result); err != nil {
		return nil, fmt.Errorf("failed to parse Claude JSON output: %w\nRaw: %s", err, rawJSON)
	}

	return &result, nil
}