import { useEffect, useState } from "react"
import { useParams } from "react-router-dom"
import { Star, MessageSquare, ThumbsUp, ThumbsDown, Briefcase } from "lucide-react"
import { useForm } from "react-hook-form"
import api from "../api"

interface Company {
  id: string
  name: string
  logo_url: string
  industry: string
  size: string
  website: string
  description: string
  avg_rating: number
  total_reviews: number
}

interface Review {
  id: string
  author_name: string
  rating: number
  title: string
  content: string
  pros: string
  cons: string
  salary_gross: number
  interview_exp: string
  comment_count?: number
  like_count?: number
  dislike_count?: number
  created_at: string
}

interface CommentItem {
  id: string
  author_name: string
  content: string
  created_at: string
  parent_comment_id?: string | null
  like_count?: number
  dislike_count?: number
}

const fallbackLogo = (name: string) =>
  `https://ui-avatars.com/api/?name=${encodeURIComponent(name || "Company")}&background=0F766E&color=FFFFFF&bold=true`

export default function CompanyDetail() {
  const { id } = useParams()
  const [company, setCompany] = useState<Company | null>(null)
  const [reviews, setReviews] = useState<Review[]>([])
  const [commentsByReview, setCommentsByReview] = useState<Record<string, CommentItem[]>>({})
  const [openComments, setOpenComments] = useState<Record<string, boolean>>({})
  const [visibleCommentCount, setVisibleCommentCount] = useState<Record<string, number>>({})
  const [newCommentByReview, setNewCommentByReview] = useState<Record<string, string>>({})
  const [replyingTo, setReplyingTo] = useState<Record<string, string | null>>({})
  const [replyContentByReview, setReplyContentByReview] = useState<Record<string, string>>({})
  const [showReviewForm, setShowReviewForm] = useState(false)
  const [sessionId] = useState(() => {
    const key = "review_session_id"
    let value = localStorage.getItem(key)
    if (!value) {
      value = crypto?.randomUUID?.() || `${Date.now()}-${Math.random()}`
      localStorage.setItem(key, value)
    }
    return value
  })

  const { register, handleSubmit, reset } = useForm()

  useEffect(() => {
    fetchData()
  }, [id])

  const fetchData = async () => {
    try {
      const [cRes, rRes] = await Promise.all([
        api.get(`/companies/${id}`),
        api.get(`/companies/${id}/reviews`)
      ])
      setCompany(cRes.data)
      setReviews(rRes.data?.data || [])
    } catch {
      setCompany(null)
      setReviews([])
    }
  }

  const loadComments = async (reviewId: string) => {
    try {
      const res = await api.get(`/reviews/${reviewId}/comments?limit=200`)
      setCommentsByReview((prev) => ({ ...prev, [reviewId]: res.data?.data || [] }))
    } catch {
      setCommentsByReview((prev) => ({ ...prev, [reviewId]: [] }))
    }
  }

  const toggleComments = async (reviewId: string) => {
    const isOpen = !!openComments[reviewId]
    setOpenComments((prev) => ({ ...prev, [reviewId]: !isOpen }))
    if (!isOpen) {
      setVisibleCommentCount((prev) => ({ ...prev, [reviewId]: 3 }))
      await loadComments(reviewId)
    }
  }

  const createComment = async (reviewId: string, parentId?: string) => {
    const content = parentId
      ? (replyContentByReview[reviewId] || "").trim()
      : (newCommentByReview[reviewId] || "").trim()
    if (!content) return

    await api.post(`/reviews/${reviewId}/comments`, {
      author_name: "Ẩn danh",
      content,
      parent_comment_id: parentId || null,
    })

    if (parentId) {
      setReplyingTo((prev) => ({ ...prev, [reviewId]: null }))
      setReplyContentByReview((prev) => ({ ...prev, [reviewId]: "" }))
    } else {
      setNewCommentByReview((prev) => ({ ...prev, [reviewId]: "" }))
    }

    await loadComments(reviewId)
  }

  const voteReview = async (reviewId: string, vote: "like" | "dislike") => {
    const res = await api.post(
      `/reviews/${reviewId}/vote`,
      { vote },
      { headers: { "X-Session-ID": sessionId } }
    )
    const like = Number(res.data?.like_count || 0)
    const dislike = Number(res.data?.dislike_count || 0)
    setReviews((prev) => prev.map((r) => (r.id === reviewId ? { ...r, like_count: like, dislike_count: dislike } : r)))
  }

  const voteComment = async (commentId: string, vote: "like" | "dislike") => {
    const res = await api.post(
      `/comments/${commentId}/vote`,
      { vote },
      { headers: { "X-Session-ID": sessionId } }
    )
    const like = Number(res.data?.like_count || 0)
    const dislike = Number(res.data?.dislike_count || 0)
    setCommentsByReview((prev) => {
      const next: Record<string, CommentItem[]> = {}
      Object.keys(prev).forEach((reviewId) => {
        next[reviewId] = prev[reviewId].map((c) => (c.id === commentId ? { ...c, like_count: like, dislike_count: dislike } : c))
      })
      return next
    })
  }

  const onSubmitReview = async (data: any) => {
    try {
      await api.post(`/companies/${id}/reviews`, {
        ...data,
        rating: Number(data.rating),
        salary_gross: data.salary_gross ? Number(data.salary_gross) : null
      })
      alert("Review submitted successfully!")
      setShowReviewForm(false)
      reset()
      fetchData()
    } catch (e) {
      alert("Failed to submit review")
    }
  }

  if (!company) return <div className="p-12 text-center text-gray-500">Loading...</div>

  return (
    <div className="max-w-4xl mx-auto space-y-8">
      {/* Header Info */}
      <div className="bg-white rounded-2xl shadow-sm border p-8 flex flex-col sm:flex-row items-center sm:items-start gap-6 relative">
        <img
          src={company.logo_url || fallbackLogo(company.name)}
          alt={company.name}
          onError={(e) => {
            e.currentTarget.onerror = null
            e.currentTarget.src = fallbackLogo(company.name)
          }}
          className="w-32 h-32 object-contain bg-white rounded-xl border p-2 shadow-sm"
        />
        <div className="flex-1 text-center sm:text-left">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">{company.name}</h1>
          <div className="flex flex-wrap items-center justify-center sm:justify-start gap-3 text-sm text-gray-600 mb-4">
            <span className="bg-blue-50 text-blue-700 px-3 py-1 rounded-full font-medium">{company.industry}</span>
            <span className="flex items-center gap-1"><Briefcase className="w-4 h-4"/> {company.size}</span>
            {company.website && <a href={company.website} target="_blank" className="text-blue-500 hover:underline">{company.website}</a>}
          </div>
          <p className="text-gray-700 leading-relaxed text-sm mb-4 line-clamp-3">
            {company.description}
          </p>
          <div className="flex items-center justify-center sm:justify-start gap-4">
            <div className="flex items-center space-x-1">
              <Star className="w-6 h-6 fill-yellow-400 text-yellow-400" />
              <span className="font-bold text-xl">{company.avg_rating.toFixed(1)}</span>
            </div>
            <span className="text-gray-400 text-sm">({company.total_reviews} reviews)</span>
            <button 
              onClick={() => setShowReviewForm(!showReviewForm)}
              className="ml-auto bg-blue-600 text-white px-6 py-2 rounded-lg font-medium hover:bg-blue-700 transition"
            >
              Viết Review
            </button>
          </div>
        </div>
      </div>

      {/* Review Form */}
      {showReviewForm && (
        <form onSubmit={handleSubmit(onSubmitReview)} className="bg-white rounded-2xl shadow-sm border p-8 space-y-4">
          <h3 className="text-xl font-bold text-gray-900 border-b pb-4">Đánh giá của bạn về {company.name}</h3>
          
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Tên hiển thị (Tùy chọn)</label>
              <input {...register("author_name")} placeholder="Ẩn danh" className="w-full border rounded-lg px-4 py-2 focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Đánh giá chung (1-5 sao) *</label>
              <select {...register("rating", { required: true })} className="w-full border rounded-lg px-4 py-2 focus:ring-2 focus:ring-blue-500 outline-none">
                <option value="5">5 Sao - Tuyệt vời</option>
                <option value="4">4 Sao - Tốt</option>
                <option value="3">3 Sao - Tạm được</option>
                <option value="2">2 Sao - Tệ</option>
                <option value="1">1 Sao - Chạy ngay đi</option>
              </select>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Tiêu đề *</label>
            <input {...register("title", { required: true })} placeholder="Ví dụ: Môi trường tốt, lương ổn" className="w-full border rounded-lg px-4 py-2 focus:ring-2 focus:ring-blue-500 outline-none" />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Chi tiết Review *</label>
            <textarea {...register("content", { required: true })} rows={4} placeholder="Chia sẻ kinh nghiệm làm việc của bạn ở đây..." className="w-full border rounded-lg px-4 py-2 focus:ring-2 focus:ring-blue-500 outline-none"></textarea>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-green-700 mb-1 flex items-center gap-1"><ThumbsUp className="w-4 h-4"/> Ưu điểm</label>
              <textarea {...register("pros")} rows={2} className="w-full border rounded-lg px-4 py-2 border-green-200 focus:ring-2 focus:ring-green-500 outline-none"></textarea>
            </div>
            <div>
              <label className="block text-sm font-medium text-red-700 mb-1 flex items-center gap-1"><ThumbsDown className="w-4 h-4"/> Nhược điểm</label>
              <textarea {...register("cons")} rows={2} className="w-full border rounded-lg px-4 py-2 border-red-200 focus:ring-2 focus:ring-red-500 outline-none"></textarea>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Mức lương (VNĐ/tháng) - Tùy chọn</label>
              <input type="number" {...register("salary_gross")} placeholder="VD: 20000000" className="w-full border rounded-lg px-4 py-2 focus:ring-2 focus:ring-blue-500 outline-none" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Kinh nghiệm phỏng vấn - Tùy chọn</label>
              <input {...register("interview_exp")} placeholder="VD: Phỏng vấn 2 vòng, thuật toán + culture fit" className="w-full border rounded-lg px-4 py-2 focus:ring-2 focus:ring-blue-500 outline-none" />
            </div>
          </div>

          <div className="pt-4 border-t flex justify-end space-x-3">
            <button type="button" onClick={() => setShowReviewForm(false)} className="px-6 py-2 rounded-lg font-medium text-gray-600 hover:bg-gray-100">Hủy</button>
            <button type="submit" className="bg-blue-600 text-white px-6 py-2 rounded-lg font-medium hover:bg-blue-700 transition">Gửi đánh giá</button>
          </div>
        </form>
      )}

      {/* Reviews List */}
      <div className="space-y-6">
        <h2 className="text-2xl font-bold text-gray-900 border-b pb-4">Tất cả bài Review ({company.total_reviews})</h2>
        {reviews.length === 0 ? (
          <p className="text-gray-500 text-center py-8">Chưa có đánh giá nào cho công ty này. Hãy là người đầu tiên!</p>
        ) : (
          reviews.map(review => {
            const comments = Array.isArray(commentsByReview[review.id]) ? commentsByReview[review.id] : []
            const topLevelComments = comments.filter((c) => !c.parent_comment_id)
            const shownTopLevelComments = topLevelComments.slice(0, visibleCommentCount[review.id] || 3)
            const hasMoreComments = topLevelComments.length > (visibleCommentCount[review.id] || 3)
            const repliesByParent = comments.reduce<Record<string, CommentItem[]>>((acc, c) => {
              if (!c.parent_comment_id) return acc
              if (!acc[c.parent_comment_id]) acc[c.parent_comment_id] = []
              acc[c.parent_comment_id].push(c)
              return acc
            }, {})

            return (
            <div key={review.id} className="bg-white rounded-2xl shadow-sm border p-6 space-y-4">
              <div className="flex justify-between items-start">
                <div className="flex items-center space-x-3">
                  <div className="w-10 h-10 rounded-full bg-slate-200 flex items-center justify-center font-bold text-slate-500">
                    {(review.author_name || "A").charAt(0).toUpperCase()}
                  </div>
                  <div>
                    <h4 className="font-bold text-gray-900">{review.author_name}</h4>
                    <p className="text-xs text-gray-500">{new Date(review.created_at).toLocaleDateString()}</p>
                  </div>
                </div>
                <div className="flex space-x-1">
                  {[...Array(5)].map((_, i) => (
                    <Star key={i} className={`w-5 h-5 ${i < review.rating ? "fill-yellow-400 text-yellow-400" : "text-gray-300"}`} />
                  ))}
                </div>
              </div>
              
              <h3 className="text-lg font-bold">{review.title}</h3>
              <p className="text-gray-700 whitespace-pre-wrap">{review.content}</p>
              
              {(review.pros || review.cons) && (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mt-4 bg-slate-50 p-4 rounded-xl border">
                  {review.pros && (
                    <div>
                      <h5 className="text-green-700 font-semibold text-sm mb-1 flex items-center gap-1"><ThumbsUp className="w-4 h-4"/> Ưu điểm</h5>
                      <p className="text-sm text-gray-600 whitespace-pre-wrap">{review.pros}</p>
                    </div>
                  )}
                  {review.cons && (
                    <div>
                      <h5 className="text-red-700 font-semibold text-sm mb-1 flex items-center gap-1"><ThumbsDown className="w-4 h-4"/> Nhược điểm</h5>
                      <p className="text-sm text-gray-600 whitespace-pre-wrap">{review.cons}</p>
                    </div>
                  )}
                </div>
              )}

              <div className="pt-4 border-t flex items-center gap-6 text-sm text-gray-500">
                <button onClick={() => toggleComments(review.id)} className="flex items-center gap-1 hover:text-blue-600 transition"><MessageSquare className="w-4 h-4"/> Bình luận ({openComments[review.id] ? comments.length : (review.comment_count || 0)})</button>
                <button onClick={() => voteReview(review.id, "like")} className="flex items-center gap-1 hover:text-green-600 transition"><ThumbsUp className="w-4 h-4"/> {review.like_count || 0}</button>
                <button onClick={() => voteReview(review.id, "dislike")} className="flex items-center gap-1 hover:text-red-600 transition"><ThumbsDown className="w-4 h-4"/> {review.dislike_count || 0}</button>
                {review.salary_gross && <span>Lương: <b>{review.salary_gross.toLocaleString()}đ</b></span>}
              </div>

              {openComments[review.id] && (
                <div className="mt-2 bg-slate-50 rounded-xl border p-4 space-y-3">
                  <div className="flex gap-2">
                    <input
                      value={newCommentByReview[review.id] || ""}
                      onChange={(e) => setNewCommentByReview((prev) => ({ ...prev, [review.id]: e.target.value }))}
                      placeholder="Viết bình luận..."
                      className="flex-1 border rounded-lg px-3 py-2 text-sm"
                    />
                    <button onClick={() => createComment(review.id)} className="bg-blue-600 text-white px-3 py-2 rounded-lg text-sm">Gửi</button>
                  </div>

                  <div className="space-y-2">
                    {shownTopLevelComments.map((c) => (
                      <div key={c.id} className="bg-white border rounded-lg p-3">
                        <div className="text-xs text-gray-500">{c.author_name} · {new Date(c.created_at).toLocaleString()}</div>
                        <p className="text-sm text-gray-700 mt-1">{c.content}</p>
                        <button
                          className="text-xs text-blue-600 hover:underline mt-1"
                          onClick={() => setReplyingTo((prev) => ({ ...prev, [review.id]: prev[review.id] === c.id ? null : c.id }))}
                        >
                          Reply
                        </button>
                        <button className="text-xs text-green-600 hover:underline mt-1 ml-3" onClick={() => voteComment(c.id, "like")}>
                          Hữu ích ({c.like_count || 0})
                        </button>
                        <button className="text-xs text-red-600 hover:underline mt-1 ml-3" onClick={() => voteComment(c.id, "dislike")}>
                          Không hữu ích ({c.dislike_count || 0})
                        </button>

                        {replyingTo[review.id] === c.id && (
                          <div className="flex gap-2 mt-2">
                            <input
                              value={replyContentByReview[review.id] || ""}
                              onChange={(e) => setReplyContentByReview((prev) => ({ ...prev, [review.id]: e.target.value }))}
                              placeholder="Nhập reply..."
                              className="flex-1 border rounded-lg px-3 py-2 text-sm"
                            />
                            <button onClick={() => createComment(review.id, c.id)} className="bg-slate-900 text-white px-3 py-2 rounded-lg text-sm">Gửi</button>
                          </div>
                        )}

                        {(repliesByParent[c.id] || []).length > 0 && (
                          <div className="mt-2 pl-4 border-l space-y-2">
                            {(repliesByParent[c.id] || []).map((r) => (
                              <div key={r.id} className="bg-slate-50 border rounded-lg p-2">
                                <div className="text-xs text-gray-500">{r.author_name} · {new Date(r.created_at).toLocaleString()}</div>
                                <p className="text-sm text-gray-700 mt-1">{r.content}</p>
                                <div className="mt-1 flex items-center gap-3 text-xs">
                                  <button className="text-green-600 hover:underline" onClick={() => voteComment(r.id, "like")}>Hữu ích ({r.like_count || 0})</button>
                                  <button className="text-red-600 hover:underline" onClick={() => voteComment(r.id, "dislike")}>Không hữu ích ({r.dislike_count || 0})</button>
                                </div>
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                    ))}
                    {hasMoreComments && (
                      <button
                        className="text-sm text-blue-600 hover:underline"
                        onClick={() => setVisibleCommentCount((prev) => ({ ...prev, [review.id]: (prev[review.id] || 3) + 3 }))}
                      >
                        Xem thêm bình luận
                      </button>
                    )}
                    {topLevelComments.length === 0 && <div className="text-sm text-gray-500">Chưa có bình luận.</div>}
                  </div>
                </div>
              )}
            </div>
            )
          })
        )}
      </div>
    </div>
  )
}
