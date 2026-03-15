import { useEffect, useState } from "react"
import axios from "axios"
import { Link } from "react-router-dom"
import { useToast } from "../components/ui/ToastProvider"
import ActionMenu from "../components/ui/ActionMenu"

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
    const res = await axios.get(`${API_BASE}/admin/reviews?page=${page}&limit=${limit}${companyParam}${createdDateParam}${seedVersionParam}`, {
      headers: { Authorization: `Bearer ${adminToken}` },
    })
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

      <div className="bg-white border rounded-2xl shadow-sm overflow-x-auto">
        <table className="w-full text-sm table-fixed">
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
                <td className="px-4 py-3 align-top w-[40%]">
                  <div className="font-medium text-slate-900 break-words [overflow-wrap:anywhere]">{r.title}</div>
                  <div className="text-slate-500 line-clamp-2 break-words [overflow-wrap:anywhere]">{r.content}</div>
                </td>
                <td className="px-4 py-3 align-top w-[18%] break-words [overflow-wrap:anywhere]">{r.company?.name || "Không có"}</td>
                <td className="px-4 py-3 align-top w-[8%]">{r.rating}</td>
                <td className="px-4 py-3 align-top w-[16%] whitespace-nowrap">{new Date(r.created_at).toLocaleString()}</td>
                <td className="px-4 py-3 align-top w-[18%]">
                  <ActionMenu menuClassName="min-w-[150px]" placement="top">
                      <Link
                        className="block w-full text-left px-3 py-2 text-sm rounded-md text-indigo-700 hover:bg-indigo-50"
                        to={`/reviews/${r.id}`}
                      >
                        Quản lý
                      </Link>
                      <a
                        className="block w-full text-left px-3 py-2 text-sm rounded-md text-blue-700 hover:bg-blue-50"
                        href={`http://localhost:5173/company/${r.company?.id}`}
                        target="_blank"
                        rel="noreferrer"
                      >
                        Portal
                      </a>
                      <button
                        className="block w-full text-left px-3 py-2 text-sm rounded-md text-rose-700 hover:bg-rose-50"
                        onClick={() => setConfirmDeleteReviewId(r.id)}
                      >
                        Xóa
                      </button>
                  </ActionMenu>
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
        <div className="fixed inset-0 bg-slate-900/40 z-50 flex items-start sm:items-center justify-center p-4 sm:p-6 overflow-y-auto">
          <div className="w-full max-w-md bg-white rounded-2xl border shadow-xl p-6 space-y-4 max-h-[90vh] overflow-y-auto">
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
