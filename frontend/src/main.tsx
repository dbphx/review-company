import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import { GoogleOAuthProvider } from '@react-oauth/google'
import App from './App.tsx'
import './index.css'
import Home from './pages/Home.tsx'
import CompanyDetail from './pages/CompanyDetail.tsx'
import Profile from './pages/Profile.tsx'
import RouteError from './pages/RouteError.tsx'
import { ToastProvider } from './components/ui/ToastProvider.tsx'

const router = createBrowserRouter([
  {
    path: "/",
    element: <App />,
    errorElement: <RouteError />,
    children: [
      {
        path: "/",
        element: <Home />,
      },
      {
        path: "/company/:id",
        element: <CompanyDetail />,
      },
      {
        path: "/profile",
        element: <Profile />,
      }
    ]
  }
])

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ToastProvider>
      <GoogleOAuthProvider clientId={import.meta.env.VITE_GOOGLE_CLIENT_ID || "MOCK_CLIENT_ID"}>
        <RouterProvider router={router} />
      </GoogleOAuthProvider>
    </ToastProvider>
  </StrictMode>,
)
