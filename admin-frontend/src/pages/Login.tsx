import { useState } from "react"
import { useNavigate } from "react-router-dom"
import axios from "axios"
import { useToast } from "../components/ui/ToastProvider"

const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:3000/api"

export default function Login() {
  const navigate = useNavigate()
  const [email, setEmail] = useState("admin@review.com")
  const [password, setPassword] = useState("admin123")
  const [error, setError] = useState("")
  const [loading, setLoading] = useState(false)
  const { pushToast } = useToast()

  const hashPassword = async (input: string) => {
    const data = new TextEncoder().encode(input)
    const digest = await crypto.subtle.digest("SHA-256", data)
    return Array.from(new Uint8Array(digest)).map((b) => b.toString(16).padStart(2, "0")).join("")
  }

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError("")
    setLoading(true)

    try {
      const passwordHash = await hashPassword(password)
      const res = await axios.post(`${API_BASE}/admin/login`, { email, password_hash: passwordHash })
      localStorage.setItem("admin_token", res.data.token)
      localStorage.setItem("admin_user", JSON.stringify(res.data.admin))
      pushToast("Đăng nhập thành công", "success")
      navigate("/")
    } catch (err: any) {
      const msg = err?.response?.data?.error || "Đăng nhập thất bại"
      setError(msg)
      pushToast(msg, "error")
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-slate-100 flex items-center justify-center p-4">
      <form onSubmit={onSubmit} className="w-full max-w-md bg-white border rounded-2xl shadow-sm p-8 space-y-5">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Admin Login</h1>
          <p className="text-sm text-slate-500 mt-1">Đăng nhập bằng tài khoản quản trị</p>
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium text-slate-700">Email</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full border rounded-lg px-3 py-2 outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium text-slate-700">Mật khẩu</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full border rounded-lg px-3 py-2 outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
        </div>

        {error ? <p className="text-sm text-red-600">{error}</p> : null}

        <button
          type="submit"
          disabled={loading}
          className="w-full bg-blue-600 text-white rounded-lg px-4 py-2 font-medium hover:bg-blue-700 disabled:opacity-60"
        >
          {loading ? "Đang đăng nhập..." : "Đăng nhập"}
        </button>
      </form>
    </div>
  )
}
