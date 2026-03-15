import { useEffect, useState } from "react"
import axios from "axios"
import { useToast } from "../components/ui/ToastProvider"
import ActionMenu from "../components/ui/ActionMenu"

interface AdminUser {
  id: string
  email: string
  name: string
  role: string
  created_at: string
}

interface AdminSessionUser {
  id?: string
  role?: string
}

const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:3000/api"

async function hashPassword(input: string) {
  const data = new TextEncoder().encode(input)
  const digest = await crypto.subtle.digest("SHA-256", data)
  return Array.from(new Uint8Array(digest)).map((b) => b.toString(16).padStart(2, "0")).join("")
}

export default function AdminUsers() {
  const [users, setUsers] = useState<AdminUser[]>([])
  const [name, setName] = useState("")
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [role, setRole] = useState("MOD")
  const [openCreateModal, setOpenCreateModal] = useState(false)
  const [saving, setSaving] = useState(false)
  const [deletingId, setDeletingId] = useState<string | null>(null)
  const [revokingId, setRevokingId] = useState<string | null>(null)
  const token = localStorage.getItem("admin_token")
  const currentAdmin: AdminSessionUser = (() => {
    try {
      const raw = localStorage.getItem("admin_user")
      if (!raw) return {}
      return JSON.parse(raw)
    } catch {
      return {}
    }
  })()
  const isAdminRole = currentAdmin.role === "ADMIN"
  const { pushToast } = useToast()

  const loadUsers = async () => {
    const res = await axios.get(`${API_BASE}/admin/users`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    setUsers(res.data?.data || [])
  }

  useEffect(() => {
    loadUsers()
  }, [])

  const createUser = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!isAdminRole) {
      pushToast("Chỉ ADMIN mới được tạo tài khoản quản trị", "error")
      return
    }
    if (!name.trim() || !email.trim() || !password.trim()) {
      pushToast("Vui lòng nhập đủ tên, email, mật khẩu", "error")
      return
    }

    setSaving(true)
    try {
      const passwordHash = await hashPassword(password)
      await axios.post(
        `${API_BASE}/admin/users`,
        { name: name.trim(), email: email.trim(), password_hash: passwordHash, role },
        { headers: { Authorization: `Bearer ${token}` } }
      )
      setName("")
      setEmail("")
      setPassword("")
      setRole("MOD")
      setOpenCreateModal(false)
      pushToast("Tạo tài khoản admin thành công", "success")
      await loadUsers()
    } catch (err: any) {
      pushToast(err?.response?.data?.error || "Tạo tài khoản admin thất bại", "error")
    } finally {
      setSaving(false)
    }
  }

  const deleteUser = async (id: string) => {
    try {
      await axios.delete(`${API_BASE}/admin/users/${id}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      pushToast("Đã xóa tài khoản admin", "success")
      await loadUsers()
    } catch (err: any) {
      pushToast(err?.response?.data?.error || "Xóa tài khoản admin thất bại", "error")
    } finally {
      setDeletingId(null)
    }
  }

  const revokeModSessions = async (id: string) => {
    setRevokingId(id)
    try {
      const res = await axios.delete(`${API_BASE}/admin/users/${id}/sessions`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      const revoked = Number(res.data?.revoked_sessions || 0)
      pushToast(`Đã thu hồi ${revoked} phiên đăng nhập của MOD`, "success")
    } catch (err: any) {
      pushToast(err?.response?.data?.error || "Thu hồi phiên đăng nhập thất bại", "error")
    } finally {
      setRevokingId(null)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-gray-900">Quản lý tài khoản Admin</h2>
        <div className="flex items-center gap-3">
          <div className="text-sm text-gray-600">Tổng số: <span className="font-semibold">{users.length}</span></div>
          <button
            type="button"
            disabled={!isAdminRole}
            onClick={() => setOpenCreateModal(true)}
            className="bg-blue-600 text-white px-4 py-2 rounded-lg font-medium hover:bg-blue-700 disabled:opacity-60"
          >
            + Thêm tài khoản
          </button>
        </div>
      </div>

      <div className="bg-white border rounded-2xl shadow-sm overflow-x-auto overflow-y-visible">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 border-b text-left text-slate-600">
            <tr>
              <th className="px-4 py-3">Tên</th>
              <th className="px-4 py-3">Email</th>
              <th className="px-4 py-3">Vai trò</th>
              <th className="px-4 py-3">Ngày tạo</th>
              <th className="px-4 py-3">Hành động</th>
            </tr>
          </thead>
          <tbody>
            {users.map((u) => (
              <tr key={u.id} className="border-b last:border-b-0">
                <td className="px-4 py-3 font-medium text-slate-900">{u.name}</td>
                <td className="px-4 py-3">{u.email}</td>
                <td className="px-4 py-3">{u.role}</td>
                <td className="px-4 py-3">{new Date(u.created_at).toLocaleString()}</td>
                <td className="px-4 py-3">
                  {isAdminRole ? (
                    <ActionMenu menuClassName="min-w-[170px]">
                        {u.role === "MOD" && (
                          <button
                            className="block w-full text-left px-3 py-2 text-sm rounded-md text-amber-700 hover:bg-amber-50 disabled:opacity-60"
                            onClick={() => revokeModSessions(u.id)}
                            disabled={revokingId === u.id}
                          >
                            {revokingId === u.id ? "Đang thu hồi..." : "Thu hồi phiên"}
                          </button>
                        )}
                        <button
                          className="block w-full text-left px-3 py-2 text-sm rounded-md text-rose-700 hover:bg-rose-50 disabled:opacity-60"
                          onClick={() => setDeletingId(u.id)}
                          disabled={u.id === currentAdmin.id}
                        >
                          {u.id === currentAdmin.id ? "Đang đăng nhập" : "Xóa"}
                        </button>
                    </ActionMenu>
                  ) : (
                    <span className="text-slate-400 text-xs">Không có quyền</span>
                  )}
                </td>
              </tr>
            ))}
            {users.length === 0 && (
              <tr>
                <td className="px-4 py-8 text-center text-gray-500" colSpan={5}>Chưa có tài khoản admin</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {deletingId && (
        <div className="fixed inset-0 bg-slate-900/40 z-50 flex items-start sm:items-center justify-center p-4 sm:p-6 overflow-y-auto">
          <div className="w-full max-w-md bg-white rounded-2xl border shadow-xl p-6 space-y-4 max-h-[90vh] overflow-y-auto">
            <h3 className="text-lg font-bold text-slate-900">Xác nhận xóa tài khoản admin</h3>
            <p className="text-sm text-slate-600">Bạn có chắc muốn xóa tài khoản này?</p>
            <div className="flex justify-end gap-3 pt-2">
              <button type="button" onClick={() => setDeletingId(null)} className="px-4 py-2 rounded-lg border">Hủy</button>
              <button type="button" onClick={() => deleteUser(deletingId)} className="px-4 py-2 rounded-lg bg-red-600 text-white hover:bg-red-700">Xóa</button>
            </div>
          </div>
        </div>
      )}

      {openCreateModal && (
        <div className="fixed inset-0 bg-slate-900/40 z-50 flex items-start sm:items-center justify-center p-4 sm:p-6 overflow-y-auto">
          <form onSubmit={createUser} className="w-full max-w-xl bg-white rounded-2xl border shadow-xl p-6 space-y-4 max-h-[90vh] overflow-y-auto">
            <h3 className="text-lg font-bold text-slate-900">Thêm tài khoản quản trị</h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Tên hiển thị" className="border rounded-lg px-3 py-2" disabled={!isAdminRole} />
              <input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="Email" className="border rounded-lg px-3 py-2" type="email" disabled={!isAdminRole} />
              <input value={password} onChange={(e) => setPassword(e.target.value)} placeholder="Mật khẩu" className="border rounded-lg px-3 py-2" type="password" disabled={!isAdminRole} />
              <select value={role} onChange={(e) => setRole(e.target.value)} className="border rounded-lg px-3 py-2" disabled={!isAdminRole}>
                <option value="MOD">MOD</option>
                <option value="ADMIN">ADMIN</option>
              </select>
            </div>
            <div className="flex justify-end gap-3 pt-2">
              <button
                type="button"
                onClick={() => setOpenCreateModal(false)}
                className="px-4 py-2 rounded-lg border"
              >
                Hủy
              </button>
              <button
                type="submit"
                disabled={saving || !isAdminRole}
                className="px-4 py-2 rounded-lg bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-60"
              >
                {saving ? "Đang tạo..." : "Tạo tài khoản"}
              </button>
            </div>
          </form>
        </div>
      )}

      {!isAdminRole && (
        <div className="bg-amber-50 border border-amber-200 rounded-xl p-3 text-sm text-amber-800">
          Bạn đang đăng nhập với vai trò MOD. MOD có thể xem danh sách quản trị nhưng không thể tạo/xóa tài khoản admin.
        </div>
      )}
    </div>
  )
}
