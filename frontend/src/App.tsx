import { useEffect, useState } from "react"
import { Outlet, useLocation } from "react-router-dom"
import axios from "axios"
import Navbar from "./components/layout/Navbar"

const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:3000/api"

function App() {
  const location = useLocation()
  const [sessionId] = useState(() => {
    const key = "visit_session_id"
    let value = localStorage.getItem(key)
    if (!value) {
      value = crypto?.randomUUID?.() || `${Date.now()}-${Math.random()}`
      localStorage.setItem(key, value)
    }
    return value
  })

  useEffect(() => {
    axios.post(
      `${API_BASE}/analytics/visits`,
      { path: location.pathname },
      { headers: { "X-Session-ID": sessionId } }
    ).catch(() => {
      // ignore tracking error in UI
    })
  }, [location.pathname, sessionId])

  return (
    <div className="min-h-screen bg-slate-50 font-sans">
      <Navbar />
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Outlet />
      </main>
    </div>
  )
}

export default App
