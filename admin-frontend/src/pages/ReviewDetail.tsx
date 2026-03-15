import { useEffect, useMemo, useState } from "react"
import { Link, useParams } from "react-router-dom"
import axios from "axios"
import { useToast } from "../components/ui/ToastProvider"

const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:3000/api"

interface CommentItem {
  id: string
  author_name: string
  content: string
  created_at: string
  parent_comment_id?: string | null
}

interface ReviewInfo {
  id: string
  title: string
  content: string
  author_name: string
  rating: number
  created_at: string
  company?: { id: string; name: string }
}

function ReviewCard({ review }: { review: ReviewInfo }) {
  return (
    <div className="bg-white rounded-2xl shadow-sm border p-6 space-y-4">
      <div className="flex justify-between items-start">
        <div className="flex items-center space-x-3">
          <div className="w-10 h-10 rounded-full bg-slate-200 flex items-center justify-center font-bold text-slate-500">
            {(review.author_name || "A").charAt(0).toUpperCase()}
          </div>
          <div>
            <h4 className="font-bold text-gray-900">{review.author_name}</h4>
            <p className="text-xs text-gray-500">{new Date(review.created_at).toLocaleString()}</p>
          </div>
        </div>
        <div className="text-yellow-500 font-semibold">★ {review.rating.toFixed(1)}</div>
      </div>
      <h3 className="text-lg font-bold">{review.title}</h3>
      <p className="text-gray-700 whitespace-pre-wrap">{review.content}</p>
      <div className="text-xs text-gray-500">{review.company ? review.company.name : ""}</div>
    </div>
  )
}

export default function ReviewDetail() {
  const { id } = useParams()
  const [review, setReview] = useState<ReviewInfo | null>(null)
  const [comments, setComments] = useState<CommentItem[]>([])
  const [newComment, setNewComment] = useState("")
  const [replyingTo, setReplyingTo] = useState<string | null>(null)
  const [replyContent, setReplyContent] = useState("")
  const [confirmDeleteCommentId, setConfirmDeleteCommentId] = useState<string | null>(null)
  const [confirmDeleteReview, setConfirmDeleteReview] = useState(false)
  const adminToken = localStorage.getItem("admin_token")
  const { pushToast } = useToast()

  const reload = () => {
    if (!id) return
    axios
      .get(`${API_BASE}/admin/reviews/${id}`, {
        headers: { Authorization: `Bearer ${adminToken}` },
      })
      .then((res) => setReview(res.data))
    axios
      .get(`${API_BASE}/admin/reviews/${id}/comments?limit=200`, {
        headers: { Authorization: `Bearer ${adminToken}` },
      })
      .then((res) => setComments(res.data.data || []))
  }

  useEffect(() => {
    reload()
  }, [id])

  const topLevelComments = useMemo(() => comments.filter((c) => !c.parent_comment_id), [comments])
  const repliesByParent = useMemo(() => {
    const map: Record<string, CommentItem[]> = {}
    comments.forEach((c) => {
      if (!c.parent_comment_id) return
      if (!map[c.parent_comment_id]) map[c.parent_comment_id] = []
      map[c.parent_comment_id].push(c)
    })
    return map
  }, [comments])

  const createComment = async (parentId?: string) => {
    if (!id) return
    const content = parentId ? replyContent.trim() : newComment.trim()
    if (!content) return

    try {
      await axios.post(`${API_BASE}/reviews/${id}/comments`, {
        author_name: "Admin",
        content,
        parent_comment_id: parentId || null,
      })
      pushToast("Đã gửi bình luận", "success")
    } catch (err: any) {
      pushToast(err?.response?.data?.error || "Gửi bình luận thất bại", "error")
      return
    }

    if (parentId) {
      setReplyingTo(null)
      setReplyContent("")
    } else {
      setNewComment("")
    }
    reload()
  }

  const deleteComment = async (commentId: string) => {
    try {
      await axios.delete(`${API_BASE}/comments/${commentId}`, {
        headers: { Authorization: `Bearer ${adminToken}` },
      })
      pushToast("Đã xóa bình luận", "success")
      reload()
    } catch (err: any) {
      pushToast(err?.response?.data?.error || "Xóa bình luận thất bại", "error")
    } finally {
      setConfirmDeleteCommentId(null)
    }
  }

  const deleteReview = async () => {
    if (!id) return
    try {
      await axios.delete(`${API_BASE}/reviews/${id}`, {
        headers: { Authorization: `Bearer ${adminToken}` },
      })
      pushToast("Đã xóa review", "success")
      window.location.href = "/reviews"
    } catch (err: any) {
      pushToast(err?.response?.data?.error || "Xóa review thất bại", "error")
    } finally {
      setConfirmDeleteReview(false)
    }
  }

  if (!review) {
    return <div className="text-gray-500">Đang tải review...</div>
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-gray-900">Chi tiết Review</h2>
        <div className="flex items-center gap-4">
          <button onClick={() => setConfirmDeleteReview(true)} className="text-red-600 hover:underline text-sm">Xóa Review</button>
          <Link to="/reviews" className="text-blue-600 hover:underline text-sm">← Quay lại danh sách</Link>
        </div>
      </div>

      <ReviewCard review={review} />

      <div className="bg-white border rounded-2xl shadow-sm p-6 space-y-4">
        <h4 className="font-semibold text-gray-900">Bình luận & Reply</h4>

        <div className="flex gap-2">
          <input
            value={newComment}
            onChange={(e) => setNewComment(e.target.value)}
            placeholder="Viết bình luận mới..."
            className="flex-1 border rounded-lg px-3 py-2"
          />
          <button onClick={() => createComment()} className="bg-blue-600 text-white px-4 py-2 rounded-lg">Gửi</button>
        </div>

        <div className="space-y-4">
          {topLevelComments.map((c) => (
            <div key={c.id} className="border rounded-xl p-4 bg-slate-50">
              <div className="text-sm font-medium text-gray-900">{c.author_name}</div>
              <div className="text-xs text-gray-500">{new Date(c.created_at).toLocaleString()}</div>
              <p className="text-sm text-gray-700 mt-2">{c.content}</p>

              <button
                className="text-xs text-blue-600 hover:underline mt-2"
                onClick={() => setReplyingTo(replyingTo === c.id ? null : c.id)}
              >
                Reply
              </button>
              <button
                className="text-xs text-red-600 hover:underline mt-2 ml-3"
                onClick={() => setConfirmDeleteCommentId(c.id)}
              >
                Xóa
              </button>

              {replyingTo === c.id && (
                <div className="flex gap-2 mt-2">
                  <input
                    value={replyContent}
                    onChange={(e) => setReplyContent(e.target.value)}
                    placeholder="Nhập reply..."
                    className="flex-1 border rounded-lg px-3 py-2 text-sm"
                  />
                  <button onClick={() => createComment(c.id)} className="bg-slate-900 text-white px-3 py-2 rounded-lg text-sm">Gửi</button>
                </div>
              )}

              {(repliesByParent[c.id] || []).length > 0 && (
                <div className="mt-3 pl-4 border-l space-y-2">
                  {(repliesByParent[c.id] || []).map((r) => (
                    <div key={r.id} className="bg-white border rounded-lg p-3">
                      <div className="text-xs text-gray-500">{r.author_name} · {new Date(r.created_at).toLocaleString()}</div>
                      <p className="text-sm text-gray-700 mt-1">{r.content}</p>
                      <button className="text-xs text-red-600 hover:underline mt-1" onClick={() => setConfirmDeleteCommentId(r.id)}>
                        Xóa
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))}
          {topLevelComments.length === 0 && <div className="text-sm text-gray-500">Chưa có bình luận.</div>}
        </div>
      </div>

      {confirmDeleteCommentId && (
        <div className="fixed inset-0 bg-slate-900/40 z-50 flex items-start sm:items-center justify-center p-4 sm:p-6 overflow-y-auto">
          <div className="w-full max-w-md bg-white rounded-2xl border shadow-xl p-6 space-y-4 max-h-[90vh] overflow-y-auto">
            <h3 className="text-lg font-bold text-slate-900">Xác nhận xóa bình luận</h3>
            <p className="text-sm text-slate-600">Bạn có chắc muốn xóa bình luận này?</p>
            <div className="flex justify-end gap-3 pt-2">
              <button type="button" onClick={() => setConfirmDeleteCommentId(null)} className="px-4 py-2 rounded-lg border">
                Hủy
              </button>
              <button
                type="button"
                onClick={() => deleteComment(confirmDeleteCommentId)}
                className="px-4 py-2 rounded-lg bg-red-600 text-white hover:bg-red-700"
              >
                Xóa
              </button>
            </div>
          </div>
        </div>
      )}

      {confirmDeleteReview && (
        <div className="fixed inset-0 bg-slate-900/40 z-50 flex items-start sm:items-center justify-center p-4 sm:p-6 overflow-y-auto">
          <div className="w-full max-w-md bg-white rounded-2xl border shadow-xl p-6 space-y-4 max-h-[90vh] overflow-y-auto">
            <h3 className="text-lg font-bold text-slate-900">Xác nhận xóa review</h3>
            <p className="text-sm text-slate-600">Bạn có chắc muốn xóa review này?</p>
            <div className="flex justify-end gap-3 pt-2">
              <button type="button" onClick={() => setConfirmDeleteReview(false)} className="px-4 py-2 rounded-lg border">
                Hủy
              </button>
              <button
                type="button"
                onClick={deleteReview}
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
