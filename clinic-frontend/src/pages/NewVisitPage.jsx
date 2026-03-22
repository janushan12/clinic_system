import { useState, useEffect, useRef } from 'react'
import { getPatients, getDoctors, createPatient, parseVisit } from '../api/api'
import '../styles/newvisit.css'

const SAMPLE_TEXT = `Patient presents with fever for 3 days, sore throat, and mild headache.
On examination: Temp 38.8°C, throat congested, tonsils enlarged with exudate.
No known drug allergies. No previous hospitalizations.

Diagnosis: Acute Tonsillitis

Prescriptions:
- Amoxicillin 500mg three times daily for 7 days
- Paracetamol 500mg every 6 hours as needed for fever
- Cetirizine 10mg once daily for 5 days

Investigations required:
- Full Blood Count (FBC)
- Throat swab culture

Advice: Adequate rest, increase fluid intake. Return if no improvement in 3 days.`

const AI_STEPS = [
  'Sending to Claude AI...',
  'Classifying drugs & dosages...',
  'Detecting lab tests...',
  'Extracting clinical notes...',
  'Generating bill...',
]

export default function NewVisitPage({ onParsed }) {
  const [patients,       setPatients]       = useState([])
  const [doctors,        setDoctors]        = useState([])
  const [patientID,      setPatientID]      = useState('')
  const [doctorID,       setDoctorID]       = useState('')
  const [rawInput,       setRawInput]       = useState('')
  const [loading,        setLoading]        = useState(false)
  const [loadingData,    setLoadingData]    = useState(true)
  const [error,          setError]          = useState('')
  const [backendError,   setBackendError]   = useState('')
  const [showNewPatient, setShowNewPatient] = useState(false)
  const [newPatient,     setNewPatient]     = useState({ full_name:'', dob:'', gender:'male', phone:'' })
  const [aiStep,         setAiStep]         = useState(0)
  const textareaRef = useRef(null)

  useEffect(() => {
    setLoadingData(true)
    Promise.allSettled([
      getPatients().then(r => setPatients(r.data || [])),
      getDoctors().then(r => {
        const list = r.data || []
        setDoctors(list)
        if (list.length > 0) setDoctorID(list[0].id)
      }),
    ]).then(results => {
      setLoadingData(false)
      if (results.some(r => r.status === 'rejected'))
        setBackendError('Cannot connect to backend at localhost:8080. Is your Go server running?')
    })
  }, [])

  useEffect(() => {
    let interval
    if (loading) {
      setAiStep(0)
      interval = setInterval(() => setAiStep(s => (s + 1) % AI_STEPS.length), 900)
    }
    return () => clearInterval(interval)
  }, [loading])

  const handleAddPatient = async () => {
    if (!newPatient.full_name || !newPatient.dob) return
    try {
      const r = await createPatient(newPatient)
      setPatients(prev => [...prev, r.data])
      setPatientID(r.data.id)
      setShowNewPatient(false)
      setNewPatient({ full_name:'', dob:'', gender:'male', phone:'' })
    } catch { setError('Failed to create patient') }
  }

  const handleParse = async () => {
    if (!patientID)      { setError('Please select a patient');    return }
    if (!doctorID)       { setError('Please select a doctor');     return }
    if (!rawInput.trim()){ setError('Please enter clinical notes'); return }
    setError('')
    setLoading(true)
    try {
      const r = await parseVisit({ patient_id: patientID, doctor_id: doctorID, raw_input: rawInput })
      onParsed(r.data)
    } catch(e) {
      setError(e.response?.data?.error || 'Parsing failed. Check API key and backend.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="new-visit-page fade-in">

      {backendError && <Banner type="warning" message={backendError} />}

      {/* Stats row */}
      <div className="stats-row">
        <StatCard icon="👤" label="Patients"  value={loadingData ? '—' : patients.length} color="var(--accent)"  />
        <StatCard icon="🩺" label="Doctors"   value={loadingData ? '—' : doctors.length}  color="var(--success)" />
        <StatCard icon="✦"  label="AI Model"  value="Claude Sonnet"                        color="var(--drug)"    small />
      </div>

      {/* Patient + Doctor */}
      <div className="two-col">
        <div className="card">
          <div className="card-header">
            <div className="card-title"><span>👤</span> Patient</div>
          </div>
          {loadingData
            ? <div className="skeleton" style={{ height:38, borderRadius:8 }} />
            : <select className="form-select" value={patientID} onChange={e => setPatientID(e.target.value)}>
                <option value="">— Select patient —</option>
                {patients.map(p => <option key={p.id} value={p.id}>{p.full_name}</option>)}
              </select>
          }
          <button className="btn-link" onClick={() => setShowNewPatient(!showNewPatient)}>
            {showNewPatient ? '✕ Cancel' : '+ Add new patient'}
          </button>
          {showNewPatient && (
            <div className="new-patient-form">
              <input className="form-input" placeholder="Full name *"
                value={newPatient.full_name}
                onChange={e => setNewPatient(p => ({ ...p, full_name: e.target.value }))} />
              <div className="two-inputs">
                <input className="form-input" type="date"
                  value={newPatient.dob}
                  onChange={e => setNewPatient(p => ({ ...p, dob: e.target.value }))} />
                <select className="form-input"
                  value={newPatient.gender}
                  onChange={e => setNewPatient(p => ({ ...p, gender: e.target.value }))}>
                  <option value="male">Male</option>
                  <option value="female">Female</option>
                  <option value="other">Other</option>
                </select>
              </div>
              <input className="form-input" placeholder="Phone number"
                value={newPatient.phone}
                onChange={e => setNewPatient(p => ({ ...p, phone: e.target.value }))} />
              <button className="btn-sm" onClick={handleAddPatient}>Add Patient</button>
            </div>
          )}
        </div>

        <div className="card">
          <div className="card-header">
            <div className="card-title"><span>🩺</span> Doctor</div>
          </div>
          {loadingData
            ? <div className="skeleton" style={{ height:38, borderRadius:8 }} />
            : <select className="form-select" value={doctorID} onChange={e => setDoctorID(e.target.value)}>
                <option value="">— Select doctor —</option>
                {doctors.map(d => <option key={d.id} value={d.id}>{d.full_name} · {d.speciality}</option>)}
              </select>
          }
          <div className="ai-hint">
            <div className="ai-hint-title">✦ AI Classification</div>
            <div className="ai-hint-text">
              Type freely — drugs, dosages, lab tests, observations — all in one block.
              Claude AI will separate and classify everything automatically.
            </div>
          </div>
        </div>
      </div>

      {/* Notes textarea */}
      <div className="card">
        <div className="card-header">
          <div className="card-title"><span>📝</span> Clinical Notes</div>
          <button className="btn-link" style={{ marginTop:0 }} onClick={() => setRawInput(SAMPLE_TEXT)}>
            Load sample
          </button>
        </div>
        <textarea
          ref={textareaRef}
          className="notes-textarea"
          value={rawInput}
          onChange={e => setRawInput(e.target.value)}
          placeholder={`Type or dictate clinical notes freely...\n\nExample:\nPatient has fever and sore throat...\nPrescribe Amoxicillin 500mg three times daily for 7 days...\nOrder Full Blood Count and throat swab...\nDiagnosis: Acute tonsillitis`}
        />
        <div className="textarea-footer">
          <span>{rawInput.length} characters · {rawInput.trim().split(/\s+/).filter(Boolean).length} words</span>
          <span>Include drug names, dosages, test names, and clinical findings</span>
        </div>
      </div>

      {error && <Banner type="danger" message={error} />}

      {/* Parse button */}
      <button className="parse-btn" onClick={handleParse} disabled={loading}>
        {loading ? (
          <div className="ai-loading">
            <div className="ai-loading-row">
              <div className="spinner" />
              <span className="ai-loading-text">{AI_STEPS[aiStep]}</span>
            </div>
            <div className="ai-progress-bar">
              <div className="ai-progress-fill" style={{ width: `${((aiStep + 1) / AI_STEPS.length) * 100}%` }} />
            </div>
          </div>
        ) : (
          <><span style={{ fontSize:18 }}>✦</span> Parse & Generate Bill</>
        )}
      </button>
    </div>
  )
}

function StatCard({ icon, label, value, color, small }) {
  return (
    <div className="stat-card" style={{ '--stat-color': color }}>
      <div className="stat-card-icon">{icon}</div>
      <div>
        <div className={`stat-card-value ${small ? 'small' : ''}`}>{value}</div>
        <div className="stat-card-label">{label}</div>
      </div>
    </div>
  )
}

function Banner({ type, message }) {
  const icons = { warning:'⚠️', danger:'✕', success:'✓' }
  return (
    <div className={`banner ${type}`}>
      {icons[type]} {message}
    </div>
  )
}