# Admin Frontend Requirements - ReactJS

Dự án Admin Frontend được chạy tách biệt với hệ thống người dùng để đảm bảo hiệu năng và bảo mật (cổng 5174).

## 1. Công Nghệ Sử Dụng
- **Framework Core:** ReactJS + Vite + TypeScript.
- **UI Library:** Lucide React, Recharts.
- **Styling:** Tailwind CSS.
- **Routing:** React Router DOM.
- **HTTP client:** Axios.

## 2. Các Tính Năng
- **Xác Thực (Authentication):** Đăng nhập bằng Email/Mật khẩu cho tài khoản quản trị.
- **Bảo vệ phiên admin:** token JWT lưu localStorage, route admin có guard `RequireAdminAuth`.
- **Dashboard Thống Kê:**
  - Tổng số: Công ty, Reviews, Điểm trung bình hệ thống từ API backend.
  - Biểu đồ: Lượng Review 7 ngày gần nhất từ API thực (`/api/reviews/stats/daily`).
- **Quản Lý Công Ty:**
  - Xem/Sửa/Thêm/Xóa thông tin công ty.
  - Cấu hình đặc quyền "Full Access": Cấp quyền cho user cụ thể được xem đầy đủ thông tin review của công ty (nếu công ty đó thuộc dạng cần bảo mật).
- **Quản Lý Review/Comment:**
  - Danh sách review có phân trang, tổng số, link qua portal.
  - Trang chi tiết review hiển thị UI gần giống portal.
  - Có thể bình luận/reply trong chi tiết review.
  - Có thể xóa review/comment (được backend kiểm tra RBAC admin).

## 3. Error UX
- Có route-level `errorElement` để hiển thị trang lỗi thân thiện.
