import { Outlet, Link, useNavigate } from "react-router-dom"
import { Building2, MessageSquare, LayoutDashboard, Settings, LogOut, Users, Activity } from "lucide-react"
import { useEffect, useState } from "react"
import axios from "axios"
import { useToast } from "../ui/ToastProvider"

const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:3000/api"

export default function AppLayout() {
  const navigate = useNavigate()
  const token = localStorage.getItem("admin_token")
  const [dataMode, setDataMode] = useState("v1")
  const { pushToast } = useToast()
  const adminName = (() => {
    try {
      const raw = localStorage.getItem("admin_user")
      if (!raw) return "Quản trị viên"
      const parsed = JSON.parse(raw)
      return parsed?.name || parsed?.email || "Quản trị viên"
    } catch {
      return "Quản trị viên"
    }
  })()
  const adminRole = (() => {
    try {
      const raw = localStorage.getItem("admin_user")
      if (!raw) return "ADMIN"
      const parsed = JSON.parse(raw)
      return parsed?.role || "ADMIN"
    } catch {
      return "ADMIN"
    }
  })()

  useEffect(() => {
    if (!token) return
    axios.get(`${API_BASE}/admin/data-mode`, { headers: { Authorization: `Bearer ${token}` } })
      .then((res) => setDataMode(res.data?.mode || "v1"))
      .catch(() => {})
  }, [token])

  const onChangeDataMode = async (mode: string) => {
    const prev = dataMode
    setDataMode(mode)
    if (adminRole !== "ADMIN") {
      setDataMode(prev)
      pushToast("Chỉ ADMIN mới được đổi mode dữ liệu", "error")
      return
    }
    try {
      const res = await axios.post(`${API_BASE}/admin/data-mode`, { mode }, { headers: { Authorization: `Bearer ${token}` } })
      setDataMode(res.data?.mode || mode)
      pushToast("Đã cập nhật mode dữ liệu", "success")
    } catch (err: any) {
      setDataMode(prev)
      pushToast(err?.response?.data?.error || "Không thể cập nhật mode dữ liệu", "error")
    }
  }

  const handleLogout = () => {
    localStorage.removeItem("admin_token")
    localStorage.removeItem("admin_user")
    navigate("/login")
  }

  return (
    <div className="min-h-screen bg-slate-100 font-sans flex">
      {/* Sidebar */}
      <aside className="w-64 bg-slate-900 text-white flex flex-col h-screen sticky top-0">
        <div className="p-6">
          <h1 className="text-xl font-bold tracking-tight text-white flex items-center gap-2">
            <Settings className="w-5 h-5 text-blue-400" />
            Bảng quản trị
          </h1>
        </div>
        
        <nav className="flex-1 px-4 space-y-2 mt-4">
          <Link to="/" className="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-slate-800 transition text-sm font-medium text-slate-300 hover:text-white">
            <LayoutDashboard className="w-4 h-4" /> Tổng quan
          </Link>
          <Link to="/reviews" className="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-slate-800 transition text-sm font-medium text-slate-300 hover:text-white">
            <MessageSquare className="w-4 h-4" /> Quản lý Reviews
          </Link>
          <Link to="/companies" className="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-slate-800 transition text-sm font-medium text-slate-300 hover:text-white">
            <Building2 className="w-4 h-4" /> Quản lý Công ty
          </Link>
          <Link to="/admin-users" className="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-slate-800 transition text-sm font-medium text-slate-300 hover:text-white">
            <Users className="w-4 h-4" /> Quản lý Admin
          </Link>
          <Link to="/active-sessions" className="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-slate-800 transition text-sm font-medium text-slate-300 hover:text-white">
            <Activity className="w-4 h-4" /> Phiên hoạt động
          </Link>
        </nav>

        <div className="p-4 border-t border-slate-800">
          <button onClick={handleLogout} className="flex items-center gap-3 w-full px-3 py-2 rounded-lg hover:bg-red-500/10 hover:text-red-400 transition text-sm font-medium text-slate-400">
            <LogOut className="w-4 h-4" /> Đăng xuất
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-x-hidden">
        <header className="h-16 bg-white border-b px-8 flex items-center justify-between sticky top-0 z-10">
          <h2 className="font-semibold text-gray-800">ReviewCT - Quản trị hệ thống</h2>
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2">
              <span className="text-xs text-gray-500">Mode dữ liệu</span>
              <select
                value={dataMode}
                onChange={(e) => onChangeDataMode(e.target.value)}
                className="border rounded-lg px-2 py-1 text-sm"
                disabled={adminRole !== "ADMIN"}
              >
                <option value="v1">V1</option>
                <option value="v2">V2</option>
                <option value="all">ALL</option>
              </select>
            </div>
            <span className="text-sm font-medium">{adminName}</span>
            <div className="w-8 h-8 rounded-full bg-blue-600 text-white flex items-center justify-center font-bold text-sm">{adminName.charAt(0).toUpperCase()}</div>
          </div>
        </header>
        <div className="p-8">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
