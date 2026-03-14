import { useEffect, useState } from "react"
import axios from "axios"
import { useToast } from "../components/ui/ToastProvider"

interface SessionItem {
  session_id: string
  admin_id: string
  role: string
  email: string
  issued_at: string
  ttl_seconds: number
}

const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:3000/api"

export default function ActiveSessions() {
  const [sessions, setSessions] = useState<SessionItem[]>([])
  const [roleFilter, setRoleFilter] = useState("")
  const [emailFilter, setEmailFilter] = useState("")
  const [loading, setLoading] = useState(false)
  const [revokingSessionId, setRevokingSessionId] = useState<string | null>(null)
  const token = localStorage.getItem("admin_token")
  const { pushToast } = useToast()

  const loadSessions = async () => {
    setLoading(true)
    try {
      const params = new URLSearchParams()
      if (roleFilter) params.set("role", roleFilter)
      if (emailFilter.trim()) params.set("email", emailFilter.trim())

      const res = await axios.get(`${API_BASE}/admin/sessions?${params.toString()}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      setSessions(res.data?.data || [])
    } catch (err: any) {
      pushToast(err?.response?.data?.error || "Không thể tải danh sách phiên", "error")
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadSessions()
  }, [roleFilter])

  const revokeSession = async (sessionId: string, role: string) => {
    if (role !== "MOD") {
      pushToast("Chỉ được thu hồi phiên của MOD", "error")
      return
    }

    setRevokingSessionId(sessionId)
    try {
      await axios.delete(`${API_BASE}/admin/sessions/${sessionId}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      pushToast("Đã thu hồi phiên MOD", "success")
      await loadSessions()
    } catch (err: any) {
      pushToast(err?.response?.data?.error || "Thu hồi phiên thất bại", "error")
    } finally {
      setRevokingSessionId(null)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-gray-900">Phiên đăng nhập đang hoạt động</h2>
        <button onClick={loadSessions} className="px-4 py-2 rounded-lg border bg-white hover:bg-slate-50">
          {loading ? "Đang tải..." : "Làm mới"}
        </button>
      </div>

      <div className="bg-white border rounded-2xl shadow-sm p-4 grid grid-cols-1 sm:grid-cols-3 gap-3">
        <select value={roleFilter} onChange={(e) => setRoleFilter(e.target.value)} className="border rounded-lg px-3 py-2">
          <option value="">Tất cả vai trò</option>
          <option value="ADMIN">ADMIN</option>
          <option value="MOD">MOD</option>
        </select>
        <input
          value={emailFilter}
          onChange={(e) => setEmailFilter(e.target.value)}
          placeholder="Lọc theo email"
          className="border rounded-lg px-3 py-2"
        />
        <button onClick={loadSessions} className="px-4 py-2 rounded-lg bg-blue-600 text-white hover:bg-blue-700">
          Áp dụng
        </button>
      </div>

      <div className="bg-white border rounded-2xl shadow-sm overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 border-b text-left text-slate-600">
            <tr>
              <th className="px-4 py-3">Email</th>
              <th className="px-4 py-3">Vai trò</th>
              <th className="px-4 py-3">Bắt đầu</th>
              <th className="px-4 py-3">TTL</th>
              <th className="px-4 py-3">Session ID</th>
              <th className="px-4 py-3">Hành động</th>
            </tr>
          </thead>
          <tbody>
            {sessions.map((s) => (
              <tr key={s.session_id} className="border-b last:border-b-0">
                <td className="px-4 py-3 font-medium text-slate-900">{s.email}</td>
                <td className="px-4 py-3">{s.role}</td>
                <td className="px-4 py-3">{s.issued_at ? new Date(s.issued_at).toLocaleString() : "-"}</td>
                <td className="px-4 py-3">{Math.max(0, Math.floor((s.ttl_seconds || 0) / 60))} phút</td>
                <td className="px-4 py-3 text-xs text-slate-500">{s.session_id}</td>
                <td className="px-4 py-3">
                  <button
                    type="button"
                    disabled={s.role !== "MOD" || revokingSessionId === s.session_id}
                    onClick={() => revokeSession(s.session_id, s.role)}
                    className="inline-flex items-center justify-center px-3 py-1.5 rounded-lg border border-red-200 bg-red-50 text-red-700 hover:bg-red-100 transition disabled:opacity-60"
                  >
                    {revokingSessionId === s.session_id ? "Đang xóa..." : "Xóa phiên"}
                  </button>
                </td>
              </tr>
            ))}
            {sessions.length === 0 && (
              <tr>
                <td className="px-4 py-8 text-center text-gray-500" colSpan={6}>Không có phiên hoạt động</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
