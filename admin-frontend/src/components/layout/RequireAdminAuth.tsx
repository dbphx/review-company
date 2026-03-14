import { Navigate, Outlet } from "react-router-dom"

export default function RequireAdminAuth() {
  const token = localStorage.getItem("admin_token")
  if (!token) {
    return <Navigate to="/login" replace />
  }
  return <Outlet />
}
