# ABC Health Clinic — AI-Powered Unified Notes & Billing System

> A full-stack clinic management system that uses Claude AI to automatically classify free-form doctor notes into drugs, lab tests, and clinical observations — then generates a complete patient bill.

## Tech Stack
| Layer | Technology |
|-------|-----------|
| Backend | Go 1.21 + Gin |
| Database | PostgreSQL 14+ |
| Frontend | React 18 + Vite |
| AI | Anthropic Claude Sonnet |


## Step 1 — Database Setup
1. Open pgAdmin → create a new database (e.g. `clinic_db`)
2. Right-click database → Query Tool
3. Run each STEP block from `migrations_step_by_step.sql` in order (Steps 1–20)

## Step 2 — Backend Setup
```bash
cd backend
cp .env.example .env   # fill in your DB credentials and API key
go mod tidy
go run main.go         # starts at http://localhost:8080
```

`.env` values needed:
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=clinic_new_db
DB_SSLMODE=disable
ANTHROPIC_API_KEY=sk-ant-...
PORT=8080
```

## Step 3 — Frontend Setup
```bash
cd frontend
npm install
npm run dev            # starts at http://localhost:3000
```

## How to Use
1. Select patient and doctor
2. Type clinical notes freely (drugs, dosages, lab tests, observations — all together)
3. Click **Parse & Generate Bill**
4. View AI-classified results in three columns
5. Review the auto-generated itemised bill
6. Print or mark as paid

## API Endpoints
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/parse | Parse notes + generate bill |
| GET | /api/visits/:id | Get visit with clinical data |
| GET | /api/visits/:id/bill | Get bill with line items |
| POST | /api/visits/:id/bill/pay | Mark bill as paid |
| GET | /api/patients | List patients |
| POST | /api/patients | Create patient |
| GET | /api/doctors | List doctors |

## Troubleshooting
| Problem | Fix |
|---------|-----|
| Failed to connect to DB | Check .env credentials, ensure PostgreSQL is running |
| API key not set | Ensure .env file is in the backend/ folder |
| Credit balance too low | Mock parser activates as fallback automatically |
| Frontend blank page | Check browser console F12 for errors |
