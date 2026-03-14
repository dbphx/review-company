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
  const [size, setSize] = useState("")
  const [submitting, setSubmitting] = useState(false)
  const { pushToast } = useToast()

  if (!open) return null

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) {
      pushToast("Tên công ty là bắt buộc", "error")
      return
    }

    setSubmitting(true)
    try {
      await api.post("/company-requests", {
        name: name.trim(),
        logo_url: logoUrl.trim(),
        size: size.trim(),
      })
      pushToast("Đã gửi yêu cầu thêm công ty", "success")
      setName("")
      setLogoUrl("")
      setSize("")
      onClose()
    } catch (err: any) {
      pushToast(err?.response?.data?.error || "Gửi yêu cầu thất bại", "error")
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 bg-slate-900/40 flex items-center justify-center p-4">
      <form onSubmit={submit} className="w-full max-w-lg bg-white rounded-2xl border shadow-xl p-6 space-y-4">
        <h3 className="text-lg font-bold text-slate-900">Yêu cầu thêm công ty mới</h3>
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Tên công ty *"
          className="w-full border rounded-lg px-3 py-2"
          required
        />
        <input
          value={logoUrl}
          onChange={(e) => setLogoUrl(e.target.value)}
          placeholder="Link logo (tùy chọn)"
          className="w-full border rounded-lg px-3 py-2"
        />
        <input
          value={size}
          onChange={(e) => setSize(e.target.value)}
          placeholder="Quy mô công ty (tùy chọn)"
          className="w-full border rounded-lg px-3 py-2"
        />
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
