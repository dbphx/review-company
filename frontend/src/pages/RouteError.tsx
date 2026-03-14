import { isRouteErrorResponse, useRouteError, Link } from "react-router-dom"

export default function RouteError() {
  const error = useRouteError()

  let title = "Đã xảy ra lỗi"
  let message = "Trang gặp lỗi không mong muốn. Vui lòng thử lại."

  if (isRouteErrorResponse(error)) {
    title = `${error.status} ${error.statusText}`
    message = typeof error.data === "string" ? error.data : message
  } else if (error instanceof Error) {
    message = error.message
  }

  return (
    <div className="min-h-[60vh] flex items-center justify-center px-4">
      <div className="max-w-lg w-full bg-white border rounded-2xl p-8 shadow-sm text-center">
        <h1 className="text-2xl font-bold text-slate-900">{title}</h1>
        <p className="text-slate-600 mt-3">{message}</p>
        <Link to="/" className="inline-block mt-6 px-4 py-2 rounded-lg bg-blue-600 text-white hover:bg-blue-700">
          Về trang chủ
        </Link>
      </div>
    </div>
  )
}
