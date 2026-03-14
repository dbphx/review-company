import { Link } from 'react-router-dom';
import { useGoogleLogin } from '@react-oauth/google';
import { UserCircle } from 'lucide-react';
import { useState, useEffect } from 'react';
import api from '../../api';

export default function Navbar() {
  const [user, setUser] = useState<any>(null);

  useEffect(() => {
    const storedUser = localStorage.getItem('user');
    if (storedUser) {
      setUser(JSON.parse(storedUser));
    }
  }, []);

  const login = useGoogleLogin({
    onSuccess: async (tokenResponse) => {
      try {
        // Send token to backend to verify and get/create user
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

  return (
    <nav className="bg-white border-b sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16">
          <div className="flex">
            <Link to="/" className="flex-shrink-0 flex items-center font-extrabold text-2xl tracking-tight text-teal-700">
              <img src="/reviewct.svg" alt="ReviewCT" className="w-9 h-9 mr-3 drop-shadow-sm" />
              <span className="bg-gradient-to-r from-teal-700 via-sky-600 to-emerald-600 bg-clip-text text-transparent">ReviewCT</span>
            </Link>
          </div>
          <div className="flex items-center gap-4">
            {user ? (
              <Link to="/profile" className="flex items-center gap-2 text-sm font-medium text-gray-700 hover:text-blue-600">
                <img src={user.picture} alt="Avatar" className="w-8 h-8 rounded-full border border-gray-200" />
                <span>{user.name}</span>
              </Link>
            ) : (
              <button 
                onClick={() => login()}
                className="flex items-center gap-2 text-sm font-medium bg-gray-50 text-gray-700 border hover:bg-gray-100 px-4 py-2 rounded-lg transition"
              >
                <UserCircle className="w-4 h-4" /> Đăng nhập
              </button>
            )}
          </div>
        </div>
      </div>
    </nav>
  );
}
