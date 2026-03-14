import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import { Users, Building2, MessageSquare, Star } from "lucide-react"
import { useEffect, useState } from 'react'
import axios from 'axios'

interface DailyReviewRow {
  day: string
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

  useEffect(() => {
    axios.get(`${API_BASE}/companies/stats/summary`).then((res) => {
      setStats(res.data)
    })

    axios.get(`${API_BASE}/reviews/stats/daily?days=7`).then((res) => {
      const mapped = ((res.data.data || []) as DailyReviewRow[]).map((row) => {
        const d = new Date(row.day)
        const weekday = d.toLocaleDateString('vi-VN', { weekday: 'short' })
        return { name: weekday, reviews: Number(row.reviews || 0) }
      })
      setDailyData(mapped)
    })
  }, [])

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
            <p className="text-2xl font-bold text-gray-900">3</p>
          </div>
        </div>
      </div>

      <div className="bg-white p-6 border rounded-2xl shadow-sm">
        <h3 className="text-lg font-bold text-gray-800 mb-6">Lượng Review 7 ngày qua</h3>
        <div className="h-[400px] w-full">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={dailyData}>
              <CartesianGrid strokeDasharray="3 3" vertical={false} />
              <XAxis dataKey="name" axisLine={false} tickLine={false} />
              <YAxis axisLine={false} tickLine={false} />
              <Tooltip cursor={{fill: '#f1f5f9'}} contentStyle={{borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)'}} />
              <Bar dataKey="reviews" fill="#3b82f6" radius={[4, 4, 0, 0]} maxBarSize={50} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  )
}
