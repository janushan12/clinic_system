import { useState } from 'react'
import NewVisitPage from './pages/NewVisitPage'
import VisitResultPage from './pages/VisitResultPage'
import './index.css'
import './styles/layout.css'

export default function App() {
  const [page, setPage]               = useState('new')
  const [visitResult, setVisitResult] = useState(null)

  const handleParsed   = (result) => { setVisitResult(result); setPage('result') }
  const handleNewVisit = ()       => { setVisitResult(null);   setPage('new')    }

  return (
    <div className="app-wrapper">
      <Sidebar activePage={page} onNewVisit={handleNewVisit} />
      <div className="app-content">
        <Topbar page={page} onNew={handleNewVisit} />
        <main className="main-content">
          {page === 'new'
            ? <NewVisitPage onParsed={handleParsed} />
            : <VisitResultPage result={visitResult} onNew={handleNewVisit} />
          }
        </main>
      </div>
    </div>
  )
}

function Sidebar({ activePage, onNewVisit }) {
  return (
    <aside className="sidebar no-print">
      <div className="sidebar-logo">
        <div className="sidebar-logo-icon">🏥</div>
        <div>
          <div className="sidebar-logo-name">ABC Health</div>
          <div className="sidebar-logo-sub">Clinic System</div>
        </div>
      </div>

      <nav className="sidebar-nav">
        <div className="nav-group">
          <div className="nav-label">Main</div>
          <NavItem icon="✦" label="New Visit"     active={activePage === 'new'}    onClick={onNewVisit} />
          <NavItem icon="📋" label="Visit Results" active={activePage === 'result'} />
        </div>
        <div className="nav-group">
          <div className="nav-label">Records</div>
          <NavItem icon="👤" label="Patients"      />
          <NavItem icon="🧪" label="Lab Orders"    />
          <NavItem icon="💊" label="Prescriptions" />
          <NavItem icon="💰" label="Billing"       />
        </div>
      </nav>

      <div className="sidebar-footer">
        <div className="sidebar-avatar">Dr</div>
        <div>
          <div className="sidebar-user-name">Doctor Portal</div>
          <div className="sidebar-user-sub">v1.0</div>
        </div>
      </div>
    </aside>
  )
}

function NavItem({ icon, label, active, onClick }) {
  return (
    <button className={`nav-item ${active ? 'active' : ''}`} onClick={onClick}>
      <span className="nav-item-icon">{icon}</span>
      {label}
    </button>
  )
}

function Topbar({ page, onNew }) {
  const titles    = { new: 'New Visit',      result: 'Visit Summary'     }
  const subtitles = {
    new:    'Enter clinical notes — AI will classify and bill automatically',
    result: 'AI classification complete — review, print or mark as paid',
  }
  return (
    <header className="topbar no-print">
      <div>
        <div className="topbar-title">{titles[page]}</div>
        <div className="topbar-sub">{subtitles[page]}</div>
      </div>
      <div className="topbar-right">
        <div className="stat-pill">
          🕐 {new Date().toLocaleDateString('en-US', { weekday:'short', month:'short', day:'numeric' })}
        </div>
        <button className="btn-primary" onClick={onNew}>
          <span style={{ fontSize: 16, lineHeight: 1 }}>+</span> New Visit
        </button>
      </div>
    </header>
  )
}