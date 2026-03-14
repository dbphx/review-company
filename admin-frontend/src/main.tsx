import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import AppLayout from './components/layout/AppLayout.tsx'
import RequireAdminAuth from './components/layout/RequireAdminAuth.tsx'
import Dashboard from './pages/Dashboard.tsx'
import Reviews from './pages/Reviews.tsx'
import ReviewDetail from './pages/ReviewDetail.tsx'
import Companies from './pages/Companies.tsx'
import AdminUsers from './pages/AdminUsers.tsx'
import ActiveSessions from './pages/ActiveSessions.tsx'
import Login from './pages/Login.tsx'
import RouteError from './pages/RouteError.tsx'
import { ToastProvider } from './components/ui/ToastProvider.tsx'
import './index.css'

const router = createBrowserRouter([
  {
    path: "/login",
    element: <Login />,
    errorElement: <RouteError />,
  },
  {
    element: <RequireAdminAuth />,
    errorElement: <RouteError />,
    children: [
      {
        path: "/",
        element: <AppLayout />,
        errorElement: <RouteError />,
        children: [
          {
            path: "/",
            element: <Dashboard />,
          },
          {
            path: "/reviews",
            element: <Reviews />,
          },
          {
            path: "/reviews/:id",
            element: <ReviewDetail />,
          },
          {
            path: "/companies",
            element: <Companies />,
          },
          {
            path: "/admin-users",
            element: <AdminUsers />,
          },
          {
            path: "/active-sessions",
            element: <ActiveSessions />,
          },
        ],
      }
    ],
  },
])

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ToastProvider>
      <RouterProvider router={router} />
    </ToastProvider>
  </StrictMode>,
)
