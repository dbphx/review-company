# Admin Frontend

React + Vite + TypeScript frontend cho quản trị.

## Chạy local
- `npm install`
- `npm run dev`

Mặc định chạy tại `http://localhost:5174`.

## Chức năng
- Login admin/mod bằng email + password hash.
- Dashboard số liệu thật từ backend.
- Biểu đồ review/visit theo ngày, visit theo tháng.
- Top công ty điểm cao/thấp theo mode dữ liệu, chọn top 10/20/50.
- Quản lý review có filter công ty/ngày/seed version.
- Quản lý công ty (thêm/sửa/xóa).
- Quản lý tài khoản admin/mod.
- Quản lý phiên hoạt động, thu hồi phiên MOD.
- Switch mode dữ liệu toàn hệ thống (`v1`, `v2`, `all`).
- Trang chi tiết review hỗ trợ comment/reply và moderation.

## Cấu hình môi trường
- `VITE_API_URL` (mặc định `http://localhost:3000/api`)
