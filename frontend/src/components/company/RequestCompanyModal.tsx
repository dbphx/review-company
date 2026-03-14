import { useState } from "react"
import api from "../../api"
import { useToast } from "../ui/ToastProvider"

type Props = {
  open: boolean
  onClose: () => void
}

export default function RequestCompanyModal({ open, onClose }: Props) {
  const [name, setName] = useState("")
  const [logoUrl, setLogoUrl] = useState("")
  const [website, setWebsite] = useState("")
  const [industry, setIndustry] = useState("")
  const [size, setSize] = useState("")
  const [description, setDescription] = useState("")
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState("")
  const { pushToast } = useToast()

  if (!open) return null

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) {
      setError("Tên công ty là bắt buộc")
      return
    }

    setSubmitting(true)
    setError("")
    try {
      await api.post("/company-requests", {
        name: name.trim(),
        logo_url: logoUrl.trim(),
        website: website.trim(),
        industry: industry.trim(),
        size: size.trim(),
        description: description.trim(),
      })
      pushToast("Đã gửi yêu cầu thêm công ty", "success")
      setName("")
      setLogoUrl("")
      setWebsite("")
      setIndustry("")
      setSize("")
      setDescription("")
      setError("")
      onClose()
    } catch (err: any) {
      setError(err?.response?.data?.error || "Gửi yêu cầu thất bại")
      pushToast(err?.response?.data?.error || "Gửi yêu cầu thất bại", "error")
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 bg-slate-900/40 flex items-center justify-center p-4">
      <form onSubmit={submit} className="w-full max-w-2xl bg-white rounded-2xl border shadow-xl p-6 space-y-4">
        <h3 className="text-lg font-bold text-slate-900">Yêu cầu thêm công ty mới</h3>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Tên công ty *"
            className="w-full border rounded-lg px-3 py-2"
            required
          />
          <input
            value={industry}
            onChange={(e) => setIndustry(e.target.value)}
            placeholder="Ngành"
            className="w-full border rounded-lg px-3 py-2"
          />
          <input
            value={website}
            onChange={(e) => setWebsite(e.target.value)}
            placeholder="Website"
            className="w-full border rounded-lg px-3 py-2"
          />
          <input
            value={logoUrl}
            onChange={(e) => setLogoUrl(e.target.value)}
            placeholder="Đường dẫn logo"
            className="w-full border rounded-lg px-3 py-2"
          />
          <input
            value={size}
            onChange={(e) => setSize(e.target.value)}
            placeholder="Quy mô"
            className="w-full border rounded-lg px-3 py-2 sm:col-span-2"
          />
        </div>
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Mô tả"
          rows={4}
          className="w-full border rounded-lg px-3 py-2"
        />
        {error ? <p className="text-sm text-red-600">{error}</p> : null}
        <div className="flex justify-end gap-3 pt-2">
          <button type="button" onClick={onClose} className="px-4 py-2 border rounded-lg">Hủy</button>
          <button type="submit" disabled={submitting} className="px-4 py-2 bg-blue-600 text-white rounded-lg disabled:opacity-60">
            {submitting ? "Đang gửi..." : "Gửi yêu cầu"}
          </button>
        </div>
      </form>
    </div>
  )
}
