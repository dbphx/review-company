import { useEffect, useState } from "react"
import axios from "axios"
import { Link } from "react-router-dom"

interface CompanyInfo {
  id: string
  name: string
}

interface ReviewItem {
  id: string
  title: string
  content: string
  rating: number
  author_name: string
  created_at: string
  company?: CompanyInfo
}

const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:3000/api"

export default function Reviews() {
  const [reviews, setReviews] = useState<ReviewItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [limit] = useState(10)
  const [companyQuery, setCompanyQuery] = useState("")

  const adminToken = localStorage.getItem("admin_token")

  const loadReviews = async () => {
    const query = companyQuery.trim()
    const companyParam = query ? `&company=${encodeURIComponent(query)}` : ""
    const res = await axios.get(`${API_BASE}/reviews?page=${page}&limit=${limit}${companyParam}`)
    setReviews(res.data.data || [])
    setTotal(res.data.total || 0)
  }

  useEffect(() => {
    loadReviews()
  }, [page, limit, companyQuery])

  useEffect(() => {
    if (page > 1 && reviews.length === 0) {
      setPage(1)
    }
  }, [reviews.length, page])

  const totalPages = Math.max(1, Math.ceil(total / limit))

  const deleteReview = async (reviewId: string) => {
    if (!window.confirm("Bạn chắc chắn muốn xóa review này?")) return
    await axios.delete(`${API_BASE}/reviews/${reviewId}`, {
      headers: { Authorization: `Bearer ${adminToken}` },
    })
    await loadReviews()
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-gray-900">Quản lý Review</h2>
        <div className="text-sm text-gray-600">Tổng số: <span className="font-semibold">{total}</span></div>
      </div>

      <div className="bg-white border rounded-2xl shadow-sm p-4">
        <input
          value={companyQuery}
          onChange={(e) => {
            setCompanyQuery(e.target.value)
            setPage(1)
          }}
          placeholder="Tìm review theo tên công ty"
          className="w-full border rounded-lg px-3 py-2"
        />
      </div>

      <div className="bg-white border rounded-2xl shadow-sm overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 border-b text-left text-slate-600">
            <tr>
              <th className="px-4 py-3">Review</th>
              <th className="px-4 py-3">Công ty</th>
              <th className="px-4 py-3">Rating</th>
              <th className="px-4 py-3">Thời gian</th>
              <th className="px-4 py-3">Hành động</th>
            </tr>
          </thead>
          <tbody>
            {reviews.map((r) => (
              <tr key={r.id} className="border-b last:border-b-0">
                <td className="px-4 py-3">
                  <div className="font-medium text-slate-900">{r.title}</div>
                  <div className="text-slate-500 line-clamp-1">{r.content}</div>
                </td>
                <td className="px-4 py-3">{r.company?.name || "N/A"}</td>
                <td className="px-4 py-3">{r.rating}</td>
                <td className="px-4 py-3">{new Date(r.created_at).toLocaleString()}</td>
                <td className="px-4 py-3 space-x-3">
                  <Link className="text-indigo-600 hover:underline" to={`/reviews/${r.id}`}>Quản lý</Link>
                  <a className="text-blue-600 hover:underline" href={`http://localhost:5173/company/${r.company?.id}`} target="_blank" rel="noreferrer">
                    Portal
                  </a>
                  <button className="text-red-600 hover:underline" onClick={() => deleteReview(r.id)}>
                    Xóa
                  </button>
                </td>
              </tr>
            ))}
            {reviews.length === 0 && (
              <tr>
                <td className="px-4 py-8 text-center text-gray-500" colSpan={5}>Không có dữ liệu review</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <div className="flex items-center justify-between">
        <button
          className="px-3 py-2 border rounded-lg disabled:opacity-50"
          onClick={() => setPage((p) => Math.max(1, p - 1))}
          disabled={page <= 1}
        >
          Trang trước
        </button>
        <span className="text-sm text-gray-600">Trang {page} / {totalPages}</span>
        <button
          className="px-3 py-2 border rounded-lg disabled:opacity-50"
          onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
          disabled={page >= totalPages}
        >
          Trang sau
        </button>
      </div>
    </div>
  )
}
