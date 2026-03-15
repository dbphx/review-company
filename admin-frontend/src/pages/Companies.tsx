import { useEffect, useState } from "react"
import axios from "axios"
import { useToast } from "../components/ui/ToastProvider"
import ActionMenu from "../components/ui/ActionMenu"

interface CompanyItem {
  id: string
  name: string
  logo_url: string
  website: string
  industry: string
  size: string
  description: string
  total_reviews: number
  avg_rating: number
}

interface CompanyPayload {
  name: string
  logo_url: string
  website: string
  industry: string
  size: string
  description: string
}

interface CompanyRequestItem {
  id: string
  name: string
  logo_url: string
  website: string
  industry: string
  size: string
  description: string
  status: "PENDING" | "APPROVED" | "REJECTED"
  created_at: string
}

const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:3000/api"

const emptyPayload: CompanyPayload = {
  name: "",
  logo_url: "",
  website: "",
  industry: "",
  size: "",
  description: "",
}

const fallbackLogo = (name: string) =>
  `https://ui-avatars.com/api/?name=${encodeURIComponent(name || "Company")}&background=0F766E&color=FFFFFF&bold=true`

const normalizeCompanyName = (name: string) =>
  (name || "").toLowerCase().trim()

const resolveLogoUrl = (name: string, logoUrl: string) => {
  const normalized = normalizeCompanyName(name)
  const isFptSoftware =
    normalized.includes("fpt software") ||
    normalized.includes("fsoft") ||
    (normalized.includes("fpt") && normalized.includes("software"))

  if (isFptSoftware) {
    return "https://upload.wikimedia.org/wikipedia/commons/1/11/FPT_logo_2010.svg"
  }

  if (normalized.includes("tma solutions")) {
    return "https://www.tma.vn/logo-menu.webp"
  }

  return logoUrl || fallbackLogo(name)
}

export default function Companies() {
  const [companies, setCompanies] = useState<CompanyItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [limit] = useState(10)
  const [query, setQuery] = useState("")
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<CompanyItem | null>(null)
  const [payload, setPayload] = useState<CompanyPayload>(emptyPayload)
  const [error, setError] = useState("")
  const [saving, setSaving] = useState(false)
  const [confirmDeleteCompany, setConfirmDeleteCompany] = useState<CompanyItem | null>(null)
  const [companyRequests, setCompanyRequests] = useState<CompanyRequestItem[]>([])
  const [requestsLoading, setRequestsLoading] = useState(false)
  const [approvingRequest, setApprovingRequest] = useState<CompanyRequestItem | null>(null)
  const [approvePayload, setApprovePayload] = useState<CompanyPayload>(emptyPayload)
  const [approveError, setApproveError] = useState("")
  const [approving, setApproving] = useState(false)
  const { pushToast } = useToast()

  const adminToken = localStorage.getItem("admin_token")

  const loadCompanies = async () => {
    const q = query.trim()
    if (q) {
      const res = await axios.get(`${API_BASE}/admin/search?q=${encodeURIComponent(q)}`, {
        headers: { Authorization: `Bearer ${adminToken}` },
      })
      const rows: CompanyItem[] = res.data || []
      setCompanies(rows)
      setTotal(rows.length)
      return
    }

    const res = await axios.get(`${API_BASE}/admin/companies/top?page=${page}&limit=${limit}`, {
      headers: { Authorization: `Bearer ${adminToken}` },
    })
    const rows: CompanyItem[] = res.data?.data || []
    setCompanies(rows)
    setTotal(res.data?.total || 0)
  }

  const loadCompanyRequests = async () => {
    setRequestsLoading(true)
    try {
      const res = await axios.get(`${API_BASE}/admin/company-requests?status=PENDING`, {
        headers: { Authorization: `Bearer ${adminToken}` },
      })
      setCompanyRequests(res.data?.data || [])
    } catch {
      setCompanyRequests([])
    } finally {
      setRequestsLoading(false)
    }
  }

  const reloadCompanyViews = async () => {
    await Promise.all([loadCompanies(), loadCompanyRequests()])
  }

  useEffect(() => {
    loadCompanies()
  }, [page, limit, query])

  useEffect(() => {
    loadCompanyRequests()
  }, [])

  useEffect(() => {
    if (page > 1 && companies.length === 0) {
      setPage(1)
    }
  }, [companies.length, page])

  const openCreate = () => {
    setEditing(null)
    setPayload(emptyPayload)
    setError("")
    setFormOpen(true)
  }

  const openEdit = (company: CompanyItem) => {
    setEditing(company)
    setPayload({
      name: company.name || "",
      logo_url: company.logo_url || "",
      website: company.website || "",
      industry: company.industry || "",
      size: company.size || "",
      description: company.description || "",
    })
    setError("")
    setFormOpen(true)
  }

  const closeForm = () => {
    setFormOpen(false)
    setEditing(null)
    setPayload(emptyPayload)
    setError("")
  }

  const saveCompany = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!payload.name.trim()) {
      setError("Tên công ty là bắt buộc")
      return
    }

    setSaving(true)
    setError("")
    try {
      if (editing) {
        await axios.put(`${API_BASE}/companies/${editing.id}`, payload, {
          headers: { Authorization: `Bearer ${adminToken}` },
        })
        pushToast("Cập nhật công ty thành công", "success")
      } else {
        await axios.post(`${API_BASE}/companies`, payload, {
          headers: { Authorization: `Bearer ${adminToken}` },
        })
        pushToast("Thêm công ty thành công", "success")
        if (query.trim()) {
          setQuery("")
        }
        if (page !== 1) {
          setPage(1)
        }
      }
      closeForm()
      await reloadCompanyViews()
    } catch (err: any) {
      const msg = err?.response?.data?.error || "Lưu công ty thất bại"
      setError(msg)
      pushToast(msg, "error")
    } finally {
      setSaving(false)
    }
  }

  const deleteCompany = async (company: CompanyItem) => {
    try {
      await axios.delete(`${API_BASE}/companies/${company.id}`, {
        headers: { Authorization: `Bearer ${adminToken}` },
      })
      await reloadCompanyViews()
      pushToast(`Đã xóa công ty ${company.name}`, "success")
    } catch (err: any) {
      pushToast(err?.response?.data?.error || "Xóa công ty thất bại", "error")
    } finally {
      setConfirmDeleteCompany(null)
    }
  }

  const resolveCompanyRequest = async (request: CompanyRequestItem, status: "APPROVED" | "REJECTED", customPayload?: CompanyPayload) => {
    try {
      const body: any = { status }
      if (status === "APPROVED" && customPayload) {
        body.name = customPayload.name
        body.logo_url = customPayload.logo_url
        body.website = customPayload.website
        body.industry = customPayload.industry
        body.size = customPayload.size
        body.description = customPayload.description
      }

      await axios.patch(
        `${API_BASE}/admin/company-requests/${request.id}`,
        body,
        { headers: { Authorization: `Bearer ${adminToken}` } }
      )
      pushToast(status === "APPROVED" ? `Đã duyệt ${request.name}` : `Đã từ chối ${request.name}`, "success")
      await reloadCompanyViews()
    } catch (err: any) {
      pushToast(err?.response?.data?.error || "Xử lý yêu cầu thất bại", "error")
    }
  }

  const openApproveRequestModal = (request: CompanyRequestItem) => {
    setApprovingRequest(request)
    setApproveError("")
    setApprovePayload({
      name: request.name || "",
      logo_url: request.logo_url || "",
      website: request.website || "",
      industry: request.industry || "",
      size: request.size || "",
      description: request.description || "",
    })
  }

  const closeApproveRequestModal = () => {
    if (approving) return
    setApprovingRequest(null)
    setApprovePayload(emptyPayload)
    setApproveError("")
  }

  const confirmApproveRequest = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!approvingRequest) return
    if (!approvePayload.name.trim()) {
      setApproveError("Tên công ty là bắt buộc")
      return
    }

    setApproving(true)
    setApproveError("")
    try {
      await resolveCompanyRequest(approvingRequest, "APPROVED", {
        ...approvePayload,
        name: approvePayload.name.trim(),
        logo_url: approvePayload.logo_url.trim(),
        website: approvePayload.website.trim(),
        industry: approvePayload.industry.trim(),
        size: approvePayload.size.trim(),
        description: approvePayload.description.trim(),
      })
      setApprovingRequest(null)
      setApprovePayload(emptyPayload)
      setApproveError("")
    } catch (err: any) {
      setApproveError(err?.response?.data?.error || "Xử lý yêu cầu thất bại")
    } finally {
      setApproving(false)
    }
  }

  const filtered = companies

  const totalPages = Math.max(1, Math.ceil(total / limit))

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <h2 className="text-2xl font-bold text-gray-900">Quản lý Công ty</h2>
        <button onClick={openCreate} className="bg-blue-600 text-white px-4 py-2 rounded-lg font-medium hover:bg-blue-700 transition">
          + Thêm công ty mới
        </button>
      </div>

      <div className="bg-white border rounded-2xl shadow-sm p-4">
        <input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Tìm theo tên, website, ngành"
          className="w-full border rounded-lg px-3 py-2"
        />
      </div>

      <div className="bg-white border rounded-2xl shadow-sm p-4 space-y-3">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold text-slate-900">Yêu cầu thêm công ty</h3>
          <span className="text-xs px-2 py-1 rounded-full bg-amber-100 text-amber-700">
            {requestsLoading ? "..." : `${companyRequests.length} chờ duyệt`}
          </span>
        </div>

        {companyRequests.length === 0 && (
          <p className="text-sm text-slate-500">Hiện không có yêu cầu nào.</p>
        )}

        {companyRequests.length > 0 && (
          <div className="space-y-3">
            {companyRequests.map((req) => (
              <div key={req.id} className="border rounded-xl p-3 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
                <div className="flex items-center gap-3 min-w-0">
                  <img
                    src={req.logo_url || fallbackLogo(req.name)}
                    onError={(e) => {
                      e.currentTarget.onerror = null
                      e.currentTarget.src = fallbackLogo(req.name)
                    }}
                    className="w-10 h-10 rounded border object-contain bg-white"
                  />
                  <div className="min-w-0">
                    <p className="font-semibold text-slate-900 truncate">{req.name}</p>
                    <p className="text-xs text-slate-500">Quy mô: {req.size || "Chưa cung cấp"}</p>
                  </div>
                </div>
                <ActionMenu menuClassName="min-w-[160px]">
                    <button
                      className="w-full text-left px-3 py-2 text-sm rounded-md text-emerald-700 hover:bg-emerald-50"
                      onClick={() => openApproveRequestModal(req)}
                    >
                      Duyệt yêu cầu
                    </button>
                    <button
                      className="w-full text-left px-3 py-2 text-sm rounded-md text-rose-700 hover:bg-rose-50"
                      onClick={() => resolveCompanyRequest(req, "REJECTED")}
                    >
                      Từ chối yêu cầu
                    </button>
                </ActionMenu>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="bg-white border rounded-2xl shadow-sm overflow-x-auto overflow-y-visible">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 border-b text-left text-slate-600">
            <tr>
              <th className="px-4 py-3">Công ty</th>
              <th className="px-4 py-3">Ngành</th>
              <th className="px-4 py-3">Website</th>
              <th className="px-4 py-3">Review</th>
              <th className="px-4 py-3">Hành động</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((c) => (
              <tr key={c.id} className="border-b last:border-b-0">
                <td className="px-4 py-3">
                  <div className="flex items-center gap-3">
                    <img
                      src={resolveLogoUrl(c.name, c.logo_url)}
                      onError={(e) => {
                        e.currentTarget.onerror = null
                        e.currentTarget.src = fallbackLogo(c.name)
                      }}
                      className="w-10 h-10 rounded border object-contain bg-white"
                    />
                    <div>
                      <div className="font-medium text-slate-900">{c.name}</div>
                      <div className="text-xs text-slate-500">{c.size || "-"}</div>
                    </div>
                  </div>
                </td>
                <td className="px-4 py-3">{c.industry || "-"}</td>
                <td className="px-4 py-3">{c.website ? <a className="text-blue-600 hover:underline" href={c.website} target="_blank" rel="noreferrer">{c.website}</a> : "-"}</td>
                <td className="px-4 py-3">{c.total_reviews} ({c.avg_rating.toFixed(1)}★)</td>
                <td className="px-4 py-3">
                  <ActionMenu menuClassName="min-w-[140px]">
                      <button
                        className="w-full text-left px-3 py-2 text-sm rounded-md text-indigo-700 hover:bg-indigo-50"
                        onClick={() => openEdit(c)}
                      >
                        Sửa
                      </button>
                      <button
                        className="w-full text-left px-3 py-2 text-sm rounded-md text-rose-700 hover:bg-rose-50"
                        onClick={() => setConfirmDeleteCompany(c)}
                      >
                        Xóa
                      </button>
                  </ActionMenu>
                </td>
              </tr>
            ))}
            {filtered.length === 0 && (
              <tr>
                <td className="px-4 py-8 text-center text-gray-500" colSpan={5}>Không có công ty phù hợp</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {!query.trim() && (
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
      )}

      {formOpen && (
        <div className="fixed inset-0 bg-slate-900/40 z-50 flex items-start sm:items-center justify-center p-4 sm:p-6 overflow-y-auto">
          <form onSubmit={saveCompany} className="w-full max-w-2xl bg-white rounded-2xl border shadow-xl p-6 space-y-4 max-h-[90vh] overflow-y-auto">
            <h3 className="text-lg font-bold text-slate-900">{editing ? "Cập nhật công ty" : "Thêm công ty mới"}</h3>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <input value={payload.name} onChange={(e) => setPayload((p) => ({ ...p, name: e.target.value }))} placeholder="Tên công ty *" className="border rounded-lg px-3 py-2" required />
              <input value={payload.industry} onChange={(e) => setPayload((p) => ({ ...p, industry: e.target.value }))} placeholder="Ngành" className="border rounded-lg px-3 py-2" />
              <input value={payload.website} onChange={(e) => setPayload((p) => ({ ...p, website: e.target.value }))} placeholder="Website" className="border rounded-lg px-3 py-2" />
              <input value={payload.logo_url} onChange={(e) => setPayload((p) => ({ ...p, logo_url: e.target.value }))} placeholder="Đường dẫn logo" className="border rounded-lg px-3 py-2" />
              <input value={payload.size} onChange={(e) => setPayload((p) => ({ ...p, size: e.target.value }))} placeholder="Quy mô" className="border rounded-lg px-3 py-2 sm:col-span-2" />
            </div>

            <textarea
              value={payload.description}
              onChange={(e) => setPayload((p) => ({ ...p, description: e.target.value }))}
              placeholder="Mô tả"
              rows={4}
              className="w-full border rounded-lg px-3 py-2"
            />

            {error ? <p className="text-sm text-red-600">{error}</p> : null}

            <div className="flex justify-end gap-3 pt-2">
              <button type="button" onClick={closeForm} className="px-4 py-2 rounded-lg border">Hủy</button>
              <button type="submit" disabled={saving} className="px-4 py-2 rounded-lg bg-blue-600 text-white disabled:opacity-60">
                {saving ? "Đang lưu..." : "Lưu"}
              </button>
            </div>
          </form>
        </div>
      )}

      {confirmDeleteCompany && (
        <div className="fixed inset-0 bg-slate-900/40 z-50 flex items-start sm:items-center justify-center p-4 sm:p-6 overflow-y-auto">
          <div className="w-full max-w-md bg-white rounded-2xl border shadow-xl p-6 space-y-4 max-h-[90vh] overflow-y-auto">
            <h3 className="text-lg font-bold text-slate-900">Xác nhận xóa công ty</h3>
            <p className="text-sm text-slate-600">
              Bạn có chắc muốn xóa <span className="font-semibold">{confirmDeleteCompany.name}</span>?
            </p>
            <div className="flex justify-end gap-3 pt-2">
              <button type="button" onClick={() => setConfirmDeleteCompany(null)} className="px-4 py-2 rounded-lg border">
                Hủy
              </button>
              <button
                type="button"
                onClick={() => deleteCompany(confirmDeleteCompany)}
                className="px-4 py-2 rounded-lg bg-red-600 text-white hover:bg-red-700"
              >
                Xóa
              </button>
            </div>
          </div>
        </div>
      )}

      {approvingRequest && (
        <div className="fixed inset-0 bg-slate-900/40 z-50 flex items-start sm:items-center justify-center p-4 sm:p-6 overflow-y-auto">
          <form onSubmit={confirmApproveRequest} className="w-full max-w-2xl bg-white rounded-2xl border shadow-xl p-6 space-y-4 max-h-[90vh] overflow-y-auto">
            <h3 className="text-lg font-bold text-slate-900">Duyệt yêu cầu và tạo công ty</h3>
            <p className="text-sm text-slate-600">Nhập đầy đủ thông tin công ty như form thêm công ty ở Admin, sau đó xác nhận tạo.</p>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <input value={approvePayload.name} onChange={(e) => setApprovePayload((p) => ({ ...p, name: e.target.value }))} placeholder="Tên công ty *" className="border rounded-lg px-3 py-2" required />
              <input value={approvePayload.industry} onChange={(e) => setApprovePayload((p) => ({ ...p, industry: e.target.value }))} placeholder="Ngành" className="border rounded-lg px-3 py-2" />
              <input value={approvePayload.website} onChange={(e) => setApprovePayload((p) => ({ ...p, website: e.target.value }))} placeholder="Website" className="border rounded-lg px-3 py-2" />
              <input value={approvePayload.logo_url} onChange={(e) => setApprovePayload((p) => ({ ...p, logo_url: e.target.value }))} placeholder="Đường dẫn logo" className="border rounded-lg px-3 py-2" />
              <input value={approvePayload.size} onChange={(e) => setApprovePayload((p) => ({ ...p, size: e.target.value }))} placeholder="Quy mô" className="border rounded-lg px-3 py-2 sm:col-span-2" />
            </div>

            <textarea
              value={approvePayload.description}
              onChange={(e) => setApprovePayload((p) => ({ ...p, description: e.target.value }))}
              placeholder="Mô tả"
              rows={4}
              className="w-full border rounded-lg px-3 py-2"
            />

            {approveError ? <p className="text-sm text-red-600">{approveError}</p> : null}

            <div className="flex justify-end gap-3 pt-2">
              <button type="button" onClick={closeApproveRequestModal} className="px-4 py-2 rounded-lg border" disabled={approving}>
                Hủy
              </button>
              <button type="submit" disabled={approving} className="px-4 py-2 rounded-lg bg-emerald-600 text-white hover:bg-emerald-700 disabled:opacity-60">
                {approving ? "Đang tạo..." : "Xác nhận tạo công ty"}
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  )
}
