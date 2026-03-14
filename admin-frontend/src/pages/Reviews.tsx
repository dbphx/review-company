import { useEffect, useState } from "react"
import axios from "axios"
import { Link } from "react-router-dom"
import { useToast } from "../components/ui/ToastProvider"

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
  const [createdDate, setCreatedDate] = useState("")
  const [seedVersion, setSeedVersion] = useState("")
  const [confirmDeleteReviewId, setConfirmDeleteReviewId] = useState<string | null>(null)
  const { pushToast } = useToast()

  const adminToken = localStorage.getItem("admin_token")

  const loadReviews = async () => {
    const query = companyQuery.trim()
    const companyParam = query ? `&company=${encodeURIComponent(query)}` : ""
    const createdDateParam = createdDate ? `&created_date=${encodeURIComponent(createdDate)}` : ""
    const seedVersionParam = seedVersion && seedVersion !== "all" ? `&seed_version=${encodeURIComponent(seedVersion)}` : ""
    const res = await axios.get(`${API_BASE}/reviews?page=${page}&limit=${limit}${companyParam}${createdDateParam}${seedVersionParam}`)
    setReviews(res.data.data || [])
    setTotal(res.data.total || 0)
  }

  useEffect(() => {
    loadReviews()
  }, [page, limit, companyQuery, createdDate, seedVersion])

  useEffect(() => {
    if (page > 1 && reviews.length === 0) {
      setPage(1)
    }
  }, [reviews.length, page])

  const totalPages = Math.max(1, Math.ceil(total / limit))

  const deleteReview = async (reviewId: string) => {
    try {
      await axios.delete(`${API_BASE}/reviews/${reviewId}`, {
        headers: { Authorization: `Bearer ${adminToken}` },
      })
      await loadReviews()
      pushToast("Đã xóa review", "success")
    } catch (err: any) {
      pushToast(err?.response?.data?.error || "Xóa review thất bại", "error")
    } finally {
      setConfirmDeleteReviewId(null)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-gray-900">Quản lý Review</h2>
        <div className="text-sm text-gray-600">Tổng số: <span className="font-semibold">{total}</span></div>
      </div>

      <div className="bg-white border rounded-2xl shadow-sm p-4 grid grid-cols-1 sm:grid-cols-2 gap-3">
        <input
          value={companyQuery}
          onChange={(e) => {
            setCompanyQuery(e.target.value)
            setPage(1)
          }}
          placeholder="Tìm review theo tên công ty"
          className="w-full border rounded-lg px-3 py-2"
        />
        <div className="flex items-center gap-2">
          <input
            type="date"
            value={createdDate}
            onChange={(e) => {
              setCreatedDate(e.target.value)
              setPage(1)
            }}
            className="w-full border rounded-lg px-3 py-2"
          />
          {createdDate && (
            <button
              type="button"
              onClick={() => {
                setCreatedDate("")
                setPage(1)
              }}
              className="px-3 py-2 border rounded-lg text-sm"
            >
              Xóa
            </button>
          )}
        </div>
        <select
          value={seedVersion}
          onChange={(e) => {
            setSeedVersion(e.target.value)
            setPage(1)
          }}
          className="w-full border rounded-lg px-3 py-2"
        >
          <option value="">Theo mode hệ thống</option>
          <option value="v2">Dữ liệu mới (v2)</option>
          <option value="v1">Dữ liệu khởi tạo (v1)</option>
          <option value="all">Tất cả dữ liệu</option>
        </select>
        <div className="sm:col-span-2 text-xs text-gray-500">
          Chọn đúng ngày để lọc review theo ngày tạo, kết hợp thêm version dữ liệu V1/live.
        </div>
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
                <td className="px-4 py-3">{r.company?.name || "Không có"}</td>
                <td className="px-4 py-3">{r.rating}</td>
                <td className="px-4 py-3">{new Date(r.created_at).toLocaleString()}</td>
                <td className="px-4 py-3">
                  <div className="flex flex-wrap items-center gap-2">
                    <Link
                      className="inline-flex items-center justify-center px-3 py-1.5 rounded-lg border border-indigo-200 bg-indigo-50 text-indigo-700 hover:bg-indigo-100 transition"
                      to={`/reviews/${r.id}`}
                    >
                      Quản lý
                    </Link>
                    <a
                      className="inline-flex items-center justify-center px-3 py-1.5 rounded-lg border border-blue-200 bg-blue-50 text-blue-700 hover:bg-blue-100 transition"
                      href={`http://localhost:5173/company/${r.company?.id}`}
                      target="_blank"
                      rel="noreferrer"
                    >
                    Portal
                    </a>
                    <button
                      className="inline-flex items-center justify-center px-3 py-1.5 rounded-lg border border-red-200 bg-red-50 text-red-700 hover:bg-red-100 transition"
                      onClick={() => setConfirmDeleteReviewId(r.id)}
                    >
                      Xóa
                    </button>
                  </div>
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

      {confirmDeleteReviewId && (
        <div className="fixed inset-0 bg-slate-900/40 z-50 flex items-center justify-center p-4">
          <div className="w-full max-w-md bg-white rounded-2xl border shadow-xl p-6 space-y-4">
            <h3 className="text-lg font-bold text-slate-900">Xác nhận xóa review</h3>
            <p className="text-sm text-slate-600">Bạn có chắc muốn xóa review này?</p>
            <div className="flex justify-end gap-3 pt-2">
              <button type="button" onClick={() => setConfirmDeleteReviewId(null)} className="px-4 py-2 rounded-lg border">
                Hủy
              </button>
              <button
                type="button"
                onClick={() => deleteReview(confirmDeleteReviewId)}
                className="px-4 py-2 rounded-lg bg-red-600 text-white hover:bg-red-700"
              >
                Xóa
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
