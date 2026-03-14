import { UserCircle, LogOut, ShieldCheck, Mail } from "lucide-react";
import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useGoogleLogin } from "@react-oauth/google";
import api from "../api";

export default function Profile() {
  const [user, setUser] = useState<any>(null);
  const navigate = useNavigate();

  useEffect(() => {
    const storedUser = localStorage.getItem('user');
    if (storedUser) {
      setUser(JSON.parse(storedUser));
    }
  }, []);

  const login = useGoogleLogin({
    onSuccess: async (tokenResponse) => {
      try {
        const res = await api.post('/auth/google', { access_token: tokenResponse.access_token });
        const userData = res.data.user;
        setUser(userData);
        localStorage.setItem('user', JSON.stringify(userData));
        localStorage.setItem('token', res.data.token);
      } catch (error) {
        console.error('Login failed', error);
      }
    },
  });

  const handleLogout = () => {
    localStorage.removeItem('user');
    localStorage.removeItem('token');
    setUser(null);
    navigate('/');
  };

  // Dữ liệu mock: công ty được phép xem
  const allowedCompanies = user?.allowed_companies || [];

  if (!user) {
    return (
      <div className="max-w-2xl mx-auto p-12 text-center bg-white rounded-2xl border shadow-sm mt-10">
        <UserCircle className="w-16 h-16 mx-auto text-gray-300 mb-4" />
        <h2 className="text-2xl font-bold mb-2">Chưa đăng nhập</h2>
        <p className="text-gray-500 mb-6">Bạn cần đăng nhập để xem hồ sơ và các đặc quyền.</p>
        <button 
          onClick={() => login()}
          className="bg-blue-600 text-white px-6 py-2 rounded-lg font-medium hover:bg-blue-700 transition"
        >
          Đăng nhập Google
        </button>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto space-y-8">
      <div className="bg-white rounded-2xl shadow-sm border p-8 relative overflow-hidden">
        <div className="absolute top-0 left-0 w-full h-32 bg-gradient-to-r from-blue-500 to-blue-700 opacity-20"></div>
        
        <div className="relative z-10 flex flex-col md:flex-row items-center md:items-start gap-6 pt-4">
          <img 
            src={user.picture || "https://api.dicebear.com/7.x/avataaars/svg?seed=Felix"} 
            alt="Avatar" 
            className="w-24 h-24 rounded-full border-4 border-white shadow-md bg-white"
          />
          <div className="flex-1 text-center md:text-left">
            <h1 className="text-3xl font-bold text-gray-900 mb-1">{user.name || "Người dùng ẩn danh"}</h1>
            <p className="text-gray-500 flex items-center justify-center md:justify-start gap-2 mb-4">
              <Mail className="w-4 h-4"/> {user.email}
            </p>
            <span className="inline-flex items-center gap-1 bg-green-50 text-green-700 px-3 py-1 rounded-full text-sm font-medium border border-green-200">
              <ShieldCheck className="w-4 h-4" /> Tài khoản Google đã xác thực
            </span>
          </div>
          
          <button 
            onClick={handleLogout}
            className="flex items-center gap-2 px-4 py-2 border rounded-lg text-red-600 hover:bg-red-50 font-medium transition"
          >
            <LogOut className="w-4 h-4" /> Đăng xuất
          </button>
        </div>
      </div>

      <div className="bg-white rounded-2xl shadow-sm border p-8">
        <h2 className="text-xl font-bold text-gray-900 border-b pb-4 mb-6">Phân Quyền Xem Của Bạn</h2>
        <p className="text-gray-600 mb-6 text-sm">
          Mặc định bạn có thể xem review ẩn danh. Dưới đây là danh sách các công ty mà Admin đã cấp quyền cho bạn xem đầy đủ thông tin (không bị che).
        </p>
        
        {allowedCompanies.length > 0 ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {allowedCompanies.map((c: any) => (
              <div key={c.id} className="p-4 border rounded-xl flex items-center justify-between bg-slate-50">
                <span className="font-medium text-gray-800">{c.name}</span>
                <span className="text-xs bg-blue-100 text-blue-700 px-2 py-1 rounded-md font-bold">FULL ACCESS</span>
              </div>
            ))}
          </div>
        ) : (
          <div className="p-4 border border-dashed rounded-xl bg-slate-50 text-center text-gray-500">
            Bạn chưa được cấp quyền xem full data đặc biệt cho công ty nào. (Xem mock ở database)
          </div>
        )}
      </div>
    </div>
  )
}
