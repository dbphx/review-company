# Admin Frontend

React + Vite + TypeScript frontend cho quản trị.

## Chạy local
- `npm install`
- `npm run dev`

Mặc định chạy tại `http://localhost:5174`.

## Chức năng
- Login admin bằng email/mật khẩu.
- Dashboard số liệu thật từ backend.
- Biểu đồ review 7 ngày lấy từ API thống kê.
- Quản lý review có phân trang.
- Trang chi tiết review (UI gần portal), hỗ trợ comment/reply.
- Xóa review/comment (backend kiểm tra RBAC admin).

## Cấu hình môi trường
- `VITE_API_URL` (mặc định `http://localhost:3000/api`)
