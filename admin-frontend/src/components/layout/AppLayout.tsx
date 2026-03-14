import { Outlet, Link, useNavigate } from "react-router-dom"
import { Building2, MessageSquare, LayoutDashboard, Settings, LogOut } from "lucide-react"

export default function AppLayout() {
  const navigate = useNavigate()
  const adminName = (() => {
    try {
      const raw = localStorage.getItem("admin_user")
      if (!raw) return "Admin User"
      const parsed = JSON.parse(raw)
      return parsed?.name || parsed?.email || "Admin User"
    } catch {
      return "Admin User"
    }
  })()

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
            Admin Panel
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
          <h2 className="font-semibold text-gray-800">Review Công Ty - Quản Trị Hệ Thống</h2>
          <div className="flex items-center gap-3">
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
