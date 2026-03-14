import { useEffect, useState } from "react"
import axios from "axios"

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

  const adminToken = localStorage.getItem("admin_token")

  const loadCompanies = async () => {
    const res = await axios.get(`${API_BASE}/companies/top?page=${page}&limit=${limit}`)
    const rows: CompanyItem[] = res.data?.data || []
    setCompanies(rows)
    setTotal(res.data?.total || 0)
  }

  useEffect(() => {
    loadCompanies()
  }, [page, limit])

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
      } else {
        await axios.post(`${API_BASE}/companies`, payload, {
          headers: { Authorization: `Bearer ${adminToken}` },
        })
      }
      closeForm()
      await loadCompanies()
    } catch (err: any) {
      setError(err?.response?.data?.error || "Lưu công ty thất bại")
    } finally {
      setSaving(false)
    }
  }

  const deleteCompany = async (company: CompanyItem) => {
    if (!window.confirm(`Xóa công ty ${company.name}?`)) return
    try {
      await axios.delete(`${API_BASE}/companies/${company.id}`, {
        headers: { Authorization: `Bearer ${adminToken}` },
      })
      await loadCompanies()
    } catch (err: any) {
      window.alert(err?.response?.data?.error || "Xóa công ty thất bại")
    }
  }

  const filtered = companies.filter((c) => {
    const q = query.trim().toLowerCase()
    if (!q) return true
    return (
      c.name.toLowerCase().includes(q) ||
      (c.website || "").toLowerCase().includes(q) ||
      (c.industry || "").toLowerCase().includes(q)
    )
  })

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

      <div className="bg-white border rounded-2xl shadow-sm overflow-hidden">
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
                      src={c.logo_url || fallbackLogo(c.name)}
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
                <td className="px-4 py-3 space-x-3">
                  <button className="text-indigo-600 hover:underline" onClick={() => openEdit(c)}>Sửa</button>
                  <button className="text-red-600 hover:underline" onClick={() => deleteCompany(c)}>Xóa</button>
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

      {formOpen && (
        <div className="fixed inset-0 bg-slate-900/40 z-50 flex items-center justify-center p-4">
          <form onSubmit={saveCompany} className="w-full max-w-2xl bg-white rounded-2xl border shadow-xl p-6 space-y-4">
            <h3 className="text-lg font-bold text-slate-900">{editing ? "Cập nhật công ty" : "Thêm công ty mới"}</h3>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <input value={payload.name} onChange={(e) => setPayload((p) => ({ ...p, name: e.target.value }))} placeholder="Tên công ty *" className="border rounded-lg px-3 py-2" required />
              <input value={payload.industry} onChange={(e) => setPayload((p) => ({ ...p, industry: e.target.value }))} placeholder="Ngành" className="border rounded-lg px-3 py-2" />
              <input value={payload.website} onChange={(e) => setPayload((p) => ({ ...p, website: e.target.value }))} placeholder="Website" className="border rounded-lg px-3 py-2" />
              <input value={payload.logo_url} onChange={(e) => setPayload((p) => ({ ...p, logo_url: e.target.value }))} placeholder="Logo URL" className="border rounded-lg px-3 py-2" />
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
    </div>
  )
}
