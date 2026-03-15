import { useEffect, useState } from "react"
import { Navigate, Outlet, useNavigate } from "react-router-dom"
import axios from "axios"

const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:3000/api"

export default function RequireAdminAuth() {
  const navigate = useNavigate()
  const token = localStorage.getItem("admin_token")
  const [checking, setChecking] = useState(true)
  const [isValid, setIsValid] = useState(false)

  useEffect(() => {
    let active = true
    if (!token) {
      setIsValid(false)
      setChecking(false)
      return
    }

    axios
      .get(`${API_BASE}/admin/data-mode`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      .then(() => {
        if (!active) return
        setIsValid(true)
      })
      .catch(() => {
        if (!active) return
        localStorage.removeItem("admin_token")
        localStorage.removeItem("admin_user")
        setIsValid(false)
        navigate("/login", { replace: true })
      })
      .finally(() => {
        if (!active) return
        setChecking(false)
      })

    return () => {
      active = false
    }
  }, [token, navigate])

  if (checking) {
    return <div className="p-6 text-sm text-gray-500">Đang kiểm tra phiên đăng nhập...</div>
  }

  if (!token) {
    return <Navigate to="/login" replace />
  }

  if (!isValid) {
    return <Navigate to="/login" replace />
  }

  return <Outlet />
}
