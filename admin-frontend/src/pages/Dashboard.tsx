import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import { Users, Building2, MessageSquare, Star } from "lucide-react"
import { useEffect, useState } from 'react'
import axios from 'axios'

interface DailyReviewRow {
  day: string
  reviews: number
}

interface DailyVisitRow {
  day: string
  visits: number
}

interface MonthlyVisitRow {
  month: string
  visits: number
}

interface CompanyRow {
  id: string
  name: string
  avg_rating: number
  total_reviews: number
}

interface ChartCompanyRow {
  name: string
  rating: number
  reviews: number
}

const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:3000/api"

export default function Dashboard() {
  const [stats, setStats] = useState({
    total_companies: 0,
    total_reviews: 0,
    avg_rating: 0,
  })
  const [dailyData, setDailyData] = useState<Array<{ name: string; reviews: number }>>([])
  const [dailyVisits, setDailyVisits] = useState<Array<{ name: string; visits: number }>>([])
  const [monthlyVisits, setMonthlyVisits] = useState<Array<{ name: string; visits: number }>>([])
  const [adminCount, setAdminCount] = useState(0)
  const [newReviewsToday, setNewReviewsToday] = useState(0)
  const [topLimit, setTopLimit] = useState(10)
  const [topRated, setTopRated] = useState<ChartCompanyRow[]>([])
  const [lowRated, setLowRated] = useState<ChartCompanyRow[]>([])

  useEffect(() => {
    const adminToken = localStorage.getItem("admin_token")
    const authHeaders = { headers: { Authorization: `Bearer ${adminToken}` } }

    axios.get(`${API_BASE}/admin/companies/stats/summary`, authHeaders).then((res) => {
      setStats(res.data)
    })

    axios.get(`${API_BASE}/admin/reviews/stats/daily?days=7`, authHeaders).then((res) => {
      const mapped = ((res.data.data || []) as DailyReviewRow[]).map((row) => {
        const d = new Date(row.day)
        const weekday = d.toLocaleDateString('vi-VN', { weekday: 'short' })
        return { name: weekday, reviews: Number(row.reviews || 0) }
      })
      setDailyData(mapped)
      const today = mapped[mapped.length - 1]
      setNewReviewsToday(Number(today?.reviews || 0))
    })

    axios.get(`${API_BASE}/analytics/visits/daily?days=7`, authHeaders).then((res) => {
      const mapped = ((res.data.data || []) as DailyVisitRow[]).map((row) => {
        const d = new Date(row.day)
        const weekday = d.toLocaleDateString('vi-VN', { weekday: 'short' })
        return { name: weekday, visits: Number(row.visits || 0) }
      })
      setDailyVisits(mapped)
    })

    axios.get(`${API_BASE}/analytics/visits/monthly?months=6`, authHeaders).then((res) => {
      const mapped = ((res.data.data || []) as MonthlyVisitRow[]).map((row) => {
        const month = String(row.month || "")
        return { name: month, visits: Number(row.visits || 0) }
      })
      setMonthlyVisits(mapped)
    })

    axios.get(`${API_BASE}/admin/companies/top?limit=${topLimit}&order=desc`, authHeaders).then((res) => {
      const rows = (res.data?.data || []) as CompanyRow[]
      setTopRated(rows.map((c) => ({
        name: c.name.length > 24 ? `${c.name.slice(0, 24)}...` : c.name,
        rating: Number(c.avg_rating || 0),
        reviews: Number(c.total_reviews || 0),
      })))
    })

    axios.get(`${API_BASE}/admin/companies/top?limit=${topLimit}&order=asc`, authHeaders).then((res) => {
      const rows = (res.data?.data || []) as CompanyRow[]
      setLowRated(rows.map((c) => ({
        name: c.name.length > 24 ? `${c.name.slice(0, 24)}...` : c.name,
        rating: Number(c.avg_rating || 0),
        reviews: Number(c.total_reviews || 0),
      })))
    })

    if (adminToken) {
      axios
        .get(`${API_BASE}/admin/users`, authHeaders)
        .then((res) => {
          setAdminCount(Number(res.data?.total || 0))
        })
        .catch(() => {
          setAdminCount(0)
        })
    }
  }, [topLimit])

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-gray-900">Dashboard</h2>
      
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        <div className="bg-white p-6 border rounded-2xl shadow-sm flex items-center space-x-4">
          <div className="p-3 bg-blue-100 text-blue-600 rounded-xl"><Building2 className="w-6 h-6"/></div>
          <div>
            <p className="text-sm font-medium text-gray-500">Tổng công ty</p>
            <p className="text-2xl font-bold text-gray-900">{stats.total_companies}</p>
          </div>
        </div>
        <div className="bg-white p-6 border rounded-2xl shadow-sm flex items-center space-x-4">
          <div className="p-3 bg-green-100 text-green-600 rounded-xl"><MessageSquare className="w-6 h-6"/></div>
          <div>
            <p className="text-sm font-medium text-gray-500">Tổng Review</p>
            <p className="text-2xl font-bold text-gray-900">{stats.total_reviews}</p>
          </div>
        </div>
        <div className="bg-white p-6 border rounded-2xl shadow-sm flex items-center space-x-4">
          <div className="p-3 bg-yellow-100 text-yellow-600 rounded-xl"><Star className="w-6 h-6"/></div>
          <div>
            <p className="text-sm font-medium text-gray-500">Điểm TB Hệ thống</p>
            <p className="text-2xl font-bold text-gray-900">{stats.avg_rating.toFixed(1)}</p>
          </div>
        </div>
        <div className="bg-white p-6 border rounded-2xl shadow-sm flex items-center space-x-4">
          <div className="p-3 bg-purple-100 text-purple-600 rounded-xl"><Users className="w-6 h-6"/></div>
          <div>
            <p className="text-sm font-medium text-gray-500">Admin/Mod</p>
            <p className="text-2xl font-bold text-gray-900">{adminCount}</p>
          </div>
        </div>

        <div className="bg-white p-6 border rounded-2xl shadow-sm flex items-center space-x-4 sm:col-span-2 lg:col-span-4">
          <div className="p-3 bg-cyan-100 text-cyan-700 rounded-xl"><MessageSquare className="w-6 h-6"/></div>
          <div>
            <p className="text-sm font-medium text-gray-500">Review mới hôm nay</p>
            <p className="text-2xl font-bold text-gray-900">{newReviewsToday}</p>
          </div>
        </div>
      </div>

      <div className="bg-white p-6 border rounded-2xl shadow-sm">
        <h3 className="text-lg font-bold text-gray-800 mb-6">Lượng Review 7 ngày qua</h3>
        <div className="h-[400px] w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={dailyData}>
                <defs>
                  <linearGradient id="reviewsDailyGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#60a5fa" />
                    <stop offset="100%" stopColor="#2563eb" />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} />
                <XAxis dataKey="name" axisLine={false} tickLine={false} />
                <YAxis axisLine={false} tickLine={false} />
                <Tooltip cursor={{fill: '#f1f5f9'}} contentStyle={{borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)'}} />
                <Bar dataKey="reviews" fill="url(#reviewsDailyGradient)" radius={[4, 4, 0, 0]} maxBarSize={50} />
              </BarChart>
            </ResponsiveContainer>
          </div>
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
        <div className="bg-white p-6 border rounded-2xl shadow-sm">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-bold text-gray-800">Top công ty điểm cao</h3>
            <select value={topLimit} onChange={(e) => setTopLimit(Number(e.target.value))} className="border rounded-lg px-2 py-1 text-sm">
              <option value={10}>Top 10</option>
              <option value={20}>Top 20</option>
              <option value={50}>Top 50</option>
            </select>
          </div>
          <div className="h-[320px] w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={topRated} layout="vertical" margin={{ left: 20, right: 20 }}>
                <defs>
                  <linearGradient id="topHighGradient" x1="0" y1="0" x2="1" y2="0">
                    <stop offset="0%" stopColor="#34d399" />
                    <stop offset="100%" stopColor="#059669" />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" horizontal={false} />
                <XAxis type="number" domain={[0, 5]} axisLine={false} tickLine={false} />
                <YAxis type="category" dataKey="name" width={150} axisLine={false} tickLine={false} />
                <Tooltip
                  formatter={(value: any, _name, item: any) => [`${Number(value || 0).toFixed(1)} ★`, `${item?.payload?.reviews || 0} review`]}
                  contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                />
                <Bar dataKey="rating" fill="url(#topHighGradient)" radius={[0, 6, 6, 0]} maxBarSize={24} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        <div className="bg-white p-6 border rounded-2xl shadow-sm">
          <h3 className="text-lg font-bold text-gray-800 mb-4">Top công ty điểm thấp</h3>
          <div className="h-[320px] w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={lowRated} layout="vertical" margin={{ left: 20, right: 20 }}>
                <defs>
                  <linearGradient id="topLowGradient" x1="0" y1="0" x2="1" y2="0">
                    <stop offset="0%" stopColor="#fda4af" />
                    <stop offset="100%" stopColor="#dc2626" />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" horizontal={false} />
                <XAxis type="number" domain={[0, 5]} axisLine={false} tickLine={false} />
                <YAxis type="category" dataKey="name" width={150} axisLine={false} tickLine={false} />
                <Tooltip
                  formatter={(value: any, _name, item: any) => [`${Number(value || 0).toFixed(1)} ★`, `${item?.payload?.reviews || 0} review`]}
                  contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                />
                <Bar dataKey="rating" fill="url(#topLowGradient)" radius={[0, 6, 6, 0]} maxBarSize={24} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
        <div className="bg-white p-6 border rounded-2xl shadow-sm">
          <h3 className="text-lg font-bold text-gray-800 mb-6">Lượt truy cập unique (IP/1h) 7 ngày qua</h3>
          <div className="h-[300px] w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={dailyVisits}>
                <defs>
                  <linearGradient id="visitsDailyGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#67e8f9" />
                    <stop offset="100%" stopColor="#0284c7" />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} />
                <XAxis dataKey="name" axisLine={false} tickLine={false} />
                <YAxis axisLine={false} tickLine={false} />
                <Tooltip cursor={{fill: '#f1f5f9'}} contentStyle={{borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)'}} />
                <Bar dataKey="visits" fill="url(#visitsDailyGradient)" radius={[4, 4, 0, 0]} maxBarSize={50} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        <div className="bg-white p-6 border rounded-2xl shadow-sm">
          <h3 className="text-lg font-bold text-gray-800 mb-6">Lượt truy cập unique (IP/1h) theo tháng (6 tháng)</h3>
          <div className="h-[300px] w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={monthlyVisits}>
                <defs>
                  <linearGradient id="visitsMonthlyGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#a5b4fc" />
                    <stop offset="100%" stopColor="#4f46e5" />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} />
                <XAxis dataKey="name" axisLine={false} tickLine={false} />
                <YAxis axisLine={false} tickLine={false} />
                <Tooltip cursor={{fill: '#f1f5f9'}} contentStyle={{borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)'}} />
                <Bar dataKey="visits" fill="url(#visitsMonthlyGradient)" radius={[4, 4, 0, 0]} maxBarSize={50} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>
    </div>
  )
}
