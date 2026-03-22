import { downloadPDF, markPaid } from "../api/api"
import { useRef, useState } from 'react'
import '../styles/result.css'

export default function VisitResultPage({ result, onNew }) {
  const visit_id      = result.visit_id     || result.visitID     || ''
  const prescriptions = result.prescriptions || []
  const lab_orders    = result.lab_orders   || result.labOrders   || []
  const clinic_notes  = result.clinic_notes || result.clinicNotes || []
  const bill          = result.bill         || {}
 
  const [billStatus, setBillStatus] = useState(bill.status || 'pending')
  const [paying,     setPaying]     = useState(false)
  const printRef = useRef()
 
  const handleMarkPaid = async () => {
    setPaying(true)
    try {
      await markPaid(visit_id)
      setBillStatus('paid')
    } catch { alert('Failed to mark as paid') }
    finally { setPaying(false) }
  }
 
  return (
    <div className="result-page fade-in">
 
      {/* Summary banner */}
      <div className="summary-banner no-print">
        <div>
          <div className="summary-banner-eyebrow">AI Classification Complete</div>
          <div className="summary-banner-title">Visit Summary</div>
          <div className="summary-banner-id">
            ID: <code className="visit-id-code">{visit_id?.slice(0,8)}…</code>
            &nbsp;·&nbsp;
            {new Date().toLocaleDateString('en-US', { weekday:'long', year:'numeric', month:'long', day:'numeric' })}
          </div>
        </div>
        <div className="banner-actions">
          <button className="btn-action light" onClick={() => window.print()}>🖨 Print</button>
          <button className="btn-action light" onClick={() => downloadPDF(visit_id)}>⬇ PDF</button>
          <button className="btn-action solid" onClick={onNew}>+ New Visit</button>
        </div>
      </div>
 
      {/* Mini stats */}
      <div className="mini-stats no-print">
        <MiniStat icon="💊" label="Drugs Prescribed" value={prescriptions.length} color="var(--drug)"    />
        <MiniStat icon="🧪" label="Lab Tests"         value={lab_orders.length}    color="var(--lab)"     />
        <MiniStat icon="📋" label="Clinical Notes"    value={clinic_notes.length}  color="var(--note)"    />
        <MiniStat icon="💰" label="Grand Total"       value={`$${(bill.grand_total||0).toFixed(2)}`} color="var(--accent)" />
      </div>
 
      {/* 3-column results */}
      <div className="results-grid">
        <ResultSection title="Drugs & Dosages" icon="💊"
          color="var(--drug)" bg="var(--drug-lt)" border="var(--drug-border)"
          count={prescriptions.length}>
          {prescriptions.length === 0 && <EmptyState text="No drugs prescribed" />}
          {prescriptions.map((p,i) => <DrugCard key={i} drug={p} />)}
        </ResultSection>
 
        <ResultSection title="Lab Tests" icon="🧪"
          color="var(--lab)" bg="var(--lab-lt)" border="var(--lab-border)"
          count={lab_orders.length}>
          {lab_orders.length === 0 && <EmptyState text="No tests ordered" />}
          {lab_orders.map((lo,i) => <LabCard key={i} order={lo} />)}
        </ResultSection>
 
        <ResultSection title="Clinical Notes" icon="📋"
          color="var(--note)" bg="var(--note-lt)" border="var(--note-border)"
          count={clinic_notes.length}>
          {clinic_notes.length === 0 && <EmptyState text="No notes recorded" />}
          {clinic_notes.map((n,i) => <NoteCard key={i} note={n} />)}
        </ResultSection>
      </div>
 
      {/* Bill */}
      <div className="bill-card">
        <div className="bill-header">
          <div className="bill-header-left">
            <div className="bill-header-icon">💰</div>
            <div>
              <div className="bill-header-title">Patient Bill</div>
              <div className="bill-header-sub">Itemised breakdown · Auto-calculated</div>
            </div>
          </div>
          <StatusBadge status={billStatus} />
        </div>
 
        <div className="bill-table-wrap">
          <table className="bill-table">
            <thead>
              <tr>
                <th style={{ textAlign:'left'   }}>Item</th>
                <th style={{ textAlign:'left'   }}>Category</th>
                <th style={{ textAlign:'center' }}>Qty</th>
                <th style={{ textAlign:'right'  }}>Unit Price</th>
                <th style={{ textAlign:'right'  }}>Total</th>
              </tr>
            </thead>
            <tbody>
              <BillRow icon="🩺" label="Consultation Fee" category="consultation" qty={1}               unit={bill.consultation_fee||0} />
              {prescriptions.map((p,i)  => <BillRow key={i} icon="💊" label={p.drug_name}  category="drug"    qty={p.quantity} unit={p.unit_price||0}  />)}
              {lab_orders.map((lo,i)    => <BillRow key={i} icon="🧪" label={lo.test_name} category="lab"     qty={1}          unit={lo.unit_price||0} />)}
            </tbody>
          </table>
        </div>
 
        <div className="bill-footer">
          <div className="totals-wrap">
            <div className="totals-block">
              <TotalRow label="Consultation" value={bill.consultation_fee||0} />
              <TotalRow label="Drugs"        value={bill.drug_total||0}       />
              <TotalRow label="Lab Tests"    value={bill.lab_total||0}        />
              {(bill.discount||0) > 0 && <TotalRow label="Discount" value={-(bill.discount||0)} color="var(--success)" />}
              <div className="grand-total-row">
                <span className="grand-total-label">Grand Total</span>
                <span className="grand-total-value">${(bill.grand_total||0).toFixed(2)}</span>
              </div>
            </div>
          </div>
 
          {billStatus === 'pending' && (
            <div className="bill-actions">
              <button className="btn-pay" onClick={handleMarkPaid} disabled={paying}>
                {paying ? '...' : '✓ Mark as Paid'}
              </button>
            </div>
          )}
          {billStatus === 'paid' && (
            <div className="bill-actions">
              <span className="paid-badge">✓ Payment received</span>
            </div>
          )}
        </div>
      </div>
 
      {/* Print view */}
      <PrintView result={result} printRef={printRef} />
      <style>{`@media print { .no-print { display:none !important; } }`}</style>
    </div>
  )
}
 
/* ── Sub-components ─────────────────────────────────────────── */
 
function ResultSection({ title, icon, color, bg, border, count, children }) {
  return (
    <div className="result-section">
      <div className="result-section-header" style={{ '--section-color':color, '--section-bg':bg, '--section-border':border }}>
        <div className="result-section-title">{icon} {title}</div>
        <div className="result-section-count">{count}</div>
      </div>
      <div className="result-section-body">{children}</div>
    </div>
  )
}
 
function DrugCard({ drug }) {
  return (
    <div className="drug-card">
      <div className="drug-card-name">{drug.drug_name}</div>
      <div className="drug-pills">
        {drug.dosage    && <span className="drug-pill">💊 {drug.dosage}</span>}
        {drug.frequency && <span className="drug-pill">🕐 {drug.frequency}</span>}
        {drug.duration  && <span className="drug-pill">📅 {drug.duration}</span>}
      </div>
      {drug.instructions && <div className="drug-instructions">📝 {drug.instructions}</div>}
      {(drug.unit_price||0) > 0 && (
        <div className="drug-price">
          ${drug.unit_price.toFixed(2)} × {drug.quantity} = ${(drug.unit_price * drug.quantity).toFixed(2)}
        </div>
      )}
    </div>
  )
}
 
function LabCard({ order }) {
  return (
    <div className="lab-card">
      <div className="lab-card-name">{order.test_name}</div>
      {order.notes && <div className="lab-card-notes">{order.notes}</div>}
      {(order.unit_price||0) > 0 && <div className="lab-card-price">${order.unit_price.toFixed(2)}</div>}
    </div>
  )
}
 
function NoteCard({ note }) {
  const catStyles = {
    observation: { color:'var(--note)',    bg:'#dcfce7' },
    diagnosis:   { color:'var(--lab)',     bg:'#e0f2fe' },
    history:     { color:'var(--warning)', bg:'var(--warning-lt)' },
    allergy:     { color:'var(--danger)',  bg:'var(--danger-lt)'  },
    other:       { color:'var(--text2)',   bg:'var(--bg)'         },
  }
  const s = catStyles[note.category] || catStyles.other
  return (
    <div className="note-card">
      <span className="note-category-badge" style={{ color:s.color, background:s.bg }}>
        {note.category}
      </span>
      <div className="note-text">{note.note_text}</div>
    </div>
  )
}
 
function BillRow({ icon, label, category, qty, unit }) {
  return (
    <tr>
      <td className="bold">{icon} {label}</td>
      <td><span className={`category-badge ${category}`}>{category}</span></td>
      <td className="muted" style={{ textAlign:'center' }}>{qty}</td>
      <td className="mono muted" style={{ textAlign:'right' }}>${unit.toFixed(2)}</td>
      <td className="mono bold" style={{ textAlign:'right' }}>${(qty * unit).toFixed(2)}</td>
    </tr>
  )
}
 
function TotalRow({ label, value, color }) {
  return (
    <div className="total-row">
      <span className="total-row-label">{label}</span>
      <span className="total-row-value" style={{ color: color || 'var(--text)' }}>
        ${Math.abs(value).toFixed(2)}
      </span>
    </div>
  )
}
 
function MiniStat({ icon, label, value, color }) {
  return (
    <div className="mini-stat" style={{ '--mini-color': color }}>
      <div className="mini-stat-icon">{icon}</div>
      <div className="mini-stat-value">{value}</div>
      <div className="mini-stat-label">{label}</div>
    </div>
  )
}
 
function StatusBadge({ status }) {
  const labels = { pending:'Pending Payment', paid:'Paid', cancelled:'Cancelled' }
  return <span className={`status-badge ${status}`}>{labels[status] || 'Pending'}</span>
}
 
function EmptyState({ text }) {
  return (
    <div className="empty-state">
      <span className="empty-state-dash">—</span>
      {text}
    </div>
  )
}
 
function PrintView({ result, printRef }) {
  const prescriptions = result.prescriptions || []
  const lab_orders    = result.lab_orders    || []
  const clinic_notes  = result.clinic_notes  || []
  const bill          = result.bill          || {}
  return (
    <div ref={printRef} className="print-only" style={{ display:'none', fontFamily:'Plus Jakarta Sans,sans-serif', padding:40 }}>
      <h1 style={{ fontSize:20, marginBottom:2 }}>ABC Health Clinic</h1>
      <p style={{ color:'#666', fontSize:12, marginBottom:20 }}>
        Date: {new Date().toLocaleDateString()} · Visit: {result.visit_id}
      </p>
      <h2 style={{ fontSize:14, borderBottom:'1px solid #eee', paddingBottom:4, marginBottom:10 }}>Prescription</h2>
      {prescriptions.map((p,i) => (
        <p key={i} style={{ fontSize:13, marginBottom:5 }}>
          <strong>{p.drug_name}</strong> {p.dosage} — {p.frequency} for {p.duration}
          {p.instructions && ` (${p.instructions})`}
        </p>
      ))}
      <h2 style={{ fontSize:14, borderBottom:'1px solid #eee', paddingBottom:4, margin:'16px 0 10px' }}>Lab Investigations</h2>
      {lab_orders.map((lo,i)   => <p key={i} style={{ fontSize:13, marginBottom:4 }}>• {lo.test_name}</p>)}
      <h2 style={{ fontSize:14, borderBottom:'1px solid #eee', paddingBottom:4, margin:'16px 0 10px' }}>Clinical Notes</h2>
      {clinic_notes.map((n,i)  => <p key={i} style={{ fontSize:13, marginBottom:4 }}><strong>{n.category}:</strong> {n.note_text}</p>)}
      <h2 style={{ fontSize:14, borderBottom:'1px solid #eee', paddingBottom:4, margin:'16px 0 10px' }}>Bill</h2>
      <table style={{ width:'100%', borderCollapse:'collapse', fontSize:13 }}>
        <tbody>
          <tr><td>Consultation</td><td style={{ textAlign:'right' }}>${(bill.consultation_fee||0).toFixed(2)}</td></tr>
          <tr><td>Drugs</td>       <td style={{ textAlign:'right' }}>${(bill.drug_total||0).toFixed(2)}</td></tr>
          <tr><td>Lab Tests</td>   <td style={{ textAlign:'right' }}>${(bill.lab_total||0).toFixed(2)}</td></tr>
          <tr style={{ fontWeight:700, borderTop:'2px solid #000' }}>
            <td style={{ paddingTop:6 }}>Grand Total</td>
            <td style={{ textAlign:'right', paddingTop:6 }}>${(bill.grand_total||0).toFixed(2)}</td>
          </tr>
        </tbody>
      </table>
    </div>
  )
}