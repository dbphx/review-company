import { useEffect, useState } from "react"
import { Link, useNavigate } from "react-router-dom"
import { Search } from "lucide-react"
import api from "../api"

interface Company {
  id: string
  name: string
  logo_url: string
  industry: string
  avg_rating: number
  total_reviews: number
}

interface ReviewItem {
  id: string
  title: string
  content: string
  rating: number
  author_name: string
  created_at: string
  company?: {
    id: string
    name: string
  }
}

const fallbackLogo = (name: string) =>
  `https://ui-avatars.com/api/?name=${encodeURIComponent(name || "Company")}&background=0F766E&color=FFFFFF&bold=true`

export default function Home() {
  const [query, setQuery] = useState("")
  const [results, setResults] = useState<Company[]>([])
  const [searching, setSearching] = useState(false)
  const [topCompanies, setTopCompanies] = useState<Company[]>([])
  const [latestReviews, setLatestReviews] = useState<ReviewItem[]>([])
  const navigate = useNavigate()

  useEffect(() => {
    // Fetch top companies on load
    api.get("/companies/top?limit=6").then((res) => {
      setTopCompanies(res.data?.data || [])
    }).catch(() => {
      setTopCompanies([])
    })
    api.get("/reviews/recent?limit=5").then((res) => {
      setLatestReviews(res.data?.data || [])
    }).catch(() => {
      setLatestReviews([])
    })
  }, [])

  useEffect(() => {
    if (query.length >= 2) {
      const timeoutId = setTimeout(() => {
        setSearching(true)
        api.get(`/search?q=${encodeURIComponent(query)}`).then((res) => {
          const payload = Array.isArray(res.data) ? res.data : []
          setResults(payload)
        }).catch(() => {
          setResults([])
        }).finally(() => {
          setSearching(false)
        })
      }, 300)
      return () => clearTimeout(timeoutId)
    } else {
      setResults([])
      setSearching(false)
    }
  }, [query])

  return (
    <div className="space-y-12">
      {/* Hero Search Section */}
      <section className="bg-blue-600 text-white rounded-2xl p-12 text-center shadow-lg relative">
        <h1 className="text-4xl font-extrabold mb-4">Tìm kiếm công ty để Review</h1>
        <p className="text-blue-100 mb-8 max-w-2xl mx-auto">Hàng ngàn review ẩn danh về môi trường làm việc, văn hóa, và mức lương.</p>
        
        <div className="max-w-2xl mx-auto relative">
          <div className="relative">
            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <Search className="h-5 w-5 text-gray-400" />
            </div>
            <input
              type="text"
              className="block w-full pl-10 pr-3 py-4 border border-transparent rounded-lg leading-5 bg-white text-gray-900 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent sm:text-lg"
              placeholder="Nhập tên công ty (VD: FPT, VNG...)"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
            />
          </div>

          {/* Autocomplete Results */}
          {query.length >= 2 && (
            <div className="absolute z-10 w-full mt-1 bg-white rounded-md shadow-lg border text-left overflow-hidden">
              <ul className="max-h-60 rounded-md py-1 text-base leading-6 shadow-xs overflow-auto focus:outline-none sm:text-sm sm:leading-5">
                {searching && <li className="py-3 px-3 text-gray-500">Đang tìm kiếm...</li>}
                {!searching && results.length === 0 && (
                  <li className="py-3 px-3 text-gray-500">Không tìm thấy công ty phù hợp.</li>
                )}
                {!searching && results.map((c) => (
                  <li
                    key={c.id}
                    className="cursor-pointer select-none relative py-3 pl-3 pr-9 hover:bg-slate-100 border-b last:border-0"
                    onClick={() => navigate(`/company/${c.id}`)}
                  >
                    <div className="flex items-center space-x-3">
                      <img
                        src={c.logo_url || fallbackLogo(c.name)}
                        onError={(e) => {
                          e.currentTarget.onerror = null
                          e.currentTarget.src = fallbackLogo(c.name)
                        }}
                        className="w-10 h-10 object-contain bg-white rounded border"
                      />
                      <div>
                        <span className="block truncate font-medium text-gray-900">{c.name}</span>
                        <span className="block truncate text-gray-500 text-xs">{c.industry}</span>
                      </div>
                    </div>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      </section>

      {/* Top Companies */}
      <section>
        <h2 className="text-2xl font-bold mb-6 text-gray-900">Công ty nổi bật</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {topCompanies?.map((c) => (
            <Link key={c.id} to={`/company/${c.id}`} className="block">
              <div className="bg-white rounded-xl shadow-sm hover:shadow-md transition-shadow p-6 border">
                <div className="flex items-center space-x-4">
                  <img
                    src={c.logo_url || fallbackLogo(c.name)}
                    alt={c.name}
                    onError={(e) => {
                      e.currentTarget.onerror = null
                      e.currentTarget.src = fallbackLogo(c.name)
                    }}
                    className="w-16 h-16 object-contain rounded border p-1"
                  />
                  <div>
                    <h3 className="font-bold text-lg text-gray-900 line-clamp-1">{c.name}</h3>
                    <p className="text-sm text-gray-500">{c.industry}</p>
                    <div className="mt-1 flex items-center space-x-1">
                      <span className="text-yellow-400">★</span>
                      <span className="font-medium text-sm">{c.avg_rating.toFixed(1)}</span>
                      <span className="text-gray-400 text-xs">({c.total_reviews} reviews)</span>
                    </div>
                  </div>
                </div>
              </div>
            </Link>
          ))}
        </div>
      </section>

      <section>
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-2xl font-bold text-gray-900">Review mới nhất</h2>
          <span className="text-sm text-gray-500">Cập nhật theo thời gian thực</span>
        </div>
        <div className="space-y-4">
          {latestReviews.length === 0 && (
            <div className="bg-white rounded-xl border p-6 text-gray-500">Chưa có review mới.</div>
          )}
          {latestReviews.map((r) => (
            <div key={r.id} className="bg-white rounded-xl border p-5 hover:shadow-sm transition">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="text-xs text-gray-500 mb-1">
                    {r.company?.name ? (
                      <Link to={`/company/${r.company.id}`} className="text-blue-600 hover:underline">
                        {r.company.name}
                      </Link>
                    ) : "Không xác định công ty"}
                  </p>
                  <h3 className="font-semibold text-gray-900">{r.title}</h3>
                </div>
                <div className="text-sm font-medium text-yellow-500">★ {r.rating.toFixed(1)}</div>
              </div>
              <p className="text-sm text-gray-600 mt-2 line-clamp-2">{r.content}</p>
              <div className="mt-3 text-xs text-gray-500">
                bởi {r.author_name} · {new Date(r.created_at).toLocaleString()}
              </div>
            </div>
          ))}
        </div>
      </section>
    </div>
  )
}
